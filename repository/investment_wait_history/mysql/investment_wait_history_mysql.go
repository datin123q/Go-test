package mysql

import (
	"context"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/orm"
	"github.com/pkg/errors"
)

type investmentWaitHistoryMysqlRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.InvestmentWaitHistoryRepository {
	return &investmentWaitHistoryMysqlRepository{
		dbFactory: dbFactory,
	}
}

func (i investmentWaitHistoryMysqlRepository) Create(ctx context.Context, investmentWaitHistory domain.InvestmentWaitHistory) error {
	db, err := i.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "err get db")
	}

	if _, err = db.Insert(investmentWaitHistory); err != nil {
		return errors.Wrap(err, "err insert contract")
	}

	return nil
}

func (i investmentWaitHistoryMysqlRepository) GetAll(ctx context.Context, options ...cdomain.InvestmentWaitHistoryQueryOption) (int64, []domain.InvestmentWaitHistory, error) {
	db, err := i.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, nil, errors.Wrap(err, "err get db")
	}

	opts := cdomain.InvestmentWaitHistoryQueryOptionFunc()
	for _, opt := range options {
		opt(&opts)
	}

	investmentWaitHistories := make([]domain.InvestmentWaitHistory, 0)

	totalCount, err := orm.Table(db, domain.InvestmentWaitHistory{}).
		Select(domain.InvestmentWaitHistory{}).
		Equal(domain.InvestmentWaitHistory{}.TableName()+".investment_wait_id", opts.InvestmentWaitId()).
		SortOrder(opts.SortBy(), opts.Order()).
		Limit(opts.Limit(), opts.Offset()).
		FindAndCount(&investmentWaitHistories)

	return totalCount, investmentWaitHistories, err
}
