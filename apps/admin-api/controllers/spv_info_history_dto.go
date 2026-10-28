package controllers

type SpvInfoHistorySearchInput struct {
	SpvId int64 `query:"spvId"` // 開発会社ID
	SearchInput
}
