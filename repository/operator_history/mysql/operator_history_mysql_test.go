package mysql

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

func Test_operatorHistoryMysqlRepository_GetAll(t *testing.T) {
	type args struct {
		ctx     context.Context
		options []cdomain.OperatorHistoryQueryOption
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		want1   []domain.OperatorHistory
		wantErr bool
	}{
		{
			name: "OperatorHistoryモデル照会 => 正常に照会されることが期待されます。",
			args: args{
				ctx: context.Background(),
				options: []cdomain.OperatorHistoryQueryOption{
					cdomain.OperatorHistoryQueryOptionFunc().WithState((*string)(lo.ToPtr(domain.LOOKUP))),
				},
			},
			want: 1,
			want1: []domain.OperatorHistory{
				{
					Id:     1000,
					UserId: 2000,
					Name:   "コ・ギルドン",
					Email:  "kildong@email.com",
					Menu:   "menu.manage_product.funding_request_detail",
					Ip:     "127.0.0.1",
					Title:  "テストページ",
					State:  domain.LOOKUP,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := operatorHistoryMysqlRepository{
				dbFactory: dbFactoryMock{},
			}
			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			got, got1, err := repo.GetAll(tt.args.ctx, tt.args.options...)
			assert.Equal(t, tt.want, got)
			assert.Equalf(t, tt.wantErr, err != nil,
				"operatorHistoryMysqlRepository.GetAll() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, true, cmp.Equal(tt.want1, got1, ignoreTime),
				"operatorHistoryMysqlRepository.GetAll() got = %v, want %v", got1, tt.want1)
		})
	}
}

func Test_operatorHistoryMysqlRepository_Create(t *testing.T) {
	type args struct {
		ctx             context.Context
		operatorHistory domain.OperatorHistory
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "OperatorHistoryモデル作成 => 正常に作成されることが期待されます。",
			args: args{
				ctx: context.Background(),
				operatorHistory: domain.OperatorHistory{
					UserId: 2000,
					Name:   "コ・ギルドン",
					Email:  "kildong@email.com",
					Menu:   "menu.manage_product.funding_request_detail",
					Ip:     "127.0.0.1",
					Title:  "テストページ",
					State:  domain.LOOKUP,
				},
			},
			wantErr: false,
		},
		{
			name: "既に存在するOperatorHistoryモデル作成 : 1000 => エラーが発生することが期待されます。",
			args: args{
				ctx: context.Background(),
				operatorHistory: domain.OperatorHistory{
					Id:     1000,
					UserId: 2000,
					Name:   "コ・ギルドン",
					Email:  "kildong@email.com",
					Menu:   "menu.manage_product.funding_request_detail",
					Ip:     "127.0.0.1",
					Title:  "テストページ",
					State:  domain.LOOKUP,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := operatorHistoryMysqlRepository{
				dbFactory: dbFactoryMock{},
			}

			err := repo.Create(tt.args.ctx, tt.args.operatorHistory)
			assert.Equalf(t, tt.wantErr, err != nil, "Create() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}
