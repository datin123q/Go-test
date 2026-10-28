package http

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	openapigo "github.com/swaggest/openapi-go"

	adomain "github.com/illuminarean/energy-share-core-api/domain"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/apirender"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/echoutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
)

const _DefaultMaxResultCount = 30

type mailHttpDelivery struct {
	cfg         *config.Config
	extractor   jwthelper.UserExtractor
	logger      factory.LoggerFactory
	mailUsecase adomain.MailUsecase
}

func New(
	api *echoutils.EchoSwagger,
	cfg *config.Config,
	extractor jwthelper.UserExtractor,
	logger factory.LoggerFactory,
	mailUsecase adomain.MailUsecase,
) *mailHttpDelivery {
	g := api.Group("/api/v1/mails")

	handler := &mailHttpDelivery{
		cfg:         cfg,
		logger:      logger,
		extractor:   extractor,
		mailUsecase: mailUsecase,
	}

	g.GET("", handler.GetAll, echoutils.AllowRoles(domain.ROLE_MESSENGER)).
		SetTags("메일").
		SetSummary("메일 발신 목록을 조회합니다").
		AddRequest(adomain.MailGetAllRequest{}).
		AddResponse(adomain.MailGetAllResponse{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "메일 발신 목록"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorParameter,
			factory.ErrorDB,
		).
		AddSecurity("Authorization")

	g.GET("/:id", handler.GetOne,
		echoutils.AllowRoles(domain.ROLE_MESSENGER, domain.ROLE_INVESTOR)).
		SetTags("메일").
		SetSummary("메일 발신 상세 정보를 조회합니다").
		AddRequest(adomain.MailGetOneRequest{}).
		AddResponse(adomain.MailDetailResponse{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "메일 발신 상세 정보"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorParameter,
			factory.ErrorDB,
		).
		AddSecurity("Authorization")

	g.GET("/failed", handler.GetAllFailed,
		echoutils.AllowRoles(domain.ROLE_MESSENGER, domain.ROLE_INVESTOR)).
		SetTags("메일").
		SetSummary("메일 발신 실패 목록을 조회합니다").
		AddRequest(adomain.MailGetAllFailedRequest{}).
		AddResponse(adomain.MailGetAllFailedResponse{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "메일 발신 실패 목록"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorParameter,
			factory.ErrorDB,
		).
		AddSecurity("Authorization")

	g.GET("/templates/:template_type", handler.GetTemplate,
		echoutils.AllowRoles(domain.ROLE_MESSENGER)).
		SetTags("메일").
		SetSummary("메일 템플릿을 가져옵니다").
		AddRequest(adomain.MailGetTemplateRequest{}).
		AddResponse(adomain.MailGetTemplateResponse{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "메일 템플릿 정보"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorParameter,
			factory.ErrorDB,
		).
		AddSecurity("Authorization")

	return handler
}

func (d mailHttpDelivery) GetAll(c echo.Context) error {
	// 유저 정보 추출
	user, err := d.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	// Request DTO 생성
	var request = adomain.MailGetAllRequest{
		Env:      strings.ToUpper(d.cfg.Env),
		JWTUser:  user,
		ClientIp: c.RealIP(),
	}

	ctx := c.Request().Context()

	if err = c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	if err := request.Validate(); err != nil {
		return factory.ErrorParameter.New(err)
	}

	if request.MaxResultCount == 0 {
		request.MaxResultCount = _DefaultMaxResultCount
	}

	response, err := d.mailUsecase.GetAll(ctx, request)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, response)
}

func (d mailHttpDelivery) GetOne(c echo.Context) error {
	// 유저 정보 추출
	user, err := d.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	// Request DTO 생성
	var request = adomain.MailGetOneRequest{
		JWTUser:  user,
		ClientIp: c.RealIP(),
	}

	ctx := c.Request().Context()

	if err = c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	response, err := d.mailUsecase.GetOne(ctx, request)
	if err != nil {
		if errors.Is(err, cdomain.ErrItemNotFound) {
			return factory.ErrorNotFound.New(err)
		}
		return factory.ErrorDB.New(err)
	}
	return apirender.RenderSucc(c, http.StatusOK, response)
}

func (d mailHttpDelivery) GetAllFailed(c echo.Context) error {
	// 유저 정보 추출
	user, err := d.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	// Request DTO 생성
	var request = adomain.MailGetAllFailedRequest{
		JWTUser:  user,
		ClientIp: c.RealIP(),
	}

	ctx := c.Request().Context()

	if err = c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	if err := c.Validate(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	if request.MaxResultCount == 0 {
		request.MaxResultCount = _DefaultMaxResultCount
	}

	response, err := d.mailUsecase.GetAllFailed(ctx, request)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, response)
}

func (d mailHttpDelivery) GetTemplate(c echo.Context) error {
	// 유저 정보 추출
	user, err := d.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	// Request DTO 생성
	var request = adomain.MailGetTemplateRequest{
		JWTUser:  user,
		ClientIp: c.RealIP(),
	}

	ctx := c.Request().Context()

	if err = c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	if err := c.Validate(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// API의 주 용도는 관리자 직접 발송 메일에서 메일 결과물을 보여주기 위한 템플릿을 가져오기 위한 것임
	// 따라서 default 로 입력할 경우 NORMAL_MAIL_TEMPLATE을 가져오도록 하였음
	if request.TemplateType == "default" {
		request.TemplateType = string(domain.NORMAL_MAIL_TEMPLATE)
	}

	response, err := d.mailUsecase.GetTemplate(ctx, request)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, response)
}
