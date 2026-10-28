package domain

import (
	"context"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

// setupWithdrawalCloseTestDB は db/sql/version/energyshares.base.schema.sql の
// withdrawal_close テーブル定義と揃える(NOWAITはsqliteが構文非対応のため
// TryLockByIdでは付けないが、それ以外の列構成は本番と一致させる)。
func setupWithdrawalCloseTestDB(t *testing.T) context.Context {
	t.Helper()

	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	require.NoError(t, err)
	engine.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, engine.Close()) })

	_, err = engine.Exec(`
		CREATE TABLE withdrawal_close (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			closer_id INTEGER NOT NULL,
			close_date DATETIME NOT NULL,
			transfer_date DATETIME NULL,
			total INTEGER NOT NULL DEFAULT 0,
			fee INTEGER NOT NULL DEFAULT 0,
			amount INTEGER NOT NULL DEFAULT 0,
			status TEXT NOT NULL DEFAULT 'WAIT'
		);
	`)
	require.NoError(t, err)

	session := engine.NewSession()
	require.NoError(t, session.Begin())
	t.Cleanup(func() {
		_ = session.Rollback()
		session.Close()
	})
	return context.WithValue(context.Background(), factory.ConextDBType, session)
}

// insertWithdrawalClose は行を挿入し、採番されたIdを返す(session.Insertの戻り値は
// 影響行数であってIdではないため、挿入後にwc.Idを読む)。
func insertWithdrawalClose(t *testing.T, ctx context.Context, wc WithdrawalClose) int64 {
	t.Helper()
	session := ctx.Value(factory.ConextDBType).(*xorm.Session)
	_, err := session.Insert(&wc)
	require.NoError(t, err)
	return wc.Id
}

func Test_WithdrawalClose_TryLockById(t *testing.T) {
	t.Run("該当行を取得できる", func(t *testing.T) {
		ctx := setupWithdrawalCloseTestDB(t)
		id := insertWithdrawalClose(t, ctx, WithdrawalClose{
			CloserId: 1,
			Status:   WITHDRAWAL_REQUEST_STATUS_WAIT,
		})

		got, err := WithdrawalClose{}.TryLockById(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, WITHDRAWAL_REQUEST_STATUS_WAIT, got.Status)
	})

	t.Run("存在しないIDはErrWithdrawalCloseNotFoundになる", func(t *testing.T) {
		ctx := setupWithdrawalCloseTestDB(t)

		_, err := WithdrawalClose{}.TryLockById(ctx, 999)
		assert.ErrorIs(t, err, ErrWithdrawalCloseNotFound)
	})
}

// MIR-1175の回帰テスト: UpdateTransferDateAndStatusが.Colsでtransfer_date/statusの
// 2列だけを更新し、TryLockByIdが読み込んだ他の列(closer_id/total/amount/fee)を
// ロック取得時点の値で上書きしないことを確認する。
func Test_WithdrawalClose_UpdateTransferDateAndStatus_他の列を上書きしない(t *testing.T) {
	ctx := setupWithdrawalCloseTestDB(t)
	id := insertWithdrawalClose(t, ctx, WithdrawalClose{
		CloserId: 7,
		Total:    3,
		Amount:   30000,
		Fee:      300,
		Status:   WITHDRAWAL_REQUEST_STATUS_WAIT,
	})

	wc, err := WithdrawalClose{}.TryLockById(ctx, id)
	require.NoError(t, err)

	// 締め後に別経路でfeeが更新されていた状況を模す(UpdateFeeById等)。
	// .Colsが効いていなければ、この後のUpdateTransferDateAndStatusでロック取得時点の
	// 古いFee(300)へ黙って巻き戻ってしまう。
	session := ctx.Value(factory.ConextDBType).(*xorm.Session)
	_, err = session.ID(id).Cols("fee").Update(&WithdrawalClose{Fee: 999})
	require.NoError(t, err)

	require.NoError(t, wc.UpdateTransferDateAndStatus(ctx, jsonutils.JSONTime{}))

	var got WithdrawalClose
	has, err := session.ID(id).Get(&got)
	require.NoError(t, err)
	require.True(t, has)
	assert.Equal(t, WITHDRAWAL_REQUEST_STATUS_SENDING, got.Status)
	assert.EqualValues(t, 7, got.CloserId, "closer_idは変わらない")
	assert.EqualValues(t, 3, got.Total, "totalは変わらない")
	assert.EqualValues(t, 30000, got.Amount, "amountは変わらない")
	assert.EqualValues(t, 999, got.Fee, "feeは.Cols対象外なので別経路の更新が保たれる")
}

func Test_WithdrawalClose_CanRequestTransfer(t *testing.T) {
	tests := []struct {
		name   string
		status WITHDRAWAL_REQUEST_STATUS_TYPE
		want   bool
	}{
		{name: "WAITは実行できる", status: WITHDRAWAL_REQUEST_STATUS_WAIT, want: true},
		{name: "SENDINGは実行できない(処理中)", status: WITHDRAWAL_REQUEST_STATUS_SENDING, want: false},
		{name: "COMPLETEは実行できない(処理済み)", status: WITHDRAWAL_REQUEST_STATUS_COMPLETE, want: false},
		{name: "空文字(未知のstatus)は実行できない", status: "", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wc := WithdrawalClose{Status: tt.status}
			assert.Equal(t, tt.want, wc.CanRequestTransfer())
		})
	}
}
