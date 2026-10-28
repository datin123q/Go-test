package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	openapigo "github.com/swaggest/openapi-go"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/apirender"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/echoutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	udomain "github.com/illuminarean/energyshares-user-api/domain"
)

type userVerifyHttpDelivery struct {
	userVerifyUsecase udomain.UserVerifyUsecase
	api               *echoutils.EchoSwagger
	extractor         jwthelper.UserExtractor
}

func New(
	api *echoutils.EchoSwagger,
	userVerifyUsecase udomain.UserVerifyUsecase,
	extractor jwthelper.UserExtractor,
) *userVerifyHttpDelivery {
	handler := &userVerifyHttpDelivery{
		api:               api,
		userVerifyUsecase: userVerifyUsecase,
		extractor:         extractor,
	}

	g := api.Group("/api/v1/users-verify")

	g.GET("/email-exists", handler.CheckEmailExists).
		SetTags("共通ユーザー認証").
		SetSummary("メール重複チェック").
		AddRequest(udomain.UserCheckEmailExistsRequest{}).
		AddResponse(nil, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "メールがすでに存在する場合はtrue、存在しない場合はfalseを返す"
		}).
		SetDescription("メール重複チェック").
		AddErrors(
			factory.ErrorParameter,
			factory.ErrorDB,
		)

	// passwordresetrequest
	g.POST("/password-reset", handler.RequestPasswordReset).
		SetTags("共通ユーザー認証").
		SetSummary("パスワードリセットリクエスト").
		AddRequest(udomain.UserRequestPasswordResetRequest{}).
		AddResponse(nil, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "パスワードリセット応答値"
		}).
		SetDescription("パスワードリセットリクエスト: 有効期限20分、再試行1回まで可能（1日合計2回）").
		AddErrors(
			factory.ErrorParameter,
			factory.ErrorUserNotExists,
			factory.ErrorSentEmail,
			factory.ErrorInvalidAuditStatus,
			factory.ErrorDB,
		)

	g.POST("/password-reset/:request_id", handler.ResetPassword).
		SetTags("共通ユーザー認証").
		SetSummary("パスワードリセット").
		AddRequest(udomain.UserResetPasswordRequest{}).
		AddResponse(nil, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "パスワードリセット / 成功時にtrue値を返す"
		}).
		AddErrors(
			factory.ErrorParameter,
			factory.ErrorInvalidPassword,
			factory.ErrorSamePassword,
			factory.ErrorExpireRequest,
			factory.ErrorInvalidAuditStatus,
			factory.ErrorDB,
		)

	// passwordresetrequest/:id
	g.GET("/password-reset/:request_id", handler.GetPasswordResetRequest).
		SetTags("共通ユーザー認証").
		SetSummary("パスワードリセットリクエスト照会").
		AddRequest(udomain.UserGetPasswordResetRequestRequest{}).
		AddResponse(udomain.UserGetPasswordResetRequestResponse{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "パスワードリセットリクエスト照会応答値"
		}).
		SetDescription("パスワードリセットリクエスト照会（有効期限切れの場合は照会できない）").
		AddErrors(
			factory.ErrorParameter,
			factory.ErrorNotFound,
			factory.ErrorExpireRequest,
			factory.ErrorDB,
		)

	g.PUT("/password", handler.ChangePassword,
		echoutils.AllowRoles(domain.ROLE_USER)).
		SetTags("共通ユーザー認証").
		SetSummary("パスワード変更").
		AddRequest(udomain.UserChangePasswordRequest{}).
		AddResponse(nil, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "パスワード変更 / 成功時にtrue値を返す"
		}).
		AddSecurity("Authorization").
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorParameter,
			factory.ErrorInvalidPassword,
			factory.ErrorSamePassword,
			factory.ErrorDB,
		)

	// :id/passwordcheck
	// TODO 確認が必要: (既存) 不一致時にErrorNotMatchPasswordエラー (修正) falseを返す
	g.POST("/password-check", handler.CheckPassword,
		echoutils.AllowRoles(domain.ROLE_USER)).
		SetTags("共通ユーザー認証").
		SetSummary("パスワード一致確認").
		AddRequest(udomain.UserCheckPasswordRequest{}).
		AddResponse(nil, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "パスワード一致確認 / 一致時はtrue、不一致時はfalseを返す"
		}).
		SetDescription("パスワード一致確認").
		AddSecurity("Authorization").
		AddErrors(
			factory.ErrorUserNotExists,
			factory.ErrorFailAuthorized,
			factory.ErrorParameter,
			factory.ErrorDB,
		)

	// :id/mobile/verification-request
	g.POST("/sms/send-code", handler.SendSmsCode,
		echoutils.AllowRoles(domain.ROLE_USER)).
		SetTags("共通ユーザー認証").
		SetSummary("SMS認証コード初回送信").
		AddRequest(udomain.UserSendSmsCodeRequest{}).
		AddResponse(udomain.UserSendSmsCodeResponse{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "SMS認証コード初回送信応答値"
		}).
		SetDescription("SMS認証コード初回送信（24時間以内3回まで可能）").
		AddSecurity("Authorization")

	// :id/mobile/verification-re-request
	g.POST("/sms/resend-code", handler.ResendSmsCode,
		echoutils.AllowRoles(domain.ROLE_USER)).
		SetTags("共通ユーザー認証").
		SetSummary("SMS認証コード再送信").
		AddRequest(udomain.UserResendSmsCodeRequest{}).
		AddResponse(udomain.UserResendSmsCodeResponse{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "SMS認証コード再送信応答値"
		}).
		SetDescription("SMS認証コード再送信（初回送信後5回まで可能）").
		AddSecurity("Authorization").
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorParameter,
			factory.ErrorDB,
		)

	// :id/mobile/verification
	// TODO 確認が必要: (既存) MobilePhoneVerificationRequestを返す (修正) boolを返す
	g.POST("/sms/check-code", handler.CheckSmsCode,
		echoutils.AllowRoles(domain.ROLE_USER)).
		SetTags("共通ユーザー認証").
		SetSummary("SMS認証コード確認").
		AddRequest(udomain.UserCheckSmsCodeRequest{}).
		AddResponse(nil, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "SMS認証コード確認 / 成功時にtrue値を返す"
		}).
		AddSecurity("Authorization").
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorParameter,
			factory.ErrorDB,
		)

	// /email-code-request
	g.POST("/email/send-code", handler.SendEmailCode,
		echoutils.AllowRoles(domain.ROLE_USER)).
		SetTags("共通ユーザー認証").
		SetSummary("メールコード送信").
		SetDescription("この関数を呼び出すと、メール変更のための認証メールが送信されます。").
		AddRequest(udomain.UserVerifyEmailChangedRequest{}).
		AddResponse(nil, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "メールコード送信成功"
		}).
		AddSecurity("Authorization").
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorParameter,
			factory.ErrorSameEmailExists,
			factory.ErrorMaxTransmissionsReached,
			factory.ErrorDB,
		)

	g.POST("/email/resend-code", handler.ResendEmailCode,
		echoutils.AllowRoles(domain.ROLE_USER)).
		SetTags("共通ユーザー認証").
		SetSummary("メールコード再送信").
		SetDescription("この関数を呼び出すと、メール変更のための認証メールが再送信されます。").
		AddRequest(udomain.UserVerifyEmailChangedRequest{}).
		AddResponse(nil, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "メールコード再送信成功"
		}).
		AddSecurity("Authorization").
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorParameter,
			factory.ErrorSameEmailExists,
			factory.ErrorMaxTransmissionsReached,
			factory.ErrorDB,
		)

	g.POST("/email/check-code", handler.CheckEmailCode,
		echoutils.AllowRoles(domain.ROLE_USER)).
		SetTags("共通ユーザー認証").
		SetSummary("メールコード確認").
		AddRequest(udomain.UserCheckEmailCodeRequest{}).
		AddResponse(nil, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "メールコード確認成功"
		}).
		AddSecurity("Authorization").
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorAlreadyExists,
			factory.ErrorParameter,
			factory.ErrorCodeMismatch,
			factory.ErrorExpireRequest,
			factory.ErrorDB,
		)

	return handler
}

// CheckEmailExists メールが存在する場合はtrue、存在しない場合はfalseを返す
func (u userVerifyHttpDelivery) CheckEmailExists(c echo.Context) error {
	ctx := c.Request().Context()

	// create & binding Request DTO
	var request = udomain.UserCheckEmailExistsRequest{}
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	response, err := u.userVerifyUsecase.CheckEmailExists(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, response)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (u userVerifyHttpDelivery) RequestPasswordReset(c echo.Context) error {
	ctx := c.Request().Context()

	// create & binding Request DTO
	var request = udomain.UserRequestPasswordResetRequest{}
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	response, err := u.userVerifyUsecase.RequestPasswordReset(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, response)
	case errors.Is(err, cdomain.ErrItemNotFound):
		return factory.ErrorUserNotExists.New(err)
	case errors.Is(err, cdomain.ErrRequestLimitExceeded):
		return factory.ErrorSentEmail.New(err)
	case errors.Is(err, cdomain.ErrAuditInvalid):
		return factory.ErrorInvalidAuditStatus.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (u userVerifyHttpDelivery) ResetPassword(c echo.Context) error {
	ctx := c.Request().Context()

	// create & binding Request DTO
	var request = udomain.UserResetPasswordRequest{}
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	err := u.userVerifyUsecase.ResetPassword(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, true)
	case errors.Is(err, cdomain.ErrBadRequest):
		return factory.ErrorParameter.New(err)
	case errors.Is(err, cdomain.ErrItemInvalid):
		return factory.ErrorInvalidPassword.New(err)
	case errors.Is(err, cdomain.ErrItemConflict):
		return factory.ErrorSamePassword.New(err)
	case errors.Is(err, cdomain.ErrAlreadyExpired):
		return factory.ErrorExpireRequest.New(err)
	case errors.Is(err, cdomain.ErrAuditInvalid):
		return factory.ErrorInvalidAuditStatus.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (u userVerifyHttpDelivery) GetPasswordResetRequest(c echo.Context) error {
	ctx := c.Request().Context()

	// create & binding Request DTO
	var request = udomain.UserGetPasswordResetRequestRequest{}
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	response, err := u.userVerifyUsecase.GetPasswordResetRequest(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, response)
	case errors.Is(err, cdomain.ErrItemNotFound):
		return factory.ErrorNotFound.New(err)
	case errors.Is(err, cdomain.ErrAlreadyExpired):
		return factory.ErrorExpireRequest.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (u userVerifyHttpDelivery) ChangePassword(c echo.Context) error {
	user, err := u.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	ctx := c.Request().Context()

	// create & binding Request DTO
	var request = udomain.UserChangePasswordRequest{
		JWTUser: user,
	}
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	err = u.userVerifyUsecase.ChangePassword(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, true)
	case errors.Is(err, cdomain.ErrItemInvalid):
		return factory.ErrorInvalidPassword.New(err)
	case errors.Is(err, cdomain.ErrItemConflict):
		return factory.ErrorSamePassword.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (u userVerifyHttpDelivery) CheckPassword(c echo.Context) error {
	user, err := u.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	ctx := c.Request().Context()

	// create & binding Request DTO
	var request = udomain.UserCheckPasswordRequest{
		JWTUser: user,
	}
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	response, err := u.userVerifyUsecase.CheckPassword(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, response)
	case errors.Is(err, cdomain.ErrItemNotFound):
		return factory.ErrorUserNotExists.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (u userVerifyHttpDelivery) SendSmsCode(c echo.Context) error {
	// TODO: (UserApiController) RequestSMSVerification
	// TODO: (CorporateApiController) RequestSMSVerification
	user, err := u.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	ctx := c.Request().Context()

	// create & binding Request DTO
	var request = udomain.UserSendSmsCodeRequest{
		JWTUser: user,
	}
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	response, err := u.userVerifyUsecase.SendSmsCode(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, response)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (u userVerifyHttpDelivery) ResendSmsCode(c echo.Context) error {
	// TODO: (UserApiController) ReRequestSMSVerification
	// TODO: (CorporateApiController) ReRequestSMSVerification
	user, err := u.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	ctx := c.Request().Context()

	// create & binding Request DTO
	var request = udomain.UserResendSmsCodeRequest{
		JWTUser: user,
	}
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	response, err := u.userVerifyUsecase.ResendSmsCode(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, response)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (u userVerifyHttpDelivery) CheckSmsCode(c echo.Context) error {
	// TODO: (a UserApiController) ConfirmSmsVerification
	return nil
}

func (u userVerifyHttpDelivery) SendEmailCode(c echo.Context) error {
	ctx := c.Request().Context()

	// ユーザー情報抽出
	user, err := u.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	// Request DTO作成
	var request = udomain.UserSendEmailCodeRequest{
		JWTUser: user,
	}

	// Request DTOバインド
	if err = c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// Request DTO妥当性検証
	if err = c.Validate(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	err = u.userVerifyUsecase.SendEmailCode(ctx, request)

	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, "")
	case errors.Is(err, cdomain.ErrAlreadyExists):
		return factory.ErrorSameEmailExists.New(err)
	case errors.Is(err, cdomain.ErrRequestLimitExceeded):
		return factory.ErrorMaxTransmissionsReached.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (u userVerifyHttpDelivery) ResendEmailCode(c echo.Context) error {
	ctx := c.Request().Context()

	// ユーザー情報抽出
	user, err := u.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	// Request DTO作成
	var request = udomain.UserResendEmailCodeRequest{
		JWTUser: user,
	}

	// Request DTOバインド
	if err = c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// Request DTO妥当性検証
	if err = c.Validate(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	err = u.userVerifyUsecase.ResendEmailCode(ctx, request)

	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, "")
	case errors.Is(err, cdomain.ErrAlreadyExists):
		return factory.ErrorSameEmailExists.New(err)
	case errors.Is(err, cdomain.ErrRequestLimitExceeded):
		return factory.ErrorMaxTransmissionsReached.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}

func (u userVerifyHttpDelivery) CheckEmailCode(c echo.Context) error {
	ctx := c.Request().Context()

	// ユーザー情報抽出
	user, err := u.extractor.Extract(c)
	if err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	// Request DTO作成
	var request = udomain.UserCheckEmailCodeRequest{
		JWTUser: user,
	}

	// Request DTOバインド
	if err = c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// Request DTO妥当性検証
	if err = c.Validate(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	err = u.userVerifyUsecase.CheckEmailCode(ctx, request)
	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, "")
	case errors.Is(err, cdomain.ErrAlreadyExists):
		return factory.ErrorAlreadyExists.New(err)
	case errors.Is(err, cdomain.ErrItemNotFound):
		return factory.ErrorCodeMismatch.New(err)
	case errors.Is(err, cdomain.ErrAlreadyExpired):
		return factory.ErrorExpireRequest.New(err)
	default:
		return factory.ErrorDB.New(err)
	}
}
