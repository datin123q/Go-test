package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/XSAM/otelsql"
	"github.com/getsentry/sentry-go"
	sentryotel "github.com/getsentry/sentry-go/otel"
	"github.com/hillfolk/goutils/ctxdb"
	"github.com/hillfolk/goutils/kafka"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
	"xorm.io/xorm/core"
	xormlog "xorm.io/xorm/log"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/echoutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/logger"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/pathutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/validator"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	_ "github.com/go-sql-driver/mysql"
	"github.com/hillfolk/goutils/cronjob"
	"github.com/hillfolk/goutils/echomiddleware"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"xorm.io/xorm"

	"energyshares-job-scheduler/controllers"
	"energyshares-job-scheduler/jobs"
	"energyshares-job-scheduler/pkg/awsutils"
	"energyshares-job-scheduler/pkg/gmo_aozora"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/contract"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/eventbus"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/s3utils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/sessionutils"
)

var (
	Version     string
	Commit      string
	AssetsPath  string
	jstLocation = mustLoadJSTLocation()
)

const lotteryCSVUploadJobName = "LotteryCSVUploadJob"
const rePublishJobName = "RePublishJob"

// jpkiSignatureRecoveryJobName は、JPKI方式の署名結果の通知(Webhook)を取りこぼした記録を拾い直すジョブ(MIR-1470)。
const jpkiSignatureRecoveryJobName = "JpkiSignatureRecoveryJob"

func RunServer(env string) {
	rootPath, err := os.Getwd()
	if err != nil {
		logrus.Error("fail to os.Getwd() ", err)
		panic(err)
	}

	AssetsPath = pathutils.GetPathOfAssets()

	logrus.WithFields(logrus.Fields{
		"Environment": env,
		"Version":     Version + Commit,
		"Starting":    jsonutils.JSONTime(time.Now()).ToString(),
		"RootPath":    rootPath,
	}).Info("ENV_INFO")

	var c config.Config
	if err := c.Read(env); err != nil {
		logrus.Error(err)
		panic(err)
	}

	// Sentry Setup
	if err := sentry.Init(sentry.ClientOptions{
		Dsn:                c.Observability.JobScheduler.SentryDSN,
		EnableTracing:      true,
		TracesSampleRate:   c.Observability.JobScheduler.SampleRate,
		AttachStacktrace:   true,
		Environment:        c.Env,
		Debug:              false,
		Release:            Version + Commit,
		IgnoreTransactions: []string{"/api/health-check"},
	}); err != nil {
		logrus.Error(errors.Wrap(err, "fail init sentry"))
	} else {
		logrus.Info("Sentry initialized")
		// JobサーバーはErrorレベルのエラーが多すぎるため停止。Cronジョブのラップ関数側でキャッチのみとする。
		// logrus.AddHook(&factory.SentryHook{})
		tp := sdktrace.NewTracerProvider(
			sdktrace.WithSpanProcessor(sentryotel.NewSentrySpanProcessor()),
			sdktrace.WithResource(resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceName("miralita-job-scheduler-backend"),
				semconv.DeploymentEnvironment(env),
			)),
		)
		defer func() { _ = tp.Shutdown(context.Background()) }()
		otel.SetTracerProvider(tp)
		otel.SetTextMapPropagator(sentryotel.NewSentryPropagator())
	}
	defer sentry.Flush(2 * time.Second)

	db, err := initDB(c)
	if err != nil {
		logrus.Error(err)
		panic(err)
	}
	defer db.Close()

	myDB, err := InitMyNumberDB(c)
	if err != nil {
		logrus.Error(err)
		panic(err)
	}
	defer myDB.Close()

	sess, err := initAws(c.AWS.Region)
	if err != nil {
		logrus.Error(err)
		panic(err)
	}

	// S3 버킷명 지정
	s3utils.InitializeS3BucketWithEkyc(c.AWS.S3_PRIVATE_BUCKET, c.AWS.S3_PUBLIC_BUCKET, c.AWS.S3_EKYC_IMAGES_BUCKET)
	// 서비스 정보 초기화
	if err := initServiceUtils(c); err != nil {
		logrus.Error(errors.Wrap(err, "fail init service utils"))
		panic(err)
	}
	// 다국어 언어팩 파일 초기화
	if err := i18nutils.InitI18nFiles(AssetsPath); err != nil {
		logrus.Errorf("failed i18n file loads :  %v ", err)

		panic(err)
	}

	InitializePdf(AssetsPath)

	InitBlogApiService(c.JobSchedulerConfig.BlogApiService)

	InitAttendanceReminderSlack(c.JobSchedulerConfig.Slack)

	if err := InitBankAccounts(c, &gmo_aozora.BankAccounts); err != nil {
		logrus.Errorf("failed gmo client info loads :  %v ", err)
		panic(err)
	}

	if err := gmo_aozora.LoadClient(db, c.Env, c); err != nil {
		logrus.Errorf("failed gmo client info loads :  %v ", err)
		panic(err)
	}
	// WebHook 구독
	if err := gmo_aozora.WebhookSubscription(gmo_aozora.START_SUBSCRIPTION); err != nil {
		logrus.Errorf("Failed to subscribe GMO Aozora webhook: %v ", err)
	}

	// Event Bus 초기화
	eventbus.InitializeEventBus(c.EventBus.Config, 60)

	e := NewRouter(c.Debug)
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Running Scheduler Server!")
	})

	// init & run cron jobs
	cronJobs, err := initJob(c, db, myDB, sess)
	if err != nil {
		logrus.Errorf("failed init cron jobs :  %v ", err)
		panic(err)
	}
	// MIR-1329: 以前は":8083"を直書きしていた。ZOLOME側・worktreeでは別ポートが
	// 必要になるため、設定値から読む。
	cronHttpPort := c.JobSchedulerConfig.CronHttpPort
	if cronHttpPort == "" {
		cronHttpPort = "8083"
	}
	go cronJobs.Start(":" + cronHttpPort)

	e.Pre(middleware.RemoveTrailingSlash())
	e.Pre(echoutils.CustomContext())
	e.Use(echoutils.CustomRecover())

	// 디버그 옵션에 따라서 추가적인 Context Field 추가를 위한 미들웨어
	e.Use(echoutils.CustomDebugContext(c.Debug)...)

	// 커스텀 로그를 위한 미들웨어
	e.Use(echoutils.CustomLogger())

	e.Use(contextDBWithName(c.JobSchedulerConfig.Service, factory.ConextDBType, db, c.Database.Logger.Kafka))
	e.Use(contextDBWithName(c.JobSchedulerConfig.Service, factory.ConextMyNumberDBType, myDB, c.MyNumberDatabase.Logger.Kafka))

	e.Use(echomiddleware.ContextAws(sess))
	e.Use(sessionutils.ContextAwsSes(sess))
	e.Use(sessionutils.ContextAwsSns(sess))

	e.HTTPErrorHandler = echoutils.CustomErrorHandler

	e.Validator = validator.NewCustomValidator()
	e.Debug = c.Debug

	if err := e.Start(":" + c.JobSchedulerConfig.HttpPort); err != nil {
		logrus.Error(err)
		panic(err)
	}
}

// InitializePdf 契約書面PDF生成用のフォントディレクトリを初期化する
// (MIR-1201でUniPDFのライセンス設定を廃止し、フォント初期化だけが残った)。
func InitializePdf(path string) {
	if err := contract.SetDefaultFont(path); err != nil {
		logrus.Error(errors.Wrap(err, "failed to set default font"))
		panic(err)
	}
}

func InitBankAccounts(accountConfig config.Config, items *gmo_aozora.Accounts) error {
	if len(*items) != 0 {
		return errors.New("already initialized bank accounts")
	}

	a := gmo_aozora.Account{}.New(
		accountConfig.JobSchedulerConfig.BankAccounts.EquityDeposit.AccountNumber,
		accountConfig.JobSchedulerConfig.BankAccounts.EquityDeposit.AccountHolder,
		accountConfig.JobSchedulerConfig.BankAccounts.EquityDeposit.BankCode,
		accountConfig.JobSchedulerConfig.BankAccounts.EquityDeposit.BranchCode,
		gmo_aozora.EQUITY_DEPOSIT,
	)
	*items = append(*items, a)

	a = gmo_aozora.Account{}.New(
		accountConfig.JobSchedulerConfig.BankAccounts.EquityWithdrawal.AccountNumber,
		accountConfig.JobSchedulerConfig.BankAccounts.EquityWithdrawal.AccountHolder,
		accountConfig.JobSchedulerConfig.BankAccounts.EquityWithdrawal.BankCode,
		accountConfig.JobSchedulerConfig.BankAccounts.EquityWithdrawal.BranchCode,
		gmo_aozora.EQUITY_WITHDRAWAL,
	)
	*items = append(*items, a)

	loanDeposit := gmo_aozora.Account{}.New(
		accountConfig.JobSchedulerConfig.BankAccounts.LoanDeposit.AccountNumber,
		accountConfig.JobSchedulerConfig.BankAccounts.LoanDeposit.AccountHolder,
		accountConfig.JobSchedulerConfig.BankAccounts.LoanDeposit.BankCode,
		accountConfig.JobSchedulerConfig.BankAccounts.LoanDeposit.BranchCode,
		gmo_aozora.LOAN_DEPOSIT,
	)
	*items = append(*items, loanDeposit)

	loanWithdrawal := gmo_aozora.Account{}.New(
		accountConfig.JobSchedulerConfig.BankAccounts.LoanWithdrawal.AccountNumber,
		accountConfig.JobSchedulerConfig.BankAccounts.LoanWithdrawal.AccountHolder,
		accountConfig.JobSchedulerConfig.BankAccounts.LoanWithdrawal.BankCode,
		accountConfig.JobSchedulerConfig.BankAccounts.LoanWithdrawal.BranchCode,
		gmo_aozora.LOAN_WITHDRAWAL,
	)
	*items = append(*items, loanWithdrawal)

	return nil
}

// exists returns whether the given file or directory exists or not
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func initDB(c config.Config) (*xorm.Engine, error) {
	sqlDB, err := otelsql.Open(
		c.Database.Driver,
		c.Database.Connection,
		otelsql.WithTracerProvider(otel.GetTracerProvider()),
		otelsql.WithAttributes(semconv.DBSystemMySQL),
	)
	if err != nil {
		return nil, errors.Wrap(err, "init instrumented database")
	}
	d := core.FromDB(sqlDB)
	db, err := xorm.NewEngineWithDB(
		c.Database.Driver,
		c.Database.Connection,
		d,
	)
	if err != nil {
		return nil, err
	}

	db.ShowSQL(c.Debug)
	db.SetLogger(logger.NewXorm())
	db.SetLogLevel(xormlog.LOG_DEBUG)
	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(60 * time.Second)

	return db, nil
}

// Jobs 을 초기화 한다.
func initJob(c config.Config, db *xorm.Engine, mydb *xorm.Engine, sess *session.Session) (*cronjob.Cron, error) {

	cj := newCronWithLocation(c.JobSchedulerConfig.Service, jstLocation)
	// Add Sentry/OTel tracing middleware before context middlewares
	cj.Use(factory.CronSentryMiddleware())
	// Add ContextDB middleware
	cj.Use(cronContextDBWithName(c.JobSchedulerConfig.Service, factory.ConextDBType, db, c.Database.Logger.Kafka))
	cj.Use(cronContextDBWithName(c.JobSchedulerConfig.Service, factory.ConextMyNumberDBType, mydb, c.Database.Logger.Kafka))
	// Add AWS middleware
	cj.Use(awsutils.ContextAws(sess))
	cj.Use(awsutils.ContextAwsSes(sess))
	cj.Use(awsutils.ContextAwsSns(sess))

	/* 변경시 아래 문서를 업데이트 하세요.
	   https://docs.google.com/spreadsheets/d/1eKQbZG7L8J42TT318_Ie8EUfVTQgAmBU4NffhwOHdfg/edit#gid=0
	*/

	aozora := &jobs.AozoraJobs{}

	// GMOあおぞら銀行 トークン更新ジョブ
	scheduleCronJob(c.JobSchedulerConfig.Cron, cj, "RefreshTokenJob", aozora.RefreshToken)

	// GMOあおぞら銀行 口座情報取得ジョブ
	scheduleCronJob(c.JobSchedulerConfig.Cron, cj, "GetAccountInfoJob", aozora.GetAccountInfo)

	// GMOあおぞら銀行 総合振込ジョブ
	scheduleCronJob(c.JobSchedulerConfig.Cron, cj, "BulkTransferJob", aozora.BulkTransferJob)

	// GMOあおぞら銀行 振込結果確認ジョブ
	scheduleCronJob(c.JobSchedulerConfig.Cron, cj, "BulkTransferResultCheckJob", aozora.BulkTransferResultCheckJob)

	// GMOあおぞら銀行 一括振込状況確認ジョブ
	scheduleCronJob(c.JobSchedulerConfig.Cron, cj, "BulkTransferStatusCheckJob", aozora.BulkTransferStatusCheckJob)

	// GMOあおぞら銀行 仮想口座入金確認ジョブ
	scheduleCronJob(c.JobSchedulerConfig.Cron, cj, "GetEquityVaTransactionJob", aozora.GetEquityVaTransactionJob)

	// 가상계좌 자동 발행 배치: 매일 9시 - STG & PROD 가상계좌 수량 부족으로 인해 해당 기능 비활성화
	// factory.CronJobWrapperWithSentry(cj, "IssueVirtualAccountJob", dailyAt0900, aozora.IssueVirtualAccountJob)

	waitingInvestment := &jobs.WaitingInvestmentJobs{}

	// 投資待機自動投資配置ジョブ
	scheduleCronJob(c.JobSchedulerConfig.Cron, cj, "AutoInvestmentJob", waitingInvestment.AutoInvestmentJob)

	// 投資待機者への投資要請期限設定ジョブ
	scheduleCronJob(c.JobSchedulerConfig.Cron, cj, "CheckInvestmentProposeJob", waitingInvestment.CheckInvestmentProposeJob)

	// ファンド状況更新ジョブ
	scheduleCronJob(c.JobSchedulerConfig.Cron, cj, "FundingOpenJob", jobs.ProductJobs{}.FundingOpenJob)

	// ファンド状態更新募集終了ジョブ
	scheduleCronJob(c.JobSchedulerConfig.Cron, cj, "FundingCloseJob", jobs.ProductJobs{}.FundingCloseJob)

	// ファンド配当/償還実施ジョブ
	scheduleCronJob(c.JobSchedulerConfig.Cron, cj, "RepaymentDividendJob", jobs.RepaymentDividendJobs{}.RepaymentDividendProcess)

	// 未使用入金確認ジョブ（開発用）
	scheduleCronJob(c.JobSchedulerConfig.Cron, cj, "VaDepositNotUseItemCheckJob", jobs.DepositCheckJobs{}.VaDepositNotUseItemCheckJob)

	// 入金依頼期限メールのバッチ処理ジョブ
	scheduleCronJob(c.JobSchedulerConfig.Cron, cj, "DepositRequestExpireJob", jobs.DepositCheckJobs{}.DepositRequestExpireJob)

	// 未使用預金の自動投資設定ジョブ
	scheduleCronJob(c.JobSchedulerConfig.Cron, cj, "DepositAutoProcessJob", jobs.DepositCheckJobs{}.DepositAutoProcessJob)

	// ユーザー住所確認ジョブ
	scheduleCronJob(c.JobSchedulerConfig.Cron, cj, "UserAddressCheckJob", jobs.UserAddressCheckJobs{}.UserAddressCheckJob)

	// マイナンバー削除ジョブ
	scheduleCronJob(c.JobSchedulerConfig.Cron, cj, "MyNumberDeleteFileJob", jobs.MyNumberJobs{}.MyNumberDeleteFileJob)

	// 失敗したメール送信リクエストの再リクエストジョブ
	scheduleCronJob(c.JobSchedulerConfig.Cron, cj, "MailRequestRetryJob", jobs.MailRetryJobs{}.MailRequestRetryJob)

	// 契約時書面作成ジョブ
	scheduleCronJob(c.JobSchedulerConfig.Cron, cj, "ContractCreateJob", jobs.ContractJobs{}.ContractCreateJob)

	// 投資家取引残高報告ジョブ
	scheduleCronJob(c.JobSchedulerConfig.Cron, cj, "CreateInvestorTransactionBalanceReportJob", jobs.ReportJobs{}.CreateInvestorTransactionBalanceReportJob)

	// 投資家支払調書作成ジョブ
	scheduleCronJob(c.JobSchedulerConfig.Cron, cj, "CreateInvestorPaymentStatementJob", jobs.ReportJobs{}.CreateInvestorPaymentStatementJob)

	// 税務署用支払調書作成ジョブ
	scheduleCronJob(c.JobSchedulerConfig.Cron, cj, "CreateTaxOfficePaymentStatementJob", jobs.ReportJobs{}.CreateTaxOfficePaymentStatementJob)

	// 税務署用支払調書合計表作成ジョブ
	scheduleCronJob(c.JobSchedulerConfig.Cron, cj, "CreateTaxOfficePaymentStatementSumJob", jobs.ReportJobs{}.CreateTaxOfficePaymentStatementSumJob)

	// ブログポストリクエストジョブ
	scheduleCronJob(c.JobSchedulerConfig.Cron, cj, "BlogPostRequest", jobs.BlogPostJobs{}.BlogPostRequest)

	// 祝日カレンダー作成ジョブ
	scheduleCronJob(c.JobSchedulerConfig.Cron, cj, "CreateHolidayCalendarJob", jobs.CalendarJobs{}.CreateHolidayCalendarJob)

	// 期間終了バナー未掲載処理ジョブ
	scheduleCronJob(c.JobSchedulerConfig.Cron, cj, "CloseBannersJob", jobs.BannerJobs{}.CloseBannersJob)

	// 勤怠連絡リマインドジョブ
	scheduleCronJob(c.JobSchedulerConfig.Cron, cj, "AttendanceReminderJob", jobs.AttendanceReminderJobs{}.AttendanceReminderJob)

	// 勤怠連絡リマインドジョブ(出勤日当日の朝)
	scheduleCronJob(c.JobSchedulerConfig.Cron, cj, "AttendanceMorningReminderJob", jobs.AttendanceReminderJobs{}.AttendanceMorningReminderJob)

	// QA, STG テスト User データ削除ジョブ
	// 他のジョブと重複しないように毎月1日の午前1時に設定
	scheduleCronJob(c.JobSchedulerConfig.Cron, cj, "DeleteTestUsersJob", jobs.TestUsersJobs{}.DeleteTestUsersJob)

	// 抽選CSV一括当落ジョブ
	scheduleCronJob(c.JobSchedulerConfig.Cron, cj, lotteryCSVUploadJobName, jobs.LotteryCSVUploadJobs{}.ProcessQueuedLotteryCSVUploadJob)

	// 未送信イベント再送ジョブ
	scheduleCronJob(c.JobSchedulerConfig.Cron, cj, rePublishJobName, jobs.EventMessageJobs{}.RePublishJob)

	// JPKI方式 署名結果の拾い直しジョブ(MIR-1470)
	// 検証結果通知(Webhook)の再送が尽きても届かなかった記録を、検証結果取得APIで確かめ直す。
	// 設定(cronの JpkiSignatureRecoveryJob)が無いとジョブは登録されない。
	scheduleCronJob(c.JobSchedulerConfig.Cron, cj, jpkiSignatureRecoveryJobName, controllers.NewNexwayJpkiControllers().RecoverStuckSignatureResults)

	return cj, nil
}

func newCronWithLocation(service string, loc *time.Location) *cronjob.Cron {
	originalLocal := time.Local
	time.Local = loc
	defer func() {
		time.Local = originalLocal
	}()

	return cronjob.New(service)
}

func mustLoadJSTLocation() *time.Location {
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return time.FixedZone("Asia/Tokyo", 9*60*60)
	}
	return loc
}

func getCronExpr(cronMap map[string]string, jobName string) string {
	if cronMap == nil {
		return ""
	}
	for key, expr := range cronMap {
		if strings.EqualFold(key, jobName) {
			if trimmed := strings.TrimSpace(expr); trimmed != "" {
				return trimmed
			}
			break
		}
	}
	return ""
}

// jobNeedsSkipIfStillRunning は、前回の実行がまだ終わっていない間は次のtickを
// スキップする保護(skipIfStillRunning)が必要なジョブかどうかを返す。RePublishJob
// (MIR-1114)は実行が長引くと同じイベントを重複して送信するため対象に加えている。
// JpkiSignatureRecoveryJob(MIR-1470)は、ネクスウェイへの問い合わせが遅いと実行が長引き、
// 重なった実行が同じ記録を同時に処理して、ネクスウェイへ二重に問い合わせるため対象に加えている
// (記録の更新は条件付きで、会員へのメールが二重になることは無い)。
func jobNeedsSkipIfStillRunning(jobName string) bool {
	return jobName == lotteryCSVUploadJobName || jobName == rePublishJobName || jobName == jpkiSignatureRecoveryJobName
}

func scheduleCronJob(cronMap map[string]string, cj *cronjob.Cron, jobName string, f func(ctx context.Context) error) {
	if spec := getCronExpr(cronMap, jobName); spec != "" {
		if jobNeedsSkipIfStillRunning(jobName) {
			f = skipIfStillRunning(jobName, f)
		}
		cj.AddAction(jobName, spec, f)
	} else {
		logrus.WithFields(logrus.Fields{
			"jobName": jobName,
			"time":    time.Now(),
		}).Info("Skipping cron job as no schedule is defined")
	}
}

func skipIfStillRunning(jobName string, next func(ctx context.Context) error) func(ctx context.Context) error {
	var mu sync.Mutex
	var running bool

	return func(ctx context.Context) error {
		mu.Lock()
		if running {
			mu.Unlock()
			logrus.WithField("jobName", jobName).Info("Skipping cron job because previous run is still in progress")
			return nil
		}
		running = true
		mu.Unlock()

		defer func() {
			mu.Lock()
			running = false
			mu.Unlock()
		}()

		return next(ctx)
	}
}

// cronjob.ContextDBはSessionにContextを渡していないため、独自関数を作成。（github.com/hillfolk/goutils/cronjob）
func cronContextDBWithName(service string, contextDBName echomiddleware.ContextDBType, xormEngine *xorm.Engine, kafkaConfig kafka.Config) cronjob.Middleware {
	db := ctxdb.New(xormEngine, service, kafkaConfig)

	return func(job, action string, next cronjob.HandlerFunc) cronjob.HandlerFunc {
		return func(ctx context.Context) error {
			session := db.NewSession(ctx).Context(ctx)
			defer session.Close()

			ctx = context.WithValue(ctx, contextDBName, session)
			return next(ctx)
		}
	}
}

// AWS 초기화 및 설정
func initAws(region string) (*session.Session, error) {

	// Initialize a session in us-west-2 that the SDK will use to load
	// credentials from the shared credentials file ~/.aws/credentials.

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region)},
	)
	return sess, err
}

// 서비스 공용 정보 초기화
func initServiceUtils(c config.Config) error {
	serviceutils.SetServiceUrlConfig(c.ServiceConfig.Url)
	serviceutils.SetServiceLocale(c.Locale)
	serviceutils.SetServiceVersion(Version)
	serviceutils.SetMailConfig(c.ServiceConfig.Mail)
	serviceutils.SetCountryCode(c.CountryCode)
	serviceutils.SetNexwayConfig(c.Nexway)
	serviceutils.SetNexwayJpkiConfig(c.Nexway)
	serviceutils.SetBankCodeJPConfig(c.BankCodeJP)
	serviceutils.SetServiceAccountConfig(c.OperatingCompany)
	serviceutils.SetCDN(c.AWS.CDN)
	serviceutils.SetPublicKeyId(c.AWS.PUBLIC_KEY_ID)
	serviceutils.SetPrivateKeyPath(c.AWS.PRIVATE_KEY_PATH)
	serviceutils.SetStatutoryBooksServiceUrl(c.StatutoryBooksService)
	serviceutils.SetEmailServiceUrl(c.EmailService.Url)
	serviceutils.SetEnv(c.Env)

	if err := serviceutils.PreloadPrivateKey(c.Env, c.AWS.Region); err != nil {
		return err
	}

	return nil
}

func InitBlogApiService(blogSvc struct {
	Url string
	Key string
}) {
	serviceutils.SetBlogApiServiceUrl(blogSvc.Url)
	serviceutils.SetBlogApiServiceKey(blogSvc.Key)
}

func InitAttendanceReminderSlack(slackConfig struct {
	AttendanceReminderWebhookUrl string
}) {
	serviceutils.SetAttendanceReminderSlackWebhookUrl(slackConfig.AttendanceReminderWebhookUrl)
}

func InitMyNumberDB(c config.Config) (*xorm.Engine, error) {

	db, err := xorm.NewEngine(c.MyNumberDatabase.Driver, c.MyNumberDatabase.Connection)

	if err != nil {
		return nil, err
	}

	db.ShowSQL(c.Debug)
	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(60 * time.Second)
	return db, nil
}

func contextDBWithName(service string, contextDBName echomiddleware.ContextDBType, xormEngine *xorm.Engine, kafkaConfig kafka.Config) echo.MiddlewareFunc {
	db := ctxdb.New(xormEngine, service, kafkaConfig)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			ctx := req.Context()

			session := db.NewSession(ctx).Context(ctx)
			defer session.Close()

			c.SetRequest(req.WithContext(context.WithValue(ctx, contextDBName, session)))

			switch req.Method {
			case "POST", "PUT", "DELETE", "PATCH":
				if err := session.Begin(); err != nil {
					fmt.Errorf("%w", err)
				}
				if err := next(c); err != nil {
					session.Rollback()
					return err
				}
				if c.Response().Status >= 500 {
					session.Rollback()
					return nil
				}
				if err := session.Commit(); err != nil {
					return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
				}
			default:
				return next(c)
			}

			return nil
		}
	}
}
