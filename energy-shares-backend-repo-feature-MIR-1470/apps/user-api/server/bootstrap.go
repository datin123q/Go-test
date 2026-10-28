package server

import (
	"context"
	"os"
	"path/filepath"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/go-resty/resty/v2"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"github.com/swaggest/openapi-go/openapi31"
	"go.uber.org/fx"
	"xorm.io/xorm"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/cipher"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/echoutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/egift"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/emailutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/filehelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/idgenerator"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/openapi"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/pathutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/pwhelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/s3utils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/sessionutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/snsutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/swagger"

	// Repository Layer
	adminUserRepository "github.com/illuminarean/energy-shares-backend-repo/repository/admin_user/mysql"
	ansisocialCheckFileRepository "github.com/illuminarean/energy-shares-backend-repo/repository/antisocial_check_file/mysql"
	attachmentRepository "github.com/illuminarean/energy-shares-backend-repo/repository/attachment/mysql"
	bannerRepository "github.com/illuminarean/energy-shares-backend-repo/repository/banner/mysql"
	bannerSettingRepository "github.com/illuminarean/energy-shares-backend-repo/repository/banner_setting/mysql"
	blogPostRepository "github.com/illuminarean/energy-shares-backend-repo/repository/blog_post/mysql"
	codeOwnershipRepository "github.com/illuminarean/energy-shares-backend-repo/repository/code_ownership/mysql"
	codeUsageRepository "github.com/illuminarean/energy-shares-backend-repo/repository/code_usage/mysql"
	contractRepository "github.com/illuminarean/energy-shares-backend-repo/repository/contract/mysql"
	corporateProfileRepository "github.com/illuminarean/energy-shares-backend-repo/repository/corporate_profile/mysql"
	corporateShareHolderRepository "github.com/illuminarean/energy-shares-backend-repo/repository/corporate_shareholder/mysql"
	depositRequestRepository "github.com/illuminarean/energy-shares-backend-repo/repository/deposit_request/mysql"
	depositRequestLackHistory "github.com/illuminarean/energy-shares-backend-repo/repository/deposit_requset_lack_history/mysql"
	dividendRepository "github.com/illuminarean/energy-shares-backend-repo/repository/dividend/mysql"
	ekycRequestLogRepository "github.com/illuminarean/energy-shares-backend-repo/repository/ekyc_request_log/mysql"
	emailCodeRequestRepository "github.com/illuminarean/energy-shares-backend-repo/repository/email_code_request/mysql"
	eventMessageRepository "github.com/illuminarean/energy-shares-backend-repo/repository/event_message/mysql"
	faqRepository "github.com/illuminarean/energy-shares-backend-repo/repository/faq/mysql"
	investmentRepository "github.com/illuminarean/energy-shares-backend-repo/repository/investment/mysql"
	investmentRequestRepository "github.com/illuminarean/energy-shares-backend-repo/repository/investment_request/mysql"
	investmentWaitHistoryRepository "github.com/illuminarean/energy-shares-backend-repo/repository/investment_wait_history/mysql"
	investmentWaitersRepository "github.com/illuminarean/energy-shares-backend-repo/repository/investment_waiters/mysql"
	loginHistoryMysqlRepository "github.com/illuminarean/energy-shares-backend-repo/repository/login_history/mysql"
	lookUpMyNumberRepository "github.com/illuminarean/energy-shares-backend-repo/repository/look_up_my_number/mysql"
	lotteryApplicationRepository "github.com/illuminarean/energy-shares-backend-repo/repository/lottery_application/mysql"
	lotteryHistoryRepository "github.com/illuminarean/energy-shares-backend-repo/repository/lottery_history/mysql"
	mailFailedRepository "github.com/illuminarean/energy-shares-backend-repo/repository/mail_failed/mysql"
	mailRetryRepository "github.com/illuminarean/energy-shares-backend-repo/repository/mail_retry/mysql"
	mapSmsReceiverRepository "github.com/illuminarean/energy-shares-backend-repo/repository/map_sms_receiver/mysql"
	marketingCodeRepository "github.com/illuminarean/energy-shares-backend-repo/repository/marketing_code/mysql"
	myNumberRepository "github.com/illuminarean/energy-shares-backend-repo/repository/my_number/mysql"
	newsRepository "github.com/illuminarean/energy-shares-backend-repo/repository/news/mysql"
	noticeMsgRepository "github.com/illuminarean/energy-shares-backend-repo/repository/notice_msg/mysql"
	noticeMsgReadRepository "github.com/illuminarean/energy-shares-backend-repo/repository/notice_msg_read/mysql"
	operatorHistoryRepository "github.com/illuminarean/energy-shares-backend-repo/repository/operator_history/mysql"
	passwordResetRequestRepository "github.com/illuminarean/energy-shares-backend-repo/repository/password_reset_request/mysql"
	policyRepository "github.com/illuminarean/energy-shares-backend-repo/repository/policy/mysql"
	policyHistoryRepository "github.com/illuminarean/energy-shares-backend-repo/repository/policy_history/mysql"
	privateProfileRepository "github.com/illuminarean/energy-shares-backend-repo/repository/private_profile/mysql"
	privateUserRepository "github.com/illuminarean/energy-shares-backend-repo/repository/private_user/mysql"
	projectRepository "github.com/illuminarean/energy-shares-backend-repo/repository/project/mysql"
	projectFileRepository "github.com/illuminarean/energy-shares-backend-repo/repository/project_file/mysql"
	projectImageRepository "github.com/illuminarean/energy-shares-backend-repo/repository/project_image/mysql"
	projectInvestmentExecuteRepository "github.com/illuminarean/energy-shares-backend-repo/repository/project_investment_execute/mysql"
	projectMilestoneRepository "github.com/illuminarean/energy-shares-backend-repo/repository/project_milestone/mysql"
	projectNewsRepository "github.com/illuminarean/energy-shares-backend-repo/repository/project_news/mysql"
	projectViewCountRepository "github.com/illuminarean/energy-shares-backend-repo/repository/project_view_count/mysql"
	refreshTokenRepository "github.com/illuminarean/energy-shares-backend-repo/repository/refresh_token/mysql"
	repaymentDividendScheduleRepository "github.com/illuminarean/energy-shares-backend-repo/repository/repayment_dividend_schedule/mysql"
	roleRepository "github.com/illuminarean/energy-shares-backend-repo/repository/role/mysql"
	shareRepository "github.com/illuminarean/energy-shares-backend-repo/repository/share/mysql"
	signUpRequestRepository "github.com/illuminarean/energy-shares-backend-repo/repository/sign_up_request/mysql"
	smsMsgRepository "github.com/illuminarean/energy-shares-backend-repo/repository/sms_msg/mysql"
	spvRepository "github.com/illuminarean/energy-shares-backend-repo/repository/spv/mysql"
	transactionInfoHistoryRepository "github.com/illuminarean/energy-shares-backend-repo/repository/transaction_info_history/mysql"
	userRepository "github.com/illuminarean/energy-shares-backend-repo/repository/user/mysql"
	userAccountRepository "github.com/illuminarean/energy-shares-backend-repo/repository/user_account/mysql"
	userAntisocialCheckRepository "github.com/illuminarean/energy-shares-backend-repo/repository/user_antisocial_check/mysql"
	userFileRepository "github.com/illuminarean/energy-shares-backend-repo/repository/user_file/mysql"
	userInfoHistoryRepository "github.com/illuminarean/energy-shares-backend-repo/repository/user_info_history/mysql"
	userLoginLockRepository "github.com/illuminarean/energy-shares-backend-repo/repository/user_login_lock/mysql"
	userNoticeMappingRepository "github.com/illuminarean/energy-shares-backend-repo/repository/user_notice_mapping/mysql"
	userTransactionRepository "github.com/illuminarean/energy-shares-backend-repo/repository/user_transaction/mysql"
	userVirtualAccountRepository "github.com/illuminarean/energy-shares-backend-repo/repository/user_virtual_account/mysql"
	withdrawalRepository "github.com/illuminarean/energy-shares-backend-repo/repository/withdrawal/mysql"

	// Service Layer
	authService "github.com/illuminarean/energy-shares-backend-repo/service/auth"
	eventMessageService "github.com/illuminarean/energy-shares-backend-repo/service/event_message"
	lookUpMyNumberService "github.com/illuminarean/energy-shares-backend-repo/service/look_up_my_number"
	lotteryValidationService "github.com/illuminarean/energy-shares-backend-repo/service/lottery_validation"
	mailService "github.com/illuminarean/energy-shares-backend-repo/service/mail" // 統合メールサービスで送信する実装
	myNumberService "github.com/illuminarean/energy-shares-backend-repo/service/my_number"
	nexwayService "github.com/illuminarean/energy-shares-backend-repo/service/nexway"
	passwordResetRequestService "github.com/illuminarean/energy-shares-backend-repo/service/password_reset_request"
	policyService "github.com/illuminarean/energy-shares-backend-repo/service/policy"
	smsMsgService "github.com/illuminarean/energy-shares-backend-repo/service/sms_msg"
	userService "github.com/illuminarean/energy-shares-backend-repo/service/user"
	userAccountService "github.com/illuminarean/energy-shares-backend-repo/service/user_account"
	UserAntisocialCheckService "github.com/illuminarean/energy-shares-backend-repo/service/user_antisocial_check"
	UserFileService "github.com/illuminarean/energy-shares-backend-repo/service/user_file"
	UserVirtualAccountService "github.com/illuminarean/energy-shares-backend-repo/service/user_virtual_account"

	// Usecase Layer
	bannerUsecase "github.com/illuminarean/energyshares-user-api/usecase/banner"
	blogPostUsecase "github.com/illuminarean/energyshares-user-api/usecase/blog_post"
	codeUsecase "github.com/illuminarean/energyshares-user-api/usecase/code"
	corporateUserUsecase "github.com/illuminarean/energyshares-user-api/usecase/corporate_user"
	faqUsecase "github.com/illuminarean/energyshares-user-api/usecase/faq"
	investmentUsecase "github.com/illuminarean/energyshares-user-api/usecase/investment"
	investmentWaitersUsecase "github.com/illuminarean/energyshares-user-api/usecase/investment_waiters"
	lotteryUsecase "github.com/illuminarean/energyshares-user-api/usecase/lottery"
	myNumberUsecase "github.com/illuminarean/energyshares-user-api/usecase/my_number"
	newsUsecase "github.com/illuminarean/energyshares-user-api/usecase/news"
	noticeMsgUsecase "github.com/illuminarean/energyshares-user-api/usecase/notice_msg"
	policyUsecase "github.com/illuminarean/energyshares-user-api/usecase/policy"
	privateUserUsecase "github.com/illuminarean/energyshares-user-api/usecase/private_user"
	userUsecase "github.com/illuminarean/energyshares-user-api/usecase/user"
	userVerifyUsecase "github.com/illuminarean/energyshares-user-api/usecase/user_verify"

	// Delivery Layer
	bannerHttpDelivery "github.com/illuminarean/energyshares-user-api/delivery/banner/http"
	blogPostHttpDelivery "github.com/illuminarean/energyshares-user-api/delivery/blog_post/http"
	codeHttpDelivery "github.com/illuminarean/energyshares-user-api/delivery/code/http"
	commonHttpDelivery "github.com/illuminarean/energyshares-user-api/delivery/common/http"
	corporateUserHttpDelivery "github.com/illuminarean/energyshares-user-api/delivery/corporate_user/http"
	faqHttpDelivery "github.com/illuminarean/energyshares-user-api/delivery/faq/http"
	healthHttpDelivery "github.com/illuminarean/energyshares-user-api/delivery/health/http"
	investmentHttpDelivery "github.com/illuminarean/energyshares-user-api/delivery/investment/http"
	investmentWaitersHttpDelivery "github.com/illuminarean/energyshares-user-api/delivery/investment_waiters/http"
	lotteryHttpDelivery "github.com/illuminarean/energyshares-user-api/delivery/lottery/http"
	myNumberHttpDelivery "github.com/illuminarean/energyshares-user-api/delivery/my_number/http"
	newsHttpDelivery "github.com/illuminarean/energyshares-user-api/delivery/news/http"
	noticeMsgHttpDelivery "github.com/illuminarean/energyshares-user-api/delivery/notice_msg/http"
	policyHttpDelivery "github.com/illuminarean/energyshares-user-api/delivery/policy/http"
	privateUserHttpDelivery "github.com/illuminarean/energyshares-user-api/delivery/private_user/http"
	userHttpDelivery "github.com/illuminarean/energyshares-user-api/delivery/user/http"
	userVerifyHttpDelivery "github.com/illuminarean/energyshares-user-api/delivery/user_verify/http"
)

type bootstrap struct {
	injections []interface{}
}

func NewBootstrap(injections ...interface{}) *bootstrap {
	return &bootstrap{
		injections: injections,
	}
}

func inject[T any](injections []interface{}, targetType T) func() T {
	targetTypeReflect := reflect.TypeOf(targetType)
	for _, injection := range injections {
		if reflect.TypeOf(injection) == targetTypeReflect {
			// 型が一致する場合、itemをT型に変換して返します。
			// ここではType Assertionを使用します。
			value, ok := injection.(T)
			if !ok {
				// 型変換に失敗した場合
				return func() T { return *new(T) }
			}
			return func() T { return value }
		}
	}

	return func() T { return *new(T) }
}

func (b *bootstrap) Invoke() *fx.App {
	return fx.New(
		fx.Provide(
			inject(b.injections, &echo.Echo{}),
			inject(b.injections, &config.Config{}),
			inject(b.injections, &xorm.Engine{}),
			inject(b.injections, &sqs.Client{}),

			// pkg injections
			cipher.New,
			jwthelper.NewUserExtractor,
			pwhelper.NewPasswordHelper,
			filehelper.NewFileHelper,
			resty.New,
			func() i18nutils.Translator {
				return i18nutils.New(pathutils.GetPathOfAssets())
			},

			func(cfg *config.Config, e *echo.Echo) echoswagger.ApiRoot {
				return lo.If(cfg.Debug, echoswagger.New(e, "/doc", &echoswagger.Info{
					Title: "Energy Shares Japan User API",
					Description: `This is API doc for Energy Shares Japan User API.
The displayed version incorporates the commit date and hash for precise version tracking.
ここをクリックするとクリーンアーキテクチャが適用されたv1スワッガーページに移動します - [https://user.dev.energyshares.jp/api/swagger/index.html](https://user.dev.energyshares.jp/api/swagger/index.html)`,
					Version: "commit: " + os.Getenv("VERSION_INFO"),
				})).Else(echoswagger.NewNop(e))
			},

			func(cfg *config.Config) *openapi31.Reflector {
				reflector := openapi31.NewReflector()
				reflector.Spec.Info.
					WithTitle("Energy Shares Japan User V1 API").
					WithVersion("1.0.0").
					WithDescription("User API V1 for Energy Shares Japan")
				reflector.Spec.SetHTTPBearerTokenSecurity("Authorization", "JWT", "JWTアクセストークンが必要です")
				reflector.Spec.Servers = []openapi31.Server{}
				return reflector
			},
			func() *config.Code {
				return config.NewCode("")
			},
			openapi.NewSwaggerUISetup,
			echoutils.NewEchoSwagger,

			// factory injections
			factory.NewTxManager,
			factory.NewDBFactory,
			factory.NewLoggerFactory,
			factory.NewEventBusFactory,
			factory.NewSnsFactory,

			// utils
			egift.New,
			emailutils.New,
			idgenerator.New,
			s3utils.New,
			sessionutils.New,
			snsutils.New,

			// Repository Layer
			adminUserRepository.New,
			ansisocialCheckFileRepository.New,
			attachmentRepository.New,
			bannerRepository.New,
			bannerSettingRepository.New,
			blogPostRepository.New,
			contractRepository.New,
			corporateProfileRepository.New,
			corporateShareHolderRepository.New,
			depositRequestLackHistory.New,
			depositRequestRepository.New,
			dividendRepository.New,
			ekycRequestLogRepository.New,
			eventMessageRepository.New,
			faqRepository.New,
			investmentRepository.New,
			investmentRequestRepository.New,
			investmentWaitersRepository.New,
			investmentWaitHistoryRepository.New,
			loginHistoryMysqlRepository.New,
			lookUpMyNumberRepository.New,
			mailFailedRepository.New,
			mailRetryRepository.New,
			mapSmsReceiverRepository.New,
			marketingCodeRepository.New,
			codeOwnershipRepository.New,
			codeUsageRepository.New,
			lotteryApplicationRepository.New,
			lotteryHistoryRepository.New,
			myNumberRepository.New,
			newsRepository.New,
			noticeMsgRepository.New,
			noticeMsgReadRepository.New,
			operatorHistoryRepository.New,
			passwordResetRequestRepository.New,
			policyRepository.New,
			policyHistoryRepository.New,
			privateProfileRepository.New,
			privateUserRepository.New,
			projectFileRepository.New,
			projectImageRepository.New,
			projectInvestmentExecuteRepository.New,
			projectMilestoneRepository.New,
			projectNewsRepository.New,
			projectRepository.New,
			projectViewCountRepository.New,
			refreshTokenRepository.New,
			repaymentDividendScheduleRepository.New,
			roleRepository.New,
			shareRepository.New,
			signUpRequestRepository.New,
			smsMsgRepository.New,
			spvRepository.New,
			transactionInfoHistoryRepository.New,
			userAccountRepository.New,
			userAntisocialCheckRepository.New,
			userFileRepository.New,
			userLoginLockRepository.New,
			userNoticeMappingRepository.New,
			userInfoHistoryRepository.New,
			func(dbFactory factory.DBFactory, history cdomain.UserInfoHistoryRepository, logger factory.LoggerFactory) cdomain.UserRepository {
				return userRepository.NewHistoryAware(userRepository.New(dbFactory), history, logger)
			},

			userTransactionRepository.New,
			userVirtualAccountRepository.New,
			emailCodeRequestRepository.New,
			withdrawalRepository.New,

			// Service Layer
			authService.New,
			eventMessageService.New,
			lookUpMyNumberService.New,
			lotteryValidationService.New,
			mailService.New,
			myNumberService.New,
			nexwayService.New,
			passwordResetRequestService.New,
			policyService.New,
			smsMsgService.New,
			userAccountService.New,
			UserAntisocialCheckService.New,
			UserFileService.New,
			userService.New,
			UserVirtualAccountService.New,

			// Usecase Layer
			bannerUsecase.New,
			lotteryUsecase.New,
			blogPostUsecase.New,
			codeUsecase.New,
			corporateUserUsecase.New,
			faqUsecase.New,
			investmentUsecase.New,
			investmentWaitersUsecase.New,
			myNumberUsecase.New,
			newsUsecase.New,
			noticeMsgUsecase.New,
			policyUsecase.New,
			privateUserUsecase.New,
			userUsecase.New,
			userVerifyUsecase.New,
		),
		fx.Invoke(
			bannerHttpDelivery.New,
			lotteryHttpDelivery.New,
			blogPostHttpDelivery.New,
			codeHttpDelivery.New,
			commonHttpDelivery.New,
			corporateUserHttpDelivery.New,
			faqHttpDelivery.New,
			healthHttpDelivery.New,
			investmentHttpDelivery.New,
			investmentWaitersHttpDelivery.New,
			myNumberHttpDelivery.New,
			newsHttpDelivery.New,
			noticeMsgHttpDelivery.New,
			policyHttpDelivery.New,
			privateUserHttpDelivery.New,
			userHttpDelivery.New,
			userVerifyHttpDelivery.New,
			registerRouter,
			generateSwagger,
			serve,
		),
	)
}

func generateSwagger(apiRoot echoswagger.ApiRoot, echo *echo.Echo) {
	swagger.GenerateSwaggerIfNeeded(
		apiRoot,
		echo,
		"user-api.dev.energyshares.jp",
		"swagger-user-api.json",
		swagger.NullableProperty{Definition: "UserWaitlistResponseItem", Property: "invest_status"},
		swagger.NullableProperty{Definition: "UserWaitlistResponseItem", Property: "project_status"},
	)
}

func serve(lifecycle fx.Lifecycle, cfg *config.Config, echo *echo.Echo, swaggerUI *openapi.SwaggerUISetup, echoSwagger *echoutils.EchoSwagger) {
	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Debug 모드의 경우에만 수행합니다.
			if cfg.IsDevelopment() {
				echoSwagger.RegisterOpenAPI()

				swaggerUIDir, err := swaggerUI.Setup()
				if err != nil {
					logrus.Errorf("failed to setup Swagger UI: %v", err)
				} else {
					echo.Static("/api/swagger", swaggerUIDir)
					echo.File("/api/swagger", filepath.Join(swaggerUIDir, "index.html"))
				}
			}

			go func() {
				if err := echo.Start(":" + cfg.UserConfig.HttpPort); err != nil {
					logrus.Errorf("fail start server : %v", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logrus.Info("stopping server")
			return nil
		},
	})
}
