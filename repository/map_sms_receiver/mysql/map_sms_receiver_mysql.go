package mysql

import (
	"context"

	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

type mapSmsReceiverMysqlRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.MapSmsReceiverRepository {
	return &mapSmsReceiverMysqlRepository{
		dbFactory: dbFactory,
	}
}

func (m mapSmsReceiverMysqlRepository) Create(ctx context.Context, mapMailReceiver domain.MapSmsReceiver) (int64, error) {
	db, err := m.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "err get db")
	}

	if _, err = db.InsertOne(&mapMailReceiver); err != nil {
		return 0, errors.Wrap(err, "err insert map_sms_receiver")
	}

	return mapMailReceiver.Id, nil
}
