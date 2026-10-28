package mysql

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"

	"github.com/stretchr/testify/assert"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

func Test_investmentWaitHistoryMysqlRepository_Create(t *testing.T) {
	type args struct {
		ctx                   context.Context
		investmentWaitHistory domain.InvestmentWaitHistory
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "InvestmentWaitHistory 모델 정상생성",
			args: args{
				ctx: context.Background(),
				investmentWaitHistory: domain.InvestmentWaitHistory{
					InvestmentWaitId: 1001,
					SubstanceType:    string(domain.INVESTMENT_WAIT_HISTORY_STATUS),
					Name:             "둘리",
					Email:            "dool2@email.com",
					PrevStatus:       domain.DEPOSIT_REQUEST,
					CurrStatus:       domain.PROPOSAL,
					ProcessMethod:    domain.USER_HISTORY_PROCESS_MEMBER,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := investmentWaitHistoryMysqlRepository{
				dbFactory: dbFactoryMock{},
			}

			err := repo.Create(tt.args.ctx, tt.args.investmentWaitHistory)
			assert.Equalf(t, tt.wantErr, err != nil, "Create() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_investmentWaitHistoryMysqlRepository_GetAll(t *testing.T) {
	type args struct {
		ctx     context.Context
		options []cdomain.InvestmentWaitHistoryQueryOption
	}
	tests := []struct {
		name    string
		args    args
		want    []domain.InvestmentWaitHistory
		wantErr bool
	}{
		{
			name: "InvestmentWaitHistory 전체조회 => 3건이 나오기를 기대합니다.",
			args: args{
				ctx: context.Background(),
				options: []cdomain.InvestmentWaitHistoryQueryOption{
					cdomain.InvestmentWaitHistoryQueryOptionFunc().WithOffset(0),
					cdomain.InvestmentWaitHistoryQueryOptionFunc().WithLimit(100),
				},
			},
			want: []domain.InvestmentWaitHistory{
				mockInvestmentWaitHistory1, mockInvestmentWaitHistory2, mockInvestmentWaitHistory3,
			},
		},
		{
			name: "특정 대기자 ID로 조회 : 2000 => \"홍길동\" 대기자에 대한 히스토리 내역이 조회되기를 기대합니다 (2건)",
			args: args{
				ctx: context.Background(),
				options: []cdomain.InvestmentWaitHistoryQueryOption{
					cdomain.InvestmentWaitHistoryQueryOptionFunc().WithOffset(0),
					cdomain.InvestmentWaitHistoryQueryOptionFunc().WithLimit(100),
					cdomain.InvestmentWaitHistoryQueryOptionFunc().WithInvestmentWaitId(2000),
				},
			},
			want: []domain.InvestmentWaitHistory{
				mockInvestmentWaitHistory1, mockInvestmentWaitHistory2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := investmentWaitHistoryMysqlRepository{
				dbFactory: dbFactoryMock{},
			}

			_, got, err := repo.GetAll(tt.args.ctx, tt.args.options...)

			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			assert.Equalf(t, tt.wantErr, err != nil, "GetAll() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, cmp.Equal(tt.want, got, ignoreTime), true, "GetAll() got = %v, want %v", got, tt.want)
		})
	}
}
