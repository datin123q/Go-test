package clean_arch_domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type CorporateProfileRepository interface {
	Create(ctx context.Context, corporateProfile domain.CorporateProfile) (int64, error)
	GetById(ctx context.Context, id int64) (domain.CorporateProfile, error)
	GetByCorpNo(ctx context.Context, corpNo string) (domain.CorporateProfile, error)
	GetByUserId(ctx context.Context, userId int64) (domain.CorporateProfile, error)
	Update(ctx context.Context, corporateProfile domain.CorporateProfile) error
}
