package http

import (
	adomain "github.com/illuminarean/energy-share-core-api/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/echoutils"

	"net/http"

	"github.com/pkg/errors"

	"github.com/illuminarean/energy-shares-backend-repo/domain"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"

	"github.com/sirupsen/logrus"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/apirender"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"

	"github.com/illuminarean/energy-share-core-api/controllers"

	"github.com/labstack/echo/v4"

	"github.com/pangpanglabs/echoswagger/v2"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
)

const _DefaultMaxResultCount = 30

type investmentHttpDelivery struct {
	investmentUsecase adomain.InvestmentUsecase
	logger            factory.LoggerFactory
	extractor         jwthelper.UserExtractor
}

func New(
	api echoswagger.ApiRoot,
	investmentUsecase adomain.InvestmentUsecase,
	logger factory.LoggerFactory,
	extractor jwthelper.UserExtractor,
) *investmentHttpDelivery {
	g := api.Group("v1/investments", "api/v1/investments")

	handler := &investmentHttpDelivery{
		investmentUsecase: investmentUsecase,
		logger:            logger,
		extractor:         extractor,
	}

	g.GET("", handler.GetAll,
		echoutils.AllowRoles(
			domain.ROLE_INVESTOR,
			domain.ROLE_PAYMENT,
		)).
		AddParamQueryNested(controllers.SearchInput{}).
		AddParamQueryNested(adomain.InvestmentGetAllQueryParam{}).
		AddResponse(http.StatusOK, "사용자 계좌 리스트 조회", adomain.InvestmentGetAllResponse{}, nil).
		SetDescription("사용자 예치금 계좌 목록을 조회합니다.").
		SetSummary("사용자 예치금 계좌 목록 조회").
		SetSecurity("Authorization")

	g.GET("/:id", handler.GetOne,
		echoutils.AllowRoles(
			domain.ROLE_INVESTOR,
			domain.ROLE_PAYMENT,
		)).
		AddParamPath("", "id", "Share Id / Investment Id").
		AddResponse(http.StatusOK, "사용자 계좌 조회", adomain.InvestmentGetOneResponse{}, nil).
		SetDescription("사용자 예치금 계좌 정보를 조회 합니다.").
		SetSummary("사용자 예치금 계좌 상세정보 조회").
		SetSecurity("Authorization")

	g.PUT("/:id/invest-status", handler.UpdateStatus,
		echoutils.AllowRoles(
			domain.ROLE_INVESTOR,
			domain.ROLE_PAYMENT,
		)).
		AddParamPath("", "id", "Share Id / Investment Id").
		AddParamBody(adomain.InvestmentUpdateStatusBody{}, "투자상태변경 입력", "", true).
		SetDescription("투자 상태를 변경 합니다. ").
		SetSummary("투자 상태를 변경 합니다.").
		SetSecurity("Authorization")

	g.GET("/list/download", handler.DownloadHistories,
		echoutils.AllowRoles(
			domain.ROLE_INVESTOR,
			domain.ROLE_PRODUCT,
		)).
		AddParamQueryNested(controllers.SearchInput{}).
		AddParamQueryNested(adomain.InvestmentDownloadHistoriesQueryParam{}).
		SetDescription("사용자 예치금 계좌 목록을 다운로드합니다.").
		SetSummary("사용자 예치금 계좌 목록 다운로드").
		SetSecurity("Authorization")

	g.GET("/:id/transaction-histories", handler.GetAllTransactionInfoHistories,
		echoutils.AllowRoles(
			domain.ROLE_INVESTOR,
			domain.ROLE_PAYMENT,
		)).
		AddParamPath("", "id", "Share Id / Investment Id").
		AddParamQueryNested(controllers.SearchInput{}).
		AddResponse(http.StatusOK, "예치금 거래 내역 조회",
			adomain.InvestmentGetAllTransactionInfoHistoriesResponse{}, nil).
		SetDescription("예치금 거래 내역 조회 합니다.").
		SetSummary("예치금 거래 내역 조회").
		SetSecurity("Authorization")

	return handler
}

func (d investmentHttpDelivery) GetAll(c echo.Context) error {
	// 유저 정보 추출
	user, err := d.extractor.Extract(c)
	if err != nil {
		return err
	}

	// Request DTO 생성
	var request = adomain.InvestmentGetAllRequest{
		JWTUser:  user,
		ClientIp: c.RealIP(),
	}

	ctx := c.Request().Context()

	// Request DTO 바인딩
	if err = c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// Request DTO 유효성 검증
	if err = c.Validate(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// MaxResultCount가 0이면 기본값으로 설정
	if request.MaxResultCount == 0 {
		request.MaxResultCount = _DefaultMaxResultCount
	}

	// FIXME Echo 기본 Logger만 사용해도 이런 것들은 찍힐 건데 굳이 코드에서 하는 이유가 있는지 확인
	d.logger.Entry(ctx).WithFields(logrus.Fields{
		"sortby":           request.Sortby,
		"order":            request.Order,
		"max_result_count": request.MaxResultCount,
		"skip_count":       request.SkipCount,
		"startAmount":      request.StartAmount,
		"endAmount":        request.EndAmount,
	}).Info("SearchStart")

	response, err := d.investmentUsecase.GetAll(ctx, request)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, response)
}

func (d investmentHttpDelivery) GetOne(c echo.Context) error {
	// 유저 정보 추출
	user, err := d.extractor.Extract(c)
	if err != nil {
		return err
	}

	if !user.HasRole(domain.ROLE_PAYMENT) {
		return factory.ErrorNotHaveRole.New(nil)
	}

	ctx := c.Request().Context()

	// Request DTO 생성
	var request = adomain.InvestmentGetOneRequest{
		JWTUser:  user,
		ClientIp: c.RealIP(),
	}

	// Request DTO 바인딩
	if err = c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// 로그 기록
	d.logger.Entry(ctx).WithFields(logrus.Fields{
		"user_id":  request.JWTUser.Id,
		"share_id": request.InvestmentId,
	}).Info("SearchStart")

	response, err := d.investmentUsecase.GetOne(ctx, request)

	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, response)
	case errors.Is(err, cdomain.ErrItemNotFound):
		return factory.ErrorNotFound.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (d investmentHttpDelivery) DownloadHistories(c echo.Context) error {
	ctx := c.Request().Context()

	request := adomain.InvestmentDownloadHistoriesRequest{}

	// Request DTO 바인딩
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	d.logger.Entry(c.Request().Context()).WithFields(logrus.Fields{
		"sortby":           request.Sortby,
		"order":            request.Order,
		"max_result_count": request.MaxResultCount,
		"skip_count":       request.SkipCount,
		"startAmount":      request.StartAmount,
		"endAmount":        request.EndAmount,
	}).Info("SearchStart")

	response, err := d.investmentUsecase.DownloadHistories(ctx, request)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	// 파일 다운로드시 파일명 설정
	c.Response().Writer.Write([]byte{0xEF, 0xBB, 0xBF}) // BOM
	c.Response().Header().Set("Content-Disposition", "attachment; filename=energyshares_invest_list.csv")
	return c.Blob(http.StatusOK, "application/csv", response.Data)
}

func (d investmentHttpDelivery) UpdateStatus(c echo.Context) error {
	// 유저 정보 추출
	user, err := d.extractor.Extract(c)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()

	// Request DTO 생성
	var request = adomain.InvestmentUpdateStatusRequest{
		JWTUser: user,
	}

	// Request DTO 바인딩
	if err = c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// Request DTO 유효성 검증
	if err = c.Validate(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	response, err := d.investmentUsecase.UpdateStatus(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, response)
	// MIR-1295: usecase側(userFacingError)がerr.Error()で日本語の案内文を
	// そのまま返すため、専用のErrorTemplate(Message: "%s")でmessageフィールド
	// に載せる。ErrCannotManuallyChangeInvestStatusはErrBadRequestを包んで
	// いない独立したsentinelなので、以下のcaseとの前後関係は結果に影響しない。
	case errors.Is(err, cdomain.ErrCannotManuallyChangeInvestStatus):
		return factory.ErrCannotManuallyChangeInvestStatus.New(err, err.Error())
	case errors.Is(err, cdomain.ErrItemNotFound):
		return factory.ErrorNotFound.New(err)
	case errors.Is(err, cdomain.ErrBadRequest), errors.Is(err, cdomain.ErrItemConflict):
		return factory.ErrorParameter.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (d investmentHttpDelivery) GetAllTransactionInfoHistories(c echo.Context) error {
	ctx := c.Request().Context()

	// Request DTO 생성
	var request = adomain.InvestmentGetAllTransactionInfoHistoriesRequest{}

	// Request DTO 바인딩
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// Request DTO 유효성 검증
	if err := c.Validate(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// 로그 기록
	d.logger.Entry(ctx).WithFields(logrus.Fields{
		"sortby":           request.Sortby,
		"order":            request.Order,
		"max_result_count": request.MaxResultCount,
		"skip_count":       request.SkipCount,
	}).Info("SearchStart")

	response, err := d.investmentUsecase.GetAllTransactionInfoHistories(ctx, request)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, response)
}
