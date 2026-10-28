package private_user

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/segmentio/ksuid"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/idgenerator"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain/mocks"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	udomain "github.com/illuminarean/energyshares-user-api/domain"
)

var privateUserTestJST = func() *time.Location {
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		panic(err)
	}
	return loc
}()

var jsonTimeComparer = cmp.Comparer(func(a, b jsonutils.JSONTime) bool {
	return a.ToString() == b.ToString()
})

func Test_privateUserUsecase_RegisterProfile(t *testing.T) {
	type args struct {
		ctx     context.Context
		request udomain.PrvUserRegisterProfileRequest
	}

	currentJSTDate := time.Now().In(privateUserTestJST)
	tooYoungBirthday := currentJSTDate.AddDate(-17, 0, 0).Format(time.DateOnly)
	tooOldBirthday := currentJSTDate.AddDate(-81, 0, 0).Format(time.DateOnly)
	exactly79Birthday := currentJSTDate.AddDate(-79, 0, 0).Format(time.DateOnly)
	exactly80Birthday := currentJSTDate.AddDate(-80, 0, 0).Format(time.DateOnly)

	tests := []struct {
		name                     string
		args                     args
		idGenerator              func() idgenerator.IdGenerator
		privateUserRepository    func() cdomain.PrivateUserRepository
		userRepository           func() cdomain.UserRepository
		privateProfileRepository func() cdomain.PrivateProfileRepository
		ekycRequestLogRepository func() cdomain.EkycRequestLogRepository
		want                     udomain.PrvUserRegisterProfileResponse
		wantErr                  bool
		wantErrIs                error
	}{
		{
			name: "Success: 基本プロフィール入力",
			args: args{
				ctx: context.Background(),
				request: udomain.PrvUserRegisterProfileRequest{
					JWTUser: jwthelper.User{
						Id: int64(1000),
					},
					BasicProfile: udomain.PrvUserBasicProfile{
						FirstName:      "first name",
						FamilyName:     "family name",
						FirstNameKana:  "first name kana",
						FamilyNameKana: "family name kana",
						Birthday:       "1999-01-01",
						Nationality:    domain.JP,
						Gender:         domain.MALE,
					},
					SignUpStatus: domain.SIGNUP_PRIVATE_PROFILE_BASIC,
				},
			},
			privateUserRepository: func() cdomain.PrivateUserRepository {
				repo := &mocks.PrivateUserRepository{}
				repo.On("GetById", context.Background(), int64(1000)).Return(domain.PrivateUser{
					User: domain.User{
						Id:           int64(1000),
						SignupStatus: domain.SIGNUP_CORPORATE_PROFILE_BASIC,
					},
				}, nil)
				return repo
			},
			userRepository: func() cdomain.UserRepository {
				repo := &mocks.UserRepository{}
				repo.On("Update", context.Background(), domain.User{
					Id:           int64(1000),
					SignupStatus: domain.SIGNUP_PRIVATE_PROFILE_BASIC,
					Name:         "family name first name",
				}).Return(nil)
				return repo
			},
			privateProfileRepository: func() cdomain.PrivateProfileRepository {
				repo := &mocks.PrivateProfileRepository{}
				expected := domain.PrivateProfile{
					FirstName:      "first name",
					FamilyName:     "family name",
					FirstNameKana:  "first name kana",
					FamilyNameKana: "family name kana",
					Birthday:       jsonutils.JSONTime(time.Date(1999, 1, 1, 0, 0, 0, 0, privateUserTestJST)),
					Nationality:    domain.JP,
					Gender:         domain.MALE,
				}
				repo.
					On("Update", context.Background(),
						mock.MatchedBy(func(profile domain.PrivateProfile) bool {
							return cmp.Equal(profile, expected, jsonTimeComparer)
						})).
					Return(nil)
				return repo
			},
		},
		{
			name: "Success: その他プロフィール入力",
			args: args{
				ctx: context.Background(),
				request: udomain.PrvUserRegisterProfileRequest{
					JWTUser: jwthelper.User{
						Id: int64(1000),
					},
					OtherProfile: udomain.PrvUserOtherProfile{
						ZipCode:     "1234567",
						Province:    "도도부현",
						City:        "시군구",
						Address:     "상세주소",
						Address2:    lo.ToPtr(""),
						MobilePhone: "01000000000",
						Job:         domain.USER_LISTED_COMPANY_EMPLOYEE,
					},
					SignUpStatus: domain.SIGNUP_PRIVATE_PROFILE_OTHER,
				},
			},
			privateUserRepository: func() cdomain.PrivateUserRepository {
				repo := &mocks.PrivateUserRepository{}
				repo.On("GetById", context.Background(), int64(1000)).Return(domain.PrivateUser{
					User: domain.User{
						Id:           int64(1000),
						SignupStatus: domain.SIGNUP_PRIVATE_PROFILE_BASIC,
					},
				}, nil)
				return repo
			},
			userRepository: func() cdomain.UserRepository {
				repo := &mocks.UserRepository{}
				repo.On("Update", context.Background(), domain.User{
					Id:           int64(1000),
					ZipCode:      "1234567",
					Province:     "도도부현",
					City:         "시군구",
					Address:      "상세주소",
					Address2:     nil,
					MobilePhone:  "01000000000",
					SignupStatus: domain.SIGNUP_PRIVATE_PROFILE_OTHER,
				}).Return(nil)
				return repo
			},
			privateProfileRepository: func() cdomain.PrivateProfileRepository {
				repo := &mocks.PrivateProfileRepository{}
				repo.On("Update", context.Background(), domain.PrivateProfile{
					Job: domain.USER_LISTED_COMPANY_EMPLOYEE,
				}).Return(nil)
				return repo
			},
		},
		{
			name: "Success: 口座入力",
			args: args{
				ctx: context.Background(),
				request: udomain.PrvUserRegisterProfileRequest{
					JWTUser: jwthelper.User{
						Id: int64(1000),
					},
					WithdrawalAccount: udomain.PrvUserWithdrawalAccount{
						BankName:    "은행명",
						BranchName:  "지점명",
						BankCode:    "123",
						BranchCode:  "456",
						DepositType: domain.ORDINARY_ACCOUNT,
						BankAccount: "1234567",
						Depositor:   "ホン・ギルドン",
					},
					SignUpStatus: domain.SIGNUP_ACCOUNT,
				},
			},
			privateUserRepository: func() cdomain.PrivateUserRepository {
				repo := &mocks.PrivateUserRepository{}
				repo.On("GetById", context.Background(), int64(1000)).Return(domain.PrivateUser{
					User: domain.User{
						Id:           int64(1000),
						SignupStatus: domain.SIGNUP_PRIVATE_PROFILE_OTHER,
					},
				}, nil)
				return repo
			},
			userRepository: func() cdomain.UserRepository {
				repo := &mocks.UserRepository{}
				repo.On("Update", context.Background(), domain.User{
					Id:           int64(1000),
					SignupStatus: domain.SIGNUP_ACCOUNT,
					BankName:     "은행명",
					BranchName:   "지점명",
					BankCode:     lo.ToPtr("123"),
					BranchCode:   lo.ToPtr("456"),
					DepositType:  domain.ORDINARY_ACCOUNT,
					BankAccount:  "1234567",
					Depositor:    "ホン・ギルドン",
				}).Return(nil)
				return repo
			},
			privateProfileRepository: func() cdomain.PrivateProfileRepository {
				repo := &mocks.PrivateProfileRepository{}
				repo.On("Update", context.Background(), domain.PrivateProfile{}).Return(nil)
				return repo
			},
		},
		{
			name: "Success: 投資適格性入力",
			args: args{
				ctx: context.Background(),
				request: udomain.PrvUserRegisterProfileRequest{
					JWTUser: jwthelper.User{
						Id: int64(1000),
					},
					InvestmentSuitability: udomain.PrvUserInvestmentSuitability{
						Asset:                   domain.ASSET3,
						AnnualIncome:            domain.ANNUALINCOME3,
						IsProInvestFinanceAsset: lo.ToPtr(true),
						IsProInvestAsset:        lo.ToPtr(true),
						InvestmentPurpose:       domain.VARIANCE_PORTFOLIO,
						HasExperience:           true,
						IsSurplusFund:           true,
					},
					SignUpStatus: domain.SIGNUP_INVESTMENT_SUITABILITY,
				},
			},
			privateUserRepository: func() cdomain.PrivateUserRepository {
				repo := &mocks.PrivateUserRepository{}
				repo.On("GetById", context.Background(), int64(1000)).Return(domain.PrivateUser{
					User: domain.User{
						Id:           int64(1000),
						SignupStatus: domain.SIGNUP_ACCOUNT,
					},
				}, nil)
				return repo
			},
			userRepository: func() cdomain.UserRepository {
				repo := &mocks.UserRepository{}
				repo.On("Update", context.Background(), domain.User{
					Id:                int64(1000),
					SignupStatus:      domain.SIGNUP_INVESTMENT_SUITABILITY,
					InvestmentPurpose: domain.VARIANCE_PORTFOLIO,
					HasExperience:     lo.ToPtr(true),
					IsSurplusFund:     lo.ToPtr(true),
				}).Return(nil)
				return repo
			},
			privateProfileRepository: func() cdomain.PrivateProfileRepository {
				repo := &mocks.PrivateProfileRepository{}
				repo.On("Update", context.Background(), domain.PrivateProfile{
					Asset:                   domain.ASSET3,
					AnnualIncome:            domain.ANNUALINCOME3,
					IsProInvestFinanceAsset: lo.ToPtr(true),
					IsProInvestAsset:        lo.ToPtr(true),
				}).Return(nil)
				return repo
			},
		},
		{
			// MIR-1121: 以前はプロフィールが空(氏名・生年月日・住所いずれも未入力)
			// でもSIGNUP_AGREEが成功し、AuditStatusがWAIT_IDENTITY_REGISTER_AUDIT
			// まで進んでいた。この状態でフロントエンドがeKYCへ空欄のまま遷移し、
			// 本人確認会社への連携に失敗して会員が滞留する事故につながっていた。
			name: "Failed: eKYC送信に必要な情報(氏名・生年月日・住所)が欠けている場合は状態遷移せず弾く",
			args: args{
				ctx: context.Background(),
				request: udomain.PrvUserRegisterProfileRequest{
					JWTUser: jwthelper.User{
						Id: int64(1000),
					},
					SignUpStatus: domain.SIGNUP_AGREE,
				},
			},
			privateUserRepository: func() cdomain.PrivateUserRepository {
				repo := &mocks.PrivateUserRepository{}
				repo.On("GetById", context.Background(), int64(1000)).Return(domain.PrivateUser{
					User: domain.User{
						Id:           int64(1000),
						SignupStatus: domain.SIGNUP_INVESTMENT_SUITABILITY,
					},
				}, nil)
				return repo
			},
			userRepository: func() cdomain.UserRepository {
				return &mocks.UserRepository{}
			},
			privateProfileRepository: func() cdomain.PrivateProfileRepository {
				return &mocks.PrivateProfileRepository{}
			},
			wantErr:   true,
			wantErrIs: cdomain.ErrProfileIncompleteForEkyc,
		},
		{
			// MIR-1121: eKYC送信の値をブラウザのローカル状態やuserの非同期取得
			// ではなく、このレスポンス自体に含めて返すようにした。GetById直後の
			// privateUserの氏名・生年月日・住所がそのままレスポンスへ渡ること
			// を固定する。
			name: "Success: 利用規約同意時にeKYC送信用の氏名・生年月日・住所をレスポンスへ含める",
			args: args{
				ctx: context.Background(),
				request: udomain.PrvUserRegisterProfileRequest{
					JWTUser: jwthelper.User{
						Id: int64(1000),
					},
					SignUpStatus: domain.SIGNUP_AGREE,
				},
			},
			privateUserRepository: func() cdomain.PrivateUserRepository {
				repo := &mocks.PrivateUserRepository{}
				repo.On("GetById", context.Background(), int64(1000)).Return(domain.PrivateUser{
					User: domain.User{
						Id:           int64(1000),
						SignupStatus: domain.SIGNUP_INVESTMENT_SUITABILITY,
						Name:         "山田 太郎",
						Province:     "東京都",
						City:         "渋谷区",
						Address:      "道玄坂1-2-3",
						Address2:     lo.ToPtr("〇〇ビル401"),
					},
					PrivateProfile: domain.PrivateProfile{
						Birthday:   jsonutils.JSONTime(time.Date(1990, 1, 2, 0, 0, 0, 0, time.UTC)),
						FamilyName: "山田",
						FirstName:  "太郎",
					},
				}, nil)
				return repo
			},
			userRepository: func() cdomain.UserRepository {
				repo := &mocks.UserRepository{}
				repo.On("Update", context.Background(), domain.User{
					Id:           int64(1000),
					SignupStatus: domain.SIGNUP_AGREE,
					AuditStatus:  domain.WAIT_IDENTITY_REGISTER_AUDIT,
					Name:         "山田 太郎",
					Province:     "東京都",
					City:         "渋谷区",
					Address:      "道玄坂1-2-3",
					Address2:     lo.ToPtr("〇〇ビル401"),
				}).Return(nil)
				return repo
			},
			privateProfileRepository: func() cdomain.PrivateProfileRepository {
				repo := &mocks.PrivateProfileRepository{}
				repo.On("Update", context.Background(), domain.PrivateProfile{
					Birthday:   jsonutils.JSONTime(time.Date(1990, 1, 2, 0, 0, 0, 0, time.UTC)),
					FamilyName: "山田",
					FirstName:  "太郎",
				}).Return(nil)
				return repo
			},
			idGenerator: func() idgenerator.IdGenerator {
				mockIdGenerator := &mocks.IdGenerator{}
				mockIdGenerator.On("GenKsuid").Return(ksuid.KSUID{})
				return mockIdGenerator
			},
			ekycRequestLogRepository: func() cdomain.EkycRequestLogRepository {
				repo := &mocks.EkycRequestLogRepository{}
				repo.On("GetByUserId", context.Background(), int64(1000)).Return(domain.EkycRequestLog{
					Status: domain.EKYC_REQUEST_STATUS_SUCCESS,
				}, nil)
				repo.On("Create", context.Background(), mock.MatchedBy(func(log domain.EkycRequestLog) bool {
					return log.UserId == int64(1000)
				})).Return(nil)
				return repo
			},
			want: udomain.PrvUserRegisterProfileResponse{
				EKycReferenceId: ksuid.KSUID{}.String(),
				FamilyName:      "山田",
				FirstName:       "太郎",
				Birthday:        "1990-01-02",
				Address:         "東京都渋谷区道玄坂1-2-3 〇〇ビル401",
			},
		},
		{
			name: "Success: 本登録保留のユーザーは完了状態が残っていても再入力できる",
			args: args{
				ctx: context.Background(),
				request: udomain.PrvUserRegisterProfileRequest{
					JWTUser: jwthelper.User{
						Id: int64(1000),
					},
					BasicProfile: udomain.PrvUserBasicProfile{
						FirstName:      "first name",
						FamilyName:     "family name",
						FirstNameKana:  "first name kana",
						FamilyNameKana: "family name kana",
						Birthday:       "1999-01-01",
						Nationality:    domain.JP,
						Gender:         domain.MALE,
					},
					SignUpStatus: domain.SIGNUP_PRIVATE_PROFILE_BASIC,
				},
			},
			// 管理画面から保留にした場合、signup_status は完了のまま残る。
			// この状態を弾くと再登録できなくなるため通す必要がある。
			privateUserRepository: func() cdomain.PrivateUserRepository {
				repo := &mocks.PrivateUserRepository{}
				repo.On("GetById", context.Background(), int64(1000)).Return(domain.PrivateUser{
					User: domain.User{
						Id:           int64(1000),
						SignupStatus: domain.SIGNUP_COMPLETE,
						AuditStatus:  domain.HOLD_AUDIT,
					},
				}, nil)
				return repo
			},
			userRepository: func() cdomain.UserRepository {
				repo := &mocks.UserRepository{}
				repo.On("Update", context.Background(), mock.Anything).Return(nil)
				return repo
			},
			privateProfileRepository: func() cdomain.PrivateProfileRepository {
				repo := &mocks.PrivateProfileRepository{}
				repo.On("Update", context.Background(), mock.Anything).Return(nil)
				return repo
			},
			want:    udomain.PrvUserRegisterProfileResponse{},
			wantErr: false,
		},
		{
			name: "Success: 仮登録へ差し戻されたユーザーは完了状態が残っていても再入力できる",
			args: args{
				ctx: context.Background(),
				request: udomain.PrvUserRegisterProfileRequest{
					JWTUser: jwthelper.User{
						Id: int64(1000),
					},
					BasicProfile: udomain.PrvUserBasicProfile{
						FirstName:      "first name",
						FamilyName:     "family name",
						FirstNameKana:  "first name kana",
						FamilyNameKana: "family name kana",
						Birthday:       "1999-01-01",
						Nationality:    domain.JP,
						Gender:         domain.MALE,
					},
					SignUpStatus: domain.SIGNUP_PRIVATE_PROFILE_BASIC,
				},
			},
			// 管理画面から仮登録へ差し戻した場合も signup_status は完了のまま残る。
			// 最初から入力し直させる操作なので通す必要がある。
			privateUserRepository: func() cdomain.PrivateUserRepository {
				repo := &mocks.PrivateUserRepository{}
				repo.On("GetById", context.Background(), int64(1000)).Return(domain.PrivateUser{
					User: domain.User{
						Id:           int64(1000),
						SignupStatus: domain.SIGNUP_COMPLETE,
						AuditStatus:  domain.PROVISIONAL_AUDIT,
					},
				}, nil)
				return repo
			},
			userRepository: func() cdomain.UserRepository {
				repo := &mocks.UserRepository{}
				repo.On("Update", context.Background(), mock.Anything).Return(nil)
				return repo
			},
			privateProfileRepository: func() cdomain.PrivateProfileRepository {
				repo := &mocks.PrivateProfileRepository{}
				repo.On("Update", context.Background(), mock.Anything).Return(nil)
				return repo
			},
			want:    udomain.PrvUserRegisterProfileResponse{},
			wantErr: false,
		},
		// 異常系
		{
			name: "Failed: 本登録完了済みのユーザーはプロフィール入力で状態を巻き戻せない",
			args: args{
				ctx: context.Background(),
				request: udomain.PrvUserRegisterProfileRequest{
					JWTUser: jwthelper.User{Id: int64(1000)},
					BasicProfile: udomain.PrvUserBasicProfile{
						FirstName:      "first name",
						FamilyName:     "family name",
						FirstNameKana:  "first name kana",
						FamilyNameKana: "family name kana",
						Birthday:       "1999-01-01",
						Nationality:    domain.JP,
						Gender:         domain.MALE,
					},
					SignUpStatus: domain.SIGNUP_PRIVATE_PROFILE_BASIC,
				},
			},
			privateUserRepository: func() cdomain.PrivateUserRepository {
				repo := &mocks.PrivateUserRepository{}
				repo.On("GetById", context.Background(), int64(1000)).Return(domain.PrivateUser{
					User: domain.User{
						Id:           int64(1000),
						SignupStatus: domain.SIGNUP_COMPLETE,
					},
				}, nil)
				return repo
			},
			// ガードが無ければ更新まで到達して成功してしまう構成にしている
			userRepository: func() cdomain.UserRepository {
				repo := &mocks.UserRepository{}
				repo.On("Update", context.Background(), mock.Anything).Return(nil)
				return repo
			},
			privateProfileRepository: func() cdomain.PrivateProfileRepository {
				repo := &mocks.PrivateProfileRepository{}
				repo.On("Update", context.Background(), mock.Anything).Return(nil)
				return repo
			},
			wantErr: true,
		},
		{
			name: "Failed: 本登録承認完了済みのユーザーは審査状態を巻き戻せない",
			args: args{
				ctx: context.Background(),
				request: udomain.PrvUserRegisterProfileRequest{
					JWTUser: jwthelper.User{
						Id: int64(1000),
					},
					InvestmentSuitability: udomain.PrvUserInvestmentSuitability{
						Asset:                   domain.ASSET3,
						AnnualIncome:            domain.ANNUALINCOME3,
						IsProInvestFinanceAsset: lo.ToPtr(true),
						IsProInvestAsset:        lo.ToPtr(true),
						InvestmentPurpose:       domain.VARIANCE_PORTFOLIO,
						HasExperience:           true,
						IsSurplusFund:           true,
					},
					SignUpStatus: domain.SIGNUP_INVESTMENT_SUITABILITY,
				},
			},
			privateUserRepository: func() cdomain.PrivateUserRepository {
				repo := &mocks.PrivateUserRepository{}
				repo.On("GetById", context.Background(), int64(1000)).Return(domain.PrivateUser{
					User: domain.User{
						Id:           int64(1000),
						SignupStatus: domain.SIGNUP_ACCOUNT,
						AuditStatus:  domain.COMPLETED_AUDIT,
					},
				}, nil)
				return repo
			},
			// ガードが無ければ更新まで到達して成功してしまう構成にしている
			userRepository: func() cdomain.UserRepository {
				repo := &mocks.UserRepository{}
				repo.On("Update", context.Background(), mock.Anything).Return(nil)
				return repo
			},
			privateProfileRepository: func() cdomain.PrivateProfileRepository {
				repo := &mocks.PrivateProfileRepository{}
				repo.On("Update", context.Background(), mock.Anything).Return(nil)
				return repo
			},
			wantErr: true,
		},
		{
			name: "Failed: 無効化されたユーザーは再入力できない",
			args: args{
				ctx: context.Background(),
				request: udomain.PrvUserRegisterProfileRequest{
					JWTUser: jwthelper.User{Id: int64(1000)},
					BasicProfile: udomain.PrvUserBasicProfile{
						FirstName:      "first name",
						FamilyName:     "family name",
						FirstNameKana:  "first name kana",
						FamilyNameKana: "family name kana",
						Birthday:       "1999-01-01",
						Nationality:    domain.JP,
						Gender:         domain.MALE,
					},
					SignUpStatus: domain.SIGNUP_PRIVATE_PROFILE_BASIC,
				},
			},
			// signup_status が未完了のままでも、無効化されたユーザーは再入力できない。
			privateUserRepository: func() cdomain.PrivateUserRepository {
				repo := &mocks.PrivateUserRepository{}
				repo.On("GetById", context.Background(), int64(1000)).Return(domain.PrivateUser{
					User: domain.User{
						Id:           int64(1000),
						SignupStatus: domain.SIGNUP_KYC,
						AuditStatus:  domain.INVALIDITY_AUDIT,
					},
				}, nil)
				return repo
			},
			// ガードが無ければ更新まで到達して成功してしまう構成にしている
			userRepository: func() cdomain.UserRepository {
				repo := &mocks.UserRepository{}
				repo.On("Update", context.Background(), mock.Anything).Return(nil)
				return repo
			},
			privateProfileRepository: func() cdomain.PrivateProfileRepository {
				repo := &mocks.PrivateProfileRepository{}
				repo.On("Update", context.Background(), mock.Anything).Return(nil)
				return repo
			},
			wantErr: true,
		},
		{
			name: "Failed: 最終承認待機中のユーザーは再入力できない",
			args: args{
				ctx: context.Background(),
				request: udomain.PrvUserRegisterProfileRequest{
					JWTUser: jwthelper.User{Id: int64(1000)},
					BasicProfile: udomain.PrvUserBasicProfile{
						FirstName:      "first name",
						FamilyName:     "family name",
						FirstNameKana:  "first name kana",
						FamilyNameKana: "family name kana",
						Birthday:       "1999-01-01",
						Nationality:    domain.JP,
						Gender:         domain.MALE,
					},
					SignUpStatus: domain.SIGNUP_PRIVATE_PROFILE_BASIC,
				},
			},
			privateUserRepository: func() cdomain.PrivateUserRepository {
				repo := &mocks.PrivateUserRepository{}
				repo.On("GetById", context.Background(), int64(1000)).Return(domain.PrivateUser{
					User: domain.User{
						Id:           int64(1000),
						SignupStatus: domain.SIGNUP_KYC,
						AuditStatus:  domain.WAIT_FINAL_APPROVAL_AUDIT,
					},
				}, nil)
				return repo
			},
			// ガードが無ければ更新まで到達して成功してしまう構成にしている
			userRepository: func() cdomain.UserRepository {
				repo := &mocks.UserRepository{}
				repo.On("Update", context.Background(), mock.Anything).Return(nil)
				return repo
			},
			privateProfileRepository: func() cdomain.PrivateProfileRepository {
				repo := &mocks.PrivateProfileRepository{}
				repo.On("Update", context.Background(), mock.Anything).Return(nil)
				return repo
			},
			wantErr: true,
		},
		{
			name: "Failed: 反社会的勢力の疑いで待機中のユーザーは再入力できない",
			args: args{
				ctx: context.Background(),
				request: udomain.PrvUserRegisterProfileRequest{
					JWTUser: jwthelper.User{Id: int64(1000)},
					BasicProfile: udomain.PrvUserBasicProfile{
						FirstName:      "first name",
						FamilyName:     "family name",
						FirstNameKana:  "first name kana",
						FamilyNameKana: "family name kana",
						Birthday:       "1999-01-01",
						Nationality:    domain.JP,
						Gender:         domain.MALE,
					},
					SignUpStatus: domain.SIGNUP_PRIVATE_PROFILE_BASIC,
				},
			},
			privateUserRepository: func() cdomain.PrivateUserRepository {
				repo := &mocks.PrivateUserRepository{}
				repo.On("GetById", context.Background(), int64(1000)).Return(domain.PrivateUser{
					User: domain.User{
						Id:           int64(1000),
						SignupStatus: domain.SIGNUP_KYC,
						AuditStatus:  domain.WAIT_ANTISOCIAL_AUDIT,
					},
				}, nil)
				return repo
			},
			// ガードが無ければ更新まで到達して成功してしまう構成にしている
			userRepository: func() cdomain.UserRepository {
				repo := &mocks.UserRepository{}
				repo.On("Update", context.Background(), mock.Anything).Return(nil)
				return repo
			},
			privateProfileRepository: func() cdomain.PrivateProfileRepository {
				repo := &mocks.PrivateProfileRepository{}
				repo.On("Update", context.Background(), mock.Anything).Return(nil)
				return repo
			},
			wantErr: true,
		},
		{
			name: "Failed: 反社会的勢力の疑いを確認中のユーザーは再入力できない",
			args: args{
				ctx: context.Background(),
				request: udomain.PrvUserRegisterProfileRequest{
					JWTUser: jwthelper.User{Id: int64(1000)},
					BasicProfile: udomain.PrvUserBasicProfile{
						FirstName:      "first name",
						FamilyName:     "family name",
						FirstNameKana:  "first name kana",
						FamilyNameKana: "family name kana",
						Birthday:       "1999-01-01",
						Nationality:    domain.JP,
						Gender:         domain.MALE,
					},
					SignUpStatus: domain.SIGNUP_PRIVATE_PROFILE_BASIC,
				},
			},
			privateUserRepository: func() cdomain.PrivateUserRepository {
				repo := &mocks.PrivateUserRepository{}
				repo.On("GetById", context.Background(), int64(1000)).Return(domain.PrivateUser{
					User: domain.User{
						Id:           int64(1000),
						SignupStatus: domain.SIGNUP_KYC,
						AuditStatus:  domain.CHECK_ANTISOCIAL_AUDIT,
					},
				}, nil)
				return repo
			},
			// ガードが無ければ更新まで到達して成功してしまう構成にしている
			userRepository: func() cdomain.UserRepository {
				repo := &mocks.UserRepository{}
				repo.On("Update", context.Background(), mock.Anything).Return(nil)
				return repo
			},
			privateProfileRepository: func() cdomain.PrivateProfileRepository {
				repo := &mocks.PrivateProfileRepository{}
				repo.On("Update", context.Background(), mock.Anything).Return(nil)
				return repo
			},
			wantErr: true,
		},
		{
			// WAIT_IDENTITY_REGISTER_AUDIT(本人確認待ち)は SIGNUP_AGREE 処理自体が
			// signup_status と同時に付与する状態(この関数のすぐ下、SIGNUP_AGREE
			// 分岐を参照)。ekyc 再検証などで SIGNUP_AGREE を再送する場合、DB上は
			// 既に SignupStatus=SIGNUP_AGREE / AuditStatus=WAIT_IDENTITY_REGISTER_AUDIT
			// になっている。ここを弾くと本人確認の再試行そのものが壊れるため、
			// 意図して対象外にしている。
			name: "Success: 本人確認待ちのユーザーはSIGNUP_AGREEを再送できる",
			args: args{
				ctx: context.Background(),
				request: udomain.PrvUserRegisterProfileRequest{
					JWTUser: jwthelper.User{
						Id: int64(1000),
					},
					SignUpStatus: domain.SIGNUP_AGREE,
				},
			},
			privateUserRepository: func() cdomain.PrivateUserRepository {
				repo := &mocks.PrivateUserRepository{}
				repo.On("GetById", context.Background(), int64(1000)).Return(domain.PrivateUser{
					User: domain.User{
						Id:           int64(1000),
						SignupStatus: domain.SIGNUP_AGREE,
						AuditStatus:  domain.WAIT_IDENTITY_REGISTER_AUDIT,
						Province:     "東京都",
						City:         "渋谷区",
						Address:      "道玄坂1-2-3",
					},
					PrivateProfile: domain.PrivateProfile{
						Birthday:   jsonutils.JSONTime(time.Date(1990, 1, 2, 0, 0, 0, 0, time.UTC)),
						FamilyName: "山田",
						FirstName:  "太郎",
					},
				}, nil)
				return repo
			},
			userRepository: func() cdomain.UserRepository {
				repo := &mocks.UserRepository{}
				repo.On("Update", context.Background(), domain.User{
					Id:           int64(1000),
					SignupStatus: domain.SIGNUP_AGREE,
					AuditStatus:  domain.WAIT_IDENTITY_REGISTER_AUDIT,
					Province:     "東京都",
					City:         "渋谷区",
					Address:      "道玄坂1-2-3",
				}).Return(nil)
				return repo
			},
			privateProfileRepository: func() cdomain.PrivateProfileRepository {
				repo := &mocks.PrivateProfileRepository{}
				repo.On("Update", context.Background(), domain.PrivateProfile{
					Birthday:   jsonutils.JSONTime(time.Date(1990, 1, 2, 0, 0, 0, 0, time.UTC)),
					FamilyName: "山田",
					FirstName:  "太郎",
				}).Return(nil)
				return repo
			},
			idGenerator: func() idgenerator.IdGenerator {
				mockIdGenerator := &mocks.IdGenerator{}
				mockIdGenerator.On("GenKsuid").Return(ksuid.KSUID{})
				return mockIdGenerator
			},
			ekycRequestLogRepository: func() cdomain.EkycRequestLogRepository {
				repo := &mocks.EkycRequestLogRepository{}
				repo.On("GetByUserId", context.Background(), int64(1000)).Return(domain.EkycRequestLog{
					Status: domain.EKYC_REQUEST_STATUS_SUCCESS,
				}, nil)
				repo.On("Create", context.Background(), mock.MatchedBy(func(log domain.EkycRequestLog) bool {
					return log.UserId == int64(1000)
				})).Return(nil)
				return repo
			},
			want: udomain.PrvUserRegisterProfileResponse{
				EKycReferenceId: ksuid.KSUID{}.String(),
				FamilyName:      "山田",
				FirstName:       "太郎",
				Birthday:        "1990-01-02",
				Address:         "東京都渋谷区道玄坂1-2-3",
			},
			wantErr: false,
		},
		{
			// WAIT_AUDIT(本登録承認待機)を今回のブロック対象に含めるかは
			// papu-nika さんへ PR で確認中(現時点では対象外のまま)。
			// 決定したらこのテストを更新する。
			name: "Success: 本登録承認待機中のユーザーは今回のスコープでは再入力できる(要確認事項)",
			args: args{
				ctx: context.Background(),
				request: udomain.PrvUserRegisterProfileRequest{
					JWTUser: jwthelper.User{Id: int64(1000)},
					BasicProfile: udomain.PrvUserBasicProfile{
						FirstName:      "first name",
						FamilyName:     "family name",
						FirstNameKana:  "first name kana",
						FamilyNameKana: "family name kana",
						Birthday:       "1999-01-01",
						Nationality:    domain.JP,
						Gender:         domain.MALE,
					},
					SignUpStatus: domain.SIGNUP_PRIVATE_PROFILE_BASIC,
				},
			},
			privateUserRepository: func() cdomain.PrivateUserRepository {
				repo := &mocks.PrivateUserRepository{}
				repo.On("GetById", context.Background(), int64(1000)).Return(domain.PrivateUser{
					User: domain.User{
						Id:           int64(1000),
						SignupStatus: domain.SIGNUP_KYC,
						AuditStatus:  domain.WAIT_AUDIT,
					},
				}, nil)
				return repo
			},
			userRepository: func() cdomain.UserRepository {
				repo := &mocks.UserRepository{}
				repo.On("Update", context.Background(), mock.Anything).Return(nil)
				return repo
			},
			privateProfileRepository: func() cdomain.PrivateProfileRepository {
				repo := &mocks.PrivateProfileRepository{}
				repo.On("Update", context.Background(), mock.Anything).Return(nil)
				return repo
			},
			want:    udomain.PrvUserRegisterProfileResponse{},
			wantErr: false,
		},
		{
			// 移行前レコードや想定外の値が audit_status に入っていても、
			// signup_status が未完了なら現状は素通りする(既存の挙動、今回は
			// 変更しない)。将来ここを閉じる場合は要検討。
			name: "Success: 想定外のAuditStatus値でもsignup_status未完了なら現状は素通りする",
			args: args{
				ctx: context.Background(),
				request: udomain.PrvUserRegisterProfileRequest{
					JWTUser: jwthelper.User{Id: int64(1000)},
					BasicProfile: udomain.PrvUserBasicProfile{
						FirstName:      "first name",
						FamilyName:     "family name",
						FirstNameKana:  "first name kana",
						FamilyNameKana: "family name kana",
						Birthday:       "1999-01-01",
						Nationality:    domain.JP,
						Gender:         domain.MALE,
					},
					SignUpStatus: domain.SIGNUP_PRIVATE_PROFILE_BASIC,
				},
			},
			privateUserRepository: func() cdomain.PrivateUserRepository {
				repo := &mocks.PrivateUserRepository{}
				repo.On("GetById", context.Background(), int64(1000)).Return(domain.PrivateUser{
					User: domain.User{
						Id:           int64(1000),
						SignupStatus: domain.SIGNUP_KYC,
						AuditStatus:  domain.AUDIT_STATUS(""),
					},
				}, nil)
				return repo
			},
			userRepository: func() cdomain.UserRepository {
				repo := &mocks.UserRepository{}
				repo.On("Update", context.Background(), mock.Anything).Return(nil)
				return repo
			},
			privateProfileRepository: func() cdomain.PrivateProfileRepository {
				repo := &mocks.PrivateProfileRepository{}
				repo.On("Update", context.Background(), mock.Anything).Return(nil)
				return repo
			},
			want:    udomain.PrvUserRegisterProfileResponse{},
			wantErr: false,
		},
		{
			name: "Failed: ユーザー情報取得失敗",
			args: args{
				ctx: context.Background(),
				request: udomain.PrvUserRegisterProfileRequest{
					JWTUser:      jwthelper.User{Id: int64(1000)},
					SignUpStatus: domain.SIGNUP_PRIVATE_PROFILE_BASIC,
				},
			},
			privateUserRepository: func() cdomain.PrivateUserRepository {
				repo := &mocks.PrivateUserRepository{}
				repo.On("GetById", context.Background(), int64(1000)).Return(domain.PrivateUser{}, errors.New("find error"))
				return repo
			},
			userRepository: func() cdomain.UserRepository {
				return &mocks.UserRepository{}
			},
			privateProfileRepository: func() cdomain.PrivateProfileRepository {
				return &mocks.PrivateProfileRepository{}
			},
			wantErr: true,
		},
		{
			name: "Failed: 登録段階の不整合",
			args: args{
				ctx: context.Background(),
				request: udomain.PrvUserRegisterProfileRequest{
					JWTUser:      jwthelper.User{Id: int64(1000)},
					SignUpStatus: domain.SIGNUP_INVESTMENT_SUITABILITY,
				},
			},
			privateUserRepository: func() cdomain.PrivateUserRepository {
				repo := &mocks.PrivateUserRepository{}
				repo.On("GetById", context.Background(), int64(1000)).Return(domain.PrivateUser{
					User: domain.User{
						Id:           int64(1000),
						SignupStatus: domain.SIGNUP_PRIVATE_PROFILE_BASIC,
					},
				}, nil)
				return repo
			},
			userRepository: func() cdomain.UserRepository {
				return &mocks.UserRepository{}
			},
			privateProfileRepository: func() cdomain.PrivateProfileRepository {
				return &mocks.PrivateProfileRepository{}
			},
			wantErr: true,
		},
		{
			name: "Failed: 年齢フォーマットエラー",
			args: args{
				ctx: context.Background(),
				request: udomain.PrvUserRegisterProfileRequest{
					JWTUser: jwthelper.User{Id: int64(1000)},
					BasicProfile: udomain.PrvUserBasicProfile{
						FirstName:      "first",
						FamilyName:     "family",
						FirstNameKana:  "first kana",
						FamilyNameKana: "family kana",
						Birthday:       "invalid-date",
						Nationality:    domain.JP,
						Gender:         domain.MALE,
					},
					SignUpStatus: domain.SIGNUP_PRIVATE_PROFILE_BASIC,
				},
			},
			privateUserRepository: func() cdomain.PrivateUserRepository {
				repo := &mocks.PrivateUserRepository{}
				repo.On("GetById", context.Background(), int64(1000)).Return(domain.PrivateUser{
					User: domain.User{
						Id:           int64(1000),
						SignupStatus: domain.SIGNUP_NONE,
					},
				}, nil)
				return repo
			},
			userRepository: func() cdomain.UserRepository {
				return &mocks.UserRepository{}
			},
			privateProfileRepository: func() cdomain.PrivateProfileRepository {
				return &mocks.PrivateProfileRepository{}
			},
			wantErr: true,
		},
		{
			name: "Failed: 年齢チェックNG（下限NG）",
			args: args{
				ctx: context.Background(),
				request: udomain.PrvUserRegisterProfileRequest{
					JWTUser: jwthelper.User{Id: int64(1000)},
					BasicProfile: udomain.PrvUserBasicProfile{
						FirstName:      "first",
						FamilyName:     "family",
						FirstNameKana:  "first kana",
						FamilyNameKana: "family kana",
						Birthday:       tooYoungBirthday,
						Nationality:    domain.JP,
						Gender:         domain.MALE,
					},
					SignUpStatus: domain.SIGNUP_PRIVATE_PROFILE_BASIC,
				},
			},
			privateUserRepository: func() cdomain.PrivateUserRepository {
				repo := &mocks.PrivateUserRepository{}
				repo.On("GetById", context.Background(), int64(1000)).Return(domain.PrivateUser{
					User: domain.User{
						Id:           int64(1000),
						SignupStatus: domain.SIGNUP_NONE,
					},
				}, nil)
				return repo
			},
			userRepository: func() cdomain.UserRepository {
				return &mocks.UserRepository{}
			},
			privateProfileRepository: func() cdomain.PrivateProfileRepository {
				return &mocks.PrivateProfileRepository{}
			},
			wantErr: true,
		},
		{
			name: "Success: 年齢チェックOK（満79歳）",
			args: args{
				ctx: context.Background(),
				request: udomain.PrvUserRegisterProfileRequest{
					JWTUser: jwthelper.User{Id: int64(1000)},
					BasicProfile: udomain.PrvUserBasicProfile{
						FirstName:      "first name",
						FamilyName:     "family name",
						FirstNameKana:  "first name kana",
						FamilyNameKana: "family name kana",
						Birthday:       exactly79Birthday,
						Nationality:    domain.JP,
						Gender:         domain.MALE,
					},
					SignUpStatus: domain.SIGNUP_PRIVATE_PROFILE_BASIC,
				},
			},
			privateUserRepository: func() cdomain.PrivateUserRepository {
				repo := &mocks.PrivateUserRepository{}
				repo.On("GetById", context.Background(), int64(1000)).Return(domain.PrivateUser{
					User: domain.User{
						Id:           int64(1000),
						SignupStatus: domain.SIGNUP_NONE,
					},
				}, nil)
				return repo
			},
			userRepository: func() cdomain.UserRepository {
				repo := &mocks.UserRepository{}
				repo.On("Update", context.Background(), domain.User{
					Id:           int64(1000),
					SignupStatus: domain.SIGNUP_PRIVATE_PROFILE_BASIC,
					Name:         "family name first name",
				}).Return(nil)
				return repo
			},
			privateProfileRepository: func() cdomain.PrivateProfileRepository {
				repo := &mocks.PrivateProfileRepository{}
				repo.On("Update", context.Background(), mock.MatchedBy(func(profile domain.PrivateProfile) bool {
					return profile.Birthday.ToFormat(time.DateOnly) == exactly79Birthday
				})).Return(nil)
				return repo
			},
			want: udomain.PrvUserRegisterProfileResponse{},
		},
		{
			name: "Failed: 年齢チェックNG（上限NG）",
			args: args{
				ctx: context.Background(),
				request: udomain.PrvUserRegisterProfileRequest{
					JWTUser: jwthelper.User{Id: int64(1000)},
					BasicProfile: udomain.PrvUserBasicProfile{
						FirstName:      "first",
						FamilyName:     "family",
						FirstNameKana:  "first kana",
						FamilyNameKana: "family kana",
						Birthday:       tooOldBirthday,
						Nationality:    domain.JP,
						Gender:         domain.MALE,
					},
					SignUpStatus: domain.SIGNUP_PRIVATE_PROFILE_BASIC,
				},
			},
			privateUserRepository: func() cdomain.PrivateUserRepository {
				repo := &mocks.PrivateUserRepository{}
				repo.On("GetById", context.Background(), int64(1000)).Return(domain.PrivateUser{
					User: domain.User{
						Id:           int64(1000),
						SignupStatus: domain.SIGNUP_NONE,
					},
				}, nil)
				return repo
			},
			userRepository: func() cdomain.UserRepository {
				return &mocks.UserRepository{}
			},
			privateProfileRepository: func() cdomain.PrivateProfileRepository {
				return &mocks.PrivateProfileRepository{}
			},
			wantErr: true,
		},
		{
			name: "Failed: 年齢チェックNG（満80歳）",
			args: args{
				ctx: context.Background(),
				request: udomain.PrvUserRegisterProfileRequest{
					JWTUser: jwthelper.User{Id: int64(1000)},
					BasicProfile: udomain.PrvUserBasicProfile{
						FirstName:      "first",
						FamilyName:     "family",
						FirstNameKana:  "first kana",
						FamilyNameKana: "family kana",
						Birthday:       exactly80Birthday,
						Nationality:    domain.JP,
						Gender:         domain.MALE,
					},
					SignUpStatus: domain.SIGNUP_PRIVATE_PROFILE_BASIC,
				},
			},
			privateUserRepository: func() cdomain.PrivateUserRepository {
				repo := &mocks.PrivateUserRepository{}
				repo.On("GetById", context.Background(), int64(1000)).Return(domain.PrivateUser{
					User: domain.User{
						Id:           int64(1000),
						SignupStatus: domain.SIGNUP_NONE,
					},
				}, nil)
				return repo
			},
			userRepository: func() cdomain.UserRepository {
				return &mocks.UserRepository{}
			},
			privateProfileRepository: func() cdomain.PrivateProfileRepository {
				return &mocks.PrivateProfileRepository{}
			},
			wantErr: true,
		},
		{
			name: "Failed: 投資適合性チェックNG",
			args: args{
				ctx: context.Background(),
				request: udomain.PrvUserRegisterProfileRequest{
					JWTUser: jwthelper.User{Id: int64(1000)},
					InvestmentSuitability: udomain.PrvUserInvestmentSuitability{
						Asset:                   domain.ASSET1,
						AnnualIncome:            domain.ANNUALINCOME1,
						IsProInvestAsset:        lo.ToPtr(false),
						IsProInvestFinanceAsset: lo.ToPtr(false),
						InvestmentPurpose:       domain.INVESTMENT_PURPOSE_ETC,
						IsSurplusFund:           true,
						HasExperience:           false,
					},
					SignUpStatus: domain.SIGNUP_INVESTMENT_SUITABILITY,
				},
			},
			privateUserRepository: func() cdomain.PrivateUserRepository {
				repo := &mocks.PrivateUserRepository{}
				repo.On("GetById", context.Background(), int64(1000)).Return(domain.PrivateUser{
					User: domain.User{
						Id:           int64(1000),
						SignupStatus: domain.SIGNUP_ACCOUNT,
					},
				}, nil)
				return repo
			},
			userRepository: func() cdomain.UserRepository {
				return &mocks.UserRepository{}
			},
			privateProfileRepository: func() cdomain.PrivateProfileRepository {
				return &mocks.PrivateProfileRepository{}
			},
			wantErr: true,
		},
		{
			name: "Failed: 無効なステータス",
			args: args{
				ctx: context.Background(),
				request: udomain.PrvUserRegisterProfileRequest{
					JWTUser:      jwthelper.User{Id: int64(1000)},
					SignUpStatus: domain.SIGNUP_COMPLETE,
				},
			},
			privateUserRepository: func() cdomain.PrivateUserRepository {
				repo := &mocks.PrivateUserRepository{}
				repo.On("GetById", context.Background(), int64(1000)).Return(domain.PrivateUser{
					User: domain.User{
						Id:           int64(1000),
						SignupStatus: domain.SIGNUP_COMPLETE,
					},
				}, nil)
				return repo
			},
			userRepository: func() cdomain.UserRepository {
				return &mocks.UserRepository{}
			},
			privateProfileRepository: func() cdomain.PrivateProfileRepository {
				return &mocks.PrivateProfileRepository{}
			},
			wantErr: true,
		},
		{
			name: "Failed: ユーザー更新失敗",
			args: args{
				ctx: context.Background(),
				request: udomain.PrvUserRegisterProfileRequest{
					JWTUser:      jwthelper.User{Id: int64(1000)},
					SignUpStatus: domain.SIGNUP_AGREE,
				},
			},
			idGenerator: func() idgenerator.IdGenerator {
				mockIdGenerator := &mocks.IdGenerator{}
				mockIdGenerator.On("GenKsuid").Return(ksuid.KSUID{})
				return mockIdGenerator
			},
			privateUserRepository: func() cdomain.PrivateUserRepository {
				repo := &mocks.PrivateUserRepository{}
				repo.On("GetById", context.Background(), int64(1000)).Return(domain.PrivateUser{
					User: domain.User{
						Id:           int64(1000),
						SignupStatus: domain.SIGNUP_INVESTMENT_SUITABILITY,
						Province:     "東京都",
						City:         "渋谷区",
						Address:      "道玄坂1-2-3",
					},
					PrivateProfile: domain.PrivateProfile{
						Birthday:   jsonutils.JSONTime(time.Date(1990, 1, 2, 0, 0, 0, 0, time.UTC)),
						FamilyName: "山田",
						FirstName:  "太郎",
					},
				}, nil)
				return repo
			},
			userRepository: func() cdomain.UserRepository {
				repo := &mocks.UserRepository{}
				repo.On("Update", context.Background(), mock.Anything).Return(errors.New("update error"))
				return repo
			},
			privateProfileRepository: func() cdomain.PrivateProfileRepository {
				return &mocks.PrivateProfileRepository{}
			},
			ekycRequestLogRepository: func() cdomain.EkycRequestLogRepository {
				repo := &mocks.EkycRequestLogRepository{}
				repo.On("GetByUserId", context.Background(), int64(1000)).Return(domain.EkycRequestLog{
					Status: domain.EKYC_REQUEST_STATUS_SUCCESS,
				}, nil)
				repo.On("Create", context.Background(), mock.MatchedBy(func(log domain.EkycRequestLog) bool {
					return log.UserId == int64(1000)
				})).Return(nil)
				return repo
			},
			wantErr: true,
		},
		{
			name: "Failed: プロフィール更新失敗",
			args: args{
				ctx: context.Background(),
				request: udomain.PrvUserRegisterProfileRequest{
					JWTUser:      jwthelper.User{Id: int64(1000)},
					SignUpStatus: domain.SIGNUP_AGREE,
				},
			},
			idGenerator: func() idgenerator.IdGenerator {
				mockIdGenerator := &mocks.IdGenerator{}
				mockIdGenerator.On("GenKsuid").Return(ksuid.KSUID{})
				return mockIdGenerator
			},
			privateUserRepository: func() cdomain.PrivateUserRepository {
				repo := &mocks.PrivateUserRepository{}
				repo.On("GetById", context.Background(), int64(1000)).Return(domain.PrivateUser{
					User: domain.User{
						Id:           int64(1000),
						SignupStatus: domain.SIGNUP_INVESTMENT_SUITABILITY,
						Province:     "東京都",
						City:         "渋谷区",
						Address:      "道玄坂1-2-3",
					},
					PrivateProfile: domain.PrivateProfile{
						Birthday:   jsonutils.JSONTime(time.Date(1990, 1, 2, 0, 0, 0, 0, time.UTC)),
						FamilyName: "山田",
						FirstName:  "太郎",
					},
				}, nil)
				return repo
			},
			userRepository: func() cdomain.UserRepository {
				repo := &mocks.UserRepository{}
				repo.On("Update", context.Background(), mock.Anything).Return(nil)
				return repo
			},
			privateProfileRepository: func() cdomain.PrivateProfileRepository {
				repo := &mocks.PrivateProfileRepository{}
				repo.On("Update", context.Background(), mock.Anything).Return(errors.New("profile update error"))
				return repo
			},
			ekycRequestLogRepository: func() cdomain.EkycRequestLogRepository {
				repo := &mocks.EkycRequestLogRepository{}
				repo.On("GetByUserId", context.Background(), int64(1000)).Return(domain.EkycRequestLog{
					Status: domain.EKYC_REQUEST_STATUS_SUCCESS,
				}, nil)
				repo.On("Create", context.Background(), mock.MatchedBy(func(log domain.EkycRequestLog) bool {
					return log.UserId == int64(1000)
				})).Return(nil)
				return repo
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := privateUserUsecase{
				userRepository:           tt.userRepository(),
				privateUserRepository:    tt.privateUserRepository(),
				privateProfileRepository: tt.privateProfileRepository(),
			}
			if tt.idGenerator != nil {
				usecase.idGenerator = tt.idGenerator()
			}
			if tt.ekycRequestLogRepository != nil {
				usecase.ekycRequestLogRepository = tt.ekycRequestLogRepository()
			}
			res, err := usecase.RegisterProfile(tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.wantErr, err != nil, "RegisterProfile() error = %v, wantErr %v", err, tt.wantErr)
			if !tt.wantErr {
				assert.Equal(t, tt.want, res)
			} else {
				assert.Equal(t, udomain.PrvUserRegisterProfileResponse{}, res)
			}
			if tt.wantErrIs != nil {
				assert.ErrorIs(t, err, tt.wantErrIs)
			}
		})
	}
}

func Test_privateUserUsecase_RegisterMyNumber(t *testing.T) {
	// TODO: マイナンバー登録復活時にもとに戻す
	t.Skip()

	// type args struct {
	// 	ctx     context.Context
	// 	request udomain.PrvUserRegisterMyNumberRequest
	// }

	// tests := []struct {
	// 	name                  string
	// 	args                  args
	// 	logger                factory.LoggerFactory
	// 	cfg                   *config.Config
	// 	satori                func() satori.Satori
	// 	userRepository        func() cdomain.UserRepository
	// 	privateUserRepository func() cdomain.PrivateUserRepository
	// 	myNumberService       func() cdomain.MyNumberService
	// 	want                  udomain.PrvUserRegisterMyNumberResponse
	// 	wantErr               bool
	// }{
	// 	{
	// 		name: "(成功) マイナンバー登録",
	// 		args: args{
	// 			ctx: context.Background(),
	// 			request: udomain.PrvUserRegisterMyNumberRequest{
	// 				JWTUser: jwthelper.User{
	// 					Id: mockMyNumber.UserId,
	// 				},
	// 				MyNumber:     mockMyNumber.MyNumber,
	// 				FrontFileUrl: mockMyNumberFront.FileUrl,
	// 				BackFileUrl:  mockMyNumberBack.FileUrl,
	// 			},
	// 		},
	// 		satori: func() satori.Satori {
	// 			satoriClient := &mocks.Satori{}
	// 			satoriClient.On("Upsert", mock.Anything, satori.UpsertRequest{
	// 				Hashcode:          "8eced811cdb57f56",
	// 				Email:             "test@example.com",
	// 				InvestmentPurpose: "説明なし",
	// 				HasExperience:     "いいえ",
	// 				IsSurplusFund:     "いいえ",
	// 				PrivateUser: satori.PrivateUser{
	// 					Gender:                  "女性",
	// 					Nationality:             "日本",
	// 					Job:                     "説明なし",
	// 					Asset:                   "説明なし",
	// 					AnnualIncome:            "説明なし",
	// 					IsProInvestAsset:        "いいえ",
	// 					IsProInvestFinanceAsset: "いいえ",
	// 				},
	// 			}).Return(nil)
	// 			return satoriClient
	// 		},
	// 		userRepository: func() cdomain.UserRepository {
	// 			repo := &mocks.UserRepository{}
	// 			updated := mockUserForSatori.User
	// 			updated.Id = mockMyNumber.UserId
	// 			updated.SignupStatus = domain.SIGNUP_MYNUMBER
	// 			updated.AuditStatus = domain.CHECK_MY_NUMBER_AUDIT
	// 			repo.On("Update", context.Background(), updated).Return(nil)
	// 			return repo
	// 		},
	// 		privateUserRepository: func() cdomain.PrivateUserRepository {
	// 			repo := &mocks.PrivateUserRepository{}
	// 			userForSatori := mockUserForSatori
	// 			userForSatori.User.Id = mockMyNumber.UserId
	// 			userForSatori.PrivateProfile.UserId = mockMyNumber.UserId
	// 			repo.On("GetById", mock.Anything, mockMyNumber.UserId).Return(userForSatori, nil)
	// 			return repo
	// 		},
	// 		myNumberService: func() cdomain.MyNumberService {
	// 			service := &mocks.MyNumberService{}
	// 			service.On("RegisterWithFiles", context.Background(),
	// 				cdomain.MyNumberRegisterInput{
	// 					UserId:        mockMyNumber.UserId,
	// 					MyNumberValue: mockMyNumber.MyNumber,
	// 					FrontFileUrl:  mockMyNumberFront.FileUrl,
	// 					BackFileUrl:   mockMyNumberBack.FileUrl,
	// 					MyNumber:      domain.MyNumber{Status: domain.UNREGISTERED},
	// 				}).Return(domain.MyNumber{
	// 				Id:        mockMyNumber.Id,
	// 				UserId:    mockMyNumber.UserId,
	// 				MyNumber:  mockMyNumber.MyNumber,
	// 				Status:    domain.WAIT_APPROVE,
	// 				FrontFile: &mockMyNumberFront,
	// 				BackFile:  &mockMyNumberBack,
	// 			}, nil)
	// 			service.On("GetByUserId", context.Background(), mockMyNumber.UserId).Return(domain.MyNumber{}, nil)
	// 			return service
	// 		},
	// 		want: udomain.PrvUserRegisterMyNumberResponse{
	// 			PrvUserMyNumber: udomain.PrvUserMyNumber{
	// 				Id:       mockMyNumber.Id,
	// 				MyNumber: mockMyNumber.MyNumber,
	// 				Status:   mockMyNumber.Status,
	// 				FrontFile: udomain.PrvUserFile{
	// 					OriginalName:  "",
	// 					FileType:      mockMyNumberFront.FileType,
	// 					FileSize:      0,
	// 					FileExtension: "",
	// 					FilePath:      "",
	// 					FileUrl:       mockMyNumberFront.FileUrl,
	// 					OwnerId:       mockMyNumberFront.OwnerId,
	// 					UpdateId:      mockMyNumberFront.UpdateId,
	// 				},
	// 				BackFile: udomain.PrvUserFile{
	// 					OriginalName:  "",
	// 					FileType:      mockMyNumberBack.FileType,
	// 					FileSize:      0,
	// 					FileExtension: "",
	// 					FilePath:      "",
	// 					FileUrl:       mockMyNumberBack.FileUrl,
	// 					OwnerId:       mockMyNumberBack.OwnerId,
	// 					UpdateId:      mockMyNumberBack.UpdateId,
	// 				},
	// 			},
	// 		},
	// 		wantErr: false,
	// 	},
	// 	{
	// 		name: "(失敗) マイナンバー登録 - マイナンバーがWAIT_APPROVE状態の場合",
	// 		args: args{
	// 			ctx: context.Background(),
	// 			request: udomain.PrvUserRegisterMyNumberRequest{
	// 				JWTUser: jwthelper.User{
	// 					Id: mockMyNumber2.UserId,
	// 				},
	// 				MyNumber:     mockMyNumber2.MyNumber,
	// 				FrontFileUrl: mockMyNumber2Front.FileUrl,
	// 				BackFileUrl:  mockMyNumber2Back.FileUrl,
	// 			},
	// 		},
	// 		satori: func() satori.Satori {
	// 			satoriClient := &mocks.Satori{}
	// 			satoriClient.On("Upsert", mock.Anything, mock.Anything).Return(nil)
	// 			return satoriClient
	// 		},
	// 		privateUserRepository: func() cdomain.PrivateUserRepository {
	// 			repo := &mocks.PrivateUserRepository{}
	// 			userForSatori := mockUserForSatori
	// 			userForSatori.User.Id = mockMyNumber2.UserId
	// 			userForSatori.PrivateProfile.UserId = mockMyNumber2.UserId
	// 			repo.On("GetById", mock.Anything, mockMyNumber2.UserId).Return(userForSatori, nil)
	// 			return repo
	// 		},
	// 		userRepository: func() cdomain.UserRepository {
	// 			repo := &mocks.UserRepository{}
	// 			repo.On("GetById", context.Background(), mockMyNumber2.UserId).Return(domain.User{
	// 				Id:           mockMyNumber2.UserId,
	// 				SignupStatus: domain.SIGNUP_MYNUMBER,
	// 			}, nil)
	// 			return repo
	// 		},
	// 		myNumberService: func() cdomain.MyNumberService {
	// 			service := &mocks.MyNumberService{}
	// 			service.On("GetByUserId", context.Background(), mockMyNumber2.UserId).Return(mockMyNumber2, nil)
	// 			return service
	// 		},
	// 		want:    udomain.PrvUserRegisterMyNumberResponse{},
	// 		wantErr: true,
	// 	},
	// }

	// for _, tt := range tests {
	// 	t.Run(tt.name, func(t *testing.T) {
	// 		usecase := privateUserUsecase{
	// 			logger: factory.NewLoggerFactory(),
	// 			cfg: &config.Config{
	// 				CommonConfig: config.CommonConfig{
	// 					AWS: struct {
	// 						Region            string
	// 						SES_Region        string
	// 						SNS_Region        string
	// 						CDN               string
	// 						S3_PRIVATE_BUCKET string
	// 						S3_PUBLIC_BUCKET  string
	// 						SES_ARN           string
	// 						SNS_ARN           string
	// 						PUBLIC_KEY_ID     string
	// 						PRIVATE_KEY_PATH  string
	//                      BASE_ENDPOINT     string
	// 					}{
	// 						S3_PRIVATE_BUCKET: "es-jp-devv2-private-file-storage",
	// 					},
	// 				},
	// 			},
	// 			userRepository:        tt.userRepository(),
	// 			privateUserRepository: tt.privateUserRepository(),
	// 			myNumberService:       tt.myNumberService(),
	// 			satori:                tt.satori(),
	// 		}
	// 		got, err := usecase.RegisterMyNumber(tt.args.ctx, tt.args.request)

	// 		assert.Equalf(t, tt.wantErr, err != nil, "RegisterMyNumber() error = %v, wantErr %v", err, tt.wantErr)
	// 		ignoreFields := cmpopts.IgnoreFields(udomain.PrvUserFile{}, "Id")
	// 		ignoreTypes := cmpopts.IgnoreTypes(jsonutils.JSONTime{})
	// 		assert.Equalf(t, cmp.Equal(tt.want, got, ignoreTypes, ignoreFields), true, "RegisterMyNumber() = %+v, want %+v", got, tt.want)
	// 	})
	// }
}

func TestPrivateUserUsecase_VerifyWithEkyc(t *testing.T) {
	randomID, _ := ksuid.NewRandom()

	type args struct {
		ctx     context.Context
		request udomain.PrvUserVerifyWithEkycRequest
	}

	tests := []struct {
		name                     string
		args                     args
		cfg                      *config.Config
		idGenerator              func() idgenerator.IdGenerator
		privateUserRepository    func() cdomain.PrivateUserRepository
		nexwayService            func() cdomain.NexwayService
		ekycRequestLogRepository func() cdomain.EkycRequestLogRepository
		want                     udomain.PrvUserVerifyWithEkycResponse
		wantErr                  bool
	}{
		{
			name: "(成功) eKYC認証成功 - Address2 がnilではない場合",
			args: args{
				ctx: context.Background(),
				request: udomain.PrvUserVerifyWithEkycRequest{
					JWTUser: jwthelper.User{Id: mockPrivateUserVerifyWithEkyc.User.Id},
					UserId:  mockPrivateUserVerifyWithEkyc.User.Id,
				},
			},
			idGenerator: func() idgenerator.IdGenerator {
				idGenerator := &mocks.IdGenerator{}
				idGenerator.On("GenKsuid").Return(randomID)
				return idGenerator
			},
			privateUserRepository: func() cdomain.PrivateUserRepository {
				repo := &mocks.PrivateUserRepository{}
				repo.On("GetById", context.Background(), mockPrivateUserVerifyWithEkyc.User.Id).Return(mockPrivateUserVerifyWithEkyc, nil)
				return repo
			},
			nexwayService: func() cdomain.NexwayService {
				service := &mocks.NexwayService{}

				address := mockPrivateUserVerifyWithEkyc.Province +
					mockPrivateUserVerifyWithEkyc.City +
					mockPrivateUserVerifyWithEkyc.Address +
					"-" + strings.TrimSpace(lo.FromPtr(mockPrivateUserVerifyWithEkyc.Address2))

				service.On("RegisterProfile", context.Background(), cdomain.NexwayProfileInput{
					ReferenceId: randomID.String(),
					Name:        mockPrivateUserVerifyWithEkyc.Name,
					Birthday:    time.Time(mockPrivateUserVerifyWithEkyc.Birthday).Format(time.DateOnly),
					Address:     address,
					Job:         string(mockPrivateUserVerifyWithEkyc.Job),
					Purpose:     string(mockPrivateUserVerifyWithEkyc.InvestmentPurpose),
				}).Return(cdomain.NexwayProfileOutput{
					ApplicantId: int64(123),
					ReferenceId: randomID.String(),
				}, nil)

				return service
			},
			ekycRequestLogRepository: func() cdomain.EkycRequestLogRepository {
				repo := &mocks.EkycRequestLogRepository{}
				repo.On("GetByUserId", context.Background(), mockPrivateUserVerifyWithEkyc.User.Id).Return(domain.EkycRequestLog{}, cdomain.ErrItemNotFound)
				repo.On("Create", context.Background(), mock.MatchedBy(func(req domain.EkycRequestLog) bool {
					return req.Email == mockPrivateUserVerifyWithEkyc.Email &&
						req.UserId == mockPrivateUserVerifyWithEkyc.UserId &&
						req.ApplicantId == int64(0) &&
						req.Status == domain.EKYC_REQUEST_STATUS_NONE
				})).Return(nil)
				return repo
			},
			want: udomain.PrvUserVerifyWithEkycResponse{
				ReferenceId: randomID.String(),
			},
			wantErr: false,
		},
		{
			name: "(成功) eKYC認証成功 - Address2がnilの場合",
			args: args{
				ctx: context.Background(),
				request: udomain.PrvUserVerifyWithEkycRequest{
					JWTUser: jwthelper.User{Id: mockPrivateUserVerifyWithEkyc2.User.Id},
					UserId:  mockPrivateUserVerifyWithEkyc2.User.Id,
				},
			},
			idGenerator: func() idgenerator.IdGenerator {
				idGenerator := &mocks.IdGenerator{}
				idGenerator.On("GenKsuid").Return(randomID)
				return idGenerator
			},
			privateUserRepository: func() cdomain.PrivateUserRepository {
				repo := &mocks.PrivateUserRepository{}
				repo.On("GetById", context.Background(), mockPrivateUserVerifyWithEkyc2.User.Id).Return(mockPrivateUserVerifyWithEkyc2, nil)
				return repo
			},
			nexwayService: func() cdomain.NexwayService {
				service := &mocks.NexwayService{}

				address := mockPrivateUserVerifyWithEkyc2.Province +
					mockPrivateUserVerifyWithEkyc2.City +
					mockPrivateUserVerifyWithEkyc2.Address

				service.On("RegisterProfile", context.Background(), cdomain.NexwayProfileInput{
					ReferenceId: randomID.String(),
					Name:        mockPrivateUserVerifyWithEkyc2.Name,
					Birthday:    time.Time(mockPrivateUserVerifyWithEkyc2.Birthday).Format(time.DateOnly),
					Address:     address,
					Job:         string(mockPrivateUserVerifyWithEkyc2.Job),
					Purpose:     string(mockPrivateUserVerifyWithEkyc2.InvestmentPurpose),
				}).Return(cdomain.NexwayProfileOutput{
					ApplicantId: int64(123),
					ReferenceId: randomID.String(),
				}, nil)

				return service
			},
			ekycRequestLogRepository: func() cdomain.EkycRequestLogRepository {
				repo := &mocks.EkycRequestLogRepository{}
				repo.On("GetByUserId", context.Background(), mockPrivateUserVerifyWithEkyc2.User.Id).Return(domain.EkycRequestLog{}, cdomain.ErrItemNotFound)
				repo.On("Create", context.Background(), mock.MatchedBy(func(req domain.EkycRequestLog) bool {
					return req.Email == mockPrivateUserVerifyWithEkyc2.Email &&
						req.UserId == mockPrivateUserVerifyWithEkyc2.UserId &&
						req.ApplicantId == int64(0) &&
						req.Status == domain.EKYC_REQUEST_STATUS_NONE
				})).Return(nil)
				return repo
			},
			want: udomain.PrvUserVerifyWithEkycResponse{
				ReferenceId: randomID.String(),
			},
			wantErr: false,
		},
		{
			name: "(成功) eKYC認証成功 - nexway要求記録再利用",
			args: args{
				ctx: context.Background(),
				request: udomain.PrvUserVerifyWithEkycRequest{
					JWTUser: jwthelper.User{Id: mockPrivateUserVerifyWithEkyc.User.Id},
					UserId:  mockPrivateUserVerifyWithEkyc.User.Id,
				},
			},
			idGenerator: func() idgenerator.IdGenerator {
				idGenerator := &mocks.IdGenerator{}
				idGenerator.On("GenKsuid").Return(randomID)
				return idGenerator
			},
			privateUserRepository: func() cdomain.PrivateUserRepository {
				repo := &mocks.PrivateUserRepository{}
				repo.On("GetById", context.Background(), mockPrivateUserVerifyWithEkyc.User.Id).Return(mockPrivateUserVerifyWithEkyc, nil)
				return repo
			},
			nexwayService: func() cdomain.NexwayService {
				return &mocks.NexwayService{}
			},
			ekycRequestLogRepository: func() cdomain.EkycRequestLogRepository {
				repo := &mocks.EkycRequestLogRepository{}
				repo.On("GetByUserId", context.Background(), mockPrivateUserVerifyWithEkyc.UserId).
					Return(domain.EkycRequestLog{
						EkycId: "test-ekyc",
						Status: domain.EKYC_REQUEST_STATUS_NONE,
					}, nil)
				return repo
			},
			want: udomain.PrvUserVerifyWithEkycResponse{
				ReferenceId: "test-ekyc",
			},
			wantErr: false,
		},
		{
			// MIR-1358: 最新の履歴がNONEだが既にUNMATCHと判定済み(判定が先に
			// 届いた)。使い回すとisEkycResultUnmatchが完了PUTを断ってしまう
			// ため、新しい行(新しいekyc_id)を作る。
			name: "(成功) eKYC認証成功 - 最新の履歴がNONE/UNMATCHなら使い回さず新しい行を作る",
			args: args{
				ctx: context.Background(),
				request: udomain.PrvUserVerifyWithEkycRequest{
					JWTUser: jwthelper.User{Id: mockPrivateUserVerifyWithEkyc.User.Id},
					UserId:  mockPrivateUserVerifyWithEkyc.User.Id,
				},
			},
			idGenerator: func() idgenerator.IdGenerator {
				idGenerator := &mocks.IdGenerator{}
				idGenerator.On("GenKsuid").Return(randomID)
				return idGenerator
			},
			privateUserRepository: func() cdomain.PrivateUserRepository {
				repo := &mocks.PrivateUserRepository{}
				repo.On("GetById", context.Background(), mockPrivateUserVerifyWithEkyc.User.Id).Return(mockPrivateUserVerifyWithEkyc, nil)
				return repo
			},
			nexwayService: func() cdomain.NexwayService {
				return &mocks.NexwayService{}
			},
			ekycRequestLogRepository: func() cdomain.EkycRequestLogRepository {
				repo := &mocks.EkycRequestLogRepository{}
				repo.On("GetByUserId", context.Background(), mockPrivateUserVerifyWithEkyc.UserId).
					Return(domain.EkycRequestLog{
						EkycId: "old-unmatched-ekyc",
						Status: domain.EKYC_REQUEST_STATUS_NONE,
						Result: domain.EKYC_RESULT_STATUS_UNMATCH,
					}, nil)
				repo.On("Create", context.Background(), mock.MatchedBy(func(req domain.EkycRequestLog) bool {
					return req.EkycId == randomID.String() && req.Status == domain.EKYC_REQUEST_STATUS_NONE
				})).Return(nil)
				return repo
			},
			want: udomain.PrvUserVerifyWithEkycResponse{
				ReferenceId: randomID.String(),
			},
			wantErr: false,
		},
		{
			// MIR-1358: 最新の履歴がNONE/MATCH(審査結果が未確定)。断るのは
			// UNMATCHだけだが、使い回すかどうかはresultが空かどうかで決まる
			// ため、MATCHも使い回さず新しい行を作る。
			name: "(成功) eKYC認証成功 - 最新の履歴がNONE/MATCHでも使い回さず新しい行を作る",
			args: args{
				ctx: context.Background(),
				request: udomain.PrvUserVerifyWithEkycRequest{
					JWTUser: jwthelper.User{Id: mockPrivateUserVerifyWithEkyc.User.Id},
					UserId:  mockPrivateUserVerifyWithEkyc.User.Id,
				},
			},
			idGenerator: func() idgenerator.IdGenerator {
				idGenerator := &mocks.IdGenerator{}
				idGenerator.On("GenKsuid").Return(randomID)
				return idGenerator
			},
			privateUserRepository: func() cdomain.PrivateUserRepository {
				repo := &mocks.PrivateUserRepository{}
				repo.On("GetById", context.Background(), mockPrivateUserVerifyWithEkyc.User.Id).Return(mockPrivateUserVerifyWithEkyc, nil)
				return repo
			},
			nexwayService: func() cdomain.NexwayService {
				return &mocks.NexwayService{}
			},
			ekycRequestLogRepository: func() cdomain.EkycRequestLogRepository {
				repo := &mocks.EkycRequestLogRepository{}
				repo.On("GetByUserId", context.Background(), mockPrivateUserVerifyWithEkyc.UserId).
					Return(domain.EkycRequestLog{
						EkycId: "old-matched-ekyc",
						Status: domain.EKYC_REQUEST_STATUS_NONE,
						Result: domain.EKYC_RESULT_STATUS_MATCH,
					}, nil)
				repo.On("Create", context.Background(), mock.MatchedBy(func(req domain.EkycRequestLog) bool {
					return req.EkycId == randomID.String() && req.Status == domain.EKYC_REQUEST_STATUS_NONE
				})).Return(nil)
				return repo
			},
			want: udomain.PrvUserVerifyWithEkycResponse{
				ReferenceId: randomID.String(),
			},
			wantErr: false,
		},
		{
			// MIR-1358: 最新の履歴がSUCCESS(今までと同じ、使い回さない)。
			name: "(成功) eKYC認証成功 - 最新の履歴がSUCCESSなら新しい行を作る(今と同じ)",
			args: args{
				ctx: context.Background(),
				request: udomain.PrvUserVerifyWithEkycRequest{
					JWTUser: jwthelper.User{Id: mockPrivateUserVerifyWithEkyc.User.Id},
					UserId:  mockPrivateUserVerifyWithEkyc.User.Id,
				},
			},
			idGenerator: func() idgenerator.IdGenerator {
				idGenerator := &mocks.IdGenerator{}
				idGenerator.On("GenKsuid").Return(randomID)
				return idGenerator
			},
			privateUserRepository: func() cdomain.PrivateUserRepository {
				repo := &mocks.PrivateUserRepository{}
				repo.On("GetById", context.Background(), mockPrivateUserVerifyWithEkyc.User.Id).Return(mockPrivateUserVerifyWithEkyc, nil)
				return repo
			},
			nexwayService: func() cdomain.NexwayService {
				return &mocks.NexwayService{}
			},
			ekycRequestLogRepository: func() cdomain.EkycRequestLogRepository {
				repo := &mocks.EkycRequestLogRepository{}
				repo.On("GetByUserId", context.Background(), mockPrivateUserVerifyWithEkyc.UserId).
					Return(domain.EkycRequestLog{
						EkycId: "old-success-ekyc",
						Status: domain.EKYC_REQUEST_STATUS_SUCCESS,
					}, nil)
				repo.On("Create", context.Background(), mock.MatchedBy(func(req domain.EkycRequestLog) bool {
					return req.EkycId == randomID.String() && req.Status == domain.EKYC_REQUEST_STATUS_NONE
				})).Return(nil)
				return repo
			},
			want: udomain.PrvUserVerifyWithEkycResponse{
				ReferenceId: randomID.String(),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := privateUserUsecase{
				cfg: &config.Config{
					CommonConfig: config.CommonConfig{
						Env: "prod",
						ServiceConfig: struct {
							Mail serviceutils.MailConfig
							Url  serviceutils.UrlConfig
						}{
							Url: serviceutils.UrlConfig{
								UserServiceDomain: "http://energyshares.jp",
							},
						},
					},
				},
				idGenerator:              tt.idGenerator(),
				privateUserRepository:    tt.privateUserRepository(),
				nexwayService:            tt.nexwayService(),
				ekycRequestLogRepository: tt.ekycRequestLogRepository(),
			}
			got, err := usecase.VerifyWithEkyc(tt.args.ctx, tt.args.request)

			assert.Equalf(t, tt.wantErr, err != nil, "VerifyWithEkyc() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want, got, "VerifyWithEkyc() = %v, want %v", got, tt.want)
		})
	}
}

// MIR-1358: ①敵対的検証の指摘。verifyWithEkyc(小文字、同意フロー
// RegisterProfileのSIGNUP_AGREEから呼ばれる)の使い回し判定を旧条件
// (Status==NONEだけ)に戻してもTestPrivateUserUsecase_VerifyWithEkyc
// (大文字)は全件成功するままだった。大文字・小文字それぞれの経路を
// 直接確かめるテストを追加する。
func Test_privateUserUsecase_verifyWithEkyc(t *testing.T) {
	randomID, _ := ksuid.NewRandom()

	tests := []struct {
		name                     string
		ekycRequestLogRepository func() cdomain.EkycRequestLogRepository
		wantEkycId               string
		wantCreateCalled         bool
	}{
		{
			name: "履歴がNONE/空なら使い回す(Createを呼ばない、同じekyc_id)",
			ekycRequestLogRepository: func() cdomain.EkycRequestLogRepository {
				repo := &mocks.EkycRequestLogRepository{}
				repo.On("GetByUserId", context.Background(), mockPrivateUserVerifyWithEkyc.User.Id).
					Return(domain.EkycRequestLog{
						EkycId: "reused-ekyc-id",
						Status: domain.EKYC_REQUEST_STATUS_NONE,
					}, nil)
				return repo
			},
			wantEkycId:       "reused-ekyc-id",
			wantCreateCalled: false,
		},
		{
			name: "履歴がNONE/UNMATCHなら使い回さず新しい行を作る",
			ekycRequestLogRepository: func() cdomain.EkycRequestLogRepository {
				repo := &mocks.EkycRequestLogRepository{}
				repo.On("GetByUserId", context.Background(), mockPrivateUserVerifyWithEkyc.User.Id).
					Return(domain.EkycRequestLog{
						EkycId: "old-unmatched-ekyc-id",
						Status: domain.EKYC_REQUEST_STATUS_NONE,
						Result: domain.EKYC_RESULT_STATUS_UNMATCH,
					}, nil)
				repo.On("Create", context.Background(), mock.MatchedBy(func(req domain.EkycRequestLog) bool {
					return req.EkycId == randomID.String() && req.Status == domain.EKYC_REQUEST_STATUS_NONE
				})).Return(nil)
				return repo
			},
			wantEkycId:       randomID.String(),
			wantCreateCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			idGenerator := &mocks.IdGenerator{}
			idGenerator.On("GenKsuid").Return(randomID)
			ekycRepo := tt.ekycRequestLogRepository()

			usecase := privateUserUsecase{
				idGenerator:              idGenerator,
				ekycRequestLogRepository: ekycRepo,
			}
			gotEkycId, err := usecase.verifyWithEkyc(context.Background(), mockPrivateUserVerifyWithEkyc)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantEkycId, gotEkycId)
			if !tt.wantCreateCalled {
				ekycRepo.(*mocks.EkycRequestLogRepository).AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
			}
		})
	}
}

func Test_isEkycResultUnmatch(t *testing.T) {
	tests := []struct {
		name string
		log  domain.EkycRequestLog
		want bool
	}{
		{"resultがUNMATCHならtrue", domain.EkycRequestLog{Result: domain.EKYC_RESULT_STATUS_UNMATCH}, true},
		{"resultがMATCHならfalse", domain.EkycRequestLog{Result: domain.EKYC_RESULT_STATUS_MATCH}, false},
		{"resultが空(判定前)ならfalse", domain.EkycRequestLog{Result: ""}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isEkycResultUnmatch(tt.log))
		})
	}
}

func Test_canReuseEkycRequestLog(t *testing.T) {
	tests := []struct {
		name string
		log  domain.EkycRequestLog
		want bool
	}{
		{"status=NONE かつ result=空 なら使い回せる", domain.EkycRequestLog{Status: domain.EKYC_REQUEST_STATUS_NONE, Result: ""}, true},
		{"status=NONE だが result=UNMATCH なら使い回せない", domain.EkycRequestLog{Status: domain.EKYC_REQUEST_STATUS_NONE, Result: domain.EKYC_RESULT_STATUS_UNMATCH}, false},
		{"status=NONE だが result=MATCH なら使い回せない", domain.EkycRequestLog{Status: domain.EKYC_REQUEST_STATUS_NONE, Result: domain.EKYC_RESULT_STATUS_MATCH}, false},
		{"status=SUCCESS なら使い回せない", domain.EkycRequestLog{Status: domain.EKYC_REQUEST_STATUS_SUCCESS, Result: ""}, false},
		{"status=FAIL なら使い回せない", domain.EkycRequestLog{Status: domain.EKYC_REQUEST_STATUS_FAIL, Result: ""}, false},
		{"履歴が見つからない(ゼロ値)なら使い回せない", domain.EkycRequestLog{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, canReuseEkycRequestLog(tt.log))
		})
	}
}

func TestPrivateUserUsecase_UpdateEkycStatus(t *testing.T) {
	type args struct {
		ctx     context.Context
		request udomain.PrvUserUpdateEkycStatusRequest
	}

	tests := []struct {
		name                     string
		args                     args
		ekycRequestLogRepository func() cdomain.EkycRequestLogRepository
		userRepository           func() cdomain.UserRepository
		wantErr                  bool
		// MIR-1358 ②レビュー指摘: 断る理由がErrItemConflictであることまで
		// 確かめる。wantErrだけだとエラーの型が変わっても(例えばDB障害の
		// エラーに化けても)テストが成功し続けてしまう。
		wantErrIs error
	}{
		{
			name: "(成功) eKYC状態更新: 本人認証成功",
			args: args{
				ctx: context.Background(),
				request: udomain.PrvUserUpdateEkycStatusRequest{
					JWTUser: jwthelper.User{Id: mockEkycRequestLogUpdateEkycStatus.UserId},
					Status:  "success",
					UserId:  mockEkycRequestLogUpdateEkycStatus.UserId,
				},
			},
			ekycRequestLogRepository: func() cdomain.EkycRequestLogRepository {
				repo := &mocks.EkycRequestLogRepository{}
				repo.On("GetByUserId", context.Background(), mockEkycRequestLogUpdateEkycStatus.UserId).Return(mockEkycRequestLogUpdateEkycStatus, nil)
				repo.On("Update", context.Background(), domain.EkycRequestLog{
					Id:          mockEkycRequestLogUpdateEkycStatus.Id,
					Email:       mockEkycRequestLogUpdateEkycStatus.Email,
					UserId:      mockEkycRequestLogUpdateEkycStatus.UserId,
					EkycId:      mockEkycRequestLogUpdateEkycStatus.EkycId,
					ApplicantId: mockEkycRequestLogUpdateEkycStatus.ApplicantId,
					Status:      domain.EKYC_REQUEST_STATUS_SUCCESS,
				}).Return(nil)
				return repo
			},
			userRepository: func() cdomain.UserRepository {
				repo := &mocks.UserRepository{}
				repo.On("GetById", context.Background(), mockUserUpdateEkycStatus.Id).Return(mockUserUpdateEkycStatus, nil)
				repo.On("GetByEmail", context.Background(), mockUserUpdateEkycStatus.Email).Return(mockUserUpdateEkycStatus, nil)
				repo.On("Update", context.Background(), domain.User{
					Id:           mockUserUpdateEkycStatus.Id,
					Email:        mockUserUpdateEkycStatus.Email,
					SignupStatus: domain.SIGNUP_COMPLETE,
					AuditStatus:  domain.WAIT_AUDIT,
				}).Return(nil)
				return repo
			},
			wantErr: false,
		},
		{
			name: "(成功) eKYC状態更新: 本人認証失敗",
			args: args{
				ctx: context.Background(),
				request: udomain.PrvUserUpdateEkycStatusRequest{
					JWTUser: jwthelper.User{Id: mockEkycRequestLogUpdateEkycStatus.UserId},
					Status:  "fail",
					UserId:  mockEkycRequestLogUpdateEkycStatus.UserId,
				},
			},
			ekycRequestLogRepository: func() cdomain.EkycRequestLogRepository {
				repo := &mocks.EkycRequestLogRepository{}
				repo.On("GetByUserId", context.Background(), mockEkycRequestLogUpdateEkycStatus.UserId).Return(mockEkycRequestLogUpdateEkycStatus, nil)
				repo.On("Update", context.Background(), domain.EkycRequestLog{
					Id:          mockEkycRequestLogUpdateEkycStatus.Id,
					Email:       mockEkycRequestLogUpdateEkycStatus.Email,
					UserId:      mockEkycRequestLogUpdateEkycStatus.UserId,
					EkycId:      mockEkycRequestLogUpdateEkycStatus.EkycId,
					ApplicantId: mockEkycRequestLogUpdateEkycStatus.ApplicantId,
					Status:      domain.EKYC_REQUEST_STATUS_FAIL,
				}).Return(nil)
				return repo
			},
			userRepository: func() cdomain.UserRepository {
				repo := &mocks.UserRepository{}
				return repo
			},
			wantErr: false,
		},
		{
			// MIR-1358: 会員10243の型。完了PUTが届いたあと判定がUNMATCHになり
			// 保留へ戻った会員に、同じ完了PUTがもう一度届くケース。
			name: "(失敗) eKYC状態更新: 最新の履歴がSUCCESS/UNMATCHなら断る(保留の会員)",
			args: args{
				ctx: context.Background(),
				request: udomain.PrvUserUpdateEkycStatusRequest{
					JWTUser: jwthelper.User{Id: mockEkycRequestLogUpdateEkycStatus.UserId},
					Status:  "success",
					UserId:  mockEkycRequestLogUpdateEkycStatus.UserId,
				},
			},
			ekycRequestLogRepository: func() cdomain.EkycRequestLogRepository {
				repo := &mocks.EkycRequestLogRepository{}
				unmatchedLog := mockEkycRequestLogUpdateEkycStatus
				unmatchedLog.Status = domain.EKYC_REQUEST_STATUS_SUCCESS
				unmatchedLog.Result = domain.EKYC_RESULT_STATUS_UNMATCH
				repo.On("GetByUserId", context.Background(), mockEkycRequestLogUpdateEkycStatus.UserId).Return(unmatchedLog, nil)
				return repo
			},
			userRepository: func() cdomain.UserRepository {
				repo := &mocks.UserRepository{}
				return repo
			},
			wantErr:   true,
			wantErrIs: cdomain.ErrItemConflict,
		},
		{
			// MIR-1358: 画面からの完了PUTより先にwebhookの判定(UNMATCH)が
			// 届いていたケース。履歴のstatusはまだNONEのまま。
			name: "(失敗) eKYC状態更新: 最新の履歴がNONE/UNMATCHなら断る(判定が先に着いた)",
			args: args{
				ctx: context.Background(),
				request: udomain.PrvUserUpdateEkycStatusRequest{
					JWTUser: jwthelper.User{Id: mockEkycRequestLogUpdateEkycStatus.UserId},
					Status:  "success",
					UserId:  mockEkycRequestLogUpdateEkycStatus.UserId,
				},
			},
			ekycRequestLogRepository: func() cdomain.EkycRequestLogRepository {
				repo := &mocks.EkycRequestLogRepository{}
				unmatchedLog := mockEkycRequestLogUpdateEkycStatus
				unmatchedLog.Result = domain.EKYC_RESULT_STATUS_UNMATCH
				repo.On("GetByUserId", context.Background(), mockEkycRequestLogUpdateEkycStatus.UserId).Return(unmatchedLog, nil)
				return repo
			},
			userRepository: func() cdomain.UserRepository {
				repo := &mocks.UserRepository{}
				return repo
			},
			wantErr:   true,
			wantErrIs: cdomain.ErrItemConflict,
		},
		{
			// MIR-1358: 保留からの正しいやり直し(失敗画面の再試行=経路A)。
			// 新しい履歴はresultが空のため、完了PUTは断られない。
			name: "(成功) eKYC状態更新: 最新の履歴がNONE/空なら保留の会員でも通す(経路A)",
			args: args{
				ctx: context.Background(),
				request: udomain.PrvUserUpdateEkycStatusRequest{
					JWTUser: jwthelper.User{Id: mockEkycRequestLogUpdateEkycStatus.UserId},
					Status:  "success",
					UserId:  mockEkycRequestLogUpdateEkycStatus.UserId,
				},
			},
			ekycRequestLogRepository: func() cdomain.EkycRequestLogRepository {
				repo := &mocks.EkycRequestLogRepository{}
				repo.On("GetByUserId", context.Background(), mockEkycRequestLogUpdateEkycStatus.UserId).Return(mockEkycRequestLogUpdateEkycStatus, nil)
				repo.On("Update", context.Background(), domain.EkycRequestLog{
					Id:          mockEkycRequestLogUpdateEkycStatus.Id,
					Email:       mockEkycRequestLogUpdateEkycStatus.Email,
					UserId:      mockEkycRequestLogUpdateEkycStatus.UserId,
					EkycId:      mockEkycRequestLogUpdateEkycStatus.EkycId,
					ApplicantId: mockEkycRequestLogUpdateEkycStatus.ApplicantId,
					Status:      domain.EKYC_REQUEST_STATUS_SUCCESS,
				}).Return(nil)
				return repo
			},
			userRepository: func() cdomain.UserRepository {
				repo := &mocks.UserRepository{}
				holdUser := mockUserUpdateEkycStatus
				holdUser.AuditStatus = domain.HOLD_AUDIT
				repo.On("GetById", context.Background(), mockUserUpdateEkycStatus.Id).Return(holdUser, nil)
				repo.On("Update", context.Background(), domain.User{
					Id:           mockUserUpdateEkycStatus.Id,
					Email:        mockUserUpdateEkycStatus.Email,
					SignupStatus: domain.SIGNUP_COMPLETE,
					AuditStatus:  domain.WAIT_AUDIT,
				}).Return(nil)
				return repo
			},
			wantErr: false,
		},
		{
			// MIR-1358: result=MATCHだが審査結果(CLEARED/DETECTED)がまだ
			// 確定していない会員。断るのはUNMATCHだけなので通す。
			name: "(成功) eKYC状態更新: 最新の履歴がNONE/MATCHなら通す(断りすぎていないことの確認)",
			args: args{
				ctx: context.Background(),
				request: udomain.PrvUserUpdateEkycStatusRequest{
					JWTUser: jwthelper.User{Id: mockEkycRequestLogUpdateEkycStatus.UserId},
					Status:  "success",
					UserId:  mockEkycRequestLogUpdateEkycStatus.UserId,
				},
			},
			ekycRequestLogRepository: func() cdomain.EkycRequestLogRepository {
				repo := &mocks.EkycRequestLogRepository{}
				matchedLog := mockEkycRequestLogUpdateEkycStatus
				matchedLog.Result = domain.EKYC_RESULT_STATUS_MATCH
				repo.On("GetByUserId", context.Background(), mockEkycRequestLogUpdateEkycStatus.UserId).Return(matchedLog, nil)
				repo.On("Update", context.Background(), domain.EkycRequestLog{
					Id:          mockEkycRequestLogUpdateEkycStatus.Id,
					Email:       mockEkycRequestLogUpdateEkycStatus.Email,
					UserId:      mockEkycRequestLogUpdateEkycStatus.UserId,
					EkycId:      mockEkycRequestLogUpdateEkycStatus.EkycId,
					ApplicantId: mockEkycRequestLogUpdateEkycStatus.ApplicantId,
					Result:      domain.EKYC_RESULT_STATUS_MATCH,
					Status:      domain.EKYC_REQUEST_STATUS_SUCCESS,
				}).Return(nil)
				return repo
			},
			userRepository: func() cdomain.UserRepository {
				repo := &mocks.UserRepository{}
				repo.On("GetById", context.Background(), mockUserUpdateEkycStatus.Id).Return(mockUserUpdateEkycStatus, nil)
				repo.On("Update", context.Background(), domain.User{
					Id:           mockUserUpdateEkycStatus.Id,
					Email:        mockUserUpdateEkycStatus.Email,
					SignupStatus: domain.SIGNUP_COMPLETE,
					AuditStatus:  domain.WAIT_AUDIT,
				}).Return(nil)
				return repo
			},
			wantErr: false,
		},
		{
			// MIR-1358: 失敗の報告は、resultがUNMATCHでも今までどおりFAILに
			// 更新する(断るのはstatus=successの完了報告だけ)。
			name: "(成功) eKYC状態更新: 本人認証失敗はresultがUNMATCHでも断らない",
			args: args{
				ctx: context.Background(),
				request: udomain.PrvUserUpdateEkycStatusRequest{
					JWTUser: jwthelper.User{Id: mockEkycRequestLogUpdateEkycStatus.UserId},
					Status:  "fail",
					UserId:  mockEkycRequestLogUpdateEkycStatus.UserId,
				},
			},
			ekycRequestLogRepository: func() cdomain.EkycRequestLogRepository {
				repo := &mocks.EkycRequestLogRepository{}
				unmatchedLog := mockEkycRequestLogUpdateEkycStatus
				unmatchedLog.Result = domain.EKYC_RESULT_STATUS_UNMATCH
				repo.On("GetByUserId", context.Background(), mockEkycRequestLogUpdateEkycStatus.UserId).Return(unmatchedLog, nil)
				repo.On("Update", context.Background(), domain.EkycRequestLog{
					Id:          mockEkycRequestLogUpdateEkycStatus.Id,
					Email:       mockEkycRequestLogUpdateEkycStatus.Email,
					UserId:      mockEkycRequestLogUpdateEkycStatus.UserId,
					EkycId:      mockEkycRequestLogUpdateEkycStatus.EkycId,
					ApplicantId: mockEkycRequestLogUpdateEkycStatus.ApplicantId,
					Result:      domain.EKYC_RESULT_STATUS_UNMATCH,
					Status:      domain.EKYC_REQUEST_STATUS_FAIL,
				}).Return(nil)
				return repo
			},
			userRepository: func() cdomain.UserRepository {
				repo := &mocks.UserRepository{}
				return repo
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ekycRepo := tt.ekycRequestLogRepository()
			userRepo := tt.userRepository()
			usecase := privateUserUsecase{
				ekycRequestLogRepository: ekycRepo,
				userRepository:           userRepo,
				logger:                   factory.NewLoggerFactory(),
			}
			err := usecase.UpdateEkycStatus(tt.args.ctx, tt.args.request)

			assert.Equalf(t, tt.wantErr, err != nil, "UpdateEkycStatus() error = %v, wantErr %v", err, tt.wantErr)
			if tt.wantErrIs != nil {
				assert.ErrorIs(t, err, tt.wantErrIs)
			}
			if tt.wantErr {
				// 断ったときは会員も履歴も更新しない
				userRepo.(*mocks.UserRepository).AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
				ekycRepo.(*mocks.EkycRequestLogRepository).AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
			}
		})
	}
}

func TestPrivateUserUsecase_GetEkycStatus(t *testing.T) {
	type args struct {
		ctx     context.Context
		request udomain.PrvUserGetEkycStatusRequest
	}

	tests := []struct {
		name                     string
		args                     args
		ekycRequestLogRepository func() cdomain.EkycRequestLogRepository
		want                     udomain.PrvUserGetEkycStatusResponse
		wantErr                  bool
	}{
		{
			name: "(成功) eKYC状態照会: eKYC状態がない場合",
			args: args{
				ctx: context.Background(),
				request: udomain.PrvUserGetEkycStatusRequest{
					UserId: int64(1),
				},
			},
			ekycRequestLogRepository: func() cdomain.EkycRequestLogRepository {
				repo := &mocks.EkycRequestLogRepository{}
				repo.On("GetByUserId", context.Background(), int64(1)).Return(domain.EkycRequestLog{
					Id:     int64(1),
					UserId: int64(1),
					Status: domain.EKYC_REQUEST_STATUS_NONE,
				}, nil)
				return repo
			},
			want: udomain.PrvUserGetEkycStatusResponse{
				Status: nil,
			},
			wantErr: false,
		},
		{
			name: "(成功) eKYC状態照会: eKYC成功状態",
			args: args{
				ctx: context.Background(),
				request: udomain.PrvUserGetEkycStatusRequest{
					UserId: int64(1),
				},
			},
			ekycRequestLogRepository: func() cdomain.EkycRequestLogRepository {
				repo := &mocks.EkycRequestLogRepository{}
				repo.On("GetByUserId", context.Background(), int64(1)).Return(domain.EkycRequestLog{
					Id:     int64(1),
					UserId: int64(1),
					Status: domain.EKYC_REQUEST_STATUS_SUCCESS,
				}, nil)
				return repo
			},
			want: udomain.PrvUserGetEkycStatusResponse{
				Status: lo.ToPtr(true),
			},
			wantErr: false,
		},
		{
			name: "(成功) eKYC状態照会: eKYC失敗状態",
			args: args{
				ctx: context.Background(),
				request: udomain.PrvUserGetEkycStatusRequest{
					UserId: int64(1),
				},
			},
			ekycRequestLogRepository: func() cdomain.EkycRequestLogRepository {
				repo := &mocks.EkycRequestLogRepository{}
				repo.On("GetByUserId", context.Background(), int64(1)).Return(domain.EkycRequestLog{
					Id:     int64(1),
					UserId: int64(1),
					Status: domain.EKYC_REQUEST_STATUS_FAIL,
				}, nil)
				return repo
			},
			want: udomain.PrvUserGetEkycStatusResponse{
				Status: lo.ToPtr(false),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := privateUserUsecase{
				ekycRequestLogRepository: tt.ekycRequestLogRepository(),
			}
			got, err := usecase.GetEkycStatus(tt.args.ctx, tt.args.request)

			assert.Equalf(t, tt.wantErr, err != nil, "GetEkycStatus() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want, got, "GetEkycStatus() = %v, want %v", got, tt.want)
		})
	}
}

func TestPrivateUserUsecase_GetSignUpStatus(t *testing.T) {
	type args struct {
		ctx     context.Context
		request udomain.PrvUserGetSignUpStatusRequest
	}

	truePtr := lo.ToPtr(true)
	falsePtr := lo.ToPtr(false)
	expectedBasicProfileBirthday := mockPrivateUserGetSignUpStatus.PrivateProfile.Birthday.ToString()

	detailedPrivateUser := domain.PrivateUser{
		User: domain.User{
			Id:                200,
			Name:              "詳細テスト",
			Email:             "detail@example.com",
			SignupStatus:      domain.SIGNUP_INVESTMENT_SUITABILITY,
			ZipCode:           "1500001",
			Address:           "東京都渋谷区神宮前1-1-1",
			Address2:          lo.ToPtr("テストマンション101"),
			Province:          "東京都",
			City:              "渋谷区",
			MobilePhone:       "08000000000",
			BankName:          "エナジー銀行",
			BranchName:        "青山支店",
			BankCode:          lo.ToPtr("0010"),
			BranchCode:        lo.ToPtr("0020"),
			BankAccount:       "1234567890",
			Depositor:         "詳細テスト",
			DepositType:       domain.ORDINARY_ACCOUNT,
			InvestmentPurpose: domain.RENEWABLE_ENERGY_INVESTMENT,
			IsSurplusFund:     truePtr,
			HasExperience:     falsePtr,
		},
		PrivateProfile: domain.PrivateProfile{
			UserId:                  200,
			FirstName:               "詳細",
			FamilyName:              "テスト",
			FirstNameKana:           "ショウサイ",
			FamilyNameKana:          "テスト",
			Birthday:                jsonutils.JSONTime(time.Date(1990, time.January, 2, 0, 0, 0, 0, privateUserTestJST)),
			Gender:                  domain.FEMALE,
			Nationality:             domain.JP,
			Job:                     domain.USER_LISTED_COMPANY_EMPLOYEE,
			Asset:                   domain.ASSET4,
			AnnualIncome:            domain.ANNUALINCOME3,
			IsProInvestAsset:        truePtr,
			IsProInvestFinanceAsset: falsePtr,
		},
	}

	privateUserOther := detailedPrivateUser
	privateUserOther.User.Id = 201
	privateUserOther.User.Email = "other@example.com"
	privateUserOther.User.SignupStatus = domain.SIGNUP_PRIVATE_PROFILE_OTHER
	privateUserOther.PrivateProfile.UserId = privateUserOther.User.Id

	privateUserAccount := detailedPrivateUser
	privateUserAccount.User.Id = 202
	privateUserAccount.User.Email = "account@example.com"
	privateUserAccount.User.SignupStatus = domain.SIGNUP_ACCOUNT
	privateUserAccount.PrivateProfile.UserId = privateUserAccount.User.Id

	privateUserInvestment := detailedPrivateUser
	privateUserInvestment.User.Id = 203
	privateUserInvestment.User.Email = "investment@example.com"
	privateUserInvestment.User.SignupStatus = domain.SIGNUP_INVESTMENT_SUITABILITY
	privateUserInvestment.PrivateProfile.UserId = privateUserInvestment.User.Id

	privateUserNone := domain.PrivateUser{
		User: domain.User{
			Id:           301,
			Email:        "none@example.com",
			SignupStatus: domain.SIGNUP_NONE,
		},
	}

	privateUserAgree := privateUserNone
	privateUserAgree.User.Id = 302
	privateUserAgree.User.Email = "agree@example.com"
	privateUserAgree.User.SignupStatus = domain.SIGNUP_AGREE

	privateUserKyc := privateUserNone
	privateUserKyc.User.Id = 303
	privateUserKyc.User.Email = "kyc@example.com"
	privateUserKyc.User.SignupStatus = domain.SIGNUP_KYC

	privateUserComplete := privateUserNone
	privateUserComplete.User.Id = 304
	privateUserComplete.User.Email = "complete@example.com"
	privateUserComplete.User.SignupStatus = domain.SIGNUP_COMPLETE

	tests := []struct {
		name                  string
		args                  args
		privateUserRepository func() cdomain.PrivateUserRepository
		want                  udomain.PrvUserGetSignUpStatusResponse
		wantErr               bool
	}{
		// 正常系
		{
			name: "Success: Status未設定",
			args: args{
				ctx: context.Background(),
				request: udomain.PrvUserGetSignUpStatusRequest{
					JWTUser: jwthelper.User{
						Id: mockPrivateUserGetSignUpStatus.User.Id,
					},
				},
			},
			privateUserRepository: func() cdomain.PrivateUserRepository {
				repo := &mocks.PrivateUserRepository{}
				repo.On("GetById", context.Background(), mockPrivateUserGetSignUpStatus.User.Id).
					Return(mockPrivateUserGetSignUpStatus, nil)
				return repo
			},
			want: udomain.PrvUserGetSignUpStatusResponse{
				Id:         mockPrivateUserGetSignUpStatus.User.Id,
				Email:      mockPrivateUserGetSignUpStatus.User.Email,
				PrevStatus: domain.SIGNUP_NONE,
				NextStatus: domain.SIGNUP_PRIVATE_PROFILE_OTHER,
				CurStatus:  domain.SIGNUP_PRIVATE_PROFILE_BASIC,
				BasicProfile: &udomain.PrvUserBasicProfile{
					FirstName:      "怜",
					FamilyName:     "直井",
					FirstNameKana:  "レイ",
					FamilyNameKana: "ナオイ",
					Birthday:       expectedBasicProfileBirthday,
					Gender:         domain.FEMALE,
					Nationality:    domain.JP,
				},
			},
			wantErr: false,
		},
		{
			name: "Success: StatusがSIGNUP_NONE",
			args: args{
				ctx: context.Background(),
				request: udomain.PrvUserGetSignUpStatusRequest{
					JWTUser: jwthelper.User{Id: privateUserNone.User.Id},
					Status:  domain.SIGNUP_NONE,
				},
			},
			privateUserRepository: func() cdomain.PrivateUserRepository {
				repo := &mocks.PrivateUserRepository{}
				repo.On("GetById", context.Background(), privateUserNone.User.Id).Return(privateUserNone, nil)
				return repo
			},
			want: udomain.PrvUserGetSignUpStatusResponse{
				Id:         privateUserNone.User.Id,
				Email:      privateUserNone.User.Email,
				NextStatus: domain.SIGNUP_PRIVATE_PROFILE_BASIC,
				CurStatus:  domain.SIGNUP_NONE,
			},
			wantErr: false,
		},
		{
			name: "Success: StatusがSIGNUP_AGREE",
			args: args{
				ctx: context.Background(),
				request: udomain.PrvUserGetSignUpStatusRequest{
					JWTUser: jwthelper.User{Id: privateUserAgree.User.Id},
					Status:  domain.SIGNUP_AGREE,
				},
			},
			privateUserRepository: func() cdomain.PrivateUserRepository {
				repo := &mocks.PrivateUserRepository{}
				repo.On("GetById", context.Background(), privateUserAgree.User.Id).Return(privateUserAgree, nil)
				return repo
			},
			want: udomain.PrvUserGetSignUpStatusResponse{
				Id:         privateUserAgree.User.Id,
				Email:      privateUserAgree.User.Email,
				PrevStatus: domain.SIGNUP_INVESTMENT_SUITABILITY,
				NextStatus: domain.SIGNUP_KYC,
				CurStatus:  domain.SIGNUP_AGREE,
			},
			wantErr: false,
		},
		{
			name: "Success: StatusがSIGNUP_KYC",
			args: args{
				ctx: context.Background(),
				request: udomain.PrvUserGetSignUpStatusRequest{
					JWTUser: jwthelper.User{Id: privateUserKyc.User.Id},
					Status:  domain.SIGNUP_KYC,
				},
			},
			privateUserRepository: func() cdomain.PrivateUserRepository {
				repo := &mocks.PrivateUserRepository{}
				repo.On("GetById", context.Background(), privateUserKyc.User.Id).Return(privateUserKyc, nil)
				return repo
			},
			want: udomain.PrvUserGetSignUpStatusResponse{
				Id:         privateUserKyc.User.Id,
				Email:      privateUserKyc.User.Email,
				PrevStatus: domain.SIGNUP_AGREE,
				NextStatus: domain.SIGNUP_COMPLETE,
				CurStatus:  domain.SIGNUP_KYC,
			},
			wantErr: false,
		},
		{
			name: "Success: StatusがSIGNUP_COMPLETE",
			args: args{
				ctx: context.Background(),
				request: udomain.PrvUserGetSignUpStatusRequest{
					JWTUser: jwthelper.User{Id: privateUserComplete.User.Id},
					Status:  domain.SIGNUP_COMPLETE,
				},
			},
			privateUserRepository: func() cdomain.PrivateUserRepository {
				repo := &mocks.PrivateUserRepository{}
				repo.On("GetById", context.Background(), privateUserComplete.User.Id).Return(privateUserComplete, nil)
				return repo
			},
			want: udomain.PrvUserGetSignUpStatusResponse{
				Id:         privateUserComplete.User.Id,
				Email:      privateUserComplete.User.Email,
				PrevStatus: domain.SIGNUP_KYC,
				CurStatus:  domain.SIGNUP_COMPLETE,
			},
			wantErr: false,
		},
		{
			name: "Success: StatusがSIGNUP_PRIVATE_PROFILE_BASIC",
			args: args{
				ctx: context.Background(),
				request: udomain.PrvUserGetSignUpStatusRequest{
					JWTUser: jwthelper.User{
						Id: mockPrivateUserGetSignUpStatus.User.Id,
					},
					Status: domain.SIGNUP_PRIVATE_PROFILE_BASIC,
				},
			},
			privateUserRepository: func() cdomain.PrivateUserRepository {
				repo := &mocks.PrivateUserRepository{}
				repo.On("GetById", context.Background(), mockPrivateUserGetSignUpStatus.User.Id).
					Return(mockPrivateUserGetSignUpStatus, nil)
				return repo
			},
			want: udomain.PrvUserGetSignUpStatusResponse{
				Id:         mockPrivateUserGetSignUpStatus.User.Id,
				Email:      mockPrivateUserGetSignUpStatus.User.Email,
				PrevStatus: domain.SIGNUP_NONE,
				NextStatus: domain.SIGNUP_PRIVATE_PROFILE_OTHER,
				CurStatus:  domain.SIGNUP_PRIVATE_PROFILE_BASIC,
				BasicProfile: &udomain.PrvUserBasicProfile{
					FirstName:      "怜",
					FamilyName:     "直井",
					FirstNameKana:  "レイ",
					FamilyNameKana: "ナオイ",
					Birthday:       expectedBasicProfileBirthday,
					Gender:         domain.FEMALE,
					Nationality:    domain.JP,
				},
			},
			wantErr: false,
		},
		{
			name: "Success: StatusがSIGNUP_PRIVATE_PROFILE_OTHER",
			args: args{
				ctx: context.Background(),
				request: udomain.PrvUserGetSignUpStatusRequest{
					JWTUser: jwthelper.User{Id: privateUserOther.User.Id},
					Status:  domain.SIGNUP_PRIVATE_PROFILE_OTHER,
				},
			},
			privateUserRepository: func() cdomain.PrivateUserRepository {
				repo := &mocks.PrivateUserRepository{}
				repo.On("GetById", context.Background(), privateUserOther.User.Id).Return(privateUserOther, nil)
				return repo
			},
			want: udomain.PrvUserGetSignUpStatusResponse{
				Id:         privateUserOther.User.Id,
				Email:      privateUserOther.User.Email,
				PrevStatus: domain.SIGNUP_PRIVATE_PROFILE_BASIC,
				NextStatus: domain.SIGNUP_ACCOUNT,
				CurStatus:  domain.SIGNUP_PRIVATE_PROFILE_OTHER,
				OtherProfile: &udomain.PrvUserOtherProfile{
					ZipCode:     privateUserOther.ZipCode,
					Address:     privateUserOther.Address,
					Province:    privateUserOther.Province,
					Address2:    privateUserOther.Address2,
					City:        privateUserOther.City,
					MobilePhone: privateUserOther.MobilePhone,
					Job:         privateUserOther.Job,
				},
			},
			wantErr: false,
		},
		{
			name: "Success: StatusがSIGNUP_ACCOUNT",
			args: args{
				ctx: context.Background(),
				request: udomain.PrvUserGetSignUpStatusRequest{
					JWTUser: jwthelper.User{Id: privateUserAccount.User.Id},
					Status:  domain.SIGNUP_ACCOUNT,
				},
			},
			privateUserRepository: func() cdomain.PrivateUserRepository {
				repo := &mocks.PrivateUserRepository{}
				repo.On("GetById", context.Background(), privateUserAccount.User.Id).Return(privateUserAccount, nil)
				return repo
			},
			want: udomain.PrvUserGetSignUpStatusResponse{
				Id:         privateUserAccount.User.Id,
				Email:      privateUserAccount.User.Email,
				PrevStatus: domain.SIGNUP_PRIVATE_PROFILE_OTHER,
				NextStatus: domain.SIGNUP_INVESTMENT_SUITABILITY,
				CurStatus:  domain.SIGNUP_ACCOUNT,
				WithdrawalAccount: &udomain.PrvUserWithdrawalAccount{
					BankName:    privateUserAccount.BankName,
					BranchName:  privateUserAccount.BranchName,
					BankAccount: privateUserAccount.BankAccount,
					BankCode:    "0010",
					BranchCode:  "0020",
					Depositor:   privateUserAccount.Depositor,
					DepositType: privateUserAccount.DepositType,
				},
			},
			wantErr: false,
		},
		{
			name: "Success: StatusがSIGNUP_INVESTMENT_SUITABILITY",
			args: args{
				ctx: context.Background(),
				request: udomain.PrvUserGetSignUpStatusRequest{
					JWTUser: jwthelper.User{Id: privateUserInvestment.User.Id},
					Status:  domain.SIGNUP_INVESTMENT_SUITABILITY,
				},
			},
			privateUserRepository: func() cdomain.PrivateUserRepository {
				repo := &mocks.PrivateUserRepository{}
				repo.On("GetById", context.Background(), privateUserInvestment.User.Id).Return(privateUserInvestment, nil)
				return repo
			},
			want: udomain.PrvUserGetSignUpStatusResponse{
				Id:         privateUserInvestment.User.Id,
				Email:      privateUserInvestment.User.Email,
				PrevStatus: domain.SIGNUP_ACCOUNT,
				NextStatus: domain.SIGNUP_AGREE,
				CurStatus:  domain.SIGNUP_INVESTMENT_SUITABILITY,
				InvestmentSuitability: &udomain.PrvUserInvestmentSuitability{
					Asset:                   privateUserInvestment.Asset,
					AnnualIncome:            privateUserInvestment.AnnualIncome,
					IsProInvestAsset:        privateUserInvestment.IsProInvestAsset,
					IsProInvestFinanceAsset: privateUserInvestment.IsProInvestFinanceAsset,
					InvestmentPurpose:       privateUserInvestment.InvestmentPurpose,
					IsSurplusFund:           true,
					HasExperience:           false,
				},
			},
			wantErr: false,
		},
		// 異常系
		{
			name: "Failed: SIGNUP_STATUSが異常値",
			args: args{
				ctx: context.Background(),
				request: udomain.PrvUserGetSignUpStatusRequest{
					JWTUser: jwthelper.User{
						Id: mockPrivateUserGetSignUpStatus.User.Id,
					},
					Status: "INVALID_STATUS",
				},
			},
			privateUserRepository: func() cdomain.PrivateUserRepository {
				repo := &mocks.PrivateUserRepository{}
				repo.On("GetById", context.Background(), mockPrivateUserGetSignUpStatus.User.Id).
					Return(mockPrivateUserGetSignUpStatus, nil)
				return repo
			},
			want:    udomain.PrvUserGetSignUpStatusResponse{},
			wantErr: true,
		},
		{
			name: "Failed: privateUserの取得に失敗",
			args: args{
				ctx: context.Background(),
				request: udomain.PrvUserGetSignUpStatusRequest{
					JWTUser: jwthelper.User{Id: 999},
				},
			},
			privateUserRepository: func() cdomain.PrivateUserRepository {
				repo := &mocks.PrivateUserRepository{}
				repo.On("GetById", context.Background(), int64(999)).Return(domain.PrivateUser{}, errors.New("repository error"))
				return repo
			},
			want:    udomain.PrvUserGetSignUpStatusResponse{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := privateUserUsecase{
				privateUserRepository: tt.privateUserRepository(),
			}
			got, err := usecase.GetSignUpStatus(tt.args.ctx, tt.args.request)

			assert.Equalf(t, tt.wantErr, err != nil, "GetSignUpStatus() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want, got, "GetSignUpStatus() = %v, want %v", got, tt.want)
		})
	}
}

func TestPrivateUserUsecase_CallEkycWebhook(t *testing.T) {
	uid, _ := uuid.FromString("1234")

	type args struct {
		ctx     context.Context
		request udomain.PrvUserCallEkycWebhookRequest
	}

	tests := []struct {
		name                      string
		args                      args
		cfg                       *config.Config
		idGenerator               func() idgenerator.IdGenerator
		mailService               func() *mocks.MailService
		userVirtualAccountService func() *mocks.UserVirtualAccountService
		ekycRequestLogRepository  func() *mocks.EkycRequestLogRepository
		myNumberRepository        func() *mocks.MyNumberRepository
		userRepository            func() *mocks.UserRepository
		wantErr                   bool
	}{
		{
			name: "eKYC成功",
			args: args{
				ctx: context.Background(),
				request: udomain.PrvUserCallEkycWebhookRequest{
					UserId: int64(1),
					Status: "success",
				},
			},
			idGenerator: func() idgenerator.IdGenerator {
				idGenerator := &mocks.IdGenerator{}
				return idGenerator
			},
			mailService: func() *mocks.MailService {
				mailService := &mocks.MailService{}
				mailService.On("SendMail", context.Background(), cdomain.SendMailRequest{
					TemplateType: domain.COMPLETED_AUDIT_STATUS_FOR_PRIVATE_TEMPLATE,
					Sender:       "info@energyshares.jp",
					Receivers:    []string{"test@example.com"},
					MailType:     domain.MAIL_SEND_PRIORITY_NORMAL,
					Details: map[string]interface{}{
						"Receiver":   "user name",
						"ServiceUrl": "http://energyshares.jp",
						"ProductUrl": "http://energyshares.jp/product",
					},
				}).Return(nil)
				return mailService
			},
			userRepository: func() *mocks.UserRepository {
				userRepository := &mocks.UserRepository{}
				userRepository.On("GetById", context.Background(), int64(1)).Return(domain.User{
					Id:    int64(1),
					Email: "test@example.com",
					Name:  "user name",
				}, nil)
				userRepository.On("Update", context.Background(), mock.MatchedBy(func(request domain.User) bool {
					ignoreType := cmpopts.IgnoreTypes(jsonutils.JSONTime{})
					return cmp.Equal(request, domain.User{
						Id:          int64(1),
						Email:       "test@example.com",
						Name:        "user name",
						IsEkyc:      true,
						AuditStatus: domain.COMPLETED_AUDIT,
					}, ignoreType)
				})).Return(nil)
				return userRepository
			},
			ekycRequestLogRepository: func() *mocks.EkycRequestLogRepository {
				ekycRequestLogRepository := &mocks.EkycRequestLogRepository{}
				ekycRequestLogRepository.On("GetByUserId", context.Background(), int64(1)).Return(domain.EkycRequestLog{
					Id:     int64(1),
					UserId: int64(1),
				}, nil)
				ekycRequestLogRepository.On("Update", context.Background(), domain.EkycRequestLog{
					Id:     int64(1),
					UserId: int64(1),
					Result: domain.EKYC_RESULT_STATUS_MATCH,
				}).Return(nil)
				return ekycRequestLogRepository
			},
			myNumberRepository: func() *mocks.MyNumberRepository {
				myNumberRepository := &mocks.MyNumberRepository{}
				myNumberRepository.On("GetByUserId", context.Background(), int64(1)).Return(domain.MyNumber{
					Id:     int64(1),
					UserId: int64(1),
					Status: domain.APPROVE,
				}, nil)
				return myNumberRepository
			},
			userVirtualAccountService: func() *mocks.UserVirtualAccountService {
				userVirtualAccountService := &mocks.UserVirtualAccountService{}
				userVirtualAccountService.On("LinkVirtualAccount", context.Background(), int64(1), domain.EQUITY).Return(nil)
				return userVirtualAccountService
			},
		},
		{
			name: "eKYC失敗",
			args: args{
				ctx: context.Background(),
				request: udomain.PrvUserCallEkycWebhookRequest{
					UserId: int64(1),
					Status: "fail",
				},
			},
			idGenerator: func() idgenerator.IdGenerator {
				idGenerator := &mocks.IdGenerator{}
				idGenerator.On("GenUuidV4").Return(uid)
				return idGenerator
			},
			mailService: func() *mocks.MailService {
				mailService := &mocks.MailService{}
				return mailService
			},
			userRepository: func() *mocks.UserRepository {
				userRepository := &mocks.UserRepository{}
				userRepository.On("GetById", context.Background(), int64(1)).Return(domain.User{
					Id:    int64(1),
					Email: "test@example.com",
					Name:  "user name",
				}, nil)
				userRepository.On("Update", context.Background(), mock.MatchedBy(func(request domain.User) bool {
					ignoreType := cmpopts.IgnoreTypes(jsonutils.JSONTime{})
					return cmp.Equal(request, domain.User{
						Id:                int64(1),
						Email:             "test@example.com",
						Name:              "user name",
						AuditStatus:       domain.HOLD_AUDIT,
						HoldUserCheckCode: lo.ToPtr(uid.String()),
					}, ignoreType)
				})).Return(nil)
				return userRepository
			},
			ekycRequestLogRepository: func() *mocks.EkycRequestLogRepository {
				ekycRequestLogRepository := &mocks.EkycRequestLogRepository{}
				ekycRequestLogRepository.On("GetByUserId", context.Background(), int64(1)).Return(domain.EkycRequestLog{
					Id:     int64(1),
					UserId: int64(1),
				}, nil)
				ekycRequestLogRepository.On("Update", context.Background(), domain.EkycRequestLog{
					Id:           int64(1),
					UserId:       int64(1),
					Result:       domain.EKYC_RESULT_STATUS_UNMATCH,
					ErrorMessage: lo.ToPtr("顔照合結果がアンマッチ(同一人物でない)\n指定した確認書類の画像でない\n免許証番号の不整合を検出しました。\n公安委員会と住所の不整合を検出しました。\n"),
				}).Return(nil)
				return ekycRequestLogRepository
			},
			myNumberRepository: func() *mocks.MyNumberRepository {
				myNumberRepository := &mocks.MyNumberRepository{}
				myNumberRepository.On("GetByUserId", context.Background(), int64(1)).Return(domain.MyNumber{
					Id:     int64(1),
					UserId: int64(1),
					Status: domain.APPROVE,
				}, nil)
				return myNumberRepository
			},
			userVirtualAccountService: func() *mocks.UserVirtualAccountService {
				userVirtualAccountService := &mocks.UserVirtualAccountService{}
				userVirtualAccountService.On("LinkVirtualAccount", context.Background(), int64(1), domain.EQUITY).Return(nil)
				return userVirtualAccountService
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := privateUserUsecase{
				cfg: &config.Config{
					CommonConfig: config.CommonConfig{
						Env: "dev",
						ServiceConfig: struct {
							Mail serviceutils.MailConfig
							Url  serviceutils.UrlConfig
						}{
							Mail: serviceutils.MailConfig{
								Info: "info@energyshares.jp",
							},
							Url: serviceutils.UrlConfig{
								UserServiceDomain: "http://energyshares.jp",
								Product:           "/product",
							},
						},
					},
				},
				idGenerator:               tt.idGenerator(),
				mailService:               tt.mailService(),
				userVirtualAccountService: tt.userVirtualAccountService(),
				myNumberRepository:        tt.myNumberRepository(),
				ekycRequestLogRepository:  tt.ekycRequestLogRepository(),
				userRepository:            tt.userRepository(),
			}
			err := usecase.CallEkycWebhook(tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.wantErr, err != nil, "CallEkycWebhook() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func TestPrivateUserUsecase_GetOne(t *testing.T) {
	type args struct {
		ctx     context.Context
		request udomain.PrvUserGetOneRequest
	}
	tests := []struct {
		name                         string
		args                         args
		privateUserRepository        func() *mocks.PrivateUserRepository
		userVirtualAccountRepository func() *mocks.UserVirtualAccountRepository
		want                         udomain.PrvUserGetOneResponse
		wantErr                      bool
	}{
		{
			name: "ユーザー情報照会成功",
			args: args{
				ctx: context.Background(),
				request: udomain.PrvUserGetOneRequest{
					JWTUser: jwthelper.User{
						Id: mockPrivateUserGetOne.UserId,
					},
				},
			},
			privateUserRepository: func() *mocks.PrivateUserRepository {
				repo := &mocks.PrivateUserRepository{}
				repo.On("GetById", context.Background(), mockPrivateUserGetOne.UserId).Return(mockPrivateUserGetOne, nil)
				return repo
			},
			userVirtualAccountRepository: func() *mocks.UserVirtualAccountRepository {
				repo := &mocks.UserVirtualAccountRepository{}
				repo.On("GetByUserIdAndLinkedStatus", context.Background(), mockPrivateUserGetOne.UserId, domain.EQUITY, domain.ON_VA_ACCOUNT).
					Return(domain.UserVirtualAccount{
						VaBankName:       domain.GMO_AOZORA_BANK_NAME,
						VaBranchName:     "あじさい支店",
						VaAccountNumber:  "5194080",
						VaHolderNameKana: "カ）ハコブネ",
					}, nil)
				return repo
			},

			want: udomain.PrvUserGetOneResponse{
				UserId:       mockPrivateUserGetOne.UserId,
				Email:        mockPrivateUserGetOne.Email,
				IsNewsletter: true,
				Name:         mockPrivateUserGetOne.FamilyName + " " + mockPrivateUserGetOne.FirstName,
				NameKana:     mockPrivateUserGetOne.FamilyNameKana + " " + mockPrivateUserGetOne.FirstNameKana,
				Created:      mockPrivateUserGetOne.User.Created,
				Nationality:  mockPrivateUserGetOne.Nationality,
				Birthday:     mockPrivateUserGetOne.Birthday.ToFormat(time.DateOnly),
				ZipCode:      mockPrivateUserGetOne.User.ZipCode,
				Address:      mockPrivateUserGetOne.User.Address,
				Address2:     mockPrivateUserGetOne.User.Address2,
				Province:     mockPrivateUserGetOne.User.Province,
				City:         mockPrivateUserGetOne.User.City,
				Job:          mockPrivateUserGetOne.PrivateProfile.Job,
				MobilePhone:  mockPrivateUserGetOne.User.MobilePhone,
				AuditStatus:  mockPrivateUserGetOne.User.AuditStatus,
				WithdrawalAccount: udomain.PrvUserGetOneWithdrawalAccount{
					BankName:    mockPrivateUserGetOne.BankName,
					BranchName:  mockPrivateUserGetOne.BranchName,
					BankCode:    lo.FromPtr(mockPrivateUserGetOne.BankCode),
					BranchCode:  lo.FromPtr(mockPrivateUserGetOne.BranchCode),
					DepositType: mockPrivateUserGetOne.DepositType,
					BankAccount: mockPrivateUserGetOne.BankAccount,
					Depositor:   mockPrivateUserGetOne.Depositor,
				},
				InvestmentSuitability: udomain.PrvUserGetOneInvestmentSuitability{
					Asset:                   mockPrivateUserGetOne.PrivateProfile.Asset,
					AnnualIncome:            mockPrivateUserGetOne.PrivateProfile.AnnualIncome,
					IsProInvestAsset:        mockPrivateUserGetOne.PrivateProfile.IsProInvestAsset,
					IsProInvestFinanceAsset: mockPrivateUserGetOne.PrivateProfile.IsProInvestFinanceAsset,
					InvestmentPurpose:       mockPrivateUserGetOne.User.InvestmentPurpose,
				},
				VirtualAccount: udomain.PrvUserGetOneVirtualAccount{
					BankName:       domain.GMO_AOZORA_BANK_NAME,
					BranchName:     "あじさい支店",
					AccountNumber:  "5194080",
					HolderNameKana: "カ）ハコブネ",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := privateUserUsecase{
				privateUserRepository:        tt.privateUserRepository(),
				userVirtualAccountRepository: tt.userVirtualAccountRepository(),
			}

			got, err := usecase.GetOne(tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.wantErr, err != nil, "GetOne() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want, got, "GetOne() = %v, want %v", got, tt.want)
		})
	}
}

func Test_privateUserUsecase_PatchOne(t *testing.T) {
	type args struct {
		ctx     context.Context
		request udomain.PrvUserPatchOneRequest
	}
	tests := []struct {
		name           string
		args           args
		userRepository func() *mocks.UserRepository
		wantErr        bool
	}{
		{
			name: "会員情報修正成功 - ニュースレターをtrue値に更新",
			args: args{
				ctx: context.Background(),
				request: udomain.PrvUserPatchOneRequest{
					JWTUser: jwthelper.User{
						Id: 1000,
					},
					IsNewsletter: lo.ToPtr(true),
				},
			},
			userRepository: func() *mocks.UserRepository {
				repo := &mocks.UserRepository{}

				repo.On("GetById", context.Background(), int64(1000)).Return(domain.User{
					Id:           1000,
					UserType:     domain.PRIVATE,
					IsNewsletter: false,
				}, nil)
				repo.On("Update", context.Background(), domain.User{
					Id:           1000,
					UserType:     domain.PRIVATE,
					IsNewsletter: true,
				}).Return(nil)
				return repo
			},
		},
		{
			name: "会員情報修正成功 - ニュースレター値をnil値に更新を試行",
			args: args{
				ctx: context.Background(),
				request: udomain.PrvUserPatchOneRequest{
					JWTUser: jwthelper.User{
						Id: 1000,
					},
				},
			},
			userRepository: func() *mocks.UserRepository {
				repo := &mocks.UserRepository{}

				repo.On("GetById", context.Background(), int64(1000)).Return(domain.User{
					Id:           1000,
					UserType:     domain.PRIVATE,
					IsNewsletter: false,
				}, nil)
				repo.On("Update", context.Background(), domain.User{
					Id:           1000,
					UserType:     domain.PRIVATE,
					IsNewsletter: false,
				}).Return(nil)
				return repo
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := privateUserUsecase{
				userRepository: tt.userRepository(),
			}

			err := usecase.PatchOne(tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.wantErr, err != nil, "PatchOne() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_buildEkycAddress(t *testing.T) {
	tests := []struct {
		name     string
		province string
		city     string
		address  string
		address2 *string
		want     string
	}{
		{
			name:     "全て揃っている場合は区切りなしで連結し、建物名の前にだけ空白が入る",
			province: "東京都",
			city:     "渋谷区",
			address:  "道玄坂1-2-3",
			address2: lo.ToPtr("〇〇ビル401"),
			want:     "東京都渋谷区道玄坂1-2-3 〇〇ビル401",
		},
		{
			name:     "建物名(address2)がnilなら末尾に空白も付かない",
			province: "東京都",
			city:     "渋谷区",
			address:  "道玄坂1-2-3",
			address2: nil,
			want:     "東京都渋谷区道玄坂1-2-3",
		},
		{
			name:     "建物名が空白のみなら無いものとして扱う",
			province: "東京都",
			city:     "渋谷区",
			address:  "道玄坂1-2-3",
			address2: lo.ToPtr("   "),
			want:     "東京都渋谷区道玄坂1-2-3",
		},
		{
			name:     "市区町村が空でも、他の項目だけで詰めて連結する(項目ごとのフォールバック)",
			province: "東京都",
			city:     "",
			address:  "道玄坂1-2-3",
			address2: nil,
			want:     "東京都道玄坂1-2-3",
		},
		{
			name:     "全て空なら空文字を返す(呼び出し元でこれを検知して送信をブロックする)",
			province: "",
			city:     "",
			address:  "",
			address2: nil,
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildEkycAddress(tt.province, tt.city, tt.address, tt.address2)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_missingEkycProfileFields(t *testing.T) {
	complete := domain.PrivateUser{
		User: domain.User{
			Province: "東京都",
			City:     "渋谷区",
			Address:  "道玄坂1-2-3",
		},
		PrivateProfile: domain.PrivateProfile{
			Birthday:   jsonutils.JSONTime(time.Date(1990, 1, 2, 0, 0, 0, 0, time.UTC)),
			FamilyName: "山田",
			FirstName:  "太郎",
		},
	}

	tests := []struct {
		name        string
		mutate      func(domain.PrivateUser) domain.PrivateUser
		wantMissing []string
	}{
		{
			name:        "全て揃っていれば欠落なし",
			mutate:      func(u domain.PrivateUser) domain.PrivateUser { return u },
			wantMissing: nil,
		},
		{
			name: "姓が空なら family_name を返す",
			mutate: func(u domain.PrivateUser) domain.PrivateUser {
				u.FamilyName = ""
				return u
			},
			wantMissing: []string{"family_name"},
		},
		{
			name: "姓が空白のみでも family_name を返す(TrimSpaceで拾う)",
			mutate: func(u domain.PrivateUser) domain.PrivateUser {
				u.FamilyName = "   "
				return u
			},
			wantMissing: []string{"family_name"},
		},
		{
			name: "名が空なら first_name を返す",
			mutate: func(u domain.PrivateUser) domain.PrivateUser {
				u.FirstName = ""
				return u
			},
			wantMissing: []string{"first_name"},
		},
		{
			name: "生年月日が未設定(ゼロ値)なら birthday を返す",
			mutate: func(u domain.PrivateUser) domain.PrivateUser {
				u.PrivateProfile.Birthday = jsonutils.JSONTime{}
				return u
			},
			wantMissing: []string{"birthday"},
		},
		{
			name: "都道府県が空なら province を返す",
			mutate: func(u domain.PrivateUser) domain.PrivateUser {
				u.Province = ""
				return u
			},
			wantMissing: []string{"province"},
		},
		{
			name: "市区町村が空なら city を返す",
			mutate: func(u domain.PrivateUser) domain.PrivateUser {
				u.City = ""
				return u
			},
			wantMissing: []string{"city"},
		},
		{
			name: "住所が空白のみでも address を返す(TrimSpaceで拾う)",
			mutate: func(u domain.PrivateUser) domain.PrivateUser {
				u.Address = "  "
				return u
			},
			wantMissing: []string{"address"},
		},
		{
			name: "複数欠けていれば全て列挙する",
			mutate: func(u domain.PrivateUser) domain.PrivateUser {
				u.FamilyName = ""
				u.Address = ""
				return u
			},
			wantMissing: []string{"family_name", "address"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := missingEkycProfileFields(tt.mutate(complete))
			assert.Equal(t, tt.wantMissing, got)
		})
	}
}
