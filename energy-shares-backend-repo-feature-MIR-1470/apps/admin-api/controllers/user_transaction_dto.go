package controllers

type UserTransactionSearchInput struct {
	UserId          *string `query:"user_id"`
	UserName        *string `query:"user_name"`
	TransactionType *string `query:"transaction_type"`
	StartAmount     *int    `query:"start_amount"`
	EndAmount       *int    `query:"end_amount"`
	SearchInput
}
