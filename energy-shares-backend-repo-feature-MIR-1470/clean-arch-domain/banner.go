package clean_arch_domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type BannerRepository interface {
	Create(ctx context.Context, banner domain.Banner) (int64, error)
	GetAll(ctx context.Context, options ...BannerQueryOption) (int64, []domain.Banner, error)
	GetById(ctx context.Context, id int64) (domain.Banner, error)
	Update(ctx context.Context, banner domain.Banner) error
	Delete(ctx context.Context, id int64) error
}

type BannerQueryOption = func(*BannerQueryOptionFields)

func BannerQueryOptionFunc() BannerQueryOptionFields {
	return BannerQueryOptionFields{}
}

type BannerQueryOptionFields struct {
	title            string
	lteStartDate     string
	gteStartDate     string
	lteEndDate       string
	gteEndDate       string
	createdStartDate string
	createdEndDate   string
	isPosted         *bool
	QueryOptions
}

func (qo BannerQueryOptionFields) WithTitle(title string) func(*BannerQueryOptionFields) {
	return func(qo *BannerQueryOptionFields) {
		qo.title = title
	}
}

func (qo BannerQueryOptionFields) WithLTEStartDate(startDate string) func(*BannerQueryOptionFields) {
	return func(qo *BannerQueryOptionFields) {
		qo.lteStartDate = startDate
	}
}

func (qo BannerQueryOptionFields) WithGTEStartDate(startDate string) func(*BannerQueryOptionFields) {
	return func(qo *BannerQueryOptionFields) {
		qo.gteStartDate = startDate
	}
}

func (qo BannerQueryOptionFields) WithLTEEndDate(endDate string) func(*BannerQueryOptionFields) {
	return func(qo *BannerQueryOptionFields) {
		qo.lteEndDate = endDate
	}
}

func (qo BannerQueryOptionFields) WithGTEEndDate(endDate string) func(*BannerQueryOptionFields) {
	return func(qo *BannerQueryOptionFields) {
		qo.gteEndDate = endDate
	}
}

func (qo BannerQueryOptionFields) WithCreatedStartDate(created string) func(*BannerQueryOptionFields) {
	return func(qo *BannerQueryOptionFields) {
		qo.createdStartDate = created
	}
}

func (qo BannerQueryOptionFields) WithCreatedEndDate(created string) func(*BannerQueryOptionFields) {
	return func(qo *BannerQueryOptionFields) {
		qo.createdEndDate = created
	}
}

func (qo BannerQueryOptionFields) WithIsPosted(isPosted *bool) func(*BannerQueryOptionFields) {
	return func(qo *BannerQueryOptionFields) {
		qo.isPosted = isPosted
	}
}
func (qo BannerQueryOptionFields) WithSortBy(sortBy string) func(*BannerQueryOptionFields) {
	return func(qo *BannerQueryOptionFields) {
		qo.sortBy = &sortBy
	}
}

func (qo BannerQueryOptionFields) WithOrder(order string) func(*BannerQueryOptionFields) {
	return func(qo *BannerQueryOptionFields) {
		qo.order = &order
	}
}

func (qo BannerQueryOptionFields) WithLimit(limit int) func(*BannerQueryOptionFields) {
	return func(qo *BannerQueryOptionFields) {
		qo.limit = &limit
	}
}

func (qo BannerQueryOptionFields) WithOffset(offset int) func(*BannerQueryOptionFields) {
	return func(qo *BannerQueryOptionFields) {
		qo.offset = &offset
	}
}

func (qo BannerQueryOptionFields) Title() string {
	return qo.title
}

func (qo BannerQueryOptionFields) LteStartDate() string {
	return qo.lteStartDate
}

func (qo BannerQueryOptionFields) GteStartDate() string {
	return qo.gteStartDate
}

func (qo BannerQueryOptionFields) LteEndDate() string {
	return qo.lteEndDate
}

func (qo BannerQueryOptionFields) GteEndDate() string {
	return qo.gteEndDate
}

func (qo BannerQueryOptionFields) CreatedStartDate() string {
	return qo.createdStartDate
}

func (qo BannerQueryOptionFields) CreatedEndDate() string {
	return qo.createdEndDate
}

func (qo BannerQueryOptionFields) IsPosted() *bool {
	return qo.isPosted
}
