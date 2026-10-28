package clean_arch_domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type AdminUserService interface {
	GetSuperAdmin(ctx context.Context) (domain.AdminUser, error)
	ExportCSV(ctx context.Context, items []domain.AdminUser) CSVData
	IsValidPeriod(ctx context.Context, adminUser domain.AdminUser) (bool, error)
}

type AdminUserRepository interface {
	GetAll(ctx context.Context, options ...AdminUserQueryOption) (int64, []domain.AdminUser, error)
	Create(ctx context.Context, user domain.AdminUser) (domain.AdminUser, error)
	Update(ctx context.Context, user domain.AdminUser) error
	Delete(ctx context.Context, userId int64) error
	GetById(ctx context.Context, userId int64) (domain.AdminUser, error)
	GetByEmail(ctx context.Context, email string) (domain.AdminUser, error)
}

type AdminUserQueryOption = func(*AdminUserQueryOptionFields)

func AdminUserQueryOptionFunc() AdminUserQueryOptionFields {
	return AdminUserQueryOptionFields{}
}

type AdminUserQueryOptionFields struct {
	name        string
	mobilePhone string
	roleCode    string
	QueryOptions
}

func (c AdminUserQueryOptionFields) WithName(name string) func(*AdminUserQueryOptionFields) {
	return func(cqo *AdminUserQueryOptionFields) {
		cqo.name = name
	}
}

func (c AdminUserQueryOptionFields) WithMobilePhone(mobilePhone string) func(*AdminUserQueryOptionFields) {
	return func(cqo *AdminUserQueryOptionFields) {
		cqo.mobilePhone = mobilePhone
	}
}

func (c AdminUserQueryOptionFields) WithRoleCode(roleCode string) func(*AdminUserQueryOptionFields) {
	return func(cqo *AdminUserQueryOptionFields) {
		cqo.roleCode = roleCode
	}
}

func (c AdminUserQueryOptionFields) WithSortBy(sortBy string) func(*AdminUserQueryOptionFields) {
	return func(qo *AdminUserQueryOptionFields) {
		qo.sortBy = &sortBy
	}
}

func (c AdminUserQueryOptionFields) WithOrder(order string) func(*AdminUserQueryOptionFields) {
	return func(qo *AdminUserQueryOptionFields) {
		qo.order = &order
	}
}

func (c AdminUserQueryOptionFields) WithOffset(offset int) func(*AdminUserQueryOptionFields) {
	return func(qo *AdminUserQueryOptionFields) {
		qo.offset = &offset
	}
}

func (c AdminUserQueryOptionFields) WithLimit(limit int) func(*AdminUserQueryOptionFields) {
	return func(qo *AdminUserQueryOptionFields) {
		qo.limit = &limit
	}
}

func (c AdminUserQueryOptionFields) Name() string {
	return c.name
}

func (c AdminUserQueryOptionFields) MobilePhone() string {
	return c.mobilePhone
}

func (c AdminUserQueryOptionFields) RoleCode() string {
	return c.roleCode
}
