package clean_arch_domain

import (
	"context"
	"fmt"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/samber/lo"
)

type InvestmentWaitHistoryRepository interface {
	Create(ctx context.Context, investmentWaitHistory domain.InvestmentWaitHistory) error
	GetAll(ctx context.Context, options ...InvestmentWaitHistoryQueryOption) (int64, []domain.InvestmentWaitHistory, error)
}

type InvestmentWaitHistoryService interface {
	CreateStatusHistory(ctx context.Context, comparer domain.InvestmentWaitersComparer, processMethod domain.USER_HISTORY_PROCESS_TYPE, updateId int64) error
}

type InvestmentWaitHistoryQueryOption = func(*InvestmentWaitHistoryQueryOptionFields)

// InvestmentWaitHistoryQueryOptionFunc Query Option
func InvestmentWaitHistoryQueryOptionFunc() InvestmentWaitHistoryQueryOptionFields {
	return InvestmentWaitHistoryQueryOptionFields{}
}

type InvestmentWaitHistoryQueryOptionFields struct {
	QueryOptions
	investmentWaitId *int64
}

func (iho InvestmentWaitHistoryQueryOptionFields) WithSortBy(sortBy string) InvestmentWaitHistoryQueryOption {
	return func(iho *InvestmentWaitHistoryQueryOptionFields) {
		iho.sortBy = &sortBy
	}
}

func (iho InvestmentWaitHistoryQueryOptionFields) WithOrder(order string) InvestmentWaitHistoryQueryOption {
	return func(iho *InvestmentWaitHistoryQueryOptionFields) {
		iho.order = &order
	}
}

func (iho InvestmentWaitHistoryQueryOptionFields) WithOffset(offset int) InvestmentWaitHistoryQueryOption {
	return func(iho *InvestmentWaitHistoryQueryOptionFields) {
		iho.offset = &offset
	}
}

func (iho InvestmentWaitHistoryQueryOptionFields) WithLimit(limit int) InvestmentWaitHistoryQueryOption {
	return func(iho *InvestmentWaitHistoryQueryOptionFields) {
		iho.limit = &limit
	}
}

func (iho InvestmentWaitHistoryQueryOptionFields) WithInvestmentWaitId(investmentWaitID int64) InvestmentWaitHistoryQueryOption {
	return func(iho *InvestmentWaitHistoryQueryOptionFields) {
		iho.investmentWaitId = &investmentWaitID
	}
}

func (iho InvestmentWaitHistoryQueryOptionFields) InvestmentWaitId() *string {
	if iho.investmentWaitId == nil {
		return nil
	}

	return lo.ToPtr(fmt.Sprintf("%v", *iho.investmentWaitId))
}
