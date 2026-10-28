package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
	"github.com/pkg/errors"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/apirender"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
)

type AuditResponseReportOutputApiController struct {
}

func (c AuditResponseReportOutputApiController) Init(g echoswagger.ApiGroup) {

	// サンプルはリアルタイムで出力されます。
	// 法定帳簿サービスから要求されます。
	// 年別に四半期別に出力されます。
	// 法定帳簿サーバーは、同じ年度に対しては重複生成されません。

	g.GET("/support-year", c.GetSupportYearList).
		AddResponse(http.StatusOK, "監査対応書類照会可能期間", []int{}, nil).
		SetDescription("監査対応書類照会可能期間を返します.").
		SetSummary("監査対応書類照会可能期間").
		SetSecurity("Authorization")

	g.GET("/trading-diary/year/:year/download", c.DownloadTradingDiaryReport).
		AddParamPath("", "year", "").
		AddResponse(http.StatusOK, "監査対応用全取引日記帳出力", domain.ReportOutput{}, nil).
		SetDescription("監査対応用取引日記帳の全年度出力を行います.").
		SetSummary("監査対応用取引日記帳の年度別記録出力").
		SetSecurity("Authorization")

	g.GET("/trading-diary/download", c.DownloadTradingDiaryReportAll).
		AddResponse(http.StatusOK, "監査対応用取引日記帳出力", domain.ReportOutput{}, nil).
		SetDescription("監査対応用取引日記帳の全年度出力を行います.").
		SetSummary("監査対応用取引日記帳の全年度出力").
		SetSecurity("Authorization")

	g.GET("/transaction-record/download", c.DownloadTransactionRecordReportAll).
		AddResponse(http.StatusOK, "監査対応用募集取扱に関する取引記録出力", domain.ReportOutput{}, nil).
		SetDescription("監査対応用募集取扱に関する取引記録を出力します.").
		SetSummary("監査対応用募集取扱に関する取引記録出力").
		SetSecurity("Authorization")

	g.GET("/transaction-record/year/:year/download", c.DownloadTransactionRecordReport).
		AddParamPath("", "year", "").
		AddResponse(http.StatusOK, "監査対応用募集取扱に関する取引記録出力", domain.ReportOutput{}, nil).
		SetDescription("監査対応用募集取扱に関する取引記録を出力します.").
		SetSummary("監査対応用募集取扱に関する取引記録出力").
		SetSecurity("Authorization")

	g.GET("/customer-account/download", c.DownloadCustomerAccountReportAll).
		AddResponse(http.StatusOK, "監査対応用顧客勘定原帳出力", domain.ReportOutput{}, nil).
		SetDescription("監査対応用顧客勘定原帳を出力します.").
		SetSummary("監査対応用顧客勘定原帳出力").
		SetSecurity("Authorization")

	g.GET("/customer-account/year/:year/download", c.DownloadCustomerAccountReport).
		AddParamPath("", "year", "").
		AddResponse(http.StatusOK, "監査対応用顧客勘定原帳出力", domain.ReportOutput{}, nil).
		SetDescription("監査対応用顧客勘定原帳を出力します.").
		SetSummary("監査対応用顧客勘定原帳出力").
		SetSecurity("Authorization")

}

// 監査対応用取引日記帳出力
func (AuditResponseReportOutputApiController) DownloadTradingDiaryReport(c echo.Context) error {

	year, err := strconv.Atoi(c.Param("year"))
	if err != nil {
		return factory.ErrorParameter.New(errors.Wrap(err, "fail to convert year"))
	}
	startDate := time.Date(year, 1, 1, 0, 0, 0, 0, time.Local)
	endDate := time.Date(year, 12, 31, 0, 0, 0, 0, time.Local)

	url, err := domain.ReportOutput{}.RequestTradingDiaryReport(&startDate, &endDate)
	if err != nil {
		return err
	}
	//TODO domain.OperatorHistory{}.Record(c.Request().Context(), tokenId, email, "", "", c.RealIP(), domain.DOWNLOAD)
	return apirender.RenderSucc(c, http.StatusOK, url)
}

func (AuditResponseReportOutputApiController) DownloadTradingDiaryReportAll(c echo.Context) error {

	url, err := domain.ReportOutput{}.RequestTradingDiaryReport(nil, nil)
	if err != nil {
		return err
	}
	return apirender.RenderSucc(c, http.StatusOK, url)
}

func (AuditResponseReportOutputApiController) DownloadAllTradingDiaryReport(c echo.Context) error {
	tokenId, email, err := jwthelper.TokenToUser(c)
	if err != nil {
		return factory.ErrorSystem.New(errors.Wrap(err, "fail to TokenToUser"))
	}

	startDate := time.Date(2022, 1, 1, 0, 0, 0, 0, time.Local)
	endDate := time.Date(2022, 12, 31, 0, 0, 0, 0, time.Local)

	url, err := domain.ReportOutput{}.RequestTradingDiaryReport(&startDate, &endDate)
	if err != nil {
		return factory.ErrorSystem.New(errors.Wrap(err, "fail to RequestTradingDiaryReport"))
	}
	domain.OperatorHistory{}.Record(c.Request().Context(), tokenId, email, "", "", c.RealIP(), domain.DOWNLOAD)
	return apirender.RenderSucc(c, http.StatusOK, url)
}

// 監査対応用募集取扱に関する取引記録出力
func (AuditResponseReportOutputApiController) DownloadTransactionRecordReportAll(c echo.Context) error {

	url, err := domain.ReportOutput{}.RequestTransactionRecord(nil, nil)
	if err != nil {
		return err
	}

	//TODO: domain.OperatorHistory{}.Record(c.Request().Context(), tokenId, email, "", "", c.RealIP(), domain.DOWNLOAD)
	return apirender.RenderSucc(c, http.StatusOK, url)
}

func (AuditResponseReportOutputApiController) DownloadTransactionRecordReport(c echo.Context) error {
	year, err := strconv.Atoi(c.Param("year"))
	if err != nil {
		return factory.ErrorParameter.New(errors.Wrap(err, "fail to convert year"))
	}
	startDate := time.Date(year, 1, 1, 0, 0, 0, 0, time.Local)
	endDate := time.Date(year, 12, 31, 0, 0, 0, 0, time.Local)

	url, err := domain.ReportOutput{}.RequestTransactionRecord(&startDate, &endDate)
	if err != nil {
		return err
	}
	// TODO: domain.OperatorHistory{}.Record(c.Request().Context(), tokenId, email, "", "", c.RealIP(), domain.DOWNLOAD)
	return apirender.RenderSucc(c, http.StatusOK, url)
}

func (AuditResponseReportOutputApiController) DownloadCustomerAccountReportAll(c echo.Context) error {

	url, err := domain.ReportOutput{}.RequestCustomerAccount(nil, nil)
	if err != nil {
		return err
	}
	// TODO: domain.OperatorHistory{}.Record(c.Request().Context(), tokenId, email, "", "", c.RealIP(), domain.DOWNLOAD)
	return apirender.RenderSucc(c, http.StatusOK, url)
}

// 監査対応用顧客勘定原帳出力
func (AuditResponseReportOutputApiController) DownloadCustomerAccountReport(c echo.Context) error {

	year, err := strconv.Atoi(c.Param("year"))
	if err != nil {
		return factory.ErrorParameter.New(errors.Wrap(err, "fail to convert year"))
	}

	startDate := time.Date(year, 1, 1, 0, 0, 0, 0, time.Local)
	endDate := time.Date(year, 12, 31, 0, 0, 0, 0, time.Local)

	url, err := domain.ReportOutput{}.RequestCustomerAccount(&startDate, &endDate)
	if err != nil {
		return err
	}
	// TODO: domain.OperatorHistory{}.Record(c.Request().Context(), tokenId, email, "", "", c.RealIP(), domain.DOWNLOAD)
	return apirender.RenderSucc(c, http.StatusOK, url)
}

// GetSupportYearList 監査対応書類照会可能期間を返します。
func (AuditResponseReportOutputApiController) GetSupportYearList(c echo.Context) error {
	startYear := 2022
	endYear := time.Now().Year()
	var years []int

	for year := startYear; year <= endYear; year++ {
		years = append(years, year)
	}
	return apirender.RenderSucc(c, http.StatusOK, years)
}
