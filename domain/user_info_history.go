package domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/commoncode"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/modelhelper"
	"github.com/sirupsen/logrus"
)

type USER_HISTORY_PROCESS_TYPE string

const (
	USER_HISTORY_PROCESS_TYPE_CODE                           = "USER_HISTORY_PROCESS_TYPE"
	USER_HISTORY_PROCESS_OPERATOR  USER_HISTORY_PROCESS_TYPE = "USER_HISTORY_PROCESS_OPERATOR" // オペレーター処理
	USER_HISTORY_PROCESS_MEMBER    USER_HISTORY_PROCESS_TYPE = "USER_HISTORY_PROCESS_MEMBER"   // メンバー直接処理
)

func (UserInfoHistory) GetUserHistoryProcessTypeCodeGroup(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(USER_HISTORY_PROCESS_OPERATOR), Name: i18nutils.Translate("commoncode.USER_HISTORY_PROCESS_TYPE.USER_HISTORY_PROCESS_OPERATOR", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_HISTORY_PROCESS_MEMBER), Name: i18nutils.Translate("commoncode.USER_HISTORY_PROCESS_TYPE.USER_HISTORY_PROCESS_MEMBER", loc), Order: len(codes) + 1})
	return commoncode.CodeGroup{
		GroupCode: USER_HISTORY_PROCESS_TYPE_CODE,
		Codes:     codes,
	}
}

type POSTAL_CODE_STATE string

const (
	POSTAL_CODE_STATE_CODE                      = "POSTAL_CODE_STATE"
	POSTAL_CODE_STATE_SENT    POSTAL_CODE_STATE = "POSTAL_CODE_STATE_SENT"    // 送信完了
	POSTAL_CODE_STATE_NOTSENT POSTAL_CODE_STATE = "POSTAL_CODE_STATE_NOTSENT" // 未送信
)

func (UserInfoHistory) GetPostalCodeStateCodeGroup(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(POSTAL_CODE_STATE_SENT), Name: i18nutils.Translate("commoncode.POSTAL_CODE_STATE.POSTAL_CODE_STATE_SENT", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(POSTAL_CODE_STATE_NOTSENT), Name: i18nutils.Translate("commoncode.POSTAL_CODE_STATE.POSTAL_CODE_STATE_NOTSENT", loc), Order: len(codes) + 1})
	return commoncode.CodeGroup{
		GroupCode: POSTAL_CODE_STATE_CODE,
		Codes:     codes,
	}
}

type CONTRACT_BLOCK_TYPE string

const (
	CONTRACT_BLOCK_TYPE_NAME                     = "TRANSACTION_TYPE"
	CONTRACT_BLOCK           CONTRACT_BLOCK_TYPE = "TRANSACTION_BLOCK"    // 取引制限
	CONTRACT_NOTBLOCK        CONTRACT_BLOCK_TYPE = "TRANSACTION_NOTBLOCK" // 取引制限なし
)

func (UserInfoHistory) GetTransactionTypeCodeGroup(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(CONTRACT_BLOCK), Name: i18nutils.Translate("commoncode.TRANSACTION_TYPE.TRANSACTION_BLOCK", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(CONTRACT_NOTBLOCK), Name: i18nutils.Translate("commoncode.TRANSACTION_TYPE.TRANSACTION_NOTBLOCK", loc), Order: len(codes) + 1})
	return commoncode.CodeGroup{
		GroupCode: CONTRACT_BLOCK_TYPE_NAME,
		Codes:     codes,
	}
}

type USER_HISTORY_CHECK_TYPE string

const (
	USER_HISTORY_CHECK_TYPE_CODE                         = "USER_HISTORY_CHECK_TYPE"
	USER_HISTORY_CHECK_YES       USER_HISTORY_CHECK_TYPE = "USER_HISTORY_CHECK_YES" // はい
	USER_HISTORY_CHECK_NO        USER_HISTORY_CHECK_TYPE = "USER_HISTORY_CHECK_NO"  // いいえ
)

func (UserInfoHistory) GetUserHistoryCheckTypeCodeGroup(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(USER_HISTORY_CHECK_YES), Name: i18nutils.Translate("commoncode.USER_HISTORY_CHECK_TYPE.USER_HISTORY_CHECK_YES", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_HISTORY_CHECK_NO), Name: i18nutils.Translate("commoncode.USER_HISTORY_CHECK_TYPE.USER_HISTORY_CHECK_NO", loc), Order: len(codes) + 1})
	return commoncode.CodeGroup{
		GroupCode: USER_HISTORY_CHECK_TYPE_CODE,
		Codes:     codes,
	}
}

type USER_HISTORY_AGREEMENT_TYPE string

const (
	USER_HISTORY_AGREEMENT_TYPE_CODE                             = "USER_HISTORY_AGREEMENT_TYPE"
	USER_HISTORY_AGREEMENT_AGREE     USER_HISTORY_AGREEMENT_TYPE = "USER_HISTORY_AGREEMENT_AGREE"
	USER_HISTORY_AGREEMENT_DISAGREE  USER_HISTORY_AGREEMENT_TYPE = "USER_HISTORY_AGREEMENT_DISAGREE"
)

func (UserInfoHistory) GetUserHistoryAgreementTypeCodeGroup(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(USER_HISTORY_AGREEMENT_AGREE), Name: i18nutils.Translate("commoncode.USER_HISTORY_AGREEMENT_TYPE.USER_HISTORY_AGREEMENT_AGREE", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_HISTORY_AGREEMENT_DISAGREE), Name: i18nutils.Translate("commoncode.USER_HISTORY_AGREEMENT_TYPE.USER_HISTORY_AGREEMENT_DISAGREE", loc), Order: len(codes) + 1})
	return commoncode.CodeGroup{
		GroupCode: USER_HISTORY_AGREEMENT_TYPE_CODE,
		Codes:     codes,
	}
}

type USER_HISTORY_DOCUMENT_TYPE string

const (
	USER_HISTORY_DOCUMENT_TYPE_CODE                            = "USER_HISTORY_DOCUMENT_TYPE"
	USER_HISTORY_DOCUMENT_ADD       USER_HISTORY_DOCUMENT_TYPE = "USER_HISTORY_DOCUMENT_ADD" // 追加
	USER_HISTORY_DOCUMENT_DEL       USER_HISTORY_DOCUMENT_TYPE = "USER_HISTORY_DOCUMENT_DEL" // 削除
)

func (UserInfoHistory) GetUserDocumentTypeCodeGroup(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(USER_HISTORY_DOCUMENT_ADD), Name: i18nutils.Translate("commoncode.USER_HISTORY_DOCUMENT_TYPE.USER_HISTORY_DOCUMENT_ADD", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_HISTORY_DOCUMENT_DEL), Name: i18nutils.Translate("commoncode.USER_HISTORY_DOCUMENT_TYPE.USER_HISTORY_DOCUMENT_DEL", loc), Order: len(codes) + 1})
	return commoncode.CodeGroup{
		GroupCode: USER_HISTORY_DOCUMENT_TYPE_CODE,
		Codes:     codes,
	}
}

type USER_HISTORY_TYPE string

const (
	USER_HISTORY_TYPE_CODE                      = "USER_HISTORY_TYPE"
	USER_HISTORY_CONTENTS     USER_HISTORY_TYPE = "USER_HISTORY_CONTENTS"     // 内容変更
	USER_HISTORY_DOCUMENT     USER_HISTORY_TYPE = "USER_HISTORY_DOCUMENT"     // 添付書類変更
	USER_HISTORY_AUDIT        USER_HISTORY_TYPE = "USER_HISTORY_AUDIT"        // 登録ステータス変更
	USER_HISTORY_TRANSACTION  USER_HISTORY_TYPE = "USER_HISTORY_TRANSACTION"  // 取引制限変更
	USER_HISTORY_INVEST_GRADE USER_HISTORY_TYPE = "USER_HISTORY_INVEST_GRADE" // 投資等級変更
	USER_HISTORY_ANTISOCIAL   USER_HISTORY_TYPE = "USER_HISTORY_ANTISOCIAL"   // 反社会的情報変更
	USER_HISTORY_SHAREHOLDER  USER_HISTORY_TYPE = "USER_HISTORY_SHAREHOLDER"  // 株主情報変更
)

func (UserInfoHistory) GetUserHistoryTypeCodeGroup(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(USER_HISTORY_CONTENTS), Name: i18nutils.Translate("commoncode.USER_HISTORY_TYPE.USER_HISTORY_CONTENTS", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_HISTORY_DOCUMENT), Name: i18nutils.Translate("commoncode.USER_HISTORY_TYPE.USER_HISTORY_DOCUMENT", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_HISTORY_AUDIT), Name: i18nutils.Translate("commoncode.USER_HISTORY_TYPE.USER_HISTORY_AUDIT", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_HISTORY_TRANSACTION), Name: i18nutils.Translate("commoncode.USER_HISTORY_TYPE.USER_HISTORY_TRANSACTION", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_HISTORY_INVEST_GRADE), Name: i18nutils.Translate("commoncode.USER_HISTORY_TYPE.USER_HISTORY_INVEST_GRADE", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_HISTORY_INVEST_GRADE), Name: i18nutils.Translate("commoncode.USER_HISTORY_TYPE.USER_HISTORY_INVEST_GRADE", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_HISTORY_ANTISOCIAL), Name: i18nutils.Translate("commoncode.USER_HISTORY_TYPE.USER_HISTORY_ANTISOCIAL", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_HISTORY_SHAREHOLDER), Name: i18nutils.Translate("commoncode.USER_HISTORY_TYPE.USER_HISTORY_SHAREHOLDER", loc), Order: len(codes) + 1})
	return commoncode.CodeGroup{
		GroupCode: USER_HISTORY_TYPE_CODE,
		Codes:     codes,
	}
}

type USER_CHANGED_CONTENT_TYPE string

const (
	USER_CHANGED_CONTENT_TYPE_CODE                                             = "USER_CHANGED_CONTENT_TYPE"
	USER_CHANGED_CONTENT_AUDIT_STATUS                USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_AUDIT_STATUS"                // 登録ステータス
	USER_CHANGED_CONTENT_TRANSACTION                 USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_TRANSACTION"                 // 取引制限
	USER_CHANGED_CONTENT_INVEST_GRADE                USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_INVEST_GRADE"                // 投資等級
	USER_CHANGED_CONTENT_EMAIL                       USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_EMAIL"                       // メール
	USER_CHANGED_CONTENT_CORP_NAME                   USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_CORP_NAME"                   // 法人名
	USER_CHANGED_CONTENT_NAME                        USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_NAME"                        // 名前
	USER_CHANGED_CONTENT_FAMILY_NAME                 USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_FAMILY_NAME"                 // 苗字
	USER_CHANGED_CONTENT_NAME2                       USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_NAME2"                       // 名前ふりがな
	USER_CHANGED_CONTENT_FAMILY_NAME2                USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_FAMILY_NAME2"                // 苗字ふりがな
	USER_CHANGED_CONTENT_MOBILEPHONE                 USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_MOBILEPHONE"                 // 電話番号
	USER_CHANGED_CONTENT_ZIPCODE                     USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_ZIPCODE"                     // 郵便番号
	USER_CHANGED_CONTENT_ADDRESS                     USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_ADDRESS"                     // 住所
	USER_CHANGED_CONTENT_ADDRESS2                    USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_ADDRESS2"                    // 詳細住所
	USER_CHANGED_CONTENT_PROVINCE                    USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_PROVINCE"                    // 都道府県
	USER_CHANGED_CONTENT_CITY                        USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_CITY"                        // 市区町村
	USER_CHANGED_CONTENT_BIRTHDAY                    USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_BIRTHDAY"                    // 生年月日
	USER_CHANGED_CONTENT_CERTIFICATION               USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_CERTIFICATION"               // 認証
	USER_CHANGED_CONTENT_POSTAL_CODE                 USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_POSTAL_CODE"                 // 郵便コード送信状態
	USER_CHANGED_CONTENT_MANAGER_POSTAL_CODE_STATUS  USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_MANAGER_POSTAL_CODE_STATUS"  // 担当者郵便コード送信状態
	USER_CHANGED_CONTENT_FINANCE_NAME                USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_FINANCE_NAME"                // 金融機関名/コード
	USER_CHANGED_CONTENT_BANK_NAME                   USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_BANK_NAME"                   // 支店名/コード
	USER_CHANGED_CONTENT_DEPOSIT_TYPE                USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_DEPOSIT_TYPE"                // 預金種類
	USER_CHANGED_CONTENT_BANK_ACCOUNT                USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_BANK_ACCOUNT"                // 口座番号
	USER_CHANGED_CONTENT_DEPOSITOR                   USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_DEPOSITOR"                   // 口座名義人
	USER_CHANGED_CONTENT_NEWSLETTER                  USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_NEWSLETTER"                  // ニュースレター受信有無
	USER_CHANGED_CONTENT_IDENTIFICATION1_FRONT       USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_IDENTIFICATION1_FRONT"       // 本人確認書類1表
	USER_CHANGED_CONTENT_IDENTIFICATION1_BACK        USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_IDENTIFICATION1_BACK"        // 本人確認書類1裏
	USER_CHANGED_CONTENT_IDENTIFICATION2_FRONT       USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_IDENTIFICATION2_FRONT"       // 本人確認書類2表
	USER_CHANGED_CONTENT_IDENTIFICATION2_BACK        USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_IDENTIFICATION2_BACK"        // 本人確認書類2裏
	USER_CHANGED_CONTENT_BANK_BOOK                   USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_BANK_BOOK"                   // 出金口座通帳写し
	USER_CHANGED_CONTENT_CERTIFIEDCOPY               USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_CERTIFIEDCOPY"               // 法人登記簿謄本
	USER_CHANGED_CONTENT_ALL_CERTIFICATE_HISTORY     USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_ALL_CERTIFICATE_HISTORY"     // 履歴事項全部証明書
	USER_CHANGED_CONTENT_POWER_OF_ATTORNEY           USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_POWER_OF_ATTORNEY"           // 委任状
	USER_CHANGED_CONTENT_BUSINESSREGISTRATION        USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_BUSINESSREGISTRATION"        // 事業者登録証
	USER_CHANGED_CONTENT_CEO_FIRSTNAME               USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_CEO_FIRSTNAME"               // 代表者名
	USER_CHANGED_CONTENT_CEO_FIRSTNAME_FURIGANA      USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_CEO_FIRSTNAME_FURIGANA"      // 代表者名ふりがな
	USER_CHANGED_CONTENT_CEO_LASTNAME                USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_CEO_LASTNAME"                // 代表者苗字
	USER_CHANGED_CONTENT_CEO_LASTNAME_FURIGANA       USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_CEO_LASTNAME_FURIGANA"       // 代表者苗字ふりがな
	USER_CHANGED_CONTENT_CEO_EMAIL                   USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_CEO_EMAIL"                   // 代表者メール
	USER_CHANGED_CONTENT_CEO_PHONE                   USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_CEO_PHONE"                   // 代表者電話番号
	USER_CHANGED_CONTENT_CEO_CERTIFICATION           USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_CEO_CERTIFICATION"           // 代表者認証
	USER_CHANGED_CONTENT_INCHARGE_FIRSTNAME          USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_INCHARGE_FIRSTNAME"          // 担当者名
	USER_CHANGED_CONTENT_INCHARGE_FIRSTNAME_FURIGANA USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_INCHARGE_FIRSTNAME_FURIGANA" // 担当者名ふりがな
	USER_CHANGED_CONTENT_INCHARGE_LASTNAME           USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_INCHARGE_LASTNAME"           // 担当者苗字
	USER_CHANGED_CONTENT_INCHARGE_LASTNAME_FURIGANA  USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_INCHARGE_LASTNAME_FURIGANA"  // 担当者苗字ふりがな
	USER_CHANGED_CONTENT_INCHARGE_EMAIL              USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_INCHARGE_EMAIL"              // 担当者メール
	USER_CHANGED_CONTENT_INCHARGE_PHONE              USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_INCHARGE_PHONE"              // 担当者電話番号
	USER_CHANGED_CONTENT_MANAGER_BIRTHDAY            USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_MANAGER_BIRTHDAY"            // 担当者生年月日
	USER_CHANGED_CONTENT_INCHARGE_DEPARTMENT         USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_INCHARGE_DEPARTMENT"         // 担当者部署
	USER_CHANGED_CONTENT_INCHARGE_POSITION           USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_INCHARGE_POSITION"           // 担当者職位
	USER_CHANGED_CONTENT_MANAGER_ZIPCODE             USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_MANAGER_ZIPCODE"             // 担当者郵便番号
	USER_CHANGED_CONTENT_MANAGER_PROVINCE            USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_MANAGER_PROVINCE"            // 担当者都道府県
	USER_CHANGED_CONTENT_MANAGER_CITY                USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_MANAGER_CITY"                // 担当者市区町村
	USER_CHANGED_CONTENT_MANAGER_ADDRESS             USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_MANAGER_ADDRESS"             // 担当者住所
	USER_CHANGED_CONTENT_MANAGER_ADDRESS2            USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_MANAGER_ADDRESS2"            // 担当者詳細住所
	USER_CHANGED_CONTENT_MANAGER_CERTIFICATION       USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_MANAGER_CERTIFICATION"       // 担当者認証
	USER_CHANGED_CONTENT_ASSET                       USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_ASSET"                       // 金融資産
	USER_CHANGED_CONTENT_ANNUAL_INCOME               USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_ANNUAL_INCOME"               // 年俸
	USER_CHANGED_CONTENT_PRO_INVEST_ASSET            USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_PRO_INVEST_ASSET"            // 3億円以上の資産保有有無
	USER_CHANGED_CONTENT_PRO_INVEST_FINANCE_ASSET    USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_PRO_INVEST_FINANCE_ASSET"    // 3億円以上の金融資産保有有無
	// USER_CHANGED_CONTENT_PERSONAL_BIRTHDAY はMIR-1173で追加。既存のUSER_CHANGED_
	// CONTENT_BIRTHDAYは株主の生年月日(CorporateShareholder.Birthday)専用で、
	// 個人投資家本人の生年月日と表示文言を共用させると紛らわしいため、法人担当者の
	// USER_CHANGED_CONTENT_MANAGER_BIRTHDAYと同じ形で専用の定数にする。
	USER_CHANGED_CONTENT_PERSONAL_BIRTHDAY                          USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_PERSONAL_BIRTHDAY"                          // 生年月日（個人投資家本人）
	USER_CHANGED_CONTENT_NATIONALITY                                USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_NATIONALITY"                                // 国籍
	USER_CHANGED_CONTENT_GENDER                                     USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_GENDER"                                     // 性別
	USER_CHANGED_CONTENT_PURPOSE                                    USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_PURPOSE"                                    // 投資目的
	USER_CHANGED_CONTENT_OPERATINGFUNDS                             USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_OPERATINGFUNDS"                             // 余裕資金運用有無
	USER_CHANGED_CONTENT_PROFESSIONAL_INVESTOR                      USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_PROFESSIONAL_INVESTOR"                      // 専門投資家制度（通知書）
	USER_CHANGED_CONTENT_IS_ADDRESS_CHANGE                          USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_IS_ADDRESS_CHANGE"                          // 住所変更（投資制限）
	USER_CHANGED_CONTENT_MANAGER_IS_ADDRESS_CHANGE                  USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_MANAGER_IS_ADDRESS_CHANGE"                  // 担当者住所変更（投資制限）
	USER_CHANGED_CONTENT_ANTISOCIAL_CHECKER                         USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_ANTISOCIAL_CHECKER"                         // 반사회 체크자
	USER_CHANGED_CONTENT_ANTISOCIAL_CHECK_DATE                      USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_ANTISOCIAL_CHECK_DATE"                      // 반사회 체크일
	USER_CHANGED_CONTENT_ANTISOCIAL_CHECK_STATUS                    USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_ANTISOCIAL_CHECK_STATUS"                    // 반사회 상태
	USER_CHANGED_CONTENT_ANTISOCIAL_CHECK_APPROVE                   USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_ANTISOCIAL_CHECK_APPROVE"                   // 승인
	USER_CHANGED_CONTENT_ANTISOCIAL_CHECK_DISAPPROVE                USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_ANTISOCIAL_CHECK_DISAPPROVE"                // 비승인
	USER_CHANGED_CONTENT_ANTISOCIAL_CHECK_DOCUMENT                  USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_ANTISOCIAL_CHECK_DOCUMENT"                  // 확인 자료
	USER_CHANGED_CONTENT_ADD                                        USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_ADD"                                        // 추가
	USER_CHANGED_CONTENT_DEL                                        USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_DEL"                                        // 삭제
	USER_CHANGED_CONTENT_INVALID_REASON                             USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_INVALID_REASON"                             // 무효처리사유
	USER_CHANGED_CONTENT_ANTISOCIAL_CHECK_DISAPPROVE_INVALID_REASON USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_ANTISOCIAL_CHECK_DISAPPROVE_INVALID_REASON" // 비승인 처리 사유
	USER_CHANGED_CONTENT_PASSWORD                                   USER_CHANGED_CONTENT_TYPE = "USER_CHANGED_CONTENT_PASSWORD"                                   // パスワード
)

// USER_HISTORY_MASKED_VALUE はパスワード等、値そのものを履歴に残してはいけない項目の表示用マスク値です。
const USER_HISTORY_MASKED_VALUE = "********"

func (UserInfoHistory) GetUserChangedContentTypeCodeGroup(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_AUDIT_STATUS), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_AUDIT_STATUS", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_TRANSACTION), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_TRANSACTION", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_INVEST_GRADE), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_INVEST_GRADE", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_EMAIL), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_EMAIL", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_CORP_NAME), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_CORP_NAME", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_NAME), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_NAME", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_FAMILY_NAME), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_FAMILY_NAME", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_NAME2), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_NAME2", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_FAMILY_NAME2), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_FAMILY_NAME2", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_MOBILEPHONE), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_MOBILEPHONE", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_ZIPCODE), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_ZIPCODE", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_ADDRESS), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_ADDRESS", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_ADDRESS2), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_ADDRESS2", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_PROVINCE), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_PROVINCE", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_CITY), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_CITY", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_BIRTHDAY), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_BIRTHDAY", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_POSTAL_CODE), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_POSTAL_CODE", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_MANAGER_POSTAL_CODE_STATUS), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_MANAGER_POSTAL_CODE_STATUS", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_FINANCE_NAME), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_FINANCE_NAME", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_BANK_NAME), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_BANK_NAME", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_DEPOSIT_TYPE), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_DEPOSIT_TYPE", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_BANK_ACCOUNT), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_BANK_ACCOUNT", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_DEPOSITOR), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_DEPOSITOR", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_NEWSLETTER), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_NEWSLETTER", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_IDENTIFICATION1_FRONT), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_IDENTIFICATION1_FRONT", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_IDENTIFICATION1_BACK), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_IDENTIFICATION1_BACK", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_IDENTIFICATION2_FRONT), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_IDENTIFICATION2_FRONT", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_IDENTIFICATION2_BACK), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_IDENTIFICATION2_BACK", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_BANK_BOOK), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_BANK_BOOK", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_CERTIFIEDCOPY), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_CERTIFIEDCOPY", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_POWER_OF_ATTORNEY), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_POWER_OF_ATTORNEY", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_ALL_CERTIFICATE_HISTORY), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_ALL_CERTIFICATE_HISTORY", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_BUSINESSREGISTRATION), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_BUSINESSREGISTRATION", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_CEO_FIRSTNAME), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_CEO_FIRSTNAME", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_CEO_FIRSTNAME_FURIGANA), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_CEO_FIRSTNAME_FURIGANA", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_CEO_LASTNAME), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_CEO_LASTNAME", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_CEO_LASTNAME_FURIGANA), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_CEO_LASTNAME_FURIGANA", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_CEO_EMAIL), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_CEO_EMAIL", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_CEO_PHONE), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_CEO_PHONE", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_CEO_CERTIFICATION), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_CEO_CERTIFICATION", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_MANAGER_CERTIFICATION), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_MANAGER_CERTIFICATION", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_CERTIFICATION), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_CERTIFICATION", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_INCHARGE_FIRSTNAME), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_INCHARGE_FIRSTNAME", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_INCHARGE_FIRSTNAME_FURIGANA), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_INCHARGE_FIRSTNAME_FURIGANA", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_INCHARGE_LASTNAME), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_INCHARGE_LASTNAME", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_INCHARGE_LASTNAME_FURIGANA), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_INCHARGE_LASTNAME_FURIGANA", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_INCHARGE_EMAIL), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_INCHARGE_EMAIL", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_INCHARGE_PHONE), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_INCHARGE_PHONE", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_MANAGER_BIRTHDAY), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_MANAGER_BIRTHDAY", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_INCHARGE_DEPARTMENT), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_INCHARGE_DEPARTMENT", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_INCHARGE_POSITION), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_INCHARGE_POSITION", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_MANAGER_ZIPCODE), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_MANAGER_ZIPCODE", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_MANAGER_PROVINCE), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_MANAGER_PROVINCE", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_MANAGER_CITY), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_MANAGER_CITY", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_MANAGER_ADDRESS), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_MANAGER_ADDRESS", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_MANAGER_ADDRESS2), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_MANAGER_ADDRESS2", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_ASSET), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_ASSET", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_ANNUAL_INCOME), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_ANNUAL_INCOME", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_PRO_INVEST_ASSET), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_PRO_INVEST_ASSET", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_PRO_INVEST_FINANCE_ASSET), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_PRO_INVEST_FINANCE_ASSET", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_PERSONAL_BIRTHDAY), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_PERSONAL_BIRTHDAY", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_NATIONALITY), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_NATIONALITY", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_GENDER), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_GENDER", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_PURPOSE), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_PURPOSE", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_OPERATINGFUNDS), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_OPERATINGFUNDS", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_PROFESSIONAL_INVESTOR), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_PROFESSIONAL_INVESTOR", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_IS_ADDRESS_CHANGE), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_IS_ADDRESS_CHANGE", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_MANAGER_IS_ADDRESS_CHANGE), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_MANAGER_IS_ADDRESS_CHANGE", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_ANTISOCIAL_CHECKER), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_ANTISOCIAL_CHECKER", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_ANTISOCIAL_CHECK_DATE), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_ANTISOCIAL_CHECK_DATE", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_ANTISOCIAL_CHECK_STATUS), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_ANTISOCIAL_CHECK_STATUS", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_ANTISOCIAL_CHECK_APPROVE), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_ANTISOCIAL_CHECK_APPROVE", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_ANTISOCIAL_CHECK_DISAPPROVE), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_ANTISOCIAL_CHECK_DISAPPROVE", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_ANTISOCIAL_CHECK_DOCUMENT), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_ANTISOCIAL_CHECK_DISAPPROVE", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_ADD), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_ADD", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_DEL), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_DEL", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_INVALID_REASON), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_INVALID_REASON", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_ANTISOCIAL_CHECK_DISAPPROVE_INVALID_REASON), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_ANTISOCIAL_CHECK_DISAPPROVE_INVALID_REASON", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(USER_CHANGED_CONTENT_PASSWORD), Name: i18nutils.Translate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_PASSWORD", loc), Order: len(codes) + 1})

	return commoncode.CodeGroup{
		GroupCode: USER_CHANGED_CONTENT_TYPE_CODE,
		Codes:     codes,
	}
}

type UserInfoHistory struct {
	Id            int64                     `json:"id"`                     // 固有ID
	UserId        int64                     `json:"user_id"`                // 状態変更されたユーザーID
	ProcessUserId int64                     `json:"process_user_id"`        // 状態変更したユーザーID
	SubstanceType string                    `json:"substance_type"`         // 内容表記種別
	ProcessMethod USER_HISTORY_PROCESS_TYPE `json:"process_method"`         // 処理方式（運用者 or 会員自身）
	Name          string                    `json:"name"`                   // 処理者名
	Email         string                    `json:"email"`                  // 処理者のメールアドレス
	Created       jsonutils.JSONTime        `json:"created" xorm:"created"` // 作成日時
}

type UserChangedContent struct {
	Id              int64  `json:"id"`
	InfoId          int64  `json:"info_id"`                           // 유저 상태변경 히스토리 아이디
	ContentType     string `json:"content_type"`                      //
	PrevContent     string `json:"prev_content"`                      // 변경 전 내용
	CurrContent     string `json:"curr_content"`                      // 변경 후 내용
	PrevContentHtml string `json:"prev_long_content" xorm:"LONGTEXT"` // 변경 전 HTML
	CurrContentHtml string `json:"curr_content_html" xorm:"LONGTEXT"` // 변경 후 HTML
	PrevUrl         string `json:"prev_url"`                          // 변경 전 URL
	CurrUrl         string `json:"curr_url"`                          // 변경 후 URL
}

type UserInfoHistoryExtends struct {
	UserInfoHistory `xorm:"extends"`
	Content         UserChangedContent `json:"content" xorm:"extends"`
}

func (UserInfoHistory) GetAll(ctx context.Context, id int64, sortby, order string, offset, limit int) (int64, []UserInfoHistoryExtends, error) {

	query := factory.DB(ctx).Table("user_info_history").Select("user_info_history.id, user_info_history.substance_type, user_info_history.process_method, user_info_history.name, user_info_history.email, user_info_history.created, user_changed_content.*").
		Join("INNER", "user_changed_content", "user_changed_content.info_id = user_info_history.id")
	if id > 0 {
		query.Where("user_id = ?", id)
	}

	if err := modelhelper.SetSortOrder(query, sortby, order); err != nil {
		factory.Logger(ctx).Error(err)
	}

	var userInfoHistoryList []UserInfoHistoryExtends
	var totalCount int64
	totalCount, err := query.Limit(limit, offset).FindAndCount(&userInfoHistoryList)
	if err != nil {
		return 0, nil, err
	}

	return totalCount, userInfoHistoryList, nil
}

func (u *UserInfoHistory) Create(ctx context.Context) (int64, error) {
	return factory.DB(ctx).Insert(u)
}

func (UserChangedContent) Create(ctx context.Context, contentType string, infoId int64, before, after, beforeHtml, afterHtml, beforeUrl, afterUrl string) {
	var userChangedContent UserChangedContent
	userChangedContent.InfoId = infoId
	userChangedContent.ContentType = contentType
	userChangedContent.PrevContent = before
	userChangedContent.CurrContent = after
	userChangedContent.PrevContentHtml = beforeHtml
	userChangedContent.CurrContentHtml = afterHtml
	userChangedContent.PrevUrl = beforeUrl
	userChangedContent.CurrUrl = afterUrl

	if _, err := factory.DB(ctx).Table("user_changed_content").Insert(&userChangedContent); err != nil {
		logrus.Error("failed to save user's changed content")
	}
}

func (UserChangedContent) GetById(ctx context.Context, id int64) (*UserChangedContent, error) {

	var ucc UserChangedContent
	query := factory.DB(ctx).Table("user_changed_content").Select("user_changed_content.*")
	if has, err := query.Where("user_changed_content.id = ?", id).Get(&ucc); err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}

	return &ucc, nil
}
