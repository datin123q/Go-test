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

// setupTransferTestDB はGetLatestByWithdrawalIdsが読む2テーブルだけを生SQLで作る。
// engine.Sync()にしないのは、Transfer構造体からスキーマを起こすと、将来transferに
// withdrawal_idという実カラムが増えたときにテストのスキーマも追随して変わってしまい、
// このバグの再発(CLAUDE.mdのReview Checklist参照)を検知できなくなるため。本番スキーマ
// (db/sql/version/energyshares.base.schema.sql)に無いカラムはここにも含めない。
func setupTransferTestDB(t *testing.T) context.Context {
	t.Helper()

	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	require.NoError(t, err)
	engine.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, engine.Close()) })

	// db/sql/version/energyshares.base.schema.sql の transfer テーブル定義と揃える
	// (withdrawal_id列は本番にも無いため含めない)。session.Insert(&Transfer{})はTransfer
	// 構造体の全フィールド分のカラムへ書き込むため、SELECTで使う列だけに絞ると失敗する。
	_, err = engine.Exec(`
		CREATE TABLE transfer (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			item_id INTEGER NOT NULL,
			request_id INTEGER NOT NULL,
			transaction_id TEXT NOT NULL,
			user_id INTEGER NULL,
			project_id INTEGER NULL,
			transfer_date DATETIME NULL,
			transfer_amount INTEGER NULL,
			account_number TEXT NULL,
			bank_code TEXT NULL,
			bank_name TEXT NULL,
			branch_code TEXT NULL,
			branch_name TEXT NULL,
			depositor TEXT NULL,
			account_type_code TEXT NULL,
			transfer_fee INTEGER NULL,
			transaction_detail_type TEXT NULL,
			created DATETIME NULL,
			updated DATETIME NULL,
			transfer_status TEXT NULL,
			repayment_date DATETIME NULL,
			error_code TEXT NULL,
			error_msg TEXT NULL
		);
	`)
	require.NoError(t, err)

	_, err = engine.Exec(`
		CREATE TABLE withdrawal_transfer_mapping (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			transaction_id TEXT NOT NULL,
			withdrawal_id INTEGER NOT NULL
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

func insertTransfer(t *testing.T, ctx context.Context, tr Transfer) {
	t.Helper()
	session := ctx.Value(factory.ConextDBType).(*xorm.Session)
	_, err := session.Insert(&tr)
	require.NoError(t, err)
}

func insertMapping(t *testing.T, ctx context.Context, m WithdrawalTransferMapping) {
	t.Helper()
	session := ctx.Value(factory.ConextDBType).(*xorm.Session)
	_, err := session.Insert(&m)
	require.NoError(t, err)
}

// MIR-1180: GetLatestByWithdrawalIdsが返すマップのキーがtr.RequestId(送金バッチのID)に
// なっていて、唯一の呼び出し元であるadmin-apiのGetAll(apps/admin-api/controllers/
// withdrawal.go)がWithdrawal.Id(出金依頼のID)で引いていたため機能していなかった不具合の
// 修正確認。GetOneはこの関数を呼ばず、GetByWithdrawalIdという別の(キーが正しい)関数を
// 使っているため対象外。アサートするフィールドは、呼び出し元(withdrawal.go)が実際に読む
// 6つ全て(TransferFee/BankName/AccountNumber/BankCode/BranchName/BranchCode)にそろえて
// いる。SELECT文の手書き列挙リストから将来これらのいずれかが抜け落ちても、このテストで
// 検知するため(code-reviewで、BankCode等をアサートしていないと抜けを検知できないと指摘された)。
func TestTransfer_GetLatestByWithdrawalIds_KeyedByWithdrawalId(t *testing.T) {
	ctx := setupTransferTestDB(t)

	// 出金依頼ID=500の投資家が、transfer_request.id=999のバッチで送金されたシナリオ。
	const (
		withdrawalId = int64(500)
		requestId    = int64(999)
	)

	insertTransfer(t, ctx, Transfer{
		ItemId:        1,
		RequestId:     requestId,
		TransactionId: "tx-500",
		BankName:      "テスト銀行",
		BankCode:      "0001",
		BranchName:    "テスト支店",
		BranchCode:    "001",
		AccountNumber: "1234567",
		TransferFee:   143,
	})
	insertMapping(t, ctx, WithdrawalTransferMapping{WithdrawalId: withdrawalId, TransactionId: "tx-500"})

	latestMap, err := Transfer{}.GetLatestByWithdrawalIds(ctx, []int64{withdrawalId})
	require.NoError(t, err)

	latest, ok := latestMap[withdrawalId]
	assert.True(t, ok, "呼び出し側と同じWithdrawal.Idで引けること")
	assert.Equal(t, "テスト銀行", latest.BankName)
	assert.Equal(t, "0001", latest.BankCode)
	assert.Equal(t, "テスト支店", latest.BranchName)
	assert.Equal(t, "001", latest.BranchCode)
	assert.Equal(t, "1234567", latest.AccountNumber)
	assert.Equal(t, int64(143), latest.TransferFee)
}

// withdrawal_transfer_mappingに1件も紐づかないwithdrawal_id(INNER JOINで除外される)は、
// マップにキー自体が存在しない。呼び出し元はこれを`ok`で判定し、既存のスナップショット値を
// そのまま保持する(上書きしない)設計のため、キーが無いことそのものを確認する。
func TestTransfer_GetLatestByWithdrawalIds_NoMappingMeansKeyAbsent(t *testing.T) {
	ctx := setupTransferTestDB(t)

	latestMap, err := Transfer{}.GetLatestByWithdrawalIds(ctx, []int64{999999})
	require.NoError(t, err)

	_, ok := latestMap[999999]
	assert.False(t, ok, "紐づく送金が無いwithdrawal_idはキーごと存在しないこと")
}

// 1つの出金依頼に複数の送金(再送金)が紐づく場合、transfer.id最大(最新)の1件だけが残ること。
func TestTransfer_GetLatestByWithdrawalIds_KeepsLatestOnRetry(t *testing.T) {
	ctx := setupTransferTestDB(t)

	const withdrawalId = int64(600)

	// 1回目の送金(失敗)と2回目の送金(再送金・成功)。IDの若い方が古い試行。
	insertTransfer(t, ctx, Transfer{ItemId: 1, RequestId: 111, TransactionId: "tx-old", BankName: "旧銀行"})
	insertTransfer(t, ctx, Transfer{ItemId: 1, RequestId: 222, TransactionId: "tx-new", BankName: "新銀行"})
	insertMapping(t, ctx, WithdrawalTransferMapping{WithdrawalId: withdrawalId, TransactionId: "tx-old"})
	insertMapping(t, ctx, WithdrawalTransferMapping{WithdrawalId: withdrawalId, TransactionId: "tx-new"})

	latestMap, err := Transfer{}.GetLatestByWithdrawalIds(ctx, []int64{withdrawalId})
	require.NoError(t, err)

	got, ok := latestMap[withdrawalId]
	require.True(t, ok)
	assert.Equal(t, "新銀行", got.BankName, "transfer.id最大(最新の送金試行)が残ること")
}

// MIR-1180: 修正前の実装はマップをRequestId(送金バッチのID)でキーしていたため、
// 「常に空振りするだけ」ではなく、withdrawal.idとtransfer_request.idの値がたまたま
// 一致する組み合わせがあると、別の投資家の口座情報が紛れ込む実害があった(①敵対的検証で
// 発見)。出金依頼A(id=7)がバッチ20で送金され、出金依頼B(id=20)がバッチ7で送金された
// というID空間の交差を再現し、修正後は各withdrawal_idに正しい投資家の情報が対応する
// ことを確認する。複数の出金依頼をまとめて引く経路もこのテストで兼ねる。
func TestTransfer_GetLatestByWithdrawalIds_DoesNotMixUpDifferentInvestors(t *testing.T) {
	ctx := setupTransferTestDB(t)

	const (
		withdrawalIdA = int64(7)
		requestIdA    = int64(20)
		withdrawalIdB = int64(20)
		requestIdB    = int64(7)
	)

	insertTransfer(t, ctx, Transfer{ItemId: 1, RequestId: requestIdA, TransactionId: "tx-a", BankName: "Aさんの銀行", AccountNumber: "1111111"})
	insertTransfer(t, ctx, Transfer{ItemId: 1, RequestId: requestIdB, TransactionId: "tx-b", BankName: "Bさんの銀行", AccountNumber: "2222222"})
	insertMapping(t, ctx, WithdrawalTransferMapping{WithdrawalId: withdrawalIdA, TransactionId: "tx-a"})
	insertMapping(t, ctx, WithdrawalTransferMapping{WithdrawalId: withdrawalIdB, TransactionId: "tx-b"})

	latestMap, err := Transfer{}.GetLatestByWithdrawalIds(ctx, []int64{withdrawalIdA, withdrawalIdB})
	require.NoError(t, err)
	require.Len(t, latestMap, 2)

	gotA, ok := latestMap[withdrawalIdA]
	require.True(t, ok)
	assert.Equal(t, "Aさんの銀行", gotA.BankName)
	assert.Equal(t, "1111111", gotA.AccountNumber)

	gotB, ok := latestMap[withdrawalIdB]
	require.True(t, ok)
	assert.Equal(t, "Bさんの銀行", gotB.BankName)
	assert.Equal(t, "2222222", gotB.AccountNumber)
}

// withdrawalIdsが空(nilを含む)の場合は空マップを返す。
func TestTransfer_GetLatestByWithdrawalIds_EmptyInput(t *testing.T) {
	ctx := setupTransferTestDB(t)

	latestMap, err := Transfer{}.GetLatestByWithdrawalIds(ctx, []int64{})
	require.NoError(t, err)
	assert.Empty(t, latestMap)

	latestMap, err = Transfer{}.GetLatestByWithdrawalIds(ctx, nil)
	require.NoError(t, err)
	assert.Empty(t, latestMap)
}
