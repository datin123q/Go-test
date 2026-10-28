package controllers

import (
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

// TestWithdrawalTransferValidationErrorはMIR-1256の検証。以前はinvestorIdが
// 数値でない場合にlogrus.Fatal(admin-api全体が落ちる)、bank_code/branch_codeが
// nilの場合にnilポインタ参照のパニックがそれぞれ起きており、1人分のデータ
// 不整合が全体を巻き込んでいた。ここではプロセスを落とさずエラーを返すだけに
// なっていることを、境界値で確認する。
//
// bank_account/deposit_typeの判定はdomain/private_user.go(522-524行目付近)の
// 既存の口座登録済み判定(4項目)に合わせたもので、②レビューで
// bank_code/branch_codeの2項目だけでは不十分と指摘され追加した。
func TestWithdrawalTransferValidationError(t *testing.T) {
	tests := []struct {
		name        string
		investorId  string
		bankCode    *string
		branchCode  *string
		bankAccount string
		depositType domain.DEPOSIT_TYPE
		wantErr     bool
	}{
		{
			name:        "投資家IDが数値で口座情報が全て登録済みならエラー無し",
			investorId:  "12345",
			bankCode:    lo.ToPtr("0001"),
			branchCode:  lo.ToPtr("001"),
			bankAccount: "1234567",
			depositType: domain.ORDINARY_ACCOUNT,
			wantErr:     false,
		},
		{
			name:        "投資家IDが数値でない文字列だとエラー",
			investorId:  "not-a-number",
			bankCode:    lo.ToPtr("0001"),
			branchCode:  lo.ToPtr("001"),
			bankAccount: "1234567",
			depositType: domain.ORDINARY_ACCOUNT,
			wantErr:     true,
		},
		{
			name:        "投資家IDが空文字だとエラー",
			investorId:  "",
			bankCode:    lo.ToPtr("0001"),
			branchCode:  lo.ToPtr("001"),
			bankAccount: "1234567",
			depositType: domain.ORDINARY_ACCOUNT,
			wantErr:     true,
		},
		{
			name:        "銀行コードがnilだとエラー",
			investorId:  "12345",
			bankCode:    nil,
			branchCode:  lo.ToPtr("001"),
			bankAccount: "1234567",
			depositType: domain.ORDINARY_ACCOUNT,
			wantErr:     true,
		},
		{
			name:        "支店コードがnilだとエラー",
			investorId:  "12345",
			bankCode:    lo.ToPtr("0001"),
			branchCode:  nil,
			bankAccount: "1234567",
			depositType: domain.ORDINARY_ACCOUNT,
			wantErr:     true,
		},
		{
			name:        "投資家IDが数値でなく銀行コード・支店コードもnilだとエラー(最初の判定で弾かれる)",
			investorId:  "not-a-number",
			bankCode:    nil,
			branchCode:  nil,
			bankAccount: "1234567",
			depositType: domain.ORDINARY_ACCOUNT,
			wantErr:     true,
		},
		{
			name:        "銀行コードが空文字だとエラー(nilではないが未登録として扱う)",
			investorId:  "12345",
			bankCode:    lo.ToPtr(""),
			branchCode:  lo.ToPtr("001"),
			bankAccount: "1234567",
			depositType: domain.ORDINARY_ACCOUNT,
			wantErr:     true,
		},
		{
			name:        "支店コードが空文字だとエラー(nilではないが未登録として扱う)",
			investorId:  "12345",
			bankCode:    lo.ToPtr("0001"),
			branchCode:  lo.ToPtr(""),
			bankAccount: "1234567",
			depositType: domain.ORDINARY_ACCOUNT,
			wantErr:     true,
		},
		{
			name:        "口座番号が空文字だとエラー",
			investorId:  "12345",
			bankCode:    lo.ToPtr("0001"),
			branchCode:  lo.ToPtr("001"),
			bankAccount: "",
			depositType: domain.ORDINARY_ACCOUNT,
			wantErr:     true,
		},
		{
			name:        "預金種別が空文字だとエラー",
			investorId:  "12345",
			bankCode:    lo.ToPtr("0001"),
			branchCode:  lo.ToPtr("001"),
			bankAccount: "1234567",
			depositType: "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := withdrawalTransferValidationError(tt.investorId, tt.bankCode, tt.branchCode, tt.bankAccount, tt.depositType)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
