package clean_arch_domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type PrivateUserRepository interface {
	GetById(ctx context.Context, id int64) (domain.PrivateUser, error)
	GetByEmail(ctx context.Context, email string) (domain.PrivateUser, error)
}
