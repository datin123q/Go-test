package http

import (
	"net/http"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/echoutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	udomain "github.com/illuminarean/energyshares-user-api/domain"
	"github.com/labstack/echo/v4"
	openapigo "github.com/swaggest/openapi-go"
)

type codeHttpDelivery struct {
	codeUsecase udomain.CodeUsecase
}

func New(
	api *echoutils.EchoSwagger,
	codeUsecase udomain.CodeUsecase,
) *codeHttpDelivery {
	handler := &codeHttpDelivery{
		codeUsecase: codeUsecase,
	}

	g := api.Group("/api/v1/codes")

	g.GET("", handler.GetAll).
		SetTags("コード").
		SetSummary("共通コードを照会します").
		AddResponse(struct {}{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "共通コードリスト応答値"
		}).
		AddErrors(
			factory.ErrorDB,
		)

	return handler
}
func (d codeHttpDelivery) GetAll(c echo.Context) error {
	ctx := c.Request().Context()
	return c.Blob(http.StatusOK, echo.MIMEApplicationJSON, []byte(d.codeUsecase.GetAll(ctx)))
}


