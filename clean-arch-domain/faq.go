package clean_arch_domain

import (
	"context"
	"errors"
	"time"

	"github.com/samber/lo"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

type FaqRepository interface {
	Create(ctx context.Context, faq domain.Faq) (int64, error)
	Update(ctx context.Context, faq domain.Faq) error
	GetAll(ctx context.Context, options ...FaqQueryOption) (int64, []domain.Faq, error)
	GetById(ctx context.Context, id int64) (domain.Faq, error)
	Delete(ctx context.Context, id int64) error
}

type FaqQueryOption = func(*FaqQueryOptionFields)

func FaqQueryOptionFunc() FaqQueryOptionFields {
	return FaqQueryOptionFields{}
}

type FaqQueryOptionFields struct {
	title        *string
	categoryType *string
	posted       *bool
	pin          *bool
	startDate    *string
	endDate      *string
	QueryOptions
}

func (qo FaqQueryOptionFields) WithTitle(title *string) func(*FaqQueryOptionFields) {
	return func(qo *FaqQueryOptionFields) {
		qo.title = title
	}
}

func (qo FaqQueryOptionFields) WithCategoryType(categoryType *string) func(*FaqQueryOptionFields) {
	return func(qo *FaqQueryOptionFields) {
		qo.categoryType = categoryType
	}
}

func (qo FaqQueryOptionFields) WithPosted(posted *bool) func(*FaqQueryOptionFields) {
	return func(qo *FaqQueryOptionFields) {
		qo.posted = posted
	}
}

func (qo FaqQueryOptionFields) WithPin(pin *bool) func(*FaqQueryOptionFields) {
	return func(qo *FaqQueryOptionFields) {
		qo.pin = pin
	}
}

func (qo FaqQueryOptionFields) WithStartDate(startDate *string) func(*FaqQueryOptionFields) {
	return func(qo *FaqQueryOptionFields) {
		qo.startDate = startDate
	}
}

func (qo FaqQueryOptionFields) WithEndDate(endDate *string) func(*FaqQueryOptionFields) {
	return func(qo *FaqQueryOptionFields) {
		qo.endDate = endDate
	}
}

func (qo FaqQueryOptionFields) WithSortBy(sortBy string) func(*FaqQueryOptionFields) {
	return func(qo *FaqQueryOptionFields) {
		qo.sortBy = &sortBy
	}
}

func (qo FaqQueryOptionFields) WithOrder(order string) func(*FaqQueryOptionFields) {
	return func(qo *FaqQueryOptionFields) {
		qo.order = &order
	}
}

func (qo FaqQueryOptionFields) WithOffset(offset int) func(*FaqQueryOptionFields) {
	return func(qo *FaqQueryOptionFields) {
		qo.offset = &offset
	}
}

func (qo FaqQueryOptionFields) WithLimit(limit int) func(*FaqQueryOptionFields) {
	return func(qo *FaqQueryOptionFields) {
		qo.limit = &limit
	}
}

func (qo FaqQueryOptionFields) Title() *string {
	return qo.title
}

func (qo FaqQueryOptionFields) CategoryType() *string {
	return qo.categoryType
}

func (qo FaqQueryOptionFields) Posted() *bool {
	return qo.posted
}

func (qo FaqQueryOptionFields) Pin() *bool {
	return qo.pin
}

func (qo FaqQueryOptionFields) StartDate() time.Time {
	if qo.startDate == nil {
		return time.Time{}
	}

	return lo.Must(parseMultiFormat(*qo.startDate))
}

func (qo FaqQueryOptionFields) EndDate() time.Time {
	if qo.endDate == nil {
		return time.Time{}
	}

	return lo.Must(parseMultiFormat(*qo.endDate))
}

func parseMultiFormat(value string) (time.Time, error) {
	var timeFormats = []string{
		time.RFC3339,
		time.DateTime,
		time.DateOnly,
		"2006-01-02T15:04:05 09:00", // 非構造化型
	}

	for _, format := range timeFormats {
		if t, err := jsonutils.ParseDateTimeJST(format, value); err == nil {
			return t, nil
		}
	}
	return time.Time{}, errors.New("unable to parse time string:" + value)
}
