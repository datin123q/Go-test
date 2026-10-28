package domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/pkg/errors"
)

type UserDeposit struct {
	User           User           `xorm:"extends"`
	DepositRequest DepositRequest `xorm:"extends"`
}

func (UserDeposit) GetDepositRequestByProjectId(ctx context.Context, projectId int64) (int64, []UserDeposit, error) {
	q := factory.DB(ctx).Table("deposit_request").
		Join("INNER", "user", "deposit_request.user_id = user.id").Select("user.*,deposit_request.*")
	q = q.Where("deposit_request.project_id = ?", projectId)
	items := make([]UserDeposit, 0, 0)
	var totalCount int64
	totalCount, err := q.FindAndCount(&items)
	if err != nil {
		return 0, nil, err
	}

	return totalCount, items, nil
}

func (UserDeposit) FailFundingToRefund(ctx context.Context, projectId int64) error {
	var transfers []Transfer
	_, items, err := UserDeposit{}.GetDepositRequestByProjectId(ctx, projectId)
	if err != nil {
		return errors.Wrap(err, "Fail Get Deposit Request By Project ID ")
	}

	transfers = make([]Transfer, 0, 0)
	for i := range items {
		if items[i].DepositRequest.DepositCheckStatus == OK_DEPOSIT && items[i].DepositRequest.DepositAmount > 0 {

			t := Transfer{
				TransactionId:         items[i].DepositRequest.ShareId,
				UserId:                items[i].DepositRequest.UserId,
				ProjectId:             items[i].DepositRequest.ProjectId,
				TransferAmount:        items[i].DepositRequest.DepositAmount,
				AccountNumber:         items[i].User.BankAccount,
				BankCode:              *items[i].User.BankCode,
				BankName:              items[i].User.BankName,
				BranchCode:            *items[i].User.BranchCode,
				BranchName:            items[i].User.BranchName,
				Depositor:             items[i].User.Depositor,
				AccountTypeCode:       DepositAccountTypeCheck(items[i].User.DepositType),
				TransactionDetailType: WITHDRAWAL_REFUND,
			}
			transfers = append(transfers, t)
		}
	}
	err = TransferRequest{}.NewTransferRequest(ctx, transfers)
	if err != nil {
		return errors.Wrap(err, "fail Funding To Refund")
	}

	return nil
}
