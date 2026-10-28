package database_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	database "github.com/illuminarean/energy-shares-backend-repo/db"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

func TestVaDepositUniqueConstraintMigrationMatchesApplication(t *testing.T) {
	const migrationPath = "migrations/20260807000000_alter_va_deposit_add_unique.sql"

	content, err := database.MigrationsFS.ReadFile(migrationPath)
	require.NoError(t, err)
	sql := string(content)
	constraintName := domain.VaDepositTransactionUniqueKeyName

	assert.Contains(t, sql, "-- +goose Up")
	assert.Contains(t, sql, "-- +goose Down")
	assert.Contains(
		t,
		sql,
		fmt.Sprintf("ADD UNIQUE KEY %s (va_account_number, item_key)", constraintName),
	)
	assert.Contains(t, sql, fmt.Sprintf("DROP KEY %s", constraintName))
	assert.Equal(t, 2, strings.Count(sql, constraintName))
	assert.Equal(t, 2, strings.Count(sql, "ALGORITHM=INPLACE"))
	assert.Equal(t, 2, strings.Count(sql, "LOCK=NONE"))
}

// TestBackfillInvestmentWaitersWinningMigrationMatchesPR1 は、MIR-1163のバックフィルが
// PR1(20260825000000)と同じ抽出条件を保っていることを確認する。バックフィル自体が対策で
// あり(INSERT...SELECTは部分成功しないため、実行後に同条件で取りこぼしを数えても常に
// 0件になり検証にならない。レビューで判明し検証ロジックは削除した)、取りこぼしの有無は
// PR4(DROP COLUMN)の直前に人が確認する運用にしている。
func TestBackfillInvestmentWaitersWinningMigrationMatchesPR1(t *testing.T) {
	const migrationPath = "migrations/20260907000000_backfill_investment_waiters_winning.sql"

	content, err := database.MigrationsFS.ReadFile(migrationPath)
	require.NoError(t, err)
	sql := string(content)

	assert.Contains(t, sql, "-- +goose Up")
	assert.Contains(t, sql, "-- +goose Down")
	assert.Contains(t, sql, "INSERT INTO investment_waiters_winning (investment_waiters_id, share_id)")
	assert.Contains(t, sql, "iw.share_id IS NOT NULL AND iw.share_id != ''")
	assert.Contains(t, sql, "NOT EXISTS (\n      SELECT 1 FROM investment_waiters_winning iww\n      WHERE iww.investment_waiters_id = iw.id\n  )")
}
