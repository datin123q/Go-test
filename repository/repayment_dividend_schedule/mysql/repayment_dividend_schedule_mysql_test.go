package mysql

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

func Test_repaymentDividendScheduleMysqlRepository_Create(t *testing.T) {
	type args struct {
		ctx                       context.Context
		repaymentDividendSchedule domain.RepaymentDividendSchedule
	}
	tests := []struct {
		name    string
		args    args
		want    error
		wantErr bool
	}{
		{
			name: "(성공) 상환/배당 스케줄 생성",
			args: args{
				ctx:                       context.Background(),
				repaymentDividendSchedule: mockRepaymentDividendScheduleCreate,
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := repaymentDividendScheduleMysqlRepository{
				dbFactory: dbFactoryMock{},
			}

			_, err := repo.Create(tt.args.ctx, tt.args.repaymentDividendSchedule)

			assert.Equalf(t, tt.wantErr, err != nil, "Create() error = %v, wantErr %v", err, tt.wantErr)
			assert.Truef(t, cmp.Equal(tt.want, err), "Create() = %v, want %v", err, tt.want)

		})
	}
}

func Test_repaymentDividendScheduleMysqlRepository_Update(t *testing.T) {
	type args struct {
		ctx                       context.Context
		repaymentDividendSchedule domain.RepaymentDividendSchedule
	}
	tests := []struct {
		name    string
		args    args
		want    error
		wantErr bool
	}{
		{
			name: "(성공) 상환/배당 스케줄 수정",
			args: args{
				ctx:                       context.Background(),
				repaymentDividendSchedule: mockRepaymentDividendSchedule1Update,
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := repaymentDividendScheduleMysqlRepository{
				dbFactory: dbFactoryMock{},
			}

			err := repo.Update(tt.args.ctx, tt.args.repaymentDividendSchedule)

			assert.Equalf(t, tt.wantErr, err != nil, "Update() error = %v, wantErr %v", err, tt.wantErr)
			assert.Truef(t, cmp.Equal(tt.want, err), "Update() = %v, want %v", err, tt.want)

		})
	}
}

func Test_repaymentDividendScheduleMysqlRepository_GetAll(t *testing.T) {
	type args struct {
		ctx     context.Context
		options []cdomain.RepaymentDividendScheduleQueryOption
	}
	tests := []struct {
		name    string
		args    args
		want    domain.RepaymentDividendSchedules
		wantErr bool
	}{
		{
			name: "전체 조회 => 4건이 조회되기를 기대합니다.",
			args: args{
				ctx:     context.Background(),
				options: []cdomain.RepaymentDividendScheduleQueryOption{},
			},
			want: domain.RepaymentDividendSchedules{
				mockRepaymentDividendSchedule1,
				mockRepaymentDividendSchedule2,
				mockRepaymentDividendSchedule3,
				mockRepaymentDividendSchedule4,
			},
			wantErr: false,
		},
		{
			name: "특정 프로젝트 ID로 조회 :2000번 => 3건이 조회되기를 기대합니다.",
			args: args{
				ctx: context.Background(),
				options: []cdomain.RepaymentDividendScheduleQueryOption{
					cdomain.RepaymentDividendQueryOptionFunc().WithProjectId(2000),
				},
			},
			want: domain.RepaymentDividendSchedules{
				mockRepaymentDividendSchedule1,
				mockRepaymentDividendSchedule2,
				mockRepaymentDividendSchedule3,
			},
			wantErr: false,
		},
		{
			name: "특정 프로젝트 ID와 상환유형으로 조회 : 2000, REPAYMENT_RETURN => 2건이 조회되기를 기대합니다.",
			args: args{
				ctx: context.Background(),
				options: []cdomain.RepaymentDividendScheduleQueryOption{
					cdomain.RepaymentDividendQueryOptionFunc().WithProjectId(2000),
					cdomain.RepaymentDividendQueryOptionFunc().WithReturnType(domain.REPAYMENT_RETURN),
				},
			},
			want: domain.RepaymentDividendSchedules{
				mockRepaymentDividendSchedule1,
				mockRepaymentDividendSchedule2,
			},
			wantErr: false,
		},
		{
			name: "특정 프로젝트 ID와 상환유형으로 조회, 배당상환 실시 일자를 최신순으로 정렬, Limit을 1건으로 줘서 단건만 출력 : 2000, REPAYMENT_RETURN, execute_date, (1,0) => 1건이 조회되기를 기대합니다.",
			args: args{
				ctx: context.Background(),
				options: []cdomain.RepaymentDividendScheduleQueryOption{
					cdomain.RepaymentDividendQueryOptionFunc().WithLimit(1),
					cdomain.RepaymentDividendQueryOptionFunc().WithProjectId(2000),
					cdomain.RepaymentDividendQueryOptionFunc().WithReturnType(domain.REPAYMENT_RETURN),
					cdomain.RepaymentDividendQueryOptionFunc().WithOrder("desc"),
					cdomain.RepaymentDividendQueryOptionFunc().WithSortBy("execute_date"),
				},
			},
			want: domain.RepaymentDividendSchedules{
				mockRepaymentDividendSchedule2,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := repaymentDividendScheduleMysqlRepository{
				dbFactory: dbFactoryMock{},
			}

			_, got, err := repo.GetAll(tt.args.ctx, tt.args.options...)

			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			assert.Equalf(t, tt.wantErr, err != nil, "GetAll() error = %v, wantErr %v", err, tt.wantErr)
			assert.Truef(t, cmp.Equal(tt.want, got, ignoreTime), "GetAll() = %v, want %v", got, tt.want)

		})
	}
}

func Test_repaymentDividendScheduleMysqlRepository_GetById(t *testing.T) {
	type args struct {
		ctx context.Context
		id  int64
	}
	tests := []struct {
		name    string
		args    args
		want    domain.RepaymentDividendSchedule
		wantErr bool
	}{
		{
			name: "원하는 ID로 조회 : 1000번 => 1건이 조회되기를 기대합니다.",
			args: args{
				ctx: context.Background(),
				id:  mockRepaymentDividendSchedule1.Id,
			},
			want:    mockRepaymentDividendSchedule1,
			wantErr: false,
		},
		{
			name: "존재하지 않는 ID로 조회: 에러를 기대합니다.",
			args: args{
				ctx: context.Background(),
				id:  9999,
			},
			want:    domain.RepaymentDividendSchedule{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := repaymentDividendScheduleMysqlRepository{
				dbFactory: dbFactoryMock{},
			}

			got, err := repo.GetById(tt.args.ctx, tt.args.id)

			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			assert.Equalf(t, tt.wantErr, err != nil, "GetById() error = %v, wantErr %v", err, tt.wantErr)
			assert.Truef(t, cmp.Equal(tt.want, got, ignoreTime), "GetById() = %v, want %v", got, tt.want)

		})
	}
}

func Test_repaymentDividendScheduleMysqlRepository_Delete(t *testing.T) {
	type args struct {
		ctx context.Context
		id  int64
	}
	tests := []struct {
		name    string
		args    args
		want    error
		wantErr bool
	}{
		{
			name: "특정 아이디로 삭제 : 1000번 => 1건이 삭제되기를 기대합니다.",
			args: args{
				ctx: context.Background(),
				id:  mockRepaymentDividendSchedule1.Id,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "존재하지 않는 아이디로 삭제 : 9999번 => 삭제는 없겠지만 에러는 발생하지 않습니다.",
			args: args{
				ctx: context.Background(),
				id:  9999,
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := repaymentDividendScheduleMysqlRepository{
				dbFactory: dbFactoryMock{},
			}

			err := repo.Delete(tt.args.ctx, tt.args.id)

			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			assert.Equalf(t, tt.wantErr, err != nil, "Delete() error = %v, wantErr %v", err, tt.wantErr)
			assert.Truef(t, cmp.Equal(tt.want, err, ignoreTime), "Delete() = %v, want %v", err, tt.want)

		})
	}
}

func Test_repaymentDividendScheduleMysqlRepository_GetSum(t *testing.T) {
	type args struct {
		ctx     context.Context
		options []cdomain.RepaymentDividendScheduleQueryOption
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "프로젝트 ID로 합계 조회 : 2000번 => RepaymentAmount 합계 300을 기대합니다.",
			args: args{
				ctx: context.Background(),
				options: []cdomain.RepaymentDividendScheduleQueryOption{
					cdomain.RepaymentDividendQueryOptionFunc().WithProjectId(2000),
					cdomain.RepaymentDividendQueryOptionFunc().WithSumColumn("repayment_amount"),
				},
			},
			want:    300,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := repaymentDividendScheduleMysqlRepository{
				dbFactory: dbFactoryMock{},
			}

			got, err := repo.GetSum(tt.args.ctx, tt.args.options...)

			assert.Equalf(t, tt.wantErr, err != nil, "GetSum() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want, int64(got), "GetSum() want = %v, got %v", tt.want, int64(got))

		})
	}
}

func Test_repaymentDividendScheduleMysqlRepository_GetSums(t *testing.T) {
	type args struct {
		ctx     context.Context
		options []cdomain.RepaymentDividendScheduleQueryOption
	}
	tests := []struct {
		name    string
		args    args
		want    []int64
		wantErr bool
	}{
		{
			name: "프로젝트 ID로 합계들 조회 : 2000번 => RepaymentAmount, DividendAmount 각각 합계 300을 기대합니다.",
			args: args{
				ctx: context.Background(),
				options: []cdomain.RepaymentDividendScheduleQueryOption{
					cdomain.RepaymentDividendQueryOptionFunc().WithProjectId(2000),
					cdomain.RepaymentDividendQueryOptionFunc().WithSumColumns([]string{"repayment_amount", "dividend_amount"}),
				},
			},
			want:    []int64{300, 300},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := repaymentDividendScheduleMysqlRepository{
				dbFactory: dbFactoryMock{},
			}

			got, err := repo.GetSums(tt.args.ctx, tt.args.options...)

			// ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			assert.Equalf(t, tt.wantErr, err != nil, "GetSums() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want[0], int64(got[0]), "GetSums() want = %v, got %v", tt.want, int64(got[0]))
			assert.Equalf(t, tt.want[1], int64(got[1]), "GetSums() want = %v, got %v", tt.want, int64(got[1]))

		})
	}
}
