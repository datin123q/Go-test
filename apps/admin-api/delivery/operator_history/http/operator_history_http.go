package http

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	openapigo "github.com/swaggest/openapi-go"

	adomain "github.com/illuminarean/energy-share-core-api/domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/apirender"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/echoutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
)

type operatorHistoryHttpDelivery struct {
	operatorHistoryUsecase adomain.OperatorHistoryUsecase
	extractor              jwthelper.UserExtractor
}

func New(
	api *echoutils.EchoSwagger,
	operatorHistoryUsecase adomain.OperatorHistoryUsecase,
	extractor jwthelper.UserExtractor,
) *operatorHistoryHttpDelivery {
	handler := &operatorHistoryHttpDelivery{
		operatorHistoryUsecase: operatorHistoryUsecase,
		extractor:              extractor,
	}

	g := api.Group("/api/v1/operatorhistories")

	g.GET("", handler.GetAll,
		echoutils.AllowRoles(domain.ROLE_SUPER_ADMIN)).
		SetTags("Operator History").
		SetSummary("운영자 로그 조회").
		SetDescription("운영자 로그를 조회합니다.").
		AddRequest(adomain.OperatorHistoryGetAllRequest{}).
		AddResponse(adomain.OperatorHistoryGetAllResponse{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "운영자 로그 목록"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorParameter,
			factory.ErrorStreamProcessing,
		).
		AddSecurity("Authorization")

	g.GET("/download", handler.Download,
		echoutils.AllowRoles(domain.ROLE_SUPER_ADMIN)).
		SetTags("Operator History").
		SetSummary("운영자 로그 엑셀파일 다운로드").
		SetDescription("운영자 로그를 엑셀파일로 다운로드합니다.").
		AddRequest(adomain.OperatorHistoryDownloadRequest{}).
		AddResponse(adomain.OperatorHistoryDownloadResponse{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "운영자 로그 엑셀파일"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorParameter,
			factory.ErrorFailDownloadCSV,
		).
		AddSecurity("Authorization")

	g.GET("/menu-codes", handler.GetMenuCodes,
		echoutils.AllowRoles(domain.ROLE_ADMIN)).
		SetTags("Operator History").
		SetSummary("메뉴 공용코드 조회").
		SetDescription("운영자 로그의 메뉴 공용코드 목록을 조회합니다.").
		AddResponse(adomain.GetCodeListResponse{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "운영자 로그 메뉴 공용코드 목록"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorParameter,
			factory.ErrorSystem,
		).
		AddSecurity("Authorization")

	g.GET("/log-codes", handler.GetLogCodes,
		echoutils.AllowRoles(domain.ROLE_ADMIN)).
		SetTags("Operator History").
		SetSummary("처리로그 공용코드 조회").
		SetDescription("운영자 로그의 처리로그 공용코드 목록을 조회합니다.").
		AddResponse(adomain.GetCodeListResponse{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "운영자 로그 처리로그 공용코드 목록"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorParameter,
			factory.ErrorSystem,
		).
		AddSecurity("Authorization")

	return handler
}

func (d operatorHistoryHttpDelivery) GetAll(c echo.Context) error {
	ctx := c.Request().Context()

	user, err := d.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	request := adomain.OperatorHistoryGetAllRequest{
		JWTUser:  user,
		ClientIp: c.RealIP(),
	}
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local).
		Format("2006-01-02 15:04:05")
	endOfDay := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, time.Local).
		Format("2006-01-02 15:04:05")
	request.StartDate = &startOfDay
	request.EndDate = &endOfDay
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	response, err := d.operatorHistoryUsecase.GetAll(ctx, request)
	if err != nil {
		return factory.ErrorStreamProcessing.New(err)
	}
	return apirender.RenderSucc(c, http.StatusOK, response)
}

func (d operatorHistoryHttpDelivery) Download(c echo.Context) error {
	ctx := c.Request().Context()

	user, err := d.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	request := adomain.OperatorHistoryDownloadRequest{
		JWTUser:  user,
		ClientIp: c.RealIP(),
	}
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	response, err := d.operatorHistoryUsecase.Download(ctx, request)
	if err != nil {
		return factory.ErrorFailDownloadCSV.New(err)
	}

	c.Response().Writer.Write([]byte{0xEF, 0xBB, 0xBF}) // BOM
	c.Response().Header().Set("Content-Disposition", "attachment; filename=energyshares_operator_history.csv")
	return c.Blob(http.StatusOK, "application/csv", response.Data)
}

func (d operatorHistoryHttpDelivery) GetMenuCodes(c echo.Context) error {
	ctx := c.Request().Context()

	user, err := d.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	request := adomain.GetCodeListRequest{
		JWTUser:  user,
		ClientIp: c.RealIP(),
	}
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if request.Locale == "" {
		request.Locale = "ja" // set default locale
	}

	response, err := d.operatorHistoryUsecase.GetMenuCodes(ctx, request)
	if err != nil {
		return factory.ErrorSystem.New(err)
	}
	return apirender.RenderSucc(c, http.StatusOK, response)
}

func (d operatorHistoryHttpDelivery) GetLogCodes(c echo.Context) error {
	ctx := c.Request().Context()

	user, err := d.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	request := adomain.GetCodeListRequest{
		JWTUser:  user,
		ClientIp: c.RealIP(),
	}
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if request.Locale == "" {
		request.Locale = "ja" // set default locale
	}

	response, err := d.operatorHistoryUsecase.GetLogCodes(ctx, request)
	if err != nil {
		return factory.ErrorSystem.New(err)
	}
	return apirender.RenderSucc(c, http.StatusOK, response)
}
