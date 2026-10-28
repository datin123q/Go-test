package dto

import (
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

type CorporateUserAdminDetail struct {
	Id                              int64                          `json:"-"`
	Uid                             string                         `json:"uid"`
	UserId                          int64                          `json:"user_id"  swagger:"desc(사용자 ID)"`
	CommonId                        string                         `json:"common_id"  swagger:"desc(공통 ID)"`
	UserType                        domain.USER_TYPE               `json:"user_type"  swagger:"desc(사용자 타입)"`
	Name                            string                         `json:"name"  swagger:"desc(이름)"`
	CorpNo                          string                         `json:"corp_no"  swagger:"desc(법인 번호)"`
	CorpType                        string                         `json:"corp_type"  swagger:"desc(법인 업종)"`
	Email                           string                         `json:"email"  swagger:"desc(이메일)"`
	LoginType                       domain.LOGIN_TYPE              `json:"login_type"  swagger:"desc(로그인 타입)"`
	InvestorType                    domain.INVESTOR_TYPE           `json:"investor_type"  swagger:"desc(투자자 타입)"`
	BlockInvest                     bool                           `json:"block_invest"  swagger:"desc(투자 차단 여부)"`
	Roles                           []string                       `json:"roles"  swagger:"desc(권한)"`
	SignUpDate                      jsonutils.JSONTime             `json:"sign_up_date"  swagger:"desc(가입일)"`
	LastLoginDate                   jsonutils.JSONTime             `json:"last_login_date"  swagger:"desc(최근 로그인 일)"`
	MobilePhone                     string                         `json:"mobile_phone"  swagger:"desc(휴대폰 번호)"`
	ZipCode                         string                         `json:"zip_code"  swagger:"desc(우편번호)"`
	Address                         string                         `json:"address"  swagger:"desc(주소)"`
	Address2                        *string                        `json:"address2"  swagger:"desc(상세 주소)"`
	Province                        string                         `json:"province"  swagger:"desc(시/도)"`
	City                            string                         `json:"city"  swagger:"desc(시/군/구)"`
	AuditStatus                     domain.AUDIT_STATUS            `json:"audit_status"  swagger:"desc(심사 상태)"`
	InvestmentPurpose               domain.INVESTMENT_PURPOSE_TYPE `json:"investment_purpose"  swagger:"desc(투자 목적)"`
	IsSurplusFund                   *bool                          `json:"is_surplus_fund"  swagger:"desc(잔여 자금 여부)"`
	HasExperience                   *bool                          `json:"has_experience" swagger:"desc(투자 경험 여부)"`
	IsNewsletter                    bool                           `json:"is_newsletter"  swagger:"desc(뉴스레터 수신 여부)"`
	NewsletterAgreeAt               jsonutils.JSONTime             `json:"newsletter_agree_at"  swagger:"desc(뉴스레터 수신 동의일)"`
	IsSmsCert                       bool                           `json:"is_sms_cert"  swagger:"desc(휴대폰 인증 여부)"`
	SmsCertAt                       jsonutils.JSONTime             `json:"sms_cert_at"  swagger:"desc(휴대폰 인증일)"`
	BankName                        string                         `json:"bank_name"  swagger:"desc(은행명)"`
	BranchName                      string                         `json:"branch_name"  swagger:"desc(지점명)"`
	BankAccount                     string                         `json:"bank_account"  swagger:"desc(계좌번호)"`
	Depositor                       string                         `json:"depositor"  swagger:"desc(예금주)"`
	BankCode                        *string                        `json:"bank_code"  swagger:"desc(은행코드)"`
	BranchCode                      *string                        `json:"branch_code"  swagger:"desc(지점코드)"`
	DepositType                     domain.DEPOSIT_TYPE            `json:"deposit_type"  swagger:"desc(예금종류)"`
	EmployeeCount                   string                         `json:"employee_count"  swagger:"desc(사원 수 기획팀 요청으로 string 값 )"`
	CapitalAmount                   string                         `json:"capital_amount"  swagger:"desc(자본금 기획팀 요청으로 string 값 )"`
	Sales                           string                         `json:"sales"  swagger:"desc(매출액 기획팀 요청으로 string 값 )"`
	RepresentativeFamilyName        string                         `json:"representative_family_name" swagger:"desc(대표자 명 성)"`
	RepresentativeFirstName         string                         `json:"representative_first_name"  swagger:"desc(대표자 명 이름)"`
	RepresentativeFamilyNameKana    string                         `json:"representative_family_name_kana"  swagger:"desc(대표자 명 성 후리가나)"`
	RepresentativeFirstNameKana     string                         `json:"representative_first_name_kana"  swagger:"desc(대표자 명 이름 후리가나)"`
	RepresentativeMobilePhone       string                         `json:"representative_mobile_phone"  swagger:"desc(대표자 핸드폰)"`
	RepresentativeEmail             string                         `json:"representative_email"  swagger:"desc(대표자 이메일)"`
	ManagerFamilyName               *string                        `json:"manager_family_name" swagger:"desc(담당자 명 성)"`
	ManagerFirstName                *string                        `json:"manager_first_name"  swagger:"desc(담당자 명 이름)"`
	ManagerFamilyNameKana           *string                        `json:"manager_family_name_kana" swagger:"desc(담당자 명 성 후리가나)"`
	ManagerFirstNameKana            *string                        `json:"manager_first_name_kana" swagger:"desc(담당자 명 이름 후리가나)"`
	ManagerEmail                    *string                        `json:"manager_email"  swagger:"desc(담당자 이메일)"`
	ManagerMobilePhone              *string                        `json:"manager_mobile_phone"  swagger:"desc(담당자 핸드폰)"`
	ManagerZipCode                  string                         `json:"manager_zip_code" swagger:"desc(담당자 우편번호)"`
	ManagerProvince                 string                         `json:"manager_province"  swagger:"desc(담당자 현 이름)"`
	ManagerCity                     string                         `json:"manager_city"  swagger:"desc(담당자 시 도시)"`
	ManagerAddress                  string                         `json:"manager_address"  swagger:"desc(담당자 주소)"`
	ManagerAddress2                 *string                        `json:"manager_address2"  swagger:"desc(담당자 주소2)"`
	ManagerBirthday                 jsonutils.JSONTime             `json:"manager_birthday" swagger:"desc(담당자 생년월일)"`
	IsManagerSmsCert                bool                           `json:"is_manager_sms_cert"  swagger:"desc(담당자 SMS 문자인증 여부)"`
	ManagerSmsCertAt                jsonutils.JSONTime             `json:"manager_sms_cert_at" swagger:"desc(담당자 SMS 문자인증 시점)"`
	Department                      *string                        `json:"department"  swagger:"desc(부서)"`
	Position                        *string                        `json:"position" swagger:"desc(직책)"`
	PostalCode                      string                         `json:"postal_code"  swagger:"desc(우편번호)"`
	PostalCodeStatus                bool                           `json:"postal_code_status"  swagger:"desc(우편번호 상태)"`
	PostalCodeSendDate              jsonutils.JSONTime             `json:"postal_code_send_date"  swagger:"desc(우편번호 발송일)"`
	CorporatePostalCode             string                         `json:"corporate_postal_code" swagger:"desc(법인 담당자 우편 코드)"`
	CorporatePostalCodeApprovalDate jsonutils.JSONTime             `json:"corporate_postal_code_approval_date" swagger:"desc(법인 담당자 우편 코드 인증 일자 )"`
	CorporatePostalStatus           bool                           `json:"corporate_postal_status"  swagger:"desc(법인 담당자 우편 코드 발송 상태 발송완료: True 미발송 False)"`
	CorporatePostalCodeSendDate     jsonutils.JSONTime             `json:"corporate_postal_code_send_date" swagger:"desc(법인 담당자 우편 코드 발송 일자 )"`
	SignUpApprovalDate              jsonutils.JSONTime             `json:"sign_up_approval_date" swagger:"desc(가입 승인 일자 )"`
	CodeApprovalDate                jsonutils.JSONTime             `json:"code_approval_date" swagger:"desc(코드 승인 일자 )"`
	InvalidReason                   *string                        `json:"invalid_reason" swagger:"desc(무효 사유 )"`
	IsRepresentativeSmsCert         bool                           `json:"is_representative_sms_cert" swagger:"desc(대표자 핸드폰 인증 일자 false 미인증)"`
	RepresentativeSmsCertAt         jsonutils.JSONTime             `json:"representative_sms_cert_at" swagger:"enum(true|false)"`
	Identification1FrontFile        []*domain.UserFile             `json:"identification1_front_file" swagger:"desc(본인 확인 서류 1 전면)"`
	Identification1BackFile         []*domain.UserFile             `json:"identification1_back_file"  swagger:"desc(본인 확인 서류 1 후면)"`
	Identification2FrontFile        []*domain.UserFile             `json:"identification2_front_file"  swagger:"desc(본인 확인 서류 2 전면)"`
	Identification2BackFile         []*domain.UserFile             `json:"identification2_back_file" swagger:"desc(본인 확인 서류 2 후면)"`
	CertifiedCopyFiles              []*domain.UserFile             `json:"certified_copy_files"  swagger:"desc(이력 사항 전부 확인서)"`
	PoaFiles                        []*domain.UserFile             `json:"poa_files"  swagger:"desc(위임장)"`
	AoiFiles                        []*domain.UserFile             `json:"aoi_files" swagger:"desc(정관)"`
	ShareholdersFiles               []*domain.UserFile             `json:"shareholders_files" swagger:"desc(주주 명부)"`
	OthersFiles                     []*domain.UserFile             `json:"others_files" swagger:"desc(기타 서류)"`
	CorporateShareholder            []*domain.CorporateShareholder `json:"shareholder_list" swagger:"desc(주주 정보 목록)"`
	EquityVaAccount                 *domain.UserVirtualAccount     `json:"equity_va_account" swagger:"desc(주식 가상 계좌 정보)"`
	AntisocialCheck                 *domain.UserAntisocialCheck    `json:"antisocial_check" swagger:"desc(안티 사회자 정보)"`
	MyNumber                        *domain.MyNumber               `json:"my_number" swagger:"desc(마이 넘버 정보)"`
	UserAccount                     *domain.UserAccount            `json:"user_account" swagger:"desc(사용자 계좌 정보)"`
	AntiResearchDate                jsonutils.JSONTime             `json:"anti_research_date" swagger:"desc(안티 사회자 조회 일자)"`
	Created                         jsonutils.JSONTime             `json:"created"`
	Updated                         jsonutils.JSONTime             `json:"updated"`
}

func (CorporateUserAdminDetail) New(user domain.CorporateUser) *CorporateUserAdminDetail {
	return &CorporateUserAdminDetail{
		Id:                              user.User.Id,
		Uid:                             user.User.Uid,
		UserId:                          user.CorporateProfile.UserId,
		CommonId:                        user.CommonId,
		UserType:                        user.UserType,
		Name:                            user.Name,
		CorpNo:                          user.CorpNo,
		CorpType:                        user.CorpType,
		Email:                           user.Email,
		LoginType:                       user.LoginType,
		InvestorType:                    user.InvestorType,
		BlockInvest:                     user.BlockInvest,
		Roles:                           user.Roles,
		SignUpDate:                      user.SignUpDate,
		LastLoginDate:                   user.LastLoginDate,
		MobilePhone:                     user.MobilePhone,
		ZipCode:                         user.ZipCode,
		Address:                         user.Address,
		Address2:                        user.Address2,
		Province:                        user.Province,
		City:                            user.City,
		AuditStatus:                     user.AuditStatus,
		InvestmentPurpose:               user.InvestmentPurpose,
		IsSurplusFund:                   user.IsSurplusFund,
		IsNewsletter:                    user.IsNewsletter,
		NewsletterAgreeAt:               user.NewsletterAgreeAt,
		IsSmsCert:                       user.IsSmsCert,
		SmsCertAt:                       user.SmsCertAt,
		BankName:                        user.BankName,
		BranchName:                      user.BranchName,
		BankAccount:                     user.BankAccount,
		Depositor:                       user.Depositor,
		BankCode:                        user.BankCode,
		BranchCode:                      user.BranchCode,
		DepositType:                     user.DepositType,
		EmployeeCount:                   user.EmployeeCount,
		CapitalAmount:                   user.CapitalAmount,
		Sales:                           user.Sales,
		RepresentativeFamilyName:        user.RepresentativeFamilyName,
		RepresentativeFirstName:         user.RepresentativeFirstName,
		RepresentativeFamilyNameKana:    user.RepresentativeFamilyNameKana,
		RepresentativeFirstNameKana:     user.RepresentativeFirstNameKana,
		RepresentativeMobilePhone:       user.RepresentativeMobilePhone,
		RepresentativeEmail:             user.RepresentativeEmail,
		ManagerFamilyName:               user.ManagerFamilyName,
		ManagerFirstName:                user.ManagerFirstName,
		ManagerFamilyNameKana:           user.ManagerFamilyNameKana,
		ManagerFirstNameKana:            user.ManagerFirstNameKana,
		ManagerEmail:                    user.ManagerEmail,
		ManagerMobilePhone:              user.ManagerMobilePhone,
		ManagerZipCode:                  user.ManagerZipCode,
		ManagerProvince:                 user.ManagerProvince,
		ManagerCity:                     user.ManagerCity,
		ManagerAddress:                  user.ManagerAddress,
		ManagerAddress2:                 user.ManagerAddress2,
		ManagerBirthday:                 user.ManagerBirthday,
		IsManagerSmsCert:                user.IsManagerSmsCert,
		ManagerSmsCertAt:                user.ManagerSmsCertAt,
		Department:                      user.Department,
		Position:                        user.Position,
		PostalCode:                      user.PostalCode,
		PostalCodeStatus:                user.PostalCodeStatus,
		PostalCodeSendDate:              user.PostalCodeSendDate,
		CorporatePostalCode:             user.CorporatePostalCode,
		CorporatePostalCodeApprovalDate: user.CorporatePostalCodeApprovalDate,
		CorporatePostalStatus:           user.CorporatePostalStatus,
		CorporatePostalCodeSendDate:     user.CorporatePostalCodeSendDate,
		SignUpApprovalDate:              user.SignUpApprovalDate,
		CodeApprovalDate:                user.CodeApprovalDate,
		InvalidReason:                   user.InvalidReason,
		IsRepresentativeSmsCert:         user.IsRepresentativeSmsCert,
		RepresentativeSmsCertAt:         user.RepresentativeSmsCertAt,
		Identification1FrontFile:        user.Identification1FrontFile,
		Identification1BackFile:         user.Identification1BackFile,
		Identification2FrontFile:        user.Identification2FrontFile,
		Identification2BackFile:         user.Identification2BackFile,
		CertifiedCopyFiles:              user.CertifiedCopyFiles,
		PoaFiles:                        user.PoaFiles,
		AoiFiles:                        user.AoiFiles,
		ShareholdersFiles:               user.ShareholdersFiles,
		OthersFiles:                     user.OthersFiles,
		CorporateShareholder:            user.CorporateShareholder,
		EquityVaAccount:                 user.EquityVaAccount,
		AntisocialCheck:                 user.AntisocialCheck,
		MyNumber:                        user.MyNumber,
		UserAccount:                     user.UserAccount,
		AntiResearchDate:                user.AntiResearchDate,
		HasExperience:                   user.HasExperience,
		Created:                         user.User.Created,
		Updated:                         user.User.Updated,
	}
}
