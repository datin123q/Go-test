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

var bannerHTTPTestJST = func() *time.Location {
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		panic(err)
	}
	return loc
}()

const bannerApi = "/api/v1/banners"

func Test_bannerHttpDelivery_Create(t *testing.T) {
	tests := []struct {
		name          string
		echoCtx       func(e *echo.Echo) echo.Context
		bannerUsecase func() *mocks.BannerUsecase
		wantErr       bool
	}{
		{
			name: "(성공) 배너 생성",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   bannerApi,
				Method: http.MethodPost,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Body: `{
					"title": "test",
					"posted": true,	
					"start_date": "2021-01-01 00:00:00",	
					"end_date": "2021-01-01 23:59:59",
					"url": "http://test.com",
					"sort_order": 1,
					"description": "test",
					"new_tab": true
				}`,
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "admin@email.com",
					"name":  "admin",
					"role":  []string{"ROLE_ADMIN", "ROLE_OPERATOR"},
				},
			}),
			bannerUsecase: func() *mocks.BannerUsecase {
				usecase := &mocks.BannerUsecase{}
				usecase.On("Create", mock.Anything, adomain.BannerCreateRequest{
					JWTUser: jwthelper.User{
						Id:    1000,
						Email: "admin@email.com",
						Name:  "admin",
						Roles: map[string]int{
							string(domain.ROLE_OPERATOR): 1,
							string(domain.ROLE_ADMIN):    1,
						},
					},
					Title:       "test",
					IsPosted:    lo.ToPtr(true),
					StartDate:   "2021-01-01 00:00:00",
					EndDate:     "2021-01-01 23:59:59",
					Url:         "http://test.com",
					SortOrder:   lo.ToPtr(int64(1)),
					Description: lo.ToPtr("test"),
					NewTab:      lo.ToPtr(true),
				}).Return(adomain.BannerCreateResponse{
					Id:          1,
					Title:       "test",
					IsPosted:    true,
					StartDate:   jsonutils.JSONTime(time.Date(2024, 1, 1, 0, 0, 0, 0, bannerHTTPTestJST)),
					EndDate:     jsonutils.JSONTime(time.Date(2024, 1, 1, 23, 59, 59, 0, bannerHTTPTestJST)),
					Url:         "http://test.com",
					SortOrder:   lo.ToPtr(int64(1)),
					Description: "test",
					NewTab:      true,
					Created:     jsonutils.JSONTime(time.Now()),
				}, nil)
				return usecase
			},
			wantErr: false,
		},
		{
			name: "(실패) 배너 생성 - 필수 값 없는 경우",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   bannerApi,
				Method: http.MethodPost,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Body: `{
					"title": "",
					"posted": true,	
					"url": "http://test.com",
				}`,
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "admin@email.com",
					"name":  "admin",
					"role":  []string{"ROLE_ADMIN", "ROLE_OPERATOR"},
				},
			}),
			bannerUsecase: func() *mocks.BannerUsecase {
				usecase := &mocks.BannerUsecase{}
				return usecase
			},
			wantErr: true,
		},
		{
			name: "(실패) 배너 생성 - IsPosted:true 인데 Order 가 없는 경우",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   bannerApi,
				Method: http.MethodPost,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Body: `{
					"title": "test",
					"posted": true,	
					"start_date": "2021-01-01 00:00:00",	
					"end_date": "2021-01-01 23:59:59",
					"url": "http://test.com",
					"new_tab": false
				}`,
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "admin@email.com",
					"name":  "admin",
					"role":  []string{"ROLE_ADMIN", "ROLE_OPERATOR"},
				},
			}),
			bannerUsecase: func() *mocks.BannerUsecase {
				usecase := &mocks.BannerUsecase{}
				return usecase
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := bannerHttpDelivery{
				bannerUsecase: tt.bannerUsecase(),
				extractor:     jwthelper.NewUserExtractor(),
			}

			e := echo.New()

			err := delivery.Create(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "Create() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_bannerHttpDelivery_GetAll(t *testing.T) {
	tests := []struct {
		name          string
		echoCtx       func(e *echo.Echo) echo.Context
		bannerUsecase func() *mocks.BannerUsecase
		wantErr       bool
	}{
		{
			name: "(성공) 배너 목록 조회",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   bannerApi,
				Method: http.MethodGet,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "admin@email.com",
					"name":  "admin",
					"role":  []string{"ROLE_ADMIN", "ROLE_OPERATOR"},
				},
				QueryParams: map[string][]string{
					"title":            {"test"},
					"max_result_count": {"10"},
					"skip_count":       {"0"},
				},
			}),
			bannerUsecase: func() *mocks.BannerUsecase {
				usecase := &mocks.BannerUsecase{}
				usecase.On("GetAll", mock.Anything, adomain.BannerGetAllRequest{
					JWTUser: jwthelper.User{
						Id:    1000,
						Email: "admin@email.com",
						Name:  "admin",
						Roles: map[string]int{
							string(domain.ROLE_OPERATOR): 1,
							string(domain.ROLE_ADMIN):    1,
						},
					},
					BannerGetAllQueryParam: adomain.BannerGetAllQueryParam{
						SearchInput: cdomain.SearchInput{
							MaxResultCount: 10,
							SkipCount:      0,
						},
						Title: lo.ToPtr("test"),
					},
				}).Return(adomain.BannerGetAllResponse{
					TotalCount:        int64(0),
					TotalBannerPosted: int64(0),
					RollingSecond:     5,
					Items:             []adomain.BannerGetAllItem{},
				}, nil)

				return usecase
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := bannerHttpDelivery{
				bannerUsecase: tt.bannerUsecase(),
				extractor:     jwthelper.NewUserExtractor(),
			}

			e := echo.New()

			err := delivery.GetAll(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "GetAll() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_bannerHttpDelivery_GetOne(t *testing.T) {
	tests := []struct {
		name          string
		echoCtx       func(e *echo.Echo) echo.Context
		bannerUsecase func() *mocks.BannerUsecase
		wantErr       bool
	}{
		{
			name: "(성공) 배너 상세 조회",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   bannerApi,
				Method: http.MethodGet,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "admin@email.com",
					"name":  "admin",
					"role":  []string{"ROLE_ADMIN", "ROLE_OPERATOR"},
				},
				Params: map[string]string{"id": "1"},
			}),
			bannerUsecase: func() *mocks.BannerUsecase {
				usecase := &mocks.BannerUsecase{}
				usecase.On("GetOne", mock.Anything, adomain.BannerGetOneRequest{
					JWTUser: jwthelper.User{
						Id:    1000,
						Email: "admin@email.com",
						Name:  "admin",
						Roles: map[string]int{
							string(domain.ROLE_OPERATOR): 1,
							string(domain.ROLE_ADMIN):    1,
						},
					},
					Id: 1,
				}).Return(adomain.BannerGetOneResponse{}, nil)

				return usecase
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := bannerHttpDelivery{
				bannerUsecase: tt.bannerUsecase(),
				extractor:     jwthelper.NewUserExtractor(),
			}

			e := echo.New()

			err := delivery.GetOne(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "GetOne() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_bannerHttpDelivery_Update(t *testing.T) {
	tests := []struct {
		name          string
		echoCtx       func(e *echo.Echo) echo.Context
		bannerUsecase func() *mocks.BannerUsecase
		wantErr       bool
	}{
		{
			name: "(성공) 배너 수정",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   bannerApi,
				Method: http.MethodPut,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Params: map[string]string{"id": "1"},
				Body: `{
					"title": "test",
					"posted": true,	
					"start_date": "2021-01-01 00:00:00",	
					"end_date": "2021-01-01 23:59:59",
					"url": "http://test.com",
					"sort_order": 1,
					"new_tab": false
				}`,
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "admin@email.com",
					"name":  "admin",
					"role":  []string{"ROLE_ADMIN", "ROLE_OPERATOR"},
				},
			}),
			bannerUsecase: func() *mocks.BannerUsecase {
				usecase := &mocks.BannerUsecase{}
				usecase.On("Update", mock.Anything, adomain.BannerUpdateRequest{
					JWTUser: jwthelper.User{
						Id:    1000,
						Email: "admin@email.com",
						Name:  "admin",
						Roles: map[string]int{
							string(domain.ROLE_OPERATOR): 1,
							string(domain.ROLE_ADMIN):    1,
						},
					},
					Id:        1,
					Title:     "test",
					IsPosted:  lo.ToPtr(true),
					StartDate: "2021-01-01 00:00:00",
					EndDate:   "2021-01-01 23:59:59",
					Url:       "http://test.com",
					SortOrder: lo.ToPtr(int64(1)),
					NewTab:    lo.ToPtr(false),
				}).Return(adomain.BannerUpdateResponse{}, nil)

				return usecase
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := bannerHttpDelivery{
				bannerUsecase: tt.bannerUsecase(),
				extractor:     jwthelper.NewUserExtractor(),
			}

			e := echo.New()

			err := delivery.Update(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "Update() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_bannerHttpDelivery_Delete(t *testing.T) {
	tests := []struct {
		name          string
		echoCtx       func(e *echo.Echo) echo.Context
		bannerUsecase func() *mocks.BannerUsecase
		wantErr       bool
	}{
		{
			name: "(성공) 배너 삭제",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   bannerApi,
				Method: http.MethodDelete,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Params: map[string]string{"id": "1"},
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "admin@email.com",
					"name":  "admin",
					"role":  []string{"ROLE_ADMIN", "ROLE_OPERATOR"},
				},
			}),
			bannerUsecase: func() *mocks.BannerUsecase {
				usecase := &mocks.BannerUsecase{}
				usecase.On("Delete", mock.Anything, adomain.BannerDeleteRequest{
					JWTUser: jwthelper.User{
						Id:    1000,
						Email: "admin@email.com",
						Name:  "admin",
						Roles: map[string]int{
							string(domain.ROLE_OPERATOR): 1,
							string(domain.ROLE_ADMIN):    1,
						},
					},
					Id: 1,
				}).Return(nil)
				return usecase
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := bannerHttpDelivery{
				bannerUsecase: tt.bannerUsecase(),
				extractor:     jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.Delete(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "Delete() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_bannerHttpDelivery_UpdateRollingTime(t *testing.T) {
	tests := []struct {
		name          string
		echoCtx       func(e *echo.Echo) echo.Context
		bannerUsecase func() *mocks.BannerUsecase
		wantErr       bool
	}{
		{
			name: "(성공) 배너 롤링 시간 수정",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   bannerApi + "/rolling-time",
				Method: http.MethodPut,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Body: `{
					"rolling_second": 5
				}`,
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "admin@email.com",
					"name":  "admin",
					"role":  []string{"ROLE_ADMIN", "ROLE_OPERATOR"},
				},
			}),
			bannerUsecase: func() *mocks.BannerUsecase {
				usecase := &mocks.BannerUsecase{}
				usecase.On("UpdateRollingTime", mock.Anything, adomain.BannerUpdateRollingTimeRequest{
					JWTUser: jwthelper.User{
						Id:    1000,
						Email: "admin@email.com",
						Name:  "admin",
						Roles: map[string]int{
							string(domain.ROLE_OPERATOR): 1,
							string(domain.ROLE_ADMIN):    1,
						},
					},
					RollingSecond: 5,
				}).Return(nil)

				return usecase
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := bannerHttpDelivery{
				bannerUsecase: tt.bannerUsecase(),
				extractor:     jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.UpdateRollingTime(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "UpdateRollingTime() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}
