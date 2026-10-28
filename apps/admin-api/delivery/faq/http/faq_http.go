package http

import (
	"net/http"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/echoutils"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	openapigo "github.com/swaggest/openapi-go"

	adomain "github.com/illuminarean/energy-share-core-api/domain"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/apirender"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
)

type faqHttpDelivery struct {
	faqUsecase adomain.FaqUsecase
	extractor  jwthelper.UserExtractor
}

func New(
	api *echoutils.EchoSwagger,
	faqUsecase adomain.FaqUsecase,
	extractor jwthelper.UserExtractor,
) *faqHttpDelivery {
	handler := &faqHttpDelivery{
		faqUsecase: faqUsecase,
		extractor:  extractor,
	}

	g := api.Group("/api/v1/faqs",
		echoutils.AllowRoles(domain.ROLE_OPERATOR))

	g.POST("", handler.Create).
		SetTags("자주묻는질문").
		SetSummary("자주묻는질문을 생성합니다.").
		AddRequest(adomain.FaqCreateRequest{}).
		AddResponse(struct {
			Result  string `json:"result" example:"success"`
			Success bool   `json:"success" example:"true"`
		}{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "생성 성공 응답 값"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorOverMaxPin,
			factory.ErrorParameter,
			factory.ErrorDB,
		).
		AddSecurity("Authorization")

	g.PUT("/:id", handler.Update).
		SetTags("자주묻는질문").
		SetSummary("자주묻는질문을 업데이트합니다.").
		AddRequest(adomain.FaqUpdateRequest{}).
		AddResponse(struct {
			Result  string `json:"result" example:"success"`
			Success bool   `json:"success" example:"true"`
		}{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "업데이트 성공 응답 값"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorOverMaxPin,
			factory.ErrorNotFound,
			factory.ErrorParameter,
			factory.ErrorDB,
		).
		AddSecurity("Authorization")

	g.GET("", handler.GetAll).
		SetTags("자주묻는질문").
		SetSummary("자주묻는질문 리스트를 조회합니다").
		AddRequest(adomain.FaqGetAllRequest{}).
		AddResponse(adomain.FaqGetAllResponse{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "자주묻는질문 리스트 응답 값"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorParameter,
			factory.ErrorDB,
		).
		AddSecurity("Authorization")

	g.GET("/:id", handler.GetOne).
		SetTags("자주묻는질문").
		SetSummary("자주묻는질문에 대해서 단건 조회합니다").
		AddRequest(adomain.FaqGetOneRequest{}).
		AddResponse(adomain.FaqGetOneResponse{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "자주묻는질문 단건 응답 값"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorNotFound,
			factory.ErrorParameter,
			factory.ErrorDB,
		).
		AddSecurity("Authorization")

	g.DELETE("/:id", handler.Delete).
		SetTags("자주묻는질문").
		SetSummary("자주묻는질문을 삭제합니다").
		AddRequest(adomain.FaqDeleteRequest{}).
		AddResponse(struct {
			Result  string `json:"result" example:"success"`
			Success bool   `json:"success" example:"true"`
		}{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "삭제성공 응답 값"
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

func (fhd faqHttpDelivery) GetAll(c echo.Context) error {
	ctx := c.Request().Context()

	// 유저 정보 추출
	user, err := fhd.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	request := adomain.FaqGetAllRequest{
		JWTUser:  user,
		ClientIp: c.RealIP(),
	}

	if err = c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	response, err := fhd.faqUsecase.GetAll(ctx, request)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, response)
}

func (fhd faqHttpDelivery) GetOne(c echo.Context) error {
	ctx := c.Request().Context()

	// 유저 정보 추출
	user, err := fhd.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	request := adomain.FaqGetOneRequest{
		JWTUser:  user,
		ClientIp: c.RealIP(),
	}

	if err = c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	response, err := fhd.faqUsecase.GetOne(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, response)
	case errors.Is(err, cdomain.ErrItemNotFound):
		return factory.ErrorNotFound.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (fhd faqHttpDelivery) Delete(c echo.Context) error {
	ctx := c.Request().Context()

	// 유저 정보 추출
	user, err := fhd.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	request := adomain.FaqDeleteRequest{
		JWTUser:  user,
		ClientIp: c.RealIP(),
	}

	if err = c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	err = fhd.faqUsecase.Delete(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, nil)
	case errors.Is(err, cdomain.ErrItemNotFound):
		return factory.ErrorNotFound.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (fhd faqHttpDelivery) Create(c echo.Context) error {
	ctx := c.Request().Context()

	user, err := fhd.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	request := adomain.FaqCreateRequest{
		JWTUser:  user,
		ClientIp: c.RealIP(),
	}

	if err = c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	if err = c.Validate(request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	err = fhd.faqUsecase.Create(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, nil)
	case errors.Is(err, cdomain.ErrRequestLimitExceeded):
		return factory.ErrorOverMaxPin.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (fhd faqHttpDelivery) Update(c echo.Context) error {
	ctx := c.Request().Context()

	user, err := fhd.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	request := adomain.FaqUpdateRequest{
		JWTUser:  user,
		ClientIp: c.RealIP(),
	}

	if err = c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	if err = c.Validate(request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	err = fhd.faqUsecase.Update(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, nil)
	case errors.Is(err, cdomain.ErrRequestLimitExceeded):
		return factory.ErrorOverMaxPin.New(err)
	case errors.Is(err, cdomain.ErrItemNotFound):
		return factory.ErrorNotFound.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}
