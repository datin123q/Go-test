package http

import (
	"net/http"

	openapigo "github.com/swaggest/openapi-go"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/echoutils"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	adomain "github.com/illuminarean/energy-share-core-api/domain"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/apirender"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
)

type newsHttpDelivery struct {
	extractor   jwthelper.UserExtractor
	newsUsecase adomain.NewsUsecase
}

func New(
	api *echoutils.EchoSwagger,
	extractor jwthelper.UserExtractor,
	newsUsecase adomain.NewsUsecase,
) *newsHttpDelivery {
	handler := &newsHttpDelivery{
		extractor:   extractor,
		newsUsecase: newsUsecase,
	}

	g := api.Group("/api/v1/news",
		echoutils.AllowRoles(domain.ROLE_OPERATOR))

	g.GET("", handler.GetAll).
		SetTags("공지사항").
		SetSummary("공지사항 리스트를 조회합니다").
		AddRequest(adomain.NewsGetAllRequest{}).
		AddResponse(adomain.NewsGetAllResponse{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "공지사항 리스트 응답 값"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorNotFound,
			factory.ErrorParameter,
			factory.ErrorDB,
		).
		AddSecurity("Authorization")

	g.POST("", handler.Create).
		SetTags("공지사항").
		SetSummary("공지사항을 생성합니다.").
		AddRequest(adomain.NewsCreateRequest{}).
		AddResponse(struct {
			Result  string `json:"result" example:"success"`
			Success bool   `json:"success" example:"true"`
		}{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "공지사항 생성성공 응답 값"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorOverMaxPin,
			factory.ErrorParameter,
			factory.ErrorDB,
		).
		AddSecurity("Authorization")

	g.PUT("/:id", handler.Update).
		SetTags("공지사항").
		SetSummary("공지사항을 수정합니다.").
		AddRequest(adomain.NewsUpdateRequest{}).
		AddResponse(struct {
			Result  string `json:"result" example:"success"`
			Success bool   `json:"success" example:"true"`
		}{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "공지사항 수정성공 응답 값"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorOverMaxPin,
			factory.ErrorNotFound,
			factory.ErrorParameter,
			factory.ErrorDB,
		).
		AddSecurity("Authorization")

	g.GET("/:id", handler.GetOne).
		SetTags("공지사항").
		SetSummary("공지사항에 대해서 단건 조회합니다").
		AddRequest(adomain.NewsGetOneRequest{}).
		AddResponse(adomain.NewsGetOneResponse{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "공지사항 단건 응답 값"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorNotFound,
			factory.ErrorParameter,
			factory.ErrorDB,
		).
		AddSecurity("Authorization")

	g.DELETE("/:id", handler.Delete).
		SetTags("공지사항").
		SetSummary("공지사항을 삭제합니다.").
		AddRequest(adomain.NewsDeleteRequest{}).
		AddResponse(struct {
			Result  string `json:"result" example:"success"`
			Success bool   `json:"success" example:"true"`
		}{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "공지사항 삭제성공 응답 값"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorNotFound,
			factory.ErrorParameter,
			factory.ErrorDB,
		).
		AddSecurity("Authorization")

	return handler
}

func (delivery *newsHttpDelivery) GetAll(c echo.Context) error {
	ctx := c.Request().Context()

	user, err := delivery.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	request := adomain.NewsGetAllRequest{
		JWTUser:  user,
		ClientIp: c.RealIP(),
	}

	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	if err := c.Validate(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// Usecase 호출
	response, err := delivery.newsUsecase.GetAll(ctx, request)
	if err != nil {
		if errors.Is(err, cdomain.ErrItemNotFound) {
			return factory.ErrorNotFound.New(err)
		}
		return factory.ErrorDB.New(err)
	}
	return apirender.RenderSucc(c, http.StatusOK, response)
}

func (delivery newsHttpDelivery) GetOne(c echo.Context) error {
	ctx := c.Request().Context()

	user, err := delivery.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	request := adomain.NewsGetOneRequest{
		JWTUser:  user,
		ClientIp: c.RealIP(),
	}

	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// Usecase 호출
	response, err := delivery.newsUsecase.GetOne(ctx, request)
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

func (delivery newsHttpDelivery) Create(c echo.Context) error {
	ctx := c.Request().Context()
	user, err := delivery.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	request := adomain.NewsCreateRequest{
		JWTUser:  user,
		ClientIp: c.RealIP(),
	}

	if err = c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	if err = c.Validate(request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	err = delivery.newsUsecase.Create(ctx, request)
	if err != nil {
		if errors.Is(err, cdomain.ErrRequestLimitExceeded) {
			return factory.ErrorOverMaxPin.New(err)
		}
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, nil)
}

func (delivery newsHttpDelivery) Update(c echo.Context) error {

	ctx := c.Request().Context()
	user, err := delivery.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	request := adomain.NewsUpdateRequest{
		JWTUser:  user,
		ClientIp: c.RealIP(),
	}

	if err = c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	if err = c.Validate(request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	err = delivery.newsUsecase.Update(ctx, request)
	if err != nil {
		if errors.Is(err, cdomain.ErrRequestLimitExceeded) {
			return factory.ErrorOverMaxPin.New(err)
		} else if errors.Is(err, cdomain.ErrItemNotFound) {
			return factory.ErrorNotFound.New(err)
		}
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, nil)
}

func (delivery newsHttpDelivery) Delete(c echo.Context) error {
	ctx := c.Request().Context()

	user, err := delivery.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	request := adomain.NewsDeleteRequest{
		JWTUser:  user,
		ClientIp: c.RealIP(),
	}

	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// Usecase 호출
	err = delivery.newsUsecase.Delete(ctx, request)
	if err != nil {
		if errors.Is(err, cdomain.ErrItemNotFound) {
			return factory.ErrorNotFound.New(err)
		}
		if errors.Is(err, cdomain.ErrBadRequest) {
			return factory.ErrorParameter.New(err)
		}
		return factory.ErrorDB.New(err)
	}
	return apirender.RenderSucc(c, http.StatusOK, nil)
}
