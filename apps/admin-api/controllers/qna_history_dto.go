package controllers

type QnaInfoHistorySearchInput struct {
	QnaId *string `query:"qna_id"`
	SearchInput
}
