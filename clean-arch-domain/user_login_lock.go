package clean_arch_domain

import (
	"context"
	"fmt"

	"github.com/samber/lo"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type UserLoginLockRepository interface {
	GetAll(ctx context.Context, options ...UserLoginLockQueryOption) (int64, []domain.UserLoginLock, error)
	GetByUserId(ctx context.Context, userId int64) (domain.UserLoginLock, error)
	Create(ctx context.Context, loginLock domain.UserLoginLock) (int64, error)
	Update(ctx context.Context, loginLock domain.UserLoginLock) error
	Delete(ctx context.Context, Id int64) error
}

type UserLoginLockQueryOption = func(*UserLoginLockQueryOptionFields)

func UserLoginLockQueryOptionFunc() UserLoginLockQueryOptionFields {
	return UserLoginLockQueryOptionFields{}
}

type UserLoginLockQueryOptionFields struct {
	userId    *int64
	startDate string
	endDate   string
	QueryOptions
}

func (f UserLoginLockQueryOptionFields) WithUserId(userId int64) func(*UserLoginLockQueryOptionFields) {
	return func(f *UserLoginLockQueryOptionFields) {
		f.userId = &userId
	}
}

func (f UserLoginLockQueryOptionFields) WithStartDate(startDate string) func(*UserLoginLockQueryOptionFields) {
	return func(f *UserLoginLockQueryOptionFields) {
		f.startDate = startDate
	}
}

func (f UserLoginLockQueryOptionFields) WithEndDate(endDate string) func(*UserLoginLockQueryOptionFields) {
	return func(f *UserLoginLockQueryOptionFields) {
		f.endDate = endDate
	}
}

func (f UserLoginLockQueryOptionFields) WithSortBy(sortBy string) func(*UserLoginLockQueryOptionFields) {
	return func(f *UserLoginLockQueryOptionFields) {
		f.sortBy = &sortBy
	}
}

func (f UserLoginLockQueryOptionFields) WithOrder(order string) func(*UserLoginLockQueryOptionFields) {
	return func(f *UserLoginLockQueryOptionFields) {
		f.order = &order
	}
}

func (f UserLoginLockQueryOptionFields) WithOffset(offset int) func(*UserLoginLockQueryOptionFields) {
	return func(f *UserLoginLockQueryOptionFields) {
		f.offset = &offset
	}
}

func (f UserLoginLockQueryOptionFields) WithLimit(limit int) func(*UserLoginLockQueryOptionFields) {
	return func(f *UserLoginLockQueryOptionFields) {
		f.limit = &limit
	}
}

func (f UserLoginLockQueryOptionFields) UserId() *string {
	if f.userId == nil {
		return nil
	}

	return lo.ToPtr(fmt.Sprintf("%v", *f.userId))
}

func (f UserLoginLockQueryOptionFields) StartDate() string {
	return f.startDate
}

func (f UserLoginLockQueryOptionFields) EndDate() string {
	return f.endDate
}
