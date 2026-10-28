package mysql

import (
	"context"
	"fmt"

	"github.com/samber/lo"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/orm"
	"github.com/pkg/errors"
)

var (
	itemNotFound = errors.New("item not found")
)

type investmentRequestMysqlRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.InvestmentRequestRepository {
	return &investmentRequestMysqlRepository{
		dbFactory: dbFactory,
	}
}

func (i investmentRequestMysqlRepository) Create(ctx context.Context, investmentRequest domain.InvestmentRequest) (int64, error) {
	db, err := i.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "err get db")
	}

	if _, err = db.InsertOne(&investmentRequest); err != nil {
		return 0, errors.Wrap(err, "err insert investment request")
	}

	return investmentRequest.Id, nil
}

func (i investmentRequestMysqlRepository) Update(ctx context.Context, investmentRequest domain.InvestmentRequest) error {
	db, err := i.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "err get db")
	}

	_, err = db.ID(investmentRequest.Id).UseBool().Update(investmentRequest)
	if err != nil {
		return errors.Wrap(err, "err update investment request")
	}

	return nil
}

func (i investmentRequestMysqlRepository) GetByInvestmentWaitersId(ctx context.Context, investmentWaitersId int64) (domain.InvestmentRequest, error) {
	db, err := i.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return domain.InvestmentRequest{}, errors.Wrap(err, "err get db")
	}

	investmentRequest := domain.InvestmentRequest{}

	has, err := orm.Table(db, domain.InvestmentRequest{}).
		Equal(domain.InvestmentRequest{}.TableName()+".investment_waiters_id", lo.ToPtr(fmt.Sprintf("%v", investmentWaitersId))).
		Get(&investmentRequest)
	if err != nil {
		return domain.InvestmentRequest{}, errors.Wrap(err, "err get investment request")
	}

	if !has {
		return domain.InvestmentRequest{}, errors.Wrap(cdomain.ErrItemNotFound, "err investment request not found")
	}

	return investmentRequest, nil
}
