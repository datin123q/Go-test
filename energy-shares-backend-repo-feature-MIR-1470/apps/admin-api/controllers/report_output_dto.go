package controllers

import (
	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type ReportOutputOpenInput struct {
	Year    string `json:"year" swagger:"desc(年度)"`
	Quarter string `json:"quarter" swagger:"desc(四半期)"`
}

type ReportOutputInput struct {
	UserId         int64  `json:"user_id" swagger:"desc(投資家ID)"`
	SpvId          int64  `json:"spv_id"  swagger:"desc(開発会社ID)"`
	SubmissionType string `json:"submission_type" swagger:"default(SUBMISSION_INVESTORS),desc(提出対象)" validate:"oneof=SUBMISSION_INVESTORS SUBMISSION_TAX_OFFICE SUBMISSION_AUDIT_RESPONSE"`
	ReportType     string `json:"report_type" swagger:"default(REPORT_TRANSACTION_BALANCE_REPORT),desc(フォーマットタイプ)" validate:"min=1"`
	Year           string `json:"year" swagger:"desc(年度)" validate:"min=1"`
	Quarter        string `json:"quarter" swagger:"enum(QUARTER_FIRST|QUARTER_SECOND|QUARTER_THIRD|QUARTER_FOURTH),default(QUARTER_FIRST),desc(四半期)"`
	IsOpen         bool   `json:"is_open" swagger:"enum(true|false),default(false),desc(公開有無)"`
}

func (roi ReportOutputInput) ToModel() (*domain.ReportOutput, error) {
	return &domain.ReportOutput{
		UserId:         roi.UserId,
		SpvId:          roi.SpvId,
		SubmissionType: domain.SUBMISSION_TYPE(roi.SubmissionType),
		ReportType:     domain.REPORT_TYPE(roi.ReportType),
		Year:           roi.Year,
		Quarter:        domain.QUARTER_TYPE(roi.Quarter),
		IsOpen:         roi.IsOpen,
	}, nil
}

type TaxOfficeSpvSearchInput struct {
	Name *string `query:"name" json:"name" swagger:"desc(会社名)"` // 会社名
}

type TaxOfficeSpvResponse struct {
	Id   int64  `json:"id" swagger:"desc(아이디)"`
	Name string `json:"name" swagger:"desc(회사명)"`
}

func (r TaxOfficeSpvResponse) ToResponse(spv domain.Spv) *TaxOfficeSpvResponse {
	return &TaxOfficeSpvResponse{
		Id:   spv.Id,
		Name: spv.Name,
	}
}

func (TaxOfficeSpvResponse) ToResponseList(spvList []domain.Spv) []*TaxOfficeSpvResponse {
	responses := make([]*TaxOfficeSpvResponse, 0, len(spvList))
	for _, spv := range spvList {
		responses = append(responses, &TaxOfficeSpvResponse{
			Id:   spv.Id,
			Name: spv.Name,
		})
	}
	return responses
}
