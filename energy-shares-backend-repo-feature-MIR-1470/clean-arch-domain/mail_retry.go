package clean_arch_domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type MailRetryRepository interface {
	Create(ctx context.Context, mailRetry domain.MailRetry) (int64, error)
}
