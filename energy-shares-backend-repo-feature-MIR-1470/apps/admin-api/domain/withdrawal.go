package domain

import (
	"context"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
)

type WithdrawalUsecase interface {
	Create(ctx context.Context, request WithdrawalCreateRequest) (WithdrawalCreateResponse, error)
	GetAll(ctx context.Context, request WithdrawalGetAllRequest) (WithdrawalGetAllResponse, error)
	GetOne(ctx context.Context, request WithdrawalGetOneRequest) (WithdrawalGetOneResponse, error)
	UpdateStatus(ctx context.Context, request WithdrawalUpdateStatusRequest) (WithdrawalUpdateStatusResponse, error)
	RequestTransfer(ctx context.Context, request WithdrawalRequestTransferRequest) (WithdrawalRequestTransferResponse, error)
}

type WithdrawalCreateRequest struct {
	JWTUser  jwthelper.User
	ClientIp string

	UserId int64 `json:"user_id" description:"出金依頼を作成する対象ユーザー ID" required:"true" validate:"gt=0" format:"int64"`
	Amount int64 `json:"amount" description:"出金依頼金額" required:"true" validate:"gt=0" format:"int64"`
}

type WithdrawalCreateResponse struct {
	Id           int64                         `json:"id" description:"出金依頼 ID" format:"int64"`
	InvestorId   string                        `json:"investor_id" description:"投資家 ID" format:"string"`
	InvestorName string                        `json:"investor_name" description:"投資家名" format:"string"`
	Amount       int64                         `json:"amount" description:"出金依頼金額" format:"int64"`
	Fee          int64                         `json:"fee" description:"振込手数料" format:"int64"`
	RequestDate  jsonutils.JSONTime            `json:"request_date" description:"出金依頼日時" format:"yyyy-MM-dd HH:mm:ss"`
	Status       domain.WITHDRAWAL_STATUS_TYPE `json:"status" description:"出金依頼状態" format:"string" enum:"INCOMPLETE,WAIT,COMPLETE,FAIL,CANCEL"`
	Created      jsonutils.JSONTime            `json:"created" description:"作成日時" format:"yyyy-MM-dd HH:mm:ss"`
	Updated      jsonutils.JSONTime            `json:"updated" description:"更新日時" format:"yyyy-MM-dd HH:mm:ss"`
}

type WithdrawalGetAllRequest struct {
	JWTUser jwthelper.User
	WithdrawalGetAllQueryParam
}

type WithdrawalGetAllResponse struct {
	Items                []WithdrawalItem `json:"items" description:"출금의뢰 리스트"`
	TotalCount           int64            `json:"total_count" description:"총 개수" format:"int64" example:"1"`
	TotalIncompleteCount int64            `json:"total_incomplete_count" description:"미완료 개수" format:"int64" example:"1"`
}

type WithdrawalGetAllQueryParam struct {
	cdomain.SearchInput
	InvestorId        *string                        `query:"investor_id" description:"투자자 ID" format:"string"`
	InvestorName      *string                        `query:"investor_name" description:"투자자명" format:"string"`
	AmountMin         *int64                         `query:"amount_min" description:"출금의뢰 최소 금액" format:"int64"`
	AmountMax         *int64                         `query:"amount_max" description:"출금의뢰 최대 금액" format:"int64"`
	RequestStartDate  *string                        `query:"request_start_date" description:"출금의뢰 요청일시 검색 최소 날짜" validate:"omitempty,datetime=2006-01-02 15:04:05" format:"yyyy-MM-dd HH:mm:ss"`
	RequestEndDate    *string                        `query:"request_end_date" description:"출금의뢰 요청일시 검색 최대 날짜" validate:"omitempty,datetime=2006-01-02 15:04:05" format:"yyyy-MM-dd HH:mm:ss"`
	CompleteStartDate *string                        `query:"complete_start_date" description:"출금의뢰 완료일시 검색 최소 날짜" validate:"omitempty,datetime=2006-01-02 15:04:05" format:"yyyy-MM-dd HH:mm:ss"`
	CompleteEndDate   *string                        `query:"complete_end_date" description:"출금의뢰 완료일시 검섹 최대 날짜" validate:"omitempty,datetime=2006-01-02 15:04:05" format:"yyyy-MM-dd HH:mm:ss"`
	Status            *domain.WITHDRAWAL_STATUS_TYPE `query:"status" description:"출금의뢰 상태" validate:"omitempty,oneof=INCOMPLETE WAIT COMPLETE FAIL CANCEL" enum:"INCOMPLETE,WAIT,COMPLETE,FAIL,CANCEL"`
}

type WithdrawalGetOneRequest struct {
	JWTUser jwthelper.User
	Id      int64 `param:"id" path:"id" description:"출금의뢰 번호" format:"int64"`
}

type WithdrawalGetOneResponse struct {
	WithdrawalDetailItem
	TransferLogs []WithdrawalTransferItem `json:"transfer_logs" description:"송금 처리 내역"`
}

type WithdrawalUpdateStatusRequest struct {
	JWTUser jwthelper.User
	Id      int64                         `param:"id" path:"id" description:"출금의뢰 번호" format:"int64"`
	Status  domain.WITHDRAWAL_STATUS_TYPE `json:"status" description:"출금의뢰 상태" format:"string" enum:"INCOMPLETE,WAIT,COMPLETE,FAIL,CANCEL"`
}

type WithdrawalUpdateStatusResponse struct {
	WithdrawalDetailItem
}

type WithdrawalRequestTransferRequest struct {
	JWTUser jwthelper.User
	Id      int64 `param:"id" path:"id" description:"출금의뢰 번호" format:"int64"`
}

type WithdrawalRequestTransferResponse struct {
	Id            int64  `json:"id" description:"출금의뢰-송금 맵핑 ID" format:"int64"`
	TransactionId string `json:"transaction_id" description:"송금 거래 ID" format:"string"`
	WithdrawalId  int64  `json:"withdrawal_id" description:"출금의뢰 ID" format:"int64"`
}

type WithdrawalItem struct {
	Id           int64                         `json:"id" description:"출금의뢰 ID" format:"int64"`
	InvestorId   string                        `json:"investor_id" description:"투자자 ID" format:"string"`
	InvestorName string                        `json:"investor_name" description:"투자자명" format:"string"`
	Amount       int64                         `json:"amount" description:"출금의뢰 금액" format:"int64"`
	Fee          int64                         `json:"fee" description:"이체 수수료" format:"int64"`
	RequestDate  jsonutils.JSONTime            `json:"request_date" description:"출금의뢰 요청일시" format:"yyyy-MM-dd HH:mm:ss"`
	CompleteDate jsonutils.JSONTime            `json:"complete_date" description:"출금의뢰 완료일시" format:"yyyy-MM-dd HH:mm:ss"`
	CloseDate    jsonutils.JSONTime            `json:"close_date" description:"출금의뢰 마감일시" format:"yyyy-MM-dd HH:mm:ss"`
	Status       domain.WITHDRAWAL_STATUS_TYPE `json:"status" description:"출금의뢰 상태" format:"string" enum:"INCOMPLETE,WAIT,COMPLETE,FAIL,CANCEL"`
	IsManual     bool                          `json:"is_manual" description:"手動出金かどうか" format:"boolean"`
	Created      jsonutils.JSONTime            `json:"created" description:"생성일" format:"yyyy-MM-dd HH:mm:ss"`
}

type WithdrawalDetailItem struct {
	WithdrawalItem
	BankName    string              `json:"bank_name" description:"은행명" format:"string"`
	BranchName  string              `json:"branch_name" description:"지점명" format:"string"`
	DepositType domain.DEPOSIT_TYPE `json:"deposit_type" description:"계좌 유형" format:"string" enum:"ORDINARY_ACCOUNT,CHECKING_ACCOUNT"`
	BankAccount string              `json:"bank_account" description:"계좌 번호" format:"string"`
	UserType    domain.USER_TYPE    `json:"user_type" description:"유저 타입" format:"string" enum:"PRIVATE,CORPORATE"`
	BranchCode  *string             `json:"branch_code" description:"은행 지점 코드" format:"string"`
}

type WithdrawalTransferItem struct {
	Id             int64                       `json:"id" description:"송금 번호" format:"int64"`
	RequestId      int64                       `json:"request_id" description:"송금 요청 ID" format:"int64"`
	TransactionId  string                      `json:"transaction_id" description:"송금 거래 ID" format:"string"`
	TransferDate   jsonutils.JSONTime          `json:"transfer_date" description:"송금 요청 일시" format:"yyyy-MM-dd HH:mm:ss"`
	TransferAmount int64                       `json:"transfer_amount" description:"송금 금액" format:"int64"`
	TransferFee    int64                       `json:"transfer_fee" description:"송금 수수료" format:"int64"`
	TransferStatus domain.TRANSFER_STATUS_TYPE `json:"transfer_status" description:"송금 상태" format:"string" enum:"TRANSFER_STATUS_WAIT,TRANSFER_STATUS_SUCCESS,TRANSFER_STATUS_FAIL,TRANSFER_STATUS_REPAYMENT,TRANSFER_STATUS_CANCEL"`
	ErrorCode      string                      `json:"error_code" description:"송금 실패 코드" format:"string"`
	ErrorMsg       string                      `json:"error_msg" description:"송금 실패 메시지" format:"string"`
}
