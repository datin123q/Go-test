package clean_arch_domain

import (
	"context"
	"time"

	"github.com/samber/lo"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type NewsRepository interface {
	Create(ctx context.Context, news domain.News) (int64, error)
	Update(ctx context.Context, news domain.News) error
	GetAll(ctx context.Context, options ...NewsQueryOption) (int64, []domain.News, error)
	GetById(ctx context.Context, id int64) (domain.News, error)
	Delete(ctx context.Context, id int64) error
}

type NewsQueryOption = func(*NewsQueryOptionFields)

func NewsQueryOptionFunc() NewsQueryOptionFields {
	return NewsQueryOptionFields{}
}

type NewsQueryOptionFields struct {
	pin         *bool
	title       *string
	posted      *bool
	closingDate *string
	startDate   *string
	endDate     *string
	QueryOptions
}

func (qo NewsQueryOptionFields) WithPin(pin *bool) func(*NewsQueryOptionFields) {
	return func(qo *NewsQueryOptionFields) {
		qo.pin = pin
	}
}

func (qo NewsQueryOptionFields) WithTitle(title string) func(*NewsQueryOptionFields) {
	return func(qo *NewsQueryOptionFields) {
		qo.title = &title
	}
}

func (qo NewsQueryOptionFields) WithPosted(posted *bool) func(*NewsQueryOptionFields) {
	return func(qo *NewsQueryOptionFields) {
		qo.posted = posted
	}
}

func (qo NewsQueryOptionFields) WithClosingDate(onlyNotExpired *bool) func(*NewsQueryOptionFields) {
	return func(qo *NewsQueryOptionFields) {
		if onlyNotExpired != nil && *onlyNotExpired == true {
			qo.closingDate = lo.ToPtr(time.Now().Format(time.DateTime))
		}
	}
}

func (qo NewsQueryOptionFields) WithStartDate(startDate string) func(*NewsQueryOptionFields) {
	return func(qo *NewsQueryOptionFields) {
		qo.startDate = &startDate
	}
}

func (qo NewsQueryOptionFields) WithEndDate(endDate string) func(*NewsQueryOptionFields) {
	return func(qo *NewsQueryOptionFields) {
		qo.endDate = &endDate
	}
}

func (qo NewsQueryOptionFields) WithSortBy(sortBy string) func(*NewsQueryOptionFields) {
	return func(qo *NewsQueryOptionFields) {
		qo.sortBy = &sortBy
	}
}

func (qo NewsQueryOptionFields) WithOrder(order string) func(*NewsQueryOptionFields) {
	return func(qo *NewsQueryOptionFields) {
		qo.order = &order
	}
}

func (qo NewsQueryOptionFields) WithOffset(offset int) func(*NewsQueryOptionFields) {
	return func(qo *NewsQueryOptionFields) {
		qo.offset = &offset
	}
}

func (qo NewsQueryOptionFields) WithLimit(limit int) func(*NewsQueryOptionFields) {
	return func(qo *NewsQueryOptionFields) {
		qo.limit = &limit
	}
}

func (qo NewsQueryOptionFields) Pin() *bool {
	return qo.pin
}

func (qo NewsQueryOptionFields) Title() *string {
	return qo.title
}

func (qo NewsQueryOptionFields) Posted() *bool {
	return qo.posted
}

func (qo NewsQueryOptionFields) ClosingDate() *string {
	return qo.closingDate
}

func (qo NewsQueryOptionFields) StartDate() *string {
	return qo.startDate
}

func (qo NewsQueryOptionFields) EndDate() *string {
	return qo.endDate
}
