package controllers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
	"github.com/sirupsen/logrus"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/apirender"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

type InvestorReportOutputApiController struct {
}

func NewInvestorReportOutputApiController() *InvestorReportOutputApiController {
	return &InvestorReportOutputApiController{}
}

func (c *InvestorReportOutputApiController) Init(g echoswagger.ApiGroup) {

	g.GET("/transaction-balance-report/list", c.GetAllTransactionBalanceReportList).
		AddParamQueryNested(SearchInput{}).
		AddResponse(http.StatusOK, "取引残高報告書リスト照会", domain.ReportOutput{}, nil).
		SetDescription("取引残高報告書リストを照会します.").
		SetSummary("取引残高報告書サンプリスト照会").
		SetSecurity("Authorization")

	g.GET("/payment-statement/list", c.GetAllPaymentStatementList).
		AddParamQueryNested(SearchInput{}).
		AddResponse(http.StatusOK, "支払調書リスト照会", domain.ReportOutput{}, nil).
		SetDescription("支払調書リストを照会します.").
		SetSummary("支払調書リスト照会").
		SetSecurity("Authorization")

	g.GET("/transaction-balance-report/year/:year/quarter/:quarter/download", c.DownloadTransactionBalanceReport).
		AddParamPath("", "year", "").
		AddParamPath("", "quarter", "").
		AddResponse(http.StatusOK, "統合投資家用取引残高報告書出力", domain.ReportOutput{}, nil).
		SetDescription("統合投資家用取引残高報告書を出力します.").
		SetSummary("統合投資家用取引残高報告書出力").
		SetSecurity("Authorization")

	g.GET("/payment-statement/year/:year/download", c.DownloadPaymentStatementReport).
		AddParamPath("", "year", "").
		AddResponse(http.StatusOK, "統合投資家用支払調書出力", domain.ReportOutput{}, nil).
		SetDescription("統合投資家用支払調書を出力します.").
		SetSummary("統合投資家用支払調書出力").
		SetSecurity("Authorization")

	g.PUT("/transaction-balance-report/update", c.UpdateTransactionBalanceReport).
		AddParamBody(ReportOutputOpenInput{}, "body", "", true).
		AddResponse(http.StatusOK, "統合投資家用取引残高報告書公開", domain.ReportOutput{}, nil).
		SetDescription("統合投資家用取引残高報告書を公開します.").
		SetSummary("統合投資家用取引残高報告書公開").
		SetSecurity("Authorization")

	g.PUT("/payment-statement/update", c.UpdatePaymentStatementReport).
		AddParamBody(ReportOutputOpenInput{}, "body", "", true).
		AddResponse(http.StatusOK, "統合支払調書公開", domain.ReportOutput{}, nil).
		SetDescription("投資家用支払調書を公開します.").
		SetSummary("投資家用支払調書公開").
		SetSecurity("Authorization")
}

// 投資家用取引残高報告書リスト
func (InvestorReportOutputApiController) GetAllTransactionBalanceReportList(c echo.Context) error {

	var v SearchInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if v.MaxResultCount == 0 {
		v.MaxResultCount = DefaultMaxResultCount
	}

	// console log
	factory.Logger(c.Request().Context()).WithFields(logrus.Fields{
		"sortby":           v.Sortby,
		"order":            v.Order,
		"max_result_count": v.MaxResultCount,
		"skip_count":       v.SkipCount,
	}).Info("Tax-office Report SearchStart")

	totalCount, items, err := domain.ReportOutput{}.GetQuarterTransactionBalanceAdminReportList(c.Request().Context(), v.Sortby, v.Order)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, apirender.ArrayResult{
		Items:      items,
		TotalCount: totalCount,
	})
}

// 投資家用支払調書リスト
func (InvestorReportOutputApiController) GetAllPaymentStatementList(c echo.Context) error {

	var v SearchInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if v.MaxResultCount == 0 {
		v.MaxResultCount = DefaultMaxResultCount
	}

	// console log
	factory.Logger(c.Request().Context()).WithFields(logrus.Fields{
		"sortby":           v.Sortby,
		"order":            v.Order,
		"max_result_count": v.MaxResultCount,
		"skip_count":       v.SkipCount,
	}).Info("Tax-office Report SearchStart")

	totalCount, items, err := domain.ReportOutput{}.GetAllInvestorPaymentStatementList(c.Request().Context(), v.Sortby, v.Order, v.SkipCount, v.MaxResultCount)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, apirender.ArrayResult{
		Items:      items,
		TotalCount: totalCount,
	})
}

// 管理者取引残高報告書出力
func (InvestorReportOutputApiController) DownloadTransactionBalanceReport(c echo.Context) error {

	year := c.Param("year")
	quarter := c.Param("quarter")

	item, err := domain.ReportOutput{}.FindAdminInvestorReportByReportType(c.Request().Context(), domain.REPORT_TRANSACTION_BALANCE_ADMIN_REPORT, year, quarter)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	if item == nil {
		return factory.ErrorNotFoundReport.New(err)
	}
	return apirender.RenderSucc(c, http.StatusOK, item.FileUrl)
}

// 投資家用支払調書出力
func (InvestorReportOutputApiController) DownloadPaymentStatementReport(c echo.Context) error {

	year := c.Param("year")

	item, err := domain.ReportOutput{}.FindAdminInvestorReportByReportType(c.Request().Context(), domain.REPORT_PAYMENT_STATEMENT, year, "")
	if err != nil {
		return factory.ErrorDB.New(err)
	} else if item == nil {
		return factory.ErrorNotFoundReport.New(nil)
	}

	return apirender.RenderSucc(c, http.StatusOK, item.FileUrl)
}

// 投資家用取引残高報告書公開
func (InvestorReportOutputApiController) UpdateTransactionBalanceReport(c echo.Context) error {

	var ro ReportOutputOpenInput
	if err := c.Bind(&ro); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(&ro); err != nil {
		return factory.ErrorParameter.New(err)
	}

	alreadyOpen, err := domain.ReportOutput{}.OpenTransactionBalanceReport(c.Request().Context(), ro.Year, ro.Quarter)
	if err != nil {
		if alreadyOpen {
			return factory.ErrorAlreadyOpenReport.New(err)
		}
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, nil)
}

// 投資家用支払調書公開
func (InvestorReportOutputApiController) UpdatePaymentStatementReport(c echo.Context) error {
	var ro ReportOutputOpenInput
	if err := c.Bind(&ro); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(&ro); err != nil {
		return factory.ErrorParameter.New(err)
	}

	alreadyOpen, err := domain.ReportOutput{}.OpenPaymentStatementReport(c.Request().Context(), ro.Year)
	if err != nil {
		if alreadyOpen {
			return factory.ErrorAlreadyOpenReport.New(err)
		}
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, nil)
}
