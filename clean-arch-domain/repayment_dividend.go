package clean_arch_domain

import (
	"context"
	"fmt"
	"time"

	"github.com/samber/lo"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type RepaymentDividendScheduleRepository interface {
	Create(ctx context.Context, repaymentDividendSchedule domain.RepaymentDividendSchedule) (domain.RepaymentDividendSchedule, error)
	Update(ctx context.Context, repaymentDividendSchedule domain.RepaymentDividendSchedule) error
	GetAll(ctx context.Context, opts ...RepaymentDividendScheduleQueryOption) (int64, domain.RepaymentDividendSchedules, error)
	GetById(ctx context.Context, id int64) (domain.RepaymentDividendSchedule, error)
	Delete(ctx context.Context, id int64) error
	GetSum(ctx context.Context, opts ...RepaymentDividendScheduleQueryOption) (float64, error)
	GetSums(ctx context.Context, opts ...RepaymentDividendScheduleQueryOption) ([]float64, error)
}

type RepaymentDividendScheduleQueryOption = func(*RepaymentDividendScheduleQueryOptionFields)

type RepaymentDividendScheduleQueryOptionFields struct {
	QueryOptions
	projectId                *int64
	returnType               *string
	beforeOrEqualExecuteDate time.Time
	afterOrEqualExecuteDate  time.Time
	sumColumn                *string
	sumColumns               *[]string
	status                   *domain.INVESTMENT_RETURN_STATUS
}

func RepaymentDividendQueryOptionFunc() RepaymentDividendScheduleQueryOptionFields {
	return RepaymentDividendScheduleQueryOptionFields{}
}

func (qo RepaymentDividendScheduleQueryOptionFields) WithSortBy(sortBy string) func(*RepaymentDividendScheduleQueryOptionFields) {
	return func(qo *RepaymentDividendScheduleQueryOptionFields) {
		qo.sortBy = &sortBy
	}
}

func (qo RepaymentDividendScheduleQueryOptionFields) WithOrder(order string) func(*RepaymentDividendScheduleQueryOptionFields) {
	return func(qo *RepaymentDividendScheduleQueryOptionFields) {
		qo.order = &order
	}
}

func (qo RepaymentDividendScheduleQueryOptionFields) WithOffset(offset int) func(*RepaymentDividendScheduleQueryOptionFields) {
	return func(qo *RepaymentDividendScheduleQueryOptionFields) {
		qo.offset = &offset
	}
}

func (qo RepaymentDividendScheduleQueryOptionFields) WithLimit(limit int) func(*RepaymentDividendScheduleQueryOptionFields) {
	return func(qo *RepaymentDividendScheduleQueryOptionFields) {
		qo.limit = &limit
	}
}

func (qo RepaymentDividendScheduleQueryOptionFields) WithProjectId(projectId int64) func(*RepaymentDividendScheduleQueryOptionFields) {
	return func(qo *RepaymentDividendScheduleQueryOptionFields) {
		qo.projectId = &projectId
	}
}

func (qo RepaymentDividendScheduleQueryOptionFields) WithStatus(status domain.INVESTMENT_RETURN_STATUS) func(*RepaymentDividendScheduleQueryOptionFields) {
	return func(qo *RepaymentDividendScheduleQueryOptionFields) {
		qo.status = &status
	}
}

func (qo RepaymentDividendScheduleQueryOptionFields) WithReturnType(returnType domain.INVESTMENT_RETURN_TYPE) func(*RepaymentDividendScheduleQueryOptionFields) {
	return func(qo *RepaymentDividendScheduleQueryOptionFields) {
		qo.returnType = lo.ToPtr(string(returnType))
	}
}

func (qo RepaymentDividendScheduleQueryOptionFields) WithBeforeOrEqualExecuteDate(executeDate time.Time) func(*RepaymentDividendScheduleQueryOptionFields) {
	return func(qo *RepaymentDividendScheduleQueryOptionFields) {
		qo.beforeOrEqualExecuteDate = executeDate
	}
}

func (qo RepaymentDividendScheduleQueryOptionFields) WithAfterOrEqualExecuteDate(executeDate time.Time) func(*RepaymentDividendScheduleQueryOptionFields) {
	return func(qo *RepaymentDividendScheduleQueryOptionFields) {
		qo.afterOrEqualExecuteDate = executeDate
	}
}

func (qo RepaymentDividendScheduleQueryOptionFields) WithSumColumn(sumColumn string) func(*RepaymentDividendScheduleQueryOptionFields) {
	return func(qo *RepaymentDividendScheduleQueryOptionFields) {
		qo.sumColumn = &sumColumn
	}
}

func (qo RepaymentDividendScheduleQueryOptionFields) WithSumColumns(sumColumns []string) func(*RepaymentDividendScheduleQueryOptionFields) {
	return func(qo *RepaymentDividendScheduleQueryOptionFields) {
		qo.sumColumns = &sumColumns
	}
}

func (qo RepaymentDividendScheduleQueryOptionFields) ProjectId() *string {
	if qo.projectId == nil {
		return nil
	}
	return lo.ToPtr(fmt.Sprintf("%v", *qo.projectId))
}

func (qo RepaymentDividendScheduleQueryOptionFields) Status() *string {
	if qo.status == nil || *qo.status == "" {
		return nil
	}
	return lo.ToPtr(fmt.Sprintf("%v", *qo.status))
}

func (qo RepaymentDividendScheduleQueryOptionFields) ReturnType() *string {
	if qo.returnType == nil {
		return nil
	}
	return qo.returnType
}

func (qo RepaymentDividendScheduleQueryOptionFields) BeforeOrEqualExecuteDate() time.Time {
	return qo.beforeOrEqualExecuteDate
}

func (qo RepaymentDividendScheduleQueryOptionFields) AfterOrEqualExecuteDate() time.Time {
	return qo.afterOrEqualExecuteDate
}

func (qo RepaymentDividendScheduleQueryOptionFields) SumColumn() *string {
	if qo.sumColumn == nil {
		return nil
	}
	return qo.sumColumn
}

func (qo RepaymentDividendScheduleQueryOptionFields) SumColumns() *[]string {
	if qo.sumColumns == nil {
		return nil
	}
	return qo.sumColumns
}
