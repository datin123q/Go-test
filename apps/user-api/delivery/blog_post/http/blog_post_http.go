package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/apirender"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	udomain "github.com/illuminarean/energyshares-user-api/domain"
)

type blogPostHttpDelivery struct {
	blogPostUsecase udomain.BlogPostUsecase
}

func New(
	api echoswagger.ApiRoot,
	blogPostUsecase udomain.BlogPostUsecase,
) *blogPostHttpDelivery {
	handler := &blogPostHttpDelivery{
		blogPostUsecase: blogPostUsecase,
	}

	g := api.Group("v1/blog-posts", "api/v1/blog-posts")
	g.GET("", handler.GetAll).
		AddParamBody(udomain.BlogPostGetAllRequest{}, "body", "ブログポストリクエスト", true).
		AddResponse(http.StatusOK, "ブログポスト応答", udomain.BlogPostGetAllResponse{}, nil).
		SetDescription("ブログポストの投稿リストを照会します。").
		SetSummary("ブログ投稿返却")
	return handler
}

func (i blogPostHttpDelivery) GetAll(c echo.Context) error {
	ctx := c.Request().Context()

	// Request DTO作成
	request := udomain.BlogPostGetAllRequest{}
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// Usecase呼び出し
	_, response, err := i.blogPostUsecase.GetAll(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, response)
	case errors.Is(err, cdomain.ErrItemNotFound):
		return factory.ErrorNotFound.New(err)
	case errors.Is(err, cdomain.ErrBadRequest):
		return factory.ErrorParameter.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}
