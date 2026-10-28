package clean_arch_domain

import (
	"context"
	"fmt"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/samber/lo"
)

type EmailCodeRequestRepository interface {
	Create(ctx context.Context, emailCodeRequest domain.EmailCodeRequest) error
	Update(ctx context.Context, emailCodeRequest domain.EmailCodeRequest) error
	GetAll(ctx context.Context, options ...EmailCodeRequestQueryOption) ([]domain.EmailCodeRequest, error)
}

type EmailCodeRequestQueryOption = func(*EmailCodeRequestQueryOptionFields)

func EmailCodeRequestQueryOptionFunc() EmailCodeRequestQueryOptionFields {
	return EmailCodeRequestQueryOptionFields{}
}

type EmailCodeRequestQueryOptionFields struct {
	userId     *int64
	firstReq   *bool
	gteCreated *string
	code       *string
	email      *string
	QueryOptions
}

func (e EmailCodeRequestQueryOptionFields) WithUserId(userId int64) func(*EmailCodeRequestQueryOptionFields) {
	return func(fields *EmailCodeRequestQueryOptionFields) {
		fields.userId = &userId
	}
}

func (e EmailCodeRequestQueryOptionFields) UserId() *string {
	if e.userId == nil {
		return nil
	}

	return lo.ToPtr(fmt.Sprintf("%v", *e.userId))
}

func (e EmailCodeRequestQueryOptionFields) WithFirstReq(firstReq bool) func(*EmailCodeRequestQueryOptionFields) {
	return func(fields *EmailCodeRequestQueryOptionFields) {
		fields.firstReq = &firstReq
	}
}

func (e EmailCodeRequestQueryOptionFields) FirstReq() *bool {
	if e.firstReq == nil {
		return nil
	}

	return e.firstReq
}

func (e EmailCodeRequestQueryOptionFields) WithGTECreated(gteCreated string) func(*EmailCodeRequestQueryOptionFields) {
	return func(fields *EmailCodeRequestQueryOptionFields) {
		fields.gteCreated = &gteCreated
	}
}

func (e EmailCodeRequestQueryOptionFields) GTECreated() *string {
	if e.gteCreated == nil {
		return nil
	}

	return lo.ToPtr(fmt.Sprintf("%v", *e.gteCreated))
}

func (e EmailCodeRequestQueryOptionFields) WithCode(code string) func(*EmailCodeRequestQueryOptionFields) {
	return func(fields *EmailCodeRequestQueryOptionFields) {
		fields.code = &code
	}
}

func (e EmailCodeRequestQueryOptionFields) Code() *string {
	if e.code == nil {
		return nil
	}

	return e.code
}

func (e EmailCodeRequestQueryOptionFields) WithEmail(email string) func(*EmailCodeRequestQueryOptionFields) {
	return func(fields *EmailCodeRequestQueryOptionFields) {
		fields.email = &email
	}
}

func (e EmailCodeRequestQueryOptionFields) Email() *string {
	if e.email == nil {
		return nil
	}

	return e.email
}

func (e EmailCodeRequestQueryOptionFields) WithSortBy(sortBy string) func(*EmailCodeRequestQueryOptionFields) {
	return func(fields *EmailCodeRequestQueryOptionFields) {
		fields.sortBy = &sortBy
	}
}

func (e EmailCodeRequestQueryOptionFields) WithOrder(order string) func(*EmailCodeRequestQueryOptionFields) {
	return func(fields *EmailCodeRequestQueryOptionFields) {
		fields.order = &order
	}
}
