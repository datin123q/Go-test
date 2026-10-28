package mysql

import (
	"context"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/pkg/errors"
)

type eventMessageMysqlRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.EventMessageRepository {
	return &eventMessageMysqlRepository{
		dbFactory: dbFactory,
	}
}

func (e eventMessageMysqlRepository) Create(ctx context.Context, eventMessage domain.EventMessage) error {
	db, err := e.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "err get db")
	}

	if _, err = db.Insert(eventMessage); err != nil {
		return errors.Wrap(err, "err insert eventMessage")
	}

	return nil
}

func (e eventMessageMysqlRepository) Update(ctx context.Context, message domain.EventMessage) error {
	db, err := e.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "err get db")
	}

	if _, err = db.ID(message.Uuid).UseBool().Update(message); err != nil {
		return errors.Wrap(err, "err update eventMessage")
	}

	return nil
}
