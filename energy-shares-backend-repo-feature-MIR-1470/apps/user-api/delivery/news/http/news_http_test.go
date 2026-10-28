package http

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/mock"

	pkgtest "github.com/illuminarean/energy-share-core-api/delivery/testing"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/validator"
	udomain "github.com/illuminarean/energyshares-user-api/domain"
	"github.com/illuminarean/energyshares-user-api/domain/mocks"
)

func Test_NewsHttpDelivery_GetAll(t *testing.T) {
	tests := []struct {
		name        string
		echoCtx     func(e *echo.Echo) echo.Context
		newsUsecase func() *mocks.NewsUsecase
		wantErr     bool
	}{
		{
			name: "GET /v1/news",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/news",
				Method: http.MethodGet,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				QueryParams: map[string][]string{
					"max_result_count": {"10"},
					"order":            {"asc"},
					"skip_count":       {"0"},
					"sortby":           {"title"},
				},
			},
			),
			newsUsecase: func() *mocks.NewsUsecase {
				m := &mocks.NewsUsecase{}
				m.On("GetAll", mock.Anything, udomain.NewsGetAllRequest{
					ClientIp: "192.0.2.1",
					NewsGetAllQueryParam: udomain.NewsGetAllQueryParam{
						SearchInput: cdomain.SearchInput{
							Sortby:         "title",
							Order:          "asc",
							SkipCount:      0,
							MaxResultCount: 10,
						},
					},
				}).Return(udomain.NewsGetAllResponse{
					TotalCount: 1,
					Items: []domain.News{
						mockNews1,
					},
				}, nil)
				return m
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := newsHttpDelivery{
				extractor:   jwthelper.NewUserExtractor(),
				newsUsecase: tt.newsUsecase(),
			}
			e := echo.New()
			e.Validator = validator.NewCustomValidator()
			err := delivery.GetAll(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "GetAll() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}
