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

type ProjectInfoHistoryApiController struct{}

func (c ProjectInfoHistoryApiController) Init(g echoswagger.ApiGroup) {
	g.GET("", c.GetAll,
		echoutils.AllowRoles(domain.ROLE_PRODUCT)).
		AddParamQueryNested(ProjectInfoHistorySearchInput{}).
		SetDescription("投資商品ログリスト情報をリクエストしています。").
		SetSummary("投資商品ログリスト情報リクエスト").
		SetSecurity("Authorization")
}

func (ProjectInfoHistoryApiController) GetAll(c echo.Context) error {

	var v ProjectInfoHistorySearchInput
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
	}).Info("ProjectInfoHistory SearchStart")

	totalCount, items, err := domain.ProjectInfoHistory{}.GetAll(c.Request().Context(), v.ProjectId, v.Sortby, v.Order, v.SkipCount, v.MaxResultCount)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, apirender.ArrayResult{
		TotalCount: totalCount,
		Items:      items,
	})
}
