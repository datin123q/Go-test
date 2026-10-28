package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	openapigo "github.com/swaggest/openapi-go"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/apirender"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/echoutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	udomain "github.com/illuminarean/energyshares-user-api/domain"
)

type noticeMsgHttpDelivery struct {
	noticeMsgUsecase udomain.NoticeMsgUsecase
	extractor        jwthelper.UserExtractor
}

func New(
	api *echoutils.EchoSwagger,
	noticeMsgUsecase udomain.NoticeMsgUsecase,
	extractor jwthelper.UserExtractor,
) *noticeMsgHttpDelivery {
	handler := &noticeMsgHttpDelivery{
		noticeMsgUsecase: noticeMsgUsecase,
		extractor:        extractor,
	}

	g := api.Group("/api/v1/notice-msgs",
		echoutils.AllowRoles(domain.ROLE_USER))

	g.GET("", handler.GetAll).
		SetTags("ユーザー通知").
		SetSummary("ユーザー通知リストを照会します").
		AddRequest(udomain.NoticeMsgGetAllRequest{}).
		AddResponse(udomain.NoticeMsgGetAllResponse{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "ユーザー通知リスト照会成功"
		}).
		AddSecurity("Authorization").
		AddErrors(
			factory.ErrorParameter,
			factory.ErrorNotFound,
			factory.ErrorDB,
		)

	g.GET("/:id", handler.GetOne).
		SetTags("ユーザー通知").
		SetSummary("ユーザー通知情報をリクエストします。").
		AddRequest(udomain.NoticeMsgGetOneRequest{}).
		AddResponse(udomain.NoticeMsgGetOneResponse{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "ユーザー通知単一項目"
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

func (d noticeMsgHttpDelivery) GetAll(c echo.Context) error {
	ctx := c.Request().Context()

	user, err := d.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	// Request DTO作成
	request := udomain.NoticeMsgGetAllRequest{
		JWTUser:  user,
		ClientIp: c.RealIP(),
	}

	// Request DTOバインド
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// Request DTO妥当性検証
	if err := c.Validate(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	response, err := d.noticeMsgUsecase.GetAll(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, response)
	case errors.Is(err, cdomain.ErrItemNotFound):
		return factory.ErrorNotFound.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (d noticeMsgHttpDelivery) GetOne(c echo.Context) error {
	ctx := c.Request().Context()

	user, err := d.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	request := udomain.NoticeMsgGetOneRequest{
		JWTUser:  user,
		ClientIp: c.RealIP(),
	}
	if err = c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err = c.Validate(request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	response, err := d.noticeMsgUsecase.GetOne(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, response)
	case errors.Is(err, cdomain.ErrItemNotFound):
		return factory.ErrorNotFound.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}
