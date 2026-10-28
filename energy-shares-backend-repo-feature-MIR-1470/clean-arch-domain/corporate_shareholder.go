package clean_arch_domain

import (
	"context"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type CorporateShareholderRepository interface {
	Create(ctx context.Context, corporateShareholder domain.CorporateShareholder) (int64, error)
	GetAll(ctx context.Context, options ...CorporateShareholderOption) (int64, []domain.CorporateShareholder, error)
	DeleteByUserId(ctx context.Context, userId int64) error
}

type CorporateShareholderQueryOptionFields struct {
	QueryOptions
	userId *int64
}

type CorporateShareholderOption = func(*CorporateShareholderQueryOptionFields)

func CorporateShareholderQueryOptionFunc() CorporateShareholderQueryOptionFields {
	return CorporateShareholderQueryOptionFields{}
}

func (cqf CorporateShareholderQueryOptionFields) WithSortBy(sortBy string) func(*CorporateShareholderQueryOptionFields) {
	return func(sqo *CorporateShareholderQueryOptionFields) {
		sqo.sortBy = &sortBy
	}
}

func (cqf CorporateShareholderQueryOptionFields) WithOrder(order string) func(*CorporateShareholderQueryOptionFields) {
	return func(sqo *CorporateShareholderQueryOptionFields) {
		sqo.order = &order
	}
}

func (cqf CorporateShareholderQueryOptionFields) WithOffset(offset int) func(*CorporateShareholderQueryOptionFields) {
	return func(sqo *CorporateShareholderQueryOptionFields) {
		sqo.offset = &offset
	}
}

func (cqf CorporateShareholderQueryOptionFields) WithLimit(limit int) func(*CorporateShareholderQueryOptionFields) {
	return func(sqo *CorporateShareholderQueryOptionFields) {
		sqo.limit = &limit
	}
}

func (cqf CorporateShareholderQueryOptionFields) WithUserId(userId int64) func(*CorporateShareholderQueryOptionFields) {
	return func(sqo *CorporateShareholderQueryOptionFields) {
		sqo.userId = &userId
	}
}

func (cqf CorporateShareholderQueryOptionFields) UserId() *int64 {
	return cqf.userId
}
