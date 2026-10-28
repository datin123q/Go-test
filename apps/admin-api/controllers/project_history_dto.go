package controllers

type ProjectInfoHistorySearchInput struct {
	ProjectId int64 `query:"projectId"` // 投資商品ID
	SearchInput
}
