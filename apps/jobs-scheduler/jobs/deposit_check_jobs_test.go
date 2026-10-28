package jobs

import (
	"context"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

// TestFindExpiredDepositRequests_WaiterOriginDetection は、中間テーブル
// (investment_waiters_winning)経由で、入金期限切れメールの種別(WAITER_REQUEST)判定に使う
// InvestmentWaitersShareId が正しく埋まることを確認する(MIR-1118)。ここが壊れると、
// 待機者由来の投資に通常の督促メールが飛ぶ。メール送信を含む depositRequestExpireJob 本体では
// なく、クエリを切り出した findExpiredDepositRequests をテスト対象にしている。
func TestFindExpiredDepositRequests_WaiterOriginDetection(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	require.NoError(t, err)
	engine.SetMaxOpenConns(1)
	t.Cleanup(func() {
		require.NoError(t, engine.Close())
	})

	_, err = engine.Exec(`
CREATE TABLE deposit_request (
    id INTEGER PRIMARY KEY,
    share_id TEXT,
    project_id INTEGER,
    expire_deposit_date DATETIME,
    deleted DATETIME
);
CREATE TABLE share (
    id TEXT PRIMARY KEY,
    invest_status TEXT,
    deleted DATETIME
);
CREATE TABLE investment_waiters_winning (
    id INTEGER PRIMARY KEY,
    investment_waiters_id INTEGER,
    share_id TEXT,
    deleted DATETIME
);
CREATE TABLE investment_waiters (
    id INTEGER PRIMARY KEY,
    project_id INTEGER,
    user_id INTEGER,
    share_id TEXT,
    deleted DATETIME
);

-- share-waiter: 繰上げ(待機者由来)の投資。期限切れ。
INSERT INTO share (id, invest_status) VALUES ('share-waiter', 'REQUEST_INVEST');
INSERT INTO deposit_request (id, share_id, project_id, expire_deposit_date)
    VALUES (1, 'share-waiter', 10, '2020-01-01 00:00:00');
INSERT INTO investment_waiters_winning (investment_waiters_id, share_id) VALUES (900, 'share-waiter');

-- share-normal: 通常投資。期限切れ。中間テーブルに紐づけ無し。
INSERT INTO share (id, invest_status) VALUES ('share-normal', 'REQUEST_INVEST');
INSERT INTO deposit_request (id, share_id, project_id, expire_deposit_date)
    VALUES (2, 'share-normal', 10, '2020-01-01 00:00:00');

-- share-deleted-winning: 当選記録はあるが論理削除済み。旧列 investment_waiters.share_id にも
-- 値が残っている(本来起き得ない組み合わせだが、ガードが無ければ旧列経由で復活してしまう
-- ことを確認する最悪ケース)。通常投資と同じ扱いになるべき。
INSERT INTO share (id, invest_status) VALUES ('share-deleted-winning', 'REQUEST_INVEST');
INSERT INTO deposit_request (id, share_id, project_id, expire_deposit_date)
    VALUES (3, 'share-deleted-winning', 10, '2020-01-01 00:00:00');
INSERT INTO investment_waiters (id, project_id, user_id, share_id) VALUES (901, 10, 1, 'share-deleted-winning');
INSERT INTO investment_waiters_winning (investment_waiters_id, share_id, deleted)
    VALUES (901, 'share-deleted-winning', '2026-08-24 00:00:00');

-- share-not-expired: 期限が未来なので対象外。
INSERT INTO share (id, invest_status) VALUES ('share-not-expired', 'REQUEST_INVEST');
INSERT INTO deposit_request (id, share_id, project_id, expire_deposit_date)
    VALUES (4, 'share-not-expired', 10, '2099-01-01 00:00:00');

-- share-already-success: 期限切れだが既に入金確定済みなので対象外。
INSERT INTO share (id, invest_status) VALUES ('share-already-success', 'SUCCESS_INVEST');
INSERT INTO deposit_request (id, share_id, project_id, expire_deposit_date)
    VALUES (5, 'share-already-success', 10, '2020-01-01 00:00:00');

-- share-zero-deleted: 当選行の deleted がゼロ時刻。xormのCondDeletedではNULLと同じ
-- 「未削除」を意味するため、待機者由来として扱われるべき。
INSERT INTO share (id, invest_status) VALUES ('share-zero-deleted', 'REQUEST_INVEST');
INSERT INTO deposit_request (id, share_id, project_id, expire_deposit_date)
    VALUES (6, 'share-zero-deleted', 10, '2020-01-01 00:00:00');
INSERT INTO investment_waiters_winning (investment_waiters_id, share_id, deleted)
    VALUES (902, 'share-zero-deleted', '0001-01-01 00:00:00');

-- share-duplicate-winning: 同じshareを指す当選行が2件(本来起き得ないが、水増しされないことを
-- 確認する)。JOINではなくスカラーサブクエリにしているため行が増えないはず。
INSERT INTO share (id, invest_status) VALUES ('share-duplicate-winning', 'REQUEST_INVEST');
INSERT INTO deposit_request (id, share_id, project_id, expire_deposit_date)
    VALUES (7, 'share-duplicate-winning', 10, '2020-01-01 00:00:00');
INSERT INTO investment_waiters_winning (investment_waiters_id, share_id) VALUES (903, 'share-duplicate-winning');
INSERT INTO investment_waiters_winning (investment_waiters_id, share_id) VALUES (904, 'share-duplicate-winning');

`)
	require.NoError(t, err)

	ctx := context.WithValue(context.Background(), factory.ConextDBType, engine)
	items, totalCount, err := findExpiredDepositRequests(ctx)
	require.NoError(t, err)
	require.EqualValues(t, 5, totalCount)
	require.Len(t, items, 5, "同一shareを指す当選行が複数あっても行が水増しされないこと")

	byShareId := make(map[string]string, len(items))
	for _, item := range items {
		byShareId[item.Share.Id] = item.InvestmentWaitersShareId
	}

	assert.Equal(t, "share-waiter", byShareId["share-waiter"])
	assert.Equal(t, "", byShareId["share-normal"])
	assert.Equal(t, "", byShareId["share-deleted-winning"])
	assert.NotContains(t, byShareId, "share-not-expired")
	assert.NotContains(t, byShareId, "share-already-success")
	assert.Equal(t, "share-zero-deleted", byShareId["share-zero-deleted"])
	assert.Equal(t, "share-duplicate-winning", byShareId["share-duplicate-winning"])
}

// TestProcessDepositRequestItems_ContinuesPastOneFailure はMIR-1251の検証。
// 1件目(user_accountはあるがuserが無く、UserAccount.GetByUserIdが
// "Not Found User Account" で失敗する)で処理が止まらず、2件目
// (NeedAmount=0で何もせず成功する)まで処理が進むことを確認する。
//
// 修正前は1件目の失敗で即座にreturnしていたため、2件目は同じ実行では
// 一切処理されず、次回のジョブ実行まで待たされていた。このテストの
// フィクスチャでは2件目に副作用を持たせていないため、「2件目まで進んだ
// こと」は失敗件数(1)が件数(2)より少ないことで確認する。
func TestProcessDepositRequestItems_ContinuesPastOneFailure(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	require.NoError(t, err)
	engine.SetMaxOpenConns(1)
	t.Cleanup(func() {
		require.NoError(t, engine.Close())
	})

	_, err = engine.Exec(`
CREATE TABLE user (
    id INTEGER PRIMARY KEY,
    name TEXT,
    user_type TEXT
);
CREATE TABLE user_account (
    id INTEGER PRIMARY KEY,
    user_id INTEGER,
    balance INTEGER,
    updated DATETIME,
    version INTEGER
);
CREATE TABLE deposit_request (
    id INTEGER PRIMARY KEY,
    user_id INTEGER,
    share_id TEXT,
    project_id INTEGER,
    project_type TEXT,
    request_amount INTEGER,
    deposit_amount INTEGER,
    refund_amount INTEGER,
    deposit_date DATETIME,
    deposit_check_status TEXT,
    update_date DATETIME,
    expire_deposit_date DATETIME,
    deposit_bank_name TEXT,
    deposit_branch_name TEXT,
    deposit_branch_code TEXT,
    deposit_account_number TEXT,
    created DATETIME,
    deleted DATETIME
);

-- 1件目: user_accountはあるがuser行が無い(GetByUserIdの内部JOINで
-- 見つからず"Not Found User Account"になる)。
INSERT INTO user_account (id, user_id, balance, version) VALUES (1, 1, 1000, 1);
INSERT INTO deposit_request (id, user_id, share_id, project_id, project_type, request_amount, deposit_amount, deposit_check_status, expire_deposit_date)
    VALUES (1, 1, 'share-1', 1, 'EQUITY', 1000, 0, 'NO_DEPOSIT', '2099-01-01 00:00:00');

-- 2件目: user・user_accountとも揃っており、request_amount=deposit_amount
-- (NeedAmount=0)なのでDepositCheckProcessは何もせず成功する。
INSERT INTO user (id, name, user_type) VALUES (2, 'user2', 'PRIVATE');
INSERT INTO user_account (id, user_id, balance, version) VALUES (2, 2, 1000, 1);
INSERT INTO deposit_request (id, user_id, share_id, project_id, project_type, request_amount, deposit_amount, deposit_check_status, expire_deposit_date)
    VALUES (2, 2, 'share-2', 2, 'EQUITY', 1000, 1000, 'NO_DEPOSIT', '2099-01-01 00:00:00');
`)
	require.NoError(t, err)

	ctx := context.WithValue(context.Background(), factory.ConextDBType, engine)

	var items []domain.DepositRequest
	require.NoError(t, engine.Table("deposit_request").Asc("id").Find(&items))
	require.Len(t, items, 2)

	failedCount := processDepositRequestItems(ctx, items)
	assert.Equal(t, 1, failedCount, "1件目だけが失敗し、2件目までループが進んでいること")
}
