package domain

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"time"

	"github.com/gofrs/uuid"
	"github.com/samber/lo"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/csvutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/modelhelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
)

type CorporateProfile struct {
	Id                              int64              `json:"-"`
	UserId                          int64              `json:"user_id" xorm:"index unique"`
	CorpNo                          string             `json:"corp_no" xorm:"not null unique" swagger:"desc(법인 번호)"`
	CorpType                        string             `json:"corp_type" xorm:"not null" swagger:"desc(법인 업종)"`
	EmployeeCount                   string             `json:"employee_count" xorm:"not null" swagger:"desc(사원 수 기획팀 요청으로 string 값 )"`
	CapitalAmount                   string             `json:"capital_amount" xorm:"not null" swagger:"desc(자본금 기획팀 요청으로 string 값 )"`
	Sales                           string             `json:"sales" xorm:"not null" swagger:"desc(매출액 기획팀 요청으로 string 값 )"`
	RepresentativeFamilyName        string             `json:"representative_family_name" xorm:"not null" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_CEO_LASTNAME" swagger:"desc(대표자 명 성)"`
	RepresentativeFirstName         string             `json:"representative_first_name" xorm:"not null" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_CEO_FIRSTNAME" swagger:"desc(대표자 명 이름)"`
	RepresentativeFamilyNameKana    string             `json:"representative_family_name_kana" xorm:"not null" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_CEO_LASTNAME_FURIGANA" swagger:"desc(대표자 명 성 후리가나)"`
	RepresentativeFirstNameKana     string             `json:"representative_first_name_kana" xorm:"not null" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_CEO_FIRSTNAME_FURIGANA" swagger:"desc(대표자 명 이름 후리가나)"`
	RepresentativeMobilePhone       string             `json:"representative_mobile_phone" xorm:"not null" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_CEO_PHONE" swagger:"desc(대표자 핸드폰)"`
	RepresentativeEmail             string             `json:"representative_email" xorm:"not null" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_CEO_EMAIL" swagger:"desc(대표자 이메일)"`
	ManagerFamilyName               *string            `json:"manager_family_name" xorm:"not null" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_INCHARGE_LASTNAME" swagger:"desc(담당자 명 성)"`
	ManagerFirstName                *string            `json:"manager_first_name" xorm:"not null" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_INCHARGE_FIRSTNAME" swagger:"desc(담당자 명 이름)"`
	ManagerFamilyNameKana           *string            `json:"manager_family_name_kana" xorm:"not null" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_INCHARGE_LASTNAME_FURIGANA" swagger:"desc(담당자 명 성 후리가나)"`
	ManagerFirstNameKana            *string            `json:"manager_first_name_kana" xorm:"not null" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_INCHARGE_FIRSTNAME_FURIGANA" swagger:"desc(담당자 명 이름 후리가나)"`
	ManagerEmail                    *string            `json:"manager_email" xorm:"not null" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_INCHARGE_EMAIL" swagger:"desc(담당자 이메일)"`
	ManagerZipCode                  string             `json:"manager_zip_code" xorm:"not null" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_MANAGER_ZIPCODE" swagger:"desc(담당자 우편번호)"`
	ManagerProvince                 string             `json:"manager_province" xorm:"not null" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_MANAGER_PROVINCE" swagger:"desc(담당자 현 이름)"`
	ManagerCity                     string             `json:"manager_city" xorm:"not null" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_MANAGER_CITY" swagger:"desc(담당자 시 도시)"`
	ManagerAddress                  string             `json:"manager_address" xorm:"not null" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_MANAGER_ADDRESS" swagger:"desc(담당자 주소)"`
	ManagerAddress2                 *string            `json:"manager_address2" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_MANAGER_ADDRESS2" swagger:"desc(담당자 주소2)"`
	ManagerMobilePhone              *string            `json:"manager_mobile_phone" xorm:"not null" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_INCHARGE_PHONE" swagger:"desc(담당자 번호)"`
	ManagerBirthday                 jsonutils.JSONTime `json:"manager_birthday" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_MANAGER_BIRTHDAY" xorm:"not null" swagger:"desc(담당자 생년월일)"`
	IsManagerSmsCert                bool               `json:"is_manager_sms_cert" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_MANAGER_CERTIFICATION" swagger:"desc(담당자 SMS 문자인증 여부)"`
	ManagerSmsCertAt                jsonutils.JSONTime `json:"manager_sms_cert_at" swagger:"desc(담당자 SMS 문자인증 시점)"`
	Department                      *string            `json:"department" xorm:"not null" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_INCHARGE_DEPARTMENT" swagger:"desc(부서)"`
	Position                        *string            `json:"position" xorm:"not null" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_INCHARGE_POSITION" swagger:"desc(직책)"`
	CorporatePostalCode             string             `json:"corporate_postal_code" swagger:"desc(법인 담당자 우편 코드)"`
	CorporatePostalCodeApprovalDate jsonutils.JSONTime `json:"corporate_postal_code_approval_date" swagger:"desc(법인 담당자 우편 코드 인증 일자 )"`

	CorporatePostalStatus       bool               `json:"corporate_postal_status" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_MANAGER_POSTAL_CODE_STATUS" swagger:"desc(법인 담당자 우편 코드 발송 상태 발송완료: True 미발송 False)"`
	CorporatePostalCodeSendDate jsonutils.JSONTime `json:"corporate_postal_code_send_date" swagger:"desc(법인 담당자 우편 코드 발송 일자 )"`

	IsRepresentativeSmsCert  bool                    `json:"is_representative_sms_cert" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_CEO_CERTIFICATION" swagger:"desc(대표자 핸드폰 인증 일자 false 미인증)"`
	RepresentativeSmsCertAt  jsonutils.JSONTime      `json:"representative_sms_cert_at" swagger:"enum(true|false)"`
	Identification1FrontFile []*UserFile             `json:"identification1_front_file" xorm:"-" swagger:"desc(본인 확인 서류 1 앞)"`
	Identification1BackFile  []*UserFile             `json:"identification1_back_file" xorm:"-" swagger:"desc(본인 확인 서류 1 뒤)"`
	Identification2FrontFile []*UserFile             `json:"identification2_front_file" xorm:"-" swagger:"desc(본인 확인 서류 2 앞)"`
	Identification2BackFile  []*UserFile             `json:"identification2_back_file" xorm:"-" swagger:"desc(본인 확인 서류 2 뒤)"`
	CertifiedCopyFiles       []*UserFile             `json:"certified_copy_files" xorm:"-" swagger:"desc(법인 등기부 등본)"`
	PoaFiles                 []*UserFile             `json:"poa_files" xorm:"-" swagger:"desc(법인 대표자 위임장)"`
	ShareholdersFiles        []*UserFile             `json:"shareholders_files" xorm:"-" swagger:"desc(주주 명부)"`
	AoiFiles                 []*UserFile             `json:"aoi_files" xorm:"-" swagger:"desc(정관)"`
	OthersFiles              []*UserFile             `json:"others_files" xorm:"-" swagger:"desc(기타 서류)"`
	CorporateShareholder     []*CorporateShareholder `json:"shareholder_list" xorm:"-" swagger:"desc(주주 정보 목록)"`
	Created                  jsonutils.JSONTime      `json:"-" xorm:"created"`
	Updated                  jsonutils.JSONTime      `json:"-" xorm:"updated"`
	Deleted                  jsonutils.JSONTime      `json:"-" xorm:"deleted"`
}

func (c CorporateProfile) TableName() string {
	return "corporate_profile"
}

func (c CorporateProfile) SelectAll() string {
	return c.TableName() + ".*"
}

// CorporateProfile 를 수정할 경우 해당 쿼리를 수정해야 함
const CORPORATE_USER_COLUMN_STRING = "" +
	"corporate_profile.user_id as user_id, corporate_profile.corp_no as corp_no, corporate_profile.corp_type as corp_type," +
	"corporate_profile.employee_count as employee_count,corporate_profile.capital_amount as capital_amount,corporate_profile.sales as sales, corporate_profile.representative_family_name as representative_family_name," +
	"corporate_profile.representative_first_name as representative_first_name,corporate_profile.representative_family_name_kana as representative_family_name_kana,corporate_profile.representative_first_name_kana as representative_first_name_kana," +
	"corporate_profile.representative_mobile_phone as representative_mobile_phone, corporate_profile.representative_email as representative_email,corporate_profile.manager_family_name as manager_family_name, corporate_profile.manager_first_name as manager_first_name," +
	"corporate_profile.manager_family_name_kana as manager_family_name_kana, corporate_profile.manager_first_name_kana as manager_first_name_kana,corporate_profile.manager_email as manager_email,corporate_profile.manager_zip_code as manager_zip_code, " +
	"corporate_profile.manager_province as manager_province, corporate_profile.manager_city as manager_city, corporate_profile.manager_address as manager_address,corporate_profile.manager_address2 as manager_address2," +
	"corporate_profile.manager_mobile_phone as manager_mobile_phone, corporate_profile.manager_birthday as manager_birthday,corporate_profile.department as department, corporate_profile.position as position, corporate_profile.corporate_postal_code as corporate_postal_code," +
	"corporate_profile.is_representative_sms_cert as is_representative_sms_cert, corporate_profile.representative_sms_cert_at as representative_sms_cert_at "

func (c *CorporateUser) MarshalJSON() ([]byte, error) {

	return json.Marshal(map[string]interface{}{
		"id":                                  c.User.Id,
		"uid":                                 c.Uid,
		"user_id":                             c.CorporateProfile.UserId,
		"common_id":                           c.CommonId,
		"user_type":                           c.UserType,
		"name":                                c.Name,
		"corp_no":                             c.CorpNo,
		"corp_type":                           c.CorpType,
		"email":                               c.Email,
		"login_type":                          c.LoginType,
		"investor_type":                       c.User.InvestorType,
		"block_invest":                        c.User.BlockInvest,
		"roles":                               c.User.Roles,
		"sign_up_date":                        c.User.SignUpDate,
		"last_login_date":                     c.User.LastLoginDate,
		"mobile_phone":                        c.User.MobilePhone,
		"zip_code":                            c.ZipCode,  // 우편번호
		"address":                             c.Address,  // 주소
		"address2":                            c.Address2, // 상세주소
		"province":                            c.Province, // 현
		"city":                                c.City,     // 시
		"audit_status":                        c.AuditStatus,
		"investment_purpose":                  c.InvestmentPurpose,
		"is_surplus_fund":                     c.IsSurplusFund,
		"is_newsletter":                       c.IsNewsletter,
		"newsletter_agree_at":                 c.NewsletterAgreeAt,
		"is_sms_cert":                         c.IsSmsCert,
		"sms_cert_at":                         c.SmsCertAt,
		"bank_name":                           c.BankName,
		"branch_name":                         c.BranchName, // 지점명
		"bank_account":                        c.BankAccount,
		"depositor":                           c.Depositor,
		"bank_code":                           c.BankCode,   // 은행 코드
		"branch_code":                         c.BranchCode, // 지점 코드
		"deposit_type":                        c.DepositType,
		"employee_count":                      c.EmployeeCount,
		"capital_amount":                      c.CapitalAmount,
		"sales":                               c.Sales,
		"representative_family_name":          c.RepresentativeFamilyName,
		"representative_first_name":           c.RepresentativeFirstName,
		"representative_family_name_kana":     c.RepresentativeFamilyNameKana,
		"representative_first_name_kana":      c.RepresentativeFirstNameKana,
		"representative_mobile_phone":         c.RepresentativeMobilePhone,
		"is_representative_sms_cert":          c.IsRepresentativeSmsCert,
		"representative_sms_cert_at":          c.RepresentativeSmsCertAt,
		"representative_email":                c.RepresentativeEmail,
		"manager_family_name":                 c.ManagerFamilyName,
		"manager_first_name":                  c.ManagerFirstName,
		"manager_family_name_kana":            c.ManagerFamilyNameKana,
		"manager_first_name_kana":             c.ManagerFirstNameKana,
		"manager_email":                       c.ManagerEmail,
		"manager_mobile_phone":                c.ManagerMobilePhone,
		"manager_zip_code":                    c.ManagerZipCode,
		"manager_province":                    c.ManagerProvince,
		"manager_city":                        c.ManagerCity,
		"manager_address":                     c.ManagerAddress,
		"manager_address2":                    c.ManagerAddress2,
		"manager_birthday":                    c.ManagerBirthday,
		"is_manager_sms_cert":                 c.IsManagerSmsCert,
		"manager_sms_cert_at":                 c.ManagerSmsCertAt,
		"department":                          c.Department,
		"position":                            c.Position,
		"identification1_front_file":          c.Identification1FrontFile,
		"identification1_back_file":           c.Identification1BackFile,
		"identification2_front_file":          c.Identification2FrontFile,
		"identification2_back_file":           c.Identification2BackFile,
		"certified_copy_files":                c.CertifiedCopyFiles,
		"poa_files":                           c.PoaFiles,
		"postal_code":                         c.PostalCode,
		"postal_code_status":                  c.PostalCodeStatus,
		"postal_code_send_date":               c.PostalCodeSendDate,
		"shareholder_list":                    c.CorporateShareholder,
		"sign_up_approval_date":               c.SignUpApprovalDate,
		"code_approval_date":                  c.CodeApprovalDate,
		"invalid_reason":                      c.InvalidReason,
		"corporate_postal_code":               c.CorporatePostalCode,
		"corporate_postal_status":             c.CorporatePostalStatus,
		"corporate_postal_code_send_date":     c.CorporatePostalCodeSendDate,
		"corporate_postal_code_approval_date": c.CorporatePostalCodeApprovalDate,
		"equity_va_account":                   c.EquityVaAccount,
		"antisocial_check":                    c.AntisocialCheck,
		"my_number":                           c.MyNumber,
		"anti_research_date":                  c.AntiResearchDate,
		"updated":                             c.User.Updated,
		"created":                             c.User.Created,
	})

}

const CORPORATE_SIGNUP_EXPIRE_TIME = time.Minute * 20

// 메일로 인증 // 회원가입 요청 메일 // 회원가입 프로세스 진행 개인 회원 가입
type CorporateSignUpRequest struct { // 메일로 전송된 URL 정보
	Id        int64              `json:"id"`
	RequestId string             `json:"request_id" xorm:"not null UUID"`
	Email     string             `json:"email" swagger:"desc(수신 메일)"`
	Url       string             `json:"url" swagger:"desc(회원 가입 주소)"`
	SendCount int                `json:"send_count" swagger:"desc(메일로 재전송된 횟수)"`
	Expire    jsonutils.JSONTime `json:"expire" swagger:"desc(만료 일자)"`
	Created   jsonutils.JSONTime `json:"created" xorm:"created"`
}
type CorporateUser struct {
	User             `xorm:"extends"`
	CorporateProfile `xorm:"extends"`
	AntisocialCheck  *UserAntisocialCheck `json:"antisocial_check" xorm:"-" swagger:"desc(반사회 세력 체크)"`
	UserAccount      *UserAccount         `json:"user_account" xorm:"-" swagger:"desc(예치금 계좌)"`
	EquityVaAccount  *UserVirtualAccount  `json:"equity_va_account" xorm:"-" swagger:"desc(사업형 가상 계좌)"`
}

// 법인회원 가입절차 확인 후 모델 생성
func (s *CorporateSignUpRequest) Create(ctx context.Context) (int64, error) {
	return factory.DB(ctx).Insert(s)
}

func (s CorporateSignUpRequest) Update(ctx context.Context) (int64, error) {
	return factory.DB(ctx).Where("email = ?", s.Email).Update(s)
}

func (CorporateSignUpRequest) CreateSignUpRequest(ctx context.Context, email string) (*CorporateSignUpRequest, error) {

	id, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}
	s := &CorporateSignUpRequest{
		RequestId: id.String(),
		Email:     email,
		Url:       fmt.Sprintf("%s/sign-up/corporate/%s", serviceutils.GetUserServiceDomain(), id.String()),
		Expire:    jsonutils.JSONTime(time.Now().Add(CORPORATE_SIGNUP_EXPIRE_TIME)),
	}
	_, err = s.Create(ctx)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (u *CorporateUser) RepresentativeVerificationConfirm(ctx context.Context) error {
	u.IsRepresentativeSmsCert = true
	u.RepresentativeSmsCertAt = jsonutils.JSONTime(time.Now())
	return u.Update(ctx)
}

func (u *CorporateUser) ManagerVerificationConfirm(ctx context.Context) error {
	u.IsManagerSmsCert = true
	u.ManagerSmsCertAt = jsonutils.JSONTime(time.Now())
	return u.Update(ctx)
}

// 법인 유저 정보 저장
func (c *CorporateProfile) Create(ctx context.Context) (cnt int64, err error) {
	return factory.DB(ctx).Insert(c)
}

// 법인 사용자 정보 조회
func (CorporateUser) GetAll(ctx context.Context, id, name, email, investorType, auditStatus *string, sortBy, order string, offset, limit int, startDate, endDate string, profileCond ProfileSearchCondition) (int64, []CorporateUser, error) {
	query := factory.DB(ctx).Table("user").
		Select("user.*,"+CORPORATE_USER_COLUMN_STRING).
		Join("INNER", "corporate_profile", "corporate_profile.user_id = user.id").
		Where("user.user_type = ?", CORPORATE)

	if err := modelhelper.SetSortOrder(query, sortBy, order); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SearchEqQuery(query, "user", "id", id); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SearchLikeQuery(query, "user", "name", name); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SearchLikeQuery(query, "user", "email", email); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SearchEqQuery(query, "user", "investor_type", investorType); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SearchEqQuery(query, "user", "audit_status", auditStatus); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SetBetweenDate(query, "user.sign_up_date", startDate, endDate); err != nil {
		factory.Logger(ctx).Error(err)
	}

	// フリガナ・生年月日(MIR-534)。DownloadCsv も同じ関数で絞る。
	if err := applyCorporateProfileSearch(query, profileCond); err != nil {
		return 0, nil, err
	}

	var userList []CorporateUser
	var totalCount int64
	totalCount, err := query.Limit(limit, offset).FindAndCount(&userList)
	if err != nil {
		return 0, nil, err
	}
	if userList == nil {
		userList = make([]CorporateUser, 0)
	}

	return totalCount, userList, nil
}

func (CorporateUser) GetDashboardList(ctx context.Context, offset, limit int) (int64, []CorporateUser, error) {
	query := factory.DB(ctx).Table("user").Select("user.*, corporate_profile.*").Join("INNER", "corporate_profile", "corporate_profile.user_id = user.id").Desc("user.sign_up_date")

	userList := make([]CorporateUser, 0, 0)
	var totalCount int64
	totalCount, err := query.Where("user.audit_status=?", WAIT_AUDIT).Desc("user.sign_up_date").Limit(limit, offset).FindAndCount(&userList)
	if err != nil {
		return 0, nil, err
	}

	return totalCount, userList, nil
}

func (CorporateUser) GetAntisocialCheckUsers(ctx context.Context, offset, limit int) (int64, []CorporateUser, error) {
	query := factory.DB(ctx).Table("user").Select("user.*, corporate_profile.*").Join("INNER", "corporate_profile", "corporate_profile.user_id = user.id").Desc("user.sign_up_date")

	userList := make([]CorporateUser, 0, 0)
	var totalCount int64
	totalCount, err := query.Where("user.audit_status=?", CHECK_ANTISOCIAL_AUDIT).Desc("user.sign_up_date").Limit(limit, offset).FindAndCount(&userList)
	if err != nil {
		return 0, nil, err
	}

	return totalCount, userList, nil
}

// 법인 사용자 개별 조회
func (CorporateUser) GetById(ctx context.Context, id int64) (*CorporateUser, error) {
	var user CorporateUser
	var err error

	query := factory.DB(ctx).Table("user").
		Select("user.*, corporate_profile.*").
		Join("INNER", "corporate_profile", "corporate_profile.user_id = user.id").
		Where("user.id = ?", id)

	if has, err := query.Get(&user); err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}

	user.Roles = []string{string(ROLE_USER)}

	user.CorporateShareholder, err = CorporateShareholder{}.FindByUserId(ctx, 100, user.User.Id)
	if err != nil {
		return nil, err
	}

	if err := user.GetIdentificationFiles(ctx); err != nil {
		return nil, err
	}

	user.EquityVaAccount, err = UserVirtualAccount{}.GetVaByUserId(ctx, user.User.Id, EQUITY)
	if err != nil {
		return nil, err
	}

	user.AntisocialCheck, err = UserAntisocialCheck{}.GetByCheckUserId(ctx, user.User.Id)
	if err != nil {
		return nil, err
	}

	user.UserAccount, err = UserAccount{}.GetByUserId(ctx, user.User.Id)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// 법인 회원 본인 확인 서류 조회
// deprecated user_file_svc의 GetCorporateUserIdentificationFiles 사용
func (cu *CorporateUser) GetIdentificationFiles(ctx context.Context) error {
	var err error

	cu.Identification1FrontFile, err = UserFile{}.FindByUserIdAndFileType(ctx, 10, cu.User.Id, IDENTIFICATION1_FRONT)
	if err != nil {
		return err
	}
	cu.Identification1BackFile, err = UserFile{}.FindByUserIdAndFileType(ctx, 10, cu.User.Id, IDENTIFICATION1_BACK)
	if err != nil {
		return err
	}
	cu.Identification2FrontFile, err = UserFile{}.FindByUserIdAndFileType(ctx, 10, cu.User.Id, IDENTIFICATION2_FRONT)
	if err != nil {
		return err
	}
	cu.Identification2BackFile, err = UserFile{}.FindByUserIdAndFileType(ctx, 10, cu.User.Id, IDENTIFICATION2_BACK)
	if err != nil {
		return err
	}
	cu.CertifiedCopyFiles, err = UserFile{}.FindByUserIdAndFileType(ctx, 10, cu.User.Id, CERTIFIED_COPY)
	if err != nil {
		return err
	}
	cu.PoaFiles, err = UserFile{}.FindByUserIdAndFileType(ctx, 10, cu.User.Id, POA)
	if err != nil {
		return err
	}
	cu.AoiFiles, err = UserFile{}.FindByUserIdAndFileType(ctx, 10, cu.User.Id, AOI)
	if err != nil {
		return err
	}
	cu.ShareholdersFiles, err = UserFile{}.FindByUserIdAndFileType(ctx, 10, cu.User.Id, SHAREHOLDERS)
	if err != nil {
		return err
	}
	cu.OthersFiles, err = UserFile{}.FindByUserIdAndFileType(ctx, 10, cu.User.Id, USER_FILE_OTHERS)
	if err != nil {
		return err
	}

	return nil
}

// 법인 사용자 생성
func (c CorporateUser) Create(ctx context.Context) (cnt int64, err error) {
	if cnt, err = c.User.Create(ctx); err != nil {
		return 0, err
	}
	c.UserId = c.User.Id
	if _, err = c.CorporateProfile.Create(ctx); err != nil {
		return 0, err
	}
	return cnt, err
}

// 법인 사용자 수정
func (u *CorporateUser) Update(ctx context.Context) (err error) {
	// MIR-1252: code_approval_dateはゼロ値でも常に更新対象に含める(MustCols)。
	// xormはデフォルトでゼロ値フィールドを更新対象から除外するため、これが無いと
	// admin側で住所確認中へ差し戻す際にCodeApprovalDateをゼロへ戻しても書き込まれず、
	// 前回確認済みの承認日が残り続けてしまう(片方の新しいコードだけの確認で
	// CheckAllPostalCodeの完了判定を満たしてしまう)。この関数の呼び出し元は
	// ほとんどがGetById等で読み込んだ完全な構造体を使っているため、通常の更新では
	// 既存の値をそのまま書き戻すだけになり、影響は無い。例外はapps/admin-api/
	// controllers/corporate.goのUpdate(法人会員情報編集)で、ToModel()が組み立てる
	// 新規構造体を使うためCodeApprovalDateがゼロ値になる。ここはcarryOverPostalCodeApprovalDates
	// が元の値を引き継ぐことで安全性を保っている。
	// UseBool()には引数を渡さないこと。UseBool(列名...)は「その列だけ」を
	// 強制するMustColsの別名になり、引数無しのUseBool()が持つ「構造体の
	// bool列を全て更新対象に含める」効果が失われる(xormのソースで確認済み)。
	// 列を追加したい場合はMustColsを別途チェーンする。
	q := factory.DB(ctx).UseBool().MustCols("code_approval_date")
	if _, err := q.ID(u.User.Id).Update(&u.User); err != nil {
		return err
	}

	q2 := factory.DB(ctx).UseBool("block_invest", "is_representative_sms_cert", "corporate_postal_status", "corporate_postal_code_approval_date", "corporate_postal_code_send_date", "manager_address")
	if _, err := q2.ID(u.CorporateProfile.Id).Update(&u.CorporateProfile); err != nil {
		return err
	}
	return nil
}

// 법인 사용자 삭제
func (CorporateUser) Delete(ctx context.Context, id int64) (err error) {
	if _, err = factory.DB(ctx).Where("corporate_profile.user_id = ?", id).Delete(&CorporateProfile{}); err != nil {
		return err
	}
	if _, err = factory.DB(ctx).ID(id).Delete(&User{}); err != nil {
		return err
	}
	return
}

func (CorporateUser) PrepareAndReturnCsv(items []CorporateUser, investGrade, basicInfo, withdrawal, loanDeposit, equityDeposit, newsletterAgreement bool) [][]string {

	type AntisocialWaitCorporate struct {
		CorpName           string `json:"corp_name" i18n:"model.corporate.corp_name"`                     // 법인명
		RepresentativeName string `json:"representative_name" i18n:"model.corporate.representative_name"` // 대표자 이름
		Email              string `json:"email" i18n:"model.user.email_id"`                               // メールアドレス(ID)
		ZipCode            string `json:"zip_code" i18n:"model.user.zip_code"`
		Address            string `json:"address" i18n:"model.corporate.address"` // 법인 주소

	}

	var corporates []AntisocialWaitCorporate
	for _, item := range items {
		c := AntisocialWaitCorporate{
			CorpName:           item.Name,
			RepresentativeName: item.RepresentativeFamilyName + "　" + item.RepresentativeFirstName,
			Email:              item.Email,
			ZipCode:            item.ZipCode,
			Address:            item.Province + " " + item.City + " " + item.Address + " " + lo.FromPtr(item.Address2),
		}
		corporates = append(corporates, c)
	}
	cols := factory.ReflectJsonFields(reflect.ValueOf(&AntisocialWaitCorporate{}).Elem())
	if len(cols) <= 0 {
		return nil
	}

	// 투자 등급
	type InvestGrade struct {
		InvestorType string `json:"investor_type" i18n:"model.user.investor_type"` // 투자 등급
	}

	var igs []InvestGrade
	if investGrade {
		cols = append(cols, factory.ReflectJsonFields(reflect.ValueOf(&InvestGrade{}).Elem())...)

		for _, item := range items {
			ig := InvestGrade{
				InvestorType: i18nutils.DefaultTranslate("commoncode.INVESTOR_TYPE." + string(item.InvestorType)),
			}
			igs = append(igs, ig)
		}
	}

	// 기본 정보
	type BasicInfo struct {
		CorpNo                      string `json:"corp_no" i18n:"model.corporate.corp_number"`                                             // 법인번호
		CorpType                    string `json:"corp_type" i18n:"model.corporate.corp_type"`                                             // 사업내용
		PhoneNumber                 string `json:"phone_number" i18n:"model.user.mobile_phone"`                                            // 전화번호
		SignUpRequest               string `json:"sign_up_request" i18n:"model.user.sign_up_request"`                                      // 회원가입 신청일
		SignUpResponse              string `json:"sign_up_response" i18n:"model.user.sign_up_response"`                                    // 회원가입 승인일
		MarketingAgree              string `json:"marketing_agree" i18n:"model.user.marketing_agree"`                                      // 마케팅 동의
		LastConnection              string `json:"last_connection" i18n:"model.user.last_login_date"`                                      // 최근 접속일
		PostalCode                  string `json:"postal_code" i18n:"model.corporate.postal_code"`                                         // 소재지 확인코드
		CodeCertified               string `json:"code_certified" i18n:"model.corporate.code_certified"`                                   // 소재지 확인코드 인증일
		PostalCodeState             string `json:"postal_code_state" i18n:"model.user.postal_code_state"`                                  // 우편물 발송상태
		PostalCodeSendDate          string `json:"postal_code_send_date" i18n:"model.user.postal_code_send_date"`                          // 우편물 발송일
		CorporatePostalCode         string `json:"corporate_postal_code" i18n:"model.corporate.corporate_postal_code"`                     // 담당자 확인 코드
		RepresentativeSmsCertAt     string `json:"representative_sms_cert_at" i18n:"model.corporate.corporate_postal_code_response"`       // 담당자 확인코드 인증일
		CorporatePostalState        string `json:"corporate_postal_status" i18n:"model.corporate.corporate_postal_code_state"`             // 담당자 우편물 발송 상태
		CorporatePostalCodeSendDate string `json:"corporate_postal_code_send_date" i18n:"model.corporate.corporate_postal_code_send_date"` // 담당자 우편물 발송 일자
		Id                          string `json:"id" i18n:"model.corporate.uid"`                                                          // 유니크 ID

	}
	var bis []BasicInfo
	if basicInfo {
		cols = append(cols, factory.ReflectJsonFields(reflect.ValueOf(&BasicInfo{}).Elem())...)

		for _, item := range items {
			bi := BasicInfo{
				CorpNo:              item.CorpNo,
				CorpType:            item.CorpType,
				PhoneNumber:         item.MobilePhone,
				PostalCode:          item.PostalCode,
				CorporatePostalCode: item.CorporatePostalCode,
				Id:                  strconv.Itoa(int(item.User.Id)),
			}

			if item.IsNewsletter {
				bi.MarketingAgree = i18nutils.DefaultTranslate("model.user.agree")
			} else {
				bi.MarketingAgree = i18nutils.DefaultTranslate("model.user.disagree")
			}

			if item.NewsletterAgreeAt.Unix() > 0 {
				bi.MarketingAgree += " : " + item.NewsletterAgreeAt.ToString()
			}

			if item.PostalCodeStatus {
				bi.PostalCodeState = i18nutils.DefaultTranslate("commoncode.POSTAL_CODE_STATE.POSTAL_CODE_STATE_SENT")
			} else {
				bi.PostalCodeState = i18nutils.DefaultTranslate("commoncode.POSTAL_CODE_STATE.POSTAL_CODE_STATE_NOTSENT")
			}

			if item.CodeApprovalDate.Unix() > 0 {
				bi.CodeCertified = item.CodeApprovalDate.ToString()
			}

			if item.PostalCodeSendDate.Unix() > 0 {
				bi.PostalCodeSendDate = item.PostalCodeSendDate.ToString()
			}

			if item.RepresentativeSmsCertAt.Unix() > 0 {
				bi.RepresentativeSmsCertAt = item.RepresentativeSmsCertAt.ToString()
			}

			if item.CorporatePostalStatus {
				bi.CorporatePostalState = i18nutils.DefaultTranslate("commoncode.POSTAL_CODE_STATE.POSTAL_CODE_STATE_SENT")
			} else {
				bi.CorporatePostalState = i18nutils.DefaultTranslate("commoncode.POSTAL_CODE_STATE.POSTAL_CODE_STATE_NOTSENT")
			}

			if item.CorporatePostalCodeSendDate.Unix() > 0 {
				bi.CorporatePostalCodeSendDate = item.CorporatePostalCodeSendDate.ToString()
			}

			if item.SignUpDate.Unix() > 0 {
				bi.SignUpRequest = item.SignUpDate.ToString()
			}

			if item.SignUpApprovalDate.Unix() > 0 {
				bi.SignUpResponse = item.SignUpApprovalDate.ToString()
			}

			if item.LastLoginDate.Unix() > 0 {
				bi.LastConnection = item.LastLoginDate.ToString()
			}

			bis = append(bis, bi)
		}
	}

	// 출금계좌
	type WithdrawalInfo struct {
		BankName    string `json:"bank_name" i18n:"model.user.withdrawal_bank_name"`       // 출금계좌 금융기관명
		BranchName  string `json:"branch_name" i18n:"model.user.withdrawal_branch_name"`   // 출금계좌 지점명
		DepositType string `json:"deposit_type" i18n:"model.user.withdrawal_deposit_type"` // 출금계좌 예금종류
		BankAccount string `json:"bank_account" i18n:"model.user.withdrawal_bank_account"` // 출금계좌 계좌번호
		Depositor   string `json:"depositor" i18n:"model.user.withdrawal_depositor"`       // 출금계좌 예금주
	}
	var wis []WithdrawalInfo
	if withdrawal {
		cols = append(cols, factory.ReflectJsonFields(reflect.ValueOf(&WithdrawalInfo{}).Elem())...)

		for _, item := range items {
			wi := WithdrawalInfo{
				BankName:    item.BankName,
				BranchName:  item.BranchName,
				DepositType: i18nutils.DefaultTranslate("commoncode.DEPOSIT_TYPE_CODE." + string(item.DepositType)),
				BankAccount: item.BankAccount,
				Depositor:   item.Depositor,
			}

			wis = append(wis, wi)
		}
	}

	// 사업형 가상계좌
	type EquityDepositInfo struct {
		BankName    string `json:"bank_name" i18n:"model.user.equity_bank_name"`       // 사업형 입금계좌 금융기관명
		BranchName  string `json:"branch_name" i18n:"model.user.equity_branch_name"`   // 사업형 입금계좌 지점명
		Depositor   string `json:"depositor" i18n:"model.user.equity_depositor"`       // 사업형 입금계좌 예금주
		BankAccount string `json:"bank_account" i18n:"model.user.equity_bank_account"` // 사업형 입금계좌 계좌번호
	}
	var edis []EquityDepositInfo
	if equityDeposit {
		cols = append(cols, factory.ReflectJsonFields(reflect.ValueOf(&EquityDepositInfo{}).Elem())...)

		for _, item := range items {
			edi := EquityDepositInfo{
				BankName:    item.EquityVaAccount.VaBankName,
				BranchName:  item.EquityVaAccount.VaBranchName,
				Depositor:   item.EquityVaAccount.VaHolderNameKana,
				BankAccount: item.EquityVaAccount.VaAccountNumber,
			}

			edis = append(edis, edi)
		}
	}

	// 뉴스레터 수신동의일
	type NewsletterAgreement struct {
		AgreementDate string `json:"agreement_date" i18n:"model.user.agreement_date"` // 뉴스레터 수신동의일
	}
	var nas []NewsletterAgreement
	if newsletterAgreement {
		cols = append(cols, factory.ReflectJsonFields(reflect.ValueOf(&NewsletterAgreement{}).Elem())...)

		for _, item := range items {
			var na NewsletterAgreement
			if item.NewsletterAgreeAt.Unix() > 0 && item.IsNewsletter {
				na.AgreementDate = item.NewsletterAgreeAt.ToString()
			}

			nas = append(nas, na)
		}
	}

	records := [][]string{}
	records = append(records, cols)

	for index, _ := range corporates {

		var datas []string
		datas = append(datas, factory.DataWriteToCsv(reflect.ValueOf(&corporates[index]).Elem())...)

		if len(igs) > 0 {
			datas = append(datas, factory.DataWriteToCsv(reflect.ValueOf(&igs[index]).Elem())...)
		}

		if len(bis) > 0 {
			datas = append(datas, factory.DataWriteToCsv(reflect.ValueOf(&bis[index]).Elem())...)
		}

		if len(wis) > 0 {
			datas = append(datas, factory.DataWriteToCsv(reflect.ValueOf(&wis[index]).Elem())...)
		}

		if len(edis) > 0 {
			datas = append(datas, factory.DataWriteToCsv(reflect.ValueOf(&edis[index]).Elem())...)
		}

		if len(nas) > 0 {
			datas = append(datas, factory.DataWriteToCsv(reflect.ValueOf(&nas[index]).Elem())...)
		}

		records = append(records, datas)
	}

	return records
}

var antisocialWaitCorporateCSVColumns = []string{
	"model.corporate.corp_name",            // 기업명 企業名
	"model.corporate.representative_name",  // 대표이사명 代表取締役氏名
	"model.corporate.manager_name",         // 거래담당자명 取引担当者氏名
	"model.corporate.manager_birthday",     // 거래담당자 생년월일 取引担当者生年月日
	"model.corporate.shareholder_name",     // 실질지배자명 実質的支配者氏名
	"model.corporate.shareholder_birthday", // 실질지배자 생년월일 実質的支配者生年月日
}

// 파일명: energyShares_anti_member_list.csv
// 지라: https://illuminarean.atlassian.net/browse/ESJP-488
func (CorporateUser) AntisocialWaitAndReturnCsv(items []CorporateUser) [][]string {

	// 컬럼 설정
	var csvHeaders []string
	for _, v := range antisocialWaitCorporateCSVColumns {
		csvHeaders = append(csvHeaders, i18nutils.DefaultTranslate(v))
	}
	records := [][]string{csvHeaders}

	// 데이터 설정
	for _, item := range items {
		for _, shareholder := range item.CorporateShareholder {
			record := []string{
				item.Name,
				item.RepresentativeFamilyName + " " + item.RepresentativeFirstName,
				*item.ManagerFamilyName + " " + *item.ManagerFirstName,
				item.ManagerBirthday.ToFormat(time.DateOnly),
				shareholder.FamilyName + " " + shareholder.FirstName,
				shareholder.Birthday.ToFormat(time.DateOnly),
			}
			records = append(records, record)

		}
	}

	return records
}

func (CorporateUser) DownloadCsv(ctx context.Context, w http.ResponseWriter, id, name, email, investorType, auditStatus *string, sortby, order string, offset, limit int, startDate, endDate string, profileCond ProfileSearchCondition, investGrade, basicInfo, withdrawal, loanDeposit, equityDeposit, newsletterAgreement bool) error {
	query := factory.DB(ctx).Table("user").Select("user.*,"+CORPORATE_USER_COLUMN_STRING).Join("INNER", "corporate_profile", "corporate_profile.user_id = user.id").Desc("user.created")

	if err := modelhelper.SetSortOrder(query, sortby, order); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SearchEqQuery(query, "user", "id", id); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SearchLikeQuery(query, "user", "name", name); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SearchLikeQuery(query, "user", "email", email); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SearchEqQuery(query, "user", "investor_type", investorType); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SearchEqQuery(query, "user", "audit_status", auditStatus); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SetBetweenDate(query, "user.sign_up_date", startDate, endDate); err != nil {
		factory.Logger(ctx).Error(err)
	}

	// 一覧(GetAll)と同じ関数で絞り、一覧と CSV の結果をそろえる(MIR-534)。
	if err := applyCorporateProfileSearch(query, profileCond); err != nil {
		return err
	}

	var userList []CorporateUser
	_, err := query.FindAndCount(&userList)
	if err != nil {
		return nil
	}

	for i, _ := range userList {
		eva, err := UserVirtualAccount{}.GetVaByUserId(ctx, userList[i].User.Id, EQUITY)
		if err != nil {
			return nil
		} else if eva == nil {
			eva = &UserVirtualAccount{}
		}
		userList[i].EquityVaAccount = eva

	}

	records := CorporateUser{}.PrepareAndReturnCsv(userList, investGrade, basicInfo, withdrawal, loanDeposit, equityDeposit, newsletterAgreement)
	csvutils.WriteHeader(w, "energyshares_cor_member_list.csv")

	return csv.NewWriter(w).WriteAll(records)
}

// 요구사항
// - 파일명 :
// energyShares_anti_member_list.csv
//
// ** 법인 회원 파일에 나오는 항목 :
// 기업명, 대표자 이름, 소재지
func (CorporateUser) AntiSocialWaitUserList(ctx context.Context, w http.ResponseWriter, searchkeys, searchvalues, sortby, order string, offset, limit int, startDate, endDate string, processId int64) error {
	// user, corporate_profile 테이블 조인하여 정보를 가져옴
	query := factory.DB(ctx).Table("user").Select("user.*,"+CORPORATE_USER_COLUMN_STRING).
		Join("INNER", "corporate_profile", "corporate_profile.user_id = user.id").
		Desc("user.created")

	if err := modelhelper.SetSortOrder(query, sortby, order); err != nil {
		factory.Logger(ctx).Error(err)
	}
	// MIR-1232: 管理画面はこのAPIへsearchKeysを一度も送っていない(確認済み)ため、
	// 許可する列名を空にしている。エラーは握りつぶさず呼び出し元(コントローラー)まで
	// 返し、フィルタが無視されたまま全件が返ることを防ぐ。
	if err := modelhelper.SearchQuery(query, "", searchkeys, searchvalues, map[string]struct{}{}); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := modelhelper.SetBetweenDate(query, "user.sign_up_date", startDate, endDate); err != nil {
		factory.Logger(ctx).Error(err)
	}

	var userList []CorporateUser
	_, err := query.Where("audit_status = ?", WAIT_ANTISOCIAL_AUDIT).Limit(limit, offset).FindAndCount(&userList)
	if err != nil {
		return err
	}

	// corporate_shareholder 데이터를 추가로 가져옴
	for i, user := range userList {
		var shareholders []*CorporateShareholder
		err = factory.DB(ctx).Table("corporate_shareholder").
			Where("corporate_shareholder.user_id = ?", user.User.Id).Find(&shareholders)
		if err != nil {
			return err
		}
		userList[i].CorporateShareholder = shareholders
	}

	records := CorporateUser{}.AntisocialWaitAndReturnCsv(userList)

	for _, item := range userList {
		// TODO-hj
		// before, err := userInfoHistoryService.GetDataForCorporateUser(ctx, item.User.Id)
		// if err != nil {
		//	return err
		// }

		item.AuditStatus = CHECK_ANTISOCIAL_AUDIT
		item.AntiResearchDate = jsonutils.JSONTime(time.Now())
		if item.Update(ctx) != nil {
			return err
		}

		// after, err := userInfoHistoryService.GetDataForCorporateUser(ctx, item.User.Id)
		// if err != nil {
		//	return err
		// }
		//
		// _, err = UserInfoHistory{}.Create(ctx, *before, *after, processId)
		// if err != nil {
		//	return err
		// }
	}

	err = query.Commit()
	if err != nil {
		return err
	}

	csvutils.WriteHeader(w, "energyShares_anti_member_list.csv")

	return csv.NewWriter(w).WriteAll(records)
}

func (c *CorporateUser) CheckAllPostalCode(ctx context.Context, postalCode, managerPostalCode string, corpType *string) error {
	if corpType == nil {
		if c.PostalCode != postalCode {
			return factory.ErrorCodeMismatch.New(nil)
		}

		if c.CorporatePostalCode != managerPostalCode {
			return factory.ErrorCorporateCodeMismatch.New(nil)
		}
		c.PostalCodeStatus = true
		c.CorporatePostalStatus = true
		c.CodeApprovalDate = jsonutils.JSONTime(time.Now())
		c.CorporatePostalCodeApprovalDate = jsonutils.JSONTime(time.Now())

	} else {
		switch CORP_TYPE(*corpType) {
		case CORP:
			if c.PostalCode != postalCode {
				return factory.ErrorCodeMismatch.New(nil)
			}
			c.PostalCodeStatus = true
			c.CodeApprovalDate = jsonutils.JSONTime(time.Now())
			break
		case MANAGER:
			if c.CorporatePostalCode != managerPostalCode {
				return factory.ErrorCorporateCodeMismatch.New(nil)
			}
			c.CorporatePostalStatus = true
			c.CorporatePostalCodeApprovalDate = jsonutils.JSONTime(time.Now())
			break
		default:
			return factory.ErrorParameter.New(nil)
		}
	}

	// MIR-1252: 本登録は法人所在地・担当者の両方の郵便コード確認が必要。
	// corpTypeを指定した個別呼び出し(旧エンドポイントをAPI越しに直接叩ける経路)では、
	// 片方だけの確認でもここが無条件で実行されており、届いたハガキ1通だけで
	// 本登録が完了・投資可能になっていた。
	//
	// 判定にはPostalCodeStatus/CorporatePostalStatusを使わない。この2つは
	// 「郵便コードを確認できた」ではなく「ハガキを発送した」を表すフラグで
	// (apps/jobs-scheduler/jobs/user_address_check_jobs.goが発送直後に両方を
	// trueにする)、利用者がハガキを受け取った時点で既に両方trueになっている。
	// これをそのまま使うと、片方の正しいコードとデタラメなもう片方のコードを
	// 送るだけで本登録が完了してしまい、修正になっていない。実際に確認できた
	// ことを表すのはCodeApprovalDate/CorporatePostalCodeApprovalDateで、両方を
	// 一括で検証してから設定するapps/user-api/usecase/corporate_user/
	// corporate_user_ucase.goのVerifyPostalCode(v1)を除けば、片方だけの確認を
	// 許すこのCheckAllPostalCodeでしか設定されない。両方セットされて初めて
	// 完了させる。
	if !c.CodeApprovalDate.IsZero() && !c.CorporatePostalCodeApprovalDate.IsZero() {
		c.AuditStatus = COMPLETED_AUDIT
		c.InvestorType = CORPORATE_INVESTOR
	}

	if err := c.Update(ctx); err != nil {
		return err
	}

	return nil

}
