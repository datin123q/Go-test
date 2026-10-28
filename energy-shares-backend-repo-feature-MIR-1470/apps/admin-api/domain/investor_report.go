package domain

import (
	"context"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
)

type InvestorReportUsecase interface {
	GetAllPaymentStatement(ctx context.Context, req InvestorReportGetAllRequest) (InvestorReportGetAllResponse, error)
	GetAllTransactionBalanceReport(ctx context.Context, req InvestorReportGetAllRequest) (InvestorReportGetAllResponse, error)
	OpenPaymentStatement(ctx context.Context, req InvestorReportOpenPaymentStatementRequest) error
	OpenTransactionBalanceReport(ctx context.Context, req InvestorReportOpenTransactionBalanceRequest) error
	DownloadPaymentStatement(ctx context.Context, req DownloadPaymentStatementRequest) (DownloadPaymentStatementResponse, error)
	DownloadTransactionBalanceReport(ctx context.Context, req DownloadTransactionBalanceReportRequest) (DownloadTransactionBalanceReportResponse, error)
}

type InvestorReportGetAllRequest struct {
	JWTUser  jwthelper.User
	ClientIp string
	cdomain.SearchInput
}

type InvestorReportGetAllResponse struct {
	TotalCount int64                 `json:"total_count" description:"법정정부 개수" example:"10"`
	Items      []domain.ReportOutput `json:"items"`
}

type InvestorReportOpenPaymentStatementRequest struct {
	JWTUser  jwthelper.User
	ClientIp string
	Year     string `json:"year" description:"공개할 연도" validate:"required,datetime=2006" example:"2024"`
}

type InvestorReportOpenTransactionBalanceRequest struct {
	JWTUser  jwthelper.User
	ClientIp string
	Year     string `json:"year" description:"공개할 연도" validate:"required,datetime=2006" example:"2024"`
	Quarter  string `json:"quarter" description:"공개할 분기" validate:"required,oneof=QUARTER_FIRST QUARTER_SECOND QUARTER_THIRD QUARTER_FOURTH" example:"QUARTER_FIRST"`
}

type DownloadPaymentStatementRequest struct {
	JWTUser  jwthelper.User
	ClientIp string
	Year     string `query:"year" description:"공개할 연도" validate:"required,datetime=2006" example:"2024"`
}

type DownloadPaymentStatementResponse struct {
	Url string `json:"url" description:"다운로드 URL"`
}

type DownloadTransactionBalanceReportRequest struct {
	JWTUser  jwthelper.User
	ClientIp string
	Year     string `query:"year" description:"연도" validate:"required,datetime=2006" example:"2024"`
	Quarter  string `query:"quarter" description:"분기" validate:"required,oneof=QUARTER_FIRST QUARTER_SECOND QUARTER_THIRD QUARTER_FOURTH" example:"QUARTER_FIRST"`
}

type DownloadTransactionBalanceReportResponse struct {
	Url string `json:"url" description:"다운로드 URL"`
}
