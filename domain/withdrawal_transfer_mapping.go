package domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/modelhelper"
)

type WithdrawalTransferMapping struct {
	Id            int64  `json:"id"`
	TransactionId string `json:"transaction_id" xorm:"not null"`
	WithdrawalId  int64  `json:"withdrawal_id" xorm:"not null"`
}

func (wtm WithdrawalTransferMapping) TableName() string {
	return "withdrawal_transfer_mapping"
}

func (wtm WithdrawalTransferMapping) SelectAll() string {
	return wtm.TableName() + ".*"
}

func (wtm *WithdrawalTransferMapping) Create(ctx context.Context) (int64, error) {
	return factory.DB(ctx).Insert(wtm)
}

func (wtm *WithdrawalTransferMapping) CreateMulti(ctx context.Context, list []WithdrawalTransferMapping) error {
	_, err := factory.DB(ctx).Insert(&list)
	return err
}

func (wtm *WithdrawalTransferMapping) GetAll(ctx context.Context, sortBy, order string, offset, limit int) (int64, []WithdrawalTransferMapping, error) {
	var items []WithdrawalTransferMapping
	q := factory.DB(ctx)
	if err := modelhelper.SetSortOrder(q, sortBy, order); err != nil {
		factory.Logger(ctx).Error(err)
	}

	totalCount, err := q.Limit(limit, offset).FindAndCount(&items)
	if err != nil {
		return 0, nil, err
	}
	return totalCount, items, nil
}
