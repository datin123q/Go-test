package http

import (
	"net/http"
	"testing"
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"

	pkgtest "github.com/illuminarean/energy-share-core-api/delivery/testing"
	adomain "github.com/illuminarean/energy-share-core-api/domain"
	"github.com/illuminarean/energy-share-core-api/domain/mocks"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/validator"
	"github.com/labstack/echo/v4"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_faqHttpDelivery_GetAll(t *testing.T) {
	tests := []struct {
		name       string
		echoCtx    func(e *echo.Echo) echo.Context
		faqUsecase func() *mocks.FaqUsecase
		wantErr    bool
	}{
		{
			name: "GET /api/v1/faqs 정상적으로 호출 (쿼리파라미터 사용)",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/faqs",
				Method: http.MethodGet,
				Claims: map[string]interface{}{
					"user":  "12378123",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_USER", "ROLE_PAYMENT"},
				},
				QueryParams: map[string][]string{
					"category_type":    {"FAQ_CATEGORY_ACCOUNT_REGISTRATION"},
					"max_result_count": {"10"},
					"order":            {"asc"},
					"skip_count":       {"0"},
					"sortby":           {"title"},
				},
			}),
			faqUsecase: func() *mocks.FaqUsecase {
				usecase := &mocks.FaqUsecase{}
				usecase.On("GetAll", mock.Anything, adomain.FaqGetAllRequest{
					JWTUser: jwthelper.User{
						Id:    12378123,
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{"ROLE_USER": 1, "ROLE_PAYMENT": 1},
					},
					ClientIp: "192.0.2.1",
					FaqGetAllQueryParam: adomain.FaqGetAllQueryParam{
						CategoryType: lo.ToPtr("FAQ_CATEGORY_ACCOUNT_REGISTRATION"),
						SearchInput: cdomain.SearchInput{
							Sortby:         "title",
							Order:          "asc",
							SkipCount:      0,
							MaxResultCount: 10,
						},
					},
				}).Return(adomain.FaqGetAllResponse{}, nil)

				return usecase
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := faqHttpDelivery{
				faqUsecase: tt.faqUsecase(),
				extractor:  jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.GetAll(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "GetAll() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_faqHttpDelivery_GetOne(t *testing.T) {
	tests := []struct {
		name       string
		echoCtx    func(e *echo.Echo) echo.Context
		faqUsecase func() *mocks.FaqUsecase
		wantErr    bool
	}{
		{
			name: "GET /api/v1/faqs/:id 정상적으로 호출",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/faqs/:id",
				Method: http.MethodGet,
				Params: map[string]string{
					"id": "1",
				},
				Claims: map[string]interface{}{
					"user":  "12378123",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_USER", "ROLE_PAYMENT"},
				},
			}),
			faqUsecase: func() *mocks.FaqUsecase {
				usecase := &mocks.FaqUsecase{}
				usecase.On("GetOne", mock.Anything, adomain.FaqGetOneRequest{
					JWTUser:  jwthelper.User{Id: 12378123, Email: "gildong@email.com", Name: "홍길동", Roles: map[string]int{"ROLE_USER": 1, "ROLE_PAYMENT": 1}},
					ClientIp: "192.0.2.1",
					Id:       1,
				}).Return(adomain.FaqGetOneResponse{}, nil)

				return usecase
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := faqHttpDelivery{
				faqUsecase: tt.faqUsecase(),
				extractor:  jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.GetOne(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "GetOne() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_faqHttpDelivery_Delete(t *testing.T) {
	tests := []struct {
		name       string
		echoCtx    func(e *echo.Echo) echo.Context
		faqUsecase func() *mocks.FaqUsecase
		wantErr    bool
	}{
		{
			name: "DELETE /api/v1/faqs/:id 정상적으로 호출",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/faqs/:id",
				Method: http.MethodDelete,
				Params: map[string]string{
					"id": "1",
				},
				Claims: map[string]interface{}{
					"user":  "12378123",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_USER", "ROLE_PAYMENT"},
				},
			}),
			faqUsecase: func() *mocks.FaqUsecase {
				usecase := &mocks.FaqUsecase{}
				usecase.On("Delete", mock.Anything, adomain.FaqDeleteRequest{
					JWTUser:  jwthelper.User{Id: 12378123, Email: "gildong@email.com", Name: "홍길동", Roles: map[string]int{"ROLE_USER": 1, "ROLE_PAYMENT": 1}},
					ClientIp: "192.0.2.1",
					Id:       1,
				}).Return(nil)

				return usecase
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := faqHttpDelivery{
				faqUsecase: tt.faqUsecase(),
				extractor:  jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.Delete(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "Delete() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_faqHttpDelivery_Create(t *testing.T) {
	tests := []struct {
		name       string
		echoCtx    func(e *echo.Echo) echo.Context
		faqUsecase func() *mocks.FaqUsecase
		wantErr    bool
	}{
		{
			name: "POST /api/v1/faqs 정상적으로 호출",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/faqs",
				Method: http.MethodPost,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Body: `{
				  "attachment_urls": [
					"https://es-jp-devv2-public-file-storage.s3.ap-northeast-2.amazonaws.com/attachment_file/2024-07-26/cqhkh82nulri2.jpg"
				  ],
				  "category_type": "FAQ_CATEGORY_ACCOUNT_REGISTRATION",
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
			faqUsecase: func() *mocks.FaqUsecase {
				usecase := &mocks.FaqUsecase{}
				jst := lo.Must(time.LoadLocation("Asia/Tokyo"))
				usecase.On("Create", mock.Anything, adomain.FaqCreateRequest{
					Pin:              lo.ToPtr(false),
					ContentHtml:      "<body><h1>테스트입니다.</h1></body>",
					Title:            "테스트 입니다.",
					Posted:           true,
					ClosingDate:      lo.ToPtr(jsonutils.JSONTime(lo.Must(time.ParseInLocation(time.DateTime, "2024-10-10 00:00:00", jst)))),
					CategoryType:     string(domain.FAQ_CATEGORY_ACCOUNT_REGISTRATION),
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
			name: "POST /api/v1/faqs 카테고리 타입을 비정상적으로 호출(FAQ_CATEGORY_ACCOUNT)",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/faqs",
				Method: http.MethodPost,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Body: `{
				  "attachment_urls": [
					"https://es-jp-devv2-public-file-storage.s3.ap-northeast-2.amazonaws.com/attachment_file/2024-07-26/cqhkh82nulri2.jpg"
				  ],
				  "category_type": "FAQ_CATEGORY_ACCOUNT",
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
			faqUsecase: func() *mocks.FaqUsecase {
				return &mocks.FaqUsecase{}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := faqHttpDelivery{
				faqUsecase: tt.faqUsecase(),
				extractor:  jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.Create(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "Create() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_faqHttpDelivery_Update(t *testing.T) {
	tests := []struct {
		name       string
		echoCtx    func(e *echo.Echo) echo.Context
		faqUsecase func() *mocks.FaqUsecase
		wantErr    bool
	}{
		{
			name: "PUT /api/v1/faqs/:id 정상적으로 호출",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/faqs/:id",
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
				  "category_type": "FAQ_CATEGORY_ACCOUNT_REGISTRATION",
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
			faqUsecase: func() *mocks.FaqUsecase {
				usecase := &mocks.FaqUsecase{}
				jst := lo.Must(time.LoadLocation("Asia/Tokyo"))
				usecase.On("Update", mock.Anything, adomain.FaqUpdateRequest{
					Id:               1,
					Pin:              lo.ToPtr(false),
					ContentHtml:      "<body><h1>테스트입니다.</h1></body>",
					Title:            "테스트 입니다.",
					Posted:           true,
					ClosingDate:      lo.ToPtr(jsonutils.JSONTime(lo.Must(time.ParseInLocation(time.DateTime, "2024-10-10 00:00:00", jst)))),
					CategoryType:     string(domain.FAQ_CATEGORY_ACCOUNT_REGISTRATION),
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
			name: "PUT /api/v1/faqs 카테고리 타입을 비정상적으로 호출(FAQ_CATEGORY_ACCOUNT)",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/faqs/:id",
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
				  "category_type": "FAQ_CATEGORY_ACCOUNT",
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
			faqUsecase: func() *mocks.FaqUsecase {
				return &mocks.FaqUsecase{}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := faqHttpDelivery{
				faqUsecase: tt.faqUsecase(),
				extractor:  jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.Update(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "Create() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}
