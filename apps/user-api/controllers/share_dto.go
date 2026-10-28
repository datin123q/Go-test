package controllers

import "github.com/illuminarean/energy-shares-backend-repo/domain"

type ShareInput struct {
	ProjectId    int64                   `json:"project_id"  swagger:"desc(ファンドプロジェクトID)"  validate:"required"`
	CategoryType domain.PROJECT_CATEGORY `json:"category_type"  swagger:"enum(SOLAR|WIND|BIOMASS),desc(ファンドカテゴリ 1:太陽光 2:風力 3:バイオマス)"   validate:"required"`
	ProjectType  domain.PROJECT_TYPE     `json:"project_type" swagger:"enum(LOAN|EQUITY),desc(1:ローン型 2:ビジネス型)" validate:"required"`
	InvestStatus domain.INVEST_STATUS    `json:"order_status" swagger:"enum(REQUEST_INVEST|PROCESS_INVEST|SUCCESS_INVEST|CANCEL_INVEST),desc(1:注文リクエスト 2:注文処理中 3:注文成功 4:注文キャンセル)" validate:"required"`
	OwnerId      int64                   `json:"owner_id" swagger:"desc(購入者)" validate:"required"`
	Desc         string                  `json:"desc" swagger:"desc(説明)" validate:"required"`
	Amount       int64                   `json:"amount" swagger:"desc(購入金額)" validate:"required"`
	Qty          int64                   `json:"qty" swagger:"desc(数量)" validate:"required"`
	Price        int64                   `json:"price" swagger:"desc(1株あたりの価格)" validate:"required"`
}

type ShareCheckInput struct {
	UserId       int64 `json:"user_id" swagger:"desc(ユーザーID)" validate:"required"`
	ProjectId    int64 `json:"project_id" swagger:"desc(ファンドID)" validate:"required"`
	Qty          int64 `json:"qty" swagger:"desc(投資口数)" validate:"required"`
	IsWaitInvest bool  `json:"is_wait_invest" swagger:"desc(キャンセル待ち状態かどうか)"`
}

type ShareInvestInput struct {
	UserId    int64 `json:"user_id" swagger:"desc(ユーザーID)"  validate:"required"`
	ProjectId int64 `json:"project_id" swagger:"desc(ファンドプロジェクトID)" validate:"required"`
	Qty       int64 `json:"qty" swagger:"desc(数量)" validate:"required"`
}

type ShareInvestStatusInput struct {
	UserId       int64                `json:"user_id" swagger:"desc(ユーザーID)"  validate:"required"`
	ProjectId    int64                `json:"project_id" swagger:"desc(ファンドプロジェクトID)" validate:"required"`
	InvestStatus domain.INVEST_STATUS `json:"order_status" swagger:"enum(REQUEST_INVEST|PROCESS_INVEST|SUCCESS_INVEST|CANCEL_INVEST),desc(1:注文リクエスト 2:注文処理中 3:注文成功 4:注文キャンセル)" validate:"required"`
}

func (d *ShareInput) ToModel() (*domain.Share, error) {

	return &domain.Share{
		ProjectId:    d.ProjectId,
		CategoryType: d.CategoryType,
		ProjectType:  d.ProjectType,
		OwnerId:      d.OwnerId,
		Desc:         d.Desc,
		InvestStatus: d.InvestStatus,
		Amount:       d.Amount,
		Qty:          d.Qty,
		Price:        d.Price,
	}, nil
}
