package http

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/labstack/echo/v4"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	pkgtest "github.com/illuminarean/energy-share-core-api/delivery/testing"
	adomain "github.com/illuminarean/energy-share-core-api/domain"
	"github.com/illuminarean/energy-share-core-api/domain/mocks"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/validator"
)

const adminUserApi = "/api/v1/admin-users"

func TestAdminUserHttpDelivery_Login(t *testing.T) {

	tests := []struct {
		name             string
		echoCtx          func(e *echo.Echo) echo.Context
		adminUserUsecase func() *mocks.AdminUserUsecase
		wantErr          bool
	}{
		{
			name: "[POST] api/v1/admin-users/login 성공",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   adminUserApi + "/login",
				Method: http.MethodPost,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Multipart: pkgtest.Multipart{},
				Body:      `{"email":"admin@energyshares.jp","password":"illuminate!@"}`,
			}),
			adminUserUsecase: func() *mocks.AdminUserUsecase {
				u := &mocks.AdminUserUsecase{}
				u.On("Login", mock.Anything, mock.MatchedBy(
					func(request adomain.AdminUserLoginRequest) bool {
						return cmp.Equal(request, adomain.AdminUserLoginRequest{
							Email:     "admin@energyshares.jp",
							Password:  "illuminate!@",
							ClientIp:  "192.0.2.1",
							UserAgent: "",
						})
					})).Return(adomain.AdminUserLoginResponse{
					AccessToken:  "access token",
					RefreshToken: "",
					Cookies:      nil,
				}, nil)
				return u
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			delivery := adminUserHttpDelivery{
				adminUserUsecase: tt.adminUserUsecase(),
				logger:           factory.NewLoggerFactory(),
				extractor:        jwthelper.NewUserExtractor(),
			}

			err := delivery.Login(tt.echoCtx(e))
			if (err != nil) != tt.wantErr {
				t.Errorf("Login() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAdminUserHttpDelivery_Logout(t *testing.T) {

	tests := []struct {
		name             string
		echoCtx          func(e *echo.Echo) echo.Context
		adminUserUsecase func() *mocks.AdminUserUsecase
		wantErr          bool
	}{
		{
			name: "api/v1/admin-users/logout 성공",
			echoCtx: func(e *echo.Echo) echo.Context {
				req := pkgtest.RequestContext(pkgtest.Request{
					Path:   adminUserApi + "/logout",
					Method: http.MethodPost,
					Claims: map[string]interface{}{
						"user":  "1",
						"email": "admin@energyshares.jp",
						"name":  "Admin",
						"role":  []string{"ROLE_SUPER_ADMIN", "ROLE_ADMIN", "ROLE_OPERATOR"},
					},
					Header: pkgtest.Header{
						Headers:   echo.HeaderContentType,
						MIMETypes: echo.MIMEApplicationJSON,
					},
				})(e)
				req.Request().AddCookie(&http.Cookie{
					Name:  "refresh_token",
					Value: "refresh_token value",
				})

				return req
			},
			adminUserUsecase: func() *mocks.AdminUserUsecase {
				u := &mocks.AdminUserUsecase{}
				u.On("Logout", mock.Anything, adomain.AdminUserLogoutRequest{
					JWTUser: jwthelper.User{
						Id:    int64(1),
						Name:  "Admin",
						Email: "admin@energyshares.jp",
						Roles: map[string]int{
							string(domain.ROLE_OPERATOR):    1,
							string(domain.ROLE_ADMIN):       1,
							string(domain.ROLE_SUPER_ADMIN): 1,
						},
					},
					RefreshToken: "refresh_token value",
					ClientIp:     "192.0.2.1",
					UserAgent:    "",
				}).Return(adomain.AdminUserLogoutResponse{}, nil)
				return u
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			delivery := adminUserHttpDelivery{
				adminUserUsecase: tt.adminUserUsecase(),
				logger:           factory.NewLoggerFactory(),
				extractor:        jwthelper.NewUserExtractor(),
			}

			err := delivery.Logout(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "Logout() error = %v, wantErr %v", err, tt.wantErr)
		})
	}

}
func TestAdminUserHttpDelivery_Create(t *testing.T) {

	tests := []struct {
		name             string
		echoCtx          func(e *echo.Echo) echo.Context
		adminUserUsecase func() *mocks.AdminUserUsecase
		wantErr          bool
	}{
		{
			name: "api/v1/admin-users 성공",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   adminUserApi,
				Method: http.MethodPost,
				Claims: map[string]interface{}{
					"user":  "1",
					"email": "admin@energyshares.jp",
					"name":  "Admin",
					"role":  []string{"ROLE_SUPER_ADMIN", "ROLE_ADMIN", "ROLE_OPERATOR"},
				},
				Body: func() string {
					data, _ := json.Marshal(map[string]interface{}{
						"email":               "operator@energyshares.jp",
						"name":                "operator",
						"password":            "wnsdyd2121!2121",
						"mobile_phone":        "01012345678",
						"validity_start_date": "2024-01-01 00:00:00",
						"validity_end_date":   "2024-12-31 23:59:59",
						"roles":               []string{"ROLE_OPERATOR", "ROLE_PRODUCT", "ROLE_PAYMENT", "ROLE_INVESTOR", "ROLE_MESSENGER"},
					})
					return string(data)
				}(),
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
			}),
			adminUserUsecase: func() *mocks.AdminUserUsecase {
				u := &mocks.AdminUserUsecase{}
				u.On("Create", mock.Anything, mock.MatchedBy(func(request adomain.AdminUserCreateRequest) bool {
					ignoreTypes := cmpopts.IgnoreTypes(map[string]int{}, jsonutils.JSONTime{})
					return cmp.Equal(request, adomain.AdminUserCreateRequest{
						Email:             "operator@energyshares.jp",
						Password:          "wnsdyd2121!2121",
						Name:              "operator",
						MobilePhone:       "01012345678",
						ValidityStartDate: "2024-01-01 00:00:00",
						ValidityEndDate:   "2024-12-31 23:59:59",
						Roles:             []string{"ROLE_OPERATOR", "ROLE_PRODUCT", "ROLE_PAYMENT", "ROLE_INVESTOR", "ROLE_MESSENGER"},
						JWTUser: jwthelper.User{
							Id:    int64(1),
							Name:  "Admin",
							Email: "admin@energyshares.jp",
							Roles: map[string]int{
								string(domain.ROLE_OPERATOR):    1,
								string(domain.ROLE_ADMIN):       1,
								string(domain.ROLE_SUPER_ADMIN): 1,
							},
						},
						ClientIp: "192.0.2.1",
					}, ignoreTypes)
				})).Return(adomain.AdminUserCreateResponse{}, nil)
				return u
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := adminUserHttpDelivery{
				adminUserUsecase: tt.adminUserUsecase(),
				logger:           factory.NewLoggerFactory(),
				extractor:        jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.Create(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "Create() error = %v, wantErr %v", err, tt.wantErr)
		})
	}

}

func TestAdminUserHttpDelivery_Delete(t *testing.T) {

	tests := []struct {
		name             string
		echoCtx          func(e *echo.Echo) echo.Context
		adminUserUsecase func() *mocks.AdminUserUsecase
		wantErr          bool
	}{
		{
			name: "api/v1/admin-users/:id 성공",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   adminUserApi,
				Method: http.MethodDelete,
				Params: map[string]string{"id": "120"},
				Claims: map[string]interface{}{
					"user":  "1",
					"email": "admin@energyshares.jp",
					"name":  "Admin",
					"role":  []string{"ROLE_SUPER_ADMIN", "ROLE_ADMIN", "ROLE_OPERATOR"},
				},
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
			}),
			adminUserUsecase: func() *mocks.AdminUserUsecase {
				u := &mocks.AdminUserUsecase{}
				u.On("Delete", mock.Anything, adomain.AdminUserDeleteRequest{
					Id: 120,
					JWTUser: jwthelper.User{
						Id:    int64(1),
						Name:  "Admin",
						Email: "admin@energyshares.jp",
						Roles: map[string]int{
							string(domain.ROLE_SUPER_ADMIN): 1,
							string(domain.ROLE_OPERATOR):    1,
							string(domain.ROLE_ADMIN):       1,
						},
					},
					ClientIp: "192.0.2.1",
				}).Return(adomain.AdminUserDeleteResponse{}, nil)
				return u
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			delivery := adminUserHttpDelivery{
				adminUserUsecase: tt.adminUserUsecase(),
				logger:           factory.NewLoggerFactory(),
				extractor:        jwthelper.NewUserExtractor(),
			}

			err := delivery.Delete(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "Delete() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func TestAdminUserHttpDelivery_Download(t *testing.T) {
	tests := []struct {
		name             string
		echoCtx          func(e *echo.Echo) echo.Context
		adminUserUsecase func() *mocks.AdminUserUsecase
		wantErr          bool
	}{
		{
			name: "api/v1/admin-users 성공",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   adminUserApi + "/download",
				Method: http.MethodGet,
				QueryParams: map[string][]string{
					"sortby":           {"desc"},
					"order":            {"id"},
					"skip_count":       {"0"},
					"max_result_count": {"30"},
				},
				Claims: map[string]interface{}{
					"user":  "1",
					"email": "admin@energyshares.jp",
					"name":  "Admin",
					"role":  []string{"ROLE_SUPER_ADMIN", "ROLE_ADMIN", "ROLE_OPERATOR"},
				},
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
			}),
			adminUserUsecase: func() *mocks.AdminUserUsecase {
				u := &mocks.AdminUserUsecase{}
				u.On("DownloadHistories", mock.Anything, adomain.AdminUserDownloadHistoriesRequest{
					AdminUserDownloadQueryParam: adomain.AdminUserDownloadQueryParam{
						Name:        nil,
						MobilePhone: nil,
					},
					SearchInput: cdomain.SearchInput{
						Sortby:         "desc",
						Order:          "id",
						SkipCount:      0,
						MaxResultCount: 30,
					},
					JWTUser: jwthelper.User{
						Id:    int64(1),
						Name:  "Admin",
						Email: "admin@energyshares.jp",
						Roles: map[string]int{
							string(domain.ROLE_OPERATOR):    1,
							string(domain.ROLE_ADMIN):       1,
							string(domain.ROLE_SUPER_ADMIN): 1,
						},
					},
					ClientIp: "192.0.2.1",
				}).Return(adomain.AdminUserDownloadHistoriesResponse{}, nil)
				return u
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := adminUserHttpDelivery{
				adminUserUsecase: tt.adminUserUsecase(),
				logger:           factory.NewLoggerFactory(),
				extractor:        jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.Download(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "Download() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func TestAdminUserHttpDelivery_GetAll(t *testing.T) {
	tests := []struct {
		name             string
		echoCtx          func(e *echo.Echo) echo.Context
		adminUserUsecase func() *mocks.AdminUserUsecase
		wantErr          bool
	}{
		{
			name: "api/v1/admin-users 성공",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   adminUserApi,
				Method: http.MethodGet,
				QueryParams: map[string][]string{
					"sortby":           {"desc"},
					"order":            {"id"},
					"skip_count":       {"0"},
					"max_result_count": {"30"},
				},
				Claims: map[string]interface{}{
					"user":  "1",
					"email": "admin@energyshares.jp",
					"name":  "Admin",
					"role":  []string{"ROLE_SUPER_ADMIN", "ROLE_ADMIN", "ROLE_OPERATOR"},
				},
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
			}),
			adminUserUsecase: func() *mocks.AdminUserUsecase {
				u := &mocks.AdminUserUsecase{}
				u.On("GetAll", mock.Anything, adomain.AdminUserGetAllRequest{
					AdminUserGetAllQueryParam: adomain.AdminUserGetAllQueryParam{
						Name:        nil,
						MobilePhone: nil,
						SearchInput: cdomain.SearchInput{
							Sortby:         "desc",
							Order:          "id",
							SkipCount:      0,
							MaxResultCount: 30,
						},
					},
					JWTUser: jwthelper.User{
						Id:    int64(1),
						Name:  "Admin",
						Email: "admin@energyshares.jp",
						Roles: map[string]int{
							string(domain.ROLE_OPERATOR):    1,
							string(domain.ROLE_ADMIN):       1,
							string(domain.ROLE_SUPER_ADMIN): 1,
						},
					},
					ClientIp: "192.0.2.1",
				}).Return(adomain.AdminUserGetAllResponse{}, nil)
				return u
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := adminUserHttpDelivery{
				adminUserUsecase: tt.adminUserUsecase(),
				logger:           factory.NewLoggerFactory(),
				extractor:        jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.GetAll(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "GetAll() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func TestAdminUserHttpDelivery_GetOne(t *testing.T) {
	tests := []struct {
		name             string
		echoCtx          func(e *echo.Echo) echo.Context
		adminUserUsecase func() *mocks.AdminUserUsecase
		wantErr          bool
	}{
		{
			name: "api/v1/admin-users/:id 성공",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   adminUserApi + "/:id",
				Method: http.MethodGet,
				Params: map[string]string{"id": "128931"},
				Claims: map[string]interface{}{
					"user":  "1",
					"email": "admin@energyshares.jp",
					"name":  "Admin",
					"role":  []string{"ROLE_SUPER_ADMIN", "ROLE_ADMIN", "ROLE_OPERATOR"},
				},
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
			}),
			adminUserUsecase: func() *mocks.AdminUserUsecase {
				u := &mocks.AdminUserUsecase{}
				u.On("GetOne", mock.Anything, adomain.AdminUserGetOneRequest{
					Id: 128931,
					JWTUser: jwthelper.User{
						Id:    int64(1),
						Name:  "Admin",
						Email: "admin@energyshares.jp",
						Roles: map[string]int{
							string(domain.ROLE_OPERATOR):    1,
							string(domain.ROLE_ADMIN):       1,
							string(domain.ROLE_SUPER_ADMIN): 1,
						},
					},
					ClientIp: "192.0.2.1",
				}).Return(adomain.AdminUserGetOneResponse{}, nil)
				return u
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := adminUserHttpDelivery{
				adminUserUsecase: tt.adminUserUsecase(),
				logger:           factory.NewLoggerFactory(),
				extractor:        jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.GetOne(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "GetOne() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func TestAdminUserHttpDelivery_Update(t *testing.T) {
	tests := []struct {
		name             string
		echoCtx          func(e *echo.Echo) echo.Context
		adminUserUsecase func() *mocks.AdminUserUsecase
		wantErr          bool
	}{
		{
			name: "api/v1/admin-users/:id 성공",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   adminUserApi + "/:id",
				Method: http.MethodPut,
				Params: map[string]string{"id": "128931"},
				Claims: map[string]interface{}{
					"user":  "1",
					"email": "admin@energyshares.jp",
					"name":  "Admin",
					"role":  []string{"ROLE_SUPER_ADMIN", "ROLE_ADMIN", "ROLE_OPERATOR"},
				},
				Body: func() string {
					data, _ := json.Marshal(map[string]interface{}{
						"email":               "operator@energyshares.jp",
						"password":            "wnsdyd2121!2121",
						"name":                "operator",
						"mobile_phone":        "01012345678",
						"validity_start_date": "2024-01-01 00:00:00",
						"validity_end_date":   "2024-12-31 23:59:59",
						"roles":               []string{"ROLE_ADMIN", "ROLE_OPERATOR", "ROLE_PRODUCT", "ROLE_PAYMENT", "ROLE_INVESTOR", "ROLE_MESSENGER"},
					})
					return string(data)
				}(),
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
			}),
			adminUserUsecase: func() *mocks.AdminUserUsecase {
				u := &mocks.AdminUserUsecase{}
				u.On("Update", mock.Anything, adomain.AdminUserUpdateRequest{
					Id:                128931,
					Email:             "operator@energyshares.jp",
					Password:          lo.ToPtr("wnsdyd2121!2121"),
					Name:              "operator",
					MobilePhone:       "01012345678",
					ValidityStartDate: "2024-01-01 00:00:00",
					ValidityEndDate:   "2024-12-31 23:59:59",
					Roles:             []string{"ROLE_ADMIN", "ROLE_OPERATOR", "ROLE_PRODUCT", "ROLE_PAYMENT", "ROLE_INVESTOR", "ROLE_MESSENGER"},
					JWTUser: jwthelper.User{
						Id:    int64(1),
						Name:  "Admin",
						Email: "admin@energyshares.jp",
						Roles: map[string]int{
							string(domain.ROLE_OPERATOR):    1,
							string(domain.ROLE_ADMIN):       1,
							string(domain.ROLE_SUPER_ADMIN): 1,
						},
					},
					ClientIp: "192.0.2.1",
				}).Return(adomain.AdminUserUpdateResponse{}, nil)
				return u
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := adminUserHttpDelivery{
				adminUserUsecase: tt.adminUserUsecase(),
				logger:           factory.NewLoggerFactory(),
				extractor:        jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.Update(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "Update() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}
