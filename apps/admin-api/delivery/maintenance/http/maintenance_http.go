package http

import (
	"net/http"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/pkg/errors"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/apirender"

	"github.com/labstack/echo/v4"

	adomain "github.com/illuminarean/energy-share-core-api/domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/echoutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	openapigo "github.com/swaggest/openapi-go"
)

type maintenanceHttpDelivery struct {
	cfg                *config.Config
	maintenanceUsecase adomain.MaintenanceUsecase
}

func New(
	api *echoutils.EchoSwagger,
	cfg *config.Config,
	maintenanceUsecase adomain.MaintenanceUsecase,
) *maintenanceHttpDelivery {
	handler := &maintenanceHttpDelivery{
		cfg:                cfg,
		maintenanceUsecase: maintenanceUsecase,
	}

	// MIR-1408: QA 用の書き換えなので、開発環境・staging のときだけ登録し、SUPER_ADMIN を必須にする。以前は処理の中の
	// IsProd() だけで止めており、環境の名前が "prod" と違えば本番でもトークン無しで書き換えられた。登録しなければ
	// 名前が想定と違っても止まる側に倒れる。
	if !cfg.IsDevelopment() {
		return handler
	}
	g := api.Group("/api/v1/maintenance")

	g.PATCH("/projects", handler.PatchProjects,
		echoutils.AllowRoles(domain.ROLE_SUPER_ADMIN)).
		SetTags("운영기능").
		SetSummary("프로젝트의 특정 필드를 업데이트 합니다.").
		AddRequest(adomain.PatchMaintenanceProjectsRequest{}).
		AddResponse(nil, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusNoContent
		}).
		AddErrors(
			factory.ErrorParameter,
			factory.ErrorFailAuthorized,
			factory.ErrorNotFound,
			factory.ErrorDB,
		)

	g.PATCH("/deposit-requests", handler.PatchDepositRequests,
		echoutils.AllowRoles(domain.ROLE_SUPER_ADMIN)).
		SetTags("운영기능").
		SetSummary("입금요청 특정 필드를 업데이트 합니다.").
		AddRequest(adomain.PatchMaintenanceDepositRequestsRequest{}).
		AddResponse(nil, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusNoContent
		}).
		AddErrors(
			factory.ErrorParameter,
			factory.ErrorFailAuthorized,
			factory.ErrorNotFound,
			factory.ErrorDB,
		)

	g.PATCH("/repayment-dividend-schedules", handler.PatchRepaymentDividendSchedules,
		echoutils.AllowRoles(domain.ROLE_SUPER_ADMIN)).
		SetTags("운영기능").
		SetSummary("배당/상환 특정 필드를 업데이트 합니다.").
		AddRequest(adomain.PatchMaintenanceRepaymentDividendSchedulesRequest{}).
		AddResponse(nil, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusNoContent
		}).
		AddErrors(
			factory.ErrorParameter,
			factory.ErrorFailAuthorized,
			factory.ErrorNotFound,
			factory.ErrorDB,
		)

	return handler
}

func (d maintenanceHttpDelivery) PatchProjects(c echo.Context) error {
	ctx := c.Request().Context()

	// 운영환경에서는 수행이 되지 않도록 작업
	if d.cfg.IsProd() {
		return factory.ErrorFailAuthorized.New(nil)
	}

	request := adomain.PatchMaintenanceProjectsRequest{}

	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	if err := request.Validate(); err != nil {
		return factory.ErrorParameter.New(err)
	}

	err := d.maintenanceUsecase.PatchProjects(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusNoContent, "")
	case errors.Is(err, cdomain.ErrItemNotFound):
		return factory.ErrorNotFound.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (d maintenanceHttpDelivery) PatchDepositRequests(c echo.Context) error {
	ctx := c.Request().Context()

	// 운영환경에서는 수행이 되지 않도록 작업
	if d.cfg.IsProd() {
		return factory.ErrorFailAuthorized.New(nil)
	}

	request := adomain.PatchMaintenanceDepositRequestsRequest{}

	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	if err := request.Validate(); err != nil {
		return factory.ErrorParameter.New(err)
	}

	err := d.maintenanceUsecase.PatchDepositRequests(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusNoContent, "")
	case errors.Is(err, cdomain.ErrItemNotFound):
		return factory.ErrorNotFound.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (d maintenanceHttpDelivery) PatchRepaymentDividendSchedules(c echo.Context) error {
	ctx := c.Request().Context()

	// 운영환경에서는 수행이 되지 않도록 작업
	if d.cfg.IsProd() {
		return factory.ErrorFailAuthorized.New(nil)
	}

	request := adomain.PatchMaintenanceRepaymentDividendSchedulesRequest{}

	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	if err := request.Validate(); err != nil {
		return factory.ErrorParameter.New(err)
	}

	err := d.maintenanceUsecase.PatchRepaymentDividendSchedules(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusNoContent, "")
	case errors.Is(err, cdomain.ErrItemNotFound):
		return factory.ErrorNotFound.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}
