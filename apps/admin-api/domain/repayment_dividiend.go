package domain

import (
	"context"
	"time"

	"github.com/illuminarean/energy-share-core-api/controllers"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
)

type RepaymentDividendUsecase interface {
	ExecuteInvestment(ctx context.Context, request ExecuteInvestmentRequest) (ExecuteInvestmentResponse, error)
	CreateRepaymentDividendSchedules(ctx context.Context, request CreateRepaymentDividendRequest) (
		CreateRepaymentDividendResponse, error)
	GetRepaymentDividendSchedulesTable(ctx context.Context, request GetRepaymentDividendTableRequest) (
		GetRepaymentDividendTableResponse, error)
	GetRepaymentDividendSchedulesTableCsv(ctx context.Context, request GetRepaymentDividendTableRequest) (
		GetRepaymentDividendTableCsvResponse, error)
	GetInvestmentAndReturnsInfo(ctx context.Context, request GetInvestmentAndReturnsInfoRequest) (
		GetInvestmentAndReturnsInfoResponse, error)
	DeleteRepaymentDividendSchedule(ctx context.Context, request DeleteRepaymentDividendRequest) error
}

type ExecuteInvestmentRequest struct {
	JWTUser   jwthelper.User
	ProjectId int64 `param:"id" path:"id" description:"출자할 프로젝트 ID" example:"1" format:"int64"`
	ExecuteInvestmentBody
}

type ExecuteInvestmentBody struct {
	Amount      int64  `json:"amount" validate:"required"  description:"출자금액" example:"10000" format:"int64"`
	ExecuteDate string `json:"execute_date" validate:"required,datetime=2006-01-02"  description:"출자실행일" example:"2024-10-10" format:"yyyy-mm-dd"`
}

func (r *ExecuteInvestmentRequest) ExecuteDateIsAfterToday() (bool, error) {
	parsedDate, err := jsonutils.ParseDateTimeJST(time.DateOnly, r.ExecuteDate)
	if err != nil {
		return false, err
	}

	now := time.Now().In(time.Local)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	isAfter := parsedDate.After(today) // Check if the parsed date is before today
	return isAfter, nil
}

type ExecuteInvestmentResponse struct {
	domain.ProjectInvestmentExecute
	// ProcessedCount は今回の実行で処理したshareの件数、RemainingCount は
	// この実行後もまだ未処理(SUCCESS_INVEST)のまま残っているshareの件数
	// (MIR-1204)。RemainingCountが0より大きければ、同じリクエストを再度
	// 送ることで続きから処理できる。
	ProcessedCount int64 `json:"processed_count" description:"desc(이번 실행에서 처리한 건수)" format:"int64"`
	RemainingCount int64 `json:"remaining_count" description:"desc(미처리로 남은 건수. 0보다 크면 이어서 실행이 필요하다)" format:"int64"`
}

type CreateRepaymentDividendRequest struct {
	JWTUser                     jwthelper.User
	ProjectId                   int64 `param:"id" path:"id" description:"배당/상환 스케쥴을 생성할 프로젝트 ID" example:"1" format:"int64"`
	CreateRepaymentDividendBody `description:"상환일 경우 dividend_amount, broker_fee, return_rate이 들어가지 않는다.\n배당일 경우 repayment_amount가 들어가지 않는다."`
}
type CreateRepaymentDividendBody struct {
	RepaymentAmount int64   `json:"repayment_amount" description:"상환액 - 배당일 경우 제외"`
	DividendAmount  int64   `json:"dividend_amount" description:"배당액 - 상환일 경우 제외"`
	BrokerFee       int64   `json:"broker_fee" description:"영업자 보수 - 상환일 경우 제외"`
	ReturnRate      float64 `json:"return_rate" description:"배당 수익률 - 장부 기입용, 상환일 경우 제외"`
	DecideDate      string  `json:"decide_date" validate:"required,datetime=2006-01-02" description:"결정 일시" example:"2024-10-10" format:"yyyy-mm-dd"`
	ExecuteDate     string  `json:"execute_date" validate:"required,datetime=2006-01-02" description:"실시 일시" example:"2024-10-12" format:"yyyy-mm-dd"`
}

type CreateRepaymentDividendResponse struct {
	domain.RepaymentDividendSchedule
	// UnmatchedInvestmentCount は、出資実行が完了しているファンドで、その後
	// SUCCESS_INVESTのまま紛れ込んだ出資の件数(MIR-1239)。この出資は分配・
	// 償還の対象にならないため、0より大きい場合は画面側で警告を表示する。
	UnmatchedInvestmentCount int64 `json:"unmatched_investment_count" description:"desc(출자실행 완료 후 남아있는 미반영 출자 건수)" format:"int64"`
}

type GetRepaymentDividendTableRequest struct {
	JWTUser   jwthelper.User
	ProjectId int64                           `param:"id" path:"id" description:"출자/배당/상환 스케쥴을 가져올 프로젝트 ID" example:"1" format:"int64"`
	Status    domain.INVESTMENT_RETURN_STATUS `param:"status" query:"status" validate:"omitempty,oneof=RETURN_SCHEDULED RETURN_COMPLETED" description:"상태" enum:"RETURN_SCHEDULED,RETURN_COMPLETED" example:"RETURN_SCHEDULED"`
	controllers.SearchInput
}

type GetRepaymentDividendTableResponse struct {
	Items      []domain.InvestmentAndReturns `json:"items"`
	TotalCount int64                         `json:"total_count"`
}

type GetRepaymentDividendTableCsvResponse struct {
	Data []byte
}

type GetInvestmentAndReturnsInfoRequest struct {
	JWTUser   jwthelper.User
	ProjectId int64 `param:"id" path:"id" description:"자금정보를 가져올 가져올 프로젝트 ID" example:"1" format:"int64"`
	ClientIp  string
}

// 출자/배당/상환 테이블
// type InvestmentAndReturnsDto struct 를 그대로 가져옴
type GetInvestmentAndReturnsInfoResponse struct {
	// 기본 정보
	ProjectAccount    string `json:"project_account" description:"펀드구좌"`       // A1. 펀드구좌
	ProjectDepositor  string `json:"project_depositor" description:"펀드명의인"`    // A2. 펀드 명의인
	ServiceAccount    string `json:"service_account" description:"ESKK 예금주"`   // 항상 "00000"
	TargetShareAmount int64  `json:"target_share_amount" description:"모집 목표액"` // 펀드가 목표로 한 금액
	RaisedAmount      int64  `json:"raised_amount" description:"모집액총액"`        // A3. 모집액총액

	// 상환 금액
	InvestedAmount    int64 `json:"invested_amount" description:"출자액"`
	InvestmentBalance int64 `json:"investment_balance" description:"출자원금잔액"` // A4. 출자원금잔액
	TotalRepayment    int64 `json:"total_repayment" description:"상환액 합계"`    // 실제 상환액은 여기서 상환적립금을 빼야 한다.
	RepaymentReserve  int64 `json:"repayment_reserve" description:"상환적립금"`   // A7. 상환적립금

	// UnprocessedShareCount は未処理(SUCCESS_INVEST)のまま残っているshareの
	// 件数(MIR-1204)。出資実行が分割実行中かどうかを画面が判定するために使う。
	// ProjectInvestmentExecuteレコードの有無だけでは「出資実行が完了したか」を
	// 判定できない(分割実行では最初のバッチで作成され、その後も残件がある)。
	UnprocessedShareCount int64 `json:"unprocessed_share_count" description:"desc(미처리로 남은 share 건수)" format:"int64"`

	// 배당 금액
	TotalNetDividendAfterBrokerFee int64 `json:"total_net_dividend_after_broker_fee" description:"현금분배액총액 - 세전"` // A5. 현금분배액총액(세전)
	TotalBrokerFee                 int64 `json:"total_broker_fee" description:"공제액"`                             // A6. 공제액
	TotalWithholdingTax            int64 `json:"total_withholding_tax" description:"원천징수 예치금"`                   // A8. 원천징수 예치금
}

type DeleteRepaymentDividendRequest struct {
	JWTUser    jwthelper.User
	ProjectId  int64 `param:"id" path:"id" description:"삭제할 스케쥴의 프로젝트 ID" example:"1" format:"int64"`
	ScheduleId int64 `param:"schedule_id" path:"schedule_id" description:"삭제할 스케쥴 ID" example:"1" format:"int64"`
}

type DeleteRepaymentDividendResponse struct {
	domain.RepaymentDividendSchedule
}
