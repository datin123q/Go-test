package domain

import (
	"context"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
)

type InvestmentUsecase interface {
	GetResult(ctx context.Context, request InvestmentGetResultRequest) (InvestmentGetResultResponse, error)
}

type InvestmentGetResultRequest struct {
	JWTUser      jwthelper.User
	InvestmentId string `param:"id"`
}

type InvestmentGetResultResponse struct {
	InvestmentId      string             `json:"investment_id" swagger:"desc(取引番号)"`
	ProjectName       string             `json:"project_name" swagger:"desc(ファンド名)"`
	InvestmentDate    jsonutils.JSONTime `json:"investment_date" swagger:"desc(取引日時)"`
	BankName          string             `json:"bank_name" swagger:"desc(銀行名)"`
	BankAccountNumber string             `json:"bank_account_number" swagger:"desc(口座番号)"`
	Depositor         string             `json:"depositor" swagger:"desc(預金者)"`
	Qty               int64              `json:"qty" swagger:"desc(投資数量)"`
	Price             int64              `json:"price" swagger:"desc(投資単価)"`
	Amount            int64              `json:"amount" swagger:"desc(投資金額)"`
}
