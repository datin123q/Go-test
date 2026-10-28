package http

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/illuminarean/energy-share-core-api/controllers"
	pkgtest "github.com/illuminarean/energy-share-core-api/delivery/testing"
	adomain "github.com/illuminarean/energy-share-core-api/domain"
	"github.com/illuminarean/energy-share-core-api/domain/mocks"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/validator"
)

func Test_repaymentDividendHttpDelivery_ExecuteInvestment(t *testing.T) {

	tests := []struct {
		name                     string
		echoCtx                  func(e *echo.Echo) echo.Context
		repaymentDividendUsecase func() *mocks.RepaymentDividendUsecase
		wantErr                  bool
	}{
		{
			name: "POST api/v1/project/:id/execute-investment 호출성공",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path: "/api/v1/project/:id/execute-investment",
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
					"role":  []string{"ROLE_USER", "ROLE_PRODUCT"},
				},
				Body: `{"amount": 1234, "execute_date": "2022-01-01"}`,
			}),
			repaymentDividendUsecase: func() *mocks.RepaymentDividendUsecase {
				projectId := int64(128931)
				usecase := &mocks.RepaymentDividendUsecase{}
				usecase.On("ExecuteInvestment", mock.Anything,
					mock.MatchedBy(func(request adomain.ExecuteInvestmentRequest) bool {
						ignoreTypes := cmpopts.IgnoreTypes(map[string]int{})
						result := cmp.Equal(request, adomain.ExecuteInvestmentRequest{
							JWTUser: jwthelper.User{
								Id:    12378123,
								Name:  "홍길동",
								Email: "gildong@email.com",
							},
							ProjectId: projectId,
							ExecuteInvestmentBody: adomain.ExecuteInvestmentBody{
								Amount:      1234,
								ExecuteDate: "2022-01-01",
							},
						}, ignoreTypes)

						return result
					})).Return(adomain.ExecuteInvestmentResponse{}, nil)

				return usecase
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			delivery := repaymentDividendHttpDelivery{
				tt.repaymentDividendUsecase(),
				factory.NewLoggerFactory(),
				jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.ExecuteInvestment(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "ExecuteInvestment() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_repaymentDividendHttpDelivery_CreateRepaymentDividendSchedules(t *testing.T) {

	tests := []struct {
		name                     string
		echoCtx                  func(e *echo.Echo) echo.Context
		repaymentDividendUsecase func() *mocks.RepaymentDividendUsecase
		wantErr                  bool
	}{
		{
			name: "POST api/v1/project/:id/repayment-dividend-schedules 호출성공",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path: "/api/v1/project/:id/repayment-dividend-schedules",
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
					"role":  []string{"ROLE_USER", "ROLE_PRODUCT"},
				},
				Body: func() string {
					data, _ := json.Marshal(adomain.CreateRepaymentDividendBody{
						RepaymentAmount: 2000,
						DividendAmount:  0,
						BrokerFee:       0,
						ReturnRate:      0,
						DecideDate:      "2024-01-11",
						ExecuteDate:     "2024-02-11",
					})
					return string(data)
				}(),
			}),
			repaymentDividendUsecase: func() *mocks.RepaymentDividendUsecase {
				usecase := &mocks.RepaymentDividendUsecase{}
				usecase.On("CreateRepaymentDividendSchedules", mock.Anything,
					mock.MatchedBy(func(request adomain.CreateRepaymentDividendRequest) bool {
						ignoreTypes := cmpopts.IgnoreTypes(map[string]int{})
						result := cmp.Equal(request, adomain.CreateRepaymentDividendRequest{
							JWTUser: jwthelper.User{
								Id:    12378123,
								Name:  "홍길동",
								Email: "gildong@email.com",
							},
							ProjectId: 128931,
							CreateRepaymentDividendBody: adomain.CreateRepaymentDividendBody{
								RepaymentAmount: 2000,
								DividendAmount:  0,
								BrokerFee:       0,
								ReturnRate:      0,
								DecideDate:      "2024-01-11",
								ExecuteDate:     "2024-02-11",
							},
						}, ignoreTypes)

						return result
					})).Return(adomain.CreateRepaymentDividendResponse{}, nil)

				return usecase
			},
		},
		{
			name: "POST api/v1/project/:id/repayment-dividend-schedules 호출 실패, DecideDate 빠짐",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path: "/api/v1/project/:id/repayment-dividend-schedules",
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
					"role":  []string{"ROLE_USER", "ROLE_PRODUCT"},
				},
				Body: func() string {
					data, _ := json.Marshal(adomain.CreateRepaymentDividendBody{
						RepaymentAmount: 2000,
						DividendAmount:  0,
						BrokerFee:       0,
						ReturnRate:      0,
						ExecuteDate:     "2024-02-11",
					})
					return string(data)
				}(),
			}),
			repaymentDividendUsecase: func() *mocks.RepaymentDividendUsecase {
				usecase := &mocks.RepaymentDividendUsecase{}
				usecase.On("CreateRepaymentDividendSchedules", mock.Anything,
					mock.MatchedBy(func(request adomain.CreateRepaymentDividendRequest) bool {
						ignoreTypes := cmpopts.IgnoreTypes(map[string]int{})
						result := cmp.Equal(request, adomain.CreateRepaymentDividendRequest{
							JWTUser: jwthelper.User{
								Id:    12378123,
								Name:  "홍길동",
								Email: "gildong@email.com",
							},
							ProjectId: 128931,
							CreateRepaymentDividendBody: adomain.CreateRepaymentDividendBody{
								RepaymentAmount: 2000,
								DividendAmount:  0,
								BrokerFee:       0,
								ReturnRate:      0,
								DecideDate:      "2024-01-11",
								ExecuteDate:     "2024-02-11",
							},
						}, ignoreTypes)

						return result
					})).Return(adomain.CreateRepaymentDividendResponse{}, nil)

				return usecase
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := repaymentDividendHttpDelivery{
				tt.repaymentDividendUsecase(),
				factory.NewLoggerFactory(),
				jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.CreateRepaymentDividendSchedules(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "CreateRepaymentDividendSchedules() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

// deliveryTestUserFacingError はMIR-1239のテスト専用。usecase層の
// userFacingError(message部分をそのまま利用者へ見せてよいエラー)と同じ
// Error()/Unwrap()の形を模して、errors.Wrapで包まれたあとも
// errors.Cause(err).Error()で日本語messageだけを取り出せることを検証する。
type deliveryTestUserFacingError struct {
	cause   error
	message string
}

func (e *deliveryTestUserFacingError) Error() string { return e.message }
func (e *deliveryTestUserFacingError) Unwrap() error { return e.cause }

// Test_repaymentDividendHttpDelivery_CreateRepaymentDividendSchedules_ErrorMapping は
// MIR-1239: usecaseが返すエラーが、10107(domain.ErrInvestmentExecutionIncomplete)
// では日本語messageつきで返り、それ以外のエラーは500(ErrorDB)のままになることを
// 検証する。
func Test_repaymentDividendHttpDelivery_CreateRepaymentDividendSchedules_ErrorMapping(t *testing.T) {
	tests := []struct {
		name           string
		usecaseErr     error
		wantErrCode    int
		wantErrStatus  int
		wantErrMessage string
	}{
		{
			// CreateRepaymentDividendSchedulesはこのエラーをerrors.Wrapで
			// 包まない(MIR-1239)ため、usecaseErrはラップされていない
			// userFacingErrorそのもの。
			name: "①出資実行の記録が無い場合、10107と日本語messageが返る",
			usecaseErr: &deliveryTestUserFacingError{
				cause:   domain.ErrInvestmentExecutionIncomplete,
				message: "出資実行が行われていません",
			},
			wantErrCode:    10107,
			wantErrStatus:  http.StatusBadRequest,
			wantErrMessage: "出資実行が行われていません",
		},
		{
			name: "②出資実行が未完了の場合、10107と日本語messageが返る",
			usecaseErr: &deliveryTestUserFacingError{
				cause:   domain.ErrInvestmentExecutionIncomplete,
				message: "出資実行が完了していません(未処理5件)。運用開始を最後まで進めてから作成してください",
			},
			wantErrCode:    10107,
			wantErrStatus:  http.StatusBadRequest,
			wantErrMessage: "出資実行が完了していません(未処理5件)。運用開始を最後まで進めてから作成してください",
		},
		{
			// ①でGetByProjectIdがErrItemNotFound以外のDBエラーを返した場合は
			// domain.ErrInvestmentExecutionIncompleteに変換されないため、
			// 500のまま。
			name:          "ErrItemNotFound以外のDBエラーは500のまま",
			usecaseErr:    errors.Wrap(errors.New("connection refused"), "failed to get investment by project id"),
			wantErrCode:   factory.ErrorDB.Code,
			wantErrStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := &mocks.RepaymentDividendUsecase{}
			usecase.On("CreateRepaymentDividendSchedules", mock.Anything, mock.Anything).
				Return(adomain.CreateRepaymentDividendResponse{}, tt.usecaseErr)

			delivery := repaymentDividendHttpDelivery{
				usecase,
				factory.NewLoggerFactory(),
				jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			echoCtx := pkgtest.RequestContext(pkgtest.Request{
				Path: "/api/v1/project/:id/repayment-dividend-schedules",
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
					"role":  []string{"ROLE_USER", "ROLE_PRODUCT"},
				},
				Body: func() string {
					data, _ := json.Marshal(adomain.CreateRepaymentDividendBody{
						RepaymentAmount: 2000,
						DecideDate:      "2024-01-11",
						ExecuteDate:     "2024-02-11",
					})
					return string(data)
				}(),
			})(e)

			err := delivery.CreateRepaymentDividendSchedules(echoCtx)
			require.Error(t, err)

			factoryErr, ok := err.(*factory.Error)
			require.True(t, ok, "expected *factory.Error, got %T: %v", err, err)
			assert.Equal(t, tt.wantErrCode, factoryErr.Code)
			assert.Equal(t, tt.wantErrStatus, factoryErr.Status())
			if tt.wantErrMessage != "" {
				assert.Equal(t, tt.wantErrMessage, factoryErr.Message)
			}
		})
	}
}

func Test_repaymentDividendHttpDelivery_GetRepaymentDividendSchedulesTable(t *testing.T) {

	tests := []struct {
		name                     string
		echoCtx                  func(e *echo.Echo) echo.Context
		repaymentDividendUsecase func() *mocks.RepaymentDividendUsecase
		wantErr                  bool
	}{
		{
			name: "GET api/v1/project/:id/repayment-dividend-schedules/table 호출성공",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path: "/api/v1/project/:id/repayment-dividend-schedules/table",
				Params: map[string]string{
					"id": "128931",
				},
				Method: http.MethodGet,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Claims: map[string]interface{}{
					"user":  "12378123",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_USER", "ROLE_PRODUCT"},
				},
			}),
			repaymentDividendUsecase: func() *mocks.RepaymentDividendUsecase {
				usecase := &mocks.RepaymentDividendUsecase{}
				usecase.On("GetRepaymentDividendSchedulesTable", mock.Anything,
					mock.MatchedBy(func(request adomain.GetRepaymentDividendTableRequest) bool {
						ignoreTypes := cmpopts.IgnoreTypes(map[string]int{})
						result := cmp.Diff(request, adomain.GetRepaymentDividendTableRequest{
							JWTUser: jwthelper.User{
								Id:    12378123,
								Name:  "홍길동",
								Email: "gildong@email.com",
							},
							ProjectId: 128931,
							SearchInput: controllers.SearchInput{
								SkipCount:      0,
								MaxResultCount: 30,
							},
						}, ignoreTypes)
						if result != "" {
							return false
						}
						return true
					})).Return(adomain.GetRepaymentDividendTableResponse{}, nil)

				return usecase
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := repaymentDividendHttpDelivery{
				tt.repaymentDividendUsecase(),
				factory.NewLoggerFactory(),
				jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.GetRepaymentDividendSchedulesTable(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "GetRepaymentDividendSchedulesTable() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_repaymentDividendHttpDelivery_GetRepaymentDividendSchedulesTableCsv(t *testing.T) {

	tests := []struct {
		name                     string
		echoCtx                  func(e *echo.Echo) echo.Context
		repaymentDividendUsecase func() *mocks.RepaymentDividendUsecase
		wantErr                  bool
	}{
		{
			name: "GET api/v1/project/:id/repayment-dividend-schedules/table-csv 호출성공",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path: "/api/v1/project/:id/repayment-dividend-schedules/table-csv",
				Params: map[string]string{
					"id": "128931",
				},
				Method: http.MethodGet,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Claims: map[string]interface{}{
					"user":  "12378123",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_USER", "ROLE_PRODUCT"},
				},
			}),
			repaymentDividendUsecase: func() *mocks.RepaymentDividendUsecase {
				usecase := &mocks.RepaymentDividendUsecase{}
				usecase.On("GetRepaymentDividendSchedulesTableCsv", mock.Anything,
					mock.MatchedBy(func(request adomain.GetRepaymentDividendTableRequest) bool {
						ignoreTypes := cmpopts.IgnoreTypes(map[string]int{})
						result := cmp.Diff(request, adomain.GetRepaymentDividendTableRequest{
							JWTUser: jwthelper.User{
								Id:    12378123,
								Name:  "홍길동",
								Email: "gildong@email.com",
							},
							ProjectId: 128931,
							SearchInput: controllers.SearchInput{
								SkipCount:      0,
								MaxResultCount: 100_000,
							},
						}, ignoreTypes)
						if result != "" {
							return false
						}
						return true
					})).Return(adomain.GetRepaymentDividendTableCsvResponse{}, nil)

				return usecase
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := repaymentDividendHttpDelivery{
				tt.repaymentDividendUsecase(),
				factory.NewLoggerFactory(),
				jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.GetRepaymentDividendSchedulesTableCsv(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "GetRepaymentDividendSchedulesTableCsv() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_repaymentDividendHttpDelivery_GetInvestmentAndReturnsInfo(t *testing.T) {

	tests := []struct {
		name                     string
		echoCtx                  func(e *echo.Echo) echo.Context
		repaymentDividendUsecase func() *mocks.RepaymentDividendUsecase
		wantErr                  bool
	}{
		{
			name: "GET api/v1/project/:id/repayment-dividend-schedules/info 호출성공",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path: "/api/v1/project/:id/repayment-dividend-schedules/info",
				Params: map[string]string{
					"id": "128931",
				},
				Method: http.MethodGet,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Claims: map[string]interface{}{
					"user":  "12378123",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_USER", "ROLE_PRODUCT"},
				},
			}),
			repaymentDividendUsecase: func() *mocks.RepaymentDividendUsecase {
				usecase := &mocks.RepaymentDividendUsecase{}
				usecase.On("GetInvestmentAndReturnsInfo", mock.Anything,
					mock.MatchedBy(func(request adomain.GetInvestmentAndReturnsInfoRequest) bool {
						ignoreTypes := cmpopts.IgnoreTypes(map[string]int{})
						ignoreFields := cmpopts.IgnoreFields(adomain.GetInvestmentAndReturnsInfoRequest{}, "ClientIp")
						result := cmp.Equal(request, adomain.GetInvestmentAndReturnsInfoRequest{
							JWTUser: jwthelper.User{
								Id:    12378123,
								Name:  "홍길동",
								Email: "gildong@email.com",
							},
							ProjectId: 128931,
						}, ignoreTypes, ignoreFields)

						return result
					})).Return(adomain.GetInvestmentAndReturnsInfoResponse{}, nil)

				return usecase
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := repaymentDividendHttpDelivery{
				tt.repaymentDividendUsecase(),
				factory.NewLoggerFactory(),
				jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.GetInvestmentAndReturnsInfo(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "GetInvestmentAndReturnsInfo() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_repaymentDividendHttpDelivery_DeleteRepaymentDividendSchedule(t *testing.T) {

	tests := []struct {
		name                     string
		echoCtx                  func(e *echo.Echo) echo.Context
		repaymentDividendUsecase func() *mocks.RepaymentDividendUsecase
		wantErr                  bool
	}{
		{
			name: "DELETE api/v1/project/:id/repayment-dividend-schedules/:schedule_id 호출성공",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path: "/api/v1/project/:id/repayment-dividend-schedules/:schedule_id",
				Params: map[string]string{
					"id":          "128931",
					"schedule_id": "128932",
				},
				Method: http.MethodDelete,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Claims: map[string]interface{}{
					"user":  "12378123",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_USER", "ROLE_PRODUCT"},
				},
			}),
			repaymentDividendUsecase: func() *mocks.RepaymentDividendUsecase {
				usecase := &mocks.RepaymentDividendUsecase{}
				usecase.On("DeleteRepaymentDividendSchedule", mock.Anything,
					mock.MatchedBy(func(request adomain.DeleteRepaymentDividendRequest) bool {
						ignoreTypes := cmpopts.IgnoreTypes(map[string]int{})
						result := cmp.Equal(request, adomain.DeleteRepaymentDividendRequest{
							JWTUser: jwthelper.User{
								Id:    12378123,
								Name:  "홍길동",
								Email: "gildong@email.com",
							},
							ProjectId:  128931,
							ScheduleId: 128932,
						}, ignoreTypes)

						return result
					})).Return(nil)

				return usecase
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := repaymentDividendHttpDelivery{
				tt.repaymentDividendUsecase(),
				factory.NewLoggerFactory(),
				jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.DeleteRepaymentDividendSchedule(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "DeleteRepaymentDividendSchedule() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}
