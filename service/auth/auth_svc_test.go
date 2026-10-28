package auth

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain/mocks"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

func Test_AuthService_CreateUserAccessToken(t *testing.T) {
	type args struct {
		ctx  context.Context
		user domain.User
	}
	tests := []struct {
		name    string
		cfg     func() *config.Config
		args    args
		want    string
		wantErr error
	}{
		{
			name: "(성공) CreateUserAccessToken 토큰 생성 -> 정상적으로 토큰",
			args: args{
				ctx:  context.Background(),
				user: mockUser1,
			},
			cfg: func() *config.Config {
				cfg := &config.Config{}
				cfg.Env = "DEV"
				return cfg
			},
			want:    mock.Anything,
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &authService{
				cfg: tt.cfg(),
			}

			got, err := svc.CreateUserAccessToken(tt.args.ctx, tt.args.user)
			if err != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("authService.CreateUserAccessToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.NotEmpty(t, got)
		})
	}
}

func TestAuthService_CreateUserRefreshToken(t *testing.T) {
	type args struct {
		ctx   context.Context
		input cdomain.UserRefreshTokenInput
	}

	testToken := "refresh-token"

	tests := []struct {
		name                   string
		cfg                    func() *config.Config
		refreshTokenRepository func() *mocks.RefreshTokenRepository
		args                   args
		want                   string
		wantErr                error
	}{
		{
			name: "(성공) CreateUserRefreshToken 토큰 생성 -> 정상적으로 토큰",
			args: args{
				ctx: context.Background(),
				input: cdomain.UserRefreshTokenInput{
					User:  mockUser1,
					IP:    "ip-address",
					Agent: "user-agent",
				},
			},
			cfg: func() *config.Config {
				cfg := &config.Config{}
				cfg.Env = "DEV"
				return cfg
			},
			refreshTokenRepository: func() *mocks.RefreshTokenRepository {
				refreshTokenRepository := &mocks.RefreshTokenRepository{}
				refreshTokenRepository.On("Create", context.Background(), mock.MatchedBy(func(refreshToken domain.RefreshToken) bool {
					ignoreFields := cmpopts.IgnoreFields(domain.RefreshToken{}, "Token", "ExpireIn")
					ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

					return cmp.Equal(refreshToken, domain.RefreshToken{
						UserId:   mockUser1.Id,
						UserType: mockUser1.UserType,
						IsActive: true,
						ExpireIn: int64(1),
						Token:    testToken,
						Email:    mockUser1.Email,
						IpAddr:   "ip-address",
						Agent:    "user-agent",
					}, ignoreFields, ignoreTime)
				})).Return(domain.RefreshToken{}, nil)

				return refreshTokenRepository
			},
			want:    testToken,
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &authService{
				cfg:                    tt.cfg(),
				refreshTokenRepository: tt.refreshTokenRepository(),
			}

			got, err := svc.CreateUserRefreshToken(tt.args.ctx, tt.args.input)
			if err != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("authService.CreateUserRefreshToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.NotEmpty(t, got)
		})
	}
}
