package mysql

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/orm"

	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

type investmentWaitersMysqlRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.InvestmentWaitersRepository {
	return &investmentWaitersMysqlRepository{
		dbFactory: dbFactory,
	}
}

func (i investmentWaitersMysqlRepository) GetAll(ctx context.Context, options ...cdomain.InvestmentWaitersQueryOption) (int64, []domain.InvestmentWaiters, error) {
	db, err := i.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, nil, errors.Wrap(err, "err get db")
	}

	opts := cdomain.InvestmentWaitersQueryOptionFunc()
	for _, opt := range options {
		opt(&opts)
	}

	investmentWaiters := make([]domain.InvestmentWaiters, 0)

	totalCount, err := orm.Table(db, domain.InvestmentWaiters{}).
		Select(
			domain.DEFAULT_INVESTMENT_WAIT_LIST,
			domain.Project{}.TableName()+".project_status AS project_status",
			// investment_waiters_winning・shareへの実JOINは、1申込に複数当選がある場合に
			// 行を水増しするため使わない(MIR-1068)。
			domain.InvestmentWaitersInvestStatusExpr,
			domain.AdvancedQtyExpr,
		).
		InnerJoin(domain.User{}, "investment_waiters.user_id = user.id").
		InnerJoin(domain.Project{}, "investment_waiters.project_id = project.id").
		SortOrderForTable(domain.InvestmentWaiters{}.TableName(), opts.SortBy(), opts.Order()).
		Limit(opts.Limit(), opts.Offset()).
		Equal(domain.InvestmentWaiters{}.TableName()+".project_id", opts.ProjectId()).
		FindAndCount(&investmentWaiters)

	return totalCount, investmentWaiters, err
}

func (i investmentWaitersMysqlRepository) GetById(ctx context.Context, id int64) (domain.InvestmentWaiters, error) {
	db, err := i.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return domain.InvestmentWaiters{}, errors.Wrap(err, "err get db")
	}

	investmentWaiters := domain.InvestmentWaiters{}

	has, err := orm.Table(db, domain.InvestmentWaiters{}).
		Select(
			domain.DEFAULT_INVESTMENT_WAIT_LIST,
			domain.Project{}.TableName()+".project_status AS project_status",
			// investment_waiters_winning・shareへの実JOINは、1申込に複数当選がある場合に
			// 行を水増しするため使わない(MIR-1068)。
			domain.InvestmentWaitersInvestStatusExpr,
			domain.AdvancedQtyExpr,
		).
		InnerJoin(domain.User{}, "investment_waiters.user_id = user.id").
		InnerJoin(domain.Project{}, "investment_waiters.project_id = project.id").
		Id(id).
		Get(&investmentWaiters)

	if err != nil {
		return domain.InvestmentWaiters{}, errors.Wrap(err, "err get investmentWaiters")
	}
	if !has {
		return domain.InvestmentWaiters{}, errors.Wrap(cdomain.ErrItemNotFound, "investmentWaiters not found")
	}

	return investmentWaiters, nil
}

func (i investmentWaitersMysqlRepository) LockById(ctx context.Context, id int64) (domain.InvestmentWaiters, error) {
	db, err := i.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return domain.InvestmentWaiters{}, errors.Wrap(err, "err get db")
	}

	investmentWaiters := domain.InvestmentWaiters{}
	// investment_waiters.* をそのまま使う。investment_waiters.share_idの実カラムは
	// PR4(MIR-1163)で撤去済みのため、ワイルドカードにこのカラムは含まれず、ShareId
	// (xorm:"<-")は常に空文字列のまま返る。以前はここで domain.InvestmentWaitersShareIdExpr を
	// 埋め込み、当選済みかどうかの判定に使える値まで解決していたが、その式が読む
	// investment_waiters_winning への相関サブクエリは、外側のSELECTにFOR UPDATEを付けても
	// ロック読みには昇格しない(MySQLはSELECTリスト内のスカラーサブクエリをロック対象に
	// 含めない)。そのため、呼び出し元のトランザクションがLockByIdより前に他のSELECTを
	// 1回でも実行していると、MySQLのREPEATABLE READでその時点のスナップショットに固定され、
	// 中間テーブル側だけ古いデータを返し続けるバグがあった(敵対的検証でMySQL 8.0にて
	// 実測・再現)。当選済みかどうかの判定は、この関数ではなく、investment_waiters_winning
	// 自体を行ロックで読むCurrentShareIdに任せる(呼び出し元はLockById→CurrentShareIdの順に
	// 同一トランザクション内で呼ぶこと)。
	has, err := db.
		SQL("SELECT investment_waiters.* FROM investment_waiters WHERE investment_waiters.id = ? AND "+domain.NotDeletedCondition("investment_waiters")+" FOR UPDATE", id).
		Get(&investmentWaiters)
	if err != nil {
		return domain.InvestmentWaiters{}, errors.Wrap(err, "err lock investmentWaiters")
	}
	if !has {
		return domain.InvestmentWaiters{}, errors.Wrap(cdomain.ErrItemNotFound, "investmentWaiters not found")
	}

	return investmentWaiters, nil
}

// LockAndSumAdvancedQty は investment_waiters_winning と share を1本のJOINクエリで
// 行ロック(FOR UPDATE)しながら読み、待機者に紐づく当選(1件以上、MIR-1068で複数可)の
// share.qty合計(繰上げ済み累計)を解決する。LockByIdと同一トランザクション内で、
// LockByIdの直後に呼ぶこと。旧列investment_waiters.share_idへのフォールバックは
// PR3(MIR-1163)で削除した。中間テーブルに当選履歴が無ければ0を返す。
//
// winningとshareを別々のクエリに分けず1本のJOINにFOR UPDATEを掛けているのは、
// 呼び出し元のトランザクションがLockByIdより前に他のSELECT(shareを含むJOIN等)を
// 実行していると、MySQLのREPEATABLE READではその時点のスナップショットに固定され、
// winning側だけロック読み(常に最新)でshare側だけ非ロック読み(スナップショット)に
// 分かれてしまうと、直後に別トランザクションが作成したshareのqtyを見落として
// 繰上げ可能残を実際より大きく計算しうるため(①の敵対的検証で発見。MIR-1118の
// CurrentShareIdが踏んだ問題と同じ構図)。
func (i investmentWaitersMysqlRepository) LockAndSumAdvancedQty(ctx context.Context, id int64) (int64, error) {
	db, err := i.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "err get db")
	}

	session := db.Table(new(domain.InvestmentWaitersWinning)).
		Join("INNER", "share", "share.id = investment_waiters_winning.share_id").
		Where(
			"investment_waiters_winning.investment_waiters_id = ? AND "+
				domain.NotDeletedCondition("investment_waiters_winning")+" AND "+
				domain.NotDeletedCondition("share"),
			id,
		)
	if session.Engine().DriverName() != "sqlite3" {
		session = session.ForUpdate()
	}

	sum, err := session.Sum(new(domain.Share), "qty")
	if err != nil {
		return 0, errors.Wrap(err, "err lock and sum investment waiters winning")
	}

	return int64(sum), nil
}

// GetWinningsByWaiterId は、待機者に紐づく当選明細一覧を中間テーブル
// (investment_waiters_winning)のid降順、つまり繰上げを実行した順の新しい順で返す
// (MIR-1068、管理画面の繰上げ履歴表示)。1待機者ぶんの明細を複数行で見せることが
// 目的のクエリなので、GetAll/GetByIdと違いinvestment_waiters_winning・share・
// deposit_requestへの実JOINで良い(水増しではなく複数当選の各行を見せることが要件)。
// deposit_requestはLEFT JOIN(繰上げ直後などdeposit_requestがまだ無い場合を考慮)。
func (i investmentWaitersMysqlRepository) GetWinningsByWaiterId(ctx context.Context, id int64) ([]domain.InvestmentWaitersWinningItem, error) {
	db, err := i.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "err get db")
	}

	items := make([]domain.InvestmentWaitersWinningItem, 0)
	err = db.Table(new(domain.InvestmentWaitersWinning)).
		Select(
			"share.id AS share_id, "+
				"share.qty AS qty, "+
				"share.price AS price, "+
				"share.amount AS amount, "+
				"share.invest_date AS invest_date, "+
				"share.invest_status AS invest_status, "+
				"deposit_request.deposit_check_status AS deposit_check_status",
		).
		Join("INNER", "share", "share.id = investment_waiters_winning.share_id AND "+domain.NotDeletedCondition("share")).
		Join("LEFT", "deposit_request", "deposit_request.share_id = share.id AND "+domain.NotDeletedCondition("deposit_request")).
		Where(
			"investment_waiters_winning.investment_waiters_id = ? AND "+domain.NotDeletedCondition("investment_waiters_winning"),
			id,
		).
		Desc("investment_waiters_winning.id").
		Find(&items)
	if err != nil {
		return nil, errors.Wrap(err, "err get winnings by waiter id")
	}

	return items, nil
}

func (i investmentWaitersMysqlRepository) Update(ctx context.Context, investmentWaiters domain.InvestmentWaiters) (int64, error) {
	db, err := i.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "err get db")
	}

	affected, err := db.ID(investmentWaiters.Id).UseBool().Update(investmentWaiters)
	if err != nil {
		return 0, errors.Wrap(err, "err update investmentWaiters")
	}

	return affected, nil
}
