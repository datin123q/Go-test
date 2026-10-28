package mysql

import (
	"context"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

func Test_transactionInfoHistoryMysqlRepository_Create(t *testing.T) {
	type args struct {
		ctx                    context.Context
		transactionInfoHistory domain.TransactionInfoHistory
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "TransactionInfoHistory 모델 생성 => 정상적으로 생성되기를 기대합니다.",
			args: args{
				ctx: context.Background(),
				transactionInfoHistory: domain.TransactionInfoHistory{
					ShareId:       "SHAREID_1",
					SubstanceType: string(domain.INVESTMENT_WAIT_HISTORY_STATUS),
					ProcessMethod: domain.USER_HISTORY_PROCESS_MEMBER,
					Name:          "둘리",
					Email:         "dool2@email.com",
				},
			},
			wantErr: false,
		},
		{
			name: "기존에 이미 있는 TransactionInfoHistory 모델 생성  : 1000 => 에러가 발생하기를 기대합니다.",
			args: args{
				ctx:                    context.Background(),
				transactionInfoHistory: mockTransactionInfoHistory1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := transactionInfoHistoryMysqlRepository{
				dbFactory: dbFactoryMock{},
			}

			err := repo.Create(tt.args.ctx, tt.args.transactionInfoHistory)
			assert.Equalf(t, tt.wantErr, err != nil, "Create() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_transactionInfoHistoryMysqlRepository_GetAll(t *testing.T) {
	type args struct {
		ctx     context.Context
		options []cdomain.TransactionInfoHistoryQueryOption
	}
	tests := []struct {
		name    string
		args    args
		want    []domain.TransactionInfoHistory
		wantErr bool
	}{
		{
			name: "TransactionInfoHistory 모델 조회 => 정상적으로 조회되기를 기대합니다.",
			args: args{
				ctx: context.Background(),
				options: []cdomain.TransactionInfoHistoryQueryOption{
					cdomain.TransactionInfoHistoryQueryOptionFunc().WithShareId("SHAREID_1"),
				},
			},
			want:    []domain.TransactionInfoHistory{mockTransactionInfoHistory1},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := transactionInfoHistoryMysqlRepository{
				dbFactory: dbFactoryMock{},
			}

			_, got, err := repo.GetAll(tt.args.ctx, tt.args.options...)

			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			assert.Equalf(t, tt.wantErr, err != nil, "GetAll() error = %v, wantErr %v", err, tt.wantErr)
			assert.Truef(t, cmp.Equal(tt.want, got, ignoreTime), "GetAll() got = %v, want %v", got, tt.want)
		})
	}
}
