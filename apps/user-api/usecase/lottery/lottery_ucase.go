package lottery

import (
	"context"
	"fmt"
	"time"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/actorcontext"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	udomain "github.com/illuminarean/energyshares-user-api/domain"
	"github.com/pkg/errors"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

const LIMITED_AMOUNT int64 = 5000000

type lotteryUsecase struct {
	cfg                      *config.Config
	txManager                factory.TxManager
	lotteryValidationService cdomain.LotteryValidationService
	lotteryRepository        cdomain.LotteryApplicationRepository
	lotteryHistoryRepository cdomain.LotteryHistoryRepository
	eventMessageService      cdomain.EventMessageService
	mailService              cdomain.MailService
	userRepository           cdomain.UserRepository
	virtualAccountRepository cdomain.UserVirtualAccountRepository
	projectRepository        cdomain.ProjectRepository
}

func New(cfg *config.Config,
	txManager factory.TxManager,
	lotteryValidationService cdomain.LotteryValidationService,
	lotteryRepository cdomain.LotteryApplicationRepository,
	lotteryHistoryRepository cdomain.LotteryHistoryRepository,
	eventMessageService cdomain.EventMessageService,
	mailService cdomain.MailService,
	userRepository cdomain.UserRepository,
	virtualAccountRepository cdomain.UserVirtualAccountRepository,
	projectRepository cdomain.ProjectRepository,
) udomain.LotteryUsecase {
	return &lotteryUsecase{
		txManager:                txManager,
		lotteryValidationService: lotteryValidationService,
		lotteryRepository:        lotteryRepository,
		lotteryHistoryRepository: lotteryHistoryRepository,
		eventMessageService:      eventMessageService,
		cfg:                      cfg,
		mailService:              mailService,
		userRepository:           userRepository,
		virtualAccountRepository: virtualAccountRepository,
		projectRepository:        projectRepository,
	}
}

func canCancelLotteryApplication(application domain.LotteryApplication, projectStatus domain.PROJECT_STATUS) bool {
	return application.CanMarkAsCanceled() && projectStatus == domain.OPEN_FUNDING
}

func (u lotteryUsecase) Application(ctx context.Context, request udomain.LotteryApplicationRequest) (udomain.LotteryApplicationResponse, error) {
	user, err := u.userRepository.GetById(ctx, request.UserID)
	if err != nil {
		return udomain.LotteryApplicationResponse{}, err
	}
	project, err := u.projectRepository.GetById(ctx, request.ProjectId)
	if err != nil {
		return udomain.LotteryApplicationResponse{}, err
	}

	if project.SelectionType != domain.PROJECT_SELECTION_LOTTERY {
		return udomain.LotteryApplicationResponse{}, cdomain.ErrBadRequest
	}

	// マイナンバーのステータスを確認
	if user.UserType == domain.PRIVATE {
		ok, err := u.lotteryValidationService.CheckMyNumberStatusByUserID(ctx, user.Id)
		if err != nil {
			return udomain.LotteryApplicationResponse{}, factory.ErrorDB.New(err)
		}
		if !ok {
			return udomain.LotteryApplicationResponse{}, factory.ErrorMyNumberNotRegistered.New(
				fmt.Errorf("fail my number load to user id:%v", user.Id),
			)
		}
	}

	// ブラックリストのステータスを確認
	if user.BlockInvest {
		return udomain.LotteryApplicationResponse{}, factory.ErrorBlockInvest.New(nil)
	}

	// 年齢確認
	if err := u.lotteryValidationService.CheckUserOverAge(ctx, &user); err != nil {
		return udomain.LotteryApplicationResponse{}, err
	}

	// プロジェクトの募集期間のみ申込可能とする
	if project.ProjectStatus != domain.OPEN_FUNDING {
		return udomain.LotteryApplicationResponse{}, cdomain.ErrBadRequest
	}

	amount := project.SharePrice * request.Qty

	// 投資家数500人未満の制限
	if project.IsInvestorLimit {
		successInvestorCount, err := u.lotteryValidationService.GetSuccessInvestorCountByProjectID(ctx, project.Id)
		if err != nil {
			return udomain.LotteryApplicationResponse{}, factory.ErrorDB.New(err)
		}

		if successInvestorCount >= int(domain.TargetInvestorCount) {
			return udomain.LotteryApplicationResponse{}, factory.ErrorOverInvestorCount.New(nil)
		}
	}

	// 投資金額の制限
	if project.IsAmountLimit {
		investedAmount, err := u.lotteryValidationService.GetInvestedAmountByUserID(ctx, project.Id, user.Id)
		if err != nil {
			return udomain.LotteryApplicationResponse{}, factory.ErrorParameter.New(err)
		}

		waitingAmount, err := u.lotteryValidationService.GetWaitingAmountByUserID(ctx, project.Id, user.Id)
		if err != nil {
			return udomain.LotteryApplicationResponse{}, factory.ErrorParameter.New(err)
		}

		totalAmount := amount + investedAmount + waitingAmount
		if totalAmount >= LIMITED_AMOUNT {
			return udomain.LotteryApplicationResponse{}, factory.ErrorOverLimitedAmount.New(nil)
		}
	}

	// 1ファンド1人に1抽選申込まで。キャンセルはカウントしない。
	alreadyApplied, err := u.lotteryRepository.IsAlreadyApplied(ctx, project.Id, user.Id)
	if err != nil {
		return udomain.LotteryApplicationResponse{}, factory.ErrorDB.New(err)
	}
	if alreadyApplied {
		return udomain.LotteryApplicationResponse{}, cdomain.ErrAlreadyExists
	}

	appliedAt := time.Now()
	if err = u.txManager.WithTx(ctx, func(txCtx context.Context) error {
		_, succeed, err := u.lotteryRepository.CreateIfNotExists(txCtx, domain.LotteryApplication{
			ProjectId: request.ProjectId,
			UserId:    request.UserID,
			Qty:       request.Qty,
			UnitPrice: project.SharePrice,
			Amount:    amount,
			AppliedAt: jsonutils.JSONTime(appliedAt),
		})
		if err != nil {
			return errors.Wrap(err, "failed to create lottery application")
		}
		if !succeed {
			return cdomain.ErrAlreadyExists
		}

		if err = u.eventMessageService.PublishEvent(txCtx, domain.EVENT_LOTTERY_REQUEST, domain.LotteryRequestDto{
			UserId:         user.Id,
			UserName:       user.Name,
			ProductName:    project.Name,
			ProductId:      project.Id,
			Recruitment:    project.OfferingType.ToRecruitment(),
			OrderDateTime:  appliedAt,
			OrderQuantity:  request.Qty,
			OrderUnitPrice: project.SharePrice,
			OrderAmount:    amount,
		}); err != nil {
			return errors.Wrap(err, "failed to publish lottery request event")
		}

		// 抽選申込金額は申込時点で project.final_funding_amount に累積する
		if err = u.projectRepository.AddFinalFundingAmount(txCtx, project.Id, amount); err != nil {
			return factory.ErrorDB.New(err)
		}

		return nil
	}); err != nil {
		if errors.Is(err, cdomain.ErrAlreadyExists) {
			return udomain.LotteryApplicationResponse{}, cdomain.ErrAlreadyExists
		}
		return udomain.LotteryApplicationResponse{}, err
	}

	err = u.mailService.SendMail(ctx, cdomain.SendMailRequest{
		TemplateType: domain.LOTTERY_APPLICATION_TEMPLATE,
		Sender:       u.cfg.ServiceConfig.Mail.Info,
		Receivers:    []string{user.Email},
		MailType:     domain.MAIL_SEND_PRIORITY_NORMAL,
		Details: map[string]interface{}{
			"ServiceUrl":        u.cfg.ServiceConfig.Url.UserServiceDomain,
			"Receiver":          user.Name,
			"ProjectName":       project.Name,
			"RequestDate":       jsonutils.JSONTime(appliedAt).ToString(),
			"SharesCount":       message.NewPrinter(language.Japanese).Sprintf("%d", request.Qty),
			"ApplicationAmount": message.NewPrinter(language.Japanese).Sprintf("%d", amount),
		},
	})
	if err != nil {
		factory.Logger(ctx).Errorf("failed to send mail: %v", err)
	}

	va, err := u.virtualAccountRepository.GetByUserIdAndLinkedStatus(ctx, user.Id, project.ProjectType, domain.ON_VA_ACCOUNT)
	if err != nil || va.Id == 0 {
		return udomain.LotteryApplicationResponse{}, errors.Wrap(err, "failed to get virtual account")
	}

	return udomain.LotteryApplicationResponse{
		ProjectName:       project.Name,
		BankName:          va.VaBankName,
		BranchName:        va.VaBranchName,
		BankAccountNumber: va.VaAccountNumber,
		Depositor:         va.VaHolderNameKana,
		Price:             project.SharePrice,
		Amount:            amount,
		Qty:               request.Qty,
	}, nil
}

func (u lotteryUsecase) CheckApplication(ctx context.Context, request udomain.LotteryCheckApplicationRequest) (udomain.LotteryCheckApplicationResponse, error) {
	hasApplied, err := u.lotteryRepository.IsAlreadyApplied(ctx, request.ProjectId, request.UserID)
	if err != nil {
		return udomain.LotteryCheckApplicationResponse{}, factory.ErrorDB.New(err)
	}

	return udomain.LotteryCheckApplicationResponse{
		HasApplied: hasApplied,
	}, nil
}

func (u lotteryUsecase) Cancel(ctx context.Context, request udomain.LotteryCancelRequest) (udomain.LotteryCancelResponse, error) {
	actor, ok := actorcontext.ActorFrom(ctx)
	if !ok {
		return udomain.LotteryCancelResponse{}, errors.Wrap(cdomain.ErrBadRequest, "actor is required")
	}

	err := u.txManager.WithTx(ctx, func(txCtx context.Context) error {
		lockedApplication, err := u.lotteryRepository.LockById(txCtx, request.LotteryApplicationID)
		if err != nil {
			return errors.Wrap(err, "failed to lock lottery application")
		}
		if lockedApplication.UserId != request.JWTUser.Id {
			return cdomain.ErrItemNotFound
		}
		lockedApplication.ComputeStatus()

		lockedProject, err := u.projectRepository.LockById(txCtx, lockedApplication.ProjectId)
		if err != nil {
			return errors.Wrap(err, "failed to lock project")
		}

		if !canCancelLotteryApplication(lockedApplication, lockedProject.ProjectStatus) {
			return errors.Wrap(cdomain.ErrBadRequest, "lottery application cannot be canceled")
		}

		success, err := u.lotteryRepository.MarkAsCanceled(txCtx, request.LotteryApplicationID)
		if err != nil {
			return errors.Wrap(err, "failed to cancel lottery application")
		}
		if !success {
			return errors.Wrap(cdomain.ErrBadRequest, "lottery application was not canceled")
		}

		if lockedProject.ProjectStatus == domain.OPEN_FUNDING {
			if err := u.projectRepository.ReduceFinalFundingAmount(txCtx, lockedProject.Id, lockedApplication.Amount); err != nil {
				return errors.Wrap(err, "failed to reduce final funding amount")
			}
		}

		if err := u.lotteryHistoryRepository.Create(txCtx, domain.LotteryHistory{
			ApplicationId: request.LotteryApplicationID,
			PrevStatus:    lockedApplication.Status,
			NewStatus:     domain.LOTTERY_STATUS_CANCELED,
			ActorId:       actor.Id,
		}); err != nil {
			return errors.Wrap(err, "failed to create lottery history")
		}

		return nil
	})
	if err != nil {
		return udomain.LotteryCancelResponse{}, err
	}

	application, err := u.lotteryRepository.GetById(ctx, request.LotteryApplicationID)
	if err != nil {
		factory.Logger(ctx).Errorf("failed to get canceled lottery application for mail: application_id=%s err=%v", request.LotteryApplicationID, err)
		return udomain.LotteryCancelResponse{}, nil
	}

	user, err := u.userRepository.GetById(ctx, application.UserId)
	if err != nil {
		factory.Logger(ctx).Errorf("failed to get user for lottery cancel mail: application_id=%s user_id=%d err=%v", request.LotteryApplicationID, application.UserId, err)
		return udomain.LotteryCancelResponse{}, nil
	}

	project, err := u.projectRepository.GetById(ctx, application.ProjectId)
	if err != nil {
		factory.Logger(ctx).Errorf("failed to get project for lottery cancel mail: application_id=%s project_id=%d err=%v", request.LotteryApplicationID, application.ProjectId, err)
		return udomain.LotteryCancelResponse{}, nil
	}

	err = u.mailService.SendMail(ctx, cdomain.SendMailRequest{
		TemplateType: domain.LOTTERY_APPLICATION_CANCEL_TEMPLATE,
		Sender:       u.cfg.ServiceConfig.Mail.Info,
		Receivers:    []string{user.Email},
		MailType:     domain.MAIL_SEND_PRIORITY_NORMAL,
		Details: map[string]interface{}{
			"ProjectName":       project.Name,
			"Receiver":          user.Name,
			"ApplicationAmount": message.NewPrinter(language.Japanese).Sprintf("%d", application.Amount),
			"CancelTime":        jsonutils.JSONTime.ToString(application.CanceledAt),
			"ServiceUrl":        u.cfg.ServiceConfig.Url.UserServiceDomain,
		},
	})
	if err != nil {
		factory.Logger(ctx).Errorf("failed to send lottery cancel mail: %v", err)
	}

	return udomain.LotteryCancelResponse{}, nil
}

func (u lotteryUsecase) GetApplications(ctx context.Context, request udomain.LotteryGetApplicationsRequest) (udomain.LotteryGetApplicationsResponse, error) {
	opts := cdomain.LotteryApplicationQueryOptionFunc()
	totalCount, applications, err := u.lotteryRepository.GetAll(ctx,
		// 同時刻の申込があるため、id で一意に定める。管理画面側も同じ並び。
		opts.WithSortBy("lottery_application.applied_at,lottery_application.id"),
		opts.WithOrder("desc,desc"),
		opts.WithUserId(request.UserID),
		opts.WithLimit(request.Limit),
		opts.WithOffset(request.Offset),
	)
	if err != nil {
		return udomain.LotteryGetApplicationsResponse{}, errors.Wrap(err, "failed to get lottery applications")
	}

	items := make([]udomain.LotteryGetApplicationsItem, len(applications))
	for i, application := range applications {
		items[i] = udomain.LotteryGetApplicationsItem{
			Id:                application.Id,
			ProjectId:         application.ProjectId,
			ProjectName:       application.ProjectName,
			ApplicationDate:   application.AppliedAt,
			LotteryStatus:     application.Status,
			ApplicationAmount: application.Amount,
			WinningAmount:     application.WinningAmount(),
			Cancelable:        canCancelLotteryApplication(application, application.ProjectStatus),
		}
	}

	return udomain.LotteryGetApplicationsResponse{
		Items:      items,
		TotalCount: totalCount,
		Limit:      request.Limit,
		Offset:     request.Offset,
	}, nil
}
