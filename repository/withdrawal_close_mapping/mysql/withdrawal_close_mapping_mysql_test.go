package mysql

import (
	"context"
	"testing"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/stretchr/testify/assert"
)

func Test_withdrawalCloseMapping_Create(t *testing.T) {
	type fields struct {
		dbFactory factory.DBFactory
	}
	type args struct {
		ctx                    context.Context
		withdrawalCloseMapping domain.WithdrawalCloseMapping
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "withdrawal close mapping 생성",
			fields: fields{
				dbFactory: dbFactoryMock{},
			},
			args: args{
				ctx: context.TODO(),
				withdrawalCloseMapping: domain.WithdrawalCloseMapping{
					CloseId:      100,
					WithdrawalId: 100,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := withdrawalCloseMappingMysqlRepository{
				dbFactory: tt.fields.dbFactory,
			}

			_, err := r.Create(tt.args.ctx, tt.args.withdrawalCloseMapping)
			assert.Equalf(t, tt.wantErr, err != nil, "withdrawalCloseMappingMysqlRepository.Create() error = %v, wantErr %v", err, tt.wantErr)

		})
	}
}

func Test_withdrawalCloseMapping_CreateMany(t *testing.T) {
	type fields struct {
		dbFactory factory.DBFactory
	}
	type args struct {
		ctx                     context.Context
		withdrawalCloseMappings []domain.WithdrawalCloseMapping
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "withdrawal close mapping 다중 생성",
			fields: fields{
				dbFactory: dbFactoryMock{},
			},
			args: args{
				ctx: context.TODO(),
				withdrawalCloseMappings: []domain.WithdrawalCloseMapping{
					{
						CloseId:      100,
						WithdrawalId: 100,
					},
					{
						CloseId:      100,
						WithdrawalId: 101,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := withdrawalCloseMappingMysqlRepository{
				dbFactory: tt.fields.dbFactory,
			}

			err := r.CreateMany(tt.args.ctx, tt.args.withdrawalCloseMappings)
			assert.Equalf(t, tt.wantErr, err != nil, "withdrawalCloseMappingMysqlRepository.CreateMany() error = %v, wantErr %v", err, tt.wantErr)

		})
	}
}

func Test_withdrawalCloseMapping_GetAll(t *testing.T) {
	type fields struct {
		dbFactory factory.DBFactory
	}
	type args struct {
		ctx     context.Context
		options []cdomain.WithdrawalCloseMappingQueryOption
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []domain.WithdrawalCloseMapping
		wantErr bool
	}{
		{
			name: "withdrawal close mapping 상세 조회 - 조건 없음",
			fields: fields{
				dbFactory: dbFactoryMock{},
			},
			args: args{
				ctx: context.TODO(),
				options: []cdomain.WithdrawalCloseMappingQueryOption{
					cdomain.WithdrawalCloseMappingQueryOptionFunc().WithSortBy("id"),
					cdomain.WithdrawalCloseMappingQueryOptionFunc().WithOrder("desc"),
					cdomain.WithdrawalCloseMappingQueryOptionFunc().WithLimit(5),
					cdomain.WithdrawalCloseMappingQueryOptionFunc().WithOffset(0),
				},
			},
			want: []domain.WithdrawalCloseMapping{
				mockWithdrawalCloseMapping3,
				mockWithdrawalCloseMapping2,
				mockWithdrawalCloseMapping1,
			},
			wantErr: false,
		},
		{
			name: "withdrawal close mapping 상세 조회 - WithdrawalId 목록",
			fields: fields{
				dbFactory: dbFactoryMock{},
			},
			args: args{
				ctx: context.TODO(),
				options: []cdomain.WithdrawalCloseMappingQueryOption{
					cdomain.WithdrawalCloseMappingQueryOptionFunc().WithSortBy("id"),
					cdomain.WithdrawalCloseMappingQueryOptionFunc().WithOrder("desc"),
					cdomain.WithdrawalCloseMappingQueryOptionFunc().WithLimit(5),
					cdomain.WithdrawalCloseMappingQueryOptionFunc().WithOffset(0),
					cdomain.WithdrawalCloseMappingQueryOptionFunc().WithWithdrawalIds([]int64{2, 3}),
				},
			},
			want: []domain.WithdrawalCloseMapping{
				mockWithdrawalCloseMapping3,
				mockWithdrawalCloseMapping2,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := withdrawalCloseMappingMysqlRepository{
				dbFactory: tt.fields.dbFactory,
			}

			_, got, err := r.GetAll(tt.args.ctx, tt.args.options...)
			assert.Equalf(t, tt.wantErr, err != nil, "withdrawalCloseMappingMysqlRepository.GetAll() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want, got, "withdrawalCloseMappingMysqlRepository.GetAll() = %v, want %v", got, tt.want)

		})
	}
}
