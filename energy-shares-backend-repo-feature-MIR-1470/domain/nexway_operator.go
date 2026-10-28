package domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

// NexwayBpoOperator Nexway eKYC確認を行ったオペレーター（確認者）情報
type NexwayBpoOperator struct {
	Id               int64              `json:"id" xorm:"pk autoincr"`
	EkycRequestLogId int64              `json:"ekyc_request_log_id" xorm:"notnull"`
	CompanyId        int64              `json:"company_id" xorm:"notnull"`
	CompanyName      string             `json:"company_name" xorm:"notnull"`
	UserId           int64              `json:"user_id" xorm:"notnull"`
	UserName         string             `json:"user_name" xorm:"notnull"`
	LoginId          string             `json:"login_id" xorm:"notnull"`
	OperatorType     string             `json:"operator_type" xorm:"notnull"` // first, second, sv
	OperationEndAt   jsonutils.JSONTime `json:"operation_end_at" xorm:"notnull"`
	Created          jsonutils.JSONTime `json:"created" xorm:"created"`
	Updated          jsonutils.JSONTime `json:"updated" xorm:"updated"`
}

func (NexwayBpoOperator) TableName() string {
	return "nexway_operators"
}

func (n NexwayBpoOperator) SelectAll() string {
	return n.TableName() + ".*"
}

// CRUD operations
func (n *NexwayBpoOperator) Create(ctx context.Context) (int64, error) {
	return factory.DB(ctx).Insert(n)
}

func (NexwayBpoOperator) GetById(ctx context.Context, id int64) (*NexwayBpoOperator, error) {
	var item NexwayBpoOperator
	has, err := factory.DB(ctx).ID(id).Get(&item)
	if err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}
	return &item, nil
}

func (NexwayBpoOperator) FindByEkycRequestLogId(ctx context.Context, ekycRequestLogId int64) ([]*NexwayBpoOperator, error) {
	var items []*NexwayBpoOperator
	err := factory.DB(ctx).Where("ekyc_request_log_id = ?", ekycRequestLogId).Find(&items)
	if err != nil {
		return nil, err
	}

	return items, nil
}

func (n *NexwayBpoOperator) Update(ctx context.Context) error {
	_, err := factory.DB(ctx).ID(n.Id).Update(n)
	return err
}

func (n *NexwayBpoOperator) Delete(ctx context.Context) error {
	_, err := factory.DB(ctx).ID(n.Id).Delete(n)
	return err
}
