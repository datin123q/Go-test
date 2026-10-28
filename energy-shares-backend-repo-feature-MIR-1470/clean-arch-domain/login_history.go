package clean_arch_domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type LoginHistoryRepository interface {
	GetAll(ctx context.Context, options ...LoginHistoryQueryOption) (int64, []domain.LoginHistory, error)
	Create(ctx context.Context, loginHistory domain.LoginHistory) (int64, error)
}

type LoginHistoryQueryOption = func(*LoginHistoryQueryOptionFields)

func LoginHistoryQueryOptionFunc() LoginHistoryQueryOptionFields {
	return LoginHistoryQueryOptionFields{}
}

type LoginHistoryQueryOptionFields struct {
	name      *string `query:"name" swagger:"desc(처리자)"`
	startDate string  `query:"start_date" swagger:"desc(로그인 타임 검색 시작 시간)"`
	endDate   string  `query:"end_date" swagger:"desc(로그인 타임 검색 종료 시간)"`
	QueryOptions
}

func (f LoginHistoryQueryOptionFields) Name() *string {
	return f.name
}

func (f LoginHistoryQueryOptionFields) StartDate() string {
	return f.startDate
}

func (f LoginHistoryQueryOptionFields) EndDate() string {
	return f.endDate
}

func (f LoginHistoryQueryOptionFields) WithName(name string) func(*LoginHistoryQueryOptionFields) {
	return func(f *LoginHistoryQueryOptionFields) {
		f.name = &name
	}
}

func (f LoginHistoryQueryOptionFields) WithStartDate(startDate string) func(*LoginHistoryQueryOptionFields) {
	return func(f *LoginHistoryQueryOptionFields) {
		f.startDate = startDate
	}
}

func (f LoginHistoryQueryOptionFields) WithEndDate(endDate string) func(*LoginHistoryQueryOptionFields) {
	return func(f *LoginHistoryQueryOptionFields) {
		f.endDate = endDate
	}
}

func (f LoginHistoryQueryOptionFields) WithSortBy(sortBy string) func(*LoginHistoryQueryOptionFields) {
	return func(f *LoginHistoryQueryOptionFields) {
		f.sortBy = &sortBy
	}
}

func (f LoginHistoryQueryOptionFields) WithOrder(order string) func(*LoginHistoryQueryOptionFields) {
	return func(f *LoginHistoryQueryOptionFields) {
		f.order = &order
	}
}

func (f LoginHistoryQueryOptionFields) WithOffset(offset int) func(*LoginHistoryQueryOptionFields) {
	return func(f *LoginHistoryQueryOptionFields) {
		f.offset = &offset
	}
}

func (f LoginHistoryQueryOptionFields) WithLimit(limit int) func(*LoginHistoryQueryOptionFields) {
	return func(f *LoginHistoryQueryOptionFields) {
		f.limit = &limit
	}
}
