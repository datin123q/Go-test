package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	openapigo "github.com/swaggest/openapi-go"

	adomain "github.com/illuminarean/energy-share-core-api/domain"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/apirender"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/echoutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
)

const _DefaultMaxResultCount = 30

type investorReportHttpDelivery struct {
	cfg                   *config.Config
	extractor             jwthelper.UserExtractor
	logger                factory.LoggerFactory
	investorReportUsecase adomain.InvestorReportUsecase
}

func New(
	api *echoutils.EchoSwagger,
	cfg *config.Config,
	extractor jwthelper.UserExtractor,
	logger factory.LoggerFactory,
	investorReportUsecase adomain.InvestorReportUsecase,
) *investorReportHttpDelivery {

	g := api.Group("/api/v1/investor-report", echoutils.AllowRoles(domain.ROLE_REPORT))

	handler := &investorReportHttpDelivery{
		cfg:                   cfg,
		logger:                logger,
		extractor:             extractor,
		investorReportUsecase: investorReportUsecase,
	}

	g.GET("/payments", handler.GetAllPaymentStatement).
		SetTags("투자자 법정장부").
		SetSummary("투자자용 지불조서를 조회합니다").
		AddRequest(adomain.InvestorReportGetAllRequest{}).
		AddResponse(adomain.InvestorReportGetAllResponse{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "투자자용 지불조서 목록"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorParameter,
			factory.ErrorDB,
		).
		AddSecurity("Authorization")

	g.GET("/transaction-balances", handler.GetAllTransactionBalanceReport).
		SetTags("투자자 법정장부").
		SetSummary("투자자용 거래잔액보고서를 조회합니다").
		AddRequest(adomain.InvestorReportGetAllRequest{}).
		AddResponse(adomain.InvestorReportGetAllResponse{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "투자자용 거래잔액보고서 목록"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorParameter,
			factory.ErrorDB,
		).
		AddSecurity("Authorization")

	g.PUT("/payments", handler.OpenPaymentStatement).
		SetTags("투자자 법정장부").
		SetSummary("투자자용 지불조서를 공개합니다").
		AddRequest(adomain.InvestorReportOpenPaymentStatementRequest{}).
		AddResponse(nil, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorParameter,
			factory.ErrorDB,
		).
		AddSecurity("Authorization")

	g.PUT("/transaction-balances", handler.OpenTransactionBalanceReport).
		SetTags("투자자 법정장부").
		SetSummary("투자자용 거래잔액보고서를 공개합니다").
		AddRequest(adomain.InvestorReportOpenTransactionBalanceRequest{}).
		AddResponse(nil, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorParameter,
			factory.ErrorDB,
		).
		AddSecurity("Authorization")

	g.GET("/payments/download", handler.DownloadPaymentStatement).
		SetTags("투자자 법정장부").
		SetSummary("투자자용 지불조서를 출력합니다").
		AddRequest(adomain.DownloadPaymentStatementRequest{}).
		AddResponse(adomain.DownloadPaymentStatementResponse{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "투자자용 지불조서 다운로드 URL"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorParameter,
			factory.ErrorDB,
		).
		AddSecurity("Authorization")

	g.GET("/transaction-balances/download", handler.DownloadTransactionBalanceReport).
		SetTags("투자자 법정장부").
		SetSummary("투자자용 거래잔액보고서를 출력합니다").
		AddRequest(adomain.DownloadTransactionBalanceReportRequest{}).
		AddResponse(adomain.DownloadTransactionBalanceReportResponse{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "투자자용 거래잔액보고서 다운로드 URL"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorParameter,
			factory.ErrorDB,
		).
		AddSecurity("Authorization")

	return handler
}

func (d investorReportHttpDelivery) GetAllPaymentStatement(c echo.Context) error {
	// 유저 정보 추출
	user, err := d.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	// Request DTO 생성
	var request = adomain.InvestorReportGetAllRequest{
		JWTUser:  user,
		ClientIp: c.RealIP(),
	}

	if err = c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	if err := c.Validate(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	if request.MaxResultCount == 0 {
		request.MaxResultCount = _DefaultMaxResultCount
	}

	ctx := c.Request().Context()
	response, err := d.investorReportUsecase.GetAllPaymentStatement(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, response)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (d investorReportHttpDelivery) GetAllTransactionBalanceReport(c echo.Context) error {
	// 유저 정보 추출
	user, err := d.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	// Request DTO 생성
	var request = adomain.InvestorReportGetAllRequest{
		JWTUser:  user,
		ClientIp: c.RealIP(),
	}

	if err = c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	if err := c.Validate(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	if request.MaxResultCount == 0 {
		request.MaxResultCount = _DefaultMaxResultCount
	}

	ctx := c.Request().Context()
	response, err := d.investorReportUsecase.GetAllTransactionBalanceReport(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, response)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (d investorReportHttpDelivery) OpenPaymentStatement(c echo.Context) error {
	// 유저 정보 추출
	user, err := d.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	// Request DTO 생성
	var request = adomain.InvestorReportOpenPaymentStatementRequest{
		JWTUser:  user,
		ClientIp: c.RealIP(),
	}

	if err = c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	if err = c.Validate(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	ctx := c.Request().Context()
	err = d.investorReportUsecase.OpenPaymentStatement(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, nil)
	case errors.Is(err, cdomain.ErrAlreadyOpenReport):
		return factory.ErrorAlreadyOpenReport.New(nil)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (d investorReportHttpDelivery) OpenTransactionBalanceReport(c echo.Context) error {
	// 유저 정보 추출
	user, err := d.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	// Request DTO 생성
	var request = adomain.InvestorReportOpenTransactionBalanceRequest{
		JWTUser:  user,
		ClientIp: c.RealIP(),
	}

	if err = c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err = c.Validate(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	ctx := c.Request().Context()
	err = d.investorReportUsecase.OpenTransactionBalanceReport(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, nil)
	case errors.Is(err, cdomain.ErrAlreadyOpenReport):
		return factory.ErrorAlreadyOpenReport.New(nil)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (d investorReportHttpDelivery) DownloadPaymentStatement(c echo.Context) error {
	// 유저 정보 추출
	user, err := d.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	// Request DTO 생성
	var request = adomain.DownloadPaymentStatementRequest{
		JWTUser:  user,
		ClientIp: c.RealIP(),
	}

	if err = c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	if err = c.Validate(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	ctx := c.Request().Context()
	response, err := d.investorReportUsecase.DownloadPaymentStatement(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, response)
	case errors.Is(err, cdomain.ErrItemNotFound):
		return factory.ErrorNotFoundReport.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (d investorReportHttpDelivery) DownloadTransactionBalanceReport(c echo.Context) error {
	// 유저 정보 추출
	user, err := d.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	// Request DTO 생성
	var request = adomain.DownloadTransactionBalanceReportRequest{
		JWTUser:  user,
		ClientIp: c.RealIP(),
	}

	if err = c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	if err = c.Validate(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	ctx := c.Request().Context()
	response, err := d.investorReportUsecase.DownloadTransactionBalanceReport(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, response)
	case errors.Is(err, cdomain.ErrItemNotFound):
		return factory.ErrorNotFoundReport.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}
