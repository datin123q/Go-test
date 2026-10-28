package domain

import (
	"context"
	"github.com/illuminarean/energy-share-core-api/controllers"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
)

type InvestmentUsecase interface {
	GetAll(ctx context.Context, request InvestmentGetAllRequest) (InvestmentGetAllResponse, error)
	GetOne(ctx context.Context, request InvestmentGetOneRequest) (InvestmentGetOneResponse, error)
	DownloadHistories(ctx context.Context, request InvestmentDownloadHistoriesRequest) (InvestmentDownloadHistoriesResponse, error)
	UpdateStatus(ctx context.Context, request InvestmentUpdateStatusRequest) (InvestmentUpdateStatusResponse, error)
	GetAllTransactionInfoHistories(ctx context.Context, request InvestmentGetAllTransactionInfoHistoriesRequest) (InvestmentGetAllTransactionInfoHistoriesResponse, error)
}

type InvestmentGetAllRequest struct {
	controllers.SearchInput
	InvestmentGetAllQueryParam
	JWTUser  jwthelper.User
	ClientIp string
}

type InvestmentGetAllQueryParam struct {
	InvestStatus *string `query:"invest_status" swagger:"desc(거래 상태),enum(REQUEST_INVEST|FAIL_INVEST|UNSUCCESSFUL_INVEST|SELECTION_INVEST|CANCEL_INVEST|SUCCESS_INVEST|OPERATION_INVEST|REFUND_INVEST)"`
	InvestorName *string `query:"investor_name" swagger:"desc(투자자 명)"`
	SpvName      *string `query:"spv_name" swagger:"desc(개발회사)"`
	ProjectName  *string `query:"project_name" swagger:"desc(상품이름)"`
	ProjectType  *string `query:"project_type" swagger:"desc(상품유형)"`
	StartAmount  *int    `query:"start_amount" swagger:"desc(거래금액 이상)"`
	EndAmount    *int    `query:"end_amount" swagger:"desc(거래금액 이하)"`
}

type InvestmentGetAllResponse struct {
	TotalCount int64                    `json:"total_count"`
	Items      []InvestmentItemResponse `json:"items"`
}

type InvestmentDownloadHistoriesRequest struct {
	controllers.SearchInput
	InvestmentDownloadHistoriesQueryParam
}

type InvestmentDownloadHistoriesQueryParam struct {
	InvestStatus *string `query:"invest_status" swagger:"desc(거래 상태),enum(REQUEST_INVEST|FAIL_INVEST|UNSUCCESSFUL_INVEST|SELECTION_INVEST|CANCEL_INVEST|SUCCESS_INVEST|OPERATION_INVEST|REFUND_INVEST)"`
	InvestorName *string `query:"investor_name" swagger:"desc(투자자 명)"`
	SpvName      *string `query:"spv_name" swagger:"desc(개발회사)"`
	ProjectName  *string `query:"project_name" swagger:"desc(상품이름)"`
	StartAmount  *int    `query:"start_amount" swagger:"desc(거래금액 이상)"`
	EndAmount    *int    `query:"end_amount" swagger:"desc(거래금액 이하)"`
}

type InvestmentDownloadHistoriesResponse struct {
	Data []byte
}

type InvestmentGetOneRequest struct {
	InvestmentId string `param:"id"`
	JWTUser      jwthelper.User
	ClientIp     string
}

type InvestmentItemResponse struct {
	Id           string               `json:"id" swagger:"desc(투자 ID)"`
	ShareId      string               `json:"share_id" swagger:"desc(지분 ID/삭제예정)"`
	SpvId        int64                `json:"spv_id" swagger:"desc(개발회사 ID)"`
	SpvName      string               `json:"spv_name" swagger:"desc(개발회사 명)"`
	UserName     string               `json:"user_name" swagger:"desc(투자자 명)"`
	InvestStatus domain.INVEST_STATUS `json:"invest_status" swagger:"desc(투자 상태)"`
	ProjectType  domain.PROJECT_TYPE  `json:"project_type" swagger:"desc(투자 유형)"`
	ProjectName  string               `json:"project_name" swagger:"desc(투자 상품명)"`
	Amount       int64                `json:"amount" swagger:"desc(투자 금액)"`
	InvestDate   jsonutils.JSONTime   `json:"invest_date" swagger:"desc(투자 일시)"`
	OwnerId      int64                `json:"owner_id" swagger:"desc(소유자 ID)"`
}

type InvestmentItemDetailResponse struct {
	Id                          string                  `json:"id" swagger:"desc(투자 ID)"`
	OwnerId                     int64                   `json:"owner_id" swagger:"desc(투자자 ID)"`
	ShareId                     string                  `json:"share_id" swagger:"desc(지분 ID)"`
	ProjectId                   int64                   `json:"project_id" swagger:"desc(프로젝트 ID)"`
	CategoryType                domain.PROJECT_CATEGORY `json:"category_type" swagger:"desc(투자 유형)"`
	ProjectType                 domain.PROJECT_TYPE     `json:"project_type" swagger:"desc(투자 유형)"`
	InvestStatus                domain.INVEST_STATUS    `json:"invest_status" swagger:"desc(투자 상태)"`
	Amount                      int64                   `json:"amount" swagger:"desc(투자 금액)"`
	Qty                         int64                   `json:"qty" swagger:"desc(투자 수량)"`
	Price                       int64                   `json:"price" swagger:"desc(투자 단가)"`
	InvestDate                  jsonutils.JSONTime      `json:"invest_date" swagger:"desc(투자 일시)"`
	SpvId                       int64                   `json:"spv_id" swagger:"desc(개발회사 ID)"`
	SpvName                     string                  `json:"spv_name" swagger:"desc(개발회사 명)"`
	ProjectName                 string                  `json:"project_name" swagger:"desc(투자 상품명)"`
	StartAt                     jsonutils.JSONTime      `json:"start_at" swagger:"desc(투자 시작일)"`
	EndAt                       jsonutils.JSONTime      `json:"end_at" swagger:"desc(투자 종료일)"`
	MainImageUrl                string                  `json:"main_image_url" swagger:"desc(투자 상품 메인 이미지)"`
	ProjectStatus               domain.PROJECT_STATUS   `json:"project_status" swagger:"desc(투자 상품 상태)"`
	Capacity                    string                  `json:"capacity" swagger:"desc(투자 상품 용량)"`
	InvestorType                domain.INVESTOR_TYPE    `json:"investor_type" swagger:"desc(투자자 유형)"`
	Name                        string                  `json:"name" swagger:"desc(투자자 명)"`
	NameKana                    string                  `json:"name_kana" swagger:"desc(투자자 명 카나))"`
	FirstName                   string                  `json:"first_name" swagger:"desc(이름)"`          // 구매자 후리가나
	FamilyName                  string                  `json:"family_name" swagger:"desc(성)"`          // 구매자 성
	FirstNameKana               string                  `json:"first_name_kana" swagger:"desc(이름 카나))"` // 구매자 후리가나
	FamilyNameKana              string                  `json:"family_name_kana" swagger:"desc(성 카나))"` // 구매자 성
	Address                     string                  `json:"address" swagger:"desc(주소)"`
	Address2                    string                  `json:"address2" swagger:"desc(주소2)"`
	City                        string                  `json:"city" swagger:"desc(도시)"`
	ZipCode                     string                  `json:"zip_code" swagger:"desc(우편번호)"`
	Province                    string                  `json:"province" swagger:"desc(주)"`
	Email                       string                  `json:"email" swagger:"desc(이메일)"`
	Phone                       string                  `json:"phone" swagger:"desc(전화번호)"`
	BankName                    string                  `json:"bank_name" swagger:"desc(은행명)"`
	BankAccountNumber           string                  `json:"bank_account_number" swagger:"desc(계좌번호)"`
	Depositor                   string                  `json:"depositor" swagger:"desc(예금주)"`
	InvestmentDate              jsonutils.JSONTime      `json:"investment_date" swagger:"desc(투자일)"`
	ContractAfterPageUrl        *string                 `json:"contract_after_page_url" swagger:"desc(계약 후 페이지 URL)"`
	ContractBeforePageUrl       *string                 `json:"contract_before_page_url" swagger:"desc(계약 전 페이지 URL)"`
	SilentPartnershipsPageUrl   string                  `json:"silent_partnerships_page_url" swagger:"desc(조용한 파트너십 페이지 URL)"`
	ProjectReportUrl            string                  `json:"project_report_url" swagger:"desc(프로젝트 리포트 URL)"`
	ContractStatus              domain.CONTRACT_STATUS  `json:"contract_status" swagger:"desc(계약 상태)"`
	ContractBeforePageCheckDate jsonutils.JSONTime      `json:"contract_before_page_check_date" swagger:"desc(계약 전 페이지 확인 일시)"`
	ContractAfterPageCheckDate  jsonutils.JSONTime      `json:"contract_after_page_check_date" swagger:"desc(계약 후 페이지 확인 일시)"`
	SilentPartnershipsCheckDate jsonutils.JSONTime      `json:"silent_partnerships_check_date" swagger:"desc(조용한 파트너십 확인 일시)"`
	ProjectReportCheckDate      jsonutils.JSONTime      `json:"project_report_check_date" swagger:"desc(프로젝트 리포트 확인 상태)"`
	PublicViewEndAt             jsonutils.JSONTime      `json:"public_view_end_at" swagger:"desc(사용자 페이지 모집마감일 표시용)"` // 삭제 예정
}

type InvestmentGetOneResponse struct {
	InvestmentItemDetailResponse
}

type InvestmentUpdateStatusRequest struct {
	JWTUser      jwthelper.User
	InvestmentId string `param:"id"`
	InvestmentUpdateStatusBody
}

type InvestmentUpdateStatusBody struct {
	InvestStatus domain.INVEST_STATUS `json:"order_status" swagger:"enum(REQUEST_INVEST|PROCESS_INVEST|SUCCESS_INVEST|CANCEL_INVEST),desc(1:주문 요청 2:주문 처리중 3:주문 성공 4:주문 취소)" validate:"required"`
}

type InvestmentUpdateStatusResponse struct {
	domain.Investment // MarshalJSON 구현해서 사용하고 있음
}

type InvestmentGetAllTransactionInfoHistoriesRequest struct {
	controllers.SearchInput
	InvestmentId string `param:"id"`
}

type InvestmentGetAllTransactionInfoHistoriesResponse struct {
	TotalCount int64                           `json:"total_count"`
	Items      []domain.TransactionInfoHistory `json:"items"`
}
