package http

import (
	"net/http"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	pkgtest "github.com/illuminarean/energy-share-core-api/delivery/testing"
	adomain "github.com/illuminarean/energy-share-core-api/domain"
	"github.com/illuminarean/energy-share-core-api/domain/mocks"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/validator"
)

func Test_NoticeMsgHttpDelivery_GetAll(t *testing.T) {
	tests := []struct {
		name             string
		echoCtx          func(e *echo.Echo) echo.Context
		noticeMsgUsecase func() *mocks.NoticeMsgUsecase
		wantErr          bool
	}{
		{
			name: "get all notice info list - success",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/notice-msgs",
				Method: http.MethodGet,
				Claims: map[string]interface{}{
					"user":  "12378123",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_INVESTOR"},
				},
				QueryParams: map[string][]string{
					"sortby":           {"title"},
					"order":            {"asc"},
					"skip_count":       {"0"},
					"max_result_count": {"10"},
				},
			}),
			noticeMsgUsecase: func() *mocks.NoticeMsgUsecase {
				usecase := new(mocks.NoticeMsgUsecase)
				usecase.On("GetAll", mock.Anything, adomain.NoticeMsgGetAllRequest{
					JWTUser: jwthelper.User{
						Id:    12378123,
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{"ROLE_INVESTOR": 1},
					},
					ClientIp: "192.0.2.1",
					NoticeMsgGetAllQueryParam: adomain.NoticeMsgGetAllQueryParam{
						SearchInput: cdomain.SearchInput{
							Sortby:         "title",
							Order:          "asc",
							SkipCount:      0,
							MaxResultCount: 10,
						},
					},
				}).Return(adomain.NoticeMsgGetAllResponse{
					TotalCount: 1,
					Items:      []domain.NoticeMsg{mockNoticeMsg1},
				}, nil)
				return usecase
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := noticeMsgHttpDelivery{
				noticeMsgUsecase: tt.noticeMsgUsecase(),
				extractor:        jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.GetAll(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "GetAll() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_NoticeMsgHttpDelivery_GetOne(t *testing.T) {
	tests := []struct {
		name             string
		echoCtx          func(e *echo.Echo) echo.Context
		noticeMsgUsecase func() *mocks.NoticeMsgUsecase
		wantErr          bool
	}{
		{
			name: "get one notice info - success",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/notice-msgs/:id",
				Method: http.MethodGet,
				Params: map[string]string{
					"id": "1",
				},
				Claims: map[string]interface{}{
					"user":  "12378123",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_INVESTOR"},
				},
			}),
			noticeMsgUsecase: func() *mocks.NoticeMsgUsecase {
				usecase := new(mocks.NoticeMsgUsecase)
				usecase.On("GetOne", mock.Anything, adomain.NoticeMsgGetOneRequest{
					Id: 1,
					JWTUser: jwthelper.User{
						Id:    12378123,
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{"ROLE_INVESTOR": 1},
					},
					ClientIp: "192.0.2.1",
				}).Return(adomain.NoticeMsgGetOneResponse{NoticeMsg: mockNoticeMsg1}, nil)
				return usecase
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := noticeMsgHttpDelivery{
				noticeMsgUsecase: tt.noticeMsgUsecase(),
				extractor:        jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.GetOne(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "GetOne() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_NoticeMsgHttpDelivery_Create(t *testing.T) {
	tests := []struct {
		name             string
		echoCtx          func(e *echo.Echo) echo.Context
		noticeMsgUsecase func() *mocks.NoticeMsgUsecase
		wantErr          bool
	}{
		{
			name: "create notice info - success",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/notice-msgs/send",
				Method: http.MethodPost,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Body: `{
					"user_ids" : [111],
					"title" : "example title",	
					"content_html" : "<body><h1>example content</h1></body>"
				}`,
				Claims: map[string]interface{}{
					"user":  "12378123",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_INVESTOR"},
				},
			}),
			noticeMsgUsecase: func() *mocks.NoticeMsgUsecase {
				usecase := new(mocks.NoticeMsgUsecase)
				usecase.On("Create", mock.Anything, adomain.NoticeMsgCreateRequest{
					UserIds: []int64{111},
					JWTUser: jwthelper.User{
						Id:    12378123,
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{"ROLE_INVESTOR": 1},
					},
					ClientIp: "192.0.2.1",
					NoticeMsgCreateRequestBase: adomain.NoticeMsgCreateRequestBase{
						Title:       "example title",
						ContentHtml: "<body><h1>example content</h1></body>",
					},
				}).Return(nil)
				return usecase
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := noticeMsgHttpDelivery{
				noticeMsgUsecase: tt.noticeMsgUsecase(),
				extractor:        jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.Create(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "Create() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_NoticeMsgHttpDelivery_CreateByProject(t *testing.T) {
	tests := []struct {
		name             string
		echoCtx          func(e *echo.Echo) echo.Context
		noticeMsgUsecase func() *mocks.NoticeMsgUsecase
		wantErr          bool
	}{
		{
			name: "create notice info by project - success",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/notice-msgs/send/projects",
				Method: http.MethodPost,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Body: `{
					"project_id" : 1,
					"title" : "example title",	
					"content_html" : "<body><h1>example content</h1></body>"
				}`,
				Claims: map[string]interface{}{
					"user":  "12378123",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_INVESTOR"},
				},
			}),
			noticeMsgUsecase: func() *mocks.NoticeMsgUsecase {
				usecase := new(mocks.NoticeMsgUsecase)
				usecase.On("CreateByProject", mock.Anything, adomain.NoticeMsgCreateByProjectRequest{
					ProjectId: 1,
					JWTUser: jwthelper.User{
						Id:    12378123,
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{"ROLE_INVESTOR": 1},
					},
					ClientIp: "192.0.2.1",
					NoticeMsgCreateRequestBase: adomain.NoticeMsgCreateRequestBase{
						Title:       "example title",
						ContentHtml: "<body><h1>example content</h1></body>",
					},
				}).Return(nil)
				return usecase
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := noticeMsgHttpDelivery{
				noticeMsgUsecase: tt.noticeMsgUsecase(),
				extractor:        jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.CreateByProject(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "CreateByProject() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_NoticeMsgHttpDelivery_CreateByNewsletter(t *testing.T) {
	tests := []struct {
		name             string
		echoCtx          func(e *echo.Echo) echo.Context
		noticeMsgUsecase func() *mocks.NoticeMsgUsecase
		wantErr          bool
	}{
		{
			name: "create notice info by newsletter - success",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/notice-msgs/send/newsletters",
				Method: http.MethodPost,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Body: `{
					"title" : "example title",	
					"content_html" : "<body><h1>example content</h1></body>"
				}`,
				Claims: map[string]interface{}{
					"user":  "12378123",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_INVESTOR"},
				},
			}),
			noticeMsgUsecase: func() *mocks.NoticeMsgUsecase {
				usecase := new(mocks.NoticeMsgUsecase)
				usecase.On("CreateByNewsletter", mock.Anything, adomain.NoticeMsgCreateByNewsletterRequest{
					JWTUser: jwthelper.User{
						Id:    12378123,
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{"ROLE_INVESTOR": 1},
					},
					ClientIp: "192.0.2.1",
					NoticeMsgCreateRequestBase: adomain.NoticeMsgCreateRequestBase{
						Title:       "example title",
						ContentHtml: "<body><h1>example content</h1></body>",
					},
				}).Return(nil)
				return usecase
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := noticeMsgHttpDelivery{
				noticeMsgUsecase: tt.noticeMsgUsecase(),
				extractor:        jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.CreateByNewsletter(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "CreateByNewsletter() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_NoticeMsgHttpDelivery_CreateByUserType(t *testing.T) {
	tests := []struct {
		name             string
		echoCtx          func(e *echo.Echo) echo.Context
		noticeMsgUsecase func() *mocks.NoticeMsgUsecase
		wantErr          bool
	}{
		{
			name: "create notice info by user type - success",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/notice-msgs/send/user-type",
				Method: http.MethodPost,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Body: `{
					"user_types" : ["PRIVATE_INVESTOR"],
					"title" : "example title",
					"content_html" : "<body><h1>example content</h1></body>"
				}`,
				Claims: map[string]interface{}{
					"user":  "12378123",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_INVESTOR"},
				},
			}),
			noticeMsgUsecase: func() *mocks.NoticeMsgUsecase {
				usecase := new(mocks.NoticeMsgUsecase)
				usecase.On("CreateByUserType", mock.Anything, adomain.NoticeMsgCreateByUserTypeRequest{
					UserTypes: []string{"PRIVATE_INVESTOR"},
					JWTUser: jwthelper.User{
						Id:    12378123,
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{"ROLE_INVESTOR": 1},
					},
					ClientIp: "192.0.2.1",
					NoticeMsgCreateRequestBase: adomain.NoticeMsgCreateRequestBase{
						Title:       "example title",
						ContentHtml: "<body><h1>example content</h1></body>",
					},
				}).Return(nil)
				return usecase
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := noticeMsgHttpDelivery{
				noticeMsgUsecase: tt.noticeMsgUsecase(),
				extractor:        jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.CreateByUserType(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "CreateByUserType() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}
