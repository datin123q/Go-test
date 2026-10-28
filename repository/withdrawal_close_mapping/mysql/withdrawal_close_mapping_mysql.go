package mysql

import (
	"context"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/orm"
	"github.com/pkg/errors"
)

type withdrawalCloseMappingMysqlRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.WithdrawalCloseMappingRepository {
	return &withdrawalCloseMappingMysqlRepository{
		dbFactory: dbFactory,
	}
}

func (r withdrawalCloseMappingMysqlRepository) Create(ctx context.Context, withdrawalCloseMapping domain.WithdrawalCloseMapping) (int64, error) {
	db, err := r.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "err get db")
	}

	if _, err = db.Insert(&withdrawalCloseMapping); err != nil {
		return 0, errors.Wrap(err, "err insert withdrawal close mapping")
	}

	return withdrawalCloseMapping.Id, nil
}

func (r withdrawalCloseMappingMysqlRepository) CreateMany(ctx context.Context, withdrawalCloseMappings []domain.WithdrawalCloseMapping) error {
	db, err := r.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "err get db")
	}

	if _, err = db.Insert(&withdrawalCloseMappings); err != nil {
		return errors.Wrap(err, "err insert withdrawal close mappings")
	}

	return nil
}

func (r withdrawalCloseMappingMysqlRepository) GetAll(ctx context.Context, options ...cdomain.WithdrawalCloseMappingQueryOption) (int64, []domain.WithdrawalCloseMapping, error) {
	db, err := r.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, nil, errors.Wrap(err, "err get db")
	}

	queryOpts := cdomain.WithdrawalCloseMappingQueryOptionFunc()
	for _, opt := range options {
		opt(&queryOpts)
	}

	query := orm.Table(db, domain.WithdrawalCloseMapping{}).
		SortOrder(queryOpts.SortBy(), queryOpts.Order()).
		Limit(queryOpts.Limit(), queryOpts.Offset())

	if len(queryOpts.WithdrawalIds()) > 0 {
		query.In(domain.WithdrawalCloseMapping{}.TableName()+".withdrawal_id", queryOpts.WithdrawalIds())
	}

	withdrawalCloseMappings := make([]domain.WithdrawalCloseMapping, 0)
	totalCount, err := query.FindAndCount(&withdrawalCloseMappings)
	if err != nil {
		return 0, nil, errors.Wrap(err, "err find and count withdrawal close mappings")
	}

	return totalCount, withdrawalCloseMappings, nil
}
