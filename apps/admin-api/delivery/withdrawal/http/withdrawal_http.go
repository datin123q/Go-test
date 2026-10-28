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

type withdrawalHttpDelivery struct {
	extractor         jwthelper.UserExtractor
	withdrawalUsecase adomain.WithdrawalUsecase
}

func New(
	api *echoutils.EchoSwagger,
	extractor jwthelper.UserExtractor,
	withdrawalUsecase adomain.WithdrawalUsecase,
) *withdrawalHttpDelivery {
	handler := &withdrawalHttpDelivery{
		extractor:         extractor,
		withdrawalUsecase: withdrawalUsecase,
	}

	g := api.Group("/api/v1/withdrawals", echoutils.AllowRoles(domain.ROLE_PAYMENT))

	g.POST("", handler.Create).
		SetTags("出金依頼").
		SetSummary("出金依頼を作成します。").
		AddRequest(adomain.WithdrawalCreateRequest{}).
		AddResponse(adomain.WithdrawalCreateResponse{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusCreated
			cu.Description = "出金依頼作成の応答値"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorParameter,
			factory.ErrorIllegalRequest,
			factory.ErrorNotFound,
			factory.ErrorDB,
		).
		AddSecurity("Authorization")

	g.GET("", handler.GetAll).
		SetTags("출금의뢰").
		SetSummary("출금의뢰 목록을 조회합니다.").
		AddRequest(adomain.WithdrawalGetAllRequest{}).
		AddResponse(adomain.WithdrawalGetAllResponse{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "출금의뢰 목록 조회 응답 값"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorParameter,
			factory.ErrorDB,
		).
		AddSecurity("Authorization")

	g.GET("/:id", handler.GetOne).
		SetTags("출금의뢰").
		SetSummary("출금의뢰를 조회합니다.").
		AddRequest(adomain.WithdrawalGetOneRequest{}).
		AddResponse(adomain.WithdrawalGetOneResponse{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "출금의뢰 조회 응답 값"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorParameter,
			factory.ErrorNotFound,
			factory.ErrorDB,
		).
		AddSecurity("Authorization")

	g.PATCH("/:id", handler.UpdateStatus).
		SetTags("출금의뢰").
		SetSummary("출금의뢰의 상태를 수정합니다. [상태: 송금대기(마감)|취소|완료]").
		AddRequest(adomain.WithdrawalUpdateStatusRequest{}).
		AddResponse(adomain.WithdrawalUpdateStatusResponse{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "출금의뢰 상태 수정 응답 값"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorIllegalRequest,
			factory.ErrorParameter,
			factory.ErrorNotFound,
			factory.ErrorDB,
		).
		AddSecurity("Authorization")

	g.POST("/:id/transfer", handler.RequestTransfer).
		SetTags("출금의뢰").
		SetSummary("출금의뢰를 송금 요청합니다. (실패 상태인 것만 재요청하는 기능)").
		AddRequest(adomain.WithdrawalRequestTransferRequest{}).
		AddResponse(adomain.WithdrawalRequestTransferResponse{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "출금의뢰의 송금요청 응답 값"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorParameter,
			factory.ErrorNotFound,
			factory.ErrorDB,
		).
		AddSecurity("Authorization")

	return handler
}

func (d *withdrawalHttpDelivery) Create(c echo.Context) error {
	ctx := c.Request().Context()

	user, err := d.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	request := adomain.WithdrawalCreateRequest{
		JWTUser:  user,
		ClientIp: c.RealIP(),
	}
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	response, err := d.withdrawalUsecase.Create(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusCreated, response)
	case errors.Is(err, cdomain.ErrItemNotFound):
		return factory.ErrorNotFound.New(err)
	case errors.Is(err, cdomain.ErrBadRequest):
		return factory.ErrorIllegalRequest.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (d *withdrawalHttpDelivery) GetAll(c echo.Context) error {
	ctx := c.Request().Context()

	user, err := d.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	// create & binding Request DTO
	var request = adomain.WithdrawalGetAllRequest{
		JWTUser: user,
		WithdrawalGetAllQueryParam: adomain.WithdrawalGetAllQueryParam{
			SearchInput: cdomain.SearchInput{
				MaxResultCount: _DefaultMaxResultCount,
			},
		},
	}
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// call usecase & handle error
	response, err := d.withdrawalUsecase.GetAll(ctx, request)

	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, response)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (d *withdrawalHttpDelivery) GetOne(c echo.Context) error {
	ctx := c.Request().Context()

	user, err := d.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	// create & binding Request DTO
	var request = adomain.WithdrawalGetOneRequest{
		JWTUser: user,
	}
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// call usecase & handle error
	response, err := d.withdrawalUsecase.GetOne(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, response)
	case errors.Is(err, cdomain.ErrItemNotFound):
		return factory.ErrorNotFound.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (d *withdrawalHttpDelivery) UpdateStatus(c echo.Context) error {
	ctx := c.Request().Context()

	user, err := d.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(nil)
	}

	// create & binding Request DTO
	var request = adomain.WithdrawalUpdateStatusRequest{
		JWTUser: user,
	}
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// call usecase & handle error
	response, err := d.withdrawalUsecase.UpdateStatus(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, response)
	case errors.Is(err, cdomain.ErrItemNotFound):
		return factory.ErrorNotFound.New(err)
	case errors.Is(err, cdomain.ErrBadRequest):
		return factory.ErrorIllegalRequest.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (d *withdrawalHttpDelivery) RequestTransfer(c echo.Context) error {
	ctx := c.Request().Context()

	user, err := d.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(nil)
	}

	// create & binding Request DTO
	var request = adomain.WithdrawalRequestTransferRequest{
		JWTUser: user,
	}
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// call usecase & handle error
	response, err := d.withdrawalUsecase.RequestTransfer(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, response)
	case errors.Is(err, cdomain.ErrItemNotFound):
		return factory.ErrorNotFound.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}
