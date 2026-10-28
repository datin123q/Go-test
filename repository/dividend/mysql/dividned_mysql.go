package mysql

import (
	"context"

	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/orm"
)

type dividendRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.DividendRepository {
	return &dividendRepository{
		dbFactory: dbFactory,
	}
}

func (d dividendRepository) CreateMany(ctx context.Context, dividends []domain.Dividend) error {
	db, err := d.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get db")
	}

	if _, err = db.Insert(&dividends); err != nil {
		return errors.Wrap(err, "failed to insert dividends")
	}

	return nil

}

func (d dividendRepository) DeleteByScheduleId(ctx context.Context, options ...cdomain.DividendQueryOption) error {
	db, err := d.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get db")
	}

	queryOpts := cdomain.DividendQueryOptionFunc()
	for _, opt := range options {
		opt(&queryOpts)
	}

	if _, err := db.Where("settlement_id = ?", *queryOpts.ScheduleId()).
		Delete(&domain.Dividend{}); err != nil {
		return err
	}

	return nil
}

func (d dividendRepository) GetAll(ctx context.Context, options ...cdomain.DividendQueryOption) ([]domain.Dividend, error) {
	db, err := d.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "err get db")
	}

	queryOpts := cdomain.DividendQueryOptionFunc()
	for _, opt := range options {
		opt(&queryOpts)
	}

	dividends := make([]domain.Dividend, 0)

	err = orm.Table(db, domain.Share{}).
		Equal(domain.Share{}.TableName()+".user_id", queryOpts.UserId()).
		Equal(domain.Share{}.TableName()+".spv_id", queryOpts.SpvId()).
		Equal(domain.Share{}.TableName()+".settlement_id", queryOpts.ScheduleId()).
		Between("dividend_date", queryOpts.StartDate(), queryOpts.EndDate()).
		SortOrder(queryOpts.SortBy(), queryOpts.Order()).
		Limit(queryOpts.Limit(), queryOpts.Offset()).
		Find(&dividends)
	if err != nil {
		return nil, errors.Wrap(err, "err find and count")
	}

	return dividends, nil
}
