package domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

type MarketingCode struct {
	Id              int64              `json:"id" xorm:"pk autoincr"`
	Code            string             `json:"code" xorm:"varchar(255) notnull"`
	Campaign        string             `json:"campaign" xorm:"varchar(255) notnull"`
	SignUpRequestId string             `json:"sign_up_request_id" xorm:"notnull"`
	Email           string             `json:"email" xorm:"varchar(255) notnull"`
	UserId          int64              `json:"user_id" xorm:"notnull"`
	Created         jsonutils.JSONTime `json:"created" xorm:"created"`
	Updated         jsonutils.JSONTime `json:"updated" xorm:"updated"`
}

func NewMarketingCode(code, campaign, reqId, email string) *MarketingCode {
	return &MarketingCode{
		Code:            code,
		Campaign:        campaign,
		SignUpRequestId: reqId,
		Email:           email,
	}
}

// タグ追加
func (m *MarketingCode) Create(ctx context.Context) error {
	_, err := factory.DB(ctx).Insert(m)
	if err != nil {
		return err
	}
	return nil
}

// タグ更新
func (m *MarketingCode) Update(ctx context.Context) error {
	_, err := factory.DB(ctx).ID(m.Id).Update(m)
	if err != nil {
		return err
	}
	return nil
}

// タグ詳細照会
func (MarketingCode) GetById(ctx context.Context, id int64) (*MarketingCode, error) {
	var item MarketingCode
	has, err := factory.DB(ctx).ID(id).Get(&item)
	if err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}
	return &item, nil
}

// 登録申請IDで照会
func (MarketingCode) GetBySignUpRequestId(ctx context.Context, id string) (*MarketingCode, error) {
	var item MarketingCode
	has, err := factory.DB(ctx).Where("sign_up_request_id = ?", id).Limit(1, 0).Get(&item)
	if err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}
	return &item, nil
}

// ユーザーIDで照会
func (MarketingCode) GetByUserId(ctx context.Context, id int64) (*MarketingCode, error) {
	var item MarketingCode
	has, err := factory.DB(ctx).Where("user_id = ?", id).Limit(1, 0).Get(&item)
	if err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}
	return &item, nil
}

// タグリスト照会
func (MarketingCode) GetAll(ctx context.Context) ([]*MarketingCode, error) {
	var items []*MarketingCode
	err := factory.DB(ctx).Find(&items)
	if err != nil {
		return nil, err
	}
	return items, nil
}
