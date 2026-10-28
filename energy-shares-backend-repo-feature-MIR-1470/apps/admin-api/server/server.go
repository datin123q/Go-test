package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/XSAM/otelsql"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/getsentry/sentry-go"
	sentryotel "github.com/getsentry/sentry-go/otel"
	_ "github.com/go-sql-driver/mysql"
	"github.com/hillfolk/goutils/ctxdb"
	"github.com/hillfolk/goutils/echomiddleware"
	"github.com/hillfolk/goutils/jwtutil"
	"github.com/hillfolk/goutils/kafka"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
	"xorm.io/xorm"
	"xorm.io/xorm/core"
	xormlog "xorm.io/xorm/log"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/echoutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/logger"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/pathutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/validator"

	"github.com/illuminarean/energy-share-core-api/tableset"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/contract"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/eventbus"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/s3utils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/sessionutils"
)

var (
	Version     string
	Commit      string
	once        sync.Once
	jwtSkipUrls map[string][]string
)

func RunServer(env string) {
	logrus.WithFields(logrus.Fields{
		"Environment": env,
		"Version":     Version + Commit,
		"Starting":    jsonutils.JSONTime(time.Now()).ToString(),
	}).Info("ENV_INFO")

	AssetsPath := pathutils.GetPathOfAssets()

	var cfg config.Config
	if err := cfg.Read(env); err != nil {
		logrus.Error(errors.Wrap(err, "fail read config"))
		panic(err)
	}

	// Sentry Setup
	if err := sentry.Init(sentry.ClientOptions{
		Dsn:                cfg.Observability.AdminAPI.SentryDSN,
		EnableTracing:      true,
		TracesSampleRate:   cfg.Observability.AdminAPI.SampleRate,
		AttachStacktrace:   true,
		Environment:        cfg.Env,
		Debug:              false,
		Release:            Version + Commit,
		IgnoreTransactions: []string{"/api/health-check"},
		BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
			for i := range event.Exception {
				exception := &event.Exception[i]
				// factoryエラーの場合はタイトルを書き換える
				if strings.Contains(exception.Type, "factory.Error") {
					switch hint.OriginalException.(type) {
					case *factory.Error:
						exception.Type = fmt.Sprintf("%d %s", hint.OriginalException.(*factory.Error).Code, hint.OriginalException.(*factory.Error).Message)
					}
				}
			}
			return event
		},
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

	// DB 테이블 초기화
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

	InitData(db, cfg.AWS.SES_Region, AssetsPath)

	// Event Bus 초기화
	defaultEventBus := eventbus.InitializeEventBus(cfg.EventBus.Config, 60)
	eventBusMap := map[eventbus.ContextEventBusType]eventbus.IEventBus{
		eventbus.ContextStatutoryLedgerEventBus: defaultEventBus,
	}
	// S3 버킷명 지정
	s3utils.InitializeS3Bucket(cfg.AWS.S3_PRIVATE_BUCKET, cfg.AWS.S3_PUBLIC_BUCKET)
	// 서비스 정보 초기화
	if err := initServiceUtils(cfg); err != nil {
		logrus.Error(errors.Wrap(err, "fail init service utils"))
		panic(err)
	}
	// 초기 등록된 유저가 없다면 모든 권한을 가지는 관리자 등록
	initAdminUser(db, cfg.Admin.Name, cfg.Admin.Email, cfg.Admin.Password)
	jwtutil.SetJwtSecret(cfg.Jwt.Secret)
	// GMO Client 정보 초기화
	InitGmoClient(db, cfg)
	// 다국어 언어팩 파일 초기화
	if err := i18nutils.InitI18nFiles(AssetsPath); err != nil {
		logrus.Error(err)
		panic(err)
	}

	e := echo.New()
	// デバッグ用の Context の項目、Kafka へのアクセスログ、DB・AWS を Context に入れるミドルウェア
	infra := append(echoutils.CustomDebugContext(cfg.Debug), echoutils.CustomKafkaLogger(defaultEventBus, cfg)...)
	infra = append(infra,
		contextDBWithName(cfg.AdminConfig.Service, factory.ConextDBType, db, cfg.Database.Logger.Kafka, eventBusMap),
		contextDBWithName(cfg.AdminConfig.Service, factory.ConextMyNumberDBType, myNumberDb, cfg.MyNumberDatabase.Logger.Kafka, eventBusMap),
		echomiddleware.ContextAws(CreateAwsSession(cfg.AWS.Region, cfg.IsLocal(), cfg.AWS.BASE_ENDPOINT)),
		sessionutils.ContextAwsSns(CreateAwsSession(cfg.AWS.SNS_Region, cfg.IsLocal(), cfg.AWS.BASE_ENDPOINT)),
		sessionutils.ContextAwsSes(CreateAwsSession(cfg.AWS.SES_Region, cfg.IsLocal(), cfg.AWS.BASE_ENDPOINT)),
	)
	configureEcho(e, &cfg, infra...)

	startCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	app := NewBootstrap(
		&cfg, // config
		e,    // echo
		db,   // EnergySharesDB xorm.Engine
	).Invoke()

	if err := app.Start(startCtx); err != nil {
		logrus.Error(errors.Wrap(err, "fail start server"))
	}

	<-app.Done()
}

// configureEcho は、要求の前処理・トークンの検査・エラーの扱いを組み立てる。route_roles_test.go も同じものでルートの
// 権限を確かめるので、ここを変えるとテストにも効く(MIR-1408)。infra(DB・AWS・Kafka などのミドルウェア)は
// Recover などの後・CORS とトークンの検査の前に差し込む。順番を変えると、401 の要求がアクセスログに残らなく
// なるなど、本番の挙動が変わる。
func configureEcho(e *echo.Echo, cfg *config.Config, infra ...echo.MiddlewareFunc) {
	e.Pre(middleware.RemoveTrailingSlash())
	e.Pre(echoutils.CustomContext())
	e.Use(otelecho.Middleware("miralita-admin-backend"))
	e.Use(echoutils.CustomRecover())
	e.Use(echoutils.RefreshTokenCookieExtend())
	e.Use(infra...)

	// CORS Setting
	e.Use(middleware.CORS())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))

	e.Use(echojwt.WithConfig(echojwt.Config{
		Skipper:      jwthelper.JwtSkipper("tabledata/admin-api"),
		ErrorHandler: jwthelper.JwtErrorHandler(),
		SigningKey:   []byte(cfg.Jwt.Secret),
		ContextKey:   "user",
	}))

	e.HTTPErrorHandler = echoutils.CustomErrorHandler

	e.Validator = validator.NewCustomValidator()
	e.Debug = cfg.Debug
}

func deferDatabase(energysharesDB, myNumberDb *xorm.Engine) {
	func(edb *xorm.Engine) {
		err := edb.Close()
		if err != nil {
			panic(errors.Wrap(err, "fail close energy-shares db"))
		}
	}(energysharesDB)

	func(mdb *xorm.Engine) {
		err := mdb.Close()
		if err != nil {
			panic(errors.Wrap(err, "fail close my-number db"))
		}
	}(myNumberDb)
}

func InitGmoClient(Db *xorm.Engine, cfg config.Config) {
	_, err := domain.GmoAozoraClient{}.InitClient(Db, cfg)
	if err != nil {
		logrus.Error(err)
		panic(err)
	}
}

// InitDB ESJP 데이터 베이스 초기화
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
	db.SetLogLevel(xormlog.LOG_DEBUG)
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
	db.SetLogLevel(xormlog.LOG_DEBUG)
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

// InitializePdf 契約書面PDF生成用のフォントディレクトリを初期化する
// (MIR-1201でUniPDFのライセンス設定を廃止し、フォント初期化だけが残った)。
func InitializePdf(path string) {
	if err := contract.SetDefaultFont(path); err != nil {
		logrus.Error(errors.Wrap(err, "failed to set default font"))
		panic(err)
	}
}

// CreateAwsSession AWS 초기화 및 설정
func CreateAwsSession(region string, isLocal bool, endpoint string) *session.Session {
	cfg := &aws.Config{
		Region: aws.String(region),
	}
	if isLocal && strings.TrimSpace(endpoint) != "" {
		cfg.Endpoint = aws.String(endpoint)
		cfg.S3ForcePathStyle = aws.Bool(true)
		cfg.Credentials = credentials.NewStaticCredentials("localstack", "localstack", "")
	}

	sess, err := session.NewSession(cfg)
	if err != nil {
		logrus.Error(errors.Wrap(err, "failed to create aws session :"+region))
		panic(err)
	}
	return sess
}

func InitData(db *xorm.Engine, region, basePath string) {

	if err := tableset.InitUserRolesTable(db, basePath); err != nil {
		panic(errors.Wrap(err, "fail init user roles table"))
	}
	if err := tableset.InitPolicyTable(db, basePath); err != nil {
		panic(errors.Wrap(err, "fail init policy table"))
	}
}

func initAdminUser(engine *xorm.Engine, name, email, password string) {

	// TODO 최초 관리자 등록 클린아키텍처 전환 필요
	count, err := engine.Table("admin_user").Select("count(admin_user.id)").Count(domain.AdminUser{})
	if err != nil {
		logrus.Error(errors.Wrap(err, "failed to count user"))
		panic(err)
	} else if count == 0 {
		adminUser := &domain.AdminUser{
			Id:                1,
			Email:             email,
			Name:              name,
			Password:          password,
			Roles:             nil,
			MobilePhone:       "",
			LastLoginDate:     jsonutils.JSONTime{},
			LoginFailCnt:      0,
			LastLoginIp:       "",
			LastLoginAgent:    "",
			RolesString:       "",
			IsSuperAdmin:      true,
			RoleUpdated:       jsonutils.JSONTime(time.Now()),
			ValidityStartDate: jsonutils.JSONTime(time.Now()),
			ValidityEndDate:   jsonutils.JSONTime(time.Now().AddDate(100, 0, 0)),
			Updated:           jsonutils.JSONTime{},
			Created:           jsonutils.JSONTime{},
			Deleted:           jsonutils.JSONTime{},
		}

		if err := adminUser.CreateAdmin(engine); err != nil {
			logrus.Error(errors.Wrap(err, "Fail the CreateAdminUser"))
			panic(err)
		}

	}
}

// 서비스 공용 정보 초기화
func initServiceUtils(c config.Config) error {
	serviceutils.SetDomain(c.Domain)
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
	serviceutils.SetStatutoryBooksServiceUrl(c.StatutoryBooksService)
	serviceutils.SetEmailServiceUrl(c.EmailService.Url)
	serviceutils.SetEnv(c.Env)

	if err := serviceutils.PreloadPrivateKey(c.Env, c.AWS.Region); err != nil {
		return err
	}

	return nil
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
