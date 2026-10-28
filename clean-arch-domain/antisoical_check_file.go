package clean_arch_domain

import (
	"context"
	"fmt"

	"github.com/samber/lo"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type AntisocialCheckFileRepository interface {
	// Create(ctx context.Context, antisocialCheckFile domain.AntisocialCheckFile) error
	GetAll(ctx context.Context, options ...AntisocialCheckFileQueryOption) (int64, domain.AntisocialCheckFiles, error)
}

type AntisocialCheckFileQueryOption = func(*AntisocialCheckFileQueryOptionFields)

func AntisocialCheckFileQueryOptionFunc() AntisocialCheckFileQueryOptionFields {
	return AntisocialCheckFileQueryOptionFields{}
}

type AntisocialCheckFileQueryOptionFields struct {
	checkId *int64
	QueryOptions
}

func (a AntisocialCheckFileQueryOptionFields) WithCheckId(checkId int64) func(*AntisocialCheckFileQueryOptionFields) {
	return func(fields *AntisocialCheckFileQueryOptionFields) {
		fields.checkId = &checkId
	}
}

func (a AntisocialCheckFileQueryOptionFields) WithSortBy(sortBy string) func(*AntisocialCheckFileQueryOptionFields) {
	return func(fields *AntisocialCheckFileQueryOptionFields) {
		fields.sortBy = &sortBy
	}
}

func (a AntisocialCheckFileQueryOptionFields) WithOrder(order string) func(*AntisocialCheckFileQueryOptionFields) {
	return func(qo *AntisocialCheckFileQueryOptionFields) {
		qo.order = &order
	}
}

func (a AntisocialCheckFileQueryOptionFields) WithOffset(offset int) func(*AntisocialCheckFileQueryOptionFields) {
	return func(fields *AntisocialCheckFileQueryOptionFields) {
		fields.offset = &offset
	}
}

func (a AntisocialCheckFileQueryOptionFields) WithLimit(limit int) func(*AntisocialCheckFileQueryOptionFields) {
	return func(fields *AntisocialCheckFileQueryOptionFields) {
		fields.limit = &limit
	}
}

func (a AntisocialCheckFileQueryOptionFields) CheckId() *string {
	if a.checkId == nil {
		return nil
	}

	return lo.ToPtr(fmt.Sprintf("%v", *a.checkId))
}
