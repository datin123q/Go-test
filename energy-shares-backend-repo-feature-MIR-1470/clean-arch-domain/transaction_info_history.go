package clean_arch_domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type TransactionInfoHistoryRepository interface {
	Create(ctx context.Context, transaction domain.TransactionInfoHistory) error
	GetAll(ctx context.Context, options ...TransactionInfoHistoryQueryOption) (int64, []domain.TransactionInfoHistory, error)
}

type TransactionInfoHistoryQueryOption = func(*TransactionInfoHistoryQueryOptionFields)

func TransactionInfoHistoryQueryOptionFunc() TransactionInfoHistoryQueryOptionFields {
	return TransactionInfoHistoryQueryOptionFields{}
}

type TransactionInfoHistoryQueryOptionFields struct {
	shareId *string
	QueryOptions
}

func (c TransactionInfoHistoryQueryOptionFields) WithShareId(shareId string) func(*TransactionInfoHistoryQueryOptionFields) {
	return func(fields *TransactionInfoHistoryQueryOptionFields) {
		fields.shareId = &shareId
	}
}

func (c TransactionInfoHistoryQueryOptionFields) WithSortBy(sortBy string) func(*TransactionInfoHistoryQueryOptionFields) {
	return func(fields *TransactionInfoHistoryQueryOptionFields) {
		fields.sortBy = &sortBy
	}
}

func (c TransactionInfoHistoryQueryOptionFields) WithOrder(order string) func(*TransactionInfoHistoryQueryOptionFields) {
	return func(qo *TransactionInfoHistoryQueryOptionFields) {
		qo.order = &order
	}
}

func (c TransactionInfoHistoryQueryOptionFields) WithOffset(offset int) func(*TransactionInfoHistoryQueryOptionFields) {
	return func(fields *TransactionInfoHistoryQueryOptionFields) {
		fields.offset = &offset
	}
}

func (c TransactionInfoHistoryQueryOptionFields) WithLimit(limit int) func(*TransactionInfoHistoryQueryOptionFields) {
	return func(fields *TransactionInfoHistoryQueryOptionFields) {
		fields.limit = &limit
	}
}

func (c TransactionInfoHistoryQueryOptionFields) ShareId() *string {
	return c.shareId
}
