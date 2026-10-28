package controllers

type TrustInfoSearchInput struct {
	SearchInput
	UserId    *string `query:"user_id" swagger:"desc(投資家ID)"`
	Name      *string `query:"name" swagger:"desc(投資家名)"`
	Email     *string `query:"email" swagger:"desc(投資家メール)"`
	StartDate string  `query:"start_date" swagger:"desc(情報検索開始日)"`
	EndDate   string  `query:"end_date" swagger:"desc(情報検索終了日)"`
}
