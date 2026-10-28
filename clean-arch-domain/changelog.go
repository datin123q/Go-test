package clean_arch_domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type ChangelogRepository interface {
	GetAll(ctx context.Context, options ...ChangelogQueryOption) ([]domain.ChangelogRow, error)
}

type ChangelogQueryOption = func(*ChangelogQueryOptionFields)

func ChangelogQueryOptionFunc() ChangelogQueryOptionFields {
	return ChangelogQueryOptionFields{}
}

type ChangelogQueryOptionFields struct {
	id *string
	KsqlQueryOptions
}

func (qo ChangelogQueryOptionFields) WithStartDate(startDate *string) func(*ChangelogQueryOptionFields) {
	return func(qo *ChangelogQueryOptionFields) {
		qo.startDate = startDate
	}
}

func (qo ChangelogQueryOptionFields) WithEndDate(endDate *string) func(*ChangelogQueryOptionFields) {
	return func(qo *ChangelogQueryOptionFields) {
		qo.endDate = endDate
	}
}

func (qo ChangelogQueryOptionFields) WithLimit(limit int) func(*ChangelogQueryOptionFields) {
	return func(qo *ChangelogQueryOptionFields) {
		qo.limit = limit
	}
}

func (qo ChangelogQueryOptionFields) WithId(id *string) func(*ChangelogQueryOptionFields) {
	return func(qo *ChangelogQueryOptionFields) {
		qo.id = id
	}
}

func (qo ChangelogQueryOptionFields) StartDate() *string {
	return qo.startDate
}

func (qo ChangelogQueryOptionFields) EndDate() *string {
	return qo.endDate
}

func (qo ChangelogQueryOptionFields) Limit() int {
	return qo.limit
}

func (qo ChangelogQueryOptionFields) Id() *string {
	return qo.id
}
