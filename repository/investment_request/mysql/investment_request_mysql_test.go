package mysql

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/stretchr/testify/assert"
)

func Test_investmentRequestMysqlRepository_Create(t *testing.T) {
	type args struct {
		ctx               context.Context
		investmentRequest domain.InvestmentRequest
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    int64
	}{
		{
			name: "InvestmentRequest 모델 생성 => 정상적으로 생성되기를 기대합니다.",
			args: args{
				ctx: context.Background(),
				investmentRequest: domain.InvestmentRequest{
					InvestmentWaitersId: 2000,
					ProjectId:           3000,
					IsExpire:            false,
				},
			},
			wantErr: false,
			want:    1001,
		},
		{
			name: "기존에 이미 있는 InvestmentRequest 모델 생성  : 1000 => 에러가 발생하기를 기대합니다.",
			args: args{
				ctx: context.Background(),
				investmentRequest: domain.InvestmentRequest{
					Id:                  1000,
					InvestmentWaitersId: 2000,
					ProjectId:           3000,
					IsExpire:            false,
				},
			},
			wantErr: true,
			want:    0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := investmentRequestMysqlRepository{
				dbFactory: dbFactoryMock{},
			}

			got, err := repo.Create(tt.args.ctx, tt.args.investmentRequest)
			assert.Equalf(t, tt.wantErr, err != nil, "Create() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want, got, "Create() = %v, want %v", got, tt.want)
		})
	}
}

func Test_investmentRequestMysqlRepository_Update(t *testing.T) {
	type args struct {
		ctx               context.Context
		investmentRequest domain.InvestmentRequest
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "InvestmentRequest 모델에서 IsExpire를 변경합니다 : false => 정상적으로 변경되기를 기대합니다.",
			args: args{
				ctx: context.Background(),
				investmentRequest: domain.InvestmentRequest{
					Id:       1000,
					IsExpire: true,
				},
			},
			wantErr: false,
		},
		{
			name: "조회할 수 없는 Id로 변경합니다 : 에러가 발생하기를 기대합니다.",
			args: args{
				ctx: context.Background(),
				investmentRequest: domain.InvestmentRequest{
					Id:       2000,
					IsExpire: true,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := investmentRequestMysqlRepository{
				dbFactory: dbFactoryMock{},
			}

			err := repo.Update(tt.args.ctx, tt.args.investmentRequest)
			assert.Equalf(t, tt.wantErr, err != nil, "Update() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_investmentRequestMysqlRepository_GetByInvestmentWaitersId(t *testing.T) {
	type args struct {
		ctx                 context.Context
		investmentWaitersId int64
	}
	tests := []struct {
		name    string
		args    args
		want    domain.InvestmentRequest
		wantErr bool
	}{
		{
			name: "InvestmentWaitersId로 조회 : 2000 => 정상적으로 1건이 조회되기를 기대합니다.",
			args: args{
				ctx:                 context.Background(),
				investmentWaitersId: 2000,
			},
			want: domain.InvestmentRequest{
				Id:                  1000,
				InvestmentWaitersId: 2000,
				ProjectId:           3000,
				IsExpire:            false,
			},
			wantErr: false,
		},
		{
			name: "조회할 수 없는 InvestmentWaitersId로 조회 : 에러가 발생하기를 기대합니다.",
			args: args{
				ctx:                 context.Background(),
				investmentWaitersId: 3000,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := investmentRequestMysqlRepository{
				dbFactory: dbFactoryMock{},
			}

			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			got, err := repo.GetByInvestmentWaitersId(tt.args.ctx, tt.args.investmentWaitersId)
			assert.Equalf(t, tt.wantErr, err != nil, "GetByInvestmentWaitersId() error = %v, wantErr %v", err, tt.wantErr)
			assert.Truef(t, cmp.Equal(tt.want, got, ignoreTime), "GetByInvestmentWaitersId() = %v, want %v", got, tt.want)
		})
	}
}
