package domain

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/csv"
	"fmt"
	"math"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/samber/lo"

	"github.com/pkg/errors"

	"github.com/gofrs/uuid"
	"golang.org/x/crypto/pbkdf2"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/csvutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/modelhelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/commoncode"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/tool"
)

const (
	MAX_SENT_SIGN_UP_REQUEST_MAIL        = 1
	MAX_SENT_EMAIL_CODE_REQUEST          = 1
	MAX_SENT_PASSWORD_RESET_REQUEST_MAIL = 1
)

// 투자자 정보 : 개인투자자, 법인 투자자,
// 개인정보 사용동의 관리용 DB 구성

type USER_TYPE string

const (
	USER_TYPE_CODE           = "USER_TYPE"
	PRIVATE        USER_TYPE = "PRIVATE"
	CORPORATE      USER_TYPE = "CORPORATE"
	ADMIN          USER_TYPE = "ADMIN"
	SYSTEM         USER_TYPE = "SYSTEM"
)

const MIN_INVESTOR_AGE_INCLUSIVE = 18
const MAX_INVESTOR_AGE_EXCLUSIVE = 80

func (User) GetUserTypeCodeGroup(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(PRIVATE), Name: i18nutils.Translate("commoncode.USER_TYPE.PRIVATE", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(CORPORATE), Name: i18nutils.Translate("commoncode.USER_TYPE.CORPORATE", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(ADMIN), Name: i18nutils.Translate("commoncode.USER_TYPE.ADMIN", loc), Order: len(codes) + 1})

	return commoncode.CodeGroup{
		GroupCode: USER_TYPE_CODE,
		Codes:     codes,
	}
}

type USER_FILE_TYPE string

const (
	IDENTIFICATION1_FRONT USER_FILE_TYPE = "IDENTIFICATION1_FRONT" // 본인 확인 서류 1 앞
	IDENTIFICATION1_BACK  USER_FILE_TYPE = "IDENTIFICATION1_BACK"  // 본인 확인 서류 1 뒤
	IDENTIFICATION2_FRONT USER_FILE_TYPE = "IDENTIFICATION2_FRONT" // 본인 확인 서류 2 앞
	IDENTIFICATION2_BACK  USER_FILE_TYPE = "IDENTIFICATION2_BACK"  // 본인 확인 서류 2 뒤
	CERTIFIED_COPY        USER_FILE_TYPE = "CERTIFIED_COPY"        // 이력사항 전부 확인 증명서
	MY_NUMBER_FRONT       USER_FILE_TYPE = "MY_NUMBER_FRONT"
	MY_NUMBER_BACK        USER_FILE_TYPE = "MY_NUMBER_BACK"
	EKYC_IDENTIFICATION   USER_FILE_TYPE = "EKYC_IDENTIFICATION"
	POA                   USER_FILE_TYPE = "POA"          // 위임장 (Power of Attorney)
	AOI                   USER_FILE_TYPE = "AOI"          // 정관 (Articles Of Incorporation)
	SHAREHOLDERS          USER_FILE_TYPE = "SHAREHOLDERS" // 주주명부
	USER_FILE_OTHERS      USER_FILE_TYPE = "OTHERS"       // 기타
)

type LOGIN_TYPE string

const (
	LOGIN_TYPE_CODE            = "LOGIN_TYPE"
	MAIL            LOGIN_TYPE = "MAIL"
)

func (User) GetLoginTypeCodeGroup(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(MAIL), Name: i18nutils.Translate("commoncode.LOGIN_TYPE.MAIL", loc), Order: len(codes) + 1})

	return commoncode.CodeGroup{
		GroupCode: LOGIN_TYPE_CODE,
		Codes:     codes,
	}
}

type INVESTOR_TYPE string

const (
	INVESTOR_TYPE_CODE                   = "INVESTOR_TYPE"
	PRIVATE_INVESTOR       INVESTOR_TYPE = "PRIVATE_INVESTOR"
	PRIVATE_PRO_INVESTOR   INVESTOR_TYPE = "PRIVATE_PRO_INVESTOR"
	CORPORATE_INVESTOR     INVESTOR_TYPE = "CORPORATE_INVESTOR"
	CORPORATE_PRO_INVESTOR INVESTOR_TYPE = "CORPORATE_PRO_INVESTOR"
)

func (User) GetInvestorTypeCodeGroup(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0)
	codes = append(codes, commoncode.Code{Code: string(PRIVATE_INVESTOR), Name: i18nutils.Translate("commoncode.INVESTOR_TYPE.PRIVATE_INVESTOR", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PRIVATE_PRO_INVESTOR), Name: i18nutils.Translate("commoncode.INVESTOR_TYPE.PRIVATE_PRO_INVESTOR", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(CORPORATE_INVESTOR), Name: i18nutils.Translate("commoncode.INVESTOR_TYPE.CORPORATE_INVESTOR", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(CORPORATE_PRO_INVESTOR), Name: i18nutils.Translate("commoncode.INVESTOR_TYPE.CORPORATE_PRO_INVESTOR", loc), Order: len(codes) + 1})

	return commoncode.CodeGroup{
		GroupCode: INVESTOR_TYPE_CODE,
		Codes:     codes,
	}
}

func (User) GetPrivateInvestorTypeCodeGroup(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0)
	codes = append(codes, commoncode.Code{Code: string(PRIVATE_INVESTOR), Name: i18nutils.Translate("commoncode.INVESTOR_TYPE.PRIVATE_INVESTOR", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PRIVATE_PRO_INVESTOR), Name: i18nutils.Translate("commoncode.INVESTOR_TYPE.PRIVATE_PRO_INVESTOR", loc), Order: len(codes) + 1})

	return commoncode.CodeGroup{
		GroupCode: "PRIVATE_INVESTOR_TYPE",
		Codes:     codes,
	}
}

func (User) GetCorporateInvestorTypeCodeGroup(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0)
	codes = append(codes, commoncode.Code{Code: string(CORPORATE_INVESTOR), Name: i18nutils.Translate("commoncode.INVESTOR_TYPE.CORPORATE_INVESTOR", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(CORPORATE_PRO_INVESTOR), Name: i18nutils.Translate("commoncode.INVESTOR_TYPE.CORPORATE_PRO_INVESTOR", loc), Order: len(codes) + 1})

	return commoncode.CodeGroup{
		GroupCode: "CORPORATE_INVESTOR_TYPE",
		Codes:     codes,
	}
}

type AUDIT_STATUS string

const (
	AUDIT_STATUS_CODE                         = "AUDIT_STATUS"
	PROVISIONAL_AUDIT            AUDIT_STATUS = "PROVISIONAL_AUDIT"            // 仮登録
	WAIT_AUDIT                   AUDIT_STATUS = "WAIT_AUDIT"                   // 本登録承認待機
	WAIT_IDENTITY_REGISTER_AUDIT AUDIT_STATUS = "WAIT_IDENTITY_REGISTER_AUDIT" // 本人確認待ち
	HOLD_AUDIT                   AUDIT_STATUS = "HOLD_AUDIT"                   // 本登録保留
	WAIT_ANTISOCIAL_AUDIT        AUDIT_STATUS = "WAIT_ANTISOCIAL_AUDIT"        // 反社待機中
	CHECK_ANTISOCIAL_AUDIT       AUDIT_STATUS = "CHECK_ANTISOCIAL_AUDIT"       // 反社確認中
	// CHECK_MY_NUMBER_AUDIT        AUDIT_STATUS = "CHECK_MY_NUMBER_AUDIT"        // マイナンバー承認待機
	WAIT_FINAL_APPROVAL_AUDIT AUDIT_STATUS = "WAIT_FINAL_APPROVAL_AUDIT" // 最終承認待機
	COMPLETED_AUDIT           AUDIT_STATUS = "COMPLETED_AUDIT"           // 本登録承認完了
	INVALIDITY_AUDIT          AUDIT_STATUS = "INVALIDITY_AUDIT"          // 無効
	CHECK_ADDRESS_AUDIT       AUDIT_STATUS = "CHECK_ADDRESS_AUDIT"       // 住所確認中（法人限定）
)

type auditStatusDescription struct {
	english string
	japan   string
}

var auditStatusDescriptions = map[AUDIT_STATUS]auditStatusDescription{
	PROVISIONAL_AUDIT:            {"Provisional Member", "仮登録"},
	WAIT_AUDIT:                   {"Waiting on Approval", "本登録承認待機中"},
	WAIT_IDENTITY_REGISTER_AUDIT: {"Waiting on Identity Verification Registration", "本人確認待ち"},
	HOLD_AUDIT:                   {"Identity Confirmation Pending", "本登録保留"},
	WAIT_ANTISOCIAL_AUDIT:        {"Waiting on Anti-social", "反社待機中"},
	CHECK_ANTISOCIAL_AUDIT:       {"Confirming Anti-Social", "反社確認中"},
	CHECK_ADDRESS_AUDIT:          {"Confirming Address", "アドレス確認中"},
	// CHECK_MY_NUMBER_AUDIT:        {"Confirming My Number", "マイナンバー承認待機"},
	WAIT_FINAL_APPROVAL_AUDIT: {"Waiting on Final Approval", "最終承認待機"},
	COMPLETED_AUDIT:           {"Approved", "本登録承認完了"},
	INVALIDITY_AUDIT:          {"Invalid", "無効"},
}

func (a AUDIT_STATUS) Description() string {
	if desc, ok := auditStatusDescriptions[a]; ok {
		return desc.english
	}
	return "No description available"
}

func (a AUDIT_STATUS) DescriptionJapan() string {
	if desc, ok := auditStatusDescriptions[a]; ok {
		return desc.japan
	}
	return "説明なし"
}

// IsFinalized はMIR-1173で追加。本登録承認完了(COMPLETED_AUDIT)または無効
// (INVALIDITY_AUDIT)を表す。個人投資家の国籍・性別・生年月日・電話番号を
// 管理画面から変更できるのはこれがfalseの間だけで、確定後は変更不可とする。
func (a AUDIT_STATUS) IsFinalized() bool {
	return a == COMPLETED_AUDIT || a == INVALIDITY_AUDIT
}

type AuditStatusCount struct {
	ProvisionalAudit          int `json:"provisional_audit" swagger:"desc(仮登録)"`
	WaitAudit                 int `json:"wait_audit" swagger:"desc(本登録承認待機中)"`
	WaitIdentityRegisterAudit int `json:"wait_identity_register_audit" swagger:"desc(本人確認待ち)"`
	HoldAudit                 int `json:"hold_audit" swagger:"desc(本登録保留)"`
	WaitAntisocialAudit       int `json:"wait_antisocial_audit" swagger:"desc(反社待機中)"`
	CheckAntisocialAudit      int `json:"check_antisocial_audit" swagger:"desc(反社確認中)"`
	CheckAddressAudit         int `json:"check_address_audit" swagger:"desc(アドレス確認中)"`
	// CheckMyNumberAudit        int `json:"check_my_number_audit" swagger:"desc(マイナンバー承認待機)"` // TODO: マイナンバー登録復活後に元に戻す
	WaitFinalApprovalAudit int `json:"wait_final_approval_audit" swagger:"desc(最終承認待機)"`
	CompletedAudit         int `json:"completed_audit" swagger:"desc(本登録承認完了)"`
	InvalidityAudit        int `json:"invalidity_audit"  swagger:"desc(無効)"`
}

func (User) GetAuditStatusCodeGroup(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0)
	codes = append(codes, commoncode.Code{Code: string(PROVISIONAL_AUDIT), Name: i18nutils.Translate("commoncode.AUDIT_STATUS.PROVISIONAL_AUDIT", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(WAIT_AUDIT), Name: i18nutils.Translate("commoncode.AUDIT_STATUS.WAIT_AUDIT", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(WAIT_IDENTITY_REGISTER_AUDIT), Name: i18nutils.Translate("commoncode.AUDIT_STATUS.WAIT_IDENTITY_REGISTER_AUDIT", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(HOLD_AUDIT), Name: i18nutils.Translate("commoncode.AUDIT_STATUS.HOLD_AUDIT", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(WAIT_ANTISOCIAL_AUDIT), Name: i18nutils.Translate("commoncode.AUDIT_STATUS.WAIT_ANTISOCIAL_AUDIT", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(CHECK_ANTISOCIAL_AUDIT), Name: i18nutils.Translate("commoncode.AUDIT_STATUS.CHECK_ANTISOCIAL_AUDIT", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(CHECK_ADDRESS_AUDIT), Name: i18nutils.Translate("commoncode.AUDIT_STATUS.CHECK_ADDRESS_AUDIT", loc), Order: len(codes) + 1})
	// codes = append(codes, commoncode.Code{Code: string(CHECK_MY_NUMBER_AUDIT), Name: i18nutils.Translate("commoncode.AUDIT_STATUS.CHECK_MY_NUMBER_AUDIT", loc), Order: len(codes) + 1}) // TODO: マイナンバー登録復活後に元に戻す
	codes = append(codes, commoncode.Code{Code: string(WAIT_FINAL_APPROVAL_AUDIT), Name: i18nutils.Translate("commoncode.AUDIT_STATUS.WAIT_FINAL_APPROVAL_AUDIT", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(COMPLETED_AUDIT), Name: i18nutils.Translate("commoncode.AUDIT_STATUS.COMPLETED_AUDIT", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(INVALIDITY_AUDIT), Name: i18nutils.Translate("commoncode.AUDIT_STATUS.INVALIDITY_AUDIT", loc), Order: len(codes) + 1})

	return commoncode.CodeGroup{
		GroupCode: AUDIT_STATUS_CODE,
		Codes:     codes,
	}
}

type NATIONALITY_TYPE string

const (
	NATIONALITY_CODE                  = "NATIONALITY_CODE"
	JP               NATIONALITY_TYPE = "JP"     // 일본국적
	OTHERS           NATIONALITY_TYPE = "OTHERS" // 기타국적
)

type nationalityTypeDescription struct {
	english string
	japan   string
}

var nationalityTypeDescriptions = map[NATIONALITY_TYPE]nationalityTypeDescription{
	JP:     {"Japan", "日本"},
	OTHERS: {"Outside of Japan", "日本以外"},
}

func (n NATIONALITY_TYPE) Description() string {
	if desc, ok := nationalityTypeDescriptions[n]; ok {
		return desc.english
	}
	return "No description available"
}

func (n NATIONALITY_TYPE) DescriptionJapan() string {
	if desc, ok := nationalityTypeDescriptions[n]; ok {
		return desc.japan
	}
	return "説明なし"
}

func (User) GetNationalityCodeGroup(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(JP), Name: i18nutils.Translate("commoncode.NATIONALITY_CODE.JP", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(OTHERS), Name: i18nutils.Translate("commoncode.NATIONALITY_CODE.OTHERS", loc), Order: len(codes) + 1})
	return commoncode.CodeGroup{
		GroupCode: NATIONALITY_CODE,
		Codes:     codes,
	}
}

type DEPOSIT_TYPE string

const (
	DEPOSIT_TYPE_CODE              = "DEPOSIT_TYPE_CODE"
	ORDINARY_ACCOUNT  DEPOSIT_TYPE = "ORDINARY_ACCOUNT" // 보통계좌
	CHECKING_ACCOUNT  DEPOSIT_TYPE = "CHECKING_ACCOUNT" // 당좌계좌
)

func (User) GetDepositTypeCodeGroup(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(ORDINARY_ACCOUNT), Name: i18nutils.Translate("commoncode.DEPOSIT_TYPE_CODE.ORDINARY_ACCOUNT", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(CHECKING_ACCOUNT), Name: i18nutils.Translate("commoncode.DEPOSIT_TYPE_CODE.CHECKING_ACCOUNT", loc), Order: len(codes) + 1})
	return commoncode.CodeGroup{
		GroupCode: DEPOSIT_TYPE_CODE,
		Codes:     codes,
	}
}

type SIGNUP_STATUS string

const (
	SIGNUP_STATUS_CODE                                    = "SIGNUP_STATUS"
	SIGNUP_NONE                             SIGNUP_STATUS = "SIGNUP_NONE"                             // 가회원 등록 완료 (A)
	SIGNUP_PRIVATE_PROFILE_BASIC            SIGNUP_STATUS = "SIGNUP_PRIVATE_PROFILE_BASIC"            // 신상 정보 입력 단계 (B-개인-1)
	SIGNUP_PRIVATE_PROFILE_OTHER            SIGNUP_STATUS = "SIGNUP_PRIVATE_PROFILE_OTHER"            // 신상 정보 입력 단계 (B-개인-2)
	SIGNUP_CORPORATE_PROFILE_BASIC          SIGNUP_STATUS = "SIGNUP_CORPORATE_PROFILE_BASIC"          // 신상 정보 입력 단계 1 (B-법인-회사)
	SIGNUP_CORPORATE_PROFILE_OTHER          SIGNUP_STATUS = "SIGNUP_CORPORATE_PROFILE_OTHER"          // 신상 정보 입력 단계 2 (B-법인-회사)
	SIGNUP_CORPORATE_PROFILE_REPRESENTATIVE SIGNUP_STATUS = "SIGNUP_CORPORATE_PROFILE_REPRESENTATIVE" // 신상 정보 입력 단계 (B-법인-대표)
	SIGNUP_CORPORATE_PROFILE_MANAGER        SIGNUP_STATUS = "SIGNUP_CORPORATE_PROFILE_MANAGER"        // 신상 정보 입력 단계 (B-법인-담당자)
	SIGNUP_CORPORATE_PROFILE_SHAREHOLDER    SIGNUP_STATUS = "SIGNUP_CORPORATE_PROFILE_SHAREHOLDER"    // 신상 정보 입력 단계 (B-법인-실질적 지배자)
	SIGNUP_ACCOUNT                          SIGNUP_STATUS = "SIGNUP_ACCOUNT"                          // 출금 계좌 정보 입력 단계 (C)
	SIGNUP_INVESTMENT_SUITABILITY           SIGNUP_STATUS = "SIGNUP_INVESTMENT_SUITABILITY"           // 투자 적합성 확인 단계 (D)
	SIGNUP_AGREE                            SIGNUP_STATUS = "SIGNUP_AGREE"                            // 약관 동의 단계 (E)
	// SIGNUP_MYNUMBER                         SIGNUP_STATUS = "SIGNUP_MYNUMBER"                         // 마이넘버 등록 단계 (F)
	SIGNUP_KYC      SIGNUP_STATUS = "SIGNUP_KYC"      // 본인 확인 단계 (G)
	SIGNUP_COMPLETE SIGNUP_STATUS = "SIGNUP_COMPLETE" // 회원 가입 완료 (H)
)

type CORP_TYPE string

const (
	CORP_TYPE_CODE           = "CORP_TYPE_CODE"
	CORP           CORP_TYPE = "CORP"           // 법인
	MANAGER        CORP_TYPE = "MANAGER"        // 담당자
	REPRESENTATIVE CORP_TYPE = "REPRESENTATIVE" // 대표자
)

func (User) GetCorpTypeCodeGroup(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(CORP), Name: i18nutils.Translate("commoncode.CORP_TYPE_CODE.CORP", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(MANAGER), Name: i18nutils.Translate("commoncode.CORP_TYPE_CODE.MANAGER", loc), Order: len(codes) + 1})
	return commoncode.CodeGroup{
		GroupCode: CORP_TYPE_CODE,
		Codes:     codes,
	}
}

type USER_JOB_TYPE string

const (
	USER_JOB_TYPE_CODE                             = "USER_JOB_TYPE"
	USER_UNLISTED_COMPANY_EXECUTIVES USER_JOB_TYPE = "USER_UNLISTED_COMPANY_EXECUTIVES"
	USER_UNLISTED_COMPANY_EMPLOYEE   USER_JOB_TYPE = "USER_UNLISTED_COMPANY_EMPLOYEE"
	USER_LISTED_COMPANY_EXECUTIVES   USER_JOB_TYPE = "USER_LISTED_COMPANY_EXECUTIVES"
	USER_LISTED_COMPANY_EMPLOYEE     USER_JOB_TYPE = "USER_LISTED_COMPANY_EMPLOYEE"
	USER_PRIVATE_FREE_BUSINESS       USER_JOB_TYPE = "USER_PRIVATE_FREE_BUSINESS"
	USER_OFFICIALS_MANAGER           USER_JOB_TYPE = "USER_OFFICIALS_MANAGER"
	USER_OFFICIALS                   USER_JOB_TYPE = "USER_OFFICIALS"
	USER_LOCAL_OFFICIALS_MANAGER     USER_JOB_TYPE = "USER_LOCAL_OFFICIALS_MANAGER"
	USER_LOCAL_OFFICIALS             USER_JOB_TYPE = "USER_LOCAL_OFFICIALS"
	USER_ORGANIZATION_EXECUTIVE      USER_JOB_TYPE = "USER_ORGANIZATION_EXECUTIVE"
	USER_ORGANIZATION_EMPLOYEE       USER_JOB_TYPE = "USER_ORGANIZATION_EMPLOYEE"
	USER_UNION_EXECUTIVE             USER_JOB_TYPE = "USER_UNION_EXECUTIVE"
	USER_UNION_EMPLOYEE              USER_JOB_TYPE = "USER_UNION_EMPLOYEE"
	USER_HOUSEWIFE                   USER_JOB_TYPE = "USER_HOUSEWIFE"
	USER_STUDENT                     USER_JOB_TYPE = "USER_STUDENT"
	USER_RETIREE                     USER_JOB_TYPE = "USER_RETIREE"
	USER_ETC                         USER_JOB_TYPE = "USER_ETC"
)

type userJobTypeDescription struct {
	english string
	japan   string
}

var userJobTypeDescriptions = map[USER_JOB_TYPE]userJobTypeDescription{
	USER_UNLISTED_COMPANY_EXECUTIVES: {"Unlisted Company Executive", "非上場会社役員"},
	USER_UNLISTED_COMPANY_EMPLOYEE:   {"Unlisted Company Employee", "非上場会社職員"},
	USER_LISTED_COMPANY_EXECUTIVES:   {"Listed Company Executive", "上場会社役員"},
	USER_LISTED_COMPANY_EMPLOYEE:     {"Listed Company Employee", "上場会社職員"},
	USER_PRIVATE_FREE_BUSINESS:       {"Private Business / Free Business", "個人事業主/自営業"},
	USER_OFFICIALS_MANAGER:           {"State Official - Management Position", "国家公務員 管理職"},
	USER_OFFICIALS:                   {"State Official", "国家公務員"},
	USER_LOCAL_OFFICIALS_MANAGER:     {"Local Public Official - Management Position", "地方公務員 管理職"},
	USER_LOCAL_OFFICIALS:             {"Local Official", "地方公務員"},
	USER_ORGANIZATION_EXECUTIVE:      {"Group Executive", "団体役員"},
	USER_ORGANIZATION_EMPLOYEE:       {"Group Staff", "団体職員"},
	USER_UNION_EXECUTIVE:             {"Union Executive", "組合役員"},
	USER_UNION_EMPLOYEE:              {"Union Staff", "組合職員"},
	USER_HOUSEWIFE:                   {"Housewife", "主婦"},
	USER_STUDENT:                     {"Student", "学生"},
	USER_RETIREE:                     {"Retiree", "年金受給者"},
	USER_ETC:                         {"Other", "その他"},
}

func (u USER_JOB_TYPE) Description() string {
	if desc, ok := userJobTypeDescriptions[u]; ok {
		return desc.english
	}
	return "No description available"
}

func (u USER_JOB_TYPE) DescriptionJapan() string {
	if desc, ok := userJobTypeDescriptions[u]; ok {
		return desc.japan
	}
	return "説明なし"
}

func (User) GetUserJobTypeCodeGroup(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(USER_UNLISTED_COMPANY_EXECUTIVES), Name: i18nutils.Translate("commoncode."+USER_JOB_TYPE_CODE+"."+string(USER_UNLISTED_COMPANY_EXECUTIVES), loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_UNLISTED_COMPANY_EMPLOYEE), Name: i18nutils.Translate("commoncode."+USER_JOB_TYPE_CODE+"."+string(USER_UNLISTED_COMPANY_EMPLOYEE), loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_LISTED_COMPANY_EXECUTIVES), Name: i18nutils.Translate("commoncode."+USER_JOB_TYPE_CODE+"."+string(USER_LISTED_COMPANY_EXECUTIVES), loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_LISTED_COMPANY_EMPLOYEE), Name: i18nutils.Translate("commoncode."+USER_JOB_TYPE_CODE+"."+string(USER_LISTED_COMPANY_EMPLOYEE), loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_PRIVATE_FREE_BUSINESS), Name: i18nutils.Translate("commoncode."+USER_JOB_TYPE_CODE+"."+string(USER_PRIVATE_FREE_BUSINESS), loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_OFFICIALS_MANAGER), Name: i18nutils.Translate("commoncode."+USER_JOB_TYPE_CODE+"."+string(USER_OFFICIALS_MANAGER), loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_OFFICIALS), Name: i18nutils.Translate("commoncode."+USER_JOB_TYPE_CODE+"."+string(USER_OFFICIALS), loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_LOCAL_OFFICIALS_MANAGER), Name: i18nutils.Translate("commoncode."+USER_JOB_TYPE_CODE+"."+string(USER_LOCAL_OFFICIALS_MANAGER), loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_LOCAL_OFFICIALS), Name: i18nutils.Translate("commoncode."+USER_JOB_TYPE_CODE+"."+string(USER_LOCAL_OFFICIALS), loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_ORGANIZATION_EXECUTIVE), Name: i18nutils.Translate("commoncode."+USER_JOB_TYPE_CODE+"."+string(USER_ORGANIZATION_EXECUTIVE), loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_ORGANIZATION_EMPLOYEE), Name: i18nutils.Translate("commoncode."+USER_JOB_TYPE_CODE+"."+string(USER_ORGANIZATION_EMPLOYEE), loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_UNION_EXECUTIVE), Name: i18nutils.Translate("commoncode."+USER_JOB_TYPE_CODE+"."+string(USER_UNION_EXECUTIVE), loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_UNION_EMPLOYEE), Name: i18nutils.Translate("commoncode."+USER_JOB_TYPE_CODE+"."+string(USER_UNION_EMPLOYEE), loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_HOUSEWIFE), Name: i18nutils.Translate("commoncode."+USER_JOB_TYPE_CODE+"."+string(USER_HOUSEWIFE), loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_STUDENT), Name: i18nutils.Translate("commoncode."+USER_JOB_TYPE_CODE+"."+string(USER_STUDENT), loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_RETIREE), Name: i18nutils.Translate("commoncode."+USER_JOB_TYPE_CODE+"."+string(USER_RETIREE), loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_ETC), Name: i18nutils.Translate("commoncode."+USER_JOB_TYPE_CODE+"."+string(USER_ETC), loc), Order: len(codes) + 1})

	return commoncode.CodeGroup{
		GroupCode: USER_JOB_TYPE_CODE,
		Codes:     codes,
	}
}

type INVESTMENT_PURPOSE_TYPE string

const (
	INVESTMENT_PURPOSE_TYPE_CODE                         = "INVESTMENT_PURPOSE_TYPE"
	VARIANCE_PORTFOLIO           INVESTMENT_PURPOSE_TYPE = "VARIANCE_PORTFOLIO"
	RAISING_FUTURE_MONEY         INVESTMENT_PURPOSE_TYPE = "RAISING_FUTURE_MONEY"
	MICRO_INVESTMENT             INVESTMENT_PURPOSE_TYPE = "MICRO_INVESTMENT"
	RENEWABLE_ENERGY_INVESTMENT  INVESTMENT_PURPOSE_TYPE = "RENEWABLE_ENERGY_INVESTMENT"
	INVESTMENT_PURPOSE_ETC       INVESTMENT_PURPOSE_TYPE = "ETC"
)

type investmentPurposeDescription struct {
	english string
	japan   string
}

var investmentPurposeDescriptions = map[INVESTMENT_PURPOSE_TYPE]investmentPurposeDescription{
	VARIANCE_PORTFOLIO:          {"Portfolio variance", "ポートフォリオの分散"},
	RAISING_FUTURE_MONEY:        {"Funding for the future", "将来のための資産形成"},
	MICRO_INVESTMENT:            {"Interest in micro-investments", "小口投資への関心"},
	RENEWABLE_ENERGY_INVESTMENT: {"Interest in investing in renewable energy", "再生可能エネルギー投資への関心"},
	INVESTMENT_PURPOSE_ETC:      {"ETC", "その他"},
}

func (i INVESTMENT_PURPOSE_TYPE) Description() string {
	if desc, ok := investmentPurposeDescriptions[i]; ok {
		return desc.english
	}
	return "No description available"
}

func (i INVESTMENT_PURPOSE_TYPE) DescriptionJapan() string {
	if desc, ok := investmentPurposeDescriptions[i]; ok {
		return desc.japan
	}
	return "説明なし"
}

func (User) GetUserInvestmentPurposeTypeCodeGroup(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(VARIANCE_PORTFOLIO), Name: i18nutils.Translate("commoncode."+INVESTMENT_PURPOSE_TYPE_CODE+"."+string(VARIANCE_PORTFOLIO), loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(RAISING_FUTURE_MONEY), Name: i18nutils.Translate("commoncode."+INVESTMENT_PURPOSE_TYPE_CODE+"."+string(RAISING_FUTURE_MONEY), loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(MICRO_INVESTMENT), Name: i18nutils.Translate("commoncode."+INVESTMENT_PURPOSE_TYPE_CODE+"."+string(MICRO_INVESTMENT), loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(RENEWABLE_ENERGY_INVESTMENT), Name: i18nutils.Translate("commoncode."+INVESTMENT_PURPOSE_TYPE_CODE+"."+string(RENEWABLE_ENERGY_INVESTMENT), loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(INVESTMENT_PURPOSE_ETC), Name: i18nutils.Translate("commoncode."+INVESTMENT_PURPOSE_TYPE_CODE+"."+string(INVESTMENT_PURPOSE_ETC), loc), Order: len(codes) + 1})

	return commoncode.CodeGroup{
		GroupCode: INVESTMENT_PURPOSE_TYPE_CODE,
		Codes:     codes,
	}
}

type UserWithRoles struct {
	User        `xorm:"extends"`
	RolesString string `json:"-" xorm:"roles_string <-"`
}

type UsersWithRoles []UserWithRoles

func (u UsersWithRoles) First() UserWithRoles {
	if u.IsEmpty() {
		return UserWithRoles{}
	}

	return u[0]
}

func (u UsersWithRoles) IsEmpty() bool {
	return len(u) == 0
}

type User struct {
	Id                 int64                   `json:"id" swagger:"desc(ID)" xorm:"pk autoincr"`
	Uid                string                  `json:"uid" xorm:"notnull unique" swagger:"desc(사용자 유니크 ID)"`
	LoginType          LOGIN_TYPE              `json:"login_type"`
	UserType           USER_TYPE               `json:"user_type"`
	CommonId           string                  `json:"common_id" xorm:"notnull UUID" swagger:"desc(공용 ID)"`
	InvestorType       INVESTOR_TYPE           `json:"investor_type" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_INVEST_GRADE" swagger:"desc(투자자 타입)"` // 투자자 타입
	Name               string                  `json:"name" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_CORP_NAME" swagger:"desc(이름)" `
	Email              string                  `json:"email" xorm:"notnull index unique" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_EMAIL" swagger:"desc(Email)"`
	Password           string                  `json:"-" xorm:"notnull"`
	Salt               string                  `json:"-" xorm:"notnull" swagger:"desc(패스워드 암호화 Salt)"`
	Roles              []string                `json:"roles" xorm:"-"`
	MobilePhone        string                  `json:"mobile_phone" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_MOBILEPHONE" swagger:"desc(휴대전화 번호)"` // 결제 인증시 휴대폰 인증
	SignUpDate         jsonutils.JSONTime      `json:"sign_up_date" swagger:"desc(회원 가입일)"`
	LastLoginDate      jsonutils.JSONTime      `json:"last_login_date"  swagger:"desc(마지막 로그인 시간)"`
	BlockInvest        bool                    `json:"block_invest" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_TRANSACTION" xorm:"block_invest" swagger:"desc(투자 허용 여부/ 투자 제한/ True 인경우 거래 불가 )"`
	ZipCode            string                  `json:"zip_code" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_ZIPCODE" swagger:"desc(우편번호)"`
	Address            string                  `json:"address" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_ADDRESS" swagger:"desc(주소)"`
	Address2           *string                 `json:"address2" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_ADDRESS2" xorm:"varchar(255)" swagger:"desc(상세주소)"`
	Province           string                  `json:"province" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_PROVINCE" swagger:"desc(현 이름)"`
	City               string                  `json:"city" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_CITY" swagger:"desc(시 또는 구, 군)"`
	AuditStatus        AUDIT_STATUS            `json:"audit_status" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_AUDIT_STATUS" swagger:"desc(공통코드 : [AUDIT_STATUS])"`
	InvestmentPurpose  INVESTMENT_PURPOSE_TYPE `json:"investment_purpose" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_PURPOSE" swagger:"desc(투자목적)"`
	IsSurplusFund      *bool                   `json:"is_surplus_fund" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_OPERATINGFUNDS" swagger:"desc(여유자금 운용 여부)"`
	IsNewsletter       bool                    `json:"is_newsletter" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_NEWSLETTER" swagger:"desc(뉴스레터수신여부)"`
	NewsletterAgreeAt  jsonutils.JSONTime      `json:"newsletter_agree_at" swagger:"desc(뉴스레터수신 동의 시점)" `
	IsSmsCert          bool                    `json:"is_sms_cert" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_CERTIFICATION" swagger:"desc(SMS 문자인증 여부)"`
	SmsCertAt          jsonutils.JSONTime      `json:"sms_cert_at" swagger:"desc(SMS 문자인증 시점)"`
	BankAccount        string                  `json:"bank_account" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_BANK_ACCOUNT" swagger:"desc(은행계좌번호)"`
	BankCode           *string                 `json:"bank_code" swagger:"desc(은행코드)"`
	BranchCode         *string                 `json:"branch_code" swagger:"desc(은행지점코드)"`
	Depositor          string                  `json:"depositor" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_DEPOSITOR" swagger:"desc(예금주)"`
	BankName           string                  `json:"bank_name" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_FINANCE_NAME" swagger:"desc(금융기관명)"`
	BranchName         string                  `json:"branch_name" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_BANK_NAME" swagger:"desc(은행지점명)"`
	DepositType        DEPOSIT_TYPE            `json:"deposit_type" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_DEPOSIT_TYPE" swagger:"desc(예금종류공통코드: [DEPOSIT_TYPE])"`
	PostalCode         string                  `json:"postal_code" swagger:"desc(우편코드)"`
	PostalCodeStatus   bool                    `json:"postal_code_status" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_POSTAL_CODE" swagger:"desc(우편코드 인증 여부 발송 완료: True 미발송:False)"`
	PostalCodeSendDate jsonutils.JSONTime      `json:"postal_code_send_date" swagger:"desc(우편 코드 전송 일자)"`
	SignUpApprovalDate jsonutils.JSONTime      `json:"sign_up_approval_date" xorm:"sign_up_approval_date"`
	InvalidReason      *string                 `json:"invalid_reason"  i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_INVALID_REASON" swagger:"desc(무효처리사유)"`
	CodeApprovalDate   jsonutils.JSONTime      `json:"code_approval_date" xorm:"code_approval_date"`
	MyNumber           *MyNumber               `json:"my_number" xorm:"-"`
	IsEkyc             bool                    `json:"is_ekyc" swagger:"desc(온라인 인증 여부)"`
	AntiResearchDate   jsonutils.JSONTime      `json:"anti_research_date" swagger:"desc(반사회 리서치 제출일)"`
	HoldUserCheckCode  *string                 `json:"hold_user_check_code" swagger:"desc(본인확인보류 사용자 확인 코드)"`
	Created            jsonutils.JSONTime      `json:"created" xorm:"created"`
	Updated            jsonutils.JSONTime      `json:"updated" xorm:"updated"`
	Deleted            jsonutils.JSONTime      `json:"deleted" xorm:"deleted"`
	HasExperience      *bool                   `json:"has_experience" swagger:"desc(투자경험 여부)"`
	SignupStatus       SIGNUP_STATUS           `json:"signup_status" swagger:"desc(회원가입 상태)"`
}

func (u User) TableName() string {
	return "user"
}

func (u User) SelectAll() string {
	return u.TableName() + ".*"
}

// ApproveSignUp はMIR-1205で追加。管理画面から本登録承認完了(COMPLETED_AUDIT)
// にする経路(個人・法人の審査ステータス変更、マイナンバー承認)で共通して使う。
// eKYCやハガキ認証などの自動経路はSignupStatusをSIGNUP_COMPLETEにしてから
// 承認へ進むが、手動登録の会員などはSignupStatusが途中のまま承認されることがあり、
// ユーザー画面はSignupStatusを見て本人確認へ案内するため、承認完了の会員が
// 本人確認画面へ誘導されてしまっていた。承認完了時は必ずSIGNUP_COMPLETEに揃える。
func (u *User) ApproveSignUp(now time.Time) {
	u.AuditStatus = COMPLETED_AUDIT
	u.SignupStatus = SIGNUP_COMPLETE
	u.SignUpApprovalDate = jsonutils.JSONTime(now)
}

func (u *User) Compare(i *User) bool {
	if u.Id != i.Id {
		return false
	}
	if u.Uid != i.Uid {
		return false
	}
	if u.LoginType != i.LoginType {
		return false
	}
	if u.UserType != i.UserType {
		return false
	}
	if u.CommonId != i.CommonId {
		return false
	}
	if u.Name != i.Name {
		return false
	}
	if u.Email != i.Email {
		return false
	}
	if u.Password != i.Password {
		return false
	}
	if u.Salt != i.Salt {
		return false
	}
	if !tool.SliceCompare(u.Roles, i.Roles) {
		return false
	}
	if u.MobilePhone != i.MobilePhone {
		return false
	}
	if u.SignUpDate != i.SignUpDate {
		return false
	}
	if u.LastLoginDate != i.LastLoginDate {
		return false
	}
	if u.BlockInvest != i.BlockInvest {
		return false
	}
	if u.ZipCode != i.ZipCode {
		return false
	}
	if u.Address != i.Address {
		return false
	}
	if u.Address2 != i.Address2 {
		return false
	}
	if u.Province != i.Province {
		return false
	}
	if u.City != i.City {
		return false
	}
	if u.AuditStatus != i.AuditStatus {
		return false
	}
	if u.InvestmentPurpose != i.InvestmentPurpose {
		return false
	}
	if u.IsSurplusFund != i.IsSurplusFund {
		return false
	}
	if u.IsNewsletter != i.IsNewsletter {
		return false
	}
	if u.NewsletterAgreeAt != i.NewsletterAgreeAt {
		return false
	}
	if u.IsSmsCert != i.IsSmsCert {
		return false
	}
	if u.SmsCertAt != i.SmsCertAt {
		return false
	}
	if u.BankAccount != i.BankAccount {
		return false
	}
	if u.BankCode != i.BankCode {
		return false
	}
	if u.BranchCode != i.BranchCode {
		return false
	}
	if u.Depositor != i.Depositor {
		return false
	}
	if u.BankName != i.BankName {
		return false
	}
	if u.BranchName != i.BranchName {
		return false
	}
	if u.DepositType != i.DepositType {
		return false
	}
	if u.PostalCode != i.PostalCode {
		return false
	}
	if u.PostalCodeStatus != i.PostalCodeStatus {
		return false
	}
	if u.PostalCodeSendDate != i.PostalCodeSendDate {
		return false
	}
	if u.SignUpApprovalDate != i.SignUpApprovalDate {
		return false
	}
	if u.InvalidReason != i.InvalidReason {
		return false
	}
	if u.CodeApprovalDate != i.CodeApprovalDate {
		return false
	}
	if u.IsEkyc != i.IsEkyc {
		return false
	}
	if u.AntiResearchDate != i.AntiResearchDate {
		return false
	}
	if u.HoldUserCheckCode != i.HoldUserCheckCode {
		return false
	}
	return true
}

type IUserRepository interface {
	FindByAcceptedNewsletterAndCompletedInvestors(ctx context.Context, projectId int64) ([]User, error)
}

const PRIVATE_SIGNUP_EMAIL_EXPIRE_TIME = time.Minute * 20

// 메일로 인증 // 회원가입 요청 메일 // 회원가입 프로세스 진행 개인 회원 가입
type SignUpRequest struct { // 메일로 전송된 URL 정보
	Id           int64              `json:"id"`
	RequestId    string             `json:"request_id" xorm:"not null UUID"`
	Email        string             `json:"email" swagger:"desc(수신 메일)"`
	Url          string             `json:"url" swagger:"desc(회원 가입 주소)"`
	SendCount    int                `json:"send_count" swagger:"desc(메일로 재전송된 횟수)"`
	UrlExpire    jsonutils.JSONTime `json:"url_expire" swagger:"desc(회원 가입 URL 만료 일자)"`
	EmailExpire  jsonutils.JSONTime `json:"email_expire" swagger:"desc(메일 인증 만료 일자)"`
	UserType     USER_TYPE          `json:"user_type" swagger:"desc(회원 유형)"`
	ReferralCode string             `json:"referral_code" swagger:"desc(紹介コード)"`
	Created      jsonutils.JSONTime `json:"created" xorm:"created"`
}

func (s SignUpRequest) TableName() string {
	return "sign_up_request"
}

func (s SignUpRequest) SelectAll() string {
	return s.TableName() + ".*"
}

// 패스워드 초기화 메일
type PasswordResetRequest struct { // 메일로 전송된 URL 정보
	Id        int64              `json:"id"`
	RequestId string             `json:"request_id" xorm:"not null UUID"`
	Email     string             `json:"email" swagger:"desc(수신 메일)"`
	Url       string             `json:"url" swagger:"desc(회원 가입 주소)"`
	SendCount int                `json:"send_count" swagger:"desc(재전송 메일 횟수)"`
	Expire    jsonutils.JSONTime `json:"expire" swagger:"desc(만료 일자)"`
	Created   jsonutils.JSONTime `json:"created" xorm:"created"`
}

// 메일로 코드 정보 전송
type EmailCodeRequest struct {
	Id       int64              `json:"id"`
	Email    string             `json:"email" swagger:"desc(수신 메일)"`
	Code     string             `json:"code" swagger:"desc(회원 가입 주소)"`
	UserId   int64              `json:"user_id"`
	FirstReq bool               `json:"first_req" swagger:"desc(최초 발송인지 여부)"`
	Expire   jsonutils.JSONTime `json:"expire" swagger:"desc(만료 일자)"`
	Created  jsonutils.JSONTime `json:"created" xorm:"created"`
}

func (e EmailCodeRequest) TableName() string {
	return "email_code_request"
}

func (e EmailCodeRequest) SelectAll() string {
	return e.TableName() + ".*"
}

func (e *EmailCodeRequest) Create(ctx context.Context) (int64, error) {
	return factory.DB(ctx).Insert(e)
}

func (e EmailCodeRequest) Update(ctx context.Context) (int64, error) {
	return factory.DB(ctx).Where("email = ?", e.Email).Update(e)
}

func (e EmailCodeRequest) UpdateById(ctx context.Context) (int64, error) {
	return factory.DB(ctx).ID(e.Id).Update(e)
}

func (EmailCodeRequest) GetByEmail(ctx context.Context, email string, code string) (*EmailCodeRequest, error) {
	var v EmailCodeRequest
	has, err := factory.DB(ctx).Table("email_code_request").Where("email = ? And code = ?", email, code).Get(&v)
	if err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}

	return &v, nil
}

// GetNotExpireByUserIdAndEmail 유효기간이 지나지 않은 이메일 코드 정보 조회
func (e *EmailCodeRequest) GetNotExpireByUserIdAndEmail(ctx context.Context, userId int64, email string) (*EmailCodeRequest, error) {
	var v EmailCodeRequest
	has, err := factory.DB(ctx).Where("user_id = ? and email = ? and expire > ?",
		userId, email, time.Now().Format(time.DateTime)).Get(&v)
	if err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}

	return &v, nil
}

// CountFirstReqByUserIdInToday 오늘 날짜에 최초 발송된 문자 개수
func (e *EmailCodeRequest) CountFirstReqByUserIdInToday(ctx context.Context, userId int64) (int64, error) {
	today := time.Now().Format(time.DateOnly)

	return factory.DB(ctx).Where("user_id = ? and first_req = true and created > ?",
		userId, today).Count(e)
}

// CountResendReqAfterFirstReqByUserId 최초 발송 이후 재발송된 문자 개수
func (e *EmailCodeRequest) CountResendReqAfterFirstReqByUserId(ctx context.Context, userId int64) (int64, error) {
	return factory.DB(ctx).Where("user_id = ? and first_req = false and id > (select max(id) from email_code_request where user_id = ? and first_req = true)",
		userId, userId).Count(e)
}

// GetUserSalt returns a random user salt token.
func GetUserSalt() (string, error) {
	// TODO utils 로 이동
	return tool.RandomString(10)
}

func UniqueIDGenerator(id int64) string {
	// TODO utils 로 이동
	return fmt.Sprintf("1%07d", id)
}

func (u *User) MobilePhoneVerificationConfirm(ctx context.Context) error {
	u.IsSmsCert = true
	u.SmsCertAt = jsonutils.JSONTime(time.Now())
	return u.Update(ctx)
}

// 메일 체크
func (User) GetByEmail(ctx context.Context, email string) (*User, error) {
	var user User

	if has, err := factory.DB(ctx).Unscoped().Table("user").Where("user.email = ?", strings.ToLower(email)).Get(&user); err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}

	return &user, nil
}

// 사용자 모델 생성
func (u *User) Create(ctx context.Context) (int64, error) {
	var uid uuid.UUID
	var err error
	if uid, err = uuid.NewV1(); err != nil {
		return 0, err
	}

	u.CommonId = uuid.NewV5(uid, u.Email).String()
	u.BlockInvest = false
	u.PostalCodeStatus = false

	if u.IsNewsletter {
		u.NewsletterAgreeAt = jsonutils.JSONTime(time.Now())
	}

	if u.Salt, err = GetUserSalt(); err != nil {
		return 0, err
	}
	u.EncodePasswd()
	count, err := factory.DB(ctx).Insert(u)
	if err != nil {
		return 0, err
	}

	u.Uid = UniqueIDGenerator(u.Id)
	if err := u.Update(ctx); err != nil {
		return 0, err
	}

	account := &UserAccount{
		UserId:  u.Id,
		Balance: lo.ToPtr(int64(0)),
	}

	if _, err := account.Create(ctx); err != nil {
		return 0, err
	}

	anti := &UserAntisocialCheck{
		CheckUserId: u.Id,
		Checker:     "",
		CheckStatus: 0,
	}

	if _, err := anti.Create(ctx); err != nil {
		return 0, err
	}

	return count, nil
}

func (u *User) CreatePrivateUser(ctx context.Context) (int64, error) {
	var err error

	// 사용자 기본 정보
	uid, err := uuid.NewV1()
	if err != nil {
		return 0, err
	}
	u.CommonId = uuid.NewV5(uid, u.Email).String()
	u.BlockInvest = false
	u.PostalCodeStatus = false

	// 사용자 추가
	count, err := factory.DB(ctx).Insert(u)
	if err != nil {
		return 0, err
	}

	// 사용자 UID 추가
	u.Uid = UniqueIDGenerator(u.Id)
	if err := u.Update(ctx); err != nil {
		return 0, err
	}

	// 사용자 계좌 생성
	account := &UserAccount{
		UserId:  u.Id,
		Balance: lo.ToPtr(int64(0)),
	}
	if _, err := account.Create(ctx); err != nil {
		return 0, err
	}

	// 반사회 검사 정보 생성
	anti := &UserAntisocialCheck{
		CheckUserId: u.Id,
		Checker:     "",
		CheckStatus: 0,
	}
	if _, err := anti.Create(ctx); err != nil {
		return 0, err
	}

	return count, nil
}

func (u *User) AcceptAntiSocialCheck(ctx context.Context, checkDate jsonutils.JSONTime) error {
	var err error
	if u.IsEkyc {
		u.AuditStatus = COMPLETED_AUDIT
	} else {
		u.AuditStatus = CHECK_ADDRESS_AUDIT
		// 우편 코드 발급
		u.PostalCode, err = modelhelper.GetPostalCode()
	}

	if err != nil {
		return err
	}

	if u.UserType == CORPORATE {
		corp, err := CorporateUser{}.GetById(ctx, u.Id)
		if err != nil {
			return err
		}

		corp.CorporatePostalCode, err = modelhelper.GetPostalCode()
		if err != nil {
			return err
		}

		// MIR-1252: 反社チェックを一度承認済みの法人会員が、管理画面で
		// CHECK_ANTISOCIAL_AUDITへ差し戻されたうえで再承認されると、ここで
		// 新しい郵便コードが発行される。UpdateAuditStatusのCHECK_ADDRESS_AUDIT
		// ケースと同じく、古い承認日をクリアしないと、新しいコードを片方だけ
		// 確認しただけで(もう片方は古い承認日が残っているため)
		// CheckAllPostalCodeの完了判定を満たしてしまう。
		//
		// uとcorp.Userは同じuser行を指すが別々に読み込んだ構造体で、呼び出し元
		// (apps/admin-api/controllers/user.go)はこのメソッドのあとuを使って
		// 別途user.Update(ctx)を呼ぶ。corp側だけクリアしてcorp.Update(ctx)で
		// NULLを書いても、そのあとのuser.Update(ctx)がuに残ったままの古い
		// 非ゼロ値を書き戻してしまう(xormは非ゼロ値のフィールドをMustCols無しでも
		// 自動的に更新対象へ含めるため)。uのCodeApprovalDateもここでゼロへ
		// 戻しておく必要がある。
		u.CodeApprovalDate = jsonutils.JSONTime{}
		corp.CodeApprovalDate = jsonutils.JSONTime{}
		corp.CorporatePostalCodeApprovalDate = jsonutils.JSONTime{}

		err = corp.Update(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (u *User) LinkVaAccounts(ctx context.Context, projectType PROJECT_TYPE) error {
	linkedAccount, err := UserVirtualAccount{}.GetVaByUserId(ctx, u.Id, projectType)
	if err != nil {
		return err
	}
	if linkedAccount != nil {
		if err := linkedAccount.UnLinkUser(ctx); err != nil {
			return err
		}
	}

	account, err := UserVirtualAccount{}.GetVaByUnLinkedAccount(ctx, projectType)
	if err != nil {
		return err
	}
	if account == nil {
		return factory.ErrorNoVirtualAccountToLink.New(err)
	}
	err = account.LinkUser(ctx, u.Id)
	if err != nil {
		return err
	}
	return nil
}

// 사용자 나이 검사
func (u *User) CheckOverAge(ctx context.Context) error {
	if u.UserType == PRIVATE {
		private, err := PrivateUser{}.GetDefaultInfoById(ctx, u.Id)
		if err != nil {
			return factory.ErrorDB.New(err)
		}
		today := time.Now()
		age := math.Floor(today.Sub(private.Birthday.Time()).Hours() / 24 / 365)
		if age >= MAX_INVESTOR_AGE_EXCLUSIVE {
			return factory.ErrorOverAgeInvestor.New(err)
		}
	}
	return nil
}

// EncodePasswd encodes password to safe format.
func (u *User) EncodePasswd() {
	newPasswd := pbkdf2.Key([]byte(u.Password), []byte(u.Salt), 10000, 50, sha256.New)
	u.Password = fmt.Sprintf("%x", newPasswd)
}

// ValidatePassword checks if given password matches the one belongs to the user.
func (u *User) ValidatePassword(passoword string) bool {
	newUser := &User{Password: passoword, Salt: u.Salt}
	newUser.EncodePasswd()
	return subtle.ConstantTimeCompare([]byte(u.Password), []byte(newUser.Password)) == 1
}

func (u *User) IsSamePassword(password string) bool {
	return u.ValidatePassword(password)
}

func (User) GetById(ctx context.Context, id int64) (*User, error) {
	var user User
	var err error

	if has, err := factory.DB(ctx).Table("user").ID(id).Get(&user); err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}

	user.Roles, err = UserRoles{}.GetRoles(ctx, user.Id)
	if err != nil {
		return nil, err
	}
	if len(user.Roles) == 0 {
		user.Roles = []string{string(ROLE_USER)}
	}

	return &user, nil
}

// LockById は user 行をロックして同一ユーザー操作を直列化する。
// MySQL では FOR UPDATE を使い、SQLite（単体テスト）では構文非対応のため省略する。
func (User) LockById(ctx context.Context, id int64) error {
	query := "SELECT id FROM user WHERE id = ?"
	engine, err := factory.Engine(ctx)
	if err != nil {
		return err
	}
	if engine.DriverName() != "sqlite3" {
		query += " FOR UPDATE"
	}

	var lockedID int64
	has, err := factory.DB(ctx).SQL(query, id).Get(&lockedID)
	if err != nil {
		return err
	}
	if !has {
		return errors.New("user not found")
	}
	return nil
}

func (User) GetByUserType(ctx context.Context, userTypes []string) ([]User, error) {
	var userList []User

	err := factory.DB(ctx).Table("user").Select("user.*").In("investor_type", userTypes).Find(&userList)
	if err != nil {
		return nil, err
	}

	return userList, err
}

// FindByAuditStatus 회원심사 상태로 조회
func (User) FindByAuditStatus(ctx context.Context, auditStatus AUDIT_STATUS) ([]User, error) {
	var userList []User

	err := factory.DB(ctx).Table("user").Select("user.*").Where("audit_status = ?", auditStatus).Find(&userList)
	if err != nil {
		return nil, err
	}

	return userList, err
}

func (User) GetAll(ctx context.Context, id, name, email, investorType, auditStatus *string, sortby, order string, offset, limit int, startDate, endDate string) (int64, []User, error) {
	query := factory.DB(ctx).Table("user").Select("user.*")

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

	var userList []User
	var totalCount int64

	totalCount, err := query.Limit(limit, offset).FindAndCount(&userList)
	if err != nil {
		return 0, nil, err
	}

	return totalCount, userList, nil
}

func (u *User) Update(ctx context.Context) (err error) {
	q := factory.DB(ctx).UseBool("block_invest", "is_surplus_fund", "is_newsletter", "is_sms_cert", "is_pro_invest_asset", "is_pro_invest_finance_asset", "postal_code_status")
	_, err = q.ID(u.Id).Update(u)
	if err != nil {
		return err
	}

	return nil
}

// UpdateWithdrawalAccount は、マイページで変えられる出金口座の6列だけを書く(預金者名は
// 変えられない)。6列は送られた値でまとめて置き換える。口座の列だけを持つ構造体を Update に
// 渡すと、UseBool(列…)(MustCols の別名で、値に関係なくその列を書く)が取引停止・メルマガ
// 同意・SMS 認証済み・ハガキ発送済みを false で上書きする(MIR-1481)。
func (u *User) UpdateWithdrawalAccount(ctx context.Context) error {
	_, err := factory.DB(ctx).ID(u.Id).
		Cols("deposit_type", "bank_name", "branch_name", "bank_account", "bank_code", "branch_code").
		Update(u)
	return err
}

func (User) GetAuditStatusCount(ctx context.Context, userType USER_TYPE) (*AuditStatusCount, error) {
	var statusCount AuditStatusCount

	q := factory.DB(ctx).Table("user").Select("" +
		"COUNT(CASE WHEN audit_status = 'PROVISIONAL_AUDIT' THEN 1 ELSE NULL END) AS 'provisional_audit'," +
		"COUNT(CASE WHEN audit_status = 'WAIT_AUDIT' THEN 1 ELSE NULL END) AS 'wait_audit'," +
		"COUNT(CASE WHEN audit_status = 'WAIT_IDENTITY_REGISTER_AUDIT' THEN 1 ELSE NULL END) AS 'wait_identity_register_audit'," +
		"COUNT(CASE WHEN audit_status = 'HOLD_AUDIT' THEN 1 ELSE NULL END) AS 'hold_audit'," +
		"COUNT(CASE WHEN audit_status = 'WAIT_ANTISOCIAL_AUDIT' THEN 1 ELSE NULL END) AS 'wait_antisocial_audit'," +
		"COUNT(CASE WHEN audit_status = 'CHECK_ANTISOCIAL_AUDIT' THEN 1 ELSE NULL END) AS 'check_antisocial_audit'," +
		"COUNT(CASE WHEN audit_status = 'CHECK_ADDRESS_AUDIT' THEN 1 ELSE NULL END) AS 'check_address_audit'," +
		// "COUNT(CASE WHEN audit_status = 'CHECK_MY_NUMBER_AUDIT' THEN 1 ELSE NULL END) AS 'check_my_number_audit'," + // TODO: マイナンバー登録復活後に元に戻す
		"COUNT(CASE WHEN audit_status = 'WAIT_FINAL_APPROVAL_AUDIT' THEN 1 ELSE NULL END) AS 'wait_final_approval_audit'," +
		"COUNT(CASE WHEN audit_status = 'COMPLETED_AUDIT' THEN 1 ELSE NULL END) AS 'completed_audit'," +
		"COUNT(CASE WHEN audit_status = 'INVALIDITY_AUDIT' THEN 1 ELSE NULL END) AS 'invalidity_audit'")
	_, err := q.Where("user_type = ?", string(userType)).Get(&statusCount)
	if err != nil {
		return nil, err
	}
	return &statusCount, err
}

func (User) GetAllEmailsAndNames(ctx context.Context, sortBy, order *string, offset, limit int) (int64, []User, error) {
	var userList []User

	q := factory.DB(ctx).Table("user").
		Select("id, email, name").
		Where("user_type != ?", ADMIN)

	if sortBy == nil {
		*sortBy = "email"
	}
	if order == nil {
		*order = "asc"
	}

	if err := modelhelper.SetSortOrder(q, *sortBy, *order); err != nil {
		factory.Logger(ctx).Error(err)
	}

	totalCount, err := q.Limit(limit, offset).FindAndCount(&userList)
	if err != nil {
		return 0, nil, nil
	}

	return totalCount, userList, err
}

// 일반회원 가입절차 확인 후 모델 생성
func (s *SignUpRequest) Create(ctx context.Context) (int64, error) {
	return factory.DB(ctx).Insert(s)
}

func (SignUpRequest) GetById(ctx context.Context, id int64) (*SignUpRequest, error) {
	var s SignUpRequest
	has, err := factory.DB(ctx).Table("sign_up_request").ID(id).Get(&s)
	if err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}

	return &s, nil
}

func (SignUpRequest) Delete(ctx context.Context, id int64) error {
	_, err := factory.DB(ctx).ID(id).Delete(&SignUpRequest{})
	return err
}

func (SignUpRequest) GetAll(ctx context.Context, offset, limit int) (int64, []SignUpRequest, error) {
	var items []SignUpRequest
	var totalCount int64

	totalCount, err := factory.DB(ctx).Limit(limit, offset).FindAndCount(&items)
	if err != nil {
		return 0, nil, nil
	}

	return totalCount, items, nil
}

func (SignUpRequest) CreateSignUpRequest(ctx context.Context, email string) (*SignUpRequest, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}
	requestId := id.String()

	s := SignUpRequest{
		RequestId:   requestId,
		Email:       email,
		Url:         serviceutils.GetUserServiceDomain() + "/sign-up/private/" + requestId,
		EmailExpire: jsonutils.JSONTime(time.Now().Add(PRIVATE_SIGNUP_EMAIL_EXPIRE_TIME)), // 20분
	}

	_, err = s.Create(ctx)
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func (SignUpRequest) GetNotEmailExpiredByEmail(ctx context.Context, email string) (*SignUpRequest, error) {
	var v SignUpRequest
	now := time.Now().Format(time.DateTime)

	has, err := factory.DB(ctx).Where("email = ? and email_expire > ?", email, now).Get(&v)
	if err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}

	return &v, nil
}

const PASSWORD_RESET_REQUEST_EXPIRE_TIME = time.Minute * 20

func (PasswordResetRequest) TableName() string {
	return "password_reset_request"
}

func (p PasswordResetRequest) SelectAll() string {
	return p.TableName() + ".*"
}

// 패스워드 초기화 메일
func (s *PasswordResetRequest) Create(ctx context.Context) (int64, error) {
	return factory.DB(ctx).Insert(s)
}

func (PasswordResetRequest) GetById(ctx context.Context, id int64) (*PasswordResetRequest, error) {
	var s PasswordResetRequest
	has, err := factory.DB(ctx).Table("password_reset_request").ID(id).Get(&s)
	if err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}

	return &s, nil
}

func (PasswordResetRequest) GetByEmail(ctx context.Context, email string) (*PasswordResetRequest, error) {
	var v PasswordResetRequest
	has, err := factory.DB(ctx).Desc("id").Where("email = ? and expire > ?",
		email, time.Now().Format(time.DateTime)).Get(&v)
	if err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}

	return &v, nil
}

// 패스워드 초기화 URL 생성
func (s *PasswordResetRequest) makeUrl() error {
	if s.RequestId == "" {
		return errors.New("request id is empty")
	}
	s.Url = serviceutils.GetUserServiceDomain() + "/reset/password/" + s.RequestId
	return nil
}

// 패스워드 초기화 메일 만료 시간 설정
func (s *PasswordResetRequest) setUpExpireTime() {
	s.Expire = jsonutils.JSONTime(time.Now().Add(PASSWORD_RESET_REQUEST_EXPIRE_TIME))
}

func (s *PasswordResetRequest) CreatePasswordResetRequest(ctx context.Context, email string) error {
	if email == "" {
		return errors.New("email is empty")
	}
	_, err := PasswordResetRequest{
		Email:  email,
		Expire: jsonutils.JSONTime(time.Now()),
	}.Update(ctx)
	if err != nil {
		return err
	}

	id, err := uuid.NewV4()
	if err != nil {
		return errors.Wrap(err, "failed to generate request id")
	}
	s.RequestId = id.String()
	s.setUpExpireTime()
	if err := s.makeUrl(); err != nil {
		return errors.New("failed to make url")
	}

	_, err = s.Create(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (PasswordResetRequest) GetAll(ctx context.Context, offset, limit int) (int64, []PasswordResetRequest, error) {
	var items []PasswordResetRequest
	var totalCount int64

	totalCount, err := factory.DB(ctx).Limit(limit, offset).FindAndCount(&items)
	if err != nil {
		return 0, nil, nil
	}

	return totalCount, items, nil
}

func (s PasswordResetRequest) Update(ctx context.Context) (int64, error) {
	return factory.DB(ctx).Where("email = ?", s.Email).Update(s)
}

func (s PasswordResetRequest) UpdateById(ctx context.Context) (int64, error) {
	return factory.DB(ctx).ID(s.Id).Update(s)
}

func (User) PrepareAndReturnCsv(items []Investment) [][]string {
	type invest struct {
		TransactionId string             `json:"transaction_id" i18n:"model.user.transaction_id"`           // 거래 번호
		PaymentStatus string             `json:"payment_status" i18n:"model.user.payment_status"`           // 거래 상태
		ProjectType   string             `json:"project_type" i18n:"model.project.project_type"`            // 투자 유형
		ProjectName   string             `json:"project_name" i18n:"model.project.project_name"`            // 상품명
		ProjectStatus string             `json:"project_status" i18n:"model.project.project_status"`        // 모집 상태
		Qty2          int64              `json:"qty2" i18n:"model.user.qty2"`                               // 거래 주(구)수
		Pay           int64              `json:"pay" i18n:"model.user.pay"`                                 // 결제 금액
		InvestDate    jsonutils.JSONTime `json:"invest_date" i18n:"model.user.invest_date" xorm:"not null"` // 거래 일시
	}

	var invests []invest
	for _, item := range items {
		inv := invest{
			TransactionId: item.Share.Id,
			PaymentStatus: string(item.Share.InvestStatus),
			ProjectType:   string(item.Share.ProjectType),
			ProjectName:   item.Project.Name,
			ProjectStatus: string(item.ProjectStatus),
			Qty2:          item.Qty,
			Pay:           item.Qty * item.Price,
			InvestDate:    item.InvestDate,
		}

		if value, ok := reflect.ValueOf(&item.InvestStatus).Elem().Interface().(INVEST_STATUS); ok {
			inv.PaymentStatus = "commoncode.INVEST_STATUS." + string(value)
		}
		if value, ok := reflect.ValueOf(&item.Share.ProjectType).Elem().Interface().(PROJECT_TYPE); ok {
			inv.ProjectType = "commoncode.PROJECT_TYPE." + string(value)
		}
		if value, ok := reflect.ValueOf(&item.ProjectStatus).Elem().Interface().(PROJECT_STATUS); ok {
			inv.ProjectStatus = "commoncode.PROJECT_STATUS." + string(value)
		}

		invests = append(invests, inv)
	}

	cols := factory.ReflectJsonFields(reflect.ValueOf(&invest{}).Elem())
	if len(cols) <= 0 {
		return nil
	}

	records := [][]string{}
	records = append(records, cols)

	for _, v := range invests {
		records = append(records, factory.DataWriteToCsv(reflect.ValueOf(&v).Elem()))
	}

	return records
}

func (User) DownloadCsv(ctx context.Context, w http.ResponseWriter, id int64, projectType, projectStatus *string, sortby, order string, offset, limit int) error {
	q := factory.DB(ctx)

	q.Table("share").Select("share.*,"+INVESTOR_PROJECT_STRINGS+","+INVESTOR_USER_STRINGS+","+INVESTOR_SPV_STRINGS).
		Join("INNER", "project", "project.id = share.project_id").
		Join("INNER", "user", "share.owner_id = user.id").
		Join("INNER", "spv", "project.spv_id = spv.id").
		Where("share.owner_id = ?", id)

	if err := modelhelper.SearchEqQuery(q, "share", "project_type", projectType); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SearchEqQuery(q, "project", "project_status", projectStatus); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SetSortOrder(q, sortby, order); err != nil {
		factory.Logger(ctx).Error(err)
	}

	var invests []Investment
	_, err := q.FindAndCount(&invests)
	if err != nil {
		return nil
	}

	records := User{}.PrepareAndReturnCsv(invests)

	fileName := fmt.Sprintf("%s_%d_%s", "energyshares_member", id, "invest_list.csv")
	csvutils.WriteHeader(w, fileName)

	return csv.NewWriter(w).WriteAll(records)
}

func (User) GetNewsLetterEmails(ctx context.Context) ([]string, error) {
	var emails []string
	err := factory.DB(ctx).Table("user").
		Select("user.email").
		Where("is_newsletter = true").
		Where("audit_status != ?", "INVALIDITY_AUDIT").
		Where("user_type IN (?, ?)", "PRIVATE", "CORPORATE").
		Find(&emails)
	if err != nil {
		return nil, err
	}

	return emails, nil
}

func (User) GetAllNewsletterAllowed(ctx context.Context) ([]User, error) {
	var userList []User

	err := factory.DB(ctx).Table("user").Select("user.*").In("is_newsletter", true).Find(&userList)
	if err != nil {
		return nil, err
	}

	return userList, err
}

func (User) CheckIfAllowNewsletterByEmail(ctx context.Context, email string) (bool, error) {
	var user User

	if has, err := factory.DB(ctx).Where("email = ? and is_newsletter = true", strings.ToLower(email)).Get(&user); err != nil {
		return false, err
	} else if has {
		return true, nil
	}

	return false, nil
}

func (User) Logout(ctx context.Context, refreshToken string) (*RefreshToken, error) {
	token, err := RefreshToken{}.Get(ctx, refreshToken)
	if err != nil {
		return nil, factory.ErrorTokenInvalid.New(err)
	}
	if token == nil {
		return nil, nil
	}
	token.ExpireIn = time.Now().Unix()
	_, err = token.Update(ctx)
	if err != nil {
		return nil, factory.ErrorDB.New(err)
	}

	return token, nil
}
