package clean_arch_domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type SpvRepository interface {
	GetById(ctx context.Context, id int64) (domain.Spv, error)
}
