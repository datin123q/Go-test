package clean_arch_domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type PolicyHistoryRepository interface {
	Create(ctx context.Context, history domain.PolicyInfoHistory) (int64, error)
	GetAll(ctx context.Context, options ...PolicyHistoryQueryOption) (int64, []domain.PolicyInfoHistory, error)
	GetById(ctx context.Context, id int64) (domain.PolicyInfoHistory, error)
	Update(ctx context.Context, history domain.PolicyInfoHistory) error
}

type PolicyHistoryQueryOption = func(*PolicyHistoryQueryOptionFields)

func PolicyHistoryQueryOptionFunc() PolicyHistoryQueryOptionFields {
	return PolicyHistoryQueryOptionFields{}
}

type PolicyHistoryQueryOptionFields struct {
	policyId *int64
	isPosted *bool
	label    *string
	QueryOptions
}

func (o PolicyHistoryQueryOptionFields) WithPolicyId(policyId int64) PolicyHistoryQueryOption {
	return func(o *PolicyHistoryQueryOptionFields) {
		o.policyId = &policyId
	}
}

func (o PolicyHistoryQueryOptionFields) WithIsPosted(isPosted *bool) PolicyHistoryQueryOption {
	return func(o *PolicyHistoryQueryOptionFields) {
		o.isPosted = isPosted
	}
}

func (o PolicyHistoryQueryOptionFields) WithLabel(label *string) PolicyHistoryQueryOption {
	return func(o *PolicyHistoryQueryOptionFields) {
		o.label = label
	}
}

func (o PolicyHistoryQueryOptionFields) WithSortBy(sortBy string) PolicyHistoryQueryOption {
	return func(o *PolicyHistoryQueryOptionFields) {
		o.sortBy = &sortBy
	}
}

func (o PolicyHistoryQueryOptionFields) WithOrder(order string) PolicyHistoryQueryOption {
	return func(o *PolicyHistoryQueryOptionFields) {
		o.order = &order
	}
}

func (o PolicyHistoryQueryOptionFields) WithOffset(offset int) PolicyHistoryQueryOption {
	return func(o *PolicyHistoryQueryOptionFields) {
		o.offset = &offset
	}
}

func (o PolicyHistoryQueryOptionFields) WithLimit(limit int) PolicyHistoryQueryOption {
	return func(o *PolicyHistoryQueryOptionFields) {
		o.limit = &limit
	}
}

func (o PolicyHistoryQueryOptionFields) PolicyId() *int64 {
	return o.policyId
}

func (o PolicyHistoryQueryOptionFields) IsPosted() *bool {
	return o.isPosted
}

func (o PolicyHistoryQueryOptionFields) Label() *string {
	return o.label
}
