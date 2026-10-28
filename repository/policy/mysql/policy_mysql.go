package mysql

import (
	"context"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/orm"
	"github.com/pkg/errors"
)

type policyMysqlRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.PolicyRepository {
	return &policyMysqlRepository{
		dbFactory: dbFactory,
	}
}

func (r policyMysqlRepository) Create(ctx context.Context, policy domain.Policy) (int64, error) {
	db, err := r.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "err get db")
	}

	if _, err = db.Insert(&policy); err != nil {
		return 0, errors.Wrap(err, "err insert policy")
	}

	return policy.Id, nil
}

func (r policyMysqlRepository) GetAll(ctx context.Context, options ...cdomain.PolicyQueryOption) (int64, []domain.Policy, error) {
	db, err := r.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, nil, errors.Wrap(err, "err get db")
	}

	queryOpts := cdomain.PolicyQueryOptionFunc()
	for _, opt := range options {
		opt(&queryOpts)
	}

	policies := make([]domain.Policy, 0)

	totalCount, err := orm.Table(db, domain.Policy{}).
		Equal(domain.Policy{}.TableName()+".policy_view_type", queryOpts.PolicyViewType()).
		Equal(domain.Policy{}.TableName()+".policy_type", queryOpts.PolicyType()).
		SortOrder(queryOpts.SortBy(), queryOpts.Order()).
		Limit(queryOpts.Limit(), queryOpts.Offset()).
		FindAndCount(&policies)

	if err != nil {
		return 0, nil, errors.Wrap(err, "err find and count")
	}

	return totalCount, policies, nil
}

func (r policyMysqlRepository) GetById(ctx context.Context, id int64) (domain.Policy, error) {
	db, err := r.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return domain.Policy{}, errors.Wrap(err, "err get db")
	}

	var policy domain.Policy
	if has, err := db.ID(id).Get(&policy); err != nil {
		return domain.Policy{}, errors.Wrap(err, "err get")
	} else if !has {
		return domain.Policy{}, cdomain.ErrItemNotFound
	}

	return policy, nil
}

func (r policyMysqlRepository) Update(ctx context.Context, policy domain.Policy) error {
	db, err := r.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "err get db")
	}

	_, err = db.ID(policy.Id).Update(&policy)
	if err != nil {
		return errors.Wrap(err, "err update")
	}

	return nil
}

func (r policyMysqlRepository) Delete(ctx context.Context, id int64) error {
	db, err := r.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "err get db")
	}

	_, err = db.ID(id).Delete(&domain.Policy{})
	if err != nil {
		return errors.Wrap(err, "err delete")
	}

	return nil
}
