package user

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/egift"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/idgenerator"
	"github.com/stretchr/testify/mock"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/google/go-cmp/cmp"

	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/s3utils"

	"github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain/mocks"

	"github.com/stretchr/testify/assert"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"

	udomain "github.com/illuminarean/energyshares-user-api/domain"
)

// matchLoginLockSortById/matchLoginLockOrderDescはMIR-1233の検証用。
// WithSortBy("id")・WithOrder("desc")が正しい向きで渡されていることを、
// モックの引数として直接検証する。引数が入れ替わっていた修正前のバグ
// (WithSortByへ"desc"を、WithOrderへ"id"を渡していた)は、このマッチャーが
// 無いとGetAllの呼び出し自体は成立してしまい、テストで検知できない。
func matchLoginLockSortById(field func(*cdomain.UserLoginLockQueryOptionFields)) bool {
	opt := cdomain.UserLoginLockQueryOptionFunc()
	field(&opt)
	return opt.SortBy() == "id"
}

func matchLoginLockOrderDesc(field func(*cdomain.UserLoginLockQueryOptionFields)) bool {
	opt := cdomain.UserLoginLockQueryOptionFunc()
	field(&opt)
	return opt.Order() == "desc"
}

func Test_userUsecase_Login(t *testing.T) {
	type args struct {
		ctx     context.Context
		request udomain.UserLoginRequest
	}

	tests := []struct {
		name                    string
		args                    args
		cfg                     *config.Config
		s3                      func() *mocks.S3
		txManager               func() *mocks.TxManager
		pwhelper                func() *mocks.PasswordHelper
		loginHistoryRepository  func() *mocks.LoginHistoryRepository
		userRepository          func() *mocks.UserRepository
		userLoginLockRepository func() *mocks.UserLoginLockRepository
		authService             func() *mocks.AuthService
		mailService             func() *mocks.MailService
		want                    udomain.UserLoginResponse
		wantErr                 bool
	}{
		{
			name: "(成功) 正常ログイン",
			args: args{
				ctx: context.Background(),
				request: udomain.UserLoginRequest{
					Email:     loginMockUser1.Email,
					Password:  password,
					ClientIp:  "ip-address",
					UserAgent: "user-agent",
				},
			},
			s3: func() *mocks.S3 {
				s3 := &mocks.S3{}
				s3.On("CreateCookieSign").Return([]*http.Cookie{}, nil)
				return s3
			},
			txManager: func() *mocks.TxManager {
				txManager := &mocks.TxManager{}
				txManager.On("WithTx", mock.Anything, mock.Anything).Return(nil)
				return txManager
			},
			pwhelper: func() *mocks.PasswordHelper {
				helper := &mocks.PasswordHelper{}
				helper.On("EncodePassword", password, loginMockUser1.Salt).Return(encodedPassword)
				helper.On("IsSamePassword", loginMockUser1.Password, encodedPassword).Return(true)
				return helper
			},
			loginHistoryRepository: func() *mocks.LoginHistoryRepository {
				repo := &mocks.LoginHistoryRepository{}
				repo.On("Create", context.Background(), mock.MatchedBy(func(history domain.LoginHistory) bool {
					return history.UserId == loginMockUser1.Id &&
						history.IpAddr == "ip-address" &&
						history.Agent == "user-agent" &&
						history.LoginResult == domain.SUCCESS_LOGIN
				})).Return(int64(1), nil)
				return repo
			},
			userRepository: func() *mocks.UserRepository {
				repo := &mocks.UserRepository{}
				repo.On("GetByEmail", context.Background(), loginMockUser1.Email).Return(loginMockUser1, nil)
				// 読み込んだ時点の古いLastLoginDate(loginMockUser1のそれ)ではなく、
				// 現在時刻に近い値で更新していることを確かめる(MIR-1373)。
				repo.On("UpdateLastLoginDate", context.Background(), loginMockUser1.Id, mock.MatchedBy(func(lastLoginDate jsonutils.JSONTime) bool {
					return time.Since(lastLoginDate.Time()) < time.Minute
				})).Return(nil)
				return repo
			},
			userLoginLockRepository: func() *mocks.UserLoginLockRepository {
				repo := &mocks.UserLoginLockRepository{}
				repo.On("GetAll", context.Background(),
					mock.MatchedBy(func(field func(*cdomain.UserLoginLockQueryOptionFields)) bool {
						opt := cdomain.UserLoginLockQueryOptionFunc()
						field(&opt)
						return *opt.UserId() == fmt.Sprintf("%v", loginMockUser1.Id)
					}),
					mock.Anything,
					mock.MatchedBy(matchLoginLockSortById),
					mock.MatchedBy(matchLoginLockOrderDesc),
					mock.Anything,
					mock.Anything,
				).Return(int64(1), []domain.UserLoginLock{loginMockUserLoginLock1}, nil)
				repo.On("Delete", context.Background(), loginMockUserLoginLock1.Id).Return(nil)
				return repo
			},
			authService: func() *mocks.AuthService {
				service := &mocks.AuthService{}
				service.On("CreateUserAccessToken", context.Background(), loginMockUser1).Return("access-token", nil)
				service.On("CreateUserRefreshToken", context.Background(), cdomain.UserRefreshTokenInput{
					User:  loginMockUser1,
					IP:    "ip-address",
					Agent: "user-agent",
				}).Return("refresh-token", nil)
				return service
			},
			mailService: func() *mocks.MailService {
				service := &mocks.MailService{}
				return service
			},
			want: udomain.UserLoginResponse{
				AccessToken:  "access-token",
				RefreshToken: "refresh-token",
				Cookies:      []*http.Cookie{},
			},
			wantErr: false,
		},
		{
			name: "(失敗) ログイン7回未満失敗",
			args: args{
				ctx: context.Background(),
				request: udomain.UserLoginRequest{
					Email:     loginMockUser1.Email,
					Password:  "incorrect-password",
					ClientIp:  "ip-address",
					UserAgent: "user-agent",
				},
			},
			s3: func() *mocks.S3 {
				s3 := &mocks.S3{}
				return s3
			},
			txManager: func() *mocks.TxManager {
				txManager := &mocks.TxManager{}
				txManager.On("WithTx", mock.Anything, mock.Anything).Return(nil)
				return txManager
			},
			pwhelper: func() *mocks.PasswordHelper {
				helper := &mocks.PasswordHelper{}
				helper.On("EncodePassword", "incorrect-password", loginMockUser1.Salt).Return("encoded-incorrect-password")
				helper.On("IsSamePassword", loginMockUser1.Password, "encoded-incorrect-password").Return(false)
				return helper
			},
			loginHistoryRepository: func() *mocks.LoginHistoryRepository {
				repo := &mocks.LoginHistoryRepository{}
				repo.On("Create", context.Background(), mock.MatchedBy(func(history domain.LoginHistory) bool {
					return history.UserId == loginMockUser1.Id &&
						history.IpAddr == "ip-address" &&
						history.Agent == "user-agent" &&
						history.LoginResult == domain.FAIL_LOGIN
				})).Return(int64(1), nil)
				return repo
			},
			userRepository: func() *mocks.UserRepository {
				repo := &mocks.UserRepository{}
				repo.On("GetByEmail", context.Background(), loginMockUser1.Email).Return(loginMockUser1, nil)

				return repo
			},
			userLoginLockRepository: func() *mocks.UserLoginLockRepository {
				repo := &mocks.UserLoginLockRepository{}
				repo.On("GetAll", context.Background(),
					mock.MatchedBy(func(field func(*cdomain.UserLoginLockQueryOptionFields)) bool {
						opt := cdomain.UserLoginLockQueryOptionFunc()
						field(&opt)
						return *opt.UserId() == fmt.Sprintf("%v", loginMockUser1.Id)
					}),
					mock.Anything,
					mock.MatchedBy(matchLoginLockSortById),
					mock.MatchedBy(matchLoginLockOrderDesc),
					mock.Anything,
					mock.Anything,
				).Return(int64(1), []domain.UserLoginLock{loginMockUserLoginLock4}, nil)
				loginMockUserLoginLock4.LoginFailCnt++
				repo.On("Update", context.Background(), mock.MatchedBy(func(lock domain.UserLoginLock) bool {
					return lock.LoginFailCnt == loginMockUserLoginLock4.LoginFailCnt
				})).Return(nil)
				return repo
			},
			authService: func() *mocks.AuthService {
				service := &mocks.AuthService{}
				return service
			},
			mailService: func() *mocks.MailService {
				service := &mocks.MailService{}
				return service
			},
			want:    udomain.UserLoginResponse{},
			wantErr: true,
		},
		{
			name: "(失敗) ログイン7回以上失敗",
			args: args{
				ctx: context.Background(),
				request: udomain.UserLoginRequest{
					Email:     loginMockUser1.Email,
					Password:  "incorrect-password",
					ClientIp:  "ip-address",
					UserAgent: "user-agent",
				},
			},
			s3: func() *mocks.S3 {
				s3 := &mocks.S3{}
				return s3
			},
			txManager: func() *mocks.TxManager {
				txManager := &mocks.TxManager{}
				txManager.On("WithTx", mock.Anything, mock.Anything).Return(nil)
				return txManager
			},
			pwhelper: func() *mocks.PasswordHelper {
				helper := &mocks.PasswordHelper{}
				helper.On("EncodePassword", "incorrect-password", loginMockUser1.Salt).Return("encoded-incorrect-password")
				helper.On("IsSamePassword", loginMockUser1.Password, "encoded-incorrect-password").Return(false)
				return helper
			},
			loginHistoryRepository: func() *mocks.LoginHistoryRepository {
				repo := &mocks.LoginHistoryRepository{}
				repo.On("Create", context.Background(), mock.MatchedBy(func(history domain.LoginHistory) bool {
					return history.UserId == loginMockUser1.Id &&
						history.IpAddr == "ip-address" &&
						history.Agent == "user-agent" &&
						history.LoginResult == domain.FAIL_LOGIN
				})).Return(int64(1), nil)
				return repo
			},
			userRepository: func() *mocks.UserRepository {
				repo := &mocks.UserRepository{}
				repo.On("GetByEmail", context.Background(), loginMockUser1.Email).Return(loginMockUser1, nil)

				return repo
			},
			userLoginLockRepository: func() *mocks.UserLoginLockRepository {
				repo := &mocks.UserLoginLockRepository{}
				repo.On("GetAll", context.Background(),
					mock.MatchedBy(func(field func(*cdomain.UserLoginLockQueryOptionFields)) bool {
						opt := cdomain.UserLoginLockQueryOptionFunc()
						field(&opt)
						return *opt.UserId() == fmt.Sprintf("%v", loginMockUser1.Id)
					}),
					mock.Anything,
					mock.MatchedBy(matchLoginLockSortById),
					mock.MatchedBy(matchLoginLockOrderDesc),
					mock.Anything,
					mock.Anything,
				).Return(int64(1), []domain.UserLoginLock{loginMockUserLoginLock6}, nil)
				loginMockUserLoginLock6.LoginFailCnt++
				repo.On("Update", context.Background(), mock.MatchedBy(func(lock domain.UserLoginLock) bool {
					return lock.LoginFailCnt == loginMockUserLoginLock6.LoginFailCnt
				})).Return(nil)
				return repo
			},
			authService: func() *mocks.AuthService {
				service := &mocks.AuthService{}
				return service
			},
			mailService: func() *mocks.MailService {
				service := &mocks.MailService{}
				service.On("SendMail", context.Background(), mock.MatchedBy(func(req cdomain.SendMailRequest) bool {
					return req.TemplateType == domain.USER_LOGIN_FAIL_LOCK_TEMPLATE
				})).Return(nil)
				return service
			},
			want:    udomain.UserLoginResponse{},
			wantErr: true,
		},
		{
			name: "(失敗) '無効'状態のアカウントでログイン試行",
			args: args{
				ctx: context.Background(),
				request: udomain.UserLoginRequest{
					Email:     loginMockUser2.Email,
					Password:  password,
					ClientIp:  "ip-address",
					UserAgent: "user-agent",
				},
			},
			s3: func() *mocks.S3 {
				s3 := &mocks.S3{}
				return s3
			},
			txManager: func() *mocks.TxManager {
				txManager := &mocks.TxManager{}
				txManager.On("WithTx", mock.Anything, mock.Anything).Return(nil)
				return txManager
			},
			pwhelper: func() *mocks.PasswordHelper {
				helper := new(mocks.PasswordHelper)
				return helper
			},
			loginHistoryRepository: func() *mocks.LoginHistoryRepository {
				repo := new(mocks.LoginHistoryRepository)
				return repo
			},
			userRepository: func() *mocks.UserRepository {
				repo := new(mocks.UserRepository)
				repo.On("GetByEmail", context.Background(), loginMockUser2.Email).Return(loginMockUser2, nil)

				return repo
			},
			userLoginLockRepository: func() *mocks.UserLoginLockRepository {
				repo := new(mocks.UserLoginLockRepository)
				return repo
			},
			authService: func() *mocks.AuthService {
				service := new(mocks.AuthService)
				return service
			},
			mailService: func() *mocks.MailService {
				service := new(mocks.MailService)
				return service
			},
			want:    udomain.UserLoginResponse{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepository := tt.userRepository()
			usecase := userUsecase{
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
				s3:                      tt.s3(),
				txManager:               tt.txManager(),
				pwhelper:                tt.pwhelper(),
				loginHistoryRepository:  tt.loginHistoryRepository(),
				userRepository:          userRepository,
				userLoginLockRepository: tt.userLoginLockRepository(),
				authService:             tt.authService(),
				mailService:             tt.mailService(),
			}

			got, err := usecase.Login(tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.wantErr, err != nil, "Login() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want, got, "Login() = %v, want %v", got, tt.want)
			// MIR-1373: UpdateLastLoginDateが呼ばれなくなる回帰(last_login_dateが
			// 更新されないままになる)を、モックの期待値未消化で検知できるようにする。
			userRepository.AssertExpectations(t)
		})
	}
}

func Test_userUsecase_CreatePreSignedUrl(t *testing.T) {
	timeNowDateOnly := time.Now().Format(time.DateOnly)

	type args struct {
		ctx     context.Context
		request udomain.UserCreatePreSignedUrlRequest
	}

	tests := []struct {
		name    string
		args    args
		s3pkg   func() *mocks.S3
		want    udomain.UserCreatePreSignedUrlResponse
		wantErr bool
	}{
		{
			name: "정상적으로 호출",
			args: args{
				ctx: context.Background(),
				request: udomain.UserCreatePreSignedUrlRequest{
					FileName: "test.jpg",
				},
			},

			s3pkg: func() *mocks.S3 {
				s3pkg := new(mocks.S3)
				s3pkg.On("CreatePreSignedUrl", context.Background(), s3utils.CreatePreSignedUrlRequest{
					BucketName:         "test-bucket",
					FilePath:           fmt.Sprintf("%s/%s", userFilePath, timeNowDateOnly),
					FileName:           "test.jpg",
					ContentType:        aws.String("image/jpeg"),
					ContentDisposition: aws.String("attachment; filename=\"test.jpg\""),
				}).Return(fmt.Sprintf("https://test-bucket/user_file/%s/test.jpg", timeNowDateOnly), nil)

				return s3pkg
			},
			want: udomain.UserCreatePreSignedUrlResponse{
				PreSignedUrl:       fmt.Sprintf("https://test-bucket/user_file/%s/test.jpg", timeNowDateOnly),
				ContentType:        "image/jpeg",
				ContentDisposition: "attachment; filename=\"test.jpg\"",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := userUsecase{
				cfg: &config.Config{
					CommonConfig: config.CommonConfig{
						AWS: struct {
							Region                  string
							SES_Region              string
							SNS_Region              string
							CDN                     string
							S3_PRIVATE_BUCKET       string
							S3_PUBLIC_BUCKET        string
							S3_EKYC_IMAGES_BUCKET   string
							SES_ARN                 string
							SNS_ARN                 string
							PUBLIC_KEY_ID           string
							PRIVATE_KEY_PATH        string
							BASE_ENDPOINT           string
							INVEST_WORKER_QUEUE_URL string
						}{
							S3_PRIVATE_BUCKET: "test-bucket",
						},
					},
				},
				s3: tt.s3pkg(),
			}

			got, err := usecase.CreatePreSignedUrl(tt.args.ctx, tt.args.request)
			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			assert.Equalf(t, tt.wantErr, err != nil, "CreatePreSignedUrl() error = %v, wantErr %v", err, tt.wantErr)
			assert.Truef(t, cmp.Equal(tt.want, got, ignoreTime), "CreatePreSignedUrl() = %v, want %v", got, tt.want)
		})
	}
}

func Test_userUsecase_RequestSignUp(t *testing.T) {
	type args struct {
		ctx     context.Context
		request udomain.UserRequestSignUpRequest
	}

	tests := []struct {
		name                    string
		args                    args
		egift                   func() *mocks.Egift
		mailService             func() *mocks.MailService
		userService             func() *mocks.UserService
		signUpRequestRepository func() *mocks.SignUpRequestRepository
		marketingCodeRepository func() *mocks.MarketingCodeRepository
		codeOwnershipRepository func() *mocks.CodeOwnershipRepository
		codeUsageRepository     func() *mocks.CodeUsageRepository
		generatedReferralCode   string
		wantErr                 bool
	}{
		{
			name: "Success: 会員登録リクエスト 既存データなし",
			args: args{
				ctx: context.Background(),
				request: udomain.UserRequestSignUpRequest{
					Email:    "gildong@email.com",
					UserType: "PRIVATE",
				},
			},
			userService: func() *mocks.UserService {
				userService := new(mocks.UserService)
				userService.On("IsDuplicateEmailIncludeDeleted", context.Background(), "gildong@email.com").Return(false, nil)
				return userService
			},
			signUpRequestRepository: func() *mocks.SignUpRequestRepository {
				signUpRequestRepository := new(mocks.SignUpRequestRepository)
				// checkSignUpRequests
				signUpRequestRepository.On("GetAll", context.Background(),
					mock.Anything,
					mock.MatchedBy(func(field func(fields *cdomain.SignUpRequestQueryOptionFields)) bool {
						opt := cdomain.SignUpRequestQueryOptionFunc()
						field(&opt)

						return opt.UrlExpire() != nil
					}),
					mock.Anything,
					mock.Anything,
				).Return(cdomain.SignUpRequests{}, nil)
				// SignUpRequests
				signUpRequestRepository.On("GetAll", context.Background(),
					mock.Anything,
					mock.MatchedBy(func(field func(fields *cdomain.SignUpRequestQueryOptionFields)) bool {
						opt := cdomain.SignUpRequestQueryOptionFunc()
						field(&opt)

						return opt.EmailExpire() != nil
					}),
					mock.Anything,
					mock.Anything,
				).Return(cdomain.SignUpRequests{}, nil)

				signUpRequestRepository.On("Create", context.Background(), mock.MatchedBy(func(request domain.SignUpRequest) bool {
					return request.Email == "gildong@email.com" &&
						strings.Contains(request.Url, "/sign-up/")
				})).Return(int64(1), nil)

				return signUpRequestRepository
			},
			mailService: func() *mocks.MailService {
				mailService := new(mocks.MailService)
				mailService.On("SendMail", context.Background(), mock.MatchedBy(func(cdomain.SendMailRequest) bool {
					return true
				})).Return(nil)

				return mailService
			},

			marketingCodeRepository: func() *mocks.MarketingCodeRepository {
				marketingCodeRepository := new(mocks.MarketingCodeRepository)
				return marketingCodeRepository
			},

			wantErr: false,
		},
		{
			name: "Success: 会員登録リクエスト 既存データなし(紹介コードの入力あり)",
			args: args{
				ctx: context.Background(),
				request: udomain.UserRequestSignUpRequest{
					Email:        "referral@email.com",
					UserType:     "PRIVATE",
					ReferralCode: "REFCODE123",
				},
			},
			userService: func() *mocks.UserService {
				userService := new(mocks.UserService)
				userService.On("IsDuplicateEmailIncludeDeleted", context.Background(), "referral@email.com").Return(false, nil)
				return userService
			},
			signUpRequestRepository: func() *mocks.SignUpRequestRepository {
				repo := new(mocks.SignUpRequestRepository)
				repo.On("GetAll", context.Background(),
					mock.Anything,
					mock.MatchedBy(func(field func(fields *cdomain.SignUpRequestQueryOptionFields)) bool {
						opt := cdomain.SignUpRequestQueryOptionFunc()
						field(&opt)

						return opt.UrlExpire() != nil
					}),
					mock.Anything,
					mock.Anything,
				).Return(cdomain.SignUpRequests{}, nil)

				repo.On("GetAll", context.Background(),
					mock.Anything,
					mock.MatchedBy(func(field func(fields *cdomain.SignUpRequestQueryOptionFields)) bool {
						opt := cdomain.SignUpRequestQueryOptionFunc()
						field(&opt)

						return opt.EmailExpire() != nil
					}),
					mock.Anything,
					mock.Anything,
				).Return(cdomain.SignUpRequests{}, nil)

				repo.On("Create", context.Background(), mock.MatchedBy(func(request domain.SignUpRequest) bool {
					return request.ReferralCode == "REFCODE123"
				})).Return(int64(1), nil)

				return repo
			},
			mailService: func() *mocks.MailService {
				mailService := new(mocks.MailService)
				mailService.On("SendMail", context.Background(), mock.Anything).Return(nil)
				return mailService
			},
			marketingCodeRepository: func() *mocks.MarketingCodeRepository {
				return new(mocks.MarketingCodeRepository)
			},
			wantErr: false,
		},
		{
			name: "Success: 会員登録リクエスト 既存データなし(マーケティングコードの入力あり)",
			args: args{
				ctx: context.Background(),
				request: udomain.UserRequestSignUpRequest{
					Email:             "gildong@email.com",
					UserType:          "PRIVATE",
					MarketingCode:     "marketing-code",
					MarketingCampaign: "marketing-campaign",
				},
			},
			marketingCodeRepository: func() *mocks.MarketingCodeRepository {
				marketingCodeRepository := new(mocks.MarketingCodeRepository)
				marketingCodeRepository.On("Create", context.Background(), mock.MatchedBy(func(marketingCode domain.MarketingCode) bool {
					ignoreFields := cmpopts.IgnoreFields(domain.MarketingCode{}, "SignUpRequestId")
					ignoreTypes := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

					return cmp.Equal(marketingCode, domain.MarketingCode{
						Code:     "marketing-code",
						Campaign: "marketing-campaign",
						Email:    "gildong@email.com",
					}, ignoreTypes, ignoreFields)

				})).Return(int64(1), nil)

				return marketingCodeRepository
			},

			userService: func() *mocks.UserService {
				userService := new(mocks.UserService)
				userService.On("IsDuplicateEmailIncludeDeleted", context.Background(), "gildong@email.com").Return(false, nil)
				return userService
			},
			signUpRequestRepository: func() *mocks.SignUpRequestRepository {
				signUpRequestRepository := new(mocks.SignUpRequestRepository)
				signUpRequestRepository.On("GetAll", context.Background(),
					mock.Anything,
					mock.MatchedBy(func(field func(fields *cdomain.SignUpRequestQueryOptionFields)) bool {
						opt := cdomain.SignUpRequestQueryOptionFunc()
						field(&opt)

						return opt.UrlExpire() != nil
					}),
					mock.Anything,
					mock.Anything,
				).Return(cdomain.SignUpRequests{}, nil)

				// SignUpRequests
				signUpRequestRepository.On("GetAll", context.Background(),
					mock.Anything,
					mock.MatchedBy(func(field func(fields *cdomain.SignUpRequestQueryOptionFields)) bool {
						opt := cdomain.SignUpRequestQueryOptionFunc()
						field(&opt)

						return opt.EmailExpire() != nil
					}),
					mock.Anything,
					mock.Anything,
				).Return(cdomain.SignUpRequests{}, nil)

				signUpRequestRepository.On("Create", context.Background(), mock.MatchedBy(func(request domain.SignUpRequest) bool {
					return request.Email == "gildong@email.com" &&
						strings.Contains(request.Url, "/sign-up/")
				})).Return(int64(1), nil)

				return signUpRequestRepository
			},
			mailService: func() *mocks.MailService {
				mailService := new(mocks.MailService)
				mailService.On("SendMail", context.Background(), mock.MatchedBy(func(cdomain.SendMailRequest) bool {
					return true
				})).Return(nil)

				return mailService
			},
			codeOwnershipRepository: func() *mocks.CodeOwnershipRepository {
				repo := new(mocks.CodeOwnershipRepository)
				repo.On("GetByCode", mock.Anything, domain.CodeTypeReferral, "12345ABCDE").Return((*domain.CodeOwnership)(nil), nil)
				repo.On("Create", mock.Anything, mock.MatchedBy(func(code domain.CodeOwnership) bool {
					return code.OwnerId == 1 && code.OwnerType == domain.PRIVATE && code.CodeType == domain.CodeTypeReferral && code.Code == "12345ABCDE"
				})).Return(nil)
				return repo
			},
			codeUsageRepository: func() *mocks.CodeUsageRepository {
				repo := new(mocks.CodeUsageRepository)
				repo.On("Create", mock.Anything, mock.MatchedBy(func(usage domain.CodeUsage) bool {
					return usage.UsedOwnerId == 1 && usage.UsedOwnerType == domain.CORPORATE && usage.CodeOwnershipId == 99 && usage.CodeInputAt != nil
				})).Return(nil)
				return repo
			},
			generatedReferralCode: "12345ABCDE",
			wantErr:               false,
		},
		{
			name: "Success: 会員登録リクエストレコード 既存データあり(メール送信回数超過なし)",
			args: args{
				ctx: context.Background(),
				request: udomain.UserRequestSignUpRequest{
					Email:    "gildong@email.com",
					UserType: "PRIVATE",
				},
			},
			userService: func() *mocks.UserService {
				userService := new(mocks.UserService)
				userService.On("IsDuplicateEmailIncludeDeleted", context.Background(), "gildong@email.com").Return(false, nil)
				return userService
			},
			signUpRequestRepository: func() *mocks.SignUpRequestRepository {
				signUpRequestRepository := new(mocks.SignUpRequestRepository)
				// checkSignUpRequests
				signUpRequestRepository.On("GetAll", context.Background(),
					mock.Anything,
					mock.MatchedBy(func(field func(fields *cdomain.SignUpRequestQueryOptionFields)) bool {
						opt := cdomain.SignUpRequestQueryOptionFunc()
						field(&opt)

						return opt.UrlExpire() != nil
					}),
					mock.Anything,
					mock.Anything,
				).Return(cdomain.SignUpRequests{}, nil)
				// SignUpRequests
				signUpRequestRepository.On("GetAll", context.Background(),
					mock.Anything,
					mock.MatchedBy(func(field func(fields *cdomain.SignUpRequestQueryOptionFields)) bool {
						opt := cdomain.SignUpRequestQueryOptionFunc()
						field(&opt)

						return opt.EmailExpire() != nil
					}),
					mock.Anything,
					mock.Anything,
				).Return(cdomain.SignUpRequests{
					{
						SendCount: 0,
					},
				}, nil)

				signUpRequestRepository.On("Update", context.Background(), mock.MatchedBy(func(request domain.SignUpRequest) bool {
					return request.SendCount == 1
				})).Return(nil)

				return signUpRequestRepository
			},
			mailService: func() *mocks.MailService {
				mailService := new(mocks.MailService)
				mailService.On("SendMail", context.Background(), mock.MatchedBy(func(cdomain.SendMailRequest) bool {
					return true
				})).Return(nil)

				return mailService
			},
			marketingCodeRepository: func() *mocks.MarketingCodeRepository {
				marketingCodeRepository := new(mocks.MarketingCodeRepository)
				return marketingCodeRepository
			},
			wantErr: false,
		},
		{
			name: "Success: 既存の会員登録リクエストの紹介コードの更新確認",
			args: args{
				ctx: context.Background(),
				request: udomain.UserRequestSignUpRequest{
					Email:        "update@email.com",
					UserType:     "PRIVATE",
					ReferralCode: "UPDATECODE",
				},
			},
			userService: func() *mocks.UserService {
				userService := new(mocks.UserService)
				userService.On("IsDuplicateEmailIncludeDeleted", context.Background(), "update@email.com").Return(false, nil)
				return userService
			},
			signUpRequestRepository: func() *mocks.SignUpRequestRepository {
				repo := new(mocks.SignUpRequestRepository)
				repo.On("GetAll", context.Background(),
					mock.Anything,
					mock.MatchedBy(func(field func(fields *cdomain.SignUpRequestQueryOptionFields)) bool {
						opt := cdomain.SignUpRequestQueryOptionFunc()
						field(&opt)

						return opt.UrlExpire() != nil
					}),
					mock.Anything,
					mock.Anything,
				).Return(cdomain.SignUpRequests{}, nil)

				repo.On("GetAll", context.Background(),
					mock.Anything,
					mock.MatchedBy(func(field func(fields *cdomain.SignUpRequestQueryOptionFields)) bool {
						opt := cdomain.SignUpRequestQueryOptionFunc()
						field(&opt)

						return opt.EmailExpire() != nil
					}),
					mock.Anything,
					mock.Anything,
				).Return(cdomain.SignUpRequests{
					{
						RequestId: "existing-request-id",
						SendCount: 0,
					},
				}, nil)

				repo.On("Update", context.Background(), mock.MatchedBy(func(request domain.SignUpRequest) bool {
					return request.RequestId == "existing-request-id" && request.ReferralCode == "UPDATECODE" && request.SendCount == 1
				})).Return(nil)

				return repo
			},
			mailService: func() *mocks.MailService {
				mailService := new(mocks.MailService)
				mailService.On("SendMail", context.Background(), mock.Anything).Return(nil)
				return mailService
			},
			marketingCodeRepository: func() *mocks.MarketingCodeRepository {
				return new(mocks.MarketingCodeRepository)
			},
			wantErr: false,
		},
		// 異常系
		{
			name: "Failed: メールアドレス重複状態取得失敗",
			args: args{
				ctx: context.Background(),
				request: udomain.UserRequestSignUpRequest{
					Email:    "gildong@email.com",
					UserType: "PRIVATE",
				},
			},
			userService: func() *mocks.UserService {
				userService := new(mocks.UserService)
				userService.On("IsDuplicateEmailIncludeDeleted", context.Background(), "gildong@email.com").Return(false, errors.New("service error"))
				return userService
			},
			signUpRequestRepository: func() *mocks.SignUpRequestRepository {
				return new(mocks.SignUpRequestRepository)
			},
			mailService: func() *mocks.MailService {
				return new(mocks.MailService)
			},
			marketingCodeRepository: func() *mocks.MarketingCodeRepository {
				return new(mocks.MarketingCodeRepository)
			},
			wantErr: true,
		},
		{
			name: "Failed: 既に登録済みのメールアドレス",
			args: args{
				ctx: context.Background(),
				request: udomain.UserRequestSignUpRequest{
					Email:    "gildong@email.com",
					UserType: "PRIVATE",
				},
			},
			userService: func() *mocks.UserService {
				userService := new(mocks.UserService)
				userService.On("IsDuplicateEmailIncludeDeleted", context.Background(), "gildong@email.com").Return(true, nil)
				return userService
			},
			signUpRequestRepository: func() *mocks.SignUpRequestRepository {
				return new(mocks.SignUpRequestRepository)
			},
			mailService: func() *mocks.MailService {
				return new(mocks.MailService)
			},
			marketingCodeRepository: func() *mocks.MarketingCodeRepository {
				return new(mocks.MarketingCodeRepository)
			},
			wantErr: true,
		},
		{
			name: "Failed: メール送信回数超過",
			args: args{
				ctx: context.Background(),
				request: udomain.UserRequestSignUpRequest{
					Email:    "gildong@email.com",
					UserType: "PRIVATE",
				},
			},
			userService: func() *mocks.UserService {
				userService := new(mocks.UserService)
				userService.On("IsDuplicateEmailIncludeDeleted", context.Background(), "gildong@email.com").Return(false, nil)
				return userService
			},
			signUpRequestRepository: func() *mocks.SignUpRequestRepository {
				signUpRequestRepository := new(mocks.SignUpRequestRepository)
				// checkSignUpRequests
				signUpRequestRepository.On("GetAll", context.Background(),
					mock.Anything,
					mock.MatchedBy(func(field func(fields *cdomain.SignUpRequestQueryOptionFields)) bool {
						opt := cdomain.SignUpRequestQueryOptionFunc()
						field(&opt)

						return opt.UrlExpire() != nil
					}),
					mock.Anything,
					mock.Anything,
				).Return(cdomain.SignUpRequests{}, nil)
				// SignUpRequests
				signUpRequestRepository.On("GetAll", context.Background(),
					mock.Anything,
					mock.MatchedBy(func(field func(fields *cdomain.SignUpRequestQueryOptionFields)) bool {
						opt := cdomain.SignUpRequestQueryOptionFunc()
						field(&opt)

						return opt.EmailExpire() != nil
					}),
					mock.Anything,
					mock.Anything,
				).Return(cdomain.SignUpRequests{
					{
						SendCount: 1, // メール送信履歴が1回ある
					},
				}, nil)

				return signUpRequestRepository
			},
			mailService: func() *mocks.MailService {
				mailService := new(mocks.MailService)
				mailService.On("SendMail", context.Background(), mock.MatchedBy(func(cdomain.SendMailRequest) bool {
					return true
				})).Return(nil)

				return mailService
			},
			marketingCodeRepository: func() *mocks.MarketingCodeRepository {
				marketingCodeRepository := new(mocks.MarketingCodeRepository)
				return marketingCodeRepository
			},
			wantErr: true,
		},
		{
			name: "Failed: 会員登録リクエスト取得失敗",
			args: args{
				ctx: context.Background(),
				request: udomain.UserRequestSignUpRequest{
					Email:    "gildong@email.com",
					UserType: "PRIVATE",
				},
			},
			userService: func() *mocks.UserService {
				userService := new(mocks.UserService)
				userService.On("IsDuplicateEmailIncludeDeleted", context.Background(), "gildong@email.com").Return(false, nil)
				return userService
			},
			signUpRequestRepository: func() *mocks.SignUpRequestRepository {
				repo := new(mocks.SignUpRequestRepository)
				repo.On("GetAll", context.Background(), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(cdomain.SignUpRequests{}, errors.New("get all error"))
				return repo
			},
			mailService: func() *mocks.MailService {
				return new(mocks.MailService)
			},
			marketingCodeRepository: func() *mocks.MarketingCodeRepository {
				return new(mocks.MarketingCodeRepository)
			},
			wantErr: true,
		},
		{
			name: "Failed: 会員登録リクエスト作成失敗",
			args: args{
				ctx: context.Background(),
				request: udomain.UserRequestSignUpRequest{
					Email:    "gildong@email.com",
					UserType: "PRIVATE",
				},
			},
			userService: func() *mocks.UserService {
				userService := new(mocks.UserService)
				userService.On("IsDuplicateEmailIncludeDeleted", context.Background(), "gildong@email.com").Return(false, nil)
				return userService
			},
			signUpRequestRepository: func() *mocks.SignUpRequestRepository {
				repo := new(mocks.SignUpRequestRepository)
				repo.On("GetAll", context.Background(), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(cdomain.SignUpRequests{}, nil)
				repo.On("Create", context.Background(), mock.Anything).Return(int64(0), errors.New("create error"))
				return repo
			},
			mailService: func() *mocks.MailService {
				return new(mocks.MailService)
			},
			marketingCodeRepository: func() *mocks.MarketingCodeRepository {
				return new(mocks.MarketingCodeRepository)
			},
			wantErr: true,
		},
		{
			name: "Failed: egiftリワード処理失敗",
			args: args{
				ctx: context.Background(),
				request: udomain.UserRequestSignUpRequest{
					Email:             "gildong@email.com",
					UserType:          "PRIVATE",
					MarketingCode:     "marketing-code",
					MarketingCampaign: "marketing-campaign",
				},
			},
			egift: func() *mocks.Egift {
				egift := new(mocks.Egift)
				egift.On("RewardSignUp", mock.Anything, "marketing-code", mock.Anything).Return(errors.New("reward error"))
				return egift
			},
			userService: func() *mocks.UserService {
				userService := new(mocks.UserService)
				userService.On("IsDuplicateEmailIncludeDeleted", context.Background(), "gildong@email.com").Return(false, nil)
				return userService
			},
			signUpRequestRepository: func() *mocks.SignUpRequestRepository {
				repo := new(mocks.SignUpRequestRepository)
				repo.On("GetAll", context.Background(), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(cdomain.SignUpRequests{}, nil)
				repo.On("Create", context.Background(), mock.Anything).Return(int64(1), nil)
				return repo
			},
			mailService: func() *mocks.MailService {
				return new(mocks.MailService)
			},
			marketingCodeRepository: func() *mocks.MarketingCodeRepository {
				return new(mocks.MarketingCodeRepository)
			},
			wantErr: true,
		},
		{
			name: "Failed: マーケティングコード生成失敗",
			args: args{
				ctx: context.Background(),
				request: udomain.UserRequestSignUpRequest{
					Email:             "gildong@email.com",
					UserType:          "PRIVATE",
					MarketingCode:     "marketing-code",
					MarketingCampaign: "marketing-campaign",
				},
			},
			egift: func() *mocks.Egift {
				egift := new(mocks.Egift)
				egift.On("RewardSignUp", mock.Anything, "marketing-code", mock.Anything).Return(nil)
				return egift
			},
			userService: func() *mocks.UserService {
				userService := new(mocks.UserService)
				userService.On("IsDuplicateEmailIncludeDeleted", context.Background(), "gildong@email.com").Return(false, nil)
				return userService
			},
			signUpRequestRepository: func() *mocks.SignUpRequestRepository {
				repo := new(mocks.SignUpRequestRepository)
				repo.On("GetAll", context.Background(), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(cdomain.SignUpRequests{}, nil)
				repo.On("Create", context.Background(), mock.Anything).Return(int64(1), nil)
				return repo
			},
			mailService: func() *mocks.MailService {
				return new(mocks.MailService)
			},
			marketingCodeRepository: func() *mocks.MarketingCodeRepository {
				repo := new(mocks.MarketingCodeRepository)
				repo.On("Create", context.Background(), mock.Anything).Return(int64(0), errors.New("marketing error"))
				return repo
			},
			wantErr: true,
		},
		{
			name: "Failed: 会員登録リクエスト更新失敗",
			args: args{
				ctx: context.Background(),
				request: udomain.UserRequestSignUpRequest{
					Email:    "gildong@email.com",
					UserType: "PRIVATE",
				},
			},
			userService: func() *mocks.UserService {
				userService := new(mocks.UserService)
				userService.On("IsDuplicateEmailIncludeDeleted", context.Background(), "gildong@email.com").Return(false, nil)
				return userService
			},
			signUpRequestRepository: func() *mocks.SignUpRequestRepository {
				repo := new(mocks.SignUpRequestRepository)
				repo.On("GetAll", context.Background(), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(cdomain.SignUpRequests{
					{
						RequestId: "request-id",
						SendCount: 0,
					},
				}, nil)
				repo.On("Update", context.Background(), mock.MatchedBy(func(req domain.SignUpRequest) bool {
					return req.RequestId == "request-id" && req.SendCount == 1
				})).Return(errors.New("update error"))
				return repo
			},
			mailService: func() *mocks.MailService {
				return new(mocks.MailService)
			},
			marketingCodeRepository: func() *mocks.MarketingCodeRepository {
				return new(mocks.MarketingCodeRepository)
			},
			wantErr: true,
		},
		{
			name: "Failed: メール送信失敗",
			args: args{
				ctx: context.Background(),
				request: udomain.UserRequestSignUpRequest{
					Email:    "gildong@email.com",
					UserType: "PRIVATE",
				},
			},
			userService: func() *mocks.UserService {
				userService := new(mocks.UserService)
				userService.On("IsDuplicateEmailIncludeDeleted", context.Background(), "gildong@email.com").Return(false, nil)
				return userService
			},
			signUpRequestRepository: func() *mocks.SignUpRequestRepository {
				repo := new(mocks.SignUpRequestRepository)
				repo.On("GetAll", context.Background(), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(cdomain.SignUpRequests{}, nil)
				repo.On("Create", context.Background(), mock.Anything).Return(int64(1), nil)
				return repo
			},
			mailService: func() *mocks.MailService {
				mailService := new(mocks.MailService)
				mailService.On("SendMail", context.Background(), mock.Anything).Return(errors.New("mail error"))
				return mailService
			},
			marketingCodeRepository: func() *mocks.MarketingCodeRepository {
				return new(mocks.MarketingCodeRepository)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			egiftService := egift.New(&config.Config{
				CommonConfig: config.CommonConfig{
					AccessTradeService: config.AccessTradeConfig{
						Stub: true,
					},
				},
			}, factory.NewLoggerFactory())

			if tt.egift != nil {
				egiftService = tt.egift()
			}

			usecase := userUsecase{
				idGenerator: idgenerator.New(),
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
				egift:                   egiftService,
				marketingCodeRepository: tt.marketingCodeRepository(),
				userService:             tt.userService(),
				signUpRequestRepository: tt.signUpRequestRepository(),
				mailService:             tt.mailService(),
			}

			_, err := usecase.RequestSignUp(tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.wantErr, err != nil, "RequestSignUp() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_userUsecase_VerifyEmail(t *testing.T) {
	type args struct {
		ctx     context.Context
		request udomain.UserVerifyEmailRequest
	}

	tests := []struct {
		name                    string
		args                    args
		signUpRequestRepository func() *mocks.SignUpRequestRepository
		wantErr                 bool
	}{
		{
			name: "会員登録要求が有効期限切れです。",
			args: args{
				ctx: context.Background(),
				request: udomain.UserVerifyEmailRequest{
					RequestId: "test-request-id",
				},
			},
			signUpRequestRepository: func() *mocks.SignUpRequestRepository {
				signUpRequestRepository := new(mocks.SignUpRequestRepository)
				signUpRequestRepository.On("GetByRequestId", context.Background(), "test-request-id").Return(domain.SignUpRequest{
					EmailExpire: jsonutils.JSONTime(time.Now().Add(-1 * time.Hour)),
					UrlExpire:   jsonutils.JSONTime(time.Now().Add(-1 * time.Hour)),
				}, nil)

				return signUpRequestRepository
			},
			wantErr: true,
		},
		{
			name: "会員登録要求がありません。",
			args: args{
				ctx: context.Background(),
				request: udomain.UserVerifyEmailRequest{
					RequestId: "test-request-id",
				},
			},
			signUpRequestRepository: func() *mocks.SignUpRequestRepository {
				signUpRequestRepository := new(mocks.SignUpRequestRepository)
				signUpRequestRepository.On("GetByRequestId", context.Background(), "test-request-id").Return(domain.SignUpRequest{}, cdomain.ErrItemNotFound)

				return signUpRequestRepository
			},
			wantErr: true,
		},
		{
			name: "正常に認証を進めます。",
			args: args{
				ctx: context.Background(),
				request: udomain.UserVerifyEmailRequest{
					RequestId: "test-request-id",
				},
			},
			signUpRequestRepository: func() *mocks.SignUpRequestRepository {
				signUpRequestRepository := new(mocks.SignUpRequestRepository)
				signUpRequestRepository.On("GetByRequestId", context.Background(), "test-request-id").Return(domain.SignUpRequest{
					EmailExpire: jsonutils.JSONTime(time.Now().Add(1 * time.Hour)),
					UrlExpire:   jsonutils.JSONTime(time.Now().Add(1 * time.Hour)),
				}, nil)

				signUpRequestRepository.On("Update", context.Background(), mock.MatchedBy(func(request domain.SignUpRequest) bool {
					return !request.UrlExpire.IsEmpty() && !request.EmailExpire.IsEmpty()
				})).Return(nil)

				return signUpRequestRepository
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := userUsecase{
				signUpRequestRepository: tt.signUpRequestRepository(),
			}

			err := usecase.VerifyEmail(tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.wantErr, err != nil, "VerifyEmail() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_userUsecase_CheckSignupRequestValidity(t *testing.T) {
	type args struct {
		ctx     context.Context
		request udomain.UserCheckSignupRequestValidityRequest
	}

	tests := []struct {
		name                    string
		args                    args
		signUpRequestRepository func() *mocks.SignUpRequestRepository
		want                    udomain.UserCheckSignupRequestValidityResponse
		wantErr                 bool
	}{
		{
			name: "メール検証が完了していない状態でメールが有効期限切れになった時",
			args: args{
				ctx: context.Background(),
				request: udomain.UserCheckSignupRequestValidityRequest{
					RequestId: "test-request-id",
				},
			},
			signUpRequestRepository: func() *mocks.SignUpRequestRepository {
				signUpRequestRepository := new(mocks.SignUpRequestRepository)
				signUpRequestRepository.On("GetByRequestId", context.Background(), "test-request-id").Return(domain.SignUpRequest{
					EmailExpire: jsonutils.JSONTime(time.Now().Add(-1 * time.Hour)),
				}, nil)

				return signUpRequestRepository
			},
			want:    udomain.UserCheckSignupRequestValidityResponse{},
			wantErr: true,
		},
		{
			name: "メール検証が完了されたがURLが有効期限切れになった時",
			args: args{
				ctx: context.Background(),
				request: udomain.UserCheckSignupRequestValidityRequest{
					RequestId: "test-request-id",
				},
			},
			signUpRequestRepository: func() *mocks.SignUpRequestRepository {
				signUpRequestRepository := new(mocks.SignUpRequestRepository)
				signUpRequestRepository.On("GetByRequestId", context.Background(), "test-request-id").Return(domain.SignUpRequest{
					EmailExpire: jsonutils.JSONTime(time.Now().Add(1 * time.Hour)),
					UrlExpire:   jsonutils.JSONTime(time.Now().Add(-1 * time.Hour)),
				}, nil)

				return signUpRequestRepository
			},
			want:    udomain.UserCheckSignupRequestValidityResponse{},
			wantErr: true,
		},
		{
			name: "正常呼び出し",
			args: args{
				ctx: context.Background(),
				request: udomain.UserCheckSignupRequestValidityRequest{
					RequestId: "test-request-id",
				},
			},
			signUpRequestRepository: func() *mocks.SignUpRequestRepository {
				signUpRequestRepository := new(mocks.SignUpRequestRepository)
				signUpRequestRepository.On("GetByRequestId", context.Background(), "test-request-id").Return(domain.SignUpRequest{
					Email:       "test@example.com",
					UserType:    domain.PRIVATE,
					EmailExpire: jsonutils.JSONTime(time.Now().Add(1 * time.Hour)),
					UrlExpire:   jsonutils.JSONTime(time.Now().Add(1 * time.Hour)),
				}, nil)

				return signUpRequestRepository
			},
			want: udomain.UserCheckSignupRequestValidityResponse{
				Email:    "test@example.com",
				UserType: domain.PRIVATE,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := userUsecase{
				signUpRequestRepository: tt.signUpRequestRepository(),
			}

			got, err := usecase.CheckSignupRequestValidity(tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.wantErr, err != nil, "CheckSignupRequestValidity() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want, got, "CheckSignupRequestValidity() = %v, want %v", got, tt.want)
		})
	}
}

func Test_userUsecase_SignUp(t *testing.T) {
	uid, _ := uuid.FromString("bddad349-d0cc-4a1d-8843-431d54fcd2c7")

	type args struct {
		ctx     context.Context
		request udomain.UserSignUpRequest
	}

	tests := []struct {
		name                          string
		args                          args
		idGenerator                   func() *mocks.IdGenerator
		pwhelper                      func() *mocks.PasswordHelper
		userRepository                func() *mocks.UserRepository
		roleRepository                func() *mocks.RoleRepository
		userAntisocialCheckRepository func() *mocks.UserAntisocialCheckRepository
		userAccountRepository         func() *mocks.UserAccountRepository
		signUpRequestRepository       func() *mocks.SignUpRequestRepository
		privateProfileRepository      func() *mocks.PrivateProfileRepository
		corporateProfileRepository    func() *mocks.CorporateProfileRepository
		userService                   func() *mocks.UserService
		mailService                   func() *mocks.MailService
		codeOwnershipRepository       func() *mocks.CodeOwnershipRepository
		codeUsageRepository           func() *mocks.CodeUsageRepository
		generatedReferralCode         string
		wantErr                       bool
	}{
		{
			name: "(個人会員) 正常に会員登録を進めます。",
			args: args{
				ctx: context.Background(),
				request: udomain.UserSignUpRequest{
					RequestId:      "test-request-id",
					UserType:       "PRIVATE",
					Email:          "gildong@email.com",
					Password:       "test-password",
					IsNewsLetter:   false,
					FirstName:      "ギルドン",
					FamilyName:     "ホン",
					FirstNameKana:  "ギルドン",
					FamilyNameKana: "ホン",
				},
			},
			idGenerator: func() *mocks.IdGenerator {
				gen := new(mocks.IdGenerator)
				gen.On("GenUuidV4").Return(uid)
				gen.On("GenUuidV5", uid, "gildong@email.com").Return(uid)
				gen.On("GenUID", int64(1)).Return("10000001")
				return gen
			},
			pwhelper: func() *mocks.PasswordHelper {
				pwhelper := new(mocks.PasswordHelper)
				pwhelper.On("EncodePassword", "test-password", mock.Anything).Return("test-password", nil)
				pwhelper.On("GenerateSalt").Return("salt")
				return pwhelper
			},
			userRepository: func() *mocks.UserRepository {
				userRepository := new(mocks.UserRepository)
				userRepository.On("Create", mock.Anything, mock.MatchedBy(func(user domain.User) bool {
					return cmp.Equal(domain.User{
						UserType:     "PRIVATE",
						Email:        "gildong@email.com",
						Name:         "ホン ギルドン",
						Password:     "test-password",
						SignupStatus: domain.SIGNUP_NONE,
						LoginType:    domain.MAIL,
						AuditStatus:  domain.PROVISIONAL_AUDIT,
					}, user,
						cmpopts.IgnoreFields(domain.User{}, "Salt", "CommonId", "Uid"),
						cmpopts.IgnoreTypes(jsonutils.JSONTime{}))
				})).Return(int64(1), nil)

				userRepository.On("Update", mock.Anything, mock.MatchedBy(func(user domain.User) bool {
					return cmp.Equal(domain.User{
						Id:           int64(1),
						Uid:          "10000001",
						UserType:     "PRIVATE",
						Email:        "gildong@email.com",
						Name:         "ホン ギルドン",
						Password:     "test-password",
						SignupStatus: domain.SIGNUP_NONE,
						LoginType:    domain.MAIL,
						AuditStatus:  domain.PROVISIONAL_AUDIT,
						InvestorType: domain.PRIVATE_INVESTOR,
					}, user,
						cmpopts.IgnoreFields(domain.User{}, "Salt", "CommonId"),
						cmpopts.IgnoreTypes(jsonutils.JSONTime{}))
				})).Return(nil)

				return userRepository
			},

			roleRepository: func() *mocks.RoleRepository {
				roleRepository := new(mocks.RoleRepository)
				roleRepository.On("Create", mock.Anything, domain.Role{
					UserId:   int64(1),
					RoleCode: string(domain.ROLE_USER),
				}).Return(int64(1), nil)

				return roleRepository
			},

			userAntisocialCheckRepository: func() *mocks.UserAntisocialCheckRepository {
				userAntisocialCheckRepository := new(mocks.UserAntisocialCheckRepository)
				userAntisocialCheckRepository.On("Create", mock.Anything, domain.UserAntisocialCheck{
					CheckUserId: int64(1),
				}).Return(nil)
				return userAntisocialCheckRepository
			},

			userAccountRepository: func() *mocks.UserAccountRepository {
				userAccountRepository := new(mocks.UserAccountRepository)
				userAccountRepository.On("Create", mock.Anything, domain.UserAccount{
					UserId: int64(1),
				}).Return(nil)

				return userAccountRepository
			},
			userService: func() *mocks.UserService {
				userService := new(mocks.UserService)
				userService.On("EncodePassword", context.Background(), "test-password", mock.MatchedBy(func(salt string) bool {
					return len(salt) == 10
				})).Return("test-password", nil)
				return userService
			},
			signUpRequestRepository: func() *mocks.SignUpRequestRepository {
				signUpRequestRepository := new(mocks.SignUpRequestRepository)
				signUpRequestRepository.On("GetByRequestId", context.Background(), "test-request-id").Return(domain.SignUpRequest{
					Email:       "gildong@email.com",
					EmailExpire: jsonutils.JSONTime(time.Now().Add(1 * time.Hour)),
				}, nil)

				signUpRequestRepository.On("Update", context.Background(), mock.MatchedBy(func(request domain.SignUpRequest) bool {
					return time.Now().After(request.UrlExpire.Time())
				})).Return(nil)

				return signUpRequestRepository
			},
			privateProfileRepository: func() *mocks.PrivateProfileRepository {
				privateProfileRepository := new(mocks.PrivateProfileRepository)
				privateProfileRepository.On("Create", context.Background(), mock.MatchedBy(func(profile domain.PrivateProfile) bool {
					return cmp.Equal(domain.PrivateProfile{
						UserId:         int64(1),
						FirstName:      "ギルドン",
						FamilyName:     "ホン",
						FirstNameKana:  "ギルドン",
						FamilyNameKana: "ホン",
						RequestId:      "test-request-id",
					}, profile, cmpopts.IgnoreTypes(jsonutils.JSONTime{}))
				})).Return(nil)

				return privateProfileRepository
			},

			corporateProfileRepository: func() *mocks.CorporateProfileRepository {
				corporateProfileRepository := new(mocks.CorporateProfileRepository)
				return corporateProfileRepository
			},

			mailService: func() *mocks.MailService {
				mailService := new(mocks.MailService)
				mailService.On("SendMail", context.Background(), mock.MatchedBy(func(cdomain.SendMailRequest) bool {
					return true
				})).Return(nil)

				return mailService
			},
			codeOwnershipRepository: func() *mocks.CodeOwnershipRepository {
				repo := new(mocks.CodeOwnershipRepository)
				repo.On("GetByCode", mock.Anything, domain.CodeTypeReferral, "67890FGHJK").Return((*domain.CodeOwnership)(nil), nil)
				repo.On("Create", mock.Anything, mock.MatchedBy(func(code domain.CodeOwnership) bool {
					return code.OwnerId == 1 && code.OwnerType == domain.PRIVATE && code.CodeType == domain.CodeTypeReferral && code.Code == "67890FGHJK"
				})).Return(nil)
				return repo
			},
			codeUsageRepository: func() *mocks.CodeUsageRepository {
				return new(mocks.CodeUsageRepository)
			},
			generatedReferralCode: "67890FGHJK",
			wantErr:               false,
		},
		{
			name: "(法人会員) 正常に会員登録を進めます。",
			args: args{
				ctx: context.Background(),
				request: udomain.UserSignUpRequest{
					RequestId:    "test-request-id",
					UserType:     "CORPORATE",
					Email:        "gildong@email.com",
					Password:     "test-password",
					IsNewsLetter: false,
					CorpName:     "illuminarean",
				},
			},
			idGenerator: func() *mocks.IdGenerator {
				gen := new(mocks.IdGenerator)
				gen.On("GenUuidV4").Return(uid)
				gen.On("GenUuidV5", uid, "gildong@email.com").Return(uid)
				gen.On("GenUID", int64(1)).Return("10000001")
				return gen
			},
			pwhelper: func() *mocks.PasswordHelper {
				pwhelper := new(mocks.PasswordHelper)
				pwhelper.On("EncodePassword", "test-password", mock.Anything).Return("test-password", nil)
				pwhelper.On("GenerateSalt").Return("salt")
				return pwhelper
			},
			userRepository: func() *mocks.UserRepository {
				userRepository := new(mocks.UserRepository)
				userRepository.On("Create", mock.Anything, mock.MatchedBy(func(user domain.User) bool {
					return cmp.Equal(domain.User{
						UserType:     "CORPORATE",
						Email:        "gildong@email.com",
						Name:         "illuminarean",
						Password:     "test-password",
						SignupStatus: domain.SIGNUP_NONE,
						LoginType:    domain.MAIL,
						AuditStatus:  domain.PROVISIONAL_AUDIT,
					}, user,
						cmpopts.IgnoreFields(domain.User{}, "Salt", "CommonId", "Uid"),
						cmpopts.IgnoreTypes(jsonutils.JSONTime{}))
				})).Return(int64(1), nil)

				userRepository.On("Update", mock.Anything, mock.MatchedBy(func(user domain.User) bool {
					return cmp.Equal(domain.User{
						Id:           int64(1),
						Uid:          "10000001",
						UserType:     "CORPORATE",
						Email:        "gildong@email.com",
						Name:         "illuminarean",
						Password:     "test-password",
						SignupStatus: domain.SIGNUP_NONE,
						LoginType:    domain.MAIL,
						AuditStatus:  domain.PROVISIONAL_AUDIT,
						InvestorType: domain.CORPORATE_INVESTOR,
					}, user,
						cmpopts.IgnoreFields(domain.User{}, "Salt", "CommonId"),
						cmpopts.IgnoreTypes(jsonutils.JSONTime{}))
				})).Return(nil)

				return userRepository
			},

			roleRepository: func() *mocks.RoleRepository {
				roleRepository := new(mocks.RoleRepository)
				roleRepository.On("Create", mock.Anything, domain.Role{
					UserId:   int64(1),
					RoleCode: string(domain.ROLE_USER),
				}).Return(int64(1), nil)

				return roleRepository
			},

			userAntisocialCheckRepository: func() *mocks.UserAntisocialCheckRepository {
				userAntisocialCheckRepository := new(mocks.UserAntisocialCheckRepository)
				userAntisocialCheckRepository.On("Create", mock.Anything, domain.UserAntisocialCheck{
					CheckUserId: int64(1),
				}).Return(nil)
				return userAntisocialCheckRepository
			},

			userAccountRepository: func() *mocks.UserAccountRepository {
				userAccountRepository := new(mocks.UserAccountRepository)
				userAccountRepository.On("Create", mock.Anything, domain.UserAccount{
					UserId: int64(1),
				}).Return(nil)

				return userAccountRepository
			},

			userService: func() *mocks.UserService {
				userService := new(mocks.UserService)
				userService.On("EncodePassword", context.Background(), "test-password", mock.MatchedBy(func(salt string) bool {
					return len(salt) == 10
				})).Return("test-password", nil)
				return userService
			},
			signUpRequestRepository: func() *mocks.SignUpRequestRepository {
				signUpRequestRepository := new(mocks.SignUpRequestRepository)
				signUpRequestRepository.On("GetByRequestId", context.Background(), "test-request-id").Return(domain.SignUpRequest{
					Email:       "gildong@email.com",
					EmailExpire: jsonutils.JSONTime(time.Now().Add(1 * time.Hour)),
				}, nil)

				signUpRequestRepository.On("Update", context.Background(), mock.MatchedBy(func(request domain.SignUpRequest) bool {
					return time.Now().After(request.UrlExpire.Time())
				})).Return(nil)

				return signUpRequestRepository
			},
			privateProfileRepository: func() *mocks.PrivateProfileRepository {
				privateProfileRepository := new(mocks.PrivateProfileRepository)
				return privateProfileRepository
			},

			corporateProfileRepository: func() *mocks.CorporateProfileRepository {
				corporateProfileRepository := new(mocks.CorporateProfileRepository)
				corporateProfileRepository.On("Create", context.Background(), mock.MatchedBy(func(profile domain.CorporateProfile) bool {
					return cmp.Equal(domain.CorporateProfile{
						UserId: int64(1),
					}, profile, cmpopts.IgnoreTypes(jsonutils.JSONTime{}))
				})).Return(int64(1), nil)
				return corporateProfileRepository
			},

			mailService: func() *mocks.MailService {
				mailService := new(mocks.MailService)
				mailService.On("SendMail", context.Background(), mock.MatchedBy(func(cdomain.SendMailRequest) bool {
					return true
				})).Return(nil)

				return mailService
			},
			codeOwnershipRepository: func() *mocks.CodeOwnershipRepository {
				repo := new(mocks.CodeOwnershipRepository)
				repo.On("GetByCode", mock.Anything, domain.CodeTypeReferral, "CORPABCDE1").Return((*domain.CodeOwnership)(nil), nil)
				repo.On("Create", mock.Anything, mock.MatchedBy(func(code domain.CodeOwnership) bool {
					return code.OwnerId == 1 && code.OwnerType == domain.CORPORATE && code.CodeType == domain.CodeTypeReferral && code.Code == "CORPABCDE1"
				})).Return(nil)
				return repo
			},
			codeUsageRepository: func() *mocks.CodeUsageRepository {
				return new(mocks.CodeUsageRepository)
			},
			generatedReferralCode: "CORPABCDE1",
			wantErr:               false,
		},
		{
			name: "(個人会員) 推奨コードを入力すると使用履歴を保存します。",
			args: args{
				ctx: context.Background(),
				request: udomain.UserSignUpRequest{
					RequestId:      "test-request-id",
					UserType:       "PRIVATE",
					Email:          "gildong@email.com",
					Password:       "test-password",
					IsNewsLetter:   false,
					FirstName:      "ギルドン",
					FamilyName:     "ホン",
					FirstNameKana:  "ギルドン",
					FamilyNameKana: "ホン",
				},
			},
			idGenerator: func() *mocks.IdGenerator {
				gen := new(mocks.IdGenerator)
				gen.On("GenUuidV4").Return(uid)
				gen.On("GenUuidV5", uid, "gildong@email.com").Return(uid)
				gen.On("GenUID", int64(1)).Return("10000001")
				return gen
			},
			pwhelper: func() *mocks.PasswordHelper {
				pwhelper := new(mocks.PasswordHelper)
				pwhelper.On("EncodePassword", "test-password", mock.Anything).Return("test-password", nil)
				pwhelper.On("GenerateSalt").Return("salt")
				return pwhelper
			},
			userRepository: func() *mocks.UserRepository {
				userRepository := new(mocks.UserRepository)
				userRepository.On("Create", mock.Anything, mock.MatchedBy(func(user domain.User) bool {
					return cmp.Equal(domain.User{
						UserType:     "PRIVATE",
						Email:        "gildong@email.com",
						Name:         "ホン ギルドン",
						Password:     "test-password",
						SignupStatus: domain.SIGNUP_NONE,
						LoginType:    domain.MAIL,
						AuditStatus:  domain.PROVISIONAL_AUDIT,
					}, user,
						cmpopts.IgnoreFields(domain.User{}, "Salt", "CommonId", "Uid"),
						cmpopts.IgnoreTypes(jsonutils.JSONTime{}))
				})).Return(int64(1), nil)

				userRepository.On("Update", mock.Anything, mock.MatchedBy(func(user domain.User) bool {
					return cmp.Equal(domain.User{
						Id:           int64(1),
						Uid:          "10000001",
						UserType:     "PRIVATE",
						Email:        "gildong@email.com",
						Name:         "ホン ギルドン",
						Password:     "test-password",
						SignupStatus: domain.SIGNUP_NONE,
						LoginType:    domain.MAIL,
						AuditStatus:  domain.PROVISIONAL_AUDIT,
						InvestorType: domain.PRIVATE_INVESTOR,
					}, user,
						cmpopts.IgnoreFields(domain.User{}, "Salt", "CommonId"),
						cmpopts.IgnoreTypes(jsonutils.JSONTime{}))
				})).Return(nil)

				return userRepository
			},

			roleRepository: func() *mocks.RoleRepository {
				roleRepository := new(mocks.RoleRepository)
				roleRepository.On("Create", mock.Anything, domain.Role{
					UserId:   int64(1),
					RoleCode: string(domain.ROLE_USER),
				}).Return(int64(1), nil)

				return roleRepository
			},

			userAntisocialCheckRepository: func() *mocks.UserAntisocialCheckRepository {
				userAntisocialCheckRepository := new(mocks.UserAntisocialCheckRepository)
				userAntisocialCheckRepository.On("Create", mock.Anything, domain.UserAntisocialCheck{
					CheckUserId: int64(1),
				}).Return(nil)
				return userAntisocialCheckRepository
			},

			userAccountRepository: func() *mocks.UserAccountRepository {
				userAccountRepository := new(mocks.UserAccountRepository)
				userAccountRepository.On("Create", mock.Anything, domain.UserAccount{
					UserId: int64(1),
				}).Return(nil)

				return userAccountRepository
			},
			userService: func() *mocks.UserService {
				userService := new(mocks.UserService)
				userService.On("EncodePassword", context.Background(), "test-password", mock.MatchedBy(func(salt string) bool {
					return len(salt) == 10
				})).Return("test-password", nil)
				return userService
			},
			signUpRequestRepository: func() *mocks.SignUpRequestRepository {
				signUpRequestRepository := new(mocks.SignUpRequestRepository)
				signUpRequestRepository.On("GetByRequestId", context.Background(), "test-request-id").Return(domain.SignUpRequest{
					Email:        "gildong@email.com",
					EmailExpire:  jsonutils.JSONTime(time.Now().Add(1 * time.Hour)),
					ReferralCode: "ABCDL67890",
				}, nil)

				signUpRequestRepository.On("Update", context.Background(), mock.MatchedBy(func(request domain.SignUpRequest) bool {
					return time.Now().After(request.UrlExpire.Time())
				})).Return(nil)

				return signUpRequestRepository
			},
			privateProfileRepository: func() *mocks.PrivateProfileRepository {
				privateProfileRepository := new(mocks.PrivateProfileRepository)
				privateProfileRepository.On("Create", context.Background(), mock.MatchedBy(func(profile domain.PrivateProfile) bool {
					return cmp.Equal(domain.PrivateProfile{
						UserId:         int64(1),
						FirstName:      "ギルドン",
						FamilyName:     "ホン",
						FirstNameKana:  "ギルドン",
						FamilyNameKana: "ホン",
						RequestId:      "test-request-id",
					}, profile, cmpopts.IgnoreTypes(jsonutils.JSONTime{}))
				})).Return(nil)

				return privateProfileRepository
			},

			corporateProfileRepository: func() *mocks.CorporateProfileRepository {
				return new(mocks.CorporateProfileRepository)
			},

			mailService: func() *mocks.MailService {
				mailService := new(mocks.MailService)
				mailService.On("SendMail", context.Background(), mock.MatchedBy(func(cdomain.SendMailRequest) bool {
					return true
				})).Return(nil)

				return mailService
			},
			codeOwnershipRepository: func() *mocks.CodeOwnershipRepository {
				repo := new(mocks.CodeOwnershipRepository)
				repo.On("GetByCode", mock.Anything, domain.CodeTypeReferral, "ABCDE12345").Return((*domain.CodeOwnership)(nil), nil)
				repo.On("Create", mock.Anything, mock.MatchedBy(func(code domain.CodeOwnership) bool {
					return code.OwnerId == 1 && code.OwnerType == domain.PRIVATE && code.CodeType == domain.CodeTypeReferral && code.Code == "ABCDE12345"
				})).Return(nil)
				repo.On("GetByCode", mock.Anything, domain.CodeTypeReferral, "ABCDL67890").Return(&domain.CodeOwnership{
					Id:        99,
					OwnerId:   10,
					OwnerType: domain.PRIVATE,
					CodeType:  domain.CodeTypeReferral,
					Code:      "ABCDL67890",
				}, nil)
				return repo
			},
			codeUsageRepository: func() *mocks.CodeUsageRepository {
				repo := new(mocks.CodeUsageRepository)
				repo.On("Create", mock.Anything, mock.MatchedBy(func(usage domain.CodeUsage) bool {
					return usage.UsedOwnerId == 1 && usage.UsedOwnerType == domain.PRIVATE && usage.CodeOwnershipId == 99 && usage.CodeInputAt != nil
				})).Return(nil)
				return repo
			},
			generatedReferralCode: "ABCDE12345",
			wantErr:               false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := userUsecase{
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
				idGenerator:                   tt.idGenerator(),
				pwhelper:                      tt.pwhelper(),
				userRepository:                tt.userRepository(),
				roleRepository:                tt.roleRepository(),
				userAntisocialCheckRepository: tt.userAntisocialCheckRepository(),
				userAccountRepository:         tt.userAccountRepository(),
				userService:                   tt.userService(),
				signUpRequestRepository:       tt.signUpRequestRepository(),
				privateProfileRepository:      tt.privateProfileRepository(),
				corporateProfileRepository:    tt.corporateProfileRepository(),
				mailService:                   tt.mailService(),
				codeOwnershipRepository:       tt.codeOwnershipRepository(),
				codeUsageRepository:           tt.codeUsageRepository(),
			}

			usecase.generateRandomString = generateRandomString

			if tt.generatedReferralCode != "" {
				code := tt.generatedReferralCode
				usecase.generateRandomString = func(_ int, _ string) (string, error) {
					return code, nil
				}
			}

			err := usecase.SignUp(tt.args.ctx, tt.args.request)

			assert.Equalf(t, tt.wantErr, err != nil, "SignUp() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}
