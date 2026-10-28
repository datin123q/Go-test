package http

import (
	"errors"
	"net/http"
	"testing"

	adomain "github.com/illuminarean/energy-share-core-api/domain"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/validator"

	"github.com/pangpanglabs/echoswagger/v2"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/illuminarean/energy-shares-backend-repo/domain"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"

	"github.com/labstack/echo/v4"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/illuminarean/energy-share-core-api/controllers"
	pkgtest "github.com/illuminarean/energy-share-core-api/delivery/testing"
	"github.com/illuminarean/energy-share-core-api/domain/mocks"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

func Test_investmentWaitersHttpDelivery_GetAll(t *testing.T) {
	tests := []struct {
		name                     string
		echoCtx                  func(e *echo.Echo) echo.Context
		investmentWaitersUsecase func() *mocks.InvestmentWaitersUsecase
		wantErr                  bool
	}{
		{
			name: "GET api/v1/investment-waiters 호출성공",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/investment-waiters",
				Method: http.MethodGet,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
			}),
			investmentWaitersUsecase: func() *mocks.InvestmentWaitersUsecase {
				usecase := &mocks.InvestmentWaitersUsecase{}
				usecase.On("GetAll", mock.Anything, adomain.InvestmentWaitersGetAllRequest{
					SearchInput: controllers.SearchInput{
						MaxResultCount: 30,
					},
				}).Return(adomain.InvestmentWaitersGetAllResponse{}, nil)

				return usecase
			},
		},
		{
			name: "GET api/v1/investment-waiters:(project_id=1283831) queryparam 필터링 => 호출성공",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/investment-waiters",
				Method: http.MethodGet,
				QueryParams: map[string][]string{
					"project_id": {"1283831"},
				},
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
			}),
			investmentWaitersUsecase: func() *mocks.InvestmentWaitersUsecase {
				usecase := &mocks.InvestmentWaitersUsecase{}
				usecase.On("GetAll", mock.Anything, adomain.InvestmentWaitersGetAllRequest{
					ProjectId: lo.ToPtr(int64(1283831)),
					SearchInput: controllers.SearchInput{
						MaxResultCount: 30,
					},
				}).Return(adomain.InvestmentWaitersGetAllResponse{}, nil)

				return usecase
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()

			delivery := New(echoswagger.NewNop(e), tt.investmentWaitersUsecase(), factory.NewLoggerFactory(), nil)

			err := delivery.GetAll(tt.echoCtx(echo.New()))
			assert.Equalf(t, tt.wantErr, err != nil, "GetAll() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func TestInvestmentWaitersHttpDelivery_GetOne(t *testing.T) {
	tests := []struct {
		name                     string
		echoCtx                  func(e *echo.Echo) echo.Context
		investmentWaitersUsecase func() *mocks.InvestmentWaitersUsecase
		wantErr                  bool
	}{
		{
			name: "GET api/v1/investment-waiters/:id 호출성공",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path: "/api/v1/investment-waiters/:id",
				Params: map[string]string{
					"id": "128931",
				},
				Method: http.MethodGet,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
			}),
			investmentWaitersUsecase: func() *mocks.InvestmentWaitersUsecase {
				usecase := &mocks.InvestmentWaitersUsecase{}
				usecase.On("GetOne", mock.Anything, adomain.InvestmentWaitersGetOneRequest{
					InvestmentWaitersId: lo.ToPtr(int64(128931)),
				}).Return(adomain.InvestmentWaitersGetOneResponse{}, nil)

				return usecase
			},
			wantErr: false,
		},
		// TODO 예외처리 에러를 하고 싶지만 기존 Controller 구조상 에러 핸들링을 하지못하는 구조로 되어있음 개선필요
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := investmentWaitersHttpDelivery{
				investmentWaitersUsecase: tt.investmentWaitersUsecase(),
				logger:                   factory.NewLoggerFactory(),
			}

			err := delivery.GetOne(tt.echoCtx(echo.New()))
			assert.Equalf(t, tt.wantErr, err != nil, "GetOne() error = %v, wantErr %v", err, tt.wantErr)
			// true
		})
	}
}

func TestInvestmentWaitersHttpDelivery_Downalod(t *testing.T) {
	tests := []struct {
		name                     string
		echoCtx                  func(e *echo.Echo) echo.Context
		investmentWaitersUsecase func() *mocks.InvestmentWaitersUsecase
		wantErr                  bool
	}{
		{
			name: "GET api/v1/list/download 호출성공 : project_id(17908234) queryparma 필터링 => 호출성공",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/list/download",
				Method: http.MethodGet,
				QueryParams: map[string][]string{
					"project_id": {"17908234"},
				},
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
			}),
			investmentWaitersUsecase: func() *mocks.InvestmentWaitersUsecase {
				usecase := &mocks.InvestmentWaitersUsecase{}
				usecase.On("Download", mock.Anything, adomain.InvestmentWaitersDownloadRequest{
					ProjectId: lo.ToPtr(int64(17908234)),
					SearchInput: controllers.SearchInput{
						MaxResultCount: 3000,
					},
				}).Return(adomain.InvestmentWaitersDownloadResponse{}, nil)

				return usecase
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := investmentWaitersHttpDelivery{
				investmentWaitersUsecase: tt.investmentWaitersUsecase(),
				logger:                   factory.NewLoggerFactory(),
			}

			err := delivery.Download(tt.echoCtx(echo.New()))
			assert.Equalf(t, tt.wantErr, err != nil, "Download() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func TestInvestmentWaitersHttpDelivery_StatusUpdate(t *testing.T) {
	tests := []struct {
		name                     string
		echoCtx                  func(e *echo.Echo) echo.Context
		investmentWaitersUsecase func() *mocks.InvestmentWaitersUsecase
		wantErr                  bool
	}{
		{
			name: "PUT api/v1/investment-waiters/:id/status-update : status(DEPOSIT_REQUEST) payload 전달=> 호출성공",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path: "/api/v1/investment-waiters/:id/status-update",
				Params: map[string]string{
					"id": "128931",
				},
				Method: http.MethodPut,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Body: `{"status":"DEPOSIT_REQUEST"}`,
				Claims: map[string]interface{}{
					"user":  "12378123",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_USER"},
				},
			}),
			investmentWaitersUsecase: func() *mocks.InvestmentWaitersUsecase {
				usecase := &mocks.InvestmentWaitersUsecase{}
				usecase.On("StatusUpdate", mock.Anything, mock.MatchedBy(
					func(request adomain.InvestmentWaitersStatusUpdateRequest) bool {
						ignoreTypes := cmpopts.IgnoreTypes(map[string]int{})
						result := cmp.Equal(request, adomain.InvestmentWaitersStatusUpdateRequest{
							InvestmentWaitersId: lo.ToPtr(int64(128931)),
							JWTUser: jwthelper.User{
								Id:    int64(12378123),
								Name:  "홍길동",
								Email: "gildong@email.com",
							},
							InvestmentWaitersStatusUpdateBody: adomain.InvestmentWaitersStatusUpdateBody{
								Status: domain.DEPOSIT_REQUEST,
							},
						}, ignoreTypes)

						return result
					})).Return(adomain.InvestmentWaitersStatusUpdateResponse{}, nil)

				return usecase
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := investmentWaitersHttpDelivery{
				investmentWaitersUsecase: tt.investmentWaitersUsecase(),
				logger:                   factory.NewLoggerFactory(),
				extractor:                jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.StatusUpdate(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "StatusUpdate() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func TestInvestmentWaitersHttpDelivery_InvestmentPropose(t *testing.T) {
	tests := []struct {
		name                     string
		echoCtx                  func(e *echo.Echo) echo.Context
		investmentWaitersUsecase func() *mocks.InvestmentWaitersUsecase
		wantErr                  bool
	}{
		{
			name: "POST api/v1/investment-waiters/:id/investment-propose => 호출성공",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path: "/api/v1/investment-waiters/:id/investment-propose",
				Params: map[string]string{
					"id": "128931",
				},
				Method: http.MethodPost,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Claims: map[string]interface{}{
					"user":  "12378123",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_USER"},
				},
			}),
			investmentWaitersUsecase: func() *mocks.InvestmentWaitersUsecase {
				usecase := &mocks.InvestmentWaitersUsecase{}
				usecase.On("InvestmentPropose", mock.Anything, mock.MatchedBy(
					func(request adomain.InvestmentWaitersInvestmentProposeRequest) bool {
						ignoreTypes := cmpopts.IgnoreTypes(map[string]int{})
						result := cmp.Equal(request, adomain.InvestmentWaitersInvestmentProposeRequest{
							InvestmentWaitersId: lo.ToPtr(int64(128931)),
							JWTUser: jwthelper.User{
								Id:    int64(12378123),
								Name:  "홍길동",
								Email: "gildong@email.com",
							},
						}, ignoreTypes)

						return result
					})).Return(adomain.InvestmentWaitersInvestmentProposeResponse{}, nil)

				return usecase
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := investmentWaitersHttpDelivery{
				investmentWaitersUsecase: tt.investmentWaitersUsecase(),
				logger:                   factory.NewLoggerFactory(),
				extractor:                jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.InvestmentPropose(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "InvestmentPropose() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func TestInvestmentWaitersHttpDelivery_InvestmentRequest(t *testing.T) {
	tests := []struct {
		name                     string
		echoCtx                  func(e *echo.Echo) echo.Context
		investmentWaitersUsecase func() *mocks.InvestmentWaitersUsecase
		wantErr                  bool
	}{
		{
			name: "POST api/v1/investment-waiters/:id/investment-request => 呼び出し成功",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path: "/api/v1/investment-waiters/:id/investment-request",
				Params: map[string]string{
					"id": "128931",
				},
				Body:   `{"qty":5}`,
				Method: http.MethodPost,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Claims: map[string]interface{}{
					"user":  "12378123",
					"email": "kwatanabe@miralita.co.jp",
					"name":  "Kei Watanabe",
					"role":  []string{"ROLE_USER"},
				},
			}),
			investmentWaitersUsecase: func() *mocks.InvestmentWaitersUsecase {
				usecase := &mocks.InvestmentWaitersUsecase{}
				usecase.On("InvestmentRequest", mock.Anything, mock.MatchedBy(
					func(request adomain.InvestmentWaitersInvestmentRequestRequest) bool {
						ignoreTypes := cmpopts.IgnoreTypes(map[string]int{})
						result := cmp.Equal(request, adomain.InvestmentWaitersInvestmentRequestRequest{
							InvestmentWaitersId: lo.ToPtr(int64(128931)),
							JWTUser: jwthelper.User{
								Id:    int64(12378123),
								Name:  "Kei Watanabe",
								Email: "kwatanabe@miralita.co.jp",
							},
							InvestmentWaitersInvestmentRequestBody: adomain.InvestmentWaitersInvestmentRequestBody{
								Qty: 5,
							},
						}, ignoreTypes)

						return result
					})).Return(adomain.InvestmentWaitersInvestmentRequestResponse{}, nil)

				return usecase
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := investmentWaitersHttpDelivery{
				investmentWaitersUsecase: tt.investmentWaitersUsecase(),
				logger:                   factory.NewLoggerFactory(),
				extractor:                jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.InvestmentRequest(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "InvestmentRequest() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func TestInvestmentWaitersHttpDelivery_Propose(t *testing.T) {
	tests := []struct {
		name                     string
		echoCtx                  func(e *echo.Echo) echo.Context
		investmentWaitersUsecase func() *mocks.InvestmentWaitersUsecase
		wantErr                  bool
	}{
		{
			name: "POST api/v1/investment-waiters/:id/propose => 호출성공",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path: "/api/v1/investment-waiters/:id/propose",
				Params: map[string]string{
					"id": "128931",
				},
				Body:   `{"qty" : 5}`,
				Method: http.MethodPost,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Claims: map[string]interface{}{
					"user":  "12378123",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_USER"},
				},
			}),
			investmentWaitersUsecase: func() *mocks.InvestmentWaitersUsecase {
				usecase := &mocks.InvestmentWaitersUsecase{}
				usecase.On("Propose", mock.Anything, mock.MatchedBy(
					func(request adomain.InvestmentWaitersProposeRequest) bool {
						ignoreTypes := cmpopts.IgnoreTypes(map[string]int{})
						result := cmp.Equal(request, adomain.InvestmentWaitersProposeRequest{
							InvestmentWaitersId: lo.ToPtr(int64(128931)),
							JWTUser: jwthelper.User{
								Id:    int64(12378123),
								Name:  "홍길동",
								Email: "gildong@email.com",
							},
							InvestmentWaitersProposeBody: adomain.InvestmentWaitersProposeBody{
								Qty: 5,
							},
						}, ignoreTypes)

						return result
					})).Return(adomain.InvestmentWaitersProposeResponse{}, nil)

				return usecase
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := investmentWaitersHttpDelivery{
				investmentWaitersUsecase: tt.investmentWaitersUsecase(),
				logger:                   factory.NewLoggerFactory(),
				extractor:                jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.Propose(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "Propose() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func TestInvestmentWaitersHttpDelivery_Cancel(t *testing.T) {
	unexpectedErr := errors.New("unexpected error")
	tests := []struct {
		name        string
		id          string
		usecaseErr  error
		callUsecase bool
		wantStatus  int
		wantErrorIs error
	}{
		{
			name:        "キャンセル待ち申込または投資取引をキャンセルできる",
			id:          "128931",
			callUsecase: true,
		},
		{
			name:        "対象が存在しない場合は404を返す",
			id:          "128931",
			usecaseErr:  cdomain.ErrItemNotFound,
			callUsecase: true,
			wantStatus:  http.StatusNotFound,
		},
		{
			name:        "キャンセル対象外の状態は400を返す",
			id:          "128931",
			usecaseErr:  cdomain.ErrBadRequest,
			callUsecase: true,
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "キャンセル済みなど競合する状態は400を返す",
			id:          "128931",
			usecaseErr:  cdomain.ErrItemConflict,
			callUsecase: true,
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:       "IDが数値でない場合は400を返す",
			id:         "invalid",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "IDが未指定の場合は400を返す",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:        "予期しないUsecaseエラーをそのまま返す",
			id:          "128931",
			usecaseErr:  unexpectedErr,
			callUsecase: true,
			wantErrorIs: unexpectedErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := new(mocks.InvestmentWaitersUsecase)
			if tt.callUsecase {
				usecase.On("Cancel", mock.Anything, mock.MatchedBy(func(request adomain.InvestmentWaitersCancelRequest) bool {
					return request.InvestmentWaitersId != nil &&
						*request.InvestmentWaitersId == int64(128931) &&
						request.JWTUser.Id == int64(12378123) &&
						request.JWTUser.Email == "admin@example.com" &&
						request.JWTUser.Name == "管理者"
				})).Return(adomain.InvestmentWaitersCancelResponse{}, tt.usecaseErr)
			}

			delivery := investmentWaitersHttpDelivery{
				investmentWaitersUsecase: usecase,
				logger:                   factory.NewLoggerFactory(),
				extractor:                jwthelper.NewUserExtractor(),
			}
			e := echo.New()
			e.Validator = validator.NewCustomValidator()
			params := map[string]string{}
			if tt.id != "" {
				params["id"] = tt.id
			}
			ctx := pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/investment-waiters/:id/cancel",
				Params: params,
				Method: http.MethodPost,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Claims: map[string]interface{}{
					"user":  "12378123",
					"email": "admin@example.com",
					"name":  "管理者",
					"role":  []string{"ROLE_INVESTOR"},
				},
			})(e)

			err := delivery.Cancel(ctx)

			switch {
			case tt.wantErrorIs != nil:
				assert.ErrorIs(t, err, tt.wantErrorIs)
			case tt.wantStatus != 0:
				var apiErr *factory.Error
				if assert.ErrorAs(t, err, &apiErr) {
					assert.Equal(t, tt.wantStatus, apiErr.Status())
				}
			default:
				assert.NoError(t, err)
				assert.Equal(t, http.StatusOK, ctx.Response().Status)
			}
			usecase.AssertExpectations(t)
		})
	}
}
