package controllers

// Quergy
type MyPageSearchInput struct {
	ProjectType   *string `query:"project_type"`
	ProjectStatus *string `query:"project_status"`
	InvestStatus  *string `query:"invest_status"`
	SearchInput
}
