package clean_arch_domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type WithdrawalService interface {
	ExportCSV(ctx context.Context, items []domain.WithdrawalDetail) CSVData
}

type WithdrawalRepository interface {
	Create(ctx context.Context, withdrawal domain.Withdrawal) (int64, error)
	GetAll(ctx context.Context, options ...WithdrawalQueryOption) (int64, []domain.WithdrawalDetail, error)
	GetById(ctx context.Context, id int64) (domain.WithdrawalDetail, error)
	GetByTransactionId(ctx context.Context, transactionId string) (domain.Withdrawal, error)
	Update(ctx context.Context, withdrawal domain.Withdrawal) error
	GetSum(ctx context.Context, options ...WithdrawalQueryOption) (int64, error)
}

type WithdrawalQueryOption = func(*WithdrawalQueryOptionFields)

func WithdrawalQueryOptionFunc() WithdrawalQueryOptionFields {
	return WithdrawalQueryOptionFields{}
}

type WithdrawalQueryOptionFields struct {
	QueryOptions
	investorId        *string
	investorName      *string
	amountMin         *int64
	amountMax         *int64
	requestStartDate  *string
	requestEndDate    *string
	completeStartDate *string
	completeEndDate   *string
	status            *domain.WITHDRAWAL_STATUS_TYPE
	closeId           *int64
	sumColumn         *string
	ids               []int64
}

func (qo WithdrawalQueryOptionFields) WithInvestorId(investorId *string) WithdrawalQueryOption {
	return func(qo *WithdrawalQueryOptionFields) {
		qo.investorId = investorId
	}
}

func (qo WithdrawalQueryOptionFields) WithInvestorName(investorName *string) WithdrawalQueryOption {
	return func(qo *WithdrawalQueryOptionFields) {
		qo.investorName = investorName
	}
}

func (qo WithdrawalQueryOptionFields) WithAmountMin(amountMin *int64) WithdrawalQueryOption {
	return func(qo *WithdrawalQueryOptionFields) {
		qo.amountMin = amountMin
	}
}

func (qo WithdrawalQueryOptionFields) WithAmountMax(amountMax *int64) WithdrawalQueryOption {
	return func(qo *WithdrawalQueryOptionFields) {
		qo.amountMax = amountMax
	}
}

func (qo WithdrawalQueryOptionFields) WithRequestStartDate(requestStartDate *string) WithdrawalQueryOption {
	return func(qo *WithdrawalQueryOptionFields) {
		qo.requestStartDate = requestStartDate
	}
}

func (qo WithdrawalQueryOptionFields) WithRequestEndDate(requestEndDate *string) WithdrawalQueryOption {
	return func(qo *WithdrawalQueryOptionFields) {
		qo.requestEndDate = requestEndDate
	}
}

func (qo WithdrawalQueryOptionFields) WithCompleteStartDate(completeStartDate *string) WithdrawalQueryOption {
	return func(qo *WithdrawalQueryOptionFields) {
		qo.completeStartDate = completeStartDate
	}
}

func (qo WithdrawalQueryOptionFields) WithCompleteEndDate(completeEndDate *string) WithdrawalQueryOption {
	return func(qo *WithdrawalQueryOptionFields) {
		qo.completeEndDate = completeEndDate
	}
}

func (qo WithdrawalQueryOptionFields) WithStatus(status *domain.WITHDRAWAL_STATUS_TYPE) func(*WithdrawalQueryOptionFields) {

	return func(qo *WithdrawalQueryOptionFields) {
		qo.status = status
	}
}

func (qo WithdrawalQueryOptionFields) WithCloseId(closeId *int64) WithdrawalQueryOption {
	return func(qo *WithdrawalQueryOptionFields) {
		qo.closeId = closeId
	}
}

func (qo WithdrawalQueryOptionFields) WithSumColumn(sumColumn string) WithdrawalQueryOption {
	return func(qo *WithdrawalQueryOptionFields) {
		qo.sumColumn = &sumColumn
	}
}

func (qo WithdrawalQueryOptionFields) WithIds(ids []int64) WithdrawalQueryOption {
	return func(qo *WithdrawalQueryOptionFields) {
		qo.ids = ids
	}
}

func (qo WithdrawalQueryOptionFields) WithSortBy(sortBy string) WithdrawalQueryOption {
	return func(qo *WithdrawalQueryOptionFields) {
		qo.sortBy = &sortBy
	}
}

func (qo WithdrawalQueryOptionFields) WithOrder(order string) WithdrawalQueryOption {
	return func(qo *WithdrawalQueryOptionFields) {
		qo.order = &order
	}
}

func (qo WithdrawalQueryOptionFields) WithOffset(offset int) WithdrawalQueryOption {
	return func(qo *WithdrawalQueryOptionFields) {
		qo.offset = &offset
	}
}

func (qo WithdrawalQueryOptionFields) WithLimit(limit int) WithdrawalQueryOption {
	return func(qo *WithdrawalQueryOptionFields) {
		qo.limit = &limit
	}
}

func (qo WithdrawalQueryOptionFields) InvestorId() *string {
	return qo.investorId
}

func (qo WithdrawalQueryOptionFields) InvestorName() *string {
	return qo.investorName
}

func (qo WithdrawalQueryOptionFields) AmountMin() *int64 {
	return qo.amountMin
}

func (qo WithdrawalQueryOptionFields) AmountMax() *int64 {
	return qo.amountMax
}

func (qo WithdrawalQueryOptionFields) RequestStartDate() *string {
	return qo.requestStartDate
}

func (qo WithdrawalQueryOptionFields) RequestEndDate() *string {
	return qo.requestEndDate
}

func (qo WithdrawalQueryOptionFields) CompleteStartDate() *string {
	return qo.completeStartDate
}

func (qo WithdrawalQueryOptionFields) CompleteEndDate() *string {
	return qo.completeEndDate
}

func (qo WithdrawalQueryOptionFields) Status() *domain.WITHDRAWAL_STATUS_TYPE {
	return qo.status
}

func (qo WithdrawalQueryOptionFields) CloseId() *int64 {
	return qo.closeId
}

func (qo WithdrawalQueryOptionFields) SumColumn() *string {
	return qo.sumColumn
}

func (qo WithdrawalQueryOptionFields) Ids() []int64 {
	return qo.ids
}
