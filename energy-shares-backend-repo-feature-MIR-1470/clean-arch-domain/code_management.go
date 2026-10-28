package clean_arch_domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type CodeOwnershipRepository interface {
	Create(ctx context.Context, codeOwnership domain.CodeOwnership) error
	GetByCode(ctx context.Context, codeType domain.CodeType, code string) (*domain.CodeOwnership, error)
}

type CodeUsageRepository interface {
	Create(ctx context.Context, codeUsage domain.CodeUsage) error
}
