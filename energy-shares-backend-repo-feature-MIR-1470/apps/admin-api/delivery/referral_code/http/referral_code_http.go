package http

import (
	"net/http"

	openapigo "github.com/swaggest/openapi-go"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	adomain "github.com/illuminarean/energy-share-core-api/domain"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/apirender"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/echoutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
)

type referralCodeHttpDelivery struct {
	extractor jwthelper.UserExtractor
	usecase   adomain.ReferralCodeUsecase
}

func New(
	api *echoutils.EchoSwagger,
	extractor jwthelper.UserExtractor,
	usecase adomain.ReferralCodeUsecase,
) *referralCodeHttpDelivery {
	handler := &referralCodeHttpDelivery{
		extractor: extractor,
		usecase:   usecase,
	}

	g := api.Group("/api/referral-codes",
		echoutils.AllowRoles(domain.ROLE_OPERATOR))

	g.POST("/issue", handler.Create).
		SetTags("紹介コード").
		SetSummary("紹介コード発行").
		AddRequest(adomain.ReferralCodeCreateRequest{}).
		AddResponse(adomain.ReferralCodeCreateResponse{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusCreated
			cu.Description = "紹介コード発行成功"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorAlreadyExists,
			factory.ErrorParameter,
			factory.ErrorDB,
		).
		AddSecurity("Authorization")

	return handler
}

func (h referralCodeHttpDelivery) Create(c echo.Context) error {
	ctx := c.Request().Context()

	user, err := h.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	request := adomain.ReferralCodeCreateRequest{
		JWTUser: user,
	}

	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	if err := request.Validate(); err != nil {
		return factory.ErrorParameter.New(err)
	}

	response, err := h.usecase.Create(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusCreated, response)
	case errors.Is(err, cdomain.ErrAlreadyExists):
		return factory.ErrorAlreadyExists.New(err)
	case errors.Is(err, cdomain.ErrBadRequest):
		return factory.ErrorParameter.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}
