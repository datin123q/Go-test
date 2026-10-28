package clean_arch_domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type PrivateProfileRepository interface {
	Create(ctx context.Context, privateProfile domain.PrivateProfile) error
	GetById(ctx context.Context, id int64) (domain.PrivateProfile, error)
	GetByUserId(ctx context.Context, userId int64) (domain.PrivateProfile, error)
	Update(ctx context.Context, privateProfile domain.PrivateProfile) error
}
