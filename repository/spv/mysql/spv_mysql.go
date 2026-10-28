package mysql

import (
	"context"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/orm"
	"github.com/pkg/errors"
)

type spvMysqlRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.SpvRepository {
	return &spvMysqlRepository{
		dbFactory: dbFactory,
	}
}

func (s spvMysqlRepository) GetById(ctx context.Context, id int64) (domain.Spv, error) {
	db, err := s.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return domain.Spv{}, errors.Wrap(err, "failed to get db")
	}

	var spv domain.Spv

	has, err := orm.Table(db, domain.Spv{}).Id(id).Get(&spv)
	if err != nil {
		return domain.Spv{}, errors.Wrap(err, "failed to get spv by id")
	}

	if !has {
		return domain.Spv{}, errors.Wrap(cdomain.ErrItemNotFound, "spv not found")
	}

	return spv, nil
}
