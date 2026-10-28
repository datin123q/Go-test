package controllers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
	"github.com/sirupsen/logrus"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/apirender"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/echoutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

type SpvInfoHistoryApiController struct{}

func (c SpvInfoHistoryApiController) Init(g echoswagger.ApiGroup) {
	g.GET("", c.GetAll,
		echoutils.AllowRoles(domain.ROLE_PRODUCT)).
		AddParamQueryNested(SpvInfoHistorySearchInput{}).
		SetDescription("開発会社ログリスト情報をリクエストしています。").
		SetSummary("開発会社ログリスト情報リクエスト").
		SetSecurity("Authorization")
}

func (SpvInfoHistoryApiController) GetAll(c echo.Context) error {

	var v SpvInfoHistorySearchInput
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
	}).Info("UserInfoHistory SearchStart")

	totalCount, items, err := domain.SpvInfoHistory{}.GetAll(c.Request().Context(), v.SpvId, v.Sortby, v.Order, v.SkipCount, v.MaxResultCount)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, apirender.ArrayResult{
		TotalCount: totalCount,
		Items:      items,
	})
}
