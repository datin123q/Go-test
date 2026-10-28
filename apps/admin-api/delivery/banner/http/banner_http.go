package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
	"github.com/pkg/errors"

	adomain "github.com/illuminarean/energy-share-core-api/domain"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/apirender"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/echoutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
)

type bannerHttpDelivery struct {
	bannerUsecase adomain.BannerUsecase
	extractor     jwthelper.UserExtractor
}

func New(
	api echoswagger.ApiRoot,
	bannerUsecase adomain.BannerUsecase,
	extractor jwthelper.UserExtractor,
) *bannerHttpDelivery {
	handler := &bannerHttpDelivery{
		bannerUsecase: bannerUsecase,
		extractor:     extractor,
	}

	g := api.Group("v1/banners", "api/v1/banners",
		echoutils.AllowRoles(domain.ROLE_OPERATOR))

	g.POST("", handler.Create).
		AddParamBody(adomain.BannerCreateRequest{}, "배너 생성 인풋", "", true).
		AddResponse(http.StatusCreated, "배너 생성", adomain.BannerCreateResponse{}, nil).
		SetDescription("배너를 생성합니다.").
		SetSummary("배너 생성").
		SetSecurity("Authorization")

	g.GET("", handler.GetAll, echoutils.AllowRoles(domain.ROLE_OPERATOR)).
		AddParamQueryNested(adomain.BannerGetAllQueryParam{}).
		AddResponse(http.StatusOK, "배너 목록 조회", adomain.BannerGetAllResponse{}, nil).
		SetDescription("배너 목록을 조회합니다.").
		SetSummary("배너 목록 조회").
		SetSecurity("Authorization")

	g.GET("/:id", handler.GetOne).
		AddParamPath("", "id", "배너 ID").
		AddResponse(http.StatusOK, "배너 상세 조회", adomain.BannerGetOneResponse{}, nil).
		SetDescription("배너 상세 정보를 조회합니다.").
		SetSummary("배너 상세 조회").
		SetSecurity("Authorization")

	g.PUT("/:id", handler.Update).
		AddParamPath("", "id", "배너 ID").
		AddParamBody(adomain.BannerUpdateRequest{}, "body", "배너 수정 인풋", true).
		AddResponse(http.StatusOK, "배너 수정", adomain.BannerUpdateResponse{}, nil).
		SetDescription("배너를 수정합니다.").
		SetSummary("배너 수정").
		SetSecurity("Authorization")

	g.PUT("/rolling-time", handler.UpdateRollingTime).
		AddParamBody(adomain.BannerUpdateRollingTimeRequest{}, "body", "배너 롤링 시간 수정 인풋", true).
		AddResponse(http.StatusOK, "배너 롤링 시간 수정", nil, nil).
		SetDescription("배너 롤링 시간을 수정합니다.").
		SetSummary("배너 롤링 시간 수정").
		SetSecurity("Authorization")

	g.DELETE("/:id", handler.Delete).
		AddParamPath("", "id", "배너 ID").
		AddResponse(http.StatusNoContent, "배너 삭제", nil, nil).
		SetDescription("배너를 삭제합니다.").
		SetSummary("배너 삭제").
		SetSecurity("Authorization")

	return handler
}

func (b bannerHttpDelivery) Create(c echo.Context) error {
	user, err := b.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	ctx := c.Request().Context()

	// create & binding Request DTO
	var request = adomain.BannerCreateRequest{
		JWTUser: user,
	}
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := request.Validate(); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// call usecase & handle error
	response, err := b.bannerUsecase.Create(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, response)
	case errors.Is(err, cdomain.ErrBadRequest):
		return factory.ErrorParameter.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (b bannerHttpDelivery) GetAll(c echo.Context) error {
	user, err := b.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	ctx := c.Request().Context()

	// create & binding Request DTO
	var request = adomain.BannerGetAllRequest{
		JWTUser: user,
	}
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// call usecase & handle error
	response, err := b.bannerUsecase.GetAll(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, response)
	case errors.Is(err, cdomain.ErrItemNotFound):
		return factory.ErrorNotFound.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (b bannerHttpDelivery) GetOne(c echo.Context) error {
	user, err := b.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	ctx := c.Request().Context()

	// create & binding Request DTO
	var request = adomain.BannerGetOneRequest{
		JWTUser: user,
	}
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// call usecase & handle error
	response, err := b.bannerUsecase.GetOne(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, response)
	case errors.Is(err, cdomain.ErrItemNotFound):
		return factory.ErrorNotFound.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (b bannerHttpDelivery) Update(c echo.Context) error {
	user, err := b.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	ctx := c.Request().Context()

	// create & binding Request DTO
	var request = adomain.BannerUpdateRequest{
		JWTUser: user,
	}
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := request.Validate(); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// call usecase & handle error
	response, err := b.bannerUsecase.Update(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, response)
	case errors.Is(err, cdomain.ErrBadRequest):
		return factory.ErrorParameter.New(err)
	case errors.Is(err, cdomain.ErrItemNotFound):
		return factory.ErrorNotFound.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (b bannerHttpDelivery) UpdateRollingTime(c echo.Context) error {
	user, err := b.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	ctx := c.Request().Context()

	// create & binding Request DTO
	var request = adomain.BannerUpdateRollingTimeRequest{
		JWTUser: user,
	}
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// call usecase & handle error
	err = b.bannerUsecase.UpdateRollingTime(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, nil)
	case errors.Is(err, cdomain.ErrItemNotFound):
		return factory.ErrorNotFound.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (b bannerHttpDelivery) Delete(c echo.Context) error {
	user, err := b.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	ctx := c.Request().Context()

	// create & binding Request DTO
	var request = adomain.BannerDeleteRequest{
		JWTUser: user,
	}
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// call usecase & handle error
	err = b.bannerUsecase.Delete(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, nil)
	default:
		return factory.ErrorDB.New(err)
	}
}
