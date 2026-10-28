package clean_arch_domain

import (
	"context"
	"fmt"

	"github.com/samber/lo"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type ShareRepository interface {
	GetById(ctx context.Context, id string) (domain.Share, error)
	LockById(ctx context.Context, id string) (domain.Share, error)
	Update(ctx context.Context, share domain.Share) error // RequestInvest, RefundInvest, OperationInvest
	GetInvestor(ctx context.Context, options ...ShareQueryOption) ([]domain.User, error)
	GetInvestorCount(ctx context.Context, projectId int64, options ...ShareQueryOption) (int64, error)
	GetSum(ctx context.Context, options ...ShareQueryOption) (float64, error)
	GetAll(ctx context.Context, options ...ShareQueryOption) ([]domain.Share, error)
}

type ShareService interface {
	RequestInvest(ctx context.Context, sharedId string, userId int64) error
	SuccessInvest(ctx context.Context, sharedId string, userId int64) error
	CancelInvest(ctx context.Context, sharedId string, userId int64) error
	FailInvest(ctx context.Context, shareId string, userId int64) error
}

type ShareQueryOption = func(*ShareQueryOptionFields)

// ShareQueryOptionFunc Query Option
func ShareQueryOptionFunc() ShareQueryOptionFields {
	return ShareQueryOptionFields{}
}

type ShareQueryOptionFields struct {
	investStatuses []string
	projectId      *int64
	userId         *int64
	sumColumn      *string
	isNewsLetter   *bool
	groupBy        *string
	QueryOptions
}

func (pqo ShareQueryOptionFields) WithInvestStatuses(investStatuses ...domain.INVEST_STATUS) ShareQueryOption {
	return func(pqo *ShareQueryOptionFields) {
		for _, investStatus := range investStatuses {
			pqo.investStatuses = append(pqo.investStatuses, string(investStatus))
		}
	}
}

func (pqo ShareQueryOptionFields) InvestStatuses() []interface{} {
	interfaceSlice := make([]interface{}, len(pqo.investStatuses))
	for i, v := range pqo.investStatuses {
		interfaceSlice[i] = v
	}
	return interfaceSlice
}

func (pqo ShareQueryOptionFields) WithProjectId(projectId int64) ShareQueryOption {
	return func(pqo *ShareQueryOptionFields) {
		pqo.projectId = &projectId
	}
}

func (pqo ShareQueryOptionFields) ProjectId() *string {
	if pqo.projectId == nil {
		return nil
	}

	return lo.ToPtr(fmt.Sprintf("%v", *pqo.projectId))
}

func (pqo ShareQueryOptionFields) WithUserId(userId int64) ShareQueryOption {
	return func(pqo *ShareQueryOptionFields) {
		pqo.userId = &userId
	}
}

func (pqo ShareQueryOptionFields) UserId() *string {
	if pqo.userId == nil {
		return nil
	}

	return lo.ToPtr(fmt.Sprintf("%v", *pqo.userId))
}

func (pqo ShareQueryOptionFields) WithSumColumn(sumColumn string) ShareQueryOption {
	return func(pqo *ShareQueryOptionFields) {
		pqo.sumColumn = &sumColumn
	}
}

func (pqo ShareQueryOptionFields) SumColumn() *string {
	if pqo.sumColumn == nil {
		return nil
	}

	return lo.ToPtr(fmt.Sprintf("%v", *pqo.sumColumn))
}

func (pqo ShareQueryOptionFields) WithIsNewsLetter(isNewsLetter bool) ShareQueryOption {
	return func(pqo *ShareQueryOptionFields) {
		pqo.isNewsLetter = &isNewsLetter
	}
}

func (pqo ShareQueryOptionFields) IsNewsLetter() *bool {
	return pqo.isNewsLetter
}

func (pqo ShareQueryOptionFields) WithGroupBy(groupBy string) ShareQueryOption {
	return func(pqo *ShareQueryOptionFields) {
		pqo.groupBy = &groupBy
	}
}

func (pqo ShareQueryOptionFields) GroupBy() *string {
	return pqo.groupBy
}

func (pqo ShareQueryOptionFields) WithSortBy(sortBy string) ShareQueryOption {
	return func(pqo *ShareQueryOptionFields) {
		pqo.sortBy = &sortBy
	}
}

func (pqo ShareQueryOptionFields) WithOrder(order string) ShareQueryOption {
	return func(pqo *ShareQueryOptionFields) {
		pqo.order = &order
	}
}

func (pqo ShareQueryOptionFields) WithOffset(offset int) ShareQueryOption {
	return func(pqo *ShareQueryOptionFields) {
		pqo.offset = &offset
	}
}

func (pqo ShareQueryOptionFields) WithLimit(limit int) ShareQueryOption {
	return func(pqo *ShareQueryOptionFields) {
		pqo.limit = &limit
	}
}
