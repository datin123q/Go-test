package investment_waiters

import (
	"context"
	"fmt"
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/idgenerator"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"

	"github.com/pkg/errors"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

var investmentWaitersCSVColumns = []string{
	"model.investment_waiters.turn",         // 待機番号
	"model.investment_waiters.email",        // メールアドレス(ID)
	"model.investment_waiters.name",         // 投資家名
	"model.investment_waiters.phone",        // 電話番号
	"model.investment_waiters.qty",          // 投資キャンセル待ち申請口数
	"model.investment_waiters.amount",       // 申込金額
	"model.investment_waiters.status",       // ステータス
	"model.investment_waiters.waiting_date", // 申請日時
}

type investmentWaitersService struct {
	cfg                             *config.Config
	idGenerator                     idgenerator.IdGenerator
	translator                      i18nutils.Translator
	userRepository                  cdomain.UserRepository
	projectRepository               cdomain.ProjectRepository
	investmentWaitersRepository     cdomain.InvestmentWaitersRepository
	investmentWaitHistoryRepository cdomain.InvestmentWaitHistoryRepository
	investmentRequestRepository     cdomain.InvestmentRequestRepository
	investmentWaitHistoryService    cdomain.InvestmentWaitHistoryService
	mailService                     cdomain.MailService
}

func New(
	cfg *config.Config,
	idGenerator idgenerator.IdGenerator,
	translator i18nutils.Translator,
	userRepository cdomain.UserRepository,
	projectRepository cdomain.ProjectRepository,
	investmentWaitersRepository cdomain.InvestmentWaitersRepository,
	investmentRequestRepository cdomain.InvestmentRequestRepository,
	investmentWaitHistoryRepository cdomain.InvestmentWaitHistoryRepository,
	investmentWaitHistoryService cdomain.InvestmentWaitHistoryService,
	mailService cdomain.MailService,
) cdomain.InvestmentWaitersService {
	return &investmentWaitersService{
		cfg:                             cfg,
		idGenerator:                     idGenerator,
		translator:                      translator,
		userRepository:                  userRepository,
		projectRepository:               projectRepository,
		investmentWaitersRepository:     investmentWaitersRepository,
		investmentRequestRepository:     investmentRequestRepository,
		investmentWaitHistoryRepository: investmentWaitHistoryRepository,
		investmentWaitHistoryService:    investmentWaitHistoryService,
		mailService:                     mailService,
	}
}

func (i investmentWaitersService) ExportCSV(items []domain.InvestmentWaiters) cdomain.CSVData {
	csvHeaders := i.translator.Translates(investmentWaitersCSVColumns)

	records := [][]string{csvHeaders}

	for _, item := range items {
		// 投資待ち者ステータスが再調整される場合、ステータス値が変更される必要があります
		status := item.GetDisplayStatus()

		record := []string{
			fmt.Sprint(item.WaitingTurn),
			item.UserEmail,
			item.UserName,
			item.UserMobilePhone,
			fmt.Sprint(item.Qty),
			fmt.Sprint(item.Amount),
			i.translator.Translate("commoncode.INVESTMENT_WAITING_STATUS." + string(status)),
			item.WaitingDate.ToFormat(time.DateTime),
		}

		records = append(records, record)
	}

	return records
}

func (i investmentWaitersService) ChangeDepositRequestStatus(ctx context.Context, investmentWaiters domain.InvestmentWaiters, userId int64) error {
	updatedInvestmentWaiters := investmentWaiters
	updatedInvestmentWaiters.Status = domain.DEPOSIT_REQUEST

	if _, err := i.investmentWaitersRepository.Update(ctx, updatedInvestmentWaiters); err != nil {
		return errors.Wrap(err, "failed to update investment waiters")
	}

	if err := i.investmentWaitHistoryService.CreateStatusHistory(
		ctx,
		domain.InvestmentWaitersComparer{Before: investmentWaiters, After: updatedInvestmentWaiters},
		domain.USER_HISTORY_PROCESS_OPERATOR,
		userId,
	); err != nil {
		return errors.Wrap(err, "failed to create status history")
	}

	return nil
}

func (i investmentWaitersService) ChangeWaitingEndStatus(ctx context.Context, investmentWaiters domain.InvestmentWaiters, userId int64) error {
	updatedInvestmentWaiters := investmentWaiters
	updatedInvestmentWaiters.Status = domain.WAITING_END

	affected, err := i.investmentWaitersRepository.Update(ctx, updatedInvestmentWaiters)
	if err != nil {
		return errors.Wrap(err, "failed to update investment waiters")
	}
	if affected == 0 {
		return cdomain.ErrItemNotAffected
	}

	investmentRequest, err := i.investmentRequestRepository.GetByInvestmentWaitersId(ctx, investmentWaiters.Id)
	if err != nil && !errors.Is(err, cdomain.ErrItemNotFound) {
		return errors.Wrap(err, "failed to get investment request")
	}
	if err == nil {
		if err = i.investmentRequestRepository.Update(ctx, investmentRequest.WithIsExpire(true)); err != nil {
			return errors.Wrap(err, "failed to update investment request")
		}
	}

	if err = i.investmentWaitHistoryService.CreateStatusHistory(
		ctx,
		domain.InvestmentWaitersComparer{Before: investmentWaiters, After: updatedInvestmentWaiters},
		domain.USER_HISTORY_PROCESS_OPERATOR,
		userId,
	); err != nil {
		return errors.Wrap(err, "failed to create status history")
	}

	return nil
}

func (i investmentWaitersService) ChangeProposalStatus(ctx context.Context, investmentWaiters domain.InvestmentWaiters, userId int64) error {
	var requestId = i.idGenerator.GenUuidV4().String()

	updatedInvestmentWaiters := investmentWaiters
	updatedInvestmentWaiters.Status = domain.PROPOSAL

	if _, err := i.investmentWaitersRepository.Update(ctx, updatedInvestmentWaiters); err != nil {
		return errors.Wrap(err, "failed to update investment waiters")
	}

	user, err := i.userRepository.GetById(ctx, investmentWaiters.UserId)
	if err != nil {
		return errors.Wrap(err, "failed to get user")
	}

	project, err := i.projectRepository.GetById(ctx, investmentWaiters.ProjectId)
	if err != nil {
		return errors.Wrap(err, "failed to get project")
	}

	url := fmt.Sprintf("%s/product/view/%v?request_id=%v", i.cfg.ServiceConfig.Url.UserServiceDomain, project.Id, requestId)

	if _, err = i.investmentRequestRepository.Create(ctx, domain.InvestmentRequest{
		InvestmentWaitersId: investmentWaiters.Id,
		ProjectId:           project.Id,
		RequestId:           requestId,
		Url:                 url,
		IsExpire:            false,
		Email:               user.Email,
		Qty:                 project.RemainingQty,
	}); err != nil {
		return errors.Wrap(err, "failed to create investment request")
	}

	// メール送信

	if err = i.mailService.SendMail(ctx, cdomain.SendMailRequest{
		TemplateType: domain.INVEST_OFFER_TEMPLATE,
		Sender:       i.cfg.ServiceConfig.Mail.Info,
		Receivers:    []string{user.Email},
		MailType:     domain.MAIL_SEND_PRIORITY_NORMAL,
		Details: map[string]interface{}{
			"Receiver":      user.Name,
			"ServiceUrl":    i.cfg.ServiceConfig.Url.UserServiceDomain,
			"InvestedQty":   message.NewPrinter(language.Japanese).Sprintf("%d", investmentWaiters.Qty),
			"InvestableQty": message.NewPrinter(language.Japanese).Sprintf("%d", project.RemainingQty),
			"ProductView":   url,
		},
	}); err != nil {
		return errors.Wrap(err, "failed to send mail")
	}

	if err = i.investmentWaitHistoryService.CreateStatusHistory(
		ctx,
		domain.InvestmentWaitersComparer{Before: investmentWaiters, After: updatedInvestmentWaiters},
		domain.USER_HISTORY_PROCESS_OPERATOR,
		userId,
	); err != nil {
		return errors.Wrap(err, "failed to create status history")
	}

	if err = i.investmentWaitHistoryRepository.Create(ctx, domain.InvestmentWaitHistory{
		InvestmentWaitId: investmentWaiters.Id,
		SubstanceType:    string(domain.INVESTMENT_WAIT_HISTORY_MAIL),
		Name:             user.Name,
		Email:            user.Email,
		MailSendDate:     jsonutils.JSONTime(time.Now()),
		ProcessMethod:    domain.USER_HISTORY_PROCESS_OPERATOR,
	}); err != nil {
		return errors.Wrap(err, "failed to create mail send history")
	}

	return nil
}

func (i investmentWaitersService) ChangeCancelledStatus(ctx context.Context, investmentWaiters domain.InvestmentWaiters, userId int64) error {
	if investmentWaiters.Status != domain.WAITING {
		return errors.Wrapf(cdomain.ErrBadRequest, "investment waiter status is not WAITING (current: %s)", investmentWaiters.Status)
	}

	// ステータス変更-> キャンセル
	updatedInvestmentWaiters := investmentWaiters
	updatedInvestmentWaiters.Status = domain.CANCELLED
	if _, err := i.investmentWaitersRepository.Update(ctx, updatedInvestmentWaiters); err != nil {
		return errors.Wrap(err, "failed to update investment waiters")
	}

	if err := i.investmentWaitHistoryService.CreateStatusHistory(
		ctx,
		domain.InvestmentWaitersComparer{Before: investmentWaiters, After: updatedInvestmentWaiters},
		domain.USER_HISTORY_PROCESS_OPERATOR,
		userId,
	); err != nil {
		return errors.Wrap(err, "failed to create status history")
	}

	if err := factory.AfterCommit(ctx, func(notificationCtx context.Context) error {
		user, err := i.userRepository.GetById(notificationCtx, investmentWaiters.UserId)
		if err != nil {
			return errors.Wrap(err, "failed to get user")
		}
		project, err := i.projectRepository.GetById(notificationCtx, investmentWaiters.ProjectId)
		if err != nil {
			return errors.Wrap(err, "failed to get project")
		}

		now := time.Now()
		if err = i.mailService.SendMail(notificationCtx, cdomain.SendMailRequest{
			TemplateType: domain.INVEST_WAITLIST_CANCEL_TEMPLATE,
			Sender:       i.cfg.ServiceConfig.Mail.Info,
			Receivers:    []string{user.Email},
			MailType:     domain.MAIL_SEND_PRIORITY_NORMAL,
			Details: map[string]interface{}{
				"Receiver":    user.Name,
				"ServiceUrl":  i.cfg.ServiceConfig.Url.UserServiceDomain,
				"ProjectName": project.Name,
				"Amount":      message.NewPrinter(language.Japanese).Sprintf("%d", investmentWaiters.Amount),
				"CancelTime":  jsonutils.JSONTime(now).ToString(),
			},
		}); err != nil {
			return errors.Wrap(err, "failed to send mail")
		}

		if err = i.investmentWaitHistoryRepository.Create(notificationCtx, domain.InvestmentWaitHistory{
			InvestmentWaitId: investmentWaiters.Id,
			SubstanceType:    string(domain.INVESTMENT_WAIT_HISTORY_MAIL),
			Name:             user.Name,
			Email:            user.Email,
			MailSendDate:     jsonutils.JSONTime(now),
			ProcessMethod:    domain.USER_HISTORY_PROCESS_OPERATOR,
		}); err != nil {
			return errors.Wrap(err, "failed to create mail send history")
		}

		return nil
	}); err != nil {
		return errors.Wrap(err, "failed to notify investment waiter cancellation")
	}

	return nil
}
