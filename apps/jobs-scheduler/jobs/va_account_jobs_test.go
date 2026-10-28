package jobs

import (
	"context"
	stderrors "errors"
	"testing"

	client "github.com/gmoaozora/gmo-aozora-api-go/corporateclient"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/width"
	"xorm.io/xorm"
)

func setupEquityVaTransactionTestDB(t *testing.T) (context.Context, *xorm.Engine) {
	t.Helper()

	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	require.NoError(t, err)
	engine.SetMaxOpenConns(1)
	require.NoError(t, engine.Sync(new(domain.UserVirtualAccount), new(domain.VaDeposit)))
	t.Cleanup(func() {
		require.NoError(t, engine.Close())
	})

	ctx := context.WithValue(context.Background(), factory.ConextDBType, engine)
	return ctx, engine
}

func equityVaTransactionForTest(accountNumber, itemKey string) client.VaTransaction {
	return client.VaTransaction{
		VaId:            "538" + accountNumber,
		TransactionDate: "2026-01-15",
		ValueDate:       "2026-01-15",
		VaBranchCode:    "538",
		VaAccountNumber: accountNumber,
		DepositAmount:   "100000",
		ItemKey:         itemKey,
	}
}

func vaDepositFromEquityTransactionForTest(t *testing.T, transaction client.VaTransaction) domain.VaDeposit {
	t.Helper()

	return domain.VaDeposit{
		VaId:                  transaction.VaId,
		TransactionDate:       transaction.TransactionDate,
		ValueDate:             transaction.ValueDate,
		VaBranchCode:          transaction.VaBranchCode,
		ItemKey:               transaction.ItemKey,
		VaAccountNumber:       transaction.VaAccountNumber,
		DepositAmount:         100000,
		DepositStatus:         domain.NOT_USE_DEPOSIT,
		TransactionDetailType: domain.DEPOSIT_VIRTUAL_ACCOUNT,
	}
}

func TestProcessEquityVaTransactions_ContinuesAfterRegisteredTransaction(t *testing.T) {
	ctx, engine := setupEquityVaTransactionTestDB(t)
	registered := equityVaTransactionForTest("1234567", "registered-key")
	newTransaction := equityVaTransactionForTest("1234568", "new-key")
	existing := vaDepositFromEquityTransactionForTest(t, registered)
	_, err := engine.Insert(&existing)
	require.NoError(t, err)

	err = ProcessEquityVaTransactions(ctx, []client.VaTransaction{registered, newTransaction})

	require.NoError(t, err)
	count, err := engine.Unscoped().Count(new(domain.VaDeposit))
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
	hasNew, err := engine.Unscoped().Where("item_key = ?", newTransaction.ItemKey).Exist(new(domain.VaDeposit))
	require.NoError(t, err)
	assert.True(t, hasNew)
}

func TestProcessEquityVaTransactions_ReturnsErrorForMismatchedRegisteredTransaction(t *testing.T) {
	ctx, engine := setupEquityVaTransactionTestDB(t)
	transaction := equityVaTransactionForTest("1234567", "conflict-key")
	existing := vaDepositFromEquityTransactionForTest(t, transaction)
	_, err := engine.Insert(&existing)
	require.NoError(t, err)
	transaction.DepositAmount = "200000"

	err = ProcessEquityVaTransactions(ctx, []client.VaTransaction{transaction})

	require.Error(t, err)
	assert.True(t, stderrors.Is(err, domain.ErrVaDepositTransactionConflict))
	count, countErr := engine.Unscoped().Count(new(domain.VaDeposit))
	require.NoError(t, countErr)
	assert.Equal(t, int64(1), count)
}

func TestProcessEquityVaTransactions_ReturnsErrorForEmptyItemKey(t *testing.T) {
	tests := []struct {
		name    string
		itemKey string
	}{
		{name: "未指定", itemKey: ""},
		{name: "空白のみ", itemKey: " \n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, engine := setupEquityVaTransactionTestDB(t)
			transaction := equityVaTransactionForTest("1234567", tt.itemKey)

			err := ProcessEquityVaTransactions(ctx, []client.VaTransaction{transaction})

			require.Error(t, err)
			assert.True(t, stderrors.Is(err, domain.ErrVaDepositItemKeyRequired))
			count, countErr := engine.Unscoped().Count(new(domain.VaDeposit))
			require.NoError(t, countErr)
			assert.Equal(t, int64(0), count)
		})
	}
}

func Test_toFullWidth(t *testing.T) {
	t.Run("전각 데이터를 처리했을 때 아무 문제 없이 기존 데이터로 유지가 되는지 확인", func(t *testing.T) {
		got := width.Widen.String(convertDakutenAndHandakuten("ＧМＯあおぞらネット銀行"))
		assert.Equalf(t, "ＧМＯあおぞらネット銀行", got, "toFullWidth() = %v, want %v", got, "ＧМＯあおぞらネット銀行")
	})

	t.Run("전각 데이터가 아닌 경우 전각으로 변환되는지 확인_1", func(t *testing.T) {
		got := width.Widen.String(convertDakutenAndHandakuten("ｱｼﾞｻｲ"))
		assert.Equalf(t, "アジサイ", got, "toFullWidth() = %v, want %v", got, "アジサイ")
	})

	t.Run("전각 데이터가 아닌 경우 전각으로 변환되는지 확인_2", func(t *testing.T) {
		got := width.Widen.String(convertDakutenAndHandakuten("ｶ)ﾊｺﾌﾞﾈ"))
		assert.Equalf(t, "カ）ハコブネ", got, "toFullWidth() = %v, want %v", got, "カ）ハコブネ")
	})
}

func Test_convertDakutenAndHandakuten(t *testing.T) {
	t.Run("다쿠텐/반탁음이 포함된 문자열을 전각으로 변환하는지 확인", func(t *testing.T) {
		got := convertDakutenAndHandakuten("ｶﾞｷﾞｸﾞｹﾞｺﾞ")
		assert.Equalf(t, "ガギグゲゴ", got, "convertDakutenAndHandakuten() = %v, want %v", got, "ガギグゲゴ")
	})

	t.Run("다쿠텐/반탁음이 포함된 문자열을 전각으로 변환하는지 확인", func(t *testing.T) {
		got := convertDakutenAndHandakuten("ｻﾞｼﾞｽﾞｾﾞｿﾞ")
		assert.Equalf(t, "ザジズゼゾ", got, "convertDakutenAndHandakuten() = %v, want %v", got, "ザジズゼゾ")
	})
}
