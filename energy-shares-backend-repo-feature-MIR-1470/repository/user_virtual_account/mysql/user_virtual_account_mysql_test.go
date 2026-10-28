package mysql

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"

	"github.com/stretchr/testify/assert"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

func Test_userVirtualAccountMysqlRepository_GetByUserIdAndLinkedStatus(t *testing.T) {
	type args struct {
		ctx          context.Context
		UserId       int64
		ProjectType  domain.PROJECT_TYPE
		LinkedStatus domain.VIRTUAL_ACCOUNT_LINK_STATUS
	}
	tests := []struct {
		name    string
		args    args
		want    domain.UserVirtualAccount
		wantErr bool
	}{
		{
			name: "특정 유저, EQUITY 타입 프로젝트, ON 상태의 가상계좌를 조회합니다 : 2000 => 1건 조회되기를 기대합니다.",
			args: args{
				ctx:          context.Background(),
				UserId:       2000,
				ProjectType:  domain.EQUITY,
				LinkedStatus: domain.ON_VA_ACCOUNT,
			},
			want:    mockUserVirtualAccount1,
			wantErr: false,
		},
		{
			name: "특정 유저, LOAN 타입 프로젝트의 가상계좌를 조회합니다 : 2000, Loan Project Type => 조회되지 않습니다.",
			args: args{
				ctx:          context.Background(),
				UserId:       2000,
				ProjectType:  domain.LOAN,
				LinkedStatus: domain.ON_VA_ACCOUNT,
			},
			want:    domain.UserVirtualAccount{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := userVirtualAccountMysqlRepository{
				dbFactory: dbFactoryMock{},
			}

			got, err := repo.GetByUserIdAndLinkedStatus(tt.args.ctx, tt.args.UserId, tt.args.ProjectType, tt.args.LinkedStatus)

			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})
			assert.Equalf(t, tt.wantErr, err != nil, "userVirtualAccountMysqlRepository.GetByUserIdAndLinkedStatus() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, true, cmp.Equal(tt.want, got, ignoreTime), "userVirtualAccountMysqlRepository.GetByUserIdAndLinkedStatus() = %v, want %v", got, tt.want)
		})
	}
}

func Test_userVirtualAccountMysqlRepository_Update(t *testing.T) {
	type args struct {
		ctx            context.Context
		virtualAccount domain.UserVirtualAccount
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "(성공) 가상 계좌 정보 업데이트",
			args: args{
				ctx: context.Background(),
				virtualAccount: domain.UserVirtualAccount{
					Id:           1000,
					UserId:       2000,
					ProjectType:  domain.EQUITY,
					LinkedStatus: domain.ON_VA_ACCOUNT,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := userVirtualAccountMysqlRepository{
				dbFactory: dbFactoryMock{},
			}

			err := c.Update(tt.args.ctx, tt.args.virtualAccount)

			assert.Equalf(t, tt.wantErr, err != nil, "userVirtualAccountMysqlRepository.Update() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}
