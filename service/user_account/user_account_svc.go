package user_account

import (
	"context"
	"time"

	"github.com/samber/lo"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

type userAccountService struct {
	cfg                                 *config.Config
	logger                              factory.LoggerFactory
	projectRepository                   cdomain.ProjectRepository
	shareRepository                     cdomain.ShareRepository
	userVirtualAccountRepository        cdomain.UserVirtualAccountRepository
	userRepository                      cdomain.UserRepository
	userAccountRepository               cdomain.UserAccountRepository
	userTransactionRepository           cdomain.UserTransactionRepository
	depositRequestRepository            cdomain.DepositRequestRepository
	depositRequestLackHistoryRepository cdomain.DepositRequestLackHistoryRepository
	depositRequestService               cdomain.DepositRequestService
	mailService                         cdomain.MailService
	shareService                        cdomain.ShareService
}

func New(
	cfg *config.Config,
	logger factory.LoggerFactory,
	projectRepository cdomain.ProjectRepository,
	shareRepository cdomain.ShareRepository,
	userVirtualAccountRepository cdomain.UserVirtualAccountRepository,
	userAccountRepository cdomain.UserAccountRepository,
	userRepository cdomain.UserRepository,
	userTransactionRepository cdomain.UserTransactionRepository,
	depositRequestRepository cdomain.DepositRequestRepository,
	depositRequestLackHistoryRepository cdomain.DepositRequestLackHistoryRepository,
	depositRequestService cdomain.DepositRequestService,
	mailService cdomain.MailService,
	shareService cdomain.ShareService,
) cdomain.UserAccountService {
	return &userAccountService{
		cfg:                                 cfg,
		logger:                              logger,
		projectRepository:                   projectRepository,
		shareRepository:                     shareRepository,
		userVirtualAccountRepository:        userVirtualAccountRepository,
		userAccountRepository:               userAccountRepository,
		userRepository:                      userRepository,
		userTransactionRepository:           userTransactionRepository,
		depositRequestRepository:            depositRequestRepository,
		depositRequestLackHistoryRepository: depositRequestLackHistoryRepository,
		depositRequestService:               depositRequestService,
		mailService:                         mailService,
		shareService:                        shareService,
	}
}

func (u userAccountService) DepositCheckProcess(ctx context.Context, depositRequest domain.DepositRequest) error {
	var plannedWithdrawalAmount int64 // 予定出金額

	// 入金要求額がない場合、該当する入金解決ロジックは実行されません。
	if depositRequest.NeedAmount() <= 0 {
		return nil
	}

	userAccount, err := u.userAccountRepository.GetByUserId(ctx, depositRequest.UserId)
	if err != nil {
		return errors.Wrap(err, "failed to get user account")
	}

	user, err := u.userRepository.GetById(ctx, depositRequest.UserId)
	if err != nil {
		return errors.Wrap(err, "failed to get user")
	}

	u.logger.Entry(ctx).WithFields(logrus.Fields{
		"user_account":    userAccount,
		"deposit_request": depositRequest,
		"need_account":    depositRequest.NeedAmount(),
		"balance":         userAccount.Balance,
	}).Info("start user account deposit check process share id - > ", depositRequest.ShareId)

	// 残高が十分な場合のみ取引が発生します。(残金 > 入金必要額)
	if lo.FromPtr(userAccount.Balance) >= depositRequest.NeedAmount() {
		plannedWithdrawalAmount = depositRequest.NeedAmount() // 予定出金額は入金要求額と同じです。

		if err = u.depositRequestService.OkDeposit(ctx, depositRequest); err != nil {
			return errors.Wrap(err, "failed to update deposit request")
		}

		// システムで自動的に実行されるため、Actor(ユーザー)は1で固定して送信します。
		if err = u.shareService.SuccessInvest(ctx, depositRequest.ShareId, 1); err != nil {
			return errors.Wrap(err, "failed to success invest")
		}

		userAccount.Balance = lo.ToPtr(lo.FromPtr(userAccount.Balance) - plannedWithdrawalAmount) // 残金から予定出金額を差し引きます。

		if err = u.userAccountRepository.Update(ctx, userAccount); err != nil {
			return errors.Wrap(err, "failed to update user account")
		}

	} else { // 残金が十分でない場合
		plannedWithdrawalAmount = lo.FromPtr(userAccount.Balance) // 予定出金額は残金と同じです。

		depositRequest.DepositAmount += plannedWithdrawalAmount
		depositRequest.DepositCheckStatus = domain.NO_DEPOSIT
		depositRequest.DepositDate = jsonutils.JSONTime(time.Now())

		if err = u.depositRequestRepository.Update(ctx, depositRequest); err != nil {
			return errors.Wrap(err, "failed to update deposit request")
		}

		userAccount.Balance = lo.ToPtr(lo.FromPtr(userAccount.Balance) - plannedWithdrawalAmount) // 残金から予定出金額を差し引きます。

		if err = u.userAccountRepository.Update(ctx, userAccount); err != nil {
			return errors.Wrap(err, "failed to update user account")
		}

		// 残高不足メール送信用の情報を入力するために、プロジェクト(project)、ファンド(share)、ユーザー仮想口座(userVirtualAccount)情報を取得します。
		project, err := u.projectRepository.GetById(ctx, depositRequest.ProjectId)
		if err != nil {
			return errors.Wrap(err, "failed to get project")
		}

		share, err := u.shareRepository.GetById(ctx, depositRequest.ShareId)
		if err != nil {
			return errors.Wrap(err, "failed to get share")
		}

		userVirtualAccount, err := u.userVirtualAccountRepository.GetByUserIdAndLinkedStatus(ctx, depositRequest.UserId, project.ProjectType, domain.ON_VA_ACCOUNT)
		if err != nil {
			return errors.Wrap(err, "failed to get user virtual account")
		}

		lackAmount := depositRequest.RequestAmount - depositRequest.DepositAmount

		investInsufficientBalanceDetails := map[string]interface{}{
			"Receiver":         user.Name,
			"ServiceUrl":       u.cfg.ServiceConfig.Url.UserServiceDomain,
			"LackAmount":       message.NewPrinter(language.Japanese).Sprintf("%d", lackAmount),
			"ProductName":      project.Name,
			"TransactionDate":  jsonutils.JSONTime.ToString(share.InvestDate),
			"BankName":         depositRequest.DepositBankName,
			"BranchName":       depositRequest.DepositBranchName,
			"AccountNumber":    depositRequest.DepositAccountNumber,
			"Depositor":        userVirtualAccount.VaHolderNameKana,
			"InvestmentAmount": message.NewPrinter(language.Japanese).Sprintf("%d", share.Amount),
			"PaymentEnd":       jsonutils.JSONTime.ToString(depositRequest.ExpireDepositDate),
		}

		// 残高不足メール送信
		if err = u.mailService.SendMail(ctx, cdomain.SendMailRequest{
			TemplateType: domain.INVEST_INSUFFICIENT_BALANCE_TEMPLATE,
			Sender:       u.cfg.ServiceConfig.Mail.Info,
			Receivers:    []string{user.Email},
			MailType:     domain.MAIL_SEND_PRIORITY_NORMAL,
			Details:      investInsufficientBalanceDetails,
		}); err != nil {
			return errors.Wrap(err, "failed to send mail")
		}

		if err = u.depositRequestLackHistoryRepository.Create(ctx, domain.DepositRequestLackHistory{
			UserId:              user.Id,
			DepositRequestId:    depositRequest.Id,
			RequestExecuteDate:  jsonutils.JSONTime(time.Now()),
			DepositExpectedDate: depositRequest.ExpireDepositDate,
			Amount:              lackAmount,
			AccountNumber:       depositRequest.DepositAccountNumber,
			ProjectType:         project.ProjectType,
			ProjectId:           project.Id,
			ShareId:             share.Id,
		}); err != nil {
			return errors.Wrap(err, "failed to create deposit request lack history")
		}
	}

	return nil
}
