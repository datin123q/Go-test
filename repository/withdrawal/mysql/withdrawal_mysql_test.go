package mysql

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithdrawalRepository_Create(t *testing.T) {
	type args struct {
		ctx        context.Context
		withdrawal domain.Withdrawal
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "withdrawal 생성",
			args: args{
				ctx: context.Background(),
				withdrawal: domain.Withdrawal{
					InvestorId:   "1",
					InvestorName: "investor1",
					Amount:       1000,
					Status:       domain.WITHDRAWAL_STATUS_WAIT,
					RequestDate:  jsonutils.JSONTime(time.Now().AddDate(0, 0, 5)),
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := withdrawalMysqlRepository{
				dbFactory: &dbFactoryMock{},
			}

			_, err := r.Create(tt.args.ctx, tt.args.withdrawal)
			if (err != nil) != tt.wantErr {
				t.Errorf("WithdrawalRepository.Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestWithdrawalRepository_GetAll(t *testing.T) {
	type args struct {
		ctx     context.Context
		options []cdomain.WithdrawalQueryOption
	}

	tests := []struct {
		name    string
		args    args
		want    []domain.WithdrawalDetail
		wantErr bool
	}{
		{
			name: "withdrawal 상세 조회 - 조건 없음",
			args: args{
				ctx: context.Background(),
				options: []cdomain.WithdrawalQueryOption{
					cdomain.WithdrawalQueryOptionFunc().WithSortBy("withdrawal.id"),
					cdomain.WithdrawalQueryOptionFunc().WithOrder("desc"),
					cdomain.WithdrawalQueryOptionFunc().WithLimit(5),
					cdomain.WithdrawalQueryOptionFunc().WithOffset(0),
				},
			},
			want: []domain.WithdrawalDetail{
				{
					mockWaitWithdrawal,
					mockUser1,
				},
				{
					mockInCompleteWithdrawal,
					mockUser2,
				},
				{
					mockCompleteWithdrawal,
					mockUser1,
				},
			},
			wantErr: false,
		},
		{
			name: "withdrawal 상세 조회 - 상태 검색",
			args: args{
				ctx: context.Background(),
				options: []cdomain.WithdrawalQueryOption{
					cdomain.WithdrawalQueryOptionFunc().WithStatus(lo.ToPtr(domain.WITHDRAWAL_STATUS_WAIT)),
					cdomain.WithdrawalQueryOptionFunc().WithSortBy("id"),
					cdomain.WithdrawalQueryOptionFunc().WithOrder("desc"),
					cdomain.WithdrawalQueryOptionFunc().WithLimit(3),
					cdomain.WithdrawalQueryOptionFunc().WithOffset(0),
				},
			},
			want: []domain.WithdrawalDetail{
				{
					mockWaitWithdrawal,
					mockUser1,
				},
			},
			wantErr: false,
		},
		{
			name: "withdrawal 상세 조회 - closeId 검색",
			args: args{
				ctx: context.Background(),
				options: []cdomain.WithdrawalQueryOption{
					cdomain.WithdrawalQueryOptionFunc().WithCloseId(lo.ToPtr(int64(1))),
					cdomain.WithdrawalQueryOptionFunc().WithSortBy("withdrawal.id"),
					cdomain.WithdrawalQueryOptionFunc().WithOrder("asc"),
					cdomain.WithdrawalQueryOptionFunc().WithLimit(3),
					cdomain.WithdrawalQueryOptionFunc().WithOffset(0),
				},
			},
			want: []domain.WithdrawalDetail{
				{
					mockCompleteWithdrawal,
					mockUser1,
				},
				{
					mockInCompleteWithdrawal,
					mockUser2,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := withdrawalMysqlRepository{
				dbFactory: &dbFactoryMock{},
			}

			_, got, err := r.GetAll(tt.args.ctx, tt.args.options...)

			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			assert.Equalf(t, tt.wantErr, err != nil, "WithdrawalRepository.GetAll() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, true, cmp.Equal(tt.want, got, ignoreTime), "GetAll() diff = %v", cmp.Diff(tt.want, got, ignoreTime))
		})
	}
}

func TestWithdrawalRepository_GetById(t *testing.T) {
	type args struct {
		ctx context.Context
		id  int64
	}

	tests := []struct {
		name    string
		args    args
		want    domain.WithdrawalDetail
		wantErr bool
	}{
		{
			name: "withdrawal 상세 조회",
			args: args{
				ctx: context.Background(),
				id:  3,
			},
			want:    domain.WithdrawalDetail{Withdrawal: mockWaitWithdrawal, User: mockUser1},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := withdrawalMysqlRepository{
				dbFactory: &dbFactoryMock{},
			}

			got, err := r.GetById(tt.args.ctx, tt.args.id)

			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			assert.Equalf(t, tt.wantErr, err != nil, "WithdrawalRepository.GetById() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, true, cmp.Equal(tt.want, got, ignoreTime), "GetById() diff = %v", cmp.Diff(tt.want, got, ignoreTime))
		})
	}
}

func TestWithdrawalRepository_GetByTransactionId(t *testing.T) {
	type args struct {
		ctx           context.Context
		transactionId string
	}

	tests := []struct {
		name    string
		args    args
		want    domain.Withdrawal
		wantErr bool
	}{
		{
			name: "withdrawal 조회 - transactionId",
			args: args{
				ctx:           context.Background(),
				transactionId: "transaction_id_1",
			},
			want:    mockCompleteWithdrawal,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := withdrawalMysqlRepository{
				dbFactory: &dbFactoryMock{},
			}

			got, err := r.GetByTransactionId(tt.args.ctx, tt.args.transactionId)

			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			assert.Equalf(t, tt.wantErr, err != nil, "WithdrawalRepository.GetByTransactionId() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, true, cmp.Equal(tt.want, got, ignoreTime), "GetByTransactionId() diff = %v", cmp.Diff(tt.want, got, ignoreTime))
		})
	}
}

func TestWithdrawalRepository_Update(t *testing.T) {
	type args struct {
		ctx        context.Context
		withdrawal domain.Withdrawal
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "withdrawal 업데이트",
			args: args{
				ctx: context.Background(),
				withdrawal: domain.Withdrawal{
					Id:           1,
					InvestorId:   "1",
					InvestorName: "investor1",
					Amount:       1000,
					Status:       domain.WITHDRAWAL_STATUS_WAIT,
					RequestDate:  jsonutils.JSONTime(time.Now().AddDate(0, 0, 5)),
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := withdrawalMysqlRepository{
				dbFactory: &dbFactoryMock{},
			}

			err := r.Update(tt.args.ctx, tt.args.withdrawal)
			assert.Equalf(t, tt.wantErr, err != nil, "WithdrawalRepository.Update() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func TestWithdrawalRepository_GetSum(t *testing.T) {
	type args struct {
		ctx     context.Context
		options []cdomain.WithdrawalQueryOption
	}

	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "withdrawal 합계 조회 - incomplete 상태인 것의 amount 합계",
			args: args{
				ctx: context.Background(),
				options: []cdomain.WithdrawalQueryOption{
					cdomain.WithdrawalQueryOptionFunc().WithInvestorId(lo.ToPtr("2")),
					cdomain.WithdrawalQueryOptionFunc().WithStatus(lo.ToPtr(domain.WITHDRAWAL_STATUS_INCOMPLETE)),
					cdomain.WithdrawalQueryOptionFunc().WithSumColumn("amount"),
				},
			},
			want:    2000,
			wantErr: false,
		},
		{
			name: "withdrawal 합계 조회 - wait 상태인 것의 fee 합계",
			args: args{
				ctx: context.Background(),
				options: []cdomain.WithdrawalQueryOption{
					cdomain.WithdrawalQueryOptionFunc().WithInvestorId(lo.ToPtr("1")),
					cdomain.WithdrawalQueryOptionFunc().WithStatus(lo.ToPtr(domain.WITHDRAWAL_STATUS_WAIT)),
					cdomain.WithdrawalQueryOptionFunc().WithSumColumn("fee"),
				},
			},
			want:    30,
			wantErr: false,
		},
		{
			name: "withdrawal 합계 조회 - Ids 에 포함되는 것들의 fee 합계",
			args: args{
				ctx: context.Background(),
				options: []cdomain.WithdrawalQueryOption{
					cdomain.WithdrawalQueryOptionFunc().WithIds([]int64{1, 2}),
					cdomain.WithdrawalQueryOptionFunc().WithSumColumn("fee"),
				},
			},
			want:    35,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := withdrawalMysqlRepository{
				dbFactory: &dbFactoryMock{},
			}

			got, err := r.GetSum(tt.args.ctx, tt.args.options...)

			assert.Equalf(t, tt.wantErr, err != nil, "WithdrawalRepository.GetSum() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want, got, "GetSum() = %v, want %v", got, tt.want)
		})
	}
}

func TestWithdrawalRepository_CreateAndGet_IsManual(t *testing.T) {
	r := withdrawalMysqlRepository{
		dbFactory: &dbFactoryMock{},
	}
	ctx := context.Background()

	id, err := r.Create(ctx, domain.Withdrawal{
		InvestorId:   "1",
		InvestorName: "investor1",
		Amount:       9999,
		Fee:          50,
		Status:       domain.WITHDRAWAL_STATUS_INCOMPLETE,
		IsManual:     true,
		RequestDate:  jsonutils.JSONTime(time.Now()),
	})
	require.NoError(t, err)
	require.NotZero(t, id)

	got, err := r.GetById(ctx, id)
	require.NoError(t, err)
	assert.True(t, got.Withdrawal.IsManual, "GetById should return is_manual=true")
	assert.Equal(t, int64(9999), got.Withdrawal.Amount)

	_, items, err := r.GetAll(ctx,
		cdomain.WithdrawalQueryOptionFunc().WithInvestorId(lo.ToPtr("1")),
		cdomain.WithdrawalQueryOptionFunc().WithSortBy("withdrawal.id"),
		cdomain.WithdrawalQueryOptionFunc().WithOrder("desc"),
		cdomain.WithdrawalQueryOptionFunc().WithLimit(10),
		cdomain.WithdrawalQueryOptionFunc().WithOffset(0),
	)
	require.NoError(t, err)

	var found *domain.WithdrawalDetail
	for i := range items {
		if items[i].Withdrawal.Id == id {
			found = &items[i]
			break
		}
	}
	require.NotNil(t, found, "GetAll should include the created withdrawal")
	assert.True(t, found.Withdrawal.IsManual, "GetAll should return is_manual=true")
}
