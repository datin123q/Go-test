package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	openapigo "github.com/swaggest/openapi-go"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/echoutils"
	udomain "github.com/illuminarean/energyshares-user-api/domain"
)

type healthHttpDelivery struct {
}

func New(
	api *echoutils.EchoSwagger,
) *healthHttpDelivery {
	handler := &healthHttpDelivery{}

	g := api.Group("/api/health-check")

	g.GET("", handler.HealthCheck).
		SetTags("Common").
		SetSummary("ヘルスチェック").
		SetDescription("サーバーヘルスチェックリクエストに対する結果を返します。").
		AddResponse(udomain.HealthCheckResponse{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "ヘルスチェック応答値"
		})

	return handler
}

func (d healthHttpDelivery) HealthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, udomain.HealthCheckResponse{
		Status: "UP",
	})
}
