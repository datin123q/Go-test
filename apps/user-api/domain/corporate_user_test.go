package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

func validWithdrawalAccount() CorporateUserRegisterWithdrawalAccountItem {
	return CorporateUserRegisterWithdrawalAccountItem{
		BankName:    "テスト銀行",
		BranchName:  "テスト支店",
		DepositType: domain.ORDINARY_ACCOUNT,
		BankAccount: "1234567",
		Depositor:   "ｶﾌﾞｼｷｶﾞｲｼｬﾃｽﾄ",
	}
}

// MIR-1171: depositorに漢字が含まれると振込バッチ全体がGMOあおぞらネット銀行に拒否される
// ため、新規登録時点(SIGNUP_ACCOUNT)で文字種を検証する。
func TestCorporateUserRegisterProfilesRequest_Validate_SignUpAccount(t *testing.T) {
	tests := []struct {
		name      string
		depositor string
		wantErr   bool
	}{
		{name: "半角カタカナは許可", depositor: "ｶﾌﾞｼｷｶﾞｲｼｬﾃｽﾄ", wantErr: false},
		{name: "漢字を含むと拒否", depositor: "株式会社テスト", wantErr: true},
		{name: "ひらがなを含むと拒否", depositor: "かぶしきがいしゃ", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			account := validWithdrawalAccount()
			account.Depositor = tt.depositor

			req := CorporateUserRegisterProfilesRequest{
				SignUpStatus:      domain.SIGNUP_ACCOUNT,
				WithdrawalAccount: account,
			}

			err := req.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
