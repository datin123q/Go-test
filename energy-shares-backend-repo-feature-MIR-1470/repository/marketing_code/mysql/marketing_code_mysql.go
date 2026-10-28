package mysql

import (
	"context"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/pkg/errors"
)

type marketingCodeMysqlRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.MarketingCodeRepository {
	return &marketingCodeMysqlRepository{
		dbFactory: dbFactory,
	}
}

func (m marketingCodeMysqlRepository) Create(ctx context.Context, marketingCode domain.MarketingCode) (int64, error) {
	db, err := m.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "err get db")
	}

	if _, err = db.Insert(&marketingCode); err != nil {
		return 0, errors.Wrap(err, "err insert marketingCode")
	}

	return marketingCode.Id, nil
}
