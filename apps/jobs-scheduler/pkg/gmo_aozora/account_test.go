package gmo_aozora

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAccounts_InitAccountInfo(t *testing.T) {
	accounts := Accounts{
		Account{}.New("1234567", "holder", "0310", "101", EQUITY_DEPOSIT),
		Account{}.New("1234567", "holder", "0310", "101", LOAN_DEPOSIT),
		Account{}.New("7654321", "holder", "0310", "102", EQUITY_WITHDRAWAL),
	}

	accounts.InitAccountInfo("1234567", "account-id", "01", "5000000")

	for _, account := range accounts[:2] {
		assert.Equal(t, "account-id", account.AccountId)
		assert.Equal(t, "01", account.AccountTypeCode)
		assert.Equal(t, "5000000", account.TransferLimitAmount)
	}

	assert.Empty(t, accounts[2].AccountId)
	assert.Empty(t, accounts[2].AccountTypeCode)
	assert.Empty(t, accounts[2].TransferLimitAmount)
}
