package clean_arch_domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type LotteryValidationService interface {
	CheckMyNumberStatusByUserID(ctx context.Context, userID int64) (bool, error)
	CheckUserOverAge(ctx context.Context, user *domain.User) error
	GetSuccessInvestorCountByProjectID(ctx context.Context, projectID int64) (int, error)
	GetInvestedAmountByUserID(ctx context.Context, projectID, userID int64) (int64, error)
	GetWaitingAmountByUserID(ctx context.Context, projectID, userID int64) (int64, error)
}
