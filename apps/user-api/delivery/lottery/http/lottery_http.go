package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/actorcontext"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/apirender"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/echoutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	udomain "github.com/illuminarean/energyshares-user-api/domain"
)

type lotteryHttpDelivery struct {
	lotteryUsecase udomain.LotteryUsecase
	extractor      jwthelper.UserExtractor
}

// projectはControllerに置かれており古いのでLotteryで作成
// projectを新しくしたらlotteryはprojectに統合する

func New(
	api echoswagger.ApiRoot,
	lotteryUsecase udomain.LotteryUsecase,
	extractor jwthelper.UserExtractor,
) *lotteryHttpDelivery {
	handler := &lotteryHttpDelivery{
		lotteryUsecase: lotteryUsecase,
		extractor:      extractor,
	}

	g := api.Group("v1/lottery", "api/v1/lottery")

	g.POST("/applications",
		handler.Application,
		echoutils.AllowRoles(domain.ROLE_USER),
		echoutils.AllowAuditStatuses(domain.COMPLETED_AUDIT),
	).AddResponse(http.StatusOK, "抽選申込", udomain.LotteryApplicationResponse{}, nil).
		SetDescription("抽選申込。1ファンド1人に1抽選申込まで。").
		AddParamBody(udomain.LotteryApplicationRequest{}, "body", "リクエストボディ", true).
		SetSummary("抽選申込").
		SetSecurity("Authorization")

	g.POST("/check-application", handler.CheckApplication,
		echoutils.AllowRoles(domain.ROLE_USER),
		echoutils.AllowAuditStatuses(domain.COMPLETED_AUDIT),
	).AddResponse(http.StatusOK, "抽選申込確認", udomain.LotteryCheckApplicationResponse{}, nil).
		SetDescription("抽選申込の有無を確認。").
		AddParamBody(udomain.LotteryCheckApplicationRequest{}, "body", "リクエストボディ", true).
		SetSummary("抽選申込確認").
		SetSecurity("Authorization")

	g.POST("/applications/:lottery_application_id/cancel",
		handler.Cancel,
		echoutils.AllowRoles(domain.ROLE_USER),
		echoutils.AllowAuditStatuses(domain.COMPLETED_AUDIT),
	).AddParamPath("", "lottery_application_id", "抽選申込ID").
		AddResponse(http.StatusOK, "抽選申込キャンセル", udomain.LotteryCancelResponse{}, nil).
		SetDescription("抽選申込をキャンセル。").
		SetSummary("抽選申込キャンセル").
		SetSecurity("Authorization")

	userLotteryGroup := api.Group("v1/users", "api/v1/users",
		echoutils.AllowRoles(domain.ROLE_USER),
		echoutils.AllowAuditStatuses(domain.COMPLETED_AUDIT),
	)

	userLotteryGroup.GET("/:user_id/lottery/applications", handler.GetApplications).
		AddParamPath("", "user_id", "ユーザーID").
		AddParamQuery("", "limit", "取得件数", false).
		AddParamQuery("", "offset", "取得開始位置", false).
		AddResponse(http.StatusOK, "抽選申込一覧", udomain.LotteryGetApplicationsResponse{}, nil).
		SetDescription("ユーザーごとの抽選申込一覧を照会します。").
		SetSummary("抽選申込一覧照会").
		SetSecurity("Authorization")

	return handler
}

func (l lotteryHttpDelivery) Application(c echo.Context) error {
	user, err := l.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}
	ctx := c.Request().Context()

	request := udomain.LotteryApplicationRequest{
		JWTUser: user,
	}
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err = request.ValidUser(c); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// call usecase & handle error
	response, err := l.lotteryUsecase.Application(ctx, request)
	if err != nil {
		var appErr *factory.Error
		if errors.As(err, &appErr) {
			return appErr
		}
		switch {
		case errors.Is(err, cdomain.ErrItemNotFound):
			return factory.ErrorNotFound.New(err)
		case errors.Is(err, cdomain.ErrAlreadyExists):
			return factory.ErrorAlreadyExists.New(err)
		case errors.Is(err, cdomain.ErrBadRequest):
			return factory.ErrorParameter.New(err)
		default:
			return factory.ErrorDB.New(err)
		}
	}

	return apirender.RenderSucc(c, http.StatusOK, response)
}

func (l lotteryHttpDelivery) CheckApplication(c echo.Context) error {
	user, err := l.extractor.Extract(c)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}
	ctx := c.Request().Context()

	request := udomain.LotteryCheckApplicationRequest{
		JWTUser: user,
	}
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err = request.ValidUser(c); err != nil {
		return factory.ErrorParameter.New(err)
	}

	response, err := l.lotteryUsecase.CheckApplication(ctx, request)
	if err != nil {
		switch {
		case errors.Is(err, cdomain.ErrItemNotFound):
			return factory.ErrorNotFound.New(err)
		case errors.Is(err, cdomain.ErrBadRequest):
			return factory.ErrorParameter.New(err)
		default:
			return factory.ErrorDB.New(err)
		}
	}

	return apirender.RenderSucc(c, http.StatusOK, response)
}

func (l lotteryHttpDelivery) Cancel(c echo.Context) error {
	user, err := l.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	ctx := actorcontext.WithActor(c.Request().Context(), user)
	lotteryApplicationID := c.Param("lottery_application_id")

	request := udomain.LotteryCancelRequest{
		JWTUser:              user,
		LotteryApplicationID: lotteryApplicationID,
	}
	if err := c.Validate(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	response, err := l.lotteryUsecase.Cancel(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, response)
	case errors.Is(err, cdomain.ErrItemNotFound):
		return factory.ErrorNotFound.New(err)
	case errors.Is(err, cdomain.ErrBadRequest):
		return factory.ErrorParameter.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (l lotteryHttpDelivery) GetApplications(c echo.Context) error {
	user, err := l.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}
	ctx := c.Request().Context()

	request := udomain.LotteryGetApplicationsRequest{
		JWTUser: user,
	}
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err = request.ValidUser(c); err != nil {
		return factory.ErrorParameter.New(err)
	}

	response, err := l.lotteryUsecase.GetApplications(ctx, request)
	if err != nil {
		switch {
		case errors.Is(err, cdomain.ErrBadRequest):
			return factory.ErrorParameter.New(err)
		default:
			return factory.ErrorDB.New(err)
		}
	}

	return apirender.RenderSucc(c, http.StatusOK, response)
}
