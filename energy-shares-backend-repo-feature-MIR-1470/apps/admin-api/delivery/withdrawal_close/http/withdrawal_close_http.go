package http

import (
	"net/http"
	"time"

	adomain "github.com/illuminarean/energy-share-core-api/domain"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/apirender"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/echoutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	"github.com/labstack/echo/v4"
	openapigo "github.com/swaggest/openapi-go"
)

const (
	_DefaultMaxResultCount = 30
)

type withdrawalCloseHttpDelivery struct {
	extractor              jwthelper.UserExtractor
	withdrawalCloseUsecase adomain.WithdrawalCloseUsecase
}

func New(
	api *echoutils.EchoSwagger,
	extractor jwthelper.UserExtractor,
	withdrawalCloseUsecase adomain.WithdrawalCloseUsecase,
) *withdrawalCloseHttpDelivery {
	handler := &withdrawalCloseHttpDelivery{
		extractor:              extractor,
		withdrawalCloseUsecase: withdrawalCloseUsecase,
	}

	g := api.Group("/api/v1/withdrawals/closed", echoutils.AllowRoles(domain.ROLE_PAYMENT))

	g.GET("", handler.GetAll).
		SetTags("출금의뢰").
		SetSummary("출금의뢰 마감 목록을 조회합니다. (송금 관리 목록)").
		AddRequest(adomain.WithdrawalCloseGetAllRequest{}).
		AddResponse(adomain.WithdrawalCloseGetAllResponse{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "출금의뢰 마감 목록 조회 응답 값"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorParameter,
			factory.ErrorDB,
		).
		AddSecurity("Authorization")

	g.GET("/:id", handler.GetWithdrawals).
		SetTags("출금의뢰").
		SetSummary("마감에 따른 출금의뢰 목록을 조회합니다. (송금 관리 상세)").
		AddRequest(adomain.WithdrawalCloseGetWithdrawalsRequest{}).
		AddResponse(adomain.WithdrawalCloseGetWithdrawalsResponse{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "출금의뢰 조회 응답 값"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorParameter,
			factory.ErrorDB,
		).
		AddSecurity("Authorization")

	g.PUT("", handler.CloseAllWithdrawals).
		SetTags("출금의뢰").
		SetSummary("출금의뢰를 일괄 마감처리 합니다.").
		AddRequest(adomain.WithdrawalCloseCloseAllWithdrawalsRequest{}).
		AddResponse(adomain.WithdrawalCloseCloseAllWithdrawalsResponse{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "출금의뢰 일괄 마감 응답 값"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorParameter,
			factory.ErrorDB,
		).
		AddSecurity("Authorization")

	g.GET("/:id/download", handler.Download).
		SetTags("출금의뢰").
		SetSummary("마감에 따른 출금의뢰 목록을 다운로드 합니다.").
		AddRequest(adomain.WithdrawalCloseDownloadRequest{}).
		AddResponse(nil, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.ContentType = "text/csv; charset=Shift_JIS"
			cu.Description = "마감된 출금의뢰 목록 다운로드 성공"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorParameter,
			factory.ErrorDB,
		).
		AddSecurity("Authorization")

	g.POST("/:id/transfer", handler.RequestTransfer).
		SetTags("출금의뢰").
		SetSummary("마감된 출금의뢰를 송금 요청합니다.").
		AddRequest(adomain.WithdrawalCloseRequestTransferRequest{}).
		AddResponse(adomain.WithdrawalCloseRequestTransferResponse{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "마감된 출금의뢰의 송금요청 응답 값"
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

func (d withdrawalCloseHttpDelivery) GetAll(c echo.Context) error {
	ctx := c.Request().Context()

	user, err := d.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	// create & binding Request DTO
	var request = adomain.WithdrawalCloseGetAllRequest{
		JWTUser: user,
		SearchInput: cdomain.SearchInput{
			MaxResultCount: _DefaultMaxResultCount,
		},
	}
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// call usecase & handle error
	response, err := d.withdrawalCloseUsecase.GetAll(ctx, request)

	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, response)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (d withdrawalCloseHttpDelivery) GetWithdrawals(c echo.Context) error {

	ctx := c.Request().Context()

	// 유저 정보 추출 & 권한 검사
	user, err := d.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	// Request DTO 바인딩 & 유효성 검증
	var request = adomain.WithdrawalCloseGetWithdrawalsRequest{
		JWTUser: user,
	}
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	response, err := d.withdrawalCloseUsecase.GetWithdrawals(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, response)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (d withdrawalCloseHttpDelivery) CloseAllWithdrawals(c echo.Context) error {
	return nil
}

func (d withdrawalCloseHttpDelivery) Download(c echo.Context) error {

	ctx := c.Request().Context()

	// 유저 정보 추출 & 권한 검사
	user, err := d.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	// Request DTO 바인딩 & 유효성 검증
	var request = adomain.WithdrawalCloseDownloadRequest{
		JWTUser:  user,
		ClientIp: c.RealIP(),
	}

	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	response, err := d.withdrawalCloseUsecase.Download(ctx, request)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	// 파일 다운로드시 파일명 및 인코딩 설정
	fileName := "energyshares_closed_withdrawal_request_" + time.Now().Format("2006_01_02_15_04_05") + ".csv"
	c.Response().Header().Set("Content-Disposition", "attachment; filename="+fileName)
	return c.Blob(http.StatusOK, "text/csv; charset=Shift_JIS", response.Data)
}

func (d withdrawalCloseHttpDelivery) RequestTransfer(c echo.Context) error {
	return nil
}
