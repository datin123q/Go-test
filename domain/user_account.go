package domain

import (
	"context"
	"encoding/csv"
	"fmt"
	"net/http"
	"reflect"
	"time"

	"github.com/samber/lo"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/csvutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/modelhelper"
)

const USER_ACCOUNT_COLUMN = "user_account.id,user_account.user_id,user.name as user_name, user.user_type as user_type,user_account.balance,user_account.updated,user_account.version"

// 사용자 어카운트에는 보유 잔액 및 입출금 내역이 존재 한다.
type UserAccount struct {
	Id                       int64              `json:"id" xorm:"pk autoincr" swagger:"desc(ID)"`
	UserId                   int64              `json:"user_id" xorm:"index" swagger:"desc(ユーザーID)"`
	UserName                 string             `json:"user_name" xorm:"<-" swagger:"desc(ユーザー名)"`
	UserType                 USER_TYPE          `json:"user_type" xorm:"<-"`
	Balance                  *int64             `json:"balance" xorm:"balance" swagger:"desc(投資可能金額)"`
	DepositBalanceCalculated *int64             `json:"deposit_balance,omitempty" xorm:"'deposit_balance' <-" swagger:"desc(投資申請中/出金申請中を含めた残高)"`
	Created                  jsonutils.JSONTime `json:"-" xorm:"created"`
	Updated                  jsonutils.JSONTime `json:"updated" xorm:"updated" swagger:"desc(更新時間)"`
	Version                  int64              `json:"version" xorm:"version" swagger:"desc(バージョン)"`
}

func (u UserAccount) IsEmpty() bool {
	return u.Id == 0
}

func (UserAccount) TableName() string {
	return "user_account"
}

func (u UserAccount) SelectAll() string {
	return u.TableName() + ".*"
}

func (a *UserAccount) Create(ctx context.Context) (int64, error) {
	q := factory.DB(ctx)
	return q.Insert(a)
}

func (a UserAccount) GetByUserId(ctx context.Context, id int64) (*UserAccount, error) {
	var account UserAccount

	q := factory.DB(ctx).Table("user_account").Select(USER_ACCOUNT_COLUMN).Join("INNER", "user", "user.id = user_account.user_id")
	if has, err := q.Where("user_id = ?", id).Get(&account); err != nil {
		return nil, err
	} else if !has {
		return nil, errors.New("Not Found User Account")
	}
	return &account, nil
}

// userAccountDepositBalanceExpr は、投資申請中・出金申請中の金額を含めた
// 残高を計算する SQL 式を返す。
//
// 画面(GetAll)と CSV(DownloadCsv)がこの式を別々に持っていたため、CSV 側の
// 更新が漏れて画面と金額・件数が食い違っていた(MIR-516)。呼び出し元を
// 増やす場合も、式を直接コピーせず必ずこの関数を呼ぶこと。
//
// 生の SQL サブクエリのため、xorm の論理削除の自動除外(deleted カラムの
// 自動条件付与)は効かない。share / deposit_request / withdrawal は
// いずれも deleted カラムを持つため、明示的に除外条件を足す。
//
// xorm の内部実装(internal/statements/statement.go の CondDeleted)は
// 「未削除」を `deleted = ゼロ時刻('0001-01-01 00:00:00') OR deleted
// IS NULL` の OR 条件で判定している。本番データ(ミラリタ)を確認した
// 限りではいずれのテーブルも削除済み行は無く、既存行は全て NULL(ゼロ
// 時刻のパターンは0件)だが、「いま0件」は将来の保証にならないため、
// xorm と同じ OR 条件に揃えておく。ZOLOME/IKKAQ 本番の実データは
// 未確認。
//
// なお、デポジット詳細画面(GetOne)との非対称は今回の変更で入れ替わる。
// withdrawal の集計(Withdrawal{}.SumInt)は Deleted フィールドを持つ
// ビーンに対する通常の xorm 呼び出しのため deleted が自動除外され、
// 今回この式にも同じ除外を明示したことで withdrawal 側は詳細画面と
// 揃う。一方 share/deposit_request の集計(SumDepositAmountOfRequestInvest
// など)は集計専用のビーン(UserShareAmountInfo、Deleted フィールドを
// 持たない)を使っており、詳細画面側は論理削除を除外しない。つまり
// share/deposit_request については、一覧・CSV(この式、除外する)と
// 詳細画面(除外しない)の非対称が新たに生じる。詳細画面側の対応は
// MIR-516 のスコープ外として別途判断する。
func userAccountDepositBalanceExpr() string {
	// xorm の CondDeleted(internal/statements/statement.go)が
	// datetime 型の deleted カラムに使うゼロ時刻。timestamp 型なら
	// 別の値(Unix epoch)になるが、本プロジェクトの deleted カラムは
	// いずれも datetime 型なのでこの値で揃える。
	const zeroDeletedTime = "0001-01-01 00:00:00"
	notDeleted := func(table string) string {
		return fmt.Sprintf("(%[1]s.deleted IS NULL OR %[1]s.deleted = '%[2]s')", table, zeroDeletedTime)
	}
	return fmt.Sprintf(`IFNULL(user_account.balance, 0) +
		(SELECT IFNULL(SUM(deposit_request.deposit_amount), 0)
		 FROM share
		 INNER JOIN deposit_request ON share.id = deposit_request.share_id
		 WHERE share.owner_id = user_account.user_id
		   AND share.invest_status IN ('%[1]s','%[2]s')
		   AND %[6]s
		   AND %[7]s) +
		(SELECT IFNULL(SUM(withdrawal.amount), 0)
		 FROM withdrawal
		 WHERE withdrawal.investor_id = user_account.user_id
		   AND withdrawal.status IN ('%[3]s','%[4]s','%[5]s')
		   AND %[8]s
		)`,
		REQUEST_INVEST,
		SUCCESS_INVEST,
		WITHDRAWAL_STATUS_INCOMPLETE,
		WITHDRAWAL_STATUS_WAIT,
		WITHDRAWAL_STATUS_FAIL,
		notDeleted("share"),
		notDeleted("deposit_request"),
		notDeleted("withdrawal"),
	)
}

// userAccountListQuery は、画面一覧(GetAll)と CSV 出力(DownloadCsv)で
// 共通のクエリを組み立てる。deposit_balance の計算式・絞り込みを1箇所に
// まとめ、両方から呼ぶことで画面と CSV の金額・件数の食い違いを防ぐ
// (MIR-516)。
//
// 返り値は Find/FindAndCount を1回呼ぶまでの使い捨て。xorm は Find 系の
// 呼び出し後に Select/OrderBy/Limit 等をリセットして内部で Close するため、
// 同じ返り値で2回目の Find/FindAndCount を呼ぶと空の結果になる。
func userAccountListQuery(ctx context.Context, sortby, order string, userId, userName *string, startBalance, endBalance *int) (xorm.Interface, error) {
	expr := userAccountDepositBalanceExpr()
	selectColumns := USER_ACCOUNT_COLUMN + "," + expr + " AS deposit_balance"
	q := factory.DB(ctx).Table("user_account").Select(selectColumns).Join("INNER", "user", "user.id = user_account.user_id")

	if err := modelhelper.SetSortOrder(q, sortby, order); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SearchEqQuery(q, "user_account", "user_id", userId); err != nil {
		return nil, errors.Wrap(err, "fail user id eq search ")
	}

	if err := modelhelper.SearchLikeQuery(q, "user", "name", userName); err != nil {
		return nil, errors.Wrap(err, "fail user name like search")
	}

	// 絞り込みも計算後の deposit_balance を対象にする。素の balance は
	// NULL の行を落としてしまい、画面(IFNULL で0に変換してから比較)と
	// 対象件数がずれるため(MIR-516)。
	if err := modelhelper.SetBetweenRange(q, "("+expr+")", startBalance, endBalance); err != nil {
		return nil, errors.Wrap(err, "fail balance between range search")
	}

	return q, nil
}

// applyCalculatedUserAccountBalance は、計算後の deposit_balance を
// Balance へ上書きする。画面と CSV の両方でこの関数を通すことで、
// 表示される金額を一致させる(MIR-516)。
func applyCalculatedUserAccountBalance(userList []UserAccount) {
	for i := range userList {
		if userList[i].DepositBalanceCalculated != nil {
			userList[i].Balance = userList[i].DepositBalanceCalculated
		}
	}
}

func (UserAccount) GetAll(ctx context.Context, sortby, order string, offset, limit int, userId, userName *string, startBalance, endBalance *int) (int64, []UserAccount, error) {
	q, err := userAccountListQuery(ctx, sortby, order, userId, userName, startBalance, endBalance)
	if err != nil {
		return 0, nil, err
	}

	var userList []UserAccount
	totalCount, err := q.Limit(limit, offset).FindAndCount(&userList)
	if err != nil {
		return 0, nil, err
	}

	applyCalculatedUserAccountBalance(userList)

	return totalCount, userList, nil
}

func (UserAccount) DownloadCsv(ctx context.Context, w http.ResponseWriter, sortby, order string, userId, userName *string, startBalance, endBalance *int) error {
	q, err := userAccountListQuery(ctx, sortby, order, userId, userName, startBalance, endBalance)
	if err != nil {
		return err
	}

	var userList []UserAccount
	// CSV は件数を使わないため FindAndCount ではなく Find にする。
	// FindAndCount は同じ WHERE で件数確認用の COUNT クエリも実行するが、
	// CSV 側はその値を使っていなかった(元のコードも戻り値を _ で捨てて
	// いた)。使わない COUNT クエリを1本減らす(MIR-516)。
	if err := q.Find(&userList); err != nil {
		return err
	}

	applyCalculatedUserAccountBalance(userList)

	records := UserAccount{}.PrepareAndReturnCsv(userList)
	csvutils.WriteHeader(w, "energyshares_invest_list.csv")

	return csv.NewWriter(w).WriteAll(records)
}

func (UserAccount) PrepareAndReturnCsv(items []UserAccount) [][]string {
	type userAccount struct {
		UserId   int64  `json:"user_id" i18n:"model.user_account.user_id"`     // 투자자 아이디
		UserName string `json:"user_name" i18n:"model.user_account.user_name"` // 투자자  명
		Balance  int64  `json:"email" i18n:"model.user_account.balance"`       // 예치금 잔고
		Updated  string `json:"phone" i18n:"model.user_account.updated"`       // 최근 업데이트
	}

	var accounts []userAccount
	for _, item := range items {
		account := userAccount{
			UserName: item.UserName,
			Balance:  lo.FromPtr(item.Balance),
			UserId:   item.UserId,
			Updated:  item.Updated.ToString(),
		}
		accounts = append(accounts, account)
	}

	cols := factory.ReflectJsonFields(reflect.ValueOf(&userAccount{}).Elem())
	if len(cols) <= 0 {
		return nil
	}

	records := [][]string{}
	records = append(records, cols)

	for _, v := range accounts {
		records = append(records, factory.DataWriteToCsv(reflect.ValueOf(&v).Elem()))
	}

	return records
}

func (a UserAccount) GetById(ctx context.Context, id int64) (*UserAccount, error) {
	var account UserAccount
	q := factory.DB(ctx).Table("user_account").Select(USER_ACCOUNT_COLUMN).Join("INNER", "user", "user.id = user_account.user_id")

	if has, err := q.ID(id).Get(&account); err != nil {
		return nil, err
	} else if !has {
		return nil, errors.New("Not Found User Account")
	}
	return &account, nil
}

// Updateは呼び出し元が現時点でゼロ件だが、version列を持つUserAccountの
// 公開メソッドとして残っているため、applyBalanceDeltaと同じ形(コピーに
// 適用し、UPDATEが実際に成立した場合のみaへ書き戻す)に揃えておく(MIR-1247)。
// 素通しのままだと、次にこれを使う実装が同じ「0行更新でも成功扱い」の
// 不具合を踏み直すことになる。
func (a *UserAccount) Update(ctx context.Context) (int64, error) {
	working := *a
	idx, err := factory.DB(ctx).ID(a.Id).Update(&working)
	if err != nil {
		return idx, err
	}
	if idx == 0 {
		return idx, errors.Wrapf(ErrVersionConflict, "account_id=%d, version=%d", a.Id, a.Version)
	}
	*a = working
	return idx, nil
}

func (a *UserAccount) WithdrawBalance(ctx context.Context, amount int64, transactionType USER_TRANSACTION_DETAIL_TYPE) (int64, error) {
	if lo.FromPtr(a.Balance) < amount {
		return 0, errors.New("lack of balance")
	}
	return a.applyBalanceDelta(ctx, -amount, true)
}

func (a *UserAccount) DepositBalance(ctx context.Context, amount int64) (int64, error) {
	return a.applyBalanceDelta(ctx, amount, false)
}

// applyBalanceDeltaは、xormの楽観ロック(version列)でaのBalanceをdeltaぶん
// 増減させる。WithdrawBalance・DepositBalanceの共通実装(MIR-1247)。
// aを直接書き換えてUPDATEを試みると、version不一致で0行更新になった場合に
// aのBalance・Version・Updatedが「更新されたつもり」のまま中途半端に残る。
// xormは影響行数を見ずにexecの成否だけでVersion・Updatedを進めるため、
// Balanceだけ戻してもVersionは進んだままになりうる(敵対的検証で実測: 復元
// 前はリトライにより残高が7000→8000に増加する事故を確認した)。ここでは
// 更新対象をaのコピー(working)に対して行い、UPDATEが実際に成立した場合に
// 限ってaへ書き戻す。version不一致・SQLエラーいずれの場合もaは一切変更
// されないため、呼び出し元がaを再取得せずにリトライしても同じ理由で
// 失敗し続けるだけで、古い状態を基準にした誤った額で成功することはない。
// (domain/lottery_csv_upload_job.goのTryMarkRunningFromQueuedも、機構は
// version列ではなくstatus条件のCASだが、影響行数を見て成功時にのみ
// receiverへ書き戻すという結末は同じ)
func (a *UserAccount) applyBalanceDelta(ctx context.Context, delta int64, allCols bool) (int64, error) {
	working := *a
	working.Balance = lo.ToPtr(lo.FromPtr(a.Balance) + delta)

	q := factory.DB(ctx).ID(a.Id)
	if allCols {
		q = q.AllCols()
	}
	idx, err := q.Update(&working)
	if err != nil {
		return idx, err
	}
	if idx == 0 {
		return idx, errors.Wrapf(ErrVersionConflict, "account_id=%d, version=%d, delta=%d", a.Id, a.Version, delta)
	}
	*a = working
	return idx, nil
}

// DepositCheckProcess
/*
TODO 참고이슈 : https://illuminarean.atlassian.net/browse/ESJP-497
	해당 로직에서 이메일을 발송하지 않는다고 되어있는데 발송중임
	투자가능금액 부족할 때 이메일은 8-6-3으로 보내는게 기획으로 보이는데 현재는 8-9로 보내고 있어서 확인필요 (내용이 유사해서 둘 중에 하나만 보낼 필요가 있을 것 같음)
	투자가능금액이 충분하면 이미 8-7로 보내고 있는데 이 부분도 재 확인필요 (구현해야한다는 기획과 달리 이미 있음)
*/
// DepositCheckProcessはMIR-1251で、残高の増減から入金要求の状態更新
// (d.Update)までを1つのトランザクションで囲うようにした。開発初期の
// 実装ではbalanceMatch/balanceLackの中だけをトランザクションにしていたが、
// その直後にあるd.Update(入金要求をOK_DEPOSITの最終状態にする書き込み)
// はトランザクションの外にあり、これが失敗すると「残高は引かれ、
// share・Contract・メールも確定しているのに、DepositCheckProcess自体は
// エラーを返す」という不整合が起きる(呼び出し元は投資が失敗したと
// 誤解する)。トランザクション境界を関数全体に広げ、抜けが無いようにした。
//
// この関数は既存のトランザクション(factory.InTransaction(ctx))があれば
// それに乗る。バッチ処理(processDepositRequestItems)から呼ばれる場合は
// 呼び出し元が1件ごとに新規トランザクションを強制するため、1件の失敗が
// 他の件に影響しない。単発呼び出し(投資申込時等、既存のリクエスト単位
// トランザクションに乗せたいケース)ではそのまま親トランザクションに参加する。
func (a *UserAccount) DepositCheckProcess(ctx context.Context, d *DepositRequest, investMailType INVEST_MAIL_TYPE) error {
	logrus.WithFields(logrus.Fields{
		"user_account":    a,
		"deposit_request": d,
		"need_account":    d.NeedAmount(),
		"balance":         a.Balance,
	}).Info("start user account deposit check process share id - > ", d.ShareId)

	if d.NeedAmount() <= 0 {
		logrus.Info("success user account deposit check process share id -> ", d.ShareId)
		return nil
	}

	needAmount := d.NeedAmount()
	availableAmount := lo.FromPtr(a.Balance)

	// balanceLackが必要とするUserVirtualAccount/Projectをトランザクションの
	// 外で先に取得しておく。Project.GetByIdは初回アクセス時にproject_view_
	// countの行を作成する副作用(書き込み)を持つため、トランザクション内で
	// 行うと書き込みロックの保持時間が延び、同じプロジェクトへ同時にアクセス
	// する別処理と競合しやすくなる(並行実行テストでsqliteの"database is
	// locked"として実測)。読み取りは先に済ませ、トランザクションの中では
	// 実際の書き込みだけを行うようにする。
	var (
		uv *UserVirtualAccount
		p  *Project
	)
	if availableAmount < needAmount {
		var err error
		uv, err = UserVirtualAccount{}.GetVaByUserId(ctx, d.UserId, EQUITY)
		if err != nil {
			return errors.Wrap(err, "fail account balance lack process: fail get user virtual account")
		}
		if uv == nil {
			return errors.New("fail account balance lack process: not found user virtual account")
		}
		p, err = Project{}.GetById(ctx, d.ProjectId)
		if err != nil {
			return errors.Wrap(err, "fail account balance lack process: fail get project")
		}
		if p == nil {
			return errors.New("fail account balance lack process: not found project")
		}
	}

	run := func(txCtx context.Context) error {
		if availableAmount >= needAmount { // 필요한 금액이 현재 남은 금액보다 적거나 같을때
			d.DepositAmount = d.DepositAmount + needAmount
			d.DepositCheckStatus = OK_DEPOSIT
			if err := a.balanceMatch(txCtx, needAmount, d, investMailType); err != nil {
				return errors.Wrap(err, "fail account balance match process")
			}
		} else {
			d.DepositAmount = d.DepositAmount + availableAmount
			d.DepositCheckStatus = NO_DEPOSIT
			if err := a.balanceLack(txCtx, availableAmount, d, investMailType, p); err != nil {
				return errors.Wrap(err, "fail account balance lack process")
			}
		}
		d.DepositDate = jsonutils.JSONTime(time.Now())
		// 입금 요청 업데이트
		if err := d.Update(txCtx); err != nil {
			return errors.Wrap(err, "fail deposit request update")
		}
		return nil
	}

	var err error
	if factory.InTransaction(ctx) {
		err = run(ctx)
	} else {
		var engine *xorm.Engine
		engine, err = factory.Engine(ctx)
		if err != nil {
			return factory.ErrorDB.New(err, "Fail Get DB Engine")
		}
		err = factory.NewTxManager(engine).WithTx(ctx, run)
	}
	if err != nil {
		return err
	}

	logrus.Info("success user account deposit check process share id -> ", d.ShareId)
	return nil
}

// balanceMatchはMIR-1251で、残高を引いてから成立処理をする順序に入れ替えた。
// 以前はUpdateInvestStatus(成立メール送信・Contract作成・deposit_requestの
// OK_DEPOSIT化を含む)を先に呼んでおり、直後のWithdrawBalanceが競合等で
// 失敗すると、share・deposit_request・メールは全て確定済みのまま残高だけが
// 引かれない状態になる。deposit_requestは既にOK_DEPOSITへ変わっているため、
// DepositAutoProcessJob(WHERE deposit_check_status = NO_DEPOSIT)にはもう
// 拾われず、再試行の機会が二度と来ない。WithdrawBalanceを先に行えば、
// それが失敗した時点で何も確定しないため、レコードは変更前のNO_DEPOSITの
// まま残り、次回また自然に拾われる。トランザクション境界は呼び出し元の
// DepositCheckProcessが持つため、ここでは新たに張らない。
func (a *UserAccount) balanceMatch(ctx context.Context, amount int64, r *DepositRequest, investMailType INVEST_MAIL_TYPE) error {
	if _, err := a.WithdrawBalance(ctx, amount, WITHDRAWAL_INVESTMENT); err != nil {
		return factory.ErrorDB.New(err, "Fail User Account Withdrawal Balance")
	}
	s, err := Share{}.GetById(ctx, r.ShareId)
	if err != nil {
		return err
	}
	if err := s.UpdateInvestStatus(ctx, SUCCESS_INVEST, 0, investMailType); err != nil {
		return factory.ErrorDB.New(err, "Fail Update Invest Status")
	}
	return nil
}

// balanceLackはMIR-1251で、balanceMatchと同じくトランザクション境界を
// 呼び出し元のDepositCheckProcessに委ねる形にした。残高を引く処理は
// 元々先頭にあり順序自体に問題は無かったが、その後のメール送信を
// AfterCommit経由にして、DBのロールバックでは取り消せないメール送信が
// コミット後まで遅延するようにしている。UserVirtualAccountの存在確認・
// Projectの取得は、Project.GetByIdの副作用(project_view_countの初回
// 作成)を書き込みロックの外に出すため、呼び出し元のDepositCheckProcessが
// トランザクションを開く前に済ませ、pとして受け取る。
func (a *UserAccount) balanceLack(ctx context.Context, amount int64, r *DepositRequest, investMailType INVEST_MAIL_TYPE, p *Project) error {
	if _, err := a.WithdrawBalance(ctx, amount, WITHDRAWAL_INVESTMENT); err != nil {
		return factory.ErrorDB.New(err, "Fail User Account Withdrawal Balance")
	}
	if investMailType == NOT_IMMEDIATE {
		// メール送信はDBのロールバックで取り消せないため、コミット後まで遅延させる。
		_ = factory.AfterCommit(ctx, func(commitCtx context.Context) error {
			mailErr := DepositRequestLackHistory{}.RequestLackDepositMail(commitCtx, r, p)
			if mailErr != nil {
				factory.Logger(commitCtx).WithFields(logrus.Fields{
					"user_id":        r.UserId,
					"project_name":   p.Name,
					"deposit_amount": r.DepositAmount,
					"lack_amount":    r.NeedAmount(),
					"err":            mailErr,
				}).Error("fail send request lack deposit mail")
			}
			return nil
		})
	}

	historyErr := DepositRequestLackHistory{}.CreateDepositRequestLackHistory(ctx, a.UserId, r.NeedAmount(), r.Id, r.ProjectId, r.ProjectType, r.DepositAccountNumber, r.ShareId, jsonutils.JSONTime(time.Now()), r.ExpireDepositDate)
	if historyErr != nil {
		return errors.Wrap(historyErr, "fail create deposit request lack history ")
	}
	return nil
}
