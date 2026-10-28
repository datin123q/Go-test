package user_verify

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain/mocks"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
	udomain "github.com/illuminarean/energyshares-user-api/domain"
)

func Test_userVerifyUsecase_CheckEmailExists(t *testing.T) {
	type args struct {
		ctx     context.Context
		request udomain.UserCheckEmailExistsRequest
	}
	tests := []struct {
		name                    string
		args                    args
		userService             func() *mocks.UserService
		signUpRequestRepository func() *mocks.SignUpRequestRepository
		want                    bool
		wantErr                 error
	}{
		{
			name: "(成功) メール重複確認成功 => 重複なし",
			args: args{
				ctx: context.Background(),
				request: udomain.UserCheckEmailExistsRequest{
					Email: "test@email.com",
				},
			},
			userService: func() *mocks.UserService {
				service := &mocks.UserService{}
				service.On("IsDuplicateEmailIncludeDeleted", context.Background(), "test@email.com").Return(false, nil)
				return service
			},
			signUpRequestRepository: func() *mocks.SignUpRequestRepository {
				signUpRequestRepository := &mocks.SignUpRequestRepository{}

				// checkEmailExpire
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

				// checkUrlExpire
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

				return signUpRequestRepository
			},
			want:    false,
			wantErr: nil,
		},
		{
			name: "(成功) メール重複確認成功 => 重複あり",
			args: args{
				ctx: context.Background(),
				request: udomain.UserCheckEmailExistsRequest{
					Email: "gildong@email.com",
				},
			},
			userService: func() *mocks.UserService {
				service := &mocks.UserService{}
				service.On("IsDuplicateEmailIncludeDeleted", context.Background(), "gildong@email.com").Return(true, nil)
				return service
			},
			signUpRequestRepository: func() *mocks.SignUpRequestRepository {
				signUpRequestRepository := &mocks.SignUpRequestRepository{}
				return signUpRequestRepository
			},
			want:    true,
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := userVerifyUsecase{
				userService:             tt.userService(),
				signUpRequestRepository: tt.signUpRequestRepository(),
			}

			got, err := usecase.CheckEmailExists(tt.args.ctx, tt.args.request)
			if (err != nil) != (tt.wantErr != nil) {
				t.Errorf("userVerifyUsecase.CheckEmailExists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equalf(t, tt.want, got, "userVerifyUsecase.CheckEmailExists() = %v, want %v", got, tt.want)
		})
	}
}

func Test_userVerifyUsecase_RequestPasswordReset(t *testing.T) {
	type args struct {
		ctx     context.Context
		request udomain.UserRequestPasswordResetRequest
	}
	tests := []struct {
		name                        string
		args                        args
		userRepository              func() *mocks.UserRepository
		passwordResetRequestService func() *mocks.PasswordResetRequestService
		mailService                 func() *mocks.MailService
		want                        udomain.UserRequestPasswordResetResponse
		wantErr                     error
	}{
		{
			name: "(成功) パスワード再設定要求成功",
			args: args{
				ctx: context.Background(),
				request: udomain.UserRequestPasswordResetRequest{
					Email: "gildong@email.com",
				},
			},
			userRepository: func() *mocks.UserRepository {
				repo := &mocks.UserRepository{}
				repo.On("GetByEmail", context.Background(), "gildong@email.com").Return(mockUser, nil)
				return repo
			},
			passwordResetRequestService: func() *mocks.PasswordResetRequestService {
				service := &mocks.PasswordResetRequestService{}
				service.On("HandlePasswordResetRequest", context.Background(), "gildong@email.com").Return(domain.PasswordResetRequest{Url: "url", RequestId: "request-id"}, nil)
				return service
			},
			mailService: func() *mocks.MailService {
				service := &mocks.MailService{}
				service.On("SendMail", context.Background(), mockSendMailPasswordReset).Return(nil)
				return service
			},
			want: udomain.UserRequestPasswordResetResponse{
				Email: "gildong@email.com",
			},
			wantErr: nil,
		},
		{
			name: "(失敗) パスワード再設定要求3回以上の試行",
			args: args{
				ctx: context.Background(),
				request: udomain.UserRequestPasswordResetRequest{
					Email: "gildong@email.com",
				},
			},
			userRepository: func() *mocks.UserRepository {
				repo := &mocks.UserRepository{}
				repo.On("GetByEmail", context.Background(), "gildong@email.com").Return(mockUser, nil)
				return repo
			},
			passwordResetRequestService: func() *mocks.PasswordResetRequestService {
				service := &mocks.PasswordResetRequestService{}
				service.On("HandlePasswordResetRequest", context.Background(), mock.Anything).Return(domain.PasswordResetRequest{}, cdomain.ErrRequestLimitExceeded)
				return service
			},
			mailService: func() *mocks.MailService {
				service := &mocks.MailService{}
				return service
			},
			want:    udomain.UserRequestPasswordResetResponse{},
			wantErr: cdomain.ErrRequestLimitExceeded,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := userVerifyUsecase{
				cfg: &config.Config{
					CommonConfig: config.CommonConfig{
						ServiceConfig: struct {
							Mail serviceutils.MailConfig
							Url  serviceutils.UrlConfig
						}{
							Url: serviceutils.UrlConfig{
								UserServiceDomain: "https://dev.energyshares.jp",
							},
						},
					},
				},
				userRepository:              tt.userRepository(),
				passwordResetRequestService: tt.passwordResetRequestService(),
				mailService:                 tt.mailService(),
			}

			got, err := usecase.RequestPasswordReset(tt.args.ctx, tt.args.request)
			if (err != nil) != (tt.wantErr != nil) {
				t.Errorf("userVerifyUsecase.RequestPasswordReset() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equalf(t, tt.want, got, "userVerifyUsecase.RequestPasswordReset() = %v, want %v", got, tt.want)
		})
	}
}

func TestUserVerifyUsecase_ResetPassword(t *testing.T) {
	type args struct {
		ctx     context.Context
		request udomain.UserResetPasswordRequest
	}
	tests := []struct {
		name                           string
		args                           args
		userRepository                 func() *mocks.UserRepository
		passwordResetRequestRepository func() *mocks.PasswordResetRequestRepository
		userService                    func() *mocks.UserService
		mailService                    func() *mocks.MailService
		smsMsgService                  func() *mocks.SmsMsgService
		wantErr                        error
	}{
		{
			name: "(成功) パスワード再設定成功",
			args: args{
				ctx: context.Background(),
				request: udomain.UserResetPasswordRequest{
					RequestId: "request-id",
					Email:     "gildong@email.com",
					Password:  "new-password",
				},
			},
			userRepository: func() *mocks.UserRepository {
				repo := &mocks.UserRepository{}
				repo.On("GetByEmail", context.Background(), "gildong@email.com").Return(mockUser, nil)
				return repo
			},
			passwordResetRequestRepository: func() *mocks.PasswordResetRequestRepository {
				repo := &mocks.PasswordResetRequestRepository{}
				repo.On("GetByRequestId", context.Background(), "request-id").
					Return(domain.PasswordResetRequest{
						RequestId: "request-id",
						Email:     "gildong@email.com",
						Expire:    jsonutils.JSONTime(time.Now().Add(time.Minute * 10)),
					}, nil)
				repo.On("Update", context.Background(), mock.Anything).Return(nil)
				return repo
			},
			userService: func() *mocks.UserService {
				service := &mocks.UserService{}
				service.On("ChangePassword", context.Background(), mockUser, "new-password").Return(nil)
				return service
			},
			mailService: func() *mocks.MailService {
				service := &mocks.MailService{}
				service.On("SendMail", context.Background(), mockSendMailResetComplete).Return(nil)
				return service
			},
			smsMsgService: func() *mocks.SmsMsgService {
				service := &mocks.SmsMsgService{}
				service.On("SendSms", context.Background(), mockSendSmsResetComplete).Return(nil)
				return service
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := userVerifyUsecase{
				cfg: &config.Config{
					CommonConfig: config.CommonConfig{
						ServiceConfig: struct {
							Mail serviceutils.MailConfig
							Url  serviceutils.UrlConfig
						}{
							Url: serviceutils.UrlConfig{
								UserServiceDomain: "https://dev.energyshares.jp",
							},
							Mail: serviceutils.MailConfig{
								Info: "info@illuminarean.com",
							},
						},
					},
				},
				userRepository:                 tt.userRepository(),
				passwordResetRequestRepository: tt.passwordResetRequestRepository(),
				userService:                    tt.userService(),
				mailService:                    tt.mailService(),
				smsMsgService:                  tt.smsMsgService(),
			}

			err := usecase.ResetPassword(tt.args.ctx, tt.args.request)
			if err != nil && errors.Is(err, tt.wantErr) {
				t.Errorf("userVerifyUsecase.ResetPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestUserVerifyUsecase_GetPasswordResetRequest(t *testing.T) {
	now := time.Now()

	type args struct {
		ctx     context.Context
		request udomain.UserGetPasswordResetRequestRequest
	}
	tests := []struct {
		name                           string
		args                           args
		passwordResetRequestRepository func() *mocks.PasswordResetRequestRepository
		want                           udomain.UserGetPasswordResetRequestResponse
		wantErr                        error
	}{
		{
			name: "(成功) 有効期限内のパスワード再設定要求照会",
			args: args{
				ctx: context.Background(),
				request: udomain.UserGetPasswordResetRequestRequest{
					RequestId: "request-id",
				},
			},
			passwordResetRequestRepository: func() *mocks.PasswordResetRequestRepository {
				repo := &mocks.PasswordResetRequestRepository{}
				repo.On("GetByRequestId", context.Background(), "request-id").
					Return(domain.PasswordResetRequest{
						RequestId: "request-id",
						Email:     "test@example.com",
						Expire:    jsonutils.JSONTime(now.Add(time.Minute * 10)),
						Url:       "https://test.com",
						SendCount: 1,
					}, nil)
				return repo
			},
			want: udomain.UserGetPasswordResetRequestResponse{
				RequestId: "request-id",
				Email:     "test@example.com",
				Expire:    jsonutils.JSONTime(now.Add(time.Minute * 10)),
				Url:       "https://test.com",
				SendCount: 1,
			},
			wantErr: nil,
		},
		{
			name: "(失敗) 有効期限切れのパスワード再設定要求照会",
			args: args{
				ctx: context.Background(),
				request: udomain.UserGetPasswordResetRequestRequest{
					RequestId: "request-id",
				},
			},
			passwordResetRequestRepository: func() *mocks.PasswordResetRequestRepository {
				repo := &mocks.PasswordResetRequestRepository{}
				repo.On("GetByRequestId", context.Background(), "request-id").
					Return(domain.PasswordResetRequest{
						RequestId: "request-id",
						Email:     "test@example.com",
						Expire:    jsonutils.JSONTime(now.Add(time.Minute * -10)),
						Url:       "https://test.com",
						SendCount: 2,
					}, nil)
				return repo
			},
			want:    udomain.UserGetPasswordResetRequestResponse{},
			wantErr: cdomain.ErrAlreadyExpired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := userVerifyUsecase{
				passwordResetRequestRepository: tt.passwordResetRequestRepository(),
			}

			got, err := usecase.GetPasswordResetRequest(tt.args.ctx, tt.args.request)
			if (err != nil) != (tt.wantErr != nil) {
				t.Errorf("userVerifyUsecase.GetPasswordResetRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equalf(t, tt.want, got, "userVerifyUsecase.GetPasswordResetRequest() = %v, want %v", got, tt.want)
		})
	}
}

func TestUserVerifyUsecase_ChangePassword(t *testing.T) {
	type args struct {
		ctx     context.Context
		request udomain.UserChangePasswordRequest
	}
	tests := []struct {
		name           string
		args           args
		userService    func() *mocks.UserService
		userRepository func() *mocks.UserRepository
		mailService    func() *mocks.MailService
		smsMsgService  func() *mocks.SmsMsgService
		want           bool
		wantErr        error
	}{
		{
			name: "(成功) パスワード変更成功",
			args: args{
				ctx: context.Background(),
				request: udomain.UserChangePasswordRequest{
					Password: "new-password",
					JWTUser: jwthelper.User{
						Id:    int64(1),
						Name:  "ホン・ギルドン",
						Email: "gildong@email.com",
					},
				},
			},
			userRepository: func() *mocks.UserRepository {
				repo := &mocks.UserRepository{}
				repo.On("GetById", context.Background(), int64(1)).Return(mockUser, nil)
				return repo
			},
			userService: func() *mocks.UserService {
				service := &mocks.UserService{}
				service.On("ChangePassword", context.Background(), mockUser, "new-password").Return(nil)
				return service
			},
			mailService: func() *mocks.MailService {
				service := &mocks.MailService{}
				service.On("SendMail", context.Background(), mockSendMailResetComplete).Return(nil)
				return service
			},
			smsMsgService: func() *mocks.SmsMsgService {
				service := &mocks.SmsMsgService{}
				service.On("SendSms", context.Background(), mockSendSmsResetComplete).Return(nil)
				return service
			},
			want:    true,
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := userVerifyUsecase{
				cfg: &config.Config{
					CommonConfig: config.CommonConfig{
						ServiceConfig: struct {
							Mail serviceutils.MailConfig
							Url  serviceutils.UrlConfig
						}{
							Url: serviceutils.UrlConfig{
								UserServiceDomain: "https://dev.energyshares.jp",
							},
							Mail: serviceutils.MailConfig{
								Info: "info@illuminarean.com",
							},
						},
					},
				},
				userRepository: tt.userRepository(),
				userService:    tt.userService(),
				mailService:    tt.mailService(),
				smsMsgService:  tt.smsMsgService(),
			}

			err := usecase.ChangePassword(tt.args.ctx, tt.args.request)
			if err != nil && errors.Is(err, tt.wantErr) {
				t.Errorf("userVerifyUsecase.ChangePassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestUserVerifyUsecase_CheckPassword(t *testing.T) {
	type args struct {
		ctx     context.Context
		request udomain.UserCheckPasswordRequest
	}
	tests := []struct {
		name           string
		args           args
		pwhelper       func() *mocks.PasswordHelper
		userRepository func() *mocks.UserRepository
		want           bool
		wantErr        error
	}{
		{
			name: "(成功) パスワード確認成功",
			args: args{
				ctx: context.Background(),
				request: udomain.UserCheckPasswordRequest{
					Password: "password",
					JWTUser: jwthelper.User{
						Id:    mockUser.Id,
						Name:  mockUser.Name,
						Email: mockUser.Email,
					},
				},
			},
			userRepository: func() *mocks.UserRepository {
				repo := &mocks.UserRepository{}
				repo.On("GetById", context.Background(), mockUser.Id).Return(mockUser, nil)
				return repo
			},
			pwhelper: func() *mocks.PasswordHelper {
				pwhelper := &mocks.PasswordHelper{}
				encodedInput := mockUser.Password
				pwhelper.On("EncodePassword", "password", mockUser.Salt).Return(encodedInput)
				pwhelper.On("IsSamePassword", mockUser.Password, encodedInput).Return(true)
				return pwhelper
			},
			want:    true,
			wantErr: nil,
		},
		{
			name: "(失敗) パスワード不一致",
			args: args{
				ctx: context.Background(),
				request: udomain.UserCheckPasswordRequest{
					Password: "abcdefg",
					JWTUser: jwthelper.User{
						Id:    mockUser.Id,
						Name:  mockUser.Name,
						Email: mockUser.Email,
					},
				},
			},
			userRepository: func() *mocks.UserRepository {
				repo := &mocks.UserRepository{}
				repo.On("GetById", context.Background(), mockUser.Id).Return(mockUser, nil)
				return repo
			},
			pwhelper: func() *mocks.PasswordHelper {
				pwhelper := &mocks.PasswordHelper{}
				pwhelper.On("EncodePassword", "abcdefg", mockUser.Salt).Return("abcdefg")
				pwhelper.On("IsSamePassword", mockUser.Password, "abcdefg").Return(false)
				return pwhelper
			},
			want:    false,
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := userVerifyUsecase{
				userRepository: tt.userRepository(),
				pwhelper:       tt.pwhelper(),
			}

			got, err := usecase.CheckPassword(tt.args.ctx, tt.args.request)
			if (err != nil) != (tt.wantErr != nil) {
				t.Errorf("userVerifyUsecase.CheckPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equalf(t, tt.want, got, "userVerifyUsecase.CheckPassword() = %v, want %v", got, tt.want)
		})
	}
}

func Test_userVerifyUsecase_SendEmailCode(t *testing.T) {
	type args struct {
		ctx     context.Context
		request udomain.UserSendEmailCodeRequest
	}

	tests := []struct {
		name                       string
		args                       args
		emailCodeRequestRepository func() *mocks.EmailCodeRequestRepository
		mailService                func() *mocks.MailService
		userService                func() *mocks.UserService
		wantErr                    bool
	}{
		{
			name: "変更しようとするメールアドレスが既に存在しているメールアドレスに変更しようとしている",
			args: args{
				ctx: context.Background(),
				request: udomain.UserSendEmailCodeRequest{
					Email: "gildong@email.com",
				},
			},
			userService: func() *mocks.UserService {
				userService := new(mocks.UserService)
				userService.On("IsDuplicateEmail", context.Background(), "gildong@email.com").Return(true, nil)
				return userService
			},
			emailCodeRequestRepository: func() *mocks.EmailCodeRequestRepository {
				emailCodeRequestRepository := new(mocks.EmailCodeRequestRepository)
				return emailCodeRequestRepository
			},
			mailService: func() *mocks.MailService {
				mailService := new(mocks.MailService)
				return mailService
			},
			wantErr: true,
		},
		{
			name: "変更しようとするメール要求履歴が既に3回ある状況",
			args: args{
				ctx: context.Background(),
				request: udomain.UserSendEmailCodeRequest{
					Email: "gildong@email.com",
					JWTUser: jwthelper.User{
						Id: 100,
					},
				},
			},
			userService: func() *mocks.UserService {
				userService := new(mocks.UserService)
				userService.On("IsDuplicateEmail", context.Background(), "gildong@email.com").Return(false, nil)
				return userService
			},
			emailCodeRequestRepository: func() *mocks.EmailCodeRequestRepository {
				emailCodeRequestRepository := new(mocks.EmailCodeRequestRepository)
				emailCodeRequestRepository.On("GetAll", context.Background(),
					mock.MatchedBy(func(field func(*cdomain.EmailCodeRequestQueryOptionFields)) bool {
						opt := cdomain.EmailCodeRequestQueryOptionFunc()
						field(&opt)
						return *opt.UserId() == "100"
					}),
					mock.MatchedBy(func(field func(*cdomain.EmailCodeRequestQueryOptionFields)) bool {
						opt := cdomain.EmailCodeRequestQueryOptionFunc()
						field(&opt)
						return *opt.FirstReq() == true
					}),
					mock.MatchedBy(func(field func(*cdomain.EmailCodeRequestQueryOptionFields)) bool {
						opt := cdomain.EmailCodeRequestQueryOptionFunc()
						field(&opt)
						return *opt.GTECreated() == time.Now().Format(time.DateOnly)
					}),
				).Return([]domain.EmailCodeRequest{
					{}, {}, {}, // 3件返却（具体的に明示しない）
				}, nil)

				return emailCodeRequestRepository
			},
			mailService: func() *mocks.MailService {
				mailService := new(mocks.MailService)
				return mailService
			},
			wantErr: true,
		},
		{
			name: "正常にメール変更要求メール進行",
			args: args{
				ctx: context.Background(),
				request: udomain.UserSendEmailCodeRequest{
					Email: "gildong@email.com",
					JWTUser: jwthelper.User{
						Id: 100,
					},
				},
			},
			userService: func() *mocks.UserService {
				userService := new(mocks.UserService)
				userService.On("IsDuplicateEmail", context.Background(), "gildong@email.com").Return(false, nil)
				return userService
			},
			emailCodeRequestRepository: func() *mocks.EmailCodeRequestRepository {
				emailCodeRequestRepository := new(mocks.EmailCodeRequestRepository)

				emailCodeRequestRepository.On("Create", context.Background(),
					mock.MatchedBy(func(request domain.EmailCodeRequest) bool {
						return cmp.Equal(domain.EmailCodeRequest{
							Email:    "gildong@email.com",
							Code:     "123456",
							UserId:   100,
							FirstReq: true,
						}, request, cmpopts.IgnoreTypes(jsonutils.JSONTime{}))
					})).Return(nil)

				emailCodeRequestRepository.On("GetAll", context.Background(),
					mock.MatchedBy(func(field func(*cdomain.EmailCodeRequestQueryOptionFields)) bool {
						opt := cdomain.EmailCodeRequestQueryOptionFunc()
						field(&opt)
						return *opt.UserId() == "100"
					}),
					mock.MatchedBy(func(field func(*cdomain.EmailCodeRequestQueryOptionFields)) bool {
						opt := cdomain.EmailCodeRequestQueryOptionFunc()
						field(&opt)
						return *opt.FirstReq() == true
					}),
					mock.MatchedBy(func(field func(*cdomain.EmailCodeRequestQueryOptionFields)) bool {
						opt := cdomain.EmailCodeRequestQueryOptionFunc()
						field(&opt)
						return *opt.GTECreated() == time.Now().Format(time.DateOnly)
					}),
				).Return([]domain.EmailCodeRequest{}, nil)

				return emailCodeRequestRepository
			},
			mailService: func() *mocks.MailService {
				mailService := new(mocks.MailService)
				mailService.On("SendMail", context.Background(), mock.MatchedBy(func(request cdomain.SendMailRequest) bool {
					return cmp.Equal(cdomain.SendMailRequest{
						TemplateType: domain.CERTIFICATION_CODE_SEND_TEMPLATE,
						MailType:     domain.MAIL_SEND_PRIORITY_FAST,
						Sender:       "info@energyshares.jp",
						Receivers:    []string{"gildong@email.com"},
						Details: map[string]interface{}{
							"Receiver":          "gildong@email.com",
							"ServiceUrl":        "https://user.dev.energyshares.com",
							"CertificationCode": "123456",
						},
					}, request)
				})).Return(nil)

				return mailService
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := userVerifyUsecase{
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
				userService:                tt.userService(),
				emailCodeRequestRepository: tt.emailCodeRequestRepository(),
				mailService:                tt.mailService(),

				// 임의로 함수를 주입받음
				generateRandomString: func(n int, charSet string) (string, error) {
					return "123456", nil
				},
			}

			err := usecase.SendEmailCode(tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.wantErr, err != nil, "SendEmailCode() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_userVerifyUsecase_ResendEmailCode(t *testing.T) {
	type args struct {
		ctx     context.Context
		request udomain.UserResendEmailCodeRequest
	}

	tests := []struct {
		name                       string
		args                       args
		userService                func() *mocks.UserService
		emailCodeRequestRepository func() *mocks.EmailCodeRequestRepository
		mailService                func() *mocks.MailService
		wantErr                    bool
	}{
		{
			name: "変更しようとするメールアドレスが既に存在しているメールアドレスに変更しようとしている",
			args: args{
				ctx: context.Background(),
				request: udomain.UserResendEmailCodeRequest{
					Email: "gildong@email.com",
				},
			},

			userService: func() *mocks.UserService {
				userService := new(mocks.UserService)
				userService.On("IsDuplicateEmail", context.Background(), "gildong@email.com").Return(true, nil)
				return userService
			},
			emailCodeRequestRepository: func() *mocks.EmailCodeRequestRepository {
				emailCodeRequestRepository := new(mocks.EmailCodeRequestRepository)
				return emailCodeRequestRepository
			},
			mailService: func() *mocks.MailService {
				mailService := new(mocks.MailService)
				return mailService
			},
			wantErr: true,
		},
		{
			name: "要求送信履歴がない場合は再送信できません。",
			args: args{
				ctx: context.Background(),
				request: udomain.UserResendEmailCodeRequest{
					Email: "gildong@email.com",
					JWTUser: jwthelper.User{
						Id: 100,
					},
				},
			},
			userService: func() *mocks.UserService {
				userService := new(mocks.UserService)
				userService.On("IsDuplicateEmail", context.Background(), "gildong@email.com").Return(false, nil)
				return userService
			},
			emailCodeRequestRepository: func() *mocks.EmailCodeRequestRepository {
				emailCodeRequestRepository := new(mocks.EmailCodeRequestRepository)
				emailCodeRequestRepository.On("GetAll", context.Background(),
					mock.MatchedBy(func(field func(*cdomain.EmailCodeRequestQueryOptionFields)) bool {
						opt := cdomain.EmailCodeRequestQueryOptionFunc()
						field(&opt)
						return *opt.UserId() == "100"
					}),
					mock.MatchedBy(func(field func(*cdomain.EmailCodeRequestQueryOptionFields)) bool {
						opt := cdomain.EmailCodeRequestQueryOptionFunc()
						field(&opt)
						return *opt.GTECreated() == time.Now().Format(time.DateOnly)
					}),
					mock.MatchedBy(func(field func(*cdomain.EmailCodeRequestQueryOptionFields)) bool {
						opt := cdomain.EmailCodeRequestQueryOptionFunc()
						field(&opt)
						return opt.SortBy() == "id"
					}),
					mock.MatchedBy(func(field func(*cdomain.EmailCodeRequestQueryOptionFields)) bool {
						opt := cdomain.EmailCodeRequestQueryOptionFunc()
						field(&opt)
						return opt.Order() == "desc"
					}),
					// 再送信は既に2件が送信されている
				).Return([]domain.EmailCodeRequest{}, nil)

				return emailCodeRequestRepository
			},
			mailService: func() *mocks.MailService {
				mailService := new(mocks.MailService)
				return mailService
			},
			wantErr: true,
		},
		{
			name: "再要求メールが既に5件進行されたためエラーを発生",
			args: args{
				ctx: context.Background(),
				request: udomain.UserResendEmailCodeRequest{
					Email: "gildong@email.com",
					JWTUser: jwthelper.User{
						Id: 100,
					},
				},
			},
			userService: func() *mocks.UserService {
				userService := new(mocks.UserService)
				userService.On("IsDuplicateEmail", context.Background(), "gildong@email.com").Return(false, nil)
				return userService
			},
			emailCodeRequestRepository: func() *mocks.EmailCodeRequestRepository {
				emailCodeRequestRepository := new(mocks.EmailCodeRequestRepository)
				emailCodeRequestRepository.On("GetAll", context.Background(),
					mock.MatchedBy(func(field func(*cdomain.EmailCodeRequestQueryOptionFields)) bool {
						opt := cdomain.EmailCodeRequestQueryOptionFunc()
						field(&opt)
						return *opt.UserId() == "100"
					}),
					mock.MatchedBy(func(field func(*cdomain.EmailCodeRequestQueryOptionFields)) bool {
						opt := cdomain.EmailCodeRequestQueryOptionFunc()
						field(&opt)
						return *opt.GTECreated() == time.Now().Format(time.DateOnly)
					}),
					mock.MatchedBy(func(field func(*cdomain.EmailCodeRequestQueryOptionFields)) bool {
						opt := cdomain.EmailCodeRequestQueryOptionFunc()
						field(&opt)
						return opt.SortBy() == "id"
					}),
					mock.MatchedBy(func(field func(*cdomain.EmailCodeRequestQueryOptionFields)) bool {
						opt := cdomain.EmailCodeRequestQueryOptionFunc()
						field(&opt)
						return opt.Order() == "desc"
					}),
					// 再送信は既に2件が送信されている
				).Return([]domain.EmailCodeRequest{
					{
						Id:       1005,
						Email:    "gildong@email.com",
						Code:     "823837",
						UserId:   100,
						FirstReq: false,
					},
					{
						Id:       1004,
						Email:    "gildong@email.com",
						Code:     "823836",
						UserId:   100,
						FirstReq: false,
					},
					{
						Id:       1003,
						Email:    "gildong@email.com",
						Code:     "823835",
						UserId:   100,
						FirstReq: false,
					},
					{
						Id:       1002,
						Email:    "gildong@email.com",
						Code:     "823834",
						UserId:   100,
						FirstReq: false,
					},
					{
						Id:       1001,
						Email:    "gildong@email.com",
						Code:     "823833",
						UserId:   100,
						FirstReq: false,
					},
					{
						Id:       1000,
						Email:    "gildong@email.com",
						Code:     "823832",
						UserId:   100,
						FirstReq: true,
					},
				}, nil)

				return emailCodeRequestRepository
			},
			mailService: func() *mocks.MailService {
				mailService := new(mocks.MailService)
				return mailService
			},
			wantErr: true,
		},
		{
			name: "正常にメール変更再要求メール進行",
			args: args{
				ctx: context.Background(),
				request: udomain.UserResendEmailCodeRequest{
					Email: "gildong@email.com",
					JWTUser: jwthelper.User{
						Id: 100,
					},
				},
			},
			userService: func() *mocks.UserService {
				userService := new(mocks.UserService)
				userService.On("IsDuplicateEmail", context.Background(), "gildong@email.com").Return(false, nil)
				return userService
			},
			emailCodeRequestRepository: func() *mocks.EmailCodeRequestRepository {
				emailCodeRequestRepository := new(mocks.EmailCodeRequestRepository)

				emailCodeRequestRepository.On("Create", context.Background(),
					mock.MatchedBy(func(request domain.EmailCodeRequest) bool {
						return cmp.Equal(domain.EmailCodeRequest{
							Email:    "gildong@email.com",
							Code:     "123456",
							UserId:   100,
							FirstReq: false,
						}, request, cmpopts.IgnoreTypes(jsonutils.JSONTime{}))
					})).Return(nil)

				emailCodeRequestRepository.On("GetAll", context.Background(),
					mock.MatchedBy(func(field func(*cdomain.EmailCodeRequestQueryOptionFields)) bool {
						opt := cdomain.EmailCodeRequestQueryOptionFunc()
						field(&opt)
						return *opt.UserId() == "100"
					}),
					mock.MatchedBy(func(field func(*cdomain.EmailCodeRequestQueryOptionFields)) bool {
						opt := cdomain.EmailCodeRequestQueryOptionFunc()
						field(&opt)
						return *opt.GTECreated() == time.Now().Format(time.DateOnly)
					}),
					mock.MatchedBy(func(field func(*cdomain.EmailCodeRequestQueryOptionFields)) bool {
						opt := cdomain.EmailCodeRequestQueryOptionFunc()
						field(&opt)
						return opt.SortBy() == "id"
					}),
					mock.MatchedBy(func(field func(*cdomain.EmailCodeRequestQueryOptionFields)) bool {
						opt := cdomain.EmailCodeRequestQueryOptionFunc()
						field(&opt)
						return opt.Order() == "desc"
					}),
					// 再送信は既に2件が送信されている
				).Return([]domain.EmailCodeRequest{
					{
						Id:       1002,
						Email:    "gildong@email.com",
						Code:     "823834",
						UserId:   100,
						FirstReq: false,
					},
					{
						Id:       1001,
						Email:    "gildong@email.com",
						Code:     "823833",
						UserId:   100,
						FirstReq: false,
					},
					{
						Id:       1000,
						Email:    "gildong@email.com",
						Code:     "823832",
						UserId:   100,
						FirstReq: true,
					},
				}, nil)

				return emailCodeRequestRepository
			},
			mailService: func() *mocks.MailService {
				mailService := new(mocks.MailService)
				mailService.On("SendMail", context.Background(), mock.MatchedBy(func(request cdomain.SendMailRequest) bool {
					return cmp.Equal(cdomain.SendMailRequest{
						TemplateType: domain.CERTIFICATION_CODE_SEND_TEMPLATE,
						MailType:     domain.MAIL_SEND_PRIORITY_FAST,
						Sender:       "info@energyshares.jp",
						Receivers:    []string{"gildong@email.com"},
						Details: map[string]interface{}{
							"Receiver":          "gildong@email.com",
							"ServiceUrl":        "https://user.dev.energyshares.com",
							"CertificationCode": "123456",
						},
					}, request)
				})).Return(nil)

				return mailService
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := userVerifyUsecase{
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
				emailCodeRequestRepository: tt.emailCodeRequestRepository(),
				mailService:                tt.mailService(),
				userService:                tt.userService(),

				// 任意に関数を注入される
				generateRandomString: func(n int, charSet string) (string, error) {
					return "123456", nil
				},
			}

			err := usecase.ResendEmailCode(tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.wantErr, err != nil, "ResendEmailCode() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_userVerifyUsecase_CheckEmailCode(t *testing.T) {
	type args struct {
		ctx     context.Context
		request udomain.UserCheckEmailCodeRequest
	}

	tests := []struct {
		name                       string
		args                       args
		mailService                func() *mocks.MailService
		userRepository             func() *mocks.UserRepository
		emailCodeRequestRepository func() *mocks.EmailCodeRequestRepository
		userService                func() *mocks.UserService
		wantErr                    bool
	}{
		{
			name: "照会される結果がないためエラー発生",
			args: args{
				ctx: context.Background(),
				request: udomain.UserCheckEmailCodeRequest{
					Email: "gildong@email.com",
					Code:  "123456",
				},
			},
			userRepository: func() *mocks.UserRepository {
				userRepository := new(mocks.UserRepository)
				return userRepository
			},
			userService: func() *mocks.UserService {
				userService := new(mocks.UserService)
				return userService
			},
			emailCodeRequestRepository: func() *mocks.EmailCodeRequestRepository {
				emailCodeRequestRepository := new(mocks.EmailCodeRequestRepository)
				emailCodeRequestRepository.On("GetAll", context.Background(),
					mock.MatchedBy(func(field func(*cdomain.EmailCodeRequestQueryOptionFields)) bool {
						opt := cdomain.EmailCodeRequestQueryOptionFunc()
						field(&opt)
						return *opt.Email() == "gildong@email.com"
					}),
					mock.MatchedBy(func(field func(*cdomain.EmailCodeRequestQueryOptionFields)) bool {
						opt := cdomain.EmailCodeRequestQueryOptionFunc()
						field(&opt)
						return *opt.Code() == "123456"
					}),
				).Return([]domain.EmailCodeRequest{}, nil)

				return emailCodeRequestRepository
			},
			mailService: func() *mocks.MailService {
				mailService := new(mocks.MailService)
				return mailService
			},
			wantErr: true,
		},
		{
			name: "照会はされたが現在時刻基準で有効期限切れのためエラー発生",
			args: args{
				ctx: context.Background(),
				request: udomain.UserCheckEmailCodeRequest{
					Email: "gildong@email.com",
					Code:  "123456",
				},
			},
			userRepository: func() *mocks.UserRepository {
				userRepository := new(mocks.UserRepository)
				return userRepository
			},

			userService: func() *mocks.UserService {
				userService := new(mocks.UserService)
				return userService
			},
			emailCodeRequestRepository: func() *mocks.EmailCodeRequestRepository {
				emailCodeRequestRepository := new(mocks.EmailCodeRequestRepository)
				emailCodeRequestRepository.On("GetAll", context.Background(),
					mock.MatchedBy(func(field func(*cdomain.EmailCodeRequestQueryOptionFields)) bool {
						opt := cdomain.EmailCodeRequestQueryOptionFunc()
						field(&opt)
						return *opt.Email() == "gildong@email.com"
					}),
					mock.MatchedBy(func(field func(*cdomain.EmailCodeRequestQueryOptionFields)) bool {
						opt := cdomain.EmailCodeRequestQueryOptionFunc()
						field(&opt)
						return *opt.Code() == "123456"
					}),
				).Return([]domain.EmailCodeRequest{
					{
						Id:     1,
						Expire: jsonutils.JSONTime(time.Now().Add(-1 * time.Minute)), // 1分経過した状況
					},
				}, nil)

				return emailCodeRequestRepository
			},
			mailService: func() *mocks.MailService {
				mailService := new(mocks.MailService)
				return mailService
			},
			wantErr: true,
		},
		{
			name: "正常にコードチェック成功",
			args: args{
				ctx: context.Background(),
				request: udomain.UserCheckEmailCodeRequest{
					JWTUser: jwthelper.User{
						Id:    100,
						Email: "gildong@email.com",
					},
					Email: "dool2@email.com",
					Code:  "123456",
				},
			},
			userRepository: func() *mocks.UserRepository {
				userRepository := new(mocks.UserRepository)
				userRepository.On("GetById", context.Background(), int64(100)).Return(domain.User{
					Id:    100,
					Email: "gildong@email.com",
					Name:  "홍길동",
				}, nil)

				userRepository.On("Update", context.Background(), domain.User{
					Id:    100,
					Name:  "홍길동",
					Email: "dool2@email.com",
				}).Return(nil)
				return userRepository
			},

			userService: func() *mocks.UserService {
				userService := new(mocks.UserService)
				userService.On("IsDuplicateEmail", context.Background(), "dool2@email.com").Return(false, nil)
				return userService
			},
			emailCodeRequestRepository: func() *mocks.EmailCodeRequestRepository {
				emailCodeRequestRepository := new(mocks.EmailCodeRequestRepository)
				emailCodeRequestRepository.On("GetAll", context.Background(),
					mock.MatchedBy(func(field func(*cdomain.EmailCodeRequestQueryOptionFields)) bool {
						opt := cdomain.EmailCodeRequestQueryOptionFunc()
						field(&opt)
						return *opt.Email() == "dool2@email.com"
					}),
					mock.MatchedBy(func(field func(*cdomain.EmailCodeRequestQueryOptionFields)) bool {
						opt := cdomain.EmailCodeRequestQueryOptionFunc()
						field(&opt)
						return *opt.Code() == "123456"
					}),
				).Return([]domain.EmailCodeRequest{
					{
						Id:     1,
						Expire: jsonutils.JSONTime(time.Now().Add(time.Minute)), // 1分残っている状況
					},
				}, nil)
				emailCodeRequestRepository.On("Update", context.Background(), mock.MatchedBy(func(request domain.EmailCodeRequest) bool {
					return request.Id == 1
				})).Return(nil)

				return emailCodeRequestRepository
			},
			mailService: func() *mocks.MailService {
				mailService := new(mocks.MailService)
				mailService.On("SendMail", context.Background(), cdomain.SendMailRequest{
					TemplateType: domain.UPDATE_USER_INFORMATION_TEMPLATE,
					Sender:       "info@energyshares.jp",
					MailType:     domain.MAIL_SEND_PRIORITY_FAST,
					Receivers:    []string{"dool2@email.com"},
					Details: map[string]interface{}{
						"Receiver":    "홍길동",
						"ServiceUrl":  "https://user.dev.energyshares.com",
						"Info":        "メールアドレス",
						"BeforePhone": "gildong@email.com",
						"AfterPhone":  "dool2@email.com",
					},
				}).Return(nil)
				return mailService
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := userVerifyUsecase{
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
				userRepository:             tt.userRepository(),
				emailCodeRequestRepository: tt.emailCodeRequestRepository(),
				userService:                tt.userService(),
				mailService:                tt.mailService(),
				logger:                     factory.NewLoggerFactory(),
			}

			err := usecase.CheckEmailCode(tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.wantErr, err != nil, "CheckEmailCode() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}
