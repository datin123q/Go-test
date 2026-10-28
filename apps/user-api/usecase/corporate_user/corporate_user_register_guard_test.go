package corporate_user

import (
	"context"
	"errors"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain/mocks"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	udomain "github.com/illuminarean/energyshares-user-api/domain"
)

// MIR-1421: 登録が完了した法人は、保留・仮登録へ差し戻されていても、登録用のルートで出金口座(名義を含む)などを
// 書き換えられない。審査が確定・進行中の状態は、完了前でも弾く。一度も登録を完了していない法人の入力は通す。
func Test_corporateUserUsecase_RegisterProfiles_ReentryGuard(t *testing.T) {
	tests := []struct {
		name         string
		signupStatus domain.SIGNUP_STATUS
		auditStatus  domain.AUDIT_STATUS
		wantBlocked  bool
	}{
		{"登録完了・審査待ちは弾く", domain.SIGNUP_COMPLETE, domain.WAIT_AUDIT, true},
		{"登録完了・承認済みは弾く", domain.SIGNUP_COMPLETE, domain.COMPLETED_AUDIT, true},
		{"登録完了・住所確認中は弾く", domain.SIGNUP_COMPLETE, domain.CHECK_ADDRESS_AUDIT, true},
		{"登録途中でも無効化は弾く", domain.SIGNUP_ACCOUNT, domain.INVALIDITY_AUDIT, true},
		{"登録途中でも反社確認中は弾く", domain.SIGNUP_ACCOUNT, domain.CHECK_ANTISOCIAL_AUDIT, true},
		{"登録途中でも反社確認待ちは弾く", domain.SIGNUP_ACCOUNT, domain.WAIT_ANTISOCIAL_AUDIT, true},
		{"登録途中でも最終承認待機は弾く", domain.SIGNUP_ACCOUNT, domain.WAIT_FINAL_APPROVAL_AUDIT, true},
		{"登録の状態が途中のまま承認済みでも弾く", domain.SIGNUP_ACCOUNT, domain.COMPLETED_AUDIT, true},
		{"登録完了で保留へ差し戻されていても弾く", domain.SIGNUP_COMPLETE, domain.HOLD_AUDIT, true},
		{"登録完了で仮登録へ差し戻されていても弾く", domain.SIGNUP_COMPLETE, domain.PROVISIONAL_AUDIT, true},
		{"登録途中(仮登録)は従来どおり入力できる", domain.SIGNUP_CORPORATE_PROFILE_SHAREHOLDER, domain.PROVISIONAL_AUDIT, false},
		{"登録途中で保留にされた法人は入力できる", domain.SIGNUP_CORPORATE_PROFILE_SHAREHOLDER, domain.HOLD_AUDIT, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			user := mockUser1
			user.SignupStatus = tt.signupStatus
			user.AuditStatus = tt.auditStatus

			userRepo := &mocks.UserRepository{}
			userRepo.On("GetById", ctx, user.Id).Return(user, nil)
			userRepo.On("Update", ctx, mock.Anything).Return(nil)
			profileRepo := &mocks.CorporateProfileRepository{}
			profileRepo.On("GetByUserId", ctx, user.Id).Return(mockCorporateProfile1, nil)
			profileRepo.On("Update", ctx, mock.Anything).Return(nil)

			usecase := corporateUserUsecase{
				userRepository:                 userRepo,
				corporateProfileRepository:     profileRepo,
				corporateShareholderRepository: &mocks.CorporateShareholderRepository{},
			}
			err := usecase.RegisterProfiles(ctx, udomain.CorporateUserRegisterProfilesRequest{
				User:         jwthelper.User{Id: user.Id},
				SignUpStatus: domain.SIGNUP_ACCOUNT,
				WithdrawalAccount: udomain.CorporateUserRegisterWithdrawalAccountItem{
					BankName: "別の銀行", BankCode: lo.ToPtr("0001"), BranchName: "支店", BranchCode: lo.ToPtr("001"),
					BankAccount: "7654321", Depositor: "ベツメイギ",
				},
			})

			if tt.wantBlocked {
				require.Error(t, err)
				assert.True(t, errors.Is(err, cdomain.ErrItemConflict), "err = %v", err)
				userRepo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
				profileRepo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
				return
			}
			require.NoError(t, err)
			userRepo.AssertCalled(t, "Update", ctx, mock.MatchedBy(func(u domain.User) bool {
				return u.Depositor == "ベツメイギ" && u.BankAccount == "7654321"
			}))
		})
	}
}
