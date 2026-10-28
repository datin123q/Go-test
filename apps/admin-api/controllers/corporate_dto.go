package controllers

import (
	"time"

	"github.com/gofrs/uuid"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
)

type CorporateUserAdminResponseDto struct {
	UserType          domain.USER_TYPE     `json:"user_type"`
	Name              string               `json:"name"`
	ZipCode           string               `json:"zip_code"`
	Address           string               `json:"address"`
	Address2          *string              `json:"address2"`
	Province          string               `json:"province"`
	City              string               `json:"city"`
	AuditStatus       domain.AUDIT_STATUS  `json:"audit_status"`
	BankAccount       string               `json:"bank_account"`
	BankName          string               `json:"bank_name"`
	BankCode          *string              `json:"bank_code"`
	BranchCode        *string              `json:"branch_code"`
	Depositor         string               `json:"depositor"`
	DepositType       domain.DEPOSIT_TYPE  `json:"deposit_type"`
	BranchName        string               `json:"branch_name"`
	MobilePhone       string               `json:"mobile_phone"`
	IsNewsletter      bool                 `json:"is_newsletter"`
	PostalCodeStatus  bool                 `json:"postal_code_status"`
	NewsletterAgreeAt jsonutils.JSONTime   `json:"newsletter_agree_at"`
	InvestorType      domain.INVESTOR_TYPE `json:"investor_type"`
}

func (CorporateUserAdminResponseDto) New(user *domain.CorporateUser) *CorporateUserAdminResponseDto {
	return &CorporateUserAdminResponseDto{
		UserType:          user.UserType,
		Name:              user.Name,
		ZipCode:           user.ZipCode,
		Address:           user.Address,
		Address2:          user.Address2,
		Province:          user.Province,
		City:              user.City,
		AuditStatus:       user.AuditStatus,
		BankAccount:       user.BankAccount,
		BankName:          user.BankName,
		BranchName:        user.BranchName,
		BankCode:          user.BankCode,
		BranchCode:        user.BranchCode,
		Depositor:         user.Depositor,
		DepositType:       user.DepositType,
		MobilePhone:       user.MobilePhone,
		IsNewsletter:      user.IsNewsletter,
		PostalCodeStatus:  user.PostalCodeStatus,
		NewsletterAgreeAt: user.NewsletterAgreeAt,
		InvestorType:      user.InvestorType,
	}
}

type CorporateSearchInput struct {
	StartDate           string  `query:"start_date"`
	EndDate             string  `query:"end_date"`
	Id                  *string `query:"id"`
	Name                *string `query:"name"`
	Email               *string `query:"email"`
	InvestorType        *string `query:"investor_type"`
	AuditStatus         *string `query:"audit_status"`
	InvestGrade         bool    `query:"investGrade"`
	BasicInfo           bool    `query:"basicInfo"`
	Withdrawal          bool    `query:"withdrawal"`
	LoanDeposit         bool    `query:"loanDeposit"`
	EquityDeposit       bool    `query:"equityDeposit"`
	NewsletterAgreement bool    `query:"newsletterAgreement"`
	ProfileSearchInput
	SearchInput
}

type AntiSocialCorporateSearchInput struct {
	StartDate string `query:"start_date"`
	EndDate   string `query:"end_date"`
	SearchInput
	QuerySearchInput
}

type CorporateUserInput struct {
	// WARNING: 管理者が「法人会員」情報を修正する際に使用するdto。フィールド追加、削除時に意図しない値が変更される可能性があるため注意が必要です。
	Name                  string                         `json:"name" form:"name" swagger:"desc(法人名)"`
	DepositType           domain.DEPOSIT_TYPE            `json:"deposit_type" form:"deposit_type" swagger:"desc(預金種類共通コード: [DEPOSIT_TYPE])"`
	ZipCode               string                         `json:"zip_code" form:"zip_code" swagger:"desc(郵便番号)"`
	Address               string                         `json:"address" form:"address" swagger:"desc(住所)"`
	Province              string                         `json:"province" form:"province" swagger:"desc(県)"`
	Address2              string                         `json:"address2" form:"address2" swagger:"desc(詳細住所)"`
	City                  string                         `json:"city" form:"city" swagger:"desc(市)"`
	BankName              string                         `json:"bank_name" form:"bank_name" swagger:"desc(銀行名)"`
	BranchName            string                         `json:"branch_name" form:"branch_name" swagger:"desc(支店名)"`
	BankAccount           string                         `json:"bank_account" form:"bank_account" validate:"omitempty,len=7,number" swagger:"desc(銀行口座番号)"`
	BankCode              *string                        `json:"bank_code" form:"bank_code" swagger:"desc(銀行コード)"`
	BranchCode            *string                        `json:"branch_code" form:"branch_code" swagger:"desc(銀行支店コード)"`
	Depositor             string                         `json:"depositor" form:"depositor" swagger:"desc(預金者)"`
	ManagerFamilyName     *string                        `json:"manager_family_name" form:"manager_family_name" swagger:"desc(担当者名姓)"`
	ManagerFirstName      *string                        `json:"manager_first_name" form:"manager_first_name" swagger:"desc(担当者名名前)"`
	ManagerFamilyNameKana *string                        `json:"manager_family_name_kana" form:"manager_family_name_kana" swagger:"desc(担当者名姓ふりがな)"`
	ManagerFirstNameKana  *string                        `json:"manager_first_name_kana" form:"manager_first_name_kana" swagger:"desc(担当者名名前ふりがな)"`
	ManagerEmail          *string                        `json:"manager_email" form:"manager_email" swagger:"desc(担当者名メール)"`
	ManagerZipCode        string                         `json:"manager_zip_code" form:"manager_zip_code" swagger:"desc(担当者郵便番号)"`
	ManagerProvince       string                         `json:"manager_province" form:"manager_province" swagger:"desc(担当者住所: 県)"`
	ManagerCity           string                         `json:"manager_city" form:"manager_city" swagger:"desc(担当者住所: 市)"`
	ManagerAddress        string                         `json:"manager_address" form:"manager_address" swagger:"desc(担当者住所)"`
	ManagerAddress2       *string                        `json:"manager_address2" form:"manager_address2" swagger:"desc(担当者詳細住所)"`
	ManagerBirthday       string                         `json:"manager_birthday" form:"manager_birthday" swagger:"desc(担当者生年月日)"`
	Department            *string                        `json:"department" form:"department" swagger:"desc(担当部門)"`
	Position              *string                        `json:"position" form:"position" swagger:"desc(担当者職務)"`
	ManagerMobilePhone    *string                        `json:"manager_mobile_phone" form:"manager_mobile_phone"  swagger:"desc(担当者電話番号)"`
	IsSurplusFund         *bool                          `json:"is_surplus_fund" form:"is_surplus_fund" swagger:"desc(余裕資金運用有無)"`
	InvestmentPurpose     domain.INVESTMENT_PURPOSE_TYPE `json:"investment_purpose" form:"investment_purpose" swagger:"desc(投資目的)"`
	PostalCodeStatus      bool                           `json:"postal_code_status" form:"postal_code_status" swagger:"desc(法人住所地郵便コード認証有無、送信完了: True | 未送信: False)"`
	CorporatePostalStatus bool                           `json:"corporate_postal_status" form:"corporate_postal_status" swagger:"desc(法人担当者郵便コード認証有無、送信完了: True | 未送信: False)"`
	HasExperience         *bool                          `json:"has_experience" form:"has_experience" swagger:"desc(投資経験有無)"`
	EmployeeCount         string                         `json:"employee_count" form:"employee_count" swagger:"desc(従業員数)"`
	CapitalAmount         string                         `json:"capital_amount" form:"capital_amount" swagger:"desc(資本金)"`
	Sales                 string                         `json:"sales" form:"sales" swagger:"desc(売上高)"`
	IsNewsletter          *bool                          `json:"is_newsletter" form:"is_newsletter" swagger:"desc(メルマガ購読有無)"`
}

func (u *CorporateUserInput) ToModel() (*domain.CorporateUser, error) {

	bd, err := jsonutils.ParseJSONTimeJST(time.DateTime, u.ManagerBirthday)
	if err != nil {
		return nil, err
	}

	return &domain.CorporateUser{
		User: domain.User{
			UserType:          domain.CORPORATE,
			Name:              u.Name,
			ZipCode:           u.ZipCode,
			Address:           u.Address,
			Address2:          &u.Address2,
			Province:          u.Province,
			City:              u.City,
			Depositor:         u.Depositor,
			InvestmentPurpose: u.InvestmentPurpose,
			IsSurplusFund:     u.IsSurplusFund,
			BankAccount:       u.BankAccount,
			BankCode:          u.BankCode,
			BranchCode:        u.BranchCode,
			BankName:          u.BankName,
			BranchName:        u.BranchName,
			DepositType:       u.DepositType,
			PostalCodeStatus:  u.PostalCodeStatus,
			HasExperience:     u.HasExperience,
		},

		CorporateProfile: domain.CorporateProfile{
			ManagerFamilyName:     u.ManagerFamilyName,
			ManagerFirstName:      u.ManagerFirstName,
			ManagerFamilyNameKana: u.ManagerFamilyNameKana,
			ManagerFirstNameKana:  u.ManagerFirstNameKana,
			ManagerEmail:          u.ManagerEmail,
			ManagerMobilePhone:    u.ManagerMobilePhone,
			Department:            u.Department,
			Position:              u.Position,
			ManagerZipCode:        u.ManagerZipCode,
			ManagerProvince:       u.ManagerProvince,
			ManagerCity:           u.ManagerCity,
			ManagerAddress:        u.ManagerAddress,
			ManagerAddress2:       u.ManagerAddress2,
			ManagerBirthday:       bd,
			CorporatePostalStatus: u.CorporatePostalStatus,
			EmployeeCount:         u.EmployeeCount,
			CapitalAmount:         u.CapitalAmount,
			Sales:                 u.Sales,
		},
	}, nil
}

type CorporateSignUpRequestInput struct {
	Email string `json:"email" form:"email" validate:"required,email"`
}

func (v *CorporateSignUpRequestInput) ToModel() (*domain.CorporateSignUpRequest, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}
	requestId := id.String()
	url := serviceutils.GetUserServiceDomain() + "/signUp/corporate" + "/" + requestId

	return &domain.CorporateSignUpRequest{
		RequestId: requestId,
		Email:     v.Email,
		Url:       url,
		Expire:    jsonutils.JSONTime(time.Now().Add(time.Minute * 20)),
	}, nil
}

type CorpNoInput struct {
	CorpNo string `json:"corp_no" form:"corp_no"  validate:"required"`
}
