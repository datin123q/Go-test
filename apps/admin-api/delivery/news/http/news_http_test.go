package http

import (
	"net/http"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	pkgtest "github.com/illuminarean/energy-share-core-api/delivery/testing"
	adomain "github.com/illuminarean/energy-share-core-api/domain"
	"github.com/illuminarean/energy-share-core-api/domain/mocks"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/validator"
)

func Test_NewsHttpDelivery_GetAll(t *testing.T) {
	tests := []struct {
		name        string
		echoCtx     func(e *echo.Echo) echo.Context
		newsUsecase func() *mocks.NewsUsecase
		wantErr     bool
	}{
		{
			name: "GET /v1/news",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/news",
				Method: http.MethodGet,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Claims: map[string]interface{}{
					"user":  "12378123",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_ADMIN", "ROLE_OPERATOR"},
				},
				QueryParams: map[string][]string{
					"max_result_count": {"10"},
					"order":            {"asc"},
					"skip_count":       {"0"},
					"sortby":           {"title"},
				},
			}),
			newsUsecase: func() *mocks.NewsUsecase {
				m := &mocks.NewsUsecase{}
				m.On("GetAll", mock.Anything, adomain.NewsGetAllRequest{
					JWTUser: jwthelper.User{
						Id:    12378123,
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{"ROLE_ADMIN": 1, "ROLE_OPERATOR": 1},
					},
					ClientIp: "192.0.2.1",
					NewsGetAllQueryParam: adomain.NewsGetAllQueryParam{
						SearchInput: cdomain.SearchInput{
							Sortby:         "title",
							Order:          "asc",
							SkipCount:      0,
							MaxResultCount: 10,
						},
					},
				}).Return(adomain.NewsGetAllResponse{
					TotalCount: 1,
					Items: []domain.News{
						mockNews1,
					},
				}, nil)
				return m
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := newsHttpDelivery{
				extractor:   jwthelper.NewUserExtractor(),
				newsUsecase: tt.newsUsecase(),
			}
			e := echo.New()
			e.Validator = validator.NewCustomValidator()
			err := delivery.GetAll(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "GetAll() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_NewsHttpDelivery_GetOne(t *testing.T) {
	tests := []struct {
		name        string
		echoCtx     func(e *echo.Echo) echo.Context
		newsUsecase func() *mocks.NewsUsecase
		wantErr     bool
	}{
		{
			name: "GET /v1/news/:id",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/v1/news/:id",
				Method: http.MethodGet,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Claims: map[string]interface{}{
					"user":  "12378123",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_ADMIN", "ROLE_OPERATOR"},
				},
				Params: map[string]string{"id": "1000"},
			},
			),
			newsUsecase: func() *mocks.NewsUsecase {
				m := &mocks.NewsUsecase{}
				m.On("GetOne", mock.Anything, adomain.NewsGetOneRequest{
					JWTUser: jwthelper.User{
						Id:    12378123,
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{"ROLE_ADMIN": 1, "ROLE_OPERATOR": 1},
					},
					ClientIp: "192.0.2.1",
					NewsId:   int64(1000),
				}).Return(adomain.NewsGetOneResponse{
					News: mockNews1,
				}, nil)
				return m
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := newsHttpDelivery{
				extractor:   jwthelper.NewUserExtractor(),
				newsUsecase: tt.newsUsecase(),
			}
			e := echo.New()
			e.Validator = validator.NewCustomValidator()
			if err := delivery.GetOne(tt.echoCtx(e)); (err != nil) != tt.wantErr {
				t.Errorf("GetOne() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_NewsHttpDelivery_Delete(t *testing.T) {
	tests := []struct {
		name        string
		echoCtx     func(e *echo.Echo) echo.Context
		newsUsecase func() *mocks.NewsUsecase
		wantErr     bool
	}{
		{
			name: "Delete /v1/news/:id",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/v1/news/:id",
				Method: http.MethodGet,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Claims: map[string]interface{}{
					"user":  "12378123",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_ADMIN", "ROLE_OPERATOR"},
				},
				Params: map[string]string{"id": "1000"},
			},
			),
			newsUsecase: func() *mocks.NewsUsecase {
				m := &mocks.NewsUsecase{}
				m.On("Delete", mock.Anything, adomain.NewsDeleteRequest{
					JWTUser: jwthelper.User{
						Id:    12378123,
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{"ROLE_ADMIN": 1, "ROLE_OPERATOR": 1},
					},
					ClientIp: "192.0.2.1",
					NewsId:   int64(1000),
				}).Return(nil)
				return m
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := newsHttpDelivery{
				extractor:   jwthelper.NewUserExtractor(),
				newsUsecase: tt.newsUsecase(),
			}
			e := echo.New()
			e.Validator = validator.NewCustomValidator()
			if err := delivery.Delete(tt.echoCtx(e)); (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_newsHttpDelivery_Create(t *testing.T) {
	tests := []struct {
		name        string
		echoCtx     func(e *echo.Echo) echo.Context
		newsUsecase func() *mocks.NewsUsecase
		wantErr     bool
	}{
		{
			name: "POST /api/v1/news 정상적으로 호출",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/news",
				Method: http.MethodPost,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Body: `{
				  "attachment_urls": [
					"https://es-jp-devv2-public-file-storage.s3.ap-northeast-2.amazonaws.com/attachment_file/2024-07-26/cqhkh82nulri2.jpg"
				  ],
				  "closing_date": "2024-10-10 09:00:00",
				  "content_html": "<body><h1>테스트입니다.</h1></body>",
				  "content_image_urls": [
					"https://es-jp-devv2-public-file-storage.s3.ap-northeast-2.amazonaws.com/content_img_file/2024-07-26/cqhkh82nulri1.jpg"
				  ],
				  "pin": false,
				  "posted": true,
				  "title": "테스트 입니다.",
				  "send_mail_status": false
				}`,
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "admin@email.com",
					"name":  "admin",
					"role":  []string{"ROLE_ADMIN", "ROLE_OPERATOR"},
				},
			}),
			newsUsecase: func() *mocks.NewsUsecase {
				usecase := &mocks.NewsUsecase{}
				jst := lo.Must(time.LoadLocation("Asia/Tokyo"))
				usecase.On("Create", mock.Anything, adomain.NewsCreateRequest{
					Pin:              lo.ToPtr(false),
					ContentHtml:      "<body><h1>테스트입니다.</h1></body>",
					Title:            "테스트 입니다.",
					Posted:           true,
					SendMailStatus:   false,
					ClosingDate:      lo.ToPtr(jsonutils.JSONTime(lo.Must(time.ParseInLocation(time.DateTime, "2024-10-10 09:00:00", jst)))),
					ContentImageUrls: []string{"https://es-jp-devv2-public-file-storage.s3.ap-northeast-2.amazonaws.com/content_img_file/2024-07-26/cqhkh82nulri1.jpg"},
					AttachmentUrls:   []string{"https://es-jp-devv2-public-file-storage.s3.ap-northeast-2.amazonaws.com/attachment_file/2024-07-26/cqhkh82nulri2.jpg"},
					JWTUser: jwthelper.User{
						Id:    1000,
						Email: "admin@email.com",
						Name:  "admin",
						Roles: map[string]int{"ROLE_ADMIN": 1, "ROLE_OPERATOR": 1},
					},
					ClientIp: "192.0.2.1",
				}).Return(nil)
				return usecase
			},
			wantErr: false,
		},
		{
			name: "POST /api/v1/news 최대 핀 개수가 초과하는 경우",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/newss",
				Method: http.MethodPost,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Body: `{
				  "attachment_urls": [
					"https://es-jp-devv2-public-file-storage.s3.ap-northeast-2.amazonaws.com/attachment_file/2024-07-26/cqhkh82nulri2.jpg"
				  ],
				  "closing_date": "2024-10-10 00:00:00",
				  "content_html": "<body><h1>테스트입니다.</h1></body>",
				  "content_image_urls": [
					"https://es-jp-devv2-public-file-storage.s3.ap-northeast-2.amazonaws.com/content_img_file/2024-07-26/cqhkh82nulri1.jpg"
				  ],
				  "pin": true,
				  "posted": true,
				  "title": "테스트 입니다.",
				  "send_mail_status": true
				}`,
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "admin@email.com",
					"name":  "admin",
					"role":  []string{"ROLE_ADMIN", "ROLE_OPERATOR"},
				},
			}),
			newsUsecase: func() *mocks.NewsUsecase {
				usecase := &mocks.NewsUsecase{}
				jst := lo.Must(time.LoadLocation("Asia/Tokyo"))
				usecase.On("Create", mock.Anything, adomain.NewsCreateRequest{
					Pin:              lo.ToPtr(true),
					ContentHtml:      "<body><h1>테스트입니다.</h1></body>",
					Title:            "테스트 입니다.",
					Posted:           true,
					SendMailStatus:   true,
					ClosingDate:      lo.ToPtr(jsonutils.JSONTime(lo.Must(time.ParseInLocation(time.DateTime, "2024-10-10 00:00:00", jst)))),
					ContentImageUrls: []string{"https://es-jp-devv2-public-file-storage.s3.ap-northeast-2.amazonaws.com/content_img_file/2024-07-26/cqhkh82nulri1.jpg"},
					AttachmentUrls:   []string{"https://es-jp-devv2-public-file-storage.s3.ap-northeast-2.amazonaws.com/attachment_file/2024-07-26/cqhkh82nulri2.jpg"},
					JWTUser: jwthelper.User{
						Id:    1000,
						Email: "admin@email.com",
						Name:  "admin",
						Roles: map[string]int{"ROLE_ADMIN": 1, "ROLE_OPERATOR": 1},
					},
					ClientIp: "192.0.2.1",
				}).Return(cdomain.ErrRequestLimitExceeded)
				return usecase
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := newsHttpDelivery{
				newsUsecase: tt.newsUsecase(),
				extractor:   jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.Create(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "Create() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_newsHttpDelivery_Update(t *testing.T) {
	tests := []struct {
		name        string
		echoCtx     func(e *echo.Echo) echo.Context
		newsUsecase func() *mocks.NewsUsecase
		wantErr     bool
	}{
		{
			name: "PUT /api/v1/newss/:id 정상적으로 호출",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/newss/:id",
				Method: http.MethodPut,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Params: map[string]string{
					"id": "1",
				},
				Body: `{
				  "attachment_urls": [
					"https://es-jp-devv2-public-file-storage.s3.ap-northeast-2.amazonaws.com/attachment_file/2024-07-26/cqhkh82nulri2.jpg"
				  ],
				  "closing_date": "2024-10-10 00:00:00",
				  "content_html": "<body><h1>테스트입니다.</h1></body>",
				  "content_image_urls": [
					"https://es-jp-devv2-public-file-storage.s3.ap-northeast-2.amazonaws.com/content_img_file/2024-07-26/cqhkh82nulri1.jpg"
				  ],
				  "pin": false,
				  "posted": true,
				  "title": "테스트 입니다."
				}`,
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "admin@email.com",
					"name":  "admin",
					"role":  []string{"ROLE_ADMIN", "ROLE_OPERATOR"},
				},
			}),
			newsUsecase: func() *mocks.NewsUsecase {
				usecase := &mocks.NewsUsecase{}
				jst := lo.Must(time.LoadLocation("Asia/Tokyo"))
				usecase.On("Update", mock.Anything, adomain.NewsUpdateRequest{
					Id:               1,
					Pin:              lo.ToPtr(false),
					ContentHtml:      "<body><h1>테스트입니다.</h1></body>",
					Title:            "테스트 입니다.",
					Posted:           true,
					ClosingDate:      lo.ToPtr(jsonutils.JSONTime(lo.Must(time.ParseInLocation(time.DateTime, "2024-10-10 00:00:00", jst)))),
					ContentImageUrls: []string{"https://es-jp-devv2-public-file-storage.s3.ap-northeast-2.amazonaws.com/content_img_file/2024-07-26/cqhkh82nulri1.jpg"},
					AttachmentUrls:   []string{"https://es-jp-devv2-public-file-storage.s3.ap-northeast-2.amazonaws.com/attachment_file/2024-07-26/cqhkh82nulri2.jpg"},
					JWTUser: jwthelper.User{
						Id:    1000,
						Email: "admin@email.com",
						Name:  "admin",
						Roles: map[string]int{"ROLE_ADMIN": 1, "ROLE_OPERATOR": 1},
					},
					ClientIp: "192.0.2.1",
				}).Return(nil)
				return usecase
			},
			wantErr: false,
		},
		{
			name: "PUT /api/v1/newss pin 최대 개수 초과",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/newss/:id",
				Method: http.MethodPut,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Params: map[string]string{
					"id": "1",
				},
				Body: `{
				  "attachment_urls": [
					"https://es-jp-devv2-public-file-storage.s3.ap-northeast-2.amazonaws.com/attachment_file/2024-07-26/cqhkh82nulri2.jpg"
				  ],
				  "closing_date": "2024-10-10 00:00:00",
				  "content_html": "<body><h1>테스트입니다.</h1></body>",
				  "content_image_urls": [
					"https://es-jp-devv2-public-file-storage.s3.ap-northeast-2.amazonaws.com/content_img_file/2024-07-26/cqhkh82nulri1.jpg"
				  ],
				  "pin": true,
				  "posted": true,
				  "title": "테스트 입니다."
				}`,
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "admin@email.com",
					"name":  "admin",
					"role":  []string{"ROLE_ADMIN", "ROLE_OPERATOR"},
				},
			}),
			newsUsecase: func() *mocks.NewsUsecase {
				usecase := &mocks.NewsUsecase{}
				jst := lo.Must(time.LoadLocation("Asia/Tokyo"))
				usecase.On("Update", mock.Anything, adomain.NewsUpdateRequest{
					Id:               1,
					Pin:              lo.ToPtr(true),
					ContentHtml:      "<body><h1>테스트입니다.</h1></body>",
					Title:            "테스트 입니다.",
					Posted:           true,
					ClosingDate:      lo.ToPtr(jsonutils.JSONTime(lo.Must(time.ParseInLocation(time.DateTime, "2024-10-10 00:00:00", jst)))),
					ContentImageUrls: []string{"https://es-jp-devv2-public-file-storage.s3.ap-northeast-2.amazonaws.com/content_img_file/2024-07-26/cqhkh82nulri1.jpg"},
					AttachmentUrls:   []string{"https://es-jp-devv2-public-file-storage.s3.ap-northeast-2.amazonaws.com/attachment_file/2024-07-26/cqhkh82nulri2.jpg"},
					JWTUser: jwthelper.User{
						Id:    1000,
						Email: "admin@email.com",
						Name:  "admin",
						Roles: map[string]int{"ROLE_ADMIN": 1, "ROLE_OPERATOR": 1},
					},
					ClientIp: "192.0.2.1",
				}).Return(cdomain.ErrRequestLimitExceeded)
				return usecase
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := newsHttpDelivery{
				newsUsecase: tt.newsUsecase(),
				extractor:   jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.Update(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "Create() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}
