package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	udomain "github.com/illuminarean/energyshares-user-api/domain"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	pkgtest "github.com/illuminarean/energy-share-core-api/delivery/testing"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/validator"
	"github.com/illuminarean/energyshares-user-api/domain/mocks"
)

const userApi = "/api/v1/users"

// 204は本文を持てない。本文を書くとnet/httpがErrBodyNotAllowedを返し、ハンドラが
// エラーを返す(本番で会員登録・メール認証のたびに記録されていた、MIR-1392)。
func assertNoContent(t *testing.T, c echo.Context) {
	t.Helper()
	assert.Equal(t, http.StatusNoContent, c.Response().Status)
	assert.Empty(t, c.Response().Writer.(*httptest.ResponseRecorder).Body.String())
}

func Test_userHttpDelivery_Login(t *testing.T) {
	tests := []struct {
		name        string
		echoCtx     func(e *echo.Echo) echo.Context
		userUsecase func() *mocks.UserUsecase
		wantErr     bool
	}{
		{
			name: "POST /api/v1/users/login 정상적으로 호출",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   userApi + "/login",
				Method: http.MethodPost,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Body: `{"email" : "test@example.com", "password": "1234"}`,
			}),
			userUsecase: func() *mocks.UserUsecase {
				usecase := &mocks.UserUsecase{}
				usecase.On("Login", mock.Anything, udomain.UserLoginRequest{
					Email:     "test@example.com",
					Password:  "1234",
					ClientIp:  "192.0.2.1",
					UserAgent: "",
				}).Return(udomain.UserLoginResponse{AccessToken: "access-token", RefreshToken: "refresh-token"}, nil)
				return usecase
			},
			wantErr: false,
		},
		{
			name: "POST /api/v1/users/login 오류 - 잠긴 계정으로 로그인 시도",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   userApi + "/login",
				Method: http.MethodPost,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Body: `{"email" : "test@example.com", "password": "1234"}`,
			}),
			userUsecase: func() *mocks.UserUsecase {
				usecase := &mocks.UserUsecase{}
				usecase.On("Login", mock.Anything, udomain.UserLoginRequest{
					Email:     "test@example.com",
					Password:  "1234",
					ClientIp:  "192.0.2.1",
					UserAgent: "",
				}).Return(udomain.UserLoginResponse{}, cdomain.ErrResourceLocked)
				return usecase
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := userHttpDelivery{
				userUsecase: tt.userUsecase(),
				extractor:   jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.Login(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "Login() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_userHttpDelivery_CreatePreSignedUrl(t *testing.T) {
	tests := []struct {
		name        string
		echoCtx     func(e *echo.Echo) echo.Context
		userUsecase func() *mocks.UserUsecase
		wantErr     bool
	}{
		{
			name: "POST /api/v1/users/pre-signed-url 정상적으로 호출",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   userApi + "/pre-signed-url",
				Method: http.MethodPost,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Body: `{"file_name" : "test.png"}`,
			}),
			userUsecase: func() *mocks.UserUsecase {
				usecase := &mocks.UserUsecase{}
				usecase.On("CreatePreSignedUrl", mock.Anything, udomain.UserCreatePreSignedUrlRequest{
					FileName: "test.png",
				}).Return(udomain.UserCreatePreSignedUrlResponse{}, nil)
				return usecase
			},
			wantErr: false,
		},
		{
			name: "POST /api/v1/users/pre-signed-url body 값을 넣지 않았을 때",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   userApi + "/pre-signed-url",
				Method: http.MethodPost,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
			}),
			userUsecase: func() *mocks.UserUsecase {
				usecase := &mocks.UserUsecase{}
				return usecase
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := userHttpDelivery{
				userUsecase: tt.userUsecase(),
				extractor:   jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.CreatePreSignedUrl(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "CreatePreSignedUrl() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_userHttpDelivery_RequestSignUp(t *testing.T) {
	tests := []struct {
		name        string
		echoCtx     func(e *echo.Echo) echo.Context
		userUsecase func() *mocks.UserUsecase
		wantErr     bool
	}{
		{
			name: "POST /api/v1/users/signup-requests 정상적으로 호출",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   userApi + "/signup-requests",
				Method: http.MethodPost,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Body: `{"email" : "gildong@email.com", "user_type":"PRIVATE", "marketing_code":"1234", "marketing_campaign":"test"}`,
			}),
			userUsecase: func() *mocks.UserUsecase {
				usecase := &mocks.UserUsecase{}
				usecase.On("RequestSignUp", mock.Anything, udomain.UserRequestSignUpRequest{
					Email:             "gildong@email.com",
					UserType:          "PRIVATE",
					MarketingCode:     "1234",
					MarketingCampaign: "test",
				}).Return(udomain.UserRequestSignUpResponse{}, nil)

				return usecase
			},
			wantErr: false,
		},
		{
			name: "POST /api/v1/users/signup-requests 호출시 필수 값을 누락해서 에러발생 (user_type)",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   userApi + "/signup-requests",
				Method: http.MethodPost,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Body: `{"email" : "gildong@email.com", "user_type":""}`,
			}),
			userUsecase: func() *mocks.UserUsecase {
				usecase := &mocks.UserUsecase{}
				return usecase
			},
			wantErr: true,
		},
		{
			name: "POST /api/v1/users/signup-requests 호출시 교차값중 하나가 누락되어 에러발생 (marketing_code, marketing_campaign)",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   userApi + "/signup-requests",
				Method: http.MethodPost,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Body: `{"email" : "gildong@email.com", "user_type":"PRIVATE", "marketing_code":"1234"}`,
			}),
			userUsecase: func() *mocks.UserUsecase {
				usecase := &mocks.UserUsecase{}
				return usecase
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := userHttpDelivery{
				userUsecase: tt.userUsecase(),
				extractor:   jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.RequestSignUp(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "RequestSignUp() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_userHttpDelivery_VerifyEmail(t *testing.T) {
	tests := []struct {
		name        string
		echoCtx     func(e *echo.Echo) echo.Context
		userUsecase func() *mocks.UserUsecase
		wantErr     bool
	}{
		{
			name: "POST /api/v1/users/signup-requests/:request_id/verify 정상적으로 호출",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   userApi + "/signup-requests/:request_id/verify",
				Method: http.MethodPost,
				Params: map[string]string{
					"request_id": "1234",
				},
			}),
			userUsecase: func() *mocks.UserUsecase {
				usecase := &mocks.UserUsecase{}
				usecase.On("VerifyEmail", mock.Anything, udomain.UserVerifyEmailRequest{
					RequestId: "1234",
				}).Return(nil)

				return usecase
			},
			wantErr: false,
		},
		{
			name: "POST /api/v1/users/signup-requests/:request_id/verify 호출시 필수 값을 누락해서 에러발생 (request_id)",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   userApi + "/signup-requests/:request_id/verify",
				Method: http.MethodPost,
			}),
			userUsecase: func() *mocks.UserUsecase {
				usecase := &mocks.UserUsecase{}
				return usecase
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := userHttpDelivery{
				userUsecase: tt.userUsecase(),
				extractor:   jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			c := tt.echoCtx(e)
			err := delivery.VerifyEmail(c)
			assert.Equalf(t, tt.wantErr, err != nil, "Verify() error = %v, wantErr %v", err, tt.wantErr)
			if !tt.wantErr {
				assertNoContent(t, c)
			}
		})
	}
}

func Test_userHttpDelivery_CheckSignupRequestValidity(t *testing.T) {
	tests := []struct {
		name        string
		echoCtx     func(e *echo.Echo) echo.Context
		userUsecase func() *mocks.UserUsecase
		wantErr     bool
	}{
		{
			name: "GET /api/v1/users/signup-requests/:request_id 정상적으로 호출",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   userApi + "/signup-requests/:request_id",
				Method: http.MethodGet,
				Params: map[string]string{
					"request_id": "1234",
				},
			}),
			userUsecase: func() *mocks.UserUsecase {
				usecase := &mocks.UserUsecase{}
				usecase.On("CheckSignupRequestValidity", mock.Anything, udomain.UserCheckSignupRequestValidityRequest{
					RequestId: "1234",
				}).Return(udomain.UserCheckSignupRequestValidityResponse{Email: "test@example.com", UserType: domain.PRIVATE}, nil)

				return usecase
			},
			wantErr: false,
		},
		{
			name: "GET /api/v1/users/signup-requests/:request_id 호출시 필수 값을 누락해서 에러발생 (request_id)",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   userApi + "/signup-requests/:request_id",
				Method: http.MethodGet,
			}),
			userUsecase: func() *mocks.UserUsecase {
				usecase := &mocks.UserUsecase{}
				return usecase
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := userHttpDelivery{
				userUsecase: tt.userUsecase(),
				extractor:   jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.CheckSignupRequestValidity(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "CheckSignupRequestValidity() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_userHttpDelivery_SingUp(t *testing.T) {
	tests := []struct {
		name        string
		echoCtx     func(e *echo.Echo) echo.Context
		userUsecase func() *mocks.UserUsecase
		wantErr     bool
	}{
		{
			name: "POST /api/v1/users/signup 정상적으로 호출",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   userApi + "/signup",
				Method: http.MethodPost,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Body: `{
					   "request_id":"1234",
					   "user_type":"PRIVATE",
					   "email":"gildong@email.com",
					   "password":"1234",
					   "is_news_letter":true,
					   "first_name":"길동",
					   "family_name":"홍",
					   "first_name_kana":"ギルドン",
					   "family_name_kana":"ホン"
					}`,
			}),
			userUsecase: func() *mocks.UserUsecase {
				usecase := &mocks.UserUsecase{}
				usecase.On("SignUp", mock.Anything, udomain.UserSignUpRequest{
					RequestId:      "1234",
					UserType:       "PRIVATE",
					Email:          "gildong@email.com",
					Password:       "1234",
					IsNewsLetter:   true,
					FirstName:      "길동",
					FamilyName:     "홍",
					FirstNameKana:  "ギルドン",
					FamilyNameKana: "ホン",
				}).Return(nil)
				return usecase
			},
			wantErr: false,
		},
		{
			name: "POST /api/v1/users/signup 호출시 필수 값을 누락해서 에러발생 (request_id)",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   userApi + "/signup",
				Method: http.MethodPost,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Body: `{
					   "user_type":"PRIVATE",
					   "email":"gildong@email.com",
					   "password":"1234",
					   "is_news_letter":true,
					   "first_name":"길동",
					   "family_name":"홍",
					   "first_name_kana":"ギルドン",
					   "family_name_kana":"ホン"
					}`,
			}),
			userUsecase: func() *mocks.UserUsecase {
				usecase := &mocks.UserUsecase{}
				return usecase
			},
			wantErr: true,
		},
		{
			name: "POST /api/v1/users/signup 호출시 필수 값을 누락해서 에러발생 (family_name)",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   userApi + "/signup",
				Method: http.MethodPost,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Body: `{
					   "request_id":"1234",	
					   "user_type":"PRIVATE",
					   "email":"gildong@email.com",
					   "password":"1234",
					   "is_news_letter":true,
					   "first_name":"길동",
					   "first_name_kana":"ギルドン",
					   "family_name_kana":"ホン"
					}`,
			}),
			userUsecase: func() *mocks.UserUsecase {
				usecase := &mocks.UserUsecase{}
				return usecase
			},
			wantErr: true,
		},
		{
			name: "POST /api/v1/users/signup 법인으로 가입",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   userApi + "/signup",
				Method: http.MethodPost,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Body: `{
					   "request_id":"1234",	
					   "user_type":"CORPORATE",
					   "email":"gildong@email.com",
					   "password":"1234",
					   "is_news_letter":true,
                       "corp_name": "illuminarean",
					   "manager_first_name": "길동",
					   "manager_family_name": "홍",
					   "manager_first_name_kana": "ギルドン",
					   "manager_family_name_kana": "ホン"
					}`,
			}),
			userUsecase: func() *mocks.UserUsecase {
				usecase := &mocks.UserUsecase{}
				usecase.On("SignUp", mock.Anything, udomain.UserSignUpRequest{
					RequestId:             "1234",
					UserType:              "CORPORATE",
					Email:                 "gildong@email.com",
					Password:              "1234",
					IsNewsLetter:          true,
					CorpName:              "illuminarean",
					ManagerFirstName:      "길동",
					ManagerFamilyName:     "홍",
					ManagerFirstNameKana:  "ギルドン",
					ManagerFamilyNameKana: "ホン",
				}).Return(nil)
				return usecase
			},
			wantErr: false,
		},
		{
			name: "POST /api/v1/users/signup 법인으로 가입했지만 회사명이 비어서 에러발생",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   userApi + "/signup",
				Method: http.MethodPost,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Body: `{
					   "request_id":"1234",	
					   "user_type":"CORPORATE",
					   "email":"gildong@email.com",
					   "password":"1234",
					   "is_news_letter":true,
					}`,
			}),
			userUsecase: func() *mocks.UserUsecase {
				usecase := &mocks.UserUsecase{}
				return usecase
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := userHttpDelivery{
				userUsecase: tt.userUsecase(),
				extractor:   jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			c := tt.echoCtx(e)
			err := delivery.SignUp(c)
			assert.Equalf(t, tt.wantErr, err != nil, "SignUp() error = %v, wantErr %v", err, tt.wantErr)
			if !tt.wantErr {
				assertNoContent(t, c)
			}
		})
	}
}
