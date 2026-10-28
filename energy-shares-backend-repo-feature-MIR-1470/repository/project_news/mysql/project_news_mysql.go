package mysql

import (
	"context"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/orm"
	"github.com/pkg/errors"
)

type projectNewsMysqlRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.ProjectNewsRepository {
	return &projectNewsMysqlRepository{
		dbFactory: dbFactory,
	}
}

func (p projectNewsMysqlRepository) GetAll(ctx context.Context, options ...cdomain.ProjectQueryOption) ([]domain.ProjectNews, error) {
	db, err := p.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get db")
	}

	opts := cdomain.ProjectQueryOptionFunc()
	for _, opt := range options {
		opt(&opts)
	}

	projectNews := make([]domain.ProjectNews, 0)

	if err = orm.Table(db, domain.ProjectNews{}).
		Select(domain.ProjectNews{}).
		Equal(domain.ProjectNews{}.TableName()+".project_id", opts.ProjectId()).
		SortOrder(opts.SortBy(), opts.Order()).
		Limit(opts.Limit(), opts.Offset()).
		Find(&projectNews); err != nil {
		return nil, errors.Wrap(err, "failed to get project news")
	}

	return projectNews, nil
}
