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

type noticeMsgHttpDelivery struct {
	noticeMsgUsecase adomain.NoticeMsgUsecase
	extractor        jwthelper.UserExtractor
}

func New(
	api *echoutils.EchoSwagger,
	noticeMsgUsecase adomain.NoticeMsgUsecase,
	extractor jwthelper.UserExtractor,
) *noticeMsgHttpDelivery {
	handler := &noticeMsgHttpDelivery{
		noticeMsgUsecase: noticeMsgUsecase,
		extractor:        extractor,
	}

	g := api.Group("/api/v1/notice-msgs",
		echoutils.AllowRoles(domain.ROLE_INVESTOR))

	g.GET("", handler.GetAll).
		SetTags("사용자알림").
		SetSummary("사용자 알림 정보 목록을 요청합니다.").
		AddRequest(adomain.NoticeMsgGetAllRequest{}).
		AddResponse(adomain.NoticeMsgGetAllResponse{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "사용자 알림 목록"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorParameter,
			factory.ErrorDB,
		).
		AddSecurity("Authorization")

	g.GET("/:id", handler.GetOne).
		SetTags("사용자알림").
		SetSummary("사용자 알림 정보를 요청합니다.").
		AddRequest(adomain.NoticeMsgGetOneRequest{}).
		AddResponse(adomain.NoticeMsgGetOneResponse{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "사용자 알림 단일 항목"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorParameter,
			factory.ErrorNotFound,
			factory.ErrorDB,
		).
		AddSecurity("Authorization")

	g.POST("/send", handler.Create).
		SetTags("사용자알림").
		SetSummary("사용자 알림 정보를 추가합니다.").
		AddRequest(adomain.NoticeMsgCreateRequest{}).
		AddResponse(struct {
			Result  string `json:"result" example:"success"`
			Success bool   `json:"success" example:"true"`
		}{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusCreated
			cu.Description = "사용자 알림 정보 추가에 대한 응답"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorParameter,
			factory.ErrorDB,
		).
		AddSecurity("Authorization")

	g.POST("/send/projects", handler.CreateByProject).
		SetTags("사용자알림").
		SetSummary("투자회원 소식 알림 정보를 추가합니다.").
		AddRequest(adomain.NoticeMsgCreateByProjectRequest{}).
		AddResponse(struct {
			Result  string `json:"result" example:"success"`
			Success bool   `json:"success" example:"true"`
		}{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusCreated
			cu.Description = "투자회원 소식 알림 정보 추가에 대한 응답"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorParameter,
			factory.ErrorDB,
		).
		AddSecurity("Authorization")

	g.POST("/send/newsletter", handler.CreateByNewsletter).
		SetTags("사용자알림").
		SetSummary("뉴스레터 소식 알림 정보를 추가합니다.").
		AddRequest(adomain.NoticeMsgCreateByNewsletterRequest{}).
		AddResponse(struct {
			Result  string `json:"result" example:"success"`
			Success bool   `json:"success" example:"true"`
		}{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusCreated
			cu.Description = "뉴스레터 소식 알림 정보 추가에 대한 응답"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorParameter,
			factory.ErrorDB,
		).
		AddSecurity("Authorization")

	g.POST("/send/user-type", handler.CreateByUserType).
		SetTags("사용자알림").
		SetSummary("사용자 유저타입별 소식 알림 정보를 추가합니다.").
		AddRequest(adomain.NoticeMsgCreateByUserTypeRequest{}).
		AddResponse(struct {
			Result  string `json:"result" example:"success"`
			Success bool   `json:"success" example:"true"`
		}{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusCreated
			cu.Description = "사용자 유저타입별 소식 알림 정보 추가에 대한 응답"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorParameter,
			factory.ErrorDB,
		).
		AddSecurity("Authorization")

	return handler
}

func (d noticeMsgHttpDelivery) GetAll(c echo.Context) error {
	ctx := c.Request().Context()

	user, err := d.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	request := adomain.NoticeMsgGetAllRequest{
		JWTUser:  user,
		ClientIp: c.RealIP(),
		NoticeMsgGetAllQueryParam: adomain.NoticeMsgGetAllQueryParam{
			SearchInput: cdomain.SearchInput{
				MaxResultCount: _DefaultMaxResultCount,
			},
		},
	}
	if err = c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	response, err := d.noticeMsgUsecase.GetAll(ctx, request)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	return apirender.RenderSucc(c, http.StatusOK, response)
}

func (d noticeMsgHttpDelivery) GetOne(c echo.Context) error {
	ctx := c.Request().Context()

	user, err := d.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	request := adomain.NoticeMsgGetOneRequest{
		JWTUser:  user,
		ClientIp: c.RealIP(),
	}
	if err = c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	response, err := d.noticeMsgUsecase.GetOne(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, response)
	case errors.Is(err, cdomain.ErrItemNotFound):
		return factory.ErrorNotFound.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (d noticeMsgHttpDelivery) Create(c echo.Context) error {
	ctx := c.Request().Context()

	user, err := d.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	request := adomain.NoticeMsgCreateRequest{
		JWTUser:  user,
		ClientIp: c.RealIP(),
	}
	if err = c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err = c.Validate(request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	err = d.noticeMsgUsecase.Create(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusCreated, nil)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (d noticeMsgHttpDelivery) CreateByProject(c echo.Context) error {
	ctx := c.Request().Context()

	user, err := d.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	request := adomain.NoticeMsgCreateByProjectRequest{
		JWTUser:  user,
		ClientIp: c.RealIP(),
	}
	if err = c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err = c.Validate(request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	err = d.noticeMsgUsecase.CreateByProject(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusCreated, nil)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (d noticeMsgHttpDelivery) CreateByNewsletter(c echo.Context) error {
	ctx := c.Request().Context()

	user, err := d.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	request := adomain.NoticeMsgCreateByNewsletterRequest{
		JWTUser:  user,
		ClientIp: c.RealIP(),
	}
	if err = c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err = c.Validate(request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	err = d.noticeMsgUsecase.CreateByNewsletter(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusCreated, nil)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (d noticeMsgHttpDelivery) CreateByUserType(c echo.Context) error {
	ctx := c.Request().Context()

	user, err := d.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	request := adomain.NoticeMsgCreateByUserTypeRequest{
		JWTUser:  user,
		ClientIp: c.RealIP(),
	}
	if err = c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err = c.Validate(request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	err = d.noticeMsgUsecase.CreateByUserType(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusCreated, nil)
	default:
		return factory.ErrorDB.New(err)
	}
}
