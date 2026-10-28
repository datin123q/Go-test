package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

func validPrivateWithdrawalAccount() PrvUserWithdrawalAccount {
	return PrvUserWithdrawalAccount{
		BankName:    "テスト銀行",
		BranchName:  "テスト支店",
		BankAccount: "1234567",
		BankCode:    "0001",
		BranchCode:  "001",
		Depositor:   "ﾔﾏﾀﾞ ﾀﾛｳ",
		DepositType: domain.ORDINARY_ACCOUNT,
	}
}

// MIR-1171: depositorに漢字が含まれると振込バッチ全体がGMOあおぞらネット銀行に拒否される。
// 個人投資家の画面ではdepositorをフリガナから自動生成し直接入力させていないが、このAPI
// 自体はリクエストの値をそのまま受け付けるため、法人と同様に新規登録時点で検証する。
func TestPrvUserRegisterProfileRequest_Validate_SignUpAccount(t *testing.T) {
	tests := []struct {
		name      string
		depositor string
		wantErr   bool
	}{
		{name: "半角カタカナは許可", depositor: "ﾔﾏﾀﾞ ﾀﾛｳ", wantErr: false},
		{name: "漢字を含むと拒否", depositor: "山田太郎", wantErr: true},
		{name: "ひらがなを含むと拒否", depositor: "やまだたろう", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			account := validPrivateWithdrawalAccount()
			account.Depositor = tt.depositor

			req := PrvUserRegisterProfileRequest{
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
