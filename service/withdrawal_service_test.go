package service

import (
	"context"
	"errors"
	"net/http"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

func setupWithdrawalServiceTestDB(t *testing.T) (context.Context, *xorm.Engine) {
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

// TestConvertWithdrawalStatus_ConcurrentCancel_OnlyOneSucceedsはMIR-1250の
// 検証。ユーザーが出金キャンセルを二度押しした状況を、同じ出金レコードを
// 別々に取得した2つのメモリ上のコピー(w1・w2、どちらもINCOMPLETEを読んだ
// 直後)への並行リクエストとして再現する。修正前はcheckStatusValidationが
// メモリ上の値だけで判定するため両方とも成立し、返金(DepositBalance)も
// 二重に走っていた。修正後は1件目のCANCELがコミットした後、2件目は
// UpdateStatusIfCurrentの条件付きUPDATEが0行に終わり
// ErrorWithdrawalStatusChangedを返すため、後続の返金処理(呼び出し元の
// UpdateStatusハンドラ)へ進めない。
func TestConvertWithdrawalStatus_ConcurrentCancel_OnlyOneSucceeds(t *testing.T) {
	ctx, engine := setupWithdrawalServiceTestDB(t)

	w := domain.Withdrawal{InvestorId: "1", InvestorName: "test", Amount: 1000, Status: domain.WITHDRAWAL_STATUS_INCOMPLETE}
	_, err := engine.Insert(&w)
	require.NoError(t, err)

	// 2つの並行リクエストが、どちらもCANCEL前の同じ内容を読んだと仮定する。
	w1 := w
	w2 := w

	s := NewWithdrawalService()

	_, err1 := s.ConvertWithdrawalStatus(ctx, &w1, domain.WITHDRAWAL_STATUS_CANCEL)
	require.NoError(t, err1, "1件目は正常にCANCELへ進めるはず")

	_, err2 := s.ConvertWithdrawalStatus(ctx, &w2, domain.WITHDRAWAL_STATUS_CANCEL)
	require.Error(t, err2, "2件目は既にCANCEL済みのため弾かれるはず(修正前はここも成功していた)")

	var factoryErr *factory.Error
	require.True(t, errors.As(err2, &factoryErr), "factory.ErrorWithdrawalStatusChanged.Newで包まれたエラーのはず")
	assert.Equal(t, factory.ErrorWithdrawalStatusChanged.Code, factoryErr.Code)
	assert.Equal(t, http.StatusConflict, factoryErr.Status(), "利用者から見える変化は500から409になること")

	// DB上のstatusは1件目のCANCELのままで、2件目による書き換えは起きていない。
	var status string
	okRow, err := engine.Table("withdrawal").Where("id = ?", w.Id).Cols("status").Get(&status)
	require.NoError(t, err)
	require.True(t, okRow)
	assert.Equal(t, "CANCEL", status)
}

// TestConvertWithdrawalStatus_WaitIsRejected はMIR-1323の検証。出金を個別に
// 締める経路(UpdateStatusのWAIT分岐)は、MIR-1320で見つかった不具合(締めの
// 紐付けが常に1になる)を抱えたまま本番で一度も使われていなかったため、機能
// ごと削除した。WAITへの遷移をcheckStatusValidationが拒否することを確認する。
func TestConvertWithdrawalStatus_WaitIsRejected(t *testing.T) {
	ctx, engine := setupWithdrawalServiceTestDB(t)

	w := domain.Withdrawal{InvestorId: "1", InvestorName: "test", Amount: 1000, Status: domain.WITHDRAWAL_STATUS_INCOMPLETE}
	_, err := engine.Insert(&w)
	require.NoError(t, err)

	s := NewWithdrawalService()
	_, err = s.ConvertWithdrawalStatus(ctx, &w, domain.WITHDRAWAL_STATUS_WAIT)
	require.Error(t, err)

	var factoryErr *factory.Error
	require.True(t, errors.As(err, &factoryErr), "factory.ErrorParameter.Newで包まれたエラーのはず")
	assert.Equal(t, factory.ErrorParameter.Code, factoryErr.Code)

	// DB上のstatusは変わらずINCOMPLETEのまま。
	var status string
	okRow, err := engine.Table("withdrawal").Where("id = ?", w.Id).Cols("status").Get(&status)
	require.NoError(t, err)
	require.True(t, okRow)
	assert.Equal(t, "INCOMPLETE", status)
}

// TestReserveWithdrawalForTransfer_RetryAfterSuccess_FailsはMIR-1275の検証。
// 管理画面の個別再送金ボタンを二度押しした状況を再現する。ReserveWithdrawalForTransferは
// withdrawal_idだけを受け取り、可否判定(現在のstatusがFAILか)をDB側の条件付きUPDATEで
// 行うため、呼び出し元がメモリ上の古いstatusを保持しているかどうかに関係なく、同じ
// withdrawal_idへの呼び出しを2回連続で行うだけで並行リクエストと同じ状況になる
// (1回目がコミットした後、2回目はDB上の実際のstatusを見て弾かれる)。修正前の
// RequestTransferByWithdrawalIdは、事前に読んだメモリ上のstatus(FAIL)だけで判定して
// unconditionalなUpdateを呼んでいたため、この2回目相当の呼び出しも成功しており、
// Transfer/WithdrawalTransferMappingが二重に作られていた。
func TestReserveWithdrawalForTransfer_RetryAfterSuccess_Fails(t *testing.T) {
	ctx, engine := setupWithdrawalServiceTestDB(t)

	w := domain.Withdrawal{InvestorId: "1", InvestorName: "test", Amount: 1000, Status: domain.WITHDRAWAL_STATUS_FAIL}
	_, err := engine.Insert(&w)
	require.NoError(t, err)

	s := NewWithdrawalService()

	err1 := s.ReserveWithdrawalForTransfer(ctx, w.Id)
	require.NoError(t, err1, "1件目は正常にFAIL→WAITへ進めるはず")

	err2 := s.ReserveWithdrawalForTransfer(ctx, w.Id)
	require.Error(t, err2, "2件目は既にWAIT済みのため弾かれるはず(修正前はここも成功し、送金が二重に作られていた)")

	var factoryErr *factory.Error
	require.True(t, errors.As(err2, &factoryErr), "factory.ErrorWithdrawalStatusChanged.Newで包まれたエラーのはず")
	assert.Equal(t, factory.ErrorWithdrawalStatusChanged.Code, factoryErr.Code)
	assert.Equal(t, http.StatusConflict, factoryErr.Status())

	// DB上のstatusは1件目のWAITのままで、2件目による書き換えは起きていない。
	var status string
	okRow, err := engine.Table("withdrawal").Where("id = ?", w.Id).Cols("status").Get(&status)
	require.NoError(t, err)
	require.True(t, okRow)
	assert.Equal(t, "WAIT", status)
}
