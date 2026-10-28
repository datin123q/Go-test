package mysql

import (
	"context"

	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

type mailRetryMysqlRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.MailRetryRepository {
	return &mailRetryMysqlRepository{
		dbFactory: dbFactory,
	}
}

func (m *mailRetryMysqlRepository) Create(ctx context.Context, mailRetry domain.MailRetry) (int64, error) {
	db, err := m.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "err get db")
	}

	if _, err = db.InsertOne(&mailRetry); err != nil {
		return 0, errors.Wrap(err, "err insert mail retry")

	}

	return mailRetry.Id, nil
}
