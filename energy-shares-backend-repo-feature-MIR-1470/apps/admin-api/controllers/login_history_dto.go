package controllers

type LoginHistorySearchInput struct {
	SearchInput
	Name      *string `query:"name" swagger:"desc(処理者)"`
	StartDate string  `query:"start_date" swagger:"desc(ログイン時間検索開始時間)"`
	EndDate   string  `query:"end_date" swagger:"desc(ログイン時間検索終了時間)"`
}
