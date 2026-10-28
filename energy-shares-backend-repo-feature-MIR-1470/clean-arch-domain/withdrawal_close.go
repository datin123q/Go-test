package clean_arch_domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type WithdrawalCloseRepository interface {
	Create(ctx context.Context, withdrawalClose domain.WithdrawalClose) (int64, error)
	Update(ctx context.Context, withdrawalClose domain.WithdrawalClose) error
	UpdateMany(ctx context.Context, withdrawalClose domain.WithdrawalClose, options ...WithdrawalCloseQueryOption) error
	GetAll(ctx context.Context, options ...WithdrawalCloseQueryOption) (int64, []domain.WithdrawalCloseDetail, error)
}

type WithdrawalCloseQueryOption = func(*WithdrawalCloseQueryOptionFields)

func WithdrawalCloseQueryOptionFunc() WithdrawalCloseQueryOptionFields {
	return WithdrawalCloseQueryOptionFields{}
}

type WithdrawalCloseQueryOptionFields struct {
	ids []int64
	QueryOptions
}

func (qo WithdrawalCloseQueryOptionFields) WithIds(ids []int64) WithdrawalCloseQueryOption {
	return func(qo *WithdrawalCloseQueryOptionFields) {
		qo.ids = ids
	}
}

func (qo WithdrawalCloseQueryOptionFields) WithSortBy(sortBy string) WithdrawalCloseQueryOption {
	return func(qo *WithdrawalCloseQueryOptionFields) {
		qo.sortBy = &sortBy
	}
}

func (qo WithdrawalCloseQueryOptionFields) WithOrder(order string) WithdrawalCloseQueryOption {
	return func(qo *WithdrawalCloseQueryOptionFields) {
		qo.order = &order
	}
}

func (qo WithdrawalCloseQueryOptionFields) WithOffset(offset int) WithdrawalCloseQueryOption {
	return func(qo *WithdrawalCloseQueryOptionFields) {
		qo.offset = &offset
	}
}

func (qo WithdrawalCloseQueryOptionFields) WithLimit(limit int) WithdrawalCloseQueryOption {
	return func(qo *WithdrawalCloseQueryOptionFields) {
		qo.limit = &limit
	}
}

func (qo WithdrawalCloseQueryOptionFields) Ids() []int64 {
	return qo.ids
}
