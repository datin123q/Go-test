package clean_arch_domain

import (
	"context"
	"fmt"

	"github.com/samber/lo"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type UserAntisocialCheckService interface {
	GetByUserId(ctx context.Context, userId int64) (domain.UserAntisocialCheck, error)
}

type UserAntisocialCheckRepository interface {
	Create(ctx context.Context, userAntisocialCheck domain.UserAntisocialCheck) error
	GetByUserId(ctx context.Context, userId int64) (domain.UserAntisocialCheck, error)
}

type UserAntisocialCheckQueryOption = func(*UserAntisocialCheckQueryOptionFields)

func UserAntisocialCheckQueryOptionFunc() UserAntisocialCheckQueryOptionFields {
	return UserAntisocialCheckQueryOptionFields{}
}

type UserAntisocialCheckQueryOptionFields struct {
	userId *int64
	QueryOptions
}

func (u UserAntisocialCheckQueryOptionFields) WithCheckUserId(userId int64) func(*UserAntisocialCheckQueryOptionFields) {
	return func(fields *UserAntisocialCheckQueryOptionFields) {
		fields.userId = &userId
	}
}

func (u UserAntisocialCheckQueryOptionFields) WithSortBy(sortBy string) func(*UserAntisocialCheckQueryOptionFields) {
	return func(fields *UserAntisocialCheckQueryOptionFields) {
		fields.sortBy = &sortBy
	}
}

func (u UserAntisocialCheckQueryOptionFields) WithOrder(order string) func(*UserAntisocialCheckQueryOptionFields) {
	return func(qo *UserAntisocialCheckQueryOptionFields) {
		qo.order = &order
	}
}

func (u UserAntisocialCheckQueryOptionFields) WithOffset(offset int) func(*UserAntisocialCheckQueryOptionFields) {
	return func(fields *UserAntisocialCheckQueryOptionFields) {
		fields.offset = &offset
	}
}

func (u UserAntisocialCheckQueryOptionFields) WithLimit(limit int) func(*UserAntisocialCheckQueryOptionFields) {
	return func(fields *UserAntisocialCheckQueryOptionFields) {
		fields.limit = &limit
	}
}

func (u UserAntisocialCheckQueryOptionFields) CheckUserId() *string {
	if u.userId == nil {
		return nil
	}

	return lo.ToPtr(fmt.Sprintf("%v", *u.userId))
}
