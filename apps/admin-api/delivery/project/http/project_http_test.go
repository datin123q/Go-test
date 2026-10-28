package http

import (
	pkgtest "github.com/illuminarean/energy-share-core-api/delivery/testing"
	adomain "github.com/illuminarean/energy-share-core-api/domain"
	"github.com/illuminarean/energy-share-core-api/domain/mocks"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/validator"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"testing"
)

func Test_projectHttpDelivery_GetStatistics(t *testing.T) {
	tests := []struct {
		name           string
		echoCtx        func(e *echo.Echo) echo.Context
		projectUsecase func() *mocks.ProjectUsecase
		wantErr        bool
	}{
		{
			name: "GET api/v1/projects/146/statistics 호출성공",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/projects/:id/statistics",
				Method: http.MethodGet,
				Params: map[string]string{"id": "146"},
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
			projectUsecase: func() *mocks.ProjectUsecase {
				usecase := &mocks.ProjectUsecase{}
				usecase.On("GetStatistics", mock.Anything, adomain.ProjectGetStatisticsRequest{
					ProjectId: 146,
				}).Return(adomain.ProjectGetStatisticsResponse{}, nil)

				return usecase
			},
			wantErr: false,
		},
		{
			name: "GET api/v1/projects/0/statistics 호출실패",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/projects/:id/statistics",
				Method: http.MethodGet,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
			}),
			projectUsecase: func() *mocks.ProjectUsecase {
				usecase := &mocks.ProjectUsecase{}
				return usecase
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := projectHttpDelivery{
				projectUsecase: tt.projectUsecase(),
				logger:         factory.NewLoggerFactory(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.GetStatistics(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "GetStatistics() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}
