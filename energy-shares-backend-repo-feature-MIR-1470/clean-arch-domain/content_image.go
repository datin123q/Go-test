package clean_arch_domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type ContentImageRepository interface {
	Create(ctx context.Context, contentImage domain.ContentImage) (int64, error)
	GetAll(ctx context.Context, options ...ContentImageQueryOption) ([]domain.ContentImage, error)
	DeleteByResourceId(ctx context.Context, resourceId string) error
	Delete(ctx context.Context, id int64) error
}

type ContentImageQueryOption = func(*ContentImageQueryOptionFields)

func ContentImageQueryOptionFunc() ContentImageQueryOptionFields {
	return ContentImageQueryOptionFields{}
}

type ContentImageQueryOptionFields struct {
	resourceId *string
	QueryOptions
}

func (c ContentImageQueryOptionFields) WithResourceId(resourceId string) func(*ContentImageQueryOptionFields) {
	return func(cqo *ContentImageQueryOptionFields) {
		cqo.resourceId = &resourceId
	}
}

func (c ContentImageQueryOptionFields) WithSortBy(sortBy string) func(*ContentImageQueryOptionFields) {
	return func(qo *ContentImageQueryOptionFields) {
		qo.sortBy = &sortBy
	}
}

func (c ContentImageQueryOptionFields) WithOrder(order string) func(*ContentImageQueryOptionFields) {
	return func(qo *ContentImageQueryOptionFields) {
		qo.order = &order
	}
}

func (c ContentImageQueryOptionFields) WithOffset(offset int) func(*ContentImageQueryOptionFields) {
	return func(qo *ContentImageQueryOptionFields) {
		qo.offset = &offset
	}
}

func (c ContentImageQueryOptionFields) WithLimit(limit int) func(*ContentImageQueryOptionFields) {
	return func(qo *ContentImageQueryOptionFields) {
		qo.limit = &limit
	}
}

func (c ContentImageQueryOptionFields) ResourceId() *string {
	return c.resourceId
}
