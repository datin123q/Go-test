package http

import (
	pkgtest "github.com/illuminarean/energy-share-core-api/delivery/testing"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/validator"
	udomain "github.com/illuminarean/energyshares-user-api/domain"
	"github.com/illuminarean/energyshares-user-api/domain/mocks"
	"github.com/labstack/echo/v4"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"testing"
)

func Test_faqHttpDelivery_GetAll(t *testing.T) {
	tests := []struct {
		name       string
		echoCtx    func(e *echo.Echo) echo.Context
		faqUsecase func() *mocks.FaqUsecase
		wantErr    bool
	}{
		{
			name: "GET /api/v1/faqs 정상적으로 호출 (쿼리파라미터 사용)",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/faqs",
				Method: http.MethodGet,
				QueryParams: map[string][]string{
					"category_type":    {"FAQ_CATEGORY_ACCOUNT_REGISTRATION"},
					"max_result_count": {"10"},
					"order":            {"asc"},
					"skip_count":       {"0"},
					"sortby":           {"title"},
				},
			}),
			faqUsecase: func() *mocks.FaqUsecase {
				usecase := &mocks.FaqUsecase{}
				usecase.On("GetAll", mock.Anything, udomain.FaqGetAllRequest{
					CategoryType: lo.ToPtr("FAQ_CATEGORY_ACCOUNT_REGISTRATION"),
					SearchInput: cdomain.SearchInput{
						Sortby:         "title",
						Order:          "asc",
						SkipCount:      0,
						MaxResultCount: 10,
					},
				}).Return(udomain.FaqGetAllResponse{}, nil)

				return usecase
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := faqHttpDelivery{
				faqUsecase: tt.faqUsecase(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.GetAll(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "GetAll() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}
