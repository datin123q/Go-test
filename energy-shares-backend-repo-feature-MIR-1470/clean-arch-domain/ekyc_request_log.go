package clean_arch_domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type EkycRequestLogRepository interface {
	Create(ctx context.Context, ekycRequestLog domain.EkycRequestLog) error
	Update(ctx context.Context, ekycRequestLog domain.EkycRequestLog) error
	GetByEkycId(ctx context.Context, ekycId string) (domain.EkycRequestLog, error)
	GetByUserId(ctx context.Context, userId int64) (domain.EkycRequestLog, error)
}
