package domain

import (
	"context"

	"github.com/illuminarean/energy-share-core-api/controllers"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
)

type InvestmentWaitersUsecase interface {
	GetAll(ctx context.Context, request InvestmentWaitersGetAllRequest) (InvestmentWaitersGetAllResponse, error)
	GetOne(ctx context.Context, request InvestmentWaitersGetOneRequest) (InvestmentWaitersGetOneResponse, error)
	Download(ctx context.Context, request InvestmentWaitersDownloadRequest) (InvestmentWaitersDownloadResponse, error)
	StatusUpdate(ctx context.Context, request InvestmentWaitersStatusUpdateRequest) (InvestmentWaitersStatusUpdateResponse, error)
	InvestmentPropose(ctx context.Context, request InvestmentWaitersInvestmentProposeRequest) (InvestmentWaitersInvestmentProposeResponse, error) // Deprecated :
	Propose(ctx context.Context, request InvestmentWaitersProposeRequest) (InvestmentWaitersProposeResponse, error)
	InvestmentRequest(ctx context.Context, request InvestmentWaitersInvestmentRequestRequest) (InvestmentWaitersInvestmentRequestResponse, error)
	Cancel(ctx context.Context, request InvestmentWaitersCancelRequest) (InvestmentWaitersCancelResponse, error)
}

type InvestmentWaitersGetAllRequest struct {
	ProjectId *int64 `query:"project_id" swagger:"desc(프로젝트 ID)"`
	controllers.SearchInput
}

// 투자대기자 DTO
type InvestmentWaitersResponse struct {
	Id              int64  `json:"id" swagger:"desc(투자 대기 신청 ID)"`
	WaitingTurn     int64  `json:"waiting_turn" swagger:"desc(대기 순번)"`
	UserName        string `json:"user_name" swagger:"desc(유저 이름)"`
	UserEmail       string `json:"user_email" swagger:"desc(유저이메일)"`
	UserMobilePhone string `json:"user_mobile_phone" swagger:"desc(유저 전화번호)"`
	Qty             int64  `json:"qty" swagger:"desc(투자 대기 신청 수량)"`
	// AdvancedQty は繰上げ済み累計数量(MIR-1068)。既存のQty(申請口数)と組み合わせて
	// 「3/10」のように併記する表示は、フロント側で組み立てる想定(一覧に列は増やさない、
	// 設計書の決定事項③)。
	AdvancedQty int64                            `json:"advanced_qty" swagger:"desc(繰上げ済み累計数量)"`
	Amount      int64                            `json:"amount" swagger:"desc(투자 대기 신청 금액)"`
	ProjectId   int64                            `json:"project_id" swagger:"desc(상품 ID)"`
	ShareId     *string                          `json:"share_id" swagger:"desc(当選後に作成されたShare ID。Share未作成の場合はnull)"`
	Status      domain.INVESTMENT_WAITING_STATUS `json:"status" swagger:"desc(表示用キャンセル待ちステータス),enum(WAITING|END_FUNDING|CANCELLED|WINNING|WINNING_AFTER_SUCCESS|WINNING_AFTER_CANCELED|WINNING_AFTER_FAILED|PROPOSAL)"`
	Type        domain.INVESTMENT_WAITING_TYPE   `json:"type" swagger:"desc(대기 상태 타입 / (구)상태)"`
	WaitingDate jsonutils.JSONTime               `json:"waiting_date" swagger:"desc(투자 대기 신청 시간)"`
}

func (i InvestmentWaitersResponse) FromInvestmentWaiters(investmentWaiters domain.InvestmentWaiters) InvestmentWaitersResponse {
	status := investmentWaiters.GetDisplayStatus()
	var shareId *string
	if investmentWaiters.ShareId != "" {
		shareId = &investmentWaiters.ShareId
	}

	return InvestmentWaitersResponse{
		Id:              investmentWaiters.Id,
		WaitingTurn:     investmentWaiters.WaitingTurn,
		UserName:        investmentWaiters.UserName,
		UserEmail:       investmentWaiters.UserEmail,
		UserMobilePhone: investmentWaiters.UserMobilePhone,
		Qty:             investmentWaiters.Qty,
		AdvancedQty:     investmentWaiters.AdvancedQty,
		Amount:          investmentWaiters.Amount,
		ProjectId:       investmentWaiters.ProjectId,
		ShareId:         shareId,
		Status:          status,
		Type:            investmentWaiters.Status, // 예전_상태 -> 현재_타입
		WaitingDate:     investmentWaiters.WaitingDate,
	}
}

type InvestmentWaitersGetAllResponse struct {
	Items      []InvestmentWaitersResponse `json:"items"`
	TotalCount int64                       `json:"total_count"`
}
type InvestmentWaitersGetOneRequest struct {
	InvestmentWaitersId *int64 `param:"id"`
}

type InvestmentWaitersGetOneResponse struct {
	InvestmentWaitersResponse
	InvestmentProposalMaxQty int64 `json:"investment_proposal_max_qty" swagger:"desc(투자제안 최대 가능구수)"`
	InvestmentProposalMinQty int64 `json:"investment_proposal_min_qty" swagger:"desc(투자제안 최소 가능구수)"`
	IsInvestmentAmountLimit  bool  `json:"is_investment_amount_limit" swagger:"desc(투자금액 제한여부)"`
	InvestmentAmountLimit    int64 `json:"investment_amount_limit" swagger:"desc(투자금액 제한금액)"`
	PricePerQty              int64 `json:"price_per_qty" swagger:"desc(구당 가격)"`
	// Winnings は当選ごとの明細(繰上げ履歴、MIR-1068)。繰上げを実行した順の新しい順。
	// omitemptyを付けない。当選が1件も無ければ空スライスのまま返し、フロント側で
	// winningsキーの有無を気にしなくてよいようにする。
	Winnings []InvestmentWaitersWinningItem `json:"winnings" swagger:"desc(繰上げ履歴)"`
}

// InvestmentWaitersWinningItem は繰上げ履歴の1明細(LotteryWinningに倣う、MIR-1068)。
type InvestmentWaitersWinningItem struct {
	ShareId            string                      `json:"investment_id" swagger:"desc(投資ID)"`
	Qty                int64                       `json:"qty" swagger:"desc(当選数量)"`
	UnitPrice          int64                       `json:"unit_price" swagger:"desc(当選単価)"`
	Amount             int64                       `json:"amount" swagger:"desc(当選金額)"`
	WinningedAt        jsonutils.JSONTime          `json:"winninged_at" swagger:"desc(当選日時)"`
	ShareStatus        domain.INVEST_STATUS        `json:"share_status" swagger:"desc(投資状態)"`
	DepositCheckStatus domain.DEPOSIT_CHECK_STATUS `json:"deposit_check_status" swagger:"desc(入金確認状態)"`
}

func (InvestmentWaitersWinningItem) FromInvestmentWaitersWinningItem(item domain.InvestmentWaitersWinningItem) InvestmentWaitersWinningItem {
	return InvestmentWaitersWinningItem{
		ShareId:            item.ShareId,
		Qty:                item.Qty,
		UnitPrice:          item.Price,
		Amount:             item.Amount,
		WinningedAt:        item.InvestDate,
		ShareStatus:        item.InvestStatus,
		DepositCheckStatus: item.DepositCheckStatus,
	}
}

type InvestmentWaitersDownloadRequest struct {
	ProjectId *int64 `query:"project_id" swagger:"desc(프로젝트 ID)"`
	controllers.SearchInput
}

type InvestmentWaitersDownloadResponse struct {
	Data []byte
}

type InvestmentWaitersStatusUpdateRequest struct {
	JWTUser             jwthelper.User
	InvestmentWaitersId *int64 `param:"id"`
	InvestmentWaitersStatusUpdateBody
}

type InvestmentWaitersStatusUpdateBody struct {
	Status domain.INVESTMENT_WAITING_TYPE `json:"status" swagger:"desc(INVESTMENT_WAITING_TYPE)" validate:"required"`
}

type InvestmentWaitersStatusUpdateResponse struct {
	InvestmentWaitersResponse
}

type InvestmentWaitersProposeRequest struct {
	JWTUser             jwthelper.User
	InvestmentWaitersId *int64 `param:"id"`
	InvestmentWaitersProposeBody
}

type InvestmentWaitersProposeBody struct {
	Qty int64 `json:"qty" swagger:"투자제안구수(프로젝트의 최소구수보다 많아야 함)" validate:"required"`
}

type InvestmentWaitersProposeResponse struct {
	InvestmentWaitersResponse
}

func (i InvestmentWaitersProposeResponse) WithInvestmentWaiters(investmentWaiters InvestmentWaitersResponse) InvestmentWaitersProposeResponse {
	i.InvestmentWaitersResponse = investmentWaiters
	return i
}

func (i InvestmentWaitersStatusUpdateResponse) WithInvestmentWaiters(investmentWaiters InvestmentWaitersResponse) InvestmentWaitersStatusUpdateResponse {
	i.InvestmentWaitersResponse = investmentWaiters
	return i
}

type InvestmentWaitersInvestmentProposeRequest struct {
	JWTUser             jwthelper.User
	InvestmentWaitersId *int64 `param:"id"`
}

type InvestmentWaitersInvestmentProposeResponse struct {
	InvestmentWaitersResponse
}

func (i InvestmentWaitersInvestmentProposeResponse) WithInvestmentWaiters(investmentWaiters InvestmentWaitersResponse) InvestmentWaitersInvestmentProposeResponse {
	i.InvestmentWaitersResponse = investmentWaiters
	return i
}

type InvestmentWaitersInvestmentRequestRequest struct {
	JWTUser             jwthelper.User
	InvestmentWaitersId *int64 `param:"id"`
	InvestmentWaitersInvestmentRequestBody
}

type InvestmentWaitersInvestmentRequestBody struct {
	Qty int64 `json:"qty" swagger:"投資提案口数(プロジェクトの最小口数以上かつ投資可能口数以内であること)" validate:"required,min=1"`
}

type InvestmentWaitersInvestmentRequestResponse struct {
	ProjectId int64 `json:"project_id" swagger:"desc(プロジェクトID)"`
	UserId    int64 `json:"user_id" swagger:"desc(ユーザーID)"`
	Qty       int64 `json:"qty" swagger:"desc(投資口数)"`
}

type InvestmentWaitersCancelRequest struct {
	JWTUser             jwthelper.User
	InvestmentWaitersId *int64 `param:"id" validate:"required"`
}
type InvestmentWaitersCancelResponse struct {
}
