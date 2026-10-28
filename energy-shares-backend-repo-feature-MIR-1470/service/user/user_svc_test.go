package user

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain/mocks"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

func Test_userService_IsDuplicateEmail(t *testing.T) {
	type args struct {
		ctx   context.Context
		email string
	}
	tests := []struct {
		name           string
		args           args
		userRepository func() *mocks.UserRepository
		want           bool
		wantErr        error
	}{
		{
			name: "(성공) 이메일 중복 확인 => 중복되지 않은 이메일",
			args: args{
				ctx:   context.Background(),
				email: mock.Anything,
			},
			userRepository: func() *mocks.UserRepository {
				userRepository := &mocks.UserRepository{}
				userRepository.On("GetByEmail", context.Background(), mock.Anything).Return(domain.User{}, cdomain.ErrItemNotFound)
				return userRepository
			},
			want:    false,
			wantErr: nil,
		},
		{
			name: "(성공) 이메일 중복 확인 => 중복된 이메일",
			args: args{
				ctx:   context.Background(),
				email: mock.Anything,
			},
			userRepository: func() *mocks.UserRepository {
				userRepository := &mocks.UserRepository{}
				userRepository.On("GetByEmail", context.Background(), mock.Anything).Return(domain.User{}, nil)
				return userRepository
			},
			want:    true,
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := userService{
				userRepository: tt.userRepository(),
			}

			got, err := service.IsDuplicateEmail(tt.args.ctx, tt.args.email)
			if err != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("userService.IsDuplicateEmail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equalf(t, tt.want, got, "userService.IsDuplicateEmail() = %v, want %v", got, tt.want)
		})
	}
}

func Test_userService_IsDuplicateEmailIncludeDeleted(t *testing.T) {
	type args struct {
		ctx   context.Context
		email string
	}
	tests := []struct {
		name           string
		args           args
		userRepository func() *mocks.UserRepository
		want           bool
		wantErr        error
	}{
		{
			name: "(성공) 이메일 중복 확인 => 중복되지 않은 이메일",
			args: args{
				ctx:   context.Background(),
				email: mock.Anything,
			},
			userRepository: func() *mocks.UserRepository {
				userRepository := &mocks.UserRepository{}
				userRepository.On("GetByEmailIncludeDeleted", context.Background(), mock.Anything).Return(domain.User{}, cdomain.ErrItemNotFound)
				return userRepository
			},
			want:    false,
			wantErr: nil,
		},
		{
			name: "(성공) 이메일 중복 확인 => 중복된 이메일",
			args: args{
				ctx:   context.Background(),
				email: mock.Anything,
			},
			userRepository: func() *mocks.UserRepository {
				userRepository := &mocks.UserRepository{}
				userRepository.On("GetByEmailIncludeDeleted", context.Background(), mock.Anything).Return(domain.User{}, nil)
				return userRepository
			},
			want:    true,
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := userService{
				userRepository: tt.userRepository(),
			}

			got, err := service.IsDuplicateEmailIncludeDeleted(tt.args.ctx, tt.args.email)
			if err != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("userService.IsDuplicateEmail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equalf(t, tt.want, got, "userService.IsDuplicateEmail() = %v, want %v", got, tt.want)
		})
	}
}
