package domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
)

type UserVerifyUsecase interface {
	CheckEmailExists(ctx context.Context, request UserCheckEmailExistsRequest) (bool, error)
	RequestPasswordReset(ctx context.Context, request UserRequestPasswordResetRequest) (UserRequestPasswordResetResponse, error)
	ResetPassword(ctx context.Context, request UserResetPasswordRequest) error
	GetPasswordResetRequest(ctx context.Context, request UserGetPasswordResetRequestRequest) (UserGetPasswordResetRequestResponse, error)
	ChangePassword(ctx context.Context, request UserChangePasswordRequest) error
	CheckPassword(ctx context.Context, request UserCheckPasswordRequest) (bool, error)
	SendSmsCode(ctx context.Context, request UserSendSmsCodeRequest) (UserSendSmsCodeResponse, error)
	ResendSmsCode(ctx context.Context, request UserResendSmsCodeRequest) (UserResendSmsCodeResponse, error)
	CheckSmsCode(ctx context.Context, request UserCheckSmsCodeRequest) (bool, error)
	SendEmailCode(ctx context.Context, request UserSendEmailCodeRequest) error
	ResendEmailCode(ctx context.Context, request UserResendEmailCodeRequest) error
	CheckEmailCode(ctx context.Context, request UserCheckEmailCodeRequest) error
}

type UserCheckEmailExistsRequest struct {
	Email string `query:"email" required:"true" validate:"required,email" format:"email" description:"メールアドレス" example:"gildong@email.com"`
}

type UserRequestPasswordResetRequest struct {
	Email string `json:"email" required:"true" validate:"required,email" format:"email" description:"メールアドレス" example:"gildong@email.com"`
}

type UserRequestPasswordResetResponse struct {
	Email string `json:"email" example:"gildong@email.com" description:"メールアドレス"`
}

type UserResetPasswordRequest struct {
	RequestId string `param:"request_id" path:"request_id" required:"true" validate:"required" description:"リクエストID" example:"682d3112-9a8b-4ee6-899d-43da0dda23c9"`
	Email     string `json:"email" required:"true" validate:"required,email" format:"email" description:"メールアドレス" example:"gildong@email.com"`
	Password  string `json:"password" required:"true" validate:"required" description:"リセットパスワード" example:"password1234!@"`
}

type UserGetPasswordResetRequestRequest struct {
	RequestId string `param:"request_id" path:"request_id" required:"true" validate:"required" description:"リクエストID" example:"682d3112-9a8b-4ee6-899d-43da0dda23c9"`
}

type UserGetPasswordResetRequestResponse struct {
	Email     string             `json:"email" example:"gildong@email.com" description:"メールアドレス"`
	RequestId string             `json:"request_id" example:"682d3112-9a8b-4ee6-899d-43da0dda23c9" description:"リクエストID"`
	Url       string             `json:"url" example:"https://ilm-es-dev.illuminarean.com/resetPwInput/20f7129c-0d75-49d0-bb3b-ed67cb1888fa" description:"パスワードリセットURL"`
	Expire    jsonutils.JSONTime `json:"expire" example:"2021-01-01 00:00:00" description:"有効期限"`
	SendCount int                `json:"send_count" description:"リクエスト回数" example:"1"`
}

type UserChangePasswordRequest struct {
	JWTUser  jwthelper.User `json:"-" swaggerignore:"true"`
	Password string         `json:"password" required:"true" validate:"required" description:"変更するパスワード" example:"password1234!@"`
}

type UserCheckPasswordRequest struct {
	JWTUser  jwthelper.User `json:"-" swaggerignore:"true"`
	Password string         `json:"password" required:"true" validate:"required" description:"リセットパスワード" example:"password1234!@"`
}

type UserSendSmsCodeRequest struct {
	JWTUser     jwthelper.User           `json:"-" swaggerignore:"true"`
	MobilePhone string                   `json:"mobile_phone" description:"携帯電話番号" validate:"required" example:"01012345678" format:"phone`
	PhoneType   domain.MOBILE_PHONE_TYPE `json:"type" description:"携帯電話番号タイプ" validate:"required" example:"DEFAULT" enum:"DEFAULT,CEO,MANAGER"`
}

type UserSendSmsCodeResponse struct {
	RequestId string `json:"request_id" description:"リクエストID" example:"682d3112-9a8b-4ee6-899d-43da0dda23c9"`
}

type UserResendSmsCodeRequest struct {
	JWTUser     jwthelper.User           `json:"-" swaggerignore:"true"`
	MobilePhone string                   `json:"mobile_phone" description:"携帯電話番号" validate:"required" example:"01012345678" format:"phone`
	PhoneType   domain.MOBILE_PHONE_TYPE `json:"type" description:"携帯電話番号タイプ" validate:"required" example:"DEFAULT" enum:"DEFAULT,CEO,MANAGER"`
}

type UserResendSmsCodeResponse struct {
	RequestId string `json:"request_id" description:"リクエストID" example:"682d3112-9a8b-4ee6-899d-43da0dda23c9"`
}

type UserCheckSmsCodeRequest struct {
	JWTUser   jwthelper.User `json:"-" swaggerignore:"true"`
	Code      string         `json:"code" description:"認証コード" validate:"required" example:"123456"`
	RequestId string         `json:"request_id" required:"true" validate:"required" description:"リクエストID" example:"682d3112-9a8b-4ee6-899d-43da0dda23c9"`
}

type UserSendEmailCodeRequest struct {
	JWTUser jwthelper.User `json:"-" swaggerignore:"true"`
	Email   string         `json:"email" required:"true" validate:"required,email" format:"email" description:"メールアドレス" example:"gildong@email.com"`
}

type UserResendEmailCodeRequest struct {
	JWTUser jwthelper.User `json:"-" swaggerignore:"true"`
	Email   string         `json:"email" required:"true" validate:"required,email" format:"email" description:"メールアドレス" example:"gildong@email.com"`
}

type UserCheckEmailCodeRequest struct {
	JWTUser jwthelper.User `json:"-" swaggerignore:"true"`
	Email   string         `json:"email" description:"メールアドレス" required:"true" format:"email" validate:"required,email" example:"gildong@email.com"`
	Code    string         `json:"code" description:"認証コード" required:"true" validate:"required" example:"123456"`
}
