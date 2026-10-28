package http

import (
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

const _DefaultMaxResultCount = 30

type adminUserHttpDelivery struct {
	adminUserUsecase adomain.AdminUserUsecase
	logger           factory.LoggerFactory
	extractor        jwthelper.UserExtractor
}

func New(
	api *echoutils.EchoSwagger,
	adminUserUsecase adomain.AdminUserUsecase,
	logger factory.LoggerFactory,
	extractor jwthelper.UserExtractor,
) *adminUserHttpDelivery {

	g := api.Group("api/v1/admin-users")

	handler := &adminUserHttpDelivery{
		adminUserUsecase: adminUserUsecase,
		logger:           logger,
		extractor:        extractor,
	}

	g.GET("", handler.GetAll,
		echoutils.AllowRoles(domain.ROLE_SUPER_ADMIN)).
		SetTags("운영자").
		SetSummary("운영자 목록 조회 (ONLY FOR SUPER ADMIN)").
		AddRequest(adomain.AdminUserGetAllRequest{}).
		AddResponse(adomain.AdminUserGetAllResponse{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "운영자 목록 조회 성공"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorParameter,
			factory.ErrorDB,
		).
		AddSecurity("Authorization")

	g.GET("/download", handler.Download,
		echoutils.AllowRoles(domain.ROLE_SUPER_ADMIN)).
		SetTags("운영자").
		SetSummary("운영자 목록 다운로드 (ONLY FOR SUPER ADMIN)").
		AddRequest(adomain.AdminUserDownloadHistoriesRequest{}).
		AddResponse(nil, func(cu *openapigo.ContentUnit) { // 다운로드 시에는 어떻게 처리해야 할 것인가
			cu.HTTPStatus = http.StatusOK
			cu.ContentType = "application/csv"
			cu.Description = "운영자 목록 다운로드 성공"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorParameter,
			factory.ErrorDB,
		).
		AddSecurity("Authorization")

	g.GET("/:id", handler.GetOne,
		echoutils.AllowRoles(domain.ROLE_ADMIN)).
		SetTags("운영자").
		SetSummary("개별 운영자 정보 조회").
		AddRequest(adomain.AdminUserGetOneRequest{}).
		AddResponse(adomain.AdminUserGetOneResponse{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "개별 운영자 정보 조회 성공"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorParameter,
			factory.ErrorNotFound,
			factory.ErrorDB,
		).
		AddSecurity("Authorization")

	g.POST("", handler.Create,
		echoutils.AllowRoles(domain.ROLE_SUPER_ADMIN)).
		SetTags("운영자").
		SetSummary("새로운 운영자 생성 (ONLY FOR SUPER ADMIN)").
		AddRequest(adomain.AdminUserCreateRequest{}).
		AddResponse(adomain.AdminUserCreateResponse{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "운영자 생성 성공"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorParameter,
			factory.ErrorAlreadyExists,
			factory.ErrorDB,
		).
		AddSecurity("Authorization")

	g.PUT("/:id", handler.Update,
		echoutils.AllowRoles(domain.ROLE_SUPER_ADMIN)).
		SetTags("운영자").
		SetSummary("운영자 정보 수정 (ONLY FOR SUPER ADMIN)").
		AddRequest(adomain.AdminUserUpdateRequest{}).
		AddResponse(adomain.AdminUserUpdateResponse{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "개별 운영자 정보 수정 성공"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorParameter,
			factory.ErrorAlreadyExists,
			factory.ErrorDB,
		).
		AddSecurity("Authorization")

	g.DELETE("/:id", handler.Delete,
		echoutils.AllowRoles(domain.ROLE_SUPER_ADMIN)).
		SetTags("운영자").
		SetSummary("개별 운영자 삭제 (ONLY FOR SUPER ADMIN)").
		AddRequest(adomain.AdminUserDeleteRequest{}).
		AddResponse(adomain.AdminUserDeleteResponse{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "운영자 삭제 성공"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorParameter,
			factory.ErrorDB,
		).
		AddSecurity("Authorization")

	g.POST("/login", handler.Login).
		SetTags("운영자").
		SetSummary("운영자 로그인").
		AddRequest(adomain.AdminUserLoginRequest{}).
		AddResponse(adomain.AdminUserLoginResponse{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "운영자 로그인 성공"
		}).
		AddErrors(
			factory.ErrorParameter,
			factory.ErrorUserNotExists,
			factory.ErrorDeletedUser,
			factory.ErrorNotMatchPassword,
			factory.ErrorExpireRequest,
			factory.ErrorNotMatchPassword,
			factory.ErrorDB,
		)

	g.POST("/logout", handler.Logout,
		echoutils.AllowRoles(domain.ROLE_ADMIN)).
		SetTags("운영자").
		SetSummary("운영자 로그아웃").
		AddRequest(adomain.AdminUserLogoutRequest{}, func(cu *openapigo.ContentUnit) {
			cu.IsDefault = false
			cu.Description = "Refresh Token은 쿠키에서 추출하므로 인풋에 넣지 않음"

		}).
		AddResponse(adomain.AdminUserLogoutResponse{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "운영자 로그아웃 성공"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorParameter,
			factory.ErrorTokenInvalid,
			factory.ErrorDB,
		).
		AddSecurity("Authorization")

	return handler
}

func (d adminUserHttpDelivery) GetAll(c echo.Context) error {
	// Only for super admin

	ctx := c.Request().Context()

	// 유저 정보 추출
	user, err := d.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	// Request DTO 바인딩 & 유효성 검증
	var request = adomain.AdminUserGetAllRequest{
		JWTUser:  user,
		ClientIp: c.RealIP(),
	}
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	if request.MaxResultCount == 0 {
		request.MaxResultCount = _DefaultMaxResultCount
	}
	response, err := d.adminUserUsecase.GetAll(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, response)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (d adminUserHttpDelivery) Download(c echo.Context) error {
	// Only for super admin

	ctx := c.Request().Context()

	// 유저 정보 추출 & 권한 검사
	user, err := d.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	// Request DTO 바인딩 & 유효성 검증
	var request = adomain.AdminUserDownloadHistoriesRequest{
		JWTUser:  user,
		ClientIp: c.RealIP(),
	}
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	response, err := d.adminUserUsecase.DownloadHistories(ctx, request)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	// 파일 다운로드시 파일명 설정
	c.Response().Writer.Write([]byte{0xEF, 0xBB, 0xBF}) // BOM
	c.Response().Header().Set("Content-Disposition", "attachment; filename=energyshares_administrators_list.csv")
	return c.Blob(http.StatusOK, "application/csv", response.Data)

}

func (d adminUserHttpDelivery) GetOne(c echo.Context) error {
	ctx := c.Request().Context()

	// 유저 정보 추출
	user, err := d.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	// Request DTO 바인딩 & 유효성 검증
	var request = adomain.AdminUserGetOneRequest{
		JWTUser:  user,
		ClientIp: c.RealIP(),
	}
	if err = c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err = c.Validate(request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	response, err := d.adminUserUsecase.GetOne(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, response)
	case errors.Is(err, cdomain.ErrItemNotFound):
		return factory.ErrorNotFound.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (d adminUserHttpDelivery) Create(c echo.Context) error {
	// Only for super admin

	ctx := c.Request().Context()

	// 유저 정보 추출 & 권한 검사
	user, err := d.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	// Request DTO 바인딩 & 유효성 검증
	var request = adomain.AdminUserCreateRequest{
		JWTUser:  user,
		ClientIp: c.RealIP(),
	}
	if err = c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err = c.Validate(request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	response, err := d.adminUserUsecase.Create(ctx, request)

	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusCreated, response)
	case errors.Is(err, cdomain.ErrAlreadyExists):
		return factory.ErrorAlreadyExists.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (d adminUserHttpDelivery) Update(c echo.Context) error {
	// Only for super admin

	ctx := c.Request().Context()

	// 유저 정보 추출 & 권한 검사
	user, err := d.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	// Request DTO 바인딩 & 유효성 검증
	var request = adomain.AdminUserUpdateRequest{
		JWTUser:  user,
		ClientIp: c.RealIP(),
	}
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	response, err := d.adminUserUsecase.Update(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, response)
	case errors.Is(err, cdomain.ErrAlreadyExists):
		return factory.ErrorAlreadyExists.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (d adminUserHttpDelivery) Delete(c echo.Context) error {
	// Only for super admin

	ctx := c.Request().Context()

	// 유저 정보 추출 & 권한 검사
	user, err := d.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	// Request DTO 바인딩 & 유효성 검증
	var request = adomain.AdminUserDeleteRequest{
		JWTUser:  user,
		ClientIp: c.RealIP(),
	}
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	response, err := d.adminUserUsecase.Delete(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, response)
	case errors.Is(err, cdomain.ErrBadRequest):
		return factory.ErrorParameter.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (d adminUserHttpDelivery) Login(c echo.Context) error {
	ctx := c.Request().Context()

	// Request DTO 바인딩 & 유효성 검증
	var request = adomain.AdminUserLoginRequest{
		ClientIp:  c.RealIP(),
		UserAgent: c.Request().UserAgent(),
	}
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	response, err := d.adminUserUsecase.Login(ctx, request)
	switch {
	case err == nil:
		// cookie 에 refresh token 저장
		setLoginCookie(c, response.RefreshToken)
		// CDN 캐시 제어를 위한 쿠키 설정
		for _, cookie := range response.Cookies {
			c.SetCookie(cookie)
		}
		return apirender.RenderSucc(c, http.StatusOK, response)
	case errors.Is(err, cdomain.ErrItemNotFound):
		return factory.ErrorUserNotExists.New(err)
	case errors.Is(err, cdomain.ErrAlreadyDeleted):
		return factory.ErrorDeletedUser.New(err)
	case errors.Is(err, cdomain.ErrBadRequest):
		return factory.ErrorNotMatchPassword.New(err)
	case errors.Is(err, cdomain.ErrAlreadyExpired):
		return factory.ErrorExpireRequest.New(err)
	case errors.Is(err, cdomain.ErrTokenInvalid):
		return factory.ErrorFailAuthorized.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}

func setLoginCookie(c echo.Context, refreshToken string) {
	refreshCookie := new(http.Cookie)
	refreshCookie.Name = "refresh_token"
	refreshCookie.Value = refreshToken
	refreshCookie.HttpOnly = true
	refreshCookie.Secure = true
	refreshCookie.Path = "/"
	refreshCookie.Expires = time.Now().Add(domain.RefreshTokenExpireTime) // 쿠키 만료 시간 == 토큰 만료 시간

	c.SetCookie(refreshCookie)
}

func (d adminUserHttpDelivery) Logout(c echo.Context) error {
	ctx := c.Request().Context()

	// 유저 정보 추출
	user, err := d.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	// Request DTO 바인딩 & 유효성 검증
	var request = adomain.AdminUserLogoutRequest{
		JWTUser:   user,
		ClientIp:  c.RealIP(),
		UserAgent: c.Request().UserAgent(),
	}
	if err := c.Validate(request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// refresh token 쿠키에서 추출
	cookie, err := c.Cookie("refresh_token")
	if err != nil || cookie.Value == "" {
		return factory.ErrorTokenInvalid.New(err) // 쿠키가 없거나 에러가 발생한 경우
	}
	request.RefreshToken = cookie.Value

	response, err := d.adminUserUsecase.Logout(ctx, request)
	if err != nil {
		switch {
		case errors.Is(err, cdomain.ErrTokenInvalid):
			return factory.ErrorTokenInvalid.New(err)
		default:
			return factory.ErrorDB.New(err)
		}
	}

	expireRefreshTokenCookie(c) // 쿠키 삭제
	return apirender.RenderSucc(c, http.StatusOK, response)
}

func expireRefreshTokenCookie(c echo.Context) {
	refreshCookie := new(http.Cookie)
	refreshCookie.Name = "refresh_token"
	refreshCookie.Value = ""
	refreshCookie.HttpOnly = true
	refreshCookie.Secure = true
	refreshCookie.Path = "/"
	refreshCookie.Expires = time.Unix(0, 0) // 쿠키 만료 시간을 과거로 설정하여 즉시 만료시킴

	c.SetCookie(refreshCookie)
}
