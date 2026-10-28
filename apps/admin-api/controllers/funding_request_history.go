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

type FundingRequestHistoryApiController struct {
}

func (c FundingRequestHistoryApiController) Init(g echoswagger.ApiGroup) {
	g.GET("", c.GetAll,
		echoutils.AllowRoles(domain.ROLE_PRODUCT)).
		AddParamQueryNested(FundingRequestHistorySearchInput{}).
		AddResponse(http.StatusOK, "ファンディング状態変更履歴リスト照会", domain.FundingRequestHistory{}, nil).
		SetDescription("ファンディング状態変更リストを照会します.").
		SetSummary("ファンディング状態変更リスト照会").
		SetSecurity("Authorization")
}

func (FundingRequestHistoryApiController) GetAll(c echo.Context) error {
	var s FundingRequestHistorySearchInput
	if err := c.Bind(&s); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if s.MaxResultCount == 0 {
		s.MaxResultCount = DefaultMaxResultCount
	}

	factory.Logger(c.Request().Context()).WithFields(logrus.Fields{
		"fundingRequestId": s.FundingRequestId,
		"sortby":           s.Sortby,
		"order":            s.Order,
		"max_result_count": s.MaxResultCount,
		"skip_count":       s.SkipCount,
	}).Info("Funding History Search Start")

	var items []domain.FundingRequestHistory
	totalCount, items, err := domain.FundingRequestHistory{}.GetAll(c.Request().Context(), s.FundingRequestId, s.Sortby, s.Order, s.SkipCount, s.MaxResultCount)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, apirender.ArrayResult{
		Items:      items,
		TotalCount: totalCount,
	})
}
