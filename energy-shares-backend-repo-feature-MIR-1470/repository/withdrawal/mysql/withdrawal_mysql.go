package mysql

import (
	"context"
	"fmt"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/orm"
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

type withdrawalMysqlRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.WithdrawalRepository {
	return &withdrawalMysqlRepository{
		dbFactory: dbFactory,
	}
}

func (r withdrawalMysqlRepository) Create(ctx context.Context, withdrawal domain.Withdrawal) (int64, error) {
	db, err := r.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "err get db")
	}

	if _, err = db.Insert(&withdrawal); err != nil {
		return 0, errors.Wrap(err, "err insert withdrawal")
	}

	return withdrawal.Id, nil
}

func (r withdrawalMysqlRepository) GetAll(ctx context.Context, options ...cdomain.WithdrawalQueryOption) (int64, []domain.WithdrawalDetail, error) {
	db, err := r.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, nil, errors.Wrap(err, "err get db")
	}

	queryOpts := cdomain.WithdrawalQueryOptionFunc()
	for _, opt := range options {
		opt(&queryOpts)
	}

	withdrawals := make([]domain.WithdrawalDetail, 0)

	sql := orm.Table(db, domain.Withdrawal{}).
		InnerJoin(domain.User{}, fmt.Sprintf("%s.id = %s.investor_id", domain.User{}.TableName(), domain.Withdrawal{}.TableName())).
		Equal(domain.Withdrawal{}.TableName()+".investor_id", queryOpts.InvestorId()).
		Equal(domain.Withdrawal{}.TableName()+".orderer", queryOpts.InvestorName()).
		Between(domain.Withdrawal{}.TableName()+".amount", queryOpts.AmountMin(), queryOpts.AmountMax()).
		Between(domain.Withdrawal{}.TableName()+".request_date", queryOpts.RequestStartDate(), queryOpts.RequestEndDate()).
		Between(domain.Withdrawal{}.TableName()+".complete_date", queryOpts.CompleteStartDate(), queryOpts.CompleteEndDate()).
		Equal(domain.Withdrawal{}.TableName()+".status", queryOpts.Status()).
		SortOrderForTable(domain.Withdrawal{}.TableName(), queryOpts.SortBy(), queryOpts.Order()).
		Limit(queryOpts.Limit(), queryOpts.Offset())

	if queryOpts.CloseId() != nil {
		sql.InnerJoin(domain.WithdrawalCloseMapping{}, fmt.Sprintf("%s.withdrawal_id = %s.id", domain.WithdrawalCloseMapping{}.TableName(), domain.Withdrawal{}.TableName())).
			Equal(domain.WithdrawalCloseMapping{}.TableName()+".close_id", queryOpts.CloseId())
	}

	totalCount, err := sql.FindAndCount(&withdrawals)
	if err != nil {
		return 0, nil, errors.Wrap(err, "err find withdrawals")
	}

	return totalCount, withdrawals, nil
}

func (r withdrawalMysqlRepository) GetById(ctx context.Context, id int64) (domain.WithdrawalDetail, error) {
	db, err := r.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return domain.WithdrawalDetail{}, errors.Wrap(err, "err get db")
	}

	var withdrawal domain.WithdrawalDetail
	has, err := orm.Table(db, domain.Withdrawal{}).
		InnerJoin(domain.User{}, fmt.Sprintf("%s.id = %s.investor_id", domain.User{}.TableName(), domain.Withdrawal{}.TableName())).
		Equal(domain.Withdrawal{}.TableName()+".id", id).
		Get(&withdrawal)
	if err != nil {
		return domain.WithdrawalDetail{}, errors.Wrap(err, "err get withdrawal")
	}
	if !has {
		return domain.WithdrawalDetail{}, cdomain.ErrItemNotFound
	}

	return withdrawal, nil
}

func (r withdrawalMysqlRepository) GetByTransactionId(ctx context.Context, transactionId string) (domain.Withdrawal, error) {
	db, err := r.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return domain.Withdrawal{}, errors.Wrap(err, "err get db")
	}

	var withdrawal domain.Withdrawal
	has, err := orm.Table(db, domain.Withdrawal{}).
		Select(domain.Withdrawal{}.SelectAll()).
		InnerJoin(domain.WithdrawalTransferMapping{}, fmt.Sprintf("%s.id = %s.withdrawal_id", domain.Withdrawal{}.TableName(), domain.WithdrawalTransferMapping{}.TableName())).
		Equal(domain.WithdrawalTransferMapping{}.TableName()+".transaction_id", transactionId).
		Get(&withdrawal)
	if err != nil {
		return domain.Withdrawal{}, errors.Wrap(err, "err get withdrawal")
	}
	if !has {
		return domain.Withdrawal{}, cdomain.ErrItemNotFound
	}

	return withdrawal, nil
}

func (r withdrawalMysqlRepository) Update(ctx context.Context, withdrawal domain.Withdrawal) error {
	db, err := r.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "err get db")
	}

	if _, err = db.ID(withdrawal.Id).Update(&withdrawal); err != nil {
		return errors.Wrap(err, "err update withdrawal")
	}

	return nil
}

func (r withdrawalMysqlRepository) GetSum(ctx context.Context, options ...cdomain.WithdrawalQueryOption) (int64, error) {
	db, err := r.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "err get db")
	}

	queryOpts := cdomain.WithdrawalQueryOptionFunc()
	for _, opt := range options {
		opt(&queryOpts)
	}

	if queryOpts.SumColumn() == nil {
		return 0, errors.New("sum column is required")
	}

	query := orm.Table(db, domain.Withdrawal{}).
		Equal(domain.Withdrawal{}.TableName()+".investor_id", queryOpts.InvestorId()).
		Equal(domain.Withdrawal{}.TableName()+".status", queryOpts.Status())

	if len(queryOpts.Ids()) > 0 {
		query.In(domain.Withdrawal{}.TableName()+".id", queryOpts.Ids())
	}

	sum, err := query.SumInt(domain.Withdrawal{}.TableName() + "." + lo.FromPtr(queryOpts.SumColumn()))
	if err != nil {
		return 0, errors.Wrap(err, "err get sum")
	}

	return sum, nil
}
