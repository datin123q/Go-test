package http

import (
	"net/http"
	"testing"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/labstack/echo/v4"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	pkgtest "github.com/illuminarean/energy-share-core-api/delivery/testing"
	adomain "github.com/illuminarean/energy-share-core-api/domain"
	"github.com/illuminarean/energy-share-core-api/domain/mocks"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/validator"
)

const userApi = "/api/v1/users"

func Test_UserHttpDelivery_GetSentMails(t *testing.T) {
	tests := []struct {
		name        string
		echoCtx     func(e *echo.Echo) echo.Context
		userUsecase func() *mocks.UserUsecase
		wantErr     bool
	}{
		{
			name: "GET /v1/users/:id/mails",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   userApi + "/:id/mails",
				Params: map[string]string{"id": "146"},
				Method: http.MethodGet,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Claims: map[string]interface{}{
					"user":  "12378123",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_ADMIN", "ROLE_INVESTOR"},
				},
				QueryParams: map[string][]string{
					"sortby":           {"created"},
					"order":            {"desc"},
					"skip_count":       {"0"},
					"max_result_count": {"10"},
				},
			}),
			userUsecase: func() *mocks.UserUsecase {
				m := &mocks.UserUsecase{}
				m.On("GetSentMails", mock.Anything, adomain.UserGetSentMailsRequest{
					JWTUser: jwthelper.User{
						Id:    12378123,
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{"ROLE_ADMIN": 1, "ROLE_INVESTOR": 1},
					},
					ClientIp: "192.0.2.1",
					UserId:   146,
					SearchInput: cdomain.SearchInput{
						Sortby:         "created",
						Order:          "desc",
						SkipCount:      0,
						MaxResultCount: 10,
					},
				}).Return(adomain.MailGetAllResponse{
					TotalCount: 1,
					Items: []adomain.MailResponse{
						mockMail1,
						mockMail2,
					},
				}, nil)
				return m
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := userHttpDelivery{
				extractor:   jwthelper.NewUserExtractor(),
				userUsecase: tt.userUsecase(),
			}
			e := echo.New()
			e.Validator = validator.NewCustomValidator()
			err := delivery.GetSentMails(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "GetSentMails() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_userHttpDelivery_DownloadUserByFilter(t *testing.T) {
	tests := []struct {
		name        string
		echoCtx     func(e *echo.Echo) echo.Context
		userUsecase func() *mocks.UserUsecase
		wantErr     bool
	}{
		{
			name: "POST /v1/users/filter/csv",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   userApi + "/filter/csv",
				Method: http.MethodPost,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Claims: map[string]interface{}{
					"user":  "1",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_ADMIN", "ROLE_INVESTOR"},
				},
				Body: `{
					"genders": ["MALE", "FEMALE"],
					"has_my_number": true,
					"Age": {
						"min": 20,
						"max": 30
					}
				}`,
			}),
			userUsecase: func() *mocks.UserUsecase {
				usecase := &mocks.UserUsecase{}
				usecase.On("DownloadUserByFilter", mock.Anything, adomain.UserDownloadUserByFilterRequest{
					JWTUser: jwthelper.User{
						Id:    1,
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{"ROLE_ADMIN": 1, "ROLE_INVESTOR": 1},
					},
					ClientIp:    "192.0.2.1",
					Genders:     []domain.GENDER{domain.MALE, domain.FEMALE},
					HasMyNumber: lo.ToPtr(true),
					Age: &adomain.NumberRange{
						Min: lo.ToPtr(int64(20)),
						Max: lo.ToPtr(int64(30)),
					}}).Return(adomain.UserDownloadUserByFilterResponse{}, nil)

				return usecase
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := userHttpDelivery{
				extractor:   jwthelper.NewUserExtractor(),
				userUsecase: tt.userUsecase(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.DownloadUserByFilter(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "DownloadUserByFilter() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}
