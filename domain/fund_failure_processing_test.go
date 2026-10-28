package domain

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

func TestWithFundFailureProcessing(t *testing.T) {
	base := context.Background()

	assert.False(t, IsFundFailureProcessing(base))
	assert.True(t, IsFundFailureProcessing(WithFundFailureProcessing(base)))
}

func TestShareResolveDepositAmountSkipsDuringFundFailureProcessing(t *testing.T) {
	share := Share{}

	err := share.ResolveDepositAmount(WithFundFailureProcessing(context.Background()))

	assert.NoError(t, err)
}

func TestSortFundFailureInvestmentWaiters(t *testing.T) {
	items := []InvestmentWaiters{
		{
			Id:          1,
			Status:      PROPOSAL,
			WaitingDate: jsonutils.JSONTime(time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC)),
			Created:     jsonutils.JSONTime(time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC)),
		},
		{
			Id:          2,
			Status:      WAITING,
			WaitingDate: jsonutils.JSONTime(time.Date(2026, 4, 2, 10, 0, 0, 0, time.UTC)),
			Created:     jsonutils.JSONTime(time.Date(2026, 4, 2, 10, 0, 0, 0, time.UTC)),
		},
		{
			Id:          3,
			Status:      DEPOSIT_REQUEST,
			WaitingDate: jsonutils.JSONTime(time.Date(2026, 4, 3, 10, 0, 0, 0, time.UTC)),
			Created:     jsonutils.JSONTime(time.Date(2026, 4, 3, 10, 0, 0, 0, time.UTC)),
		},
		{
			Id:          4,
			Status:      WAITING,
			WaitingDate: jsonutils.JSONTime(time.Date(2026, 4, 4, 10, 0, 0, 0, time.UTC)),
			Created:     jsonutils.JSONTime(time.Date(2026, 4, 4, 10, 0, 0, 0, time.UTC)),
		},
	}

	sortFundFailureInvestmentWaiters(items)

	assert.Equal(t, []int64{4, 2, 3, 1}, []int64{items[0].Id, items[1].Id, items[2].Id, items[3].Id})
}

func TestSortFundFailureShares(t *testing.T) {
	items := []Share{
		{
			Id:         "share-1",
			InvestDate: jsonutils.JSONTime(time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC)),
			Created:    jsonutils.JSONTime(time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC)),
		},
		{
			Id:         "share-3",
			InvestDate: jsonutils.JSONTime(time.Date(2026, 4, 3, 10, 0, 0, 0, time.UTC)),
			Created:    jsonutils.JSONTime(time.Date(2026, 4, 3, 10, 0, 0, 0, time.UTC)),
		},
		{
			Id:         "share-2",
			InvestDate: jsonutils.JSONTime(time.Date(2026, 4, 3, 10, 0, 0, 0, time.UTC)),
			Created:    jsonutils.JSONTime(time.Date(2026, 4, 2, 10, 0, 0, 0, time.UTC)),
		},
	}

	sortFundFailureShares(items)

	assert.Equal(t, []string{"share-3", "share-2", "share-1"}, []string{items[0].Id, items[1].Id, items[2].Id})
}
