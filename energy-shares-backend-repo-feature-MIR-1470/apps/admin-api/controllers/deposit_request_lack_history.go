package controllers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/apirender"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/echoutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

type DepositRequestLackHistoryApiController struct{}

func (c DepositRequestLackHistoryApiController) Init(g echoswagger.ApiGroup) {

	g.GET("/:id", c.GetOne,
		echoutils.AllowRoles(domain.ROLE_PAYMENT)).
		AddParamQueryNested(SearchInput{}).
		AddParamPath("", "id", "").
		SetDescription("不足入金要求の詳細情報を提供します.").
		AddResponse(http.StatusOK, "不足入金要求データ", domain.DepositRequestLackHistory{}, nil).
		SetSummary("不足入金要求照会").
		SetSecurity("Authorization")
}

func (DepositRequestLackHistoryApiController) GetOne(c echo.Context) error {

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	var v SearchInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if v.MaxResultCount == 0 {
		v.MaxResultCount = DefaultMaxResultCount
	}

	totalCount, items, err := domain.DepositRequestLackHistory{}.GetByUserId(c.Request().Context(), id, v.Sortby, v.Order, v.SkipCount, v.MaxResultCount)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, apirender.ArrayResult{
		TotalCount: totalCount,
		Items:      items,
	})
}
