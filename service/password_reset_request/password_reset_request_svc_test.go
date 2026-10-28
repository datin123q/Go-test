package password_reset_request

import (
	"context"
	"testing"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain/mocks"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
)

func Test_passwordResetRequestService_HandlePasswordResetRequest(t *testing.T) {
	type args struct {
		ctx   context.Context
		email string
	}
	tests := []struct {
		name                           string
		args                           args
		passwordResetRequestRepository func() *mocks.PasswordResetRequestRepository
		want                           domain.PasswordResetRequest
		wantErr                        error
	}{
		{
			name: "(성공) 비밀번호 재설정 요청 => 최초 요청",
			args: args{
				ctx:   context.Background(),
				email: "test@email.com",
			},
			passwordResetRequestRepository: func() *mocks.PasswordResetRequestRepository {
				passwordResetRequestRepository := &mocks.PasswordResetRequestRepository{}
				passwordResetRequestRepository.On("GetLatestByEmail", context.Background(), "test@email.com").Return(domain.PasswordResetRequest{}, cdomain.ErrItemNotFound)
				passwordResetRequestRepository.On("Create", context.Background(), mock.Anything).Return(mockPasswordResetRequest1.Id, nil)
				return passwordResetRequestRepository
			},
			want:    mockPasswordResetRequest1,
			wantErr: nil,
		},
		{
			name: "(성공) 비밀번호 재설정 요청 => 두번째 요청",
			args: args{
				ctx:   context.Background(),
				email: "test@email.com",
			},
			passwordResetRequestRepository: func() *mocks.PasswordResetRequestRepository {
				passwordResetRequestRepository := &mocks.PasswordResetRequestRepository{}
				passwordResetRequestRepository.On("GetLatestByEmail", context.Background(), "test@email.com").Return(mockPasswordResetRequest1, nil)
				passwordResetRequestRepository.On("Update", context.Background(), mock.Anything).Return(nil)
				return passwordResetRequestRepository
			},
			want:    mockPasswordResetRequest2,
			wantErr: nil,
		},
		{
			name: "(실패) 비밀번호 재설정 요청 => 세번째 요청",
			args: args{
				ctx:   context.Background(),
				email: "test@email.com",
			},
			passwordResetRequestRepository: func() *mocks.PasswordResetRequestRepository {
				passwordResetRequestRepository := &mocks.PasswordResetRequestRepository{}
				passwordResetRequestRepository.On("GetLatestByEmail", context.Background(), "test@email.com").Return(mockPasswordResetRequest2, nil)
				return passwordResetRequestRepository
			},
			want:    domain.PasswordResetRequest{},
			wantErr: cdomain.ErrRequestLimitExceeded,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := passwordResetRequestService{
				cfg: &config.Config{
					CommonConfig: config.CommonConfig{
						ServiceConfig: struct {
							Mail serviceutils.MailConfig
							Url  serviceutils.UrlConfig
						}{
							Url: serviceutils.UrlConfig{
								UserServiceDomain: "https://dev.energyshares.jp",
							},
						},
					},
				},
				passwordResetRequestRepository: tt.passwordResetRequestRepository(),
			}

			got, err := service.HandlePasswordResetRequest(tt.args.ctx, tt.args.email)
			if err != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("passwordResetReqeustService.HandlePasswordResetRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equalf(t, got.SendCount, tt.want.SendCount, "passwordResetRequestService.HandlePasswordResetRequest() = %v, want %v", got.SendCount, tt.want.SendCount)
		})
	}
}
