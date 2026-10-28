package clean_arch_domain

import (
	"context"

	"github.com/samber/lo"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

var CompositeUserCSVHeader = []string{
	"氏名",          // Name
	"メールアドレス(ID)", // Email
	"会員承認ステータス",   // AuditStatuses
	"国籍",          // Nationalities
	"性別",          // Genders
	"年齢",          // Birthday (YYYY-MM-DD)
	"職業",          // Jobs
	"会員申請日",       // SignUpRequestDate (YYYY-MM-DD HH:MM:SS)
	"会員登録承認日",     // SignUpApprovalDate (YYYY-MM-DD HH:MM:SS)
	"メルマガ申請会員",    // AllowsNewsletter (同意/非同意)
	"最終アクセス日時",    // LastLoginDate (YYYY-MM-DD HH:MM:SS)
	"マイナンバー登録状態",  // MyNumberStatus (同意/非同意)
	"金融資産",        // Assets
	"年収",          // AnnualIncomes
	"総資産",         // TotalAsset
	"投資可能金額",      // Deposit
	"運用中金額",       // InvestOperationAmount
	"投資完了金額",      // InvestSuccessAmount
	"投資申請金額",      // InvestRequestAmount
	"入金必要額",       // UnpaidAmount
	"会員区分",        // UserTypes
}

type CompositeUser struct {
	Id                    int64                       `json:"id"`                      // user ID
	Email                 string                      `json:"email"`                   // メールアドレス
	Name                  string                      `json:"name"`                    // 名前
	AuditStatus           *domain.AUDIT_STATUS        `json:"audit_status"`            // 会員承認状態
	UserType              *domain.USER_TYPE           `json:"user_type"`               // 会員区分
	Nationality           *domain.NATIONALITY_TYPE    `json:"nationality"`             // 国籍
	Gender                *domain.GENDER              `json:"gender"`                  // 性別
	Job                   *domain.USER_JOB_TYPE       `json:"job"`                     // 職業
	Asset                 *domain.ASSETS_STATUS       `json:"asset"`                   // 資産
	AnnualIncome          *domain.ANNUALINCOME_STATUS `json:"annual_income"`           // 年収
	MyNumberStatus        *domain.MY_NUMBER_STATUS    `json:"my_number_status"`        // マイナンバー登録状態
	AllowsNewsletter      *bool                       `json:"allows_newsletter"`       // ニュースレター受信同意
	Birthday              *jsonutils.JSONTime         `json:"birthday"`                // 生年月日
	SignUpRequestDate     *jsonutils.JSONTime         `json:"sign_up_request_date"`    // 会員申請日
	SignUpApprovalDate    *jsonutils.JSONTime         `json:"sign_up_approval_date"`   // 会員登録承認日
	LastLoginDate         *jsonutils.JSONTime         `json:"last_login_date"`         // 最終ログイン日時
	TotalAsset            *int64                      `json:"total_asset"`             // 総資産
	InvestOperationAmount *int64                      `json:"invest_operation_amount"` // 運用中の金額
	Deposit               *int64                      `json:"deposit"`                 // 投資可能金額
	InvestSuccessAmount   *int64                      `json:"invest_success_amount"`   // 投資完了金額
	InvestRequestAmount   *int64                      `json:"invest_request_amount"`   // 投資申請金額
	UnpaidAmount          *int64                      `json:"unpaid_amount"`           // 入金必要額
}

type CompositeUserRepository interface {
	GetAll(ctx context.Context, options ...CompositeUserQueryOption) (int64, []CompositeUser, error)
}

type CompositeUserQueryOption = func(*CompositeUserQueryOptionFields)

func CompositeUserQueryOptionFunc() CompositeUserQueryOptionFields {
	return CompositeUserQueryOptionFields{}
}

type CompositeUserQueryOptionFields struct {
	// 文字列オプション
	emails        []string
	auditStatuses []string
	noAuditStatus *string
	userTypes     []string
	nationalities []string
	genders       []string
	jobs          []string
	assets        []string
	annualIncomes []string

	// true, false
	allowsNewsletter *bool
	hasMyNumber      *bool

	// 日付範囲
	minSignUpRequestDate  *string
	maxSignUpRequestDate  *string
	minSignUpApprovalDate *string
	maxSignUpApprovalDate *string
	minLastLoginDate      *string
	maxLastLoginDate      *string
	minBirthday           *string
	maxBirthday           *string

	// 数値範囲
	minTotalAsset            *int64
	maxTotalAsset            *int64
	minInvestOperationAmount *int64
	maxInvestOperationAmount *int64
	minDeposit               *int64
	maxDeposit               *int64
	minInvestSuccessAmount   *int64
	maxInvestSuccessAmount   *int64
	minInvestRequestAmount   *int64
	maxInvestRequestAmount   *int64
	minUnpaidAmount          *int64
	maxUnpaidAmount          *int64
}

func (qo CompositeUserQueryOptionFields) WithEmails(emails ...string) func(*CompositeUserQueryOptionFields) {
	return func(qo *CompositeUserQueryOptionFields) {
		qo.emails = emails
	}
}

func (qo CompositeUserQueryOptionFields) WithAuditStatuses(auditStatuses ...domain.AUDIT_STATUS) func(*CompositeUserQueryOptionFields) {
	return func(qo *CompositeUserQueryOptionFields) {
		for _, auditStatus := range auditStatuses {
			qo.auditStatuses = append(qo.auditStatuses, string(auditStatus))
		}
	}
}

func (qo CompositeUserQueryOptionFields) WithNoAuditStatus(noAuditStatus domain.AUDIT_STATUS) func(*CompositeUserQueryOptionFields) {
	return func(qo *CompositeUserQueryOptionFields) {
		qo.noAuditStatus = lo.ToPtr(string(noAuditStatus))
	}
}

func (qo CompositeUserQueryOptionFields) WithUserTypes(userTypes ...domain.USER_TYPE) func(*CompositeUserQueryOptionFields) {
	return func(qo *CompositeUserQueryOptionFields) {
		for _, userType := range userTypes {
			qo.userTypes = append(qo.userTypes, string(userType))
		}
	}
}

func (qo CompositeUserQueryOptionFields) WithNationalities(nationalities ...domain.NATIONALITY_TYPE) func(*CompositeUserQueryOptionFields) {
	return func(qo *CompositeUserQueryOptionFields) {
		for _, nationality := range nationalities {
			qo.nationalities = append(qo.nationalities, string(nationality))
		}
	}
}

func (qo CompositeUserQueryOptionFields) WithGenders(genders ...domain.GENDER) func(*CompositeUserQueryOptionFields) {
	return func(qo *CompositeUserQueryOptionFields) {
		for _, gender := range genders {
			qo.genders = append(qo.genders, string(gender))
		}
	}
}

func (qo CompositeUserQueryOptionFields) WithJobs(jobs ...domain.USER_JOB_TYPE) func(*CompositeUserQueryOptionFields) {
	return func(qo *CompositeUserQueryOptionFields) {
		for _, job := range jobs {
			qo.jobs = append(qo.jobs, string(job))
		}
	}
}

func (qo CompositeUserQueryOptionFields) WithAssets(assets ...domain.ASSETS_STATUS) func(*CompositeUserQueryOptionFields) {
	return func(qo *CompositeUserQueryOptionFields) {
		for _, asset := range assets {
			qo.assets = append(qo.assets, string(asset))
		}
	}
}

func (qo CompositeUserQueryOptionFields) WithAnnualIncomes(annualIncomes ...domain.ANNUALINCOME_STATUS) func(*CompositeUserQueryOptionFields) {
	return func(qo *CompositeUserQueryOptionFields) {
		for _, annualIncome := range annualIncomes {
			qo.annualIncomes = append(qo.annualIncomes, string(annualIncome))
		}
	}
}

func (qo CompositeUserQueryOptionFields) WithAllowsNewsletter(allowsNewsletter *bool) func(*CompositeUserQueryOptionFields) {
	return func(qo *CompositeUserQueryOptionFields) {
		qo.allowsNewsletter = allowsNewsletter
	}
}

func (qo CompositeUserQueryOptionFields) WithHasMyNumber(hasMyNumber *bool) func(*CompositeUserQueryOptionFields) {
	return func(qo *CompositeUserQueryOptionFields) {
		qo.hasMyNumber = hasMyNumber
	}
}

func (qo CompositeUserQueryOptionFields) WithMinSignUpRequestDate(minSignUpRequestDate *string) func(*CompositeUserQueryOptionFields) {
	return func(qo *CompositeUserQueryOptionFields) {
		qo.minSignUpRequestDate = minSignUpRequestDate
	}
}

func (qo CompositeUserQueryOptionFields) WithMaxSignUpRequestDate(maxSignUpRequestDate *string) func(*CompositeUserQueryOptionFields) {
	return func(qo *CompositeUserQueryOptionFields) {
		qo.maxSignUpRequestDate = maxSignUpRequestDate
	}
}

func (qo CompositeUserQueryOptionFields) WithMinSignUpApprovalDate(minSignUpApprovalDate *string) func(*CompositeUserQueryOptionFields) {
	return func(qo *CompositeUserQueryOptionFields) {
		qo.minSignUpApprovalDate = minSignUpApprovalDate
	}
}

func (qo CompositeUserQueryOptionFields) WithMaxSignUpApprovalDate(maxSignUpApprovalDate *string) func(*CompositeUserQueryOptionFields) {
	return func(qo *CompositeUserQueryOptionFields) {
		qo.maxSignUpApprovalDate = maxSignUpApprovalDate
	}
}

func (qo CompositeUserQueryOptionFields) WithMinLastLoginDate(minLastLoginDate *string) func(*CompositeUserQueryOptionFields) {
	return func(qo *CompositeUserQueryOptionFields) {
		qo.minLastLoginDate = minLastLoginDate
	}
}

func (qo CompositeUserQueryOptionFields) WithMaxLastLoginDate(maxLastLoginDate *string) func(*CompositeUserQueryOptionFields) {
	return func(qo *CompositeUserQueryOptionFields) {
		qo.maxLastLoginDate = maxLastLoginDate
	}
}

func (qo CompositeUserQueryOptionFields) WithMinBirthday(minBirthday *string) func(*CompositeUserQueryOptionFields) {
	return func(qo *CompositeUserQueryOptionFields) {
		qo.minBirthday = minBirthday
	}
}

func (qo CompositeUserQueryOptionFields) WithMaxBirthday(maxBirthday *string) func(*CompositeUserQueryOptionFields) {
	return func(qo *CompositeUserQueryOptionFields) {
		qo.maxBirthday = maxBirthday
	}
}

func (qo CompositeUserQueryOptionFields) WithMinTotalAsset(minTotalAsset *int64) func(*CompositeUserQueryOptionFields) {
	return func(qo *CompositeUserQueryOptionFields) {
		qo.minTotalAsset = minTotalAsset
	}
}

func (qo CompositeUserQueryOptionFields) WithMaxTotalAsset(maxTotalAsset *int64) func(*CompositeUserQueryOptionFields) {
	return func(qo *CompositeUserQueryOptionFields) {
		qo.maxTotalAsset = maxTotalAsset
	}
}

func (qo CompositeUserQueryOptionFields) WithMinInvestOperationAmount(minInvestOperationAmount *int64) func(*CompositeUserQueryOptionFields) {
	return func(qo *CompositeUserQueryOptionFields) {
		qo.minInvestOperationAmount = minInvestOperationAmount
	}
}

func (qo CompositeUserQueryOptionFields) WithMaxInvestOperationAmount(maxInvestOperationAmount *int64) func(*CompositeUserQueryOptionFields) {
	return func(qo *CompositeUserQueryOptionFields) {
		qo.maxInvestOperationAmount = maxInvestOperationAmount
	}
}

func (qo CompositeUserQueryOptionFields) WithMinDeposit(minDeposit *int64) func(*CompositeUserQueryOptionFields) {
	return func(qo *CompositeUserQueryOptionFields) {
		qo.minDeposit = minDeposit
	}
}

func (qo CompositeUserQueryOptionFields) WithMaxDeposit(maxDeposit *int64) func(*CompositeUserQueryOptionFields) {
	return func(qo *CompositeUserQueryOptionFields) {
		qo.maxDeposit = maxDeposit
	}
}

func (qo CompositeUserQueryOptionFields) WithMinInvestSuccessAmount(minInvestSuccessAmount *int64) func(*CompositeUserQueryOptionFields) {
	return func(qo *CompositeUserQueryOptionFields) {
		qo.minInvestSuccessAmount = minInvestSuccessAmount
	}
}

func (qo CompositeUserQueryOptionFields) WithMaxInvestSuccessAmount(maxInvestSuccessAmount *int64) func(*CompositeUserQueryOptionFields) {
	return func(qo *CompositeUserQueryOptionFields) {
		qo.maxInvestSuccessAmount = maxInvestSuccessAmount
	}
}

func (qo CompositeUserQueryOptionFields) WithMinInvestRequestAmount(minInvestRequestAmount *int64) func(*CompositeUserQueryOptionFields) {
	return func(qo *CompositeUserQueryOptionFields) {
		qo.minInvestRequestAmount = minInvestRequestAmount
	}
}

func (qo CompositeUserQueryOptionFields) WithMaxInvestRequestAmount(maxInvestRequestAmount *int64) func(*CompositeUserQueryOptionFields) {
	return func(qo *CompositeUserQueryOptionFields) {
		qo.maxInvestRequestAmount = maxInvestRequestAmount
	}
}

func (qo CompositeUserQueryOptionFields) WithMinUnpaidAmount(minUnpaidAmount *int64) func(*CompositeUserQueryOptionFields) {
	return func(qo *CompositeUserQueryOptionFields) {
		qo.minUnpaidAmount = minUnpaidAmount
	}
}

func (qo CompositeUserQueryOptionFields) WithMaxUnpaidAmount(maxUnpaidAmount *int64) func(*CompositeUserQueryOptionFields) {
	return func(qo *CompositeUserQueryOptionFields) {
		qo.maxUnpaidAmount = maxUnpaidAmount
	}
}

func (qo CompositeUserQueryOptionFields) Emails() []interface{} {
	interfaceSlice := make([]interface{}, len(qo.emails))
	for i, v := range qo.emails {
		interfaceSlice[i] = v
	}
	return interfaceSlice
}

func (qo CompositeUserQueryOptionFields) AuditStatuses() []interface{} {
	interfaceSlice := make([]interface{}, len(qo.auditStatuses))
	for i, v := range qo.auditStatuses {
		interfaceSlice[i] = v
	}
	return interfaceSlice
}
func (qo CompositeUserQueryOptionFields) NoAuditStatus() *string {
	return qo.noAuditStatus
}

func (qo CompositeUserQueryOptionFields) UserTypes() []interface{} {
	interfaceSlice := make([]interface{}, len(qo.userTypes))
	for i, v := range qo.userTypes {
		interfaceSlice[i] = v
	}
	return interfaceSlice
}

func (qo CompositeUserQueryOptionFields) Nationalities() []interface{} {
	interfaceSlice := make([]interface{}, len(qo.nationalities))
	for i, v := range qo.nationalities {
		interfaceSlice[i] = v
	}
	return interfaceSlice
}
func (qo CompositeUserQueryOptionFields) Genders() []interface{} {
	interfaceSlice := make([]interface{}, len(qo.genders))
	for i, v := range qo.genders {
		interfaceSlice[i] = v
	}
	return interfaceSlice
}

func (qo CompositeUserQueryOptionFields) Jobs() []interface{} {
	interfaceSlice := make([]interface{}, len(qo.jobs))
	for i, v := range qo.jobs {
		interfaceSlice[i] = v
	}
	return interfaceSlice
}
func (qo CompositeUserQueryOptionFields) Assets() []interface{} {
	interfaceSlice := make([]interface{}, len(qo.assets))
	for i, v := range qo.assets {
		interfaceSlice[i] = v
	}
	return interfaceSlice
}

func (qo CompositeUserQueryOptionFields) AnnualIncomes() []interface{} {
	interfaceSlice := make([]interface{}, len(qo.annualIncomes))
	for i, v := range qo.annualIncomes {
		interfaceSlice[i] = v
	}
	return interfaceSlice
}

func (qo CompositeUserQueryOptionFields) AllowsNewsletter() *bool {
	return qo.allowsNewsletter
}

func (qo CompositeUserQueryOptionFields) HasMyNumber() *bool {
	return qo.hasMyNumber
}

func (qo CompositeUserQueryOptionFields) MinSignUpRequestDate() *string {
	return qo.minSignUpRequestDate
}

func (qo CompositeUserQueryOptionFields) MaxSignUpRequestDate() *string {
	return qo.maxSignUpRequestDate
}

func (qo CompositeUserQueryOptionFields) MinSignUpApprovalDate() *string {
	return qo.minSignUpApprovalDate
}

func (qo CompositeUserQueryOptionFields) MaxSignUpApprovalDate() *string {
	return qo.maxSignUpApprovalDate
}

func (qo CompositeUserQueryOptionFields) MinLastLoginDate() *string {
	return qo.minLastLoginDate
}

func (qo CompositeUserQueryOptionFields) MaxLastLoginDate() *string {
	return qo.maxLastLoginDate
}

func (qo CompositeUserQueryOptionFields) MinBirthday() *string {
	return qo.minBirthday
}

func (qo CompositeUserQueryOptionFields) MaxBirthday() *string {
	return qo.maxBirthday
}

func (qo CompositeUserQueryOptionFields) MinTotalAsset() *int64 {
	return qo.minTotalAsset
}

func (qo CompositeUserQueryOptionFields) MaxTotalAsset() *int64 {
	return qo.maxTotalAsset
}

func (qo CompositeUserQueryOptionFields) MinInvestOperationAmount() *int64 {
	return qo.minInvestOperationAmount
}

func (qo CompositeUserQueryOptionFields) MaxInvestOperationAmount() *int64 {
	return qo.maxInvestOperationAmount
}

func (qo CompositeUserQueryOptionFields) MinDeposit() *int64 {
	return qo.minDeposit
}

func (qo CompositeUserQueryOptionFields) MaxDeposit() *int64 {
	return qo.maxDeposit
}

func (qo CompositeUserQueryOptionFields) MinInvestSuccessAmount() *int64 {
	return qo.minInvestSuccessAmount
}

func (qo CompositeUserQueryOptionFields) MaxInvestSuccessAmount() *int64 {
	return qo.maxInvestSuccessAmount
}

func (qo CompositeUserQueryOptionFields) MinInvestRequestAmount() *int64 {
	return qo.minInvestRequestAmount
}

func (qo CompositeUserQueryOptionFields) MaxInvestRequestAmount() *int64 {
	return qo.maxInvestRequestAmount
}

func (qo CompositeUserQueryOptionFields) MinUnpaidAmount() *int64 {
	return qo.minUnpaidAmount
}

func (qo CompositeUserQueryOptionFields) MaxUnpaidAmount() *int64 {
	return qo.maxUnpaidAmount
}
