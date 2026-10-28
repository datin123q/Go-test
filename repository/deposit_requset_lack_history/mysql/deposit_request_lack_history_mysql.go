package mysql

import (
	"context"

	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

type depositRequestLackHistoryMysqlRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.DepositRequestLackHistoryRepository {
	return &depositRequestLackHistoryMysqlRepository{
		dbFactory: dbFactory,
	}
}

func (d depositRequestLackHistoryMysqlRepository) Create(ctx context.Context, depositRequestLackHistory domain.DepositRequestLackHistory) error {
	db, err := d.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "err get db")
	}

	if _, err = db.Insert(depositRequestLackHistory); err != nil {
		return errors.Wrap(err, "err insert depositRequestLackHistory")
	}

	return nil
}
