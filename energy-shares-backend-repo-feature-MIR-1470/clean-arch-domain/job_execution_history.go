package clean_arch_domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type JobExecutionHistoryRepository interface {
	Create(ctx context.Context, history domain.JobExecutionHistory) (int64, error)
	GetById(ctx context.Context, id int64) (domain.JobExecutionHistory, error)
	Update(ctx context.Context, history domain.JobExecutionHistory) error
}
