package clean_arch_domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type ContractRepository interface {
	Create(ctx context.Context, contract domain.Contract) (int64, error)
	GetAll(ctx context.Context, opts ...ContractQueryOption) (domain.Contracts, error)
}

type ContractQueryOption = func(*ContractQueryOptionFields)

func ContractQueryOptionFunc() ContractQueryOptionFields {
	return ContractQueryOptionFields{}
}

type ContractQueryOptionFields struct {
	shareId *string
	QueryOptions
}

func (c ContractQueryOptionFields) WithShareId(shareId string) func(*ContractQueryOptionFields) {
	return func(cqo *ContractQueryOptionFields) {
		cqo.shareId = &shareId
	}
}

func (c ContractQueryOptionFields) WithSortBy(sortBy string) func(*ContractQueryOptionFields) {
	return func(qo *ContractQueryOptionFields) {
		qo.sortBy = &sortBy
	}
}

func (c ContractQueryOptionFields) WithOrder(order string) func(*ContractQueryOptionFields) {
	return func(qo *ContractQueryOptionFields) {
		qo.order = &order
	}
}

func (c ContractQueryOptionFields) WithOffset(offset int) func(*ContractQueryOptionFields) {
	return func(qo *ContractQueryOptionFields) {
		qo.offset = &offset
	}
}

func (c ContractQueryOptionFields) WithLimit(limit int) func(*ContractQueryOptionFields) {
	return func(qo *ContractQueryOptionFields) {
		qo.limit = &limit
	}
}

func (c ContractQueryOptionFields) ShareId() *string {
	return c.shareId
}
