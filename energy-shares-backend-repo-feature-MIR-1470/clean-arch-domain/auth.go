package clean_arch_domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type AuthService interface {
	CreateUserAccessToken(ctx context.Context, user domain.User) (string, error)
	CreateUserRefreshToken(ctx context.Context, input UserRefreshTokenInput) (string, error)
	CreateAdminAccessToken(ctx context.Context, admin domain.AdminUser) (string, error)
	CreateAdminRefreshToken(ctx context.Context, input AdminRefreshTokenInput) (string, error)
}

type UserRefreshTokenInput struct {
	User  domain.User
	IP    string
	Agent string
}

type AdminRefreshTokenInput struct {
	Admin domain.AdminUser
	IP    string
	Agent string
}
