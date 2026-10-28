package http

import (
	"net/http"
	"testing"

	"github.com/samber/lo"

	adomain "github.com/illuminarean/energy-share-core-api/domain"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/validator"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	pkgtest "github.com/illuminarean/energy-share-core-api/delivery/testing"
	"github.com/illuminarean/energy-share-core-api/domain/mocks"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

func Test_mailHttpDelivery_GetAll(t *testing.T) {
	tests := []struct {
		name        string
		echoCtx     func(e *echo.Echo) echo.Context
		mailUsecase func() *mocks.MailUsecase
		wantErr     bool
	}{
		{
			name: "GET api/v1/mails 호출성공",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path: "/api/v1/mails",
				Claims: map[string]interface{}{
					"user":  "12378123",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_ADMIN", "ROLE_OPERATOR"},
				},
				Method: http.MethodGet,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
			}),
			mailUsecase: func() *mocks.MailUsecase {
				usecase := &mocks.MailUsecase{}
				usecase.On("GetAll", mock.Anything, adomain.MailGetAllRequest{
					JWTUser: jwthelper.User{
						Id:    12378123,
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{"ROLE_ADMIN": 1, "ROLE_OPERATOR": 1},
					},
					ClientIp: "192.0.2.1",
					SearchInput: cdomain.SearchInput{
						MaxResultCount: 30,
					},
				}).Return(adomain.MailGetAllResponse{}, nil)

				return usecase
			},
		},
		{
			name: "GET api/v1/mails 호출성공 - 쿼리: Status SENT, MaxResultCount 10",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path: "/api/v1/mails",
				Claims: map[string]interface{}{
					"user":  "12378123",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_ADMIN", "ROLE_OPERATOR"},
				},
				QueryParams: map[string][]string{
					"status":           {"SENT"},
					"max_result_count": {"10"},
				},
				Method: http.MethodGet,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
			}),
			mailUsecase: func() *mocks.MailUsecase {
				usecase := &mocks.MailUsecase{}
				usecase.On("GetAll", mock.Anything, adomain.MailGetAllRequest{
					JWTUser: jwthelper.User{
						Id:    12378123,
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{"ROLE_ADMIN": 1, "ROLE_OPERATOR": 1},
					},
					ClientIp: "192.0.2.1",
					SearchInput: cdomain.SearchInput{
						MaxResultCount: 10,
					},
					MailGetAllQueryParam: adomain.MailGetAllQueryParam{
						Status: lo.ToPtr(string(domain.MAIL_SENT)),
					},
				}).Return(adomain.MailGetAllResponse{}, nil)

				return usecase
			},
		},
		{
			name: "GET api/v1/mails 호출실패 - 쿼리: Status는 SENT FAILED PENDING 중 하나여야 함",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path: "/api/v1/mails",
				Claims: map[string]interface{}{
					"user":  "12378123",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_ADMIN", "ROLE_OPERATOR"},
				},
				QueryParams: map[string][]string{
					"status":           {"SUCCESS"},
					"max_result_count": {"10"},
				},
				Method: http.MethodGet,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
			}),
			mailUsecase: func() *mocks.MailUsecase {
				usecase := &mocks.MailUsecase{}
				usecase.On("GetAll", mock.Anything, adomain.MailGetAllRequest{
					JWTUser: jwthelper.User{
						Id:    12378123,
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{"ROLE_ADMIN": 1, "ROLE_OPERATOR": 1},
					},
					ClientIp: "192.0.2.1",
					SearchInput: cdomain.SearchInput{
						MaxResultCount: 10,
					},
					MailGetAllQueryParam: adomain.MailGetAllQueryParam{
						Status: lo.ToPtr("SUCCESS"),
					},
				}).Return(adomain.MailGetAllResponse{}, nil)

				return usecase
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			delivery := mailHttpDelivery{
				cfg:         &config.Config{},
				logger:      factory.NewLoggerFactory(),
				extractor:   jwthelper.NewUserExtractor(),
				mailUsecase: tt.mailUsecase(),
			}
			e := echo.New()
			e.Validator = validator.NewCustomValidator()
			err := delivery.GetAll(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "GetAll() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_mailHttpDelivery_GetOne(t *testing.T) {
	tests := []struct {
		name        string
		echoCtx     func(e *echo.Echo) echo.Context
		mailUsecase func() *mocks.MailUsecase
		wantErr     bool
	}{
		{
			name: "GET api/v1/mails/:id 호출성공",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path: "/api/v1/mails/:id",
				Claims: map[string]interface{}{
					"user":  "12378123",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_ADMIN", "ROLE_OPERATOR"},
				},
				Params: map[string]string{"id": "1000"},
				Method: http.MethodGet,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
			}),
			mailUsecase: func() *mocks.MailUsecase {
				usecase := &mocks.MailUsecase{}
				usecase.On("GetOne", mock.Anything, adomain.MailGetOneRequest{
					JWTUser: jwthelper.User{
						Id:    12378123,
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{"ROLE_ADMIN": 1, "ROLE_OPERATOR": 1},
					},
					ClientIp: "192.0.2.1",
					MailId:   "1000",
				}).Return(adomain.MailDetailResponse{}, nil)

				return usecase
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			delivery := mailHttpDelivery{
				cfg:         &config.Config{},
				logger:      factory.NewLoggerFactory(),
				extractor:   jwthelper.NewUserExtractor(),
				mailUsecase: tt.mailUsecase(),
			}
			e := echo.New()
			e.Validator = validator.NewCustomValidator()
			err := delivery.GetOne(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "GetAll() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_mailHttpDelivery_GetAllFailed(t *testing.T) {
	tests := []struct {
		name        string
		echoCtx     func(e *echo.Echo) echo.Context
		mailUsecase func() *mocks.MailUsecase
		wantErr     bool
	}{
		{
			name: "GET api/v1/mails/failed 호출성공",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path: "/api/v1/mails/failed",
				Claims: map[string]interface{}{
					"user":  "12378123",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_ADMIN", "ROLE_OPERATOR"},
				},
				Method: http.MethodGet,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
			}),
			mailUsecase: func() *mocks.MailUsecase {
				usecase := &mocks.MailUsecase{}
				usecase.On("GetAllFailed", mock.Anything, adomain.MailGetAllFailedRequest{
					JWTUser: jwthelper.User{
						Id:    12378123,
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{"ROLE_ADMIN": 1, "ROLE_OPERATOR": 1},
					},
					ClientIp: "192.0.2.1",
					SearchInput: cdomain.SearchInput{
						MaxResultCount: 30,
					},
				}).Return(adomain.MailGetAllFailedResponse{}, nil)

				return usecase
			},
		},
		{
			name: "GET api/v1/mails/failed 호출성공 - 쿼리: NORMAL_MAIL",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path: "/api/v1/mails/failed",
				Claims: map[string]interface{}{
					"user":  "12378123",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_ADMIN", "ROLE_OPERATOR"},
				},
				QueryParams: map[string][]string{
					"template_name":    {"NORMAL_MAIL"},
					"max_result_count": {"10"},
				},
				Method: http.MethodGet,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
			}),
			mailUsecase: func() *mocks.MailUsecase {
				usecase := &mocks.MailUsecase{}
				usecase.On("GetAllFailed", mock.Anything, adomain.MailGetAllFailedRequest{
					JWTUser: jwthelper.User{
						Id:    12378123,
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{"ROLE_ADMIN": 1, "ROLE_OPERATOR": 1},
					},
					ClientIp: "192.0.2.1",
					SearchInput: cdomain.SearchInput{
						MaxResultCount: 10,
					},
					MailGetAllFailedQueryParam: adomain.MailGetAllFailedQueryParam{
						TemplateName: lo.ToPtr("NORMAL_MAIL"),
					},
				}).Return(adomain.MailGetAllFailedResponse{}, nil)

				return usecase
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			delivery := mailHttpDelivery{
				cfg:         &config.Config{},
				logger:      factory.NewLoggerFactory(),
				extractor:   jwthelper.NewUserExtractor(),
				mailUsecase: tt.mailUsecase(),
			}
			e := echo.New()
			e.Validator = validator.NewCustomValidator()
			err := delivery.GetAllFailed(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "GetAllFailed() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_mailHttpDelivery_GetTemplate(t *testing.T) {
	tests := []struct {
		name        string
		echoCtx     func(e *echo.Echo) echo.Context
		mailUsecase func() *mocks.MailUsecase
		wantErr     bool
	}{
		{
			name: "GET api/v1/mails/templates/:template_type 호출성공. NORMAL_MAIL",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path: "/api/v1/mails/templates/:template_type",
				Claims: map[string]interface{}{
					"user":  "12378123",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_ADMIN", "ROLE_OPERATOR"},
				},
				Params: map[string]string{"template_type": string(domain.NORMAL_MAIL_TEMPLATE)},
				Method: http.MethodGet,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
			}),
			mailUsecase: func() *mocks.MailUsecase {
				usecase := &mocks.MailUsecase{}
				usecase.On("GetTemplate", mock.Anything, adomain.MailGetTemplateRequest{
					JWTUser: jwthelper.User{
						Id:    12378123,
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{"ROLE_ADMIN": 1, "ROLE_OPERATOR": 1},
					},
					ClientIp:     "192.0.2.1",
					TemplateType: string(domain.NORMAL_MAIL_TEMPLATE),
				}).Return(adomain.MailGetTemplateResponse{}, nil)
				return usecase
			},
		},
		{
			name: "GET api/v1/mails/templates/:template_type 호출성공 - default 는 NORMAL_MAIL",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path: "/api/v1/mails/templates/:template_type",
				Claims: map[string]interface{}{
					"user":  "12378123",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_ADMIN", "ROLE_OPERATOR"},
				},
				Params: map[string]string{"template_type": "default"},
				Method: http.MethodGet,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
			}),
			mailUsecase: func() *mocks.MailUsecase {
				usecase := &mocks.MailUsecase{}
				usecase.On("GetTemplate", mock.Anything, adomain.MailGetTemplateRequest{
					JWTUser: jwthelper.User{
						Id:    12378123,
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{"ROLE_ADMIN": 1, "ROLE_OPERATOR": 1},
					},
					ClientIp:     "192.0.2.1",
					TemplateType: string(domain.NORMAL_MAIL_TEMPLATE),
				}).Return(adomain.MailGetTemplateResponse{}, nil)
				return usecase
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			delivery := mailHttpDelivery{
				cfg:         &config.Config{},
				logger:      factory.NewLoggerFactory(),
				extractor:   jwthelper.NewUserExtractor(),
				mailUsecase: tt.mailUsecase(),
			}
			e := echo.New()
			e.Validator = validator.NewCustomValidator()
			err := delivery.GetTemplate(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "GetTemplate() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}
