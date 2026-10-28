package serviceutils

import (
	"os"
	"strings"

	"github.com/pkg/errors"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/bankcodejp"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/nexway"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/operatingcompany"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/paramutils"
)

var (
	SERVICE_DOMAIN       string = "localhost"
	SERVICE_LOCALE       string = "ja"
	SERVICE_VERSION      string = "1.0.0"
	SERVICE_NAME         string = "energyshares-core-api"
	SERVICE_COUNTRY_CODE string = "+81"

	SERVICE_JOB_SCHEDULER    string = ""
	SERVICE_PUBLIC_KEY_ID    string = "public_key_id"
	SERVICE_PRIVATE_KEY_PATH string = "key/private.pem"
	SERVICE_PRIVATE_KEY      string
	SERVICE_CDN              string = ""

	STATUTORY_BOOKS_SERVICE_URL = "http://localhost:3000"
	EMAIL_SERVICE_URL           = "http://localhost:5000"

	ENV string

	// Operating Company
	SERVICE_ACCOUNT_NUMBER  string
	SERVICE_ACCOUNT_HOLDER  string
	SERVICE_BRANCH_CODE     string
	SERVICE_BRANCH_NAME     string
	SERVICE_COMPANY_ADDRESS string
	SERVICE_COMPANY_PHONE   string
	SERVICE_CORP_NUMBER     string
	SERVICE_CORP_NAME       string
	SERVICE_CEO_NAME        string

	// Nexway
	SERVICE_NEXWAY_ID             string
	SERVICE_NEXWAY_PWD            string
	SERVICE_NEXWAY_REGIST_API     string
	SERVICE_NEXWAY_FILE_API       string
	SERVICE_NEXWAY_ORDER_API      string
	SERVICE_NEXWAY_APPLICANTS_API string
	SERVICE_NEXWAY_TOKEN          string
	// 本人確認BPOサービス JPKI版(ProTechとは別契約・別トークン)
	SERVICE_NEXWAY_JPKI_BPO_BASE_URL string
	SERVICE_NEXWAY_JPKI_BPO_TOKEN    string

	// BankCodeJP
	SERVICE_BANK_CODE_JP_API_KEY string

	// Service URL
	USER_SERVICE_DOMAIN string
	PRODUCT_URL         string
	MY_PAGE_URL         string

	// Service Email
	SERVICE_EMAILS       []string
	SERVICE_INFO_MAIL    string
	SERVICE_SUPPORT_MAIL string
	SERVICE_PRIVACY_MAIL string
	SERVICE_LOAN_MAIL    string
	SERVICE_CS_MAIL      string
	SERVICE_MASTER_MAIL  string
	SERVICE_ADMIN_MAIL   string

	BLOG_API_URL string
	BLOG_API_KEY string

	ATTENDANCE_REMINDER_SLACK_WEBHOOK_URL string
)

func init() {
	SERVICE_LOCALE = "ja"
	SERVICE_VERSION = "1.0.0"
	SERVICE_ACCOUNT_NUMBER = "000000"
}

func GetDomain() string {
	return SERVICE_DOMAIN
}

func SetDomain(key string) {
	SERVICE_DOMAIN = key
}

func GetPublicKeyId() string {
	return SERVICE_PUBLIC_KEY_ID
}

func SetPublicKeyId(key string) {
	SERVICE_PUBLIC_KEY_ID = key
}

func GetPrivateKeyPath() string {
	return SERVICE_PRIVATE_KEY_PATH
}

func SetPrivateKeyPath(key string) {
	SERVICE_PRIVATE_KEY_PATH = key
}

func SetPrivateKey(key string) {
	SERVICE_PRIVATE_KEY = key
}

func GetPrivateKey() string {
	return SERVICE_PRIVATE_KEY
}

func GetCDN() string {
	return SERVICE_CDN
}

func SetCDN(url string) {
	SERVICE_CDN = url
}

func SetServiceLocale(locale string) {
	SERVICE_LOCALE = locale
}

func GetServiceLocale() string {
	return SERVICE_LOCALE
}

func SetServiceVersion(v string) {
	SERVICE_VERSION = v
}

func GetServiceVersion() string {
	return SERVICE_VERSION
}

func SetServiceName(n string) {
	SERVICE_NAME = n
}

func GetServiceName() string {
	return SERVICE_NAME
}

// func SetServiceProductUrl(serviceProductUrl string) {
//	SERVICE_PRODUCT_URL = serviceProductUrl
// }
//
// func GetServiceProductUrl() string {
//	return SERVICE_PRODUCT_URL
// }

func SetCountryCode(code string) {
	SERVICE_COUNTRY_CODE = code
}

func GetServiceCountryCode() string {
	return SERVICE_COUNTRY_CODE
}

func GetServiceAccountNo() string {
	return SERVICE_ACCOUNT_NUMBER
}

func GetServiceAccountHolder() string {
	return SERVICE_ACCOUNT_HOLDER
}

func GetServiceAccountBranchName() string {
	return SERVICE_BRANCH_NAME
}

func GetServiceAccountBranchCode() string {
	return SERVICE_BRANCH_CODE
}

func GetServiceCompanyAddress() string {
	return SERVICE_COMPANY_ADDRESS
}

func GetServiceCompanyPhone() string {
	return SERVICE_COMPANY_PHONE
}

func GetServiceCorpNumber() string {
	return SERVICE_CORP_NUMBER
}

func GetServiceCorpName() string {
	return SERVICE_CORP_NAME
}

func GetServiceCeoName() string {
	return SERVICE_CEO_NAME
}

func SetServiceAccountConfig(c operatingcompany.OperatingCompanyConfig) {
	SERVICE_ACCOUNT_NUMBER = c.AccountNumber
	SERVICE_ACCOUNT_HOLDER = c.AccountHolder
	SERVICE_BRANCH_CODE = c.BranchCode
	SERVICE_BRANCH_NAME = c.BranchName
	SERVICE_COMPANY_ADDRESS = c.Address
	SERVICE_COMPANY_PHONE = c.Phone
	SERVICE_CORP_NUMBER = c.CorpNumber
	SERVICE_CORP_NAME = c.CorpName
	SERVICE_CEO_NAME = c.CeoName
}

func SetNexwayConfig(c nexway.Config) {
	SERVICE_NEXWAY_ID = c.Id
	SERVICE_NEXWAY_PWD = c.Password
	SERVICE_NEXWAY_REGIST_API = c.FormRegistrationApi
	SERVICE_NEXWAY_FILE_API = c.FormFileRegistrationApi
	SERVICE_NEXWAY_ORDER_API = c.FormOrderApi
	SERVICE_NEXWAY_APPLICANTS_API = c.ApplicantsApi
	SERVICE_NEXWAY_TOKEN = c.Token
	SERVICE_NEXWAY_JPKI_BPO_BASE_URL = c.JpkiBpoBaseUrl
	SERVICE_NEXWAY_JPKI_BPO_TOKEN = c.JpkiBpoToken
}

func GetServiceNexwayId() string {
	return SERVICE_NEXWAY_ID
}

func GetServiceNexwayPwd() string {
	return SERVICE_NEXWAY_PWD
}

func GetServiceNexwayRegistAPI() string {
	return SERVICE_NEXWAY_REGIST_API
}

func GetServiceNexwayFileAPI() string {
	return SERVICE_NEXWAY_FILE_API
}

func GetServiceNexwayOrderAPI() string {
	return SERVICE_NEXWAY_ORDER_API
}

func GetServiceNexwayApplicantsAPI() string {
	return SERVICE_NEXWAY_APPLICANTS_API
}

func GetServiceNexwayToken() string {
	return SERVICE_NEXWAY_TOKEN
}

// GetServiceNexwayJpkiBpoBaseUrl は、本人確認BPOサービス JPKI版のベースURL(ホスト部まで)。
func GetServiceNexwayJpkiBpoBaseUrl() string {
	return SERVICE_NEXWAY_JPKI_BPO_BASE_URL
}

// GetServiceNexwayJpkiBpoToken は、本人確認BPOサービス JPKI版のAPIトークン。
func GetServiceNexwayJpkiBpoToken() string {
	return SERVICE_NEXWAY_JPKI_BPO_TOKEN
}

func SetMailConfig(c MailConfig) {
	SERVICE_INFO_MAIL = c.Info
	SERVICE_SUPPORT_MAIL = c.Support
	SERVICE_PRIVACY_MAIL = c.Privacy
	SERVICE_LOAN_MAIL = c.Loan
	SERVICE_ADMIN_MAIL = c.Admin
	SERVICE_MASTER_MAIL = c.Master
	SERVICE_CS_MAIL = c.Cs
	SERVICE_EMAILS = append(SERVICE_EMAILS, c.Info)
	SERVICE_EMAILS = append(SERVICE_EMAILS, c.Support)
	SERVICE_EMAILS = append(SERVICE_EMAILS, c.Privacy)
	SERVICE_EMAILS = append(SERVICE_EMAILS, c.Loan)
	SERVICE_EMAILS = append(SERVICE_EMAILS, c.Admin)
	SERVICE_EMAILS = append(SERVICE_EMAILS, c.Master)
	SERVICE_EMAILS = append(SERVICE_EMAILS, c.Cs)
}
func GetServiceEmails() []string {
	return SERVICE_EMAILS
}
func GetServiceInfoMail() string {
	return SERVICE_INFO_MAIL
}
func GetServiceSupportMail() string {
	return SERVICE_SUPPORT_MAIL
}
func GetServicePrivacyMail() string {
	return SERVICE_PRIVACY_MAIL
}
func GetServiceLoanMail() string {
	return SERVICE_LOAN_MAIL
}
func GetServiceCsMail() string {
	return SERVICE_CS_MAIL
}
func GetMasterEmail() string {
	return SERVICE_MASTER_MAIL
}
func GetServiceAdminMail() string {
	return SERVICE_INFO_MAIL
}

func SetServiceUrlConfig(c UrlConfig) {
	USER_SERVICE_DOMAIN = c.UserServiceDomain
	PRODUCT_URL = c.Product
	MY_PAGE_URL = c.MyPage
}

func GetUserServiceDomain() string {
	return USER_SERVICE_DOMAIN
}
func GetProductUrl() string {
	return PRODUCT_URL
}
func GetMyPageUrl() string {
	return MY_PAGE_URL
}

func SetBankCodeJPConfig(c bankcodejp.Config) {
	SERVICE_BANK_CODE_JP_API_KEY = c.ApiKey
}

func GetBankCodeJPApiKey() string {
	return SERVICE_BANK_CODE_JP_API_KEY
}

func SetJobSchedulerUrl(url string) {
	SERVICE_JOB_SCHEDULER = url
}

func GetJobSchedulerUrl() string {
	return SERVICE_JOB_SCHEDULER
}

func SetStatutoryBooksServiceUrl(url string) {
	STATUTORY_BOOKS_SERVICE_URL = url
}

func GetStatutoryBooksServiceUrl() string {
	return STATUTORY_BOOKS_SERVICE_URL
}

func SetEmailServiceUrl(url string) {
	EMAIL_SERVICE_URL = url
}

func GetEmailServiceUrl() string {
	return EMAIL_SERVICE_URL
}

func GetEnv() string {
	return ENV
}

func SetEnv(env string) {
	ENV = strings.ToUpper(env)
}

// ENV が prod であるかをチェックする関数。大文字小文字の区別はありません。
func EnvIsLocal() bool {
	return strings.ToUpper(ENV) == "LOCAL"
}
func EnvIsDev() bool {
	return strings.ToUpper(ENV) == "DEV"
}
func EnvIsQa() bool {
	return strings.ToUpper(ENV) == "QA"
}
func EnvIsStg() bool {
	return strings.ToUpper(ENV) == "STG"
}
func EnvIsProd() bool {
	return strings.ToUpper(ENV) == "PROD"
}

func LoadPrivateKey() (string, error) {
	if SERVICE_PRIVATE_KEY == "" {
		return "", errors.New("private key is not loaded")
	}
	return SERVICE_PRIVATE_KEY, nil
}

func PreloadPrivateKey(env, region string) error {
	if SERVICE_PRIVATE_KEY != "" {
		return nil
	}

	key, err := fetchPrivateKey(env, region)
	if err != nil {
		return err
	}

	SERVICE_PRIVATE_KEY = key
	return nil
}

func fetchPrivateKey(env, region string) (string, error) {
	env = strings.TrimSpace(env)
	if strings.ToUpper(env) == "LOCAL" {
		path := SERVICE_PRIVATE_KEY_PATH
		if path == "" {
			return "", errors.New("private key path is empty")
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return "", errors.Wrap(err, "failed to read private key file")
		}
		if len(data) == 0 {
			return "", errors.New("private key file is empty")
		}
		return string(data), nil
	}

	store, err := paramutils.NewStore(region)
	if err != nil {
		return "", errors.Wrap(err, "failed to create parameter store session")
	}
	param, err := store.ParamRead(env, "/app", "common-backend/cloudfront-private-key-pem")
	if err != nil {
		return "", errors.Wrap(err, "failed to read private key from parameter store")
	}
	if param.Parameter == nil || param.Parameter.Value == nil {
		return "", errors.New("private key parameter is empty")
	}
	value := *param.Parameter.Value
	if value == "" {
		return "", errors.New("private key parameter is empty")
	}
	return value, nil
}

func SetBlogApiServiceUrl(url string) {
	BLOG_API_URL = url
}

func GetBlogApiServiceUrl() string {
	return BLOG_API_URL
}

func SetBlogApiServiceKey(key string) {
	BLOG_API_KEY = key
}

func GetBlogApiServiceKey() string {
	return BLOG_API_KEY
}

func SetAttendanceReminderSlackWebhookUrl(url string) {
	ATTENDANCE_REMINDER_SLACK_WEBHOOK_URL = url
}

func GetAttendanceReminderSlackWebhookUrl() string {
	return ATTENDANCE_REMINDER_SLACK_WEBHOOK_URL
}
