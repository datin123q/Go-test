package clean_arch_domain

import (
	"context"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type MarketingCodeRepository interface {
	Create(ctx context.Context, marketingCode domain.MarketingCode) (int64, error)
}
