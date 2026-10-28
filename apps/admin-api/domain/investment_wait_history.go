package domain

import (
	"context"
	"github.com/illuminarean/energy-share-core-api/controllers"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

// Deprecated : v1.17.0 이후 사용하지 않습니다.
type InvestmentWaitHistoryUsecase interface {
	GetAll(ctx context.Context, request InvestmentWaitHistoryGetAllRequest) (InvestmentWaitHistoryGetAllResponse, error) // Deprecated
}

type InvestmentWaitHistoryGetAllRequest struct {
	InvestmentWaitId *int64 `query:"investment_wait_id"` // 투자 대기자 목록 아이디
	controllers.SearchInput
}

type InvestmentWaitHistoryGetAllResponse struct {
	Items      []domain.InvestmentWaitHistory `json:"items"`
	TotalCount int64                          `json:"total_count"`
}

func (i InvestmentWaitHistoryGetAllResponse) WithTotalCount(totalCount int64) InvestmentWaitHistoryGetAllResponse {
	i.TotalCount = totalCount
	return i
}

func (i InvestmentWaitHistoryGetAllResponse) WithItems(items []domain.InvestmentWaitHistory) InvestmentWaitHistoryGetAllResponse {
	i.Items = items
	return i
}
