package mysql

import (
	"context"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/orm"
	"github.com/pkg/errors"
)

type projectImageMysqlRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.ProjectImageRepository {
	return &projectImageMysqlRepository{
		dbFactory: dbFactory,
	}
}

func (p projectImageMysqlRepository) GetById(ctx context.Context, id int64) (domain.ProjectImage, error) {
	db, err := p.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return domain.ProjectImage{}, errors.Wrap(err, "failed to get db")
	}

	var projectImage domain.ProjectImage

	has, err := orm.Table(db, domain.ProjectImage{}).Id(id).Get(&projectImage)
	if err != nil {
		return domain.ProjectImage{}, errors.Wrap(err, "failed to get project image by id")
	}

	if !has {
		return domain.ProjectImage{}, errors.Wrap(cdomain.ErrItemNotFound, "project image not found")
	}

	return projectImage, nil
}
