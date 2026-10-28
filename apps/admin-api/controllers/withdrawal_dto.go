package controllers

import (
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

type WithdrawalSearchInput struct {
	SearchInput
	InvestorId        *string                       `query:"investor_id" swagger:"desc(投資家ID)"`
	Orderer           *string                       `query:"orderer" swagger:"desc(投資家)" validate:"required"`
	AmountStart       *int                          `query:"amount_start" swagger:"desc(出金リクエスト最小金額)"`
	AmountEnd         *int                          `query:"amount_end" swagger:"desc(出金リクエスト最大金額)"`
	RequestStartDate  string                        `query:"request_start_date" swagger:"desc(出金依頼リクエスト開始日)"`
	RequestEndDate    string                        `query:"request_end_date" swagger:"desc(出金依頼リクエスト終了日)"`
	CompleteStartDate string                        `query:"complete_start_date" swagger:"desc(出金依頼完了開始日)"`
	CompleteEndDate   string                        `query:"complete_end_date" swagger:"desc(出金依頼完了終了日)"`
	Status            domain.WITHDRAWAL_STATUS_TYPE `query:"status" swagger:"enum(INCOMPLETE|WAIT|COMPLETE|FAIL),default(INCOMPLETE),desc(出金依頼状態)"`
}

type WithdrawalDto struct {
	Id           int64                         `json:"id" swagger:"desc(出金番号)"`
	InvestorId   string                        `json:"investor_id" swagger:"desc(投資家ID)"`
	Orderer      string                        `json:"orderer" swagger:"desc(投資家名)"`
	Amount       int64                         `json:"amount" swagger:"desc(出金リクエスト金額)"`
	Fee          int64                         `json:"fee" swagger:"desc(送金手数料)"`
	RequestDate  jsonutils.JSONTime            `json:"request_date" swagger:"desc(リクエスト日)"`
	CompleteDate jsonutils.JSONTime            `json:"complete_date" swagger:"desc(完了日)"`
	CloseDate    jsonutils.JSONTime            `json:"close_date" swagger:"desc(締切日)"`
	Status       domain.WITHDRAWAL_STATUS_TYPE `json:"status" swagger:"desc(ステータス)"`
	Created      jsonutils.JSONTime            `json:"created" swagger:"desc(作成日)"`
	Updated      jsonutils.JSONTime            `json:"updated" swagger:"desc(更新日)"`
	Deleted      jsonutils.JSONTime            `json:"deleted" swagger:"desc(削除時に Deleted に時間が入力されると照会されない。1年後に削除できるようにすることができる)"`
}

func (w WithdrawalDto) New(m *domain.Withdrawal) *WithdrawalDto {
	return &WithdrawalDto{
		Id:           m.Id,
		InvestorId:   m.InvestorId,
		Orderer:      m.InvestorName,
		Amount:       m.Amount,
		Fee:          m.Fee,
		RequestDate:  m.RequestDate,
		CompleteDate: m.CompleteDate,
		CloseDate:    m.CloseDate,
		Status:       m.Status,
		Created:      m.Created,
		Updated:      m.Updated,
		Deleted:      m.Deleted,
	}
}

type WithdrawalDetailDto struct {
	Id           int64                         `json:"id" swagger:"desc(出金番号)"`
	InvestorId   string                        `json:"investor_id" swagger:"desc(投資家ID)"`
	Orderer      string                        `json:"orderer" swagger:"desc(投資家名)"`
	Amount       int64                         `json:"amount" swagger:"desc(出金リクエスト金額)"`
	Fee          int64                         `json:"fee" swagger:"desc(送金手数料)"`
	RequestDate  jsonutils.JSONTime            `json:"request_date" swagger:"desc(リクエスト日)"`
	CompleteDate jsonutils.JSONTime            `json:"complete_date" swagger:"desc(完了日)"`
	CloseDate    jsonutils.JSONTime            `json:"close_date" swagger:"desc(締切日)"`
	Status       domain.WITHDRAWAL_STATUS_TYPE `json:"status" swagger:"desc(ステータス)"`
	IsManual     bool                          `json:"is_manual" swagger:"desc(手動出金かどうか)"`
	BankName     string                        `json:"bank_name" swagger:"desc(銀行名)"`
	BankCode     *string                       `json:"bank_code" swagger:"desc(銀行コード)"`
	BranchName   string                        `json:"branch_name" swagger:"desc(支店名)"`
	DepositType  domain.DEPOSIT_TYPE           `json:"deposit_type" swagger:"desc(預金種類)"`
	BankAccount  string                        `json:"bank_account" swagger:"desc(口座番号)"`
	UserType     domain.USER_TYPE              `json:"user_type" swagger:"desc(預金者区分)"`
	BranchCode   *string                       `json:"branch_code" swagger:"desc(銀行支店コード)"`
	Transfer     []*TransferDto                `json:"transfer" swagger:"desc(送金内容)"`
	// MIR-1170: 失敗の理由。出金依頼がFAIL/CANCELで、理由を出す対象のときだけ入る。
	FailureKind    string `json:"failure_kind,omitempty" swagger:"desc(失敗の理由の種類)"`
	FailureMessage string `json:"failure_message,omitempty" swagger:"desc(失敗の理由(画面にそのまま出す))"`
}

func (w WithdrawalDetailDto) New(m *domain.WithdrawalDetail) *WithdrawalDetailDto {
	return &WithdrawalDetailDto{
		Id:           m.Withdrawal.Id,
		InvestorId:   m.InvestorId,
		Orderer:      m.InvestorName,
		Amount:       m.Amount,
		Fee:          m.Fee,
		RequestDate:  m.RequestDate,
		CompleteDate: m.CompleteDate,
		CloseDate:    m.CloseDate,
		Status:       m.Status,
		IsManual:     m.IsManual,
		BankName:     m.BankName,
		BankCode:     m.BankCode,
		BranchName:   m.BranchName,
		DepositType:  m.DepositType,
		BankAccount:  m.BankAccount,
		UserType:     m.UserType,
		BranchCode:   m.BranchCode,
		Transfer:     nil,
	}
}

func (w WithdrawalDetailDto) NewList(m []domain.WithdrawalDetail) []*WithdrawalDetailDto {
	var list []*WithdrawalDetailDto
	for _, v := range m {
		list = append(list, w.New(&v))
	}
	return list
}

type WithdrawalCloseDetailDto struct {
	Id           int64                                 `json:"id" swagger:"desc(締切ID)"`
	Total        int64                                 `json:"total" swagger:"desc(総出金依頼数)"`
	Amount       int64                                 `json:"amount" swagger:"desc(総出金金額)"`
	Fee          int64                                 `json:"fee" swagger:"desc(総手数料)"`
	CloserId     int64                                 `json:"closer_id" swagger:"desc(締切者ID)"`
	CloserName   string                                `json:"closer_name" swagger:"desc(締切者名)"`
	CloseDate    jsonutils.JSONTime                    `json:"close_date" swagger:"desc(締切日)"`
	TransferDate jsonutils.JSONTime                    `json:"transfer_date" swagger:"desc(送金日)"`
	Status       domain.WITHDRAWAL_REQUEST_STATUS_TYPE `json:"status" swagger:"enum(WAIT|COMPLETE|SENDING),default(WAIT),desc(送金状態)"`
}

func (w WithdrawalCloseDetailDto) New(m *domain.WithdrawalCloseDetail) *WithdrawalCloseDetailDto {
	return &WithdrawalCloseDetailDto{
		Id:           m.Id,
		Total:        m.Total,
		Amount:       m.Amount,
		Fee:          m.Fee,
		CloserId:     m.CloserId,
		CloserName:   m.CloserName,
		CloseDate:    m.CloseDate,
		TransferDate: m.TransferDate,
		Status:       m.Status,
	}
}

func (w WithdrawalCloseDetailDto) NewList(m []domain.WithdrawalCloseDetail) []*WithdrawalCloseDetailDto {
	var list []*WithdrawalCloseDetailDto
	for _, v := range m {
		list = append(list, w.New(&v))
	}
	return list
}
