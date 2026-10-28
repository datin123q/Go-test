package controllers

type VirtualAccountSearchInput struct {
	SearchInput
	StartDate string `query:"start_date" swagger:"desc(開始日)"`
	EndDate   string `query:"end_date" swagger:"desc(終了日)"`
}
type AccountTransactionInput struct {
	UserId                int64  `json:"user_id" form:"user_id"`
	UserAccountId         int64  `json:"user_account_id" form:"user_account_id"`
	AccountNumber         string `json:"account_number" form:"account_number"`
	TransactionId         string `json:"transaction_id" form:"transaction_id"`
	Amount                int64  `json:"amount" form:"amount"`
	TransactionType       string `json:"transaction_type" form:"transaction_type"`
	TransactionDetailType string `json:"transaction_detail_type" form:"transaction_detail_type"`
	TransactionDate       string `json:"transaction_date" form:"transaction_date"`
}

type VirtualAccountInput struct {
	UserId           int64 `json:"user_id"`
	UserAccountId    int64 `json:"user_account_id"`
	VirtualAccountId int64 `json:"virtual_account_id"`
}
