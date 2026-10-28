package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	openapigo "github.com/swaggest/openapi-go"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/echoutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	udomain "github.com/illuminarean/energyshares-user-api/domain"
)

type investmentWaitersHttpDelivery struct {
	investmentWaitersUsecase udomain.InvestmentWaitersUsecase
	extractor                jwthelper.UserExtractor
}

func New(
	api *echoutils.EchoSwagger,
	investmentWaitersUsecase udomain.InvestmentWaitersUsecase,
	extractor jwthelper.UserExtractor,
) *investmentWaitersHttpDelivery {
	handler := &investmentWaitersHttpDelivery{
		investmentWaitersUsecase: investmentWaitersUsecase,
		extractor:                extractor,
	}

	g := api.Group("/api/v1/users", echoutils.AllowRoles(domain.ROLE_USER))

	g.POST("/:id/waitlist/:waitlist_id/cancel", handler.Cancel).
		SetTags("キャンセル待ち").
		SetSummary("キャンセル待ち申請を取り消します").
		SetDescription("ユーザー本人の待機中のキャンセル待ち申請を取り消します。").
		AddRequest(udomain.InvestmentWaitersCancelRequest{}).
		AddResponse(nil, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "キャンセル待ち申請の取消成功"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorParameter,
			factory.ErrorPermissionDenied,
			factory.ErrorNotFound,
			factory.ErrorDB,
		).
		AddSecurity("Authorization")

	return handler
}

func (i investmentWaitersHttpDelivery) Cancel(c echo.Context) error {
	user, err := i.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	request := udomain.InvestmentWaitersCancelRequest{
		JWTUser:  user,
		ClientIp: c.RealIP(),
	}
	if err = c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err = c.Validate(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	_, err = i.investmentWaitersUsecase.Cancel(c.Request().Context(), request)
	switch {
	case err == nil:
		return c.NoContent(http.StatusOK)
	case errors.Is(err, cdomain.ErrItemNotFound):
		return factory.ErrorNotFound.New(err)
	case errors.Is(err, cdomain.ErrForbidden):
		return factory.ErrorPermissionDenied.New(err)
	case errors.Is(err, cdomain.ErrBadRequest):
		return factory.ErrorParameter.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}
