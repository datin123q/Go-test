package domain

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/commoncode"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"
)

func projectStatusPtrForInvestmentWaitersTest(status PROJECT_STATUS) *PROJECT_STATUS {
	return &status
}

func investStatusPtrForInvestmentWaitersTest(status INVEST_STATUS) *INVEST_STATUS {
	return &status
}

// InvestmentWaiterIsStillPending は Cancel API の二重処理防止ガードが使う判定式なので、
// 引数だけで完結する形で3ケースを固定する(MIR-1118)。
func TestInvestmentWaiterIsStillPending(t *testing.T) {
	tests := []struct {
		name        string
		status      INVESTMENT_WAITING_TYPE
		advancedQty int64
		want        bool
	}{
		{
			name:        "WAITINGかつ繰上げ済み累計が0なら処理対象",
			status:      WAITING,
			advancedQty: 0,
			want:        true,
		},
		{
			name:        "WAITINGでも繰上げ済み累計が1以上(当選済み)なら処理対象外",
			status:      WAITING,
			advancedQty: 1,
			want:        false,
		},
		{
			name:        "WAITING以外は当選の有無にかかわらず処理対象外",
			status:      WAITING_END,
			advancedQty: 0,
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, InvestmentWaiterIsStillPending(tt.status, tt.advancedQty))
		})
	}
}

// InvestmentWaiterCanAdvance は WaitInvest の繰上げ可否判定が使う判定式なので、
// 引数だけで完結する形で境界値を固定する(MIR-1068)。
func TestInvestmentWaiterCanAdvance(t *testing.T) {
	tests := []struct {
		name       string
		status     INVESTMENT_WAITING_TYPE
		remaining  int64
		hasWinning bool
		want       bool
	}{
		{
			name:      "WAITINGかつ繰上げ可能残が正なら処理対象",
			status:    WAITING,
			remaining: 1,
			want:      true,
		},
		{
			name:      "WAITINGでも繰上げ可能残が0なら処理対象外(申込口数を使い切っている)",
			status:    WAITING,
			remaining: 0,
			want:      false,
		},
		{
			name:      "繰上げ可能残が負値でも処理対象外(境界値の安全側)",
			status:    WAITING,
			remaining: -1,
			want:      false,
		},
		{
			name:       "WAITING_ENDでも繰上げ実績があり残があれば処理対象(再繰上げ)",
			status:     WAITING_END,
			remaining:  1,
			hasWinning: true,
			want:       true,
		},
		{
			name:       "WAITING_ENDで繰上げ実績が無ければ処理対象外(打ち切り)",
			status:     WAITING_END,
			remaining:  1,
			hasWinning: false,
			want:       false,
		},
		{
			name:       "WAITING_ENDで繰上げ実績があっても残が無ければ処理対象外",
			status:     WAITING_END,
			remaining:  0,
			hasWinning: true,
			want:       false,
		},
		{
			name:      "CANCELLEDは残の有無にかかわらず処理対象外",
			status:    CANCELLED,
			remaining: 1,
			want:      false,
		},
		{
			name:      "PROPOSALは残の有無にかかわらず処理対象外",
			status:    PROPOSAL,
			remaining: 1,
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, InvestmentWaiterCanAdvance(tt.status, tt.remaining, tt.hasWinning))
		})
	}
}

func TestInvestmentWaiters_GetStatusCodeGroup(t *testing.T) {
	require.NoError(t, i18nutils.InitI18nFiles("../assets"))

	want := commoncode.CodeGroup{
		GroupCode: INVESTMENT_WAITING_STATUS_CODE,
		Codes: []commoncode.Code{
			{Code: string(INVESTMENT_WAITING_STATUS_WAITING), Name: "キャンセル待ち", Order: 1},
			{Code: string(INVESTMENT_WAITING_STATUS_END_FUNDING), Name: "募集終了", Order: 2},
			{Code: string(INVESTMENT_WAITING_STATUS_CANCELLED), Name: "キャンセル", Order: 3},
			{Code: string(INVESTMENT_WAITING_STATUS_WINNING), Name: "当選（未入金）", Order: 4},
			{Code: string(INVESTMENT_WAITING_STATUS_WINNING_AFTER_SUCCESS), Name: "当選（確定）", Order: 5},
			{Code: string(INVESTMENT_WAITING_STATUS_WINNING_AFTER_CANCELED), Name: "当選（キャンセル）", Order: 6},
			{Code: string(INVESTMENT_WAITING_STATUS_WINNING_AFTER_FAILED), Name: "当選（不成立）", Order: 7},
			{Code: string(INVESTMENT_WAITING_STATUS_PROPOSAL), Name: "投資提案中", Order: 8},
		},
	}

	assert.Equal(t, want, (InvestmentWaiters{}).GetStatusCodeGroup("ja"))
}

func TestInvestmentWaiters_GetDisplayStatus(t *testing.T) {
	tests := []struct {
		name   string
		waiter InvestmentWaiters
		want   INVESTMENT_WAITING_STATUS
	}{
		{
			name:   "WAITINGかつProjectStatus未設定の場合はキャンセル待ち",
			waiter: InvestmentWaiters{Status: WAITING},
			want:   INVESTMENT_WAITING_STATUS_WAITING,
		},
		{
			name:   "WAITINGかつ募集開始前の場合はキャンセル待ち",
			waiter: InvestmentWaiters{Status: WAITING, ProjectStatus: projectStatusPtrForInvestmentWaitersTest(STANDBY_FUNDING)},
			want:   INVESTMENT_WAITING_STATUS_WAITING,
		},
		{
			name:   "WAITINGかつ募集中の場合はキャンセル待ち",
			waiter: InvestmentWaiters{Status: WAITING, ProjectStatus: projectStatusPtrForInvestmentWaitersTest(OPEN_FUNDING)},
			want:   INVESTMENT_WAITING_STATUS_WAITING,
		},
		{
			name:   "WAITINGかつ募集終了の場合はキャンセル待ち",
			waiter: InvestmentWaiters{Status: WAITING, ProjectStatus: projectStatusPtrForInvestmentWaitersTest(END_FUNDING)},
			want:   INVESTMENT_WAITING_STATUS_WAITING,
		},
		{
			name:   "WAITINGかつ成立の場合は募集終了",
			waiter: InvestmentWaiters{Status: WAITING, ProjectStatus: projectStatusPtrForInvestmentWaitersTest(SUCCESS_FUNDING)},
			want:   INVESTMENT_WAITING_STATUS_END_FUNDING,
		},
		{
			name:   "WAITINGかつ不成立の場合は募集終了",
			waiter: InvestmentWaiters{Status: WAITING, ProjectStatus: projectStatusPtrForInvestmentWaitersTest(FAIL_FUNDING)},
			want:   INVESTMENT_WAITING_STATUS_END_FUNDING,
		},
		{
			name:   "WAITINGかつ運用中の場合は募集終了",
			waiter: InvestmentWaiters{Status: WAITING, ProjectStatus: projectStatusPtrForInvestmentWaitersTest(OPERATION_FUNDING)},
			want:   INVESTMENT_WAITING_STATUS_END_FUNDING,
		},
		{
			name:   "WAITINGかつ運用終了の場合は募集終了(MIR-272)",
			waiter: InvestmentWaiters{Status: WAITING, ProjectStatus: projectStatusPtrForInvestmentWaitersTest(OPERATION_END_FUNDING)},
			want:   INVESTMENT_WAITING_STATUS_END_FUNDING,
		},
		{
			name:   "WAITINGかつ償還完了の場合は募集終了",
			waiter: InvestmentWaiters{Status: WAITING, ProjectStatus: projectStatusPtrForInvestmentWaitersTest(CLOSE_FUNDING)},
			want:   INVESTMENT_WAITING_STATUS_END_FUNDING,
		},
		{
			name:   "WAITINGかつ無効の場合は募集終了",
			waiter: InvestmentWaiters{Status: WAITING, ProjectStatus: projectStatusPtrForInvestmentWaitersTest(INVALID_FUNDING)},
			want:   INVESTMENT_WAITING_STATUS_END_FUNDING,
		},
		{
			name:   "DEPOSIT_REQUESTの場合は当選未入金",
			waiter: InvestmentWaiters{Status: DEPOSIT_REQUEST},
			want:   INVESTMENT_WAITING_STATUS_WINNING,
		},
		{
			name:   "PROPOSALの場合は投資提案中",
			waiter: InvestmentWaiters{Status: PROPOSAL},
			want:   INVESTMENT_WAITING_STATUS_PROPOSAL,
		},
		{
			name:   "CANCELLEDの場合はキャンセル",
			waiter: InvestmentWaiters{Status: CANCELLED},
			want:   INVESTMENT_WAITING_STATUS_CANCELLED,
		},
		{
			name:   "WAITING_ENDかつShareなしの場合は募集終了",
			waiter: InvestmentWaiters{Status: WAITING_END},
			want:   INVESTMENT_WAITING_STATUS_END_FUNDING,
		},
		{
			name:   "WAITING_ENDかつ投資申請中の場合は当選未入金",
			waiter: InvestmentWaiters{Status: WAITING_END, InvestStatus: investStatusPtrForInvestmentWaitersTest(REQUEST_INVEST)},
			want:   INVESTMENT_WAITING_STATUS_WINNING,
		},
		{
			name:   "WAITING_ENDかつ投資完了の場合は当選確定",
			waiter: InvestmentWaiters{Status: WAITING_END, InvestStatus: investStatusPtrForInvestmentWaitersTest(SUCCESS_INVEST)},
			want:   INVESTMENT_WAITING_STATUS_WINNING_AFTER_SUCCESS,
		},
		{
			name:   "WAITING_ENDかつ運用中の場合は当選確定",
			waiter: InvestmentWaiters{Status: WAITING_END, InvestStatus: investStatusPtrForInvestmentWaitersTest(OPERATION_INVEST)},
			want:   INVESTMENT_WAITING_STATUS_WINNING_AFTER_SUCCESS,
		},
		{
			name:   "WAITING_ENDかつ運用終了の場合は当選確定",
			waiter: InvestmentWaiters{Status: WAITING_END, InvestStatus: investStatusPtrForInvestmentWaitersTest(REFUND_INVEST)},
			want:   INVESTMENT_WAITING_STATUS_WINNING_AFTER_SUCCESS,
		},
		{
			name:   "WAITING_ENDかつ投資キャンセルの場合は当選キャンセル",
			waiter: InvestmentWaiters{Status: WAITING_END, InvestStatus: investStatusPtrForInvestmentWaitersTest(CANCEL_INVEST)},
			want:   INVESTMENT_WAITING_STATUS_WINNING_AFTER_CANCELED,
		},
		{
			name:   "WAITING_ENDかつ投資不成立の場合は当選不成立",
			waiter: InvestmentWaiters{Status: WAITING_END, InvestStatus: investStatusPtrForInvestmentWaitersTest(FAIL_INVEST)},
			want:   INVESTMENT_WAITING_STATUS_WINNING_AFTER_FAILED,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.waiter.GetDisplayStatus())
		})
	}
}

func TestInvestmentWaiters_GetByUserId_Statuses(t *testing.T) {
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
CREATE TABLE project (
    id INTEGER PRIMARY KEY,
    name TEXT,
    project_status TEXT
);
CREATE TABLE share (
    id TEXT PRIMARY KEY,
    qty INTEGER,
    invest_status TEXT,
    deleted DATETIME
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
INSERT INTO user (id, investor_type, name, email, mobile_phone)
VALUES (1, 'PRIVATE', 'test user', 'test@example.com', '00000000000');
INSERT INTO project (id, name, project_status)
VALUES (10, 'test project', 'OPEN_FUNDING');
INSERT INTO project (id, name, project_status)
VALUES (11, 'project without status', NULL);
INSERT INTO share (id, invest_status)
VALUES
    ('share-1', 'SUCCESS_INVEST'),
    ('share-2', 'CANCEL_INVEST'),
    ('share-3', 'REQUEST_INVEST'),
    ('share-4', 'SUCCESS_INVEST'),
    ('share-5', 'SUCCESS_INVEST'),
    ('share-legacy-resurrected', 'SUCCESS_INVEST');
-- MIR-1068の複数当選ケース(待機者110)用。AdvancedQtyExprの検証にqtyを使うため
-- 他のshareとは別にqtyを指定する。
INSERT INTO share (id, qty, invest_status)
VALUES
    ('share-multi-a', 3, 'CANCEL_INVEST'),
    ('share-multi-b', 4, 'SUCCESS_INVEST');
INSERT INTO investment_waiters (id, project_id, user_id, waiting_turn, qty, amount, status)
VALUES
    (100, 10, 1, 1, 1, 10000, 'WAITING'),
    (101, 10, 1, 2, 2, 20000, 'WAITING_END'),
    (102, 11, 1, 3, 3, 30000, 'WAITING'),
    (103, 10, 1, 4, 4, 40000, 'WAITING_END'),
    (104, 10, 1, 5, 5, 50000, 'WAITING_END'),
    (105, 10, 1, 6, 6, 60000, 'WAITING_END'),
    (106, 10, 1, 7, 7, 70000, 'WAITING_END'),
    (107, 10, 1, 8, 8, 80000, 'WAITING_END');
-- 109 は106と同じく当選記録が論理削除済みだが、さらに旧列 share_id にも値が残っている
-- (本来起き得ない組み合わせ)。旧列へのCOALESCEフォールバックはPR3(MIR-1163)で削除済みの
-- ため、旧列に何が入っていても share-legacy-resurrected が復活しないことを確認する。
INSERT INTO investment_waiters (id, project_id, user_id, waiting_turn, qty, amount, status, share_id)
VALUES
    (109, 10, 1, 10, 10, 100000, 'WAITING_END', 'share-legacy-resurrected');
-- 110(MIR-1068): 2回繰上げ済み(複数当選)。investment_waiters_winningへの実JOINが
-- 残っていれば、この待機者だけ行が2件に水増しされる。
INSERT INTO investment_waiters (id, project_id, user_id, waiting_turn, qty, amount, status)
VALUES
    (110, 10, 1, 11, 10, 100000, 'WAITING');
INSERT INTO investment_waiters_winning (investment_waiters_id, share_id)
VALUES
    (101, 'share-1'),
    (103, 'share-2'),
    (104, 'share-3'),
    (110, 'share-multi-a'),
    (110, 'share-multi-b');
-- 106 は当選記録があるが論理削除済み。募集失敗等による打ち切りと区別が付かず、
-- 集計・判定のいずれにも混ざらないことを確認する対象(MIR-1118)。
INSERT INTO investment_waiters_winning (investment_waiters_id, share_id, deleted)
VALUES
    (106, 'share-4', '2026-08-25 00:00:00');
-- 107 はゼロ時刻(0001-01-01)。xormのCondDeletedではNULLと同じ「未削除」を意味するため、
-- 106とは逆に当選済みとして扱われるべき(MIR-516で確認済みの判定と同じ)。
INSERT INTO investment_waiters_winning (investment_waiters_id, share_id, deleted)
VALUES
    (107, 'share-5', '0001-01-01 00:00:00');
INSERT INTO investment_waiters_winning (investment_waiters_id, share_id, deleted)
VALUES
    (109, 'share-legacy-resurrected', '2026-08-25 00:00:00');
`)
	require.NoError(t, err)

	ctx := context.WithValue(context.Background(), factory.ConextDBType, engine)
	totalCount, items, err := (InvestmentWaiters{}).GetByUserId(ctx, 1, "investment_waiters.id", "ASC", 0, 10)
	require.NoError(t, err)
	require.Equal(t, int64(10), totalCount)
	require.Len(t, items, 10)

	require.NotNil(t, items[0].ProjectStatus)
	assert.Equal(t, OPEN_FUNDING, *items[0].ProjectStatus)
	assert.Nil(t, items[0].InvestStatus)
	assert.Equal(t, INVESTMENT_WAITING_STATUS_WAITING, items[0].GetDisplayStatus())

	require.NotNil(t, items[1].ProjectStatus)
	assert.Equal(t, OPEN_FUNDING, *items[1].ProjectStatus)
	require.NotNil(t, items[1].InvestStatus)
	assert.Equal(t, SUCCESS_INVEST, *items[1].InvestStatus)
	assert.Equal(t, "share-1", items[1].ShareId)
	assert.Equal(t, INVESTMENT_WAITING_STATUS_WINNING_AFTER_SUCCESS, items[1].GetDisplayStatus())

	assert.Nil(t, items[2].ProjectStatus)
	assert.Nil(t, items[2].InvestStatus)

	assert.Equal(t, INVESTMENT_WAITING_STATUS_WINNING_AFTER_CANCELED, items[3].GetDisplayStatus())

	assert.Equal(t, INVESTMENT_WAITING_STATUS_WINNING, items[4].GetDisplayStatus())

	assert.Equal(t, INVESTMENT_WAITING_STATUS_END_FUNDING, items[5].GetDisplayStatus())

	// 論理削除済みの当選記録は「未当選」と同じ扱いになり、打ち切りと判定される。
	assert.Nil(t, items[6].InvestStatus)
	assert.Equal(t, "", items[6].ShareId)
	assert.Equal(t, INVESTMENT_WAITING_STATUS_END_FUNDING, items[6].GetDisplayStatus())

	// ゼロ時刻の当選記録は「未削除」を意味するため、当選済みとして扱われる。
	require.NotNil(t, items[7].InvestStatus)
	assert.Equal(t, SUCCESS_INVEST, *items[7].InvestStatus)
	assert.Equal(t, "share-5", items[7].ShareId)
	assert.Equal(t, INVESTMENT_WAITING_STATUS_WINNING_AFTER_SUCCESS, items[7].GetDisplayStatus())

	// 中間テーブルに論理削除済みの当選履歴が1件でもあれば、旧列に値が残っていても
	// 当選済みとして解決されないこと(旧列へのCOALESCEフォールバックはPR3で削除済み)。
	assert.Nil(t, items[8].InvestStatus)
	assert.Equal(t, "", items[8].ShareId)
	assert.Equal(t, INVESTMENT_WAITING_STATUS_END_FUNDING, items[8].GetDisplayStatus())

	// 複数当選(MIR-1068)でも行が水増しされず1件のままで、invest_statusは最新
	// (中間テーブルのid降順で先頭)の当選のものになること。AdvancedQtyは
	// CANCEL_INVEST(入金期限切れ)の分も含めた合計(3+4=7)になること
	// (設計書「入金期限切れでキャンセルされた口数は繰上げ済みとして消化する」の確認)。
	require.NotNil(t, items[9].InvestStatus)
	assert.Equal(t, SUCCESS_INVEST, *items[9].InvestStatus)
	assert.Equal(t, int64(7), items[9].AdvancedQty)
}

// TestInvestmentWaiters_QueriesWithoutShareIdColumn は、investment_waiters.share_id の
// 実カラムを撤去した後(PR4、MIR-1163)も、DEFAULT_INVESTMENT_WAIT_LIST を使う一覧・詳細系の
// クエリがエラー無く動作することを確認する(実際に本番へ反映する前の安全確認)。
//
// engine.Sync() を使う既存テストでは、xormが構造体の全フィールド(ShareId含む)に対応する
// カラムを作ってしまうため、実カラムが無い状態を再現できない。ここでは生SQLで
// share_id カラムを持たない investment_waiters テーブルを作り、xormにその構造体の
// SELECT列リストを暗黙に組み立てさせるのではなく DEFAULT_INVESTMENT_WAIT_LIST 等の
// 明示的な Select を経由する既存メソッドだけを呼ぶ形で確認する。
func TestInvestmentWaiters_QueriesWithoutShareIdColumn(t *testing.T) {
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
CREATE TABLE project (
    id INTEGER PRIMARY KEY,
    name TEXT,
    project_status TEXT
);
CREATE TABLE share (
    id TEXT PRIMARY KEY,
    qty INTEGER,
    invest_status TEXT,
    deleted DATETIME
);
-- share_id カラムを持たない。PR4(MIR-1163)でDROP COLUMNした後の本番スキーマを模す。
CREATE TABLE investment_waiters (
    id INTEGER PRIMARY KEY,
    project_id INTEGER,
    user_id INTEGER,
    waiting_turn INTEGER,
    qty INTEGER,
    amount INTEGER,
    status TEXT,
    waiting_date DATETIME,
    created DATETIME,
    deleted DATETIME
);
CREATE TABLE investment_waiters_winning (
    id INTEGER PRIMARY KEY,
    investment_waiters_id INTEGER,
    share_id TEXT,
    deleted DATETIME
);
INSERT INTO user (id, investor_type, name, email, mobile_phone)
VALUES (1, 'PRIVATE', 'test user', 'test@example.com', '00000000000');
INSERT INTO project (id, name, project_status)
VALUES (10, 'test project', 'OPEN_FUNDING');
INSERT INTO share (id, invest_status)
VALUES ('share-1', 'SUCCESS_INVEST');
INSERT INTO investment_waiters (id, project_id, user_id, waiting_turn, qty, amount, status, created)
VALUES
    (100, 10, 1, 1, 1, 10000, 'WAITING', '2026-09-01 00:00:00'),
    (101, 10, 1, 2, 2, 20000, 'WAITING_END', '2026-09-02 00:00:00');
INSERT INTO investment_waiters_winning (investment_waiters_id, share_id)
VALUES (101, 'share-1');
`)
	require.NoError(t, err)

	ctx := context.WithValue(context.Background(), factory.ConextDBType, engine)

	totalCount, byUser, err := (InvestmentWaiters{}).GetByUserId(ctx, 1, "investment_waiters.id", "ASC", 0, 10)
	require.NoError(t, err)
	require.Equal(t, int64(2), totalCount)
	require.Len(t, byUser, 2)
	assert.Equal(t, "", byUser[0].ShareId)
	assert.Equal(t, "share-1", byUser[1].ShareId)

	byId, err := (InvestmentWaiters{}).GetById(ctx, 101)
	require.NoError(t, err)
	require.NotNil(t, byId)
	assert.Equal(t, "share-1", byId.ShareId)

	projectIdStr := "10"
	allCount, all, err := (InvestmentWaiters{}).GetAll(ctx, &projectIdStr, "investment_waiters.id", "ASC", 0, 10)
	require.NoError(t, err)
	require.Equal(t, int64(2), allCount)
	require.Len(t, all, 2)

	byProject, err := (InvestmentWaiters{}).GetByProjectIdAndStatuses(ctx, 10, []INVESTMENT_WAITING_TYPE{WAITING})
	require.NoError(t, err)
	require.Len(t, byProject, 1)

	byProjectNoGroup, err := (InvestmentWaiters{}).GetByProjectIdAndStatusesWithoutGrouping(ctx, 10, []INVESTMENT_WAITING_TYPE{WAITING, WAITING_END})
	require.NoError(t, err)
	require.Len(t, byProjectNoGroup, 2)

	byUserAndProject, err := (InvestmentWaiters{}).GetByUserIdAndProjectID(ctx, 10, 1)
	require.NoError(t, err)
	require.Len(t, byUserAndProject, 2)
}

func TestInvestmentWaiters_GetTotalAmountForWaitingStatus(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	require.NoError(t, err)
	engine.SetMaxOpenConns(1)
	t.Cleanup(func() {
		require.NoError(t, engine.Close())
	})

	_, err = engine.Exec(`
CREATE TABLE investment_waiters (
    id INTEGER PRIMARY KEY,
    project_id INTEGER,
    user_id INTEGER,
    amount INTEGER,
    status TEXT,
    deleted DATETIME
);
INSERT INTO investment_waiters (id, project_id, user_id, amount, status)
VALUES
    (1, 10, 1, 100000, 'WAITING'),
    (2, 10, 1, 200000, 'WAITING'),
    (3, 10, 1, 300000, 'WAITING_END'),
    (4, 10, 1, 400000, 'CANCELLED'),
    (5, 10, 2, 500000, 'WAITING'),
    (6, 11, 1, 600000, 'WAITING');
`)
	require.NoError(t, err)

	ctx := context.WithValue(context.Background(), factory.ConextDBType, engine)

	// 同一ユーザー・同一プロジェクトの WAITING のみ合算する（複数回申込を想定）
	amount, err := (InvestmentWaiters{}).GetTotalAmountForWaitingStatus(ctx, 10, 1)
	require.NoError(t, err)
	assert.Equal(t, int64(300000), amount)

	amount, err = (InvestmentWaiters{}).GetTotalAmountForWaitingStatus(ctx, 10, 2)
	require.NoError(t, err)
	assert.Equal(t, int64(500000), amount)

	amount, err = (InvestmentWaiters{}).GetTotalAmountForWaitingStatus(ctx, 99, 1)
	require.NoError(t, err)
	assert.Equal(t, int64(0), amount)
}

// TestInvestmentWaitersWinning_Create は、繰上げ時に investment_service.go が呼ぶ
// InvestmentWaitersWinning.Create が実際に中間テーブルへ1行挿入することを確認する(MIR-1118)。
func TestInvestmentWaitersWinning_Create(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	require.NoError(t, err)
	engine.SetMaxOpenConns(1)
	t.Cleanup(func() {
		require.NoError(t, engine.Close())
	})
	require.NoError(t, engine.Sync(new(InvestmentWaitersWinning)))

	ctx := context.WithValue(context.Background(), factory.ConextDBType, engine)

	winning := InvestmentWaitersWinning{InvestmentWaitersId: 42, ShareId: "share-abc"}
	affected, err := winning.Create(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(1), affected)
	assert.NotZero(t, winning.Id, "自動採番されたIdがセットされること")

	var got InvestmentWaitersWinning
	has, err := engine.Where("investment_waiters_id = ?", 42).Get(&got)
	require.NoError(t, err)
	require.True(t, has)
	assert.Equal(t, "share-abc", got.ShareId)
	assert.True(t, got.Deleted.IsZero(), "作成直後は論理削除されていないこと")
}
