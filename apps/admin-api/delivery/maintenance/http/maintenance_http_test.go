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
	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/validator"
)

func Test_MaintenanceHttpDelivery_PatchProjects(t *testing.T) {
	tests := []struct {
		name               string
		echoCtx            func(e *echo.Echo) echo.Context
		cfg                *config.Config
		maintenanceUsecase func() *mocks.MaintenanceUsecase
		wantErr            bool
	}{
		{
			name: "PATCH /api/v1/maintenance/projects 정상적으로 호출",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/maintenance/projects",
				Method: http.MethodPatch,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Body: `{"project_id": 123214, "end_at": "2024-07-09 00:00:00"}`,
			}),
			cfg: &config.Config{},
			maintenanceUsecase: func() *mocks.MaintenanceUsecase {
				usecase := &mocks.MaintenanceUsecase{}
				usecase.On("PatchProjects", mock.Anything, adomain.PatchMaintenanceProjectsRequest{
					ProjectId: 123214,
					EndAt:     "2024-07-09 00:00:00",
				}).
					Return(nil)
				return usecase
			},
			wantErr: false,
		},
		{
			name: "PATCH /api/v1/maintenance/projects endAt을 잘못된 포맷으로 요청",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/maintenance/projects",
				Method: http.MethodPatch,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Body: `{"project_id": 123214, "end_at": "2024-07-09"}`,
			}),
			cfg: &config.Config{},
			maintenanceUsecase: func() *mocks.MaintenanceUsecase {
				usecase := &mocks.MaintenanceUsecase{}
				return usecase
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := maintenanceHttpDelivery{
				cfg:                tt.cfg,
				maintenanceUsecase: tt.maintenanceUsecase(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.PatchProjects(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "PatchProjects() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_MaintenanceHttpDelivery_PatchDepositRequests(t *testing.T) {
	tests := []struct {
		name               string
		echoCtx            func(e *echo.Echo) echo.Context
		cfg                *config.Config
		maintenanceUsecase func() *mocks.MaintenanceUsecase
		wantErr            bool
	}{
		{
			name: "PATCH /api/v1/maintenance/deposit-requests 정상적으로 호출",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/maintenance/deposit-requests",
				Method: http.MethodPatch,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Body: `{"deposit_request_id": 123214, "expire_deposit_date": "2024-07-09 00:00:00"}`,
			}),
			cfg: &config.Config{},
			maintenanceUsecase: func() *mocks.MaintenanceUsecase {
				usecase := &mocks.MaintenanceUsecase{}
				usecase.On("PatchDepositRequests", mock.Anything, adomain.PatchMaintenanceDepositRequestsRequest{
					DepositRequestId:  123214,
					ExpireDepositDate: "2024-07-09 00:00:00",
				}).
					Return(nil)
				return usecase
			},
			wantErr: false,
		},
		{

			name: "PATCH /api/v1/maintenance/deposit-requests expire_deposit_date를 잘못된 포맷으로 요청",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/maintenance/deposit-requests",
				Method: http.MethodPatch,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Body: `{"deposit_request_id": 123214, "expire_deposit_date": "2024-07-09"}`,
			}),
			cfg: &config.Config{},
			maintenanceUsecase: func() *mocks.MaintenanceUsecase {
				usecase := &mocks.MaintenanceUsecase{}
				return usecase
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := maintenanceHttpDelivery{
				cfg:                tt.cfg,
				maintenanceUsecase: tt.maintenanceUsecase(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.PatchDepositRequests(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "PatchDepositRequests() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_MaintenanceHttpDelivery_PatchRepaymentDividendSchedules(t *testing.T) {
	tests := []struct {
		name               string
		echoCtx            func(e *echo.Echo) echo.Context
		cfg                *config.Config
		maintenanceUsecase func() *mocks.MaintenanceUsecase
		wantErr            bool
	}{
		{
			name: "PATCH /api/v1/maintenance/repayment-dividend-schedules 정상적으로 호출",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/maintenance/repayment-dividend-schedules",
				Method: http.MethodPatch,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Body: `{"repayment_dividend_schedule_id": 123214, "execute_date": "2024-07-09 00:00:00"}`,
			}),
			cfg: &config.Config{},
			maintenanceUsecase: func() *mocks.MaintenanceUsecase {
				usecase := &mocks.MaintenanceUsecase{}
				usecase.On("PatchRepaymentDividendSchedules", mock.Anything, adomain.PatchMaintenanceRepaymentDividendSchedulesRequest{
					RepaymentDividendScheduleId: 123214,
					ExecuteDate:                 "2024-07-09 00:00:00",
				}).
					Return(nil)
				return usecase
			},
			wantErr: false,
		},
		{
			name: "PATCH /api/v1/maintenance/repayment-dividend-schedules 운영환경에서는 수행되지 않음",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/maintenance/repayment-dividend-schedules",
				Method: http.MethodPatch,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Body: `{"repayment_dividend_schedule_id": 123214, "execute_date": "2024-07-09 00:00:00"}`,
			}),
			cfg: &config.Config{
				CommonConfig: config.CommonConfig{
					Env: "prod",
				},
			},
			maintenanceUsecase: func() *mocks.MaintenanceUsecase {
				usecase := &mocks.MaintenanceUsecase{}
				return usecase
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := maintenanceHttpDelivery{
				cfg:                tt.cfg,
				maintenanceUsecase: tt.maintenanceUsecase(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.PatchRepaymentDividendSchedules(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "PatchRepaymentDividendSchedules() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}
