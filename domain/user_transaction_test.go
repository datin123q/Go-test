package domain

import (
	"context"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

// MIR-1202: WithdrawalUserTransactionWithBalanceは、WithdrawalUserTransactionから
// CalculateCustomerLedgerBalance呼び出しだけを外に出したもの(呼び出し元が算出済みの
// balanceをそのまま使う)で、実際に作られるUserTransaction行の形はWithdrawalUserTransaction
// と同じでなければならない。金額の符号(出金なのでマイナス)を落とすと、出資金が
// 顧客勘定元帳にプラスの入金として記録されてしまうため、ここで固定する。
func TestUserTransaction_WithdrawalUserTransactionWithBalance_CreatesNegativeAmountWithdrawal(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, engine.Close()) })
	require.NoError(t, engine.Sync(new(UserTransaction)))

	ctx := context.WithValue(context.Background(), factory.ConextDBType, engine)

	err = UserTransaction{}.WithdrawalUserTransactionWithBalance(
		ctx,
		2000,               // userId
		123456,             // balance(呼び出し元が既に算出済みの値)
		5000,               // amount
		1,                  // projectId
		EQUITY,             // projectType
		"",                 // accountNumber
		WITHDRAWAL_INVESTMENT,
		jsonutils.JSONTime(time.Date(2026, 9, 8, 12, 0, 0, 0, time.UTC)),
	)
	require.NoError(t, err)

	var got UserTransaction
	has, err := factory.DB(ctx).Where("user_id = ?", 2000).Get(&got)
	require.NoError(t, err)
	require.True(t, has)

	assert.Equal(t, int64(-5000), got.Amount, "出金なので符号がマイナスになること")
	assert.Equal(t, int64(123456), got.Balance, "呼び出し元が渡したbalanceがそのまま記録されること")
	assert.Equal(t, WITHDRAWAL_TRANSACTION, got.TransactionType)
	assert.Equal(t, WITHDRAWAL_INVESTMENT, got.TransactionDetailType)
	assert.Equal(t, int64(1), got.ProjectId)
}

// WithdrawalUserTransactionはWithdrawalUserTransactionWithBalanceへ委譲するだけになった
// (MIR-1202、二重に構築ロジックを持たないため)。CalculateCustomerLedgerBalance経由で
// user_asset_viewから算出したbalanceが、そのまま委譲先へ渡って記録されることを確認する。
func TestUserTransaction_WithdrawalUserTransaction_DelegatesToWithBalance(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, engine.Close()) })
	require.NoError(t, engine.Sync(new(UserTransaction)))

	// user_asset_viewは本番ではMySQLの複合ビューだが、CalculateCustomerLedgerBalanceは
	// このテーブル名から素直にGetするだけなので、同じ列を持つテーブルとして代用する。
	_, err = engine.Exec(`
CREATE TABLE user_asset_view (
    user_id INTEGER PRIMARY KEY,
    deposit INTEGER,
    invest_operation_amount INTEGER,
    invest_success_amount INTEGER,
    invest_request_amount INTEGER,
    withdrawal_request_amount INTEGER,
    unpaid_amount INTEGER,
    total_asset INTEGER
);
INSERT INTO user_asset_view (user_id, deposit, invest_operation_amount, invest_success_amount, invest_request_amount, withdrawal_request_amount, unpaid_amount, total_asset)
VALUES (3000, 10000, 999999, 20000, 30000, 40000, 0, 0);
`)
	require.NoError(t, err)

	ctx := context.WithValue(context.Background(), factory.ConextDBType, engine)

	// CalculateCustomerLedgerBalanceはDeposit+WithdrawalRequestAmount+InvestRequestAmount+
	// InvestSuccessAmountの和(InvestOperationAmountは含まない)なので、期待値は
	// 10000+40000+30000+20000=100000。999999(invest_operation_amount)が混ざらないことも
	// あわせて確認する。
	err = UserTransaction{}.WithdrawalUserTransaction(
		ctx,
		3000, // userId
		2000, // amount
		1,    // projectId
		EQUITY,
		"",
		WITHDRAWAL_INVESTMENT,
		jsonutils.JSONTime(time.Date(2026, 9, 8, 12, 0, 0, 0, time.UTC)),
	)
	require.NoError(t, err)

	var got UserTransaction
	has, err := factory.DB(ctx).Where("user_id = ?", 3000).Get(&got)
	require.NoError(t, err)
	require.True(t, has)

	assert.Equal(t, int64(-2000), got.Amount)
	assert.Equal(t, int64(100000), got.Balance, "CalculateCustomerLedgerBalanceの算出結果がそのまま委譲先へ渡ること")
}
