package server

import (
	"context"
	"os"
	"path/filepath"
	"reflect"

	"github.com/go-resty/resty/v2"
	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"github.com/swaggest/openapi-go/openapi31"
	"go.uber.org/fx"
	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/cipher"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/echoutils"
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
	attachmentRepository "github.com/illuminarean/energy-shares-backend-repo/repository/attachment/mysql"
	bannerRepository "github.com/illuminarean/energy-shares-backend-repo/repository/banner/mysql"
	bannerSettingRepository "github.com/illuminarean/energy-shares-backend-repo/repository/banner_setting/mysql"
	changelogRepository "github.com/illuminarean/energy-shares-backend-repo/repository/changelog/ksql"
	codeOwnershipRepository "github.com/illuminarean/energy-shares-backend-repo/repository/code_ownership/mysql"
	compositeUserRepository "github.com/illuminarean/energy-shares-backend-repo/repository/composite_user/mysql"
	contentImageRepository "github.com/illuminarean/energy-shares-backend-repo/repository/content_image/mysql"
	contractRepository "github.com/illuminarean/energy-shares-backend-repo/repository/contract/mysql"
	depositRequestRepository "github.com/illuminarean/energy-shares-backend-repo/repository/deposit_request/mysql"
	depositRequestLackHistory "github.com/illuminarean/energy-shares-backend-repo/repository/deposit_requset_lack_history/mysql"
	dividendRepository "github.com/illuminarean/energy-shares-backend-repo/repository/dividend/mysql"
	ekycRequestLogRepository "github.com/illuminarean/energy-shares-backend-repo/repository/ekyc_request_log/mysql"
	eventMessageRepository "github.com/illuminarean/energy-shares-backend-repo/repository/event_message/mysql"
	faqRepository "github.com/illuminarean/energy-shares-backend-repo/repository/faq/mysql"
	investmentRepository "github.com/illuminarean/energy-shares-backend-repo/repository/investment/mysql"
	investmentRequestRepository "github.com/illuminarean/energy-shares-backend-repo/repository/investment_request/mysql"
	investmentWaitHistoryRepository "github.com/illuminarean/energy-shares-backend-repo/repository/investment_wait_history/mysql"
	investmentWaitersRepository "github.com/illuminarean/energy-shares-backend-repo/repository/investment_waiters/mysql"
	loginHistoryMysqlRepository "github.com/illuminarean/energy-shares-backend-repo/repository/login_history/mysql"
	lotteryApplicationRepository "github.com/illuminarean/energy-shares-backend-repo/repository/lottery_application/mysql"
	lotteryHistoryRepository "github.com/illuminarean/energy-shares-backend-repo/repository/lottery_history/mysql"
	mailFailedRepository "github.com/illuminarean/energy-shares-backend-repo/repository/mail_failed/mysql"
	mailRetryRepository "github.com/illuminarean/energy-shares-backend-repo/repository/mail_retry/mysql"
	myNumberRepository "github.com/illuminarean/energy-shares-backend-repo/repository/my_number/mysql"
	newsRepository "github.com/illuminarean/energy-shares-backend-repo/repository/news/mysql"
	noticeMsgRepository "github.com/illuminarean/energy-shares-backend-repo/repository/notice_msg/mysql"
	noticeMsgReadRepository "github.com/illuminarean/energy-shares-backend-repo/repository/notice_msg_read/mysql"
	operatorHistoryKsqlRepository "github.com/illuminarean/energy-shares-backend-repo/repository/operator_history/ksql"
	operatorHistoryRepository "github.com/illuminarean/energy-shares-backend-repo/repository/operator_history/mysql"
	passwordResetRequestRepository "github.com/illuminarean/energy-shares-backend-repo/repository/password_reset_request/mysql"
	policyRepository "github.com/illuminarean/energy-shares-backend-repo/repository/policy/mysql"
	policyHistoryRepository "github.com/illuminarean/energy-shares-backend-repo/repository/policy_history/mysql"
	projectRepository "github.com/illuminarean/energy-shares-backend-repo/repository/project/mysql"
	projectFileRepository "github.com/illuminarean/energy-shares-backend-repo/repository/project_file/mysql"
	projectImageRepository "github.com/illuminarean/energy-shares-backend-repo/repository/project_image/mysql"
	projectInvestmentExecuteRepository "github.com/illuminarean/energy-shares-backend-repo/repository/project_investment_execute/mysql"
	projectMilestoneRepository "github.com/illuminarean/energy-shares-backend-repo/repository/project_milestone/mysql"
	projectNewsRepository "github.com/illuminarean/energy-shares-backend-repo/repository/project_news/mysql"
	projectViewCountRepository "github.com/illuminarean/energy-shares-backend-repo/repository/project_view_count/mysql"
	refreshTokenRepository "github.com/illuminarean/energy-shares-backend-repo/repository/refresh_token/mysql"
	repaymentDividendScheduleRepository "github.com/illuminarean/energy-shares-backend-repo/repository/repayment_dividend_schedule/mysql"
	reportOutputRepository "github.com/illuminarean/energy-shares-backend-repo/repository/report_output/mysql"
	roleRepository "github.com/illuminarean/energy-shares-backend-repo/repository/role/mysql"
	shareRepository "github.com/illuminarean/energy-shares-backend-repo/repository/share/mysql"
	spvRepository "github.com/illuminarean/energy-shares-backend-repo/repository/spv/mysql"
	transactionInfoHistoryRepository "github.com/illuminarean/energy-shares-backend-repo/repository/transaction_info_history/mysql"
	transferRepository "github.com/illuminarean/energy-shares-backend-repo/repository/transfer/mysql"
	userRepository "github.com/illuminarean/energy-shares-backend-repo/repository/user/mysql"
	userAccountRepository "github.com/illuminarean/energy-shares-backend-repo/repository/user_account/mysql"
	userLoginLockRepository "github.com/illuminarean/energy-shares-backend-repo/repository/user_login_lock/mysql"
	userNoticeMappingRepository "github.com/illuminarean/energy-shares-backend-repo/repository/user_notice_mapping/mysql"
	userTransactionRepository "github.com/illuminarean/energy-shares-backend-repo/repository/user_transaction/mysql"
	userVirtualAccountRepository "github.com/illuminarean/energy-shares-backend-repo/repository/user_virtual_account/mysql"
	withdrawalRepository "github.com/illuminarean/energy-shares-backend-repo/repository/withdrawal/mysql"
	withdrawalCloseRepository "github.com/illuminarean/energy-shares-backend-repo/repository/withdrawal_close/mysql"
	withdrawalCloseMappingRepository "github.com/illuminarean/energy-shares-backend-repo/repository/withdrawal_close_mapping/mysql"

	// Service Layer
	adminUserService "github.com/illuminarean/energy-shares-backend-repo/service/admin_user"
	authService "github.com/illuminarean/energy-shares-backend-repo/service/auth"
	depositRequestService "github.com/illuminarean/energy-shares-backend-repo/service/deposit_request"
	eventMessageService "github.com/illuminarean/energy-shares-backend-repo/service/event_message"
	investmentService "github.com/illuminarean/energy-shares-backend-repo/service/investment"
	investmentWaitHistoryService "github.com/illuminarean/energy-shares-backend-repo/service/investment_wait_history"
	investmentWaitersService "github.com/illuminarean/energy-shares-backend-repo/service/investment_waiters"
	lotteryService "github.com/illuminarean/energy-shares-backend-repo/service/lottery"
	mailService "github.com/illuminarean/energy-shares-backend-repo/service/mail"
	myNumberService "github.com/illuminarean/energy-shares-backend-repo/service/my_number"
	nexwayService "github.com/illuminarean/energy-shares-backend-repo/service/nexway"
	policyService "github.com/illuminarean/energy-shares-backend-repo/service/policy"
	projectService "github.com/illuminarean/energy-shares-backend-repo/service/project"
	shareService "github.com/illuminarean/energy-shares-backend-repo/service/share"
	userAccountService "github.com/illuminarean/energy-shares-backend-repo/service/user_account"
	userVirtualAccountService "github.com/illuminarean/energy-shares-backend-repo/service/user_virtual_account"
	withdrawalService "github.com/illuminarean/energy-shares-backend-repo/service/withdrawal"

	// Usecase Layer
	adminUserUsecase "github.com/illuminarean/energy-share-core-api/usecase/admin_user"
	bannerUsecase "github.com/illuminarean/energy-share-core-api/usecase/banner"
	changelogUsecase "github.com/illuminarean/energy-share-core-api/usecase/changelog"
	faqUsecase "github.com/illuminarean/energy-share-core-api/usecase/faq"
	fileUsecase "github.com/illuminarean/energy-share-core-api/usecase/file"
	investmentUsecase "github.com/illuminarean/energy-share-core-api/usecase/investment"
	investmentWaitHistoryUsecase "github.com/illuminarean/energy-share-core-api/usecase/investment_wait_history"
	investmentWaitersUsecase "github.com/illuminarean/energy-share-core-api/usecase/investment_waiters"
	investorReportUsecase "github.com/illuminarean/energy-share-core-api/usecase/investor_report"
	lotteryUsecase "github.com/illuminarean/energy-share-core-api/usecase/lottery"
	mailsUsecase "github.com/illuminarean/energy-share-core-api/usecase/mail"
	maintenanceUsecase "github.com/illuminarean/energy-share-core-api/usecase/maintenance"
	newsUsecase "github.com/illuminarean/energy-share-core-api/usecase/news"
	noticeMsgUsecase "github.com/illuminarean/energy-share-core-api/usecase/notice_msg"
	operatorHistoryUsecase "github.com/illuminarean/energy-share-core-api/usecase/operator_history"
	policyUsecase "github.com/illuminarean/energy-share-core-api/usecase/policy"
	privateUserUsecase "github.com/illuminarean/energy-share-core-api/usecase/private_user"
	projectUsecase "github.com/illuminarean/energy-share-core-api/usecase/project"
	referralCodeUsecase "github.com/illuminarean/energy-share-core-api/usecase/referral_code"
	repaymentDividendUsecase "github.com/illuminarean/energy-share-core-api/usecase/repayment_dividend"
	userUsecase "github.com/illuminarean/energy-share-core-api/usecase/user"
	withdrawalUsecase "github.com/illuminarean/energy-share-core-api/usecase/withdrawal"
	withdrawalCloseUsecase "github.com/illuminarean/energy-share-core-api/usecase/withdrawal_close"

	// Delivery Layer
	adminUserHttpDelivery "github.com/illuminarean/energy-share-core-api/delivery/admin_user/http"
	bannerHttpDelivery "github.com/illuminarean/energy-share-core-api/delivery/banner/http"
	faqHttpDelivery "github.com/illuminarean/energy-share-core-api/delivery/faq/http"
	fileHttpDelivery "github.com/illuminarean/energy-share-core-api/delivery/file/http"
	healthHttpDelivery "github.com/illuminarean/energy-share-core-api/delivery/health/http"
	investmentHttpDelivery "github.com/illuminarean/energy-share-core-api/delivery/investment/http"
	investmentWaitersHttpDelivery "github.com/illuminarean/energy-share-core-api/delivery/investment_waiters/http"
	investorReportHttpDelivery "github.com/illuminarean/energy-share-core-api/delivery/investor_report/http"
	lotteryHttpDelivery "github.com/illuminarean/energy-share-core-api/delivery/lottery/http"
	mailsHttpDelivery "github.com/illuminarean/energy-share-core-api/delivery/mail/http"
	maintenanceHttpDelivery "github.com/illuminarean/energy-share-core-api/delivery/maintenance/http"
	newsHttpDelivery "github.com/illuminarean/energy-share-core-api/delivery/news/http"
	noticeMsgHttpDelivery "github.com/illuminarean/energy-share-core-api/delivery/notice_msg/http"
	operatorHistoryHttpDelivery "github.com/illuminarean/energy-share-core-api/delivery/operator_history/http"
	policyHttpDelivery "github.com/illuminarean/energy-share-core-api/delivery/policy/http"
	privateUserHttpDelivery "github.com/illuminarean/energy-share-core-api/delivery/private_user/http"
	projectHttpDelivery "github.com/illuminarean/energy-share-core-api/delivery/project/http"
	qnaHttpDelivery "github.com/illuminarean/energy-share-core-api/delivery/qna/http"
	referralCodeHttpDelivery "github.com/illuminarean/energy-share-core-api/delivery/referral_code/http"
	repaymentDividnedHttpDelivery "github.com/illuminarean/energy-share-core-api/delivery/repayment_dividend/http"
	userHttpDelivery "github.com/illuminarean/energy-share-core-api/delivery/user/http"
	withdrawalHttpDelivery "github.com/illuminarean/energy-share-core-api/delivery/withdrawal/http"
	withdrawalCloseHttpDelivery "github.com/illuminarean/energy-share-core-api/delivery/withdrawal_close/http"
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
			// 타입이 일치하면, item을 T 타입으로 변환하여 반환합니다.
			// 여기서는 Type Assertion을 사용합니다.
			value, ok := injection.(T)
			if !ok {
				// 타입 변환에 실패한 경우
				return func() T { return *new(T) }
			}
			return func() T { return value }
		}
	}

	return func() T { return *new(T) }
}

func (b *bootstrap) Invoke() *fx.App {
	return fx.New(
		appOptions(b.injections),
		fx.Invoke(
			generateSwagger,
			serve,
		),
	)
}

// appOptions は部品(Provide)とルートの登録(Invoke)。route_roles_test.go もこれでルートを組み立てて権限の表と
// 突き合わせるので、delivery はここに足す(MIR-1408)。サーバーの起動と swagger の書き出しは Invoke() の側に置く。
func appOptions(injections []interface{}) fx.Option {
	return fx.Options(
		fx.Provide(
			inject(injections, &echo.Echo{}),
			inject(injections, &config.Config{}),
			inject(injections, &xorm.Engine{}),
			// pkg injections
			cipher.New,
			jwthelper.NewUserExtractor,
			pwhelper.NewPasswordHelper,
			filehelper.NewFileHelper,
			resty.New,
			func() i18nutils.Translator {
				return i18nutils.New(pathutils.GetPathOfAssets())
			},

			// 향후에는 해당 부분을 삭제하고 openapi 패키지를 이용해서 swagger를 운영하려고 합니다.
			func(cfg *config.Config, e *echo.Echo) echoswagger.ApiRoot {
				return lo.If(cfg.Debug, echoswagger.New(e, "/doc", &echoswagger.Info{
					Title: "Energy Shares Japan Admin API",
					Description: `This is API doc for Energy Shares Japan Admin API.
The displayed version incorporates the commit date and hash for precise version tracking.
여기를 클릭하시면 클린 아키텍처가 적용된 v1 스웨거 페이지로 이동합니다 - [https://admin.dev.energyshares.jp/api/swagger/index.html](https://admin.dev.energyshares.jp/api/swagger/index.html)`,
					Version: "commit: " + os.Getenv("VERSION_INFO"),
				})).Else(echoswagger.NewNop(e))
			},

			func(cfg *config.Config) *openapi31.Reflector {
				reflector := openapi31.NewReflector()
				reflector.Spec.Info.
					WithTitle("Energy Shares Japan Admin V1 API").
					WithVersion("1.0.0").
					WithDescription("Admin API V1 for Energy Shares Japan")
				reflector.Spec.SetHTTPBearerTokenSecurity("Authorization", "JWT", "JWT Access Token을 필요로 합니다")
				reflector.Spec.Servers = []openapi31.Server{}
				return reflector
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
			s3utils.New,
			snsutils.New,
			emailutils.New,
			sessionutils.New,
			idgenerator.New,

			// Repository Layer
			adminUserRepository.New,
			attachmentRepository.New,
			bannerRepository.New,
			bannerSettingRepository.New,
			changelogRepository.New,
			contentImageRepository.New,

			contractRepository.New,
			codeOwnershipRepository.New,
			depositRequestLackHistory.New,
			depositRequestRepository.New,
			dividendRepository.New,
			ekycRequestLogRepository.New,
			eventMessageRepository.New,
			faqRepository.New,
			investmentRepository.New,
			investmentRequestRepository.New,
			investmentWaitHistoryRepository.New,
			investmentWaitersRepository.New,
			loginHistoryMysqlRepository.New,
			mailFailedRepository.New,
			mailRetryRepository.New,
			myNumberRepository.New,
			newsRepository.New,
			noticeMsgRepository.New,
			noticeMsgReadRepository.New,
			operatorHistoryKsqlRepository.New,
			operatorHistoryRepository.New,
			passwordResetRequestRepository.New,
			policyRepository.New,
			policyHistoryRepository.New,
			projectFileRepository.New,
			projectImageRepository.New,
			projectInvestmentExecuteRepository.New,
			projectMilestoneRepository.New,
			projectNewsRepository.New,
			projectRepository.New,
			projectViewCountRepository.New,
			reportOutputRepository.New,
			refreshTokenRepository.New,
			repaymentDividendScheduleRepository.New,
			roleRepository.New,
			shareRepository.New,
			spvRepository.New,
			transactionInfoHistoryRepository.New,
			transferRepository.New,
			userAccountRepository.New,
			compositeUserRepository.New,
			userLoginLockRepository.New,
			userNoticeMappingRepository.New,
			userRepository.New,
			userTransactionRepository.New,
			userVirtualAccountRepository.New,
			withdrawalRepository.New,
			withdrawalCloseRepository.New,
			withdrawalCloseMappingRepository.New,
			lotteryApplicationRepository.New,
			lotteryHistoryRepository.New,

			// Service Layer
			adminUserService.New,
			authService.New,
			depositRequestService.New,
			eventMessageService.New,
			investmentService.New,
			investmentWaitHistoryService.New,
			investmentWaitersService.New,
			lotteryService.New,
			mailService.New,
			myNumberService.New,
			nexwayService.New,
			policyService.New,
			projectService.New,
			shareService.New,
			userAccountService.New,
			userVirtualAccountService.New,
			withdrawalService.New,

			// Usecase Layer
			adminUserUsecase.New,
			lotteryUsecase.New,
			bannerUsecase.New,
			changelogUsecase.New,
			faqUsecase.New,
			fileUsecase.New,
			investmentUsecase.New,
			investmentWaitHistoryUsecase.New,
			investmentWaitersUsecase.New,
			mailsUsecase.New,
			maintenanceUsecase.New,
			newsUsecase.New,
			referralCodeUsecase.New,
			noticeMsgUsecase.New,
			operatorHistoryUsecase.New,
			policyUsecase.New,
			privateUserUsecase.New,
			projectUsecase.New,
			repaymentDividendUsecase.New,
			investorReportUsecase.New,
			userUsecase.New,
			withdrawalUsecase.New,
			withdrawalCloseUsecase.New,
		),
		fx.Invoke(
			adminUserHttpDelivery.New,
			bannerHttpDelivery.New,
			faqHttpDelivery.New,
			fileHttpDelivery.New,
			healthHttpDelivery.New,
			investorReportHttpDelivery.New,
			investmentHttpDelivery.New,
			investmentWaitersHttpDelivery.New,
			mailsHttpDelivery.New,
			maintenanceHttpDelivery.New,
			newsHttpDelivery.New,
			referralCodeHttpDelivery.New,
			noticeMsgHttpDelivery.New,
			operatorHistoryHttpDelivery.New,
			policyHttpDelivery.New,
			privateUserHttpDelivery.New,
			projectHttpDelivery.New,
			qnaHttpDelivery.New,
			repaymentDividnedHttpDelivery.New,
			userHttpDelivery.New,
			withdrawalHttpDelivery.New,
			withdrawalCloseHttpDelivery.New,
			lotteryHttpDelivery.New,

			registerRouter,
		),
	)
}

// generateSwagger と serve ではルートを足さない。appOptions の外なので route_roles_test.go の表に出てこず、権限の
// 確認が無いルートを足しても気づけない(MIR-1408)。開発環境・staging の /api/swagger* だけは例外。
func generateSwagger(apiRoot echoswagger.ApiRoot, echo *echo.Echo) {
	swagger.GenerateSwaggerIfNeeded(apiRoot, echo, "admin-api.dev.energyshares.jp", "swagger-admin-api.json")
}

func serve(lifecycle fx.Lifecycle, cfg *config.Config, echo *echo.Echo, swaggerUI *openapi.SwaggerUISetup, echoSwagger *echoutils.EchoSwagger) {
	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {

			if cfg.IsDevelopment() {
				echoSwagger.RegisterOpenAPI()

				swaggerUIDir, err := swaggerUI.Setup()
				if err != nil {
					logrus.Errorf("failed to setup Swagger UI: %v", err)
				} else {
					echo.Static("/api/swagger", swaggerUIDir)
					echo.File(cfg.Swagger.OpenApiSpecPath, filepath.Join(swaggerUIDir, "openapi.yaml"))
				}
			}

			go func() {
				if err := echo.Start(":" + cfg.AdminConfig.HttpPort); err != nil {
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
