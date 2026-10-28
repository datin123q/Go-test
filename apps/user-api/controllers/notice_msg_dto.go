package controllers

type NoticeSearchInput struct {
	Title     *string `query:"title"`
	StartDate string  `query:"start_date" swagger:"desc(開始日)"`
	EndDate   string  `query:"end_date" swagger:"desc(終了日)"`
	SearchInput
}
