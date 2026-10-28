package domain

import (
	"context"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
)

type UserUsecase interface {
	GetSentMails(ctx context.Context, request UserGetSentMailsRequest) (MailGetAllResponse, error)
	DownloadUserByFilter(ctx context.Context, request UserDownloadUserByFilterRequest) (UserDownloadUserByFilterResponse, error)
}

type UserGetSentMailsRequest struct {
	JWTUser  jwthelper.User
	ClientIp string
	UserId   int64 `param:"id" path:"id" description:"desc(사용자 아이디)" example:"1" format:"int64"`
	MailGetAllQueryParam
	Env string
	cdomain.SearchInput
}

type UserDownloadUserByFilterRequest struct {
	JWTUser  jwthelper.User
	ClientIp string

	// 문자열 옵션
	Emails        []string                     `json:"emails" description:"이메일"`
	AuditStatuses []domain.AUDIT_STATUS        `json:"audit_statuses" description:"심사 상태 (AUDIT_STATUS)" enum:"PROVISIONAL_AUDIT,WAIT_AUDIT,HOLD_AUDIT,WAIT_ANTISOCIAL_AUDIT,CHECK_ANTISOCIAL_AUDIT,CHECK_ADDRESS_AUDIT,CHECK_MY_NUMBER_AUDIT,COMPLETED_AUDIT,INVALIDITY_AUDIT"`
	UserTypes     []domain.USER_TYPE           `json:"user_types" description:"회원 구분 (USER_TYPE)" enum:"PRIVATE,CORPORATE"`
	Nationalities []domain.NATIONALITY_TYPE    `json:"nationalities" description:"국가 (NATIONALITY_TYPE)" enum:"JP,OTHERS"`
	Genders       []domain.GENDER              `json:"genders" description:"성별 (GENDER)" enum:"FEMALE,MALE"`
	Jobs          []domain.USER_JOB_TYPE       `json:"jobs" description:"직업 (USER_JOB_TYPE)"`
	Assets        []domain.ASSETS_STATUS       `json:"assets" description:"자산 (ASSETS_STATUS)"`
	AnnualIncomes []domain.ANNUALINCOME_STATUS `json:"annual_incomes" description:"연봉 (ANNUALINCOME_STATUS)"`

	// true, false
	AllowsNewsletter *bool `json:"allows_newsletter" description:"뉴스레터 수신 동의 [true|false]"`
	HasMyNumber      *bool `json:"has_my_number" description:"마이 넘버 제출 여부 [true|false]"`

	// date ~ date
	SignUpRequestDate  *DateRange `json:"sign_up_request_date" description:"가입 요청일"`
	SignUpApprovalDate *DateRange `json:"sign_up_approval_date" description:"가입 승인일"`
	LastLoginDate      *DateRange `json:"last_login_date" description:"최종 로그인"`

	// number ~ number
	Age                   *NumberRange `json:"age" description:"연령대"`
	TotalAsset            *NumberRange `json:"total_asset" description:"총 자산"`
	InvestOperationAmount *NumberRange `json:"invest_operation_amount" description:"운용 중 금액"`
	Deposit               *NumberRange `json:"deposit" description:"투자 가능 금액(예치금)"`
	InvestSuccessAmount   *NumberRange `json:"invest_success_amount" description:"투자 완료 금액"`
	InvestRequestAmount   *NumberRange `json:"invest_request_amount" description:"투자 신청 금액"`
	UnpaidAmount          *NumberRange `json:"unpaid_amount" description:"입금 필요 금액"`
}

type DateRange struct {
	StartDate jsonutils.JSONTime `json:"start_date" description:"시작일(yyyy-MM-dd 00:00:00)"`
	EndDate   jsonutils.JSONTime `json:"end_date" description:"종료일(yyyy-MM-dd 23:59:59)"`
}

type NumberRange struct {
	Min *int64 `json:"min" description:"최소값"`
	Max *int64 `json:"max" description:"최대값(~2140000000)"`
}

type UserDownloadUserByFilterResponse struct {
	Data []byte
}
