package clean_arch_domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type OperatorHistoryRepository interface {
	GetAll(ctx context.Context, options ...OperatorHistoryQueryOption) (int64, []domain.OperatorHistory, error)
	Create(ctx context.Context, history domain.OperatorHistory) error
}

type OperatorHistoryQueryOption = func(*OperatorHistoryQueryOptionFields)

func OperatorHistoryQueryOptionFunc() OperatorHistoryQueryOptionFields {
	return OperatorHistoryQueryOptionFields{}
}

type OperatorHistoryQueryOptionFields struct {
	name      *string
	menu      *string
	state     *string
	startDate *string
	endDate   *string
	QueryOptions
}

func (qo OperatorHistoryQueryOptionFields) WithName(name *string) func(*OperatorHistoryQueryOptionFields) {
	return func(qo *OperatorHistoryQueryOptionFields) {
		qo.name = name
	}
}

func (qo OperatorHistoryQueryOptionFields) WithMenu(menu *string) func(*OperatorHistoryQueryOptionFields) {
	return func(qo *OperatorHistoryQueryOptionFields) {
		qo.menu = menu
	}
}

func (qo OperatorHistoryQueryOptionFields) WithState(state *string) func(*OperatorHistoryQueryOptionFields) {
	return func(qo *OperatorHistoryQueryOptionFields) {
		qo.state = state
	}
}

func (qo OperatorHistoryQueryOptionFields) WithStartDate(startDate *string) func(*OperatorHistoryQueryOptionFields) {
	return func(qo *OperatorHistoryQueryOptionFields) {
		qo.startDate = startDate
	}
}

func (qo OperatorHistoryQueryOptionFields) WithEndDate(endDate *string) func(*OperatorHistoryQueryOptionFields) {
	return func(qo *OperatorHistoryQueryOptionFields) {
		qo.endDate = endDate
	}
}

func (qo OperatorHistoryQueryOptionFields) WithSortBy(sortBy string) func(*OperatorHistoryQueryOptionFields) {
	return func(qo *OperatorHistoryQueryOptionFields) {
		qo.sortBy = &sortBy
	}
}

func (qo OperatorHistoryQueryOptionFields) WithOrder(order string) func(*OperatorHistoryQueryOptionFields) {
	return func(qo *OperatorHistoryQueryOptionFields) {
		qo.order = &order
	}
}

func (qo OperatorHistoryQueryOptionFields) WithOffset(offset int) func(*OperatorHistoryQueryOptionFields) {
	return func(qo *OperatorHistoryQueryOptionFields) {
		qo.offset = &offset
	}
}

func (qo OperatorHistoryQueryOptionFields) WithLimit(limit int) func(*OperatorHistoryQueryOptionFields) {
	return func(qo *OperatorHistoryQueryOptionFields) {
		qo.limit = &limit
	}
}

func (qo OperatorHistoryQueryOptionFields) Name() *string {
	return qo.name
}

func (qo OperatorHistoryQueryOptionFields) Menu() *string {
	return qo.menu
}

func (qo OperatorHistoryQueryOptionFields) State() *string {
	return qo.state
}

func (qo OperatorHistoryQueryOptionFields) StartDate() *string {
	return qo.startDate
}

func (qo OperatorHistoryQueryOptionFields) EndDate() *string {
	return qo.endDate
}

type OperatorHistoryKsqlRepository interface {
	GetAll(ctx context.Context, options ...OperatorHistoryKsqlQueryOption) (int64, []domain.OperatorHistoryRowResponse, error)
}

type OperatorHistoryKsqlQueryOption = func(*OperatorHistoryKsqlQueryOptionFields)

func OperatorHistoryKsqlQueryOptionFunc() OperatorHistoryKsqlQueryOptionFields {
	return OperatorHistoryKsqlQueryOptionFields{}
}

type OperatorHistoryKsqlQueryOptionFields struct {
	KsqlQueryOptions
}

func (qo OperatorHistoryKsqlQueryOptionFields) WithStartDate(startDate *string) func(*OperatorHistoryKsqlQueryOptionFields) {
	return func(qo *OperatorHistoryKsqlQueryOptionFields) {
		qo.startDate = startDate
	}
}

func (qo OperatorHistoryKsqlQueryOptionFields) WithEndDate(endDate *string) func(*OperatorHistoryKsqlQueryOptionFields) {
	return func(qo *OperatorHistoryKsqlQueryOptionFields) {
		qo.endDate = endDate
	}
}

func (qo OperatorHistoryKsqlQueryOptionFields) WithLimit(limit int) func(*OperatorHistoryKsqlQueryOptionFields) {
	return func(qo *OperatorHistoryKsqlQueryOptionFields) {
		qo.limit = limit
	}
}

func (qo OperatorHistoryKsqlQueryOptionFields) StartDate() *string {
	return qo.startDate
}

func (qo OperatorHistoryKsqlQueryOptionFields) EndDate() *string {
	return qo.endDate
}

func (qo OperatorHistoryKsqlQueryOptionFields) Limit() int {
	return qo.limit
}
