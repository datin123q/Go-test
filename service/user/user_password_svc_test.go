package user

import (
	"context"
	"testing"

	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain/mocks"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

func Test_userService_ChangePassword(t *testing.T) {
	type args struct {
		ctx      context.Context
		user     domain.User
		password string
	}
	tests := []struct {
		name           string
		args           args
		userRepository func() *mocks.UserRepository
		pwhelper       func() *mocks.PasswordHelper
		wantErr        error
	}{
		{
			name: "(성공) 비밀번호 변경 => 성공",
			args: args{
				ctx:      context.Background(),
				user:     mockUserChangePasswordBefore,
				password: "new_password",
			},
			userRepository: func() *mocks.UserRepository {
				userRepository := &mocks.UserRepository{}

				userRepository.On("Update", context.Background(), mockUserChangePasswordAfter).Return(nil)
				return userRepository
			},
			pwhelper: func() *mocks.PasswordHelper {
				pwhelper := &mocks.PasswordHelper{}
				pwhelper.On("GenerateSalt").Return("salt")
				pwhelper.On("EncodePassword", "new_password", "salt").Return(mockUserChangePasswordAfter.Password)
				pwhelper.On("IsSamePassword", mockUserChangePasswordBefore.Password, mockUserChangePasswordAfter.Password).Return(false)
				return pwhelper
			},
			wantErr: nil,
		},
		{
			name: "(실패) 비밀번호 변경 => 동일한 비밀번호",
			args: args{
				ctx:      context.Background(),
				user:     mockUserChangePasswordBefore,
				password: "password",
			},
			userRepository: func() *mocks.UserRepository {
				userRepository := &mocks.UserRepository{}
				return userRepository
			},
			pwhelper: func() *mocks.PasswordHelper {
				pwhelper := &mocks.PasswordHelper{}
				pwhelper.On("GenerateSalt").Return("salt")
				pwhelper.On("EncodePassword", "password", "salt").Return("encoded_password")
				pwhelper.On("IsSamePassword", mockUserChangePasswordBefore.Password, "encoded_password").Return(true)
				return pwhelper
			},
			wantErr: cdomain.ErrItemConflict,
		},
		{
			name: "(실패) 비밀번호 변경 => 이메일을 포함한 비밀번호",
			args: args{
				ctx:      context.Background(),
				user:     mockUserChangePasswordBefore,
				password: "password" + mockUserChangePasswordBefore.Email,
			},
			userRepository: func() *mocks.UserRepository {
				userRepository := &mocks.UserRepository{}
				return userRepository
			},
			pwhelper: func() *mocks.PasswordHelper {
				pwhelper := &mocks.PasswordHelper{}
				pwhelper.On("GenerateSalt").Return("salt")
				pwhelper.On("EncodePassword", "password"+mockUserChangePasswordBefore.Email, "salt").Return("encoded_password")
				return pwhelper
			},
			wantErr: cdomain.ErrItemInvalid,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := userService{
				pwhelper:       tt.pwhelper(),
				userRepository: tt.userRepository(),
			}

			err := service.ChangePassword(tt.args.ctx, tt.args.user, tt.args.password)
			if err != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("userService.ChangePassword() error = %v, wantErr %v", err, tt.wantErr)
			}

		})
	}
}
