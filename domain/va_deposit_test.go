package domain

import (
	"context"
	stderrors "errors"
	"fmt"
	"strings"
	"testing"

	"github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
	"xorm.io/xorm/contexts"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

type vaDepositDuplicateInsertHook struct {
	engine   *xorm.Engine
	winner   VaDeposit
	injected bool
}

func (h *vaDepositDuplicateInsertHook) BeforeProcess(c *contexts.ContextHook) (context.Context, error) {
	if h.injected || !strings.Contains(strings.ToUpper(c.SQL), "INSERT INTO `VA_DEPOSIT`") {
		return c.Ctx, nil
	}

	h.injected = true
	if _, err := h.engine.Insert(&h.winner); err != nil {
		return c.Ctx, err
	}

	return c.Ctx, &mysql.MySQLError{
		Number:  1062,
		Message: "Duplicate entry for key 'uq_va_deposit_account_item_key'",
	}
}

func (h *vaDepositDuplicateInsertHook) AfterProcess(*contexts.ContextHook) error {
	return nil
}

func setupVaDepositTestDB(t *testing.T) (context.Context, *xorm.Engine) {
	t.Helper()

	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	require.NoError(t, err)
	engine.SetMaxOpenConns(1)
	require.NoError(t, engine.Sync(new(VaDeposit)))
	t.Cleanup(func() {
		require.NoError(t, engine.Close())
	})

	ctx := context.WithValue(context.Background(), factory.ConextDBType, engine)
	return ctx, engine
}

func vaDepositForRegistrationTest() VaDeposit {
	return VaDeposit{
		UserId:          1001,
		VaId:            "test-va-id-001",
		TransactionDate: "2026-01-15",
		ValueDate:       "2026-01-15",
		VaAccountNumber: "1234567",
		ItemKey:         "test-item-key-001",
		DepositAmount:   100000,
	}
}

func TestValidateVaDepositItemKey(t *testing.T) {
	tests := []struct {
		name    string
		itemKey string
		wantErr bool
	}{
		{name: "正常値", itemKey: "test-item-key-001", wantErr: false},
		{name: "空文字", itemKey: "", wantErr: true},
		{name: "半角空白のみ", itemKey: "   ", wantErr: true},
		{name: "改行とタブのみ", itemKey: "\n\t", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVaDepositItemKey(tt.itemKey)
			if tt.wantErr {
				assert.ErrorIs(t, err, ErrVaDepositItemKeyRequired)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestVaDepositCreateOrConfirmExisting_RejectsEmptyItemKeyBeforeDBAccess(t *testing.T) {
	deposit := vaDepositForRegistrationTest()
	deposit.ItemKey = " "

	created, err := deposit.CreateOrConfirmExisting(context.Background())

	assert.False(t, created)
	assert.ErrorIs(t, err, ErrVaDepositItemKeyRequired)
}

func TestVaDepositCreate_RejectsEmptyItemKeyBeforeDBAccess(t *testing.T) {
	deposit := vaDepositForRegistrationTest()
	deposit.ItemKey = ""

	count, err := deposit.Create(context.Background())

	assert.Equal(t, int64(0), count)
	assert.ErrorIs(t, err, ErrVaDepositItemKeyRequired)
}

func TestVaDepositCreateOrConfirmExisting_CreatesNewTransaction(t *testing.T) {
	ctx, engine := setupVaDepositTestDB(t)
	deposit := vaDepositForRegistrationTest()

	created, err := deposit.CreateOrConfirmExisting(ctx)

	require.NoError(t, err)
	assert.True(t, created)
	count, err := engine.Unscoped().Count(new(VaDeposit))
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestVaDepositCreateOrConfirmExisting_ConfirmsMatchingTransaction(t *testing.T) {
	ctx, engine := setupVaDepositTestDB(t)
	existing := vaDepositForRegistrationTest()
	existing.RemitterNameKana = "既存振込人名"
	_, err := engine.Insert(&existing)
	require.NoError(t, err)

	incoming := vaDepositForRegistrationTest()
	incoming.UserId = 9999
	incoming.RemitterNameKana = "後着側振込人名"
	created, err := incoming.CreateOrConfirmExisting(ctx)

	require.NoError(t, err)
	assert.False(t, created)
	count, err := engine.Unscoped().Count(new(VaDeposit))
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestVaDepositCreateOrConfirmExisting_ConfirmsConcurrentWinningTransaction(t *testing.T) {
	ctx, engine := setupVaDepositTestDB(t)
	incoming := vaDepositForRegistrationTest()
	hook := &vaDepositDuplicateInsertHook{
		engine: engine,
		winner: vaDepositForRegistrationTest(),
	}
	engine.AddHook(hook)

	created, err := incoming.CreateOrConfirmExisting(ctx)

	require.NoError(t, err)
	assert.False(t, created)
	assert.True(t, hook.injected)
	count, err := engine.Unscoped().Count(new(VaDeposit))
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestVaDepositCreateOrConfirmExisting_RejectsMismatchedTransaction(t *testing.T) {
	ctx, engine := setupVaDepositTestDB(t)
	existing := vaDepositForRegistrationTest()
	_, err := engine.Insert(&existing)
	require.NoError(t, err)

	incoming := vaDepositForRegistrationTest()
	incoming.VaId = "different-va-id"
	incoming.DepositAmount = 200000
	created, err := incoming.CreateOrConfirmExisting(ctx)

	assert.False(t, created)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrVaDepositTransactionConflict)
	assert.Contains(t, err.Error(), "mismatched_fields=va_id,deposit_amount")
	count, countErr := engine.Unscoped().Count(new(VaDeposit))
	require.NoError(t, countErr)
	assert.Equal(t, int64(1), count)
}

func TestVaDepositGetByTransactionKey_IncludesSoftDeletedRecord(t *testing.T) {
	ctx, engine := setupVaDepositTestDB(t)
	existing := vaDepositForRegistrationTest()
	_, err := engine.Insert(&existing)
	require.NoError(t, err)
	_, err = engine.ID(existing.Id).Delete(new(VaDeposit))
	require.NoError(t, err)

	got, err := (VaDeposit{}).GetByTransactionKey(ctx, existing.VaAccountNumber, existing.ItemKey)

	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, existing.Id, got.Id)
	assert.False(t, got.Deleted.IsEmpty())
}

func TestIsVaDepositTransactionDuplicateError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "対象制約の1062",
			err: &mysql.MySQLError{
				Number:  1062,
				Message: "Duplicate entry '1234567-test-item-key-001' for key 'va_deposit.uq_va_deposit_account_item_key'",
			},
			want: true,
		},
		{
			name: "ラップされた対象制約の1062",
			err: fmt.Errorf("insert failed: %w", &mysql.MySQLError{
				Number:  1062,
				Message: "Duplicate entry 'value' for key 'uq_va_deposit_account_item_key'",
			}),
			want: true,
		},
		{
			name: "別制約の1062",
			err: &mysql.MySQLError{
				Number:  1062,
				Message: "Duplicate entry 'value' for key 'PRIMARY'",
			},
			want: false,
		},
		{
			name: "対象制約名を含む別制約の1062",
			err: &mysql.MySQLError{
				Number:  1062,
				Message: "Duplicate entry 'value' for key 'backup_uq_va_deposit_account_item_key'",
			},
			want: false,
		},
		{
			name: "対象制約でも1062以外",
			err: &mysql.MySQLError{
				Number:  1048,
				Message: "Column cannot be null: uq_va_deposit_account_item_key",
			},
			want: false,
		},
		{
			name: "MySQL以外のエラー",
			err:  stderrors.New("db error"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsVaDepositTransactionDuplicateError(tt.err))
		})
	}
}
