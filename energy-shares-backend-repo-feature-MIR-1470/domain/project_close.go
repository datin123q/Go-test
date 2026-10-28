package domain

import (
	"context"
	"strings"

	"github.com/pkg/errors"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

// MIR-272: 運用終了・償還完了へ進める前の前提チェック。どちらも、チェックの前に
// ファンドの行をロックし、チェックに使う読み取りも FOR UPDATE にする。ロックが無いと、
// チェックが通った直後に入金の自動照合が申込中→成立にした出資が、どこからも処理されずに
// 浮く(お金は引かれているのに分配も返金も無い)。MySQL の REPEATABLE READ では、同じ
// トランザクションの先頭で読んだ値(ファンドの取得など)が固定されるため、通常の読み取りでは
// ロックを取った後の最新の状態を見られない。

// closeConfirmCtxKey は、償還完了で「元本が残っている/償還予定が無い」確認に同意済みであることを
// context で運ぶためのキー。ChangeStatus の引数を増やさずに、API だけが立てられるようにする。
type closeConfirmCtxKey struct{}

// WithRemainingPrincipalConfirmed は、償還完了のときに元本が残っていても進める(画面の確認に
// 同意した)ことを示す context を返す。
func WithRemainingPrincipalConfirmed(ctx context.Context) context.Context {
	return context.WithValue(ctx, closeConfirmCtxKey{}, true)
}

func isRemainingPrincipalConfirmed(ctx context.Context) bool {
	v, _ := ctx.Value(closeConfirmCtxKey{}).(bool)
	return v
}

// operationEndBlockingInvestStatuses は、出資実行が済んでいない出資の状態。運用終了にすると出資実行
// (運用中のときだけ動く)も予定の作成(出資実行の完了が前提)もできなくなるため、残っていれば拒否する。
var operationEndBlockingInvestStatuses = []INVEST_STATUS{SUCCESS_INVEST, REQUEST_INVEST}

// lockProjectRowForTransition は、ファンドの行をロックして、そのときの状態を返す。to への遷移が許されない
// 状態(画面を開いている間に別の管理者やジョブが変えた)なら、前提チェックの前に
// ErrorProjectStatusChanged を返す。確認の文言などを、すでに通らない遷移に対して出さないため。
// SQLite(単体テスト)は FOR UPDATE が構文非対応のため省略する。
func lockProjectRowForTransition(ctx context.Context, projectId int64, to PROJECT_STATUS) error {
	suffix, err := forUpdateSuffix(ctx)
	if err != nil {
		return err
	}

	var current PROJECT_STATUS
	has, err := factory.DB(ctx).SQL("SELECT project_status FROM project WHERE id = ? AND deleted IS NULL"+suffix, projectId).Get(&current)
	if err != nil {
		return err
	}
	if !has {
		return factory.ErrorNotFound.New(errors.New("project not found"))
	}

	for _, from := range allowedFromStatuses(to) {
		if from == current {
			return nil
		}
	}
	return factory.ErrorProjectStatusChanged.New(errors.Errorf("project status is %s, cannot change to %s", current, to))
}

// forUpdateSuffix は、MySQL のときだけ FOR UPDATE を返す(SQLite は構文非対応)。
func forUpdateSuffix(ctx context.Context) (string, error) {
	engine, err := factory.Engine(ctx)
	if err != nil {
		return "", err
	}
	if engine.DriverName() == "sqlite3" {
		return "", nil
	}
	return " FOR UPDATE", nil
}

// countInvestmentsBlockingOperationEnd は、運用終了の前提(出資実行が要る出資が残っていない)に
// 引っかかる出資の件数を返す。
func countInvestmentsBlockingOperationEnd(ctx context.Context, projectId int64) (int64, error) {
	suffix, err := forUpdateSuffix(ctx)
	if err != nil {
		return 0, err
	}

	// 状態の数はプロダクトごとに違う(ZOLOME は自己募集の当選も数える)ため、プレースホルダーは配列の長さから作る。
	args := []interface{}{projectId}
	for _, status := range operationEndBlockingInvestStatuses {
		args = append(args, status)
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?, ", len(operationEndBlockingInvestStatuses)), ", ")

	var ids []string
	err = factory.DB(ctx).
		SQL("SELECT id FROM share WHERE project_id = ? AND invest_status IN ("+placeholders+") AND deleted IS NULL"+suffix, args...).
		Find(&ids)
	if err != nil {
		return 0, err
	}
	return int64(len(ids)), nil
}

// checkOperationEnd は、運用終了にしてよいかを判定する(MIR-272)。引数だけで決まる。
func checkOperationEnd(blockingInvestCount int64) error {
	if blockingInvestCount > 0 {
		return factory.ErrorProjectOperationEndBlocked.New(errors.Errorf("%d investments still need to be executed", blockingInvestCount))
	}
	return nil
}

// CloseCheck は、償還完了にする前の確認の材料(MIR-272)。
type CloseCheck struct {
	// UnexecutedScheduleCount は、未実施(RETURN_SCHEDULED)の分配・償還予定の件数。0 でなければ拒否する。
	UnexecutedScheduleCount int64 `json:"unexecuted_schedule_count"`
	// HasSchedule は、分配・償還予定(分配だけの予定も含む)が1件でもあるか。
	HasSchedule bool `json:"has_schedule"`
	// RemainingPrincipal は、最後の予定(execute_date 降順・id 降順)の元本残高。予定が無ければ 0。
	// 分配だけの予定も、直前の予定の残元本(無ければ出資額)を引き継ぐ。
	RemainingPrincipal int64 `json:"remaining_principal"`
	// BalanceUnknown は、最後の予定の元本残高が NULL のとき true(予定の作成は必ず値を書くため通常は起きない)。
	// 残元本が分からないので、確認が要る側に倒す。
	BalanceUnknown bool `json:"balance_unknown"`
	// ConfirmationRequired は、NeedsConfirmation の結果(GetCloseCheck が詰める)。画面は確認が要るかをこれだけで
	// 決める。判定を画面にも書くと、条件を足したときにずれて、画面から閉じられなくなる。
	ConfirmationRequired bool `json:"confirmation_required"`
}

// NeedsConfirmation は、元本が残っている(ジョブが出資を償還済みにする条件 <= 0 の逆)か、
// 分配・償還の予定が1件も無い(または残元本が分からない)ために、押す前に確認が要るかを返す。
// 償還の途中(予定を全部作り終える前)に押すと、以後は口数0で予定を作れず、残りの元本を
// システムから払えなくなり、元にも戻せないため。
func (c CloseCheck) NeedsConfirmation() bool {
	return !c.HasSchedule || c.BalanceUnknown || c.RemainingPrincipal > 0
}

// Check は、償還完了へ進めてよいかを判定する(MIR-272)。未実施の予定があれば拒否、確認が要るのに
// 同意が無ければ ErrorProjectCloseNeedsConfirmation を返す。
func (c CloseCheck) Check(confirmed bool) error {
	if c.UnexecutedScheduleCount > 0 {
		return factory.ErrorProjectCloseBlockedBySchedule.New(errors.New("repayment dividend schedule is not completed"))
	}
	if c.NeedsConfirmation() && !confirmed {
		return factory.ErrorProjectCloseNeedsConfirmation.New(errors.Errorf("remaining principal %d (has schedule: %t, balance unknown: %t)", c.RemainingPrincipal, c.HasSchedule, c.BalanceUnknown))
	}
	return nil
}

// GetCloseCheck は、償還完了にする前の確認の材料を返す。lock が true のときは、読み取りを
// FOR UPDATE にする(状態を変える処理の中で使う)。画面に残元本を出すだけなら false。
func (Project) GetCloseCheck(ctx context.Context, projectId int64, lock bool) (CloseCheck, error) {
	suffix := ""
	if lock {
		var err error
		if suffix, err = forUpdateSuffix(ctx); err != nil {
			return CloseCheck{}, err
		}
	}

	var scheduledIds []int64
	if err := factory.DB(ctx).
		SQL("SELECT id FROM repayment_dividend_schedule WHERE project_id = ? AND status = ? AND deleted IS NULL"+suffix,
			projectId, RETURN_SCHEDULED).
		Find(&scheduledIds); err != nil {
		return CloseCheck{}, errors.Wrap(err, "fail to get count of repayment dividend schedule")
	}

	// investment_balance は NULL を許す列。構造体(int64)で受けると NULL が 0 に見えるため、ポインタで受ける。
	var latest []struct {
		InvestmentBalance *int64 `xorm:"investment_balance"`
	}
	if err := factory.DB(ctx).
		SQL("SELECT investment_balance FROM repayment_dividend_schedule WHERE project_id = ? AND deleted IS NULL ORDER BY execute_date DESC, id DESC LIMIT 1"+suffix,
			projectId).
		Find(&latest); err != nil {
		return CloseCheck{}, errors.Wrap(err, "fail to get latest repayment dividend schedule")
	}

	check := CloseCheck{UnexecutedScheduleCount: int64(len(scheduledIds)), HasSchedule: len(latest) > 0}
	if len(latest) > 0 {
		if latest[0].InvestmentBalance == nil {
			check.BalanceUnknown = true
		} else {
			check.RemainingPrincipal = *latest[0].InvestmentBalance
		}
	}
	check.ConfirmationRequired = check.NeedsConfirmation()
	return check, nil
}
