package withdrawal

import (
	"context"
	"strconv"
	"time"

	adomain "github.com/illuminarean/energy-share-core-api/domain"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

type withdrawalUsecase struct {
	cfg                       *config.Config
	txManager                 factory.TxManager
	mailService               cdomain.MailService
	userRepository            cdomain.UserRepository
	userAccountRepository     cdomain.UserAccountRepository
	operatorHistoryRepository cdomain.OperatorHistoryRepository
	withdrawalRepository      cdomain.WithdrawalRepository
	transferRepository        cdomain.TransferRepository
	withdrawalService         cdomain.WithdrawalService
	withdrawBalance           func(context.Context, *domain.UserAccount, int64, domain.USER_TRANSACTION_DETAIL_TYPE) (int64, error)
	requestFee                func(context.Context, domain.User, int64) (int64, error)
	now                       func() time.Time
}

func New(
	cfg *config.Config,
	txManager factory.TxManager,
	mailService cdomain.MailService,
	userRepository cdomain.UserRepository,
	userAccountRepository cdomain.UserAccountRepository,
	operatorHistoryRepository cdomain.OperatorHistoryRepository,
	withdrawalRepository cdomain.WithdrawalRepository,
	transferRepository cdomain.TransferRepository,
	withdrawalService cdomain.WithdrawalService,
) adomain.WithdrawalUsecase {
	return &withdrawalUsecase{
		cfg:                       cfg,
		txManager:                 txManager,
		mailService:               mailService,
		userRepository:            userRepository,
		userAccountRepository:     userAccountRepository,
		operatorHistoryRepository: operatorHistoryRepository,
		withdrawalRepository:      withdrawalRepository,
		transferRepository:        transferRepository,
		withdrawalService:         withdrawalService,
		withdrawBalance: func(ctx context.Context, account *domain.UserAccount, amount int64, transactionType domain.USER_TRANSACTION_DETAIL_TYPE) (int64, error) {
			return account.WithdrawBalance(ctx, amount, transactionType)
		},
		requestFee: domain.Transfer{}.RequestFee,
		now:        time.Now,
	}
}

func (u withdrawalUsecase) Create(ctx context.Context, request adomain.WithdrawalCreateRequest) (adomain.WithdrawalCreateResponse, error) {
	if request.Amount <= 0 {
		return adomain.WithdrawalCreateResponse{}, errors.Wrap(cdomain.ErrBadRequest, "withdrawal amount must be greater than zero")
	}

	user, err := u.userRepository.GetById(ctx, request.UserId)
	if err != nil {
		return adomain.WithdrawalCreateResponse{}, errors.Wrap(err, "failed to get user")
	}
	if user.AuditStatus != domain.COMPLETED_AUDIT || user.UserType == domain.ADMIN {
		return adomain.WithdrawalCreateResponse{}, errors.Wrap(cdomain.ErrBadRequest, "user is invalid for withdrawal request")
	}

	account, err := u.userAccountRepository.GetByUserId(ctx, request.UserId)
	if err != nil {
		return adomain.WithdrawalCreateResponse{}, errors.Wrap(err, "failed to get user account")
	}
	if lo.FromPtr(account.Balance) < request.Amount {
		return adomain.WithdrawalCreateResponse{}, errors.Wrap(cdomain.ErrBadRequest, "insufficient balance")
	}

	requestFee := u.requestFee
	if requestFee == nil {
		requestFee = domain.Transfer{}.RequestFee
	}

	fee, err := requestFee(ctx, user, request.Amount)
	if err != nil {
		return adomain.WithdrawalCreateResponse{}, errors.Wrap(err, "failed to request withdrawal fee")
	}

	nowFn := u.now
	if nowFn == nil {
		nowFn = time.Now
	}
	withdrawBalance := u.withdrawBalance
	if withdrawBalance == nil {
		withdrawBalance = func(ctx context.Context, account *domain.UserAccount, amount int64, transactionType domain.USER_TRANSACTION_DETAIL_TYPE) (int64, error) {
			return account.WithdrawBalance(ctx, amount, transactionType)
		}
	}

	now := jsonutils.JSONTime(nowFn())
	withdrawal := domain.Withdrawal{
		InvestorId:   strconv.FormatInt(user.Id, 10),
		InvestorName: user.Name,
		Amount:       request.Amount,
		Fee:          fee,
		RequestDate:  now,
		Status:       domain.WITHDRAWAL_STATUS_INCOMPLETE,
		IsManual:     true,
		Created:      now,
		Updated:      now,
	}

	if err := u.txManager.WithTx(ctx, func(txCtx context.Context) error {
		if _, err := withdrawBalance(txCtx, &account, withdrawal.Amount, domain.WITHDRAWAL_INVESTOR); err != nil {
			return errors.Wrap(err, "failed to withdraw user balance")
		}

		withdrawalID, err := u.withdrawalRepository.Create(txCtx, withdrawal)
		if err != nil {
			return errors.Wrap(err, "failed to create withdrawal")
		}
		withdrawal.Id = withdrawalID

		if err := u.operatorHistoryRepository.Create(txCtx, domain.OperatorHistory{
			UserId: request.JWTUser.Id,
			Name:   request.JWTUser.Name,
			Email:  request.JWTUser.Email,
			Menu:   "menu.manage_accounting.withdrawal_request",
			Ip:     request.ClientIp,
			State:  domain.REGISTER,
		}); err != nil {
			return errors.Wrap(err, "failed to create operator history")
		}

		return nil
	}); err != nil {
		return adomain.WithdrawalCreateResponse{}, err
	}

	if err := u.mailService.SendMail(ctx, cdomain.SendMailRequest{
		TemplateType: domain.WITHDRAWAL_REQUEST_TEMPLATE,
		MailType:     domain.MAIL_SEND_PRIORITY_NORMAL,
		Sender:       u.cfg.ServiceConfig.Mail.Info,
		Receivers:    []string{user.Email},
		Details: map[string]interface{}{
			"Receiver":    user.Name,
			"ServiceUrl":  u.cfg.ServiceConfig.Url.UserServiceDomain,
			"RequestDate": withdrawal.RequestDate.Time().Format(time.DateOnly),
			"Amount":      strconv.FormatInt(withdrawal.Amount, 10),
		},
	}); err != nil {
		factory.Logger(ctx).WithError(err).Error("failed to send withdrawal request mail")
	}

	return adomain.WithdrawalCreateResponse{
		Id:           withdrawal.Id,
		InvestorId:   withdrawal.InvestorId,
		InvestorName: withdrawal.InvestorName,
		Amount:       withdrawal.Amount,
		Fee:          withdrawal.Fee,
		RequestDate:  withdrawal.RequestDate,
		Status:       withdrawal.Status,
		Created:      withdrawal.Created,
		Updated:      withdrawal.Updated,
	}, nil
}

func (u withdrawalUsecase) GetAll(ctx context.Context, request adomain.WithdrawalGetAllRequest) (adomain.WithdrawalGetAllResponse, error) {
	opts := cdomain.WithdrawalQueryOptionFunc()
	totalCount, withdrawals, err := u.withdrawalRepository.GetAll(ctx,
		opts.WithSortBy(request.Sortby),
		opts.WithOrder(request.Order),
		opts.WithOffset(request.SkipCount),
		opts.WithLimit(request.MaxResultCount),
		opts.WithInvestorId(request.InvestorId),
		opts.WithInvestorName(request.InvestorName),
		opts.WithAmountMin(request.AmountMin),
		opts.WithAmountMax(request.AmountMax),
		opts.WithRequestStartDate(request.RequestStartDate),
		opts.WithRequestEndDate(request.RequestEndDate),
		opts.WithCompleteStartDate(request.CompleteStartDate),
		opts.WithCompleteEndDate(request.CompleteEndDate),
		opts.WithStatus(request.Status),
	)
	if err != nil {
		return adomain.WithdrawalGetAllResponse{}, errors.Wrap(err, "failed to get all withdrawals")
	}

	items := lo.Map(withdrawals, func(wd domain.WithdrawalDetail, _ int) adomain.WithdrawalItem {
		return adomain.WithdrawalItem{
			Id:           wd.Withdrawal.Id,
			InvestorId:   wd.Withdrawal.InvestorId,
			InvestorName: wd.Withdrawal.InvestorName,
			Amount:       wd.Withdrawal.Amount,
			Fee:          wd.Withdrawal.Fee,
			RequestDate:  wd.Withdrawal.RequestDate,
			CompleteDate: wd.Withdrawal.CompleteDate,
			Status:       wd.Withdrawal.Status,
			IsManual:     wd.Withdrawal.IsManual,
			Created:      wd.Withdrawal.Created,
		}
	})

	// withdrawal_transfer_mapping 経由で紐づく transfer がある場合、最新の transfer 情報で上書き
	for idx, wd := range withdrawals {
		if wd.Withdrawal.Status != domain.WITHDRAWAL_STATUS_COMPLETE {
			continue
		}
		_, transfers, err := u.transferRepository.GetAll(ctx,
			cdomain.TransferQueryOptionFunc().WithWithdrawalId(lo.ToPtr(wd.Withdrawal.Id)),
			cdomain.TransferQueryOptionFunc().WithOrder("desc"),
			cdomain.TransferQueryOptionFunc().WithSortBy("id"),
			cdomain.TransferQueryOptionFunc().WithLimit(1),
		)
		if err != nil {
			return adomain.WithdrawalGetAllResponse{}, errors.Wrap(err, "failed to get transfers for withdrawal")
		}
		if len(transfers) == 0 {
			continue
		}
		latest := transfers[0]
		items[idx].Amount = latest.TransferAmount
		items[idx].Fee = latest.TransferFee
	}

	totalIncompleteCount, _, err := u.withdrawalRepository.GetAll(ctx,
		opts.WithOffset(0),
		opts.WithLimit(99999),
		opts.WithStatus(lo.ToPtr(domain.WITHDRAWAL_STATUS_INCOMPLETE)),
	)
	if err != nil {
		return adomain.WithdrawalGetAllResponse{}, err
	}

	return adomain.WithdrawalGetAllResponse{
		Items:                items,
		TotalCount:           totalCount,
		TotalIncompleteCount: totalIncompleteCount,
	}, nil
}

func (u withdrawalUsecase) GetOne(ctx context.Context, request adomain.WithdrawalGetOneRequest) (adomain.WithdrawalGetOneResponse, error) {
	// 출금의뢰 조회
	withdrawal, err := u.withdrawalRepository.GetById(ctx, request.Id)
	if err != nil {
		return adomain.WithdrawalGetOneResponse{}, errors.Wrap(err, "failed to get withdrawal")
	}

	// 송금 처리 내역 조회
	opts := cdomain.TransferQueryOptionFunc()
	_, transfers, err := u.transferRepository.GetAll(ctx,
		opts.WithWithdrawalId(lo.ToPtr(request.Id)),
		opts.WithOrder("desc"),
		opts.WithSortBy("id"),
		opts.WithLimit(1),
	)
	if err != nil {
		return adomain.WithdrawalGetOneResponse{}, errors.Wrap(err, "failed to get transfers")
	}

	transferLogs := lo.Map(transfers, func(t domain.Transfer, _ int) adomain.WithdrawalTransferItem {
		return adomain.WithdrawalTransferItem{
			Id:             t.Id,
			RequestId:      t.RequestId,
			TransactionId:  t.TransactionId,
			TransferDate:   t.TransferDate,
			TransferAmount: t.TransferAmount,
			TransferFee:    t.TransferFee,
			TransferStatus: t.TransferStatus,
			ErrorCode:      t.ErrorCode,
			ErrorMsg:       t.ErrorMsg,
		}
	})

	if len(transfers) > 0 && withdrawal.Withdrawal.Status == domain.WITHDRAWAL_STATUS_COMPLETE {
		latest := transfers[0]
		withdrawal.Withdrawal.Amount = latest.TransferAmount
		withdrawal.Withdrawal.Fee = latest.TransferFee
		if latest.BankName != "" {
			withdrawal.User.BankName = latest.BankName
		}
		if latest.BankCode != "" {
			withdrawal.User.BankCode = &latest.BankCode
		}
		if latest.BranchName != "" {
			withdrawal.User.BranchName = latest.BranchName
		}
		if latest.BranchCode != "" {
			withdrawal.User.BranchCode = &latest.BranchCode
		}
		if latest.AccountNumber != "" {
			withdrawal.User.BankAccount = latest.AccountNumber
		}
		if latest.Depositor != "" {
			withdrawal.User.Depositor = latest.Depositor
		}
		switch latest.AccountTypeCode {
		case "1":
			withdrawal.User.DepositType = domain.ORDINARY_ACCOUNT
		case "2":
			withdrawal.User.DepositType = domain.CHECKING_ACCOUNT
		}
	}

	// 출금의뢰 상세 정보 반환
	return adomain.WithdrawalGetOneResponse{
		WithdrawalDetailItem: adomain.WithdrawalDetailItem{
			WithdrawalItem: adomain.WithdrawalItem{
				Id:           withdrawal.Withdrawal.Id,
				InvestorId:   withdrawal.Withdrawal.InvestorId,
				InvestorName: withdrawal.Withdrawal.InvestorName,
				Amount:       withdrawal.Withdrawal.Amount,
				Fee:          withdrawal.Withdrawal.Fee,
				RequestDate:  withdrawal.Withdrawal.RequestDate,
				CompleteDate: withdrawal.Withdrawal.CompleteDate,
				CloseDate:    withdrawal.Withdrawal.CloseDate,
				Status:       withdrawal.Withdrawal.Status,
				IsManual:     withdrawal.Withdrawal.IsManual,
				Created:      withdrawal.Withdrawal.Created,
			},
			BankName:    withdrawal.User.BankName,
			BranchName:  withdrawal.User.BranchName,
			BranchCode:  withdrawal.User.BranchCode,
			DepositType: withdrawal.User.DepositType,
			BankAccount: withdrawal.User.BankAccount,
			UserType:    withdrawal.User.UserType,
		},
		TransferLogs: transferLogs,
	}, nil
}

func (u withdrawalUsecase) UpdateStatus(ctx context.Context, request adomain.WithdrawalUpdateStatusRequest) (adomain.WithdrawalUpdateStatusResponse, error) {
	// TODO implement me
	panic("implement me")
}

func (u withdrawalUsecase) RequestTransfer(ctx context.Context, request adomain.WithdrawalRequestTransferRequest) (adomain.WithdrawalRequestTransferResponse, error) {
	// TODO implement me
	panic("implement me")
}
