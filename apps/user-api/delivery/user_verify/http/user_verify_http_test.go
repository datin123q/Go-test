package http

import (
	"net/http"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	pkgtest "github.com/illuminarean/energy-share-core-api/delivery/testing"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/validator"
	udomain "github.com/illuminarean/energyshares-user-api/domain"
	"github.com/illuminarean/energyshares-user-api/domain/mocks"
)

const userVerifyApi = "/api/v1/users-verify"

func Test_userVerifyHttpDelivery_GetPasswordResetRequest(t *testing.T) {
	tests := []struct {
		name              string
		echoCtx           func(e *echo.Echo) echo.Context
		userVerifyUsecase func() *mocks.UserVerifyUsecase
		wantErr           bool
	}{
		{
			name: "GET /api/v1/users-verify/password-reset/:request_id 성공",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   userVerifyApi + "/password-reset/:request_id",
				Method: http.MethodGet,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Params: map[string]string{
					"request_id": "1234",
				},
			}),
			userVerifyUsecase: func() *mocks.UserVerifyUsecase {
				usecase := &mocks.UserVerifyUsecase{}
				usecase.On("GetPasswordResetRequest", mock.Anything, udomain.UserGetPasswordResetRequestRequest{
					RequestId: "1234",
				}).Return(udomain.UserGetPasswordResetRequestResponse{
					Email:     "test@test.com",
					RequestId: "1234",
					Url:       "http://test.com",
					Expire:    jsonutils.JSONTime{},
					SendCount: 1,
				}, nil)
				return usecase
			},
			wantErr: false,
		},
		{
			name: "GET /api/v1/users-verify/password-reset/:request_id 존재하지 않는 request_id 일때",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   userVerifyApi + "/password-reset/:request_id",
				Method: http.MethodGet,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Params: map[string]string{
					"request_id": "1234",
				},
			}),
			userVerifyUsecase: func() *mocks.UserVerifyUsecase {
				usecase := &mocks.UserVerifyUsecase{}
				usecase.On("GetPasswordResetRequest", mock.Anything, udomain.UserGetPasswordResetRequestRequest{
					RequestId: "1234",
				}).Return(udomain.UserGetPasswordResetRequestResponse{}, cdomain.ErrItemNotFound)
				return usecase
			},
			wantErr: true,
		},
		{
			name: "GET /api/v1/users-verify/password-reset/:request_id 만료된 요청일 때",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   userVerifyApi + "/password-reset/:request_id",
				Method: http.MethodGet,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Params: map[string]string{
					"request_id": "1234",
				},
			}),
			userVerifyUsecase: func() *mocks.UserVerifyUsecase {
				usecase := &mocks.UserVerifyUsecase{}
				usecase.On("GetPasswordResetRequest", mock.Anything, udomain.UserGetPasswordResetRequestRequest{
					RequestId: "1234",
				}).Return(udomain.UserGetPasswordResetRequestResponse{}, cdomain.ErrAlreadyExpired)
				return usecase
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := userVerifyHttpDelivery{
				userVerifyUsecase: tt.userVerifyUsecase(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.GetPasswordResetRequest(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "GetPasswordResetRequest() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_userVerifyHttpDelivery_RequestPasswordReset(t *testing.T) {
	tests := []struct {
		name              string
		echoCtx           func(e *echo.Echo) echo.Context
		userVerifyUsecase func() *mocks.UserVerifyUsecase
		wantErr           bool
	}{
		{
			name: "POST /api/v1/users-verify/password-reset 성공",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   userVerifyApi + "/password-reset",
				Method: http.MethodPost,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Body: `{"email": "test@test.com"}`,
			}),
			userVerifyUsecase: func() *mocks.UserVerifyUsecase {
				usecase := &mocks.UserVerifyUsecase{}
				usecase.On("RequestPasswordReset", mock.Anything, udomain.UserRequestPasswordResetRequest{
					Email: "test@test.com",
				}).Return(udomain.UserRequestPasswordResetResponse{Email: "test@test.com"}, nil)
				return usecase
			},
			wantErr: false,
		},
		{
			name: "POST /api/v1/users-verify/password-reset 존재하지 않는 유저일 때",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   userVerifyApi + "/password-reset",
				Method: http.MethodPost,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Body: `{"email": "test@test.com"}`,
			}),
			userVerifyUsecase: func() *mocks.UserVerifyUsecase {
				usecase := &mocks.UserVerifyUsecase{}
				usecase.On("RequestPasswordReset", mock.Anything, udomain.UserRequestPasswordResetRequest{
					Email: "test@test.com",
				}).Return(udomain.UserRequestPasswordResetResponse{}, cdomain.ErrItemNotFound)
				return usecase
			},
			wantErr: true,
		},
		{
			name: "POST /api/v1/users-verify/password-reset 요청 횟수를 초과했을 때",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   userVerifyApi + "/password-reset",
				Method: http.MethodPost,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Body: `{"email": "test@test.com"}`,
			}),
			userVerifyUsecase: func() *mocks.UserVerifyUsecase {
				usecase := &mocks.UserVerifyUsecase{}
				usecase.On("RequestPasswordReset", mock.Anything, udomain.UserRequestPasswordResetRequest{
					Email: "test@test.com",
				}).Return(udomain.UserRequestPasswordResetResponse{}, cdomain.ErrRequestLimitExceeded)
				return usecase
			},
			wantErr: true,
		},
		{
			name: "POST /api/v1/users-verify/password-reset 만료된 요청일 때",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   userVerifyApi + "/password-reset",
				Method: http.MethodPost,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Body: `{"email": "test@test.com"}`,
			}),
			userVerifyUsecase: func() *mocks.UserVerifyUsecase {
				usecase := &mocks.UserVerifyUsecase{}
				usecase.On("RequestPasswordReset", mock.Anything, udomain.UserRequestPasswordResetRequest{
					Email: "test@test.com",
				}).Return(udomain.UserRequestPasswordResetResponse{}, cdomain.ErrAlreadyExpired)
				return usecase
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := userVerifyHttpDelivery{
				userVerifyUsecase: tt.userVerifyUsecase(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.RequestPasswordReset(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "RequestPasswordReset() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_userVerifyHttpDelivery_ResetPassword(t *testing.T) {
	tests := []struct {
		name              string
		echoCtx           func(e *echo.Echo) echo.Context
		userVerifyUsecase func() *mocks.UserVerifyUsecase
		wantErr           bool
	}{
		{
			name: "POST /api/v1/users-verify/password-reset/:request_id 성공",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   userVerifyApi + "/password-reset/:request_id",
				Method: http.MethodPost,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Params: map[string]string{
					"request_id": "1234",
				},
				Body: `{"email": "test@test.com", "password" : "Qwer1234!@#$"}`,
			}),
			userVerifyUsecase: func() *mocks.UserVerifyUsecase {
				usecase := &mocks.UserVerifyUsecase{}
				usecase.On("ResetPassword", mock.Anything, udomain.UserResetPasswordRequest{
					RequestId: "1234",
					Email:     "test@test.com",
					Password:  "Qwer1234!@#$",
				}).Return(nil)
				return usecase
			},
			wantErr: false,
		},
		{
			name: "POST /api/v1/users-verify/password 유효하지 않은 비밀번호 형식일 때",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   userVerifyApi + "/password-reset/:request_id",
				Method: http.MethodPost,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Params: map[string]string{
					"request_id": "1234",
				},
				Body: `{"email": "test@test.com", "password" : "gildong@email.com"}`,
			}),
			userVerifyUsecase: func() *mocks.UserVerifyUsecase {
				usecase := &mocks.UserVerifyUsecase{}
				usecase.On("ResetPassword", mock.Anything, udomain.UserResetPasswordRequest{
					RequestId: "1234",
					Email:     "test@test.com",
					Password:  "gildong@email.com",
				}).Return(cdomain.ErrItemInvalid)
				return usecase
			},
			wantErr: true,
		},
		{
			name: "POST /api/v1/users-verify/password 기존 비밀번호와 동일할 때",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   userVerifyApi + "/password-reset/:request_id",
				Method: http.MethodPost,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Params: map[string]string{
					"request_id": "1234",
				},
				Body: `{"email": "test@test.com", "password" : "password"}`,
			}),
			userVerifyUsecase: func() *mocks.UserVerifyUsecase {
				usecase := &mocks.UserVerifyUsecase{}
				usecase.On("ResetPassword", mock.Anything, udomain.UserResetPasswordRequest{
					RequestId: "1234",
					Email:     "test@test.com",
					Password:  "password",
				}).Return(cdomain.ErrItemConflict)
				return usecase
			},
			wantErr: true,
		},
		{
			name: "POST /api/v1/users-verify/password 만료된 요청일 때",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   userVerifyApi + "/password-reset/:request_id",
				Method: http.MethodPost,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Params: map[string]string{
					"request_id": "1234",
				},
				Body: `{"email": "test@test.com", "password" : "password"}`,
			}),
			userVerifyUsecase: func() *mocks.UserVerifyUsecase {
				usecase := &mocks.UserVerifyUsecase{}
				usecase.On("ResetPassword", mock.Anything, udomain.UserResetPasswordRequest{
					RequestId: "1234",
					Email:     "test@test.com",
					Password:  "password",
				}).Return(cdomain.ErrAlreadyExpired)
				return usecase
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := userVerifyHttpDelivery{
				userVerifyUsecase: tt.userVerifyUsecase(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.ResetPassword(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "ResetPassword() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_userVerifyHttpDelivery_ChangePassword(t *testing.T) {
	tests := []struct {
		name              string
		echoCtx           func(e *echo.Echo) echo.Context
		userVerifyUsecase func() *mocks.UserVerifyUsecase
		wantErr           bool
	}{
		{
			name: "PUT /api/v1/users-verify/password 성공",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   userVerifyApi + "/password",
				Method: http.MethodPut,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Body: `{"password" : "Qwer1234!@#$"}`,
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_USER"},
				},
			}),
			userVerifyUsecase: func() *mocks.UserVerifyUsecase {
				usecase := &mocks.UserVerifyUsecase{}
				usecase.On("ChangePassword", mock.Anything, udomain.UserChangePasswordRequest{
					JWTUser: jwthelper.User{
						Id:    int64(1000),
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{
							"ROLE_USER": 1,
						},
					},
					Password: "Qwer1234!@#$",
				}).Return(nil)
				return usecase
			},
			wantErr: false,
		},
		{
			name: "PUT /api/v1/users-verify/password 유효하지 않은 비밀번호 형식일 때",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   userVerifyApi + "/password",
				Method: http.MethodPut,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Body: `{"password" : "gildong@email.com"}`,
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_USER"},
				},
			}),
			userVerifyUsecase: func() *mocks.UserVerifyUsecase {
				usecase := &mocks.UserVerifyUsecase{}
				usecase.On("ChangePassword", mock.Anything, udomain.UserChangePasswordRequest{
					JWTUser: jwthelper.User{
						Id:    int64(1000),
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{
							"ROLE_USER": 1,
						},
					},
					Password: "gildong@email.com",
				}).Return(cdomain.ErrItemInvalid)
				return usecase
			},
			wantErr: true,
		},
		{
			name: "PUT /api/v1/users-verify/password 기존 비밀번호와 동일할 때",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   userVerifyApi + "/password",
				Method: http.MethodPut,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Body: `{"password" : "password"}`,
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_USER"},
				},
			}),
			userVerifyUsecase: func() *mocks.UserVerifyUsecase {
				usecase := &mocks.UserVerifyUsecase{}
				usecase.On("ChangePassword", mock.Anything, udomain.UserChangePasswordRequest{
					JWTUser: jwthelper.User{
						Id:    int64(1000),
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{
							"ROLE_USER": 1,
						},
					},
					Password: "password",
				}).Return(cdomain.ErrItemConflict)
				return usecase
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := userVerifyHttpDelivery{
				userVerifyUsecase: tt.userVerifyUsecase(),
				extractor:         jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.ChangePassword(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "ChangePassword() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_userVerifyHttpDelivery_SendEmailCode(t *testing.T) {
	tests := []struct {
		name              string
		echoCtx           func(e *echo.Echo) echo.Context
		userVerifyUsecase func() *mocks.UserVerifyUsecase
		wantErr           bool
	}{
		{
			name: "POST /api/v1/users-verify/email/send-code 정상적으로 호출",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   userVerifyApi + "/email/send-code",
				Method: http.MethodPost,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Body: `{"email":"gildong@email.com"}`,
				Claims: map[string]interface{}{
					"user":  "100",
					"email": "test@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_USER"},
				},
			}),
			userVerifyUsecase: func() *mocks.UserVerifyUsecase {
				usecase := &mocks.UserVerifyUsecase{}
				usecase.On("SendEmailCode", mock.Anything, udomain.UserSendEmailCodeRequest{
					JWTUser: jwthelper.User{
						Id:    100,
						Email: "test@email.com",
						Name:  "홍길동",
						Roles: map[string]int{"ROLE_USER": 1},
					},
					Email: "gildong@email.com",
				}).Return(nil)
				return usecase
			},
			wantErr: false,
		},
		{
			name: "POST /api/v1/users-verify/email/send-code 호출시 이메일 형식에 안맞게 요청해서 실패 ",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   userVerifyApi + "/email/send-code",
				Method: http.MethodPost,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Body: `{"email":"gildong"}`,
			}),
			userVerifyUsecase: func() *mocks.UserVerifyUsecase {
				usecase := &mocks.UserVerifyUsecase{}
				usecase.On("SendEmailCode", mock.Anything, udomain.UserSendEmailCodeRequest{
					Email: "gildong@email.com",
				}).Return(nil)
				return usecase
			},
			wantErr: true,
		},
		{
			name: "POST /api/v1/users/email/send-code 이메일 값이 없으면 실패",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   userVerifyApi + "/email/send-code",
				Method: http.MethodPost,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
			}),
			userVerifyUsecase: func() *mocks.UserVerifyUsecase {
				usecase := &mocks.UserVerifyUsecase{}
				usecase.On("SendEmailCode", mock.Anything, udomain.UserSendEmailCodeRequest{
					Email: "gildong@email.com",
				}).Return(nil)
				return usecase
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := userVerifyHttpDelivery{
				userVerifyUsecase: tt.userVerifyUsecase(),
				extractor:         jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.SendEmailCode(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "VerifyEmailRequest() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_userVerifyHttpDelivery_ResendEmailCode(t *testing.T) {
	tests := []struct {
		name              string
		echoCtx           func(e *echo.Echo) echo.Context
		userVerifyUsecase func() *mocks.UserVerifyUsecase
		wantErr           bool
	}{
		{
			name: "POST /api/v1/users/email/resend-code 정상적으로 호출",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   userVerifyApi + "/email/resend-code",
				Method: http.MethodPost,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Body: `{"email":"gildong@email.com"}`,
				Claims: map[string]interface{}{
					"user":  "100",
					"email": "test@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_USER"},
				},
			}),
			userVerifyUsecase: func() *mocks.UserVerifyUsecase {
				usecase := &mocks.UserVerifyUsecase{}
				usecase.On("ResendEmailCode", mock.Anything, udomain.UserResendEmailCodeRequest{
					JWTUser: jwthelper.User{
						Id:    100,
						Email: "test@email.com",
						Name:  "홍길동",
						Roles: map[string]int{"ROLE_USER": 1},
					},
					Email: "gildong@email.com",
				}).Return(nil)
				return usecase
			},
			wantErr: false,
		},
		{
			name: "POST /api/v1/users/email/resend-code 호출시 이메일 형식에 안맞게 요청해서 실패 ",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   userVerifyApi + "/email/resend-code",
				Method: http.MethodPost,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Body: `{"email":"gildong"}`,
			}),
			userVerifyUsecase: func() *mocks.UserVerifyUsecase {
				usecase := &mocks.UserVerifyUsecase{}
				usecase.On("ResendEmailCode", mock.Anything, udomain.UserResendEmailCodeRequest{
					Email: "gildong@email.com",
				}).Return(nil)
				return usecase
			},
			wantErr: true,
		},
		{
			name: "POST /api/v1/users/email/resend-code 이메일 값이 없으면 실패",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   userVerifyApi + "/email/resend-code",
				Method: http.MethodPost,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
			}),
			userVerifyUsecase: func() *mocks.UserVerifyUsecase {
				usecase := &mocks.UserVerifyUsecase{}
				usecase.On("ResendEmailCode", mock.Anything, udomain.UserResendEmailCodeRequest{
					Email: "gildong@email.com",
				}).Return(nil)
				return usecase
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := userVerifyHttpDelivery{
				userVerifyUsecase: tt.userVerifyUsecase(),
				extractor:         jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.ResendEmailCode(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "ReverifyEmailRequest() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_userVerifyHttpDelivery_CheckEmailCode(t *testing.T) {
	tests := []struct {
		name              string
		echoCtx           func(e *echo.Echo) echo.Context
		userVerifyUsecase func() *mocks.UserVerifyUsecase
		wantErr           bool
	}{
		{
			name: "POST /api/v1/users/email/check-code 정상적으로 호출",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   userVerifyApi + "/email/check-code",
				Method: http.MethodPost,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Body: `{"email":"gildong@email.com", "code":"123456"}`,
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "dool2@email.com",
					"name":  "둘리",
					"role":  []string{"ROLE_USER"},
				},
			}),

			userVerifyUsecase: func() *mocks.UserVerifyUsecase {
				usecase := &mocks.UserVerifyUsecase{}
				usecase.On("CheckEmailCode", mock.Anything, udomain.UserCheckEmailCodeRequest{
					JWTUser: jwthelper.User{
						Id:    1000,
						Email: "dool2@email.com",
						Name:  "둘리",
						Roles: map[string]int{"ROLE_USER": 1},
					},
					Email: "gildong@email.com",
					Code:  "123456",
				}).Return(nil)

				return usecase
			},
			wantErr: false,
		},
		{
			name: "POST /api/v1/users/email/check-code 호출시 이메일 형식에 안맞아서 실패",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   userVerifyApi + "/email/check-code",
				Method: http.MethodPost,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Body: `{"email":"gildong", "code":"123456"}`,
			}),

			userVerifyUsecase: func() *mocks.UserVerifyUsecase {
				usecase := &mocks.UserVerifyUsecase{}
				return usecase
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := userVerifyHttpDelivery{
				userVerifyUsecase: tt.userVerifyUsecase(),
				extractor:         jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.CheckEmailCode(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "CheckEmailCode() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}
