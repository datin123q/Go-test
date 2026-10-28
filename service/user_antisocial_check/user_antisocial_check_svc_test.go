package user_antisocial_check

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain/mocks"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

func Test_UserAntisocialCheckService_GetByUserId(t *testing.T) {
	type args struct {
		ctx    context.Context
		userId int64
	}
	tests := []struct {
		name                          string
		args                          args
		userAntisocialCheckRepository func() *mocks.UserAntisocialCheckRepository
		antisocialCheckFileRepository func() *mocks.AntisocialCheckFileRepository
		want                          domain.UserAntisocialCheck
		wantErr                       bool
	}{
		{
			name: "(성공) UserAntisocialCheck 모델 조회 (Id:1000) -> 정상적으로 조회되기를 기대합니다.",
			args: args{
				ctx:    context.Background(),
				userId: mockUserAntisocialCheck1.CheckUserId,
			},
			userAntisocialCheckRepository: func() *mocks.UserAntisocialCheckRepository {
				userAntisocialCheckRepository := &mocks.UserAntisocialCheckRepository{}
				userAntisocialCheckRepository.On("GetByUserId", context.Background(),
					mockUserAntisocialCheck1.CheckUserId).
					Return(mockUserAntisocialCheck1, nil)
				return userAntisocialCheckRepository
			},
			antisocialCheckFileRepository: func() *mocks.AntisocialCheckFileRepository {
				antisocialCheckFileRepository := &mocks.AntisocialCheckFileRepository{}
				antisocialCheckFileRepository.On("GetAll", context.Background(),
					mock.AnythingOfType("func(*clean_arch_domain.AntisocialCheckFileQueryOptionFields)"),
					mock.AnythingOfType("func(*clean_arch_domain.AntisocialCheckFileQueryOptionFields)"),
					mock.AnythingOfType("func(*clean_arch_domain.AntisocialCheckFileQueryOptionFields)"),
					mock.AnythingOfType("func(*clean_arch_domain.AntisocialCheckFileQueryOptionFields)")).
					Return(int64(2), mocAntisocialCheckFiles1, nil)
				return antisocialCheckFileRepository
			},
			want:    mockUserAntisocialCheck2,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := userAntisocialCheckService{
				userAntisocialCheckRepository: tt.userAntisocialCheckRepository(),
				antisocialCheckFileRepository: tt.antisocialCheckFileRepository(),
			}

			got, err := service.GetByUserId(tt.args.ctx, tt.args.userId)

			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})
			ignoreFields := cmpopts.IgnoreFields(domain.AntisocialCheckFile{}, "FileUrl")

			assert.Equalf(t, tt.wantErr, err != nil, "userAntisocialCheckService.GetByUserId() = %v, wantErr %v", err, tt.wantErr)
			if diff := cmp.Diff(tt.want, got, ignoreTime, ignoreFields); diff != "" {
				t.Errorf("lookUpMyNumberRepository.GetByMyNumberId() mismatch (-want +got):\n%s", diff)
			}

		})
	}
}
