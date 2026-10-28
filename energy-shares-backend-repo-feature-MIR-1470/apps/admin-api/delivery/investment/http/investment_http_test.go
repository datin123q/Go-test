package http

import (
	adomain "github.com/illuminarean/energy-share-core-api/domain"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/validator"

	"net/http"
	"testing"

	"github.com/illuminarean/energy-share-core-api/controllers"

	"github.com/pangpanglabs/echoswagger/v2"

	"github.com/samber/lo"

	"github.com/illuminarean/energy-shares-backend-repo/domain"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	pkgtest "github.com/illuminarean/energy-share-core-api/delivery/testing"
	"github.com/illuminarean/energy-share-core-api/domain/mocks"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

func Test_investmentHttpDelivery_GetAll(t *testing.T) {
	tests := []struct {
		name              string
		echoCtx           func(e *echo.Echo) echo.Context
		investmentUsecase func() *mocks.InvestmentUsecase
		wantErr           bool
	}{
		{
			name: "GET api/v1/investments 호출성공",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/investments",
				Method: http.MethodGet,
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
			investmentUsecase: func() *mocks.InvestmentUsecase {
				usecase := &mocks.InvestmentUsecase{}
				usecase.On("GetAll", mock.Anything, mock.MatchedBy(func(request adomain.InvestmentGetAllRequest) bool {
					ignoreTypes := cmpopts.IgnoreTypes(map[string]int{})
					result := cmp.Equal(request, adomain.InvestmentGetAllRequest{
						JWTUser: jwthelper.User{
							Id:    12378123,
							Name:  "홍길동",
							Email: "gildong@email.com",
						},
						ClientIp: "192.0.2.1",
						SearchInput: controllers.SearchInput{
							MaxResultCount: 30,
						},
					}, ignoreTypes)

					return result
				})).Return(adomain.InvestmentGetAllResponse{}, nil)

				return usecase
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()

			delivery := New(
				echoswagger.NewNop(e),
				tt.investmentUsecase(),
				factory.NewLoggerFactory(),
				jwthelper.NewUserExtractor(),
			)

			e.Validator = validator.NewCustomValidator()

			err := delivery.GetAll(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "GetAll() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_investmentHttpDelivery_GetOne(t *testing.T) {
	tests := []struct {
		name              string
		echoCtx           func(e *echo.Echo) echo.Context
		investmentUsecase func() *mocks.InvestmentUsecase
		wantErr           bool
	}{
		{
			name: "GET api/v1/investments:(investment_id=1283831) 호출성공",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/investments",
				Method: http.MethodGet,
				Params: map[string]string{"id": "1283831"},
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Claims: map[string]interface{}{
					"user":  "12378123",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_USER", "ROLE_PAYMENT"},
				},
			}),
			investmentUsecase: func() *mocks.InvestmentUsecase {
				usecase := &mocks.InvestmentUsecase{}
				usecase.On("GetOne", mock.Anything, mock.MatchedBy(func(request adomain.InvestmentGetOneRequest) bool {
					ignoreTypes := cmpopts.IgnoreTypes(map[string]int{})

					return cmp.Equal(request, adomain.InvestmentGetOneRequest{
						JWTUser: jwthelper.User{
							Id:    12378123,
							Name:  "홍길동",
							Email: "gildong@email.com",
						},
						ClientIp:     "192.0.2.1",
						InvestmentId: "1283831",
					}, ignoreTypes)
				})).Return(adomain.InvestmentGetOneResponse{}, nil)

				return usecase
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := investmentHttpDelivery{
				investmentUsecase: tt.investmentUsecase(),
				logger:            factory.NewLoggerFactory(),
				extractor:         jwthelper.NewUserExtractor(),
			}

			err := delivery.GetOne(tt.echoCtx(echo.New()))
			assert.Equalf(t, tt.wantErr, err != nil, "GetOne() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func TestInvestmentHttpDelivery_UpdateStatus(t *testing.T) {
	tests := []struct {
		name              string
		echoCtx           func(e *echo.Echo) echo.Context
		investmentUsecase func() *mocks.InvestmentUsecase
		wantErr           bool
	}{
		{
			name: "PUT api/v1/investments/:id/invest-status : status(REQUEST_INVEST) payload 전달=> 호출성공",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path: "/api/v1/investments/:id/invest-status",
				Params: map[string]string{
					"id": "128931",
				},
				Method: http.MethodPut,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Body: `{"order_status":"REQUEST_INVEST"}`,
				Claims: map[string]interface{}{
					"user":  "12378123",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_USER"},
				},
			}),
			investmentUsecase: func() *mocks.InvestmentUsecase {
				usecase := &mocks.InvestmentUsecase{}
				usecase.On("UpdateStatus", mock.Anything, mock.MatchedBy(
					func(request adomain.InvestmentUpdateStatusRequest) bool {
						ignoreTypes := cmpopts.IgnoreTypes(map[string]int{})
						result := cmp.Equal(request, adomain.InvestmentUpdateStatusRequest{
							InvestmentId: "128931",
							JWTUser: jwthelper.User{
								Id:    int64(12378123),
								Name:  "홍길동",
								Email: "gildong@email.com",
							},
							InvestmentUpdateStatusBody: adomain.InvestmentUpdateStatusBody{
								InvestStatus: domain.REQUEST_INVEST,
							},
						}, ignoreTypes)

						return result
					})).Return(adomain.InvestmentUpdateStatusResponse{}, nil)

				return usecase
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := investmentHttpDelivery{
				investmentUsecase: tt.investmentUsecase(),
				logger:            factory.NewLoggerFactory(),
				extractor:         jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.UpdateStatus(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "UpdateStatus() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

// investmentUserFacingErrorForTest はMIR-1295のテスト専用。usecase層の
// userFacingError(message部分をそのまま利用者へ見せてよいエラー)と同じ
// Error()/Unwrap()の形を模して、errors.Isによる判別とmessageの伝播を
// 検証する。
type investmentUserFacingErrorForTest struct {
	cause   error
	message string
}

func (e *investmentUserFacingErrorForTest) Error() string { return e.message }
func (e *investmentUserFacingErrorForTest) Unwrap() error { return e.cause }

func TestInvestmentHttpDelivery_UpdateStatus_ErrorMapping(t *testing.T) {
	tests := []struct {
		name        string
		usecaseErr  error
		wantStatus  int
		wantCode    int
		wantMessage string
	}{
		{name: "不正なステータス遷移はパラメータエラーを返す", usecaseErr: cdomain.ErrBadRequest, wantStatus: http.StatusBadRequest},
		{name: "同一ステータスへの遷移はパラメータエラーを返す", usecaseErr: cdomain.ErrItemConflict, wantStatus: http.StatusBadRequest},
		{
			// MIR-1295: 許可されない遷移は専用のエラーコード(10108)を返し、
			// usecaseが組み立てた日本語の案内文がmessageフィールドに
			// そのまま載ることを確認する。
			name: "許可されない遷移は日本語の案内文をmessageに載せて返す",
			usecaseErr: &investmentUserFacingErrorForTest{
				cause:   cdomain.ErrCannotManuallyChangeInvestStatus,
				message: "この出資は運用中のため、取消には変更できません",
			},
			wantStatus:  http.StatusBadRequest,
			wantCode:    10108,
			wantMessage: "この出資は運用中のため、取消には変更できません",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			investmentUsecase := new(mocks.InvestmentUsecase)
			investmentUsecase.On("UpdateStatus", mock.Anything, mock.Anything).
				Return(adomain.InvestmentUpdateStatusResponse{}, tt.usecaseErr)

			delivery := investmentHttpDelivery{
				investmentUsecase: investmentUsecase,
				logger:            factory.NewLoggerFactory(),
				extractor:         jwthelper.NewUserExtractor(),
			}
			e := echo.New()
			e.Validator = validator.NewCustomValidator()
			ctx := pkgtest.RequestContext(pkgtest.Request{
				Path: "/api/v1/investments/:id/invest-status",
				Params: map[string]string{
					"id": "SHARE-1",
				},
				Method: http.MethodPut,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Body: `{"order_status":"CANCEL_INVEST"}`,
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "admin@example.com",
					"name":  "管理者",
					"role":  []string{"ROLE_INVESTOR"},
				},
			})(e)

			err := delivery.UpdateStatus(ctx)

			var apiErr *factory.Error
			if assert.ErrorAs(t, err, &apiErr) {
				assert.Equal(t, tt.wantStatus, apiErr.Status())
				if tt.wantCode != 0 {
					assert.Equal(t, tt.wantCode, apiErr.Code)
					assert.Equal(t, tt.wantMessage, apiErr.Message)
				}
			}
		})
	}
}

func Test_investmentHttpDelivery_DownloadHistories(t *testing.T) {
	tests := []struct {
		name              string
		echoCtx           func(e *echo.Echo) echo.Context
		investmentUsecase func() *mocks.InvestmentUsecase
		wantErr           bool
	}{
		{
			name: "GET api/v1/investments/download => 호출성공",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/investments/download",
				Method: http.MethodGet,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Claims: map[string]interface{}{
					"user":  "12378123",
					"email": "gildong@email.com",
					"role":  []string{"ROLE_USER"},
				},
			}),
			investmentUsecase: func() *mocks.InvestmentUsecase {
				usecase := &mocks.InvestmentUsecase{}
				usecase.On("DownloadHistories", mock.Anything, mock.MatchedBy(func(request adomain.InvestmentDownloadHistoriesRequest) bool {
					return cmp.Equal(request, adomain.InvestmentDownloadHistoriesRequest{})
				})).Return(adomain.InvestmentDownloadHistoriesResponse{}, nil)
				return usecase
			},
		},
		{
			name: "GET api/v1/investments/download : invest_status(REQUEST_INVEST), investor_name(솔라리언트) queryparam => 호출성공",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/investments/download",
				Method: http.MethodGet,
				QueryParams: map[string][]string{
					"invest_status": {"REQUEST_INVEST"},
					"investor_name": {"솔라리언트"},
				},
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Claims: map[string]interface{}{
					"user":  "12378123",
					"email": "gildong@email.com",
					"role":  []string{"ROLE_USER"},
				},
			}),
			investmentUsecase: func() *mocks.InvestmentUsecase {
				usecase := &mocks.InvestmentUsecase{}
				usecase.On("DownloadHistories", mock.Anything, mock.MatchedBy(func(request adomain.InvestmentDownloadHistoriesRequest) bool {
					return cmp.Equal(request, adomain.InvestmentDownloadHistoriesRequest{
						InvestmentDownloadHistoriesQueryParam: adomain.InvestmentDownloadHistoriesQueryParam{
							InvestStatus: lo.ToPtr(string(domain.REQUEST_INVEST)),
							InvestorName: lo.ToPtr("솔라리언트"),
						},
					})
				})).Return(adomain.InvestmentDownloadHistoriesResponse{}, nil)
				return usecase
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := investmentHttpDelivery{
				investmentUsecase: tt.investmentUsecase(),
				logger:            factory.NewLoggerFactory(),
				extractor:         jwthelper.NewUserExtractor(),
			}

			err := delivery.DownloadHistories(tt.echoCtx(echo.New()))
			assert.Equalf(t, tt.wantErr, err != nil, "DownloadHistories() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func TestInvestmentHttpDelivery_GetAllTransactionInfoHistories(t *testing.T) {
	tests := []struct {
		name              string
		echoCtx           func(e *echo.Echo) echo.Context
		investmentUsecase func() *mocks.InvestmentUsecase
		wantErr           bool
	}{
		{
			name: "GET api/v1/investments/:id/transactions => 호출성공",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/investments/:id/transactions",
				Params: map[string]string{"id": "1278312"},
				Method: http.MethodGet,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
			}),
			investmentUsecase: func() *mocks.InvestmentUsecase {
				usecase := &mocks.InvestmentUsecase{}
				usecase.On("GetAllTransactionInfoHistories", mock.Anything, mock.MatchedBy(
					func(request adomain.InvestmentGetAllTransactionInfoHistoriesRequest) bool {
						return cmp.Equal(request, adomain.InvestmentGetAllTransactionInfoHistoriesRequest{
							InvestmentId: "1278312",
						})
					})).Return(adomain.InvestmentGetAllTransactionInfoHistoriesResponse{}, nil)
				return usecase
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := investmentHttpDelivery{
				investmentUsecase: tt.investmentUsecase(),
				logger:            factory.NewLoggerFactory(),
				extractor:         jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.GetAllTransactionInfoHistories(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "GetAllTransactionInfoHistories() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}
