package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	openapigo "github.com/swaggest/openapi-go"

	adomain "github.com/illuminarean/energy-share-core-api/domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/apirender"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/echoutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
)

type qnaHttpDelivery struct {
	cfg              *config.Config
	changelogUsecase adomain.ChangelogUsecase
	extractor        jwthelper.UserExtractor
}

func New(
	api *echoutils.EchoSwagger,
	cfg *config.Config,
	changelogUsecase adomain.ChangelogUsecase,
	extractor jwthelper.UserExtractor,
) *qnaHttpDelivery {
	handler := &qnaHttpDelivery{
		cfg:              cfg,
		changelogUsecase: changelogUsecase,
		extractor:        extractor,
	}

	g := api.Group("/api/v1/qna",
		echoutils.AllowRoles(domain.ROLE_INVESTOR))

	g.GET("/changelogs", handler.GetChangelogs).
		SetTags("QNA").
		SetSummary("변경로그 목록을 조회").
		SetDescription("1:1문의 변경로그 목록을 조회합니다.").
		AddRequest(adomain.ChangelogGetAllRequest{}).
		AddResponse(adomain.ChangelogGetAllResponse{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "변경로그 목록"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorParameter,
			factory.ErrorStreamProcessing,
		).
		AddSecurity("Authorization")

	return handler
}

func (d qnaHttpDelivery) GetChangelogs(c echo.Context) error {
	ctx := c.Request().Context()

	user, err := d.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	request := adomain.ChangelogGetAllRequest{
		JWTUser:  user,
		ClientIp: c.RealIP(),
	}
	if err = c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	response, err := d.changelogUsecase.GetAll(ctx, request)
	if err != nil {
		return factory.ErrorStreamProcessing.New(err)
	}
	return apirender.RenderSucc(c, http.StatusOK, response)
}
