package http

import (
	"net/http"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	pkgtest "github.com/illuminarean/energy-share-core-api/delivery/testing"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/validator"
	udomain "github.com/illuminarean/energyshares-user-api/domain"
	"github.com/illuminarean/energyshares-user-api/domain/mocks"
)

const privateUserApi = "/api/v1/private-users"

func Test_privateUserHttpDelivery_RegisterProfile(t *testing.T) {
	tests := []struct {
		name               string
		echoCtx            func(e *echo.Echo) echo.Context
		privateUserUsecase func() *mocks.PrivateUserUsecase
		wantErr            bool
		// wantFactoryErrorCode が0でなければ、返ったerrorが*factory.Errorで
		// このCodeを持つことまで検証する(MIR-1121: 20037のような専用コードは
		// wantErrだけだと別のエラー種別へのマッピング崩れを見逃すため)。
		wantFactoryErrorCode int
	}{
		{
			name: "(성공) 개인 회원 기본 정보1(basic_profile) 등록",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   privateUserApi + "/signup/profiles",
				Method: http.MethodPut,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Params: map[string]string{
					"id": "1000",
				},
				Body: `{
				   "sign_up_status":"SIGNUP_PRIVATE_PROFILE_BASIC",
				   "basic_profile": {
						"birthday": "1996-02-09",
						"family_name": "family_name",
						"family_name_kana": "family_name_kana",
						"first_name": "first_name",
						"first_name_kana": "first_name_kana",
						"gender": "MALE",
						"nationality": "JP"
					}
				}`,
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_USER"},
				},
			}),
			privateUserUsecase: func() *mocks.PrivateUserUsecase {
				usecase := &mocks.PrivateUserUsecase{}
				usecase.On("RegisterProfile", mock.Anything, udomain.PrvUserRegisterProfileRequest{
					JWTUser: jwthelper.User{
						Id:    int64(1000),
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{
							"ROLE_USER": 1,
						},
					},
					SignUpStatus: domain.SIGNUP_PRIVATE_PROFILE_BASIC,
					BasicProfile: udomain.PrvUserBasicProfile{
						Birthday:       "1996-02-09",
						FamilyName:     "family_name",
						FamilyNameKana: "family_name_kana",
						FirstName:      "first_name",
						FirstNameKana:  "first_name_kana",
						Gender:         domain.MALE,
						Nationality:    domain.JP,
					},
				}).Return(udomain.PrvUserRegisterProfileResponse{}, nil)
				return usecase
			},
			wantErr: false,
		},
		{
			name: "(성공) 개인 회원 기본 정보2(other_profile) 등록",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   privateUserApi + "/signup/profiles",
				Method: http.MethodPut,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Params: map[string]string{
					"id": "1000",
				},
				Body: `{
				   "sign_up_status":"SIGNUP_PRIVATE_PROFILE_OTHER",
				   "other_profile": {
						"address": "address",
						"city": "city",
						"job": "USER_OFFICIALS",
						"mobile_phone": "010-1111-2222",
						"province": "province",
						"zip_code": "6894401"
					}
				}`,
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_USER"},
				},
			}),
			privateUserUsecase: func() *mocks.PrivateUserUsecase {
				usecase := &mocks.PrivateUserUsecase{}
				usecase.On("RegisterProfile", mock.Anything, udomain.PrvUserRegisterProfileRequest{
					JWTUser: jwthelper.User{
						Id:    int64(1000),
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{
							"ROLE_USER": 1,
						},
					},
					SignUpStatus: domain.SIGNUP_PRIVATE_PROFILE_OTHER,
					OtherProfile: udomain.PrvUserOtherProfile{
						Address:     "address",
						City:        "city",
						Job:         domain.USER_OFFICIALS,
						MobilePhone: "010-1111-2222",
						Province:    "province",
						ZipCode:     "6894401",
					},
				}).Return(udomain.PrvUserRegisterProfileResponse{}, nil)
				return usecase
			},
			wantErr: false,
		},
		{
			name: "(성공) 개인 회원 출금계좌 정보(withdrawal_account) 등록",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   privateUserApi + "/signup/profiles",
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
				   "withdrawal_account": {
						"bank_account": "2434551",
						"bank_code": "1111",
						"bank_name": "bank_name",
						"branch_code": "2222",
						"branch_name": "branch_name",
						"depositor": "depositor",
						"deposit_type": "ORDINARY_ACCOUNT"
				  }
				}`,
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_USER"},
				},
			}),
			privateUserUsecase: func() *mocks.PrivateUserUsecase {
				usecase := &mocks.PrivateUserUsecase{}
				usecase.On("RegisterProfile", mock.Anything, udomain.PrvUserRegisterProfileRequest{
					JWTUser: jwthelper.User{
						Id:    int64(1000),
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{
							"ROLE_USER": 1,
						},
					},
					SignUpStatus: domain.SIGNUP_ACCOUNT,
					WithdrawalAccount: udomain.PrvUserWithdrawalAccount{
						BankAccount: "2434551",
						BankCode:    "1111",
						BankName:    "bank_name",
						BranchCode:  "2222",
						BranchName:  "branch_name",
						Depositor:   "depositor",
						DepositType: domain.ORDINARY_ACCOUNT,
					},
				}).Return(udomain.PrvUserRegisterProfileResponse{}, nil)
				return usecase
			},
			wantErr: false,
		},
		{
			name: "(성공) 개인 회원 투자 적합성 정보(investment_suitability) 등록",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   privateUserApi + "/signup/profiles",
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
				   "investment_suitability": {
						"annual_income": "ANNUALINCOME3",
						"asset": "ASSET4",
						"investment_purpose": "ETC",
						"is_pro_invest_asset": true,
						"is_pro_invest_finance_asset": false,
						"is_surplus_fund": true,
						"has_experience": true
					}
				}`,
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_USER"},
				},
			}),
			privateUserUsecase: func() *mocks.PrivateUserUsecase {
				usecase := &mocks.PrivateUserUsecase{}
				usecase.On("RegisterProfile", mock.Anything, udomain.PrvUserRegisterProfileRequest{
					JWTUser: jwthelper.User{
						Id:    int64(1000),
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{
							"ROLE_USER": 1,
						},
					},
					SignUpStatus: domain.SIGNUP_INVESTMENT_SUITABILITY,
					InvestmentSuitability: udomain.PrvUserInvestmentSuitability{
						AnnualIncome:            domain.ANNUALINCOME3,
						Asset:                   domain.ASSET4,
						InvestmentPurpose:       domain.INVESTMENT_PURPOSE_ETC,
						IsProInvestAsset:        lo.ToPtr(true),
						IsProInvestFinanceAsset: lo.ToPtr(false),
						IsSurplusFund:           true,
						HasExperience:           true,
					},
				}).Return(udomain.PrvUserRegisterProfileResponse{}, nil)
				return usecase
			},
			wantErr: false,
		},
		{
			name: "(성공) 개인 회원 약관 동의 등록",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   privateUserApi + "/signup/profiles",
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
			privateUserUsecase: func() *mocks.PrivateUserUsecase {
				usecase := &mocks.PrivateUserUsecase{}
				usecase.On("RegisterProfile", mock.Anything, udomain.PrvUserRegisterProfileRequest{
					JWTUser: jwthelper.User{
						Id:    int64(1000),
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{
							"ROLE_USER": 1,
						},
					},
					SignUpStatus: domain.SIGNUP_AGREE,
				}).Return(udomain.PrvUserRegisterProfileResponse{}, nil)
				return usecase
			},
			wantErr: false,
		},
		{
			name: "(실패) 개인 회원 투자 적합성 정보(investment_suitability) 등록 - 투자자 자격 미달",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   privateUserApi + "/signup/profiles",
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
				   "investment_suitability": {
						"annual_income": "ANNUALINCOME1",
						"asset": "ASSET1",
						"investment_purpose": "ETC",
						"is_pro_invest_asset": true,
						"is_pro_invest_finance_asset": true,
						"is_surplus_fund": true
					}
				}`,
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_USER"},
				},
			}),
			privateUserUsecase: func() *mocks.PrivateUserUsecase {
				usecase := &mocks.PrivateUserUsecase{}
				usecase.On("RegisterProfile", mock.Anything, udomain.PrvUserRegisterProfileRequest{
					JWTUser: jwthelper.User{
						Id:    int64(1000),
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{
							"ROLE_USER": 1,
						},
					},
					SignUpStatus: domain.SIGNUP_INVESTMENT_SUITABILITY,
					InvestmentSuitability: udomain.PrvUserInvestmentSuitability{
						AnnualIncome:            domain.ANNUALINCOME1,
						Asset:                   domain.ASSET1,
						InvestmentPurpose:       domain.INVESTMENT_PURPOSE_ETC,
						IsProInvestAsset:        lo.ToPtr(true),
						IsProInvestFinanceAsset: lo.ToPtr(true),
						IsSurplusFund:           true,
					},
				}).Return(udomain.PrvUserRegisterProfileResponse{}, errors.Wrap(cdomain.ErrServiceUnavailable, "asset and annual income is not valid"))
				return usecase
			},
			wantErr: true,
		},
		{
			name: "(실패) 개인 회원 이용 약관 동의(SIGNUP_AGREE) - eKYC 전송에 필요한 정보 누락",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   privateUserApi + "/signup/profiles",
				Method: http.MethodPut,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Params: map[string]string{
					"id": "1000",
				},
				Body: `{"sign_up_status":"SIGNUP_AGREE"}`,
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_USER"},
				},
			}),
			privateUserUsecase: func() *mocks.PrivateUserUsecase {
				usecase := &mocks.PrivateUserUsecase{}
				usecase.On("RegisterProfile", mock.Anything, udomain.PrvUserRegisterProfileRequest{
					JWTUser: jwthelper.User{
						Id:    int64(1000),
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{
							"ROLE_USER": 1,
						},
					},
					SignUpStatus: domain.SIGNUP_AGREE,
				}).Return(udomain.PrvUserRegisterProfileResponse{}, errors.Wrap(cdomain.ErrProfileIncompleteForEkyc, "profile is incomplete for ekyc: missing family_name,first_name,birthday,province,city,address"))
				return usecase
			},
			wantErr:              true,
			wantFactoryErrorCode: factory.ErrorProfileIncompleteForEkyc.Code,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := privateUserHttpDelivery{
				privateUserUsecase: tt.privateUserUsecase(),
				extractor:          jwthelper.NewUserExtractor(),
			}

			e := echo.New()

			err := delivery.RegisterProfile(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "RegisterProfile() error = %v, wantErr %v", err, tt.wantErr)
			if tt.wantFactoryErrorCode != 0 {
				factoryErr, ok := err.(*factory.Error)
				require.True(t, ok, "error should be *factory.Error, got %T", err)
				assert.Equal(t, tt.wantFactoryErrorCode, factoryErr.Code)
			}
		})
	}
}

func Test_privateUserHttpDelivery_GetSignUpStatus(t *testing.T) {
	tests := []struct {
		name               string
		echoCtx            func(e *echo.Echo) echo.Context
		privateUserUsecase func() *mocks.PrivateUserUsecase
		wantErr            bool
	}{
		{
			name: "(성공) 개인 회원 가입 상태 조회 - 조건 없음",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   privateUserApi + "/signup/step",
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
			privateUserUsecase: func() *mocks.PrivateUserUsecase {
				usecase := &mocks.PrivateUserUsecase{}
				usecase.On("GetSignUpStatus", mock.Anything, udomain.PrvUserGetSignUpStatusRequest{
					JWTUser: jwthelper.User{
						Id:    int64(1000),
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{"ROLE_USER": 1},
					},
				}).Return(udomain.PrvUserGetSignUpStatusResponse{}, nil)
				return usecase
			},
		},
		{
			name: "(성공) 개인 회원 가입 상태 조회 - status=SIGNUP_PRIVATE_PROFILE_BASIC",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   privateUserApi + "/signup/step",
				Method: http.MethodGet,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				QueryParams: map[string][]string{
					"status": {"SIGNUP_PRIVATE_PROFILE_BASIC"},
				},
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_USER"},
				},
			}),
			privateUserUsecase: func() *mocks.PrivateUserUsecase {
				usecase := &mocks.PrivateUserUsecase{}
				usecase.On("GetSignUpStatus", mock.Anything, udomain.PrvUserGetSignUpStatusRequest{
					JWTUser: jwthelper.User{
						Id:    int64(1000),
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{"ROLE_USER": 1},
					},
					Status: domain.SIGNUP_PRIVATE_PROFILE_BASIC,
				}).Return(udomain.PrvUserGetSignUpStatusResponse{}, nil)
				return usecase
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := privateUserHttpDelivery{
				privateUserUsecase: tt.privateUserUsecase(),
				extractor:          jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()
			err := delivery.GetSignUpStatus(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "GetSignUpStatus() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_privateUserHttpDelivery_RegisterMyNumber(t *testing.T) {
	tests := []struct {
		name               string
		echoCtx            func(e *echo.Echo) echo.Context
		privateUserUsecase func() *mocks.PrivateUserUsecase
		wantErr            bool
	}{
		{
			name: "(성공) 개인 회원 마이넘버 등록",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   privateUserApi + "/signup/my-number",
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
				   "my_number": "123456789012",
				   "front_file_url": "front_file_url",
				   "back_file_url": "back_file_url"
				}`,
			}),
			privateUserUsecase: func() *mocks.PrivateUserUsecase {
				usecase := &mocks.PrivateUserUsecase{}
				usecase.On("RegisterMyNumber", mock.Anything, udomain.PrvUserRegisterMyNumberRequest{
					JWTUser: jwthelper.User{
						Id:    int64(1000),
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{"ROLE_USER": 1},
					},
					MyNumber:     "123456789012",
					FrontFileUrl: "front_file_url",
					BackFileUrl:  "back_file_url",
				}).Return(udomain.PrvUserRegisterMyNumberResponse{
					PrvUserMyNumber: udomain.PrvUserMyNumber{
						Id:        1,
						MyNumber:  "123456789012",
						FrontFile: udomain.PrvUserFile{},
						BackFile:  udomain.PrvUserFile{},
						Status:    domain.WAIT_APPROVE,
					},
				}, nil)
				return usecase
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := privateUserHttpDelivery{
				privateUserUsecase: tt.privateUserUsecase(),
				extractor:          jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()
			err := delivery.RegisterMyNumber(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "RegisterMyNumber() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_privateUserHttpDelivery_VerifyWithEkyc(t *testing.T) {
	tests := []struct {
		name               string
		echoCtx            func(e *echo.Echo) echo.Context
		privateUserUsecase func() *mocks.PrivateUserUsecase
		wantErr            bool
	}{
		{
			name: "(성공) eKYC reference ID 발급",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   privateUserApi + "/signup/ekyc",
				Method: http.MethodPost,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Body: `{
					"user_id": 1000
				}`,
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_USER"},
				},
			}),
			privateUserUsecase: func() *mocks.PrivateUserUsecase {
				usecase := &mocks.PrivateUserUsecase{}
				usecase.On("VerifyWithEkyc", mock.Anything, udomain.PrvUserVerifyWithEkycRequest{
					JWTUser: jwthelper.User{
						Id:    int64(1000),
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{"ROLE_USER": 1},
					},
					UserId: 1000,
				}).Return(udomain.PrvUserVerifyWithEkycResponse{ReferenceId: "EKYC-REF"}, nil)
				return usecase
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := privateUserHttpDelivery{
				privateUserUsecase: tt.privateUserUsecase(),
				extractor:          jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()
			err := delivery.VerifyWithEkyc(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "VerifyWithEkyc() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_privateUserHttpDelivery_UpdateEkycStatus(t *testing.T) {
	tests := []struct {
		name               string
		echoCtx            func(e *echo.Echo) echo.Context
		privateUserUsecase func() *mocks.PrivateUserUsecase
		wantErr            bool
	}{
		{
			name: "(성공) eKYC 상태 수정",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   privateUserApi + "/signup/ekyc",
				Method: http.MethodPut,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Body: `{
				   "status": "success",
					"user_id": 1000
				}`,
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_USER"},
				},
			}),
			privateUserUsecase: func() *mocks.PrivateUserUsecase {
				usecase := &mocks.PrivateUserUsecase{}
				usecase.On("UpdateEkycStatus", mock.Anything, udomain.PrvUserUpdateEkycStatusRequest{
					JWTUser: jwthelper.User{
						Id:    int64(1000),
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{"ROLE_USER": 1},
					},
					Status: "success",
					UserId: 1000,
				}).Return(nil)
				return usecase
			},
		},
		{
			name:    "(실패) 본인 이외의 user_id 지정",
			wantErr: true,
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   privateUserApi + "/signup/ekyc",
				Method: http.MethodPut,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Body: `{
				   "status": "success",
					"user_id": 9999
				}`,
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_USER"},
				},
			}),
			privateUserUsecase: func() *mocks.PrivateUserUsecase {
				usecase := &mocks.PrivateUserUsecase{}
				usecase.On("UpdateEkycStatus", mock.Anything, udomain.PrvUserUpdateEkycStatusRequest{
					JWTUser: jwthelper.User{
						Id:    int64(1000),
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{"ROLE_USER": 1},
					},
					Status: "success",
					UserId: 9999,
				}).Return(errors.Wrap(cdomain.ErrForbidden, "cannot update other user's eKYC status"))
				return usecase
			},
		},
		{
			name:    "(실패) eKYC 상태가 이미 완료",
			wantErr: true,
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   privateUserApi + "/signup/ekyc",
				Method: http.MethodPut,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Body: `{
				   "status": "success",
					"user_id": 1000
				}`,
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_USER"},
				},
			}),
			privateUserUsecase: func() *mocks.PrivateUserUsecase {
				usecase := &mocks.PrivateUserUsecase{}
				usecase.On("UpdateEkycStatus", mock.Anything, udomain.PrvUserUpdateEkycStatusRequest{
					JWTUser: jwthelper.User{
						Id:    int64(1000),
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{"ROLE_USER": 1},
					},
					Status: "success",
					UserId: 1000,
				}).Return(errors.Wrap(cdomain.ErrItemConflict, "user's signup status is already SIGNUP_COMPLETE"))
				return usecase
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := privateUserHttpDelivery{
				privateUserUsecase: tt.privateUserUsecase(),
				extractor:          jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()
			err := delivery.UpdateEkycStatus(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "UpdateEkycStatus() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_privateUserHttpDelivery_GetEkycStatus(t *testing.T) {
	tests := []struct {
		name               string
		echoCtx            func(e *echo.Echo) echo.Context
		privateUserUsecase func() *mocks.PrivateUserUsecase
		wantErr            bool
	}{
		{
			name: "(성공) eKYC 상태 조회",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   privateUserApi + "/signup/ekyc/status",
				Method: http.MethodGet,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				QueryParams: map[string][]string{
					"user_id": {"1000"},
				},
			}),
			privateUserUsecase: func() *mocks.PrivateUserUsecase {
				usecase := &mocks.PrivateUserUsecase{}
				usecase.On("GetEkycStatus", mock.Anything, udomain.PrvUserGetEkycStatusRequest{
					UserId: 1000,
				}).Return(udomain.PrvUserGetEkycStatusResponse{
					Status: nil,
				}, nil)
				return usecase
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := privateUserHttpDelivery{
				privateUserUsecase: tt.privateUserUsecase(),
				extractor:          jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()
			err := delivery.GetEkycStatus(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "GetEkycStatus() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_privateUserHttpDelivery_CallEkycWebhook(t *testing.T) {
	tests := []struct {
		name               string
		echoCtx            func(e *echo.Echo) echo.Context
		privateUserUsecase func() *mocks.PrivateUserUsecase
		wantErr            bool
	}{
		{
			name: "eKYC 성공",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   privateUserApi + "/signup/ekyc/webhook",
				Method: http.MethodPost,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Body: `{
					"user_id": 1000,
					"status": "success"
				}`,
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_USER"},
				},
			}),
			privateUserUsecase: func() *mocks.PrivateUserUsecase {
				usecase := &mocks.PrivateUserUsecase{}
				usecase.On("CallEkycWebhook", mock.Anything, udomain.PrvUserCallEkycWebhookRequest{
					Status: "success",
					UserId: 1000,
				}).Return(nil)
				return usecase
			},
			wantErr: false,
		},
		{
			name: "eKYC 실패",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   privateUserApi + "/signup/ekyc/webhook",
				Method: http.MethodPost,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Body: `{
					"user_id": 1000,
					"status": "fail"
				}`,
				Claims: map[string]interface{}{
					"user":  "1000",
					"email": "gildong@email.com",
					"name":  "홍길동",
					"role":  []string{"ROLE_USER"},
				},
			}),
			privateUserUsecase: func() *mocks.PrivateUserUsecase {
				usecase := &mocks.PrivateUserUsecase{}
				usecase.On("CallEkycWebhook", mock.Anything, udomain.PrvUserCallEkycWebhookRequest{
					Status: "fail",
					UserId: 1000,
				}).Return(nil)
				return usecase
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := privateUserHttpDelivery{
				privateUserUsecase: tt.privateUserUsecase(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.CallEkycWebhook(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "CallEkycWebhook() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_privateUserHttpDelivery_GetOne(t *testing.T) {
	tests := []struct {
		name               string
		echoCtx            func(e *echo.Echo) echo.Context
		privateUserUsecase func() *mocks.PrivateUserUsecase
		wantErr            bool
	}{
		{
			name: "GET /api/v1/private-users/detail 상세조회 성공",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   privateUserApi + "/detail",
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
			privateUserUsecase: func() *mocks.PrivateUserUsecase {
				usecase := &mocks.PrivateUserUsecase{}
				usecase.On("GetOne", mock.Anything, udomain.PrvUserGetOneRequest{
					JWTUser: jwthelper.User{
						Id:    int64(1000),
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{"ROLE_USER": 1},
					},
				}).Return(udomain.PrvUserGetOneResponse{}, nil)
				return usecase
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := privateUserHttpDelivery{
				privateUserUsecase: tt.privateUserUsecase(),
				extractor:          jwthelper.NewUserExtractor(),
			}

			e := echo.New()

			err := delivery.GetOne(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "GetOne() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_privateUserHttpDelivery_PatchOne(t *testing.T) {
	tests := []struct {
		name               string
		echoCtx            func(e *echo.Echo) echo.Context
		privateUserUsecase func() *mocks.PrivateUserUsecase
		wantErr            bool
	}{
		{
			name: "PATCH /api/v1/private-users 회원정보 수정 성공 (is_newsletter=true)",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   privateUserApi,
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
			privateUserUsecase: func() *mocks.PrivateUserUsecase {
				usecase := &mocks.PrivateUserUsecase{}
				usecase.On("PatchOne", mock.Anything, udomain.PrvUserPatchOneRequest{
					JWTUser: jwthelper.User{
						Id:    int64(1000),
						Email: "gildong@email.com",
						Name:  "홍길동",
						Roles: map[string]int{"ROLE_USER": 1},
					},
					IsNewsletter: lo.ToPtr(true),
				}).Return(nil)

				return usecase
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := privateUserHttpDelivery{
				privateUserUsecase: tt.privateUserUsecase(),
				extractor:          jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()
			err := delivery.PatchOne(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "PatchOne() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}
