package controllers

import (
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

type TransferDto struct {
	Id             int64                       `json:"id" swagger:"desc(ID)"`
	RequestId      int64                       `json:"request_id" swagger:"desc(transfer request ID)"`
	TransactionId  string                      `json:"transaction_id" swagger:"desc(transaction ID)"`
	TransferDate   jsonutils.JSONTime          `json:"transfer_date" swagger:"desc(送金リクエスト日時)"`
	TransferAmount int64                       `json:"transfer_amount" swagger:"desc(送金金額)"`
	TransferFee    int64                       `json:"transfer_fee" swagger:"desc(送金手数料)"`
	TransferStatus domain.TRANSFER_STATUS_TYPE `json:"transfer_status" swagger:"desc(送金状態)"`
	ErrorCode      string                      `json:"error_code" swagger:"desc(送金失敗コード)"`
	ErrorMsg       string                      `json:"error_msg" swagger:"desc(送金失敗メッセージ)"`
}

func (t *TransferDto) New(transfer *domain.Transfer) *TransferDto {
	return &TransferDto{
		Id:             transfer.Id,
		RequestId:      transfer.RequestId,
		TransactionId:  transfer.TransactionId,
		TransferDate:   transfer.TransferDate,
		TransferAmount: transfer.TransferAmount,
		TransferFee:    transfer.TransferFee,
		TransferStatus: transfer.TransferStatus,
		ErrorCode:      transfer.ErrorCode,
		ErrorMsg:       transfer.ErrorMsg,
	}
}

func (t *TransferDto) NewList(transferList []domain.Transfer) []*TransferDto {
	var list []*TransferDto
	for _, v := range transferList {
		list = append(list, t.New(&v))
	}

	return list
}
