package controllers

type FundingRequestHistorySearchInput struct {
	FundingRequestId int64 `query:"fundingRequestId"`
	SearchInput
}
