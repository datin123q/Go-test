package clean_arch_domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type MapSmsReceiverRepository interface {
	Create(ctx context.Context, receiver domain.MapSmsReceiver) (int64, error)
}
