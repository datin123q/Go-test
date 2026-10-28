package controllers

import "github.com/illuminarean/energy-shares-backend-repo/domain"

type InvestmentWaitersInput struct {
	Qty int64 `json:"qty" swagger:"desc(数量)" validate:"required"`
}

type InvestmentWaitersSearchInput struct {
	ProjectId *string `query:"project_id" swagger:"desc(プロジェクトID)"`
	SearchInput
}

type InvestmentWaitersStatusUpdateInput struct {
	Status domain.INVESTMENT_WAITING_TYPE `json:"status" swagger:"desc(INVESTMENT_WAITING_TYPE)" validate:"required"`
}
