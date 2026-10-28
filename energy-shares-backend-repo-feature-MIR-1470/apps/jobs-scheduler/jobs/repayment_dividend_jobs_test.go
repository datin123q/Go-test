package jobs

import (
	"context"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

func setupHasDividendRowsTestDB(t *testing.T) context.Context {
	t.Helper()

	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	require.NoError(t, err)
	engine.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, engine.Close()) })

	_, err = engine.Exec(`
		CREATE TABLE user (
			id INTEGER PRIMARY KEY AUTOINCREMENT
		);
		CREATE TABLE dividend (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER,
			settlement_id INTEGER,
			deleted DATETIME
		);
	`)
	require.NoError(t, err)

	return context.WithValue(context.Background(), factory.ConextDBType, engine)
}

// TestHasDividendRowsForSchedule_ExcludesSoftDeleted は、論理削除済みの配当行を
// 「配当行がある」と誤判定しないことを確認する(MIR-1249)。bean無しでCount()を
// 呼ぶとxormが論理削除の除外条件を付けないため、全行が論理削除されたスケジュールを
// hasAny=trueと判定してしまい、このガード自体が防ごうとしている「誰にも支払われ
// ないままCompletedになる」事象をすり抜けてしまう。
func TestHasDividendRowsForSchedule_ExcludesSoftDeleted(t *testing.T) {
	ctx := setupHasDividendRowsTestDB(t)
	engine := ctx.Value(factory.ConextDBType).(*xorm.Engine)

	_, err := engine.Exec(`INSERT INTO user (id) VALUES (1)`)
	require.NoError(t, err)
	_, err = engine.Exec(`INSERT INTO dividend (id, user_id, settlement_id, deleted) VALUES (1, 1, 100, '2026-01-01 00:00:00')`)
	require.NoError(t, err)

	hasAny, err := hasDividendRowsForSchedule(ctx, 100)
	require.NoError(t, err)
	assert.False(t, hasAny, "論理削除済みの行しか無いスケジュールはhasAny=falseであるべき")
}

func TestHasDividendRowsForSchedule_TrueWhenLiveRowsExist(t *testing.T) {
	ctx := setupHasDividendRowsTestDB(t)
	engine := ctx.Value(factory.ConextDBType).(*xorm.Engine)

	_, err := engine.Exec(`INSERT INTO user (id) VALUES (1)`)
	require.NoError(t, err)
	_, err = engine.Exec(`INSERT INTO dividend (id, user_id, settlement_id, deleted) VALUES (1, 1, 100, NULL)`)
	require.NoError(t, err)

	hasAny, err := hasDividendRowsForSchedule(ctx, 100)
	require.NoError(t, err)
	assert.True(t, hasAny)
}

// TestHasDividendRowsForSchedule_FalseWhenUserRowMissing は、GetByRepaymentDividendScheduleId
// と同じINNER JOIN userを効かせていることを確認する(MIR-1249)。JOIN無しで数えると、
// user行が無い孤立したdividend行まで「ある」と数えてしまい、GetByRepaymentDividendScheduleId
// (JOIN経由でその行を除外する)の0件という結果と食い違う。
func TestHasDividendRowsForSchedule_FalseWhenUserRowMissing(t *testing.T) {
	ctx := setupHasDividendRowsTestDB(t)
	engine := ctx.Value(factory.ConextDBType).(*xorm.Engine)

	_, err := engine.Exec(`INSERT INTO dividend (id, user_id, settlement_id, deleted) VALUES (1, 999, 100, NULL)`)
	require.NoError(t, err)

	hasAny, err := hasDividendRowsForSchedule(ctx, 100)
	require.NoError(t, err)
	assert.False(t, hasAny, "user行が無い孤立行はGetByRepaymentDividendScheduleIdのJOINと同じく除外されるべき")
}

func TestHasDividendRowsForSchedule_FalseWhenNoRows(t *testing.T) {
	ctx := setupHasDividendRowsTestDB(t)

	hasAny, err := hasDividendRowsForSchedule(ctx, 999)
	require.NoError(t, err)
	assert.False(t, hasAny)
}
