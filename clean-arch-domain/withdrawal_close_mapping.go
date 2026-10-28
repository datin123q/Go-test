package clean_arch_domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type WithdrawalCloseMappingRepository interface {
	Create(ctx context.Context, withdrawalCloseMapping domain.WithdrawalCloseMapping) (int64, error)
	CreateMany(ctx context.Context, withdrawalCloseMappings []domain.WithdrawalCloseMapping) error
	GetAll(ctx context.Context, options ...WithdrawalCloseMappingQueryOption) (int64, []domain.WithdrawalCloseMapping, error)
}

type WithdrawalCloseMappingQueryOption = func(*WithdrawalCloseMappingQueryOptionFields)

func WithdrawalCloseMappingQueryOptionFunc() WithdrawalCloseMappingQueryOptionFields {
	return WithdrawalCloseMappingQueryOptionFields{}
}

type WithdrawalCloseMappingQueryOptionFields struct {
	QueryOptions
	withdrawalIds []int64
}

func (qo WithdrawalCloseMappingQueryOptionFields) WithWithdrawalIds(withdrawalIds []int64) func(*WithdrawalCloseMappingQueryOptionFields) {
	return func(qo *WithdrawalCloseMappingQueryOptionFields) {
		qo.withdrawalIds = withdrawalIds
	}
}

func (qo WithdrawalCloseMappingQueryOptionFields) WithSortBy(sortBy string) func(*WithdrawalCloseMappingQueryOptionFields) {
	return func(qo *WithdrawalCloseMappingQueryOptionFields) {
		qo.sortBy = &sortBy
	}
}

func (qo WithdrawalCloseMappingQueryOptionFields) WithOrder(order string) func(*WithdrawalCloseMappingQueryOptionFields) {
	return func(qo *WithdrawalCloseMappingQueryOptionFields) {
		qo.order = &order
	}
}

func (qo WithdrawalCloseMappingQueryOptionFields) WithOffset(offset int) func(*WithdrawalCloseMappingQueryOptionFields) {
	return func(qo *WithdrawalCloseMappingQueryOptionFields) {
		qo.offset = &offset
	}
}

func (qo WithdrawalCloseMappingQueryOptionFields) WithLimit(limit int) func(*WithdrawalCloseMappingQueryOptionFields) {
	return func(qo *WithdrawalCloseMappingQueryOptionFields) {
		qo.limit = &limit
	}
}

func (qo WithdrawalCloseMappingQueryOptionFields) WithdrawalIds() []int64 {
	return qo.withdrawalIds
}
