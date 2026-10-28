package deposit_request

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type depositRequestService struct {
	depositRequestRepository  cdomain.DepositRequestRepository
	userAccountRepository     cdomain.UserAccountRepository
	userTransactionRepository cdomain.UserTransactionRepository
}

func New(
	depositRequestRepository cdomain.DepositRequestRepository,
	userAccountRepository cdomain.UserAccountRepository,
	userTransactionRepository cdomain.UserTransactionRepository,
) cdomain.DepositRequestService {
	return &depositRequestService{
		depositRequestRepository:  depositRequestRepository,
		userAccountRepository:     userAccountRepository,
		userTransactionRepository: userTransactionRepository,
	}
}

func (d depositRequestService) OkDeposit(ctx context.Context, depositRequest domain.DepositRequest) error {
	if err := d.depositRequestRepository.Update(ctx, depositRequest.
		WithDepositCheckStatus(domain.OK_DEPOSIT).
		WithDepositAmount(depositRequest.RequestAmount).
		WithDepositDate(jsonutils.JSONTime(time.Now()))); err != nil {
		return errors.Wrap(err, "err update deposit request")
	}

	return nil
}

func (d depositRequestService) CancelDeposit(ctx context.Context, depositRequest domain.DepositRequest) error {
	if depositRequest.DepositCheckStatus == domain.CANCEL_DEPOSIT {
		return errors.New("already been canceled")
	}

	// 今回返金する金額 : 入金金額 - 既に返金済みの金額
	amount := depositRequest.DepositAmount - depositRequest.RefundAmount

	if amount > 0 {
		account, err := domain.UserAccount{}.GetByUserId(ctx, depositRequest.UserId)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to get user account by userId %d", depositRequest.UserId))
		}

		_, err = account.DepositBalance(ctx, amount)
		if err != nil {
			return errors.Wrap(err, "fail Deposit To User Account")
		}

		if domain.IsFundFailureProcessing(ctx) {
			depositRequest.RefundAmount = depositRequest.RefundAmount + amount
			if err := d.depositRequestRepository.Update(ctx, depositRequest.WithDepositCheckStatus(domain.CANCEL_DEPOSIT)); err != nil {
				return errors.Wrap(err, "err update deposit request")
			}
			return nil
		}

		// 投資申請中のリクエストを取得
		drs, err := d.depositRequestRepository.GetNoDepositRequestByUserId(ctx, depositRequest.UserId)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to get deposit request by userId %d", depositRequest.UserId))
		}

		// 投資申請中のリクエストに返金した金額を充当する
		for _, dr := range drs {
			if depositRequest.Id == dr.Id {
				// 自身のリクエストはスキップ
				continue
			}

			if err := account.DepositCheckProcess(ctx, &dr, domain.NOT_IMMEDIATE); err != nil {
				return errors.Wrap(err, "fail Deposit Check process share_id:"+dr.ShareId)
			}
		}

		// 入金要求から払い戻し要求があった場合、払い戻し要求金額は既存 (払い戻し要求額 + 実際の入金額)
		depositRequest.RefundAmount = depositRequest.RefundAmount + amount
	}

	if err := d.depositRequestRepository.Update(ctx, depositRequest.WithDepositCheckStatus(domain.CANCEL_DEPOSIT)); err != nil {
		return errors.Wrap(err, "err update deposit request")
	}

	return nil
}
