package mysql

import (
	"context"

	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

type smsMsgMysqlRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.SmsMsgRepository {
	return &smsMsgMysqlRepository{
		dbFactory: dbFactory,
	}
}

func (s *smsMsgMysqlRepository) Create(ctx context.Context, smsMsg domain.SmsMsg) (int64, error) {
	db, err := s.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "err get db")
	}

	if _, err = db.InsertOne(&smsMsg); err != nil {
		return 0, errors.Wrap(err, "failed to insert sms_msg")
	}

	return smsMsg.Id, nil
}

func (s *smsMsgMysqlRepository) Update(ctx context.Context, smsMsg domain.SmsMsg) error {
	db, err := s.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "err get db")
	}

	_, err = db.ID(smsMsg.Id).Update(smsMsg)
	if err != nil {
		return errors.Wrap(err, "failed to update sms_msg")
	}

	return nil
}
