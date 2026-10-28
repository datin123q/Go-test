package mysql

import (
	"context"

	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/orm"
)

type roleMysqlRepository struct {
	dbFactory factory.DBFactory
}

func (r roleMysqlRepository) GetAll(ctx context.Context, options ...cdomain.RoleQueryOption) (int64, []domain.Role, error) {
	db, err := r.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, nil, errors.Wrap(err, "failed to get db")
	}

	opts := cdomain.RoleQueryOptionFunc()
	for _, opt := range options {
		opt(&opts)
	}

	roles := make([]domain.Role, 0)

	totalCount, err := orm.Table(db, domain.Role{}).
		Select(domain.Role{}.SelectAll()).
		Equal(domain.Role{}.TableName()+".user_id", opts.UserId()).
		SortOrder(opts.SortBy(), opts.Order()).
		Limit(opts.Limit(), opts.Offset()).
		FindAndCount(&roles)
	if err != nil {
		return 0, nil, errors.Wrap(err, "failed to get roles")
	}

	return totalCount, roles, nil

}

func (r roleMysqlRepository) Create(ctx context.Context, role domain.Role) (int64, error) {
	db, err := r.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "failed to get db")
	}

	_, err = db.InsertOne(&role)
	if err != nil {
		return 0, errors.Wrap(err, "failed to insert role")
	}

	return role.Id, nil
}

func (r roleMysqlRepository) GetById(ctx context.Context, id int64) (domain.Role, error) {
	db, err := r.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return domain.Role{}, errors.Wrap(err, "failed to get db")
	}

	role := domain.Role{}
	has, err := orm.Table(db, domain.Role{}).
		Select(domain.Role{}.SelectAll()).
		Equal(domain.Role{}.TableName()+".id", id).
		Limit(1, 0).
		Get(&role)
	if err != nil {
		return domain.Role{}, errors.Wrap(err, "failed to get role")
	}

	if !has {
		return domain.Role{}, cdomain.ErrItemNotFound
	}

	return role, nil
}

func (r roleMysqlRepository) GetByUserId(ctx context.Context, userId int64) ([]domain.Role, error) {
	db, err := r.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get db")
	}

	roles := make([]domain.Role, 0)
	err = orm.Table(db, domain.Role{}).
		Select(domain.Role{}.SelectAll()).
		Equal(domain.Role{}.TableName()+".user_id", userId).
		Find(&roles)

	if err != nil {
		return nil, errors.Wrap(err, "failed to get roles")
	}

	return roles, nil
}

func (r roleMysqlRepository) DeleteByUserId(ctx context.Context, userId int64) error {
	db, err := r.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get db")
	}

	affected, err := db.Where(domain.Role{}.TableName()+".user_id = ?", userId).Delete(&domain.Role{})
	if err != nil {
		return errors.Wrap(err, "failed to delete roles")
	}
	if affected == 0 {
		return cdomain.ErrItemNotAffected
	}
	return nil
}

func (r roleMysqlRepository) Delete(ctx context.Context, id int64) error {
	db, err := r.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get db")
	}

	affected, err := db.ID(id).Delete(&domain.Role{})
	if err != nil {
		return errors.Wrap(err, "failed to delete role")
	}
	if affected == 0 {
		return cdomain.ErrItemNotAffected
	}
	return nil
}

func New(dbFactory factory.DBFactory) cdomain.RoleRepository {
	return &roleMysqlRepository{
		dbFactory: dbFactory,
	}
}
