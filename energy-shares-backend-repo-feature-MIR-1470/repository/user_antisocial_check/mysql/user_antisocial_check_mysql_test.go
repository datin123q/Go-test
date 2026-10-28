package mysql

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

func Test_userAntisocialCheckMysqlRepository_Create(t *testing.T) {
	type args struct {
		ctx                 context.Context
		userAntisocialCheck domain.UserAntisocialCheck
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "(성공) 유저 생성",
			args: args{
				ctx:                 context.Background(),
				userAntisocialCheck: mockUserAntisocialCheck2,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := userAntisocialCheckRepository{
				dbFactory: dbFactoryMock{},
			}
			err := repo.Create(tt.args.ctx, tt.args.userAntisocialCheck)

			assert.Equalf(t, tt.wantErr, err != nil, "userMysqlRepository.Create() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_userAntisocialCheckMysqlRepository_GetByUserId(t *testing.T) {
	type args struct {
		ctx    context.Context
		userId int64
	}
	tests := []struct {
		name    string
		args    args
		want    domain.UserAntisocialCheck
		wantErr bool
	}{
		{
			name: "AntisocialCheckFile 모델 조회 (User:2000) -> 정상적으로 조회되기를 기대합니다.",
			args: args{
				ctx:    context.Background(),
				userId: mockUserAntisocialCheck1.CheckUserId,
			},
			want:    mockUserAntisocialCheck1,
			wantErr: false,
		},
		{
			name: "AntisocialCheckFile 모델 조회 (User:2001) -> 존재하지 않는 User를 조회하므로 에러가 발생합니다.",
			args: args{
				ctx:    context.Background(),
				userId: 2001,
			},
			want:    domain.UserAntisocialCheck{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := userAntisocialCheckRepository{
				dbFactory: dbFactoryMock{},
			}

			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})
			ignoreFields := cmpopts.IgnoreFields(domain.UserAntisocialCheck{}, "Id")

			got, err := repo.GetByUserId(tt.args.ctx, tt.args.userId)
			assert.Equalf(t, tt.wantErr, err != nil, "userAntisocialCheckRepository.GetAll() error = %v, wantErr %v", err, tt.wantErr)
			// t.Log(cmp.Diff(tt.want, got, ignoreTime, ignoreFields))
			assert.Equalf(t, true, cmp.Equal(tt.want, got, ignoreTime, ignoreFields), "userAntisocialCheckRepository.GetAll() got = %v, want %v", got, tt.want)
		})
	}
}
