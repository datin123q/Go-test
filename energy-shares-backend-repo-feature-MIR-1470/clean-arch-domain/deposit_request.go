package clean_arch_domain

import (
	"context"
	"fmt"

	"github.com/samber/lo"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

type DepositRequestRepository interface {
	GetDepositRequestByShareId(ctx context.Context, shareId string) (domain.DepositRequest, error)
	GetById(ctx context.Context, id int64) (domain.DepositRequest, error)
	Update(ctx context.Context, depositRequest domain.DepositRequest) error
	GetAll(ctx context.Context, options ...DepositRequestQueryOption) ([]domain.DepositRequest, error)
	GetNoDepositRequestByUserId(ctx context.Context, userId int64) ([]domain.DepositRequest, error)
}

type DepositRequestLackHistoryRepository interface {
	Create(ctx context.Context, depositRequestLackHistory domain.DepositRequestLackHistory) error
}

type DepositRequestService interface {
	OkDeposit(ctx context.Context, depositRequest domain.DepositRequest) error
	CancelDeposit(ctx context.Context, depositRequest domain.DepositRequest) error
}

type DepositRequestQueryOption = func(*DepositRequestQueryOptionFields)

func DepositRequestQueryOptionFunc() DepositRequestQueryOptionFields {
	return DepositRequestQueryOptionFields{}
}

type DepositRequestQueryOptionFields struct {
	userId            *int64
	expireDepositDate *jsonutils.JSONTime
	QueryOptions
}

func (c DepositRequestQueryOptionFields) WithUserId(userId int64) func(*DepositRequestQueryOptionFields) {
	return func(fields *DepositRequestQueryOptionFields) {
		fields.userId = &userId
	}
}

func (c DepositRequestQueryOptionFields) WithSortBy(sortBy string) func(*DepositRequestQueryOptionFields) {
	return func(fields *DepositRequestQueryOptionFields) {
		fields.sortBy = &sortBy
	}
}

func (c DepositRequestQueryOptionFields) WithOrder(order string) func(*DepositRequestQueryOptionFields) {
	return func(qo *DepositRequestQueryOptionFields) {
		qo.order = &order
	}
}

func (c DepositRequestQueryOptionFields) WithOffset(offset int) func(*DepositRequestQueryOptionFields) {
	return func(fields *DepositRequestQueryOptionFields) {
		fields.offset = &offset
	}
}

func (c DepositRequestQueryOptionFields) WithLimit(limit int) func(*DepositRequestQueryOptionFields) {
	return func(fields *DepositRequestQueryOptionFields) {
		fields.limit = &limit
	}
}

func (c DepositRequestQueryOptionFields) WithExpireDepositDate(expireDepositDate jsonutils.JSONTime) func(*DepositRequestQueryOptionFields) {
	return func(fields *DepositRequestQueryOptionFields) {
		fields.expireDepositDate = &expireDepositDate
	}
}

func (c DepositRequestQueryOptionFields) UserId() *string {
	if c.userId == nil {
		return nil
	}

	return lo.ToPtr(fmt.Sprintf("%v", *c.userId))
}

func (c DepositRequestQueryOptionFields) ExpireDepositDate() string {
	if c.expireDepositDate == nil {
		return ""
	}

	return c.expireDepositDate.ToString()
}
