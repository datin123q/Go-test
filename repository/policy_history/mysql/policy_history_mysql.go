package mysql

import (
	"context"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/orm"
	"github.com/pkg/errors"
)

type policyHistoryRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.PolicyHistoryRepository {
	return &policyHistoryRepository{
		dbFactory: dbFactory,
	}
}

func (p policyHistoryRepository) Create(ctx context.Context, history domain.PolicyInfoHistory) (int64, error) {
	db, err := p.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "err get db")
	}

	if _, err = db.Insert(&history); err != nil {
		return 0, errors.Wrap(err, "err insert policy")
	}

	return history.Id, nil
}

func (p policyHistoryRepository) GetAll(ctx context.Context, options ...cdomain.PolicyHistoryQueryOption) (int64, []domain.PolicyInfoHistory, error) {
	db, err := p.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, nil, errors.Wrap(err, "err get db")
	}

	queryOpts := cdomain.PolicyHistoryQueryOptionFunc()
	for _, opt := range options {
		opt(&queryOpts)
	}

	histories := make([]domain.PolicyInfoHistory, 0)

	totalCount, err := orm.Table(db, domain.PolicyInfoHistory{}).
		Equal(domain.PolicyInfoHistory{}.TableName()+".policy_id", queryOpts.PolicyId()).
		Equal(domain.PolicyInfoHistory{}.TableName()+".is_posted", queryOpts.IsPosted()).
		Equal(domain.PolicyInfoHistory{}.TableName()+".label", queryOpts.Label()).
		SortOrder(queryOpts.SortBy(), queryOpts.Order()).
		Limit(queryOpts.Limit(), queryOpts.Offset()).
		FindAndCount(&histories)
	if err != nil {
		return 0, nil, errors.Wrap(err, "err get all policy")
	}

	return totalCount, histories, nil
}

func (p policyHistoryRepository) GetById(ctx context.Context, id int64) (domain.PolicyInfoHistory, error) {
	db, err := p.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return domain.PolicyInfoHistory{}, errors.Wrap(err, "err get db")
	}

	var history domain.PolicyInfoHistory
	if has, err := db.ID(id).Get(&history); err != nil {
		return domain.PolicyInfoHistory{}, errors.Wrap(err, "err get")
	} else if !has {
		return domain.PolicyInfoHistory{}, cdomain.ErrItemNotFound
	}

	return history, nil
}

func (p policyHistoryRepository) Update(ctx context.Context, history domain.PolicyInfoHistory) error {
	db, err := p.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "err get db")
	}

	if _, err = db.ID(history.Id).UseBool().Update(&history); err != nil {
		return errors.Wrap(err, "err update policy")
	}

	return nil
}
