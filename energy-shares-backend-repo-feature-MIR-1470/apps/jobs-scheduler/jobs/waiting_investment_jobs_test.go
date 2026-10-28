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

// TestFindWaitingInvestmentItems_ResolvesDepositCheckStatusAndAdvancedQty は、中間テーブル
// (investment_waiters_winning)経由で当選済みの待機者について、deposit_check_status が
// deposit_request まで正しく解決されること(MIR-1118)、繰上げ済み累計(AdvancedQty)が
// share.qtyの合計として正しく解決されること(MIR-1068)、1申込に複数当選があっても行が
// 水増しされないこと(MIR-1068)を確認する。
func TestFindWaitingInvestmentItems_ResolvesDepositCheckStatusAndAdvancedQty(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	require.NoError(t, err)
	engine.SetMaxOpenConns(1)
	t.Cleanup(func() {
		require.NoError(t, engine.Close())
	})

	_, err = engine.Exec(`
CREATE TABLE user (
    id INTEGER PRIMARY KEY,
    investor_type TEXT,
    name TEXT,
    email TEXT,
    mobile_phone TEXT
);
CREATE TABLE investment_waiters (
    id INTEGER PRIMARY KEY,
    project_id INTEGER,
    user_id INTEGER,
    waiting_turn INTEGER,
    qty INTEGER,
    amount INTEGER,
    status TEXT,
    waiting_date DATETIME,
    deleted DATETIME,
    share_id TEXT NULL
);
CREATE TABLE investment_waiters_winning (
    id INTEGER PRIMARY KEY,
    investment_waiters_id INTEGER,
    share_id TEXT,
    deleted DATETIME
);
CREATE TABLE deposit_request (
    id INTEGER PRIMARY KEY,
    share_id TEXT,
    deposit_check_status TEXT,
    deleted DATETIME
);
CREATE TABLE share (
    id TEXT PRIMARY KEY,
    qty INTEGER,
    deleted DATETIME
);

INSERT INTO user (id, investor_type, name, email, mobile_phone)
VALUES (1, 'PRIVATE', 'test user', 'test@example.com', '00000000000');

-- waiter-winning-table: 新コードで繰上げ当選済み(1回)。中間テーブル経由で deposit_request と
-- 結び付く。
INSERT INTO investment_waiters (id, project_id, user_id, waiting_turn, qty, amount, status)
VALUES (200, 10, 1, 1, 1, 10000, 'DEPOSIT_REQUEST');
INSERT INTO share (id, qty) VALUES ('share-winning', 1);
INSERT INTO investment_waiters_winning (investment_waiters_id, share_id) VALUES (200, 'share-winning');
INSERT INTO deposit_request (share_id, deposit_check_status) VALUES ('share-winning', 'REQUEST_INVEST');

-- waiter-not-won: 未当選のまま。deposit_check_status・advanced_qty とも空/0であるべき。
INSERT INTO investment_waiters (id, project_id, user_id, waiting_turn, qty, amount, status)
VALUES (202, 10, 1, 3, 3, 30000, 'WAITING');

-- waiter-multi-won(MIR-1068): 2回繰上げ済み。中間テーブルが2行あっても
-- findWaitingInvestmentItemsの結果は1行のままで、advanced_qtyは合計(3+4=7)になるべき。
INSERT INTO investment_waiters (id, project_id, user_id, waiting_turn, qty, amount, status)
VALUES (204, 10, 1, 4, 10, 100000, 'WAITING');
INSERT INTO share (id, qty) VALUES ('share-multi-a', 3);
INSERT INTO share (id, qty) VALUES ('share-multi-b', 4);
INSERT INTO investment_waiters_winning (investment_waiters_id, share_id) VALUES (204, 'share-multi-a');
INSERT INTO investment_waiters_winning (investment_waiters_id, share_id) VALUES (204, 'share-multi-b');

-- waiter-deleted-deposit-request: 当選済みだが対応するdeposit_requestが論理削除済み。
-- deposit_check_statusは空であるべき(論理削除済みの行を拾わないこと、①の敵対的検証で
-- 指摘された論理削除考慮漏れの回帰テスト)。advanced_qtyはdeposit_requestの状態と無関係に
-- share.qtyがそのまま合計されるべき。
INSERT INTO investment_waiters (id, project_id, user_id, waiting_turn, qty, amount, status)
VALUES (206, 10, 1, 5, 2, 20000, 'WAITING');
INSERT INTO share (id, qty) VALUES ('share-deleted-deposit', 2);
INSERT INTO investment_waiters_winning (investment_waiters_id, share_id) VALUES (206, 'share-deleted-deposit');
INSERT INTO deposit_request (share_id, deposit_check_status, deleted) VALUES ('share-deleted-deposit', 'CANCEL_DEPOSIT', '2026-01-01 00:00:00');
`)
	require.NoError(t, err)

	ctx := context.WithValue(context.Background(), factory.ConextDBType, engine)
	items, err := findWaitingInvestmentItems(ctx, 10)
	require.NoError(t, err)
	require.Len(t, items, 4)

	type got struct {
		depositCheckStatus string
		advancedQty        int64
	}
	byId := make(map[int64]got, len(items))
	for _, item := range items {
		byId[item.Id] = got{depositCheckStatus: string(item.DepositCheckStatus), advancedQty: item.AdvancedQty}
	}

	assert.Equal(t, got{depositCheckStatus: "REQUEST_INVEST", advancedQty: 1}, byId[200])
	assert.Equal(t, got{depositCheckStatus: "", advancedQty: 0}, byId[202])
	assert.Equal(t, got{depositCheckStatus: "", advancedQty: 7}, byId[204])
	assert.Equal(t, got{depositCheckStatus: "", advancedQty: 2}, byId[206])
}
