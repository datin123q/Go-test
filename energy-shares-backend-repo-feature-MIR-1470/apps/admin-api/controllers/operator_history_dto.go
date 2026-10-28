package controllers

// Quergy
type OperatorHistorySearchInput struct {
	Name      *string `query:"name"`
	Menu      *string `query:"menu"`
	State     *string `query:"state"`
	StartDate string  `query:"start_date"` // 処理日時開始
	EndDate   string  `query:"end_date"`   // 処理日時終了
	SearchInput
}
