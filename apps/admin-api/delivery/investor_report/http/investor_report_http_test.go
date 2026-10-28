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

func Test_InvestorReportDelivery_GetAllPaymentStatement(t *testing.T) {
	tests := []struct {
		name                  string
		echoCtx               func(e *echo.Echo) echo.Context
		investorReportUsecase func() *mocks.InvestorReportUsecase
		wantErr               bool
	}{
		{
			name: "GET /v1/investor_report/payments",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/investor_report/payments",
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
					"skip_count":       {"0"},
				},
			}),
			investorReportUsecase: func() *mocks.InvestorReportUsecase {
				m := &mocks.InvestorReportUsecase{}
				m.On("GetAllPaymentStatement", mock.Anything, adomain.InvestorReportGetAllRequest{
					JWTUser: jwthelper.User{
						Id:    12378123,
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{"ROLE_ADMIN": 1, "ROLE_OPERATOR": 1},
					},
					ClientIp: "192.0.2.1",
					SearchInput: cdomain.SearchInput{
						SkipCount:      0,
						MaxResultCount: 10,
					},
				}).Return(adomain.InvestorReportGetAllResponse{
					TotalCount: 1,
					Items:      []domain.ReportOutput{},
				}, nil)
				return m
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := investorReportHttpDelivery{
				extractor:             jwthelper.NewUserExtractor(),
				investorReportUsecase: tt.investorReportUsecase(),
			}
			e := echo.New()
			e.Validator = validator.NewCustomValidator()
			err := delivery.GetAllPaymentStatement(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "GetAllPaymentStatement() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_InvestorReportDelivery_GetAllTransactionBalanceReport(t *testing.T) {
	tests := []struct {
		name                  string
		echoCtx               func(e *echo.Echo) echo.Context
		investorReportUsecase func() *mocks.InvestorReportUsecase
		wantErr               bool
	}{
		{
			name: "GET /v1/investor_report/transaction-balances",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/investor_report/transaction-balances",
				Method: http.MethodGet,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Claims: map[string]interface{}{
					"user":  "12378123",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_ADMIN", "ROLE_REPORT"},
				},
				QueryParams: map[string][]string{
					"max_result_count": {"10"},
					"skip_count":       {"0"},
				},
			}),
			investorReportUsecase: func() *mocks.InvestorReportUsecase {
				m := &mocks.InvestorReportUsecase{}
				m.On("GetAllTransactionBalanceReport", mock.Anything, adomain.InvestorReportGetAllRequest{
					JWTUser: jwthelper.User{
						Id:    12378123,
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{"ROLE_ADMIN": 1, "ROLE_REPORT": 1},
					},
					ClientIp: "192.0.2.1",
					SearchInput: cdomain.SearchInput{
						SkipCount:      0,
						MaxResultCount: 10,
					},
				}).Return(adomain.InvestorReportGetAllResponse{
					TotalCount: 1,
					Items:      []domain.ReportOutput{},
				}, nil)
				return m
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := investorReportHttpDelivery{
				extractor:             jwthelper.NewUserExtractor(),
				investorReportUsecase: tt.investorReportUsecase(),
			}
			e := echo.New()
			e.Validator = validator.NewCustomValidator()
			err := delivery.GetAllTransactionBalanceReport(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "GetAllTransactionBalanceReport() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_InvestorReportDelivery_OpenPaymentStatement(t *testing.T) {
	tests := []struct {
		name                  string
		echoCtx               func(e *echo.Echo) echo.Context
		investorReportUsecase func() *mocks.InvestorReportUsecase
		wantErr               bool
	}{
		{
			name: "PUT /v1/investor_report/payments",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/investor_report/payments",
				Method: http.MethodPut,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Claims: map[string]interface{}{
					"user":  "12378123",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_ADMIN", "ROLE_REPORT"},
				},
				Body: `{
					"year": "2024",
					"quarter": "QUARTER_FIRST"
				}`,
			}),
			investorReportUsecase: func() *mocks.InvestorReportUsecase {
				m := &mocks.InvestorReportUsecase{}
				m.On("OpenPaymentStatement", mock.Anything, adomain.InvestorReportOpenPaymentStatementRequest{
					JWTUser: jwthelper.User{
						Id:    12378123,
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{"ROLE_ADMIN": 1, "ROLE_REPORT": 1},
					},
					ClientIp: "192.0.2.1",
					Year:     "2024",
				}).Return(nil)
				return m
			},
			wantErr: false,
		},
		{
			name: "PUT /v1/investor_report/payments - validation error - 실패: year가 연도가 아님(숫자 4자리)",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/investor_report/payments",
				Method: http.MethodPut,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Claims: map[string]interface{}{
					"user":  "12378123",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_ADMIN", "ROLE_REPORT"},
				},
				Body: `{
					"year": "20241",
				}`,
			}),
			investorReportUsecase: func() *mocks.InvestorReportUsecase {
				m := &mocks.InvestorReportUsecase{}
				m.On("OpenPaymentStatement", mock.Anything, adomain.InvestorReportOpenPaymentStatementRequest{
					JWTUser: jwthelper.User{
						Id:    12378123,
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{"ROLE_ADMIN": 1, "ROLE_REPORT": 1},
					},
					ClientIp: "192.0.2.1",
					Year:     "20241",
				}).Return(nil)
				return m
			},
			wantErr: true,
		},
		{
			name: "PUT /v1/investor_report/payments - validation error - 실패: year가 숫자가 아님",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/investor_report/payments",
				Method: http.MethodPut,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Claims: map[string]interface{}{
					"user":  "12378123",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_ADMIN", "ROLE_REPORT"},
				},
				Body: `{
					"year": "not-a-number",
				}`,
			}),
			investorReportUsecase: func() *mocks.InvestorReportUsecase {
				m := &mocks.InvestorReportUsecase{}
				m.On("OpenPaymentStatement", mock.Anything, adomain.InvestorReportOpenPaymentStatementRequest{
					JWTUser: jwthelper.User{
						Id:    12378123,
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{"ROLE_ADMIN": 1, "ROLE_REPORT": 1},
					},
					ClientIp: "192.0.2.1",
					Year:     "not-a-number",
				}).Return(nil)
				return m
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := investorReportHttpDelivery{
				extractor:             jwthelper.NewUserExtractor(),
				investorReportUsecase: tt.investorReportUsecase(),
			}
			e := echo.New()
			e.Validator = validator.NewCustomValidator()
			err := delivery.OpenPaymentStatement(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "OpenPaymentStatement() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_InvestorReportDelivery_OpenTransactionBalanceReport(t *testing.T) {
	tests := []struct {
		name                  string
		echoCtx               func(e *echo.Echo) echo.Context
		investorReportUsecase func() *mocks.InvestorReportUsecase
		wantErr               bool
	}{
		{
			name: "PUT /v1/investor_report/transaction-balances",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/investor_report/transaction-balances",
				Method: http.MethodPut,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Claims: map[string]interface{}{
					"user":  "12378123",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_ADMIN", "ROLE_REPORT"},
				},
				Body: `{
					"year": "2024",
					"quarter": "QUARTER_FIRST"
				}`,
			}),
			investorReportUsecase: func() *mocks.InvestorReportUsecase {
				m := &mocks.InvestorReportUsecase{}
				m.On("OpenTransactionBalanceReport", mock.Anything, adomain.InvestorReportOpenTransactionBalanceRequest{
					JWTUser: jwthelper.User{
						Id:    12378123,
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{"ROLE_ADMIN": 1, "ROLE_REPORT": 1},
					},
					ClientIp: "192.0.2.1",
					Year:     "2024",
					Quarter:  "QUARTER_FIRST",
				}).Return(nil)
				return m
			},
			wantErr: false,
		},
		{
			name: "PUT /v1/investor_report/transaction-balances - 실패: year가 연도가 아님(숫자 4자리)",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/investor_report/transaction-balances",
				Method: http.MethodPut,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Claims: map[string]interface{}{
					"user":  "12378123",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_ADMIN", "ROLE_REPORT"},
				},
				Body: `{
					"year": "20241",
					"quarter": "QUARTER_FIRST"
				}`,
			}),
			investorReportUsecase: func() *mocks.InvestorReportUsecase {
				m := &mocks.InvestorReportUsecase{}
				m.On("OpenTransactionBalanceReport", mock.Anything, adomain.InvestorReportOpenTransactionBalanceRequest{
					JWTUser: jwthelper.User{
						Id:    12378123,
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{"ROLE_ADMIN": 1, "ROLE_REPORT": 1},
					},
					ClientIp: "192.0.2.1",
					Year:     "2024",
					Quarter:  "QUARTER_FIRST",
				}).Return(nil)
				return m
			},
			wantErr: true,
		},
		{
			name: "PUT /v1/investor_report/transaction-balances - 실패: quarter가 기대값이 아님",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/investor_report/transaction-balances",
				Method: http.MethodPut,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Claims: map[string]interface{}{
					"user":  "12378123",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_ADMIN", "ROLE_REPORT"},
				},
				Body: `{
					"year": "2024",
					"quarter": "QUARTER_FIFTH"
				}`,
			}),
			investorReportUsecase: func() *mocks.InvestorReportUsecase {
				m := &mocks.InvestorReportUsecase{}
				m.On("OpenTransactionBalanceReport", mock.Anything, adomain.InvestorReportOpenTransactionBalanceRequest{
					JWTUser: jwthelper.User{
						Id:    12378123,
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{"ROLE_ADMIN": 1, "ROLE_REPORT": 1},
					},
					ClientIp: "192.0.2.1",
					Year:     "2024",
					Quarter:  "QUARTER_FIRST",
				}).Return(nil)
				return m
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := investorReportHttpDelivery{
				extractor:             jwthelper.NewUserExtractor(),
				investorReportUsecase: tt.investorReportUsecase(),
			}
			e := echo.New()
			e.Validator = validator.NewCustomValidator()
			err := delivery.OpenTransactionBalanceReport(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "OpenTransactionBalanceReport() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_InvestorReportDelivery_DownloadPaymentStatement(t *testing.T) {
	tests := []struct {
		name                  string
		echoCtx               func(e *echo.Echo) echo.Context
		investorReportUsecase func() *mocks.InvestorReportUsecase
		wantErr               bool
	}{
		{
			name: "GET /v1/investor_report/payments/download",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/investor_report/payments/download",
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
					"year": {"2024"},
				},
			}),
			investorReportUsecase: func() *mocks.InvestorReportUsecase {
				m := &mocks.InvestorReportUsecase{}
				m.On("DownloadPaymentStatement", mock.Anything, adomain.DownloadPaymentStatementRequest{
					JWTUser: jwthelper.User{
						Id:    12378123,
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{"ROLE_ADMIN": 1, "ROLE_OPERATOR": 1},
					},
					ClientIp: "192.0.2.1",
					Year:     "2024",
				}).Return(adomain.DownloadPaymentStatementResponse{}, nil)
				return m
			},
			wantErr: false,
		},
		{
			name: "GET /v1/investor_report/payments/download - validation error - 실패: year가 연도가 아님(숫자 4자리): ",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/investor_report/payments/download",
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
					"year": {"20241"},
				},
			}),
			investorReportUsecase: func() *mocks.InvestorReportUsecase {
				m := &mocks.InvestorReportUsecase{}
				m.On("DownloadPaymentStatement", mock.Anything, adomain.DownloadPaymentStatementRequest{
					JWTUser: jwthelper.User{
						Id:    12378123,
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{"ROLE_ADMIN": 1, "ROLE_OPERATOR": 1},
					},
					ClientIp: "192.0.2.1",
					Year:     "20241",
				}).Return(adomain.DownloadPaymentStatementResponse{}, nil)
				return m
			},
			wantErr: true,
		},
		{
			name: "GET /v1/investor_report/payments/download - validation error - 실패: year가 숫자가 아님 ): ",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/investor_report/payments/download",
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
					"year": {"not-a-number"},
				},
			}),
			investorReportUsecase: func() *mocks.InvestorReportUsecase {
				m := &mocks.InvestorReportUsecase{}
				m.On("DownloadPaymentStatement", mock.Anything, adomain.DownloadPaymentStatementRequest{
					JWTUser: jwthelper.User{
						Id:    12378123,
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{"ROLE_ADMIN": 1, "ROLE_OPERATOR": 1},
					},
					ClientIp: "192.0.2.1",
					Year:     "not-a-number",
				}).Return(adomain.DownloadPaymentStatementResponse{}, nil)
				return m
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := investorReportHttpDelivery{
				extractor:             jwthelper.NewUserExtractor(),
				investorReportUsecase: tt.investorReportUsecase(),
			}
			e := echo.New()
			e.Validator = validator.NewCustomValidator()
			err := delivery.DownloadPaymentStatement(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "DownloadPaymentStatement() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_InvestorReportDelivery_DownloadTransactionBalanceReport(t *testing.T) {
	tests := []struct {
		name                  string
		echoCtx               func(e *echo.Echo) echo.Context
		investorReportUsecase func() *mocks.InvestorReportUsecase
		wantErr               bool
	}{
		{
			name: "GET /v1/investor_report/transaction-balances/download",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/investor_report/transaction-balances/download",
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
					"year":    {"2024"},
					"quarter": {"QUARTER_FIRST"},
				},
			}),
			investorReportUsecase: func() *mocks.InvestorReportUsecase {
				m := &mocks.InvestorReportUsecase{}
				m.On("DownloadTransactionBalanceReport", mock.Anything, adomain.DownloadTransactionBalanceReportRequest{
					JWTUser: jwthelper.User{
						Id:    12378123,
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{"ROLE_ADMIN": 1, "ROLE_OPERATOR": 1},
					},
					ClientIp: "192.0.2.1",
					Year:     "2024",
					Quarter:  "QUARTER_FIRST",
				}).Return(adomain.DownloadTransactionBalanceReportResponse{}, nil)
				return m
			},
			wantErr: false,
		},
		{
			name: "GET /v1/investor_report/transaction-balances/download - validation error - 실패: year가 연도가 아님(숫자 4자리)",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/investor_report/transaction-balances/download",
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
					"year":    {"20241"},
					"quarter": {"QUARTER_FIFTH"},
				},
			}),
			investorReportUsecase: func() *mocks.InvestorReportUsecase {
				m := &mocks.InvestorReportUsecase{}
				m.On("DownloadTransactionBalanceReport", mock.Anything, adomain.DownloadTransactionBalanceReportRequest{
					JWTUser: jwthelper.User{
						Id:    12378123,
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{"ROLE_ADMIN": 1, "ROLE_OPERATOR": 1},
					},
					ClientIp: "192.0.2.1",
					Year:     "20241",
					Quarter:  "QUARTER_FIFTH",
				}).Return(adomain.DownloadTransactionBalanceReportResponse{}, nil)
				return m
			},
			wantErr: true,
		},
		{
			name: "GET /v1/investor_report/transaction-balances/download - validation error - quarter: QUARTER_FIFTH",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/investor_report/transaction-balances/download",
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
					"year":    {"2024"},
					"quarter": {"QUARTER_FIFTH"},
				},
			}),
			investorReportUsecase: func() *mocks.InvestorReportUsecase {
				m := &mocks.InvestorReportUsecase{}
				m.On("DownloadTransactionBalanceReport", mock.Anything, adomain.DownloadTransactionBalanceReportRequest{
					JWTUser: jwthelper.User{
						Id:    12378123,
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{"ROLE_ADMIN": 1, "ROLE_OPERATOR": 1},
					},
					ClientIp: "192.0.2.1",
					Year:     "2024",
					Quarter:  "QUARTER_FIFTH",
				}).Return(adomain.DownloadTransactionBalanceReportResponse{}, nil)
				return m
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := investorReportHttpDelivery{
				extractor:             jwthelper.NewUserExtractor(),
				investorReportUsecase: tt.investorReportUsecase(),
			}
			e := echo.New()
			e.Validator = validator.NewCustomValidator()
			err := delivery.DownloadTransactionBalanceReport(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "DownloadTransactionBalanceReport() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}
