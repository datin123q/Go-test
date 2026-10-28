package mysql

import (
	"context"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

func Test_privateProfileMysqlRepository_Create(t *testing.T) {
	type args struct {
		ctx            context.Context
		privateProfile domain.PrivateProfile
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "PrivateProfile 모델 생성 => 정상적으로 생성되기를 기대합니다.",
			args: args{
				ctx:            context.Background(),
				privateProfile: mockPrivateProfile1,
			},
			wantErr: false,
		},
		{
			name: "PrivateProfile 모델 생성: UserID not unique => 에러가 발생하기를 기대합니다.",
			args: args{
				ctx:            context.Background(),
				privateProfile: mockPrivateProfile2,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := privateProfileMysqlRepository{
				dbFactory: dbFactoryMock{},
			}

			err := repo.Create(tt.args.ctx, tt.args.privateProfile)
			assert.Equalf(t, tt.wantErr, err != nil, "Create() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_privateProfileMysqlRepository_GetById(t *testing.T) {
	type args struct {
		ctx context.Context
		id  int64
	}
	tests := []struct {
		name    string
		args    args
		want    domain.PrivateProfile
		wantErr bool
	}{
		{
			name: "(성공) 자동증가되는 ID 값(1000)으로 조회 => 결과 1건 있음",
			args: args{
				ctx: context.Background(),
				id:  1000,
			},
			want:    mockPrivateProfile5,
			wantErr: false,
		},
		{
			name: "(실패) 특정 유저 조회 {Id: 1001} => 결과 0건 있음",
			args: args{
				ctx: context.Background(),
				id:  2001,
			},
			want:    domain.PrivateProfile{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := privateProfileMysqlRepository{
				dbFactory: dbFactoryMock{},
			}
			got, err := repo.GetById(tt.args.ctx, tt.args.id)

			// jsonutils.JSONTime{}
			ignoreTime := cmp.FilterPath(func(path cmp.Path) bool {
				return path.Last().Type() == reflect.TypeOf(jsonutils.JSONTime{})
			}, cmp.Ignore())
			ignoreFields := cmpopts.IgnoreFields(domain.PrivateProfile{}, "Id")

			assert.Equalf(t, tt.wantErr, err != nil, "privateProfileMysqlRepository.GetById() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, cmp.Equal(tt.want, got, ignoreTime, ignoreFields), true, "privateProfileMysqlRepository.GetById() = %+v, want %+v", got, tt.want)
		})
	}
}

func Test_privateProfileMysqlRepository_GetByUserId(t *testing.T) {
	type args struct {
		ctx    context.Context
		userId int64
	}
	tests := []struct {
		name    string
		args    args
		want    domain.PrivateProfile
		wantErr bool
	}{
		{
			name: "(성공) 특정 유저 조회 {UserId: 2003} => 결과 1건 있음",
			args: args{
				ctx:    context.Background(),
				userId: 2003,
			},
			want:    mockPrivateProfile5,
			wantErr: false,
		},
		{
			name: "(실패) 특정 유저 조회 {UserId: 2002} => 결과 0건 있음",
			args: args{
				ctx:    context.Background(),
				userId: 2002,
			},
			want:    domain.PrivateProfile{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := privateProfileMysqlRepository{
				dbFactory: dbFactoryMock{},
			}
			got, err := repo.GetByUserId(tt.args.ctx, tt.args.userId)

			// jsonutils.JSONTime{}
			ignoreTime := cmp.FilterPath(func(path cmp.Path) bool {
				return path.Last().Type() == reflect.TypeOf(jsonutils.JSONTime{})
			}, cmp.Ignore())
			ignoreFields := cmpopts.IgnoreFields(domain.PrivateProfile{}, "Id")

			assert.Equalf(t, tt.wantErr, err != nil, "privateProfileMysqlRepository.GetByUserId() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, cmp.Equal(tt.want, got, ignoreTime, ignoreFields), true, "privateProfileMysqlRepository.GetByUserId() = %+v, want %+v", got, tt.want)
		})
	}
}

func Test_privateProfileMysqlRepository_Update(t *testing.T) {
	type args struct {
		ctx            context.Context
		privateProfile domain.PrivateProfile
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "PrivateProfile 정상 수정",
			args: args{
				ctx: context.Background(),
				privateProfile: domain.PrivateProfile{
					Id:               1000,
					UserId:           2003,
					Nationality:      domain.JP,
					IsProInvestAsset: lo.ToPtr(false),
					Gender:           domain.MALE,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := privateProfileMysqlRepository{
				dbFactory: dbFactoryMock{},
			}

			err := repo.Update(tt.args.ctx, tt.args.privateProfile)
			assert.Equalf(t, tt.wantErr, err != nil, "Update() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}
