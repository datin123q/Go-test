package mysql

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

func Test_corporateProfileMysqlRepository_Create(t *testing.T) {
	type args struct {
		ctx              context.Context
		corporateProfile domain.CorporateProfile
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "CorporateProfile 생성",
			args: args{
				ctx: context.Background(),
				corporateProfile: domain.CorporateProfile{
					UserId:                1003,
					CorpNo:                "1234",
					CorpType:              "フィルム暖房製造",
					ManagerFamilyName:     lo.ToPtr("Lee"),
					ManagerFirstName:      lo.ToPtr("Soojin"),
					ManagerFamilyNameKana: lo.ToPtr("リ"),
					ManagerFirstNameKana:  lo.ToPtr("スジン"),
					ManagerEmail:          lo.ToPtr("soojin.lee@example.com"),
					ManagerMobilePhone:    lo.ToPtr("010-1122-3344"),
					Department:            lo.ToPtr("Marketing"),
					Position:              lo.ToPtr("Senior Manager"),
				},
			},
			want:    4,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := corporateProfileMysqlRepository{
				dbFactory: dbFactoryMock{},
			}
			got, err := c.Create(tt.args.ctx, tt.args.corporateProfile)
			assert.Equalf(t, tt.wantErr, err != nil, "corporateProfileMysqlRepository.Create() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want, got, "corporateProfileMysqlRepository.Create() = %v, want %v", got, tt.want)
		})
	}
}

func Test_corporateProfileMysqlRepository_GetById(t *testing.T) {
	type args struct {
		ctx context.Context
		id  int64
	}
	tests := []struct {
		name    string
		args    args
		want    domain.CorporateProfile
		wantErr bool
	}{
		{
			name: "CorporateProfile 조회",
			args: args{
				ctx: context.Background(),
				id:  1,
			},
			want: mockCorporateProfile1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := corporateProfileMysqlRepository{
				dbFactory: dbFactoryMock{},
			}
			got, err := c.GetById(tt.args.ctx, tt.args.id)
			assert.Equalf(t, tt.wantErr, err != nil, "corporateProfileMysqlRepository.GetById() error = %v, wantErr %v", err, tt.wantErr)
			assert.Truef(t, cmp.Equal(tt.want, got, cmpopts.IgnoreTypes(jsonutils.JSONTime{})), "corporateProfileMysqlRepository.GetById() = %v, want %v", got, tt.want)
		})
	}
}

func Test_corporateProfileMysqlRepository_GetByCorpNo(t *testing.T) {
	type args struct {
		ctx    context.Context
		corpNo string
	}
	tests := []struct {
		name    string
		args    args
		want    domain.CorporateProfile
		wantErr bool
	}{
		{
			name: "회사번호로(1234567890) CorporateProfile 조회",
			args: args{
				ctx:    context.Background(),
				corpNo: "1234567890",
			},
			want:    mockCorporateProfile1,
			wantErr: false,
		},
		{
			name: "회사번호로(1234) CorporateProfile 조회 => 번호를 찾을 수 없어서 에러가 발생",
			args: args{
				ctx:    context.Background(),
				corpNo: "1234",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := corporateProfileMysqlRepository{
				dbFactory: dbFactoryMock{},
			}
			got, err := c.GetByCorpNo(tt.args.ctx, tt.args.corpNo)
			assert.Equalf(t, tt.wantErr, err != nil, "corporateProfileMysqlRepository.GetByCorpNo() error = %v, wantErr %v", err, tt.wantErr)
			assert.Truef(t, cmp.Equal(tt.want, got, cmpopts.IgnoreTypes(jsonutils.JSONTime{})), "corporateProfileMysqlRepository.GetByCorpNo() = %v, want %v", got, tt.want)
		})
	}
}

func Test_corporateProfileMysqlRepository_Update(t *testing.T) {
	type args struct {
		ctx              context.Context
		corporateProfile domain.CorporateProfile
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "CorporateProfile 정상 수정",
			args: args{
				ctx: context.Background(),
				corporateProfile: domain.CorporateProfile{
					Id:               3,
					CorpType:         "new type",
					IsManagerSmsCert: true,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := corporateProfileMysqlRepository{
				dbFactory: dbFactoryMock{},
			}

			err := repo.Update(tt.args.ctx, tt.args.corporateProfile)
			assert.Equalf(t, tt.wantErr, err != nil, "Update() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_corporateProfileMysqlRepository_GetByUserId(t *testing.T) {
	type args struct {
		ctx    context.Context
		userId int64
	}
	tests := []struct {
		name    string
		args    args
		want    domain.CorporateProfile
		wantErr bool
	}{
		{
			name: "사용자ID로(1000) CorporateProfile 조회",
			args: args{
				ctx:    context.Background(),
				userId: 1000,
			},
			want:    mockCorporateProfile1,
			wantErr: false,
		},
		{
			name: "사용자ID로(1004) CorporateProfile 조회 => 조회되지 않음",
			args: args{
				ctx:    context.Background(),
				userId: 1004,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := corporateProfileMysqlRepository{
				dbFactory: dbFactoryMock{},
			}
			got, err := c.GetByUserId(tt.args.ctx, tt.args.userId)
			assert.Equalf(t, tt.wantErr, err != nil, "corporateProfileMysqlRepository.GetByUserId() error = %v, wantErr %v", err, tt.wantErr)
			assert.Truef(t, cmp.Equal(tt.want, got, cmpopts.IgnoreTypes(jsonutils.JSONTime{})), "corporateProfileMysqlRepository.GetByUserId() = %v, want %v", got, tt.want)
		})
	}
}
