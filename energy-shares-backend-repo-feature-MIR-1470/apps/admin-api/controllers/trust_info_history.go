package controllers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/apirender"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/echoutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

type TrustInfoHistoryApiController struct{}

func (c TrustInfoHistoryApiController) Init(g echoswagger.ApiGroup) {

	g.GET("", c.GetAll,
		echoutils.AllowRoles(domain.ROLE_PAYMENT)).
		AddParamQueryNested(SearchInput{}).
		SetDescription("信託情報ログリスト照会").
		SetSummary("信託情報ログリスト照会").
		SetSecurity("Authorization")
}

func (TrustInfoHistoryApiController) GetAll(c echo.Context) error {
	var v SearchInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if v.MaxResultCount == 0 {
		v.MaxResultCount = DefaultMaxResultCount
	}

	totalCount, items, err := domain.TrustInfoHistory{}.GetAll(c.Request().Context(), v.Sortby, v.Order, v.SkipCount, v.MaxResultCount)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	latestHistory, err := domain.TrustInfoHistory{}.GetLatest(c.Request().Context())

	type result struct {
		Items         []domain.TrustInfoHistory `json:"items"`
		TotalCount    int64                     `json:"total_count"`
		LastPrintDate jsonutils.JSONTime        `json:"lastPrintDate"`
	}

	return apirender.RenderSucc(c, http.StatusOK, result{
		TotalCount:    totalCount,
		Items:         items,
		LastPrintDate: latestHistory.Created,
	})
}
