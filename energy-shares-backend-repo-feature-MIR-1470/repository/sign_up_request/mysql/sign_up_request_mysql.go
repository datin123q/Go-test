package mysql

import (
	"context"

	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/orm"
)

type signUpRequestMysqlRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.SignUpRequestRepository {
	return &signUpRequestMysqlRepository{
		dbFactory: dbFactory,
	}
}

func (s signUpRequestMysqlRepository) Create(ctx context.Context, signUpRequest domain.SignUpRequest) (int64, error) {
	db, err := s.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "failed to get db")
	}

	_, err = db.Insert(&signUpRequest)
	if err != nil {
		return 0, errors.Wrap(err, "failed to create user")
	}

	return signUpRequest.Id, nil
}

func (s signUpRequestMysqlRepository) Update(ctx context.Context, signUpRequest domain.SignUpRequest) error {
	db, err := s.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get db")
	}

	_, err = db.ID(signUpRequest.Id).Update(&signUpRequest)
	if err != nil {
		return errors.Wrap(err, "failed to update user")
	}

	return nil
}

func (s signUpRequestMysqlRepository) GetByRequestId(ctx context.Context, requestId string) (domain.SignUpRequest, error) {
	db, err := s.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return domain.SignUpRequest{}, errors.Wrap(err, "failed to get db")
	}

	var signUpRequest domain.SignUpRequest

	has, err := orm.Table(db, domain.SignUpRequest{}).
		Equal(domain.SignUpRequest{}.TableName()+".request_id", requestId).
		Get(&signUpRequest)
	if err != nil {
		return domain.SignUpRequest{}, errors.Wrap(err, "failed to get sign up request")
	}

	if !has {
		return domain.SignUpRequest{}, errors.Wrap(cdomain.ErrItemNotFound, "sign up request not found")
	}

	return signUpRequest, nil
}

func (s signUpRequestMysqlRepository) GetAll(ctx context.Context, options ...cdomain.SignUpRequestQueryOption) (cdomain.SignUpRequests, error) {
	db, err := s.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get db")
	}

	queryOpts := cdomain.SignUpRequestQueryOptionFunc()
	for _, opt := range options {
		opt(&queryOpts)
	}

	signUpRequests := make(cdomain.SignUpRequests, 0)

	if err = orm.Table(db, domain.SignUpRequest{}).
		Equal(domain.SignUpRequest{}.TableName()+".email", queryOpts.Email()).
		GT(domain.SignUpRequest{}.TableName()+".url_expire", queryOpts.UrlExpire()).
		GT(domain.SignUpRequest{}.TableName()+".email_expire", queryOpts.EmailExpire()).
		SortOrder(queryOpts.SortBy(), queryOpts.Order()).
		Find(&signUpRequests); err != nil {
		return nil, errors.Wrap(err, "err find and count")
	}

	return signUpRequests, nil
}

func (s signUpRequestMysqlRepository) Delete(ctx context.Context, id int64) error {
	db, err := s.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get db")
	}

	_, err = db.ID(id).Delete(&domain.SignUpRequest{})
	if err != nil {
		return errors.Wrap(err, "failed to delete sign up request")
	}

	return nil
}
