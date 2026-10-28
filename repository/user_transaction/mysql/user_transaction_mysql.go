package mysql

import (
	"context"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/pkg/errors"
)

type userTransactionMysqlRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.UserTransactionRepository {
	return &userTransactionMysqlRepository{
		dbFactory: dbFactory,
	}
}

func (m userTransactionMysqlRepository) Create(ctx context.Context, userTransaction domain.UserTransaction) (int64, error) {
	db, err := m.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "err get db")
	}

	if _, err = db.Insert(&userTransaction); err != nil {
		return 0, errors.Wrap(err, "err insert userTransaction")
	}

	return userTransaction.Id, nil
}
