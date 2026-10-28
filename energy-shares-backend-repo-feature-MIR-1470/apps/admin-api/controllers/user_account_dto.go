package controllers

import "github.com/illuminarean/energy-shares-backend-repo/domain"

type UserAccountDetailResponse struct {
	domain.UserAccount
	Withdrawal          int64                      `json:"withdrawal" swagger:"desc(出金申請金額)"`
	UserShareAmountInfo domain.UserShareAmountInfo `json:"user_share_amount_info" swagger:"desc(投資金額情報)"`
	UserTotalAsset      int64                      `json:"user_total_asset" swagger:"desc(顧客資産総額)"`
	UserTotalUnpaid     int64                      `json:"user_total_unpaid" swagger:"desc(顧客未払額総額)"`
}

type TransactionSearchInput struct {
	ProjectType *string `query:"project_type" swagger:"desc(프로젝트 타입)"`
	SearchInput
	QuerySearchInput
}
type VirtualAccountSearchInput struct {
	SearchInput
	VaBranchCode    *string `query:"va_branch_code" swagger:"desc(지점 번호)"`
	VaAccountNumber *string `query:"va_account_number" swagger:"desc(가상계좌 번호)"`
	LinkStatus      *string `query:"linked_status" swagger:"desc(가상계좌 상태)"`
	ProjectType     *string `query:"project_type" swagger:"desc(상품 유형),enum(LOAN|EQUITY)"`
	UserId          *string `query:"user_id" swagger:"desc(투자자 ID)"`
	StartDate       string  `query:"start_date" swagger:"desc(업데이트 일시 시작일)"`
	EndDate         string  `query:"end_date" swagger:"desc(업데이트 일시 종료일)"`
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

type UserAccountSearchInput struct {
	SearchInput
	UserId       *string `query:"user_id" swagger:"desc(투자자 ID)"`
	UserName     *string `query:"user_name" swagger:"desc(투자자 명)"`
	StartBalance *int    `query:"start_balance" swagger:"desc(예치금 엔 이상 )"`
	EndBalance   *int    `query:"end_balance" swagger:"desc(예치금 엔 이하 )"`
}
