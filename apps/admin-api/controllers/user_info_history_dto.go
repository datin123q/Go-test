package controllers

type UserInfoHistorySearchInput struct {
	UserId int64 `query:"userId"` // 会員ID
	SearchInput
}
