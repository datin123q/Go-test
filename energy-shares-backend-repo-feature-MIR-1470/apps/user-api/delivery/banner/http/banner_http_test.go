package http

import (
	"net/http"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	pkgtest "github.com/illuminarean/energy-share-core-api/delivery/testing"
	udomain "github.com/illuminarean/energyshares-user-api/domain"
	"github.com/illuminarean/energyshares-user-api/domain/mocks"
)

const bannerApi = "/api/v1/banners"

func Test_bannerHttpDelivery_GetAll(t *testing.T) {
	tests := []struct {
		name          string
		echoCtx       func(e *echo.Echo) echo.Context
		bannerUsecase func() *mocks.BannerUsecase
		wantErr       bool
	}{
		{
			name: "(성공) 게시 중인 배너 목록 조회",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   bannerApi,
				Method: http.MethodGet,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
			}),
			bannerUsecase: func() *mocks.BannerUsecase {
				usecase := &mocks.BannerUsecase{}
				usecase.On("GetAll", mock.Anything).Return(udomain.BannerGetAllResponse{
					TotalCount:    int64(0),
					RollingSecond: 5,
					Items:         []udomain.BannerGetAllItem{},
				}, nil)

				return usecase
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := bannerHttpDelivery{
				bannerUsecase: tt.bannerUsecase(),
			}

			e := echo.New()

			err := delivery.GetAll(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "GetAll() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}
