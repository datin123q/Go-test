package http

import (
	"net/http"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	pkgtest "github.com/illuminarean/energy-share-core-api/delivery/testing"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/validator"
	udomain "github.com/illuminarean/energyshares-user-api/domain"
	"github.com/illuminarean/energyshares-user-api/domain/mocks"
)

const myNumberApi = "/api/v1/my-number"

func Test_myNumberHttpDelivery_Create(t *testing.T) {
	tests := []struct {
		name            string
		echoCtx         func(e *echo.Echo) echo.Context
		myNumberUsecase func() *mocks.MyNumberUsecase
		wantErr         bool
	}{
		{
			name: "(성공) 마이넘버 등록",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   myNumberApi,
				Method: http.MethodPost,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_USER"},
				},
				Body: `{
				   "my_number": "123456789012",
				   "front_file_url": "front_file_url",
				   "back_file_url": "back_file_url"
				}`,
			}),
			myNumberUsecase: func() *mocks.MyNumberUsecase {
				usecase := &mocks.MyNumberUsecase{}
				usecase.On("Create", mock.Anything, udomain.MyNumberCreateRequest{
					JWTUser: jwthelper.User{
						Id:    int64(1000),
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{"ROLE_USER": 1},
					},
					MyNumber:     "123456789012",
					FrontFileUrl: "front_file_url",
					BackFileUrl:  "back_file_url",
				}).Return(udomain.MyNumberCreateResponse{
					MyNumberBody: udomain.MyNumberBody{
						Id:        1,
						MyNumber:  "123456789012",
						FrontFile: &udomain.UserFileBody{},
						BackFile:  &udomain.UserFileBody{},
						Status:    domain.WAIT_APPROVE,
					},
				}, nil)
				return usecase
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := myNumberHttpDelivery{
				myNumberUsecase: tt.myNumberUsecase(),
				extractor:       jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()
			err := delivery.Create(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "Create() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_myNumberHttpDelivery_Get(t *testing.T) {
	tests := []struct {
		name            string
		echoCtx         func(e *echo.Echo) echo.Context
		myNumberUsecase func() *mocks.MyNumberUsecase
		wantErr         bool
	}{
		{
			name: "(성공) 마이넘버 조회",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   myNumberApi,
				Method: http.MethodGet,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_USER"},
				},
			}),
			myNumberUsecase: func() *mocks.MyNumberUsecase {
				usecase := &mocks.MyNumberUsecase{}
				usecase.On("Get", mock.Anything, udomain.MyNumberGetRequest{
					JWTUser: jwthelper.User{
						Id:    int64(1000),
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{"ROLE_USER": 1},
					},
				}).Return(udomain.MyNumberGetResponse{
					MyNumberBody: &udomain.MyNumberBody{
						Id:        1,
						MyNumber:  "123456789012",
						FrontFile: &udomain.UserFileBody{},
						BackFile:  &udomain.UserFileBody{},
						Status:    domain.WAIT_APPROVE,
					},
				}, nil)
				return usecase
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := myNumberHttpDelivery{
				myNumberUsecase: tt.myNumberUsecase(),
				extractor:       jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()
			err := delivery.Get(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "Get() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}
