package domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/modelhelper"
)

type ConsultNote struct {
	Id         int64              `json:"id"`
	UserId     int64              `json:"user_id" xorm:"index"`        // 고객 ID
	OwnerId    int64              `json:"owner_id" xorm:"not null"`    // 작성자 ID
	Owner      string             `json:"owner" xorm:"not null"`       // 작성자 이름 (수정자 ID 기준으로 조회)
	ModifierId int64              `json:"modifier_id" xorm:"not null"` // 수정자 ID
	Title      string             `json:"title" xorm:"not null"`       // 제목
	Content    string             `json:"content" xorm:"not null"`     // 내용
	Created    jsonutils.JSONTime `json:"created" xorm:"created"`      // 생성일
	Updated    jsonutils.JSONTime `json:"updated" xorm:"updated"`      // 수정일
	Deleted    jsonutils.JSONTime `json:"deleted" xorm:"deleted"`      // 삭제일
}

func (ConsultNote) GetAll(ctx context.Context, sortBy, order string, offset, limit int) (int64, []ConsultNote, error) {
	q := factory.DB(ctx).Table("consult_note").
		Select("consult_note.*")

	if err := modelhelper.SetSortOrder(q, sortBy, order); err != nil {
		factory.Logger(ctx).Error(err)
	}

	var items []ConsultNote
	var totalCount int64
	totalCount, err := q.Limit(limit, offset).FindAndCount(&items)
	if err != nil {
		return 0, nil, err
	}

	return totalCount, items, nil
}

func (ConsultNote) GetAllByUserId(ctx context.Context, userId int64, sortBy, order string, offset, limit int) (int64, []ConsultNote, error) {
	q := factory.DB(ctx).Table("consult_note").
		Select("consult_note.*").
		Where("user_id=?", userId)

	if err := modelhelper.SetSortOrder(q, sortBy, order); err != nil {
		factory.Logger(ctx).Error(err)
	}

	var items []ConsultNote
	var totalCount int64

	totalCount, err := q.Limit(limit, offset).FindAndCount(&items)
	if err != nil {
		return 0, nil, err
	}

	return totalCount, items, nil
}

func (cn *ConsultNote) Create(ctx context.Context) (int64, error) {
	return factory.DB(ctx).Insert(cn)
}

func (cn *ConsultNote) Update(ctx context.Context) (int64, error) {
	return factory.DB(ctx).ID(cn.Id).Update(cn)
}

func (ConsultNote) GetById(ctx context.Context, id int64) (*ConsultNote, error) {
	cn := &ConsultNote{}
	has, err := factory.DB(ctx).ID(id).Get(cn)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return cn, nil
}

func (cn *ConsultNote) Delete(ctx context.Context) error {
	_, err := factory.DB(ctx).ID(cn.Id).Delete(cn)
	return err
}
