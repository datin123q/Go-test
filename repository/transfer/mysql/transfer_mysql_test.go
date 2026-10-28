package mysql

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func Test_transferRepository_GetAll(t *testing.T) {
	type args struct {
		ctx     context.Context
		options []cdomain.TransferQueryOption
	}

	tests := []struct {
		name    string
		args    args
		want    []domain.Transfer
		wantErr bool
	}{
		{
			name: "transfer 전체 조회 - 조건 없음",
			args: args{
				ctx: context.Background(),
				options: []cdomain.TransferQueryOption{
					cdomain.TransferQueryOptionFunc().WithSortBy("id"),
					cdomain.TransferQueryOptionFunc().WithOrder("asc"),
					cdomain.TransferQueryOptionFunc().WithOffset(0),
					cdomain.TransferQueryOptionFunc().WithLimit(10),
				},
			},
			want: []domain.Transfer{
				mockTransfer1,
				mockTransfer2,
				mockTransfer3,
			},
			wantErr: false,
		},
		{
			name: "transfer 전체 조회 - withdrawal_id",
			args: args{
				ctx: context.Background(),
				options: []cdomain.TransferQueryOption{
					cdomain.TransferQueryOptionFunc().WithWithdrawalId(&mockWithdrawalTransferMapping1.WithdrawalId),
					cdomain.TransferQueryOptionFunc().WithSortBy("id"),
					cdomain.TransferQueryOptionFunc().WithOrder("desc"),
				},
			},
			want: []domain.Transfer{
				mockTransfer1,
			},
			wantErr: false,
		},
		{
			name: "transfer 전체 조회 - request_id",
			args: args{
				ctx: context.Background(),
				options: []cdomain.TransferQueryOption{
					cdomain.TransferQueryOptionFunc().WithRequestId(lo.ToPtr(int64(1))),
				},
			},
			want: []domain.Transfer{
				mockTransfer1,
				mockTransfer3,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := transferMysqlRepository{
				dbFactory: dbFactoryMock{},
			}

			_, got, err := repo.GetAll(tt.args.ctx, tt.args.options...)

			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			assert.Equalf(t, tt.wantErr, err != nil, "GetAll() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, true, cmp.Equal(tt.want, got, ignoreTime), "GetAll() diff = %v", cmp.Diff(tt.want, got, ignoreTime))
		})
	}
}
