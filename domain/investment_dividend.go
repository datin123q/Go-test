package domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/modelhelper"
)

// 마이페이지 배당 내역 명세
type InvestmentDividendDetail struct {
	Id                    int64                        `json:"id"` // 분배 ID
	UserId                int64                        `json:"user_id"`
	ProjectId             int64                        `json:"project_id"`
	ProjectType           PROJECT_TYPE                 `json:"project_type" swagger:"desc(투자유형)"`
	ProjectName           string                       `json:"project_name" swagger:"desc(상품명)"`
	ProjectStatus         PROJECT_STATUS               `json:"project_status" swagger:"desc(모집상태),enum(STANDBY_FUNDING|OPEN_FUNDING|END_FUNDING|SUCCESS_FUNDING|FAIL_FUNDING|OPERATION_FUNDING|OPERATION_END_FUNDING|CLOSE_FUNDING|INVALID_FUNDING)"`
	TransactionDetailType USER_TRANSACTION_DETAIL_TYPE `json:"transaction_detail_type" swagger:"desc(종류)"`
	TransactionDate       jsonutils.JSONTime           `json:"transaction_date" swagger:"desc(일시)"`
	DividendDate          jsonutils.JSONTime           `json:"dividend_date" swagger:"desc(배당 일시)"`
	Amount                int64                        `json:"amount" swagger:"desc(금액)"`
	Tax                   int64                        `json:"tax" swagger:"desc(원천 징수세)"`
	Deleted               jsonutils.JSONTime           `json:"-" xorm:"deleted"` // 소프트 딜리트 필드 - JSON 마샬링 시 제외
}

func (InvestmentDividendDetail) GetDividendsByUserId(ctx context.Context, userId int64, startDate string, endDate string, detailType, projectId, projectName *string, sortby string, order string, offset int, limit int) (int64, []InvestmentDividendDetail, error) {

	q := factory.DB(ctx).Table("dividend").Select("dividend.*,dividend.id as id,dividend.project_id as project_id,dividend.transaction_type as transaction_detail_type, project.name as project_name, project.project_status as project_status,dividend.project_type as project_type, dividend.tax as tax").
		Join("INNER", "project", "dividend.project_id = project.id").
		Join("INNER", "repayment_dividend_schedule", "dividend.settlement_id = repayment_dividend_schedule.id").
		Where("dividend.user_id = ? AND repayment_dividend_schedule.status =?", userId, RETURN_COMPLETED)

	if err := modelhelper.SearchEqQuery(q, "dividend", "transaction_type", detailType); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SearchEqQuery(q, "dividend", "project_id", projectId); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SearchLikeQuery(q, "project", "name", projectName); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SetBetweenDate(q, "dividend.transaction_date", startDate, endDate); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SetSortOrderForTable(q, "dividend", sortby, order); err != nil {
		factory.Logger(ctx).Error(err)
	}

	var items []InvestmentDividendDetail
	var totalCount int64
	totalCount, err := q.Limit(limit, offset).FindAndCount(&items)
	if err != nil {
		return 0, nil, err
	}
	return totalCount, items, nil
}

// 마이페이지 배당 내역 명세
type ProjectDividendDetail struct {
	Id                 int64          `json:"id"`
	UserId             int64          `json:"user_id" xorm:"<-"`
	ProjectId          int64          `json:"project_id" xorm:"<-"`
	ProjectStatus      PROJECT_STATUS `json:"project_status" xorm:"<-" swagger:"desc(모집상태),enum(STANDBY_FUNDING|OPEN_FUNDING|END_FUNDING|SUCCESS_FUNDING|FAIL_FUNDING|OPERATION_FUNDING|OPERATION_END_FUNDING|CLOSE_FUNDING|INVALID_FUNDING)"`
	ProjectType        PROJECT_TYPE   `json:"project_type" xorm:"<-" swagger:"desc(투자유형),enum(LOAN/EQUITY))" `
	ProjectName        string         `json:"project_name" xorm:"<-" swagger:"desc(상품명)"`
	InvestmentAmount   int64          `json:"investment_amount" xorm:"<-" swagger:"desc(투자금액)"`
	DividendAmount     int64          `json:"dividend_amount" xorm:"<-" swagger:"desc(배당금 합계)"`
	RepaymentAmount    int64          `json:"repayment_amount" xorm:"<-" swagger:"desc(상환금 합계)"`
	TotalAmount        int64          `json:"total_amount" xorm:"<-" swagger:"desc(분배/상환금 전체 합계)"`
	TotalRevenueAmount int64          `json:"total_revenue_amount" xorm:"<-" swagger:"desc(분배/상환금 수익 세후 합계)"`
	TotalTax           int64          `json:"total_tax" xorm:"<-" swagger:"desc(분배/상환금 원천징수 세금 합계)"`
}

func (ProjectDividendDetail) GetByUserProjectByDividends(ctx context.Context, userId int64, sortby, order string, offset, limit int) (int64, []ProjectDividendDetail, error) {

	q := factory.DB(ctx).Table("project_dividend_detail").Where("user_id = ?", userId)

	if err := modelhelper.SetSortOrder(q, sortby, order); err != nil {
		factory.Logger(ctx).Error(err)
	}

	var items []ProjectDividendDetail
	var totalCount int64
	totalCount, err := q.Limit(limit, offset).FindAndCount(&items)
	if err != nil {
		return 0, nil, err
	}
	return totalCount, items, nil
}
