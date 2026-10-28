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

type bannerHttpDelivery struct {
	bannerUsecase udomain.BannerUsecase
}

func New(
	api echoswagger.ApiRoot,
	bannerUsecase udomain.BannerUsecase,
) *bannerHttpDelivery {
	handler := &bannerHttpDelivery{
		bannerUsecase: bannerUsecase,
	}

	g := api.Group("v1/banners", "api/v1/banners")

	g.GET("", handler.GetAll).
		AddResponse(http.StatusOK, "掲載中のバナーリスト照会", udomain.BannerGetAllResponse{}, nil).
		SetDescription("掲載中のバナーリストを照会します。").
		SetSummary("掲載中のバナーリスト照会")

	return handler
}

func (b bannerHttpDelivery) GetAll(c echo.Context) error {
	ctx := c.Request().Context()

	// call usecase & handle error
	response, err := b.bannerUsecase.GetAll(ctx)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, response)
	case errors.Is(err, cdomain.ErrItemNotFound):
		return factory.ErrorNotFound.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}
