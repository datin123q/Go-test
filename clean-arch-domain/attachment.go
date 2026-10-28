package clean_arch_domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type AttachmentRepository interface {
	Create(ctx context.Context, attachment domain.Attachment) (int64, error)
	GetAll(ctx context.Context, options ...AttachmentQueryOption) ([]domain.Attachment, error)
	DeleteByResourceId(ctx context.Context, resourceId string) error
	Delete(ctx context.Context, id int64) error
}

type AttachmentQueryOption = func(*AttachmentQueryOptionFields)

func AttachmentQueryOptionFunc() AttachmentQueryOptionFields {
	return AttachmentQueryOptionFields{}
}

type AttachmentQueryOptionFields struct {
	resourceId *string
	QueryOptions
}

func (c AttachmentQueryOptionFields) WithResourceId(resourceId string) func(*AttachmentQueryOptionFields) {
	return func(cqo *AttachmentQueryOptionFields) {
		cqo.resourceId = &resourceId
	}
}

func (c AttachmentQueryOptionFields) WithSortBy(sortBy string) func(*AttachmentQueryOptionFields) {
	return func(qo *AttachmentQueryOptionFields) {
		qo.sortBy = &sortBy
	}
}

func (c AttachmentQueryOptionFields) WithOrder(order string) func(*AttachmentQueryOptionFields) {
	return func(qo *AttachmentQueryOptionFields) {
		qo.order = &order
	}
}

func (c AttachmentQueryOptionFields) WithOffset(offset int) func(*AttachmentQueryOptionFields) {
	return func(qo *AttachmentQueryOptionFields) {
		qo.offset = &offset
	}
}

func (c AttachmentQueryOptionFields) WithLimit(limit int) func(*AttachmentQueryOptionFields) {
	return func(qo *AttachmentQueryOptionFields) {
		qo.limit = &limit
	}
}

func (c AttachmentQueryOptionFields) ResourceId() *string {
	return c.resourceId
}
