package clean_arch_domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
)

type PolicyService interface {
	CreatePolicyHistory(ctx context.Context, input PolicyHistoryInput) error
}

type PolicyRepository interface {
	Create(ctx context.Context, policy domain.Policy) (int64, error)
	GetAll(ctx context.Context, options ...PolicyQueryOption) (int64, []domain.Policy, error)
	GetById(ctx context.Context, id int64) (domain.Policy, error)
	Update(ctx context.Context, policy domain.Policy) error
	Delete(ctx context.Context, id int64) error
}

type PolicyHistoryInput struct {
	Before   domain.Policy
	After    domain.Policy
	Modifier jwthelper.User
}

type PolicyQueryOption = func(*PolicyQueryOptionFields)

func PolicyQueryOptionFunc() PolicyQueryOptionFields {
	return PolicyQueryOptionFields{}
}

type PolicyQueryOptionFields struct {
	policyViewType *string
	policyType     *string
	QueryOptions
}

func (o PolicyQueryOptionFields) WithPolicyViewType(policyViewType *string) func(*PolicyQueryOptionFields) {
	return func(o *PolicyQueryOptionFields) {
		o.policyViewType = policyViewType
	}
}

func (o PolicyQueryOptionFields) WithPolicyType(policyType *string) func(*PolicyQueryOptionFields) {
	return func(o *PolicyQueryOptionFields) {
		o.policyType = policyType
	}
}

func (o PolicyQueryOptionFields) WithSortBy(sortBy string) func(*PolicyQueryOptionFields) {
	return func(o *PolicyQueryOptionFields) {
		o.sortBy = &sortBy
	}
}

func (o PolicyQueryOptionFields) WithOrder(order string) func(*PolicyQueryOptionFields) {
	return func(o *PolicyQueryOptionFields) {
		o.order = &order
	}
}

func (o PolicyQueryOptionFields) WithOffset(offset int) func(*PolicyQueryOptionFields) {
	return func(o *PolicyQueryOptionFields) {
		o.offset = &offset
	}
}

func (o PolicyQueryOptionFields) WithLimit(limit int) func(*PolicyQueryOptionFields) {
	return func(o *PolicyQueryOptionFields) {
		o.limit = &limit
	}
}

func (o PolicyQueryOptionFields) PolicyViewType() *string {
	return o.policyViewType
}

func (o PolicyQueryOptionFields) PolicyType() *string {
	return o.policyType
}
