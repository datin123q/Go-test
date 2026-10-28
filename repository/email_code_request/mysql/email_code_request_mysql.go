package mysql

import (
	"context"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/orm"
	"github.com/pkg/errors"
)

type emailCodeRequestRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.EmailCodeRequestRepository {
	return &emailCodeRequestRepository{
		dbFactory: dbFactory,
	}
}

func (e emailCodeRequestRepository) Create(ctx context.Context, emailCodeRequest domain.EmailCodeRequest) error {
	db, err := e.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get db")
	}

	if _, err = db.Insert(emailCodeRequest); err != nil {
		return errors.Wrap(err, "failed to insert emailCodeRequest")
	}

	return nil
}

func (e emailCodeRequestRepository) Update(ctx context.Context, emailCodeRequest domain.EmailCodeRequest) error {
	db, err := e.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get db")
	}

	if _, err = db.Update(emailCodeRequest); err != nil {
		return errors.Wrap(err, "failed to update emailCodeRequest")
	}

	return nil
}

func (e emailCodeRequestRepository) GetAll(ctx context.Context, options ...cdomain.EmailCodeRequestQueryOption) ([]domain.EmailCodeRequest, error) {
	db, err := e.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "err get db")
	}

	queryOpts := cdomain.EmailCodeRequestQueryOptionFunc()
	for _, opt := range options {
		opt(&queryOpts)
	}

	emailCodeRequests := make([]domain.EmailCodeRequest, 0)

	err = orm.Table(db, domain.EmailCodeRequest{}).
		Equal(domain.EmailCodeRequest{}.TableName()+".user_id", queryOpts.UserId()).
		Equal(domain.EmailCodeRequest{}.TableName()+".first_req", queryOpts.FirstReq()).
		Equal(domain.EmailCodeRequest{}.TableName()+".code", queryOpts.Code()).
		Equal(domain.EmailCodeRequest{}.TableName()+".email", queryOpts.Email()).
		GTE(domain.EmailCodeRequest{}.TableName()+".created", queryOpts.GTECreated()).
		SortOrder(queryOpts.SortBy(), queryOpts.Order()).
		Limit(queryOpts.Limit(), queryOpts.Offset()).
		Find(&emailCodeRequests)
	if err != nil {
		return nil, errors.Wrap(err, "err find and count")
	}

	return emailCodeRequests, nil
}
