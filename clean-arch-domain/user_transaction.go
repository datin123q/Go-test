package clean_arch_domain

import (
	"context"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type UserTransactionRepository interface {
	Create(ctx context.Context, transaction domain.UserTransaction) (int64, error)
}
