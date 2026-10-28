package clean_arch_domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type RoleRepository interface {
	GetAll(ctx context.Context, options ...RoleQueryOption) (int64, []domain.Role, error)
	Create(ctx context.Context, role domain.Role) (int64, error)
	GetById(ctx context.Context, id int64) (domain.Role, error)
	GetByUserId(ctx context.Context, userId int64) ([]domain.Role, error)
	DeleteByUserId(ctx context.Context, userId int64) error
	Delete(ctx context.Context, id int64) error
}

type RoleQueryOptionFields struct {
	userId int64
	QueryOptions
}

type RoleQueryOption = func(*RoleQueryOptionFields)

func RoleQueryOptionFunc() RoleQueryOptionFields {
	return RoleQueryOptionFields{}
}

func (c RoleQueryOptionFields) WithUserId(userId int64) func(*RoleQueryOptionFields) {
	return func(cqo *RoleQueryOptionFields) {
		cqo.userId = userId
	}
}

func (c RoleQueryOptionFields) UserId() int64 {
	return c.userId
}
