package corporate_user

import (
	"context"
	"errors"
	"testing"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/filehelper"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain/mocks"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/s3utils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
	udomain "github.com/illuminarean/energyshares-user-api/domain"
)

func Test_corporateUserUsecase_RegisterProfiles(t *testing.T) {
	type args struct {
		ctx     context.Context
		request udomain.CorporateUserRegisterProfilesRequest
	}
	tests := []struct {
		name                           string
		args                           args
		userRepository                 func() *mocks.UserRepository
		corporateProfileRepository     func() *mocks.CorporateProfileRepository
		corporateShareholderRepository func() *mocks.CorporateShareholderRepository
		wantErr                        bool
	}{
		{
			name: "基本プロフィール登録成功",
			args: args{
				ctx: context.Background(),
				request: udomain.CorporateUserRegisterProfilesRequest{
					User: jwthelper.User{
						Id: 1000,
					},
					BasicProfile: udomain.CorporateUserBasicProfileItem{
						CorpNo:   "1234567890123",
						CorpName: "イルミナリアン",
						CorpType: "IT企業",
					},
					SignUpStatus: domain.SIGNUP_CORPORATE_PROFILE_BASIC,
				},
			},
			userRepository: func() *mocks.UserRepository {
				repo := &mocks.UserRepository{}

				copyUser1 := mockUser1
				copyUser1.SignupStatus = domain.SIGNUP_NONE
				repo.On("GetById", context.Background(), int64(1000)).Return(copyUser1, nil)

				copyUser1.Name = "イルミナリアン"
				copyUser1.SignupStatus = domain.SIGNUP_CORPORATE_PROFILE_BASIC

				repo.On("Update", context.Background(), copyUser1).Return(nil)
				return repo
			},

			corporateProfileRepository: func() *mocks.CorporateProfileRepository {
				repo := &mocks.CorporateProfileRepository{}

				copyCorporateProfile1 := mockCorporateProfile1

				repo.On("GetByUserId", context.Background(), int64(1000)).Return(copyCorporateProfile1, nil)
				repo.On("GetByCorpNo", context.Background(), "1234567890123").Return(domain.CorporateProfile{}, cdomain.ErrItemNotFound)

				copyCorporateProfile1.CorpNo = "1234567890123"
				copyCorporateProfile1.CorpType = "IT企業"

				repo.On("Update", context.Background(), copyCorporateProfile1).Return(nil)
				return repo
			},
			corporateShareholderRepository: func() *mocks.CorporateShareholderRepository {
				repo := &mocks.CorporateShareholderRepository{}
				return repo
			},
			wantErr: false,
		},
		{
			name: "その他プロフィール登録成功",
			args: args{
				ctx: context.Background(),
				request: udomain.CorporateUserRegisterProfilesRequest{
					User: jwthelper.User{
						Id: 1000,
					},
					OtherProfile: udomain.CorporateUserOtherProfileItem{
						ZipCode:     "1234567",
						Province:    "都道府県",
						City:        "市群区",
						Address:     "詳細住所",
						Address2:    lo.ToPtr("詳細住所2"),
						MobilePhone: "01012345678",
					},
					SignUpStatus: domain.SIGNUP_CORPORATE_PROFILE_OTHER,
				},
			},
			userRepository: func() *mocks.UserRepository {
				repo := &mocks.UserRepository{}

				copyUser1 := mockUser1
				copyUser1.SignupStatus = domain.SIGNUP_CORPORATE_PROFILE_OTHER

				repo.On("GetById", context.Background(), int64(1000)).Return(copyUser1, nil)

				copyUser1.ZipCode = "1234567"
				copyUser1.Province = "都道府県"
				copyUser1.City = "市群区"
				copyUser1.Address = "詳細住所"
				copyUser1.Address2 = lo.ToPtr("詳細住所2")
				copyUser1.MobilePhone = "01012345678"

				repo.On("Update", context.Background(), copyUser1).Return(nil)
				return repo
			},

			corporateProfileRepository: func() *mocks.CorporateProfileRepository {
				repo := &mocks.CorporateProfileRepository{}

				copyCorporateProfile1 := mockCorporateProfile1

				repo.On("GetByUserId", context.Background(), int64(1000)).Return(copyCorporateProfile1, nil)
				repo.On("Update", context.Background(), copyCorporateProfile1).Return(nil)
				return repo
			},
			corporateShareholderRepository: func() *mocks.CorporateShareholderRepository {
				repo := &mocks.CorporateShareholderRepository{}
				return repo
			},
			wantErr: false,
		},
		{
			name: "代表プロフィール登録成功",
			args: args{
				ctx: context.Background(),
				request: udomain.CorporateUserRegisterProfilesRequest{
					User: jwthelper.User{
						Id: 1000,
					},
					RepresentativeProfile: udomain.CorporateUserRepresentativeProfileItem{
						FamilyName:     "ホン",
						FirstName:      "ギルドン",
						FamilyNameKana: "ホン",
						FirstNameKana:  "ギルドン",
						Email:          "gildong@email.com",
						MobilePhone:    "01012345678",
					},
					SignUpStatus: domain.SIGNUP_CORPORATE_PROFILE_REPRESENTATIVE,
				},
			},
			userRepository: func() *mocks.UserRepository {
				repo := &mocks.UserRepository{}

				copyUser1 := mockUser1

				copyUser1.SignupStatus = domain.SIGNUP_CORPORATE_PROFILE_MANAGER
				repo.On("GetById", context.Background(), int64(1000)).Return(copyUser1, nil)

				// SignupStatusは後段階なのでSIGNUP_CORPORATE_PROFILE_REPRESENTATIVEに変更されない。
				repo.On("Update", context.Background(), copyUser1).Return(nil)
				return repo
			},

			corporateProfileRepository: func() *mocks.CorporateProfileRepository {
				repo := &mocks.CorporateProfileRepository{}

				copyCorporateProfile1 := mockCorporateProfile1

				repo.On("GetByUserId", context.Background(), int64(1000)).Return(copyCorporateProfile1, nil)

				copyCorporateProfile1.RepresentativeFamilyName = "ホン"
				copyCorporateProfile1.RepresentativeFirstName = "ギルドン"
				copyCorporateProfile1.RepresentativeFamilyNameKana = "ホン"
				copyCorporateProfile1.RepresentativeFirstNameKana = "ギルドン"
				copyCorporateProfile1.RepresentativeEmail = "gildong@email.com"
				copyCorporateProfile1.RepresentativeMobilePhone = "01012345678"

				repo.On("Update", context.Background(), copyCorporateProfile1).Return(nil)
				return repo
			},
			corporateShareholderRepository: func() *mocks.CorporateShareholderRepository {
				repo := &mocks.CorporateShareholderRepository{}
				return repo
			},
			wantErr: false,
		},
		{
			name: "マネージャープロフィール登録成功",
			args: args{
				ctx: context.Background(),
				request: udomain.CorporateUserRegisterProfilesRequest{
					User: jwthelper.User{
						Id: 1000,
					},
					ManagerProfile: udomain.CorporateUserManagerProfileItem{
						FamilyName:     "ホン",
						FirstName:      "ギルドン",
						FamilyNameKana: "ホン",
						FirstNameKana:  "ギルドン",
						Email:          "gildong@email.com",
						MobilePhone:    "01012345678",
						Department:     "部署名",
						Position:       "職責",
						ZipCode:        "1234567",
						Province:       "都道府県",
						City:           "市群区",
						Address:        "詳細住所",
						Address2:       lo.ToPtr(""),
						Birthday:       "2000-01-01",
					},
					SignUpStatus: domain.SIGNUP_CORPORATE_PROFILE_MANAGER,
				},
			},
			userRepository: func() *mocks.UserRepository {
				repo := &mocks.UserRepository{}

				copyUser1 := mockUser1
				copyUser1.SignupStatus = domain.SIGNUP_CORPORATE_PROFILE_MANAGER

				repo.On("GetById", context.Background(), int64(1000)).Return(copyUser1, nil)
				repo.On("Update", context.Background(), copyUser1).Return(nil)
				return repo
			},

			corporateProfileRepository: func() *mocks.CorporateProfileRepository {
				repo := &mocks.CorporateProfileRepository{}

				copyCorporateProfile1 := mockCorporateProfile1

				repo.On("GetByUserId", context.Background(), int64(1000)).Return(copyCorporateProfile1, nil)

				copyCorporateProfile1.ManagerFamilyName = lo.ToPtr("ホン")
				copyCorporateProfile1.ManagerFirstName = lo.ToPtr("ギルドン")
				copyCorporateProfile1.ManagerFamilyNameKana = lo.ToPtr("ホン")
				copyCorporateProfile1.ManagerFirstNameKana = lo.ToPtr("ギルドン")
				copyCorporateProfile1.ManagerEmail = lo.ToPtr("gildong@email.com")
				copyCorporateProfile1.ManagerMobilePhone = lo.ToPtr("01012345678")
				copyCorporateProfile1.Department = lo.ToPtr("部署名")
				copyCorporateProfile1.Position = lo.ToPtr("職責")
				copyCorporateProfile1.ManagerZipCode = "1234567"
				copyCorporateProfile1.ManagerProvince = "都道府県"
				copyCorporateProfile1.ManagerCity = "市群区"
				copyCorporateProfile1.ManagerAddress = "詳細住所"
				copyCorporateProfile1.ManagerAddress2 = lo.ToPtr("")
				copyCorporateProfile1.ManagerBirthday = mustParseCorporateUserDate("2000-01-01")

				repo.On("Update", context.Background(), copyCorporateProfile1).Return(nil)
				return repo
			},
			corporateShareholderRepository: func() *mocks.CorporateShareholderRepository {
				repo := &mocks.CorporateShareholderRepository{}
				return repo
			},
			wantErr: false,
		},
		{
			name: "実質的支配者プロフィール登録成功",
			args: args{
				ctx: context.Background(),
				request: udomain.CorporateUserRegisterProfilesRequest{
					User: jwthelper.User{
						Id: 1000,
					},
					ShareholderProfiles: []udomain.CorporateUserShareholderProfileItem{
						{
							FamilyName:     "ホン",
							FirstName:      "ギルドン",
							FamilyNameKana: "ホン",
							FirstNameKana:  "ギルドン",
							ZipCode:        "1234567",
							Province:       "都道府県",
							City:           "市群区",
							Address:        "詳細住所",
							Address2:       lo.ToPtr("詳細住所2"),
							Birthday:       "2000-01-01",
						},
					},
					SignUpStatus: domain.SIGNUP_CORPORATE_PROFILE_SHAREHOLDER,
				},
			},
			userRepository: func() *mocks.UserRepository {
				repo := &mocks.UserRepository{}

				copyUser1 := mockUser1
				copyUser1.SignupStatus = domain.SIGNUP_CORPORATE_PROFILE_MANAGER
				repo.On("GetById", context.Background(), int64(1000)).Return(copyUser1, nil)

				copyUser1.SignupStatus = domain.SIGNUP_CORPORATE_PROFILE_SHAREHOLDER
				repo.On("Update", context.Background(), copyUser1).Return(nil)
				return repo
			},

			corporateProfileRepository: func() *mocks.CorporateProfileRepository {
				repo := &mocks.CorporateProfileRepository{}

				copyCorporateProfile1 := mockCorporateProfile1

				repo.On("GetByUserId", context.Background(), int64(1000)).Return(copyCorporateProfile1, nil)
				repo.On("Update", context.Background(), copyCorporateProfile1).Return(nil)
				return repo
			},

			corporateShareholderRepository: func() *mocks.CorporateShareholderRepository {
				repo := &mocks.CorporateShareholderRepository{}
				repo.On("DeleteByUserId", context.Background(), int64(1000)).Return(nil)
				repo.On("Create", context.Background(), domain.CorporateShareholder{
					UserId:         1000,
					FamilyName:     "ホン",
					FirstName:      "ギルドン",
					FamilyNameKana: "ホン",
					FirstNameKana:  "ギルドン",
					ZipCode:        "1234567",
					Province:       "都道府県",
					City:           "市群区",
					Address:        "詳細住所",
					Address2:       "詳細住所2",
					Birthday:       mustParseCorporateUserDate("2000-01-01"),
				}).Return(int64(1), nil)

				return repo
			},
			wantErr: false,
		},
		{
			name: "ユーザー状態より前の状態のプロフィール更新を試行(OTHER->MANAGER)",
			args: args{
				ctx: context.Background(),
				request: udomain.CorporateUserRegisterProfilesRequest{
					User: jwthelper.User{
						Id: 1000,
					},
					SignUpStatus: domain.SIGNUP_CORPORATE_PROFILE_MANAGER,
				},
			},
			userRepository: func() *mocks.UserRepository {
				repo := &mocks.UserRepository{}

				copyUser1 := mockUser1
				copyUser1.SignupStatus = domain.SIGNUP_CORPORATE_PROFILE_OTHER
				repo.On("GetById", context.Background(), int64(1000)).Return(copyUser1, nil)
				return repo
			},

			corporateProfileRepository: func() *mocks.CorporateProfileRepository {
				repo := &mocks.CorporateProfileRepository{}
				return repo
			},
			corporateShareholderRepository: func() *mocks.CorporateShareholderRepository {
				repo := &mocks.CorporateShareholderRepository{}
				return repo
			},
			wantErr: true,
		},
		{
			name: "法人出金口座登録成功",
			args: args{
				ctx: context.Background(),
				request: udomain.CorporateUserRegisterProfilesRequest{
					User: jwthelper.User{
						Id: 1000,
					},
					SignUpStatus: domain.SIGNUP_ACCOUNT,
					WithdrawalAccount: udomain.CorporateUserRegisterWithdrawalAccountItem{
						BankName:    "銀行名",
						BranchName:  "支店名",
						BankCode:    lo.ToPtr("123"),
						BranchCode:  lo.ToPtr("456"),
						DepositType: domain.ORDINARY_ACCOUNT,
						BankAccount: "1234567",
						Depositor:   "ホン・ギルドン",
					},
				},
			},
			userRepository: func() *mocks.UserRepository {
				repo := &mocks.UserRepository{}

				copyUser1 := mockUser1
				repo.On("GetById", context.Background(), int64(1000)).Return(copyUser1, nil)

				copyUser1.BankName = "銀行名"
				copyUser1.BranchName = "支店名"
				copyUser1.BankCode = lo.ToPtr("123")
				copyUser1.BranchCode = lo.ToPtr("456")
				copyUser1.DepositType = domain.ORDINARY_ACCOUNT
				copyUser1.BankAccount = "1234567"
				copyUser1.Depositor = "ホン・ギルドン"
				copyUser1.SignupStatus = domain.SIGNUP_ACCOUNT

				repo.On("Update", context.Background(), copyUser1).Return(nil)
				return repo
			},
			corporateProfileRepository: func() *mocks.CorporateProfileRepository {
				repo := &mocks.CorporateProfileRepository{}

				copyCorporateProfile1 := mockCorporateProfile1

				repo.On("GetByUserId", context.Background(), int64(1000)).Return(copyCorporateProfile1, nil)
				repo.On("Update", context.Background(), copyCorporateProfile1).Return(nil)
				return repo
			},
			corporateShareholderRepository: func() *mocks.CorporateShareholderRepository {
				repo := &mocks.CorporateShareholderRepository{}
				return repo
			},
			wantErr: false,
		},
		{
			name: "法人投資適合性確認成功",
			args: args{
				ctx: context.Background(),
				request: udomain.CorporateUserRegisterProfilesRequest{
					User: jwthelper.User{
						Id: mockUser1.Id,
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
				},
			},
			userRepository: func() *mocks.UserRepository {
				repo := &mocks.UserRepository{}

				copyUser1 := mockUser1
				repo.On("GetById", context.Background(), copyUser1.Id).Return(copyUser1, nil)

				copyUser1.InvestmentPurpose = domain.VARIANCE_PORTFOLIO
				copyUser1.HasExperience = lo.ToPtr(true)
				copyUser1.IsSurplusFund = lo.ToPtr(true)
				copyUser1.SignupStatus = domain.SIGNUP_INVESTMENT_SUITABILITY

				repo.On("Update", context.Background(), copyUser1).Return(nil)
				return repo
			},

			corporateProfileRepository: func() *mocks.CorporateProfileRepository {
				repo := &mocks.CorporateProfileRepository{}

				copyCorporateProfile1 := mockCorporateProfile1
				repo.On("GetByUserId", mock.Anything, int64(1000)).Return(copyCorporateProfile1, nil)

				copyCorporateProfile1.EmployeeCount = "10"
				copyCorporateProfile1.CapitalAmount = "100000"
				copyCorporateProfile1.Sales = "50000"

				repo.On("Update", context.Background(), copyCorporateProfile1).Return(nil)
				return repo
			},
			corporateShareholderRepository: func() *mocks.CorporateShareholderRepository {
				repo := &mocks.CorporateShareholderRepository{}
				return repo
			},
			wantErr: false,
		},
		{
			name: "約款同意成功",
			args: args{
				ctx: context.Background(),
				request: udomain.CorporateUserRegisterProfilesRequest{
					User: jwthelper.User{
						Id: mockUser1.Id,
					},
					SignUpStatus: domain.SIGNUP_AGREE,
				},
			},
			userRepository: func() *mocks.UserRepository {
				repo := &mocks.UserRepository{}

				copyUser1 := mockUser1
				repo.On("GetById", context.Background(), copyUser1.Id).Return(copyUser1, nil)

				copyUser1.SignupStatus = domain.SIGNUP_AGREE
				repo.On("Update", context.Background(), copyUser1).Return(nil)
				return repo
			},
			corporateProfileRepository: func() *mocks.CorporateProfileRepository {
				repo := &mocks.CorporateProfileRepository{}

				copyCorporateProfile1 := mockCorporateProfile1

				repo.On("GetByUserId", context.Background(), int64(1000)).Return(copyCorporateProfile1, nil)
				repo.On("Update", context.Background(), copyCorporateProfile1).Return(nil)
				return repo
			},
			corporateShareholderRepository: func() *mocks.CorporateShareholderRepository {
				repo := &mocks.CorporateShareholderRepository{}
				return repo
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := corporateUserUsecase{
				userRepository:                 tt.userRepository(),
				corporateProfileRepository:     tt.corporateProfileRepository(),
				corporateShareholderRepository: tt.corporateShareholderRepository(),
			}
			err := usecase.RegisterProfiles(tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.wantErr, err != nil, "RegisterProfiles() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_corporateUserUsecase_VerifyWithKyc(t *testing.T) {
	type args struct {
		ctx     context.Context
		request udomain.CorporateUserVerifyWithKycRequest
	}
	tests := []struct {
		name                       string
		args                       args
		cfg                        *config.Config
		userRepository             func() *mocks.UserRepository
		corporateProfileRepository func() *mocks.CorporateProfileRepository
		s3pkg                      func() *mocks.S3
		userFileRepository         func() *mocks.UserFileRepository
		wantErr                    bool
	}{

		{
			name: "全て同名のファイルなら登録対象が無いので弾く",
			args: args{
				ctx: context.Background(),
				request: udomain.CorporateUserVerifyWithKycRequest{
					JWTUser: jwthelper.User{
						Id: 1000,
					},
					// 二重送信などで全て既存と同名だと、追加対象が無くなる。
					// xorm は空スライスの Insert で ErrNoElementsOnSlice を返すため、
					// CreateMany へ到達する前に弾く必要がある。
					CorporateUserFileUrlItem: udomain.CorporateUserFileUrlItem{
						CertifiedCopyFileUrl: lo.ToPtr("https://cdn.dev.energyshares.jp/user_file/2024-05-09/kyc_certified_copy.jpeg"),
					},
				},
			},
			userRepository: func() *mocks.UserRepository {
				repo := &mocks.UserRepository{}
				copyUser1 := mockUser1
				copyUser1.SignupStatus = domain.SIGNUP_AGREE
				repo.On("GetById", context.Background(), int64(1000)).Return(copyUser1, nil)
				repo.On("Update", context.Background(), mock.Anything).Return(nil)
				return repo
			},
			s3pkg: func() *mocks.S3 {
				client := &mocks.S3{}
				client.On("GetHeadObject", s3utils.BUCKET_PRIVATE_NAME, mock.Anything).Return(mockS3ObjectInfo1, nil)
				return client
			},
			userFileRepository: func() *mocks.UserFileRepository {
				repo := &mocks.UserFileRepository{}
				repo.On("ExistsByFileName", context.Background(), "kyc_certified_copy").Return(true, nil)
				repo.On("CreateMany", context.Background(), mock.Anything).Return(nil)
				return repo
			},
			corporateProfileRepository: func() *mocks.CorporateProfileRepository {
				repo := &mocks.CorporateProfileRepository{}
				repo.On("GetByUserId", context.Background(), int64(1000)).Return(mockCorporateProfile1, nil)
				return repo
			},
			wantErr: true,
		},
		{
			name: "KYCファイルアップロード成功",
			args: args{
				ctx: context.Background(),
				request: udomain.CorporateUserVerifyWithKycRequest{
					JWTUser: jwthelper.User{
						Id: 1000,
					},
					// file_name はテーブル全体で一意なので、書類ごとに別ファイルになる
					CorporateUserFileUrlItem: udomain.CorporateUserFileUrlItem{
						CertifiedCopyFileUrl:        lo.ToPtr("https://cdn.dev.energyshares.jp/user_file/2024-05-09/kyc_certified_copy.jpeg"),
						Identification1FrontFileUrl: lo.ToPtr("https://cdn.dev.energyshares.jp/user_file/2024-05-09/kyc_identification1_front.jpeg"),
						Identification1BackFileUrl:  lo.ToPtr("https://cdn.dev.energyshares.jp/user_file/2024-05-09/kyc_identification1_back.jpeg"),
						Identification2FrontFileUrl: lo.ToPtr("https://cdn.dev.energyshares.jp/user_file/2024-05-09/kyc_identification2_front.jpeg"),
						Identification2BackFileUrl:  lo.ToPtr("https://cdn.dev.energyshares.jp/user_file/2024-05-09/kyc_identification2_back.jpeg"),
						PoaFileUrl:                  lo.ToPtr("https://cdn.dev.energyshares.jp/user_file/2024-05-09/kyc_poa.jpeg"),
					},
				},
			},
			userRepository: func() *mocks.UserRepository {
				repo := &mocks.UserRepository{}

				copyUser1 := mockUser1
				copyUser1.SignupStatus = domain.SIGNUP_AGREE
				repo.On("GetById", context.Background(), int64(1000)).Return(copyUser1, nil)

				copyUser1.SignupStatus = domain.SIGNUP_COMPLETE
				copyUser1.AuditStatus = domain.WAIT_AUDIT
				repo.On("Update", context.Background(), copyUser1).Return(nil)
				return repo
			},
			s3pkg: func() *mocks.S3 {
				client := &mocks.S3{}
				client.On("GetHeadObject", s3utils.BUCKET_PRIVATE_NAME, mock.Anything).Return(mockS3ObjectInfo1, nil)

				return client
			},
			userFileRepository: func() *mocks.UserFileRepository {
				repo := &mocks.UserFileRepository{}

				// 同名の既存ファイルは無い（初回提出）
				repo.On("ExistsByFileName", context.Background(), mock.Anything).Return(false, nil)
				repo.On("CreateMany", context.Background(), mock.MatchedBy(func(userFiles []domain.UserFile) bool {
					ignoreTypes := cmpopts.IgnoreTypes(jsonutils.JSONTime{})
					ignoreFields := cmpopts.IgnoreFields(domain.UserFile{}, "Id")

					return cmp.Equal(mockUserFilesForVerifyWithKyc, userFiles, ignoreTypes, ignoreFields)
				})).Return(nil)
				return repo
			},
			corporateProfileRepository: func() *mocks.CorporateProfileRepository {
				repo := &mocks.CorporateProfileRepository{}
				repo.On("GetByUserId", context.Background(), int64(1000)).Return(mockCorporateProfile1, nil)
				return repo
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := corporateUserUsecase{
				cfg: &config.Config{
					CommonConfig: config.CommonConfig{
						ServiceConfig: struct {
							Mail serviceutils.MailConfig
							Url  serviceutils.UrlConfig
						}{
							Url: serviceutils.UrlConfig{
								UserServiceDomain: "https://user.dev.energyshares.com",
							},
							Mail: serviceutils.MailConfig{
								Info: "info@energyshares.jp",
							},
						},
					},
				},
				fileHelper:                 filehelper.NewFileHelper(),
				logger:                     factory.NewLoggerFactory(),
				userRepository:             tt.userRepository(),
				s3pkg:                      tt.s3pkg(),
				userFileRepository:         tt.userFileRepository(),
				corporateProfileRepository: tt.corporateProfileRepository(),
			}
			err := usecase.VerifyWithKyc(tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.wantErr, err != nil, "VerifyWithKyc() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_corporateUserUsecase_GetOne(t *testing.T) {
	type args struct {
		ctx     context.Context
		request udomain.CorporateUserGetOneRequest
	}
	tests := []struct {
		name                           string
		args                           args
		userRepository                 func() *mocks.UserRepository
		corporateProfileRepository     func() *mocks.CorporateProfileRepository
		corporateShareholderRepository func() *mocks.CorporateShareholderRepository
		userFileService                func() *mocks.UserFileService
		userVirtualAccountRepository   func() *mocks.UserVirtualAccountRepository
		want                           udomain.CorporateUserGetOneResponse
		wantErr                        bool
	}{
		{
			name: "ユーザー情報照会成功",
			args: args{
				ctx: context.Background(),
				request: udomain.CorporateUserGetOneRequest{
					User: jwthelper.User{
						Id: 1001,
					},
				},
			},
			userRepository: func() *mocks.UserRepository {
				repo := &mocks.UserRepository{}
				copyUser := mockUser3
				copyUser.SignupStatus = domain.SIGNUP_COMPLETE
				repo.On("GetById", context.Background(), int64(1001)).Return(copyUser, nil)
				return repo
			},

			corporateProfileRepository: func() *mocks.CorporateProfileRepository {
				repo := &mocks.CorporateProfileRepository{}
				repo.On("GetByUserId", context.Background(), int64(1001)).Return(mockCorporateProfile2, nil)
				return repo
			},

			corporateShareholderRepository: func() *mocks.CorporateShareholderRepository {
				repo := &mocks.CorporateShareholderRepository{}
				repo.On("GetAll", context.Background(), mock.Anything).Return(int64(0), []domain.CorporateShareholder{mockCorporateShareholder1}, nil)
				return repo
			},

			userFileService: func() *mocks.UserFileService {
				svc := &mocks.UserFileService{}
				svc.On("GetCorporateUserIdentificationFiles", context.Background(), int64(1001)).
					Return(cdomain.CorporateUserIdentificationFiles{
						Front1:        &mockUserFile1,
						Back1:         &mockCertifiedCopyFile,
						Front2:        &mockUserFile1,
						Back2:         &mockCertifiedCopyFile,
						CertifiedCopy: &mockUserFile1,
						POA:           &mockCertifiedCopyFile,
					}, nil)

				return svc
			},

			userVirtualAccountRepository: func() *mocks.UserVirtualAccountRepository {
				repo := &mocks.UserVirtualAccountRepository{}
				repo.On("GetByUserIdAndLinkedStatus", context.Background(), int64(1001), domain.EQUITY, domain.ON_VA_ACCOUNT).
					Return(domain.UserVirtualAccount{
						VaBankName:       domain.GMO_AOZORA_BANK_NAME,
						VaBranchName:     "あじさい支店",
						VaAccountNumber:  "5194080",
						VaHolderNameKana: "カ）ハコブネカ）ハコブネ",
					}, nil)
				return repo
			},

			want: udomain.CorporateUserGetOneResponse{
				UserId:            1001,
				Email:             "gildong@email.com",
				IsNewsletter:      true,
				InvestmentPurpose: string(domain.RAISING_FUTURE_MONEY),
				CorporateUserBasicProfileItem: udomain.CorporateUserBasicProfileItem{
					CorpNo:   "1234567890123",
					CorpName: "イルミナリアン",
					CorpType: "IT企業",
				},
				CorporateUserOtherProfileItem: udomain.CorporateUserOtherProfileItem{
					ZipCode:     "1234567",
					Province:    "都道府県",
					City:        "市",
					Address:     "詳細住所",
					Address2:    lo.ToPtr("詳細住所2"),
					MobilePhone: "01012345678",
				},
				Representative: udomain.CorporateUserRepresentativeProfileItem{
					FamilyName:     "ホン",
					FirstName:      "ギルドン",
					FamilyNameKana: "ホン",
					FirstNameKana:  "ギルドン",
					Email:          "gildong@email.com",
					MobilePhone:    "01012345678",
				},
				Manager: udomain.CorporateUserManagerProfileItem{
					FamilyName:     "ホン",
					FirstName:      "ギルドン",
					FamilyNameKana: "ホン",
					FirstNameKana:  "ギルドン",
					Email:          "gildong@email.com",
					MobilePhone:    "01012345678",
					Department:     "部署",
					Position:       "職責",
					ZipCode:        "1234567",
					Province:       "都道府県",
					City:           "市",
					Address:        "詳細住所",
					Address2:       lo.ToPtr("詳細住所2"),
					Birthday:       "2000-01-01",
				},
				Shareholders: []udomain.CorporateUserShareholderProfileItem{
					{
						FamilyName:     "ホン",
						FirstName:      "ギルドン",
						FamilyNameKana: "ホン",
						FirstNameKana:  "ギルドン",
						ZipCode:        "1234567",
						Province:       "都道府県",
						City:           "市",
						Address:        "詳細住所",
						Address2:       lo.ToPtr("詳細住所2"),
						Birthday:       "2000-01-01",
						Uid:            "0001",
					},
				},
				WithdrawalAccount: udomain.CorporateUserRegisterWithdrawalAccountItem{
					BankName:    mockUser3.BankName,
					BranchName:  mockUser3.BranchName,
					BankCode:    mockUser3.BankCode,
					BranchCode:  mockUser3.BranchCode,
					DepositType: mockUser3.DepositType,
					BankAccount: mockUser3.BankAccount,
					Depositor:   mockUser3.Depositor,
				},
				InvestmentSuitability: udomain.CorporateUserConfirmInvestmentSuitabilityItem{
					EmployeeCount:     mockCorporateProfile2.EmployeeCount,
					CapitalAmount:     mockCorporateProfile2.CapitalAmount,
					Sales:             mockCorporateProfile2.Sales,
					InvestmentPurpose: mockUser3.InvestmentPurpose,
					HasExperience:     mockUser3.HasExperience,
					IsSurplusFund:     mockUser3.IsSurplusFund,
					InvestorType:      mockUser3.InvestorType,
				},
				VirtualAccount: udomain.CorporateUserVirtualAccountItem{
					BankName:       domain.GMO_AOZORA_BANK_NAME,
					BranchName:     "あじさい支店",
					AccountNumber:  "5194080",
					HolderNameKana: "カ）ハコブネカ）ハコブネ",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := corporateUserUsecase{
				userRepository:                 tt.userRepository(),
				corporateProfileRepository:     tt.corporateProfileRepository(),
				corporateShareholderRepository: tt.corporateShareholderRepository(),
				userFileService:                tt.userFileService(),
				userVirtualAccountRepository:   tt.userVirtualAccountRepository(),
			}
			got, err := usecase.GetOne(tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.wantErr, err != nil, "GetOne() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want, got, "GetOne() = %v, want %v", got, tt.want)
		})
	}
}

func Test_corporateUserUsecase_GetSignUpStatus_Part1(t *testing.T) {
	type args struct {
		ctx     context.Context
		request udomain.CorporateUserGetSignUpStatusRequest
	}
	tests := []struct {
		name                           string
		args                           args
		userRepository                 func() *mocks.UserRepository
		corporateProfileRepository     func() *mocks.CorporateProfileRepository
		corporateShareholderRepository func() *mocks.CorporateShareholderRepository
		want                           udomain.CorporateUserGetSignUpStatusResponse
		wantErr                        bool
	}{
		{
			name: "会員登録状態照会成功 (Statusを別途要求しなかった時)- SIGNUP_NONE",
			args: args{
				ctx: context.Background(),
				request: udomain.CorporateUserGetSignUpStatusRequest{
					User: jwthelper.User{
						Id: 1001,
					},
				},
			},
			userRepository: func() *mocks.UserRepository {
				repo := &mocks.UserRepository{}

				copyUser := mockUser2
				copyUser.SignupStatus = domain.SIGNUP_NONE
				repo.On("GetById", context.Background(), int64(1001)).Return(copyUser, nil)
				return repo
			},
			corporateProfileRepository: func() *mocks.CorporateProfileRepository {
				repo := &mocks.CorporateProfileRepository{}
				repo.On("GetByUserId", context.Background(), int64(1001)).Return(mockCorporateProfile2, nil)
				return repo
			},
			corporateShareholderRepository: func() *mocks.CorporateShareholderRepository {
				repo := &mocks.CorporateShareholderRepository{}
				return repo
			},

			want: udomain.CorporateUserGetSignUpStatusResponse{
				UserId:     1001,
				Email:      "gildong@email.com",
				PrevStatus: domain.SIGNUP_NONE,
				CurStatus:  domain.SIGNUP_NONE,
				NextStatus: domain.SIGNUP_CORPORATE_PROFILE_BASIC,
			},
			wantErr: false,
		},
		{
			name: "会員登録状態照会成功 - SIGNUP_BASIC",
			args: args{
				ctx: context.Background(),
				request: udomain.CorporateUserGetSignUpStatusRequest{
					User: jwthelper.User{
						Id: 1001,
					},
					Status: domain.SIGNUP_CORPORATE_PROFILE_BASIC,
				},
			},
			userRepository: func() *mocks.UserRepository {
				repo := &mocks.UserRepository{}

				copyUser := mockUser2
				repo.On("GetById", context.Background(), int64(1001)).Return(copyUser, nil)
				return repo
			},
			corporateProfileRepository: func() *mocks.CorporateProfileRepository {
				repo := &mocks.CorporateProfileRepository{}
				repo.On("GetByUserId", context.Background(), int64(1001)).Return(mockCorporateProfile2, nil)
				return repo
			},
			corporateShareholderRepository: func() *mocks.CorporateShareholderRepository {
				repo := &mocks.CorporateShareholderRepository{}
				return repo
			},

			want: udomain.CorporateUserGetSignUpStatusResponse{
				UserId:     1001,
				Email:      "gildong@email.com",
				PrevStatus: domain.SIGNUP_NONE,
				CurStatus:  domain.SIGNUP_CORPORATE_PROFILE_BASIC,
				NextStatus: domain.SIGNUP_CORPORATE_PROFILE_OTHER,
				BasicProfile: &udomain.CorporateUserBasicProfileItem{
					CorpNo:   "1234567890123",
					CorpName: "イルミナリアン",
					CorpType: "IT企業",
				},
			},
			wantErr: false,
		},
		{
			name: "会員登録状態照会成功 - SIGNUP_OTHER",
			args: args{
				ctx: context.Background(),
				request: udomain.CorporateUserGetSignUpStatusRequest{
					User: jwthelper.User{
						Id: 1001,
					},
					Status: domain.SIGNUP_CORPORATE_PROFILE_OTHER,
				},
			},
			userRepository: func() *mocks.UserRepository {
				repo := &mocks.UserRepository{}

				copyUser := mockUser2
				repo.On("GetById", context.Background(), int64(1001)).Return(copyUser, nil)
				return repo
			},
			corporateProfileRepository: func() *mocks.CorporateProfileRepository {
				repo := &mocks.CorporateProfileRepository{}
				repo.On("GetByUserId", context.Background(), int64(1001)).Return(mockCorporateProfile2, nil)
				return repo
			},
			corporateShareholderRepository: func() *mocks.CorporateShareholderRepository {
				repo := &mocks.CorporateShareholderRepository{}
				return repo
			},

			want: udomain.CorporateUserGetSignUpStatusResponse{
				UserId:     1001,
				Email:      "gildong@email.com",
				PrevStatus: domain.SIGNUP_CORPORATE_PROFILE_BASIC,
				CurStatus:  domain.SIGNUP_CORPORATE_PROFILE_OTHER,
				NextStatus: domain.SIGNUP_CORPORATE_PROFILE_REPRESENTATIVE,
				OtherProfile: &udomain.CorporateUserOtherProfileItem{
					ZipCode:     "1234567",
					Province:    "都道府県",
					City:        "市",
					Address:     "詳細住所",
					Address2:    lo.ToPtr("詳細住所2"),
					MobilePhone: "01012345678",
				},
			},
			wantErr: false,
		},
		{
			name: "会員登録状態照会成功 - SIGN_UP_REPRESENTATIVE",
			args: args{
				ctx: context.Background(),
				request: udomain.CorporateUserGetSignUpStatusRequest{
					User: jwthelper.User{
						Id: 1001,
					},
					Status: domain.SIGNUP_CORPORATE_PROFILE_REPRESENTATIVE,
				},
			},
			userRepository: func() *mocks.UserRepository {
				repo := &mocks.UserRepository{}

				copyUser := mockUser2
				repo.On("GetById", context.Background(), int64(1001)).Return(copyUser, nil)
				return repo
			},
			corporateProfileRepository: func() *mocks.CorporateProfileRepository {
				repo := &mocks.CorporateProfileRepository{}
				repo.On("GetByUserId", context.Background(), int64(1001)).Return(mockCorporateProfile2, nil)
				return repo
			},
			corporateShareholderRepository: func() *mocks.CorporateShareholderRepository {
				repo := &mocks.CorporateShareholderRepository{}
				return repo
			},

			want: udomain.CorporateUserGetSignUpStatusResponse{
				UserId:     1001,
				Email:      "gildong@email.com",
				PrevStatus: domain.SIGNUP_CORPORATE_PROFILE_OTHER,
				CurStatus:  domain.SIGNUP_CORPORATE_PROFILE_REPRESENTATIVE,
				NextStatus: domain.SIGNUP_CORPORATE_PROFILE_MANAGER,
				RepresentativeProfile: &udomain.CorporateUserRepresentativeProfileItem{
					FamilyName:     "ホン",
					FirstName:      "ギルドン",
					FamilyNameKana: "ホン",
					FirstNameKana:  "ギルドン",
					Email:          "gildong@email.com",
					MobilePhone:    "01012345678",
				},
			},
			wantErr: false,
		},
		{
			name: "会員登録状態照会成功 - SIGN_UP_MANAGER",
			args: args{
				ctx: context.Background(),
				request: udomain.CorporateUserGetSignUpStatusRequest{
					User: jwthelper.User{
						Id: 1001,
					},
					Status: domain.SIGNUP_CORPORATE_PROFILE_MANAGER,
				},
			},
			userRepository: func() *mocks.UserRepository {
				repo := &mocks.UserRepository{}

				copyUser := mockUser2
				repo.On("GetById", context.Background(), int64(1001)).Return(copyUser, nil)
				return repo
			},
			corporateProfileRepository: func() *mocks.CorporateProfileRepository {
				repo := &mocks.CorporateProfileRepository{}
				repo.On("GetByUserId", context.Background(), int64(1001)).Return(mockCorporateProfile2, nil)
				return repo
			},
			corporateShareholderRepository: func() *mocks.CorporateShareholderRepository {
				repo := &mocks.CorporateShareholderRepository{}
				return repo
			},

			want: udomain.CorporateUserGetSignUpStatusResponse{
				UserId:     1001,
				Email:      "gildong@email.com",
				PrevStatus: domain.SIGNUP_CORPORATE_PROFILE_REPRESENTATIVE,
				CurStatus:  domain.SIGNUP_CORPORATE_PROFILE_MANAGER,
				NextStatus: domain.SIGNUP_CORPORATE_PROFILE_SHAREHOLDER,
				ManagerProfile: &udomain.CorporateUserManagerProfileItem{
					FamilyName:     "ホン",
					FirstName:      "ギルドン",
					FamilyNameKana: "ホン",
					FirstNameKana:  "ギルドン",
					Email:          "gildong@email.com",
					MobilePhone:    "01012345678",
					Department:     "部署",
					Position:       "職責",
					ZipCode:        "1234567",
					Province:       "都道府県",
					City:           "市",
					Address:        "詳細住所",
					Address2:       lo.ToPtr("詳細住所2"),
					Birthday:       "2000-01-01",
				},
			},
			wantErr: false,
		},
		{
			name: "会員登録状態照会成功 - SIGNUP_CORPORATE_PROFILE_SHAREHOLDER",
			args: args{
				ctx: context.Background(),
				request: udomain.CorporateUserGetSignUpStatusRequest{
					User: jwthelper.User{
						Id: 1001,
					},
					Status: domain.SIGNUP_CORPORATE_PROFILE_SHAREHOLDER,
				},
			},
			userRepository: func() *mocks.UserRepository {
				repo := &mocks.UserRepository{}

				copyUser := mockUser2
				repo.On("GetById", context.Background(), int64(1001)).Return(copyUser, nil)
				return repo
			},
			corporateProfileRepository: func() *mocks.CorporateProfileRepository {
				repo := &mocks.CorporateProfileRepository{}
				repo.On("GetByUserId", context.Background(), int64(1001)).Return(mockCorporateProfile2, nil)
				return repo
			},
			corporateShareholderRepository: func() *mocks.CorporateShareholderRepository {
				repo := &mocks.CorporateShareholderRepository{}
				repo.On("GetAll", context.Background(),
					mock.MatchedBy(func(field func(*cdomain.CorporateShareholderQueryOptionFields)) bool {
						opt := cdomain.CorporateShareholderQueryOptionFunc()
						field(&opt)
						return lo.FromPtr(opt.UserId()) == int64(1001)
					}),
				).Return(int64(0), []domain.CorporateShareholder{mockCorporateShareholder1}, nil)
				return repo
			},

			want: udomain.CorporateUserGetSignUpStatusResponse{
				UserId:     1001,
				Email:      "gildong@email.com",
				PrevStatus: domain.SIGNUP_CORPORATE_PROFILE_MANAGER,
				CurStatus:  domain.SIGNUP_CORPORATE_PROFILE_SHAREHOLDER,
				NextStatus: domain.SIGNUP_ACCOUNT,
				ShareholderProfiles: []udomain.CorporateUserShareholderProfileItem{
					{
						FamilyName:     "ホン",
						FirstName:      "ギルドン",
						FamilyNameKana: "ホン",
						FirstNameKana:  "ギルドン",
						ZipCode:        "1234567",
						Province:       "都道府県",
						City:           "市",
						Address:        "詳細住所",
						Address2:       lo.ToPtr("詳細住所2"),
						Birthday:       "2000-01-01",
						Uid:            "0001",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := corporateUserUsecase{
				userRepository:                 tt.userRepository(),
				corporateProfileRepository:     tt.corporateProfileRepository(),
				corporateShareholderRepository: tt.corporateShareholderRepository(),
			}
			got, err := usecase.GetSignUpStatus(tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.wantErr, err != nil, "GetSignUpStatus() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want, got, "GetSignUpStatus() = %v, want %v", got, tt.want)
		})
	}
}

// Test_corporateUserUsecase_GetSignUpStatus_Part2 - 会員登録状態照会テスト
// 不要な依存性注入を減らすのは良いことではないかと思い分離した
// SIGNUP_ACCOUNT, SIGNUP_INVESTMENT_SUITABILITY, SIGNUP_AGREE, SIGNUP_KYC
func Test_corporateUserUsecase_GetSignUpStatus_Part2(t *testing.T) {
	type args struct {
		ctx     context.Context
		request udomain.CorporateUserGetSignUpStatusRequest
	}
	tests := []struct {
		name                       string
		args                       args
		userRepository             func() *mocks.UserRepository
		corporateProfileRepository func() *mocks.CorporateProfileRepository
		userFileService            func() *mocks.UserFileService
		want                       udomain.CorporateUserGetSignUpStatusResponse
		wantErr                    bool
	}{
		{
			name: "会員登録状態照会成功 - SIGNUP_ACCOUNT",
			args: args{
				ctx: context.Background(),
				request: udomain.CorporateUserGetSignUpStatusRequest{
					User: jwthelper.User{
						Id: 1001,
					},
					Status: domain.SIGNUP_ACCOUNT,
				},
			},
			userRepository: func() *mocks.UserRepository {
				repo := &mocks.UserRepository{}

				copyUser := mockUser3
				repo.On("GetById", context.Background(), int64(1001)).Return(copyUser, nil)
				return repo
			},
			corporateProfileRepository: func() *mocks.CorporateProfileRepository {
				repo := &mocks.CorporateProfileRepository{}
				repo.On("GetByUserId", context.Background(), int64(1001)).Return(mockCorporateProfile2, nil)
				return repo
			},
			userFileService: func() *mocks.UserFileService {
				return &mocks.UserFileService{}
			},

			want: udomain.CorporateUserGetSignUpStatusResponse{
				UserId:     1001,
				Email:      "gildong@email.com",
				PrevStatus: domain.SIGNUP_CORPORATE_PROFILE_SHAREHOLDER,
				CurStatus:  domain.SIGNUP_ACCOUNT,
				NextStatus: domain.SIGNUP_INVESTMENT_SUITABILITY,
				WithdrawalAccount: &udomain.CorporateUserRegisterWithdrawalAccountItem{
					BankName:    mockUser3.BankName,
					BranchName:  mockUser3.BranchName,
					BankCode:    mockUser3.BankCode,
					BranchCode:  mockUser3.BranchCode,
					DepositType: mockUser3.DepositType,
					BankAccount: mockUser3.BankAccount,
					Depositor:   mockUser3.Depositor,
				},
			},
			wantErr: false,
		},
		{
			name: "会員登録状態照会成功 - SIGNUP_INVESTMENT_SUITABILITY",
			args: args{
				ctx: context.Background(),
				request: udomain.CorporateUserGetSignUpStatusRequest{
					User: jwthelper.User{
						Id: 1001,
					},
					Status: domain.SIGNUP_INVESTMENT_SUITABILITY,
				},
			},
			userRepository: func() *mocks.UserRepository {
				repo := &mocks.UserRepository{}

				copyUser := mockUser3
				repo.On("GetById", context.Background(), int64(1001)).Return(copyUser, nil)
				return repo
			},
			corporateProfileRepository: func() *mocks.CorporateProfileRepository {
				repo := &mocks.CorporateProfileRepository{}
				repo.On("GetByUserId", context.Background(), int64(1001)).Return(mockCorporateProfile3, nil)
				return repo
			},
			userFileService: func() *mocks.UserFileService {
				return &mocks.UserFileService{}
			},

			want: udomain.CorporateUserGetSignUpStatusResponse{
				UserId:     1001,
				Email:      "gildong@email.com",
				PrevStatus: domain.SIGNUP_ACCOUNT,
				CurStatus:  domain.SIGNUP_INVESTMENT_SUITABILITY,
				NextStatus: domain.SIGNUP_AGREE,
				InvestmentSuitability: &udomain.CorporateUserConfirmInvestmentSuitabilityItem{
					EmployeeCount:     mockCorporateProfile3.EmployeeCount,
					CapitalAmount:     mockCorporateProfile3.CapitalAmount,
					Sales:             mockCorporateProfile3.Sales,
					InvestmentPurpose: mockUser3.InvestmentPurpose,
					HasExperience:     mockUser3.HasExperience,
					IsSurplusFund:     mockUser3.IsSurplusFund,
					InvestorType:      mockUser3.InvestorType,
				},
			},
			wantErr: false,
		},
		{
			name: "会員登録状態照会成功 - SIGNUP_AGREE",
			args: args{
				ctx: context.Background(),
				request: udomain.CorporateUserGetSignUpStatusRequest{
					User: jwthelper.User{
						Id: 1001,
					},
					Status: domain.SIGNUP_AGREE,
				},
			},
			userRepository: func() *mocks.UserRepository {
				repo := &mocks.UserRepository{}

				copyUser := mockUser3
				repo.On("GetById", context.Background(), int64(1001)).Return(copyUser, nil)
				return repo
			},
			corporateProfileRepository: func() *mocks.CorporateProfileRepository {
				repo := &mocks.CorporateProfileRepository{}
				repo.On("GetByUserId", context.Background(), int64(1001)).Return(mockCorporateProfile3, nil)
				return repo
			},
			userFileService: func() *mocks.UserFileService {
				return &mocks.UserFileService{}
			},

			want: udomain.CorporateUserGetSignUpStatusResponse{
				UserId:     1001,
				Email:      "gildong@email.com",
				PrevStatus: domain.SIGNUP_INVESTMENT_SUITABILITY,
				CurStatus:  domain.SIGNUP_AGREE,
				NextStatus: domain.SIGNUP_KYC,
			},
			wantErr: false,
		},
		{
			name: "会員登録状態照会成功 - SIGNUP_COMPLETE",
			args: args{
				ctx: context.Background(),
				request: udomain.CorporateUserGetSignUpStatusRequest{
					User: jwthelper.User{
						Id: 1001,
					},
					Status: domain.SIGNUP_COMPLETE,
				},
			},
			userRepository: func() *mocks.UserRepository {
				repo := &mocks.UserRepository{}

				copyUser := mockUser3
				repo.On("GetById", context.Background(), int64(1001)).Return(copyUser, nil)
				return repo
			},
			corporateProfileRepository: func() *mocks.CorporateProfileRepository {
				repo := &mocks.CorporateProfileRepository{}
				repo.On("GetByUserId", context.Background(), int64(1001)).Return(mockCorporateProfile3, nil)
				return repo
			},
			userFileService: func() *mocks.UserFileService {
				svc := &mocks.UserFileService{}
				svc.On("GetCorporateUserIdentificationFiles", context.Background(), int64(1001)).
					Return(cdomain.CorporateUserIdentificationFiles{
						Front1:        &mockUserFile1,
						Back1:         &mockCertifiedCopyFile,
						Front2:        &mockUserFile1,
						Back2:         &mockCertifiedCopyFile,
						CertifiedCopy: &mockUserFile1,
						POA:           &mockCertifiedCopyFile,
					}, nil)

				return svc
			},

			want: udomain.CorporateUserGetSignUpStatusResponse{
				UserId:     1001,
				Email:      "gildong@email.com",
				PrevStatus: domain.SIGNUP_AGREE,
				CurStatus:  domain.SIGNUP_COMPLETE,
				NextStatus: domain.SIGNUP_COMPLETE,
				VerifyWithKyc: &udomain.CorporateUserVerifyWithKycItem{
					Identification1FrontFile: &mockUserFile1,
					Identification1BackFile:  &mockCertifiedCopyFile,
					Identification2FrontFile: &mockUserFile1,
					Identification2BackFile:  &mockCertifiedCopyFile,
					CertifiedCopyFile:        &mockUserFile1,
					PoaFile:                  &mockCertifiedCopyFile,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := corporateUserUsecase{
				userRepository:             tt.userRepository(),
				corporateProfileRepository: tt.corporateProfileRepository(),
				userFileService:            tt.userFileService(),
			}
			got, err := usecase.GetSignUpStatus(tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.wantErr, err != nil, "GetSignUpStatus() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want, got, "GetSignUpStatus() = %v, want %v", got, tt.want)
		})
	}
}

func Test_corporateUserUsecase_VerifyPostalCode(t *testing.T) {
	type args struct {
		ctx     context.Context
		request udomain.CorporateUserVerifyPostalCodeRequest
	}
	tests := []struct {
		name                       string
		args                       args
		userRepository             func() *mocks.UserRepository
		corporateProfileRepository func() *mocks.CorporateProfileRepository
		mailService                func() *mocks.MailService
		wantErr                    bool
	}{
		{
			name: "法人の郵便コードが一致しない場合は失敗する",
			args: args{
				ctx: context.Background(),
				request: udomain.CorporateUserVerifyPostalCodeRequest{
					User: jwthelper.User{
						Id: 1000,
					},
					CorporatePostalCode: "999999",
					ManagerPostalCode:   "654321",
				},
			},
			userRepository: func() *mocks.UserRepository {
				repo := &mocks.UserRepository{}
				repo.On("GetById", context.Background(), int64(1000)).Return(domain.User{
					Id:          1000,
					Name:        "ホン・ギルドン",
					Email:       "gildong@email.com",
					PostalCode:  "123456",
					AuditStatus: domain.CHECK_ADDRESS_AUDIT,
				}, nil)
				repo.On("Update", context.Background(), mock.Anything).Return(nil)
				return repo
			},
			corporateProfileRepository: func() *mocks.CorporateProfileRepository {
				repo := &mocks.CorporateProfileRepository{}
				repo.On("GetByUserId", context.Background(), int64(1000)).Return(domain.CorporateProfile{
					Id:                  2000,
					UserId:              1000,
					CorporatePostalCode: "654321",
				}, nil)
				repo.On("Update", context.Background(), mock.Anything).Return(nil)
				return repo
			},
			mailService: func() *mocks.MailService {
				svc := &mocks.MailService{}
				svc.On("SendMail", context.Background(), mock.Anything).Return(nil)
				return svc
			},
			wantErr: true,
		},
		{
			name: "担当者の郵便コードが一致しない場合は失敗する",
			args: args{
				ctx: context.Background(),
				request: udomain.CorporateUserVerifyPostalCodeRequest{
					User: jwthelper.User{
						Id: 1000,
					},
					CorporatePostalCode: "123456",
					ManagerPostalCode:   "999999",
				},
			},
			userRepository: func() *mocks.UserRepository {
				repo := &mocks.UserRepository{}
				repo.On("GetById", context.Background(), int64(1000)).Return(domain.User{
					Id:          1000,
					Name:        "ホン・ギルドン",
					Email:       "gildong@email.com",
					PostalCode:  "123456",
					AuditStatus: domain.CHECK_ADDRESS_AUDIT,
				}, nil)
				repo.On("Update", context.Background(), mock.Anything).Return(nil)
				return repo
			},
			corporateProfileRepository: func() *mocks.CorporateProfileRepository {
				repo := &mocks.CorporateProfileRepository{}
				repo.On("GetByUserId", context.Background(), int64(1000)).Return(domain.CorporateProfile{
					Id:                  2000,
					UserId:              1000,
					CorporatePostalCode: "654321",
				}, nil)
				repo.On("Update", context.Background(), mock.Anything).Return(nil)
				return repo
			},
			mailService: func() *mocks.MailService {
				svc := &mocks.MailService{}
				svc.On("SendMail", context.Background(), mock.Anything).Return(nil)
				return svc
			},
			wantErr: true,
		},
		{
			name: "本登録保留の法人は郵便コード再送で承認状態に戻せない",
			args: args{
				ctx: context.Background(),
				request: udomain.CorporateUserVerifyPostalCodeRequest{
					User: jwthelper.User{
						Id: 1000,
					},
					CorporatePostalCode: "123456",
					ManagerPostalCode:   "654321",
				},
			},
			// ガードが無ければコードが一致して承認完了まで到達してしまう構成にしている
			userRepository: func() *mocks.UserRepository {
				repo := &mocks.UserRepository{}
				repo.On("GetById", context.Background(), int64(1000)).Return(domain.User{
					Id:          1000,
					Name:        "ホン・ギルドン",
					Email:       "gildong@email.com",
					PostalCode:  "123456",
					AuditStatus: domain.HOLD_AUDIT,
				}, nil)
				repo.On("Update", context.Background(), mock.Anything).Return(nil)
				return repo
			},
			corporateProfileRepository: func() *mocks.CorporateProfileRepository {
				repo := &mocks.CorporateProfileRepository{}
				repo.On("GetByUserId", context.Background(), int64(1000)).Return(domain.CorporateProfile{
					Id:                  2000,
					UserId:              1000,
					CorporatePostalCode: "654321",
				}, nil)
				repo.On("Update", context.Background(), mock.Anything).Return(nil)
				return repo
			},
			mailService: func() *mocks.MailService {
				svc := &mocks.MailService{}
				svc.On("SendMail", context.Background(), mock.Anything).Return(nil)
				return svc
			},
			wantErr: true,
		},
		{
			name: "本登録承認完了済みの法人は郵便コード再送で承認日を上書きできない",
			args: args{
				ctx: context.Background(),
				request: udomain.CorporateUserVerifyPostalCodeRequest{
					User: jwthelper.User{
						Id: 1000,
					},
					CorporatePostalCode: "123456",
					ManagerPostalCode:   "654321",
				},
			},
			userRepository: func() *mocks.UserRepository {
				repo := &mocks.UserRepository{}
				repo.On("GetById", context.Background(), int64(1000)).Return(domain.User{
					Id:          1000,
					Name:        "ホン・ギルドン",
					Email:       "gildong@email.com",
					PostalCode:  "123456",
					AuditStatus: domain.COMPLETED_AUDIT,
				}, nil)
				repo.On("Update", context.Background(), mock.Anything).Return(nil)
				return repo
			},
			corporateProfileRepository: func() *mocks.CorporateProfileRepository {
				repo := &mocks.CorporateProfileRepository{}
				repo.On("GetByUserId", context.Background(), int64(1000)).Return(domain.CorporateProfile{
					Id:                  2000,
					UserId:              1000,
					CorporatePostalCode: "654321",
				}, nil)
				repo.On("Update", context.Background(), mock.Anything).Return(nil)
				return repo
			},
			mailService: func() *mocks.MailService {
				svc := &mocks.MailService{}
				svc.On("SendMail", context.Background(), mock.Anything).Return(nil)
				return svc
			},
			wantErr: true,
		},
		{
			name: "郵便番号検証成功",
			args: args{
				ctx: context.Background(),
				request: udomain.CorporateUserVerifyPostalCodeRequest{
					User: jwthelper.User{
						Id: 1000,
					},
					CorporatePostalCode: "123456",
					ManagerPostalCode:   "654321",
				},
			},
			userRepository: func() *mocks.UserRepository {
				repo := &mocks.UserRepository{}
				repo.On("GetById", context.Background(), int64(1000)).Return(domain.User{
					Id:          1000,
					Name:        "ホン・ギルドン",
					Email:       "gildong@email.com",
					PostalCode:  "123456",
					AuditStatus: domain.CHECK_ADDRESS_AUDIT,
				}, nil)

				repo.On("Update", context.Background(), mock.MatchedBy(func(user domain.User) bool {
					if user.PostalCodeStatus == true &&
						user.AuditStatus == domain.COMPLETED_AUDIT {
						return true
					}

					return false
				})).Return(nil)

				return repo

			},

			corporateProfileRepository: func() *mocks.CorporateProfileRepository {
				repo := &mocks.CorporateProfileRepository{}
				repo.On("GetByUserId", context.Background(), int64(1000)).Return(domain.CorporateProfile{
					Id:                  2000,
					UserId:              1000,
					CorporatePostalCode: "654321",
				}, nil)

				repo.On("Update", context.Background(), mock.MatchedBy(func(corporateProfile domain.CorporateProfile) bool {
					if corporateProfile.CorporatePostalStatus == true {
						return true
					}

					return false
				})).Return(nil)

				return repo
			},

			mailService: func() *mocks.MailService {
				mailService := &mocks.MailService{}
				mailService.On("SendMail", context.Background(), cdomain.SendMailRequest{
					TemplateType: domain.CODE_AUDIT_COMPLETE_FOR_CORPORATE_TEMPLATE,
					Sender:       "info@energyshares.jp",
					Receivers:    []string{"gildong@email.com"},
					MailType:     domain.MAIL_SEND_PRIORITY_NORMAL,
					Details: map[string]interface{}{
						"Receiver":   "ホン・ギルドン",
						"ServiceUrl": "https://dev.energyshares.jp",
						"ProductUrl": "https://dev.energyshares.jp/product/list",
					},
				}).Return(nil)
				return mailService
			},
			wantErr: false,
		},
		{
			name: "郵便番号検証失敗 - ユーザー情報照会失敗",
			args: args{
				ctx: context.Background(),
				request: udomain.CorporateUserVerifyPostalCodeRequest{
					User: jwthelper.User{
						Id: 1000,
					},
					CorporatePostalCode: "123456",
					ManagerPostalCode:   "654321",
				},
			},
			// ケース名どおり取得失敗を検証する。
			// 以前はユーザー取得が成功して担当者コード不一致で落ちており、
			// 名前と内容が食い違っていた（不一致は専用のケースで検証する）。
			userRepository: func() *mocks.UserRepository {
				repo := &mocks.UserRepository{}
				repo.On("GetById", context.Background(), int64(1000)).
					Return(domain.User{}, errors.New("find error"))

				return repo
			},

			corporateProfileRepository: func() *mocks.CorporateProfileRepository {
				repo := &mocks.CorporateProfileRepository{}
				repo.On("GetByUserId", context.Background(), int64(1000)).Return(domain.CorporateProfile{
					Id:                  2000,
					UserId:              1000,
					CorporatePostalCode: "654321",
				}, nil)

				return repo
			},
			mailService: func() *mocks.MailService {
				mailService := &mocks.MailService{}
				return mailService
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 失敗時に副作用が無いことを確かめるため、モックを変数に束縛しておく
			userRepository := tt.userRepository()
			mailService := tt.mailService()

			usecase := corporateUserUsecase{
				cfg: &config.Config{
					CommonConfig: config.CommonConfig{
						ServiceConfig: struct {
							Mail serviceutils.MailConfig
							Url  serviceutils.UrlConfig
						}{
							Mail: serviceutils.MailConfig{
								Info: "info@energyshares.jp",
							},
							Url: serviceutils.UrlConfig{
								UserServiceDomain: "https://dev.energyshares.jp",
								Product:           "/product/list",
							},
						},
					},
				},
				userRepository:             userRepository,
				corporateProfileRepository: tt.corporateProfileRepository(),
				mailService:                mailService,
			}

			err := usecase.VerifyPostalCode(tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.wantErr, err != nil, "VerifyPostalCode() error = %v, wantErr %v", err, tt.wantErr)

			// 失敗したときは承認状態の更新も承認メールの送信も起こしてはいけない
			if tt.wantErr {
				userRepository.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
				mailService.AssertNotCalled(t, "SendMail", mock.Anything, mock.Anything)
			}
		})
	}
}

func Test_corporateUserUsecase_PatchOne(t *testing.T) {
	type args struct {
		ctx     context.Context
		request udomain.CorporateUserPatchOneRequest
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
				request: udomain.CorporateUserPatchOneRequest{
					JwtUser: jwthelper.User{
						Id: 1000,
					},
					IsNewsletter: lo.ToPtr(true),
				},
			},
			userRepository: func() *mocks.UserRepository {
				repo := &mocks.UserRepository{}

				repo.On("GetById", context.Background(), int64(1000)).Return(domain.User{
					Id:           1000,
					UserType:     domain.CORPORATE,
					IsNewsletter: false,
				}, nil)
				repo.On("Update", context.Background(), domain.User{
					Id:           1000,
					UserType:     domain.CORPORATE,
					IsNewsletter: true,
				}).Return(nil)
				return repo
			},
		},
		{
			name: "会員情報修正成功 - ニュースレター値をnil値に更新を試行",
			args: args{
				ctx: context.Background(),
				request: udomain.CorporateUserPatchOneRequest{
					JwtUser: jwthelper.User{
						Id: 1000,
					},
				},
			},
			userRepository: func() *mocks.UserRepository {
				repo := &mocks.UserRepository{}

				repo.On("GetById", context.Background(), int64(1000)).Return(domain.User{
					Id:           1000,
					UserType:     domain.CORPORATE,
					IsNewsletter: false,
				}, nil)
				repo.On("Update", context.Background(), domain.User{
					Id:           1000,
					UserType:     domain.CORPORATE,
					IsNewsletter: false,
				}).Return(nil)
				return repo
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := corporateUserUsecase{
				userRepository: tt.userRepository(),
			}
			err := usecase.PatchOne(tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.wantErr, err != nil, "PatchOne() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_corporateUserUsecase_GetDocuments(t *testing.T) {
	type args struct {
		ctx     context.Context
		request udomain.CorporateUserGetDocumentsRequest
	}

	tests := []struct {
		name            string
		args            args
		userFileService func() *mocks.UserFileService
		want            udomain.CorporateUserGetDocumentsResponse
		wantErr         bool
	}{
		{
			name: "(成功) 登録したファイル一覧照会",
			args: args{
				ctx: context.Background(),
				request: udomain.CorporateUserGetDocumentsRequest{
					JwtUser: jwthelper.User{
						Id: 1000,
					},
				},
			},
			userFileService: func() *mocks.UserFileService {
				repo := &mocks.UserFileService{}
				repo.On("GetCorporateUserIdentificationFiles", context.Background(), int64(1000)).
					Return(cdomain.CorporateUserIdentificationFiles{
						CertifiedCopy: &mockCertifiedCopyFile,
						POA:           &mockPOAFile,
					}, nil)
				return repo
			},
			want: udomain.CorporateUserGetDocumentsResponse{
				CorporateUserVerifyWithKycItem: udomain.CorporateUserVerifyWithKycItem{
					Identification1FrontFile: nil,
					Identification1BackFile:  nil,
					Identification2FrontFile: nil,
					Identification2BackFile:  nil,
					CertifiedCopyFile:        &mockCertifiedCopyFile,
					PoaFile:                  &mockPOAFile,
					AoiFile:                  nil,
					ShareholdersFile:         nil,
					OthersFile:               nil,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := corporateUserUsecase{
				userFileService: tt.userFileService(),
			}
			got, err := usecase.GetDocuments(tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.wantErr, err != nil, "GetDocuments() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want, got, "GetDocuments() = %v, want %v", got, tt.want)
		})
	}
}

func Test_corporateUserUsecase_UpdateDocuments(t *testing.T) {
	type args struct {
		ctx     context.Context
		request udomain.CorporateUserUpdateDocumentsRequest
	}

	const (
		existingFileName = "cou7922nulrn470ulh4g"
		replacedFileName = "replaced_certified_copy"
		existingUrl      = "https://cdn.dev.energyshares.jp/user_file/2024-05-09/cou7922nulrn470ulh4g.jpeg"
		replacedUrl      = "https://cdn.dev.energyshares.jp/user_file/2024-05-09/replaced_certified_copy.jpeg"
	)

	certifiedCopyRequest := func(url string) args {
		return args{
			ctx: context.Background(),
			request: udomain.CorporateUserUpdateDocumentsRequest{
				JwtUser: jwthelper.User{Id: 1000},
				CorporateUserFileUrlItem: udomain.CorporateUserFileUrlItem{
					CertifiedCopyFileUrl: lo.ToPtr(url),
				},
			},
		}
	}

	holdUser := func() *mocks.UserRepository {
		repo := &mocks.UserRepository{}
		repo.On("GetById", context.Background(), int64(1000)).Return(mockUser1, nil)
		repo.On("Update", context.Background(), mock.Anything).Return(nil)
		return repo
	}

	s3ok := func() *mocks.S3 {
		client := &mocks.S3{}
		client.On("GetHeadObject", s3utils.BUCKET_PRIVATE_NAME, mock.Anything).Return(mockS3ObjectInfo1, nil)
		return client
	}

	tests := []struct {
		name               string
		args               args
		userRepository     func() *mocks.UserRepository
		s3pkg              func() *mocks.S3
		userFileRepository func() *mocks.UserFileRepository
		wantCreate         bool
		wantDelete         bool
		wantWaitAudit      bool
		wantNoStatusChange bool
		wantErr            bool
	}{
		{
			name:           "(成功) 差し戻された書類を別のファイルに差し替えられる",
			args:           certifiedCopyRequest(replacedUrl),
			userRepository: holdUser,
			s3pkg:          s3ok,
			// 同名の行が無いので、差し替えとして新しい行を作る
			userFileRepository: func() *mocks.UserFileRepository {
				repo := &mocks.UserFileRepository{}
				repo.On("ExistsByFileName", context.Background(), replacedFileName).Return(false, nil)
				repo.On("CreateMany", context.Background(), mock.MatchedBy(func(files []domain.UserFile) bool {
					if len(files) != 1 {
						return false
					}
					f := files[0]
					return f.FileType == domain.CERTIFIED_COPY &&
						f.FileName == replacedFileName &&
						f.FileExtension == ".jpeg" &&
						f.FilePath == "/user_file/2024-05-09/" &&
						f.UserId == 1000 &&
						f.OwnerId == 1000 &&
						f.UpdateId == 1000 &&
						f.UserType == domain.CORPORATE
				})).Return(nil)
				return repo
			},
			wantCreate:    true,
			wantWaitAudit: true,
			wantErr:       false,
		},
		{
			name: "(成功) 既存の書類を残したまま一部だけ差し替えられる",
			args: args{
				ctx: context.Background(),
				request: udomain.CorporateUserUpdateDocumentsRequest{
					JwtUser: jwthelper.User{Id: 1000},
					CorporateUserFileUrlItem: udomain.CorporateUserFileUrlItem{
						// 既存と同名（差し替え対象外）
						Identification1FrontFileUrl: lo.ToPtr(existingUrl),
						// 差し替えたい書類
						CertifiedCopyFileUrl: lo.ToPtr(replacedUrl),
					},
				},
			},
			userRepository: holdUser,
			s3pkg:          s3ok,
			// 差し替えた1件だけを作る
			userFileRepository: func() *mocks.UserFileRepository {
				repo := &mocks.UserFileRepository{}
				repo.On("ExistsByFileName", context.Background(), existingFileName).Return(true, nil)
				repo.On("ExistsByFileName", context.Background(), replacedFileName).Return(false, nil)
				repo.On("CreateMany", context.Background(), mock.MatchedBy(func(files []domain.UserFile) bool {
					return len(files) == 1 &&
						files[0].FileType == domain.CERTIFIED_COPY &&
						files[0].FileName == replacedFileName
				})).Return(nil)
				return repo
			},
			wantCreate: true,
			wantErr:    false,
		},
		{
			name: "(成功) 同一リクエスト内で同じファイルを重ねても行は1件だけ作る",
			args: args{
				ctx: context.Background(),
				request: udomain.CorporateUserUpdateDocumentsRequest{
					JwtUser: jwthelper.User{Id: 1000},
					CorporateUserFileUrlItem: udomain.CorporateUserFileUrlItem{
						// 同じファイルを2つの枠に入れる。
						// file_name は一意なので、2件作ると unique 制約に違反する。
						CertifiedCopyFileUrl:        lo.ToPtr(replacedUrl),
						Identification1FrontFileUrl: lo.ToPtr(replacedUrl),
					},
				},
			},
			userRepository: holdUser,
			s3pkg:          s3ok,
			userFileRepository: func() *mocks.UserFileRepository {
				repo := &mocks.UserFileRepository{}
				repo.On("ExistsByFileName", context.Background(), replacedFileName).Return(false, nil)
				repo.On("CreateMany", context.Background(), mock.MatchedBy(func(files []domain.UserFile) bool {
					return len(files) == 1 && files[0].FileName == replacedFileName
				})).Return(nil)
				return repo
			},
			wantCreate: true,
			wantErr:    false,
		},
		{
			name:           "(失敗) 同名のファイルは審査状態を動かさず弾く",
			args:           certifiedCopyRequest(existingUrl),
			userRepository: holdUser,
			s3pkg:          s3ok,
			// file_name の unique 制約に違反するため作らない。
			// 追加対象が無いので審査待ちへ遷移させずにエラーで返す。
			// リポジトリ側は論理削除済みの行も Unscoped で数えるため、
			// 管理画面で添付を削除したあとの再送信もここに入る。
			userFileRepository: func() *mocks.UserFileRepository {
				repo := &mocks.UserFileRepository{}
				repo.On("ExistsByFileName", context.Background(), existingFileName).Return(true, nil)
				repo.On("CreateMany", context.Background(), mock.Anything).Return(nil)
				return repo
			},
			wantCreate:         false,
			wantNoStatusChange: true,
			wantErr:            true,
		},
		{
			name: "(失敗) 本登録保留以外は書類を更新できない",
			args: certifiedCopyRequest(replacedUrl),
			userRepository: func() *mocks.UserRepository {
				repo := &mocks.UserRepository{}
				completed := mockUser1
				completed.AuditStatus = domain.COMPLETED_AUDIT
				repo.On("GetById", context.Background(), int64(1000)).Return(completed, nil)
				repo.On("Update", context.Background(), mock.Anything).Return(nil)
				return repo
			},
			s3pkg: s3ok,
			userFileRepository: func() *mocks.UserFileRepository {
				repo := &mocks.UserFileRepository{}
				repo.On("ExistsByFileName", context.Background(), mock.Anything).Return(false, nil)
				repo.On("CreateMany", context.Background(), mock.Anything).Return(nil)
				return repo
			},
			wantCreate: false,
			wantErr:    true,
		},
		{
			name:           "(失敗) 重複確認でエラーが出たら書類を作らない",
			args:           certifiedCopyRequest(replacedUrl),
			userRepository: holdUser,
			s3pkg:          s3ok,
			userFileRepository: func() *mocks.UserFileRepository {
				repo := &mocks.UserFileRepository{}
				repo.On("ExistsByFileName", context.Background(), replacedFileName).
					Return(false, errors.New("db error"))
				repo.On("CreateMany", context.Background(), mock.Anything).Return(nil)
				return repo
			},
			wantCreate: false,
			wantErr:    true,
		},
		{
			name: "(成功) 既存ファイルの削除だけを行い、審査状態は動かさない",
			args: args{
				ctx: context.Background(),
				request: udomain.CorporateUserUpdateDocumentsRequest{
					JwtUser:       jwthelper.User{Id: 1000},
					DeleteFileIds: []int64{2000},
				},
			},
			userRepository: holdUser,
			s3pkg:          s3ok,
			userFileRepository: func() *mocks.UserFileRepository {
				repo := &mocks.UserFileRepository{}
				repo.On("GetById", context.Background(), int64(2000)).
					Return(&domain.UserFile{Id: 2000, UserId: 1000, FileType: domain.CERTIFIED_COPY}, nil)
				repo.On("Delete", context.Background(), int64(2000)).Return(nil)
				return repo
			},
			wantCreate:         false,
			wantDelete:         true,
			wantNoStatusChange: true,
			wantErr:            false,
		},
		{
			// フィールドが省略されるとポインタは nil になるが、クライアントが
			// 明示的に空文字を送るとポインタは非nilになる。どちらも「入力なし」
			// として扱い、削除だけの正当なリクエストを弾かないことを確かめる。
			name: "(成功) 空文字のファイルURLが明示的に送られても削除だけのリクエストとして成立する",
			args: args{
				ctx: context.Background(),
				request: udomain.CorporateUserUpdateDocumentsRequest{
					JwtUser:       jwthelper.User{Id: 1000},
					DeleteFileIds: []int64{2000},
					CorporateUserFileUrlItem: udomain.CorporateUserFileUrlItem{
						CertifiedCopyFileUrl: lo.ToPtr(""),
					},
				},
			},
			userRepository: holdUser,
			s3pkg:          s3ok,
			userFileRepository: func() *mocks.UserFileRepository {
				repo := &mocks.UserFileRepository{}
				repo.On("GetById", context.Background(), int64(2000)).
					Return(&domain.UserFile{Id: 2000, UserId: 1000, FileType: domain.CERTIFIED_COPY}, nil)
				repo.On("Delete", context.Background(), int64(2000)).Return(nil)
				return repo
			},
			wantCreate:         false,
			wantDelete:         true,
			wantNoStatusChange: true,
			wantErr:            false,
		},
		{
			name: "(成功) 削除と差し替えを同時に行う",
			args: args{
				ctx: context.Background(),
				request: udomain.CorporateUserUpdateDocumentsRequest{
					JwtUser:       jwthelper.User{Id: 1000},
					DeleteFileIds: []int64{2000},
					CorporateUserFileUrlItem: udomain.CorporateUserFileUrlItem{
						CertifiedCopyFileUrl: lo.ToPtr(replacedUrl),
					},
				},
			},
			userRepository: holdUser,
			s3pkg:          s3ok,
			userFileRepository: func() *mocks.UserFileRepository {
				repo := &mocks.UserFileRepository{}
				repo.On("GetById", context.Background(), int64(2000)).
					Return(&domain.UserFile{Id: 2000, UserId: 1000, FileType: domain.CERTIFIED_COPY}, nil)
				repo.On("Delete", context.Background(), int64(2000)).Return(nil)
				repo.On("ExistsByFileName", context.Background(), replacedFileName).Return(false, nil)
				repo.On("CreateMany", context.Background(), mock.Anything).Return(nil)
				return repo
			},
			wantCreate:    true,
			wantDelete:    true,
			wantWaitAudit: true,
			wantErr:       false,
		},
		{
			name: "(失敗) 削除と同時に指定した差し替えが重複でスキップされたら、削除も含めて弾く",
			args: args{
				ctx: context.Background(),
				request: udomain.CorporateUserUpdateDocumentsRequest{
					JwtUser:       jwthelper.User{Id: 1000},
					DeleteFileIds: []int64{2000},
					CorporateUserFileUrlItem: udomain.CorporateUserFileUrlItem{
						// 既存と同名(差し替えとして成立しない)
						CertifiedCopyFileUrl: lo.ToPtr(existingUrl),
					},
				},
			},
			userRepository: holdUser,
			s3pkg:          s3ok,
			// 差し替えのつもりが全滅した場合、削除だけを成功させて
			// 気付かれないまま終わらせてはいけない。Delete は呼ばれない。
			userFileRepository: func() *mocks.UserFileRepository {
				repo := &mocks.UserFileRepository{}
				repo.On("GetById", context.Background(), int64(2000)).
					Return(&domain.UserFile{Id: 2000, UserId: 1000, FileType: domain.CERTIFIED_COPY}, nil)
				repo.On("Delete", context.Background(), mock.Anything).Return(nil)
				repo.On("ExistsByFileName", context.Background(), existingFileName).Return(true, nil)
				repo.On("CreateMany", context.Background(), mock.Anything).Return(nil)
				return repo
			},
			wantCreate:         false,
			wantDelete:         false,
			wantNoStatusChange: true,
			wantErr:            true,
		},
		{
			name: "(失敗) 他人のファイルは削除できない",
			args: args{
				ctx: context.Background(),
				request: udomain.CorporateUserUpdateDocumentsRequest{
					JwtUser:       jwthelper.User{Id: 1000},
					DeleteFileIds: []int64{2000},
				},
			},
			userRepository: holdUser,
			s3pkg:          s3ok,
			userFileRepository: func() *mocks.UserFileRepository {
				repo := &mocks.UserFileRepository{}
				// 所有者が別ユーザー(9999)のファイル
				repo.On("GetById", context.Background(), int64(2000)).
					Return(&domain.UserFile{Id: 2000, UserId: 9999}, nil)
				repo.On("Delete", context.Background(), mock.Anything).Return(nil)
				repo.On("CreateMany", context.Background(), mock.Anything).Return(nil)
				return repo
			},
			wantCreate:         false,
			wantDelete:         false,
			wantNoStatusChange: true,
			wantErr:            true,
		},
		{
			name: "(失敗) 法人書類以外の種別(マイナンバー等)は削除できない",
			args: args{
				ctx: context.Background(),
				request: udomain.CorporateUserUpdateDocumentsRequest{
					JwtUser:       jwthelper.User{Id: 1000},
					DeleteFileIds: []int64{2000},
				},
			},
			userRepository: holdUser,
			s3pkg:          s3ok,
			userFileRepository: func() *mocks.UserFileRepository {
				repo := &mocks.UserFileRepository{}
				// 所有者は自分だが、この画面が扱わない種別(マイナンバー)
				repo.On("GetById", context.Background(), int64(2000)).
					Return(&domain.UserFile{Id: 2000, UserId: 1000, FileType: domain.MY_NUMBER_FRONT}, nil)
				repo.On("Delete", context.Background(), mock.Anything).Return(nil)
				repo.On("CreateMany", context.Background(), mock.Anything).Return(nil)
				return repo
			},
			wantCreate:         false,
			wantDelete:         false,
			wantNoStatusChange: true,
			wantErr:            true,
		},
		{
			name: "(失敗) 存在しないファイルは削除できない",
			args: args{
				ctx: context.Background(),
				request: udomain.CorporateUserUpdateDocumentsRequest{
					JwtUser:       jwthelper.User{Id: 1000},
					DeleteFileIds: []int64{99999},
				},
			},
			userRepository: holdUser,
			s3pkg:          s3ok,
			userFileRepository: func() *mocks.UserFileRepository {
				repo := &mocks.UserFileRepository{}
				repo.On("GetById", context.Background(), int64(99999)).
					Return(nil, cdomain.ErrItemNotFound)
				repo.On("Delete", context.Background(), mock.Anything).Return(nil)
				repo.On("CreateMany", context.Background(), mock.Anything).Return(nil)
				return repo
			},
			wantCreate:         false,
			wantDelete:         false,
			wantNoStatusChange: true,
			wantErr:            true,
		},
		{
			name: "(失敗) 削除も差し替えも無ければ審査状態を動かさず弾く",
			args: args{
				ctx: context.Background(),
				request: udomain.CorporateUserUpdateDocumentsRequest{
					JwtUser: jwthelper.User{Id: 1000},
				},
			},
			userRepository: holdUser,
			s3pkg:          s3ok,
			userFileRepository: func() *mocks.UserFileRepository {
				repo := &mocks.UserFileRepository{}
				repo.On("Delete", context.Background(), mock.Anything).Return(nil)
				repo.On("CreateMany", context.Background(), mock.Anything).Return(nil)
				return repo
			},
			wantCreate:         false,
			wantDelete:         false,
			wantNoStatusChange: true,
			wantErr:            true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userFileRepository := tt.userFileRepository()
			userRepository := tt.userRepository()

			usecase := corporateUserUsecase{
				cfg: &config.Config{
					CommonConfig: config.CommonConfig{
						ServiceConfig: struct {
							Mail serviceutils.MailConfig
							Url  serviceutils.UrlConfig
						}{
							Url: serviceutils.UrlConfig{
								UserServiceDomain: "https://user.dev.energyshares.com",
							},
							Mail: serviceutils.MailConfig{
								Info: "info@energyshares.jp",
							},
						},
					},
				},
				fileHelper:         filehelper.NewFileHelper(),
				logger:             factory.NewLoggerFactory(),
				userRepository:     userRepository,
				s3pkg:              tt.s3pkg(),
				userFileRepository: userFileRepository,
			}
			err := usecase.UpdateDocuments(tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.wantErr, err != nil, "UpdateDocuments() error = %v, wantErr %v", err, tt.wantErr)

			// 「作る」側も固定する。作らないケースで作られていないことも確かめる
			if tt.wantCreate {
				userFileRepository.AssertCalled(t, "CreateMany", mock.Anything, mock.Anything)
			} else {
				userFileRepository.AssertNotCalled(t, "CreateMany", mock.Anything, mock.Anything)
			}

			// 削除も同様に、削除したケースとしないケースを固定する。
			if tt.wantDelete {
				userFileRepository.AssertCalled(t, "Delete", mock.Anything, mock.Anything)
			} else {
				userFileRepository.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
			}

			// 弾くケースでは審査状態も動かさない。
			// ガードが userRepository.Update より後ろへ動くと落ちる。
			if tt.wantNoStatusChange {
				userRepository.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
			}

			// 差し替えたら審査待ちへ戻す
			if tt.wantWaitAudit {
				userRepository.AssertCalled(t, "Update", mock.Anything, mock.MatchedBy(func(u domain.User) bool {
					return u.AuditStatus == domain.WAIT_AUDIT
				}))
			}
		})
	}
}
