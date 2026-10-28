package maintenance

import (
	"context"
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"

	"github.com/samber/lo"

	"github.com/pkg/errors"

	"github.com/illuminarean/energy-share-core-api/domain"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
)

type maintenanceUsecase struct {
	projectRepository                   cdomain.ProjectRepository
	depositRequestRepository            cdomain.DepositRequestRepository
	repaymentDividendScheduleRepository cdomain.RepaymentDividendScheduleRepository
}

func New(
	projectRepository cdomain.ProjectRepository,
	depositRequestRepository cdomain.DepositRequestRepository,
	repaymentDividendScheduleRepository cdomain.RepaymentDividendScheduleRepository,

) domain.MaintenanceUsecase {
	return &maintenanceUsecase{
		projectRepository:                   projectRepository,
		depositRequestRepository:            depositRequestRepository,
		repaymentDividendScheduleRepository: repaymentDividendScheduleRepository,
	}
}

func (m maintenanceUsecase) PatchProjects(ctx context.Context, request domain.PatchMaintenanceProjectsRequest) error {
	project, err := m.projectRepository.GetById(ctx, request.ProjectId)
	if err != nil {
		return errors.Wrap(err, "err get project by id")
	}

	if lo.IsNotEmpty(request.EndAt) {
		project.EndAt = lo.Must(parseMultiFormat(request.EndAt))
	}

	if err = m.projectRepository.Update(ctx, project); err != nil {
		return errors.Wrap(err, "err update project")
	}

	return nil
}

func (m maintenanceUsecase) PatchDepositRequests(ctx context.Context, request domain.PatchMaintenanceDepositRequestsRequest) error {
	depositRequest, err := m.depositRequestRepository.GetById(ctx, request.DepositRequestId)
	if err != nil {
		return errors.Wrap(err, "err get deposit request by id")
	}

	if lo.IsNotEmpty(request.ExpireDepositDate) {
		depositRequest.ExpireDepositDate = lo.Must(parseMultiFormat(request.ExpireDepositDate))
	}

	if err = m.depositRequestRepository.Update(ctx, depositRequest); err != nil {
		return errors.Wrap(err, "err update depositRequest")
	}

	return nil
}

func (m maintenanceUsecase) PatchRepaymentDividendSchedules(ctx context.Context, request domain.PatchMaintenanceRepaymentDividendSchedulesRequest) error {
	repaymentDividendSchedule, err := m.repaymentDividendScheduleRepository.GetById(ctx, request.RepaymentDividendScheduleId)
	if err != nil {
		return errors.Wrap(err, "err get repaymentDividendSchedule by id")
	}

	if lo.IsNotEmpty(request.ExecuteDate) {
		repaymentDividendSchedule.ExecuteDate = lo.Must(parseMultiFormat(request.ExecuteDate))
	}

	if err = m.repaymentDividendScheduleRepository.Update(ctx, repaymentDividendSchedule); err != nil {
		return errors.Wrap(err, "err update repaymentDividendSchedule")
	}

	return nil
}

func parseMultiFormat(value string) (jsonutils.JSONTime, error) {
	var timeFormats = []string{
		time.RFC3339,
		time.DateTime,
		time.DateOnly,
		"2006-01-02T15:04:05 09:00", // 비정형 타입
	}

	for _, format := range timeFormats {
		if t, err := jsonutils.ParseJSONTimeJST(format, value); err == nil {
			return jsonutils.JSONTime(t), nil
		}
	}
	return jsonutils.JSONTime{}, errors.New("unable to parse time string:" + value)
}
