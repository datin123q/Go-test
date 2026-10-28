package clean_arch_domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type NoticeMsgRepository interface {
	GetAll(ctx context.Context, options ...NoticeMsgQueryOption) (int64, []domain.NoticeMsg, error)
	GetById(ctx context.Context, id int64) (domain.NoticeMsg, error)
	Create(ctx context.Context, noticeMsg domain.NoticeMsg) (int64, error)
}

type NoticeMsgQueryOption = func(*NoticeMsgQueryOptionFields)

func NoticeMsgQueryOptionFunc() NoticeMsgQueryOptionFields {
	return NoticeMsgQueryOptionFields{}
}

type NoticeMsgQueryOptionFields struct {
	title     *string
	startDate *string
	endDate   *string
	msgIds    []int64
	QueryOptions
}

func (qo NoticeMsgQueryOptionFields) WithTitle(title *string) func(*NoticeMsgQueryOptionFields) {
	return func(qo *NoticeMsgQueryOptionFields) {
		qo.title = title
	}
}

func (qo NoticeMsgQueryOptionFields) WithStartDate(startDate *string) func(*NoticeMsgQueryOptionFields) {
	return func(qo *NoticeMsgQueryOptionFields) {
		qo.startDate = startDate
	}
}

func (qo NoticeMsgQueryOptionFields) WithEndDate(endDate *string) func(*NoticeMsgQueryOptionFields) {
	return func(qo *NoticeMsgQueryOptionFields) {
		qo.endDate = endDate
	}
}

func (qo NoticeMsgQueryOptionFields) WithMsgIds(msgIds []int64) func(*NoticeMsgQueryOptionFields) {
	return func(qo *NoticeMsgQueryOptionFields) {
		qo.msgIds = msgIds
	}
}

func (qo NoticeMsgQueryOptionFields) WithSortBy(sortBy string) func(*NoticeMsgQueryOptionFields) {
	return func(qo *NoticeMsgQueryOptionFields) {
		qo.sortBy = &sortBy
	}
}

func (qo NoticeMsgQueryOptionFields) WithOrder(order string) func(*NoticeMsgQueryOptionFields) {
	return func(qo *NoticeMsgQueryOptionFields) {
		qo.order = &order
	}
}

func (qo NoticeMsgQueryOptionFields) WithOffset(offset int) func(*NoticeMsgQueryOptionFields) {
	return func(qo *NoticeMsgQueryOptionFields) {
		qo.offset = &offset
	}
}

func (qo NoticeMsgQueryOptionFields) WithLimit(limit int) func(*NoticeMsgQueryOptionFields) {
	return func(qo *NoticeMsgQueryOptionFields) {
		qo.limit = &limit
	}
}

func (qo NoticeMsgQueryOptionFields) Title() *string {
	return qo.title
}

func (qo NoticeMsgQueryOptionFields) StartDate() *string {
	return qo.startDate
}

func (qo NoticeMsgQueryOptionFields) EndDate() *string {
	return qo.endDate
}

func (qo NoticeMsgQueryOptionFields) MsgIds() []int64 {
	return qo.msgIds
}
