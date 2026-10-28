package controllers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

func newFailureReasonTestCtx(t *testing.T) (context.Context, *xorm.Engine) {
	t.Helper()
	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	require.NoError(t, err)
	engine.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, engine.Close()) })
	return context.WithValue(context.Background(), factory.ConextDBType, engine), engine
}

func withdrawalDetail(id int64, status domain.WITHDRAWAL_STATUS_TYPE) domain.WithdrawalDetail {
	return domain.WithdrawalDetail{Withdrawal: domain.Withdrawal{Id: id, Status: status}}
}

// MIR-1170: 理由は、DTOの並び(=items)と同じ順の出金依頼に、出金依頼IDで対応づけて付く。
// 別の出金依頼の理由を取り違えて出さないことを確かめる。
func TestApplyWithdrawalFailureReasons_MapsReasonsByWithdrawalId(t *testing.T) {
	ctx, engine := newFailureReasonTestCtx(t)
	for _, ddl := range []string{
		`CREATE TABLE transfer (id INTEGER PRIMARY KEY AUTOINCREMENT, item_id INTEGER NOT NULL, request_id INTEGER NOT NULL, transaction_id TEXT NOT NULL, transfer_status TEXT NULL, error_code TEXT NULL, error_msg TEXT NULL, created DATETIME NULL)`,
		`CREATE TABLE withdrawal_transfer_mapping (id INTEGER PRIMARY KEY AUTOINCREMENT, transaction_id TEXT NOT NULL, withdrawal_id INTEGER NOT NULL)`,
		`CREATE TABLE transfer_request (id INTEGER PRIMARY KEY AUTOINCREMENT, request_status TEXT NULL, error_code TEXT NULL, error_msg TEXT NULL, result_code TEXT NULL)`,
		`INSERT INTO transfer_request (id, request_status, error_code) VALUES (1, 'TRANSFER_REQUEST_FAIL', '220011')`,
		// 出金依頼11が原因(220001)、12は巻き込まれ。13はCOMPLETE。
		`INSERT INTO transfer (item_id, request_id, transaction_id, transfer_status, error_code, error_msg) VALUES (1, 1, 'tx-11', 'TRANSFER_STATUS_FAIL', '220001', '原因')`,
		`INSERT INTO transfer (item_id, request_id, transaction_id, transfer_status, error_code, error_msg) VALUES (2, 1, 'tx-12', 'TRANSFER_STATUS_FAIL', '', '')`,
		`INSERT INTO withdrawal_transfer_mapping (transaction_id, withdrawal_id) VALUES ('tx-11', 11), ('tx-12', 12)`,
	} {
		_, err := engine.Exec(ddl)
		require.NoError(t, err)
	}

	// 並びをあえて入れ替え(13・12・11)、位置ではなく出金依頼IDで対応づくことを確かめる。
	items := []domain.WithdrawalDetail{
		withdrawalDetail(13, domain.WITHDRAWAL_STATUS_COMPLETE),
		withdrawalDetail(12, domain.WITHDRAWAL_STATUS_FAIL),
		withdrawalDetail(11, domain.WITHDRAWAL_STATUS_FAIL),
	}
	dtos := WithdrawalDetailDto{}.NewList(items)

	applyWithdrawalFailureReasons(ctx, items, dtos)

	assert.Equal(t, int64(13), dtos[0].Id)
	assert.Empty(t, dtos[0].FailureKind, "COMPLETEには理由を出さない")
	assert.Empty(t, dtos[0].FailureMessage)

	assert.Equal(t, int64(12), dtos[1].Id)
	assert.Equal(t, "BATCH_CAUSED_BY_OTHER", dtos[1].FailureKind)
	assert.Equal(t, "同じ振込の出金ID 11 のエラーで不成立", dtos[1].FailureMessage)

	assert.Equal(t, int64(11), dtos[2].Id)
	assert.Equal(t, "TRANSFER_ERROR", dtos[2].FailureKind)
	assert.Equal(t, "220001 原因", dtos[2].FailureMessage)
}

// 取得に失敗しても一覧は落とさない。理由を空のまま、DTOをそのまま返す。
func TestApplyWithdrawalFailureReasons_QueryFailureLeavesListIntact(t *testing.T) {
	ctx, _ := newFailureReasonTestCtx(t) // テーブルが無く、クエリは必ず失敗する
	items := []domain.WithdrawalDetail{withdrawalDetail(21, domain.WITHDRAWAL_STATUS_FAIL)}
	dtos := WithdrawalDetailDto{}.NewList(items)

	require.NotPanics(t, func() { applyWithdrawalFailureReasons(ctx, items, dtos) })

	require.Len(t, dtos, 1)
	assert.Equal(t, int64(21), dtos[0].Id)
	assert.Empty(t, dtos[0].FailureKind)
	assert.Empty(t, dtos[0].FailureMessage)
}

// 対象が無ければDBを読まない(ctxにDBが無くてもpanicしない)。
func TestApplyWithdrawalFailureReasons_NoTargetsNoQuery(t *testing.T) {
	items := []domain.WithdrawalDetail{withdrawalDetail(31, domain.WITHDRAWAL_STATUS_COMPLETE)}
	dtos := WithdrawalDetailDto{}.NewList(items)

	require.NotPanics(t, func() { applyWithdrawalFailureReasons(context.Background(), items, dtos) })
	assert.Empty(t, dtos[0].FailureKind)
}
