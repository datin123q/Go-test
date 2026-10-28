package controllers

import (
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

type WithdrawalInput struct {
	Amount int64 `json:"amount" form:"amount" swagger:"desc(出金要求金額)"`
}

func (ni *WithdrawalInput) ToModel() (*domain.Withdrawal, error) {

	return &domain.Withdrawal{
		Amount:   ni.Amount,
		IsManual: false,
	}, nil
}

type WithdrawalDto struct {
	Id          int64                         `json:"id" swagger:"desc(出金番号)"`
	InvestorId  string                        `json:"investor_id" swagger:"desc(投資家ID)"`
	Orderer     string                        `json:"orderer" swagger:"desc(投資家名)"`
	Amount      int64                         `json:"amount" swagger:"desc(出金要求金額)"`
	Status      domain.WITHDRAWAL_STATUS_TYPE `json:"status" swagger:"desc(ステータス)"`
	RequestDate jsonutils.JSONTime            `json:"request_date" swagger:"desc(リクエスト日付)"`
	Created     jsonutils.JSONTime            `json:"created" swagger:"desc(作成日付)"`
	Updated     jsonutils.JSONTime            `json:"updated" swagger:"desc(更新日付)"`
}

func (w WithdrawalDto) New(m *domain.Withdrawal) *WithdrawalDto {
	return &WithdrawalDto{
		Id:          m.Id,
		InvestorId:  m.InvestorId,
		Orderer:     m.InvestorName,
		Amount:      m.Amount,
		Status:      m.Status,
		RequestDate: m.RequestDate,
		Created:     m.Created,
		Updated:     m.Updated,
	}
}

func (w WithdrawalDto) NewByDetail(m *domain.WithdrawalDetail) *WithdrawalDto {
	return &WithdrawalDto{
		Id:          m.Withdrawal.Id,
		InvestorId:  m.InvestorId,
		Orderer:     m.InvestorName,
		Amount:      m.Amount,
		Status:      m.Status,
		RequestDate: m.RequestDate,
		Created:     m.Withdrawal.Created,
		Updated:     m.Withdrawal.Updated,
	}
}

func (w WithdrawalDto) NewList(m []domain.Withdrawal) []*WithdrawalDto {
	var list []*WithdrawalDto
	for _, v := range m {
		list = append(list, w.New(&v))
	}
	return list
}

func (w WithdrawalDto) NewListByDetail(m []domain.WithdrawalDetail) []*WithdrawalDto {
	var list []*WithdrawalDto
	for _, v := range m {
		list = append(list, w.NewByDetail(&v))
	}
	return list
}

type WithdrawalDetailDto struct {
	Id              int64                         `json:"id" swagger:"desc(出金番号)"`
	InvestorId      string                        `json:"investor_id" swagger:"desc(投資家ID)"`
	Orderer         string                        `json:"orderer" swagger:"desc(投資家名)"`
	Depositor       string                        `json:"depositor" swagger:"desc(預金者名)"`
	Amount          int64                         `json:"amount" swagger:"desc(出金要求金額)"`
	Status          domain.WITHDRAWAL_STATUS_TYPE `json:"status" swagger:"desc(ステータス)"`
	Fee             int64                         `json:"fee" swagger:"desc(送金手数料)"`
	BankName        string                        `json:"bank_name" swagger:"desc(銀行名)"`
	BankCode        *string                       `json:"bank_code" swagger:"desc(銀行コード)"`
	BranchName      string                        `json:"branch_name" swagger:"desc(支店名)"`
	DepositType     domain.DEPOSIT_TYPE           `json:"deposit_type" swagger:"desc(預金種類)"`
	BankAccount     string                        `json:"bank_account" swagger:"desc(口座番号)"`
	BranchCode      *string                       `json:"branch_code" swagger:"desc(銀行支店コード)"`
	RequestDate     jsonutils.JSONTime            `json:"request_date" swagger:"desc(リクエスト日付)"`
	CompleteDate    jsonutils.JSONTime            `json:"complete_date" swagger:"desc(完了日付)"`
	Created         jsonutils.JSONTime            `json:"created" swagger:"desc(作成日付)"`
	Updated         jsonutils.JSONTime            `json:"updated" swagger:"desc(更新日付)"`
	ExpectedBalance int64                         `json:"expected_balance" swagger:"desc(予想残高)"`
	CancelledReason *string                       `json:"cancelled_reason" swagger:"desc(キャンセル理由)"`
}

func (w WithdrawalDetailDto) New(m *domain.WithdrawalDetail) *WithdrawalDetailDto {
	return &WithdrawalDetailDto{
		Id:              m.Withdrawal.Id,
		InvestorId:      m.InvestorId,
		Orderer:         m.InvestorName,
		Depositor:       m.Depositor,
		Amount:          m.Amount,
		Status:          m.Status,
		Fee:             m.Fee,
		BankName:        m.BankName,
		BankCode:        m.BankCode,
		BranchName:      m.BranchName,
		DepositType:     m.DepositType,
		BankAccount:     m.BankAccount,
		BranchCode:      m.BranchCode,
		RequestDate:     m.RequestDate,
		CompleteDate:    m.CompleteDate,
		Created:         m.Withdrawal.Created,
		Updated:         m.Withdrawal.Updated,
		ExpectedBalance: 0,
		CancelledReason: nil,
	}
}

func (w WithdrawalDetailDto) NewList(m []domain.WithdrawalDetail) []*WithdrawalDetailDto {
	var list []*WithdrawalDetailDto
	for _, v := range m {
		list = append(list, w.New(&v))
	}
	return list
}
