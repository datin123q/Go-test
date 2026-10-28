package http

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	openapigo "github.com/swaggest/openapi-go"

	adomain "github.com/illuminarean/energy-share-core-api/domain"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/apirender"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/echoutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
)

const (
	_DefaultMaxResultCount = 30
)

type repaymentDividendHttpDelivery struct {
	repaymentDividendUsecase adomain.RepaymentDividendUsecase
	logger                   factory.LoggerFactory
	extractor                jwthelper.UserExtractor
}

func New(
	api *echoutils.EchoSwagger,
	repaymentDividendUsecase adomain.RepaymentDividendUsecase,
	logger factory.LoggerFactory,
	extractor jwthelper.UserExtractor,
) *repaymentDividendHttpDelivery {
	g := api.Group("/api/v1/projects",
		echoutils.AllowRoles(domain.ROLE_PRODUCT))

	handler := &repaymentDividendHttpDelivery{
		repaymentDividendUsecase: repaymentDividendUsecase,
		logger:                   logger,
		extractor:                extractor,
	}

	// 대출실행
	g.POST("/:id/execute-investment", handler.ExecuteInvestment).
		SetTags("출자/배당/상환").
		SetSummary("모집 금액 출자를 실행 합니다.").
		AddRequest(adomain.ExecuteInvestmentRequest{}).
		AddResponse(adomain.ExecuteInvestmentResponse{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "출자 응답 값"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorParameter,
			factory.ErrExecuteInvestmentDateMismatch,
			factory.ErrExecuteInvestmentAmountMismatch,
			factory.ErrExecuteInvestmentLocked,
			factory.ErrorDB,
		).
		AddSecurity("Authorization")

	// // 배당상환 리뉴얼: 배당/상환하기 - https://illuminarean.atlassian.net/browse/ESJP-365
	g.POST("/:id/repayment-dividend-schedules", handler.CreateRepaymentDividendSchedules).
		SetTags("출자/배당/상환").
		SetSummary("배당/상환 일정 추가.").
		AddRequest(adomain.CreateRepaymentDividendRequest{}).
		AddResponse(adomain.CreateRepaymentDividendResponse{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "배당/상환 일정 추가 응답값"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorParameter,
			factory.ErrParameterRepaymentAmountLogic,
			factory.ErrParameterDividendAmountLogic,
			factory.ErrParameterRepaymentAmountExceeds,
			factory.ErrInvestmentExecutionIncomplete,
			factory.ErrorDB,
		).
		AddSecurity("Authorization")

	// 배당상환 리뉴얼 - 출자/배당/상환 테이블 정보 가져오기
	g.GET("/:id/repayment-dividend-schedules/table", handler.GetRepaymentDividendSchedulesTable).
		SetTags("출자/배당/상환").
		SetSummary("출자/배당/상환 목록을 요청합니다.").
		AddRequest(adomain.GetRepaymentDividendTableRequest{}).
		AddResponse(adomain.GetRepaymentDividendTableResponse{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "출자/배당/상환 목록"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorParameter,
			factory.ErrorDB,
		).
		AddSecurity("Authorization")

	// 배당상환 리뉴얼 - 출자/배당/상환 테이블 정보 다운로드
	g.GET("/:id/repayment-dividend-schedules/table-csv", handler.GetRepaymentDividendSchedulesTableCsv).
		SetTags("출자/배당/상환").
		SetSummary("출자/배당/상환 목록을 다운로드 합니다.").
		AddRequest(adomain.GetRepaymentDividendTableRequest{}).
		AddResponse(adomain.GetRepaymentDividendTableResponse{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "출자/배당/상환 목록 CSV 다운로드"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorParameter,
			factory.ErrorDB,
		).
		AddSecurity("Authorization")

	// 배당상환 리뉴얼 - 자금관리 정보 가저오기
	g.GET("/:id/repayment-dividend-schedules/info", handler.GetInvestmentAndReturnsInfo).
		SetTags("출자/배당/상환").
		SetSummary("출자/배당/상환 정보를 조회 합니다.").
		AddRequest(adomain.GetInvestmentAndReturnsInfoRequest{}).
		AddResponse(adomain.GetInvestmentAndReturnsInfoResponse{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "출자/배당/상환 정보"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorParameter,
			factory.ErrorDB,
		).
		AddSecurity("Authorization")

	// 배당상환 리뉴얼 - 스케쥴 삭제하기
	g.DELETE("/:id/repayment-dividend-schedules/:schedule_id", handler.DeleteRepaymentDividendSchedule).
		SetTags("출자/배당/상환").
		SetSummary("배당/상환 일정을 삭제 합니다.").
		AddRequest(adomain.DeleteRepaymentDividendRequest{}).
		AddResponse(nil, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorParameter,
			factory.ErrorDB,
		).
		AddSecurity("Authorization")

	return handler
}

func (d repaymentDividendHttpDelivery) ExecuteInvestment(c echo.Context) error {

	// リクエストキャンセルを切り離し、独立した 300 秒タイムアウトを設定する。
	ctx, cancel := context.WithTimeout(context.WithoutCancel(c.Request().Context()), 300*time.Second)
	defer cancel()
	user, err := d.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	request := adomain.ExecuteInvestmentRequest{
		JWTUser: user,
	}
	if err = c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err = c.Validate(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// request.ExecuteDate는 오늘 또는 그 이전 날짜여야 한다.
	isAfter, err := request.ExecuteDateIsAfterToday()
	if err != nil {
		return factory.ErrorParameter.New(err)
	}
	if isAfter {
		return factory.ErrorParameter.New(errors.New("invalid date. it must be today or before"))
	}

	if request.Amount == 0 {
		return factory.ErrorParameter.New(errors.New("invalid amount. it must be greater than 0"))
	}

	// 출자 실행
	response, err := d.repaymentDividendUsecase.ExecuteInvestment(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, response)
	case errors.Is(err, cdomain.ErrBadRequest):
		return factory.ErrorParameter.New(err)
	case errors.Is(err, cdomain.ErrExecuteInvestmentDateMismatch):
		// 利用者にそのまま見せる日本語の文言をmessageへ載せる(MIR-1204)。
		return factory.ErrExecuteInvestmentDateMismatch.New(err, err.Error())
	case errors.Is(err, cdomain.ErrExecuteInvestmentAmountMismatch):
		return factory.ErrExecuteInvestmentAmountMismatch.New(err, err.Error())
	case errors.Is(err, cdomain.ErrItemConflict):
		// 同じプロジェクトへの出資実行が他のリクエストによって実行中(MIR-1204)。
		// 409を返し、画面側が「他の処理が実行中」と判定して自動繰り返しを
		// 止められるようにする(500だと画面共通のエラーページに飛んでしまう)。
		return factory.ErrExecuteInvestmentLocked.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (d repaymentDividendHttpDelivery) CreateRepaymentDividendSchedules(c echo.Context) error {
	ctx := c.Request().Context()
	user, err := d.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	// Request DTO 생성, 바인딩, 유효성 검증
	request := adomain.CreateRepaymentDividendRequest{
		JWTUser: user,
	}
	if err = c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err = c.Validate(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// 배당상환 스케쥴 생성 작업(계산, 스케쥴 추가, 스케쥴 기반 dividend 생성)
	response, err := d.repaymentDividendUsecase.CreateRepaymentDividendSchedules(ctx, request)

	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, response)
	// MIR-1239: usecase側(CreateRepaymentDividendSchedules)がこのエラーを
	// errors.Wrapで包まないため、err.Error()がuserFacingError自体の
	// Error()(=日本語message)をそのまま返す。domain.ErrInvestmentExecutionIncomplete
	// はcdomain.ErrBadRequestを包んでいない独立したsentinelなので、
	// cdomain.ErrBadRequestのcaseとの前後関係自体は結果に影響しない。
	case errors.Is(err, domain.ErrInvestmentExecutionIncomplete):
		return factory.ErrInvestmentExecutionIncomplete.New(err, err.Error())
	case errors.Is(err, cdomain.ErrBadRequest):
		return factory.ErrorParameter.New(err)
	case errors.Is(err, domain.ErrParameterRepaymentAmountLogic):
		return factory.ErrParameterRepaymentAmountLogic.New(err)
	case errors.Is(err, domain.ErrParameterDividendAmountLogic):
		return factory.ErrParameterDividendAmountLogic.New(err)
	case errors.Is(err, domain.ErrParameterRepaymentAmountExceeds):
		return factory.ErrParameterRepaymentAmountExceeds.New(err)

	default:
		return factory.ErrorDB.New(err)
	}
}

func (d repaymentDividendHttpDelivery) GetRepaymentDividendSchedulesTable(c echo.Context) error {
	ctx := c.Request().Context()
	// 유저 정보 추출
	user, err := d.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	// Request DTO 생성, 바인딩, 유효성 검증
	request := adomain.GetRepaymentDividendTableRequest{
		JWTUser: user,
	}
	if err = c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err = c.Validate(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if request.MaxResultCount == 0 {
		request.MaxResultCount = _DefaultMaxResultCount
	}

	// UseCase 호출
	response, err := d.repaymentDividendUsecase.GetRepaymentDividendSchedulesTable(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, response)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (d repaymentDividendHttpDelivery) GetRepaymentDividendSchedulesTableCsv(c echo.Context) error {
	ctx := c.Request().Context()
	user, err := d.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	// Request DTO 생성, 바인딩, 유효성 검증
	request := adomain.GetRepaymentDividendTableRequest{
		JWTUser: user,
	}

	if err = c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err = c.Validate(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// CSV 다운로드를 위한 설정
	request.SkipCount = 0
	request.MaxResultCount = 100_000

	// UseCase 호출
	response, err := d.repaymentDividendUsecase.GetRepaymentDividendSchedulesTableCsv(ctx, request)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	// 파일 다운로드시 파일명 설정
	c.Response().Writer.Write([]byte{0xEF, 0xBB, 0xBF}) // BOM
	c.Response().Header().Set("Content-Disposition", "attachment; filename=energyshares_repayment_dividend.csv")
	return c.Blob(http.StatusOK, "application/csv", response.Data)
}

func (d repaymentDividendHttpDelivery) GetInvestmentAndReturnsInfo(c echo.Context) error {
	ctx := c.Request().Context()
	user, err := d.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	// Request DTO 생성, 바인딩, 유효성 검증
	request := adomain.GetInvestmentAndReturnsInfoRequest{
		JWTUser:  user,
		ClientIp: c.RealIP(),
	}
	if err = c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err = c.Validate(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// UseCase 호출
	response, err := d.repaymentDividendUsecase.GetInvestmentAndReturnsInfo(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, response)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (d repaymentDividendHttpDelivery) DeleteRepaymentDividendSchedule(c echo.Context) error {
	ctx := c.Request().Context()
	// 유저 정보 추출
	user, err := d.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	// Request DTO 생성, 바인딩, 유효성 검증
	request := adomain.DeleteRepaymentDividendRequest{
		JWTUser: user,
	}

	if err = c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err = c.Validate(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// UseCase 호출
	err = d.repaymentDividendUsecase.DeleteRepaymentDividendSchedule(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, nil)
	default:
		return factory.ErrorDB.New(err)
	}
}
