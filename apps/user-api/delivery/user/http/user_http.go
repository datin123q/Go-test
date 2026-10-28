package http

import (
	"net/http"
	"time"

	"github.com/pkg/errors"
	openapigo "github.com/swaggest/openapi-go"

	"github.com/labstack/echo/v4"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/apirender"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/echoutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	udomain "github.com/illuminarean/energyshares-user-api/domain"
)

type userHttpDelivery struct {
	userUsecase      udomain.UserUsecase
	noticeMsgUsecase udomain.NoticeMsgUsecase
	api              *echoutils.EchoSwagger
	extractor        jwthelper.UserExtractor
}

func New(
	api *echoutils.EchoSwagger,
	userUsecase udomain.UserUsecase,
	noticeMsgUsecase udomain.NoticeMsgUsecase,
	extractor jwthelper.UserExtractor,
) *userHttpDelivery {
	handler := &userHttpDelivery{
		userUsecase:      userUsecase,
		noticeMsgUsecase: noticeMsgUsecase,
		extractor:        extractor,
		api:              api,
	}

	g := api.Group("/api/v1/users")

	g.POST("/login", handler.Login).
		SetTags("共通ユーザー").
		SetSummary("ログインを実行します").
		AddRequest(udomain.UserLoginRequest{}).
		AddResponse(udomain.UserLoginResponse{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "ログイン実行成功"
		}).
		AddErrors(
			factory.ErrorUserNotExists,
			factory.ErrorInvalidityAudit,
			factory.ErrorUserLocked,
			factory.ErrorTooManyRequests,
			factory.ErrorNotMatchPassword,
			factory.ErrorParameter,
			factory.ErrorDB,
		)

	g.POST("/signup-requests", handler.RequestSignUp).
		SetTags("共通ユーザー").
		SetSummary("(仮登録) メール認証及び加入リクエスト").
		AddRequest(udomain.UserRequestSignUpRequest{}).
		AddResponse(udomain.UserRequestSignUpResponse{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "加入リクエスト成功"
		}).
		AddErrors(
			factory.ErrorAlreadyExists,
			factory.ErrorExpireRequest,
			factory.ErrorSentEmail,
			factory.ErrorParameter,
			factory.ErrorDB,
		)

	g.GET("/signup-requests/:request_id", handler.CheckSignupRequestValidity).
		SetTags("共通ユーザー").
		SetSummary("会員加入リクエスト妥当性チェック（メール検証）").
		AddRequest(udomain.UserCheckSignupRequestValidityRequest{}).
		AddResponse(udomain.UserCheckSignupRequestValidityResponse{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "会員加入リクエスト妥当性チェック成功（無効な場合はエラーを返す）"
		}).
		AddErrors(
			factory.ErrorParameter,
			factory.ErrorNotFound,
			factory.ErrorAlreadyExists,
			factory.ErrorExpireRequest,
			factory.ErrorDB,
		)

	g.POST("/signup-requests/:request_id/verify", handler.VerifyEmail).
		SetTags("共通ユーザー").
		SetSummary("会員加入リクエストメール認証").
		AddRequest(udomain.UserVerifyEmailRequest{}).
		AddResponse(nil, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "会員加入リクエストのメール認証成功"
		}).
		AddErrors(
			factory.ErrorParameter,
			factory.ErrorExpireRequest,
			factory.ErrorNotFound,
			factory.ErrorDB,
		)

	g.POST("/signup", handler.SignUp).
		SetTags("共通ユーザー").
		SetSummary("仮登録会員加入").
		AddRequest(udomain.UserSignUpRequest{}).
		AddResponse(nil, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "（仮登録）会員加入成功"
		}).
		AddErrors(
			factory.ErrorParameter,
			factory.ErrorSignUpEmailMismatch,
			factory.ErrorNotFound,
			factory.ErrorDB,
		)

	g.POST("/pre-signed-url", handler.CreatePreSignedUrl,
		echoutils.AllowRoles(domain.ROLE_USER)).
		SetTags("共通ユーザー").
		SetSummary("ユーザードメインpre-signed URL生成").
		AddRequest(udomain.UserCreatePreSignedUrlRequest{}).
		AddResponse(udomain.UserCreatePreSignedUrlResponse{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "pre-signed URL生成成功"
		}).
		AddSecurity("Authorization").
		AddErrors(
			factory.ErrorParameter,
			factory.ErrorDB,
		)

	return handler
}

func (u userHttpDelivery) Login(c echo.Context) error {
	ctx := c.Request().Context()

	// Request DTO作成
	var request = udomain.UserLoginRequest{
		ClientIp:  c.RealIP(),
		UserAgent: c.Request().UserAgent(),
	}

	// Request DTOバインド
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// Request DTO妥当性検証
	if err := c.Validate(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	response, err := u.userUsecase.Login(ctx, request)
	switch {
	case err == nil:
		// クッキーにリフレッシュトークンを保存
		setLoginCookie(c, response.RefreshToken)
		// CDNキャッシュ制御のためのクッキー設定
		for _, cookie := range response.Cookies {
			c.SetCookie(cookie)
		}
		return apirender.RenderSucc(c, http.StatusOK, response)
	case errors.Is(err, cdomain.ErrItemNotFound): // 無効なメールでログイン試行
		return factory.ErrorUserNotExists.New(err)
	case errors.Is(err, cdomain.ErrAuditInvalid): // 「無効」状態のアカウントでログイン試行
		return factory.ErrorInvalidityAudit.New(err)
	case errors.Is(err, cdomain.ErrResourceLocked): // ロックされたアカウントでログイン試行
		return factory.ErrorUserLocked.New(err)
	case errors.Is(err, cdomain.ErrTooManyRequests): // 4回以上失敗したアカウントでログイン試行
		return factory.ErrorTooManyRequests.New(err)
	case errors.Is(err, cdomain.ErrItemInvalid): // パスワード不一致
		return factory.ErrorNotMatchPassword.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (u userHttpDelivery) RequestSignUp(c echo.Context) error {
	ctx := c.Request().Context()

	// Request DTO作成
	var request = udomain.UserRequestSignUpRequest{}

	// Request DTOバインド
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// Request DTO妥当性検証
	if err := request.Valid(c); err != nil {
		return factory.ErrorParameter.New(err)
	}

	response, err := u.userUsecase.RequestSignUp(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, response)
	case errors.Is(err, cdomain.ErrAlreadyExists):
		return factory.ErrorAlreadyExists.New(err)
	case errors.Is(err, cdomain.ErrAlreadyExpired):
		return factory.ErrorExpireRequest.New(err)
	case errors.Is(err, cdomain.ErrRequestLimitExceeded):
		return factory.ErrorSentEmail.New(err)
	case errors.Is(err, cdomain.ErrBadRequest):
		return factory.ErrorParameter.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (u userHttpDelivery) CheckSignupRequestValidity(c echo.Context) error {
	ctx := c.Request().Context()

	// Request DTO作成
	var request = udomain.UserCheckSignupRequestValidityRequest{}

	// Request DTOバインド
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// Request DTO妥当性検証
	if err := c.Validate(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	response, err := u.userUsecase.CheckSignupRequestValidity(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, response)
	case errors.Is(err, cdomain.ErrItemNotFound):
		return factory.ErrorNotFound.New(err)
	case errors.Is(err, cdomain.ErrAlreadyExists):
		return factory.ErrorAlreadyExists.New(err)
	case errors.Is(err, cdomain.ErrAlreadyExpired):
		return factory.ErrorExpireRequest.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (u userHttpDelivery) SignUp(c echo.Context) error {
	ctx := c.Request().Context()

	// Request DTO作成
	var request = udomain.UserSignUpRequest{}

	// Request DTOバインド
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// Request DTO妥当性検証
	if err := request.Valid(c); err != nil {
		return factory.ErrorParameter.New(err)
	}

	err := u.userUsecase.SignUp(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusNoContent, "")
	case errors.Is(err, cdomain.ErrBadRequest):
		return factory.ErrorParameter.New(err)
	case errors.Is(err, cdomain.ErrItemConflict):
		return factory.ErrorSignUpEmailMismatch.New(err)
	case errors.Is(err, cdomain.ErrItemNotFound):
		return factory.ErrorNotFound.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (u userHttpDelivery) CreatePreSignedUrl(c echo.Context) error {
	ctx := c.Request().Context()

	// Request DTO作成
	var request = udomain.UserCreatePreSignedUrlRequest{}

	// Request DTOバインド
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// Request DTO妥当性検証
	if err := c.Validate(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	response, err := u.userUsecase.CreatePreSignedUrl(ctx, request)
	if err != nil {
		return err
	}

	return apirender.RenderSucc(c, http.StatusOK, response)
}

func (u userHttpDelivery) VerifyEmail(c echo.Context) error {
	ctx := c.Request().Context()

	// Request DTO作成
	var request = udomain.UserVerifyEmailRequest{}

	// Request DTOバインド
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// Request DTO妥当性検証
	if err := c.Validate(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	err := u.userUsecase.VerifyEmail(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusNoContent, "") // No Content
	case errors.Is(err, cdomain.ErrAlreadyExpired):
		return factory.ErrorExpireRequest.New(err)
	case errors.Is(err, cdomain.ErrItemNotFound):
		return factory.ErrorNotFound.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}

func setLoginCookie(c echo.Context, refreshToken string) {
	cookie := new(http.Cookie)
	cookie.Name = "refresh_token"
	cookie.Value = refreshToken

	if serviceutils.EnvIsDev() {
		cookie.Expires = time.Now().Add(domain.TestRefreshTokenCookieExpireTime)
	} else {
		cookie.Expires = time.Now().Add(domain.RefreshTokenCookieExpireTime)
	}

	cookie.Path = "/"
	cookie.HttpOnly = true
	cookie.Secure = true
	c.SetCookie(cookie)
}
