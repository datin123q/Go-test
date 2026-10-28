package service

import (
	"context"

	"github.com/go-faster/errors"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/sirupsen/logrus"
)

type InvestService struct{}

// Invest 投資当選処理
func (s InvestService) Invest(
	ctx context.Context, p *domain.Project, u *domain.User, qty int64, amount int64, investMailType domain.INVEST_MAIL_TYPE,
) error {
	share, err := p.InvestProject(u.Id, qty, p.SharePrice, amount, false, false, false)
	if err != nil {
		return factory.ErrorFailInvest.New(errors.Wrap(err, "fail invest project"))
	}

	// share作成
	if _, err := share.Create(ctx); err != nil {
		return factory.ErrorDB.New(errors.Wrap(err, "fail create share"))
	}

	// 投資待機確認
	if err := p.CheckInvestmentWait(ctx, amount, qty); err != nil {
		return errors.Wrap(err, "fail check investment wait")
	}

	// deposit作成
	deposit, err := domain.DepositRequest{}.NewDepositRequest(ctx, share, false)
	if err != nil {
		return errors.Wrap(err, "fail create deposit request")
	}

	// 投資家数確認
	if p.ProjectType == domain.EQUITY {
		account, err := domain.UserAccount{}.GetByUserId(ctx, u.Id)
		if err != nil {
			return factory.ErrorDB.New(errors.Wrap(err, "fail get user account"))
		}

		if err = account.DepositCheckProcess(ctx, deposit, investMailType); err != nil {
			return factory.ErrorDB.New(errors.Wrap(err, "fail deposit check process"))
		}
	}

	factory.Logger(ctx).WithFields(logrus.Fields{
		"share_id":               share.Id,
		"user_id":                share.OwnerId,
		"project_id":             share.ProjectId,
		"project_type":           share.ProjectType,
		"invest_status":          share.InvestStatus,
		"invest_date":            share.InvestDate,
		"invest_amount":          share.Amount,
		"invest_qty":             share.Qty,
		"deposit_request_id":     deposit.Id,
		"deposit_expire_date":    deposit.ExpireDepositDate,
		"deposit_account_number": deposit.DepositAccountNumber,
		"deposit_check_status":   deposit.DepositCheckStatus,
		"request_amount":         deposit.RequestAmount,
		"deposit_amount":         deposit.DepositAmount,
		"refund_amount":          deposit.RefundAmount,
		"deposit_date":           deposit.DepositDate,
	}).Info("Success Invest")

	return nil
}
