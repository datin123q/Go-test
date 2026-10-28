package mysql

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/samber/lo"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/orm"
)

type adminUserMysqlRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.AdminUserRepository {
	return &adminUserMysqlRepository{
		dbFactory: dbFactory,
	}
}

func (a adminUserMysqlRepository) GetAll(ctx context.Context, options ...cdomain.AdminUserQueryOption) (int64, []domain.AdminUser, error) {
	db, err := a.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, []domain.AdminUser{}, errors.Wrap(err, "err get db")
	}

	opts := cdomain.AdminUserQueryOptionFunc()
	for _, opt := range options {
		opt(&opts)
	}

	adminUsers := make([]domain.AdminUser, 0)

	totalCnt, err := orm.Table(db, domain.AdminUser{}).Select(
		fmt.Sprintf("group_concat(%v order by role_code ASC) AS roles_string", domain.Role{}.TableName()+".role_code"),
		domain.AdminUser{}.SelectAll(),
	).
		LeftJoin(domain.Role{}, "admin_user.id = role.user_id").
		Like(domain.AdminUser{}.TableName()+".name", lo.ToPtr(opts.Name())).
		Like(domain.AdminUser{}.TableName()+".mobile_phone", lo.ToPtr(opts.MobilePhone())).
		Equal(domain.Role{}.TableName()+".role_code", opts.RoleCode()).
		GroupBy(domain.AdminUser{}.TableName()+".id").
		SortOrder(opts.SortBy(), opts.Order()).
		Limit(opts.Limit(), opts.Offset()).
		FindAndCount(&adminUsers)

	if err != nil {
		return 0, []domain.AdminUser{}, errors.Wrap(err, "err get admin users")
	}

	for i := range adminUsers {
		if adminUsers[i].RolesString == "" {
			return 0, []domain.AdminUser{}, errors.Wrap(cdomain.ErrItemNotFound, "roles not found")
		}
		adminUsers[i].Roles = strings.Split(adminUsers[i].RolesString, ",")
		adminUsers[i].IsSuperAdmin = lo.Contains(adminUsers[i].Roles, string(domain.ROLE_SUPER_ADMIN))
	}

	return totalCnt, adminUsers, nil
}

func (a adminUserMysqlRepository) Create(ctx context.Context, user domain.AdminUser) (domain.AdminUser, error) {
	db, err := a.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return domain.AdminUser{}, errors.Wrap(err, "err get db")
	}

	if _, err = db.Insert(&user); err != nil {
		return domain.AdminUser{}, errors.Wrap(err, "err insert admin user")
	}
	return user, nil
}

func (a adminUserMysqlRepository) GetById(ctx context.Context, userId int64) (domain.AdminUser, error) {
	db, err := a.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return domain.AdminUser{}, errors.Wrap(err, "err get db")
	}

	user := domain.AdminUser{}

	has, err := orm.Table(db, domain.AdminUser{}).Select(
		fmt.Sprintf("group_concat(%v order by role_code ASC) AS roles_string", domain.Role{}.TableName()+".role_code"),
		domain.AdminUser{}.SelectAll(),
	).LeftJoin(domain.Role{},
		fmt.Sprintf("%s.id = %s.user_id", domain.AdminUser{}.TableName(), domain.Role{}.TableName())).
		GroupBy(domain.AdminUser{}.TableName() + ".id").
		Id(userId).Get(&user)
	if err != nil {
		return domain.AdminUser{}, errors.Wrap(err, "failed to get share")
	}

	if !has {
		return domain.AdminUser{}, cdomain.ErrItemNotFound
	}
	if user.RolesString == "" {
		return domain.AdminUser{}, errors.Wrap(cdomain.ErrItemNotFound, "roles not found")
	}
	user.Roles = strings.Split(user.RolesString, ",")
	user.IsSuperAdmin = lo.Contains(user.Roles, string(domain.ROLE_SUPER_ADMIN))

	return user, nil
}

func (a adminUserMysqlRepository) Update(ctx context.Context, user domain.AdminUser) error {
	db, err := a.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "err get db")
	}

	has, err := db.ID(user.Id).UseBool().Update(&user)
	if err != nil {
		return errors.Wrap(err, "err update share")
	}
	if has == 0 {
		return cdomain.ErrItemNotFound
	}
	return nil
}

func (a adminUserMysqlRepository) Delete(ctx context.Context, userId int64) error {
	db, err := a.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "err get db")
	}

	has, err := db.ID(userId).Delete(&domain.AdminUser{})
	if err != nil {
		return errors.Wrap(err, "err update share")
	}
	if has == 0 {
		return cdomain.ErrItemNotFound
	}
	return nil
}

func (a adminUserMysqlRepository) GetByEmail(ctx context.Context, email string) (domain.AdminUser, error) {
	db, err := a.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return domain.AdminUser{}, errors.Wrap(err, "err get db")
	}

	user := domain.AdminUser{}

	has, err := orm.Table(db, domain.AdminUser{}).Select(
		fmt.Sprintf("group_concat(%v order by role_code ASC) AS roles_string", domain.Role{}.TableName()+".role_code"),
		domain.AdminUser{}.SelectAll(),
	).
		LeftJoin(domain.Role{}, "admin_user.id = role.user_id").
		Equal(domain.AdminUser{}.TableName()+".email", email).
		GroupBy(domain.AdminUser{}.TableName()+".id").
		Limit(1, 0).
		Get(&user)
	if err != nil {
		return domain.AdminUser{}, errors.Wrap(err, "failed to get share")
	}

	if !has {
		return domain.AdminUser{}, cdomain.ErrItemNotFound
	}

	if user.RolesString == "" {
		return domain.AdminUser{}, errors.Wrap(cdomain.ErrItemNotFound, "roles not found")
	}
	user.Roles = strings.Split(user.RolesString, ",")
	user.IsSuperAdmin = lo.Contains(user.Roles, string(domain.ROLE_SUPER_ADMIN))

	return user, nil
}
