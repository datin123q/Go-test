package mysql

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/stretchr/testify/assert"
)

func Test_withdrawalClose_Create(t *testing.T) {
	type args struct {
		ctx             context.Context
		withdrawalClose domain.WithdrawalClose
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "withdrawal close 생성",
			args: args{
				ctx: context.TODO(),
				withdrawalClose: domain.WithdrawalClose{
					CloserId:  int64(1),
					CloseDate: jsonutils.JSONTime(time.Now().Local()),
					Total:     int64(100),
					Amount:    int64(100),
					Fee:       int64(0),
					Status:    domain.WITHDRAWAL_REQUEST_STATUS_WAIT,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := withdrawalCloseMysqlRepository{
				dbFactory: dbFactoryMock{},
			}

			_, err := r.Create(tt.args.ctx, tt.args.withdrawalClose)
			assert.Equalf(t, tt.wantErr, err != nil, "withdrawalCloseMysqlRepository.Create() error = %v, wantErr %v", err, tt.wantErr)

		})
	}
}

func Test_withdrawalClose_Update(t *testing.T) {
	type args struct {
		ctx             context.Context
		withdrawalClose domain.WithdrawalClose
	}

	now := jsonutils.JSONTime(time.Now().Local())
	tests := []struct {
		name    string
		args    args
		want    domain.WithdrawalClose
		wantErr bool
	}{
		{
			name: "withdrawal close 업데이트",
			args: args{
				ctx: context.TODO(),
				withdrawalClose: domain.WithdrawalClose{
					Id:           mockWithdrawalClose1.Id,
					TransferDate: now,
					Status:       domain.WITHDRAWAL_REQUEST_STATUS_SENDING,
				},
			},
			want: domain.WithdrawalClose{
				Id:           mockWithdrawalClose1.Id,
				CloserId:     mockWithdrawalClose1.CloserId,
				CloseDate:    mockWithdrawalClose1.CloseDate,
				TransferDate: now,
				Total:        mockWithdrawalClose1.Total,
				Amount:       mockWithdrawalClose1.Amount,
				Fee:          mockWithdrawalClose1.Fee,
				Status:       domain.WITHDRAWAL_REQUEST_STATUS_SENDING,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session, _ := dbFactoryMock{}.EnergySharesDB(tt.args.ctx)
			ctx := context.WithValue(tt.args.ctx, factory.ConextDBType, session)

			repo := withdrawalCloseMysqlRepository{
				dbFactory: factory.NewDBFactory(),
			}

			err := repo.Update(ctx, tt.args.withdrawalClose)
			assert.Equalf(t, tt.wantErr, err != nil, "withdrawalCloseMysqlRepository.Update() error = %v, wantErr %v", err, tt.wantErr)

			ignoreTypes := cmpopts.IgnoreTypes(jsonutils.JSONTime{})
			result, err := repo.GetById(ctx, tt.args.withdrawalClose.Id)
			assert.Equalf(t, cmp.Equal(tt.want, result, ignoreTypes), true, "withdrawalCloseMysqlRepository.GetById() = %+v, want %+v", result, tt.want)
			assert.Equalf(t, tt.want.TransferDate.ToFormat(time.DateOnly), result.TransferDate.ToFormat(time.DateOnly), "withdrawalCloseMysqlRepository.GetById() = %+v, want %+v", result.TransferDate.ToFormat(time.DateOnly), tt.want.TransferDate.ToFormat(time.DateOnly))
		})
	}
}

func Test_withdrawalClose_UpdateMany(t *testing.T) {
	type args struct {
		ctx             context.Context
		options         []cdomain.WithdrawalCloseQueryOption
		withdrawalClose domain.WithdrawalClose
	}

	ids := []int64{mockWithdrawalClose1.Id, mockWithdrawalClose2.Id}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "withdrawal close 상태 업데이트",
			args: args{
				ctx: context.TODO(),
				options: []cdomain.WithdrawalCloseQueryOption{
					cdomain.WithdrawalCloseQueryOptionFunc().WithIds(ids),
				},
				withdrawalClose: domain.WithdrawalClose{
					Status: domain.WITHDRAWAL_REQUEST_STATUS_COMPLETE,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session, _ := dbFactoryMock{}.EnergySharesDB(tt.args.ctx)
			ctx := context.WithValue(tt.args.ctx, factory.ConextDBType, session)

			repo := withdrawalCloseMysqlRepository{
				dbFactory: factory.NewDBFactory(),
			}

			err := repo.UpdateMany(ctx, tt.args.withdrawalClose, tt.args.options...)
			assert.Equalf(t, tt.wantErr, err != nil, "withdrawalCloseMysqlRepository.UpdateMany() error = %v, wantErr %v", err, tt.wantErr)

			for _, id := range ids {
				result, err := repo.GetById(ctx, id)
				assert.Equalf(t, nil, err, "withdrawalCloseMysqlRepository.GetById() error = %v, wantErr %v", err, nil)
				assert.Equalf(t, cmp.Equal(tt.args.withdrawalClose.Status, result.Status), true, "withdrawalCloseMysqlRepository.GetById() = %+v, want %+v", result.Status, tt.args.withdrawalClose.Status)
			}
		})
	}
}

func Test_withdrawalClose_GetAll(t *testing.T) {
	type args struct {
		ctx     context.Context
		options []cdomain.WithdrawalCloseQueryOption
	}

	tests := []struct {
		name    string
		args    args
		want    []domain.WithdrawalCloseDetail
		wantErr bool
	}{
		{
			name: "withdrawal close 전체 조회 - 조건 없음",
			args: args{
				ctx:     context.TODO(),
				options: []cdomain.WithdrawalCloseQueryOption{},
			},
			want: []domain.WithdrawalCloseDetail{
				{
					WithdrawalClose: mockWithdrawalClose1,
					CloserName:      mockAdminUser1.Name,
				},
				{
					WithdrawalClose: mockWithdrawalClose2,
					CloserName:      mockAdminUser1.Name,
				},
			},
			wantErr: false,
		},
		{
			name: "withdrawal close 전체 조회 - 조건 있음",
			args: args{
				ctx: context.TODO(),
				options: []cdomain.WithdrawalCloseQueryOption{
					cdomain.WithdrawalCloseQueryOptionFunc().WithSortBy("id"),
					cdomain.WithdrawalCloseQueryOptionFunc().WithOrder("desc"),
					cdomain.WithdrawalCloseQueryOptionFunc().WithLimit(1),
					cdomain.WithdrawalCloseQueryOptionFunc().WithOffset(0),
				},
			},
			want: []domain.WithdrawalCloseDetail{
				{
					WithdrawalClose: mockWithdrawalClose2,
					CloserName:      mockAdminUser1.Name,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := withdrawalCloseMysqlRepository{
				dbFactory: dbFactoryMock{},
			}

			_, got, err := r.GetAll(tt.args.ctx, tt.args.options...)

			for _, g := range got {
				fmt.Printf(" %+v\n", g)
			}

			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			assert.Equalf(t, tt.wantErr, err != nil, "GetAll() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, true, cmp.Equal(tt.want, got, ignoreTime), "GetAll() diff = %v", cmp.Diff(tt.want, got, ignoreTime))
		})
	}
}
