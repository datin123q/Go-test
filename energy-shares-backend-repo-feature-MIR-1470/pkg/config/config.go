package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hillfolk/goutils/echomiddleware"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/bankcodejp"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/eventbus"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/nexway"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/operatingcompany"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/paramutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
)

type CommonConfig struct {
	Env      string
	Nexway   nexway.Config
	Database struct {
		Driver     string
		Connection string
		Logger     struct {
			Kafka echomiddleware.KafkaConfig
		}
	}
	MyNumberDatabase struct {
		Driver     string
		Connection string
		Logger     struct {
			Kafka echomiddleware.KafkaConfig
		}
	}
	EventBus struct {
		Config eventbus.Config `yaml:"config" mapstructure:"config"`
	}
	EventStream struct {
		KsqlDB KsqlDBConfig
	}
	AWS struct {
		Region                  string
		SES_Region              string
		SNS_Region              string
		CDN                     string
		S3_PRIVATE_BUCKET       string
		S3_PUBLIC_BUCKET        string
		S3_EKYC_IMAGES_BUCKET   string // eKYC画像専用バケット
		SES_ARN                 string
		SNS_ARN                 string
		PUBLIC_KEY_ID           string
		PRIVATE_KEY_PATH        string
		BASE_ENDPOINT           string // ローカル環境用
		INVEST_WORKER_QUEUE_URL string // 先着申込排他制御用キューURL
	}
	Observability struct {
		UserAPI struct {
			SentryDSN  string
			SampleRate float64
		}
		AdminAPI struct {
			SentryDSN  string
			SampleRate float64
		}
		JobScheduler struct {
			SentryDSN  string
			SampleRate float64
		}
	}
	Admin struct {
		Name     string
		Email    string
		Password string
	}
	ServiceConfig struct {
		Mail serviceutils.MailConfig
		Url  serviceutils.UrlConfig
	}
	Jwt struct {
		Secret string
	}
	Locale                string
	CountryCode           string
	Domain                string
	JobScheduler          string
	StatutoryBooksService string
	AccessTradeService    AccessTradeConfig
	EmailService          EmailServiceConfig
	BankCodeJP            bankcodejp.Config
	MailTemplateMap       map[string]string
	OperatingCompany      operatingcompany.OperatingCompanyConfig
	Debug                 bool
	Swagger               SwaggerConfig

	GmoAozoraConfig struct {
		ClientId     string
		ClientSecret string
		RedirectUri  string
	}
}

type AccessTradeConfig struct {
	Url  string
	Stub bool
}

type EmailServiceConfig struct {
	Url string
}

type AdminConfig struct {
	Service  string
	HttpPort string
}

type UserConfig struct {
	Service  string
	HttpPort string
}

type JobSchedulerConfig struct {
	Service      string
	HttpPort     string
	CronHttpPort string
	BankAccounts struct {
		LoanDeposit      Account
		LoanWithdrawal   Account
		EquityDeposit    Account
		EquityWithdrawal Account
	}
	BlogApiService struct {
		Url string
		Key string
	}
	Slack struct {
		AttendanceReminderWebhookUrl string
	}
	Cron map[string]string `yaml:"cron" mapstructure:"cron"`
}

type WorkerConfig struct {
	Service                       string // サービス名
	MaxNumberOfSqsMessages        int32  // 一度に受信するSQSのメッセージ数
	WaitTimeSecondsForSqsMessages int32  // SQSメッセージ到着待機時間(秒)
	MaxRetryCount                 int32  // 最大リトライ回数
	ShutdownGraceTimeSeconds      int32  // シャットダウン時の猶予時間(秒)
}

type Account struct {
	AccountNumber string
	BankCode      string
	BranchCode    string
	AccountHolder string
}

type Config struct {
	CommonConfig
	AdminConfig        AdminConfig        `yaml:"adminConfig"`
	UserConfig         UserConfig         `yaml:"userConfig"`
	JobSchedulerConfig JobSchedulerConfig `yaml:"jobSchedulerConfig"`
	WorkerConfig       WorkerConfig       `yaml:"workerConfig"`
}

type KsqlDBConfig struct {
	Url       string
	Resources KsqlDBResources
}

type KsqlDBResources struct {
	Changelog struct {
		Answer string `yaml:"answer"`
	}
	AccessLog string `yaml:"accessLog"`
}

type SwaggerConfig struct {
	Version         string `yaml:"version"`
	OpenApiSpecPath string `yaml:"openApiSpecPath"`
}

func (c *Config) Read(env string) error {
	// アプリケーション実行時に-eオプションがない場合、パラメータストアの値を取得する。
	// ローカルでテストする場合は、config.init.ymlファイルをadmin-apiディレクトリにコピーして、go run main.go runを実行する。
	if env == "" {
		if err := c.loadConfigFromParamStore(); err != nil {
			return err
		}
		return nil
	}
	// アプリケーション実行時に-eオプションがある場合、ローカル実行と判定してローカルファイルを読む。
	if err := c.loadConfigFromLocalFile(env); err != nil {
		return err
	}
	return nil
}

const ESJP_PREFIX = "/mira/backend"

// loadConfigFromParamStoreメソッドはパラメータストアから設定を取得して構造体にアンマーシャルするメソッドです。
// -eオプションがない場合に実行されます。
func (c *Config) loadConfigFromParamStore() error {

	configPath := "."
	executablePath, err := filepath.Abs(filepath.Dir("."))
	if err != nil {
		return fmt.Errorf("error finding the absolute path: %v", err)
	}

	viper.AddConfigPath(filepath.Join(executablePath, configPath))
	viper.SetConfigName("config.base") // 基本設定ファイル名を設定します。 ('config.base.yml')

	// 初期設定ファイルをロードします。この情報をもとにParameter Storeから設定を取得します。
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("fatal error reading  config file: %s \n", err)
	}
	if err := viper.Unmarshal(c); err != nil {
		return fmt.Errorf("fatal error unmarshaling config file: %s \n", err)
	}

	// Parameter Storeから設定を取得します。
	store, err := paramutils.NewStore(c.AWS.Region)
	if err != nil {
		return err
	}
	param, err := store.ParamRead(c.Env, ESJP_PREFIX, "")
	if err != nil {
		return err
	}

	// logging for debug
	logrus.WithFields(logrus.Fields{
		"Name":         param.Parameter.Name,
		"Version":      param.Parameter.Version,
		"Type":         param.Parameter.Type,
		"LastModified": param.Parameter.LastModifiedDate.Format(time.DateTime),
	}).Info("AWS_PARAM_STORE_CONFIG_LOADED")

	// viperを使用してParameter Storeから取得した設定を構造体にアンマーシャルします。
	viper.SetConfigType("yaml") // 設定ファイルタイプを指定
	if err := viper.ReadConfig(bytes.NewReader([]byte(*param.Parameter.Value))); err != nil {
		return err
	}

	if err := viper.Unmarshal(c); err != nil {
		return fmt.Errorf("fatal error unmarshaling config file: %s \n", err)
	}

	return nil
}

// loadConfigFromLocalFileメソッドはローカルファイルから設定を取得して構造体にアンマーシャルするメソッドです。
// -eオプションがある場合に実行されます。
func (c *Config) loadConfigFromLocalFile(env string) error {

	configPath := "../../config"
	executablePath, err := filepath.Abs(filepath.Dir("."))
	if err != nil {
		return fmt.Errorf("error finding the absolute path: %v", err)
	}

	viper.AddConfigPath(filepath.Join(executablePath, configPath))
	viper.SetConfigName(fmt.Sprintf("config.%s", env)) // 例: 'dev'環境の場合'config.dev'

	// ローカルの環境別設定ファイルをロードします。
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("fatal error config file: %s \n", err)
	}

	// 設定を構造体にアンマーシャルします。
	if err := viper.Unmarshal(c); err != nil {
		return fmt.Errorf("fatal error unmarshaling config file: %s \n", err)
	}

	return nil
}

func (c *Config) IsLocal() bool {
	return strings.ToUpper(c.Env) == "LOCAL"
}

// 本番以外(envが"LOCAL", "DEV", "QA", "STG")の場合にtrueを返す
func (c *Config) IsDevelopment() bool {
	return strings.ToUpper(c.Env) == "LOCAL" ||
		strings.ToUpper(c.Env) == "DEV" ||
		strings.ToUpper(c.Env) == "QA" ||
		c.IsStage()
}

func (c *Config) IsStage() bool {
	return strings.ToUpper(c.Env) == "STG"
}

func (c *Config) IsProd() bool {
	return strings.ToUpper(c.Env) == "PROD"
}

type Code struct {
	codeMap    map[string]interface{}
	jsonOutput string
}

func NewCode(configPath string) *Code {
	if configPath == "" {
		configPath = "../../config"
	}

	code := &Code{
		codeMap: make(map[string]interface{}),
	}

	if err := code.loadCodes(configPath); err != nil {
		panic(err)
	}

	if err := code.generateJSON(); err != nil {
		panic(err)
	}

	return code
}

func (c *Code) loadCodes(configPath string) error {
	files, err := filepath.Glob(filepath.Join(configPath, "code/*/code.yaml"))
	if err != nil {
		return errors.Wrap(err, "failed to find config files")
	}

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return errors.Wrap(err, "failed to read config file")
		}

		var configMap map[string]interface{}
		if err := yaml.Unmarshal(data, &configMap); err != nil {
			return errors.Wrapf(err, "failed to parse yaml file: %s", file)
		}

		// ドメイン名自体をマップのキーとして使用せず、configMapそのものを保存
		domain := strings.Split(filepath.Dir(file), string(filepath.Separator))
		domainName := domain[len(domain)-1]

		// トップレベルキーはそのまま保持
		c.codeMap[domainName] = configMap
	}

	return nil
}

func (c *Code) generateJSON() error {
	// JSON マーシャリング前にトップレベルマップを作成
	result := make(map[string]interface{})
	for _, value := range c.codeMap {
		if m, ok := value.(map[string]interface{}); ok {
			for k, v := range m {
				result[k] = v
			}
		}
	}

	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return errors.Wrap(err, "failed to marshal to JSON")
	}

	c.jsonOutput = string(jsonBytes)
	return nil
}

func (c *Code) JsonOutput() string {
	return c.jsonOutput
}
