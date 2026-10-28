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
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const withdrawalApi = "/api/v1/withdrawals"

func Test_WithdrawalHttpDelivery_GetAll(t *testing.T) {
	tests := []struct {
		name              string
		echoCtx           func(e *echo.Echo) echo.Context
		withdrawalUsecase func() *mocks.WithdrawalUsecase
		wantErr           bool
	}{
		{
			name: "GET /v1/withdrawals",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   withdrawalApi,
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
					"sortby":             {"id"},
					"order":              {"desc"},
					"skip_count":         {"0"},
					"max_result_count":   {"10"},
					"investor_id":        {"1"},
					"request_start_date": {"2025-01-01 00:00:00"},
					"request_end_date":   {"2025-12-31 23:59:59"},
				},
			}),
			withdrawalUsecase: func() *mocks.WithdrawalUsecase {
				m := &mocks.WithdrawalUsecase{}
				m.On("GetAll", mock.Anything, adomain.WithdrawalGetAllRequest{
					JWTUser: jwthelper.User{
						Id:    12378123,
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{"ROLE_ADMIN": 1, "ROLE_PAYMENT": 1},
					},
					WithdrawalGetAllQueryParam: adomain.WithdrawalGetAllQueryParam{
						SearchInput: cdomain.SearchInput{
							Sortby:         "id",
							Order:          "desc",
							SkipCount:      0,
							MaxResultCount: 10,
						},
						InvestorId:       lo.ToPtr("1"),
						RequestStartDate: lo.ToPtr("2025-01-01 00:00:00"),
						RequestEndDate:   lo.ToPtr("2025-12-31 23:59:59"),
					},
				}).Return(adomain.WithdrawalGetAllResponse{
					TotalCount:           1,
					TotalIncompleteCount: 1,
					Items: []adomain.WithdrawalItem{
						mockWithdrawalItem1,
					},
				}, nil)
				return m
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := withdrawalHttpDelivery{
				extractor:         jwthelper.NewUserExtractor(),
				withdrawalUsecase: tt.withdrawalUsecase(),
			}

			e := echo.New()
			err := delivery.GetAll(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_WithdrawalHttpDelivery_Create(t *testing.T) {
	tests := []struct {
		name              string
		echoCtx           func(e *echo.Echo) echo.Context
		withdrawalUsecase func() *mocks.WithdrawalUsecase
		wantStatus        int
		wantErr           bool
	}{
		{
			name: "POST /v1/withdrawals",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   withdrawalApi,
				Method: http.MethodPost,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Claims: map[string]interface{}{
					"user":  "12378123",
					"email": "gildong@email.com",
					"name":  "山田太郎",
					"role":  []string{"ROLE_ADMIN", "ROLE_PAYMENT"},
				},
				Body: `{
					"user_id": 1,
					"amount": 1000000
				}`,
			}),
			withdrawalUsecase: func() *mocks.WithdrawalUsecase {
				m := &mocks.WithdrawalUsecase{}
				m.On("Create", mock.Anything, mock.MatchedBy(func(request adomain.WithdrawalCreateRequest) bool {
					return request.JWTUser.Id == 12378123 &&
						request.JWTUser.Email == "gildong@email.com" &&
						request.JWTUser.Name == "山田太郎" &&
						request.UserId == 1 &&
						request.Amount == 1000000
				})).Return(mockWithdrawalCreateResponse1, nil)
				return m
			},
			wantStatus: http.StatusCreated,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := withdrawalHttpDelivery{
				extractor:         jwthelper.NewUserExtractor(),
				withdrawalUsecase: tt.withdrawalUsecase(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			ctx := tt.echoCtx(e)
			err := delivery.Create(ctx)
			assert.Equalf(t, tt.wantErr, err != nil, "error = %v, wantErr %v", err, tt.wantErr)
			assert.Equal(t, tt.wantStatus, ctx.Response().Status)
		})
	}
}

func Test_WithdrawalHttpDelivery_GetOne(t *testing.T) {
	tests := []struct {
		name              string
		echoCtx           func(e *echo.Echo) echo.Context
		withdrawalUsecase func() *mocks.WithdrawalUsecase
		wantErr           bool
	}{
		{
			name: "GET /v1/withdrawals/:id",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   withdrawalApi + "/:id",
				Method: http.MethodGet,
				Params: map[string]string{"id": "1"},
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
			}),
			withdrawalUsecase: func() *mocks.WithdrawalUsecase {
				m := &mocks.WithdrawalUsecase{}
				m.On("GetOne", mock.Anything, adomain.WithdrawalGetOneRequest{
					JWTUser: jwthelper.User{
						Id:    12378123,
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{"ROLE_ADMIN": 1, "ROLE_PAYMENT": 1},
					},
					Id: 1,
				}).Return(adomain.WithdrawalGetOneResponse{
					WithdrawalDetailItem: mockWithdrawalDetailItem1,
					TransferLogs:         []adomain.WithdrawalTransferItem{},
				}, nil)
				return m
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := withdrawalHttpDelivery{
				extractor:         jwthelper.NewUserExtractor(),
				withdrawalUsecase: tt.withdrawalUsecase(),
			}

			e := echo.New()
			err := delivery.GetOne(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_WithdrawalHttpDelivery_UpdateStatus(t *testing.T) {
	tests := []struct {
		name              string
		echoCtx           func(e *echo.Echo) echo.Context
		withdrawalUsecase func() *mocks.WithdrawalUsecase
		wantErr           bool
	}{
		{
			name: "PUT /v1/withdrawals/:id",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   withdrawalApi + "/:id",
				Method: http.MethodPut,
				Params: map[string]string{"id": "1"},
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
				Body: `{
					"status": "COMPLETE"	
				}`,
			}),
			withdrawalUsecase: func() *mocks.WithdrawalUsecase {
				m := &mocks.WithdrawalUsecase{}
				mockWithdrawalDetailItem1.Status = domain.WITHDRAWAL_STATUS_COMPLETE
				m.On("UpdateStatus", mock.Anything, adomain.WithdrawalUpdateStatusRequest{
					JWTUser: jwthelper.User{
						Id:    12378123,
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{"ROLE_ADMIN": 1, "ROLE_PAYMENT": 1},
					},
					Id:     1,
					Status: domain.WITHDRAWAL_STATUS_COMPLETE,
				}).Return(adomain.WithdrawalUpdateStatusResponse{
					WithdrawalDetailItem: mockWithdrawalDetailItem1,
				}, nil)
				return m
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := withdrawalHttpDelivery{
				extractor:         jwthelper.NewUserExtractor(),
				withdrawalUsecase: tt.withdrawalUsecase(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.UpdateStatus(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_WithdrawalHttpDelivery_RequestTransfer(t *testing.T) {
	tests := []struct {
		name              string
		echoCtx           func(e *echo.Echo) echo.Context
		withdrawalUsecase func() *mocks.WithdrawalUsecase
		wantErr           bool
	}{
		{
			name: "POST /v1/withdrawals/:id/transfer",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   withdrawalApi + "/:id/transfer",
				Method: http.MethodPost,
				Params: map[string]string{"id": "1"},
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
			}),
			withdrawalUsecase: func() *mocks.WithdrawalUsecase {
				m := &mocks.WithdrawalUsecase{}
				m.On("RequestTransfer", mock.Anything, adomain.WithdrawalRequestTransferRequest{
					JWTUser: jwthelper.User{
						Id:    12378123,
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{"ROLE_ADMIN": 1, "ROLE_PAYMENT": 1},
					},
					Id: 1,
				}).Return(adomain.WithdrawalRequestTransferResponse{
					Id:            1,
					TransactionId: "transaction_id_1",
					WithdrawalId:  1,
				}, nil)
				return m
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := withdrawalHttpDelivery{
				extractor:         jwthelper.NewUserExtractor(),
				withdrawalUsecase: tt.withdrawalUsecase(),
			}

			e := echo.New()

			err := delivery.RequestTransfer(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}
