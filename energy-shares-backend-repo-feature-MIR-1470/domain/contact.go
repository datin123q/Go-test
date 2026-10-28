package domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/modelhelper"
)

type Contact struct {
	Id       int64              `json:"id" `
	Name     string             `json:"name"`
	Email    string             `json:"email"`
	Title    string             `json:"title"`
	Contents string             `json:"contents" xorm:"LONGTEXT"`
	Created  jsonutils.JSONTime `json:"created" xorm:"created"`
	Updated  jsonutils.JSONTime `json:"updated" xorm:"updated"`
	Deleted  jsonutils.JSONTime `json:"deleted" xorm:"deleted"`
}

func (c *Contact) Create(ctx context.Context) (int64, error) {

	return factory.DB(ctx).Insert(c)
}

func (Contact) GetById(ctx context.Context, id int64) (*Contact, error) {

	var c Contact
	if has, err := factory.DB(ctx).ID(id).Get(&c); err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}
	return &c, nil
}

func (Contact) GetAll(ctx context.Context, sortby, order string, offset, limit int, name, title, email *string, startDate, endDate string) (int64, []Contact, error) {
	var err error
	q := factory.DB(ctx).Table("contact")

	if err := modelhelper.SetSortOrder(q, sortby, order); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SearchLikeQuery(q, "contact", "title", title); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SearchLikeQuery(q, "contact", "name", name); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SearchEqQuery(q, "contact", "email", email); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SetBetweenDate(q, "created", startDate, endDate); err != nil {
		factory.Logger(ctx).Error(err)
	}

	var items []Contact
	var totalCount int64

	totalCount, err = q.Limit(limit, offset).FindAndCount(&items)

	if err != nil {
		return 0, nil, err
	}

	return totalCount, items, nil
}
