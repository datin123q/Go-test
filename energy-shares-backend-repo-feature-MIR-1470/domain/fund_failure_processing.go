package domain

import (
	"context"
	"sort"
)

type fundFailureProcessingContextKey struct{}

func WithFundFailureProcessing(ctx context.Context) context.Context {
	return context.WithValue(ctx, fundFailureProcessingContextKey{}, true)
}

func IsFundFailureProcessing(ctx context.Context) bool {
	value, ok := ctx.Value(fundFailureProcessingContextKey{}).(bool)
	return ok && value
}

func sortFundFailureInvestmentWaiters(items []InvestmentWaiters) {
	sort.SliceStable(items, func(i, j int) bool {
		left := items[i]
		right := items[j]

		leftPriority := fundFailureInvestmentWaiterPriority(left.Status)
		rightPriority := fundFailureInvestmentWaiterPriority(right.Status)
		if leftPriority != rightPriority {
			return leftPriority < rightPriority
		}

		if !left.WaitingDate.Time().Equal(right.WaitingDate.Time()) {
			return left.WaitingDate.Time().After(right.WaitingDate.Time())
		}

		if !left.Created.Time().Equal(right.Created.Time()) {
			return left.Created.Time().After(right.Created.Time())
		}

		return left.Id > right.Id
	})
}

func fundFailureInvestmentWaiterPriority(status INVESTMENT_WAITING_TYPE) int {
	switch status {
	case WAITING:
		return 0
	case DEPOSIT_REQUEST:
		return 1
	case PROPOSAL:
		return 2
	default:
		return 3
	}
}

func sortFundFailureShares(items []Share) {
	sort.SliceStable(items, func(i, j int) bool {
		left := items[i]
		right := items[j]

		if !left.InvestDate.Time().Equal(right.InvestDate.Time()) {
			return left.InvestDate.Time().After(right.InvestDate.Time())
		}

		if !left.Created.Time().Equal(right.Created.Time()) {
			return left.Created.Time().After(right.Created.Time())
		}

		return left.Id > right.Id
	})
}
