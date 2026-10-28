package http

import (
	"net/http"

	adomain "github.com/illuminarean/energy-share-core-api/domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/echoutils"

	"github.com/pkg/errors"

	"github.com/sirupsen/logrus"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/apirender"

	"github.com/labstack/echo/v4"

	"github.com/pangpanglabs/echoswagger/v2"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
)

const (
	_DefaultMaxResultCount   = 30
	_DefaultMaxDownloadCount = 3000
)

type investmentWaitersHttpDelivery struct {
	investmentWaitersUsecase adomain.InvestmentWaitersUsecase
	logger                   factory.LoggerFactory
	extractor                jwthelper.UserExtractor
}

func New(
	api echoswagger.ApiRoot,
	investmentWaitersUsecase adomain.InvestmentWaitersUsecase,
	logger factory.LoggerFactory,
	extractor jwthelper.UserExtractor,
) *investmentWaitersHttpDelivery {
	handler := &investmentWaitersHttpDelivery{
		investmentWaitersUsecase: investmentWaitersUsecase,
		logger:                   logger,
		extractor:                extractor,
	}

	g := api.Group("v1/investment-waiters", "api/v1/investment-waiters",
		echoutils.AllowRoles(domain.ROLE_PRODUCT))

	g.GET("", handler.GetAll).
		AddParamQueryNested(adomain.InvestmentWaitersGetAllRequest{}).
		AddResponse(http.StatusOK, "投資待機者一覧", adomain.InvestmentWaitersGetAllResponse{}, nil).
		SetDescription("投資待機者一覧を照会").
		SetSummary("投資待機者一覧照会").
		SetSecurity("Authorization")

	g.GET("/:id", handler.GetOne).
		AddParamPath("", "id", "").
		AddResponse(http.StatusOK, "投資待機者詳細情報", adomain.InvestmentWaitersGetOneResponse{}, nil).
		SetDescription("投資待機者のコンテンツを照会").
		SetSummary("投資待機者詳細情報照会").
		SetSecurity("Authorization")

	g.GET("/list/download", handler.Download).
		AddParamQueryNested(adomain.InvestmentWaitersDownloadRequest{}).
		SetDescription("投資待機者一覧をダウンロード").
		SetSummary("投資待機者一覧ダウンロードリクエスト").
		SetSecurity("Authorization")

	g.PUT("/:id/status-update", handler.StatusUpdate).
		AddParamPath("", "id", "").
		AddParamBody(adomain.InvestmentWaitersStatusUpdateBody{}, "投資待機者ステータス変更", "", true).
		SetDescription("投資待機者の投資待機ステータスを変更").
		SetSummary("投資待機者の投資待機ステータスを変更").
		SetSecurity("Authorization")

	// Deprecated
	g.POST("/:id/investment-propose", handler.InvestmentPropose).
		AddParamPath("", "id", "").
		SetDescription("投資待機者に投資提案リクエストを実施").
		SetSummary("投資待機者に投資提案リクエストを実施").
		SetSecurity("Authorization").
		SetDeprecated()

	g.POST("/:id/propose", handler.Propose).
		AddParamPath("", "id", "").
		AddParamBody(adomain.InvestmentWaitersProposeBody{}, "投資待機者への投資提案リクエスト", "", true).
		SetDescription("投資待機者に投資提案リクエストを実施").
		SetSummary("投資待機者に投資提案リクエストを実施").
		SetSecurity("Authorization")

	g.POST("/:id/investment-request", handler.InvestmentRequest).
		AddParamPath("", "id", "").
		AddParamBody(adomain.InvestmentWaitersInvestmentRequestBody{}, "投資待機者の投資申請", "", true).
		SetDescription("投資待機者に投資申請を実施").
		SetSummary("投資待機者に投資申請を実施").
		SetSecurity("Authorization")

	g.POST("/:id/cancel", handler.Cancel).
		AddParamPath("", "id", "").
		SetDescription("当選前のキャンセル待ち申込をキャンセル").
		SetSummary("キャンセル待ち申込をキャンセル").
		SetSecurity("Authorization")

	return handler
}

func (d investmentWaitersHttpDelivery) GetAll(c echo.Context) error {
	ctx := c.Request().Context()

	// Request DTO 생성
	request := adomain.InvestmentWaitersGetAllRequest{}

	// Request DTO 바인딩
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// MaxResultCount 기본값 설정
	if request.MaxResultCount == 0 {
		request.MaxResultCount = _DefaultMaxResultCount
	}

	d.logger.Entry(ctx).WithFields(logrus.Fields{
		"sortby":           request.Sortby,
		"order":            request.Order,
		"max_result_count": request.MaxResultCount,
		"skip_count":       request.SkipCount,
	}).Info("SearchStart")

	response, err := d.investmentWaitersUsecase.GetAll(ctx, request)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, response)
}

func (d investmentWaitersHttpDelivery) GetOne(c echo.Context) error {
	ctx := c.Request().Context()

	// Request DTO 생성
	request := adomain.InvestmentWaitersGetOneRequest{}

	// Request DTO 바인딩
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	response, err := d.investmentWaitersUsecase.GetOne(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, response)
	case errors.Is(err, cdomain.ErrItemNotFound):
		return factory.ErrorNotFound.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (d investmentWaitersHttpDelivery) Download(c echo.Context) error {
	ctx := c.Request().Context()

	// Request DTO 생성
	request := adomain.InvestmentWaitersDownloadRequest{}

	// Request DTO 바인딩
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// MaxResultCount 기본값 설정
	if request.MaxResultCount == 0 {
		request.MaxResultCount = _DefaultMaxDownloadCount
	}

	// 로그 생성
	d.logger.Entry(c.Request().Context()).WithFields(logrus.Fields{
		"sortby":           request.Sortby,
		"order":            request.Order,
		"max_result_count": request.MaxResultCount,
		"skip_count":       request.SkipCount,
	}).Info("SearchStart")

	response, err := d.investmentWaitersUsecase.Download(ctx, request)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	// 파일 다운로드시 파일명 설정
	c.Response().Writer.Write([]byte{0xEF, 0xBB, 0xBF}) // BOM
	c.Response().Header().Set("Content-Disposition", "attachment; filename=energyshares_investment_waiters_list.csv")
	return c.Blob(http.StatusOK, "application/csv", response.Data)
}

func (d investmentWaitersHttpDelivery) StatusUpdate(c echo.Context) error {
	// 유저 정보 추출
	user, err := d.extractor.Extract(c)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()

	// Request DTO 생성
	request := adomain.InvestmentWaitersStatusUpdateRequest{
		JWTUser: user,
	}

	// Request DTO 바인딩
	if err = c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// 유효성 검사
	if err = c.Validate(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	response, err := d.investmentWaitersUsecase.StatusUpdate(ctx, request)
	if err != nil {
		return err
	}

	return apirender.RenderSucc(c, http.StatusOK, response)
}

// Deprecated: Use Propose instead
func (d investmentWaitersHttpDelivery) InvestmentPropose(c echo.Context) error {
	// 유저 정보 추출
	user, err := d.extractor.Extract(c)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()

	// Request DTO 생성
	request := adomain.InvestmentWaitersInvestmentProposeRequest{
		JWTUser: user,
	}

	// Request DTO 바인딩
	if err = c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	response, err := d.investmentWaitersUsecase.InvestmentPropose(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, response)
	case errors.Is(err, cdomain.ErrItemNotFound):
		return factory.ErrorNotFound.New(err)
	default:
		return err
	}
}

func (d investmentWaitersHttpDelivery) Propose(c echo.Context) error {
	// 유저 정보 추출
	user, err := d.extractor.Extract(c)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()

	// Request DTO 생성
	request := adomain.InvestmentWaitersProposeRequest{
		JWTUser: user,
	}

	// Request DTO 바인딩
	if err = c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	response, err := d.investmentWaitersUsecase.Propose(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, response)
	case errors.Is(err, cdomain.ErrItemNotFound):
		return factory.ErrorNotFound.New(err)
	case errors.Is(err, cdomain.ErrBadRequest):
		return factory.ErrorParameter.New(err)
	default:
		return err
	}
}

func (d investmentWaitersHttpDelivery) InvestmentRequest(c echo.Context) error {
	user, err := d.extractor.Extract(c)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()

	request := adomain.InvestmentWaitersInvestmentRequestRequest{
		JWTUser: user,
	}

	if err = c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	if err = c.Validate(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	response, err := d.investmentWaitersUsecase.InvestmentRequest(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, response)
	case errors.Is(err, cdomain.ErrItemNotFound):
		return factory.ErrorNotFound.New(err)
	case errors.Is(err, cdomain.ErrBadRequest):
		return factory.ErrorParameter.New(err)
	case errors.Is(err, cdomain.ErrConcurrentUpdate):
		// MIR-1068: 別々の待機者への同時繰上げでMySQLがデッドロックすることがある。
		// データは壊れないので、汎用の500ではなく再試行を促すエラーを返す。
		return factory.ErrorConcurrentUpdate.New(err)
	default:
		return err
	}
}

func (d investmentWaitersHttpDelivery) Cancel(c echo.Context) error {
	user, err := d.extractor.Extract(c)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()

	request := adomain.InvestmentWaitersCancelRequest{
		JWTUser: user,
	}

	if err = c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	if err = c.Validate(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	response, err := d.investmentWaitersUsecase.Cancel(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, response)
	case errors.Is(err, cdomain.ErrItemNotFound):
		return factory.ErrorNotFound.New(err)
	case errors.Is(err, cdomain.ErrBadRequest), errors.Is(err, cdomain.ErrItemConflict):
		return factory.ErrorParameter.New(err)
	default:
		return err
	}
}
