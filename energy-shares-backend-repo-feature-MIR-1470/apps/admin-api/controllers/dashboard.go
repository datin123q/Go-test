package controllers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/apirender"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/echoutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

type DashboardApiController struct{}

func (c DashboardApiController) Init(g echoswagger.ApiGroup) {
	g.GET("", c.GetDashboard,
		echoutils.AllowRoles(domain.ROLE_ADMIN)).
		AddParamQueryNested(SearchInput{}).
		AddResponse(http.StatusOK, "ダッシュボードリスト照会", domain.Dashboard{}, nil).
		SetDescription("ダッシュボードリストを照会します.").
		SetSummary("ダッシュボードリスト照会").
		SetSecurity("Authorization")
}

func (DashboardApiController) GetDashboard(c echo.Context) error {
	var v SearchInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if v.MaxResultCount == 0 {
		v.MaxResultCount = 5
	}

	item, err := domain.Dashboard{}.GetDashboard(c.Request().Context(), v.SkipCount, v.MaxResultCount)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, item)
}
