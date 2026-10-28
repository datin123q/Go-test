package domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/commoncode"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/modelhelper"
)

const WITHHOLDING_TAX = 0.2042

type INVESTMENT_RETURN_TYPE string

const (
	INVESTMENT_RETURN_TYPE_CODE                        = "INVESTMENT_RETURN_TYPE"
	INVESTMENT_EXECUTE          INVESTMENT_RETURN_TYPE = "INVESTMENT_EXECUTE"
	DIVIDEND_RETURN             INVESTMENT_RETURN_TYPE = "DIVIDEND_RETURN"
	REPAYMENT_RETURN            INVESTMENT_RETURN_TYPE = "REPAYMENT_RETURN"
)

func (r RepaymentDividendSchedule) GetInvestmentReturnTypeCodeGroup(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(INVESTMENT_EXECUTE), Name: i18nutils.Translate("commoncode.INVESTMENT_RETURN_TYPE_CODE.INVESTMENT_EXECUTE", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(DIVIDEND_RETURN), Name: i18nutils.Translate("commoncode.INVESTMENT_RETURN_TYPE_CODE.DIVIDEND_RETURN", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(REPAYMENT_RETURN), Name: i18nutils.Translate("commoncode.INVESTMENT_RETURN_TYPE_CODE.REPAYMENT_RETURN", loc), Order: len(codes) + 1})
	return commoncode.CodeGroup{
		GroupCode: INVESTMENT_RETURN_TYPE_CODE,
		Codes:     codes,
	}
}

type INVESTMENT_RETURN_STATUS string

const (
	INVESTMENT_RETURN_STATUS_CODE                          = "INVESTMENT_RETURN_STATUS"
	RETURN_SCHEDULED              INVESTMENT_RETURN_STATUS = "RETURN_SCHEDULED" // 예정
	RETURN_COMPLETED              INVESTMENT_RETURN_STATUS = "RETURN_COMPLETED" // 완료
)

func (r RepaymentDividendSchedule) GetInvestmentReturnStatusCodeGroup(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(RETURN_SCHEDULED), Name: i18nutils.Translate("commoncode.INVESTMENT_RETURN_STATUS_CODE.RETURN_SCHEDULED", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(RETURN_COMPLETED), Name: i18nutils.Translate("commoncode.INVESTMENT_RETURN_STATUS_CODE.RETURN_COMPLETED", loc), Order: len(codes) + 1})
	return commoncode.CodeGroup{
		GroupCode: INVESTMENT_RETURN_STATUS_CODE,
		Codes:     codes,
	}
}

// 출자배당상환 클린 아키텍처 전환시 함께 제거될 구조체
type SearchInput struct {
	Sortby         string `query:"sortby"`
	Order          string `query:"order"`
	SkipCount      int    `query:"skip_count"`
	MaxResultCount int    `query:"max_result_count"`
}

// 배당상환 리뉴얼 - https://illuminarean.atlassian.net/browse/ESJP-365
type RepaymentDividendSchedule struct {
	// 기본정보
	Id                 int64        `json:"id" xorm:"id autoincr pk"`                         // 고유 식별자
	ProjectId          int64        `json:"project_id" xorm:"project_id"`                     // 프로젝트 ID
	ProjectType        PROJECT_TYPE `json:"project_type" xorm:"project_type"`                 // 프로젝트 유형 (사업형, 대출형 - 현재는 사업형만 있음)
	ProjectShareQty    int64        `json:"project_share_qty" xorm:"project_share_qty"`       // 전체 구좌수
	ProjectShareAmount int64        `json:"project_share_amount" xorm:"project_share_amount"` // 전체 투자액

	// 공통정보
	DecideDate  jsonutils.JSONTime       `json:"decide_date" xorm:"decide_date"`   // 배당, 상환을 결정한 일시
	ExecuteDate jsonutils.JSONTime       `json:"execute_date" xorm:"execute_date"` // 배당, 상환을 실시할 일시
	ReturnType  INVESTMENT_RETURN_TYPE   `json:"return_type" xorm:"return_type"`   // 배당 또는 상환 유형 DIVIDEND_RETURN, REPAYMENT_RETURN
	Status      INVESTMENT_RETURN_STATUS `json:"status" xorm:"status"`             // "SCHEDULED", "COMPLETED"

	// 상환정보
	RepaymentAmount          int64 `json:"repayment_amount" xorm:"repayment_amount"`                     // 상환액
	RepaymentPerShare        int64 `json:"repayment_per_share" xorm:"repayment_per_share"`               // 구좌당 상환액
	PreviousRepaymentReserve int64 `json:"previous_repayment_reserve" xorm:"previous_repayment_reserve"` // 이전 상환적립금
	RepaymentReserve         int64 `json:"repayment_reserve" xorm:"repayment_reserve"`                   // 상환적립금
	InvestmentBalance        int64 `json:"investment_balance" xorm:"investment_balance"`                 // 원금잔액

	// 배당정보
	DividendAmount            int64   `json:"dividend_amount" xorm:"dividend_amount"`                             // 배당액
	NetDividendAfterBrokerFee int64   `json:"net_dividend_after_broker_fee" xorm:"net_dividend_after_broker_fee"` // 공제 후 순배당액
	BrokerFee                 int64   `json:"broker_fee" xorm:"broker_fee"`                                       // 영업자 수수료
	ReturnRate                float64 `json:"return_rate" xorm:"return_rate"`                                     // 수익율
	DividendPerShare          int64   `json:"dividend_per_share" xorm:"dividend_per_share"`                       // 구좌당 배당액
	WithholdingTax            int64   `json:"withholding_tax" xorm:"withholding_tax"`                             // 원천징수 예치액

	// 법정장부를 위한 부가정보
	ProjectSpvId int64 `json:"-" xorm:"project_spv_id"` // 프로젝트 SPV ID
	NumInvestors int64 `json:"-" xorm:"num_investors"`  // 투자자 수

	// 기본날짜정보
	Created jsonutils.JSONTime `json:"created" xorm:"created"` // 생성 날짜
	Updated jsonutils.JSONTime `json:"updated" xorm:"updated"` // 업데이트 날짜
	Deleted jsonutils.JSONTime `json:"-" xorm:"deleted"`       // 삭제 날짜
}

type RepaymentDividendSchedules []RepaymentDividendSchedule

func (r RepaymentDividendSchedules) isEmpty() bool {
	return len(r) == 0
}

func (r RepaymentDividendSchedules) First() RepaymentDividendSchedule {
	if r.isEmpty() {
		return RepaymentDividendSchedule{}
	}
	return r[0]
}

func (RepaymentDividendSchedule) TableName() string {
	return "repayment_dividend_schedule"
}

func (r RepaymentDividendSchedule) SelectAll() string {
	return r.TableName() + ".*"
}

// 배당 상환 스케쥴 가져오기
func (RepaymentDividendSchedule) GetByProjectId(ctx context.Context, projectId int64, search SearchInput) (int64, []RepaymentDividendSchedule, error) {
	var result []RepaymentDividendSchedule
	q := factory.DB(ctx).Table("repayment_dividend_schedule").
		Where("project_id = ?", projectId)

	// 정렬 설정
	if len(search.Sortby) <= 0 {
		search.Sortby = "execute_date"
	}
	if len(search.Order) <= 0 {
		search.Order = "asc"
	}

	if err := modelhelper.SetSortOrder(q, search.Sortby, search.Order); err != nil {
		factory.Logger(ctx).Error(err)
	}

	total, err := q.Limit(search.MaxResultCount, search.SkipCount).FindAndCount(&result)
	if err != nil {
		return 0, nil, err
	}

	return total, result, nil
}

// 배당 상환 마지막 스케쥴 가져오기
func (RepaymentDividendSchedule) GetLastSchedultByProjectIdCompleted(ctx context.Context, projectId int64) (*RepaymentDividendSchedule, error) {
	var result RepaymentDividendSchedule
	q := factory.DB(ctx).Table("repayment_dividend_schedule").
		Where("project_id = ? AND status = 'RETURN_COMPLETED'", projectId).
		Desc("execute_date", "id"). // execute_date, id 기준으로 내림차순 정렬
		Limit(1)                    // 하나의 레코드만 선택

	found, err := q.Get(&result)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil // 레코드가 없는 경우 nil 반환
	}
	return &result, nil
}

func (r *RepaymentDividendSchedule) Update(ctx context.Context) error {
	_, err := factory.DB(ctx).ID(r.Id).AllCols().Update(r)
	return err
}
func (r *RepaymentDividendSchedule) Completed(ctx context.Context) error {

	beforeProject, err := Project{}.GetById(ctx, r.ProjectId)
	if err != nil {
		return err
	}

	r.Status = RETURN_COMPLETED
	if err = r.Update(ctx); err != nil {
		return err
	}

	afterProject, err := Project{}.GetById(ctx, r.ProjectId)
	if err != nil {
		return err
	}

	_, err = ProjectInfoHistory{}.Create(ctx, beforeProject, afterProject, 0)
	if err != nil {
		return err
	}

	return nil
}

type InvestmentAndReturns struct {
	Id                int64                    `json:"id" description:"출자 또는 배당/상환 스케쥴 ID"`
	ProjectId         int64                    `json:"project_id" description:"프로젝트 ID"`
	Type              INVESTMENT_RETURN_TYPE   `json:"type" description:"구분: INVESTMENT_EXECUTE, REPAYMENT_RETURN, DIVIDEND_RETURN"`
	DecideDate        string                   `json:"decide_date" description:"결정일"`
	ExecuteDate       string                   `json:"execute_date" description:"실시일"`
	Status            INVESTMENT_RETURN_STATUS `json:"status" description:"상태: RETURN_SCHEDULED, RETURN_COMPLETED"`
	Repayment         int64                    `json:"repayment" description:"상환액"`
	Dividend          int64                    `json:"dividend" description:"배당액"`
	BrokerFee         int64                    `json:"broker_fee" description:"공제액"`
	WithholdingTax    int64                    `json:"withholding_tax" description:"원천징수 보유액세"`
	RepaymentReserve  int64                    `json:"repayment_reserve" description:"상환적립금"`
	InvestmentBalance int64                    `json:"investment_balance" description:"투자원금잔액"`
}
