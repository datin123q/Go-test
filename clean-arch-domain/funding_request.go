package clean_arch_domain

import (
	"context"
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/samber/lo"
)

type FundingRequestRepository interface {
	Create(ctx context.Context, fundingRequest domain.FundingRequest) (int64, error)
	Update(ctx context.Context, fundingRequest domain.FundingRequest) error
	GetAll(ctx context.Context, options ...FundingRequestQueryOption) (int64, []domain.FundingRequest, error)
	GetById(ctx context.Context, id int64) (domain.FundingRequest, error)
}

type FundingRequestQueryOption = func(*FundingRequestQueryOptionFields)

func FundingRequestQueryOptionFunc() FundingRequestQueryOptionFields {
	return FundingRequestQueryOptionFields{}
}

type FundingRequestQueryOptionFields struct {
	id            int64
	email         *string
	phone         *string
	requestStatus *domain.REQUEST_STATUS
	startDate     *string
	endDate       *string
	QueryOptions
}

func (qo FundingRequestQueryOptionFields) WithId(id int64) func(*FundingRequestQueryOptionFields) {
	return func(qo *FundingRequestQueryOptionFields) {
		qo.id = id
	}
}

func (qo FundingRequestQueryOptionFields) WithEmail(email *string) func(*FundingRequestQueryOptionFields) {
	return func(qo *FundingRequestQueryOptionFields) {
		qo.email = email
	}
}

func (qo FundingRequestQueryOptionFields) WithPhone(phone *string) func(*FundingRequestQueryOptionFields) {
	return func(qo *FundingRequestQueryOptionFields) {
		qo.phone = phone
	}
}

func (qo FundingRequestQueryOptionFields) WithRequestStatus(requestStatus *domain.REQUEST_STATUS) func(*FundingRequestQueryOptionFields) {
	return func(qo *FundingRequestQueryOptionFields) {
		qo.requestStatus = requestStatus
	}
}

func (qo FundingRequestQueryOptionFields) WithStartDate(startDate *string) func(*FundingRequestQueryOptionFields) {
	return func(qo *FundingRequestQueryOptionFields) {
		qo.startDate = startDate
	}
}

func (qo FundingRequestQueryOptionFields) WithEndDate(endDate *string) func(*FundingRequestQueryOptionFields) {
	return func(qo *FundingRequestQueryOptionFields) {
		qo.endDate = endDate
	}
}

func (qo FundingRequestQueryOptionFields) WithSortBy(sortBy string) func(*FundingRequestQueryOptionFields) {
	return func(qo *FundingRequestQueryOptionFields) {
		qo.sortBy = &sortBy
	}
}

func (qo FundingRequestQueryOptionFields) WithOrder(order string) func(*FundingRequestQueryOptionFields) {
	return func(qo *FundingRequestQueryOptionFields) {
		qo.order = &order
	}
}

func (qo FundingRequestQueryOptionFields) WithOffset(offset int) func(*FundingRequestQueryOptionFields) {
	return func(qo *FundingRequestQueryOptionFields) {
		qo.offset = &offset
	}
}

func (qo FundingRequestQueryOptionFields) WithLimit(limit int) func(*FundingRequestQueryOptionFields) {
	return func(qo *FundingRequestQueryOptionFields) {
		qo.limit = &limit
	}
}

func (qo FundingRequestQueryOptionFields) Id() int64 {
	return qo.id
}

func (qo FundingRequestQueryOptionFields) Email() *string {
	return qo.email
}

func (qo FundingRequestQueryOptionFields) Phone() *string {
	return qo.phone
}

func (qo FundingRequestQueryOptionFields) RequestStatus() *domain.REQUEST_STATUS {
	return qo.requestStatus
}

func (qo FundingRequestQueryOptionFields) StartDate() time.Time {
	if qo.endDate == nil {
		return time.Time{}
	}

	return lo.Must(parseMultiFormat(*qo.endDate))
}

func (qo FundingRequestQueryOptionFields) EndDate() time.Time {
	if qo.endDate == nil {
		return time.Time{}
	}

	return lo.Must(parseMultiFormat(*qo.endDate))
}
