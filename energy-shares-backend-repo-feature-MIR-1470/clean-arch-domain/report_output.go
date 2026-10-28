package clean_arch_domain

import (
	"context"

	"github.com/samber/lo"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type ReportOutputRepository interface {
	GetAll(ctx context.Context, options ...ReportOutputQueryOption) (int64, []domain.ReportOutput, error)
	GetById(ctx context.Context, id int64) (domain.ReportOutput, error)
	Update(ctx context.Context, reportOutput domain.ReportOutput) error
}

type ReportOutputQueryOption = func(*ReportOutputQueryOptionFields)

func ReportOutputQueryOptionFunc() ReportOutputQueryOptionFields {
	return ReportOutputQueryOptionFields{}
}

type ReportOutputQueryOptionFields struct {
	submissionType *string
	reportTypes    []string
	groupBy        *string
	year           *string
	quarter        *string
	QueryOptions
}

func (qo ReportOutputQueryOptionFields) WithSubmissionType(submissionType domain.SUBMISSION_TYPE) func(*ReportOutputQueryOptionFields) {
	return func(qo *ReportOutputQueryOptionFields) {
		qo.submissionType = lo.ToPtr(string(submissionType))
	}
}

func (qo ReportOutputQueryOptionFields) SubmissionType() *string {
	return qo.submissionType
}

func (qo ReportOutputQueryOptionFields) WithReportTypes(reportTypes ...domain.REPORT_TYPE) func(*ReportOutputQueryOptionFields) {
	return func(qo *ReportOutputQueryOptionFields) {
		for _, reportType := range reportTypes {
			qo.reportTypes = append(qo.reportTypes, string(reportType))
		}
	}
}

func (qo ReportOutputQueryOptionFields) ReportTypes() []interface{} {
	interfaceSlice := make([]interface{}, len(qo.reportTypes))
	for i, v := range qo.reportTypes {
		interfaceSlice[i] = v
	}
	return interfaceSlice
}

func (qo ReportOutputQueryOptionFields) WithGroupBy(groupBy string) func(*ReportOutputQueryOptionFields) {
	return func(qo *ReportOutputQueryOptionFields) {
		qo.groupBy = &groupBy
	}
}

func (qo ReportOutputQueryOptionFields) GroupBy() *string {
	return qo.groupBy
}

func (qo ReportOutputQueryOptionFields) WithYear(year string) func(*ReportOutputQueryOptionFields) {
	return func(qo *ReportOutputQueryOptionFields) {
		qo.year = &year
	}
}

func (qo ReportOutputQueryOptionFields) Year() *string {
	return qo.year
}

func (qo ReportOutputQueryOptionFields) WithQuarter(quarter string) func(*ReportOutputQueryOptionFields) {
	return func(qo *ReportOutputQueryOptionFields) {
		qo.quarter = &quarter
	}
}

func (qo ReportOutputQueryOptionFields) Quarter() *string {
	return qo.quarter
}

func (qo ReportOutputQueryOptionFields) WithSortBy(sortBy string) func(*ReportOutputQueryOptionFields) {
	return func(qo *ReportOutputQueryOptionFields) {
		qo.sortBy = &sortBy
	}
}

func (qo ReportOutputQueryOptionFields) WithOrder(order string) func(*ReportOutputQueryOptionFields) {
	return func(qo *ReportOutputQueryOptionFields) {
		qo.order = &order
	}
}

func (qo ReportOutputQueryOptionFields) WithOffset(offset int) func(*ReportOutputQueryOptionFields) {
	return func(qo *ReportOutputQueryOptionFields) {
		qo.offset = &offset
	}
}

func (qo ReportOutputQueryOptionFields) WithLimit(limit int) func(*ReportOutputQueryOptionFields) {
	return func(qo *ReportOutputQueryOptionFields) {
		qo.limit = &limit
	}
}
