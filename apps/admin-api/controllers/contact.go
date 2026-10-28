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
)

type ContactApiController struct {
}

func (c ContactApiController) Init(g echoswagger.ApiGroup) {

	g.GET("", c.GetAll,
		echoutils.AllowRoles(domain.ROLE_INVESTOR)).
		AddParamQueryNested(ContactSearchInput{}).
		SetDescription("問い合わせリスト照会").
		SetSummary("問い合わせリスト照会").
		SetSecurity("Authorization")

	g.GET("/:id", c.GetOne,
		echoutils.AllowRoles(domain.ROLE_INVESTOR)).
		AddParamPath("", "id", "").
		SetDescription("問い合わせ詳細照会").
		SetSummary("問い合わせ詳細照会").
		SetSecurity("Authorization")
}

func (ContactApiController) GetAll(c echo.Context) error {

	var v ContactSearchInput
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
	}).Info("ContactSearchStart")

	totalCount, items, err := domain.Contact{}.GetAll(c.Request().Context(), v.Sortby, v.Order, v.SkipCount, v.MaxResultCount, v.Name, v.Title, v.Email, v.StartDate, v.EndDate)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, apirender.ArrayResult{
		Items:      items,
		TotalCount: totalCount,
	})
}

func (ContactApiController) GetOne(c echo.Context) error {

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	contact, err := domain.Contact{}.GetById(c.Request().Context(), id)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	if contact == nil {
		return factory.ErrorNotFound.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, contact)
}
