package domain

import (
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

// TestWithdrawalBankInfoErrorはMIR-1256の検証。銀行コード・支店コード・口座番号・
// 預金種別の4項目のうちいずれかが未登録(nilまたは空文字)ならエラーを返す。
// domain/private_user.go(522-524行目付近)の既存の口座登録済み判定と揃えている。
func TestWithdrawalBankInfoError(t *testing.T) {
	tests := []struct {
		name        string
		bankCode    *string
		branchCode  *string
		bankAccount string
		depositType DEPOSIT_TYPE
		wantErr     bool
	}{
		{
			name:        "4項目全て登録済みならエラー無し",
			bankCode:    lo.ToPtr("0001"),
			branchCode:  lo.ToPtr("001"),
			bankAccount: "1234567",
			depositType: ORDINARY_ACCOUNT,
			wantErr:     false,
		},
		{
			name:        "銀行コードがnilだとエラー",
			bankCode:    nil,
			branchCode:  lo.ToPtr("001"),
			bankAccount: "1234567",
			depositType: ORDINARY_ACCOUNT,
			wantErr:     true,
		},
		{
			name:        "銀行コードが空文字だとエラー",
			bankCode:    lo.ToPtr(""),
			branchCode:  lo.ToPtr("001"),
			bankAccount: "1234567",
			depositType: ORDINARY_ACCOUNT,
			wantErr:     true,
		},
		{
			name:        "支店コードがnilだとエラー",
			bankCode:    lo.ToPtr("0001"),
			branchCode:  nil,
			bankAccount: "1234567",
			depositType: ORDINARY_ACCOUNT,
			wantErr:     true,
		},
		{
			name:        "口座番号が空文字だとエラー",
			bankCode:    lo.ToPtr("0001"),
			branchCode:  lo.ToPtr("001"),
			bankAccount: "",
			depositType: ORDINARY_ACCOUNT,
			wantErr:     true,
		},
		{
			name:        "預金種別が空文字だとエラー",
			bankCode:    lo.ToPtr("0001"),
			branchCode:  lo.ToPtr("001"),
			bankAccount: "1234567",
			depositType: "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := WithdrawalBankInfoError(tt.bankCode, tt.branchCode, tt.bankAccount, tt.depositType)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestRequestFee_ReturnsErrorWithoutPanicWhenBankInfoMissingはMIR-1256の検証。
// 以前はuser.BankCode/BranchCodeをnilチェック無しで参照しており、会員本人の
// 口座情報が未登録なだけでこの関数を呼んだuser-apiプロセス全体がパニックで
// 落ちていた。ここではHTTPリクエストへ進む前にエラーを返し、パニックしない
// ことを確認する(job-schedulerへの実際のHTTP呼び出しは発生しないため、
// サーバーの起動は不要)。
func TestRequestFee_ReturnsErrorWithoutPanicWhenBankInfoMissing(t *testing.T) {
	user := User{Id: 1, Name: "test", BankCode: nil, BranchCode: nil}

	assert.NotPanics(t, func() {
		_, err := Transfer{}.RequestFee(nil, user, 1000) //nolint:staticcheck // 検証対象はHTTPへ進む前のガードのみ
		assert.Error(t, err)
	})
}
