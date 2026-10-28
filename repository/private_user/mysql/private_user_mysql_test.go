package mysql

import (
	"context"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

func Test_privateUserMysqlRepository_GetById(t *testing.T) {
	type args struct {
		ctx context.Context
		id  int64
	}
	tests := []struct {
		name    string
		args    args
		want    domain.PrivateUser
		wantErr bool
	}{
		{
			name: "(성공) 특정 유저 조회 {Id: 1000} => 결과 1건 있음",
			args: args{
				ctx: context.Background(),
				id:  1000,
			},
			want: domain.PrivateUser{
				User:           mockUser1,
				PrivateProfile: mockPrivateProfile1,
			},
			wantErr: false,
		},
		{
			name: "(성공) 특정 유저 조회 {Id: 2000} => 결과 1건 있음",
			args: args{
				ctx: context.Background(),
				id:  2000,
			},
			want: domain.PrivateUser{
				User:           mockUser2,
				PrivateProfile: mockPrivateProfile2,
			},
			wantErr: false,
		},
		{
			name: "(실패) 특정 유저 조회 {Id: 1001} => 결과 0건 있음",
			args: args{
				ctx: context.Background(),
				id:  1001,
			},
			want:    domain.PrivateUser{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := privateUserMysqlRepository{
				dbFactory: dbFactoryMock{},
			}
			got, err := repo.GetById(tt.args.ctx, tt.args.id)

			// jsonutils.JSONTime{}
			ignoreTime := cmp.FilterPath(func(path cmp.Path) bool {
				return path.Last().Type() == reflect.TypeOf(jsonutils.JSONTime{})
			}, cmp.Ignore())
			ignoreFields := cmpopts.IgnoreFields(domain.PrivateProfile{}, "Id")

			assert.Equalf(t, tt.wantErr, err != nil, "privateUserMysqlRepository.GetById() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, cmp.Equal(tt.want, got, ignoreTime, ignoreFields), true, "privateUserMysqlRepository.GetById() = %+v, want %+v", got, tt.want)
		})
	}
}

func Test_privateUserMysqlRepository_GetByEmail(t *testing.T) {
	type args struct {
		ctx   context.Context
		email string
	}
	tests := []struct {
		name    string
		args    args
		want    domain.PrivateUser
		wantErr bool
	}{
		{
			name: "(성공)특정 유저를 조회 {Email: gildong@email.com} => 결과 1건 있음",
			args: args{
				ctx:   context.Background(),
				email: mockUser1.Email,
			},
			want: domain.PrivateUser{
				User:           mockUser1,
				PrivateProfile: mockPrivateProfile1,
			},
			wantErr: false,
		},
		{
			name: "(성공) 특정 유저를 조회 {Email: dool2@email.com} => 결과 0건 있음",
			args: args{
				ctx:   context.Background(),
				email: "dool2@email.com",
			},
			want:    domain.PrivateUser{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := privateUserMysqlRepository{
				dbFactory: dbFactoryMock{},
			}
			got, err := repo.GetByEmail(tt.args.ctx, tt.args.email)

			// jsonutils.JSONTime{}
			ignoreTime := cmp.FilterPath(func(path cmp.Path) bool {
				return path.Last().Type() == reflect.TypeOf(jsonutils.JSONTime{})
			}, cmp.Ignore())
			ignoreFields := cmpopts.IgnoreFields(domain.PrivateProfile{}, "Id")

			assert.Equalf(t, tt.wantErr, err != nil, "privateUserMysqlRepository.GetByEmail() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, cmp.Equal(tt.want, got, ignoreTime, ignoreFields), true, "privateUserMysqlRepository.GetByEmail() = %v, want %v", got, tt.want)
		})
	}
}
