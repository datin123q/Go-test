package controllers

import (
	"github.com/illuminarean/energy-shares-backend-repo/service"
)

type JapanBankSearchRequest struct {
	Keyword string  `query:"keyword" swagger:"desc(銀行名)" validate:"required"`
	Cursor  *string `query:"cursor" swagger:"desc(次ページカーソル)"`
	Limit   int64   `query:"limit" swagger:"desc(照会最大数、デフォルト: 2000)"`
}

type JapanBankBranchSearchRequest struct {
	Keyword string  `query:"keyword" swagger:"desc(支店名)" validate:"required"`
	Cursor  *string `query:"cursor" swagger:"desc(次ページカーソル)"`
	Limit   int64   `query:"limit" swagger:"desc(照会最大数、デフォルト: 2000)"`
}

type JapanBank struct {
	Code          string `json:"code" swagger:"desc(銀行コード番号)"`
	Name          string `json:"name" swagger:"desc(銀行名)"`
	HalfWidthKana string `json:"halfWidthKana" swagger:"desc(銀行名半角カタカナ)"`
}
type JapanBankResponse struct {
	Items      []JapanBank `json:"items" swagger:"desc(銀行リスト)"`
	TotalCount int64       `json:"total_count" swagger:"desc(銀行リスト総数)"`
	HasNext    bool        `json:"hasNext" swagger:"desc(次ページ存在有無)"`
	NextCursor *string     `json:"nextCursor" swagger:"desc(次ページカーソル)"`
}

func (jbr JapanBankResponse) New(bankResult service.JapanBankResult) JapanBankResponse {
	var list []JapanBank
	for _, b := range bankResult.Banks {
		bank := JapanBank{
			Code:          b.Code,
			Name:          b.Name,
			HalfWidthKana: b.HalfWidthKana,
		}
		list = append(list, bank)
	}

	return JapanBankResponse{
		Items:      list,
		TotalCount: bankResult.Size,
		HasNext:    bankResult.HasNext,
		NextCursor: bankResult.NextCursor,
	}
}

type JapanBankBranch struct {
	Code          string `json:"code" swagger:"desc(支店コード番号)"`
	Name          string `json:"name" swagger:"desc(支店名)"`
	HalfWidthKana string `json:"halfWidthKana" swagger:"desc(支店名半角カタカナ)"`
}

type JapanBankBranchResponse struct {
	Items      []JapanBankBranch `json:"items" swagger:"desc(支店リスト)"`
	TotalCount int64             `json:"total_count" swagger:"desc(支店リスト総数)"`
	HasNext    bool              `json:"hasNext" swagger:"desc(次ページ存在有無)"`
	NextCursor *string           `json:"nextCursor" swagger:"desc(次ページカーソル)"`
}

func (jbr JapanBankBranchResponse) New(branchResult service.JapanBankBranchResult) JapanBankBranchResponse {
	var list []JapanBankBranch
	for _, b := range branchResult.Branches {
		branch := JapanBankBranch{
			Code:          b.Code,
			Name:          b.Name,
			HalfWidthKana: b.HalfWidthKana,
		}
		list = append(list, branch)
	}

	return JapanBankBranchResponse{
		Items:      list,
		TotalCount: branchResult.Size,
		HasNext:    branchResult.HasNext,
		NextCursor: branchResult.NextCursor,
	}
}
