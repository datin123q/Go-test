package lottery

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	adomain "github.com/illuminarean/energy-share-core-api/domain"
	adminservice "github.com/illuminarean/energy-share-core-api/service"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
)

type lotteryWinExecutionResult struct {
	application          domain.LotteryApplication
	createdShare         *domain.Share
	createdDepositResult *domain.DepositRequest
}

type lotteryLoseExecutionResult struct {
	application        domain.LotteryApplication
	shouldSendLostMail bool
}

func buildLotteryCSVBatchActor(requestedBy string) jwthelper.User {
	return jwthelper.User{
		Id:    0,
		Name:  requestedBy,
		Email: requestedBy,
	}
}

func (l lotteryUsecase) executeWinForBatchTx(
	txCtx context.Context,
	request adomain.LotteryWinRequest,
	historyTitle string,
) (lotteryWinExecutionResult, error) {
	lockApplication := domain.LotteryApplication{}
	has, err := factory.DB(txCtx).
		SQL("SELECT lottery_application.* FROM lottery_application WHERE lottery_application.id = ? FOR UPDATE", request.LotteryApplicationID).
		Get(&lockApplication)
	if err != nil {
		return lotteryWinExecutionResult{}, errors.Wrap(err, "failed to lock lottery application")
	}
	if !has {
		return lotteryWinExecutionResult{}, errors.Wrap(cdomain.ErrItemNotFound, "lottery application not found")
	}

	lockedProject := domain.Project{}
	has, err = factory.DB(txCtx).
		SQL("SELECT "+domain.PROJECT_DETAIL_COLUMN_LIST+" FROM project WHERE project.id = ? FOR UPDATE", lockApplication.ProjectId).
		Get(&lockedProject)
	if err != nil {
		return lotteryWinExecutionResult{}, errors.Wrap(err, "failed to lock project")
	}
	if !has {
		return lotteryWinExecutionResult{}, errors.Wrap(cdomain.ErrItemNotFound, "project not found")
	}

	txApplication, err := l.lotteryRepository.GetById(txCtx, request.LotteryApplicationID)
	if err != nil {
		return lotteryWinExecutionResult{}, errors.Wrap(err, "failed to get lottery application in tx")
	}
	if !txApplication.CanceledAt.IsZero() {
		return lotteryWinExecutionResult{}, errors.Wrap(cdomain.ErrBadRequest, "lottery application is canceled")
	}

	txProject := &lockedProject
	if err := validateWinningProject(txProject); err != nil {
		return lotteryWinExecutionResult{}, err
	}

	txUser, err := domain.User{}.GetById(txCtx, txApplication.UserId)
	if err != nil {
		return lotteryWinExecutionResult{}, errors.Wrap(err, "failed to get user in tx")
	}
	if err := validateWinningUser(txUser); err != nil {
		return lotteryWinExecutionResult{}, err
	}
	if txUser.UserType == domain.PRIVATE {
		hasMyNumber, err := domain.MyNumber{}.CheckMyNumberStatusByUserId(txCtx, txUser.Id)
		if err != nil {
			return lotteryWinExecutionResult{}, errors.Wrap(err, "failed to check my number status")
		}
		if !hasMyNumber {
			return lotteryWinExecutionResult{}, errors.Wrap(cdomain.ErrBadRequest, "my number is not registered")
		}
	}
	if err := txUser.CheckOverAge(txCtx); err != nil {
		return lotteryWinExecutionResult{}, errors.Wrap(cdomain.ErrBadRequest, err.Error())
	}
	if _, err := (domain.UserVirtualAccount{}).GetVaByUserId(txCtx, txUser.Id, txProject.ProjectType); err != nil {
		return lotteryWinExecutionResult{}, errors.Wrap(err, "failed to get linked virtual account")
	}

	txExistingWinningQty := winningQtyTotal(txApplication)
	if txExistingWinningQty+request.Qty > txApplication.Qty {
		return lotteryWinExecutionResult{}, errors.Wrap(cdomain.ErrBadRequest, "winning qty exceeds lottery application qty")
	}

	amount := txProject.SharePrice * request.Qty
	if !txProject.CheckShareQty(request.Qty, amount) {
		return lotteryWinExecutionResult{}, errors.Wrap(cdomain.ErrBadRequest, "insufficient remaining project qty or amount")
	}

	share, err := adminservice.InvestmentService{}.Invest(
		txCtx,
		txProject,
		txUser,
		request.Qty,
		domain.LOTTERY_MANUAL,
		true,
		false,
		false, // isWaitInvestFinalPortion --- 抽選は繰上げの分割対象ではないため常にfalse(MIR-1068)
	)
	if err != nil {
		return lotteryWinExecutionResult{}, errors.Wrap(err, "failed to invest by lottery winning")
	}

	depositRequest, err := domain.DepositRequest{}.GetDepositRequestByShareId(txCtx, share.Id)
	if err != nil {
		return lotteryWinExecutionResult{}, errors.Wrap(err, "failed to get deposit request in tx")
	}

	if _, err := (&domain.LotteryWinning{
		LotteryApplicationId: txApplication.Id,
		ShareId:              share.Id,
	}).Create(txCtx); err != nil {
		return lotteryWinExecutionResult{}, errors.Wrap(err, "failed to create lottery winning")
	}

	if txApplication.WinningAt.IsZero() {
		updated, err := l.lotteryRepository.MarkAsWinning(txCtx, txApplication.Id)
		if err != nil {
			return lotteryWinExecutionResult{}, errors.Wrap(err, "failed to mark lottery application as winning")
		}
		if !updated {
			return lotteryWinExecutionResult{}, errors.Wrap(cdomain.ErrBadRequest, "failed to update lottery application winning status")
		}
	}

	if _, err := (&domain.OperatorHistory{
		UserId: request.JWTUser.Id,
		Name:   request.JWTUser.Name,
		Email:  request.JWTUser.Email,
		Menu:   "menu.manage_accounting.payment_list",
		Ip:     "",
		Title:  historyTitle + " share_id=" + share.Id,
		State:  domain.CHANGE,
	}).Create(txCtx); err != nil {
		return lotteryWinExecutionResult{}, errors.Wrap(err, "failed to create operator history")
	}

	if err := l.createLotteryHistory(
		txCtx,
		txApplication.Id,
		txApplication.Status,
		lotteryStatusAfterWinning(txApplication, request.Qty),
		request.JWTUser.Id,
	); err != nil {
		return lotteryWinExecutionResult{}, errors.Wrap(err, "failed to create lottery history")
	}

	return lotteryWinExecutionResult{
		application:          txApplication,
		createdShare:         share,
		createdDepositResult: depositRequest,
	}, nil
}

func (l lotteryUsecase) sendWinningNotification(ctx context.Context, result lotteryWinExecutionResult) {
	if result.createdShare == nil || result.createdDepositResult == nil {
		return
	}

	project, err := domain.Project{}.GetById(ctx, result.application.ProjectId)
	if err != nil {
		factory.Logger(ctx).Errorf("failed to get project after commit for lottery winning mail: application_id=%s project_id=%d err=%v", result.application.Id, result.application.ProjectId, err)
		return
	}
	if project == nil {
		factory.Logger(ctx).Errorf("project not found after commit for lottery winning mail: application_id=%s project_id=%d", result.application.Id, result.application.ProjectId)
		return
	}

	user, err := domain.User{}.GetById(ctx, result.application.UserId)
	if err != nil {
		factory.Logger(ctx).Errorf("failed to get user after commit for lottery winning mail: application_id=%s user_id=%d err=%v", result.application.Id, result.application.UserId, err)
		return
	}
	if user == nil {
		factory.Logger(ctx).Errorf("user not found after commit for lottery winning mail: application_id=%s user_id=%d", result.application.Id, result.application.UserId)
		return
	}

	switch result.createdShare.InvestStatus {
	case domain.SUCCESS_INVEST:
		err := l.mailService.SendMail(ctx, cdomain.SendMailRequest{
			TemplateType: domain.LOTTERY_WINNING_IMMEDIATE_SUCCESS_TEMPLATE,
			Sender:       l.cfg.ServiceConfig.Mail.Info,
			Receivers:    []string{user.Email},
			MailType:     domain.MAIL_SEND_PRIORITY_NORMAL,
			Details: map[string]interface{}{
				"Receiver":          user.Name,
				"ProjectName":       project.Name,
				"WinningAmount":     message.NewPrinter(language.Japanese).Sprintf("%d", result.createdShare.Amount),
				"ApplicationAmount": message.NewPrinter(language.Japanese).Sprintf("%d", result.application.Amount),
				"ServiceUrl":        l.cfg.ServiceConfig.Url.UserServiceDomain,
			},
		})
		if err != nil {
			factory.Logger(ctx).Errorf("failed to send lottery winning immediate success mail: %v", err)
		}
	case domain.REQUEST_INVEST:
		err := l.mailService.SendMail(ctx, cdomain.SendMailRequest{
			TemplateType: domain.LOTTERY_WINNING_INSUFFICIENT_BALANCE_TEMPLATE,
			Sender:       l.cfg.ServiceConfig.Mail.Info,
			Receivers:    []string{user.Email},
			MailType:     domain.MAIL_SEND_PRIORITY_NORMAL,
			Details: map[string]interface{}{
				"Receiver":          user.Name,
				"ProjectName":       project.Name,
				"WinningAmount":     message.NewPrinter(language.Japanese).Sprintf("%d", result.createdShare.Amount),
				"ApplicationAmount": message.NewPrinter(language.Japanese).Sprintf("%d", result.application.Amount),
				"LackAmount":        message.NewPrinter(language.Japanese).Sprintf("%d", result.createdDepositResult.NeedAmount()),
				"PaymentEnd":        jsonutils.JSONTime.ToString(result.createdDepositResult.ExpireDepositDate),
				"ServiceUrl":        l.cfg.ServiceConfig.Url.UserServiceDomain,
			},
		})
		if err != nil {
			factory.Logger(ctx).Errorf("failed to send lottery winning insufficient balance mail: %v", err)
		}
	}
}

func (l lotteryUsecase) executeLoseForBatchTx(
	txCtx context.Context,
	request adomain.LotteryLoseRequest,
	historyTitle string,
) (lotteryLoseExecutionResult, error) {
	lockApplication := domain.LotteryApplication{}
	has, err := factory.DB(txCtx).
		SQL("SELECT lottery_application.* FROM lottery_application WHERE lottery_application.id = ? FOR UPDATE", request.LotteryApplicationID).
		Get(&lockApplication)
	if err != nil {
		return lotteryLoseExecutionResult{}, errors.Wrap(err, "failed to lock lottery application")
	}
	if !has {
		return lotteryLoseExecutionResult{}, errors.Wrap(cdomain.ErrItemNotFound, "lottery application not found")
	}

	lockedProject := domain.Project{}
	has, err = factory.DB(txCtx).
		SQL("SELECT "+domain.PROJECT_DETAIL_COLUMN_LIST+" FROM project WHERE project.id = ? FOR UPDATE", lockApplication.ProjectId).
		Get(&lockedProject)
	if err != nil {
		return lotteryLoseExecutionResult{}, errors.Wrap(err, "failed to lock project")
	}
	if !has {
		return lotteryLoseExecutionResult{}, errors.Wrap(cdomain.ErrItemNotFound, "project not found")
	}
	if err := validateWinningProject(&lockedProject); err != nil {
		return lotteryLoseExecutionResult{}, err
	}

	txApplication, err := l.lotteryRepository.GetById(txCtx, request.LotteryApplicationID)
	if err != nil {
		return lotteryLoseExecutionResult{}, errors.Wrap(err, "failed to get lottery application in tx")
	}
	if err := validateLosingApplication(txApplication); err != nil {
		return lotteryLoseExecutionResult{}, err
	}
	existingWinningQty := winningQtyTotal(txApplication)
	if existingWinningQty > 0 && existingWinningQty < txApplication.Qty {
		return lotteryLoseExecutionResult{}, errors.Wrap(cdomain.ErrBadRequest, "partial winning application cannot be lost by csv batch")
	}

	shouldSendLostMail := txApplication.LostAt.IsZero()
	if shouldSendLostMail {
		updated, err := l.lotteryRepository.MarkAsLost(txCtx, txApplication.Id)
		if err != nil {
			return lotteryLoseExecutionResult{}, errors.Wrap(err, "failed to mark lottery application as lost")
		}
		if !updated {
			return lotteryLoseExecutionResult{}, errors.Wrap(cdomain.ErrBadRequest, "failed to update lottery application lost status")
		}
	}

	if _, err := (&domain.OperatorHistory{
		UserId: request.JWTUser.Id,
		Name:   request.JWTUser.Name,
		Email:  request.JWTUser.Email,
		Menu:   "menu.manage_accounting.payment_list",
		Ip:     "",
		Title:  historyTitle,
		State:  domain.CHANGE,
	}).Create(txCtx); err != nil {
		return lotteryLoseExecutionResult{}, errors.Wrap(err, "failed to create operator history")
	}

	if err := l.createLotteryHistory(
		txCtx,
		txApplication.Id,
		txApplication.Status,
		lotteryStatusAfterLosing(txApplication),
		request.JWTUser.Id,
	); err != nil {
		return lotteryLoseExecutionResult{}, errors.Wrap(err, "failed to create lottery history")
	}

	return lotteryLoseExecutionResult{
		application:        txApplication,
		shouldSendLostMail: shouldSendLostMail,
	}, nil
}

func (l lotteryUsecase) sendLoseNotification(ctx context.Context, result lotteryLoseExecutionResult) {
	if !result.shouldSendLostMail {
		return
	}

	project, err := domain.Project{}.GetById(ctx, result.application.ProjectId)
	if err != nil {
		factory.Logger(ctx).Errorf("failed to get project after commit for lottery lost mail: application_id=%s project_id=%d err=%v", result.application.Id, result.application.ProjectId, err)
		return
	}
	if project == nil {
		factory.Logger(ctx).Errorf("project not found after commit for lottery lost mail: application_id=%s project_id=%d", result.application.Id, result.application.ProjectId)
		return
	}

	user, err := domain.User{}.GetById(ctx, result.application.UserId)
	if err != nil {
		factory.Logger(ctx).Errorf("failed to get user after commit for lottery lost mail: application_id=%s user_id=%d err=%v", result.application.Id, result.application.UserId, err)
		return
	}
	if user == nil {
		factory.Logger(ctx).Errorf("user not found after commit for lottery lost mail: application_id=%s user_id=%d", result.application.Id, result.application.UserId)
		return
	}

	err = l.mailService.SendMail(ctx, cdomain.SendMailRequest{
		TemplateType: domain.LOTTERY_LOST_TEMPLATE,
		Sender:       l.cfg.ServiceConfig.Mail.Info,
		Receivers:    []string{user.Email},
		MailType:     domain.MAIL_SEND_PRIORITY_NORMAL,
		Details: map[string]interface{}{
			"Receiver":          user.Name,
			"ProjectName":       project.Name,
			"ApplicationAmount": message.NewPrinter(language.Japanese).Sprintf("%d", result.application.Amount),
		},
	})
	if err != nil {
		factory.Logger(ctx).Errorf("failed to send lottery lost mail: %v", err)
	}
}

func (l lotteryUsecase) revalidateLotteryCSVUploadDetail(
	ctx context.Context,
	job domain.LotteryCSVUploadJob,
	detail domain.LotteryCSVUploadJobDetail,
) error {
	application, err := l.lotteryRepository.GetById(ctx, detail.LotteryApplicationID)
	if err != nil {
		return errors.Wrap(err, "failed to get lottery application")
	}
	if application.ProjectId != job.ProjectID {
		return errors.Wrap(cdomain.ErrBadRequest, "lottery application does not belong to job project")
	}
	if detail.ExpectedLotteryStatus != nil && *detail.ExpectedLotteryStatus != application.Status {
		return errors.Wrap(cdomain.ErrBadRequest, "lottery status does not match current lottery application status")
	}
	if detail.ExpectedAppliedAt != nil && *detail.ExpectedAppliedAt != formatLotteryCSVAppliedAt(application.AppliedAt.Time()) {
		return errors.Wrap(cdomain.ErrBadRequest, "applied at does not match current lottery application applied at")
	}
	if detail.InputAppliedQty != nil && *detail.InputAppliedQty != application.Qty {
		return errors.Wrap(cdomain.ErrBadRequest, "applied qty does not match current lottery application qty")
	}

	switch detail.ProcessType {
	case domain.LOTTERY_CSV_UPLOAD_PROCESS_TYPE_WIN:
		if detail.InputWinningQty <= 0 {
			return errors.Wrap(cdomain.ErrBadRequest, "winning qty must be greater than zero")
		}
		if err := validateWinningApplication(application, detail.InputWinningQty); err != nil {
			return err
		}
	case domain.LOTTERY_CSV_UPLOAD_PROCESS_TYPE_LOSE:
		if detail.InputWinningQty != 0 {
			return errors.Wrap(cdomain.ErrBadRequest, "losing row must have zero winning qty")
		}
		if winningQtyTotal(application) > 0 && winningQtyTotal(application) < application.Qty {
			return errors.Wrap(cdomain.ErrBadRequest, "partial winning application cannot be lost by csv batch")
		}
		if err := validateLosingApplication(application); err != nil {
			return err
		}
	default:
		return errors.Wrap(cdomain.ErrBadRequest, "unsupported csv upload process type")
	}

	return nil
}

func updateLotteryCSVUploadJobFailure(
	ctx context.Context,
	job *domain.LotteryCSVUploadJob,
	details []domain.LotteryCSVUploadJobDetail,
	rowMessages map[int]string,
	jobMessage string,
) error {
	now := jsonutils.JSONTime(time.Now())
	for i := range details {
		detail := details[i]
		detail.Status = domain.LOTTERY_CSV_UPLOAD_DETAIL_STATUS_FAILED
		detail.ProcessedAt = now
		if msg, ok := rowMessages[detail.RowNumber]; ok {
			detail.ErrorMessage = &msg
		} else {
			detail.ErrorMessage = &jobMessage
		}
		if err := detail.Update(ctx); err != nil {
			return errors.Wrap(err, "failed to update lottery csv upload job detail as failed")
		}
	}

	job.SuccessRowCount = 0
	job.FailedRowCount = len(details)
	job.Status = domain.LOTTERY_CSV_UPLOAD_JOB_STATUS_FAILED
	job.FinishedAt = now
	job.ErrorMessage = &jobMessage
	if err := job.Update(ctx); err != nil {
		return errors.Wrap(err, "failed to update lottery csv upload job as failed")
	}
	return nil
}

func (l lotteryUsecase) ProcessCSVUploadJob(ctx context.Context, jobID string) error {
	job, err := (domain.LotteryCSVUploadJob{}).GetByJobID(ctx, jobID)
	if err != nil {
		return errors.Wrap(err, "failed to get lottery csv upload job")
	}
	if job == nil {
		return errors.Wrap(cdomain.ErrItemNotFound, "lottery csv upload job not found")
	}
	if job.Status != domain.LOTTERY_CSV_UPLOAD_JOB_STATUS_QUEUED && job.Status != domain.LOTTERY_CSV_UPLOAD_JOB_STATUS_RUNNING {
		return errors.Wrap(cdomain.ErrBadRequest, "lottery csv upload job is not queued or running")
	}
	if job.Status == domain.LOTTERY_CSV_UPLOAD_JOB_STATUS_RUNNING {
		return nil
	}

	startedAt := job.StartedAt
	if startedAt.IsZero() {
		startedAt = jsonutils.JSONTime(time.Now())
	}
	claimed, err := job.TryMarkRunningFromQueued(ctx, startedAt)
	if err != nil {
		return errors.Wrap(err, "failed to mark lottery csv upload job as running")
	}
	if !claimed {
		return nil
	}

	project, err := l.projectRepository.GetById(ctx, job.ProjectID)
	if err != nil {
		return errors.Wrap(err, "failed to get project")
	}
	if err := validateWinningProject(&project); err != nil {
		msg := err.Error()
		job.Status = domain.LOTTERY_CSV_UPLOAD_JOB_STATUS_FAILED
		job.FinishedAt = jsonutils.JSONTime(time.Now())
		job.ErrorMessage = &msg
		if updateErr := job.Update(ctx); updateErr != nil {
			return errors.Wrap(updateErr, "failed to update lottery csv upload job after project validation failure")
		}
		return err
	}

	details, err := (domain.LotteryCSVUploadJobDetail{}).GetByJobID(ctx, job.Id)
	if err != nil {
		return errors.Wrap(err, "failed to get lottery csv upload job details")
	}

	pendingDetails := make([]domain.LotteryCSVUploadJobDetail, 0)
	for _, detail := range details {
		if detail.Status == domain.LOTTERY_CSV_UPLOAD_DETAIL_STATUS_PENDING {
			pendingDetails = append(pendingDetails, detail)
		}
	}

	if len(pendingDetails) == 0 {
		job.SuccessRowCount = 0
		job.FailedRowCount = 0
		job.Status = domain.LOTTERY_CSV_UPLOAD_JOB_STATUS_COMPLETED
		job.FinishedAt = jsonutils.JSONTime(time.Now())
		job.ErrorMessage = nil
		if err := job.Update(ctx); err != nil {
			return errors.Wrap(err, "failed to complete lottery csv upload job without pending rows")
		}
		return nil
	}

	revalidationMessages := make(map[int]string)
	for _, detail := range pendingDetails {
		if err := l.revalidateLotteryCSVUploadDetail(ctx, *job, detail); err != nil {
			revalidationMessages[detail.RowNumber] = err.Error()
		}
	}
	if len(revalidationMessages) > 0 {
		messages := make([]string, 0, len(revalidationMessages))
		for _, detail := range pendingDetails {
			if msg, ok := revalidationMessages[detail.RowNumber]; ok {
				messages = append(messages, fmt.Sprintf("row %d: %s", detail.RowNumber, msg))
			}
		}
		jobMessage := strings.Join(messages, "\n")
		if err := updateLotteryCSVUploadJobFailure(ctx, job, pendingDetails, revalidationMessages, jobMessage); err != nil {
			return err
		}
		return nil
	}

	actor := buildLotteryCSVBatchActor(job.RequestedBy)
	winResults := make(map[int]lotteryWinExecutionResult)
	loseResults := make(map[int]lotteryLoseExecutionResult)

	err = l.txManager.WithTx(ctx, func(txCtx context.Context) error {
		for _, detail := range pendingDetails {
			switch detail.ProcessType {
			case domain.LOTTERY_CSV_UPLOAD_PROCESS_TYPE_WIN:
				result, err := l.executeWinForBatchTx(
					txCtx,
					adomain.LotteryWinRequest{
						JWTUser:              actor,
						LotteryApplicationID: detail.LotteryApplicationID,
						Qty:                  detail.InputWinningQty,
					},
					fmt.Sprintf(
						"lottery csv win job_id=%s application_id=%s requested_by=%s",
						escapeCSVFormulaString(job.JobID),
						escapeCSVFormulaString(detail.LotteryApplicationID),
						escapeCSVFormulaString(job.RequestedBy),
					),
				)
				if err != nil {
					return errors.Wrapf(err, "row %d", detail.RowNumber)
				}
				winResults[detail.RowNumber] = result
			case domain.LOTTERY_CSV_UPLOAD_PROCESS_TYPE_LOSE:
				result, err := l.executeLoseForBatchTx(
					txCtx,
					adomain.LotteryLoseRequest{
						JWTUser:              actor,
						LotteryApplicationID: detail.LotteryApplicationID,
					},
					fmt.Sprintf(
						"lottery csv lose job_id=%s application_id=%s requested_by=%s",
						escapeCSVFormulaString(job.JobID),
						escapeCSVFormulaString(detail.LotteryApplicationID),
						escapeCSVFormulaString(job.RequestedBy),
					),
				)
				if err != nil {
					return errors.Wrapf(err, "row %d", detail.RowNumber)
				}
				loseResults[detail.RowNumber] = result
			default:
				return errors.Wrapf(cdomain.ErrBadRequest, "row %d has unsupported process type", detail.RowNumber)
			}
		}
		return nil
	})
	if err != nil {
		jobMessage := err.Error()
		if updateErr := updateLotteryCSVUploadJobFailure(ctx, job, pendingDetails, nil, jobMessage); updateErr != nil {
			return updateErr
		}
		return err
	}

	processedAt := jsonutils.JSONTime(time.Now())
	for i := range pendingDetails {
		detail := pendingDetails[i]
		detail.Status = domain.LOTTERY_CSV_UPLOAD_DETAIL_STATUS_SUCCESS
		detail.ProcessedAt = processedAt
		detail.ErrorCode = nil
		detail.ErrorMessage = nil
		if result, ok := winResults[detail.RowNumber]; ok {
			if result.createdShare != nil {
				detail.CreatedShareID = &result.createdShare.Id
			}
			if result.createdDepositResult != nil {
				detail.CreatedDepositRequestID = &result.createdDepositResult.Id
			}
		}
		if err := detail.Update(ctx); err != nil {
			return errors.Wrap(err, "failed to update lottery csv upload job detail as success")
		}
	}

	job.SuccessRowCount = len(pendingDetails)
	job.FailedRowCount = 0
	job.Status = domain.LOTTERY_CSV_UPLOAD_JOB_STATUS_COMPLETED
	job.FinishedAt = processedAt
	job.ErrorMessage = nil
	if err := job.Update(ctx); err != nil {
		return errors.Wrap(err, "failed to update lottery csv upload job as completed")
	}

	for _, result := range winResults {
		l.sendWinningNotification(ctx, result)
	}
	for _, result := range loseResults {
		l.sendLoseNotification(ctx, result)
	}

	return nil
}
