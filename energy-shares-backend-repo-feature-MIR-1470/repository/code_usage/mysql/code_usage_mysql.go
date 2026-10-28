package mysql

import (
	"context"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/pkg/errors"
)

type codeUsageMysqlRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.CodeUsageRepository {
	return &codeUsageMysqlRepository{
		dbFactory: dbFactory,
	}
}

func (r codeUsageMysqlRepository) Create(ctx context.Context, codeUsage domain.CodeUsage) error {
	db, err := r.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "err get db")
	}

	if _, err = db.InsertOne(&codeUsage); err != nil {
		return errors.Wrap(err, "err insert code usage")
	}

	return nil
}
