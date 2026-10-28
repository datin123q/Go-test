package domain

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/commoncode"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/emailutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/modelhelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
)

type MAIL_TYPE string

const (
	MAIL_TYPE_CODE                      = "MAIL_TYPE"
	MAIL_SEND_PRIORITY_FAST   MAIL_TYPE = "FAST"   // 統合メールサービスへのメール送信要求時にFASTに設定
	MAIL_SEND_PRIORITY_NORMAL MAIL_TYPE = "NORMAL" // 統合メールサービスへのメール送信要求時にNORMALに設定
)

func (Mail) GetMailTypeCodeGroup(loc string) commoncode.CodeGroup {

	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(MAIL_SEND_PRIORITY_FAST), Name: i18nutils.Translate("commoncode.MAIL_TYPE.MAIL_SEND_PRIORITY_FAST", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(MAIL_SEND_PRIORITY_NORMAL), Name: i18nutils.Translate("commoncode.MAIL_TYPE.MAIL_SEND_PRIORITY_NORMAL", loc), Order: len(codes) + 1})

	return commoncode.CodeGroup{
		GroupCode: MAIL_TYPE_CODE,
		Codes:     codes,
	}
}

type INVEST_MAIL_TYPE string

const (
	IMMEDIATE        INVEST_MAIL_TYPE = "IMMEDIATE"
	NOT_IMMEDIATE    INVEST_MAIL_TYPE = "NOT_IMMEDIATE"
	WAITER_IMMEDIATE INVEST_MAIL_TYPE = "WAITER_IMMEDIATE"
	WAITER_REQUEST   INVEST_MAIL_TYPE = "WAITER_REQUEST"
	LOTTERY_MANUAL   INVEST_MAIL_TYPE = "LOTTERY_MANUAL"
)

type TEMPLATE_TYPE string

func (t TEMPLATE_TYPE) TemplateName() string {
	// NOTE: 環境移行前は"ESJP_" + string(t) + "_JA_" + envを使っていた
	return string(t)
}

type MAIL_STATUS string

const (
	MAIL_STATUS_CODE             = "MAIL_STATUS"
	MAIL_SENT        MAIL_STATUS = "SENT"    // 統合メールサービスでの送信成功
	MAIL_FAILED      MAIL_STATUS = "FAILED"  // 統合メールサービスでの送信失敗
	MAIL_PENDING     MAIL_STATUS = "PENDING" // 統合メールサービスでの送信待機中
)

func (Mail) GetMailStatusCodeGroup(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(MAIL_SENT), Name: i18nutils.Translate("commoncode.MAIL_STATUS.MAIL_SENT", loc), Order: 1})
	codes = append(codes, commoncode.Code{Code: string(MAIL_FAILED), Name: i18nutils.Translate("commoncode.MAIL_STATUS.MAIL_FAILED", loc), Order: 2})
	codes = append(codes, commoncode.Code{Code: string(MAIL_PENDING), Name: i18nutils.Translate("commoncode.MAIL_STATUS.MAIL_PENDING", loc), Order: 3})

	return commoncode.CodeGroup{
		GroupCode: MAIL_STATUS_CODE,
		Codes:     codes,
	}
}

const (
	NORMAL_MAIL_TEMPLATE                               TEMPLATE_TYPE = "NORMAL_MAIL"                               // 일반 메일
	SIGN_UP_TEMPLATE                                   TEMPLATE_TYPE = "SIGN_UP"                                   // 회원 가입 요청 메일, 회원가입 > (1) 이메일주소 본인인증 시
	PROVISIONAL_PRIVATE_SIGN_UP_COMPLETE_TEMPLATE      TEMPLATE_TYPE = "PROVISIONAL_PRIVATE_SIGN_UP_COMPLETE"      // [개인] 가회원 가입완료 個人仮登録完了
	PROVISIONAL_CORPORATE_SIGN_UP_COMPLETE_TEMPLATE    TEMPLATE_TYPE = "PROVISIONAL_CORPORATE_SIGN_UP_COMPLETE"    // [법인] 가회원 가입완료 法人仮登録完了
	ANTISOCIAL_CHECK_DISAPPROVE_TEMPLATE               TEMPLATE_TYPE = "ANTISOCIAL_CHECK_DISAPPROVE"               // (2-5, 2-6) Admin 개인 및 법인 회원관리 > 반사회체크 :  [비승인] 저장 시 = 가입상태 : ‘무효' 변경 시점
	SIGN_UP_AUDIT_ADDRESS_CHECK_FOR_CORPORATE_TEMPLATE TEMPLATE_TYPE = "SIGN_UP_AUDIT_ADDRESS_CHECK_FOR_CORPORATE" // 법인회원 반사회체크 승인, (2-8) Admin 법인회원관리 > 반사회체크 : [승인] 저장 시 = 가입상태 : ‘주소확인중' 변경 시점
	INVALIDITY_TO_USER_TEMPLATE                        TEMPLATE_TYPE = "INVALIDITY_TO_USER"                        // 개인, 법인 반사회체크 비인증 무효(탈퇴), (2-11) Admin 개인/법인회원 관리 > 가입상태 : ’무효’ 처리 시
	HOLD_AUDIT_STATUS_FOR_CORPORATE_TEMPLATE           TEMPLATE_TYPE = "HOLD_AUDIT_STATUS_FOR_CORPORATE"           // 本登録保留（法人会員用）
	RETRY_SIGN_UP_TEMPLATE                             TEMPLATE_TYPE = "RETRY_SIGN_UP"                             // 再本登録
	JPKI_RETRY_VERIFY_TEMPLATE                         TEMPLATE_TYPE = "JPKI_RETRY_VERIFY"                         // マイナンバーカードでの本人確認のやり直し(MIR-1470)。miralita-mail-service(SendGrid)にテンプレートの追加が要る
	CODE_AUDIT_COMPLETE_FOR_CORPORATE_TEMPLATE         TEMPLATE_TYPE = "CODE_AUDIT_COMPLETE_FOR_CORPORATE"         // 법인회원 본인확인코드 인증완료, (2-10) 법인투자회원이 본인확인코드 인증 완료 시 (최초로그인 성공) = 가입상태 : ‘승인완료' 변경 시점
	COMPLETED_AUDIT_STATUS_FOR_PRIVATE_TEMPLATE        TEMPLATE_TYPE = "COMPLETED_AUDIT_STATUS_FOR_PRIVATE"        // (3. 본인확인 승인 완료) 3A. 개인회원 본등록 승인 완료
	PASSWORD_RESET_TEMPLATE                            TEMPLATE_TYPE = "PASSWORD_RESET"                            // 패스워드 변경요청, (3) 로그인 > 비밀번호 재설정 요청 시
	PASSWORD_RESET_COMPLETE_TEMPLATE                   TEMPLATE_TYPE = "PASSWORD_RESET_COMPLETE"                   // 패스워드 변경완료, (4) 로그인 & 마이페이지 > 비밀번호 변경(재설정) 완료 시
	USER_LOGIN_FAIL_LOCK_TEMPLATE                      TEMPLATE_TYPE = "USER_LOGIN_FAIL_LOCK"                      // 로그인 실패로 계정 잠금, 로그인 7회 실패로 계정 잠금 시
	CERTIFICATION_CODE_SEND_TEMPLATE                   TEMPLATE_TYPE = "CERTIFICATION_CODE_SEND"                   // (5-1) 마이페이지 > 회원정보 이메일 인증코드 발송
	UPDATE_USER_INFORMATION_TEMPLATE                   TEMPLATE_TYPE = "UPDATE_USER_INFORMATION"                   // 회원 정보 수정(전화번호/이메일), (5-2) 마이페이지 > 회원정보 (전화번호/이메일) 수정 저장 시
	FUNDING_REQUEST_TEMPLATE                           TEMPLATE_TYPE = "FUNDING_REQUEST"                           // 펀딩 신청, (6-1) 대출의뢰 완료 시
	FUNDING_SUCCESS_TEMPLATE                           TEMPLATE_TYPE = "FUNDING_SUCCESS"                           // 펀딩 적합, (6-2) (Admin) 대출의뢰 상세 > 상태 : ‘승인완료’ 변경 시
	FUNDING_FAIL_TEMPLATE                              TEMPLATE_TYPE = "FUNDING_FAIL"                              // 펀딩 부적합, (6-3) (Admin) 대출의뢰 상세 > 상태 : ‘ 무효’ 변경 시
	PROJECT_STATUS_END_TEMPLATE                        TEMPLATE_TYPE = "PROJECT_STATUS_END"                        // (8-1) (Admin) 투자상품 관리 > 모집 상태 : ‘모집종료’ 로 변경 시
	PROJECT_STATUS_SUCCESS_TEMPLATE                    TEMPLATE_TYPE = "PROJECT_STATUS_SUCCESS"                    // (8-2) (Admin) 투자상품 관리 > 모집 상태 : ‘성립’으로 변경 시
	PROJECT_STATUS_FAIL_TEMPLATE                       TEMPLATE_TYPE = "PROJECT_STATUS_FAIL"                       // (8-3) (Admin) 투자상품 관리 > 모집 상태 : ‘불성립’으로 변경 시, (Admin) 거래/정산 관리 > 거래상태 : ‘투자 불성립’으로 변경 시
	PROJECT_STATUS_OPERATING_TEMPLATE                  TEMPLATE_TYPE = "PROJECT_STATUS_OPERATING"                  // (8-4) (Admin) 투자상품 관리 > 모집 상태 : ‘운용중’으로 변경 시
	PROJECT_STATUS_OPERATION_CLOSE_TEMPLATE            TEMPLATE_TYPE = "PROJECT_STATUS_OPERATION_CLOSE"            // (8-5) (Admin) 투자상품 관리 > 모집 상태 : ‘운용종료’로 변경 시, (Admin) 거래/정산 관리 > 거래상태 : ‘환불완료’로 변경 시
	INVEST_REQUEST_EQUITY_TEMPLATE                     TEMPLATE_TYPE = "INVEST_REQUEST_EQUITY"                     // (8-6-1) (Admin) 거래/정산 관리 > 거래상태 : ‘투자 신청중(=입금대기중)’ 일 경우 (회원이 [투자하기] 버튼 클릭 시) 사업형
	INVEST_INSUFFICIENT_BALANCE_TEMPLATE               TEMPLATE_TYPE = "INVEST_INSUFFICIENT_BALANCE"               // (8-6-3) (Admin) 예치금 < 투자신청금액 상태에서 예치금이 선 처리되고 잔여 분에 대한 입금요청 메일
	INVEST_SUCCESS_TEMPLATE                            TEMPLATE_TYPE = "INVEST_SUCCESS"                            // 투자대기 -> 투자완료, (8-7) (Admin) 거래/정산 관리 > 거래상태 : ‘투자완료’로 변경 시 (입금 확인 시 자동처리 / 운영자 처리)
	INVEST_IMMEDIATE_SUCCESS_TEMPLATE                  TEMPLATE_TYPE = "INVEST_IMMEDIATE_SUCCESS"                  // (8-7-1) (Admin) 예치금 <= 투자신청금액 상태에서 즉시 투자완료가 되어 발송되는 확인 메일 (8-7과 다름)
	INVEST_CANCEL_TEMPLATE                             TEMPLATE_TYPE = "INVEST_CANCEL"                             // 투자대기 -> 투자취소, (8-8-1) (Admin) 거래/정산 관리 > 거래상태 : ‘투자취소’로 변경 시 (미입금 확인 시 자동처리 / 운영자 수동처리)
	INVEST_INVALIDITY_TEMPLATE                         TEMPLATE_TYPE = "INVEST_INVALIDITY"                         // 투자무효, (8-8-2) (Admin) 투자상품 관리 > 투자상품 상세 > ‘무효’ 버튼 클릭 시
	INVEST_STANDBY_TEMPLATE                            TEMPLATE_TYPE = "INVEST_STANDBY"                            // (8-11) 투자상품 상세페이지 회원이 [투자대기] 버튼 클릭 시
	INVEST_STANDBY_TERMINATION_EQUITY_TEMPLATE         TEMPLATE_TYPE = "INVEST_STANDBY_TERMINATION_EQUITY"         // (8-12-1) (Admin) 대기자 목록 상세 > 대기상태 : ‘입금요청’ 로 변경 시 사업형
	INVEST_STANDBY_CANCELED_TEMPLATE                   TEMPLATE_TYPE = "INVEST_STANDBY_CANCELED"                   // (4. 취소대기) 4E. 순번이 된 투자가가 기한 내에 입금하지 않았을 때
	INVEST_STANDBY_IMMEDIATE_SUCCESS_TEMPLATE          TEMPLATE_TYPE = "INVEST_STANDBY_IMMEDIATE_SUCCESS"          // (4. 취소대기) 4B. 순번이 된 투자가의 예치금이 충분하여 바로 투자완료처리
	INVEST_WAITLIST_CANCEL_TEMPLATE                    TEMPLATE_TYPE = "INVEST_WAITLIST_CANCEL"                    // キャンセル待ち申込のキャンセル
	INVEST_OFFER_TEMPLATE                              TEMPLATE_TYPE = "INVEST_OFFER"                              // (8-13) (Admin) 대기자 목록 상세 > [투자제안 메일 전송] 버튼 클릭 시 (대기상태 : ‘투자 제안중’)
	INVEST_RECRUITMENT_SUCCESS_TEMPLATE                TEMPLATE_TYPE = "INVEST_RECRUITMENT_SUCCESS"                // (8-14) (Admin) 대기자 목록의 회원에게 모집 ‘성립’ 시
	LOTTERY_APPLICATION_TEMPLATE                       TEMPLATE_TYPE = "LOTTERY_APPLICATION"                       // 抽選申込
	LOTTERY_WINNING_IMMEDIATE_SUCCESS_TEMPLATE         TEMPLATE_TYPE = "LOTTERY_WINNING_IMMEDIATE_SUCCESS"         // 抽選当選（入金済み）
	LOTTERY_WINNING_INSUFFICIENT_BALANCE_TEMPLATE      TEMPLATE_TYPE = "LOTTERY_WINNING_INSUFFICIENT_BALANCE"      // 抽選当選（残高不足）
	LOTTERY_LOST_TEMPLATE                              TEMPLATE_TYPE = "LOTTERY_LOST"                              // 抽選落選
	LOTTERY_APPLICATION_CANCEL_TEMPLATE                TEMPLATE_TYPE = "LOTTERY_APPLICATION_CANCEL"                // 抽選申込キャンセル
	PROJECT_NEWS_TEMPLATE                              TEMPLATE_TYPE = "PROJECT_NEWS"                              // 투자상품 뉴스 업데이트, 동의 필요, (9) 내 투자상품 [소식] 업데이트 시
	NEWS_UPLOAD_TEMPLATE                               TEMPLATE_TYPE = "NEWS_UPLOAD"                               // 소식 업로드, 동의필요, (9-1) (Admin) 소식  업로드 시
	QNA_TEMPLATE                                       TEMPLATE_TYPE = "QNA"                                       // 문의하기 답변, (5-6) (Admin) 문의/답변 관리> 답변 최초등록 시
	CONTACT_REQUEST_TEMPLATE                           TEMPLATE_TYPE = "CONTACT_REQUEST"                           // 사용자 문의하기, (10) (운영자가 받는 메일) 사용자가 문의 등록 시
	MY_NUMBER_RESULT_TEMPLATE                          TEMPLATE_TYPE = "MY_NUMBER_RESULT"                          // 마이넘버 등록 결과, (5-7) (Admin) 마이넘버 관리 목록 > [비승인] 선택 시
	WITHDRAWAL_REQUEST_TEMPLATE                        TEMPLATE_TYPE = "WITHDRAWAL_REQUEST"                        // (5-8) 마이페이지 > 출금의뢰 신청 시
	WITHDRAWAL_CANCEL_TEMPLATE                         TEMPLATE_TYPE = "WITHDRAWAL_CANCEL"                         // (Admin) 거래/정산 관리 > 출금 의뢰 관리: 출금의뢰 취소 시
	REISSUE_ACCOUNT_TEMPLATE                           TEMPLATE_TYPE = "REISSUE_ACCOUNT"                           // (11-1, 11-2) (Admin) 투자회원 관리 > 개인/법인 회원 관리 : 가상계좌 재발급 성공시 자동 발송
)

type Mail struct {
	Id           int64              `json:"id" xorm:"pk autoincr"`
	Status       MAIL_STATUS        `json:"status" swagger:"desc(메일 상태)"`
	TemplateType TEMPLATE_TYPE      `json:"template_type" swagger:"desc(템플릿 타입)"`
	Subject      string             `json:"subject" swagger:"desc(제목)"`
	Name         string             `json:"name" swagger:"발송자 이름"`
	Sender       string             `json:"sender" swagger:"desc(발송자 메일)"`
	Content      string             `json:"content" xorm:"LONGTEXT" swagger:"desc(메일 내용)"`
	MailType     MAIL_TYPE          `json:"mail_type" swagger:"desc(메일 종류)"`
	Count        int64              `json:"count" swagger:"desc(발송 수량)"`
	SendDate     jsonutils.JSONTime `json:"send_date" swagger:"desc(전송 일시)"`
	Updated      jsonutils.JSONTime `json:"updated" xorm:"updated"`
	Created      jsonutils.JSONTime `json:"created" xorm:"created"` // 作成日
}

// SendDirectMailは統合メールサービスへのメール送信を要求するリクエストを送信する。
func (Mail) SendDirectMail(ctx context.Context, templateName, sender string, receivers, templateDataList []string, mailType MAIL_TYPE, opts ...emailutils.MailDataOption) error {
	// 統合メールサービスへのメール送信要求
	serviceEnv := serviceutils.GetEnv()
	if serviceEnv == "GITHUB" {
		serviceEnv = "DEV"
	}

	if len(templateDataList) <= 0 {
		return errors.New("No Template Data")
	}
	templateFields, err := TemplateDataToMap(templateDataList[0]) // 현재 모든 사용케이스에서 모든 템플릿 데이터는 같다.
	if err != nil {
		return errors.Wrap(err, "fail unmarshal template data")
	}

	mailData := emailutils.NewMailData(templateFields, sender, receivers, opts...)

	// 統合メールサービスへのリクエスト準備と要求
	emailRequest := emailutils.EmailRequest{
		ServiceName:  "ESJP", // 固定値
		ServiceEnv:   serviceEnv,
		TemplateName: templateName,
		MailData:     []emailutils.MailData{*mailData},
	}

	for i := range emailRequest.MailData {
		emailRequest.MailData[i].SendType = string(mailType)
	}

	_, err = emailutils.SendEmailRequest(ctx, emailRequest)
	if err != nil {
		factory.Logger(ctx).WithFields(logrus.Fields{
			"method": "SendDirectMail",
			"err":    err,
		}).Error("fail send mail request")

		// 統合メールサービスを使用して再送信するためにmail_retryテーブルに保存
		var mailRetry MailRetry
		err := mailRetry.EmailRequestToMailRetry(emailRequest)
		if err != nil {
			factory.Logger(ctx).WithFields(logrus.Fields{
				"method": "SendDirectMail",
				"err":    err,
			}).Error("fail convert email request to mail retry")
			return errors.Wrap(err, "fail convert email request to mail retry")
		}
		_, err = mailRetry.Create(ctx)
		if err != nil {
			factory.Logger(ctx).WithFields(logrus.Fields{
				"method": "SendDirectMail",
				"err":    err,
			}).Error("fail create mail retry")
			return errors.Wrap(err, "fail create mail retry")
		}
	}
	return nil
}

// TemplateDataToMapはtemplateDataListの要素をmap[string]interface{}に変換する。
// すでにマーシャリングされているデータを再度アンマーシャリングしてmap[string]interface{}に変換するレガシー互換の一時的な関数である。
func TemplateDataToMap(templateData string) (map[string]interface{}, error) {

	// jsonパースエラーを防ぐためにスペースに置き換え
	cleanedTemplateData := strings.NewReplacer(
		"\t", " ", // タブ
		"\b", " ", // バックスペース
		"\f", " ", // フォームフィード
		"\r", " ", // キャリッジリターン
	).Replace(templateData)
	var result map[string]interface{}

	if err := json.Unmarshal([]byte(cleanedTemplateData), &result); err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON: %v", err)
	}

	return result, nil
}

// 일반 메일 전송
func (Mail) NormalEmail(ctx context.Context, sender, title, content string, receivers []string) error {

	if len(receivers) <= 0 {
		return errors.New("No Recipient")
	}

	td, err := json.Marshal(map[string]string{
		"ServiceUrl": serviceutils.GetUserServiceDomain(),
		"Subject":    title,
		"Content":    content,
	})
	if err != nil {
		return errors.Wrap(err, "failed to marshal normal email template data")
	}

	var templateDataList []string
	for i := 0; i < len(receivers); i++ {
		templateDataList = append(templateDataList, string(td))
	}

	templateName := Mail{}.ConvertTemplateName(NORMAL_MAIL_TEMPLATE)

	return Mail{}.SendDirectMail(ctx, templateName, sender, receivers, templateDataList, MAIL_SEND_PRIORITY_FAST)
}

// 회원 무효 (탈퇴)시 전송 메일, (2-11) Admin 개인/법인회원 관리 > 가입상태 : ’무효’ 처리 시
func (Mail) SendAuditInvalidityToUserEmail(ctx context.Context, sender, email, name, reason string) error {

	td := fmt.Sprintf("{ \"Receiver\":\"%s\", \"ServiceUrl\":\"%s\", \"Reason\":\"%s\" }", name, serviceutils.GetUserServiceDomain(), reason)
	templateDataList := []string{td}

	templateName := Mail{}.ConvertTemplateName(INVALIDITY_TO_USER_TEMPLATE)

	return Mail{}.SendDirectMail(ctx, templateName, sender, []string{email}, templateDataList, MAIL_SEND_PRIORITY_NORMAL)
}

// (2-5, 2-6) Admin 개인 및 법인 회원관리 > 반사회체크 :  [비승인] 저장 시 = 가입상태 : ‘무효' 변경 시점
func (Mail) AntisocialCheckDisapproveEmail(ctx context.Context, sender, email, name string) error {

	td := fmt.Sprintf("{ \"Receiver\":\"%s\", \"ServiceUrl\":\"%s\", \"ServiceMail\": \"%s\"}", name, serviceutils.GetUserServiceDomain(), serviceutils.GetServiceInfoMail())
	templateDataList := []string{td}

	templateName := Mail{}.ConvertTemplateName(ANTISOCIAL_CHECK_DISAPPROVE_TEMPLATE)

	return Mail{}.SendDirectMail(ctx, templateName, sender, []string{email}, templateDataList, MAIL_SEND_PRIORITY_NORMAL)
}

// 반사회 세력 확인후 주소 체크용 주소 발송 완료, (2-8) Admin 법인회원관리 > 반사회체크 : [승인] 저장 시 = 가입상태 : ‘주소확인중' 변경 시점
func (Mail) SignUpAuditAddressCheckForCorporateEmail(ctx context.Context, sender, email, name string) error {

	td := fmt.Sprintf("{ \"Receiver\":\"%s\", \"ServiceUrl\":\"%s\" }", name, serviceutils.GetUserServiceDomain())
	templateDataList := []string{td}

	templateName := Mail{}.ConvertTemplateName(SIGN_UP_AUDIT_ADDRESS_CHECK_FOR_CORPORATE_TEMPLATE)

	return Mail{}.SendDirectMail(ctx, templateName, sender, []string{email}, templateDataList, MAIL_SEND_PRIORITY_NORMAL)
}

// 법인 본인확인코드 인증완료, (2-10) 법인투자회원이 본인확인코드 인증 완료 시 (최초로그인 성공) = 가입상태 : ‘승인완료' 변경 시점
func (Mail) CodeAuditCompleteForCorporateEmail(ctx context.Context, sender, email, name string) error {

	td := fmt.Sprintf("{ \"Receiver\":\"%s\", \"ServiceUrl\":\"%s\", \"ProductUrl\":\"%s\" }", name, serviceutils.GetUserServiceDomain(), serviceutils.GetUserServiceDomain()+serviceutils.GetProductUrl())
	templateDataList := []string{td}

	templateName := Mail{}.ConvertTemplateName(CODE_AUDIT_COMPLETE_FOR_CORPORATE_TEMPLATE)

	return Mail{}.SendDirectMail(ctx, templateName, sender, []string{email}, templateDataList, MAIL_SEND_PRIORITY_NORMAL)
}

func (Mail) HoldAuditStatusForCorporateEmail(ctx context.Context, sender, email, name string) error {

	td := fmt.Sprintf("{ \"Receiver\":\"%s\", \"ServiceUrl\":\"%s\" }", name, serviceutils.GetUserServiceDomain())
	templateDataList := []string{td}

	templateName := Mail{}.ConvertTemplateName(HOLD_AUDIT_STATUS_FOR_CORPORATE_TEMPLATE)

	return Mail{}.SendDirectMail(ctx, templateName, sender, []string{email}, templateDataList, MAIL_SEND_PRIORITY_NORMAL)
}

// CompletedAuditStatusForPrivateMail (3. 본인확인 승인 완료) 3A. 개인회원 본등록 승인 완료
func (Mail) CompletedAuditStatusForPrivateMail(ctx context.Context, sender, email, name string) error {

	td := fmt.Sprintf("{ \"Receiver\":\"%s\", \"ServiceUrl\":\"%s\", \"ProductUrl\":\"%s\" }", name, serviceutils.GetUserServiceDomain(), serviceutils.GetUserServiceDomain()+serviceutils.GetProductUrl())
	templateDataList := []string{td}

	templateName := Mail{}.ConvertTemplateName(COMPLETED_AUDIT_STATUS_FOR_PRIVATE_TEMPLATE)

	return Mail{}.SendDirectMail(ctx, templateName, sender, []string{email}, templateDataList, MAIL_SEND_PRIORITY_NORMAL)
}

// 비밀번호 변경 신청 메일, (3) 로그인 > 비밀번호 재설정 요청 시
func (Mail) PasswordResetMail(ctx context.Context, sender, email, name, urls string) error {

	td := fmt.Sprintf("{ \"Receiver\":\"%s\", \"PasswordResetUrl\":\"%s\", \"ServiceUrl\":\"%s\" }", name, urls, serviceutils.GetUserServiceDomain())
	templateDataList := []string{td}

	templateName := Mail{}.ConvertTemplateName(PASSWORD_RESET_TEMPLATE)

	return Mail{}.SendDirectMail(ctx, templateName, sender, []string{email}, templateDataList, MAIL_SEND_PRIORITY_FAST)
}

// (5-1) 마이페이지 > 회원정보 이메일 인증코드 발송
func (Mail) CertificationCodeSendMail(ctx context.Context, sender, email, name, code string) error {

	td := fmt.Sprintf("{ \"Receiver\":\"%s\", \"ServiceUrl\":\"%s\", \"CertificationCode\":\"%s\" }", name, serviceutils.GetUserServiceDomain(), code)
	templateDataList := []string{td}

	templateName := Mail{}.ConvertTemplateName(CERTIFICATION_CODE_SEND_TEMPLATE)

	return Mail{}.SendDirectMail(ctx, templateName, sender, []string{email}, templateDataList, MAIL_SEND_PRIORITY_NORMAL)
}

// 회원 정보 수정(전화번호/이메일), (5-2) 마이페이지 > 회원정보 (전화번호/이메일) 수정 저장 시
func (Mail) UpdateUserInformationMail(ctx context.Context, sender, email, name, beforePhone, afterPhone string, isEmail bool) error {

	var info string
	if isEmail {
		info = i18nutils.DefaultTranslate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_EMAIL")
	} else {
		info = i18nutils.DefaultTranslate("commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_MOBILEPHONE")
	}

	td := fmt.Sprintf("{ \"Receiver\":\"%s\", \"ServiceUrl\":\"%s\", \"Info\":\"%s\", \"BeforePhone\":\"%s\", \"AfterPhone\":\"%s\" }", name, serviceutils.GetUserServiceDomain(), info, beforePhone, afterPhone)
	templateDataList := []string{td}

	templateName := Mail{}.ConvertTemplateName(UPDATE_USER_INFORMATION_TEMPLATE)

	return Mail{}.SendDirectMail(ctx, templateName, sender, []string{email}, templateDataList, MAIL_SEND_PRIORITY_NORMAL)
}

// 펀딩신청 -> 대출의뢰 완료, (6-1) 대출의뢰 완료 시
func (Mail) FundingRequestMail(ctx context.Context, sender, email, name, time string) error {

	td := fmt.Sprintf("{ \"Receiver\":\"%s\", \"Time\":\"%s\" , \"ServiceUrl\":\"%s\" }", name, time, serviceutils.GetUserServiceDomain())
	templateDataList := []string{td}

	templateName := Mail{}.ConvertTemplateName(FUNDING_REQUEST_TEMPLATE)

	return Mail{}.SendDirectMail(ctx, templateName, sender, []string{email}, templateDataList, MAIL_SEND_PRIORITY_NORMAL)
}

// 펀딩 적합 -> 대출의뢰 적합, (6-2) (Admin) 대출의뢰 상세 > 상태 : ‘승인완료’ 변경 시
func (Mail) FundingSuccessMail(ctx context.Context, sender, email, name, time string) error {

	td := fmt.Sprintf("{ \"Receiver\":\"%s\", \"ServiceUrl\":\"%s\" }", name, serviceutils.GetUserServiceDomain())
	templateDataList := []string{td}

	templateName := Mail{}.ConvertTemplateName(FUNDING_SUCCESS_TEMPLATE)

	return Mail{}.SendDirectMail(ctx, templateName, sender, []string{email}, templateDataList, MAIL_SEND_PRIORITY_NORMAL)
}

// 펀딩 부적합 -> 대출의뢰 부적합, (6-3) (Admin) 대출의뢰 상세 > 상태 : ‘ 무효’ 변경 시
func (Mail) FundingFailMail(ctx context.Context, sender, email, name, time string) error {

	td := fmt.Sprintf("{ \"Receiver\":\"%s\", \"ServiceUrl\":\"%s\" }", name, serviceutils.GetUserServiceDomain())
	templateDataList := []string{td}

	templateName := Mail{}.ConvertTemplateName(FUNDING_FAIL_TEMPLATE)

	return Mail{}.SendDirectMail(ctx, templateName, sender, []string{email}, templateDataList, MAIL_SEND_PRIORITY_NORMAL)
}

// 투자완료, (8-7) (Admin) 거래/정산 관리 > 거래상태 : ‘투자완료’로 변경 시 (입금 확인 시 자동처리 / 운영자 처리)
func (Mail) InvestSuccessMail(ctx context.Context, sender, email, name, projectName, tradingTime, depositDate, depositAmount string) error {

	da := modelhelper.ConvertNumberToPrice(depositAmount)

	// 상품명, 거래일시, 입금일시, 입금금액
	td := fmt.Sprintf("{ \"Receiver\":\"%s\", \"ServiceUrl\":\"%s\", \"ProjectName\":\"%s\", \"TradingTime\":\"%s\", \"DepositDate\":\"%s\", \"DepositAmount\":\"%s\" }", name, serviceutils.GetUserServiceDomain(), projectName, tradingTime, depositDate, da)
	templateDataList := []string{td}

	templateName := Mail{}.ConvertTemplateName(INVEST_SUCCESS_TEMPLATE)

	return Mail{}.SendDirectMail(ctx, templateName, sender, []string{email}, templateDataList, MAIL_SEND_PRIORITY_NORMAL)
}

// (8-7-1) (Admin) 예치금 <= 투자신청금액 상태에서 즉시 투자완료가 되어 발송되는 확인 메일 (8-7과 다름)
func (Mail) InvestImmediateSuccessMail(ctx context.Context, sender, email, name, projectName, depositAmount string) error {

	da := modelhelper.ConvertNumberToPrice(depositAmount)

	// 상품명, 거래일시, 입금일시, 입금금액
	td := fmt.Sprintf("{ \"Receiver\":\"%s\", \"ServiceUrl\":\"%s\", \"ProjectName\":\"%s\", \"DepositAmount\":\"%s\" }", name, serviceutils.GetUserServiceDomain(), projectName, da)
	templateDataList := []string{td}

	templateName := Mail{}.ConvertTemplateName(INVEST_IMMEDIATE_SUCCESS_TEMPLATE)

	return Mail{}.SendDirectMail(ctx, templateName, sender, []string{email}, templateDataList, MAIL_SEND_PRIORITY_NORMAL)
}

// 투자취소, (8-8-1) (Admin) 거래/정산 관리 > 거래상태 : ‘투자취소’로 변경 시 (미입금 확인 시 자동처리 / 운영자 수동처리)
func (Mail) InvestCancelMail(ctx context.Context, sender, email, name, projectName, tradingTime, cancelTime string) error {

	// 상품명, 거래일시, 취소일시
	td := fmt.Sprintf("{ \"Receiver\":\"%s\", \"ProjectName\":\"%s\", \"TradingTime\":\"%s\", \"CancelTime\":\"%s\", \"ServiceUrl\":\"%s\" }", name, projectName, tradingTime, cancelTime, serviceutils.GetUserServiceDomain())
	templateDataList := []string{td}

	templateName := Mail{}.ConvertTemplateName(INVEST_CANCEL_TEMPLATE)

	return Mail{}.SendDirectMail(ctx, templateName, sender, []string{email}, templateDataList, MAIL_SEND_PRIORITY_NORMAL)
}

// (8-8-2) (Admin) 투자상품 관리 > 투자상품 상세 > ‘무효’ 버튼 클릭 시
func (Mail) InvestCancelDueToInvalidityMail(ctx context.Context, sender, email, name, projectName, tradingTime, cancelTime string) error {

	// 상품명, 거래일시, 취소일시
	td := fmt.Sprintf("{ \"Receiver\":\"%s\", \"ProjectName\":\"%s\", \"TradingTime\":\"%s\", \"CancelTime\":\"%s\", \"ServiceUrl\":\"%s\" }", name, projectName, tradingTime, cancelTime, serviceutils.GetUserServiceDomain())
	templateDataList := []string{td}

	templateName := Mail{}.ConvertTemplateName(INVEST_INVALIDITY_TEMPLATE)

	return Mail{}.SendDirectMail(ctx, templateName, sender, []string{email}, templateDataList, MAIL_SEND_PRIORITY_NORMAL)
}

// 프로젝트 뉴스 업데이트, (9) 내 투자상품 [소식] 업데이트 시
func (Mail) ProjectNewsMail(ctx context.Context, sender, projectName, title, content, time string, receivers []string, projectId int64) error {

	projectView := fmt.Sprintf("%s/product/view/%d?tab=news", serviceutils.GetUserServiceDomain(), projectId)

	td := fmt.Sprintf("{ \"ProjectName\":\"%s\", \"Title\":\"%s\", \"Content\":\"%s\", \"Time\":\"%s\", \"ProjectView\":\"%s\", \"ServiceUrl\":\"%s\" }", projectName, title, transformHtml(content), time, projectView, serviceutils.GetUserServiceDomain())

	var templateDataList []string
	for i := 0; i < len(receivers); i++ {
		templateDataList = append(templateDataList, td)
	}

	templateName := Mail{}.ConvertTemplateName(PROJECT_NEWS_TEMPLATE)

	return Mail{}.SendDirectMail(ctx, templateName, sender, receivers, templateDataList, MAIL_SEND_PRIORITY_NORMAL)
}

// 뉴스 업로드 메일, (9-1) (Admin) 소식  업로드 시
func (Mail) NewsUploadMail(ctx context.Context, sender, title, content, time string, receivers []string) error {

	newsView := fmt.Sprintf("%s/user-support?tab=0", serviceutils.GetUserServiceDomain())

	td := fmt.Sprintf("{ \"Title\":\"%s\", \"Content\":\"%s\", \"Time\":\"%s\", \"NewsView\":\"%s\", \"ServiceUrl\":\"%s\" }", title, transformHtml(content), time, newsView, serviceutils.GetUserServiceDomain())

	var templateDataList []string
	for i := 0; i < len(receivers); i++ {
		templateDataList = append(templateDataList, td)
	}

	templateName := Mail{}.ConvertTemplateName(NEWS_UPLOAD_TEMPLATE)

	return Mail{}.SendDirectMail(ctx, templateName, sender, receivers, templateDataList, MAIL_SEND_PRIORITY_NORMAL)
}

// 문의하기 답변, (5-6) (Admin) 문의/답변 관리> 답변 최초등록 시
func (Mail) QnaMail(ctx context.Context, sender, email, name, title, qnaTime, answerTime string, qnaId int64) error {

	qnaView := fmt.Sprintf("%s/my-page/contact/view/%d", serviceutils.GetUserServiceDomain(), qnaId)

	td := fmt.Sprintf("{ \"Receiver\":\"%s\", \"Title\":\"%s\", \"QnaTime\":\"%s\", \"AnswerTime\":\"%s\", \"QnaView\":\"%s\", \"ServiceUrl\":\"%s\" }", name, title, qnaTime, answerTime, qnaView, serviceutils.GetUserServiceDomain())
	templateDataList := []string{td}

	templateName := Mail{}.ConvertTemplateName(QNA_TEMPLATE)

	return Mail{}.SendDirectMail(ctx, templateName, sender, []string{email}, templateDataList, MAIL_SEND_PRIORITY_NORMAL)
}

// 사용자 문의하기, (10) (운영자가 받는 메일) 사용자가 문의 등록 시
func (Mail) ContactMail(ctx context.Context, sender, receiver, name, email, contactDatetime, content, replyTo, replyToName string) error {

	td := fmt.Sprintf("{ \"ContactTitle\":\"%s\", \"Receiver\":\"%s\", \"UserEmail\":\"%s\", \"QnaTime\":\"%s\", \"Content\":\"%s\", \"ServiceUrl\":\"%s\" }", name, name, email, contactDatetime, transformHtml(content), serviceutils.GetUserServiceDomain())

	receivers := []string{receiver}

	var templateDataList []string
	for i := 0; i < len(receivers); i++ {
		templateDataList = append(templateDataList, td)
	}

	templateName := Mail{}.ConvertTemplateName(CONTACT_REQUEST_TEMPLATE)

	return Mail{}.SendDirectMail(ctx, templateName, sender, receivers, templateDataList, MAIL_SEND_PRIORITY_NORMAL,
		emailutils.WithReplyTo(replyTo), emailutils.WithReplyToName(replyToName))
}

// 마이넘버 등록 결과, (5-7) (Admin) 마이넘버 관리 목록 > [비승인] 선택 시
func (Mail) MyNumberResult(ctx context.Context, sender, email, name string) error {

	td := fmt.Sprintf("{ \"Receiver\":\"%s\", \"ServiceUrl\":\"%s\" }", name, serviceutils.GetUserServiceDomain())
	templateDataList := []string{td}

	templateName := Mail{}.ConvertTemplateName(MY_NUMBER_RESULT_TEMPLATE)

	return Mail{}.SendDirectMail(ctx, templateName, sender, []string{email}, templateDataList, MAIL_SEND_PRIORITY_NORMAL)
}

// WithdrawalRequestMail (5-8) 마이페이지 > 출금의뢰 신청 시
func (Mail) WithdrawalRequestMail(ctx context.Context, sender, email, name, requestDate, amount string) error {

	amount = modelhelper.ConvertNumberToPrice(amount)
	td := fmt.Sprintf("{ \"Receiver\":\"%s\", \"ServiceUrl\":\"%s\", \"RequestDate\":\"%s\", \"Amount\":\"%s\" }", name, serviceutils.GetUserServiceDomain(), requestDate, amount)
	templateDataList := []string{td}

	templateName := Mail{}.ConvertTemplateName(WITHDRAWAL_REQUEST_TEMPLATE)

	return Mail{}.SendDirectMail(ctx, templateName, sender, []string{email}, templateDataList, MAIL_SEND_PRIORITY_NORMAL)
}

// (8-1) (Admin) 투자상품 관리 > 모집 상태 : ‘모집종료’ 로 변경 시
func (Mail) ProjectStatusEndMail(ctx context.Context, sender, email, name, projectName, transactionDate string) error {

	td := fmt.Sprintf("{ \"Receiver\":\"%s\", \"ServiceUrl\":\"%s\", \"ProjectName\":\"%s\", \"TransactionDate\":\"%s\" }", name, serviceutils.GetUserServiceDomain(), projectName, transactionDate)
	templateDataList := []string{td}

	templateName := Mail{}.ConvertTemplateName(PROJECT_STATUS_END_TEMPLATE)

	return Mail{}.SendDirectMail(ctx, templateName, sender, []string{email}, templateDataList, MAIL_SEND_PRIORITY_NORMAL)
}

// MIR-942: ProjectView系の差し込みJSONの組み立てをテストできる形に切り出す
// (ProjectStatusSuccessMail・ProjectStatusOperatingMailで共用。ProductView
// の有無以外は同じ形)。
func buildProjectStatusMailWithProductViewTemplateData(name, productView, projectName string) (string, error) {
	// MIR-942: ファンド名に二重引用符等が入ってもJSONが壊れないよう、新しく
	// 足すProjectNameはjson.Marshalで組み立てる(他の既存フィールドの
	// fmt.Sprintf部分は変更しない)。
	projectNameJSON, err := marshalMailTemplateField(projectName)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("{ \"Receiver\":\"%s\", \"ServiceUrl\":\"%s\", \"ProductView\":\"%s\", \"ProjectName\":%s }", name, serviceutils.GetUserServiceDomain(), productView, projectNameJSON), nil
}

// MIR-942: ProjectNameだけを足す差し込みJSONの組み立てをテストできる形に
// 切り出す(ProjectStatusFailMail・ProjectStatusOperationCloseMailで共用)。
func buildProjectStatusMailTemplateData(name, projectName string) (string, error) {
	// MIR-942: ファンド名に二重引用符等が入ってもJSONが壊れないよう、新しく
	// 足すProjectNameはjson.Marshalで組み立てる。
	projectNameJSON, err := marshalMailTemplateField(projectName)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("{ \"Receiver\":\"%s\", \"ServiceUrl\":\"%s\", \"ProjectName\":%s }", name, serviceutils.GetUserServiceDomain(), projectNameJSON), nil
}

// (8-2) (Admin) 투자상품 관리 > 모집 상태 : ‘성립’으로 변경 시
func (Mail) ProjectStatusSuccessMail(ctx context.Context, sender, email, name, projectName, productView string) error {

	td, err := buildProjectStatusMailWithProductViewTemplateData(name, productView, projectName)
	if err != nil {
		return err
	}
	templateDataList := []string{td}

	templateName := Mail{}.ConvertTemplateName(PROJECT_STATUS_SUCCESS_TEMPLATE)

	return Mail{}.SendDirectMail(ctx, templateName, sender, []string{email}, templateDataList, MAIL_SEND_PRIORITY_NORMAL)
}

// (8-3) (Admin) 투자상품 관리 > 모집 상태 : ‘불성립’으로 변경 시
// (Admin) 거래/정산 관리 > 거래상태 : ‘투자 불성립’으로 변경 시
func (Mail) ProjectStatusFailMail(ctx context.Context, sender, email, name, projectName string) error {

	td, err := buildProjectStatusMailTemplateData(name, projectName)
	if err != nil {
		return err
	}
	templateDataList := []string{td}

	templateName := Mail{}.ConvertTemplateName(PROJECT_STATUS_FAIL_TEMPLATE)

	return Mail{}.SendDirectMail(ctx, templateName, sender, []string{email}, templateDataList, MAIL_SEND_PRIORITY_NORMAL)
}

// (8-4) (Admin) 투자상품 관리 > 모집 상태 : ‘운용중’으로 변경 시
func (Mail) ProjectStatusOperatingMail(ctx context.Context, sender, email, name, projectName, productView string) error {

	td, err := buildProjectStatusMailWithProductViewTemplateData(name, productView, projectName)
	if err != nil {
		return err
	}
	templateDataList := []string{td}

	templateName := Mail{}.ConvertTemplateName(PROJECT_STATUS_OPERATING_TEMPLATE)

	return Mail{}.SendDirectMail(ctx, templateName, sender, []string{email}, templateDataList, MAIL_SEND_PRIORITY_NORMAL)
}

// (8-5) (Admin) 투자상품 관리 > 모집 상태 : ‘운용종료’로 변경 시
// (Admin) 거래/정산 관리 > 거래상태 : ‘환불완료’로 변경 시
func (Mail) ProjectStatusOperationCloseMail(ctx context.Context, sender, email, name, projectName string) error {

	td, err := buildProjectStatusMailTemplateData(name, projectName)
	if err != nil {
		return err
	}
	templateDataList := []string{td}

	templateName := Mail{}.ConvertTemplateName(PROJECT_STATUS_OPERATION_CLOSE_TEMPLATE)

	return Mail{}.SendDirectMail(ctx, templateName, sender, []string{email}, templateDataList, MAIL_SEND_PRIORITY_NORMAL)
}

// (8-6) (Admin) 거래/정산 관리 > 거래상태 : ‘투자 신청중(=입금대기중)’ 일 경우 (회원이 [투자하기] 버튼 클릭 시)
func (Mail) InvestRequestMail(ctx context.Context, sender, email, name, projectName, transactionDate, paymentEnd, investmentAmount, bankName, branchName, accountNumber, vaHolderName string, projectType PROJECT_TYPE) error {

	ia := modelhelper.ConvertNumberToPrice(investmentAmount)
	// 투자상품명, 거래일시, 입금기한, 총 투자금액, 은행명, 지점명, 계좌번호, 가상계좌 명의인 이름
	td := fmt.Sprintf("{ \"Receiver\":\"%s\", \"ServiceUrl\":\"%s\", \"ProjectName\":\"%s\", \"TransactionDate\":\"%s\", \"PaymentEnd\":\"%s\", \"InvestmentAmount\":\"%s\", \"BankName\":\"%s\", \"BranchName\":\"%s\", \"AccountNumber\":\"%s\", \"VaHolderName\":\"%s\" }", name, serviceutils.GetUserServiceDomain(), projectName, transactionDate, paymentEnd, ia, bankName, branchName, accountNumber, vaHolderName)
	templateDataList := []string{td}

	if projectType != EQUITY {
		return fmt.Errorf("invalid project type: %s", projectType)
	}

	templateName := Mail{}.ConvertTemplateName(INVEST_REQUEST_EQUITY_TEMPLATE)

	return Mail{}.SendDirectMail(ctx, templateName, sender, []string{email}, templateDataList, MAIL_SEND_PRIORITY_NORMAL)
}

// InvestInsufficientBalanceMail [1.입금 요청]-[1B. 0 < 예치금 < 투자 신청 금액]: 예치금이 선 처리되고 잔여분에 대한 입금 요청
// 기존 기획: 8-6-3
func (Mail) InvestInsufficientBalanceMail(ctx context.Context, sender, email, name, lackAmount, productName, transactionDate, bankName, branchName, accountNumber, depositor, investmentAmount, paymentEnd string) error {

	la := modelhelper.ConvertNumberToPrice(lackAmount)
	ia := modelhelper.ConvertNumberToPrice(investmentAmount)

	// 이름, 입금 의뢰금, 펀드명, 거래일시, 입금기한, 총 투자금액, 가상 계좌 정보(은행명, 지점명, 구좌번호, 구좌명의인)
	td := fmt.Sprintf("{ \"Receiver\":\"%s\", \"ServiceUrl\":\"%s\", \"LackAmount\":\"%s\", \"ProductName\":\"%s\", \"TransactionDate\":\"%s\", \"BankName\":\"%s\", \"BranchName\":\"%s\", \"AccountNumber\":\"%s\", \"Depositor\":\"%s\", \"InvestmentAmount\":\"%s\", \"PaymentEnd\":\"%s\" }",
		name, serviceutils.GetUserServiceDomain(), la, productName, transactionDate, bankName, branchName, accountNumber, depositor, ia, paymentEnd)
	templateDataList := []string{td}

	templateName := Mail{}.ConvertTemplateName(INVEST_INSUFFICIENT_BALANCE_TEMPLATE)

	return Mail{}.SendDirectMail(ctx, templateName, sender, []string{email}, templateDataList, MAIL_SEND_PRIORITY_NORMAL)
}

// (8-11) 투자상품 상세페이지 회원이 [투자대기] 버튼 클릭 시
func (Mail) InvestStandbyMail(ctx context.Context, sender, email, name, projectName, requestDate, sharesCount, investmentAmount string) error {

	sc := modelhelper.ConvertNumberToPrice(sharesCount)
	ia := modelhelper.ConvertNumberToPrice(investmentAmount)

	// 투자상품명, 신청일시, 투자구수, 총 투자금액
	td := fmt.Sprintf("{ \"Receiver\":\"%s\", \"ServiceUrl\":\"%s\", \"ProjectName\":\"%s\", \"RequestDate\":\"%s\", \"SharesCount\":\"%s\", \"InvestmentAmount\":\"%s\" }", name, serviceutils.GetUserServiceDomain(), projectName, requestDate, sc, ia)
	templateDataList := []string{td}

	templateName := Mail{}.ConvertTemplateName(INVEST_STANDBY_TEMPLATE)

	return Mail{}.SendDirectMail(ctx, templateName, sender, []string{email}, templateDataList, MAIL_SEND_PRIORITY_NORMAL)
}

// (8-12) (Admin) 대기자 목록 상세 > 대기상태 : ‘입금요청’ 로 변경 시 사업형, 대출형
func (Mail) InvestStandbyTerminationMail(ctx context.Context, sender, email, name, projectName, requestDate, paymentEnd, investmentAmount, bankName, branchName, accountNumber, vaHolderName string, projectType PROJECT_TYPE) error {

	ia := modelhelper.ConvertNumberToPrice(investmentAmount)

	// 투자상품명, 신청일시, 입금기한, 총 투자금액, 은행명, 지점명, 계좌번호, 가상계좌 명의인 이름
	td := fmt.Sprintf("{ \"Receiver\":\"%s\", \"ServiceUrl\":\"%s\", \"ProjectName\":\"%s\", \"RequestDate\":\"%s\", \"PaymentEnd\":\"%s\", \"InvestmentAmount\":\"%s\", \"BankName\":\"%s\", \"BranchName\":\"%s\", \"AccountNumber\":\"%s\", \"VaHolderName\":\"%s\" }", name, serviceutils.GetUserServiceDomain(), projectName, requestDate, paymentEnd, ia, bankName, branchName, accountNumber, vaHolderName)
	templateDataList := []string{td}

	if projectType != EQUITY {
		return fmt.Errorf("invalid project type: %s", projectType)
	}

	templateName := Mail{}.ConvertTemplateName(INVEST_STANDBY_TERMINATION_EQUITY_TEMPLATE)

	return Mail{}.SendDirectMail(ctx, templateName, sender, []string{email}, templateDataList, MAIL_SEND_PRIORITY_NORMAL)
}

// InvestStandbyCanceledMail (4. 취소대기) 4E. 순번이 된 투자가가 기한 내에 입금하지 않았을 때
func (Mail) InvestStandbyCanceledMail(ctx context.Context, sender, email, name, projectName string) error {

	// 투자상품명
	td := fmt.Sprintf("{ \"Receiver\":\"%s\", \"ServiceUrl\":\"%s\", \"ProjectName\":\"%s\" }", name, serviceutils.GetUserServiceDomain(), projectName)
	templateDataList := []string{td}

	templateName := Mail{}.ConvertTemplateName(INVEST_STANDBY_CANCELED_TEMPLATE)

	return Mail{}.SendDirectMail(ctx, templateName, sender, []string{email}, templateDataList, MAIL_SEND_PRIORITY_NORMAL)
}

// MIR-942/971: InvestStandbyImmediateSuccessMailの差し込みJSONの組み立てを
// テストできる形に切り出す。
//
// 【判断】依頼ではInvestmentAmountだけをjson.Marshalで組み立て、ProjectName
// (この関数では既存のフィールド)は対象外とされていたが、テーブル駆動テストで
// 実際にファンド名に「"」が入ると壊れることを確認したため、ProjectNameも
// 他の5関数と同じくjson.Marshalで組み立てる形にした。
func buildInvestStandbyImmediateSuccessMailTemplateData(name, projectName, investmentAmount string) (string, error) {
	projectNameJSON, err := marshalMailTemplateField(projectName)
	if err != nil {
		return "", err
	}

	// MIR-971: 投資金額を足す(InvestImmediateSuccessMailの渡し方に揃える。
	// 申込金額は今回入れない)。
	ia := modelhelper.ConvertNumberToPrice(investmentAmount)
	// MIR-942: 新しく足すInvestmentAmountもjson.Marshalで組み立てる。
	investmentAmountJSON, err := marshalMailTemplateField(ia)
	if err != nil {
		return "", err
	}

	// 투자상품명, 투자금액
	return fmt.Sprintf("{ \"Receiver\":\"%s\", \"ServiceUrl\":\"%s\", \"ProjectName\":%s, \"InvestmentAmount\":%s }", name, serviceutils.GetUserServiceDomain(), projectNameJSON, investmentAmountJSON), nil
}

// InvestStandbyImmediateSuccessMail (4. 취소대기) 4B. 순번이 된 투자가의 예치금이 충분하여 바로 투자완료처리
func (Mail) InvestStandbyImmediateSuccessMail(ctx context.Context, sender, email, name, projectName, investmentAmount string) error {

	td, err := buildInvestStandbyImmediateSuccessMailTemplateData(name, projectName, investmentAmount)
	if err != nil {
		return err
	}
	templateDataList := []string{td}

	templateName := Mail{}.ConvertTemplateName(INVEST_STANDBY_IMMEDIATE_SUCCESS_TEMPLATE)

	return Mail{}.SendDirectMail(ctx, templateName, sender, []string{email}, templateDataList, MAIL_SEND_PRIORITY_NORMAL)
}

// (8-13) (Admin) 대기자 목록 상세 > [투자제안 메일 전송] 버튼 클릭 시 (대기상태 : ‘투자 제안중’)
func (Mail) InvestOfferMail(ctx context.Context, sender, email, name, investedQty, investableQty, productView string) error {

	iq := modelhelper.ConvertNumberToPrice(investedQty)
	itq := modelhelper.ConvertNumberToPrice(investableQty)

	// 투자한 구수, 투자 가능한 구수
	td := fmt.Sprintf("{ \"Receiver\":\"%s\", \"ServiceUrl\":\"%s\", \"InvestedQty\":\"%s\", \"InvestableQty\":\"%s\", \"ProductView\":\"%s\" }", name, serviceutils.GetUserServiceDomain(), iq, itq, productView)
	templateDataList := []string{td}

	templateName := Mail{}.ConvertTemplateName(INVEST_OFFER_TEMPLATE)

	return Mail{}.SendDirectMail(ctx, templateName, sender, []string{email}, templateDataList, MAIL_SEND_PRIORITY_NORMAL)
}

// (8-14) (Admin) 대기자 목록의 회원에게 모집 ‘성립’ 시
// MIR-942: InvestRecruitmentSuccessMailの差し込みJSONの組み立てをテスト
// できる形に切り出す。
func buildInvestRecruitmentSuccessMailTemplateData(name, projectName string) (string, error) {
	productUrl := serviceutils.GetUserServiceDomain() + serviceutils.GetProductUrl()

	// MIR-942: ファンド名に二重引用符等が入ってもJSONが壊れないよう、新しく
	// 足すProjectNameはjson.Marshalで組み立てる。
	projectNameJSON, err := marshalMailTemplateField(projectName)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("{ \"Receiver\":\"%s\", \"ServiceUrl\":\"%s\", \"ProductUrl\":\"%s\", \"ProjectName\":%s }", name, serviceutils.GetUserServiceDomain(), productUrl, projectNameJSON), nil
}

func (Mail) InvestRecruitmentSuccessMail(ctx context.Context, sender, email, name, projectName string) error {
	td, err := buildInvestRecruitmentSuccessMailTemplateData(name, projectName)
	if err != nil {
		return err
	}
	templateDataList := []string{td}

	templateName := Mail{}.ConvertTemplateName(INVEST_RECRUITMENT_SUCCESS_TEMPLATE)

	return Mail{}.SendDirectMail(ctx, templateName, sender, []string{email}, templateDataList, MAIL_SEND_PRIORITY_NORMAL)
}

// (11-1, 11-2) (Admin) 투자회원 관리 > 개인/법인 회원 관리 : 가상계좌 재발급 성공시 자동 발송
func (Mail) ReissueAccountMail(ctx context.Context, sender, email, name, accountNumber string) error {

	td := fmt.Sprintf("{ \"Receiver\":\"%s\", \"ServiceUrl\":\"%s\", \"AccountNumber\":\"%s\", \"ServiceMail\":\"%s\" }", name, serviceutils.GetUserServiceDomain(), accountNumber, serviceutils.GetServiceInfoMail())
	templateDataList := []string{td}

	templateName := Mail{}.ConvertTemplateName(REISSUE_ACCOUNT_TEMPLATE)

	return Mail{}.SendDirectMail(ctx, templateName, sender, []string{email}, templateDataList, MAIL_SEND_PRIORITY_NORMAL)
}

// WithdrawalCancelMail (Admin) 거래/정산 관리 > 출금의뢰 관리 : 출금의뢰 취소 시 자동 발송
func (Mail) WithdrawalCancelMail(ctx context.Context, email, name string) error {
	td := fmt.Sprintf("{ \"Receiver\":\"%s\", \"ServiceUrl\":\"%s\", \"ServiceMail\":\"%s\" }", name, serviceutils.GetUserServiceDomain(), serviceutils.GetServiceInfoMail())
	templateDataList := []string{td}

	templateName := Mail{}.ConvertTemplateName(WITHDRAWAL_CANCEL_TEMPLATE)

	return Mail{}.SendDirectMail(ctx, templateName, serviceutils.GetServiceInfoMail(), []string{email}, templateDataList, MAIL_SEND_PRIORITY_NORMAL)
}

func (Mail) RetrySignUpEmail(ctx context.Context, sender, email, name string, reasons []string) error {
	data := map[string]interface{}{
		"Receiver":   name,
		"ServiceUrl": serviceutils.GetUserServiceDomain(),
		"Reasons":    reasons,
	}

	td, err := json.Marshal(data)
	if err != nil {
		return errors.Wrap(err, "fail marshal retry sign up email template data")
	}

	templateDataList := []string{string(td)}

	templateName := Mail{}.ConvertTemplateName(RETRY_SIGN_UP_TEMPLATE)

	return Mail{}.SendDirectMail(ctx, templateName, sender, []string{email}, templateDataList, MAIL_SEND_PRIORITY_NORMAL)
}

// JpkiRetryVerifyEmail は、マイナンバーカードでの本人確認(JPKI方式)が完了しなかった会員へ、もう一度
// 本人確認をお願いするメールを送る(MIR-1470)。署名の失敗・中断・有効期限切れが確定したとき、または
// 署名結果の通知が届かず、定期ジョブが失敗と確定したときに送る。会員が画面を離れたあとに確定すると、
// 会員は失敗に気付けないため。
//
// 再本登録メール(RetrySignUpEmail)とは別のテンプレート(JPKI_RETRY_VERIFY)を使う。再本登録は、
// 本登録の入力からやり直す案内だが、こちらは、マイページの本人確認の案内から、もう一度
// 本人確認を進めてもらう案内で、会員の状態は変えない。
//
// reasons は利用者向けの理由(service.JpkiSignatureResult.MemberReasons)。空にはしない。
func (Mail) JpkiRetryVerifyEmail(ctx context.Context, sender, email, name string, reasons []string) error {
	td, err := jpkiRetryVerifyTemplateData(name, reasons)
	if err != nil {
		return err
	}

	templateName := Mail{}.ConvertTemplateName(JPKI_RETRY_VERIFY_TEMPLATE)

	return Mail{}.SendDirectMail(ctx, templateName, sender, []string{email}, []string{td}, MAIL_SEND_PRIORITY_NORMAL)
}

// jpkiRetryVerifyTemplateData は、JPKI_RETRY_VERIFY テンプレートへ渡す値(JSON)を作る。
// テンプレートの変数は、再本登録メールと同じ Receiver・ServiceUrl・Reasons。
func jpkiRetryVerifyTemplateData(name string, reasons []string) (string, error) {
	data := map[string]interface{}{
		"Receiver":   name,
		"ServiceUrl": serviceutils.GetUserServiceDomain(),
		"Reasons":    reasons,
	}
	td, err := json.Marshal(data)
	if err != nil {
		return "", errors.Wrap(err, "fail marshal jpki retry verify email template data")
	}
	return string(td), nil
}

func (Mail) ConvertTemplateName(templateName TEMPLATE_TYPE) string {
	// NOTE: 環境移行前は"ESJP_" + string(templateName) + "_JA_" + envを使っていた
	return string(templateName)
}

// MIR-942: ファンド名などの差し込み値にダブルクォートやバックスラッシュが
// 含まれていても、fmt.Sprintfで組み立てるテンプレート用JSONが壊れないように
// する。json.Marshalした文字列は前後のダブルクォートを含んだ形で返るため、
// 呼び出し側ではfmt.Sprintfの書式に"%s"ではなく%sとして埋め込む
// (既存のフィールドのfmt.Sprintf組み立て自体は対象外。この関数は新しく
// 足す値だけに使う)。
func marshalMailTemplateField(value string) (string, error) {
	b, err := json.Marshal(value)
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal mail template field")
	}
	return string(b), nil
}

type MailSendOption struct {
	// 문자열 옵션
	AuditStatus  string
	UserType     string
	Nationality  string
	Gender       string
	Job          string
	Asset        string
	AnnualIncome string

	// true, false
	AllowsNewsletter *bool
	HasMyNumber      *bool

	// date ~ date
	SignUpRequestDate  []string
	SignUpApprovalDate []string
	LastLoginDate      []string

	// number ~ number
	Age                   []int64
	TotalAsset            []int64
	InvestOperationAmount []int64
	Deposit               []int64
	InvestSuccessAmount   []int64
	InvestRequestAmount   []int64
	UnpaidAmount          []int64
}

func (option *MailSendOption) IsNotEmpty() bool {
	if len(option.AuditStatus) > 0 {
		return true
	}
	if len(option.UserType) > 0 {
		return true
	}
	if len(option.Nationality) > 0 {
		return true
	}
	if len(option.Gender) > 0 {
		return true
	}
	if len(option.Job) > 0 {
		return true
	}
	if len(option.Asset) > 0 {
		return true
	}
	if len(option.AnnualIncome) > 0 {
		return true
	}
	if option.AllowsNewsletter != nil {
		return true
	}
	if option.HasMyNumber != nil {
		return true
	}
	if len(option.SignUpRequestDate) > 0 {
		return true
	}
	if len(option.SignUpApprovalDate) > 0 {
		return true
	}
	if len(option.LastLoginDate) > 0 {
		return true
	}
	if len(option.Age) > 0 {
		return true
	}
	if len(option.TotalAsset) > 0 {
		return true
	}
	if len(option.InvestOperationAmount) > 0 {
		return true
	}
	if len(option.Deposit) > 0 {
		return true
	}
	if len(option.InvestSuccessAmount) > 0 {
		return true
	}
	if len(option.InvestRequestAmount) > 0 {
		return true
	}
	if len(option.UnpaidAmount) > 0 {
		return true
	}

	return false
}

func (Mail) GetReceivers(ctx context.Context, options *MailSendOption) ([]User, error) {

	q := factory.DB(ctx).Table("user").
		Distinct("user.email").
		Join("left", "private_profile", "user.id = private_profile.user_id").
		Join("left", "user_asset_view", "user.id = user_asset_view.user_id")

	// search condition
	q.Where("user.user_type != ?", ADMIN)
	if len(options.AuditStatus) > 0 {
		q.And(fmt.Sprintf("user.audit_status IN (%v)", options.AuditStatus))
	} else {
		q.And("user.audit_status != ?", INVALIDITY_AUDIT) // 특정하지 않으면 무효상태는 항상 배제
	}
	if len(options.UserType) > 0 {
		q.And(fmt.Sprintf("user.user_type IN (%v)", options.UserType))
	}
	if len(options.Nationality) > 0 {
		q.And(fmt.Sprintf("private_profile.nationality IN (%v)", options.Nationality))
	}
	if len(options.Gender) > 0 {
		q.And(fmt.Sprintf("private_profile.gender IN (%v)", options.Gender))
	}
	if len(options.Job) > 0 {
		q.And(fmt.Sprintf("private_profile.job IN (%v)", options.Job))
	}
	if len(options.Asset) > 0 {
		q.And(fmt.Sprintf("private_profile.asset IN (%v)", options.Asset))
	}
	if len(options.AnnualIncome) > 0 {
		q.And(fmt.Sprintf("private_profile.annual_income IN (%v)", options.AnnualIncome))
	}
	if options.AllowsNewsletter != nil {
		q.And("user.is_newsletter = ?", *options.AllowsNewsletter)
	}
	if options.HasMyNumber != nil {
		q.Join("left", "my_number", "user.id = my_number.user_id and AND (my_number.deleted IS NULL OR my_number.deleted = '0001-01-01 00:00:00')")
		if *options.HasMyNumber {
			q.And("my_number.status = ?", APPROVE)
		} else {
			q.And("my_number.status IS NULL OR my_number.status != ?", APPROVE)
		}
	}
	if len(options.SignUpRequestDate) > 0 {
		q.And("user.created BETWEEN ? AND ?", options.SignUpRequestDate[0], options.SignUpRequestDate[1])
	}
	if len(options.SignUpApprovalDate) > 0 {
		q.And("user.sign_up_approval_date BETWEEN ? AND ?", options.SignUpApprovalDate[0], options.SignUpApprovalDate[1])
	}
	if len(options.LastLoginDate) > 0 {
		q.And("user.last_login_date BETWEEN ? AND ?", options.LastLoginDate[0], options.LastLoginDate[1])
	}
	if len(options.Age) > 0 {
		today := time.Now()
		minBirthDate := time.Date(today.Year()-int(options.Age[1]), today.Month(), today.Day(), 0, 0, 0, 0, time.Local)
		maxBirthDate := time.Date(today.Year()-int(options.Age[0]), today.Month(), today.Day(), 23, 59, 59, 999999999, time.Local)
		q.And("private_profile.birthday BETWEEN ? AND ?", minBirthDate, maxBirthDate)
	}
	if len(options.Deposit) > 0 {
		q.And("COALESCE(user_asset_view.deposit, 0) BETWEEN ? AND ?", options.Deposit[0], options.Deposit[1])
	}
	if len(options.InvestOperationAmount) > 0 {
		q.And("COALESCE(user_asset_view.invest_operation_amount, 0) BETWEEN ? AND ?", options.InvestOperationAmount[0], options.InvestOperationAmount[1])
	}
	if len(options.InvestSuccessAmount) > 0 {
		q.And("COALESCE(user_asset_view.invest_success_amount, 0) BETWEEN ? AND ?", options.InvestSuccessAmount[0], options.InvestSuccessAmount[1])
	}
	if len(options.InvestRequestAmount) > 0 {
		q.And("COALESCE(user_asset_view.invest_request_amount, 0) BETWEEN ? AND ?", options.InvestRequestAmount[0], options.InvestRequestAmount[1])
	}
	if len(options.TotalAsset) > 0 {
		q.And("COALESCE(user_asset_view.total_asset, 0) BETWEEN ? AND ?", options.TotalAsset[0], options.TotalAsset[1])
	}
	if len(options.UnpaidAmount) > 0 {
		q.And("COALESCE(user_asset_view.unpaid_amount, 0) BETWEEN ? AND ?", options.UnpaidAmount[0], options.UnpaidAmount[1])
	}

	var receivers []User
	err := q.Find(&receivers)
	if err != nil {
		return nil, err
	}

	return receivers, nil
}

func transformHtml(html string) string {
	html = strings.ReplaceAll(html, "\"", "'")
	html = strings.ReplaceAll(html, "\n", "<br>")
	html = strings.ReplaceAll(html, "\r", "")

	return html
}
