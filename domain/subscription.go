package domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

type SubscriptionMail struct {
	Id      int64              `json:"id"`
	Email   string             `json:"email"`
	Created jsonutils.JSONTime `json:"created" xorm:"created"`
}

func (SubscriptionMail) GetSubscriptionEmail(ctx context.Context) ([]string, error) {
	q := factory.DB(ctx).Table("subscription_mail").Select("subscription_mail.email")

	var items []string
	err := q.Iterate(&SubscriptionMail{}, func(idx int, bean interface{}) error {
		items = append(items, bean.(*SubscriptionMail).Email)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return items, nil
}
