package controllers

import (
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

type MyNumberStatusInput struct {
	Status domain.MY_NUMBER_STATUS `json:"status" form:"status" swagger:"desc(マイナンバー状態変更)" validate:"oneof=APPROVE DISAPPROVED WAIT_APPROVE UNREGISTERED"`
}

func (m *MyNumberStatusInput) ToModel() (*domain.MyNumber, error) {
	return &domain.MyNumber{
		Status:       m.Status,
		ApprovalDate: jsonutils.JSONTime(time.Now()),
	}, nil
}

type MyNumberSearchInput struct {
	Status            *string `query:"status" swagger:"desc(マイナンバー状態)"`
	UserId            *string `query:"user_id" swagger:"desc(ユーザーID)"`
	MyNumber          *string `query:"my_number" swagger:"desc(マイナンバー)"`
	ApprovalStartDate string  `query:"approval_start_date" swagger:"desc(承認開始日)"`
	ApprovalEndDate   string  `query:"approval_end_date" swagger:"desc(承認終了日)"`
	RequestStartDate  string  `query:"request_start_date" swagger:"desc(リクエスト開始日)"`
	RequestEndDate    string  `query:"request_end_date" swagger:"desc(リクエスト終了日)"`
	SearchInput
}

func (m *MyNumberSearchInput) ToModel() *domain.MyNumberSearchOption {

	return &domain.MyNumberSearchOption{
		Status:            m.Status,
		UserId:            m.UserId,
		ApprovalStartDate: m.ApprovalStartDate,
		ApprovalEndDate:   m.ApprovalEndDate,
		RequestStartDate:  m.RequestStartDate,
		RequestEndDate:    m.RequestEndDate,
	}
}

type DuplicatedMyNumberInfo struct {
	Id     int64                   `json:"id" swagger:"desc(ID)"`
	UserId int64                   `json:"user_id" swagger:"desc(ユーザーID)"`
	Status domain.MY_NUMBER_STATUS `json:"status" swagger:"desc(マイナンバー承認状態: 共通コードMY_NUMBER_STATUS)"`
}

type DuplicatedMyNumberResponse struct {
	HasDuplicate bool                     `json:"has_duplicate" swagger:"desc(重複の有無)"`
	TotalCount   int                      `json:"total_count" swagger:"desc(重複したマイナンバーの数)"`
	MyNumber     []DuplicatedMyNumberInfo `json:"my_number" swagger:"desc(マイナンバーリスト)"`
}

func (DuplicatedMyNumberResponse) New(myNumbers []domain.MyNumber) DuplicatedMyNumberResponse {
	var myNumberInfo []DuplicatedMyNumberInfo
	for _, myNumber := range myNumbers {
		myNumberInfo = append(myNumberInfo, DuplicatedMyNumberInfo{
			Id:     myNumber.Id,
			UserId: myNumber.UserId,
			Status: myNumber.Status,
		})
	}

	return DuplicatedMyNumberResponse{
		HasDuplicate: len(myNumbers) > 0,
		TotalCount:   len(myNumbers),
		MyNumber:     myNumberInfo,
	}
}
