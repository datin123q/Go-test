package controllers

import (
	"encoding/json"
	"testing"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithdrawalDetailDto_NewList_IsManualJSON(t *testing.T) {
	items := []domain.WithdrawalDetail{
		{
			Withdrawal: domain.Withdrawal{
				Id:           1,
				InvestorId:   "1",
				InvestorName: "手動出金ユーザー",
				Amount:       12000,
				Status:       domain.WITHDRAWAL_STATUS_INCOMPLETE,
				IsManual:     true,
			},
		},
		{
			Withdrawal: domain.Withdrawal{
				Id:           2,
				InvestorId:   "2",
				InvestorName: "通常出金ユーザー",
				Amount:       5000,
				Status:       domain.WITHDRAWAL_STATUS_INCOMPLETE,
				IsManual:     false,
			},
		},
	}

	got := WithdrawalDetailDto{}.NewList(items)
	require.Len(t, got, 2)
	assert.True(t, got[0].IsManual)
	assert.False(t, got[1].IsManual)

	data, err := json.Marshal(got)
	require.NoError(t, err)

	var decoded []map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &decoded))
	require.Len(t, decoded, 2)

	assert.Contains(t, decoded[0], "is_manual")
	assert.Contains(t, decoded[1], "is_manual")
	assert.Equal(t, true, decoded[0]["is_manual"])
	assert.Equal(t, false, decoded[1]["is_manual"])
}
