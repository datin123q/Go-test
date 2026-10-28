package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/samber/lo"
	"github.com/sirupsen/logrus"

	"github.com/pkg/errors"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

type Dividend struct {
	Id                int64                        `json:"id" xorm:"pk autoincr not null"`
	UserId            int64                        `json:"user_id" xorm:"user_id" swagger:"desc(사용자 ID)"`
	SpvId             int64                        `json:"spv_id" xorm:"spv_id" swagger:"desc(SPV ID)"`
	ProjectId         int64                        `json:"project_id" xorm:"project_id" swagger:"desc(상품 ID)"`
	ProjectType       PROJECT_TYPE                 `json:"project_type" xorm:"project_type" swagger:"desc(상품 타입)"`
	SettlementId      int64                        `json:"settlement_id" xorm:"settlement_id" swagger:"desc(결산  ID)"`
	InvestmentAmount  int64                        `json:"investment_amount" xorm:"investment_amount" swagger:"desc(투자금)"`
	InvestmentQty     int64                        `json:"investment_qty" xorm:"investment_qty" swagger:"desc(투자 구수)"`
	Amount            int64                        `json:"amount" xorm:"amount" swagger:"desc(배당/상환금)"`
	Qty               int64                        `json:"qty" xorm:"qty" swagger:"desc(투자 구수)"`
	DepositAmount     int64                        `json:"deposit_amount" xorm:"deposit_amount" swagger:"desc(세후 송금 금액)"`
	Tax               int64                        `json:"tax" xorm:"tax" swagger:"desc(원천 징수 세금)"`
	DividendDate      jsonutils.JSONTime           `json:"dividend_date" xorm:"dividend_date" swagger:"desc(배당일)"`
	ConfirmedDate     jsonutils.JSONTime           `json:"confirmed_date" xorm:"confirmed_date" swagger:"desc(확정일)"`
	TransactionStatus TRANSACTION_STATUS           `json:"transaction_status" xorm:"transaction_status" swagger:"desc(출금 상태)"`
	TransactionType   USER_TRANSACTION_DETAIL_TYPE `json:"transaction_type" xorm:"transaction_type" swagger:"desc(거래 타입)"`
	TransactionDate   jsonutils.JSONTime           `json:"transaction_date" xorm:"transaction_date" swagger:"desc(거래 일시)"`
	TransactionId     string                       `json:"transaction_id" xorm:"UUID NOT NULL INDEX" swagger:"desc(거래 ID)"`
	Created           jsonutils.JSONTime           `json:"-" xorm:"created"`
	Updated           jsonutils.JSONTime           `json:"-" xorm:"updated"`
	Deleted           jsonutils.JSONTime           `json:"-" xorm:"deleted"`
}

type UserDividend struct {
	User     `xorm:"extends"`
	Dividend `xorm:"extends"`
}

// 법정 장부를 위한 투자자별 배당 목록 조회
// Deprecated: Use DividendRepository.GetAll instead
func (Dividend) GetAllByUserId(ctx context.Context, userId int64, startDate, endDate string) ([]Dividend, error) {
	var items []Dividend
	q := factory.DB(ctx).Table("dividend").
		Where("user_id = ?", userId).
		And("dividend_date >= ?", startDate).
		And("dividend_date <= ?", endDate)
	_, err := q.FindAndCount(&items)
	if err != nil {
		return nil, err
	}
	return items, nil
}

// 법정 장부를 위한 개발회사의 배당 목록 조회
// Deprecated: Use DividendRepository.GetAll instead
func (Dividend) GetAllBySpvId(ctx context.Context, spvId int64, startDate, endDate string) ([]Dividend, error) {
	var items []Dividend
	q := factory.DB(ctx).Table("dividend").
		Where("spv_id = ?", spvId).
		And("dividend_date >= ?", startDate).
		And("dividend_date <= ?", endDate)
	_, err := q.FindAndCount(&items)
	if err != nil {
		return nil, err
	}
	return items, nil
}

// GetByRepaymentDividendScheduleId は、指定したスケジュールのうち支払い済み
// (transaction_status = TRANSACTION_OK)でない行だけを返す(MIR-1249)。以前は
// settlement_idだけで絞っており、支払い済みの行も含めて返していたため、
// スケジュールを再実行すると既に成立した配当・償還までもう一度処理して
// しまっていた。TRANSACTION_NOだけに絞らないのは、transaction_status列が
// NULL許容かつデフォルト無しで(energyshares.base.schema.sql)、過去データや
// TRANSACTION_FAILの行が混じっている可能性があるため。それらを対象から
// 漏らすと、支払われていないのにスケジュールがCompletedになってしまう。
func (UserDividend) GetByRepaymentDividendScheduleId(ctx context.Context, scheduleId int64) ([]UserDividend, error) {
	q := factory.DB(ctx)
	var items []UserDividend
	q = q.Table("dividend").
		Select("user.*,dividend.*").
		Join("INNER", "user", "dividend.user_id = user.id")
	q.Where("dividend.settlement_id = ? AND (dividend.transaction_status IS NULL OR dividend.transaction_status != ?)", scheduleId, TRANSACTION_OK)
	err := q.Find(&items)
	if err != nil {
		return nil, err
	}

	return items, nil
}

func (d *Dividend) Update(ctx context.Context) error {
	q := factory.DB(ctx)
	_, err := q.ID(d.Id).Update(d)
	if err != nil {
		return err
	}
	return nil
}

// DoTransactionの第2戻り値ledgerEventFailedは、支払い自体(残高加算・
// transaction_status書込み)は成立したが、顧客元帳イベント(CreateCustomerLedgerDividendEvent)
// の作成にコミット後失敗したことを表す(MIR-1249)。呼び出し元は、この場合を
// エラー(err != nil、支払いそのものの失敗)と混同せず、別枠(eventFailed)として
// 扱うこと。GetByRepaymentDividendScheduleIdはTRANSACTION_OKの行を除外するため
// 再試行では拾えず、この帳簿イベントは失われたままになる。
func (d *UserDividend) DoTransaction(ctx context.Context, status TRANSACTION_STATUS) (ledgerEventFailed bool, err error) {
	d.Dividend.TransactionDate = jsonutils.JSONTime(time.Now())
	d.Dividend.TransactionStatus = status

	if status != TRANSACTION_OK {
		return false, d.Dividend.Update(ctx)
	}

	// 残高加算からtransaction_status書込み(Dividend.Update)までを1つの
	// トランザクションで包む(MIR-1249)。以前は個々のUpdate/Insertがそれぞれ
	// 即コミットされており、途中(残高加算後・TRANSACTION_OK書込み前)で失敗すると、
	// 残高だけ加算済みでtransaction_statusは初期値TRANSACTION_NOのまま残った。
	// この状態は「未着手」の行と区別が付かないため、
	// GetByRepaymentDividendScheduleId(未処理行のみを再取得して再試行する)が
	// 誤ってこの行を拾い直し、二重加算してしまう。
	run := func(txCtx context.Context) error {
		p, err := Project{}.GetById(txCtx, d.ProjectId)
		if err != nil {
			return errors.Wrap(err, "failed to get project by id")
		}
		switch d.ProjectType {
		case EQUITY:
			account, err := UserAccount{}.GetByUserId(txCtx, d.UserId)
			if err != nil {
				return errors.Wrap(err, "failed user account by use.")
			}
			originalBalance := lo.FromPtr(account.Balance)
			account.Balance = lo.ToPtr(originalBalance + d.DepositAmount)

			depositRequests, err := DepositRequest{}.GetNoDepositRequestByUserId(txCtx, d.UserId)
			if err != nil {
				return errors.Wrap(err, "failed to get no deposit request by user id")
			}

			var consumedAmount int64
			for i := range depositRequests {
				before := lo.FromPtr(account.Balance)
				if err := account.DepositCheckProcess(txCtx, &depositRequests[i], NOT_IMMEDIATE); err != nil {
					return errors.Wrap(err, "fail to deposit check process")
				}
				after := lo.FromPtr(account.Balance)
				if after != before {
					consumedAmount += before - after
					if consumedAmount >= d.DepositAmount {
						break
					}
				}
			}

			if consumedAmount == 0 {
				account.Balance = lo.ToPtr(originalBalance)
				if _, err = account.DepositBalance(txCtx, d.DepositAmount); err != nil {
					return errors.Wrap(err, "failed to deposit balance")
				}
			}

			// user_transaction には実入金額（源泉徴収後）を記録する
			if err := (UserTransaction{}).DepositUserTransaction(txCtx, d.UserId, d.DepositAmount, d.ProjectId, d.ProjectType, "", d.Dividend.TransactionType, jsonutils.JSONTime(d.TransactionDate.Time())); err != nil {
				return errors.Wrap(err, "failed to create user transaction")
			}

			// CreateCustomerLedgerDividendEventは外部イベントバスへ同期publishする
			// (EventMessage.PublishEvent)。トランザクション内で直接呼ぶと、後続の
			// 失敗でロールバックされた際にpublish自体は取り消せずDBだけ「未処理」に
			// 戻り、再試行で二重発行する。factory.AfterCommitでコミット後にのみ
			// 実行する。ここでのエラー・panic(0除算 domain/event_message.go:732 など)
			// はDoTransactionのerrには反映させず(残高加算とTRANSACTION_OK書込みは
			// 既にコミット済みで支払いは成立済みのため、errに乗せると「未払い」と
			// 誤って扱われる)、代わりに外側のledgerEventFailedへ記録する。
			factory.AfterCommit(txCtx, func(afterCtx context.Context) (afterErr error) {
				defer func() {
					if r := recover(); r != nil {
						afterErr = nil
						ledgerEventFailed = true
						factory.Logger(afterCtx).WithFields(logrus.Fields{
							"dividend_id": d.Dividend.Id,
							"user_id":     d.UserId,
							"panic":       fmt.Sprintf("%v", r),
						}).Error("Panic recovered while creating customer ledger event after commit (payment already succeeded, not retried)")
					}
				}()
				if err := CreateCustomerLedgerDividendEvent(afterCtx, d, p.Name, account); err != nil {
					ledgerEventFailed = true
					factory.Logger(afterCtx).WithFields(logrus.Fields{
						"dividend_id": d.Dividend.Id,
						"user_id":     d.UserId,
						"err":         err,
					}).Error("Fail to create customer ledger event after commit (payment already succeeded, not retried)")
				}
				return nil
			})
		}

		return d.Dividend.Update(txCtx)
	}

	// 呼び出し元(ctx)が既にトランザクションを持っている場合でも、常に自前で
	// 新しいWithTxを張る(既存トランザクションを再利用しない)。factory.AfterCommit
	// は渡されたctxが*factory.TxContextであることを前提に「コミット後にのみ実行」
	// を保証する仕組みで、呼び出し元のトランザクションをそのまま使い回すと、
	// 型アサーションが外れてAfterCommitが即時実行(コミット前の同期実行)に
	// 落ちてしまい、上のコメントが防ごうとしている二重発行がそのまま起きる。
	// 配当1件ずつを独立してコミットするというこのチケットの設計そのものにも、
	// 常に新規トランザクションを張ることが必要(呼び出し元のトランザクションに
	// 合流すると、1件の失敗が他の会員の分まで巻き戻してしまう)。
	engine, err := factory.Engine(ctx)
	if err != nil {
		return false, errors.Wrap(err, "failed to get db engine")
	}
	// ledgerEventFailedはWithTx内のAfterCommitコールバックで書き込まれる。
	// 戻り値2つの式の中でWithTx呼び出しとまとめて書くと、Go仕様上「関数呼び出し
	// 同士の左から右」以外の評価順は未規定のため、先にWithTxを完了させてから
	// ledgerEventFailedを読む(文を分ける)ことで評価順を確定させる。
	err = factory.NewTxManager(engine).WithTx(ctx, run)
	return ledgerEventFailed, err
}
