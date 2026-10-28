package domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

type JobExecuteTime struct {
	Id                  int64              `json:"id" xorm:"pk autoincr"`
	DepositCheckTime    jsonutils.JSONTime `json:"deposit_check_time" xorm:"deposit_check_time"`       // 最後の入金内訳照会時間
	VaAccountIssueTime  jsonutils.JSONTime `json:"va_account_issue_time" xorm:"va_account_issue_time"` // 最後の口座発行時間
	TransferRequestTime jsonutils.JSONTime `json:"transfer_request_time" xorm:"transfer_request_time"`
	TransferResultTime  jsonutils.JSONTime `json:"transfer_result_time" xorm:"transfer_result_time"`
	DepositProcessTime  jsonutils.JSONTime `json:"deposit_process_time" xorm:"deposit_process_time"` // 最後の預り金自動入金処理バッチ実行時間
	TransferStatusTime  jsonutils.JSONTime `json:"transfer_status_time" xorm:"transfer_status_time"` // 最後の送金状態照会バッチ実行時間
}

func (JobExecuteTime) GetJobLastExecuteTime(ctx context.Context) (*JobExecuteTime, error) {

	var res JobExecuteTime
	db := factory.DB(ctx)
	if has, err := db.ID(1).Get(&res); err != nil {
		return nil, err
	} else if !has {
		res = JobExecuteTime{}
		if _, err := db.InsertOne(&res); err != nil {
			if has, getErr := db.ID(1).Get(&res); getErr != nil {
				return nil, getErr
			} else if !has {
				return nil, err
			}
		}
	}

	return &res, nil
}

func (j *JobExecuteTime) Update(ctx context.Context) error {
	if _, err := factory.DB(ctx).ID(1).Update(j); err != nil {
		return err
	}
	return nil
}

func (j *JobExecuteTime) UpdateVaAccountIssueTime(ctx context.Context, t jsonutils.JSONTime) error {
	j.VaAccountIssueTime = t
	if err := j.Update(ctx); err != nil {
		return err
	}
	return nil
}

func (j *JobExecuteTime) UpdateTransferRequestTime(ctx context.Context, t jsonutils.JSONTime) error {
	j.TransferRequestTime = t
	if err := j.Update(ctx); err != nil {
		return err
	}
	return nil
}

func (j *JobExecuteTime) UpdateTransferStatusTime(ctx context.Context, t jsonutils.JSONTime) error {
	j.TransferStatusTime = t
	if err := j.Update(ctx); err != nil {
		return err
	}
	return nil
}

func (j *JobExecuteTime) UpdateDepositCheckTime(ctx context.Context, t jsonutils.JSONTime) error {
	j.DepositCheckTime = t
	if err := j.Update(ctx); err != nil {
		return err
	}
	return nil
}

func (j *JobExecuteTime) UpdateTransferResultTime(ctx context.Context, t jsonutils.JSONTime) error {
	j.TransferResultTime = t
	if err := j.Update(ctx); err != nil {
		return err
	}
	return nil
}

func (j *JobExecuteTime) UpdateDepositProcessTime(ctx context.Context, t jsonutils.JSONTime) error {
	j.DepositProcessTime = t
	if err := j.Update(ctx); err != nil {
		return err
	}
	return nil
}
