package mysql

import (
	"context"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_corporateShareholderMysqlRepository_Create(t *testing.T) {
	type args struct {
		ctx                  context.Context
		corporateShareholder domain.CorporateShareholder
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "CorporateShareholder 생성",
			args: args{
				ctx: context.Background(),
				corporateShareholder: domain.CorporateShareholder{
					UserId: 1000,
					Uid:    "UID1",
				},
			},
			want:    2,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := corporateShareholderMysqlRepository{
				dbFactory: dbFactoryMock{},
			}
			got, err := c.Create(tt.args.ctx, tt.args.corporateShareholder)
			assert.Equalf(t, tt.wantErr, err != nil, "corporateShareholderMysqlRepository.Create() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want, got, "corporateShareholderMysqlRepository.Create() = %v, want %v", got, tt.want)
		})
	}
}

func Test_corporateShareholderMysqlRepository_GetAll(t *testing.T) {
	type args struct {
		ctx     context.Context
		options []cdomain.CorporateShareholderOption
	}
	tests := []struct {
		name    string
		args    args
		want    []domain.CorporateShareholder
		wantErr bool
	}{
		{
			name: "CorporateShareholder 특정유저(1000)로 조회",
			args: args{
				ctx: context.Background(),
				options: []cdomain.CorporateShareholderOption{
					cdomain.CorporateShareholderQueryOptionFields{}.WithUserId(1000),
				},
			},
			want:    []domain.CorporateShareholder{mockCorporateShareholder1},
			wantErr: false,
		},
		{
			name: "CorporateShareholder 특정유저(1001)로 조회 => 아무것도 조회되지 않음",
			args: args{
				ctx: context.Background(),
				options: []cdomain.CorporateShareholderOption{
					cdomain.CorporateShareholderQueryOptionFields{}.WithUserId(1001),
				},
			},
			want:    []domain.CorporateShareholder{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := corporateShareholderMysqlRepository{
				dbFactory: dbFactoryMock{},
			}
			_, got, err := c.GetAll(tt.args.ctx, tt.args.options...)

			assert.Equalf(t, tt.wantErr, err != nil, "corporateShareholderMysqlRepository.GetAll() error = %v, wantErr %v", err, tt.wantErr)
			assert.Truef(t, cmp.Equal(tt.want, got, cmpopts.IgnoreTypes(jsonutils.JSONTime{})), "corporateShareholderMysqlRepository.GetAll() = %v, want %v", got, tt.want)
		})
	}

}

func Test_corporateShareholderMysqlRepository_DeleteByUserId(t *testing.T) {
	type args struct {
		ctx    context.Context
		userId int64
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "유저(1000) CorporateShareholder 삭제",
			args: args{
				ctx:    context.Background(),
				userId: 1000,
			},
			wantErr: false,
		},
		{
			name: "유저값이 비어있을 때 CorporateShareholder 삭제요청",
			args: args{
				ctx: context.Background(),
			},
			wantErr: true,
		},
		{
			name: "없는 유저(1001)로 CorporateShareholder 삭제요청",
			args: args{
				ctx:    context.Background(),
				userId: 1001,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := corporateShareholderMysqlRepository{
				dbFactory: dbFactoryMock{},
			}
			err := c.DeleteByUserId(tt.args.ctx, tt.args.userId)
			assert.Equalf(t, tt.wantErr, err != nil, "corporateShareholderMysqlRepository.DeleteByUserId() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}
