package service

import (
	"context"
	"github.com/samber/lo"
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/service"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

type UserAdminService struct {
	withdrawalService service.IWithdrawalService
}

func NewUserAdminService() *UserAdminService {
	return &UserAdminService{
		withdrawalService: service.NewWithdrawalService(),
	}
}

func (s *UserAdminService) UploadVaCsv(ctx context.Context, ProjectType domain.PROJECT_TYPE, csvLines [][]string) ([]domain.UserVirtualAccount, error) {
	items := make([]domain.UserVirtualAccount, 0)
	for i, line := range csvLines {
		if i == 0 {
			continue
		}
		va := domain.UserVirtualAccount{
			VaBankName:       domain.GMO_AOZORA_BANK_NAME,
			VaBranchCode:     line[0],
			VaBranchName:     line[1],
			VaAccountNumber:  line[2],
			VaHolderNameKana: line[3],
			ProjectType:      ProjectType,
			LinkedStatus:     domain.OFF_VA_ACCOUNT,
			UpdateDate:       jsonutils.JSONTime(time.Now()),
		}

		_, err := va.Create(ctx)
		if err != nil {
			return nil, err
		}
		items = append(items, va)
	}

	return items, nil
}

func (s *UserAdminService) GetUserTotalAsset(ctx context.Context, userId int64) (int64, error) {

	// 투자금 (= 운용중 자산 + 투자 완료금)
	shareInfo, err := domain.UserShareAmountInfo{}.GetUserProjectShareAmount(ctx, userId)
	if err != nil {
		return 0, err
	}
	shareTotal := shareInfo.InvestOperationAmount + shareInfo.InvestSuccessAmount

	// 예치금
	account, err := domain.UserAccount{}.GetByUserId(ctx, userId)
	if err != nil {
		return 0, err
	}

	// 출금 의뢰금 (= 미완료 + 마감 + 송금 실패)
	withdrawalTotal, err := s.withdrawalService.GetTotalWithdrawalAmountByUserId(ctx, userId)
	if err != nil {
		return 0, err
	}

	// 자산 총액 = 투자금 + 예치금 + 출금 의뢰금
	totalAsset := shareTotal + lo.FromPtr(account.Balance) + withdrawalTotal

	return totalAsset, nil
}
