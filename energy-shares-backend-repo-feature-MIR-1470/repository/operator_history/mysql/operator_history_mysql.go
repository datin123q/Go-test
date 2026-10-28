package mysql

import (
	"context"

	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/orm"
)

type operatorHistoryMysqlRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.OperatorHistoryRepository {
	return &operatorHistoryMysqlRepository{
		dbFactory: dbFactory,
	}
}

func (repo operatorHistoryMysqlRepository) GetAll(ctx context.Context, options ...cdomain.OperatorHistoryQueryOption) (int64, []domain.OperatorHistory, error) {
	db, err := repo.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, []domain.OperatorHistory{}, errors.Wrap(err, "failed to get db")
	}

	queryOpts := cdomain.OperatorHistoryQueryOptionFunc()
	for _, opt := range options {
		opt(&queryOpts)
	}

	// find all operator histories
	var operatorHistories []domain.OperatorHistory
	count, err := orm.Table(db, domain.OperatorHistory{}).
		Select(domain.OPERATOR_HISTORY_STRINGS).
		InnerJoin(domain.AdminUser{}, "admin_user.id = operator_history.user_id").
		Like(domain.AdminUser{}.TableName()+".name", queryOpts.Name()).
		Equal(domain.OperatorHistory{}.TableName()+".menu", queryOpts.Menu()).
		Equal(domain.OperatorHistory{}.TableName()+".state", queryOpts.State()).
		Between(domain.OperatorHistory{}.TableName()+".created", queryOpts.StartDate(), queryOpts.EndDate()).
		SortOrder(queryOpts.SortBy(), queryOpts.Order()).
		Limit(queryOpts.Limit(), queryOpts.Offset()).
		FindAndCount(&operatorHistories)
	if err != nil {
		return 0, []domain.OperatorHistory{}, errors.Wrap(err, "err find and count")
	}

	return count, operatorHistories, nil
}

func (repo operatorHistoryMysqlRepository) Create(ctx context.Context, operatorHistory domain.OperatorHistory) error {
	db, err := repo.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "err get db")
	}

	if _, err = db.Insert(operatorHistory); err != nil {
		return errors.Wrap(err, "err insert operatorHistory")
	}

	return nil
}
