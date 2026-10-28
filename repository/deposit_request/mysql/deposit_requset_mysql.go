package mysql

import (
	"context"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/orm"
	"github.com/pkg/errors"
)

type depositRequestMysqlRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.DepositRequestRepository {
	return &depositRequestMysqlRepository{
		dbFactory: dbFactory,
	}
}

func (d depositRequestMysqlRepository) GetById(ctx context.Context, id int64) (domain.DepositRequest, error) {
	db, err := d.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return domain.DepositRequest{}, errors.Wrap(err, "err get db")
	}

	depositRequest := domain.DepositRequest{}

	has, err := orm.Table(db, domain.DepositRequest{}).
		Select(domain.DepositRequest{}).
		Id(id).
		Get(&depositRequest)
	if err != nil {
		return domain.DepositRequest{}, errors.Wrap(err, "err get deposit request")
	}

	if !has {
		return domain.DepositRequest{}, errors.New("err deposit request not found")
	}

	return depositRequest, nil
}

func (d depositRequestMysqlRepository) Update(ctx context.Context, depositRequest domain.DepositRequest) error {
	db, err := d.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "err get db")
	}

	if _, err = db.ID(depositRequest.Id).UseBool().Update(depositRequest); err != nil {
		return errors.Wrap(err, "err update deposit request")
	}

	return nil
}

func (d depositRequestMysqlRepository) GetDepositRequestByShareId(ctx context.Context, shareId string) (domain.DepositRequest, error) {
	db, err := d.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return domain.DepositRequest{}, errors.Wrap(err, "err get db")
	}

	depositRequest := domain.DepositRequest{}

	has, err := orm.Table(db, domain.DepositRequest{}).
		Select(domain.DepositRequest{}).
		Equal(domain.DepositRequest{}.TableName()+".share_id", &shareId).
		Get(&depositRequest)
	if err != nil {
		return domain.DepositRequest{}, errors.Wrap(err, "err get deposit request")
	}

	if !has {
		return domain.DepositRequest{}, factory.ErrorNotFound.New(nil)
	}

	return depositRequest, nil
}

func (d depositRequestMysqlRepository) GetAll(ctx context.Context, options ...cdomain.DepositRequestQueryOption) ([]domain.DepositRequest, error) {
	db, err := d.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "err get db")
	}

	queryOpts := cdomain.DepositRequestQueryOptionFunc()
	for _, opt := range options {
		opt(&queryOpts)
	}

	depositRequests := make([]domain.DepositRequest, 0)

	if err = orm.Table(db, domain.DepositRequest{}).
		Equal(domain.DepositRequest{}.TableName()+".user_id", queryOpts.UserId()).
		GT(domain.DepositRequest{}.TableName()+".expire_deposit_date", queryOpts.ExpireDepositDate()).
		SortOrder(queryOpts.SortBy(), queryOpts.Order()).
		Find(&depositRequests); err != nil {
		return nil, errors.Wrap(err, "err find and count")
	}

	return depositRequests, nil
}

func (d depositRequestMysqlRepository) GetNoDepositRequestByUserId(ctx context.Context, userId int64) ([]domain.DepositRequest, error) {
	db, err := d.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "err get db")
	}

	depositRequests := make([]domain.DepositRequest, 0)

	if err = orm.Table(db, domain.DepositRequest{}).
		Equal(domain.DepositRequest{}.TableName()+".user_id", &userId).
		Equal(domain.DepositRequest{}.TableName()+".deposit_check_status", domain.NO_DEPOSIT).
		Asc(domain.DepositRequest{}.TableName() + ".created").
		Find(&depositRequests); err != nil {
		return nil, errors.Wrap(err, "err find deposit requests by userId")
	}

	return depositRequests, nil
}
