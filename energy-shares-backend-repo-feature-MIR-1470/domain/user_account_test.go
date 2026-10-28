package domain

import (
	"context"
	"encoding/csv"
	"net/http/httptest"
	"strconv"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

// setupUserAccountListTest は、GetAll / DownloadCsv が共通で使う
// deposit_balance の計算式を検証するための sqlite を用意する。
//
// 用意するユーザー:
//   - id=1: balance が NULL、投資申請中・出金申請中も無い(deposit_balance = 0)
//   - id=2: balance = 50000、投資申請中・出金申請中は無い(deposit_balance は
//     balance のまま変わらない)
//   - id=3: balance = 10000、REQUEST_INVEST の share に紐づく
//     deposit_request が 20000(deposit_balance = 30000)
//   - id=4: balance = 5000、WAIT 状態の withdrawal が 7000
//     (deposit_balance = 12000)
func setupUserAccountListTest(t *testing.T) context.Context {
	t.Helper()

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
CREATE TABLE share (
    id TEXT PRIMARY KEY,
    owner_id INTEGER,
    invest_status TEXT,
    deleted DATETIME
);
CREATE TABLE deposit_request (
    id INTEGER PRIMARY KEY,
    share_id TEXT,
    deposit_amount INTEGER,
    deleted DATETIME
);
CREATE TABLE withdrawal (
    id INTEGER PRIMARY KEY,
    investor_id INTEGER,
    amount INTEGER,
    status TEXT,
    deleted DATETIME
);

INSERT INTO user (id, name, user_type) VALUES
    (1, 'balance null user', 'PRIVATE'),
    (2, 'no deposit or withdrawal user', 'PRIVATE'),
    (3, 'has request invest user', 'PRIVATE'),
    (4, 'has waiting withdrawal user', 'PRIVATE');

INSERT INTO user_account (id, user_id, balance, version) VALUES
    (1, 1, NULL, 1),
    (2, 2, 50000, 1),
    (3, 3, 10000, 1),
    (4, 4, 5000, 1);

INSERT INTO share (id, owner_id, invest_status) VALUES
    ('share-3', 3, 'REQUEST_INVEST');
INSERT INTO deposit_request (id, share_id, deposit_amount) VALUES
    (1, 'share-3', 20000);

INSERT INTO withdrawal (id, investor_id, amount, status) VALUES
    (1, 4, 7000, 'WAIT');
`)
	require.NoError(t, err)

	return context.WithValue(context.Background(), factory.ConextDBType, engine)
}

// CSV の列位置は PrepareAndReturnCsv 内の userAccount 構造体のフィールド順
// (UserId, UserName, Balance, Updated)に対応する。ヘッダーの文言は
// i18n(serviceutils.GetServiceLocale)の設定次第で変わり、他のテストが
// 実行順序次第でロケールを切り替えるため、ヘッダー文字列を手がかりに
// 列を探す方法はテスト同士で不安定になる。構造体のフィールド順の方が
// 変更されにくく、変わった場合も構造体定義を見ればすぐ気づける。
const (
	csvColUserId  = 0
	csvColBalance = 2
)

// downloadUserAccountCsvRows は DownloadCsv の出力をパースして、
// ヘッダー行と、行ごとの [userId, balance] の並びを返す。
func downloadUserAccountCsvRows(t *testing.T, ctx context.Context, sortby, order string, userId, userName *string, startBalance, endBalance *int) (header []string, userIds, balances []int64) {
	t.Helper()

	w := httptest.NewRecorder()
	err := UserAccount{}.DownloadCsv(ctx, w, sortby, order, userId, userName, startBalance, endBalance)
	require.NoError(t, err)

	rows, err := csv.NewReader(w.Body).ReadAll()
	require.NoError(t, err)
	require.NotEmpty(t, rows, "ヘッダー行が無い")

	header = rows[0]
	require.Len(t, header, 4, "列数が変わった場合はcsvColUserId/csvColBalanceの見直しが必要")

	for _, row := range rows[1:] {
		id, err := strconv.ParseInt(row[csvColUserId], 10, 64)
		require.NoError(t, err)
		balance, err := strconv.ParseInt(row[csvColBalance], 10, 64)
		require.NoError(t, err)
		userIds = append(userIds, id)
		balances = append(balances, balance)
	}
	return header, userIds, balances
}

// downloadUserAccountCsv は DownloadCsv の出力を [userId]balance の map
// にして返す。
func downloadUserAccountCsv(t *testing.T, ctx context.Context, sortby, order string, userId, userName *string, startBalance, endBalance *int) map[int64]int64 {
	t.Helper()

	_, userIds, balances := downloadUserAccountCsvRows(t, ctx, sortby, order, userId, userName, startBalance, endBalance)

	result := map[int64]int64{}
	for i, id := range userIds {
		result[id] = balances[i]
	}
	return result
}

func TestUserAccount_GetAll_And_DownloadCsv_MatchBalance(t *testing.T) {
	ctx := setupUserAccountListTest(t)

	_, screenList, err := UserAccount{}.GetAll(ctx, "user_account.id", "ASC", 0, 10, nil, nil, nil, nil)
	require.NoError(t, err)
	require.Len(t, screenList, 4)

	screenBalance := map[int64]int64{}
	for _, u := range screenList {
		screenBalance[u.UserId] = *u.Balance
	}

	csvBalance := downloadUserAccountCsv(t, ctx, "user_account.id", "ASC", nil, nil, nil, nil)

	// balance が NULL のユーザーも含め、画面と CSV の金額が一致すること。
	assert.Equal(t, screenBalance, csvBalance)
	assert.Equal(t, int64(0), screenBalance[1], "balanceがNULLのユーザーはdeposit_balanceも0")
	assert.Equal(t, int64(50000), screenBalance[2], "投資申請中・出金申請中が無ければ balance のまま")
	assert.Equal(t, int64(30000), screenBalance[3], "投資申請中の deposit_request 分が加算される")
	assert.Equal(t, int64(12000), screenBalance[4], "出金申請中(WAIT)の withdrawal 分が加算される")
}

func TestUserAccount_GetAll_And_DownloadCsv_MatchCountOnBalanceRange(t *testing.T) {
	tests := []struct {
		name         string
		startBalance *int
		endBalance   *int
		wantUserIds  []int64
	}{
		{
			name:         "上限・下限の両方を指定",
			startBalance: lo.ToPtr(0),
			endBalance:   lo.ToPtr(40000),
			wantUserIds:  []int64{1, 3, 4}, // 2(50000)だけ範囲外
		},
		{
			name:         "下限のみ指定",
			startBalance: lo.ToPtr(20000),
			endBalance:   nil,
			wantUserIds:  []int64{2, 3}, // 30000,50000 が該当
		},
		{
			name:         "上限のみ指定",
			startBalance: nil,
			endBalance:   lo.ToPtr(0),
			wantUserIds:  []int64{1}, // deposit_balance = 0 のユーザーのみ
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := setupUserAccountListTest(t)

			totalCount, screenList, err := UserAccount{}.GetAll(ctx, "user_account.id", "ASC", 0, 10, nil, nil, tt.startBalance, tt.endBalance)
			require.NoError(t, err)

			var screenUserIds []int64
			for _, u := range screenList {
				screenUserIds = append(screenUserIds, u.UserId)
			}
			assert.ElementsMatch(t, tt.wantUserIds, screenUserIds, "画面側の絞り込み結果")
			assert.Equal(t, int64(len(tt.wantUserIds)), totalCount)

			csvBalance := downloadUserAccountCsv(t, ctx, "user_account.id", "ASC", nil, nil, tt.startBalance, tt.endBalance)
			var csvUserIds []int64
			for userId := range csvBalance {
				csvUserIds = append(csvUserIds, userId)
			}
			assert.ElementsMatch(t, tt.wantUserIds, csvUserIds, "CSV側の絞り込み結果は画面と一致すること")
		})
	}
}

// TestUserAccount_GetAll_And_DownloadCsv_MatchOnUserNameSearch は、
// userAccountListQuery が userId/userName の絞り込みも画面と CSV の
// 両方に反映することを確かめる(MIR-516 は balance 側の更新漏れが原因
// だったが、同じ「片方だけ更新する」ミスは userId/userName 側でも
// 起こりうるため)。
func TestUserAccount_GetAll_And_DownloadCsv_MatchOnUserNameSearch(t *testing.T) {
	ctx := setupUserAccountListTest(t)
	userName := lo.ToPtr("waiting")

	_, screenList, err := UserAccount{}.GetAll(ctx, "user_account.id", "ASC", 0, 10, nil, userName, nil, nil)
	require.NoError(t, err)

	var screenUserIds []int64
	for _, u := range screenList {
		screenUserIds = append(screenUserIds, u.UserId)
	}
	assert.Equal(t, []int64{4}, screenUserIds, "「waiting」を含むのは id=4 のユーザーのみ")

	csvBalance := downloadUserAccountCsv(t, ctx, "user_account.id", "ASC", nil, userName, nil, nil)
	var csvUserIds []int64
	for userId := range csvBalance {
		csvUserIds = append(csvUserIds, userId)
	}
	assert.Equal(t, []int64{4}, csvUserIds, "CSV側もuserNameの絞り込み結果が画面と一致すること")
}

func TestUserAccount_GetAll_And_DownloadCsv_MatchSortOrder(t *testing.T) {
	ctx := setupUserAccountListTest(t)

	_, screenList, err := UserAccount{}.GetAll(ctx, "user_account.id", "DESC", 0, 10, nil, nil, nil, nil)
	require.NoError(t, err)

	var screenOrder []int64
	for _, u := range screenList {
		screenOrder = append(screenOrder, u.UserId)
	}

	_, csvOrder, _ := downloadUserAccountCsvRows(t, ctx, "user_account.id", "DESC", nil, nil, nil, nil)

	assert.Equal(t, screenOrder, csvOrder, "sortby/order が同じなら画面とCSVの並び順も一致すること")
	assert.Equal(t, []int64{4, 3, 2, 1}, screenOrder)
}

func TestUserAccount_GetAll_PagingDoesNotAffectTotalCount(t *testing.T) {
	ctx := setupUserAccountListTest(t)

	totalCount, page1, err := UserAccount{}.GetAll(ctx, "user_account.id", "ASC", 0, 2, nil, nil, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(4), totalCount, "limit を絞っても totalCount は全件数のまま")
	require.Len(t, page1, 2)

	totalCount2, page2, err := UserAccount{}.GetAll(ctx, "user_account.id", "ASC", 2, 2, nil, nil, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(4), totalCount2)
	require.Len(t, page2, 2)

	assert.Equal(t, []int64{1, 2}, []int64{page1[0].UserId, page1[1].UserId})
	assert.Equal(t, []int64{3, 4}, []int64{page2[0].UserId, page2[1].UserId})
}

// TestUserAccount_DepositBalanceExpr_ExcludesLogicalDelete は、
// deposit_balance の式が share / deposit_request / withdrawal の
// いずれについても、論理削除(deleted に実際の削除時刻が入っている)行を
// 計算から除外することを確かめる(MIR-516)。式は生の SQL サブクエリの
// ため xorm の自動除外は効かず、`deleted IS NULL OR deleted =
// '0001-01-01 00:00:00'`(xorm の CondDeleted と同じ判定)を式自身に
// 明示している。
//
// 各ケースは対象テーブルに削除しない行をもう1行追加してから片方だけを
// 削除する。1行だけのフィクスチャで「削除した行を除外」を検証すると、
// 実装が誤って「該当ユーザーの行を全て除外」していても同じ結果になり
// 区別できないため。
//
// 「ゼロ時刻」ケースは deleted に '0001-01-01 00:00:00' を入れるが、
// これは xorm の CondDeleted 上は NULL と同じ「未削除」を表す値であり、
// 「削除された」ことにはならない。加算されたままになるのが正しい挙動
// (もし NULL のみで判定していたら、ゼロ時刻が書き込まれた行を削除済み
// と誤判定して残高を過少に計算してしまう)。
func TestUserAccount_DepositBalanceExpr_ExcludesLogicalDelete(t *testing.T) {
	tests := []struct {
		name         string
		setupSQL     string
		deleteSQL    string
		targetUserId int64
		wantBalance  int64
	}{
		{
			name: "withdrawalが実際の削除時刻で論理削除されていれば加算しない",
			setupSQL: `INSERT INTO withdrawal (id, investor_id, amount, status) VALUES
				(2, 4, 3000, 'WAIT');`,
			deleteSQL:    `UPDATE withdrawal SET deleted = '2026-08-24 00:00:00' WHERE id = 1;`,
			targetUserId: 4,
			wantBalance:  8000, // 5000(balance) + 3000(残す方) 。削除した7000は含めない
		},
		{
			name: "withdrawalがゼロ時刻(xormのCondDeletedでは未削除扱い)なら加算し続ける",
			setupSQL: `INSERT INTO withdrawal (id, investor_id, amount, status) VALUES
				(2, 4, 3000, 'WAIT');`,
			deleteSQL:    `UPDATE withdrawal SET deleted = '0001-01-01 00:00:00' WHERE id = 1;`,
			targetUserId: 4,
			wantBalance:  15000, // 5000(balance) + 7000(ゼロ時刻=未削除) + 3000
		},
		{
			name: "shareが実際の削除時刻で論理削除されていれば加算しない",
			setupSQL: `INSERT INTO share (id, owner_id, invest_status) VALUES ('share-4', 3, 'REQUEST_INVEST');
				INSERT INTO deposit_request (id, share_id, deposit_amount) VALUES (2, 'share-4', 15000);`,
			deleteSQL:    `UPDATE share SET deleted = '2026-08-24 00:00:00' WHERE id = 'share-3';`,
			targetUserId: 3,
			wantBalance:  25000, // 10000(balance) + 15000(残す方)。削除したshare-3経由の20000は含めない
		},
		{
			name: "shareがゼロ時刻(xormのCondDeletedでは未削除扱い)なら加算し続ける",
			setupSQL: `INSERT INTO share (id, owner_id, invest_status) VALUES ('share-4', 3, 'REQUEST_INVEST');
				INSERT INTO deposit_request (id, share_id, deposit_amount) VALUES (2, 'share-4', 15000);`,
			deleteSQL:    `UPDATE share SET deleted = '0001-01-01 00:00:00' WHERE id = 'share-3';`,
			targetUserId: 3,
			wantBalance:  45000, // 10000(balance) + 20000(ゼロ時刻=未削除) + 15000
		},
		{
			name: "deposit_requestが実際の削除時刻で論理削除されていれば加算しない",
			setupSQL: `INSERT INTO share (id, owner_id, invest_status) VALUES ('share-4', 3, 'REQUEST_INVEST');
				INSERT INTO deposit_request (id, share_id, deposit_amount) VALUES (2, 'share-4', 15000);`,
			deleteSQL:    `UPDATE deposit_request SET deleted = '2026-08-24 00:00:00' WHERE id = 1;`,
			targetUserId: 3,
			wantBalance:  25000,
		},
		{
			name: "deposit_requestがゼロ時刻(xormのCondDeletedでは未削除扱い)なら加算し続ける",
			setupSQL: `INSERT INTO share (id, owner_id, invest_status) VALUES ('share-4', 3, 'REQUEST_INVEST');
				INSERT INTO deposit_request (id, share_id, deposit_amount) VALUES (2, 'share-4', 15000);`,
			deleteSQL:    `UPDATE deposit_request SET deleted = '0001-01-01 00:00:00' WHERE id = 1;`,
			targetUserId: 3,
			wantBalance:  45000, // 10000(balance) + 20000(ゼロ時刻=未削除) + 15000
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := setupUserAccountListTest(t)

			engine := ctx.Value(factory.ConextDBType).(*xorm.Engine)
			_, err := engine.Exec(tt.setupSQL)
			require.NoError(t, err)
			_, err = engine.Exec(tt.deleteSQL)
			require.NoError(t, err)

			_, screenList, err := UserAccount{}.GetAll(ctx, "user_account.id", "ASC", 0, 10, nil, nil, nil, nil)
			require.NoError(t, err)

			var target *UserAccount
			for i := range screenList {
				if screenList[i].UserId == tt.targetUserId {
					target = &screenList[i]
				}
			}
			require.NotNil(t, target)

			assert.Equal(t, tt.wantBalance, *target.Balance)
		})
	}
}

// setupUserAccountBalanceTest は WithdrawBalance・DepositBalance の楽観ロック
// (version列)の挙動を検証するための最小限のsqliteを用意する。user_account
// 1件のみ。
func setupUserAccountBalanceTest(t *testing.T, balance int64) context.Context {
	t.Helper()

	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	require.NoError(t, err)
	engine.SetMaxOpenConns(1)
	t.Cleanup(func() {
		require.NoError(t, engine.Close())
	})

	_, err = engine.Exec(`
CREATE TABLE user_account (
    id INTEGER PRIMARY KEY,
    user_id INTEGER,
    balance INTEGER,
    updated DATETIME,
    version INTEGER
);
INSERT INTO user_account (id, user_id, balance, version) VALUES (1, 1, ?, 1);
`, balance)
	require.NoError(t, err)

	return context.WithValue(context.Background(), factory.ConextDBType, engine)
}

// MIR-1247: WithdrawBalance は xorm の楽観ロック(version列)でUPDATEするが、
// version不一致で0行更新になってもエラーを返さなかった(呼び出し元が
// 影響行数を見ていなかったことと合わさり、残高が減らないまま出資が
// 成立していた)。ここでは同じ行を2つの構造体に読み込み、片方で先に
// WithdrawBalanceを成功させてversionを進めたあと、もう片方(version が
// 古いまま)で同じ操作を行うことで、実際の競合を再現する。
func TestUserAccount_WithdrawBalance_VersionConflict(t *testing.T) {
	ctx := setupUserAccountBalanceTest(t, 10000)

	var first, stale UserAccount
	has, err := factory.DB(ctx).ID(int64(1)).Select(UserAccount{}.SelectAll()).Get(&first)
	require.NoError(t, err)
	require.True(t, has)
	has, err = factory.DB(ctx).ID(int64(1)).Select(UserAccount{}.SelectAll()).Get(&stale)
	require.NoError(t, err)
	require.True(t, has)

	// 別プロセスによる先行更新を、first側のWithdrawBalanceで再現する。
	idx, err := first.WithdrawBalance(ctx, 3000, WITHDRAWAL_INVESTMENT)
	require.NoError(t, err)
	require.EqualValues(t, 1, idx)

	staleUpdatedBefore := stale.Updated

	// staleはversionが古いまま(先行更新を知らない)なので、0行更新になる。
	idx, err = stale.WithdrawBalance(ctx, 2000, WITHDRAWAL_INVESTMENT)
	assert.Zero(t, idx, "version不一致なら0行更新になること")
	require.Error(t, err, "0行更新はエラーとして呼び出し元に伝わること")
	assert.ErrorIs(t, err, ErrVersionConflict, "呼び出し元がerrors.Isで判別できること")
	// 失敗時にBalance・Version・Updatedがstale側で一切変更されないことの
	// 固定(理由はapplyBalanceDeltaのdocコメントを参照)。
	assert.EqualValues(t, 10000, lo.FromPtr(stale.Balance), "更新に失敗したらBalanceは変化しないこと")
	assert.EqualValues(t, 1, stale.Version, "更新に失敗したらVersionも変化しないこと")
	assert.Equal(t, staleUpdatedBefore, stale.Updated, "更新に失敗したらUpdatedも変化しないこと")

	var current UserAccount
	has, err = factory.DB(ctx).ID(int64(1)).Select(UserAccount{}.SelectAll()).Get(&current)
	require.NoError(t, err)
	require.True(t, has)
	assert.EqualValues(t, 7000, lo.FromPtr(current.Balance), "staleの更新は失敗しているので、firstの分だけ減った残高のままであること")

	// リトライしても同じ理由で失敗し続け、残高が壊れないことを固定する。
	for i := 0; i < 3; i++ {
		idx, err = stale.WithdrawBalance(ctx, 2000, WITHDRAWAL_INVESTMENT)
		assert.Zero(t, idx)
		assert.Error(t, err, "リトライ %d回目も同じ理由で失敗し続けること", i+1)
	}
	has, err = factory.DB(ctx).ID(int64(1)).Select(UserAccount{}.SelectAll()).Get(&current)
	require.NoError(t, err)
	require.True(t, has)
	assert.EqualValues(t, 7000, lo.FromPtr(current.Balance), "リトライを繰り返しても残高が変化しないこと")
}

// 対照実験: versionが一致していれば従来どおり正常に減算されることを確認する。
func TestUserAccount_WithdrawBalance_Success(t *testing.T) {
	ctx := setupUserAccountBalanceTest(t, 10000)

	var account UserAccount
	has, err := factory.DB(ctx).ID(int64(1)).Select(UserAccount{}.SelectAll()).Get(&account)
	require.NoError(t, err)
	require.True(t, has)

	idx, err := account.WithdrawBalance(ctx, 3000, WITHDRAWAL_INVESTMENT)
	require.NoError(t, err)
	assert.EqualValues(t, 1, idx)

	var current UserAccount
	has, err = factory.DB(ctx).ID(int64(1)).Select(UserAccount{}.SelectAll()).Get(&current)
	require.NoError(t, err)
	require.True(t, has)
	assert.EqualValues(t, 7000, lo.FromPtr(current.Balance))
}

// 境界値: 残高ちょうど全額の出金は許可され、残高は0になる。
func TestUserAccount_WithdrawBalance_ExactBalance_Succeeds(t *testing.T) {
	ctx := setupUserAccountBalanceTest(t, 10000)

	var account UserAccount
	has, err := factory.DB(ctx).ID(int64(1)).Select(UserAccount{}.SelectAll()).Get(&account)
	require.NoError(t, err)
	require.True(t, has)

	idx, err := account.WithdrawBalance(ctx, 10000, WITHDRAWAL_INVESTMENT)
	require.NoError(t, err)
	assert.EqualValues(t, 1, idx)
	assert.EqualValues(t, 0, lo.FromPtr(account.Balance))

	var current UserAccount
	has, err = factory.DB(ctx).ID(int64(1)).Select(UserAccount{}.SelectAll()).Get(&current)
	require.NoError(t, err)
	require.True(t, has)
	assert.EqualValues(t, 0, lo.FromPtr(current.Balance), "全額出金後、DB上のbalanceも0であること(AllCols()でゼロ値も書き込まれること)")
}

// 境界値: 残高を1円でも超える出金は、DBへ触れる前に弾かれる。
func TestUserAccount_WithdrawBalance_ExceedsBalance_Fails(t *testing.T) {
	ctx := setupUserAccountBalanceTest(t, 10000)

	var account UserAccount
	has, err := factory.DB(ctx).ID(int64(1)).Select(UserAccount{}.SelectAll()).Get(&account)
	require.NoError(t, err)
	require.True(t, has)

	idx, err := account.WithdrawBalance(ctx, 10001, WITHDRAWAL_INVESTMENT)
	assert.Zero(t, idx)
	require.Error(t, err)
	assert.EqualValues(t, 10000, lo.FromPtr(account.Balance), "残高不足で弾かれた場合はaが変更されないこと")

	var current UserAccount
	has, err = factory.DB(ctx).ID(int64(1)).Select(UserAccount{}.SelectAll()).Get(&current)
	require.NoError(t, err)
	require.True(t, has)
	assert.EqualValues(t, 10000, lo.FromPtr(current.Balance), "DBのbalanceも変化しないこと")
}

// MIR-1247: DepositBalanceもWithdrawBalanceと同じ理由(xormが影響行数を
// 見ずにVersionを進める)で、version不一致時にBalance・Versionを戻さないと
// リトライ時に残高が壊れる。WithdrawBalanceと対称に固定する。
func TestUserAccount_DepositBalance_VersionConflict(t *testing.T) {
	ctx := setupUserAccountBalanceTest(t, 10000)

	var first, stale UserAccount
	has, err := factory.DB(ctx).ID(int64(1)).Select(UserAccount{}.SelectAll()).Get(&first)
	require.NoError(t, err)
	require.True(t, has)
	has, err = factory.DB(ctx).ID(int64(1)).Select(UserAccount{}.SelectAll()).Get(&stale)
	require.NoError(t, err)
	require.True(t, has)

	idx, err := first.DepositBalance(ctx, 3000)
	require.NoError(t, err)
	require.EqualValues(t, 1, idx)
	// DepositBalanceはallCols=falseの経路(WithdrawBalanceと非対称)なので、
	// 成功時にfirst自身へ書き戻される(*a = working)ことをここで固定する。
	assert.EqualValues(t, 13000, lo.FromPtr(first.Balance), "成功したらfirstのBalanceにも反映されること")
	assert.EqualValues(t, 2, first.Version, "成功したらfirstのVersionも進むこと")

	staleUpdatedBefore := stale.Updated

	idx, err = stale.DepositBalance(ctx, 2000)
	assert.Zero(t, idx, "version不一致なら0行更新になること")
	require.Error(t, err, "0行更新はエラーとして呼び出し元に伝わること")
	assert.ErrorIs(t, err, ErrVersionConflict, "呼び出し元がerrors.Isで判別できること")
	assert.EqualValues(t, 10000, lo.FromPtr(stale.Balance), "更新に失敗したらBalanceは変化しないこと")
	assert.EqualValues(t, 1, stale.Version, "更新に失敗したらVersionも変化しないこと")
	assert.Equal(t, staleUpdatedBefore, stale.Updated, "更新に失敗したらUpdatedも変化しないこと")

	var current UserAccount
	has, err = factory.DB(ctx).ID(int64(1)).Select(UserAccount{}.SelectAll()).Get(&current)
	require.NoError(t, err)
	require.True(t, has)
	assert.EqualValues(t, 13000, lo.FromPtr(current.Balance), "staleの更新は失敗しているので、firstの分だけ増えた残高のままであること")
}

// MIR-1247: applyBalanceDeltaはerr != nilをidx == 0より先に判定する。この順序を
// 誤ると、デッドロック等の本物のSQLエラーが「バージョン不一致」の固定文言に
// 置き換わってしまい、原因の切り分けができなくなる(敵対的検証で指摘された)。
// ここではテーブル自体を落としてSQLレベルのエラーを発生させ、固定文言に
// 化けないこと・aが一切変更されないことを固定する。
func TestUserAccount_WithdrawBalance_SqlError(t *testing.T) {
	ctx := setupUserAccountBalanceTest(t, 10000)

	var account UserAccount
	has, err := factory.DB(ctx).ID(int64(1)).Select(UserAccount{}.SelectAll()).Get(&account)
	require.NoError(t, err)
	require.True(t, has)

	engine, ok := ctx.Value(factory.ConextDBType).(*xorm.Engine)
	require.True(t, ok)
	_, err = engine.Exec(`DROP TABLE user_account`)
	require.NoError(t, err)

	balanceBefore := lo.FromPtr(account.Balance)
	versionBefore := account.Version
	updatedBefore := account.Updated

	idx, err := account.WithdrawBalance(ctx, 3000, WITHDRAWAL_INVESTMENT)
	assert.Zero(t, idx)
	require.Error(t, err)
	assert.NotContains(t, err.Error(), "not match version", "SQLレベルのエラーは固定文言に化けず、元のエラー内容が呼び出し元に伝わること")
	assert.EqualValues(t, balanceBefore, lo.FromPtr(account.Balance), "SQLエラー時もaは一切変更されないこと")
	assert.EqualValues(t, versionBefore, account.Version)
	assert.Equal(t, updatedBefore, account.Updated)
}

// MIR-1247: UserAccount.Updateは呼び出し元がまだ無いが、applyBalanceDeltaと
// 同じくversion列を持つUserAccountの公開メソッドであるため、同じ「0行更新
// でも成功扱い」の穴を埋めておく。WithdrawBalance側の
// TestUserAccount_WithdrawBalance_VersionConflictと同じ形で固定する。
func TestUserAccount_Update_VersionConflict(t *testing.T) {
	ctx := setupUserAccountBalanceTest(t, 10000)

	var first, stale UserAccount
	has, err := factory.DB(ctx).ID(int64(1)).Select(UserAccount{}.SelectAll()).Get(&first)
	require.NoError(t, err)
	require.True(t, has)
	has, err = factory.DB(ctx).ID(int64(1)).Select(UserAccount{}.SelectAll()).Get(&stale)
	require.NoError(t, err)
	require.True(t, has)

	first.Balance = lo.ToPtr(int64(7000))
	idx, err := first.Update(ctx)
	require.NoError(t, err)
	require.EqualValues(t, 1, idx)

	stale.Balance = lo.ToPtr(int64(9000))
	idx, err = stale.Update(ctx)
	assert.Zero(t, idx, "version不一致なら0行更新になること")
	require.Error(t, err, "0行更新はエラーとして呼び出し元に伝わること")
	assert.ErrorIs(t, err, ErrVersionConflict)
	assert.EqualValues(t, 9000, lo.FromPtr(stale.Balance), "更新に失敗した場合、呼び出し前にセットした値も含めstaleは変化しないこと")
	assert.EqualValues(t, 1, stale.Version)

	var current UserAccount
	has, err = factory.DB(ctx).ID(int64(1)).Select(UserAccount{}.SelectAll()).Get(&current)
	require.NoError(t, err)
	require.True(t, has)
	assert.EqualValues(t, 7000, lo.FromPtr(current.Balance), "staleの更新は失敗しているので、firstの分だけ反映された状態のままであること")
}
