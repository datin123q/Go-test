package clean_arch_domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type NoticeMsgReadRepository interface {
	GetAll(ctx context.Context, options ...NoticeMsgReadQueryOption) (int64, []domain.NoticeMsgRead, error)
	Create(ctx context.Context, noticeMsgRead domain.NoticeMsgRead) (int64, error)
	Update(ctx context.Context, noticeMsgRead domain.NoticeMsgRead) error
}

type NoticeMsgReadQueryOption = func(*NoticeMsgReadQueryOptionFields)

func NoticeMsgReadQueryOptionFunc() NoticeMsgReadQueryOptionFields {
	return NoticeMsgReadQueryOptionFields{}
}

type NoticeMsgReadQueryOptionFields struct {
	msgId  *int64
	userId *int64
	isRead *bool
	QueryOptions
}

func (qo NoticeMsgReadQueryOptionFields) WithMsgId(msgId int64) func(*NoticeMsgReadQueryOptionFields) {
	return func(qo *NoticeMsgReadQueryOptionFields) {
		qo.msgId = &msgId
	}
}

func (qo NoticeMsgReadQueryOptionFields) WithUserId(userId int64) func(*NoticeMsgReadQueryOptionFields) {
	return func(qo *NoticeMsgReadQueryOptionFields) {
		qo.userId = &userId
	}
}

func (qo NoticeMsgReadQueryOptionFields) WithIsRead(isRead bool) func(*NoticeMsgReadQueryOptionFields) {
	return func(qo *NoticeMsgReadQueryOptionFields) {
		qo.isRead = &isRead
	}
}

func (qo NoticeMsgReadQueryOptionFields) WithSortBy(sortBy string) func(*NoticeMsgReadQueryOptionFields) {
	return func(qo *NoticeMsgReadQueryOptionFields) {
		qo.sortBy = &sortBy
	}
}

func (qo NoticeMsgReadQueryOptionFields) WithOrder(order string) func(*NoticeMsgReadQueryOptionFields) {
	return func(qo *NoticeMsgReadQueryOptionFields) {
		qo.order = &order
	}
}

func (qo NoticeMsgReadQueryOptionFields) WithOffset(offset int) func(*NoticeMsgReadQueryOptionFields) {
	return func(qo *NoticeMsgReadQueryOptionFields) {
		qo.offset = &offset
	}
}

func (qo NoticeMsgReadQueryOptionFields) WithLimit(limit int) func(*NoticeMsgReadQueryOptionFields) {
	return func(qo *NoticeMsgReadQueryOptionFields) {
		qo.limit = &limit
	}
}

func (qo NoticeMsgReadQueryOptionFields) MsgId() *int64 {
	return qo.msgId
}

func (qo NoticeMsgReadQueryOptionFields) UserId() *int64 {
	return qo.userId
}

func (qo NoticeMsgReadQueryOptionFields) IsRead() *bool {
	return qo.isRead
}
