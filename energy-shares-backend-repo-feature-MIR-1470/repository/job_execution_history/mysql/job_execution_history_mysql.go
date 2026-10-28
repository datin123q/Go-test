package mysql

import (
	"context"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/pkg/errors"
)

type jobExecutionHistoryMysqlRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.JobExecutionHistoryRepository {
	return &jobExecutionHistoryMysqlRepository{
		dbFactory: dbFactory,
	}
}

func (repo jobExecutionHistoryMysqlRepository) Create(ctx context.Context, history domain.JobExecutionHistory) (int64, error) {
	db, err := repo.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "failed to get db")
	}

	if _, err = db.Insert(&history); err != nil {
		return 0, errors.Wrap(err, "failed to insert job execution history")
	}

	return history.Id, nil
}

func (repo jobExecutionHistoryMysqlRepository) GetById(ctx context.Context, id int64) (domain.JobExecutionHistory, error) {
	db, err := repo.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return domain.JobExecutionHistory{}, errors.Wrap(err, "failed to get db")
	}

	var history domain.JobExecutionHistory
	has, err := db.ID(id).Get(&history)
	if err != nil {
		return domain.JobExecutionHistory{}, errors.Wrap(err, "failed to get job execution history")
	}
	if !has {
		return domain.JobExecutionHistory{}, cdomain.ErrItemNotFound
	}

	return history, nil
}

func (repo jobExecutionHistoryMysqlRepository) Update(ctx context.Context, history domain.JobExecutionHistory) error {
	db, err := repo.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get db")
	}

	if _, err = db.ID(history.Id).UseBool().AllCols().Update(&history); err != nil {
		return errors.Wrap(err, "failed to update job execution history")
	}

	return nil
}
