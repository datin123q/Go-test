package controllers

import (
	"net/http"
	"strconv"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/apirender"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
	"github.com/sirupsen/logrus"
)

type UserInfoHistoryApiController struct{}

func (c UserInfoHistoryApiController) Init(g echoswagger.ApiGroup) {
	g.GET("", c.GetAll).
		AddParamQueryNested(UserInfoHistorySearchInput{}).
		SetDescription("会員ログリスト情報をリクエストしています。").
		SetSummary("会員ログリスト情報リクエスト").
		SetSecurity("Authorization")

	g.GET("/content/:id", c.GetContent).
		AddParamPath("", "id", "").
		SetDescription("変更された詳細内容情報をリクエストしています。").
		SetSummary("変更された詳細内容情報リクエスト").
		SetSecurity("Authorization")
}

func (UserInfoHistoryApiController) GetAll(c echo.Context) error {

	var v UserInfoHistorySearchInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if v.MaxResultCount == 0 {
		v.MaxResultCount = DefaultMaxResultCount
	}

	// console log
	factory.Logger(c.Request().Context()).WithFields(logrus.Fields{
		"userId":           v.UserId,
		"sortby":           v.Sortby,
		"order":            v.Order,
		"max_result_count": v.MaxResultCount,
		"skip_count":       v.SkipCount,
	}).Info("UserInfoHistory SearchStart")

	totalCount, items, err := domain.UserInfoHistory{}.GetAll(c.Request().Context(), v.UserId, v.Sortby, v.Order, v.SkipCount, v.MaxResultCount)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, apirender.ArrayResult{
		TotalCount: totalCount,
		Items:      items,
	})
}

func (UserInfoHistoryApiController) GetContent(c echo.Context) error {

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	ucc, err := domain.UserChangedContent{}.GetById(c.Request().Context(), id)
	if err != nil {
		return factory.ErrorDB.New(err)
	} else if ucc == nil {
		return factory.ErrorNotFound.New(nil)
	}

	return apirender.RenderSucc(c, http.StatusOK, ucc)
}
