package domain

import (
	"context"
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

type MaintenanceUsecase interface {
	PatchProjects(ctx context.Context, request PatchMaintenanceProjectsRequest) error
	PatchDepositRequests(ctx context.Context, request PatchMaintenanceDepositRequestsRequest) error
	PatchRepaymentDividendSchedules(ctx context.Context, request PatchMaintenanceRepaymentDividendSchedulesRequest) error
}

type PatchMaintenanceProjectsRequest struct {
	ProjectId int64  `json:"project_id" example:"1231258" description:"프로젝트 ID"`
	EndAt     string `json:"end_at" example:"2024-07-09 00:00:00" description:"모집 종료일" format:"yyyy-mm-dd HH:mm:ss"`
}

func (p PatchMaintenanceProjectsRequest) Validate() error {
	if lo.IsNotEmpty(p.EndAt) {
		if _, err := jsonutils.ParseDateTimeJST(time.DateTime, p.EndAt); err != nil {
			return errors.Wrap(err, "err end_at format yyyy-mm-dd hh:mm:ss")
		}
	}

	return nil
}

type PatchMaintenanceDepositRequestsRequest struct {
	DepositRequestId  int64  `json:"deposit_request_id" example:"1231258" description:"입금요청 ID"`
	ExpireDepositDate string `json:"expire_deposit_date" example:"2024-07-09 00:00:00" description:"입금 완료 예정일" format:"yyyy-mm-dd HH:mm:ss"`
}

func (p PatchMaintenanceDepositRequestsRequest) Validate() error {
	if lo.IsNotEmpty(p.ExpireDepositDate) {
		if _, err := jsonutils.ParseDateTimeJST(time.DateTime, p.ExpireDepositDate); err != nil {
			return errors.Wrap(err, "err expire_deposit_date format yyyy-mm-dd hh:mm:ss")
		}
	}

	return nil
}

type PatchMaintenanceRepaymentDividendSchedulesRequest struct {
	RepaymentDividendScheduleId int64  `json:"repayment_dividend_schedule_id" example:"1231258" description:"배당 상환 ID"`
	ExecuteDate                 string `json:"execute_date" example:"2024-07-09 00:00:00" description:"배당/상환 실시 일시" format:"yyyy-mm-dd HH:mm:ss"`
}

func (p PatchMaintenanceRepaymentDividendSchedulesRequest) Validate() error {
	if lo.IsNotEmpty(p.ExecuteDate) {
		if _, err := jsonutils.ParseDateTimeJST(time.DateTime, p.ExecuteDate); err != nil {
			return errors.Wrap(err, "err execute_date format yyyy-mm-dd hh:mm:ss")
		}
	}

	return nil
}
