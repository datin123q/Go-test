package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	openapigo "github.com/swaggest/openapi-go"

	adomain "github.com/illuminarean/energy-share-core-api/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/echoutils"
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
		SetSummary("헬스 체크").
		SetDescription("서버 헬스 체크 요청에 대한 결과를 반환합니다.").
		AddResponse(adomain.HealthCheckResponse{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "헬스 체크 응답값"
		})

	return handler
}

func (d healthHttpDelivery) HealthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, adomain.HealthCheckResponse{
		Status: "UP",
	})
}
