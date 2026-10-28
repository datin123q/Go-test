package mysql

import (
	"context"

	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/orm"
)

type antisocialCheckFileRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.AntisocialCheckFileRepository {
	return &antisocialCheckFileRepository{
		dbFactory: dbFactory,
	}
}

func (a antisocialCheckFileRepository) GetAll(ctx context.Context, options ...cdomain.AntisocialCheckFileQueryOption) (
	int64, domain.AntisocialCheckFiles, error) {
	db, err := a.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, nil, errors.Wrap(err, "err get db")
	}

	queryOpts := cdomain.AntisocialCheckFileQueryOptionFunc()
	for _, opt := range options {
		opt(&queryOpts)
	}

	antisocialCheckFiles := make([]domain.AntisocialCheckFile, 0)

	total, err := orm.Table(db, domain.AntisocialCheckFile{}).
		Equal(domain.AntisocialCheckFile{}.TableName()+".check_id", queryOpts.CheckId()).
		SortOrder(queryOpts.SortBy(), queryOpts.Order()).
		Limit(queryOpts.Limit(), queryOpts.Offset()).
		FindAndCount(&antisocialCheckFiles)
	if err != nil {
		return 0, nil, errors.Wrap(err, "err find and count")
	}

	return total, antisocialCheckFiles, nil
}
