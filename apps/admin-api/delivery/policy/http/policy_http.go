package http

import (
	"net/http"

	adomain "github.com/illuminarean/energy-share-core-api/domain"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/apirender"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/echoutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	openapigo "github.com/swaggest/openapi-go"
)

const (
	_DefaultMaxResultCount = 30
)

type policyHttpDelivery struct {
	extractor     jwthelper.UserExtractor
	policyUsecase adomain.PolicyUsecase
}

func New(
	api *echoutils.EchoSwagger,
	extractor jwthelper.UserExtractor,
	policyUsecase adomain.PolicyUsecase,
) *policyHttpDelivery {
	handler := &policyHttpDelivery{
		extractor:     extractor,
		policyUsecase: policyUsecase,
	}

	g := api.Group("/api/v1/policies", echoutils.AllowRoles(domain.ROLE_PRODUCT))

	g.PUT("/:id", handler.Update).
		SetTags("약관").
		SetSummary("약관을 수정합니다").
		AddRequest(adomain.PolicyUpdateRequest{}).
		AddResponse(adomain.PolicyUpdateResponse{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "업데이트 성공 응답 값"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorNotFound,
			factory.ErrorParameter,
			factory.ErrorDB,
		).
		AddSecurity("Authorization")

	g.GET("", handler.GetAll).
		SetTags("약관").
		SetSummary("약관의 리스트를 조회합니다").
		AddRequest(adomain.PolicyGetAllRequest{}).
		AddResponse(adomain.PolicyGetAllResponse{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "약관 리스트 응답 값"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorParameter,
			factory.ErrorDB,
		).
		AddSecurity("Authorization")

	g.GET("/:id", handler.GetOne).
		SetTags("약관").
		SetSummary("약관에 대해서 단건 조회합니다").
		AddRequest(adomain.PolicyGetOneRequest{}).
		AddResponse(adomain.PolicyGetOneResponse{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "약관 단건 응답 값"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorNotFound,
			factory.ErrorParameter,
			factory.ErrorDB,
		).
		AddSecurity("Authorization")

	g.GET("/:id/histories", handler.GetAllHistory).
		SetTags("약관").
		SetSummary("약관의 히스토리를 조회합니다").
		AddRequest(adomain.PolicyGetAllHistoryRequest{}).
		AddResponse(adomain.PolicyGetAllHistoryResponse{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "약관 히스토리 리스트 응답 값"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorParameter,
			factory.ErrorDB,
		).
		AddSecurity("Authorization")

	g.PUT("/:id/histories/:history_id", handler.UpdateHistory).
		SetTags("약관").
		SetSummary("약관의 히스토리를 수정합니다.").
		AddRequest(adomain.PolicyUpdateHistoryRequest{}).
		AddResponse(adomain.PolicyUpdateHistoryResponse{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "약관 히스토리 업데이트 응답 값"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorNotFound,
			factory.ErrorParameter,
			factory.ErrorDB,
		).
		AddSecurity("Authorization")

	g.DELETE("/:id", handler.Delete).
		SetTags("약관").
		SetSummary("약관을 삭제합니다").
		AddRequest(adomain.PolicyDeleteRequest{}).
		AddResponse(struct {
			Result  string `json:"result" example:"success"`
			Success bool   `json:"success" example:"true"`
		}{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "삭제 성공 응답 값"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorNotFound,
			factory.ErrorParameter,
			factory.ErrorDB,
		).
		AddSecurity("Authorization")

	g.GET("/:id/check-label", handler.CheckLabelDuplicate).
		SetTags("약관").
		SetSummary("약관에 중복되는 라벨이 있는지 확인합니다").
		AddRequest(adomain.PolicyCheckLabelDuplicateRequest{}).
		AddResponse(adomain.PolicyCheckLabelDuplicateResponse{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "약관 단건 응답 값"
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

func (h *policyHttpDelivery) Update(c echo.Context) error {
	ctx := c.Request().Context()

	user, err := h.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(nil)
	}

	// create & binding Request DTO
	var request = adomain.PolicyUpdateRequest{
		JWTUser: user,
	}
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// call usecase & handle error
	response, err := h.policyUsecase.Update(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, response)
	case errors.Is(err, cdomain.ErrItemNotFound):
		return factory.ErrorNotFound.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (h *policyHttpDelivery) GetAll(c echo.Context) error {
	ctx := c.Request().Context()

	user, err := h.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	// create & binding Request DTO
	var request = adomain.PolicyGetAllRequest{
		JWTUser: user,
		PolicyGetAllQueryParam: adomain.PolicyGetAllQueryParam{
			SearchInput: cdomain.SearchInput{
				MaxResultCount: _DefaultMaxResultCount,
			},
		},
	}
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// call usecase & handle error
	response, err := h.policyUsecase.GetAll(ctx, request)

	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, response)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (h *policyHttpDelivery) GetOne(c echo.Context) error {
	ctx := c.Request().Context()

	user, err := h.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	// create & binding Request DTO
	var request = adomain.PolicyGetOneRequest{
		JWTUser: user,
	}
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// call usecase & handle error
	response, err := h.policyUsecase.GetOne(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, response)
	case errors.Is(err, cdomain.ErrItemNotFound):
		return factory.ErrorNotFound.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (h *policyHttpDelivery) GetAllHistory(c echo.Context) error {
	ctx := c.Request().Context()

	user, err := h.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	// create & binding Request DTO
	var request = adomain.PolicyGetAllHistoryRequest{
		JWTUser: user,
		PolicyGetAllHistoryQueryParam: adomain.PolicyGetAllHistoryQueryParam{
			SearchInput: cdomain.SearchInput{
				MaxResultCount: _DefaultMaxResultCount,
			},
		},
	}
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// call usecase & handle error
	response, err := h.policyUsecase.GetAllHistory(ctx, request)

	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, response)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (h *policyHttpDelivery) UpdateHistory(c echo.Context) error {
	ctx := c.Request().Context()

	user, err := h.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(nil)
	}

	// create & binding Request DTO
	var request = adomain.PolicyUpdateHistoryRequest{
		JWTUser: user,
	}
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := request.Valid(c); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// call usecase & handle error
	response, err := h.policyUsecase.UpdateHistory(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, response)
	case errors.Is(err, cdomain.ErrItemNotFound):
		return factory.ErrorNotFound.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (h *policyHttpDelivery) Delete(c echo.Context) error {
	ctx := c.Request().Context()

	user, err := h.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	// create & binding Request DTO
	var request = adomain.PolicyDeleteRequest{
		JWTUser: user,
	}
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// call usecase & handle error
	err = h.policyUsecase.Delete(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, nil)
	case errors.Is(err, cdomain.ErrItemNotFound):
		return factory.ErrorNotFound.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (h *policyHttpDelivery) CheckLabelDuplicate(c echo.Context) error {
	ctx := c.Request().Context()

	user, err := h.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	// create & binding Request DTO
	var request = adomain.PolicyCheckLabelDuplicateRequest{
		JWTUser: user,
	}
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// call usecase & handle error
	response, err := h.policyUsecase.CheckLabelDuplicate(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, response)
	case errors.Is(err, cdomain.ErrItemNotFound):
		return factory.ErrorNotFound.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}
