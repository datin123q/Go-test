package http

import (
	"github.com/illuminarean/energy-shares-backend-repo/pkg/apirender"
	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
	"net/http"
)

type commonHttpDelivery struct{}

func New(
	api echoswagger.ApiRoot,
) *commonHttpDelivery {
	handler := &commonHttpDelivery{}

	// 別のドメインを持たない
	g := api.Group("v1/utility", "api/v1")

	g.GET("/csrf", handler.GetCsrfToken).
		AddResponse(http.StatusNoContent, "CSRFトークン照会", "", nil).
		SetSummary("CSRFトークン照会").
		SetDescription("CSRF認証用のトークン照会（ヘッダーに含む）")

	return handler
}

func (d commonHttpDelivery) GetCsrfToken(c echo.Context) error {
	return apirender.RenderSucc(c, http.StatusNoContent, "")
}
