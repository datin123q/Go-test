package mysql

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/orm"
	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

type projectMilestoneMysqlRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.ProjectMilestoneRepository {
	return &projectMilestoneMysqlRepository{
		dbFactory: dbFactory,
	}
}

func (p projectMilestoneMysqlRepository) GetAll(ctx context.Context, options ...cdomain.ProjectQueryOption) ([]domain.ProjectMilestone, error) {
	db, err := p.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get db")
	}

	opts := cdomain.ProjectQueryOptionFunc()
	for _, opt := range options {
		opt(&opts)
	}

	projectMilestones := make([]domain.ProjectMilestone, 0)

	if err = orm.Table(db, domain.ProjectMilestone{}).
		Select(domain.ProjectMilestone{}).
		Equal(domain.ProjectMilestone{}.TableName()+".project_id", opts.ProjectId()).
		SortOrder(opts.SortBy(), opts.Order()).
		Limit(opts.Limit(), opts.Offset()).
		Find(&projectMilestones); err != nil {
		return nil, errors.Wrap(err, "failed to get project milestone")
	}

	return projectMilestones, nil
}
