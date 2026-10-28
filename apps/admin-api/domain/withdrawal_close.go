package domain

import (
	"context"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
)

type WithdrawalCloseUsecase interface {
	GetAll(ctx context.Context, request WithdrawalCloseGetAllRequest) (WithdrawalCloseGetAllResponse, error)
	GetWithdrawals(ctx context.Context, request WithdrawalCloseGetWithdrawalsRequest) (WithdrawalCloseGetWithdrawalsResponse, error)
	Download(ctx context.Context, request WithdrawalCloseDownloadRequest) (WithdrawalCloseDownloadResponse, error)
	RequestTransfer(ctx context.Context, request WithdrawalCloseRequestTransferRequest) (WithdrawalCloseRequestTransferResponse, error)
	CloseAllWithdrawals(ctx context.Context, request WithdrawalCloseCloseAllWithdrawalsRequest) (WithdrawalCloseCloseAllWithdrawalsResponse, error)
}

type WithdrawalCloseGetAllRequest struct {
	JWTUser jwthelper.User
	cdomain.SearchInput
}

type WithdrawalCloseGetAllResponse struct {
	Items      []WithdrawalCloseGetAllItem `json:"items" description:"출금의뢰 마감 목록"`
	TotalCount int64                       `json:"total_count" description:"총 개수" format:"int64"`
}

type WithdrawalCloseGetAllItem struct {
	Id           int64                                 `json:"id" description:"마감 ID" format:"int64"`
	Total        int64                                 `json:"total" description:"총 출금의뢰 수" format:"int64"`
	Amount       int64                                 `json:"amount" description:"총 출금의뢰 금액" format:"int64"`
	Fee          int64                                 `json:"fee" description:"총 수수료" format:"int64"`
	CloserId     int64                                 `json:"closer_id" description:"마감자 ID" format:"int64"`
	CloserName   string                                `json:"closer_name" description:"마감자명" format:"string"`
	CloseDate    jsonutils.JSONTime                    `json:"close_date" description:"마감일" format:"yyyy-MM-dd HH:mm:ss"`
	TransferDate jsonutils.JSONTime                    `json:"transfer_date" description:"송금일" format:"yyyy-MM-dd HH:mm:ss"`
	Status       domain.WITHDRAWAL_REQUEST_STATUS_TYPE `json:"status" description:"출금의뢰 상태" format:"string" enum:"WAIT COMPLETE SENDING"`
}

type WithdrawalCloseGetWithdrawalsRequest struct {
	JWTUser jwthelper.User
	Id      int64 `param:"id" path:"id" description:"출금의뢰 마감 ID" format:"int64"`
	cdomain.SearchInput
}

type WithdrawalCloseGetWithdrawalsResponse struct {
	Items      []WithdrawalCloseGetWithdrawalsItem `json:"items" description:"출금의뢰 목록"`
	TotalCount int64                               `json:"total_count" description:"총 개수" format:"int64"`
}

type WithdrawalCloseGetWithdrawalsItem struct {
	Id           int64                         `json:"id" description:"출금의뢰 ID" format:"int64"`
	InvestorId   string                        `json:"investor_id" description:"투자자 ID" format:"string"`
	InvestorName string                        `json:"investor_name" description:"투자자명" format:"string"`
	Amount       int64                         `json:"amount" description:"출금의뢰 금액" format:"int64"`
	Fee          int64                         `json:"fee" description:"이체 수수료" format:"int64"`
	RequestDate  jsonutils.JSONTime            `json:"request_date" description:"출금의뢰 요청일시" format:"yyyy-MM-dd HH:mm:ss"`
	CompleteDate jsonutils.JSONTime            `json:"complete_date" description:"출금의뢰 완료일시" format:"yyyy-MM-dd HH:mm:ss"`
	Status       domain.WITHDRAWAL_STATUS_TYPE `json:"status" description:"출금의뢰 상태" format:"string" enum:"INCOMPLETE,WAIT,COMPLETE,FAIL,CANCEL"`
	Created      jsonutils.JSONTime            `json:"created" description:"생성일" format:"yyyy-MM-dd HH:mm:ss"`
	CloseDate    jsonutils.JSONTime            `json:"close_date" description:"출금의뢰 마감일시" format:"yyyy-MM-dd HH:mm:ss"`
}

type WithdrawalCloseDownloadRequest struct {
	JWTUser  jwthelper.User
	ClientIp string
	Id       int64 `param:"id" path:"id" description:"출금의뢰 마감 ID" format:"int64"`
	cdomain.SearchInput
}

type WithdrawalCloseDownloadResponse struct {
	Data []byte `json:"data" description:"다운로드 데이터"`
}

type WithdrawalCloseRequestTransferRequest struct {
	JWTUser jwthelper.User
	Id      int64 `param:"id" path:"id" description:"출금의뢰 마감 ID" format:"int64"`
}

type WithdrawalCloseRequestTransferResponse struct {
	TotalCount int64 `json:"total_count" description:"총 개수" format:"int64"`
}

type WithdrawalCloseCloseAllWithdrawalsRequest struct {
	JWTUser jwthelper.User
}

type WithdrawalCloseCloseAllWithdrawalsResponse struct {
	TotalCount int64 `json:"total_count" description:"총 개수" format:"int64"`
}
