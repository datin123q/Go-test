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

func Test_contractMysqlRepository_Create(t *testing.T) {
	type args struct {
		ctx      context.Context
		contract domain.Contract
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "Contract作成 / mockデータはすでに2つ作成されている。新しいContract作成時、idは3である必要がある。",
			args: args{
				ctx: context.Background(),
				contract: domain.Contract{
					ShareId:     "SHARE_ID",
					ShareQty:    4,
					ShareAmount: 1000,
					Status:      domain.CONTRACT_SIGNED,
					ProjectId:   1234,
					UserName:    "Jinyong Kim",
				},
			},
			want: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := contractMysqlRepository{
				dbFactory: dbFactoryMock{},
			}
			got, err := c.Create(tt.args.ctx, tt.args.contract)

			assert.NoErrorf(t, err, "contractMysqlRepository.Create() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want, got, "contractMysqlRepository.Create() = %v, want %v", got, tt.want)
		})
	}
}

func Test_contractMysqlRepository_GetAll(t *testing.T) {
	type args struct {
		ctx     context.Context
		options []cdomain.ContractQueryOption
	}
	tests := []struct {
		name    string
		args    args
		want    domain.Contracts
		wantErr bool
	}{
		{
			name: "Contractリスト照会",
			args: args{
				ctx: context.Background(),
				options: []cdomain.ContractQueryOption{
					cdomain.ContractQueryOptionFunc().WithLimit(100),
					cdomain.ContractQueryOptionFunc().WithOffset(0),
				},
			},
			want: domain.Contracts{
				{
					Id:          1,
					ShareId:     "SHARE_ID1",
					ShareQty:    4,
					ShareAmount: 1000,
					Status:      domain.CONTRACT_SIGNED,
					ProjectId:   1234,
					UserName:    "Jinyong Kim",
				},

				{
					Id:          2,
					ShareId:     "SHARE_ID2",
					ShareQty:    5,
					ShareAmount: 2000,
					Status:      domain.CONTRACT_SIGNED,
					ProjectId:   1235,
					UserName:    "Jinyong Kim",
				},
			},
		},
		{
			name: "SHARE_ID2を持つContractリスト照会",
			args: args{
				ctx: context.Background(),
				options: []cdomain.ContractQueryOption{
					cdomain.ContractQueryOptionFunc().WithLimit(100),
					cdomain.ContractQueryOptionFunc().WithOffset(0),
					cdomain.ContractQueryOptionFunc().WithShareId("SHARE_ID2"),
				},
			},
			want: domain.Contracts{
				{
					Id:          2,
					ShareId:     "SHARE_ID2",
					ShareQty:    5,
					ShareAmount: 2000,
					Status:      domain.CONTRACT_SIGNED,
					ProjectId:   1235,
					UserName:    "Jinyong Kim",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := contractMysqlRepository{
				dbFactory: dbFactoryMock{},
			}
			got, err := c.GetAll(tt.args.ctx, tt.args.options...)

			// go-cmpオプションを使用してCreated、Updated、Deletedフィールドを比較しないように設定
			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			assert.NoErrorf(t, err, "contractMysqlRepository.GetAll() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, cmp.Equal(tt.want, got, ignoreTime), true, "contractMysqlRepository.GetAll() = %v, want %v", got, tt.want)
		})
	}
}
