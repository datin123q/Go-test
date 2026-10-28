package clean_arch_domain

import (
	"context"

	"github.com/samber/lo"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

const maxSentSignUpRequestMail = 1

type SignUpRequestRepository interface {
	Create(ctx context.Context, signUpRequest domain.SignUpRequest) (int64, error)
	Update(ctx context.Context, signUpRequest domain.SignUpRequest) error
	GetByRequestId(ctx context.Context, requestId string) (domain.SignUpRequest, error)
	GetAll(ctx context.Context, opts ...SignUpRequestQueryOption) (SignUpRequests, error)
	Delete(ctx context.Context, id int64) error
}

type SignUpRequests []domain.SignUpRequest

func (s SignUpRequests) IsEmpty() bool {
	if len(s) == 0 {
		return true
	}

	return false
}

func (s SignUpRequests) IsSentMailLimitExceeded() bool {
	if s.IsEmpty() {
		return false
	}

	return s[0].SendCount >= maxSentSignUpRequestMail
}

type SignUpRequestQueryOption = func(*SignUpRequestQueryOptionFields)

func SignUpRequestQueryOptionFunc() SignUpRequestQueryOptionFields {
	return SignUpRequestQueryOptionFields{}
}

type SignUpRequestQueryOptionFields struct {
	email       *string
	urlExpire   *jsonutils.JSONTime
	emailExpire *jsonutils.JSONTime
	QueryOptions
}

func (sqf SignUpRequestQueryOptionFields) WithEmail(email string) SignUpRequestQueryOption {
	return func(sqf *SignUpRequestQueryOptionFields) {
		sqf.email = &email
	}
}

func (sqf SignUpRequestQueryOptionFields) WithSortBy(sortBy string) SignUpRequestQueryOption {
	return func(sqf *SignUpRequestQueryOptionFields) {
		sqf.sortBy = &sortBy
	}
}

func (sqf SignUpRequestQueryOptionFields) WithOrder(order string) SignUpRequestQueryOption {
	return func(sqf *SignUpRequestQueryOptionFields) {
		sqf.order = &order
	}
}

func (sqf SignUpRequestQueryOptionFields) WithLimit(limit int) SignUpRequestQueryOption {
	return func(sqf *SignUpRequestQueryOptionFields) {
		sqf.limit = &limit
	}
}

func (sqf SignUpRequestQueryOptionFields) WithOffset(offset int) SignUpRequestQueryOption {
	return func(sqf *SignUpRequestQueryOptionFields) {
		sqf.offset = &offset
	}
}

func (sqf SignUpRequestQueryOptionFields) WithUrlExpire(urlExpire jsonutils.JSONTime) SignUpRequestQueryOption {
	return func(sqf *SignUpRequestQueryOptionFields) {
		sqf.urlExpire = &urlExpire
	}
}

func (sqf SignUpRequestQueryOptionFields) WithEmailExpire(emailExpire jsonutils.JSONTime) SignUpRequestQueryOption {
	return func(sqf *SignUpRequestQueryOptionFields) {
		sqf.emailExpire = &emailExpire
	}
}

func (sqf SignUpRequestQueryOptionFields) Email() *string {
	return sqf.email
}

func (sqf SignUpRequestQueryOptionFields) UrlExpire() *string {
	if sqf.urlExpire == nil {
		return nil
	}

	return lo.ToPtr(sqf.urlExpire.ToString())
}

func (sqf SignUpRequestQueryOptionFields) EmailExpire() *string {
	if sqf.emailExpire == nil {
		return nil
	}

	return lo.ToPtr(sqf.emailExpire.ToString())
}
