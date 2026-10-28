package http

import (
	"net/http"
	"testing"

	pkgtest "github.com/illuminarean/energy-share-core-api/delivery/testing"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/validator"
	udomain "github.com/illuminarean/energyshares-user-api/domain"
	"github.com/illuminarean/energyshares-user-api/domain/mocks"
	"github.com/labstack/echo/v4"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const policyApi = "/api/v1/policies"

func Test_policyHttpDelivery_GetAll(t *testing.T) {
	tests := []struct {
		name          string
		echoCtx       func(e *echo.Echo) echo.Context
		policyUsecase func() *mocks.PolicyUsecase
		wantErr       bool
	}{
		{
			name: "(성공) 약관 전체 조회 성공",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   policyApi,
				Method: http.MethodGet,
				QueryParams: map[string][]string{
					"policy_view_type": {"POLICY_VIEW_JOIN"},
					"max_result_count": {"10"},
					"order":            {"asc"},
					"skip_count":       {"0"},
					"sortby":           {"id"},
				},
			}),
			policyUsecase: func() *mocks.PolicyUsecase {
				usecase := &mocks.PolicyUsecase{}
				usecase.On("GetAll", mock.Anything, udomain.PolicyGetAllRequest{
					PolicyGetAllQueryParam: udomain.PolicyGetAllQueryParam{
						SearchInput: cdomain.SearchInput{
							Sortby:         "id",
							Order:          "asc",
							SkipCount:      0,
							MaxResultCount: 10,
						},
						PolicyViewType: lo.ToPtr(string(domain.POLICY_VIEW_JOIN)),
						PolicyType:     nil,
					},
				}).Return(udomain.PolicyGetAllResponse{}, nil)
				return usecase
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := policyHttpDelivery{
				policyUsecase: tt.policyUsecase(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.GetAll(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "GetAll() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_policyHttpDelivery_GetAllHistory(t *testing.T) {
	tests := []struct {
		name          string
		echoCtx       func(e *echo.Echo) echo.Context
		policyUsecase func() *mocks.PolicyUsecase
		wantErr       bool
	}{
		{
			name: "(성공) 약관 히스토리 조회 성공",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   policyApi + "/histories",
				Method: http.MethodGet,
				QueryParams: map[string][]string{
					"policy_view_type": {"POLICY_VIEW_JOIN"},
					"policy_type":      {"POLICY_CAUTION"},
					"max_result_count": {"10"},
					"order":            {"asc"},
					"skip_count":       {"0"},
					"sortby":           {"id"},
				},
			}),
			policyUsecase: func() *mocks.PolicyUsecase {
				usecase := &mocks.PolicyUsecase{}
				usecase.On("GetAllHistory", mock.Anything, udomain.PolicyGetAllHistoryRequest{
					PolicyGetAllHistoryQueryParam: udomain.PolicyGetAllHistoryQueryParam{
						PolicyViewType: lo.ToPtr(string(domain.POLICY_VIEW_JOIN)),
						PolicyType:     lo.ToPtr(string(domain.POLICY_CAUTION)),
						SearchInput: cdomain.SearchInput{
							Sortby:         "id",
							Order:          "asc",
							SkipCount:      0,
							MaxResultCount: 10,
						},
					},
				}).Return(udomain.PolicyGetAllHistoryResponse{}, nil)
				return usecase
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := policyHttpDelivery{
				policyUsecase: tt.policyUsecase(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.GetAllHistory(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "GetAllHistory() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}
