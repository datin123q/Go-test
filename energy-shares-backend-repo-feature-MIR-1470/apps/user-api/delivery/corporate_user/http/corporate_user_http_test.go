package http

import (
	"net/http"
	"testing"

	"github.com/pkg/errors"

	"github.com/labstack/echo/v4"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	pkgtest "github.com/illuminarean/energy-share-core-api/delivery/testing"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/validator"
	udomain "github.com/illuminarean/energyshares-user-api/domain"
	"github.com/illuminarean/energyshares-user-api/domain/mocks"
)

func Test_corporateUserHttpDelivery_RegisterProfiles(t *testing.T) {
	tests := []struct {
		name                 string
		echoCtx              func(e *echo.Echo) echo.Context
		corporateUserUsecase func() *mocks.CorporateUserUsecase
		wantErr              bool
	}{
		{
			name: "POST /api/v1/corporate-users/signup/profiles 기본정보(basic_profile) 등록 성공",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/corporate-users/signup/profiles",
				Method: http.MethodPut,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Params: map[string]string{
					"id": "1000",
				},
				Body: `{
				   "sign_up_status":"SIGNUP_CORPORATE_PROFILE_BASIC",
				   "basic_profile":{
					  "corp_no":"1234567890123",
					  "corp_name":"회사명",
					  "corp_type":"사업내용"
				   }
				}`,
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_USER"},
				},
			}),
			corporateUserUsecase: func() *mocks.CorporateUserUsecase {
				usecase := &mocks.CorporateUserUsecase{}
				usecase.On("RegisterProfiles", mock.Anything,
					udomain.CorporateUserRegisterProfilesRequest{
						User: jwthelper.User{
							Id:    int64(1000),
							Email: "gildong@email.com",
							Name:  "홍길동",
							Roles: map[string]int{
								"ROLE_USER": 1,
							},
						},
						SignUpStatus: domain.SIGNUP_CORPORATE_PROFILE_BASIC,
						BasicProfile: udomain.CorporateUserBasicProfileItem{
							CorpNo:   "1234567890123",
							CorpName: "회사명",
							CorpType: "사업내용",
						},
					}).Return(nil)
				return usecase
			},
			wantErr: false,
		},
		{
			name: "POST /api/v1/corporate-users/signup/profiles 기본정보(basic_profile) 등록 실패 - 법인이 이미 존재",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/corporate-users/signup/profiles",
				Method: http.MethodPut,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Params: map[string]string{
					"id": "1000",
				},
				Body: `{
				   "sign_up_status":"SIGNUP_CORPORATE_PROFILE_BASIC",
				   "basic_profile":{
					  "corp_no":"1234567890123",
					  "corp_name":"회사명",
					  "corp_type":"사업내용"
				   }
				}`,
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_USER"},
				},
			}),
			corporateUserUsecase: func() *mocks.CorporateUserUsecase {
				usecase := &mocks.CorporateUserUsecase{}
				usecase.On("RegisterProfiles", mock.Anything,
					udomain.CorporateUserRegisterProfilesRequest{
						User: jwthelper.User{
							Id:    int64(1000),
							Email: "gildong@email.com",
							Name:  "홍길동",
							Roles: map[string]int{
								"ROLE_USER": 1,
							},
						},
						SignUpStatus: domain.SIGNUP_CORPORATE_PROFILE_BASIC,
						BasicProfile: udomain.CorporateUserBasicProfileItem{
							CorpNo:   "1234567890123",
							CorpName: "회사명",
							CorpType: "사업내용",
						},
					}).Return(errors.Wrap(cdomain.ErrAlreadyExists, "already existing corporate number"))
				return usecase
			},
			wantErr: true,
		},
		{
			name: "POST /api/v1/corporate-users/signup/profiles 기타정보(other_profile) 등록",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/corporate-users/signup/profiles",
				Method: http.MethodPut,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Params: map[string]string{
					"id": "1000",
				},
				Body: `{
				   "sign_up_status":"SIGNUP_CORPORATE_PROFILE_OTHER",
				   "other_profile":{
					  "zip_code":"1234567",
					  "province":"도도부현",
					  "city":"시군구",
					  "address":"상세주소",
					  "address2":"상세주소2",
					  "mobile_phone":"012345678"
				   }
				}`,
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_USER"},
				},
			}),
			corporateUserUsecase: func() *mocks.CorporateUserUsecase {
				usecase := &mocks.CorporateUserUsecase{}
				usecase.On("RegisterProfiles", mock.Anything,
					udomain.CorporateUserRegisterProfilesRequest{
						User: jwthelper.User{
							Id:    int64(1000),
							Email: "gildong@email.com",
							Name:  "홍길동",
							Roles: map[string]int{
								"ROLE_USER": 1,
							},
						},
						SignUpStatus: domain.SIGNUP_CORPORATE_PROFILE_OTHER,
						OtherProfile: udomain.CorporateUserOtherProfileItem{
							ZipCode:     "1234567",
							Province:    "도도부현",
							City:        "시군구",
							Address:     "상세주소",
							Address2:    lo.ToPtr("상세주소2"),
							MobilePhone: "012345678",
						},
					}).Return(nil)
				return usecase
			},
			wantErr: false,
		},
		{
			name: "POST /api/v1/corporate-users/signup/profiles 실질적 지배자(shareholders) 등록",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/corporate-users/signup/profiles",
				Method: http.MethodPut,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Params: map[string]string{
					"id": "1000",
				},
				Body: `{
				   "sign_up_status":"SIGNUP_CORPORATE_PROFILE_SHAREHOLDER",
				   "shareholder_profiles":[
					  {
						 "family_name":"홍",
						 "first_name":"길동",
						 "family_name_kana":"ホン",
						 "first_name_kana":"ギルドン",
						 "zip_code":"1234567",
						 "province":"도도부현",
						 "city":"시군구",
						 "address":"상세주소",
						 "address2":"상세주소2",
						 "birthday" : "2000-01-01"
					  }
				   ]
				}`,
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_USER"},
				},
			}),
			corporateUserUsecase: func() *mocks.CorporateUserUsecase {
				usecase := &mocks.CorporateUserUsecase{}
				usecase.On("RegisterProfiles", mock.Anything,
					udomain.CorporateUserRegisterProfilesRequest{
						User: jwthelper.User{
							Id:    int64(1000),
							Email: "gildong@email.com",
							Name:  "홍길동",
							Roles: map[string]int{
								"ROLE_USER": 1,
							},
						},
						SignUpStatus: domain.SIGNUP_CORPORATE_PROFILE_SHAREHOLDER,
						ShareholderProfiles: []udomain.CorporateUserShareholderProfileItem{
							{
								FamilyName:     "홍",
								FirstName:      "길동",
								FamilyNameKana: "ホン",
								FirstNameKana:  "ギルドン",
								ZipCode:        "1234567",
								Province:       "도도부현",
								City:           "시군구",
								Address:        "상세주소",
								Address2:       lo.ToPtr("상세주소2"),
								Birthday:       "2000-01-01",
							},
						},
					}).Return(nil)
				return usecase
			},
			wantErr: false,
		},
		{
			name: "POST /api/v1/corporate-users/signup/profiles 기타정보(other_profile) 등록, 우편번호가 7자리가 아님",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/corporate-users/signup/profiles",
				Method: http.MethodPut,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Body: `{
				   "sign_up_status":"SIGNUP_CORPORATE_PROFILE_OTHER",
				   "other_profile":{
					  "zip_code":"12345678",
					  "province":"도도부현",
					  "city":"시군구",
					  "address":"상세주소",
					  "address2":"상세주소2",
					  "mobile_phone":"012345678"
				   }
				}`,
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_USER"},
				},
			}),
			corporateUserUsecase: func() *mocks.CorporateUserUsecase {
				usecase := &mocks.CorporateUserUsecase{}
				return usecase
			},
			wantErr: true,
		},
		{
			name: "PUT /api/v1/corporate-users/signup/profiles 볍인 회원 출금 계좌 정보 등록 성공",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/corporate-users/signup/profiles",
				Method: http.MethodPut,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Params: map[string]string{
					"id": "1000",
				},
				Body: `{
					"sign_up_status":"SIGNUP_ACCOUNT",
					"withdrawal_account":{
						"bank_name":    "은행명",
						"branch_name":  "지점명",
						"bank_code":    "123",
						"branch_code":  "456",
						"deposit_type": "ORDINARY_ACCOUNT",
						"bank_account": "1234567",
						"depositor":    "HONGGILDONG"
					}
				}`,
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_USER"},
				},
			}),
			corporateUserUsecase: func() *mocks.CorporateUserUsecase {
				usecase := &mocks.CorporateUserUsecase{}
				usecase.On("RegisterProfiles", mock.Anything,
					udomain.CorporateUserRegisterProfilesRequest{
						User: jwthelper.User{
							Id:    int64(1000),
							Email: "gildong@email.com",
							Name:  "홍길동",
							Roles: map[string]int{
								"ROLE_USER": 1,
							},
						},
						SignUpStatus: domain.SIGNUP_ACCOUNT,
						WithdrawalAccount: udomain.CorporateUserRegisterWithdrawalAccountItem{
							BankName:    "은행명",
							BranchName:  "지점명",
							BankCode:    lo.ToPtr("123"),
							BranchCode:  lo.ToPtr("456"),
							DepositType: domain.ORDINARY_ACCOUNT,
							BankAccount: "1234567",
							Depositor:   "HONGGILDONG",
						},
					}).Return(nil)
				return usecase
			},
			wantErr: false,
		},
		{
			name: "PUT /api/v1/corporate-users/signup/withdrawal-account 볍인 회원 출금 계좌 정보 등록 실패 (계좌번호 최소 7자리)",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/corporate-users/signup/profiles",
				Method: http.MethodPut,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Params: map[string]string{
					"id": "1000",
				},
				Body: `{
					"sign_up_status":"SIGNUP_ACCOUNT",
					"withdrawal_account":{
						"bank_name":    "은행명",
						"branch_name":  "지점명",
						"bank_code":    "123",
						"branch_code":  "456",
						"deposit_type": "ORDINARY_ACCOUNT",
						"bank_account": "123456",
						"depositor":    "HONGGILDONG"
					}
				}`,
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_USER"},
				},
			}),
			corporateUserUsecase: func() *mocks.CorporateUserUsecase {
				usecase := &mocks.CorporateUserUsecase{}
				usecase.On("RegisterWithdrawalAccount", mock.Anything,
					udomain.CorporateUserRegisterProfilesRequest{
						User: jwthelper.User{
							Id:    int64(1000),
							Email: "gildong@email.com",
							Name:  "홍길동",
							Roles: map[string]int{
								"ROLE_USER": 1,
							},
						},
						SignUpStatus: domain.SIGNUP_ACCOUNT,
						WithdrawalAccount: udomain.CorporateUserRegisterWithdrawalAccountItem{
							BankName:    "은행명",
							BranchName:  "지점명",
							BankCode:    lo.ToPtr("123"),
							BranchCode:  lo.ToPtr("456"),
							DepositType: domain.ORDINARY_ACCOUNT,
							BankAccount: "123456",
							Depositor:   "HONGGILDONG",
						},
					}).Return(nil)
				return usecase
			},
			wantErr: true,
		},
		{
			name: "PUT /api/v1/corporate-users//signup/profiles 볍인 투자적합성 확인 성공",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/corporate-users/signup/profiles",
				Method: http.MethodPut,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Params: map[string]string{
					"id": "1000",
				},
				Body: `{
					"sign_up_status":"SIGNUP_INVESTMENT_SUITABILITY",
					"investment_suitability":{
						"employee_count":     "10",
						"capital_amount":     "100000",
						"sales":              "50000",
						"investment_purpose": "VARIANCE_PORTFOLIO",
						"has_experience":     true,
						"is_surplus_fund":    true
					}	
				}`,
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_USER"},
				},
			}),
			corporateUserUsecase: func() *mocks.CorporateUserUsecase {
				usecase := &mocks.CorporateUserUsecase{}
				usecase.On("RegisterProfiles", mock.Anything,
					udomain.CorporateUserRegisterProfilesRequest{
						User: jwthelper.User{
							Id:    int64(1000),
							Email: "gildong@email.com",
							Name:  "홍길동",
							Roles: map[string]int{
								"ROLE_USER": 1,
							},
						},
						SignUpStatus: domain.SIGNUP_INVESTMENT_SUITABILITY,
						InvestmentSuitability: udomain.CorporateUserConfirmInvestmentSuitabilityItem{
							EmployeeCount:     "10",
							CapitalAmount:     "100000",
							Sales:             "50000",
							InvestmentPurpose: domain.VARIANCE_PORTFOLIO,
							HasExperience:     lo.ToPtr(true),
							IsSurplusFund:     lo.ToPtr(true),
						},
					}).Return(nil)
				return usecase
			},
			wantErr: false,
		},
		{
			name: "PUT /api/v1/corporate-users/signup/profiles 볍인 투자적합성 확인 실패(투자목적 이상",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/corporate-users/signup/profiles",
				Method: http.MethodPut,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Params: map[string]string{
					"id": "1000",
				},
				Body: `{
					"sign_up_status":"SIGNUP_INVESTMENT_SUITABILITY",
					"investment_suitability":{
						"employee_count":     "10",
						"capital_amount":     "100000",
						"sales":              "50000",
						"investment_purpose": "GOOD_PORTFOLIO",
						"has_experience":     true,
						"is_surplus_fund":    true
					}	
				}`,
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_USER"},
				},
			}),
			corporateUserUsecase: func() *mocks.CorporateUserUsecase {
				usecase := &mocks.CorporateUserUsecase{}
				usecase.On("RegisterProfiles", mock.Anything,
					udomain.CorporateUserRegisterProfilesRequest{
						User: jwthelper.User{
							Id:    int64(1000),
							Email: "gildong@email.com",
							Name:  "홍길동",
							Roles: map[string]int{
								"ROLE_USER": 1,
							},
						},
						SignUpStatus: domain.SIGNUP_INVESTMENT_SUITABILITY,
						InvestmentSuitability: udomain.CorporateUserConfirmInvestmentSuitabilityItem{
							EmployeeCount:     "10",
							CapitalAmount:     "100000",
							Sales:             "50000",
							InvestmentPurpose: domain.INVESTMENT_PURPOSE_TYPE("GOOD_PORTFOLIO"),
							HasExperience:     lo.ToPtr(true),
							IsSurplusFund:     lo.ToPtr(true),
						},
					}).Return(nil)
				return usecase
			},
			wantErr: true,
		},
		{
			name: "PUT /api/v1/corporate-users/signup/profiles 약관 동의 성공",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/corporate-users/signup/profiles",
				Method: http.MethodPut,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Params: map[string]string{
					"id": "1000",
				},
				Body: `{
					"sign_up_status":"SIGNUP_AGREE"
				}`,
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_USER"},
				},
			}),
			corporateUserUsecase: func() *mocks.CorporateUserUsecase {
				usecase := &mocks.CorporateUserUsecase{}
				usecase.On("RegisterProfiles", mock.Anything,
					udomain.CorporateUserRegisterProfilesRequest{
						User: jwthelper.User{
							Id:    int64(1000),
							Email: "gildong@email.com",
							Name:  "홍길동",
							Roles: map[string]int{
								"ROLE_USER": 1,
							},
						},
						SignUpStatus: domain.SIGNUP_AGREE,
					}).Return(nil)
				return usecase
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := corporateUserHttpDelivery{
				corporateUserUsecase: tt.corporateUserUsecase(),
				extractor:            jwthelper.NewUserExtractor(),
			}

			e := echo.New()

			err := delivery.RegisterProfiles(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "RegisterProfiles() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_corporateUserHttpDelivery_GetOne(t *testing.T) {
	tests := []struct {
		name                 string
		echoCtx              func(e *echo.Echo) echo.Context
		corporateUserUsecase func() *mocks.CorporateUserUsecase
		wantErr              bool
	}{
		{
			name: "GET /api/v1/corporate-users/detail 상세조회 성공",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/corporate-users/detail",
				Method: http.MethodGet,
				Params: map[string]string{
					"id": "1000",
				},
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_USER"},
				},
			}),
			corporateUserUsecase: func() *mocks.CorporateUserUsecase {
				usecase := &mocks.CorporateUserUsecase{}
				usecase.On("GetOne", mock.Anything, udomain.CorporateUserGetOneRequest{
					User: jwthelper.User{
						Id:    int64(1000),
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{"ROLE_USER": 1},
					},
				}).Return(udomain.CorporateUserGetOneResponse{}, nil)
				return usecase
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := corporateUserHttpDelivery{
				corporateUserUsecase: tt.corporateUserUsecase(),
				extractor:            jwthelper.NewUserExtractor(),
			}

			e := echo.New()

			err := delivery.GetOne(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "GetOne() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_corporateUserHttpDelivery_GetSignUpStatus(t *testing.T) {
	tests := []struct {
		name                 string
		echoCtx              func(e *echo.Echo) echo.Context
		corporateUserUsecase func() *mocks.CorporateUserUsecase
		wantErr              bool
	}{
		{
			name: "GET /api/v1/corporate-users/signup/status 회원가입 상태 조회 성공",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/corporate-users/signup/status",
				Method: http.MethodGet,
				QueryParams: map[string][]string{
					"status": {"SIGNUP_CORPORATE_PROFILE_BASIC"},
				},
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_USER"},
				},
			}),
			corporateUserUsecase: func() *mocks.CorporateUserUsecase {
				usecase := &mocks.CorporateUserUsecase{}
				usecase.On("GetSignUpStatus", mock.Anything, udomain.CorporateUserGetSignUpStatusRequest{
					User: jwthelper.User{
						Id:    int64(1000),
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{"ROLE_USER": 1},
					},
					Status: domain.SIGNUP_CORPORATE_PROFILE_BASIC,
				}).Return(udomain.CorporateUserGetSignUpStatusResponse{}, nil)
				return usecase
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := corporateUserHttpDelivery{
				corporateUserUsecase: tt.corporateUserUsecase(),
				extractor:            jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()
			err := delivery.GetSignUpStatus(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "GetSignUpStatus() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_corporateUserHttpDelivery_VerifyPostalCode(t *testing.T) {
	tests := []struct {
		name                 string
		echoCtx              func(e *echo.Echo) echo.Context
		corporateUserUsecase func() *mocks.CorporateUserUsecase
		wantErrCode          int
		wantErr              bool
	}{
		{
			name: "POST /api/v1/corporate-users/verify-postal-code 우편번호 확인 성공",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/corporate-users/verify-postal-code",
				Method: http.MethodPost,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Body: `{
					"corporate_postal_code":"123456",
					"manager_postal_code":"654321"
				}`,
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_USER"},
				},
			}),
			corporateUserUsecase: func() *mocks.CorporateUserUsecase {
				usecase := &mocks.CorporateUserUsecase{}
				usecase.On("VerifyPostalCode", mock.Anything, udomain.CorporateUserVerifyPostalCodeRequest{
					User: jwthelper.User{
						Id:    int64(1000),
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{"ROLE_USER": 1},
					},
					CorporatePostalCode: "123456",
					ManagerPostalCode:   "654321",
				}).Return(nil)

				return usecase
			},
			wantErr: false,
		},
		{
			name: "POST /api/v1/corporate-users/verify-postal-code 審査状態が不正な場合は 20030 を返す",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/corporate-users/verify-postal-code",
				Method: http.MethodPost,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Body: `{
					"corporate_postal_code":"123456",
					"manager_postal_code":"654321"
				}`,
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_USER"},
				},
			}),
			corporateUserUsecase: func() *mocks.CorporateUserUsecase {
				usecase := &mocks.CorporateUserUsecase{}
				usecase.On("VerifyPostalCode", mock.Anything, mock.Anything).
					Return(errors.Wrap(cdomain.ErrItemConflict, "invalid audit status"))

				return usecase
			},
			// コード不一致(20015)ではなく無効な会員状態(20030)として返す
			wantErrCode: factory.ErrorInvalidAuditStatus.Code,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := corporateUserHttpDelivery{
				corporateUserUsecase: tt.corporateUserUsecase(),
				extractor:            jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()
			err := delivery.VerifyPostalCode(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "VerifyPostalCode() error = %v, wantErr %v", err, tt.wantErr)

			// 状態エラーがコード不一致として返らないことを確かめる
			if tt.wantErrCode != 0 {
				var apiErr *factory.Error
				if assert.ErrorAs(t, err, &apiErr) {
					assert.Equal(t, tt.wantErrCode, apiErr.Code)
				}
			}
		})
	}
}

func Test_corporateUserHttpDelivery_PatchOne(t *testing.T) {
	tests := []struct {
		name                 string
		echoCtx              func(e *echo.Echo) echo.Context
		corporateUserUsecase func() *mocks.CorporateUserUsecase
		wantErr              bool
	}{
		{
			name: "PATCH /api/v1/corporate-users/patch 회원정보 수정 성공 (is_newsletter=true)",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/corporate-users",
				Method: http.MethodPatch,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Body: `{
					"is_newsletter": true
				}`,
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_USER"},
				},
			}),
			corporateUserUsecase: func() *mocks.CorporateUserUsecase {
				usecase := &mocks.CorporateUserUsecase{}
				usecase.On("PatchOne", mock.Anything, udomain.CorporateUserPatchOneRequest{
					JwtUser: jwthelper.User{
						Id:    int64(1000),
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{"ROLE_USER": 1},
					},
					IsNewsletter: lo.ToPtr(true),
				}).Return(nil)

				return usecase
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := corporateUserHttpDelivery{
				corporateUserUsecase: tt.corporateUserUsecase(),
				extractor:            jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()
			err := delivery.PatchOne(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "PatchOne() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_corporateUserHttpDelivery_GetDocuments(t *testing.T) {
	tests := []struct {
		name                 string
		echoCtx              func(e *echo.Echo) echo.Context
		corporateUserUsecase func() *mocks.CorporateUserUsecase
		wantErr              bool
	}{
		{
			name: "GET /api/v1/corporate-users/documents 등록한 문서 목록 조회 성공",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/corporate-users/documents",
				Method: http.MethodGet,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_USER"},
				},
			}),
			corporateUserUsecase: func() *mocks.CorporateUserUsecase {
				usecase := &mocks.CorporateUserUsecase{}
				usecase.On("GetDocuments", mock.Anything, udomain.CorporateUserGetDocumentsRequest{
					JwtUser: jwthelper.User{
						Id:    int64(1000),
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{"ROLE_USER": 1},
					},
				}).Return(udomain.CorporateUserGetDocumentsResponse{}, nil)
				return usecase
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := corporateUserHttpDelivery{
				corporateUserUsecase: tt.corporateUserUsecase(),
				extractor:            jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()
			err := delivery.GetDocuments(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "GetDocuments() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_corporateUserHttpDelivery_UpdateDocuments(t *testing.T) {
	tests := []struct {
		name                 string
		echoCtx              func(e *echo.Echo) echo.Context
		corporateUserUsecase func() *mocks.CorporateUserUsecase
		wantErr              bool
	}{
		{
			name: "PUT /api/v1/corporate-users/documents 문서 등록 성공",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/corporate-users/documents",
				Method: http.MethodPut,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_USER"},
				},
				Body: `{
					"certified_copy_file_url": "http://certified_copy_file_url.com",
					"poa_file_url": ""
				}`,
			}),
			corporateUserUsecase: func() *mocks.CorporateUserUsecase {
				usecase := &mocks.CorporateUserUsecase{}
				usecase.On("UpdateDocuments", mock.Anything, udomain.CorporateUserUpdateDocumentsRequest{
					JwtUser: jwthelper.User{
						Id:    int64(1000),
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{"ROLE_USER": 1},
					},
					CorporateUserFileUrlItem: udomain.CorporateUserFileUrlItem{
						CertifiedCopyFileUrl: lo.ToPtr("http://certified_copy_file_url.com"),
						PoaFileUrl:           lo.ToPtr(""),
					},
				}).Return(nil)
				return usecase
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := corporateUserHttpDelivery{
				corporateUserUsecase: tt.corporateUserUsecase(),
				extractor:            jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()
			err := delivery.UpdateDocuments(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "UpdateDocuments() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_corporateUserHttpDelivery_VerifyWithKyc(t *testing.T) {
	tests := []struct {
		name                 string
		echoCtx              func(e *echo.Echo) echo.Context
		corporateUserUsecase func() *mocks.CorporateUserUsecase
		wantErr              bool
	}{
		{
			name: "POST /api/v1/corporate-users/signup/kyc KYC 문서 등록 성공",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/corporate-users/signup/kyc",
				Method: http.MethodPost,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_USER"},
				},
				Body: `{
					"identification1_front_file_url": "http://identification1_front_file_url.com",
					"identification2_front_file_url": "http://identification2_front_file_url.com",
					"certified_copy_file_url": "http://certified_copy_file_url.com",
					"poa_file_url": "http://poa_file_url.com"
				}`,
			}),
			corporateUserUsecase: func() *mocks.CorporateUserUsecase {
				usecase := &mocks.CorporateUserUsecase{}
				usecase.On("VerifyWithKyc", mock.Anything, udomain.CorporateUserVerifyWithKycRequest{
					JWTUser: jwthelper.User{
						Id:    int64(1000),
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{"ROLE_USER": 1},
					},
					CorporateUserFileUrlItem: udomain.CorporateUserFileUrlItem{
						Identification1FrontFileUrl: lo.ToPtr("http://identification1_front_file_url.com"),
						Identification2FrontFileUrl: lo.ToPtr("http://identification2_front_file_url.com"),
						CertifiedCopyFileUrl:        lo.ToPtr("http://certified_copy_file_url.com"),
						PoaFileUrl:                  lo.ToPtr("http://poa_file_url.com"),
					},
				}).Return(nil)
				return usecase
			},
			wantErr: false,
		},
		{
			name: "POST /api/v1/corporate-users/signup/kyc KYC 문서 등록 실패 - 필수 파일 누락",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/corporate-users/signup/kyc",
				Method: http.MethodPost,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_USER"},
				},
				Body: `{
					"identification1_front_file_url": "http://identification1_front_file_url.com"
				}`,
			}),
			corporateUserUsecase: func() *mocks.CorporateUserUsecase {
				usecase := &mocks.CorporateUserUsecase{}
				usecase.On("VerifyWithKyc", mock.Anything, udomain.CorporateUserVerifyWithKycRequest{
					JWTUser: jwthelper.User{
						Id:    int64(1000),
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{"ROLE_USER": 1},
					},
					CorporateUserFileUrlItem: udomain.CorporateUserFileUrlItem{
						Identification1FrontFileUrl: lo.ToPtr("http://identification1_front_file_url.com"),
					},
				}).Return(nil)
				return usecase
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := corporateUserHttpDelivery{
				corporateUserUsecase: tt.corporateUserUsecase(),
				extractor:            jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()
			err := delivery.VerifyWithKyc(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "VerifyWithKyc() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}
