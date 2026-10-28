package controllers

import (
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

func calcPlanedOperationDays(start, end jsonutils.JSONTime) int {
	s := time.Time(start)
	e := time.Time(end)
	if s.IsZero() || e.IsZero() {
		return 0
	}
	days := int(e.Sub(s).Hours()/24) + 1 // 初日を含む
	if days < 1 {
		return 0
	}
	return days
}

type ProjectSearchInput struct {
	SearchInput
	ProjectSearchItem
	StartDate string `query:"start_date" swagger:"desc(시작일 등록일자)"`
	EndDate   string `query:"end_date" swagger:"desc(종료일 등록일자)"`
}

type ProjectSearchItem struct {
	ProjectId         *string `query:"id" swagger:"desc(프로젝트 ID)"`
	ProjectName       *string `query:"name" swagger:"desc(프로젝트 이름)"`
	ProjectStatus     *string `query:"project_status" swagger:"desc(모집상태)"`
	BlockInvest       *bool   `query:"block_invest" swagger:"desc(투자 제한),enum(true|false)"`
	TargetShareAmount *string `query:"target_share_amount" swagger:"desc(목표 금)"`
	ProjectType       *string `query:"project_type" swagger:"dsec(투자 유형)"`
	CategoryType      *string `query:"category_type" swagger:"desc(발전 유형)"`
	OutOfDateProject  bool    `query:"out_of_date_project" swagger:"desc(게시일이 지난 상품)"`
}

type ProjectListSearchInput struct {
	ListType       domain.PROJECT_STATUS_LIST_TYPE `query:"ListType" swagger:"PROJECT_STATUS_LIST_TYPE"`
	ProjectName    *string                         `query:"project_name"`
	SkipCount      int                             `query:"skip_count"`
	MaxResultCount int                             `query:"max_result_count"`
	StartDate      string                          `query:"start_date"`
	EndDate        string                          `query:"end_date"`
	SearchInput
}

type ProjectBayPage struct {
	User               *domain.User               `json:"user"`
	Project            *domain.Project            `json:"project"`
	UserVirtualAccount *domain.UserVirtualAccount `json:"user_virtual_account"`
}

type ProjectItemsResponseDto struct {
	Id                          int64                         `json:"id"`
	CommonId                    string                        `json:"common_id" swagger:"desc(공용 ID)"`
	SpvId                       int64                         `json:"spv_id" i18n:"commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_SPV" swagger:"desc(개발회사 ID)"`
	Name                        string                        `json:"name"  i18n:"commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_NAME" swagger:"desc(상품명)"`
	Description                 string                        `json:"description" xorm:"not null" i18n:"commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_DESCRIPTION" swagger:"desc(상품 설명)"` // 등록자? 담당자?
	OfferingType                domain.OFFERING_TYPE          `json:"offering_type"  swagger:"desc(사모 공모 여부)"`                                                                                              // 펀딩 타입
	SelectionType               domain.PROJECT_SELECTION_TYPE `json:"selection_type" swagger:"desc(선정 방식),enum(LOTTERY|FIRST_COME_FIRST_SERVED)"`
	StartAt                     jsonutils.JSONTime            `json:"start_at" i18n:"commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_FUNDRAISING_START_DATE" swagger:"desc(모집 시작일 , 펀드 개시일)" ` // 프로젝트 모집 시작일
	EndAt                       jsonutils.JSONTime            `json:"end_at" i18n:"commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_FUNDRAISING_END_DATE" swagger:"desc(모집 종료일)"`               // 프로젝트 모집 종료일
	PlanedOperationStartDate    jsonutils.JSONTime            `json:"planed_operation_start_date" swagger:"desc(予定運用開始日)"`
	PlanedOperationEndDate      jsonutils.JSONTime            `json:"planed_operation_end_date" swagger:"desc(予定運用終了日)"`
	PlanedOperationDays         int                           `json:"planed_operation_days" swagger:"desc(予定運用期間日数)"`
	ProjectType                 domain.PROJECT_TYPE           `json:"project_type" i18n:"commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_INVESTMENT_METHOD" swagger:"desc(프로젝트 타입 1:LOAN 2:EQUITY)"` // Loan Equity Fund
	ProjectStatus               domain.PROJECT_STATUS         `json:"project_status" i18n:"commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_PROJECT_STATUS" swagger:"desc(프로젝트 상태(펀딩 상태) 1:모집대기 2:모집중 3:모집성공 4:모집 실패 5:투자 제한)"`
	Country                     domain.PROJECT_COUNTRY_TYPE   `json:"country" swagger:"enum(JP|US) desc(국가)"`
	CategoryType                domain.PROJECT_CATEGORY       `json:"category_type" swagger:"desc(발전 유형) "` // 펀팅 카테고리
	DividendRate                float64                       `json:"dividend_rate" swagger:"desc(배당율)"`
	OperationTerm               int                           `json:"operation_term" swagger:"desc(운용기간)"`
	CurrentShareAmount          int64                         `json:"current_share_amount" swagger:"desc(모집된 Share 금액)"`                                                                                                // 모집된 주식 금액
	CurrentShareQty             int64                         `json:"current_share_qty" swagger:"desc(현재 모집된 Share 수량)"`                                                                                                // 모집된 주식 수량
	TargetMinShareAmount        int64                         `json:"target_min_share_amount" i18n:"commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_MINIMUM_FUNDRAISING_AMOUNT" swagger:"desc(목표 모금액)"` // 목표 모금액
	TargetShareQty              int64                         `json:"target_share_qty" i18n:"commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_SHARES" swagger:"desc(목표 주식수 총 발생 주식)"`                    // 목표 주식수
	TargetShareAmount           int64                         `json:"target_share_amount" i18n:"commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_TARGET_FUNDRAISING_AMOUNT" swagger:"desc(상한 모금액)"`      // 총 대출액
	MinShareQty                 int64                         `json:"min_share_qty" swagger:"desc(最小購入口数)"`
	SharePrice                  int64                         `json:"share_price" swagger:"desc(1口あたりの価格)"`
	Posted                      bool                          `json:"posted"  i18n:"commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_POSTED_STATE" swagger:"desc(펀드 게시여부)"` // 펀딩 디스플레이 여부 펀딩 시작후 수정 불가
	MainImageUrl                string                        `json:"main_image_url" i18n:"commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_MAINIMAGE_URL" swagger:"desc(펀딩 메인 이미지 URL)"`
	IsWaitInvest                bool                          `json:"is_wait_invest"  swagger:"desc(투자 대기중 상태 True:투자 대기 상품 False: 투자 허용)"`
	IsInvestorLimit             bool                          `json:"is_investor_limit"  swagger:"desc(투자자수 499명 제한 여부)"`
	Created                     jsonutils.JSONTime            `json:"created" `
	Updated                     jsonutils.JSONTime            `json:"updated" `
	PublicProgressInvestPercent float64                       `json:"public_progress_invest_percent" swagger:"desc(상품 상태별 투자율%)"`
	TransactionAmount           int                           `json:"transaction_amount"  swagger:"desc(거래 모집 긍액)"`
	ProjectProgressPercent      int                           `json:"project_progress_percent" swagger:"desc(사용자 프로그래스바용 모집액%)"`
	RemainingQty                int64                         `json:"remaining_qty"  swagger:"desc(해당 프로젝트 구매 가능한 구수)"`
	CurrentInvestorCount        int64                         `json:"current_investor_count"  swagger:"desc(해당 프로젝트 투자자 인원 수)"`
	RaisedAmount                int64                         `json:"raised_amount"   swagger:"desc(모집 조달 금액)"`
	ProjectReportDate           jsonutils.JSONTime            `json:"project_report_date"  swagger:"desc(펀드 보고서 등록 날짜)"`
	PublicViewShareAmount       int64                         `json:"public_view_share_amount" swagger:"desc(상품 상태별 Share 금액)"`
	DividendType                *domain.DIVIDEND_TYPE         `json:"dividend_type" swagger:"desc(分配方式)"`
	CapitalRaisingEntity        *domain.CapitalRaisingEntity  `json:"capital_raising_entity" swagger:"desc(資金調達者)"`
}

func (ProjectItemsResponseDto) New(p *domain.Project) *ProjectItemsResponseDto {
	return &ProjectItemsResponseDto{
		Id:                          p.Id,
		CommonId:                    p.CommonId,
		SpvId:                       p.SpvId,
		Name:                        p.Name,
		Description:                 p.Description,
		OfferingType:                p.OfferingType,
		SelectionType:               p.SelectionType,
		StartAt:                     p.StartAt,
		EndAt:                       p.EndAt,
		PlanedOperationStartDate:    p.PlanedOperationStartDate,
		PlanedOperationEndDate:      p.PlanedOperationEndDate,
		PlanedOperationDays:         calcPlanedOperationDays(p.PlanedOperationStartDate, p.PlanedOperationEndDate),
		ProjectType:                 p.ProjectType,
		ProjectStatus:               p.ProjectStatus,
		Country:                     p.Country,
		CategoryType:                p.CategoryType,
		DividendRate:                p.DividendRate,
		OperationTerm:               p.OperationTerm,
		CurrentShareAmount:          p.CurrentShareAmount,
		CurrentShareQty:             p.CurrentShareQty,
		TargetMinShareAmount:        p.TargetMinShareAmount,
		TargetShareQty:              p.TargetShareQty,
		TargetShareAmount:           p.TargetShareAmount,
		MinShareQty:                 p.MinShareQty,
		SharePrice:                  p.SharePrice,
		Posted:                      p.Posted,
		MainImageUrl:                p.MainImageUrl,
		IsWaitInvest:                p.IsWaitInvest,
		IsInvestorLimit:             p.IsInvestorLimit,
		Created:                     p.Created,
		Updated:                     p.Updated,
		ProjectProgressPercent:      p.ProjectProgressPercent,
		RemainingQty:                p.RemainingQty,
		CurrentInvestorCount:        p.CurrentInvestorCount,
		RaisedAmount:                p.RaisedAmount,
		ProjectReportDate:           p.ProjectReportDate,
		PublicViewShareAmount:       p.PublicViewShareAmount,
		PublicProgressInvestPercent: p.PublicProgressInvestPercent,
		DividendType:                p.DividendType,
		CapitalRaisingEntity:        p.CapitalRaisingEntity,
	}
}

func (ProjectItemsResponseDto) NewList(projects []domain.Project) []*ProjectItemsResponseDto {
	var list []*ProjectItemsResponseDto
	for i := range projects {
		list = append(list, ProjectItemsResponseDto{}.New(&projects[i]))
	}
	return list
}

type ProjectDetailResponseDto struct {
	Id                          int64                         `json:"id" swagger:"desc(상품 ID)"`
	CommonId                    string                        `json:"common_id" swagger:"desc(공용 ID)"`
	SpvId                       int64                         `json:"spv_id" swagger:"desc(개발회사 ID)"`
	Spv                         *domain.Spv                   `json:"spv"  swagger:"desc(개발회사 SPV DEVELOPER)"`
	Name                        string                        `json:"name" swagger:"desc(상품명)"`
	CategoryType                domain.PROJECT_CATEGORY       `json:"category_type"  swagger:"desc(펀딩 카테고리) "` // 펀팅 카테고리
	SelectionType               domain.PROJECT_SELECTION_TYPE `json:"selection_type" swagger:"desc(선정 방식),enum(LOTTERY|FIRST_COME_FIRST_SERVED)"`
	ProjectInfoHtml             string                        `json:"project_info_html" swagger:"desc(프로젝트 정보)"`     // 프로젝트 컨텐츠 Html
	CompanyInfoHtml             string                        `json:"company_info_html" swagger:"desc(문의 하기,회사 정보)"` // 회사 컨텐츠 Html
	Description                 string                        `json:"description" swagger:"desc(상품 설명)"`             // 등록자? 담당자?
	StartAt                     jsonutils.JSONTime            `json:"start_at"  swagger:"desc(모집 시작일 , 펀드 개시일)" `    // 프로젝트 모집 시작일
	EndAt                       jsonutils.JSONTime            `json:"end_at"  swagger:"desc(모집 종료일)"`                // 프로젝트 모집 종료일
	PlanedOperationStartDate    jsonutils.JSONTime            `json:"planed_operation_start_date" swagger:"desc(予定運用開始日)"`
	PlanedOperationEndDate      jsonutils.JSONTime            `json:"planed_operation_end_date" swagger:"desc(予定運用終了日)"`
	PlanedOperationDays         int                           `json:"planed_operation_days" swagger:"desc(予定運用期間日数)"`
	ProjectType                 domain.PROJECT_TYPE           `json:"project_type"  swagger:"desc(프로젝트 타입 1:LOAN 2:EQUITY)"` // Loan Equity Fund
	ProjectStatus               domain.PROJECT_STATUS         `json:"project_status"  swagger:"desc(프로젝트 상태(펀딩 상태) 1:모집대기 2:모집중 3:모집성공 4:모집 실패 5:투자 제한)"`
	OwnerId                     int64                         `json:"owner_id"  swagger:"desc(등록자)"` // 등록자? 담당자?
	UpdateId                    int64                         `json:"update_id"  swagger:"desc(최근 수정자)"`
	Country                     domain.PROJECT_COUNTRY_TYPE   `json:"country" swagger:"desc(국가)"`
	Capacity                    string                        `json:"capacity"  swagger:"desc(개발 용량,KW,formP DC 규모)"`          // 개발 용량 주식 대출 펀드
	DividendRate                float64                       `json:"dividend_rate" swagger:"desc(배당율)" validate:"required"`   // 배당율
	OperationTerm               int                           `json:"operation_term" swagger:"desc(운용기간)" validate:"required"` // 운용기間
	OperationTermDays           int                           `json:"operation_term_days" swagger:"desc(運用期間日数)"`
	CurrentShareAmount          int64                         `json:"current_share_amount" swagger:"desc(모집된 Share 금액)"` // 모집된 주식 금액
	CurrentShareQty             int64                         `json:"current_share_qty" swagger:"desc(현재 모집된 Share 수량)"` // 모집된 주식 수량
	MinShareQty                 int64                         `json:"min_share_qty"  swagger:"desc(최소 구매 수량)"`           // 최소 구매 수량
	SharePrice                  int64                         `json:"share_price"  swagger:"desc(1구좌당 가격)"`              // 1구좌 당 가격
	TargetMinShareAmount        int64                         `json:"target_min_share_amount"  swagger:"desc(목표 모금액)"`   // 목표 모금액
	TargetShareQty              int64                         `json:"target_share_qty"  swagger:"desc(목표 주식수 총 발생 주식)"`  // 목표 주식수
	TargetShareAmount           int64                         `json:"target_share_amount"  swagger:"desc(상한 모금액)"`       // 총 대출액
	Posted                      bool                          `json:"posted"  swagger:"desc(펀드 게시여부)"`                   // 펀딩 디스플레이 여부 펀딩 시작후 수정 불가
	ViewCount                   int64                         `json:"view_count" swagger:"desc(퍼블릭 조회 카운트)"`             // (퍼블릭 조회 카운트)
	BlockInvest                 bool                          `json:"block_invest"  swagger:"desc(투자 제한 True:투자제한됨 False: 투자 허용)"`
	MainImageId                 int64                         `json:"main_image_id" swagger:"desc(프로젝트 메인 이미지 고유 아이디)"`
	MainImageUrl                string                        `json:"main_image_url"  swagger:"desc(펀딩 메인 이미지 URL)"`
	MainImage                   *domain.ProjectImage          `json:"main_image"  swagger:"desc(프로젝트 이미지)"` // 프로젝트 메인 이미지
	SubImageInfoList            []*domain.SubImageInfo        `json:"sub_image_info_list" swagger:"desc(サブ画像情報リスト)"`
	Attachments                 []*domain.Attachment          `json:"attachments"  swagger:"desc(프로젝트 첨부파일 목록 )"` // 첨부 파일 // 서브 자료
	News                        []*domain.ProjectNews         `json:"project_news"  swagger:"desc(프로젝트 뉴스 목록)"`   // 첨부 파일 // 서브 자료
	IsWaitInvest                bool                          `json:"is_wait_invest"  swagger:"desc(투자 대기중 상태 True:투자 대기 상품 False: 투자 허용)"`
	IsManualInvest              bool                          `json:"is_manual_invest"  swagger:"desc(수동투자 여부)"`
	IsInvestorLimit             bool                          `json:"is_investor_limit"  swagger:"desc(투자자수 제한 여부)"`
	IsAmountLimit               bool                          `json:"is_amount_limit"  swagger:"desc(투자금액 제한 여부)"`
	IsVisibleInvestmentInfo     bool                          `json:"is_visible_investment_info" swagger:"desc(투자정보 공개 여부)"`
	TransactionCount            int64                         `json:"transaction_count" swagger:"desc(거래 카운트)"`
	CautionInfoHtml             string                        `json:"caution_info_html" swagger:"desc(유의 사항 )"` // 유의 사항  Html
	ContractBeforePage          *domain.ProjectFile           `json:"contract_before_page"  swagger:"desc(계약성립 전 서면)"`
	ContractAfterPage           *domain.ProjectFile           `json:"contract_after_page"  swagger:"desc(계약성립 시 서면)"`
	SilentPartnerships          *domain.ProjectFile           `json:"silent_partnerships"  swagger:"desc(익명조합 계약서)"`
	ProjectReport               *domain.ProjectFile           `json:"project_report"  swagger:"desc(펀드 보고서)"`
	RemainingQty                int64                         `json:"remaining_qty"  swagger:"desc(해당 프로젝트 구매 가능한 구수)"`
	CurrentInvestorCount        int64                         `json:"current_investor_count"  swagger:"desc(해당 프로젝트 투자자 인원 수)"`
	RaisedAmount                int64                         `json:"raised_amount"   swagger:"desc(모집 조달 금액)"`
	InvestWaitRemainingQty      int64                         `json:"invest_wait_remaining_qty" swagger:"desc(해당 프로젝트 투자 대기 가능한 구수)"`
	PublicViewShareAmount       int64                         `json:"public_view_share_amount" swagger:"desc(상품 상태별 Share 금액)"`
	PublicProgressInvestPercent float64                       `json:"public_progress_invest_percent" swagger:"desc(상품 상태별 투자율%)"`
	ProjectProgressPercent      int                           `json:"project_progress_percent" swagger:"desc(프로젝트 진행율)"`
	ProjectMilestones           []domain.ProjectMilestone     `json:"project_milestones" xorm:"-" swagger:"desc(프로젝트 마일스톤 목록)"`
	Created                     jsonutils.JSONTime            `json:"created" `
	Updated                     jsonutils.JSONTime            `json:"updated" `
	DividendType                *domain.DIVIDEND_TYPE         `json:"dividend_type" swagger:"desc(分配方式)"`
	CapitalRaisingEntity        *domain.CapitalRaisingEntity  `json:"capital_raising_entity" swagger:"desc(資金調達者)"`
}

func (ProjectDetailResponseDto) ToProjectPubResponseDto(p *domain.Project) *ProjectDetailResponseDto {
	return &ProjectDetailResponseDto{
		Id:                          p.Id,
		CommonId:                    p.CommonId,
		SpvId:                       p.SpvId,
		Spv:                         p.Spv,
		Name:                        p.Name,
		CategoryType:                p.CategoryType,
		SelectionType:               p.SelectionType,
		ProjectInfoHtml:             p.ProjectInfoHtml,
		CompanyInfoHtml:             p.CompanyInfoHtml,
		Description:                 p.Description,
		StartAt:                     p.StartAt,
		EndAt:                       p.EndAt,
		PlanedOperationStartDate:    p.PlanedOperationStartDate,
		PlanedOperationEndDate:      p.PlanedOperationEndDate,
		PlanedOperationDays:         calcPlanedOperationDays(p.PlanedOperationStartDate, p.PlanedOperationEndDate),
		ProjectType:                 p.ProjectType,
		ProjectStatus:               p.ProjectStatus,
		OwnerId:                     p.OwnerId,
		UpdateId:                    p.UpdateId,
		Country:                     p.Country,
		Capacity:                    p.Capacity,
		DividendRate:                p.DividendRate,
		OperationTerm:               p.OperationTerm,
		OperationTermDays:           calcPlanedOperationDays(p.PlanedOperationStartDate, p.PlanedOperationEndDate),
		CurrentShareAmount:          p.CurrentShareAmount,
		CurrentShareQty:             p.CurrentShareQty,
		MinShareQty:                 p.MinShareQty,
		SharePrice:                  p.SharePrice,
		TargetMinShareAmount:        p.TargetMinShareAmount,
		TargetShareQty:              p.TargetShareQty,
		TargetShareAmount:           p.TargetShareAmount,
		Posted:                      p.Posted,
		ViewCount:                   p.ViewCount,
		BlockInvest:                 p.BlockInvest,
		MainImageId:                 p.MainImageId,
		MainImageUrl:                p.MainImageUrl,
		MainImage:                   p.MainImage,
		SubImageInfoList:            p.SubImageInfoList,
		Attachments:                 p.Attachments,
		News:                        p.News,
		IsWaitInvest:                p.IsWaitInvest,
		IsManualInvest:              p.IsManualInvest,
		IsInvestorLimit:             p.IsInvestorLimit,
		IsAmountLimit:               p.IsAmountLimit,
		IsVisibleInvestmentInfo:     p.IsVisibleInvestmentInfo,
		TransactionCount:            p.TransactionCount,
		CautionInfoHtml:             p.CautionInfoHtml,
		ContractBeforePage:          p.ContractBeforePage,
		ContractAfterPage:           p.ContractAfterPage,
		SilentPartnerships:          p.SilentPartnerships,
		ProjectReport:               p.ProjectReport,
		RemainingQty:                p.RemainingQty,
		CurrentInvestorCount:        p.CurrentInvestorCount,
		RaisedAmount:                p.RaisedAmount,
		InvestWaitRemainingQty:      p.InvestWaitRemainingQty,
		PublicViewShareAmount:       p.PublicViewShareAmount,
		PublicProgressInvestPercent: p.PublicProgressInvestPercent,
		ProjectProgressPercent:      p.ProjectProgressPercent,
		ProjectMilestones:           p.ProjectMilestones,
		Created:                     p.Created,
		Updated:                     p.Updated,
		DividendType:                p.DividendType,
		CapitalRaisingEntity:        p.CapitalRaisingEntity,
	}
}

// InvestProjectSharesResponse 投資リクエスト レスポンス
type InvestProjectSharesResponse struct {
	ProjectName       string `json:"project_name" swagger:"desc(プロジェクト名)"`
	BankName          string `json:"bank_name" swagger:"desc(銀行名)"`
	BankAccountNumber string `json:"bank_account_number" swagger:"desc(口座番号)"`
	BranchName        string `json:"branch_name" swagger:"desc(支店名)"`
	Depositor         string `json:"depositor" swagger:"desc(預金者)"`
	Qty               int64  `json:"qty" swagger:"desc(投資数量)"`
	Price             int64  `json:"price" swagger:"desc(投資単価)"`
	Amount            int64  `json:"amount" swagger:"desc(投資金額)"`
}
