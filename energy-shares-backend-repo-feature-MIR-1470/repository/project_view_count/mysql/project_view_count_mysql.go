package mysql

import (
	"context"
	"fmt"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/orm"
	"github.com/samber/lo"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/pkg/errors"
)

type projectViewCountMysqlRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.ProjectViewCountRepository {
	return &projectViewCountMysqlRepository{
		dbFactory: dbFactory,
	}
}

func (p projectViewCountMysqlRepository) Create(ctx context.Context, projectViewCount domain.ProjectViewCount) error {
	db, err := p.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get db")
	}

	if _, err = db.Insert(&projectViewCount); err != nil {
		return errors.Wrap(err, "failed to insert project view count")
	}

	return nil
}

func (p projectViewCountMysqlRepository) GetByProjectId(ctx context.Context, projectId int64) (domain.ProjectViewCount, error) {
	db, err := p.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return domain.ProjectViewCount{}, errors.Wrap(err, "failed to get db")
	}

	projectViewCount := domain.ProjectViewCount{}

	has, err := orm.Table(db, domain.ProjectViewCount{}).
		Equal(domain.ProjectViewCount{}.TableName()+".project_id", lo.ToPtr(fmt.Sprintf("%v", projectId))).
		Get(&projectViewCount)
	if err != nil {
		return domain.ProjectViewCount{}, errors.Wrap(err, "failed to get project view count")
	}

	if !has {
		return domain.ProjectViewCount{}, errors.Wrap(cdomain.ErrItemNotFound, "failed to get project view count")
	}

	return projectViewCount, nil
}
