package http

import (
	"net/http"
	"testing"

	pkgtest "github.com/illuminarean/energy-share-core-api/delivery/testing"
	adomain "github.com/illuminarean/energy-share-core-api/domain"
	"github.com/illuminarean/energy-share-core-api/domain/mocks"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/validator"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const withdrawalCloseApi = "/api/v1/withdrawals/closed"

func Test_withdrawalHttpDelivery_GetAll(t *testing.T) {
	tests := []struct {
		name                   string
		echoCtx                func(e *echo.Echo) echo.Context
		withdrawalCloseUsecase func() *mocks.WithdrawalCloseUsecase
		wantErr                bool
	}{
		{
			name: "(성공) 마감된 출금의뢰 목록 조회",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   withdrawalCloseApi,
				Method: http.MethodGet,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Claims: map[string]interface{}{
					"user":  "12378123",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_ADMIN", "ROLE_PAYMENT"},
				},
				QueryParams: map[string][]string{
					"sortby":           {"id"},
					"order":            {"desc"},
					"skip_count":       {"0"},
					"max_result_count": {"10"},
				},
			}),
			withdrawalCloseUsecase: func() *mocks.WithdrawalCloseUsecase {
				m := &mocks.WithdrawalCloseUsecase{}
				m.On("GetAll", mock.Anything, adomain.WithdrawalCloseGetAllRequest{
					JWTUser: jwthelper.User{
						Id:    12378123,
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{"ROLE_ADMIN": 1, "ROLE_PAYMENT": 1},
					},
					SearchInput: cdomain.SearchInput{
						Sortby:         "id",
						Order:          "desc",
						SkipCount:      0,
						MaxResultCount: 10,
					},
				}).Return(adomain.WithdrawalCloseGetAllResponse{
					TotalCount: 1,
					Items: []adomain.WithdrawalCloseGetAllItem{
						mockWithdrawalCloseGetAllItem1,
					},
				}, nil)
				return m
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := withdrawalCloseHttpDelivery{
				extractor:              jwthelper.NewUserExtractor(),
				withdrawalCloseUsecase: tt.withdrawalCloseUsecase(),
			}

			e := echo.New()
			err := delivery.GetAll(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_withdrawalHttpDelivery_GetWithdrawals(t *testing.T) {
	tests := []struct {
		name                   string
		echoCtx                func(e *echo.Echo) echo.Context
		withdrawalCloseUsecase func() *mocks.WithdrawalCloseUsecase
		wantErr                bool
	}{
		{
			name: "(성공) 마감에 따른 출금의뢰 목록 조회",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   withdrawalCloseApi + "/:id",
				Method: http.MethodGet,
				Params: map[string]string{"id": "1"},
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
					"role":  []string{"ROLE_ADMIN", "ROLE_PAYMENT"},
				},
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
			}),
			withdrawalCloseUsecase: func() *mocks.WithdrawalCloseUsecase {
				u := &mocks.WithdrawalCloseUsecase{}
				u.On("GetWithdrawals", mock.Anything, adomain.WithdrawalCloseGetWithdrawalsRequest{
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
							string(domain.ROLE_PAYMENT): 1,
							string(domain.ROLE_ADMIN):   1,
						},
					},
					Id: 1,
				}).Return(adomain.WithdrawalCloseGetWithdrawalsResponse{}, nil)
				return u
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := withdrawalCloseHttpDelivery{
				withdrawalCloseUsecase: tt.withdrawalCloseUsecase(),
				extractor:              jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.GetWithdrawals(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "GetWithdrawals() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_withdrawalHttpDelivery_Download(t *testing.T) {
	tests := []struct {
		name                   string
		echoCtx                func(e *echo.Echo) echo.Context
		withdrawalCloseUsecase func() *mocks.WithdrawalCloseUsecase
		wantErr                bool
	}{
		{
			name: "(성공) 마감에 따른 출금의뢰 목록 다운로드",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   withdrawalCloseApi + "/:id/download",
				Method: http.MethodGet,
				Params: map[string]string{"id": "1"},
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
					"role":  []string{"ROLE_ADMIN", "ROLE_PAYMENT"},
				},
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
			}),
			withdrawalCloseUsecase: func() *mocks.WithdrawalCloseUsecase {
				u := &mocks.WithdrawalCloseUsecase{}
				u.On("Download", mock.Anything, adomain.WithdrawalCloseDownloadRequest{
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
							string(domain.ROLE_PAYMENT): 1,
							string(domain.ROLE_ADMIN):   1,
						},
					},
					ClientIp: "192.0.2.1",
					Id:       1,
				}).Return(adomain.WithdrawalCloseDownloadResponse{}, nil)
				return u
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := withdrawalCloseHttpDelivery{
				withdrawalCloseUsecase: tt.withdrawalCloseUsecase(),
				extractor:              jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.Download(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "Download() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}
