package main

import (
	"context"
	"energyshares-invest-worker/controllers"
	"energyshares-invest-worker/usecase"
	"reflect"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/go-faster/errors"
	"github.com/hillfolk/goutils/ctxdb"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/awsutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/emailutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/eventbus"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/idgenerator"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/logger"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"go.uber.org/fx"
	"xorm.io/xorm"

	_ "github.com/go-sql-driver/mysql"
	projectRepository "github.com/illuminarean/energy-shares-backend-repo/repository/project/mysql"
	xorm_log "xorm.io/xorm/log"
)

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{DisableHTMLEscape: true})

	// 環境変数読み込み
	viper.AutomaticEnv()

	// アプリケーション起動
	if err := run(viper.GetString("ENV")); err != nil {
		logrus.Error(errors.Wrap(err, "fail run application"))
	}
}

// run アプリケーション起動
func run(env string) error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	logrus.WithFields(logrus.Fields{
		"Environment": env,
		"Starting":    time.Now().Format(time.DateTime),
	}).Info("ENV_INFO")

	// 設定ファイル読み込み
	var cfg config.Config
	if err := cfg.Read(env); err != nil {
		logrus.Error(errors.Wrap(err, "fail read config"))
		panic(err)
	}
	initServiceUtils(cfg)

	// DB設定
	db, err := initDB(cfg)
	if err != nil {
		logrus.Error(err)
		panic(err)
	}
	ctx = context.WithValue(ctx, factory.ConextDBType, db)
	defer db.Close()

	// AWS設定
	awsCfg, err := awsutils.InitAwsV2(ctx, cfg.AWS.Region)
	if err != nil {
		logrus.Error(err)
		panic(err)
	}

	// SQSクライアント設定
	sqsClient := awsutils.InitSQSClient(awsCfg, cfg.IsLocal(), cfg.CommonConfig.AWS.BASE_ENDPOINT)

	// Event Bus設定
	bus := eventbus.InitializeEventBus(cfg.EventBus.Config, 60)
	ctx = context.WithValue(ctx, eventbus.ContextStatutoryLedgerEventBus, bus)

	// アプリ起動（DB情報を含むcontextを渡す）
	app := NewBootstrap(&cfg, db, sqsClient).Invoke(ctx, cfg.WorkerConfig.ShutdownGraceTimeSeconds)
	app.Run()

	return nil
}

func initDB(c config.Config) (*xorm.Engine, error) {
	db, err := xorm.NewEngine(c.Database.Driver, c.Database.Connection)
	if err != nil {
		return nil, err
	}

	db.ShowSQL(c.Debug)
	db.SetLogger(logger.NewXorm())
	db.SetLogLevel(xorm_log.LOG_DEBUG)
	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(60 * time.Second)

	ctxDB := ctxdb.New(db, c.WorkerConfig.Service, c.Database.Logger.Kafka)
	return ctxDB.Engine, nil
}

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
			value, ok := injection.(T)
			if !ok {
				return func() T { return *new(T) }
			}
			return func() T { return value }
		}
	}
	return func() T { return *new(T) }
}

func (b *bootstrap) Invoke(baseCtx context.Context, shutdownGraceTimeSeconds int32) *fx.App {
	return fx.New(
		// シャットダウンタイムアウトを設定
		fx.StopTimeout(time.Duration(shutdownGraceTimeSeconds)*time.Second),

		fx.Provide(
			inject(b.injections, &config.Config{}),
			inject(b.injections, &xorm.Engine{}),
			inject(b.injections, &sqs.Client{}),

			fx.Annotate(
				func(cfg *config.Config) config.WorkerConfig {
					return cfg.WorkerConfig
				},
				fx.ResultTags(`name:"workerConfig"`),
			),

			fx.Annotate(
				func(cfg *config.Config) string {
					return cfg.AWS.INVEST_WORKER_QUEUE_URL
				},
				fx.ResultTags(`name:"queueURL"`),
			),

			fx.Annotate(
				controllers.NewSqsWorker,
				fx.ParamTags("", `name:"workerConfig"`, `name:"queueURL"`, ""),
			),

			// factory injections
			factory.NewTxManager,
			factory.NewDBFactory,
			factory.NewLoggerFactory,
			factory.NewEventBusFactory,

			// utils
			emailutils.New,
			idgenerator.New,

			// Repository Layer
			projectRepository.New,

			// Usecase Layer
			usecase.NewInvestUsecase,
		),

		fx.Invoke(
			func(lifecycle fx.Lifecycle, worker *controllers.SqsWorker) {
				lifecycle.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						logrus.Info("Start SQS worker polling...")
						go worker.StartPolling(baseCtx)
						return nil
					},
					OnStop: func(ctx context.Context) error {
						logrus.Info("Stop SQS worker polling...")

						// シャットダウン時の猶予時間を設定
						workerCtx, cancel := context.WithTimeout(baseCtx, time.Duration(shutdownGraceTimeSeconds)*time.Second)
						defer cancel()

						if err := worker.Shutdown(workerCtx); err != nil {
							logrus.WithError(err).Error("Failed to gracefully shutdown worker")
							return err
						}

						// workerCancel() // ワーカーのcontextをキャンセル
						logrus.Info("SQS worker polling stopped")
						return nil
					},
				})
			},
		),
	)
}

func initServiceUtils(c config.Config) {
	serviceutils.SetMailConfig(c.ServiceConfig.Mail)
	serviceutils.SetEmailServiceUrl(c.EmailService.Url)
	serviceutils.SetEnv(c.Env)
	serviceutils.SetServiceUrlConfig(c.ServiceConfig.Url)
	serviceutils.SetServiceName(c.UserConfig.Service)
	serviceutils.SetServiceLocale(c.Locale)
	serviceutils.SetCountryCode(c.CountryCode)
	serviceutils.SetServiceAccountConfig(c.OperatingCompany)
	serviceutils.SetJobSchedulerUrl(c.JobScheduler)
	serviceutils.SetCDN(c.AWS.CDN)
	serviceutils.SetPublicKeyId(c.AWS.PUBLIC_KEY_ID)
	serviceutils.SetPrivateKeyPath(c.AWS.PRIVATE_KEY_PATH)
	serviceutils.SetDomain(c.Domain)
	serviceutils.SetStatutoryBooksServiceUrl(c.StatutoryBooksService)
}
