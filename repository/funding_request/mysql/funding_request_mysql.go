package mysql

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/orm"
	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

type fundingRequestRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.FundingRequestRepository {
	return &fundingRequestRepository{
		dbFactory: dbFactory,
	}
}

func (f fundingRequestRepository) Create(ctx context.Context, fundingRequest domain.FundingRequest) (int64, error) {
	db, err := f.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "failed to get db")
	}

	_, err = db.InsertOne(&fundingRequest)
	if err != nil {
		return 0, errors.Wrap(err, "failed to insert")
	}

	return fundingRequest.Id, nil
}

func (f fundingRequestRepository) Update(ctx context.Context, fundingRequest domain.FundingRequest) error {
	db, err := f.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get db")
	}

	_, err = db.ID(fundingRequest.Id).Update(&fundingRequest)
	if err != nil {
		return errors.Wrap(err, "failed to update")
	}

	return nil
}

func (f fundingRequestRepository) GetAll(ctx context.Context, options ...cdomain.FundingRequestQueryOption) (int64, []domain.FundingRequest, error) {
	db, err := f.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, nil, errors.Wrap(err, "failed to get db")
	}

	queryOpts := cdomain.FundingRequestQueryOptionFunc()
	for _, opt := range options {
		opt(&queryOpts)
	}

	var fundingRequests []domain.FundingRequest

	count, err := orm.Table(db, domain.FundingRequest{}).
		Equal(domain.FundingRequest{}.TableName()+".id", queryOpts.Id()).
		Equal(domain.FundingRequest{}.TableName()+"email", queryOpts.Email()).
		Equal(domain.FundingRequest{}.TableName()+"phone", queryOpts.Phone()).
		Equal(domain.FundingRequest{}.TableName()+"request_status", queryOpts.RequestStatus()).
		Between("created", queryOpts.StartDate(), queryOpts.EndDate()).
		SortOrder(queryOpts.SortBy(), queryOpts.Order()).
		Limit(queryOpts.Limit(), queryOpts.Offset()).
		FindAndCount(&fundingRequests)
	if err != nil {
		return 0, nil, errors.Wrap(err, "err find and count")
	}

	return count, fundingRequests, nil
}

func (f fundingRequestRepository) GetById(ctx context.Context, id int64) (domain.FundingRequest, error) {
	db, err := f.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return domain.FundingRequest{}, errors.Wrap(err, "failed to get db")
	}

	var fundingRequest domain.FundingRequest
	has, err := db.ID(id).Get(&fundingRequest)
	if err != nil {
		return domain.FundingRequest{}, errors.Wrap(err, "failed to get")
	}
	if !has {
		return domain.FundingRequest{}, cdomain.ErrItemNotFound
	}

	return fundingRequest, nil
}
