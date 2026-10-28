package clean_arch_domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type UserRoleService interface {
	GetRoles(ctx context.Context, id int64) ([]string, error)
}

type UserRoleQueryOption = func(*UserRoleQueryOptionFields)

type UserRoleQueryOptionFields struct {
}

func UserRoleQueryOptionFunc() UserRoleQueryOptionFields {
	return UserRoleQueryOptionFields{}
}

type UserRoleRepository interface {
	GetAll(ctx context.Context, options ...UserRoleQueryOption) (int64, []domain.AdminUser, error)
	Create(ctx context.Context, user domain.UserRoles) (int64, error)
	GetAllUserRoles(ctx context.Context, options ...UserRoleQueryOption) (int64, []domain.UserRoles, error)
	GetRoles(ctx context.Context, userId int64) ([]domain.UserRoles, error)
}
