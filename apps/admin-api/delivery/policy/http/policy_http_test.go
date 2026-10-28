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

const policyApi = "/api/v1/policies"

func Test_policyHttpDelivery_Update(t *testing.T) {
	tests := []struct {
		name          string
		echoCtx       func(e *echo.Echo) echo.Context
		policyUsecase func() adomain.PolicyUsecase
		wantErr       bool
	}{
		{
			name: "(성공) 약관 수정 성공",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   policyApi + "/:id",
				Method: http.MethodPut,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "admin@email.com",
					"name":  "admin",
					"role":  []string{"ROLE_ADMIN", "ROLE_PRODUCT"},
				},
				Params: map[string]string{
					"id": "1",
				},
				Body: `{
					"policy_view_type": "POLICY_VIEW_JOIN",
					"content_html": "<h1>test</h1>",
					 "attachment_urls": [
						"https://es-jp-devv2-public-file-storage.s3.ap-northeast-2.amazonaws.com/attachment_file/2024-07-26/cqhkh82nulri2.jpg"
					],
					"is_posted": true,
					"label": "2024-07-26"
				}`,
			}),
			policyUsecase: func() adomain.PolicyUsecase {
				usecase := &mocks.PolicyUsecase{}
				usecase.On("Update", mock.Anything, adomain.PolicyUpdateRequest{
					JWTUser: jwthelper.User{
						Id:    1000,
						Email: "admin@email.com",
						Name:  "admin",
						Roles: map[string]int{"ROLE_ADMIN": 1, "ROLE_PRODUCT": 1},
					},
					Id:             1,
					PolicyViewType: domain.POLICY_VIEW_JOIN,
					ContentHtml:    "<h1>test</h1>",
					AttachmentUrls: []string{"https://es-jp-devv2-public-file-storage.s3.ap-northeast-2.amazonaws.com/attachment_file/2024-07-26/cqhkh82nulri2.jpg"},
				}).Return(adomain.PolicyUpdateResponse{}, nil)
				return usecase
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := policyHttpDelivery{
				policyUsecase: tt.policyUsecase(),
				extractor:     jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.Update(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "Update() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

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
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "admin@email.com",
					"name":  "admin",
					"role":  []string{"ROLE_ADMIN", "ROLE_PRODUCT"},
				},
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
				usecase.On("GetAll", mock.Anything, adomain.PolicyGetAllRequest{
					JWTUser: jwthelper.User{
						Id:    1000,
						Email: "admin@email.com",
						Name:  "admin",
						Roles: map[string]int{"ROLE_ADMIN": 1, "ROLE_PRODUCT": 1},
					},
					PolicyGetAllQueryParam: adomain.PolicyGetAllQueryParam{
						SearchInput: cdomain.SearchInput{
							Sortby:         "id",
							Order:          "asc",
							SkipCount:      0,
							MaxResultCount: 10,
						},
						PolicyViewType: lo.ToPtr(string(domain.POLICY_VIEW_JOIN)),
						PolicyType:     nil,
					},
				}).Return(adomain.PolicyGetAllResponse{}, nil)
				return usecase
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := policyHttpDelivery{
				policyUsecase: tt.policyUsecase(),
				extractor:     jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.GetAll(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "GetAll() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_policyHttpDelivery_GetOne(t *testing.T) {
	tests := []struct {
		name          string
		echoCtx       func(e *echo.Echo) echo.Context
		policyUsecase func() *mocks.PolicyUsecase
		wantErr       bool
	}{
		{
			name: "(성공) 약관 단건 조회 성공",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   policyApi + "/:id",
				Method: http.MethodGet,
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "admin@email.com",
					"name":  "admin",
					"role":  []string{"ROLE_ADMIN", "ROLE_PRODUCT"},
				},
				Params: map[string]string{
					"id": "1",
				},
			}),
			policyUsecase: func() *mocks.PolicyUsecase {
				usecase := &mocks.PolicyUsecase{}
				usecase.On("GetOne", mock.Anything, adomain.PolicyGetOneRequest{
					JWTUser: jwthelper.User{
						Id:    1000,
						Email: "admin@email.com",
						Name:  "admin",
						Roles: map[string]int{"ROLE_ADMIN": 1, "ROLE_PRODUCT": 1},
					},
					Id: 1,
				}).Return(adomain.PolicyGetOneResponse{}, nil)
				return usecase
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := policyHttpDelivery{
				policyUsecase: tt.policyUsecase(),
				extractor:     jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.GetOne(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "GetOne() error = %v, wantErr %v", err, tt.wantErr)
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
				Path:   policyApi + "/:id/histories",
				Method: http.MethodGet,
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "admin@email.com",
					"name":  "admin",
					"role":  []string{"ROLE_ADMIN", "ROLE_PRODUCT"},
				},
				Params: map[string]string{
					"id": "1",
				},
				QueryParams: map[string][]string{
					"max_result_count": {"10"},
					"order":            {"asc"},
					"skip_count":       {"0"},
					"sortby":           {"id"},
				},
			}),
			policyUsecase: func() *mocks.PolicyUsecase {
				usecase := &mocks.PolicyUsecase{}
				usecase.On("GetAllHistory", mock.Anything, adomain.PolicyGetAllHistoryRequest{
					JWTUser: jwthelper.User{
						Id:    1000,
						Email: "admin@email.com",
						Name:  "admin",
						Roles: map[string]int{"ROLE_ADMIN": 1, "ROLE_PRODUCT": 1},
					},
					Id: int64(1),
					PolicyGetAllHistoryQueryParam: adomain.PolicyGetAllHistoryQueryParam{
						SearchInput: cdomain.SearchInput{
							Sortby:         "id",
							Order:          "asc",
							SkipCount:      0,
							MaxResultCount: 10,
						},
					},
				}).Return(adomain.PolicyGetAllHistoryResponse{}, nil)
				return usecase
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := policyHttpDelivery{
				policyUsecase: tt.policyUsecase(),
				extractor:     jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.GetAllHistory(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "GetAllHistory() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_policyHttpDelivery_UpdateHistory(t *testing.T) {
	tests := []struct {
		name          string
		echoCtx       func(e *echo.Echo) echo.Context
		policyUsecase func() *mocks.PolicyUsecase
		wantErr       bool
	}{
		{
			name: "(성공) 약관 히스토리 수정",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   policyApi + "/:id/histories/:history_id",
				Method: http.MethodPut,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "admin@email.com",
					"name":  "admin",
					"role":  []string{"ROLE_ADMIN", "ROLE_PRODUCT"},
				},
				Params: map[string]string{
					"id":         "1",
					"history_id": "1",
				},
				Body: `{
					"is_posted": true,
					"label": "2024-07-26"
				}`,
			}),
			policyUsecase: func() *mocks.PolicyUsecase {
				usecase := &mocks.PolicyUsecase{}
				usecase.On("UpdateHistory", mock.Anything, adomain.PolicyUpdateHistoryRequest{
					JWTUser: jwthelper.User{
						Id:    1000,
						Email: "admin@email.com",
						Name:  "admin",
						Roles: map[string]int{"ROLE_ADMIN": 1, "ROLE_PRODUCT": 1},
					},
					Id:        1,
					HistoryId: 1,
					IsPosted:  true,
					Label:     lo.ToPtr("2024-07-26"),
				}).Return(adomain.PolicyUpdateHistoryResponse{}, nil)
				return usecase
			},
			wantErr: false,
		},
		{
			name: "(실패) 약관 히스토리 수정 - is_posted 가 true 일 때 label 이 없는 경우",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   policyApi + "/:id/histories/:history_id",
				Method: http.MethodPut,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "admin@email.com",
					"name":  "admin",
					"role":  []string{"ROLE_ADMIN", "ROLE_PRODUCT"},
				},
				Params: map[string]string{
					"id":         "1",
					"history_id": "1",
				},
				Body: `{
					"is_posted": true
				}`,
			}),
			policyUsecase: func() *mocks.PolicyUsecase {
				usecase := &mocks.PolicyUsecase{}
				usecase.On("UpdateHistory", mock.Anything, adomain.PolicyUpdateHistoryRequest{
					JWTUser: jwthelper.User{
						Id:    1000,
						Email: "admin@email.com",
						Name:  "admin",
						Roles: map[string]int{"ROLE_ADMIN": 1, "ROLE_PRODUCT": 1},
					},
					Id:        1,
					HistoryId: 1,
					IsPosted:  true,
					Label:     nil,
				}).Return(adomain.PolicyUpdateHistoryResponse{}, nil)
				return usecase
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := policyHttpDelivery{
				policyUsecase: tt.policyUsecase(),
				extractor:     jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.UpdateHistory(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "UpdateHistory() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_policyHttpDelivery_Delete(t *testing.T) {
	tests := []struct {
		name          string
		echoCtx       func(e *echo.Echo) echo.Context
		policyUsecase func() *mocks.PolicyUsecase
		wantErr       bool
	}{
		{
			name: "(성공) 약관 삭제 성공",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   policyApi + "/:id",
				Method: http.MethodDelete,
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "admin@email.com",
					"name":  "admin",
					"role":  []string{"ROLE_ADMIN", "ROLE_PRODUCT"},
				},
				Params: map[string]string{
					"id": "1",
				},
			}),
			policyUsecase: func() *mocks.PolicyUsecase {
				usecase := &mocks.PolicyUsecase{}
				usecase.On("Delete", mock.Anything, adomain.PolicyDeleteRequest{
					JWTUser: jwthelper.User{
						Id:    1000,
						Email: "admin@email.com",
						Name:  "admin",
						Roles: map[string]int{"ROLE_ADMIN": 1, "ROLE_PRODUCT": 1},
					},
					Id: 1,
				}).Return(nil)
				return usecase
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := policyHttpDelivery{
				policyUsecase: tt.policyUsecase(),
				extractor:     jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.Delete(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "Delete() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}
