package controllers

import (
	"fmt"
	"strings"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

type MailSendCsvRequest struct {
	Subject string `json:"subject" form:"subject" swagger:"desc(メールタイトル)" validate:"required"`
	Content string `json:"content" form:"content" swagger:"desc(メールコンテンツ)"`
	Sender  string `json:"sender" form:"sender" swagger:"desc(送信者メール)" validate:"required"`
}

type DateRange struct {
	StartDate jsonutils.JSONTime `json:"start_date" swagger:"desc(開始日)"`
	EndDate   jsonutils.JSONTime `json:"end_date" swagger:"desc(終了日)"`
}

type NumberRange struct {
	Min *int64 `json:"min" swagger:"desc(最小値)"`
	Max *int64 `json:"max" swagger:"desc(最大値)"`
}

type MailSendOptionRequest struct {
	// 基本情報
	Subject  string   `json:"subject" swagger:"desc(メールタイトル)" validate:"required"`
	Content  string   `json:"content" swagger:"desc(メールコンテンツ)"`
	Sender   string   `json:"sender" swagger:"desc(送信者メール)" validate:"required"`
	Receiver []string `json:"emails" swagger:"desc(受信者メール)"`

	// 文字列オプション
	AuditStatus  []string `json:"audit_statuses" swagger:"desc(審査状態: AUDIT_STATUS)"`
	UserType     []string `json:"user_types" swagger:"desc(会員区分: USER_TYPE)"`
	Nationality  []string `json:"nationalities" swagger:"desc(国: NATIONALITY_TYPE)"`
	Gender       []string `json:"genders" swagger:"desc(性別: GENDER)"`
	Job          []string `json:"jobs" swagger:"desc(職業: USER_JOB_TYPE)"`
	Asset        []string `json:"assets" swagger:"desc(資産: ASSETS_STATUS)"`
	AnnualIncome []string `json:"annual_incomes" swagger:"desc(年収: ANNUALINCOME_STATUS)"`

	// true, false
	AllowsNewsletter *bool `json:"allows_newsletter" swagger:"desc(ニュースレター受信同意 [true|false])"`
	HasMyNumber      *bool `json:"has_my_number" swagger:"desc(マイナンバー提出有無 [true|false])"`

	// date ~ date
	SignUpRequestDate  *DateRange `json:"sign_up_request_date" swagger:"desc(加入要求日)"`
	SignUpApprovalDate *DateRange `json:"sign_up_approval_date" swagger:"desc(加入承認日)"`
	LastLoginDate      *DateRange `json:"last_login_date" swagger:"desc(最終ログイン)"`

	// number ~ number
	Age                   *NumberRange `json:"age" swagger:"desc(年代)"`
	TotalAsset            *NumberRange `json:"total_asset" swagger:"desc(総資産)"`
	InvestOperationAmount *NumberRange `json:"invest_operation_amount" swagger:"desc(運用中の金額)"`
	Deposit               *NumberRange `json:"deposit" swagger:"desc(預託金、投資可能金額)"`
	InvestSuccessAmount   *NumberRange `json:"invest_success_amount" swagger:"desc(投資完了金額)"`
	InvestRequestAmount   *NumberRange `json:"invest_request_amount" swagger:"desc(投資申請金額)"`
	UnpaidAmount          *NumberRange `json:"unpaid_amount" swagger:"desc(入金必要金額)"`
}

func (m *MailSendOptionRequest) ToDomain() *domain.MailSendOption {

	defaultStartDate := "1800-01-01 00:00:00"
	defaultEndDate := "9999-12-31 23:59:59"
	defaultMin := int64(0)
	defaultMax := int64(2140000000)

	var AuditStatus string
	var UserType string
	var Nationality string
	var Gender string
	var Job string
	var Asset string
	var AnnualIncome string
	var SignUpRequestDate []string
	var SignUpApprovalDate []string
	var LastLoginDate []string
	var Age []int64
	var TotalAsset []int64
	var InvestOperationAmount []int64
	var Deposit []int64
	var InvestSuccessAmount []int64
	var InvestRequestAmount []int64
	var UnpaidAmount []int64

	if len(m.AuditStatus) > 0 {
		AuditStatus = generateInClause(m.AuditStatus)
	}
	if len(m.UserType) > 0 {
		UserType = generateInClause(m.UserType)
	}
	if len(m.Nationality) > 0 {
		Nationality = generateInClause(m.Nationality)
	}
	if len(m.Gender) > 0 {
		Gender = generateInClause(m.Gender)
	}
	if len(m.Job) > 0 {
		Job = generateInClause(m.Job)
	}
	if len(m.Asset) > 0 {
		Asset = generateInClause(m.Asset)
	}
	if len(m.AnnualIncome) > 0 {
		AnnualIncome = generateInClause(m.AnnualIncome)
	}
	if m.SignUpRequestDate != nil && (!m.SignUpRequestDate.StartDate.IsEmpty() || !m.SignUpRequestDate.EndDate.IsEmpty()) {
		if m.SignUpRequestDate.StartDate.IsEmpty() { // start date
			SignUpRequestDate = append(SignUpRequestDate, defaultStartDate)
		} else {
			SignUpRequestDate = append(SignUpRequestDate, m.SignUpRequestDate.StartDate.ToString())
		}
		if m.SignUpRequestDate.EndDate.IsEmpty() { // end date
			SignUpRequestDate = append(SignUpRequestDate, defaultEndDate)
		} else {
			SignUpRequestDate = append(SignUpRequestDate, m.SignUpRequestDate.EndDate.ToString())
		}
	}
	if m.SignUpApprovalDate != nil && (!m.SignUpApprovalDate.StartDate.IsEmpty() || !m.SignUpApprovalDate.EndDate.IsEmpty()) {
		if m.SignUpApprovalDate.StartDate.IsEmpty() { // start date
			SignUpApprovalDate = append(SignUpApprovalDate, defaultStartDate)
		} else {
			SignUpApprovalDate = append(SignUpApprovalDate, m.SignUpApprovalDate.StartDate.ToString())
		}
		if m.SignUpApprovalDate.EndDate.IsEmpty() { // end date
			SignUpApprovalDate = append(SignUpApprovalDate, defaultEndDate)
		} else {
			SignUpApprovalDate = append(SignUpApprovalDate, m.SignUpApprovalDate.EndDate.ToString())
		}
	}
	if m.LastLoginDate != nil && (!m.LastLoginDate.StartDate.IsEmpty() || !m.LastLoginDate.EndDate.IsEmpty()) {
		if m.LastLoginDate.StartDate.IsEmpty() { // start date
			LastLoginDate = append(LastLoginDate, defaultStartDate)
		} else {
			LastLoginDate = append(LastLoginDate, m.LastLoginDate.StartDate.ToString())
		}
		if m.LastLoginDate.EndDate.IsEmpty() { // end date
			LastLoginDate = append(LastLoginDate, defaultEndDate)
		} else {
			LastLoginDate = append(LastLoginDate, m.LastLoginDate.EndDate.ToString())
		}
	}
	if m.Age != nil && (m.Age.Min != nil || m.Age.Max != nil) {
		if m.Age.Min == nil { // min
			Age = append(Age, defaultMin)
		} else {
			Age = append(Age, *m.Age.Min)
		}
		if m.Age.Max == nil { // max
			Age = append(Age, int64(100))
		} else {
			Age = append(Age, *m.Age.Max)
		}
	}
	if m.TotalAsset != nil && (m.TotalAsset.Min != nil || m.TotalAsset.Max != nil) {
		if m.TotalAsset.Min == nil { // min
			TotalAsset = append(TotalAsset, defaultMin)
		} else {
			TotalAsset = append(TotalAsset, *m.TotalAsset.Min)
		}
		if m.TotalAsset.Max == nil { // max
			TotalAsset = append(TotalAsset, defaultMax)
		} else {
			TotalAsset = append(TotalAsset, *m.TotalAsset.Max)
		}
	}
	if m.InvestOperationAmount != nil && (m.InvestOperationAmount.Min != nil || m.InvestOperationAmount.Max != nil) {
		if m.InvestOperationAmount.Min == nil { // min
			InvestOperationAmount = append(InvestOperationAmount, defaultMin)
		} else {
			InvestOperationAmount = append(InvestOperationAmount, *m.InvestOperationAmount.Min)
		}
		if m.InvestOperationAmount.Max == nil { // max
			InvestOperationAmount = append(InvestOperationAmount, defaultMax)
		} else {
			InvestOperationAmount = append(InvestOperationAmount, *m.InvestOperationAmount.Max)
		}
	}
	if m.Deposit != nil && (m.Deposit.Min != nil || m.Deposit.Max != nil) {
		if m.Deposit.Min == nil { // min
			Deposit = append(Deposit, defaultMin)
		} else {
			Deposit = append(Deposit, *m.Deposit.Min)
		}
		if m.Deposit.Max == nil { // max
			Deposit = append(Deposit, defaultMax)
		} else {
			Deposit = append(Deposit, *m.Deposit.Max)
		}
	}
	if m.InvestSuccessAmount != nil && (m.InvestSuccessAmount.Min != nil || m.InvestSuccessAmount.Max != nil) {
		if m.InvestSuccessAmount.Min == nil { // min
			InvestSuccessAmount = append(InvestSuccessAmount, defaultMin)
		} else {
			InvestSuccessAmount = append(InvestSuccessAmount, *m.InvestSuccessAmount.Min)
		}
		if m.InvestSuccessAmount.Max == nil { // max
			InvestSuccessAmount = append(InvestSuccessAmount, defaultMax)
		} else {
			InvestSuccessAmount = append(InvestSuccessAmount, *m.InvestSuccessAmount.Max)
		}
	}
	if m.InvestRequestAmount != nil && (m.InvestRequestAmount.Min != nil || m.InvestRequestAmount.Max != nil) {
		if m.InvestRequestAmount.Min == nil { // min
			InvestRequestAmount = append(InvestRequestAmount, defaultMin)
		} else {
			InvestRequestAmount = append(InvestRequestAmount, *m.InvestRequestAmount.Min)
		}
		if m.InvestRequestAmount.Max == nil { // max
			InvestRequestAmount = append(InvestRequestAmount, defaultMax)
		} else {
			InvestRequestAmount = append(InvestRequestAmount, *m.InvestRequestAmount.Max)
		}
	}
	if m.UnpaidAmount != nil && (m.UnpaidAmount.Min != nil || m.UnpaidAmount.Max != nil) {
		if m.UnpaidAmount.Min == nil { // min
			UnpaidAmount = append(UnpaidAmount, defaultMin)
		} else {
			UnpaidAmount = append(UnpaidAmount, *m.UnpaidAmount.Min)
		}
		if m.UnpaidAmount.Max == nil { // max
			UnpaidAmount = append(UnpaidAmount, defaultMax)
		} else {
			UnpaidAmount = append(UnpaidAmount, *m.UnpaidAmount.Max)
		}
	}

	return &domain.MailSendOption{
		AuditStatus:           AuditStatus,
		UserType:              UserType,
		Nationality:           Nationality,
		Gender:                Gender,
		Job:                   Job,
		Asset:                 Asset,
		AnnualIncome:          AnnualIncome,
		AllowsNewsletter:      m.AllowsNewsletter,
		HasMyNumber:           m.HasMyNumber,
		SignUpRequestDate:     SignUpRequestDate,
		SignUpApprovalDate:    SignUpApprovalDate,
		LastLoginDate:         LastLoginDate,
		Age:                   Age,
		TotalAsset:            TotalAsset,
		InvestOperationAmount: InvestOperationAmount,
		Deposit:               Deposit,
		InvestSuccessAmount:   InvestSuccessAmount,
		InvestRequestAmount:   InvestRequestAmount,
		UnpaidAmount:          UnpaidAmount,
	}
}

type MailReceiversResponse struct {
	Id    int64  `json:"id" swagger:"desc(ユーザーID)"`
	Name  string `json:"name" swagger:"desc(ユーザー名)"`
	Email string `json:"email" swagger:"desc(ユーザーメール)"`
}

func (m MailReceiversResponse) ToDto(user *domain.User) *MailReceiversResponse {

	return &MailReceiversResponse{
		Id:    user.Id,
		Name:  user.Name,
		Email: user.Email,
	}
}

func (m MailReceiversResponse) ToDtoList(users []domain.User) []*MailReceiversResponse {

	var list []*MailReceiversResponse
	for _, user := range users {
		list = append(list, m.ToDto(&user))
	}

	return list
}

func generateInClause(arr []string) string {
	var quotedStrings []string
	for _, s := range arr {
		quotedStrings = append(quotedStrings, fmt.Sprintf("'%s'", s))
	}
	return strings.Join(quotedStrings, ",")
}

type CountReceiversWithOptionResponse struct {
	Count int `json:"count" swagger:"desc(受信者数)"`
}

type CountReceiversFromCsv struct {
	Count int `json:"count" swagger:"desc(受信者数)"`
}
