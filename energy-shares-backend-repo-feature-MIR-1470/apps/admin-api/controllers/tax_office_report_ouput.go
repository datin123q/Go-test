package controllers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/apirender"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

type TaxOfficeReportOutputApiController struct {
}

func (c TaxOfficeReportOutputApiController) Init(g echoswagger.ApiGroup) {

	g.GET("/payment-statement/spvs/:spv_id/support-year", c.GetTaxOfficeReportGroupByYear).
		AddParamPath("", "spv_id", "開発会社ID").
		AddResponse(http.StatusOK, "税務署用フォーマット出力年度照会", []string{}, nil).
		SetDescription("支払調書が存在する年度を提供します。").
		SetSummary("支払調書が存在する年度を提供").
		SetSecurity("Authorization")

	g.GET("/spvs", c.GetSpvList).
		AddParamQueryNested(TaxOfficeSpvSearchInput{}).
		AddResponse(http.StatusOK, "支払調書ページの開発会社リスト照会", TaxOfficeSpvResponse{}, nil).
		SetDescription("支払調書ページの開発会社リストを照会します。").
		SetSummary("支払調書ページの開発会社リスト照会").
		SetSecurity("Authorization")

	g.GET("/payment-statement/spvs/:spv_id/year/:year/download", c.DownloadPaymentStatementReport).
		AddParamPath("", "spv_id", "開発会社ID").
		AddParamPath("", "year", "").
		AddResponse(http.StatusOK, "統合税務署用支払調書出力", domain.ReportOutput{}, nil).
		SetDescription("統合税務署用支払調書を出力します。").
		SetSummary("統合税務署用支払調書出力").
		SetSecurity("Authorization")

	g.GET("/payment-statement-sum/spvs/:spv_id/year/:year/download", c.DownloadPaymentStatementSumReport).
		AddParamPath("", "spv_id", "開発会社ID").
		AddParamPath("", "year", "").
		AddResponse(http.StatusOK, "税務署用支払調書合計出力", domain.ReportOutput{}, nil).
		SetDescription("税務署用支払調書合計を出力します。").
		SetSummary("税務署用支払調書合計出力").
		SetSecurity("Authorization")
}

func (TaxOfficeReportOutputApiController) GetTaxOfficeReportGroupByYear(c echo.Context) error {
	spvId, err := strconv.ParseInt(c.Param("spv_id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	totalCount, items, err := domain.ReportOutput{}.GetAllTaxOfficeReportGroupByYear(c.Request().Context(), spvId)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, apirender.ArrayResult{
		Items:      items,
		TotalCount: totalCount,
	})
}

func (TaxOfficeReportOutputApiController) GetOne(c echo.Context) error {

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	reportOutput, err := domain.ReportOutput{}.GetById(c.Request().Context(), id)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	if reportOutput == nil {
		return factory.ErrorNotFound.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, reportOutput)
}

func (TaxOfficeReportOutputApiController) Update(c echo.Context) error {

	var toi ReportOutputInput
	if err := c.Bind(&toi); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(&toi); err != nil {
		return factory.ErrorParameter.New(err)
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}
	reportOutput, err := toi.ToModel()
	if err != nil {
		return factory.ErrorNotFound.New(err)
	}
	reportOutput.Id = id

	if err := reportOutput.Update(c.Request().Context()); err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, reportOutput)
}

func (TaxOfficeReportOutputApiController) Delete(c echo.Context) error {

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	reportOutput, err := domain.ReportOutput{}.GetById(c.Request().Context(), id)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	if reportOutput == nil {
		return factory.ErrorNotFound.New(err)
	}

	if err := reportOutput.Delete(c.Request().Context(), id); err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, reportOutput)
}

// 税務署用支払調書出力
func (TaxOfficeReportOutputApiController) DownloadPaymentStatementReport(c echo.Context) error {
	spvId, err := strconv.ParseInt(c.Param("spv_id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}
	year := c.Param("year")
	if year == "" {
		return factory.ErrorParameter.New(nil)
	}

	item, err := domain.ReportOutput{}.FindAdminTaxOfficeReportByReportType(c.Request().Context(), domain.REPORT_PAYMENT_STATEMENT, spvId, year)
	if err != nil {
		return factory.ErrorDB.New(err)
	} else if item == nil {
		return factory.ErrorNotFoundReport.New(nil)
	}

	return apirender.RenderSucc(c, http.StatusOK, item.FileUrl)
}

// 税務署用支払調書合計出力
func (TaxOfficeReportOutputApiController) DownloadPaymentStatementSumReport(c echo.Context) error {
	spvId, err := strconv.ParseInt(c.Param("spv_id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}
	year := c.Param("year")
	if year == "" {
		return factory.ErrorParameter.New(nil)
	}

	item, err := domain.ReportOutput{}.FindAdminTaxOfficeReportByReportType(c.Request().Context(), domain.REPORT_PAYMENT_STATEMENT_SUM, spvId, year)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	if item == nil {
		return factory.ErrorNotFoundReport.New(nil)
	}

	return apirender.RenderSucc(c, http.StatusOK, item.FileUrl)
}

func (TaxOfficeReportOutputApiController) GetSpvList(c echo.Context) error {
	const MAX_RESULT_COUNT = 1000 // 最大1000個

	var v TaxOfficeSpvSearchInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	_, items, err := domain.Spv{}.GetAll(c.Request().Context(), nil, v.Name, nil, nil, "name", "desc", 0, MAX_RESULT_COUNT)
	if err != nil {
		return factory.ErrorDB.New(err)
	} else if items == nil {
		return factory.ErrorNotFound.New(nil)
	}

	return apirender.RenderSucc(c, http.StatusOK, TaxOfficeSpvResponse{}.ToResponseList(items))
}
