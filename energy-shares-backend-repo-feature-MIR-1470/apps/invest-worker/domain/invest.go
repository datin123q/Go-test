package domain

import "context"

type InvestUsecase interface {
	ProcessInvestmentRequest(ctx context.Context, request InvestRequest,
	) error
}

type InvestRequest struct {
	ProjectId int64 `json:"project_id"`
	UserId    int64 `json:"user_id"`
	Qty       int64 `json:"qty"`
}
