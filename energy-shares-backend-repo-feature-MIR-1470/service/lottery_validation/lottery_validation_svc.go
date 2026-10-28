package lottery_validation

import (
	"context"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type lotteryValidationService struct{}

func New() cdomain.LotteryValidationService {
	return &lotteryValidationService{}
}

func (s lotteryValidationService) CheckMyNumberStatusByUserID(ctx context.Context, userID int64) (bool, error) {
	return domain.MyNumber{}.CheckMyNumberStatusByUserId(ctx, userID)
}

func (s lotteryValidationService) CheckUserOverAge(ctx context.Context, user *domain.User) error {
	if user == nil {
		return nil
	}

	return user.CheckOverAge(ctx)
}

func (s lotteryValidationService) GetSuccessInvestorCountByProjectID(ctx context.Context, projectID int64) (int, error) {
	_, shares, err := domain.Share{}.GetByProjectId(ctx, projectID)
	if err != nil {
		return 0, err
	}

	successInvestorCount := 0
	for _, share := range shares {
		if share.InvestStatus == domain.SUCCESS_INVEST ||
			share.InvestStatus == domain.OPERATION_INVEST ||
			share.InvestStatus == domain.REFUND_INVEST {
			successInvestorCount++
		}
	}

	return successInvestorCount, nil
}

func (s lotteryValidationService) GetInvestedAmountByUserID(ctx context.Context, projectID, userID int64) (int64, error) {
	return domain.Share{}.GetAmountByUserId(ctx, projectID, userID)
}

func (s lotteryValidationService) GetWaitingAmountByUserID(ctx context.Context, projectID, userID int64) (int64, error) {
	return domain.InvestmentWaiters{}.GetTotalAmountForWaitingStatus(ctx, projectID, userID)
}
