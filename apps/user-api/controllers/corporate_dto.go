package controllers

import (
	"strconv"

	"github.com/samber/lo"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

type CorporateSearchInput struct {
	StartDate    string  `query:"start_date"`
	EndDate      string  `query:"end_date"`
	Id           *string `query:"id"`
	Name         *string `query:"name"`
	InvestorType *string `query:"investor_type"`
	AuditStatus  *string `query:"audit_status"`
	InvestGrade  bool    `query:"investGrade"`
	BasicInfo    bool    `query:"basicInfo"`
	Withdrawal   bool    `query:"withdrawal"`
	Deposit      bool    `query:"deposit"`
	SearchInput
}

type AntiSocialCorporateSearchInput struct {
	StartDate string `query:"start_date"`
	EndDate   string `query:"end_date"`
	SearchInput
}

type CorporateUserResponse struct {
	Email                        string                         `json:"email" form:"email"`
	Password                     string                         `json:"password" form:"password"`
	Name                         string                         `json:"name" form:"name"`
	CorpNo                       string                         `json:"corp_no" form:"corp_no"`
	CorpType                     string                         `json:"corp_type" form:"corp_type"`
	ZipCode                      string                         `json:"zip_code" form:"zip_code"`
	Address                      string                         `json:"address" form:"address"`
	Address2                     string                         `json:"address2" form:"address2" `
	Province                     string                         `json:"province" form:"province"`
	City                         string                         `json:"city" form:"city"`
	MobilePhone                  string                         `json:"mobile_phone" form:"mobile_phone"`
	InvestmentPurpose            string                         `json:"investment_purpose" form:"investment_purpose" swagger:"desc(投資目的)"`
	IsSurplusFund                string                         `json:"is_surplus_fund" form:"is_surplus_fund" swagger:"desc(余裕資金運用有無)"`
	IsNewsletter                 string                         `json:"is_newsletter" form:"is_newsletter" swagger:"desc(ニュースレター受信希望)"`
	DepositType                  domain.DEPOSIT_TYPE            `json:"deposit_type" form:"deposit_type" swagger:"desc(預金種類 共通コード : [DEPOSIT_TYPE])"`
	BankName                     string                         `json:"bank_name" form:"bank_name" swagger:"desc(銀行名)"`
	BranchName                   string                         `json:"branch_name" form:"branch_name" swagger:"desc(支店名)"`
	BankAccount                  string                         `json:"bank_account" form:"bank_account" swagger:"desc(銀行口座番号)"`
	BankCode                     *string                        `json:"bank_code" form:"bank_code" swagger:"desc(銀行コード)"`
	BranchCode                   *string                        `json:"branch_code" form:"branch_code" swagger:"desc(銀行支店コード)"`
	Depositor                    string                         `json:"depositor" form:"depositor" swagger:"desc(預金者名)"`
	EmployeeCount                string                         `json:"employee_count" form:"employee_count" swagger:"desc(従業員数 - 企画チーム要求によりstring値)"`
	CapitalAmount                string                         `json:"capital_amount" form:"capital_amount"  swagger:"desc(資本金 - 企画チーム要求によりstring値)"`
	Sales                        string                         `json:"sales" form:"sales" swagger:"desc(売上額 - 企画チーム要求によりstring値)"`
	RepresentativeFamilyName     string                         `json:"representative_family_name" form:"representative_family_name" swagger:"desc(代表者名 姓)"`
	RepresentativeFirstName      string                         `json:"representative_first_name" form:"representative_first_name"  swagger:"desc(代表者名 名前)"`
	RepresentativeFamilyNameKana string                         `json:"representative_family_name_kana" form:"representative_family_name_kana" swagger:"desc(代表者名 姓ふりがな)"`
	RepresentativeFirstNameKana  string                         `json:"representative_first_name_kana" form:"representative_first_name_kana"  swagger:"desc(代表者名 名前ふりがな)"`
	RepresentativeMobilePhone    string                         `json:"representative_mobile_phone" form:"representative_mobile_phone" swagger:"desc(代表者携帯電話)"`
	RepresentativeEmail          string                         `json:"representative_email" form:"representative_email" swagger:"desc(代表者メール)"`
	RepresentativeSmsCertAt      jsonutils.JSONTime             `json:"representative_sms_cert_at" form:"representative_sms_cert_at" swagger:"desc(대표자 SMS 인증 시간)"`
	ManagerFamilyName            *string                        `json:"manager_family_name" form:"manager_family_name" swagger:"desc(担当者姓)"`
	ManagerFirstName             *string                        `json:"manager_first_name" form:"manager_first_name" swagger:"desc(担当者名前)"`
	ManagerFamilyNameKana        *string                        `json:"manager_family_name_kana" form:"manager_family_name_kana" swagger:"desc(担当者名 姓ふりがな)"`
	ManagerFirstNameKana         *string                        `json:"manager_first_name_kana" form:"manager_first_name_kana"  swagger:"desc(担当者名 名前ふりがな)"`
	ManagerEmail                 *string                        `json:"manager_email" form:"manager_email" swagger:"desc(担当者メール)"`
	ManagerMobilePhone           *string                        `json:"manager_mobile_phone" form:"manager_mobile_phone" swagger:"desc(担当者名 携帯電話)"`
	ManagerZipCode               string                         `json:"manager_zip_code" form:"manager_zip_code" swagger:"desc(担当者郵便番号)"`
	ManagerProvince              string                         `json:"manager_province" form:"manager_province" swagger:"desc(担当者都道府県名)"`
	ManagerCity                  string                         `json:"manager_city" form:"manager_city" swagger:"desc(担当者市区町村)"`
	ManagerAddress               string                         `json:"manager_address" form:"manager_address" swagger:"desc(担当者住所)"`
	ManagerAddress2              *string                        `json:"manager_address2" form:"manager_address2" swagger:"desc(担当者詳細住所)"`
	ManagerBirthday              jsonutils.JSONTime             `json:"manager_birthday" form:"manager_birthday" swagger:"desc(담당자 생년월일)"`
	Department                   *string                        `json:"department" form:"department" swagger:"desc(部署)"`
	Position                     *string                        `json:"position" form:"position" swagger:"desc(役職)"`
	RequestId                    string                         `json:"request_id" form:"request_id" swagger:"desc(リクエストID)"`
	AuditStatus                  domain.AUDIT_STATUS            `json:"audit_status" form:"audit_status" swagger:"desc(審査ステータス)"`
	Identification1FrontFile     []*domain.UserFile             `json:"identification1_front_file" form:"identification1_front_file" swagger:"desc(本人確認書類1 正面)"`
	Identification1BackFile      []*domain.UserFile             `json:"identification1_back_file"  form:"identification1_back_file" swagger:"desc(本人確認書類1 裏面)"`
	Identification2FrontFile     []*domain.UserFile             `json:"identification2_front_file" form:"identification2_front_file" swagger:"desc(本人確認書類2 正面)"`
	Identification2BackFile      []*domain.UserFile             `json:"identification2_back_file" form:"identification2_back_file" swagger:"desc(本人確認書類2 裏面)"`
	CertifiedCopyFiles           []*domain.UserFile             `json:"certified_copy_files"  form:"certified_copy_files" swagger:"desc(登記事項全部証明書)"`
	PoaFiles                     []*domain.UserFile             `json:"poa_files"  form:"poa_files" swagger:"desc(委任状)"`
	CorporatePostalStatus        bool                           `json:"corporate_postal_status" form:"corporate_postal_status" swagger:"desc(法人担当者郵便コード発送ステータス, 完了:True|未発送:False)"`
	InvestorType                 domain.INVESTOR_TYPE           `json:"investor_type" form:"investor_type" swagger:"desc(投資家タイプ)"`
	PostalCodeStatus             bool                           `json:"postal_code_status" form:"postal_code_status" swagger:"desc(担当者郵便コード発送ステータス, 完了:True|未発送:False)"`
	CorporateShareholder         []*domain.CorporateShareholder `json:"shareholder_list" xorm:"-" swagger:"desc(株主情報リスト)"`
	SignUpDate                   jsonutils.JSONTime             `json:"sign_up_date" form:"sign_up_date" swagger:"desc(登録日)"`
	UserId                       int64                          `json:"user_id" form:"user_id" swagger:"desc(사용자 ID)"`
	UserType                     domain.USER_TYPE               `json:"user_type" form:"user_type" swagger:"desc(사용자 유형)"`
}

func (u CorporateUserResponse) ToDto(cu *domain.CorporateUser) CorporateUserResponse {
	return CorporateUserResponse{
		Email:                        cu.Email,
		Password:                     cu.Password,
		Name:                         cu.Name,
		CorpNo:                       cu.CorpNo,
		CorpType:                     cu.CorpType,
		ZipCode:                      cu.ZipCode,
		Address:                      cu.Address,
		Address2:                     *cu.Address2,
		Province:                     cu.Province,
		City:                         cu.City,
		MobilePhone:                  cu.MobilePhone,
		InvestmentPurpose:            string(cu.InvestmentPurpose),
		IsSurplusFund:                strconv.FormatBool(lo.FromPtr(cu.IsSurplusFund)),
		IsNewsletter:                 strconv.FormatBool(cu.IsNewsletter),
		DepositType:                  cu.DepositType,
		BankName:                     cu.BankName,
		BranchName:                   cu.BranchName,
		BankAccount:                  cu.BankAccount,
		BankCode:                     cu.BankCode,
		BranchCode:                   cu.BranchCode,
		Depositor:                    cu.Depositor,
		EmployeeCount:                cu.EmployeeCount,
		CapitalAmount:                cu.CapitalAmount,
		Sales:                        cu.Sales,
		RepresentativeFamilyName:     cu.RepresentativeFamilyName,
		RepresentativeFirstName:      cu.RepresentativeFirstName,
		RepresentativeFamilyNameKana: cu.RepresentativeFamilyNameKana,
		RepresentativeFirstNameKana:  cu.RepresentativeFirstNameKana,
		RepresentativeMobilePhone:    cu.RepresentativeMobilePhone,
		RepresentativeEmail:          cu.RepresentativeEmail,
		RepresentativeSmsCertAt:      cu.RepresentativeSmsCertAt,
		ManagerFamilyName:            cu.ManagerFamilyName,
		ManagerFirstName:             cu.ManagerFirstName,
		ManagerFamilyNameKana:        cu.ManagerFamilyNameKana,
		ManagerFirstNameKana:         cu.ManagerFirstNameKana,
		ManagerEmail:                 cu.ManagerEmail,
		ManagerMobilePhone:           cu.ManagerMobilePhone,
		ManagerZipCode:               cu.ManagerZipCode,
		ManagerProvince:              cu.ManagerProvince,
		ManagerCity:                  cu.ManagerCity,
		ManagerAddress:               cu.ManagerAddress,
		ManagerAddress2:              cu.ManagerAddress2,
		ManagerBirthday:              cu.ManagerBirthday,
		Department:                   cu.Department,
		Position:                     cu.Position,
		AuditStatus:                  cu.AuditStatus,
		CertifiedCopyFiles:           cu.CertifiedCopyFiles,
		Identification1FrontFile:     cu.Identification1FrontFile,
		Identification1BackFile:      cu.Identification1BackFile,
		Identification2FrontFile:     cu.Identification2FrontFile,
		Identification2BackFile:      cu.Identification2BackFile,
		PoaFiles:                     cu.PoaFiles,
		CorporatePostalStatus:        cu.CorporatePostalStatus,
		InvestorType:                 cu.InvestorType,
		PostalCodeStatus:             cu.PostalCodeStatus,
		CorporateShareholder:         cu.CorporateShareholder,
		SignUpDate:                   cu.SignUpDate,
		UserId:                       cu.User.Id,
		UserType:                     cu.UserType,
	}
}
