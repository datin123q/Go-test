package controllers

import "github.com/illuminarean/energy-shares-backend-repo/domain"

type ReportOutputInput struct {
	UserId         int64  `json:"user_id" swagger:"desc(投資家ID)"`
	SpvId          int64  `json:"spv_id" swagger:"desc(開発会社ID)"`
	SubmissionType string `json:"submission_type" swagger:"default(SUBMISSION_INVESTORS),desc(提出対象)" validate:"required"`
	ReportType     string `json:"report_type" swagger:"default(REPORT_TRANSACTION_BALANCE_REPORT),desc(レポートタイプ)" validate:"required"`
	Year           string `json:"year" swagger:"desc(年度)" validate:"required"`
	Quater         string `json:"quater" swagger:"enum(QUARTER_FIRST|QUARTER_SECOND|QUARTER_THIRD|QUARTER_FOURTH),default(QUARTER_FIRST),desc(四半期)"`
	IsOpen         bool   `json:"is_open" swagger:"enum(true|false),default(false),desc(公開有無)"`
}

func (roi ReportOutputInput) ToModel() (*domain.ReportOutput, error) {
	return &domain.ReportOutput{
		UserId:         roi.UserId,
		SpvId:          roi.SpvId,
		SubmissionType: domain.SUBMISSION_TYPE(roi.SubmissionType),
		ReportType:     domain.REPORT_TYPE(roi.ReportType),
		Year:           roi.Year,
		Quarter:        domain.QUARTER_TYPE(roi.Quater),
		IsOpen:         roi.IsOpen,
	}, nil
}
