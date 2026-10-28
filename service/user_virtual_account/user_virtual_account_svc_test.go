package user_virtual_account

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"

	"github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain/mocks"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

func Test_userVirtualAccountService_LinkVirtualAccount(t *testing.T) {
	type args struct {
		ctx         context.Context
		userId      int64
		projectType domain.PROJECT_TYPE
	}

	tests := []struct {
		name                         string
		args                         args
		userVirtualAccountRepository func() *mocks.UserVirtualAccountRepository
		wantErr                      error
	}{
		{
			name: "(성공) 가상 계좌 연결",
			args: args{
				ctx:         context.Background(),
				userId:      int64(1),
				projectType: domain.EQUITY,
			},
			userVirtualAccountRepository: func() *mocks.UserVirtualAccountRepository {
				userRepository := &mocks.UserVirtualAccountRepository{}
				userRepository.On("GetByUserIdAndLinkedStatus", context.Background(), int64(1), domain.EQUITY, domain.ON_VA_ACCOUNT).
					Return(domain.UserVirtualAccount{
						Id:           1,
						UserId:       1,
						ProjectType:  domain.EQUITY,
						LinkedStatus: domain.ON_VA_ACCOUNT,
					}, nil)
				userRepository.On("Update", context.Background(), mock.MatchedBy(func(request domain.UserVirtualAccount) bool {
					ignoreType := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

					return cmp.Equal(request, domain.UserVirtualAccount{
						Id:           1,
						UserId:       0,
						ProjectType:  domain.EQUITY,
						LinkedStatus: domain.OFF_VA_ACCOUNT,
						UpdateDate:   jsonutils.JSONTime(time.Now()),
					}, ignoreType)
				})).Return(nil)
				userRepository.On("GetByUserIdAndLinkedStatus", context.Background(), int64(0), domain.EQUITY, domain.OFF_VA_ACCOUNT).
					Return(domain.UserVirtualAccount{
						Id:           1,
						UserId:       0,
						ProjectType:  domain.EQUITY,
						LinkedStatus: domain.OFF_VA_ACCOUNT,
					}, nil)
				userRepository.On("Update", context.Background(), mock.MatchedBy(func(request domain.UserVirtualAccount) bool {
					ignoreType := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

					return cmp.Equal(request, domain.UserVirtualAccount{
						Id:           1,
						UserId:       1,
						LinkedStatus: domain.ON_VA_ACCOUNT,
						ProjectType:  domain.EQUITY,
						UpdateDate:   jsonutils.JSONTime(time.Now()),
					}, ignoreType)
				})).Return(nil)
				return userRepository
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := userVirtualAccountService{
				userVirtualAccountRepository: tt.userVirtualAccountRepository(),
			}

			err := service.LinkVirtualAccount(tt.args.ctx, tt.args.userId, tt.args.projectType)
			if err != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("userVirtualAccountService.LinkVirtualAccount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
