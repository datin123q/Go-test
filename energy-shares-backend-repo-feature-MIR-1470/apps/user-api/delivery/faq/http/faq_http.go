package http

import (
	"net/http"

	openapigo "github.com/swaggest/openapi-go"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/apirender"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/echoutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	udomain "github.com/illuminarean/energyshares-user-api/domain"
	"github.com/labstack/echo/v4"
)

type faqHttpDelivery struct {
	faqUsecase udomain.FaqUsecase
}

func New(
	api *echoutils.EchoSwagger,
	faqUsecase udomain.FaqUsecase,
) *faqHttpDelivery {
	handler := &faqHttpDelivery{
		faqUsecase: faqUsecase,
	}

	g := api.Group("/api/v1/faqs")

	g.GET("", handler.GetAll).
		SetTags("よくある質問").
		SetSummary("よくある質問リストを照会します").
		AddRequest(udomain.FaqGetAllRequest{}).
		AddResponse(udomain.FaqGetAllResponse{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "よくある質問リスト応答値"
		}).
		AddErrors(
			factory.ErrorNotFound,
			factory.ErrorParameter,
			factory.ErrorDB,
		)

	return handler
}

func (d faqHttpDelivery) GetAll(c echo.Context) error {
	ctx := c.Request().Context()

	// Request DTO作成
	var request udomain.FaqGetAllRequest

	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// Usecase呼び出し
	response, err := d.faqUsecase.GetAll(ctx, request)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, response)

}
