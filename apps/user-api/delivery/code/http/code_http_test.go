package http

import (
	"net/http"
	"testing"

	pkgtest "github.com/illuminarean/energy-share-core-api/delivery/testing"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/validator"
	"github.com/illuminarean/energyshares-user-api/domain/mocks"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_codeHttpDelivery_GetAll(t *testing.T) {
	tests := []struct {
		name       string
		echoCtx    func(e *echo.Echo) echo.Context
		codeUsecase func() *mocks.CodeUsecase
		got string
		wantErr    bool
	}{
		{
			name: "GET /api/v1/codes 정상적으로 호출",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/codes",
				Method: http.MethodGet,
			}),
			codeUsecase: func() *mocks.CodeUsecase {
				usecase := &mocks.CodeUsecase{}
				usecase.On("GetAll", mock.Anything).Return(``, nil)

				return usecase
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := codeHttpDelivery{
				codeUsecase: tt.codeUsecase(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.GetAll(tt.echoCtx(e))

			assert.NoErrorf(t, err, "codeHttpDelivery.GetAll() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}
