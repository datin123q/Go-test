package http

import (
	"net/http"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	pkgtest "github.com/illuminarean/energy-share-core-api/delivery/testing"
	udomain "github.com/illuminarean/energy-share-core-api/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/validator"
)

func TestHealthHttpDelivery_HealthCheck(t *testing.T) {
	tests := []struct {
		name    string
		echoCtx func(e *echo.Echo) echo.Context
		want    udomain.HealthCheckResponse
		wantErr bool
	}{
		{
			name: "헬스 체크 요청에 정상 응답",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/health-check",
				Method: http.MethodGet,
			}),
			want: udomain.HealthCheckResponse{
				Status: "UP",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := healthHttpDelivery{}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.HealthCheck(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "HealthCheck() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}
