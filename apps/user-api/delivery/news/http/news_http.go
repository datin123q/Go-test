package http

import (
	"net/http"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/echoutils"
	openapigo "github.com/swaggest/openapi-go"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/apirender"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	udomain "github.com/illuminarean/energyshares-user-api/domain"
)

type newsHttpDelivery struct {
	extractor   jwthelper.UserExtractor
	newsUsecase udomain.NewsUsecase
}

func New(
	api *echoutils.EchoSwagger,
	extractor jwthelper.UserExtractor,
	newsUsecase udomain.NewsUsecase,

) *newsHttpDelivery {
	handler := &newsHttpDelivery{
		extractor:   extractor,
		newsUsecase: newsUsecase,
	}

	g := api.Group("/api/v1/news")
	g.GET("", handler.GetAll).
		SetTags("お知らせ").
		SetSummary("お知らせリストを照会します").
		AddRequest(udomain.NewsGetAllRequest{}).
		AddResponse(udomain.NewsGetAllResponse{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "お知らせリスト応答値"
		}).
		AddErrors(
			factory.ErrorNotFound,
			factory.ErrorParameter,
			factory.ErrorDB,
		)
	return handler
}

func (delivery newsHttpDelivery) GetAll(c echo.Context) error {
	ctx := c.Request().Context()

	request := udomain.NewsGetAllRequest{
		ClientIp: c.RealIP(),
	}
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	if err := c.Validate(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// Usecase呼び出し
	response, err := delivery.newsUsecase.GetAll(ctx, request)
	if err != nil {
		if errors.Is(err, cdomain.ErrItemNotFound) {
			return factory.ErrorNotFound.New(err)
		}
		if errors.Is(err, cdomain.ErrBadRequest) {
			return factory.ErrorParameter.New(err)
		}
		return factory.ErrorDB.New(err)
	}
	return apirender.RenderSucc(c, http.StatusOK, response)
}
