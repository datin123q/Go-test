package mysql

import (
	"context"

	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

type lotteryHistoryRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.LotteryHistoryRepository {
	return &lotteryHistoryRepository{dbFactory: dbFactory}
}

func (r lotteryHistoryRepository) Create(ctx context.Context, history domain.LotteryHistory) error {
	db, err := r.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "err get db")
	}

	if _, err = db.Insert(&history); err != nil {
		return errors.Wrap(err, "err insert lottery history")
	}

	return nil
}
