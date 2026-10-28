package controllers

type DepositCheckSearchInput struct {
	SearchInput
	InvestorId            *string `query:"investor_id" swagger:"desc(投資家ID)"`
	InvestmentName        *string `query:"investor_name" swagger:"desc(投資家名)"`
	DepositCheckStatus    *string `query:"deposit_check_status" swagger:"desc(入金額確認状態)"`
	StartInvestmentAmount *int    `query:"start_investment_amount" swagger:"desc(投資金額以上)"`
	EndInvestmentAmount   *int    `query:"end_investment_amount" swagger:"desc(余剰不足金額以下)"`
	StartDepositAmount    *int    `query:"start_deposit_amount" swagger:"desc(入金確認金額以上)"`
	EndDepositAmount      *int    `query:"end_deposit_amount" swagger:"desc(入金確認金額以下)"`
}

type DepositMatchSearchInput struct {
	SearchInput
	InvestorId              *string `query:"investor_id" swagger:"desc(投資家ID)"`
	InvestorName            *string `query:"investor_name" swagger:"desc(投資家名)"`
	DepositMatchStatus      *string `query:"deposit_match_status" swagger:"desc(余剰不足状況)"`
	StartDepositMatchAmount *int    `query:"start_deposit_match_amount" swagger:"desc(余剰不足金以上)"`
	EndDepositMatchAmount   *int    `query:"end_deposit_match_amount" swagger:"desc(余剰不足金金額以下)"`
	StartDepositAmount      *int    `query:"start_deposit_amount" swagger:"desc(入金確認金額以上)"`
	EndDepositAmount        *int    `query:"end_deposit_amount" swagger:"desc(入金確認金額以下)"`
}
