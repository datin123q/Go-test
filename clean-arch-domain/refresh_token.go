package clean_arch_domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type RefreshTokenRepository interface {
	Create(ctx context.Context, refreshToken domain.RefreshToken) (domain.RefreshToken, error)
	GetByToken(ctx context.Context, token string) (domain.RefreshToken, error)
	Update(ctx context.Context, refreshToken domain.RefreshToken) error
}
