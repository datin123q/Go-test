package mysql

import (
	"context"

	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/orm"
)

type passwordResetRequestMysqlRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.PasswordResetRequestRepository {
	return &passwordResetRequestMysqlRepository{
		dbFactory: dbFactory,
	}
}

func (p passwordResetRequestMysqlRepository) Create(ctx context.Context, passwordResetRequest domain.PasswordResetRequest) (int64, error) {
	db, err := p.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "err get db")
	}

	if _, err = db.InsertOne(&passwordResetRequest); err != nil {
		return 0, errors.Wrap(err, "failed to insert password_reset_request")
	}

	return passwordResetRequest.Id, nil
}

func (p passwordResetRequestMysqlRepository) Update(ctx context.Context, passwordResetRequest domain.PasswordResetRequest) error {
	db, err := p.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "err get db")
	}

	_, err = db.ID(passwordResetRequest.Id).Update(passwordResetRequest)
	if err != nil {
		return errors.Wrap(err, "failed to update password_reset_request")
	}

	return nil
}

func (p passwordResetRequestMysqlRepository) GetLatestByEmail(ctx context.Context, email string) (domain.PasswordResetRequest, error) {
	db, err := p.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return domain.PasswordResetRequest{}, errors.Wrap(err, "failed to get db")
	}

	var passwordResetReq domain.PasswordResetRequest

	has, err := orm.Table(db, domain.PasswordResetRequest{}).
		Equal(domain.PasswordResetRequest{}.TableName()+".email", email).
		Desc(domain.PasswordResetRequest{}.TableName() + ".id").
		Get(&passwordResetReq)
	if err != nil {
		return domain.PasswordResetRequest{}, errors.Wrap(err, "failed to get unexpired password_reset_request by email")
	}
	if !has {
		return domain.PasswordResetRequest{}, errors.Wrap(cdomain.ErrItemNotFound, "password_reset_request not found")
	}

	return passwordResetReq, nil
}

func (p passwordResetRequestMysqlRepository) GetByRequestId(ctx context.Context, requestId string) (domain.PasswordResetRequest, error) {
	db, err := p.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return domain.PasswordResetRequest{}, errors.Wrap(err, "failed to get db")
	}

	var passwordResetReq domain.PasswordResetRequest

	has, err := orm.Table(db, domain.PasswordResetRequest{}).
		Equal(domain.PasswordResetRequest{}.TableName()+".request_id", requestId).
		Get(&passwordResetReq)
	if err != nil {
		return domain.PasswordResetRequest{}, errors.Wrap(err, "failed to get password_reset_request by request_id")
	}
	if !has {
		return domain.PasswordResetRequest{}, errors.Wrap(cdomain.ErrItemNotFound, "password_reset_request not found")
	}

	return passwordResetReq, nil
}
