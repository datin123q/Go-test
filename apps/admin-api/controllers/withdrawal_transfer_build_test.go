package controllers

import (
	"context"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

func setupBuildWithdrawalTransfersForCloseTestDB(t *testing.T) (context.Context, *xorm.Engine) {
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

// TestBuildWithdrawalTransfersForClose_SkipsInvalidAndMarksFailはMIR-1256の検証。
// 1回目の敵対的検証で「対象から外すだけ(continue)では、外した出金依頼がWAITのまま
// 復旧不能になる(個別再送金・キャンセルどちらの経路にも進めず、引き落とし済みの
// 残高を戻せない)」という重大な指摘を受け、外す際にその場でFAILへ更新するよう
// 直した。ここでは有効な1件・投資家IDが数値でない1件・銀行コードが未登録
// (nilと空文字の両方)の1件を混在させ、有効な1件だけがtransferに積まれ、
// 残り2件はskippedWithdrawalIdsに入り、かつDB上のstatusが実際にFAILへ
// 更新されていることを確認する。
func TestBuildWithdrawalTransfersForClose_SkipsInvalidAndMarksFail(t *testing.T) {
	ctx, engine := setupBuildWithdrawalTransfersForCloseTestDB(t)

	valid := domain.Withdrawal{InvestorId: "1", InvestorName: "valid", Amount: 1000, Status: domain.WITHDRAWAL_STATUS_WAIT}
	invalidId := domain.Withdrawal{InvestorId: "not-a-number", InvestorName: "invalid-id", Amount: 2000, Status: domain.WITHDRAWAL_STATUS_WAIT}
	invalidBank := domain.Withdrawal{InvestorId: "3", InvestorName: "invalid-bank", Amount: 3000, Status: domain.WITHDRAWAL_STATUS_WAIT}
	invalidDepositType := domain.Withdrawal{InvestorId: "4", InvestorName: "invalid-deposit-type", Amount: 4000, Status: domain.WITHDRAWAL_STATUS_WAIT}
	_, err := engine.Insert(&valid, &invalidId, &invalidBank, &invalidDepositType)
	require.NoError(t, err)

	items := []domain.WithdrawalDetail{
		{
			Withdrawal: valid,
			User:       domain.User{BankCode: lo.ToPtr("0001"), BranchCode: lo.ToPtr("001"), BankAccount: "1234567", DepositType: domain.ORDINARY_ACCOUNT},
		},
		{
			Withdrawal: invalidId,
			User:       domain.User{BankCode: lo.ToPtr("0001"), BranchCode: lo.ToPtr("001"), BankAccount: "1234567", DepositType: domain.ORDINARY_ACCOUNT},
		},
		{
			// 銀行コードが空文字(未登録)。nilではない点がポイント。
			Withdrawal: invalidBank,
			User:       domain.User{BankCode: lo.ToPtr(""), BranchCode: lo.ToPtr("001"), BankAccount: "1234567", DepositType: domain.ORDINARY_ACCOUNT},
		},
		{
			// 預金種別が空文字(未登録)。②レビューでbank_code/branch_codeの2項目
			// だけでは不十分と指摘され、bank_account/deposit_typeも見るようにした。
			Withdrawal: invalidDepositType,
			User:       domain.User{BankCode: lo.ToPtr("0001"), BranchCode: lo.ToPtr("001"), BankAccount: "1234567", DepositType: ""},
		},
	}

	transfers, mappings, skippedIds, err := buildWithdrawalTransfersForClose(ctx, items)
	require.NoError(t, err)

	require.Len(t, transfers, 1, "有効な1件だけがtransferに積まれるはず")
	assert.EqualValues(t, 1, transfers[0].UserId, "investor_idの数値がUserIdへ入っているはず")
	assert.Equal(t, valid.Amount, transfers[0].TransferAmount)
	require.Len(t, mappings, 1)
	assert.Equal(t, valid.Id, mappings[0].WithdrawalId)
	assert.Equal(t, transfers[0].TransactionId, mappings[0].TransactionId)

	assert.ElementsMatch(t, []int64{invalidId.Id, invalidBank.Id, invalidDepositType.Id}, skippedIds,
		"投資家IDが数値でない・銀行コードが空文字・預金種別が空文字の3件がskipされるはず")

	// スキップされた3件はDB上でFAILへ更新されている。
	for _, id := range []int64{invalidId.Id, invalidBank.Id, invalidDepositType.Id} {
		var status string
		ok, getErr := engine.Table("withdrawal").Where("id = ?", id).Cols("status").Get(&status)
		require.NoError(t, getErr)
		require.True(t, ok)
		assert.Equal(t, "FAIL", status, "id=%d はFAILへ更新されているはず", id)
	}

	// 有効だった1件のstatusはこの関数では変更されない(WAITのまま)。
	var validStatus string
	ok, getErr := engine.Table("withdrawal").Where("id = ?", valid.Id).Cols("status").Get(&validStatus)
	require.NoError(t, getErr)
	require.True(t, ok)
	assert.Equal(t, "WAIT", validStatus)
}

// TestBuildWithdrawalTransfersForClose_AllInvalid_ReturnsEmptyはMIR-1256の検証。
// 全件が無効な場合、transfersは空スライスで返り、呼び出し元(RequestTransferByClosedId)
// はこれを見て締めのステータスを進めずエラーを返す。
func TestBuildWithdrawalTransfersForClose_AllInvalid_ReturnsEmpty(t *testing.T) {
	ctx, engine := setupBuildWithdrawalTransfersForCloseTestDB(t)

	invalid := domain.Withdrawal{InvestorId: "not-a-number", InvestorName: "invalid", Amount: 1000, Status: domain.WITHDRAWAL_STATUS_WAIT}
	_, err := engine.Insert(&invalid)
	require.NoError(t, err)

	items := []domain.WithdrawalDetail{
		{Withdrawal: invalid, User: domain.User{BankCode: lo.ToPtr("0001"), BranchCode: lo.ToPtr("001")}},
	}

	transfers, mappings, skippedIds, err := buildWithdrawalTransfersForClose(ctx, items)
	require.NoError(t, err)
	assert.Empty(t, transfers)
	assert.Empty(t, mappings)
	assert.Equal(t, []int64{invalid.Id}, skippedIds)
}
