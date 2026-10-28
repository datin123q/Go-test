package investment_waiters

import (
	"context"

	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	udomain "github.com/illuminarean/energyshares-user-api/domain"
)

type investmentWaitersUsecase struct {
	txManager                       factory.TxManager
	investmentWaitersRepository     cdomain.InvestmentWaitersRepository
	investmentWaitHistoryRepository cdomain.InvestmentWaitHistoryRepository
}

func New(
	txManager factory.TxManager,
	investmentWaitersRepository cdomain.InvestmentWaitersRepository,
	investmentWaitHistoryRepository cdomain.InvestmentWaitHistoryRepository,
) udomain.InvestmentWaitersUsecase {
	return &investmentWaitersUsecase{
		txManager:                       txManager,
		investmentWaitersRepository:     investmentWaitersRepository,
		investmentWaitHistoryRepository: investmentWaitHistoryRepository,
	}
}

// Cancel はユーザー本人のキャンセル待ち申請を取り消す。
func (i investmentWaitersUsecase) Cancel(ctx context.Context, request udomain.InvestmentWaitersCancelRequest) (udomain.InvestmentWaitersCancelResponse, error) {
	var response udomain.InvestmentWaitersCancelResponse

	if request.JWTUser.Id != request.UserID {
		return response, errors.Wrap(cdomain.ErrForbidden, "cannot cancel another user's investment waiter")
	}

	if err := i.txManager.WithTx(ctx, func(txCtx context.Context) error {
		investmentWaiters, err := i.investmentWaitersRepository.LockById(txCtx, request.InvestmentWaiterID)
		if err != nil {
			return errors.Wrap(err, "failed to lock investment waiter")
		}

		if investmentWaiters.UserId != request.UserID {
			return errors.Wrap(cdomain.ErrForbidden, "investment waiter is owned by another user")
		}

		if investmentWaiters.Status != domain.WAITING {
			return errors.Wrapf(
				cdomain.ErrBadRequest,
				"investment waiter cannot be canceled (status: %s)",
				investmentWaiters.Status,
			)
		}

		// 当選済みかどうかは中間テーブル(投資済みかどうかの一次情報)を行ロックで確認する
		// (MIR-1118)。旧列investment_waiters.share_idへのフォールバックはPR3(MIR-1163)で
		// 削除した。MIR-1068で繰上げが複数回になっても、1件でも当選実績があればキャンセル
		// 不可という判定自体は変えない(一部繰上げ済みの残りだけをキャンセルする機能は無い)。
		advancedQty, err := i.investmentWaitersRepository.LockAndSumAdvancedQty(txCtx, investmentWaiters.Id)
		if err != nil {
			return errors.Wrap(err, "failed to resolve advanced qty")
		}
		if !domain.InvestmentWaiterIsStillPending(investmentWaiters.Status, advancedQty) {
			return errors.Wrapf(
				cdomain.ErrBadRequest,
				"investment waiter cannot be canceled (status: %s, advanced qty: %d)",
				investmentWaiters.Status,
				advancedQty,
			)
		}

		updatedInvestmentWaiters := investmentWaiters
		updatedInvestmentWaiters.Status = domain.CANCELLED

		affected, err := i.investmentWaitersRepository.Update(txCtx, updatedInvestmentWaiters)
		if err != nil {
			return errors.Wrap(err, "failed to cancel investment waiter")
		}
		if affected == 0 {
			return errors.Wrap(cdomain.ErrItemNotAffected, "investment waiter was not canceled")
		}

		if err = i.investmentWaitHistoryRepository.Create(txCtx, domain.InvestmentWaitHistory{
			InvestmentWaitId: investmentWaiters.Id,
			SubstanceType:    string(domain.INVESTMENT_WAIT_HISTORY_STATUS),
			Name:             request.JWTUser.Name,
			Email:            request.JWTUser.Email,
			PrevStatus:       investmentWaiters.Status,
			CurrStatus:       updatedInvestmentWaiters.Status,
			ProcessMethod:    domain.USER_HISTORY_PROCESS_MEMBER,
		}); err != nil {
			return errors.Wrap(err, "failed to create investment waiter status history")
		}

		return nil
	}); err != nil {
		return response, err
	}

	return response, nil
}
