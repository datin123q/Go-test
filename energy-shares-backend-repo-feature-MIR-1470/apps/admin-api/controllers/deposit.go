package controllers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
	"github.com/sirupsen/logrus"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/apirender"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/echoutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
)

type DepositMatchDataApiController struct{}

func (c DepositMatchDataApiController) Init(g echoswagger.ApiGroup) {

	g.GET("", c.GetAll).
		AddParamQueryNested(DepositMatchSearchInput{}).
		SetDescription("余剰不足金リストを照会します.").
		AddResponse(http.StatusOK, "余剰不足金データ", domain.DepositMatchData{}, nil).
		SetSummary("余剰不足金リスト照会").
		SetSecurity("Authorization")

	g.GET("/download", c.GetDownlaod).
		AddParamQueryNested(DepositMatchSearchInput{}).
		SetDescription("余剰不足金リストをダウンロードします.").
		AddResponse(http.StatusOK, "余剰不足金リストダウンロード", domain.DepositMatchData{}, nil).
		SetSummary("余剰不足金リストダウンロード").
		SetSecurity("Authorization")

	g.GET("/:id", c.GetOne).
		AddParamPath("", "id", "").
		SetDescription("顧客が入金した金額と投資した金額の余剰不足金の詳細情報を提供します.").
		AddResponse(http.StatusOK, "余剰不足金データ", domain.DepositMatchData{}, nil).
		SetSummary("余剰不足金詳細照会").
		SetSecurity("Authorization")
}

type DepositCheckApiController struct{}

func (c DepositCheckApiController) Init(g echoswagger.ApiGroup) {
	g.GET("", c.GetAll,
		echoutils.AllowRoles(domain.ROLE_PAYMENT)).
		AddParamQueryNested(DepositCheckSearchInput{}).
		SetDescription("入金額リストを照会します.").
		AddResponse(http.StatusOK, "入金額確認データ", domain.DepositCheckData{}, nil).
		SetSummary("入金額確認リスト照会").
		SetSecurity("Authorization")

	g.GET("/download", c.GetDownlaod,
		echoutils.AllowRoles(domain.ROLE_PAYMENT)).
		AddParamQueryNested(DepositCheckSearchInput{}).
		SetDescription("入金額リストをダウンロードします.").
		AddResponse(http.StatusOK, "入金額確認リストダウンロード", domain.DepositCheckData{}, nil).
		SetSummary("入金額確認ダウンロード").
		SetSecurity("Authorization")
}

func (DepositCheckApiController) GetAll(c echo.Context) error {
	var v DepositCheckSearchInput
	var items []domain.DepositCheckData
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if v.MaxResultCount == 0 {
		v.MaxResultCount = DefaultMaxResultCount
	}
	userId, email, err := jwthelper.TokenToUser(c)
	if err != nil {
		return err
	}
	factory.Logger(c.Request().Context()).WithFields(logrus.Fields{
		"sortby":                v.Sortby,
		"order":                 v.Order,
		"max_result_count":      v.MaxResultCount,
		"skip_count":            v.SkipCount,
		"startInvestmentAmount": v.StartInvestmentAmount,
		"endInvestmentAmount":   v.EndInvestmentAmount,
		"startDepositAmount":    v.StartDepositAmount,
		"endDepositAmount":      v.EndDepositAmount,
	}).Info("SearchStart")

	cnt, items, err := domain.DepositCheckData{}.GetAll(c.Request().Context(), v.Sortby, v.Order, v.SkipCount, v.MaxResultCount, v.InvestorId, v.InvestmentName, v.DepositCheckStatus, v.StartInvestmentAmount, v.EndInvestmentAmount, v.StartDepositAmount, v.EndDepositAmount)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	if items == nil {
		items = make([]domain.DepositCheckData, 0)
	}

	err = domain.OperatorHistory{}.Record(c.Request().Context(), userId, email,
		"menu.manage_accounting.receiving_amount_list", "", c.RealIP(), domain.LOOKUP)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, apirender.ArrayResult{
		Items:      items,
		TotalCount: cnt,
	})
}

func (DepositCheckApiController) GetOne(c echo.Context) error {
	var item *domain.DepositCheckData
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}
	userId, email, err := jwthelper.TokenToUser(c)
	if err != nil {
		return err
	}

	item, err = domain.DepositCheckData{}.GetById(c.Request().Context(), id)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	if item == nil {
		return factory.ErrorNotFound.New(err)
	}

	err = domain.OperatorHistory{}.Record(c.Request().Context(), userId, email,
		"menu.manage_accounting.receiving_amount", "ID :"+item.ShareId, c.RealIP(), domain.LOOKUP)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	return apirender.RenderSucc(c, http.StatusOK, item)
}

func (DepositCheckApiController) GetDownlaod(c echo.Context) error {
	userId, email, err := jwthelper.TokenToUser(c)
	if err != nil {
		return err
	}

	var v DepositCheckSearchInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}

	factory.Logger(c.Request().Context()).WithFields(logrus.Fields{
		"sortby": v.Sortby,
		"order":  v.Order,
	}).Info("SearchStart")

	err = domain.DepositCheckData{}.DownloadCsv(c.Request().Context(), c.Response().Writer, v.Sortby, v.Order, v.InvestorId, v.InvestmentName, v.DepositCheckStatus, v.StartInvestmentAmount, v.EndInvestmentAmount, v.StartDepositAmount, v.EndDepositAmount)
	if err != nil {
		return factory.ErrorFailDownloadCSV.New(err)
	}

	err = domain.OperatorHistory{}.Record(c.Request().Context(), userId, email,
		"menu.manage_accounting.receiving_amount_list", "", c.RealIP(), domain.DOWNLOAD)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return nil
}

func (DepositMatchDataApiController) GetAll(c echo.Context) error {
	var items []domain.DepositMatchData
	var v DepositMatchSearchInput

	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if v.MaxResultCount == 0 {
		v.MaxResultCount = DefaultMaxResultCount
	}

	userId, email, err := jwthelper.TokenToUser(c)
	if err != nil {
		return err
	}

	factory.Logger(c.Request().Context()).WithFields(logrus.Fields{
		"sortby":           v.Sortby,
		"order":            v.Order,
		"max_result_count": v.MaxResultCount,
		"skip_count":       v.SkipCount,
	}).Info("SearchStart")

	cnt, items, err := domain.DepositMatchData{}.GetAll(c.Request().Context(), v.Sortby, v.Order, v.SkipCount, v.MaxResultCount, v.InvestorId, v.InvestorName, v.DepositMatchStatus, v.StartDepositMatchAmount, v.EndDepositMatchAmount, v.StartDepositAmount, v.EndDepositAmount)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	if items == nil {
		items = make([]domain.DepositMatchData, 0)
	}

	err = domain.OperatorHistory{}.Record(c.Request().Context(), userId, email,
		"menu.manage_accounting.excess_and_shortage_list", "", c.RealIP(), domain.LOOKUP)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, apirender.ArrayResult{
		Items:      items,
		TotalCount: cnt,
	})
}

func (DepositMatchDataApiController) GetOne(c echo.Context) error {

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}
	userId, email, err := jwthelper.TokenToUser(c)
	if err != nil {
		return err
	}

	item, err := domain.DepositMatchData{}.GetById(c.Request().Context(), id)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	if item == nil {
		return factory.ErrorNotFound.New(err)
	}

	err = domain.OperatorHistory{}.Record(c.Request().Context(), userId, email,
		"menu.manage_accounting.excess_and_shortage_detail", "ID :"+strconv.FormatInt(item.UserId, 10), c.RealIP(), domain.LOOKUP)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, item)
}

func (DepositMatchDataApiController) GetDownlaod(c echo.Context) error {
	userId, email, err := jwthelper.TokenToUser(c)
	if err != nil {
		return err
	}

	var v DepositMatchSearchInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}

	factory.Logger(c.Request().Context()).WithFields(logrus.Fields{
		"sortby": v.Sortby,
		"order":  v.Order,
	}).Info("SearchStart")

	err = domain.DepositMatchData{}.DownloadCsv(c.Request().Context(), c.Response().Writer, v.Sortby, v.Order, v.InvestorId, v.InvestorName, v.DepositMatchStatus, v.StartDepositMatchAmount, v.EndDepositMatchAmount, v.StartDepositAmount, v.EndDepositAmount)
	if err != nil {
		return factory.ErrorFailDownloadCSV.New(err)
	}

	err = domain.OperatorHistory{}.Record(c.Request().Context(), userId, email,
		"menu.manage_accounting.excess_and_shortage_list", "", c.RealIP(), domain.DOWNLOAD)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	return nil
}
