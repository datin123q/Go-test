package controllers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithdrawalInput_ToModel_IsManualFalse(t *testing.T) {
	got, err := (&WithdrawalInput{Amount: 10000}).ToModel()

	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, int64(10000), got.Amount)
	assert.False(t, got.IsManual, "user-facing withdrawal must set is_manual=false")
}
