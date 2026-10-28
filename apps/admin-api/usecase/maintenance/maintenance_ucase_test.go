package maintenance

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/illuminarean/energy-shares-backend-repo/domain"

	"github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain/mocks"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"

	adomain "github.com/illuminarean/energy-share-core-api/domain"
)

func Test_maintenanceUsecase_PatchProjects(t *testing.T) {
	type args struct {
		ctx     context.Context
		request adomain.PatchMaintenanceProjectsRequest
	}

	tests := []struct {
		name              string
		args              args
		projectRepository func() cdomain.ProjectRepository
		wantErr           bool
	}{
		{
			name: "プロジェクトフィールド更新 (EndAt)",
			args: args{
				ctx: context.Background(),
				request: adomain.PatchMaintenanceProjectsRequest{
					ProjectId: 1231258,
					EndAt:     "2024-07-09 00:00:00",
				},
			},
			projectRepository: func() cdomain.ProjectRepository {
				repo := new(mocks.ProjectRepository)

				endAt, _ := parseMultiFormat("2024-07-09 00:00:00")

				repo.On("GetById", context.Background(), int64(1231258)).Return(domain.Project{
					Id:    int64(1231258),
					EndAt: endAt,
				}, nil)
				repo.On("Update", context.Background(), domain.Project{
					Id:    int64(1231258),
					EndAt: endAt,
				}).Return(nil)
				return repo
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := maintenanceUsecase{
				projectRepository: tt.projectRepository(),
			}

			err := usecase.PatchProjects(tt.args.ctx, tt.args.request)
			assert.NoErrorf(t, err, "maintenanceUsecase.PatchProjects() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_maintenanceUsecase_PatchDepositRequests(t *testing.T) {
	type args struct {
		ctx     context.Context
		request adomain.PatchMaintenanceDepositRequestsRequest
	}

	tests := []struct {
		name                     string
		args                     args
		depositRequestRepository func() cdomain.DepositRequestRepository
		wantErr                  bool
	}{
		{
			name: "入金要求フィールド更新 (ExpireDepositDate)",
			args: args{
				ctx: context.Background(),
				request: adomain.PatchMaintenanceDepositRequestsRequest{
					DepositRequestId:  1231258,
					ExpireDepositDate: "2024-07-09 00:00:00",
				},
			},
			depositRequestRepository: func() cdomain.DepositRequestRepository {
				repo := new(mocks.DepositRequestRepository)

				expireDepositDate, _ := parseMultiFormat("2024-07-09 00:00:00")

				repo.On("GetById", context.Background(), int64(1231258)).Return(domain.DepositRequest{
					Id:                int64(1231258),
					ExpireDepositDate: expireDepositDate,
				}, nil)
				repo.On("Update", context.Background(), domain.DepositRequest{
					Id:                int64(1231258),
					ExpireDepositDate: expireDepositDate,
				}).Return(nil)
				return repo
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := maintenanceUsecase{
				depositRequestRepository: tt.depositRequestRepository(),
			}

			err := usecase.PatchDepositRequests(tt.args.ctx, tt.args.request)
			assert.NoErrorf(t, err, "maintenanceUsecase.PatchDepositRequests() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_maintenanceUsecase_PatchRepaymentDividendSchedules(t *testing.T) {
	type args struct {
		ctx     context.Context
		request adomain.PatchMaintenanceRepaymentDividendSchedulesRequest
	}

	tests := []struct {
		name                                string
		args                                args
		repaymentDividendScheduleRepository func() cdomain.RepaymentDividendScheduleRepository
		wantErr                             bool
	}{
		{
			name: "返済配当スケジュールフィールド更新 (ExecuteDate)",
			args: args{
				ctx: context.Background(),
				request: adomain.PatchMaintenanceRepaymentDividendSchedulesRequest{
					RepaymentDividendScheduleId: 1231258,
					ExecuteDate:                 "2024-07-09 00:00:00",
				},
			},
			repaymentDividendScheduleRepository: func() cdomain.RepaymentDividendScheduleRepository {
				repo := new(mocks.RepaymentDividendScheduleRepository)

				executeDate, _ := parseMultiFormat("2024-07-09 00:00:00")

				repo.On("GetById", context.Background(), int64(1231258)).Return(domain.RepaymentDividendSchedule{
					Id:          int64(1231258),
					ExecuteDate: executeDate,
				}, nil)
				repo.On("Update", context.Background(), domain.RepaymentDividendSchedule{
					Id:          int64(1231258),
					ExecuteDate: executeDate,
				}).Return(nil)
				return repo
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := maintenanceUsecase{
				repaymentDividendScheduleRepository: tt.repaymentDividendScheduleRepository(),
			}

			err := usecase.PatchRepaymentDividendSchedules(tt.args.ctx, tt.args.request)
			assert.NoErrorf(t, err, "maintenanceUsecase.PatchRepaymentDividendSchedules() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}
