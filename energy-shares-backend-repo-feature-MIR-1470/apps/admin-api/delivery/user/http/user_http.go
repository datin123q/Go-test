package http

import (
	"net/http"

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

const _DefaultMaxResultCount = 30

type userHttpDelivery struct {
	extractor   jwthelper.UserExtractor
	userUsecase adomain.UserUsecase
}

func New(
	api *echoutils.EchoSwagger,
	extractor jwthelper.UserExtractor,
	userUsecase adomain.UserUsecase,
) *userHttpDelivery {
	handler := &userHttpDelivery{
		extractor:   extractor,
		userUsecase: userUsecase,
	}

	g := api.Group("/api/v1/users")

	g.GET("/:id/mails", handler.GetSentMails,
		echoutils.AllowRoles(domain.ROLE_INVESTOR)).
		SetTags("공통유저").
		SetSummary("특정 사용자에게 발신한 메일을 조회합니다.").
		AddRequest(adomain.UserGetSentMailsRequest{}).
		AddResponse(adomain.MailGetAllResponse{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "특정 사용자에게 발신한 메일 조회 응답 값"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorParameter,
			factory.ErrorNotFound,
			factory.ErrorDB,
		).
		AddSecurity("Authorization")

	// 프론트에서 GET 인 경우 body 를 사용할 수 없어 POST 사용
	g.POST("/filter/csv", handler.DownloadUserByFilter,
		echoutils.AllowRoles(domain.ROLE_INVESTOR)).
		SetTags("공통유저").
		SetSummary("필터에 해당하는 사용자들을 CSV 형태로 다운로드합니다.").
		AddRequest(adomain.UserDownloadUserByFilterRequest{}).
		AddResponse(adomain.UserDownloadUserByFilterRequest{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "필터에 해당하는 사용자들 목록 CSV 파일"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorParameter,
			factory.ErrorDB,
		).
		AddSecurity("Authorization")

	return handler
}

func (delivery *userHttpDelivery) GetSentMails(c echo.Context) error {
	ctx := c.Request().Context()

	user, err := delivery.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	request := adomain.UserGetSentMailsRequest{
		JWTUser:  user,
		ClientIp: c.RealIP(),
	}

	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	if request.MaxResultCount == 0 {
		request.MaxResultCount = _DefaultMaxResultCount
	}

	// Usecase 호출
	response, err := delivery.userUsecase.GetSentMails(ctx, request)
	if err != nil {
		if errors.Is(err, cdomain.ErrItemNotFound) {
			return factory.ErrorNotFound.New(err)
		}
		if errors.Is(err, cdomain.ErrBadRequest) {
			return factory.ErrorParameter.New(err)
		}
		return factory.ErrorDB.New(err)
	}
	return apirender.RenderSucc(c, http.StatusOK, response)
}

func (delivery *userHttpDelivery) DownloadUserByFilter(c echo.Context) error {
	ctx := c.Request().Context()

	user, err := delivery.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	request := adomain.UserDownloadUserByFilterRequest{
		JWTUser:  user,
		ClientIp: c.RealIP(),
	}

	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// Usecase 호출
	response, err := delivery.userUsecase.DownloadUserByFilter(ctx, request)
	switch {
	case err == nil:
		c.Response().Writer.Write([]byte{0xEF, 0xBB, 0xBF}) // BOM
		c.Response().Header().Set("Content-Disposition", "attachment; filename=energyshares_mail_list.csv")
		return c.Blob(http.StatusOK, "application/csv", response.Data)
	default:
		return factory.ErrorDB.New(err)
	}
}
