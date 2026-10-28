package controllers

type JobExecuteTimeSearchInput struct {
	RequestTime REQUEST_TIME_TYPE `query:"request_time" swagger:"enum(DEPOSIT_CHECK|ISSUE_VIRTUAL_ACCOUNT|TRANSFER_REQUEST|TRANSFER_RESULT)"` // 最後の入金履歴照会時刻
}
