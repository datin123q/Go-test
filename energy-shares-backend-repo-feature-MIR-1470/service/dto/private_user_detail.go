package dto

import (
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

type PrivateUserAdminDetail struct {
	Id                       int64                          `json:"id"`
	Uid                      string                         `json:"uid"`
	UserId                   int64                          `json:"user_id"`
	CommonId                 string                         `json:"common_id"`
	UserType                 domain.USER_TYPE               `json:"user_type"`
	Name                     string                         `json:"name"`
	NameKana                 string                         `json:"name_kana"`
	FirstName                string                         `json:"first_name"`
	FamilyName               string                         `json:"family_name"`
	FirstNameKana            string                         `json:"first_name_kana"`
	FamilyNameKana           string                         `json:"family_name_kana"`
	Email                    string                         `json:"email"`
	Birthday                 jsonutils.JSONTime             `json:"birthday"`
	InvestorType             domain.INVESTOR_TYPE           `json:"investor_type"`
	BlockInvest              bool                           `json:"block_invest"`
	Roles                    []string                       `json:"roles"`
	SignUpDate               jsonutils.JSONTime             `json:"sign_up_date"`
	LoginType                domain.LOGIN_TYPE              `json:"login_type"`
	LastLoginDate            jsonutils.JSONTime             `json:"last_login_date"`
	MobilePhone              string                         `json:"mobile_phone"`
	ZipCode                  string                         `json:"zip_code"`
	Address                  string                         `json:"address"`
	Address2                 *string                        `json:"address2"`
	Province                 string                         `json:"province"`
	City                     string                         `json:"city"`
	AuditStatus              domain.AUDIT_STATUS            `json:"audit_status"`
	BankName                 string                         `json:"bank_name"`
	BranchName               string                         `json:"branch_name"`
	BankCode                 *string                        `json:"bank_code"`
	BranchCode               *string                        `json:"branch_code"`
	DepositType              domain.DEPOSIT_TYPE            `json:"deposit_type"`
	Depositor                string                         `json:"depositor"` // 預金者
	BankAccount              string                         `json:"bank_account"`
	Asset                    domain.ASSETS_STATUS           `json:"asset"`
	AnnualIncome             domain.ANNUALINCOME_STATUS     `json:"annual_income"`
	IsProInvestAsset         *bool                          `json:"is_pro_invest_asset"`
	IsProInvestFinanceAsset  *bool                          `json:"is_pro_invest_finance_asset"`
	Nationality              domain.NATIONALITY_TYPE        `json:"nationality"`
	Identification1FrontFile []*domain.UserFile             `json:"identification1_front_file" xorm:"-" swagger:"desc(본인확인서류 전면 1)"`
	Identification1BackFile  []*domain.UserFile             `json:"identification1_back_file" xorm:"-" swagger:"desc(본인확인서류 후면 1)"`
	Identification2FrontFile []*domain.UserFile             `json:"identification2_front_file" xorm:"-" swagger:"desc(본인확인서류 전면 2)"`
	Identification2BackFile  []*domain.UserFile             `json:"identification2_back_file" xorm:"-" swagger:"desc(본인확인서류 후면 2)"`
	EkycIdentificationFiles  []*domain.UserFile             `json:"ekyc_identification_files"`
	InvestmentPurpose        domain.INVESTMENT_PURPOSE_TYPE `json:"investment_purpose"`
	IsSurplusFund            *bool                          `json:"is_surplus_fund"`
	HasExperience            *bool                          `json:"has_experience"`
	IsNewsletter             bool                           `json:"is_newsletter"`
	NewsletterAgreeAt        jsonutils.JSONTime             `json:"newsletter_agree_at"`
	IsSmsCert                bool                           `json:"is_sms_cert"`
	SmsCertAt                jsonutils.JSONTime             `json:"sms_cert_at"`
	PostalCode               string                         `json:"postal_code"`
	PostalCodeStatus         bool                           `json:"postal_code_status"`
	PostalCodeSendDate       jsonutils.JSONTime             `json:"postal_code_send_date"`
	Job                      domain.USER_JOB_TYPE           `json:"job"`
	SignUpApprovalDate       jsonutils.JSONTime             `json:"sign_up_approval_date"`
	CodeApprovalDate         jsonutils.JSONTime             `json:"code_approval_date"`
	EquityVaAccount          *domain.UserVirtualAccount     `json:"equity_va_account"`
	AntisocialCheck          *domain.UserAntisocialCheck    `json:"antisocial_check"`
	InvalidReason            *string                        `json:"invalid_reason"`
	MyNumber                 *domain.MyNumber               `json:"my_number"`
	Gender                   domain.GENDER                  `json:"gender"`
	UserAccount              *domain.UserAccount            `json:"user_account"`
	IsEkyc                   bool                           `json:"is_ekyc"`
	AntiResearchDate         jsonutils.JSONTime             `json:"anti_research_date"`
	RequestId                string                         `json:"request_id"`
	EkycResult               *EkycResult                    `json:"ekyc_result"`
	HasEkycImage             bool                           `json:"has_ekyc_image"`
	DuplicateUsers           []DuplicateUser                `json:"duplicate_users"`
	Updated                  jsonutils.JSONTime             `json:"updated"`
	Created                  jsonutils.JSONTime             `json:"created"`
}

type DuplicateUser struct {
	CheckedUserID   int64  `json:"checked_user_id"`
	DuplicateUserID int64  `json:"duplicate_user_id"`
	DuplicateReason string `json:"duplicate_reason"`
}

type EkycResult struct {
	Result  domain.EKYC_RESULT_STATUS `json:"result"`  // MATCH, UNMATCH
	Error   *string                   `json:"error"`   // error message
	Created jsonutils.JSONTime        `json:"created"` // ekyc 제출일자
}

func (PrivateUserAdminDetail) New(user *domain.PrivateUser, EquityAccount *domain.UserVirtualAccount, AntisocialCheck *domain.UserAntisocialCheck, UserAccount *domain.UserAccount, myNumber *domain.MyNumber, ekycReq *domain.EkycRequestLog) *PrivateUserAdminDetail {
	name := user.FamilyName + " " + user.FirstName
	nameKana := user.FamilyNameKana + " " + user.FirstNameKana

	dto := &PrivateUserAdminDetail{
		Id:                      user.User.Id,
		Uid:                     user.Uid,
		UserId:                  user.PrivateProfile.UserId,
		CommonId:                user.CommonId,
		UserType:                user.UserType,
		Name:                    name,
		NameKana:                nameKana,
		FirstName:               user.FirstName,
		FamilyName:              user.FamilyName,
		FirstNameKana:           user.FirstNameKana,
		FamilyNameKana:          user.FamilyNameKana,
		Email:                   user.Email,
		Birthday:                user.Birthday,
		InvestorType:            user.InvestorType,
		BlockInvest:             user.BlockInvest,
		Roles:                   user.Roles,
		SignUpDate:              user.SignUpDate,
		LoginType:               user.LoginType,
		LastLoginDate:           user.LastLoginDate,
		MobilePhone:             user.MobilePhone,
		ZipCode:                 user.ZipCode,
		Address:                 user.Address,
		Address2:                user.Address2,
		Province:                user.Province,
		City:                    user.City,
		AuditStatus:             user.AuditStatus,
		BankName:                user.BankName,
		BranchName:              user.BranchName,
		BankCode:                user.BankCode,
		BranchCode:              user.BranchCode,
		DepositType:             user.DepositType,
		Depositor:               user.Depositor,
		BankAccount:             user.BankAccount,
		Asset:                   user.Asset,
		AnnualIncome:            user.AnnualIncome,
		IsProInvestAsset:        user.IsProInvestAsset,
		IsProInvestFinanceAsset: user.IsProInvestFinanceAsset,
		Nationality:             user.Nationality,
		InvestmentPurpose:       user.InvestmentPurpose,
		IsSurplusFund:           user.IsSurplusFund,
		HasExperience:           user.HasExperience,
		IsNewsletter:            user.IsNewsletter,
		NewsletterAgreeAt:       user.NewsletterAgreeAt,
		IsSmsCert:               user.IsSmsCert,
		SmsCertAt:               user.SmsCertAt,
		PostalCode:              user.PostalCode,
		PostalCodeStatus:        user.PostalCodeStatus,
		PostalCodeSendDate:      user.PostalCodeSendDate,
		Job:                     user.Job,
		SignUpApprovalDate:      user.SignUpApprovalDate,
		CodeApprovalDate:        user.CodeApprovalDate,
		EquityVaAccount:         EquityAccount,
		AntisocialCheck:         AntisocialCheck,
		InvalidReason:           user.InvalidReason,
		MyNumber:                myNumber,
		UserAccount:             UserAccount,
		Gender:                  user.Gender,
		IsEkyc:                  user.IsEkyc,
		AntiResearchDate:        user.AntiResearchDate,
		RequestId:               user.RequestId,
		Updated:                 user.User.Updated,
		Created:                 user.User.Created,
	}

	if ekycReq != nil {
		dto.EkycResult = &EkycResult{
			Result:  ekycReq.Result,
			Error:   ekycReq.ErrorMessage,
			Created: ekycReq.Created,
		}
	}

	return dto
}

type PrivateUserCsv struct {
	Id                      int64                          `json:"id"`
	Uid                     string                         `json:"uid"`
	UserId                  int64                          `json:"user_id"`
	CommonId                string                         `json:"common_id"`
	UserType                domain.USER_TYPE               `json:"user_type"`
	Name                    string                         `json:"name"`
	NameKana                string                         `json:"name_kana"`
	FirstName               string                         `json:"first_name"`
	FamilyName              string                         `json:"family_name"`
	FirstNameKana           string                         `json:"first_name_kana"`
	FamilyNameKana          string                         `json:"family_name_kana"`
	Email                   string                         `json:"email"`
	Birthday                jsonutils.JSONTime             `json:"birthday"`
	InvestorType            domain.INVESTOR_TYPE           `json:"investor_type"`
	BlockInvest             bool                           `json:"block_invest"`
	Roles                   []string                       `json:"roles"`
	SignUpDate              jsonutils.JSONTime             `json:"sign_up_date"`
	LoginType               domain.LOGIN_TYPE              `json:"login_type"`
	LastLoginDate           jsonutils.JSONTime             `json:"last_login_date"`
	MobilePhone             string                         `json:"mobile_phone"`
	ZipCode                 string                         `json:"zip_code"`
	Address                 string                         `json:"address"`
	Address2                *string                        `json:"address2"`
	Province                string                         `json:"province"`
	City                    string                         `json:"city"`
	AuditStatus             domain.AUDIT_STATUS            `json:"audit_status"`
	BankName                string                         `json:"bank_name"`
	BranchName              string                         `json:"branch_name"`
	BankCode                *string                        `json:"bank_code"`
	BranchCode              *string                        `json:"branch_code"`
	DepositType             domain.DEPOSIT_TYPE            `json:"deposit_type"`
	Depositor               string                         `json:"depositor"` // 預金者
	BankAccount             string                         `json:"bank_account"`
	Asset                   domain.ASSETS_STATUS           `json:"asset"`
	AnnualIncome            domain.ANNUALINCOME_STATUS     `json:"annual_income"`
	IsProInvestAsset        *bool                          `json:"is_pro_invest_asset"`
	IsProInvestFinanceAsset *bool                          `json:"is_pro_invest_finance_asset"`
	Nationality             domain.NATIONALITY_TYPE        `json:"nationality"`
	InvestmentPurpose       domain.INVESTMENT_PURPOSE_TYPE `json:"investment_purpose"`
	IsSurplusFund           *bool                          `json:"is_surplus_fund"`
	IsNewsletter            bool                           `json:"is_newsletter"`
	NewsletterAgreeAt       jsonutils.JSONTime             `json:"newsletter_agree_at"`
	IsSmsCert               bool                           `json:"sms_cert"`
	SmsCertAt               jsonutils.JSONTime             `json:"sms_cert_at"`
	PostalCode              string                         `json:"postal_code"`
	PostalCodeStatus        bool                           `json:"postal_code_status"`
	PostalCodeSendDate      jsonutils.JSONTime             `json:"postal_code_send_date"`
	Job                     domain.USER_JOB_TYPE           `json:"job"`
	SignUpApprovalDate      jsonutils.JSONTime             `json:"sign_up_approval_date"`
	CodeApprovalDate        jsonutils.JSONTime             `json:"code_approval_date"`
	EquityVaAccount         *domain.UserVirtualAccount     `json:"equity_va_account"`
	InvalidReason           *string                        `json:"invalid_reason"`
	Gender                  domain.GENDER                  `json:"gender"`
	RequestId               string                         `json:"request_id"`
	Updated                 jsonutils.JSONTime             `json:"updated"`
	Created                 jsonutils.JSONTime             `json:"created"`
}

func (PrivateUserCsv) New(user *domain.PrivateUser, EquityAccount *domain.UserVirtualAccount) *PrivateUserCsv {
	name := user.FamilyName + " " + user.FirstName
	nameKana := user.FamilyNameKana + " " + user.FirstNameKana

	return &PrivateUserCsv{
		Id:                      user.User.Id,
		Uid:                     user.Uid,
		UserId:                  user.PrivateProfile.UserId,
		CommonId:                user.CommonId,
		UserType:                user.UserType,
		Name:                    name,
		NameKana:                nameKana,
		FirstName:               user.FirstName,
		FamilyName:              user.FamilyName,
		FirstNameKana:           user.FirstNameKana,
		FamilyNameKana:          user.FamilyNameKana,
		Email:                   user.Email,
		Birthday:                user.Birthday,
		InvestorType:            user.InvestorType,
		BlockInvest:             user.BlockInvest,
		Roles:                   user.Roles,
		SignUpDate:              user.SignUpDate,
		LoginType:               user.LoginType,
		LastLoginDate:           user.LastLoginDate,
		MobilePhone:             user.MobilePhone,
		ZipCode:                 user.ZipCode,
		Address:                 user.Address,
		Address2:                user.Address2,
		Province:                user.Province,
		City:                    user.City,
		AuditStatus:             user.AuditStatus,
		BankName:                user.BankName,
		BranchName:              user.BranchName,
		BankCode:                user.BankCode,
		BranchCode:              user.BranchCode,
		DepositType:             user.DepositType,
		Depositor:               user.Depositor,
		BankAccount:             user.BankAccount,
		Asset:                   user.Asset,
		AnnualIncome:            user.AnnualIncome,
		IsProInvestAsset:        user.IsProInvestAsset,
		IsProInvestFinanceAsset: user.IsProInvestFinanceAsset,
		Nationality:             user.Nationality,
		InvestmentPurpose:       user.InvestmentPurpose,
		IsSurplusFund:           user.IsSurplusFund,
		IsNewsletter:            user.IsNewsletter,
		NewsletterAgreeAt:       user.NewsletterAgreeAt,
		IsSmsCert:               user.IsSmsCert,
		SmsCertAt:               user.SmsCertAt,
		PostalCode:              user.PostalCode,
		PostalCodeStatus:        user.PostalCodeStatus,
		PostalCodeSendDate:      user.PostalCodeSendDate,
		Job:                     user.Job,
		SignUpApprovalDate:      user.SignUpApprovalDate,
		CodeApprovalDate:        user.CodeApprovalDate,
		EquityVaAccount:         EquityAccount,
		InvalidReason:           user.InvalidReason,
		Gender:                  user.Gender,
		RequestId:               user.RequestId,
		Updated:                 user.User.Updated,
		Created:                 user.User.Created,
	}
}
