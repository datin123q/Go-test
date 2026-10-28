package domain

import (
	"context"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

// MIR-1481: マイページの出金口座の変更は、口座の列だけを書く。取引停止中の会員が口座を
// 変えても、取引停止・メルマガ同意・SMS 認証済み・ハガキ発送済みは変わらない。
func TestUser_UpdateWithdrawalAccount(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	require.NoError(t, err)
	engine.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, engine.Close()) })
	require.NoError(t, engine.Sync2(new(User)))
	ctx := context.WithValue(context.Background(), factory.ConextDBType, engine)

	agreedAt := time.Date(2025, 4, 1, 3, 0, 0, 0, time.UTC)
	u := User{
		Uid: "u1", Email: "u1@example.invalid", Name: "テスト 太郎", Address: "1-1", Depositor: "テスト タロウ",
		BlockInvest: true, IsNewsletter: true, NewsletterAgreeAt: jsonutils.JSONTime(agreedAt), IsSmsCert: true, PostalCodeStatus: true,
		IsSurplusFund: lo.ToPtr(true), AuditStatus: COMPLETED_AUDIT,
		DepositType: "ORDINARY_ACCOUNT", BankName: "旧銀行", BranchName: "旧支店", BankAccount: "1111111", BankCode: lo.ToPtr("0001"),
		BranchCode: lo.ToPtr("001"),
	}
	_, err = engine.UseBool().Insert(&u)
	require.NoError(t, err)

	// UpdateWithdrawalAccount のハンドラと同じく、口座の列だけを持つ新しい構造体を渡す。
	in := &User{Id: u.Id, DepositType: "CHECKING_ACCOUNT", BankName: "新銀行", BranchName: "新支店", BankAccount: "2222222",
		BankCode: lo.ToPtr("0005"), BranchCode: lo.ToPtr("123")}
	require.NoError(t, in.UpdateWithdrawalAccount(ctx))

	var got User
	_, err = engine.ID(u.Id).Get(&got)
	require.NoError(t, err)
	assert.True(t, got.BlockInvest, "取引停止が外れない")
	assert.True(t, got.IsNewsletter, "メルマガ同意が消えない")
	assert.True(t, agreedAt.Equal(got.NewsletterAgreeAt.Time()))
	assert.True(t, got.IsSmsCert, "SMS 認証済みが消えない")
	assert.True(t, got.PostalCodeStatus, "ハガキ発送済みが消えない")
	assert.Equal(t, lo.ToPtr(true), got.IsSurplusFund)
	assert.Equal(t, COMPLETED_AUDIT, got.AuditStatus)
	assert.Equal(t, "テスト 太郎", got.Name)
	assert.Equal(t, "1-1", got.Address)
	assert.Equal(t, "テスト タロウ", got.Depositor, "預金者名は変えない")

	assert.Equal(t, DEPOSIT_TYPE("CHECKING_ACCOUNT"), got.DepositType)
	assert.Equal(t, "新銀行", got.BankName)
	assert.Equal(t, "新支店", got.BranchName)
	assert.Equal(t, "2222222", got.BankAccount)
	assert.Equal(t, lo.ToPtr("0005"), got.BankCode)
	assert.Equal(t, lo.ToPtr("123"), got.BranchCode)

	// 6列は送られた値でまとめて置き換える。空の列も書く(古い支店名・コードを残さない)。
	in = &User{Id: u.Id, DepositType: "ORDINARY_ACCOUNT", BankName: "別銀行", BankAccount: "3333333"}
	require.NoError(t, in.UpdateWithdrawalAccount(ctx))
	got = User{}
	_, err = engine.ID(u.Id).Get(&got)
	require.NoError(t, err)
	assert.Equal(t, "別銀行", got.BankName)
	assert.Equal(t, "", got.BranchName)
	assert.Nil(t, got.BankCode)
	assert.Nil(t, got.BranchCode)
	assert.True(t, got.BlockInvest)
}
