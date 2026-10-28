package domain

import (
	"context"
	"errors"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

// setupWithdrawalUpdateStatusTestDBはUpdateStatusIfCurrentの検証用に、
// withdrawalテーブルだけを持つ最小限のsqliteを用意する。
func setupWithdrawalUpdateStatusTestDB(t *testing.T) (context.Context, *xorm.Engine) {
	t.Helper()

	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	require.NoError(t, err)
	engine.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, engine.Close()) })

	_, err = engine.Exec(`
		CREATE TABLE withdrawal (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			investor_id TEXT NOT NULL,
			orderer TEXT NOT NULL,
			amount INTEGER NOT NULL DEFAULT 0,
			fee INTEGER NOT NULL DEFAULT 0,
			request_date DATETIME,
			complete_date DATETIME,
			close_date DATETIME,
			status TEXT NOT NULL,
			is_manual BOOLEAN NOT NULL DEFAULT 0,
			created DATETIME,
			updated DATETIME,
			deleted DATETIME
		);
	`)
	require.NoError(t, err)

	return context.WithValue(context.Background(), factory.ConextDBType, engine), engine
}

// MIR-1250: UpdateStatusIfCurrentは、想定した現在のステータス(expectedCurrent)
// とDB上の実際の値が一致する場合のみ更新する。一致すれば通常のUpdateと同じ
// 結果になることを確認する。
func TestWithdrawal_UpdateStatusIfCurrent_MatchesCurrentStatus(t *testing.T) {
	ctx, engine := setupWithdrawalUpdateStatusTestDB(t)

	w := Withdrawal{InvestorId: "1", InvestorName: "test", Amount: 1000, Status: WITHDRAWAL_STATUS_INCOMPLETE}
	_, err := engine.Insert(&w)
	require.NoError(t, err)

	w.Status = WITHDRAWAL_STATUS_CANCEL
	err = w.UpdateStatusIfCurrent(ctx, WITHDRAWAL_STATUS_INCOMPLETE)
	require.NoError(t, err)

	var status string
	ok, err := engine.Table("withdrawal").Where("id = ?", w.Id).Cols("status").Get(&status)
	require.NoError(t, err)
	require.True(t, ok)
	assert.Equal(t, "CANCEL", status)
}

// MIR-1250: DB上のステータスが想定(expectedCurrent)と既に異なる場合は、
// 0行更新としてErrWithdrawalStatusChangedを返し、実際には何も書き換えない
// ことを確認する。これが、二重キャンセルの2件目が弾かれる仕組みそのもの。
//
// 並行リクエストの2件目を模して、DB上は既にCANCEL済みの行に対し、
// (1件目がまだCANCELへ進める前に読んでいた)INCOMPLETEを期待値として
// 呼び出す。呼び出し側のwのAmountを実際のDB値(1000)とは異なる値に
// しておき、更新が一切効かなかったこと(DB側が呼び出し側の値で上書き
// されていないこと)をAmountの不一致で確認する。
func TestWithdrawal_UpdateStatusIfCurrent_ReturnsErrorWhenStatusAlreadyChanged(t *testing.T) {
	ctx, engine := setupWithdrawalUpdateStatusTestDB(t)

	w := Withdrawal{InvestorId: "1", InvestorName: "test", Amount: 1000, Status: WITHDRAWAL_STATUS_CANCEL}
	_, err := engine.Insert(&w)
	require.NoError(t, err)

	// 呼び出し側の手元のコピーは、DBの実際の値(Amount: 1000)とは異なる
	// 値を持たせておく。更新が誤って通ってしまった場合はこの値でDBが
	// 上書きされるはずなので、0行更新であることをAmountの不一致で確認できる。
	w.Amount = 9999
	err = w.UpdateStatusIfCurrent(ctx, WITHDRAWAL_STATUS_INCOMPLETE)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrWithdrawalStatusChanged))

	var amount int64
	ok, err := engine.Table("withdrawal").Where("id = ?", w.Id).Cols("amount").Get(&amount)
	require.NoError(t, err)
	require.True(t, ok)
	assert.Equal(t, int64(1000), amount, "0行更新のはずなので、呼び出し側が持っていた別の値(9999)でDBが上書きされていないこと")
}
