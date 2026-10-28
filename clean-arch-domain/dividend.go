package clean_arch_domain

import (
	"context"
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type DividendRepository interface {
	CreateMany(ctx context.Context, dividends []domain.Dividend) error
	DeleteByScheduleId(ctx context.Context, options ...DividendQueryOption) error
	GetAll(ctx context.Context, options ...DividendQueryOption) ([]domain.Dividend, error)
}

type DividendQueryOption = func(*DividendQueryOptionFields)

func DividendQueryOptionFunc() DividendQueryOptionFields {
	return DividendQueryOptionFields{}
}

type DividendQueryOptionFields struct {
	userId     *int64
	spvId      *int64
	scheduleId *int64
	startDate  time.Time
	endDate    time.Time
	QueryOptions
}

func (qo DividendQueryOptionFields) WithUserId(userId int64) func(*DividendQueryOptionFields) {
	return func(qo *DividendQueryOptionFields) {
		qo.userId = &userId
	}
}

func (qo DividendQueryOptionFields) WithSpvId(spvId int64) func(*DividendQueryOptionFields) {
	return func(qo *DividendQueryOptionFields) {
		qo.spvId = &spvId
	}
}

func (qo DividendQueryOptionFields) WithScheduleId(scheduleId int64) func(*DividendQueryOptionFields) {
	return func(qo *DividendQueryOptionFields) {
		qo.scheduleId = &scheduleId
	}
}

func (qo DividendQueryOptionFields) withStartDate(startDate time.Time) func(*DividendQueryOptionFields) {
	return func(qo *DividendQueryOptionFields) {
		qo.startDate = startDate
	}
}

func (qo DividendQueryOptionFields) withEndDate(endDate time.Time) func(*DividendQueryOptionFields) {
	return func(qo *DividendQueryOptionFields) {
		qo.endDate = endDate
	}
}

func (qo DividendQueryOptionFields) WithSortBy(sortBy string) func(*DividendQueryOptionFields) {
	return func(qo *DividendQueryOptionFields) {
		qo.sortBy = &sortBy
	}
}

func (qo DividendQueryOptionFields) WithOrder(order string) func(*DividendQueryOptionFields) {
	return func(qo *DividendQueryOptionFields) {
		qo.order = &order
	}
}

func (qo DividendQueryOptionFields) WithOffset(offset int) func(*DividendQueryOptionFields) {
	return func(iqo *DividendQueryOptionFields) {
		qo.offset = &offset
	}
}

func (qo DividendQueryOptionFields) WithLimit(limit int) func(*DividendQueryOptionFields) {
	return func(qo *DividendQueryOptionFields) {
		qo.limit = &limit
	}
}

func (qo DividendQueryOptionFields) UserId() *int64 {
	return qo.userId
}

func (qo DividendQueryOptionFields) SpvId() *int64 {
	return qo.spvId
}

func (qo DividendQueryOptionFields) ScheduleId() *int64 {
	return qo.scheduleId
}

func (qo DividendQueryOptionFields) StartDate() time.Time {
	return qo.startDate
}

func (qo DividendQueryOptionFields) EndDate() time.Time {
	return qo.endDate
}
