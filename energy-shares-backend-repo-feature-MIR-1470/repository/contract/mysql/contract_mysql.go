package mysql

import (
	"context"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/orm"
	"github.com/pkg/errors"
)

type contractMysqlRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.ContractRepository {
	return &contractMysqlRepository{
		dbFactory: dbFactory,
	}
}

func (c contractMysqlRepository) Create(ctx context.Context, contract domain.Contract) (int64, error) {
	db, err := c.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "err get db")
	}

	if _, err = db.InsertOne(&contract); err != nil {
		return 0, errors.Wrap(err, "err insert contract")
	}

	return contract.Id, nil
}

func (c contractMysqlRepository) GetAll(ctx context.Context, options ...cdomain.ContractQueryOption) (domain.Contracts, error) {
	q, err := c.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "err get db")
	}

	opts := cdomain.ContractQueryOptionFunc()
	for _, opt := range options {
		opt(&opts)
	}

	contracts := make([]domain.Contract, 0)

	err = orm.Table(q, domain.Contract{}).
		Select(domain.Contract{}).
		SortOrder(opts.SortBy(), opts.Order()).
		Equal(domain.Contract{}.TableName()+".share_id", opts.ShareId()).
		Limit(opts.Limit(), opts.Offset()).
		Find(&contracts)
	if err != nil {
		return nil, errors.Wrap(err, "err get contracts")
	}

	return contracts, nil
}
