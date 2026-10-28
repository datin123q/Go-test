package mysql

import (
	"context"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/orm"
	"github.com/pkg/errors"
)

type loginHistoryMysqlRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.LoginHistoryRepository {
	return &loginHistoryMysqlRepository{
		dbFactory: dbFactory,
	}
}

func (l loginHistoryMysqlRepository) GetAll(ctx context.Context, options ...cdomain.LoginHistoryQueryOption) (int64, []domain.LoginHistory, error) {
	db, err := l.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, nil, errors.Wrap(err, "failed to get db")
	}

	opts := cdomain.LoginHistoryQueryOptionFunc()
	for _, opt := range options {
		opt(&opts)
	}

	loginHistories := make([]domain.LoginHistory, 0)

	totalCount, err := orm.Table(db, domain.LoginHistory{}).
		Select(domain.LoginHistory{}.SelectAll()).
		Like(domain.LoginHistory{}.TableName()+".name", opts.Name()).
		Between("created", opts.StartDate(), opts.EndDate()).
		SortOrder(opts.SortBy(), opts.Order()).
		Limit(opts.Limit(), opts.Offset()).
		FindAndCount(&loginHistories)
	if err != nil {
		return 0, nil, errors.Wrap(err, "failed to get login histories")
	}

	return totalCount, loginHistories, nil
}

func (l loginHistoryMysqlRepository) Create(ctx context.Context, loginHistory domain.LoginHistory) (int64, error) {
	db, err := l.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "failed to get db")
	}

	if _, err = db.Insert(&loginHistory); err != nil {
		return 0, errors.Wrap(err, "failed to insert login history")
	}

	return loginHistory.Id, nil
}
