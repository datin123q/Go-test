package server

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/XSAM/otelsql"
	"github.com/getsentry/sentry-go"
	"github.com/hillfolk/goutils/ctxdb"
	"github.com/hillfolk/goutils/kafka"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/awsutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/echoutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/eventbus"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/logger"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/pathutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/validator"

	"xorm.io/xorm/core"
	xorm_log "xorm.io/xorm/log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	_ "github.com/go-sql-driver/mysql"
	"github.com/hillfolk/goutils/echomiddleware"
	"github.com/hillfolk/goutils/jwtutil"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"xorm.io/xorm"

	sentryotel "github.com/getsentry/sentry-go/otel"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/pdfgen"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/contract"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/s3utils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/sessionutils"
)

var (
	Version    string
	Commit     string
	AssetsPath string
)

func RunServer(env string) {
	logrus.WithFields(logrus.Fields{
		"Environment": env,
		"Version":     Version + Commit,
		"Starting":    jsonutils.JSONTime(time.Now()).ToString(),
	}).Info("ENV_INFO")

	AssetsPath = pathutils.GetPathOfAssets()

	ctx := context.Background()

	var cfg config.Config
	if err := cfg.Read(env); err != nil {
		logrus.Error(errors.Wrap(err, "fail read config"))
		panic(err)
	}

	// Sentry Setup
	if err := sentry.Init(sentry.ClientOptions{
		Dsn:                cfg.Observability.UserAPI.SentryDSN,
		EnableTracing:      true,
		TracesSampleRate:   cfg.Observability.UserAPI.SampleRate,
		AttachStacktrace:   true,
		Environment:        cfg.Env,
		Debug:              false,
		Release:            Version + Commit,
		IgnoreTransactions: []string{"/api/health-check"},
	}); err != nil {
		logrus.Error(errors.Wrap(err, "fail init sentry"))
	} else {
		logrus.Info("Sentry initialized")
		logrus.AddHook(&factory.SentryHook{})
		tp := sdktrace.NewTracerProvider(
			sdktrace.WithSpanProcessor(sentryotel.NewSentrySpanProcessor()),
			sdktrace.WithResource(resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceName("miralita-user-backend"),
				semconv.DeploymentEnvironment(env),
			)),
		)
		defer func() { _ = tp.Shutdown(context.Background()) }()
		otel.SetTracerProvider(tp)
		otel.SetTextMapPropagator(sentryotel.NewSentryPropagator())
	}
	defer sentry.Flush(2 * time.Second)

	InitializePdf(AssetsPath)

	// DBテーブル初期化
	db, err := InitDB(cfg)
	if err != nil {
		logrus.Error(err)
		panic(err)
	}
	defer db.Close()

	myNumberDb, err := InitMyNumberDB(cfg)
	if err != nil {
		logrus.Error(err)
		panic(err)
	}
	defer myNumberDb.Close()

	sess, err := initAws(cfg.AWS.Region, cfg.IsLocal(), cfg.AWS.BASE_ENDPOINT)
	if err != nil {
		logrus.Error(err)
		panic(err)
	}
	ses, err := initAws(cfg.AWS.SES_Region, cfg.IsLocal(), cfg.AWS.BASE_ENDPOINT)
	if err != nil {
		logrus.Error(err)
		panic(err)
	}
	sns, err := initAws(cfg.AWS.SNS_Region, cfg.IsLocal(), cfg.AWS.BASE_ENDPOINT)
	if err != nil {
		logrus.Error(err)
		panic(err)
	}

	// AWS設定
	awsCfg, err := awsutils.InitAwsV2(ctx, cfg.AWS.Region)
	if err != nil {
		logrus.Error(err)
		panic(err)
	}

	// SQSクライアント設定
	sqsClient := awsutils.InitSQSClient(awsCfg, cfg.IsLocal(), cfg.CommonConfig.AWS.BASE_ENDPOINT)

	// Event Bus 初期化
	defaultEventBus := eventbus.InitializeEventBus(cfg.EventBus.Config, 60)
	eventBusMap := map[eventbus.ContextEventBusType]eventbus.IEventBus{
		eventbus.ContextStatutoryLedgerEventBus: defaultEventBus,
	}
	// S3バケット名指定
	s3utils.InitializeS3Bucket(cfg.AWS.S3_PRIVATE_BUCKET, cfg.AWS.S3_PUBLIC_BUCKET)
	// サービス情報初期化
	if err := initServiceUtils(cfg); err != nil {
		logrus.Error(errors.Wrap(err, "fail init service utils"))
		panic(err)
	}

	// 多言語言語パックファイル初期化
	if err := i18nutils.InitI18nFiles(AssetsPath); err != nil {
		logrus.Error(err)
		panic(err)
	}

	jwtutil.SetJwtSecret(cfg.Jwt.Secret)
	e := echo.New()

	e.Pre(middleware.RemoveTrailingSlash())
	e.Pre(echoutils.CustomContext())
	e.Use(otelecho.Middleware("miralita-user-backend"))
	e.Use(echoutils.CustomRecover())
	e.Use(echoutils.RefreshTokenCookieExtend())

	// デバッグオプションに応じた追加的なContextField追加用ミドルウェア
	e.Use(echoutils.CustomDebugContext(cfg.Debug)...)

	// カスタムログ用ミドルウェア
	e.Use(echoutils.CustomLogger())

	e.Use(contextDBWithName(cfg.UserConfig.Service, factory.ConextDBType, db, cfg.Database.Logger.Kafka, eventBusMap))
	e.Use(contextDBWithName(cfg.UserConfig.Service, factory.ConextMyNumberDBType, myNumberDb, cfg.MyNumberDatabase.Logger.Kafka, eventBusMap))

	e.Use(echomiddleware.ContextAws(sess))
	e.Use(sessionutils.ContextAwsSns(sns))
	e.Use(sessionutils.ContextAwsSes(ses))

	// CORS Setting
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{http.MethodGet, http.MethodOptions, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete},
		AllowHeaders:     []string{"*", echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderXCSRFToken, echo.HeaderAuthorization, echo.HeaderXRequestedWith, "locale", "x-xsrf-token"},
		AllowCredentials: true,
	}))

	e.Use(middleware.CSRFWithConfig(
		middleware.CSRFConfig{
			Skipper: func(c echo.Context) bool {
				if cfg.IsDevelopment() {
					return true
				}

				if c.Path() == "/api/adebis/itp" {
					return true
				}

				return false
			},
			TokenLength: 32,
			TokenLookup: "header:X-XSRF-TOKEN",
			ContextKey:  "csrf",
			CookieName:  "_csrf",
			CookiePath:  "/",
			ErrorHandler: func(err error, c echo.Context) error {
				return factory.ErrorCsrfInvalid.New(err)
			},
		}))

	e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
		Skipper: func(c echo.Context) bool {
			if cfg.IsDevelopment() {
				return true
			}

			return false
		},
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "SAMEORIGIN",
		HSTSMaxAge:            3600,
		ContentSecurityPolicy: "default-src 'self'",
	}))

	e.Use(echojwt.WithConfig(echojwt.Config{
		Skipper:      jwthelper.JwtSkipper("tabledata/user-api"),
		ErrorHandler: jwthelper.JwtErrorHandler(),
		SigningKey:   []byte(cfg.Jwt.Secret),
		ContextKey:   "user",
	}))

	e.HTTPErrorHandler = echoutils.CustomErrorHandler

	e.Validator = validator.NewCustomValidator()
	e.Debug = cfg.Debug

	startCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	app := NewBootstrap(
		&cfg,      // config
		e,         // echo
		db,        // EnergySharesDB xorm.Engine
		sqsClient, // SQS Client
	).Invoke()

	if err := app.Start(startCtx); err != nil {
		logrus.Error(errors.Wrap(err, "fail start server"))
	}

	<-app.Done()
}

// AWS初期化及び設定。ローカル環境ではBASE_ENDPOINTが指定されていればそこに接続する
func initAws(region string, isLocal bool, endpoint string) (*session.Session, error) {
	awsCfg := &aws.Config{
		Region: aws.String(region),
	}
	if isLocal && strings.TrimSpace(endpoint) != "" {
		awsCfg.Endpoint = aws.String(endpoint)
		awsCfg.S3ForcePathStyle = aws.Bool(true)
		awsCfg.Credentials = credentials.NewStaticCredentials("localstack", "localstack", "")
	}

	return session.NewSession(awsCfg)
}

// InitDB ESJP データベース初期化
func InitDB(c config.Config) (*xorm.Engine, error) {
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
	db.SetLogLevel(xorm_log.LOG_DEBUG)
	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(60 * time.Second)

	return db, nil
}

func InitMyNumberDB(c config.Config) (*xorm.Engine, error) {

	db, err := xorm.NewEngine(c.MyNumberDatabase.Driver, c.MyNumberDatabase.Connection)

	if err != nil {
		return nil, err
	}

	db.ShowSQL(c.Debug)
	db.SetLogger(logger.NewXorm())
	db.SetLogLevel(xorm_log.LOG_DEBUG)
	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(60 * time.Second)

	err = db.Sync(new(domain.LookUpMyNumber))

	if err != nil {
		logrus.Error("Fail the MyNumberDBDataSync:" + err.Error())
		return nil, err
	}
	return db, nil
}

// サービス共用情報初期化
func initServiceUtils(c config.Config) error {
	serviceutils.SetServiceName(c.UserConfig.Service)
	serviceutils.SetServiceUrlConfig(c.ServiceConfig.Url)
	serviceutils.SetServiceLocale(c.Locale)
	serviceutils.SetServiceVersion(Version)
	serviceutils.SetMailConfig(c.ServiceConfig.Mail)
	serviceutils.SetCountryCode(c.CountryCode)
	serviceutils.SetServiceAccountConfig(c.OperatingCompany)
	serviceutils.SetNexwayConfig(c.Nexway)
	serviceutils.SetBankCodeJPConfig(c.BankCodeJP)
	serviceutils.SetJobSchedulerUrl(c.JobScheduler)
	serviceutils.SetCDN(c.AWS.CDN)
	serviceutils.SetPublicKeyId(c.AWS.PUBLIC_KEY_ID)
	serviceutils.SetPrivateKeyPath(c.AWS.PRIVATE_KEY_PATH)
	serviceutils.SetEmailServiceUrl(c.EmailService.Url)
	serviceutils.SetEnv(c.Env)

	if err := serviceutils.PreloadPrivateKey(c.Env, c.AWS.Region); err != nil {
		return err
	}

	return nil
}

// InitializePdf PDF生成用のフォントを初期化する
// (MIR-1201でUniPDFのライセンス設定を廃止し、フォント初期化だけが残った)。
func InitializePdf(path string) {
	if err := contract.SetDefaultFont(path); err != nil {
		logrus.Error(errors.Wrap(err, "failed to set default font"))
		panic(err)
	}
	pdfgen.SetDefaultFont(path)
}

func contextDBWithName(service string, contextDBName echomiddleware.ContextDBType, xormEngine *xorm.Engine,
	kafkaConfig kafka.Config, eventBusMap map[eventbus.ContextEventBusType]eventbus.IEventBus) echo.MiddlewareFunc {
	db := ctxdb.New(xormEngine, service, kafkaConfig)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			ctx := req.Context()

			session := db.NewSession(ctx).Context(ctx)
			defer session.Close()

			ctx = context.WithValue(ctx, contextDBName, session)
			for k, v := range eventBusMap {
				ctx = context.WithValue(ctx, k, v)
			}

			switch req.Method {
			case "POST", "PUT", "DELETE", "PATCH":
				ctx, ownsAfterCommit := factory.WithAfterCommit(ctx)
				c.SetRequest(req.WithContext(ctx))

				if err := session.Begin(); err != nil {
					return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
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
				if ownsAfterCommit {
					afterCommitCtx := context.WithValue(c.Request().Context(), contextDBName, xormEngine)
					if err := factory.RunAfterCommit(ctx, afterCommitCtx); err != nil {
						factory.Logger(afterCommitCtx).WithError(err).Error("failed to run after commit callbacks")
					}
				}
			default:
				c.SetRequest(req.WithContext(ctx))
				return next(c)
			}

			return nil
		}
	}
}
