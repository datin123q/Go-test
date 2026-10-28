package controllers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
	"github.com/sirupsen/logrus"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/apirender"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
)

type LoginHistoryApiController struct{}

func (c LoginHistoryApiController) Init(g echoswagger.ApiGroup) {
	g.GET("", c.GetAll).
		AddParamQueryNested(LoginHistorySearchInput{}).
		AddResponse(http.StatusOK, "ユーザーアクセス履歴リスト", domain.LoginHistory{}, nil).
		SetDescription("ユーザーのログイン履歴リストを照会します.").
		SetSummary("ユーザーログイン履歴照会").
		SetSecurity("Authorization")

	g.GET("/download", c.Download).
		AddParamQueryNested(SearchInput{}).
		AddResponse(http.StatusOK, "ユーザーアクセス履歴リストダウンロード", domain.LoginHistory{}, nil).
		SetDescription("ユーザーのログイン履歴をダウンロードします.").
		SetSummary("ユーザーログイン履歴ダウンロード").
		SetSecurity("Authorization")
}

func (LoginHistoryApiController) GetAll(c echo.Context) error {
	var s LoginHistorySearchInput
	if err := c.Bind(&s); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if s.MaxResultCount == 0 {
		s.MaxResultCount = DefaultMaxResultCount
	}

	userId, email, err := jwthelper.TokenToUser(c)
	if err != nil {
		return err
	}

	factory.Logger(c.Request().Context()).WithFields(logrus.Fields{
		"sortby":           s.Sortby,
		"order":            s.Order,
		"max_result_count": s.MaxResultCount,
		"skip_count":       s.SkipCount,
	}).Info("SearchStart")

	var items []domain.LoginHistory
	totalCount, items, err := domain.LoginHistory{}.GetAll(c.Request().Context(), s.Sortby, s.Order, s.SkipCount, s.MaxResultCount, s.Name, s.StartDate, s.EndDate)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	err = domain.OperatorHistory{}.Record(c.Request().Context(), userId, email,
		"menu.manage_users.user_access_log_list", "", c.RealIP(), domain.LOOKUP)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, apirender.ArrayResult{
		Items:      items,
		TotalCount: totalCount,
	})
}

func (LoginHistoryApiController) Download(c echo.Context) error {
	var v LoginHistorySearchInput
	if err := c.Bind(&v); err != nil {
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

	err = domain.LoginHistory{}.DownloadCsv(c.Request().Context(), c.Response().Writer, v.Sortby, v.Order, v.SkipCount, v.MaxResultCount, v.Name, v.StartDate, v.EndDate)
	if err != nil {
		return factory.ErrorFailDownloadCSV.New(err)
	}

	err = domain.OperatorHistory{}.Record(c.Request().Context(), userId, email,
		"menu.manage_users.user_access_log_list", "", c.RealIP(), domain.DOWNLOAD)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return nil
}
