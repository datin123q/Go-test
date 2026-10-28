package clean_arch_domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type LotteryHistoryRepository interface {
	Create(ctx context.Context, history domain.LotteryHistory) error
}
