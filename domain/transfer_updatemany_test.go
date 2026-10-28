package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MIR-1183の回帰テスト。同じctx(同じ*xorm.Session)へ2回連続でUpdateManyを呼んでも
// 両方成功し、各回のcountが正しいことを検証する。旧実装(自前でBegin/Commit/
// defer Closeを呼ぶ形)がセッションを閉じても、xormのClose()はConnをプールへ
// 返すだけで完全には破棄しないため、この2回呼び出しの再現力はsqlite上では弱い
// (敵対的検証で確認済み)。セッションを閉じないという設計上の意図はコード側の
// コメントで担保し、本テストはcountの正しさを主眼とする。
func TestTransfer_UpdateMany(t *testing.T) {
	t.Run("同じセッションへ複数回呼んでも成功する", func(t *testing.T) {
		ctx := setupTransferTestDB(t)
		insertTransfer(t, ctx, Transfer{RequestId: 1, TransactionId: "t1", TransferStatus: TRANSFER_STATUS_WAIT})
		insertTransfer(t, ctx, Transfer{RequestId: 1, TransactionId: "t2", TransferStatus: TRANSFER_STATUS_WAIT})

		count1, err := Transfer{}.UpdateMany(ctx, []Transfer{{Id: 1, TransferStatus: TRANSFER_STATUS_SUCCESS}})
		require.NoError(t, err)
		assert.EqualValues(t, 1, count1)

		count2, err := Transfer{}.UpdateMany(ctx, []Transfer{{Id: 2, TransferStatus: TRANSFER_STATUS_SUCCESS}})
		require.NoError(t, err)
		assert.EqualValues(t, 1, count2)
	})

	t.Run("countが実際の更新件数を返す(以前は常に0を返していた。MySQLでは実際に値が変わった行数)", func(t *testing.T) {
		ctx := setupTransferTestDB(t)
		insertTransfer(t, ctx, Transfer{RequestId: 1, TransactionId: "t1", TransferStatus: TRANSFER_STATUS_WAIT})
		insertTransfer(t, ctx, Transfer{RequestId: 1, TransactionId: "t2", TransferStatus: TRANSFER_STATUS_WAIT})

		count, err := Transfer{}.UpdateMany(ctx, []Transfer{
			{Id: 1, TransferStatus: TRANSFER_STATUS_SUCCESS},
			{Id: 2, TransferStatus: TRANSFER_STATUS_SUCCESS},
		})

		require.NoError(t, err)
		assert.EqualValues(t, 2, count)
	})
}
