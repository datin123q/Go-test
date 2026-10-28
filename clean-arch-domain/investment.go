package clean_arch_domain

import (
	"context"
	"fmt"

	"github.com/samber/lo"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type InvestmentRepository interface {
	GetAll(ctx context.Context, options ...InvestmentQueryOption) (int64, []domain.Investment, error)
	GroupByOwner(ctx context.Context, options ...InvestmentQueryOption) ([]domain.Investment, error)
	GetById(ctx context.Context, id string) (domain.Investment, error)
}

type InvestmentService interface {
	GetDetail(ctx context.Context, investmentId string) (domain.InvestmentDetail, error)
	ExportCSV(items []domain.Investment) CSVData
}

type InvestmentQueryOption = func(*InvestmentQueryOptionFields)

// InvestmentQueryOptionFunc Query Option
func InvestmentQueryOptionFunc() InvestmentQueryOptionFields {
	return InvestmentQueryOptionFields{}
}

type InvestmentQueryOptionFields struct {
	projectId    *int64
	investStatus *string
	investorName *string
	spvName      *string
	projectName  *string
	projectType  *string
	startAmount  *int
	endAmount    *int
	QueryOptions
}

func (iqo InvestmentQueryOptionFields) WithInvestStatus(investStatus domain.INVEST_STATUS) func(*InvestmentQueryOptionFields) {
	return func(iqo *InvestmentQueryOptionFields) {
		if investStatus != "" {
			iqo.investStatus = lo.ToPtr(string(investStatus))
		}
	}
}

func (iqo InvestmentQueryOptionFields) WithInvestorName(investorName *string) func(*InvestmentQueryOptionFields) {
	return func(iqo *InvestmentQueryOptionFields) {
		iqo.investorName = investorName
	}
}

func (iqo InvestmentQueryOptionFields) WithSpvName(spvName *string) func(*InvestmentQueryOptionFields) {
	return func(iqo *InvestmentQueryOptionFields) {
		iqo.spvName = spvName
	}
}

func (iqo InvestmentQueryOptionFields) WithProjectName(projectName *string) func(*InvestmentQueryOptionFields) {
	return func(iqo *InvestmentQueryOptionFields) {
		iqo.projectName = projectName
	}
}

func (iqo InvestmentQueryOptionFields) WithProjectType(projectType *string) func(*InvestmentQueryOptionFields) {
	return func(iqo *InvestmentQueryOptionFields) {
		iqo.projectType = projectType
	}
}

func (iqo InvestmentQueryOptionFields) WithStartAmount(startAmount *int) func(*InvestmentQueryOptionFields) {
	return func(iqo *InvestmentQueryOptionFields) {
		iqo.startAmount = startAmount
	}
}

func (iqo InvestmentQueryOptionFields) WithEndAmount(endAmount *int) func(*InvestmentQueryOptionFields) {
	return func(iqo *InvestmentQueryOptionFields) {
		iqo.endAmount = endAmount
	}
}

func (iqo InvestmentQueryOptionFields) WithSortBy(sortBy string) func(*InvestmentQueryOptionFields) {
	return func(iqo *InvestmentQueryOptionFields) {
		iqo.sortBy = &sortBy
	}
}

func (iqo InvestmentQueryOptionFields) WithOrder(order string) func(*InvestmentQueryOptionFields) {
	return func(iqo *InvestmentQueryOptionFields) {
		iqo.order = &order
	}
}

func (iqo InvestmentQueryOptionFields) WithOffset(offset int) func(*InvestmentQueryOptionFields) {
	return func(iqo *InvestmentQueryOptionFields) {
		iqo.offset = &offset
	}
}

func (iqo InvestmentQueryOptionFields) WithLimit(limit int) func(*InvestmentQueryOptionFields) {
	return func(iqo *InvestmentQueryOptionFields) {
		iqo.limit = &limit
	}
}

func (iqo InvestmentQueryOptionFields) InvestStatus() *string {
	return iqo.investStatus
}

func (iqo InvestmentQueryOptionFields) InvestorName() *string {
	return iqo.investorName
}

func (iqo InvestmentQueryOptionFields) SpvName() *string {
	return iqo.spvName
}

func (iqo InvestmentQueryOptionFields) ProjectName() *string {
	return iqo.projectName
}

func (iqo InvestmentQueryOptionFields) ProjectType() *string {
	return iqo.projectType
}

func (iqo InvestmentQueryOptionFields) StartAmount() *int {
	return iqo.startAmount
}

func (iqo InvestmentQueryOptionFields) EndAmount() *int {
	return iqo.endAmount
}

func (iqo InvestmentQueryOptionFields) ProjectId() *string {
	if iqo.projectId == nil {
		return nil
	}

	return lo.ToPtr(fmt.Sprintf("%d", *iqo.projectId))
}

func (iqo InvestmentQueryOptionFields) WithProjectId(projectId int64) func(*InvestmentQueryOptionFields) {
	return func(iqo *InvestmentQueryOptionFields) {
		iqo.projectId = &projectId
	}
}
