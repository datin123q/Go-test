package share

import (
	"context"
	stderrors "errors"
	"strconv"
	"time"

	"github.com/samber/lo"

	"github.com/pkg/errors"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

type shareService struct {
	cfg                          *config.Config
	shareRepository              cdomain.ShareRepository
	adminUserRepository          cdomain.AdminUserRepository
	userRepository               cdomain.UserRepository
	userAccountRepository        cdomain.UserAccountRepository
	userVirtualAccountRepository cdomain.UserVirtualAccountRepository
	projectRepository            cdomain.ProjectRepository
	spvRepository                cdomain.SpvRepository
	contractRepository           cdomain.ContractRepository
	operatorHistoryRepository    cdomain.OperatorHistoryRepository
	depositRequestRepository     cdomain.DepositRequestRepository
	depositRequestService        cdomain.DepositRequestService
	eventMessageService          cdomain.EventMessageService
	projectService               cdomain.ProjectService
	mailService                  cdomain.MailService
}

func New(
	cfg *config.Config,
	shareRepository cdomain.ShareRepository,
	adminUserRepository cdomain.AdminUserRepository,
	userRepository cdomain.UserRepository,
	userAccountRepository cdomain.UserAccountRepository,
	userVirtualAccountRepository cdomain.UserVirtualAccountRepository,
	projectRepository cdomain.ProjectRepository,
	spvRepository cdomain.SpvRepository,
	contractRepository cdomain.ContractRepository,
	operatorHistoryRepository cdomain.OperatorHistoryRepository,
	depositRequestRepository cdomain.DepositRequestRepository,
	depositRequestService cdomain.DepositRequestService,
	projectService cdomain.ProjectService,
	eventMessageService cdomain.EventMessageService,
	mailService cdomain.MailService,
) cdomain.ShareService {
	return &shareService{
		cfg:                          cfg,
		shareRepository:              shareRepository,
		adminUserRepository:          adminUserRepository,
		userRepository:               userRepository,
		userAccountRepository:        userAccountRepository,
		userVirtualAccountRepository: userVirtualAccountRepository,
		projectRepository:            projectRepository,
		spvRepository:                spvRepository,
		contractRepository:           contractRepository,
		operatorHistoryRepository:    operatorHistoryRepository,
		depositRequestRepository:     depositRequestRepository,
		depositRequestService:        depositRequestService,
		projectService:               projectService,
		eventMessageService:          eventMessageService,
		mailService:                  mailService,
	}
}

func (s shareService) RequestInvest(ctx context.Context, sharedId string, userId int64) error {
	share, err := s.shareRepository.GetById(ctx, sharedId)
	if err != nil {
		return errors.Wrap(err, "fail to get share")
	}

	depositRequest, err := s.depositRequestRepository.GetDepositRequestByShareId(ctx, sharedId)
	if err != nil {
		return errors.Wrap(err, "fail to get deposit request")
	}

	project, err := s.projectRepository.GetById(ctx, share.ProjectId)
	if err != nil {
		return errors.Wrap(err, "fail to get project")
	}

	owner, err := s.userRepository.GetById(ctx, share.OwnerId)
	if err != nil {
		return errors.Wrap(err, "fail to get user")
	}

	userAccount, err := s.userAccountRepository.GetByUserId(ctx, owner.Id)

	userVirtualAccount, err := s.userVirtualAccountRepository.GetByUserIdAndLinkedStatus(ctx, owner.Id, project.ProjectType, domain.ON_VA_ACCOUNT)
	if err != nil {
		return errors.Wrap(err, "fail to get user virtual account")
	}

	share.InvestStatus = domain.REQUEST_INVEST
	if err = s.shareRepository.Update(ctx, share); err != nil {
		return errors.Wrap(err, "fail to update share")
	}

	investRequestDetails := map[string]interface{}{
		"Receiver":         owner.Name,
		"ServiceUrl":       s.cfg.ServiceConfig.Url.UserServiceDomain,
		"ProjectName":      project.Name,
		"TransactionDate":  jsonutils.JSONTime.ToString(share.InvestDate),
		"PaymentEnd":       jsonutils.JSONTime.ToString(depositRequest.ExpireDepositDate),
		"InvestmentAmount": message.NewPrinter(language.Japanese).Sprintf("%d", share.Amount),
		"BankName":         depositRequest.DepositBankName,
		"BranchName":       depositRequest.DepositBranchName,
		"AccountNumber":    depositRequest.DepositAccountNumber,
		"VaHolderName":     userVirtualAccount.VaHolderNameKana,
	}

	investRequestMailRequest := cdomain.SendMailRequest{
		Sender:    s.cfg.ServiceConfig.Mail.Info,
		Receivers: []string{owner.Email},
		MailType:  domain.MAIL_SEND_PRIORITY_NORMAL,
		Details:   investRequestDetails,
	}

	// プロジェクト型がEQUITYで、デポジット金が0円の場合
	if project.ProjectType == domain.EQUITY && lo.FromPtr(userAccount.Balance) == 0 {
		investRequestMailRequest.TemplateType = domain.INVEST_REQUEST_EQUITY_TEMPLATE

		if err = s.mailService.SendMail(ctx, investRequestMailRequest); err != nil {
			return errors.Wrap(err, "fail to send mail")
		}
	}

	// 残高が十分でない場合、残高不足メール送信
	if depositRequest.NeedAmount() > lo.FromPtr(userAccount.Balance) {
		lackAmount := share.Amount - (lo.FromPtr(userAccount.Balance) + depositRequest.DepositAmount)

		investInsufficientBalanceDetails := map[string]interface{}{
			"Receiver":    owner.Name,
			"ServiceUrl":  s.cfg.ServiceConfig.Url.UserServiceDomain,
			"LackAmount":  message.NewPrinter(language.Japanese).Sprintf("%d", lackAmount),
			"ProductName": project.Name,
			// 변경사항 적용 https://github.com/illuminarean/energy-shares-backend-repo/commit/d68f482e77c2b6f5137eba205ceebd94e7d88e3d
			"TransactionDate":  jsonutils.JSONTime.ToString(share.InvestDate),
			"BankName":         depositRequest.DepositBankName,
			"BranchName":       depositRequest.DepositBranchName,
			"AccountNumber":    depositRequest.DepositAccountNumber,
			"Depositor":        userVirtualAccount.VaHolderNameKana,
			"InvestmentAmount": message.NewPrinter(language.Japanese).Sprintf("%d", share.Amount),
			"PaymentEnd":       jsonutils.JSONTime.ToString(depositRequest.ExpireDepositDate),
		}

		if err = s.mailService.SendMail(ctx, cdomain.SendMailRequest{
			TemplateType: domain.INVEST_INSUFFICIENT_BALANCE_TEMPLATE,
			Sender:       s.cfg.ServiceConfig.Mail.Info,
			Receivers:    []string{owner.Email},
			MailType:     domain.MAIL_SEND_PRIORITY_NORMAL,
			Details:      investInsufficientBalanceDetails,
		}); err != nil {
			return errors.Wrap(err, "fail to send mail")
		}
	}

	// イベント発行
	if err = s.eventMessageService.PublishEvent(ctx, domain.EVENT_INVESTMENT_REQUEST, domain.InvestmentRequestDto{
		UserId:         owner.Id,
		UserName:       owner.Name,
		InvestmentId:   share.Id,
		ProductName:    project.Name,
		ProductId:      project.Id,
		Recruitment:    project.OfferingType.ToRecruitment(),
		OrderDateTime:  share.InvestDate.Time(),
		OrderQuantity:  share.Qty,
		OrderUnitPrice: share.Price,
		OrderAmount:    share.Amount,
	}); err != nil {
		return errors.Wrap(err, "fail to publish event")
	}

	return nil
}

func (s shareService) SuccessInvest(ctx context.Context, sharedId string, userId int64) error {
	share, err := s.shareRepository.GetById(ctx, sharedId)
	if err != nil {
		return errors.Wrap(err, "fail to get share")
	}

	depositRequest, err := s.depositRequestRepository.GetDepositRequestByShareId(ctx, sharedId)
	if err != nil {
		return errors.Wrap(err, "fail to get deposit request")
	}

	project, err := s.projectRepository.GetById(ctx, share.ProjectId)
	if err != nil {
		return errors.Wrap(err, "fail to get project")
	}

	spv, err := s.spvRepository.GetById(ctx, project.SpvId)
	if err != nil {
		return errors.Wrap(err, "fail to get spv")
	}

	owner, err := s.userRepository.GetById(ctx, share.OwnerId)
	if err != nil {
		return errors.Wrap(err, "fail to get owner")
	}

	actor, err := s.adminUserRepository.GetById(ctx, userId)
	if err != nil {
		return errors.Wrap(err, "fail to get actor")
	}

	// 入金要求確認処理
	if err = s.depositRequestService.OkDeposit(ctx, depositRequest); err != nil {
		return errors.Wrap(err, "fail to ok deposit")
	}

	// 投資更新
	share.InvestStatus = domain.SUCCESS_INVEST
	share.UpdateId = userId
	share.CommitmentDate = jsonutils.JSONTime(time.Now()) // コミットメント日

	if err = s.shareRepository.Update(ctx, share); err != nil {
		return errors.Wrap(err, "fail to update share")
	}

	// 契約書作成
	if _, err = s.contractRepository.Create(ctx, domain.Contract{
		ShareId:     share.Id,
		ShareQty:    share.Qty,
		ShareAmount: share.Amount,
		Status:      domain.CONTRACT_NONE,
		ProjectId:   project.Id,
		UserName:    owner.Name,
	}); err != nil {
		return errors.Wrap(err, "fail to create contract")
	}

	// 投資申請承認ログ適用
	if err = s.operatorHistoryRepository.Create(ctx, domain.OperatorHistory{
		UserId: userId,
		Email:  actor.Email,
		Menu:   "commoncode.INVEST_STATUS.SUCCESS_INVEST",
		Ip:     "::1",
		State:  domain.CHANGE,
	}); err != nil {
		return errors.Wrap(err, "fail to create operator history")
	}

	// 更新されたshare/depositRequestを使用してメールとイベントを送信
	updatedShare, err := s.shareRepository.GetById(ctx, sharedId)
	if err != nil {
		return errors.Wrap(err, "fail to get updated share")
	}

	updatedDepositRequest, err := s.depositRequestRepository.GetDepositRequestByShareId(ctx, sharedId)
	if err != nil {
		return errors.Wrap(err, "fail to get updated deposit request")
	}

	// 投資完了メール送信(入金確認時に自動処理/オペレーター処理)
	if err = s.mailService.SendMail(ctx, cdomain.SendMailRequest{
		TemplateType: domain.INVEST_SUCCESS_TEMPLATE,
		Sender:       s.cfg.ServiceConfig.Mail.Info,
		Receivers:    []string{owner.Email},
		MailType:     domain.MAIL_SEND_PRIORITY_NORMAL,
		Details: map[string]interface{}{
			"Receiver":      owner.Name,
			"ServiceUrl":    s.cfg.ServiceConfig.Url.UserServiceDomain,
			"ProjectName":   project.Name,
			"TradingTime":   updatedShare.Updated.Time().Format(time.DateTime),
			"DepositDate":   updatedDepositRequest.DepositDate.ToString(),
			"DepositAmount": strconv.FormatInt(updatedDepositRequest.DepositAmount, 10)},
	}); err != nil {
		return errors.Wrap(err, "fail to send mail")
	}

	// 投資成功イベント発行
	if err = s.eventMessageService.PublishEvent(ctx, domain.EVENT_INVESTMENT_COMPLETED, domain.InvestmentCompletedDto{
		UserId:          updatedShare.OwnerId,
		UserName:        owner.Name,
		InvestmentId:    updatedShare.Id,
		CompanyName:     spv.Name,
		ProductName:     project.Name,
		ProductId:       project.Id,
		Recruitment:     project.OfferingType.ToRecruitment(),
		CommitDateTime:  updatedDepositRequest.DepositDate.Time(),
		CommitQuantity:  updatedShare.Qty,
		CommitUnitPrice: updatedShare.Price,
		CommitAmount:    updatedShare.Amount,
	}); err != nil {
		return errors.Wrap(err, "fail to publish event")
	}

	return nil
}

func (s shareService) CancelInvest(ctx context.Context, sharedId string, userId int64) error {
	share, err := s.shareRepository.GetById(ctx, sharedId)
	if err != nil {
		return errors.Wrap(err, "fail to get share")
	}
	if !share.CanCancelInvest() {
		return errors.Wrapf(cdomain.ErrBadRequest, "investment cannot be canceled from status %s", share.InvestStatus)
	}

	project, err := s.projectRepository.GetById(ctx, share.ProjectId)
	if err != nil {
		return errors.Wrap(err, "fail to get project")
	}

	depositRequest, err := s.depositRequestRepository.GetDepositRequestByShareId(ctx, sharedId)
	if err != nil {
		return errors.Wrap(err, "fail to get deposit request")
	}

	actor, err := s.adminUserRepository.GetById(ctx, userId)
	if err != nil {
		return errors.Wrap(err, "fail to get user")
	}

	// 入金要求キャンセル処理
	if err = s.depositRequestService.CancelDeposit(ctx, depositRequest); err != nil {
		return errors.Wrap(err, "fail to cancel equity deposit")
	}

	// 投資キャンセル後のShare更新
	share.InvestStatus = domain.CANCEL_INVEST
	share.UpdateId = userId
	share.CommitmentDate = jsonutils.JSONTime(time.Now()) // コミットメント日

	if err = s.shareRepository.Update(ctx, share); err != nil {
		return errors.Wrap(err, "fail to update share")
	}

	isUpdated := false
	if share.ShareType == domain.MANUAL_SHARE {
		project.IsManualInvest = false
		isUpdated = true
	}

	// キャンセル後、キャンセル待ち者が存在しない場合は、キャンセル待ち状態を解除する
	if project.IsWaitInvest {
		waiters, err := domain.InvestmentWaiters{}.GetByProjectIdAndStatuses(ctx, project.Id, []domain.INVESTMENT_WAITING_TYPE{domain.WAITING, domain.DEPOSIT_REQUEST, domain.PROPOSAL})
		if err != nil {
			return errors.Wrap(err, "fail to get investment waiters")
		}
		if len(waiters) == 0 {
			checkProject, err := domain.Project{}.GetById(ctx, project.Id)
			if err != nil {
				return err
			}
			if !checkProject.HasReachedTarget() && checkProject.ProjectStatus != domain.END_FUNDING {
				// キャンセル待ち者が存在せず、かつ募集条件に到達していない場合は、キャンセル待ち状態を解除する
				// ただし、募集終了ステータスの場合はキャンセル待ち状態を維持する
				project.IsWaitInvest = false
				isUpdated = true
			}
		}
	}

	if isUpdated {
		if err = s.projectRepository.Update(ctx, project); err != nil {
			return errors.Wrap(err, "fail to update project")
		}
	}

	// 投資申請キャンセルログ適用
	if err = s.operatorHistoryRepository.Create(ctx, domain.OperatorHistory{
		UserId: userId,
		Email:  actor.Email,
		Menu:   "commoncode.INVEST_STATUS.CANCEL_INVEST",
		Ip:     "::1",
		State:  domain.CHANGE,
	}); err != nil {
		return errors.Wrap(err, "fail to create operator history")
	}

	// 投資キャンセルイベント発行。 DepositCheckStatusがCANCEL_DEPOSITの状態でここに来ることはできません。
	category := domain.EVENT_INVESTMENT_CANCELED
	if depositRequest.DepositCheckStatus == domain.NO_DEPOSIT {
		category = domain.EVENT_INVESTMENT_VOIDED
	}

	mailErr := factory.AfterCommit(ctx, func(notificationCtx context.Context) error {
		owner, err := s.userRepository.GetById(notificationCtx, share.OwnerId)
		if err != nil {
			return errors.Wrap(err, "fail to get owner")
		}

		// プロジェクト状態がINVALID_FUNDINGの場合は"無効な投資テンプレート"、それ以外は"投資キャンセルテンプレート"
		templateType := lo.If(project.ProjectStatus == domain.INVALID_FUNDING, domain.INVEST_INVALIDITY_TEMPLATE).Else(domain.INVEST_CANCEL_TEMPLATE)

		if err = s.mailService.SendMail(notificationCtx, cdomain.SendMailRequest{
			TemplateType: templateType,
			Sender:       s.cfg.ServiceConfig.Mail.Info,
			Receivers:    []string{owner.Email},
			MailType:     domain.MAIL_SEND_PRIORITY_NORMAL,
			Details: map[string]interface{}{
				"Receiver":    owner.Name,
				"ServiceUrl":  s.cfg.ServiceConfig.Url.UserServiceDomain,
				"ProjectName": project.Name,
				"TradingTime": share.Updated.ToString(),
				"CancelTime":  jsonutils.JSONTime(time.Now()).ToString(),
			},
		}); err != nil {
			return errors.Wrap(err, "fail to send mail")
		}

		return nil
	})
	eventErr := factory.AfterCommit(ctx, func(notificationCtx context.Context) error {
		if err := s.eventMessageService.PublishEvent(notificationCtx, category, domain.InvestmentCancelDto{
			RequestDatetime: share.CommitmentDate.Time(),
			InvestmentId:    share.Id,
		}); err != nil {
			return errors.Wrap(err, "fail to publish event")
		}

		return nil
	})
	if notificationErr := stderrors.Join(mailErr, eventErr); notificationErr != nil {
		return errors.Wrap(notificationErr, "fail to notify investment cancellation")
	}

	return nil
}

func (s shareService) FailInvest(ctx context.Context, shareId string, userId int64) error {
	depositRequest, err := s.depositRequestRepository.GetDepositRequestByShareId(ctx, shareId)
	if err != nil {
		return errors.Wrap(err, "fail to get deposit request")
	}

	actor, err := s.adminUserRepository.GetById(ctx, userId)
	if err != nil {
		return errors.Wrap(err, "fail to get actor")
	}

	share, err := s.shareRepository.GetById(ctx, shareId)
	if err != nil {
		return errors.Wrap(err, "fail to get share")
	}

	// 입금 요청 취소 처리
	if err = s.depositRequestService.CancelDeposit(ctx, depositRequest); err != nil {
		return errors.Wrap(err, "fail to cancel equity deposit")
	}

	share.InvestStatus = domain.FAIL_INVEST
	share.UpdateId = userId
	share.CommitmentDate = jsonutils.JSONTime(time.Now()) // コミットメント日

	if err = s.shareRepository.Update(ctx, share); err != nil {
		return errors.Wrap(err, "fail to update share")
	}

	if err = s.operatorHistoryRepository.Create(ctx, domain.OperatorHistory{
		UserId: actor.Id,
		Email:  actor.Email,
		Menu:   "commoncode.INVEST_STATUS.FAIL_INVEST",
		Ip:     "::1",
		State:  domain.CHANGE,
	}); err != nil {
		return errors.Wrap(err, "fail to create operator history")
	}

	// 投資キャンセルイベント発行。 DepositCheckStatusがCANCEL_DEPOSITの状態でここに来ることはできません。
	category := domain.EVENT_INVESTMENT_CANCELED
	if depositRequest.DepositCheckStatus == domain.NO_DEPOSIT {
		category = domain.EVENT_INVESTMENT_VOIDED
	}

	if err = s.eventMessageService.PublishEvent(ctx, category, domain.InvestmentCancelDto{
		RequestDatetime: share.CommitmentDate.Time(),
		InvestmentId:    share.Id,
	}); err != nil {
		return errors.Wrap(err, "fail to publish event")
	}

	return nil
}
