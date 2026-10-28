package http

import (
	"net/http"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	pkgtest "github.com/illuminarean/energy-share-core-api/delivery/testing"
	adomain "github.com/illuminarean/energy-share-core-api/domain"
	"github.com/illuminarean/energy-share-core-api/domain/mocks"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/validator"
)

const referralCodeApi = "/api/referral-codes/issue"

func Test_referralCodeHttpDelivery_Create(t *testing.T) {
	tests := []struct {
		name                string
		echoCtx             func(e *echo.Echo) echo.Context
		referralCodeUsecase func() *mocks.ReferralCodeUsecase
		wantErr             bool
	}{
		{
			name: "(成功) 紹介コード発行",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   referralCodeApi,
				Method: http.MethodPost,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Body: `{
					"referral_code": "SPRING2024",
					"notes": "春のキャンペーン用紹介コード"
				}`,
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "operator@email.com",
					"name":  "オペレーター",
					"role":  []string{"ROLE_OPERATOR"},
				},
			}),
			referralCodeUsecase: func() *mocks.ReferralCodeUsecase {
				usecase := &mocks.ReferralCodeUsecase{}
				usecase.On("Create", mock.Anything, adomain.ReferralCodeCreateRequest{
					JWTUser: jwthelper.User{
						Id:    1000,
						Email: "operator@email.com",
						Name:  "オペレーター",
						Roles: map[string]int{
							string(domain.ROLE_OPERATOR): 1,
						},
					},
					ReferralCode: "SPRING2024",
					Notes:        "春のキャンペーン用紹介コード",
				}).Return(adomain.ReferralCodeCreateResponse{
					ID:        1,
					OwnerType: domain.SYSTEM,
					OwnerID:   0,
					CodeType:  domain.CodeTypeReferral,
					Code:      "SPRING2024",
					Notes:     lo.ToPtr("春のキャンペーン用紹介コード"),
				}, nil)
				return usecase
			},
			wantErr: false,
		},
		{
			name: "(成功) 紹介コード発行 - notesなし",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   referralCodeApi,
				Method: http.MethodPost,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Body: `{
					"referral_code": "SUMMER2024"
				}`,
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "operator@email.com",
					"name":  "オペレーター",
					"role":  []string{"ROLE_OPERATOR"},
				},
			}),
			referralCodeUsecase: func() *mocks.ReferralCodeUsecase {
				usecase := &mocks.ReferralCodeUsecase{}
				usecase.On("Create", mock.Anything, adomain.ReferralCodeCreateRequest{
					JWTUser: jwthelper.User{
						Id:    1000,
						Email: "operator@email.com",
						Name:  "オペレーター",
						Roles: map[string]int{
							string(domain.ROLE_OPERATOR): 1,
						},
					},
					ReferralCode: "SUMMER2024",
					Notes:        "",
				}).Return(adomain.ReferralCodeCreateResponse{
					ID:        2,
					OwnerType: domain.SYSTEM,
					OwnerID:   0,
					CodeType:  domain.CodeTypeReferral,
					Code:      "SUMMER2024",
					Notes:     nil,
				}, nil)
				return usecase
			},
			wantErr: false,
		},
		{
			name: "(失敗) 紹介コード発行 - referral_codeが空",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   referralCodeApi,
				Method: http.MethodPost,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Body: `{
					"referral_code": "",
					"notes": "テスト"
				}`,
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "operator@email.com",
					"name":  "オペレーター",
					"role":  []string{"ROLE_OPERATOR"},
				},
			}),
			referralCodeUsecase: func() *mocks.ReferralCodeUsecase {
				return &mocks.ReferralCodeUsecase{}
			},
			wantErr: true,
		},
		{
			name: "(失敗) 紹介コード発行 - referral_codeが空白のみ",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   referralCodeApi,
				Method: http.MethodPost,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Body: `{
					"referral_code": "",
					"notes": "テスト"
				}`,
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "operator@email.com",
					"name":  "オペレーター",
					"role":  []string{"ROLE_OPERATOR"},
				},
			}),
			referralCodeUsecase: func() *mocks.ReferralCodeUsecase {
				return &mocks.ReferralCodeUsecase{}
			},
			wantErr: true,
		},
		{
			name: "(失敗) 紹介コード発行 - 重複エラー",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   referralCodeApi,
				Method: http.MethodPost,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Body: `{
					"referral_code": "DUPLICATE2024",
					"notes": "重複テスト"
				}`,
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "operator@email.com",
					"name":  "オペレーター",
					"role":  []string{"ROLE_OPERATOR"},
				},
			}),
			referralCodeUsecase: func() *mocks.ReferralCodeUsecase {
				usecase := &mocks.ReferralCodeUsecase{}
				usecase.On("Create", mock.Anything, adomain.ReferralCodeCreateRequest{
					JWTUser: jwthelper.User{
						Id:    1000,
						Email: "operator@email.com",
						Name:  "オペレーター",
						Roles: map[string]int{
							string(domain.ROLE_OPERATOR): 1,
						},
					},
					ReferralCode: "DUPLICATE2024",
					Notes:        "重複テスト",
				}).Return(adomain.ReferralCodeCreateResponse{}, cdomain.ErrAlreadyExists)
				return usecase
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := referralCodeHttpDelivery{
				usecase:   tt.referralCodeUsecase(),
				extractor: jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.Create(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "Create() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}
