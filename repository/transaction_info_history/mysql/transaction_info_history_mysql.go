package mysql

import (
	"context"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/orm"
	"github.com/pkg/errors"
)

type transactionInfoHistoryMysqlRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.TransactionInfoHistoryRepository {
	return &transactionInfoHistoryMysqlRepository{
		dbFactory: dbFactory,
	}
}

func (m transactionInfoHistoryMysqlRepository) Create(ctx context.Context, transactionInfoHistory domain.TransactionInfoHistory) error {
	db, err := m.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "err get db")
	}

	if _, err = db.Insert(transactionInfoHistory); err != nil {
		return errors.Wrap(err, "err insert transactionInfoHistory")
	}

	return nil
}

func (m transactionInfoHistoryMysqlRepository) GetAll(ctx context.Context, options ...cdomain.TransactionInfoHistoryQueryOption) (int64, []domain.TransactionInfoHistory, error) {
	db, err := m.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, nil, errors.Wrap(err, "err get db")
	}

	queryOpts := cdomain.TransactionInfoHistoryQueryOptionFunc()
	for _, opt := range options {
		opt(&queryOpts)
	}

	transactionInfoHistories := make([]domain.TransactionInfoHistory, 0)

	totalCount, err := orm.Table(db, domain.TransactionInfoHistory{}).
		Equal(domain.TransactionInfoHistory{}.TableName()+".share_id", queryOpts.ShareId()).
		SortOrder(queryOpts.SortBy(), queryOpts.Order()).
		FindAndCount(&transactionInfoHistories)

	if err != nil {
		return 0, nil, errors.Wrap(err, "err find and count")
	}

	return totalCount, transactionInfoHistories, nil
}
