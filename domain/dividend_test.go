package domain

import (
	"context"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

// setupDividendScheduleTestDB は GetByRepaymentDividendScheduleId が参照する
// dividend / user の最小構成を用意する。
func setupDividendScheduleTestDB(t *testing.T) context.Context {
	t.Helper()

	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	require.NoError(t, err)
	engine.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, engine.Close()) })

	_, err = engine.Exec(`
		CREATE TABLE user (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT
		);
		CREATE TABLE dividend (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER,
			settlement_id INTEGER,
			transaction_status TEXT,
			deleted DATETIME
		);
	`)
	require.NoError(t, err)

	return context.WithValue(context.Background(), factory.ConextDBType, engine)
}

// TestUserDividend_GetByRepaymentDividendScheduleId_ExcludesPaid は、
// 既に支払い済み(TRANSACTION_OK)の行だけを対象から除外することを確認する
// (MIR-1249)。TRANSACTION_NOだけに絞ると、NULLやTRANSACTION_FAILの行
// (過去データや異常系で発生しうる)が漏れて未払いのままスケジュールが
// Completedになってしまうため、TRANSACTION_OK以外は全て対象に含める。
func TestUserDividend_GetByRepaymentDividendScheduleId_ExcludesPaid(t *testing.T) {
	ctx := setupDividendScheduleTestDB(t)
	engine := ctx.Value(factory.ConextDBType).(*xorm.Engine)

	// user.id と dividend.id をあえてずらしておく。両テーブルに重複する
	// idカラムがSelect("user.*,dividend.*")でどちらの値に解決されるかを
	// このテストで検出できるようにするため(同じ値だと取り違えても気付けない)。
	_, err := engine.Exec(`INSERT INTO user (id, name) VALUES (101, 'user1'), (102, 'user2'), (103, 'user3'), (104, 'user4')`)
	require.NoError(t, err)
	_, err = engine.Exec(`
		INSERT INTO dividend (id, user_id, settlement_id, transaction_status) VALUES
		(1, 101, 100, 'TRANSACTION_NO'),
		(2, 102, 100, 'TRANSACTION_OK'),
		(3, 103, 100, 'TRANSACTION_FAIL'),
		(4, 104, 100, NULL),
		(5, 101, 200, 'TRANSACTION_NO')
	`)
	require.NoError(t, err)

	items, err := UserDividend{}.GetByRepaymentDividendScheduleId(ctx, 100)
	require.NoError(t, err)

	var gotIds []int64
	for _, item := range items {
		gotIds = append(gotIds, item.Dividend.Id)
	}
	assert.ElementsMatch(t, []int64{1, 3, 4}, gotIds, "TRANSACTION_OK以外(NULL・TRANSACTION_FAILを含む)を対象にし、支払い済みの行や別スケジュールの行は含めないこと")
}

// TestUserDividend_GetByRepaymentDividendScheduleId_ReturnsEmptyWhenAllPaid は、
// スケジュール内の全員が支払い済みなら空を返すことを確認する
// (再実行しても誰も再処理されない=二重加算が起きない、を保証する)。
func TestUserDividend_GetByRepaymentDividendScheduleId_ReturnsEmptyWhenAllPaid(t *testing.T) {
	ctx := setupDividendScheduleTestDB(t)
	engine := ctx.Value(factory.ConextDBType).(*xorm.Engine)

	_, err := engine.Exec(`INSERT INTO user (id, name) VALUES (1, 'user1')`)
	require.NoError(t, err)
	_, err = engine.Exec(`INSERT INTO dividend (id, user_id, settlement_id, transaction_status) VALUES (1, 1, 100, 'TRANSACTION_OK')`)
	require.NoError(t, err)

	items, err := UserDividend{}.GetByRepaymentDividendScheduleId(ctx, 100)
	require.NoError(t, err)
	assert.Empty(t, items)
}
