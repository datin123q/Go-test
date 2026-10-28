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

// setupBalanceMatchTest はbalanceMatchの検証用に、user_account・share・
// deposit_requestだけを持つ最小限のsqliteを用意する。balanceMatchが失敗する
// ケース(WithdrawBalanceの残高不足)では、share・deposit_requestには一切
// 触れないため、Share.UpdateInvestStatusが要求する他のテーブル
// (project・user・admin_user・contract等)は不要。
func setupBalanceMatchTest(t *testing.T, accountBalance int64) context.Context {
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
    project_id INTEGER,
    invest_status TEXT,
    deleted DATETIME
);
CREATE TABLE deposit_request (
    id INTEGER PRIMARY KEY,
    user_id INTEGER,
    share_id TEXT,
    project_id INTEGER,
    request_amount INTEGER,
    deposit_amount INTEGER,
    deposit_check_status TEXT,
    deleted DATETIME
);

INSERT INTO user (id, name, user_type) VALUES (1, 'test user', 'PRIVATE');
INSERT INTO user_account (id, user_id, balance, version) VALUES (1, 1, ?, 1);
INSERT INTO share (id, owner_id, project_id, invest_status) VALUES ('share-1', 1, 1, 'REQUEST_INVEST');
INSERT INTO deposit_request (id, user_id, share_id, project_id, request_amount, deposit_amount, deposit_check_status)
    VALUES (1, 1, 'share-1', 1, 1000, 0, 'NO_DEPOSIT');
`, accountBalance)
	require.NoError(t, err)

	return context.WithValue(context.Background(), factory.ConextDBType, engine)
}

// MIR-1251: balanceMatchはWithdrawBalanceを先に行うようにした。WithdrawBalance
// が失敗した場合(このテストでは残高不足で即座にエラーを返すケースを使う。
// 実運用ではxormの楽観ロック(version列)の競合で同様に失敗しうる)、
// share・deposit_requestのどちらも一切変更されないことを検証する。
//
// 修正前の実装(UpdateInvestStatusを先に呼ぶ順序)だと、WithdrawBalanceが
// 失敗してもshareはSUCCESS_INVESTに、deposit_requestはOK_DEPOSITに
// 変わってしまい、しかも変わった時点でdeposit_check_status=NO_DEPOSITの
// 検索条件に二度と一致しなくなるため再試行もされない。
//
// このテストのsqliteフィクスチャにはadmin_user・project・contract等、
// UpdateInvestStatusのSUCCESS_INVEST分岐が完走するのに必要なテーブルを
// 意図的に用意していない。もしUpdateInvestStatusが(修正前のように)
// WithdrawBalanceより先に呼ばれれば、"lack of balance"ではなく
// "fail to get super admin id"(admin_userテーブルが無いため)のような
// 別のエラーになるはずなので、エラー内容がWithdrawBalance由来であることの
// 確認自体が「WithdrawBalanceが先に呼ばれている」ことの検証を兼ねる。
func TestUserAccount_BalanceMatch_WithdrawBalanceFailure_LeavesShareAndDepositRequestUnchanged(t *testing.T) {
	// 残高(500)より必要金額(1000)の方が大きいので、WithdrawBalanceは
	// "lack of balance" で即座に失敗する。
	ctx := setupBalanceMatchTest(t, 500)

	account, err := UserAccount{}.GetByUserId(ctx, 1)
	require.NoError(t, err)
	require.NotNil(t, account)

	r, err := DepositRequest{}.GetDepositRequestByShareId(ctx, "share-1")
	require.NoError(t, err)
	require.NotNil(t, r)

	err = account.balanceMatch(ctx, 1000, r, NOT_IMMEDIATE)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "lack of balance", "WithdrawBalanceが先に呼ばれ、そこで失敗している必要がある")

	var investStatus string
	ok, err := factory.DB(ctx).Table("share").Where("id = ?", "share-1").Cols("invest_status").Get(&investStatus)
	require.NoError(t, err)
	require.True(t, ok)
	assert.Equal(t, "REQUEST_INVEST", investStatus, "WithdrawBalance失敗時はshareの状態が変わってはいけない")

	var depositCheckStatus string
	ok, err = factory.DB(ctx).Table("deposit_request").Where("id = ?", 1).Cols("deposit_check_status").Get(&depositCheckStatus)
	require.NoError(t, err)
	require.True(t, ok)
	assert.Equal(t, "NO_DEPOSIT", depositCheckStatus, "WithdrawBalance失敗時はdeposit_requestの状態が変わってはいけない(次回また拾われる必要がある)")

	var balance int64
	ok, err = factory.DB(ctx).Table("user_account").Where("id = ?", 1).Cols("balance").Get(&balance)
	require.NoError(t, err)
	require.True(t, ok)
	assert.Equal(t, int64(500), balance, "残高も変わっていないこと")
}

// TestUserAccount_DepositCheckProcess_RollsBackWithdrawalWhenLaterStepFails
// は、DepositCheckProcessがトランザクション境界を持つこと自体を検証する。
// 前のテスト(WithdrawBalanceFailure)はWithdrawBalance自体が失敗する
// ケースだったため、DBへの書き込みが一度も起きておらず、ロールバックの
// 検証にはなっていなかった(WithTxを外してもテストが通ってしまう)。
//
// このテストでは残高を十分に用意してWithdrawBalanceを実際に成功させ
// (残高が実際に減る)、その後のShare.UpdateInvestStatusをadmin_user
// テーブルが無い状態にして失敗させる。DepositCheckProcessがトランザクション
// で囲われていれば、WithdrawBalanceの書き込みもロールバックされ、残高は
// 元の値に戻るはずである。
func TestUserAccount_DepositCheckProcess_RollsBackWithdrawalWhenLaterStepFails(t *testing.T) {
	ctx := setupBalanceMatchTest(t, 1000)

	account, err := UserAccount{}.GetByUserId(ctx, 1)
	require.NoError(t, err)

	r, err := DepositRequest{}.GetDepositRequestByShareId(ctx, "share-1")
	require.NoError(t, err)

	// request_amount=1000, deposit_amount=0 なので NeedAmount=1000、残高も
	// ちょうど1000あるためWithdrawBalance自体は成功する。その後の
	// Share.UpdateInvestStatusがadmin_userテーブル不在で失敗するはず。
	err = account.DepositCheckProcess(ctx, r, NOT_IMMEDIATE)
	require.Error(t, err, "admin_userテーブルが無いためUpdateInvestStatusの途中で失敗するはず")
	assert.NotContains(t, err.Error(), "lack of balance", "今回はWithdrawBalance自体は成功しているはず")

	var balance int64
	ok, err := factory.DB(ctx).Table("user_account").Where("id = ?", 1).Cols("balance").Get(&balance)
	require.NoError(t, err)
	require.True(t, ok)
	assert.Equal(t, int64(1000), balance, "WithdrawBalanceは成功したが、後続の失敗でロールバックされ残高は元に戻っているべき")

	var version int64
	ok, err = factory.DB(ctx).Table("user_account").Where("id = ?", 1).Cols("version").Get(&version)
	require.NoError(t, err)
	require.True(t, ok)
	assert.Equal(t, int64(1), version, "versionもロールバックされ元の値のままのはず")

	var depositCheckStatus string
	ok, err = factory.DB(ctx).Table("deposit_request").Where("id = ?", 1).Cols("deposit_check_status").Get(&depositCheckStatus)
	require.NoError(t, err)
	require.True(t, ok)
	assert.Equal(t, "NO_DEPOSIT", depositCheckStatus, "deposit_requestもロールバックされ、次回また拾われる状態のままのはず")
}
