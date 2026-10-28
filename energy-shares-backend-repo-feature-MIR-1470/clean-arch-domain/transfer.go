package clean_arch_domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type TransferRepository interface {
	GetAll(ctx context.Context, options ...TransferQueryOption) (int64, []domain.Transfer, error)
}

type TransferQueryOption = func(*TransferQueryOptionFields)

func TransferQueryOptionFunc() TransferQueryOptionFields {
	return TransferQueryOptionFields{}
}

type TransferQueryOptionFields struct {
	QueryOptions
	requestId    *int64
	withdrawalId *int64
}

func (qo TransferQueryOptionFields) WithRequestId(requestId *int64) func(*TransferQueryOptionFields) {
	return func(fields *TransferQueryOptionFields) {
		fields.requestId = requestId
	}
}

func (qo TransferQueryOptionFields) WithWithdrawalId(withdrawalId *int64) func(*TransferQueryOptionFields) {
	return func(fields *TransferQueryOptionFields) {
		fields.withdrawalId = withdrawalId
	}
}

func (qo TransferQueryOptionFields) WithSortBy(sortBy string) func(*TransferQueryOptionFields) {
	return func(qo *TransferQueryOptionFields) {
		qo.sortBy = &sortBy
	}
}

func (qo TransferQueryOptionFields) WithOrder(order string) func(*TransferQueryOptionFields) {
	return func(qo *TransferQueryOptionFields) {
		qo.order = &order
	}
}

func (qo TransferQueryOptionFields) WithLimit(limit int) func(*TransferQueryOptionFields) {
	return func(qo *TransferQueryOptionFields) {
		qo.limit = &limit
	}
}

func (qo TransferQueryOptionFields) WithOffset(offset int) func(*TransferQueryOptionFields) {
	return func(qo *TransferQueryOptionFields) {
		qo.offset = &offset
	}
}

func (qo TransferQueryOptionFields) RequestId() *int64 {
	return qo.requestId
}

func (qo TransferQueryOptionFields) WithdrawalId() *int64 {
	return qo.withdrawalId
}
