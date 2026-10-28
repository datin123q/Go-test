package clean_arch_domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type UserNoticeMappingRepository interface {
	GetAll(ctx context.Context, options ...UserNoticeMappingQueryOption) (int64, []domain.UserNoticeMapping, error)
	Create(ctx context.Context, mapping domain.UserNoticeMapping) (int64, error)
	CountByMsgIds(ctx context.Context, msgIds []int64) (int64, []domain.NoticeMsgUserCount, error)
}

type UserNoticeMappingQueryOption = func(*UserNoticeMappingQueryOptionFields)

func UserNoticeMappingQueryOptionFunc() UserNoticeMappingQueryOptionFields {
	return UserNoticeMappingQueryOptionFields{}
}

type UserNoticeMappingQueryOptionFields struct {
	msgId  *int64
	userId *int64
	QueryOptions
}

func (qo UserNoticeMappingQueryOptionFields) WithMsgId(msgId int64) func(*UserNoticeMappingQueryOptionFields) {
	return func(qo *UserNoticeMappingQueryOptionFields) {
		qo.msgId = &msgId
	}
}

func (qo UserNoticeMappingQueryOptionFields) WithUserId(userId int64) func(*UserNoticeMappingQueryOptionFields) {
	return func(qo *UserNoticeMappingQueryOptionFields) {
		qo.userId = &userId
	}
}

func (qo UserNoticeMappingQueryOptionFields) WithSortBy(sortBy string) func(*UserNoticeMappingQueryOptionFields) {
	return func(qo *UserNoticeMappingQueryOptionFields) {
		qo.sortBy = &sortBy
	}
}

func (qo UserNoticeMappingQueryOptionFields) WithOrder(order string) func(*UserNoticeMappingQueryOptionFields) {
	return func(qo *UserNoticeMappingQueryOptionFields) {
		qo.order = &order
	}
}

func (qo UserNoticeMappingQueryOptionFields) WithOffset(offset int) func(*UserNoticeMappingQueryOptionFields) {
	return func(qo *UserNoticeMappingQueryOptionFields) {
		qo.offset = &offset
	}
}

func (qo UserNoticeMappingQueryOptionFields) WithLimit(limit int) func(*UserNoticeMappingQueryOptionFields) {
	return func(qo *UserNoticeMappingQueryOptionFields) {
		qo.limit = &limit
	}
}

func (qo UserNoticeMappingQueryOptionFields) MsgId() *int64 {
	return qo.msgId
}

func (qo UserNoticeMappingQueryOptionFields) UserId() *int64 {
	return qo.userId
}
