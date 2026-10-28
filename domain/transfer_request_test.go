package domain

import (
	"context"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

// setupNewTransferRequestTest はNewTransferRequestのチャンク分割(MIR-1181)を
// 検証するためのsqliteを用意する。engine.Sync2で構造体からテーブルを生成する
// ため、CLAUDE.mdが警告する「Sync2が本番に無い列まで作ってしまう」問題は
// 本来ならFind/Getにも掛かる。ただしこのテストが検証したいのはチャンク分割の
// 重複挿入・件数・ItemId採番であり、Transfer/TransferRequestには現時点で
// xorm:"<-"(JOIN専用)フィールドが無いため実害はない。将来そうしたフィールドが
// 増えたら、この前提も合わせて見直すこと。
func setupNewTransferRequestTest(t *testing.T) context.Context {
	t.Helper()

	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, engine.Close())
	})

	require.NoError(t, engine.Sync2(new(Transfer), new(TransferRequest)))

	return context.WithValue(context.Background(), factory.ConextDBType, engine)
}

func makeTransfers(count int) []Transfer {
	transfers := make([]Transfer, count)
	for i := range transfers {
		transfers[i] = Transfer{
			TransactionId:  "tx",
			TransferAmount: 1000,
			TransferStatus: TRANSFER_STATUS_WAIT,
		}
	}
	return transfers
}

func Test_NewTransferRequest_ChunkBoundaries(t *testing.T) {
	// チャンクはTRANSFER_LIMIT(98)件ちょうどで区切る。GMOの総合振込1リクエスト
	// あたりの明細数上限は9,999件で確認済みのため98はGMO側の制約ではないが、
	// 定数名と実際のチャンクサイズを一致させている(MIR-1181)。98件は単一
	// チャンクになる境目、99件は2チャンク目ができる境目、196件はちょうど
	// 98件×2に収まる境目、197件は3チャンク目ができる境目にあたる。
	tests := []struct {
		name          string
		count         int
		wantRequests  int
		wantTransfers int
	}{
		{name: "98件(TRANSFER_LIMITちょうど)は1件のTransferRequestになる", count: 98, wantRequests: 1, wantTransfers: 98},
		{name: "99件は2件のTransferRequestに分かれる(98件+1件)", count: 99, wantRequests: 2, wantTransfers: 99},
		{name: "100件は2件のTransferRequestに分かれ、重複しない", count: 100, wantRequests: 2, wantTransfers: 100},
		{name: "196件はちょうど2件のTransferRequest(98件ずつ)に収まる", count: 196, wantRequests: 2, wantTransfers: 196},
		{name: "197件は3件のTransferRequestに分かれる(98件+98件+1件)", count: 197, wantRequests: 3, wantTransfers: 197},
		{name: "198件は3件のTransferRequestに分かれ、重複しない", count: 198, wantRequests: 3, wantTransfers: 198},
		{name: "199件は3件のTransferRequestに分かれ、重複しない", count: 199, wantRequests: 3, wantTransfers: 199},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := setupNewTransferRequestTest(t)
			transfers := makeTransfers(tt.count)

			err := (TransferRequest{}).NewTransferRequest(ctx, transfers)
			require.NoError(t, err)

			var transferCount int64
			transferCount, err = factory.DB(ctx).Count(&Transfer{})
			require.NoError(t, err)
			require.Equal(t, int64(tt.wantTransfers), transferCount,
				"重複挿入がある場合、この件数はwantTransfersの倍数になる")

			var requestCount int64
			requestCount, err = factory.DB(ctx).Count(&TransferRequest{})
			require.NoError(t, err)
			require.Equal(t, int64(tt.wantRequests), requestCount)

			// 各TransferRequestに紐づくTransfer件数がrequest_total_cntと一致すること
			var requests []TransferRequest
			require.NoError(t, factory.DB(ctx).Find(&requests))
			totalFromRequests := 0
			for _, r := range requests {
				var cnt int64
				cnt, err = factory.DB(ctx).Where("request_id = ?", r.Id).Count(&Transfer{})
				require.NoError(t, err)
				require.Equal(t, int64(r.RequestTotalCnt), cnt,
					"request_total_cntは実際に紐づくTransfer件数と一致するべき")
				totalFromRequests += r.RequestTotalCnt
			}
			require.Equal(t, tt.wantTransfers, totalFromRequests)
		})
	}
}

func Test_NewTransferRequest_ItemIdIsSequentialWithinEachChunk(t *testing.T) {
	// 重複挿入時代は、後半チャンクの分だけまだItemIdが採番されておらず0で
	// 入っていた。各チャンク内で1始まりの連番になっていることを確認する。
	ctx := setupNewTransferRequestTest(t)
	transfers := makeTransfers(150)

	err := (TransferRequest{}).NewTransferRequest(ctx, transfers)
	require.NoError(t, err)

	var requests []TransferRequest
	require.NoError(t, factory.DB(ctx).OrderBy("id").Find(&requests))
	require.Len(t, requests, 2)

	for _, r := range requests {
		var chunkTransfers []Transfer
		require.NoError(t, factory.DB(ctx).Where("request_id = ?", r.Id).OrderBy("id").Find(&chunkTransfers))
		for i, tr := range chunkTransfers {
			require.Equal(t, i+1, tr.ItemId, "チャンク内のItemIdは1始まりの連番であるべき")
		}
	}
}

func Test_NewTransferRequest_RollsBackWithOuterTransaction(t *testing.T) {
	// 修正前はTransfer.InsertManyが自前でsession.Commit()/Close()していたため、
	// factory.NewTxManager().WithTx()の中でNewTransferRequestを呼んだ時点で
	// 外側のトランザクションが早期コミットされ、後続の失敗でRollbackしても
	// 既に挿入されたTransfer/TransferRequestは残ってしまっていた(MIR-1181)。
	// この試験は、WithTxの外側から見てロールバックが機能することを確認する。
	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, engine.Close()) })
	require.NoError(t, engine.Sync2(new(Transfer), new(TransferRequest)))

	txErr := factory.NewTxManager(engine).WithTx(context.Background(), func(txCtx context.Context) error {
		if err := (TransferRequest{}).NewTransferRequest(txCtx, makeTransfers(3)); err != nil {
			return err
		}
		// NewTransferRequestの後、あえて失敗させてロールバックさせる。
		return errors.New("forced failure after NewTransferRequest")
	})
	require.Error(t, txErr)

	// トランザクション外(素のengine)から見て、挿入した行が残っていないこと。
	transferCount, err := engine.Count(&Transfer{})
	require.NoError(t, err)
	require.Equal(t, int64(0), transferCount, "NewTransferRequestの後の失敗でロールバックされるべき")

	requestCount, err := engine.Count(&TransferRequest{})
	require.NoError(t, err)
	require.Equal(t, int64(0), requestCount, "NewTransferRequestの後の失敗でロールバックされるべき")
}

// TestTransferRequest_UpdateTransferTotalFee_ZeroIsWrittenは、手数料合計が本当に0円に
// なったとき(MIR-1123、受取口座が全員GMOあおぞら宛て)、通常のUpdate()ではXORMがゼロ値
// のフィールドをUPDATE文から除いてしまい古い合計が残るのに対し、UpdateTransferTotalFee
// はMustColsでこの1カラムだけ強制的に書き込むことを確認する。
func TestTransferRequest_UpdateTransferTotalFee_ZeroIsWritten(t *testing.T) {
	ctx := setupNewTransferRequestTest(t)

	req := TransferRequest{TransferTotalFee: 286}
	id, err := factory.DB(ctx).Insert(&req)
	require.NoError(t, err)
	require.Equal(t, int64(1), id)

	// ApplyNoも同時に非ゼロ値へ変える。MustColsは「この1カラムだけ強制する」
	// 仕組みで、Cols(全体を絞り込む)とは違い他の非ゼロ値カラムも通常どおり
	// 自動検出される。ApplyNoが更新されることを確かめ、実装が誤ってColsに
	// 書き換えられても(その場合ApplyNoが書き込まれなくなる)検出できるようにする。
	req.TransferTotalFee = 0
	req.ApplyNo = "A123"
	require.NoError(t, req.UpdateTransferTotalFee(ctx))

	var got TransferRequest
	has, err := factory.DB(ctx).ID(req.Id).Get(&got)
	require.NoError(t, err)
	require.True(t, has)
	require.Equal(t, int64(0), got.TransferTotalFee, "UpdateTransferTotalFeeは0円を実際にDBへ書き込むべき")
	require.Equal(t, "A123", got.ApplyNo, "MustColsは他の非ゼロ値カラムの自動検出を妨げないべき")
}

// 対比: 通常のUpdate()だと、ゼロ値(TransferTotalFee=0)はXORMの既定で
// UPDATE対象から外れ、古い値(286)が残ることを確認する(退行の再発防止)。
func TestTransferRequest_Update_DoesNotWriteZeroTransferTotalFee(t *testing.T) {
	ctx := setupNewTransferRequestTest(t)

	req := TransferRequest{TransferTotalFee: 286}
	id, err := factory.DB(ctx).Insert(&req)
	require.NoError(t, err)
	require.Equal(t, int64(1), id)

	req.TransferTotalFee = 0
	require.NoError(t, req.Update(ctx))

	var got TransferRequest
	has, err := factory.DB(ctx).ID(req.Id).Get(&got)
	require.NoError(t, err)
	require.True(t, has)
	require.Equal(t, int64(286), got.TransferTotalFee, "通常のUpdate()はゼロ値を書き込まないため古い値が残る(xormの既定動作)")
}

func Test_Transfer_InsertMany_WorksOnPlainEngine(t *testing.T) {
	// 修正前はctxがEngine(Sessionでない)のとき、何も挿入せずエラーも返さず
	// 0件を返していた。修正後はEngineでも正しく挿入されることを確認する
	// (MIR-1181)。
	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, engine.Close()) })
	require.NoError(t, engine.Sync2(new(Transfer)))

	ctx := context.WithValue(context.Background(), factory.ConextDBType, engine)

	count, err := (Transfer{}).InsertMany(ctx, makeTransfers(3))
	require.NoError(t, err)
	require.Equal(t, int64(3), count)

	var actual int64
	actual, err = factory.DB(ctx).Count(&Transfer{})
	require.NoError(t, err)
	require.Equal(t, int64(3), actual)
}
