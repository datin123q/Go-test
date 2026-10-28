package domain

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/commoncode"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/csvutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/modelhelper"
)

type MenuGroup struct {
	Codes []commoncode.Code `json:"codes"`
}

type OPERATOR_STATE string

const (
	OPERATOR_STATE_CODE                = "OPERATOR_STATE"
	LOGIN               OPERATOR_STATE = "LOGIN"    // ログイン
	LOGOUT              OPERATOR_STATE = "LOGOUT"   // ログアウト
	REGISTER            OPERATOR_STATE = "REGISTER" // 登録
	LOOKUP              OPERATOR_STATE = "LOOKUP"   // 照会
	CHANGE              OPERATOR_STATE = "CHANGE"   // 変更
	DELETE              OPERATOR_STATE = "DELETE"   // 削除
	UPLOAD              OPERATOR_STATE = "UPLOAD"   // アップロード
	DOWNLOAD            OPERATOR_STATE = "DOWNLOAD" // ダウンロード
)

type OperatorHistory struct {
	Id      int64              `json:"id"`
	UserId  int64              `json:"user_id" xorm:"not null"`
	Name    string             `json:"name" xorm:"<-"`
	Email   string             `json:"email"`
	Menu    string             `json:"menu" xorm:"not null"`
	Ip      string             `json:"ip" xorm:"not null"`
	Title   string             `json:"title"`
	State   OPERATOR_STATE     `json:"state" xorm:"not null"`
	Created jsonutils.JSONTime `json:"created" xorm:"created"`
}

type OperatorHistoryRowRequest struct {
	Ts    int64                `json:"ts" description:"날짜 타임스탬프"`
	Actor OperatorHistoryActor `json:"actor" description:"행위자"`
	Path  string               `json:"path" description:"경로"`
	IP    string               `json:"ip" description:"IP"`
	Log   OperatorHistoryLog   `json:"log" description:"로그"`
}

type OperatorHistoryRowResponse struct {
	Date      string               `json:"date" description:"날짜"`
	Actor     OperatorHistoryActor `json:"actor" description:"행위자"`
	Operation string               `json:"operation" description:"작업"`
	IP        string               `json:"ip" description:"IP"`
	Log       OperatorHistoryLog   `json:"log" description:"로그"`
}

type OperatorHistoryLog struct {
	LogType   string    `json:"log_type"`
	RequestID string    `json:"request_id"`
	RemoteIP  string    `json:"remote_ip"`
	Host      string    `json:"host"`
	Path      string    `json:"path"`
	URI       string    `json:"uri"`
	Method    string    `json:"method"`
	Referer   string    `json:"referer"`
	UserAgent string    `json:"user_agent"`
	Status    int       `json:"status"`
	Latency   string    `json:"latency"`
	BytesIn   string    `json:"bytes_in"`
	BytesOut  int64     `json:"bytes_out"`
	Date      time.Time `json:"date"`
}

func (l OperatorHistoryLog) ToOperatorHistoryLog(inf interface{}) (OperatorHistoryLog, error) {
	var log OperatorHistoryLog
	switch inf.(type) {
	case map[string]interface{}:
		infJson, err := json.Marshal(inf)
		if err != nil {
			return log, err
		}
		err = json.Unmarshal(infJson, &log)
		if err != nil {
			return log, err
		}
	default:
		return log, fmt.Errorf("not supported type: %s", reflect.TypeOf(inf).String())
	}
	return log, nil
}

func (l OperatorHistoryLog) String() string {
	return fmt.Sprintf("{LogType:%s,RequestID:%s,RemoteIP:%s,Host:%s,Operation:%s,URI:%s,Method:%s,Referer:%s,UserAgent:%s,Status:%d,Latency:%s,BytesIn:%s,BytesOut:%d,Date:%s}",
		l.LogType, l.RequestID, l.RemoteIP, l.Host, l.Path, l.URI, l.Method, l.Referer, l.UserAgent, l.Status, l.Latency, l.BytesIn, l.BytesOut, l.Date)
}

type OperatorHistoryActor struct {
	Id    int64
	Name  string
	Email string
}

func (a OperatorHistoryActor) ToOperatorHistoryActor(inf interface{}) (OperatorHistoryActor, error) {
	switch inf.(type) {
	case map[string]interface{}:
		actorMap := inf.(map[string]interface{})
		actor := OperatorHistoryActor{
			Id:    int64(actorMap["ID"].(float64)),
			Name:  actorMap["NAME"].(string),
			Email: actorMap["EMAIL"].(string),
		}
		return actor, nil
	default:
		return OperatorHistoryActor{}, fmt.Errorf("not supported type: %s", reflect.TypeOf(inf).String())
	}
}

func (a OperatorHistoryActor) String() string {
	return fmt.Sprintf("{Id:%d,Name:%s,Email:%s}", a.Id, a.Name, a.Email)
}

func (o OperatorHistory) TableName() string {
	return "operator_history"
}

func (o OperatorHistory) SelectAll() string {
	return o.TableName() + ".*"
}

const (
	OPERATOR_HISTORY_STRINGS string = `
		operator_history.id, 
		operator_history.user_id,
		operator_history.menu,
		operator_history.ip,
		operator_history.title,
		operator_history.state,
		operator_history.created,
		admin_user.name as name,
		admin_user.email as email
	`
)

func (OperatorHistory) GetOperatorStateCodeGroup(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0, 0)

	codes = append(codes, commoncode.Code{Code: string(LOGIN), Name: i18nutils.Translate("commoncode.OPERATOR_STATE.LOGIN", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(LOGOUT), Name: i18nutils.Translate("commoncode.OPERATOR_STATE.LOGOUT", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(REGISTER), Name: i18nutils.Translate("commoncode.OPERATOR_STATE.REGISTER", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(CHANGE), Name: i18nutils.Translate("commoncode.OPERATOR_STATE.CHANGE", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(DELETE), Name: i18nutils.Translate("commoncode.OPERATOR_STATE.DELETE", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(DOWNLOAD), Name: i18nutils.Translate("commoncode.OPERATOR_STATE.DOWNLOAD", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(LOOKUP), Name: i18nutils.Translate("commoncode.OPERATOR_STATE.LOOKUP", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(UPLOAD), Name: i18nutils.Translate("commoncode.OPERATOR_STATE.UPLOAD", loc), Order: len(codes) + 1})

	return commoncode.CodeGroup{
		GroupCode: OPERATOR_STATE_CODE,
		Codes:     codes,
	}
}

// GetLogCodeList - メニューコード一覧（公用コード機能実装前に一時的に保持されるコード）
func (OperatorHistory) GetMenuCodes(loc string) []commoncode.Code {
	codes := make([]commoncode.Code, 0)
	codes = append(codes, commoncode.Code{Code: "menu.dashboard", Name: i18nutils.Translate("menu.dashboard", loc), Order: len(codes) + 1})

	codes = append(codes, commoncode.Code{Code: "menu.manage_accounting.3_manage_accounting", Name: i18nutils.Translate("menu.manage_accounting.3_manage_accounting", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_accounting.deposit_detail", Name: i18nutils.Translate("menu.manage_accounting.deposit_detail", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_accounting.deposit_list", Name: i18nutils.Translate("menu.manage_accounting.deposit_list", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_accounting.excess_and_shortage", Name: i18nutils.Translate("menu.manage_accounting.excess_and_shortage", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_accounting.excess_and_shortage_detail", Name: i18nutils.Translate("menu.manage_accounting.excess_and_shortage_detail", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_accounting.excess_and_shortage_list", Name: i18nutils.Translate("menu.manage_accounting.excess_and_shortage_list", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_accounting.payment_detail", Name: i18nutils.Translate("menu.manage_accounting.payment_detail", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_accounting.payment_list", Name: i18nutils.Translate("menu.manage_accounting.payment_list", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_accounting.receiving_amount", Name: i18nutils.Translate("menu.manage_accounting.receiving_amount", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_accounting.receiving_amount_list", Name: i18nutils.Translate("menu.manage_accounting.receiving_amount_list", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_accounting.withdrawal_manage", Name: i18nutils.Translate("menu.manage_accounting.withdrawal_manage", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_accounting.withdrawal_request", Name: i18nutils.Translate("menu.manage_accounting.withdrawal_request", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_accounting.withdrawal_request_detail", Name: i18nutils.Translate("menu.manage_accounting.withdrawal_request_detail", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_accounting.manage_withdrawal_close", Name: i18nutils.Translate("menu.manage_accounting.manage_withdrawal_close", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_accounting.manage_withdrawal_download", Name: i18nutils.Translate("menu.manage_accounting.manage_withdrawal_download", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_accounting.trust_info.list", Name: i18nutils.Translate("menu.manage_accounting.trust_info.list", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_accounting.send_manage", Name: i18nutils.Translate("menu.manage_accounting.send_manage", loc), Order: len(codes) + 1})
	// 抽選一覧
	codes = append(codes, commoncode.Code{Code: "menu.manage_accounting.project_lottery.list", Name: i18nutils.Translate("menu.manage_accounting.project_lottery.list", loc), Order: len(codes) + 1})


	codes = append(codes, commoncode.Code{Code: "menu.manage_mail_sms.5_manage_mail_sms", Name: i18nutils.Translate("menu.manage_mail_sms.5_manage_mail_sms", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_mail_sms.manage_emails", Name: i18nutils.Translate("menu.manage_mail_sms.manage_emails", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_mail_sms.manage_emails_manage_detail", Name: i18nutils.Translate("menu.manage_mail_sms.manage_emails_manage_detail", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_mail_sms.manage_emails_manage_send", Name: i18nutils.Translate("menu.manage_mail_sms.manage_emails_manage_send", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_mail_sms.manage_emails_send", Name: i18nutils.Translate("menu.manage_mail_sms.manage_emails_send", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_mail_sms.manage_sms", Name: i18nutils.Translate("menu.manage_mail_sms.manage_sms", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_mail_sms.manage_sms_manage_detail", Name: i18nutils.Translate("menu.manage_mail_sms.manage_sms_manage_detail", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_mail_sms.manage_sms_manage_send", Name: i18nutils.Translate("menu.manage_mail_sms.manage_sms_manage_send", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_mail_sms.manage_sms_send", Name: i18nutils.Translate("menu.manage_mail_sms.manage_sms_send", loc), Order: len(codes) + 1})

	codes = append(codes, commoncode.Code{Code: "menu.manage_main.contact_detail", Name: i18nutils.Translate("menu.manage_main.contact_detail", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_main.contact_manage", Name: i18nutils.Translate("menu.manage_main.contact_manage", loc), Order: len(codes) + 1})

	codes = append(codes, commoncode.Code{Code: "menu.manage_mynumber.8_manage", Name: i18nutils.Translate("menu.manage_mynumber.8_manage", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_mynumber_list", Name: i18nutils.Translate("menu.manage_mynumber_list", loc), Order: len(codes) + 1})

	codes = append(codes, commoncode.Code{Code: "menu.manage_mypage.notice_manage", Name: i18nutils.Translate("menu.manage_mypage.notice_manage", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_mypage.notice_send.detail", Name: i18nutils.Translate("menu.manage_mypage.notice_send.detail", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_mypage.notice_send.list", Name: i18nutils.Translate("menu.manage_mypage.notice_send.list", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_mypage.notice_send.notice", Name: i18nutils.Translate("menu.manage_mypage.notice_send.notice", loc), Order: len(codes) + 1})

	codes = append(codes, commoncode.Code{Code: "menu.manage_operating_system.6_manage_operating_system", Name: i18nutils.Translate("menu.manage_operating_system.6_manage_operating_system", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_operating_system.display_code", Name: i18nutils.Translate("menu.manage_operating_system.display_code", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_operating_system.look_up_administrators_access_history", Name: i18nutils.Translate("menu.manage_operating_system.look_up_administrators_access_history", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_operating_system.manage_newsletter_subscriber", Name: i18nutils.Translate("menu.manage_operating_system.manage_newsletter_subscriber", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_operating_system.mange_administrators", Name: i18nutils.Translate("menu.manage_operating_system.mange_administrators", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_operating_system.mange_administrators_detail", Name: i18nutils.Translate("menu.manage_operating_system.mange_administrators_detail", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_operating_system.mange_administrators_list", Name: i18nutils.Translate("menu.manage_operating_system.mange_administrators_list", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_operating_system.mange_administrators_register", Name: i18nutils.Translate("menu.manage_operating_system.mange_administrators_register", loc), Order: len(codes) + 1})

	codes = append(codes, commoncode.Code{Code: "menu.manage_printout.7_manage", Name: i18nutils.Translate("menu.manage_printout.7_manage", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_printout.audit_response", Name: i18nutils.Translate("menu.manage_printout.audit_response", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_printout.for_investors", Name: i18nutils.Translate("menu.manage_printout.for_investors", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_printout.for_tax_office", Name: i18nutils.Translate("menu.manage_printout.for_tax_office", loc), Order: len(codes) + 1})

	codes = append(codes, commoncode.Code{Code: "menu.manage_product.2_manage_product", Name: i18nutils.Translate("menu.manage_product.2_manage_product", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_product.developer", Name: i18nutils.Translate("menu.manage_product.developer", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_product.developer_detail", Name: i18nutils.Translate("menu.manage_product.developer_detail", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_product.developer_list", Name: i18nutils.Translate("menu.manage_product.developer_list", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_product.developer_register", Name: i18nutils.Translate("menu.manage_product.developer_register", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_product.funding_request", Name: i18nutils.Translate("menu.manage_product.funding_request", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_product.funding_request_detail", Name: i18nutils.Translate("menu.manage_product.funding_request_detail", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_product.funding_request_list", Name: i18nutils.Translate("menu.manage_product.funding_request_list", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_product.product", Name: i18nutils.Translate("menu.manage_product.product", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_product.product_detail", Name: i18nutils.Translate("menu.manage_product.product_detail", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_product.product_list", Name: i18nutils.Translate("menu.manage_product.product_list", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_product.product_register", Name: i18nutils.Translate("menu.manage_product.product_register", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_product.tab1", Name: i18nutils.Translate("menu.manage_product.tab1", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_product.tab2_waiting", Name: i18nutils.Translate("menu.manage_product.tab2_waiting", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_product.funding_requester", Name: i18nutils.Translate("menu.manage_product.funding_requester", loc), Order: len(codes) + 1})

	codes = append(codes, commoncode.Code{Code: "menu.manage_project.request", Name: i18nutils.Translate("menu.manage_project.request", loc), Order: len(codes) + 1})

	codes = append(codes, commoncode.Code{Code: "menu.manage_qna.detail", Name: i18nutils.Translate("menu.manage_qna.detail", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_qna.list", Name: i18nutils.Translate("menu.manage_qna.list", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_qna.title", Name: i18nutils.Translate("menu.manage_qna.title", loc), Order: len(codes) + 1})

	codes = append(codes, commoncode.Code{Code: "menu.manage_system_data_manage", Name: i18nutils.Translate("menu.manage_system_data_manage", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_terms_manage", Name: i18nutils.Translate("menu.manage_terms_manage", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_terms_manage_detail", Name: i18nutils.Translate("menu.manage_terms_manage_detail", loc), Order: len(codes) + 1})

	codes = append(codes, commoncode.Code{Code: "menu.manage_users.4_manage_users", Name: i18nutils.Translate("menu.manage_users.4_manage_users", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_users.corporate_investors", Name: i18nutils.Translate("menu.manage_users.corporate_investors", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_users.corporate_investors_detail", Name: i18nutils.Translate("menu.manage_users.corporate_investors_detail", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_users.corporate_investors_list", Name: i18nutils.Translate("menu.manage_users.corporate_investors_list", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_users.corporate_investors_shares_detail", Name: i18nutils.Translate("menu.manage_users.corporate_investors_shares_detail", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_users.individual_investors", Name: i18nutils.Translate("menu.manage_users.individual_investors", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_users.individual_investors_detail", Name: i18nutils.Translate("menu.manage_users.individual_investors_detail", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_users.individual_investors_list", Name: i18nutils.Translate("menu.manage_users.individual_investors_list", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_users.pre_corp_investors", Name: i18nutils.Translate("menu.manage_users.pre_corp_investors", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_users.pre_corp_investors_list", Name: i18nutils.Translate("menu.manage_users.pre_corp_investors_list", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_users.pre_indv_investors", Name: i18nutils.Translate("menu.manage_users.pre_indv_investors", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_users.pre_indv_investors_list", Name: i18nutils.Translate("menu.manage_users.pre_indv_investors_list", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_users.user_access_log", Name: i18nutils.Translate("menu.manage_users.user_access_log", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_users.user_access_log_list", Name: i18nutils.Translate("menu.manage_users.user_access_log_list", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_users.user_list_download", Name: i18nutils.Translate("menu.manage_users.user_list_download", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_users.referral_code_issuance", Name: i18nutils.Translate("menu.manage_users.referral_code_issuance", loc), Order: len(codes) + 1})

	codes = append(codes, commoncode.Code{Code: "menu.manage_virtualaccount_manage", Name: i18nutils.Translate("menu.manage_virtualaccount_manage", loc), Order: len(codes) + 1})

	codes = append(codes, commoncode.Code{Code: "menu.manage_website.1_manage_website", Name: i18nutils.Translate("menu.manage_website.1_manage_website", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_website.manage_faq", Name: i18nutils.Translate("menu.manage_website.manage_faq", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_website.manage_faq_detail", Name: i18nutils.Translate("menu.manage_website.manage_faq_detail", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_website.manage_faq_list", Name: i18nutils.Translate("menu.manage_website.manage_faq_list", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_website.manage_faq_register", Name: i18nutils.Translate("menu.manage_website.manage_faq_register", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_website.manage_column", Name: i18nutils.Translate("menu.manage_website.manage_column", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_website.manage_column_detail", Name: i18nutils.Translate("menu.manage_website.manage_column_detail", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_website.manage_column_list", Name: i18nutils.Translate("menu.manage_website.manage_column_list", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_website.manage_main_banner", Name: i18nutils.Translate("menu.manage_website.manage_main_banner", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_website.manage_news", Name: i18nutils.Translate("menu.manage_website.manage_news", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_website.manage_news_detail", Name: i18nutils.Translate("menu.manage_website.manage_news_detail", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_website.manage_news_list", Name: i18nutils.Translate("menu.manage_website.manage_news_list", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_website.manage_news_register", Name: i18nutils.Translate("menu.manage_website.manage_news_register", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_website.manage_product_recommendation", Name: i18nutils.Translate("menu.manage_website.manage_product_recommendation", loc), Order: len(codes) + 1})

	codes = append(codes, commoncode.Code{Code: "commoncode.INVEST_STATUS.CANCEL_INVEST", Name: i18nutils.Translate("commoncode.INVEST_STATUS.CANCEL_INVEST", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "commoncode.INVEST_STATUS.SUCCESS_INVEST", Name: i18nutils.Translate("commoncode.INVEST_STATUS.SUCCESS_INVEST", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "commoncode.INVEST_STATUS.FAIL_INVEST", Name: i18nutils.Translate("commoncode.INVEST_STATUS.FAIL_INVEST", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "commoncode.LOG_CONTENT_TYPE.ADMIN_ROLE_DELETED", Name: i18nutils.Translate("commoncode.LOG_CONTENT_TYPE.ADMIN_ROLE_DELETED", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "commoncode.LOG_CONTENT_TYPE.CONTRACT_BATCH_EXECUTE", Name: i18nutils.Translate("commoncode.LOG_CONTENT_TYPE.CONTRACT_BATCH_EXECUTE", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "commoncode.LOG_CONTENT_TYPE.FUNDING_REQUEST_ACCEPT", Name: i18nutils.Translate("commoncode.LOG_CONTENT_TYPE.FUNDING_REQUEST_ACCEPT", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "commoncode.LOG_CONTENT_TYPE.GMO_API_REQUEST", Name: i18nutils.Translate("commoncode.LOG_CONTENT_TYPE.GMO_API_REQUEST", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "commoncode.LOG_CONTENT_TYPE.LOGIN", Name: i18nutils.Translate("commoncode.LOG_CONTENT_TYPE.LOGIN", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "commoncode.LOG_CONTENT_TYPE.PAYMENT_STATEMENT_BATCH_EXECUTE", Name: i18nutils.Translate("commoncode.LOG_CONTENT_TYPE.PAYMENT_STATEMENT_BATCH_EXECUTE", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "commoncode.LOG_CONTENT_TYPE.VA_BATCH_EXECUTE", Name: i18nutils.Translate("commoncode.LOG_CONTENT_TYPE.VA_BATCH_EXECUTE", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "commoncode.LOG_CONTENT_TYPE.EXECESS_AND_SHORTAGE_TARGET_EXECUTE", Name: i18nutils.Translate("commoncode.LOG_CONTENT_TYPE.EXECESS_AND_SHORTAGE_TARGET_EXECUTE", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "waiting.manual_investment.registration", Name: i18nutils.Translate("waiting.manual_investment.registration", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "commoncode.FAQ_CATEGORY_TYPE.FAQ_CATEGORY_ACCOUNT_REGISTRATION", Name: i18nutils.Translate("commoncode.FAQ_CATEGORY_TYPE.FAQ_CATEGORY_ACCOUNT_REGISTRATION", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "commoncode.FAQ_CATEGORY_TYPE.FAQ_CATEGORY_PUT_TAKE_MONEY", Name: i18nutils.Translate("commoncode.FAQ_CATEGORY_TYPE.FAQ_CATEGORY_PUT_TAKE_MONEY", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "commoncode.FAQ_CATEGORY_TYPE.FAQ_CATEGORY_INVESTING", Name: i18nutils.Translate("commoncode.FAQ_CATEGORY_TYPE.FAQ_CATEGORY_INVESTING", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "commoncode.FAQ_CATEGORY_TYPE.FAQ_CATEGORY_FUNDS", Name: i18nutils.Translate("commoncode.FAQ_CATEGORY_TYPE.FAQ_CATEGORY_FUNDS", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "commoncode.FAQ_CATEGORY_TYPE.FAQ_CATEGORY_REPAY_DISTRIBUTION", Name: i18nutils.Translate("commoncode.FAQ_CATEGORY_TYPE.FAQ_CATEGORY_REPAY_DISTRIBUTION", loc), Order: len(codes) + 1})

	return codes
}

// GetLogCodeList - ログコード一覧（公用コード機能実装前に一時的に保持されるコード）
func (OperatorHistory) GetLogCodeList(loc string) []commoncode.Code {
	codes := make([]commoncode.Code, 0)
	codes = append(codes, commoncode.Code{Code: "menu.dashboard", Name: i18nutils.Translate("menu.dashboard", loc), Order: len(codes) + 1})

	codes = append(codes, commoncode.Code{Code: "menu.manage_accounting.deposit_detail", Name: i18nutils.Translate("menu.manage_accounting.deposit_detail", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_accounting.deposit_list", Name: i18nutils.Translate("menu.manage_accounting.deposit_list", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_accounting.excess_and_shortage_detail", Name: i18nutils.Translate("menu.manage_accounting.excess_and_shortage_detail", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_accounting.excess_and_shortage_list", Name: i18nutils.Translate("menu.manage_accounting.excess_and_shortage_list", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_accounting.payment_list", Name: i18nutils.Translate("menu.manage_accounting.payment_list", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_accounting.receiving_amount", Name: i18nutils.Translate("menu.manage_accounting.receiving_amount", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_accounting.receiving_amount_list", Name: i18nutils.Translate("menu.manage_accounting.receiving_amount_list", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_accounting.trust_info.list", Name: i18nutils.Translate("menu.manage_accounting.trust_info.list", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_accounting.manage_withdrawal_close", Name: i18nutils.Translate("menu.manage_accounting.manage_withdrawal_close", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_accounting.manage_withdrawal_download", Name: i18nutils.Translate("menu.manage_accounting.manage_withdrawal_download", loc), Order: len(codes) + 1})

	codes = append(codes, commoncode.Code{Code: "menu.manage_mail_sms.manage_emails", Name: i18nutils.Translate("menu.manage_mail_sms.manage_emails", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_mail_sms.manage_emails_manage_detail", Name: i18nutils.Translate("menu.manage_mail_sms.manage_emails_manage_detail", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_mail_sms.manage_emails_manage_send", Name: i18nutils.Translate("menu.manage_mail_sms.manage_emails_manage_send", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_mail_sms.manage_sms", Name: i18nutils.Translate("menu.manage_mail_sms.manage_sms", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_mail_sms.manage_sms_manage_detail", Name: i18nutils.Translate("menu.manage_mail_sms.manage_sms_manage_detail", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_mail_sms.manage_sms_manage_send", Name: i18nutils.Translate("menu.manage_mail_sms.manage_sms_manage_send", loc), Order: len(codes) + 1})

	codes = append(codes, commoncode.Code{Code: "menu.manage_mynumber.8_manage", Name: i18nutils.Translate("menu.manage_mynumber.8_manage", loc), Order: len(codes) + 1})

	codes = append(codes, commoncode.Code{Code: "menu.manage_operating_system.display_code", Name: i18nutils.Translate("menu.manage_operating_system.display_code", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_operating_system.look_up_administrators_access_history", Name: i18nutils.Translate("menu.manage_operating_system.look_up_administrators_access_history", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_operating_system.manage_newsletter_subscriber", Name: i18nutils.Translate("menu.manage_operating_system.manage_newsletter_subscriber", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_operating_system.mange_administrators_detail", Name: i18nutils.Translate("menu.manage_operating_system.mange_administrators_detail", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_operating_system.mange_administrators_list", Name: i18nutils.Translate("menu.manage_operating_system.mange_administrators_list", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_operating_system.mange_administrators_register", Name: i18nutils.Translate("menu.manage_operating_system.mange_administrators_register", loc), Order: len(codes) + 1})

	codes = append(codes, commoncode.Code{Code: "menu.manage_product.developer_detail", Name: i18nutils.Translate("menu.manage_product.developer_detail", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_product.developer_list", Name: i18nutils.Translate("menu.manage_product.developer_list", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_product.developer_register", Name: i18nutils.Translate("menu.manage_product.developer_register", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_product.funding_request", Name: i18nutils.Translate("menu.manage_product.funding_request", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_product.funding_request_detail", Name: i18nutils.Translate("menu.manage_product.funding_request_detail", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_product.funding_request_list", Name: i18nutils.Translate("menu.manage_product.funding_request_list", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_product.product_detail", Name: i18nutils.Translate("menu.manage_product.product_detail", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_product.product_list", Name: i18nutils.Translate("menu.manage_product.product_list", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_product.product_register", Name: i18nutils.Translate("menu.manage_product.product_register", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_product.tab1", Name: i18nutils.Translate("menu.manage_product.tab1", loc), Order: len(codes) + 1})

	codes = append(codes, commoncode.Code{Code: "menu.manage_qna.detail", Name: i18nutils.Translate("menu.manage_qna.detail", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_qna.list", Name: i18nutils.Translate("menu.manage_qna.list", loc), Order: len(codes) + 1})

	codes = append(codes, commoncode.Code{Code: "menu.manage_users.corporate_investors_detail", Name: i18nutils.Translate("menu.manage_users.corporate_investors_detail", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_users.corporate_investors_list", Name: i18nutils.Translate("menu.manage_users.corporate_investors_list", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_users.individual_investors_detail", Name: i18nutils.Translate("menu.manage_users.individual_investors_detail", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_users.individual_investors_list", Name: i18nutils.Translate("menu.manage_users.individual_investors_list", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_users.pre_corp_investors", Name: i18nutils.Translate("menu.manage_users.pre_corp_investors", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_users.pre_indv_investors", Name: i18nutils.Translate("menu.manage_users.pre_indv_investors", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_users.user_access_log_list", Name: i18nutils.Translate("menu.manage_users.user_access_log_list", loc), Order: len(codes) + 1})

	codes = append(codes, commoncode.Code{Code: "menu.manage_virtualaccount_manage", Name: i18nutils.Translate("menu.manage_virtualaccount_manage", loc), Order: len(codes) + 1})

	codes = append(codes, commoncode.Code{Code: "menu.manage_website.manage_faq_detail", Name: i18nutils.Translate("menu.manage_website.manage_faq_detail", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_website.manage_faq_list", Name: i18nutils.Translate("menu.manage_website.manage_faq_list", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_website.manage_faq_register", Name: i18nutils.Translate("menu.manage_website.manage_faq_register", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_website.manage_column", Name: i18nutils.Translate("menu.manage_website.manage_column", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_website.manage_column_detail", Name: i18nutils.Translate("menu.manage_website.manage_column_detail", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_website.manage_column_list", Name: i18nutils.Translate("menu.manage_website.manage_column_list", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_website.manage_main_banner", Name: i18nutils.Translate("menu.manage_website.manage_main_banner", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_website.manage_news_detail", Name: i18nutils.Translate("menu.manage_website.manage_news_detail", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "menu.manage_website.manage_product_recommendation", Name: i18nutils.Translate("menu.manage_website.manage_product_recommendation", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "commoncode.INVEST_STATUS.CANCEL_INVEST", Name: i18nutils.Translate("commoncode.INVEST_STATUS.CANCEL_INVEST", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "commoncode.INVEST_STATUS.SUCCESS_INVEST", Name: i18nutils.Translate("commoncode.INVEST_STATUS.SUCCESS_INVEST", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "commoncode.INVEST_STATUS.FAIL_INVEST", Name: i18nutils.Translate("commoncode.INVEST_STATUS.FAIL_INVEST", loc), Order: len(codes) + 1})

	codes = append(codes, commoncode.Code{Code: "commoncode.LOG_CONTENT_TYPE.FUNDING_REQUEST_ACCEPT", Name: i18nutils.Translate("commoncode.LOG_CONTENT_TYPE.FUNDING_REQUEST_ACCEPT", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "commoncode.LOG_CONTENT_TYPE.VA_BATCH_EXECUTE", Name: i18nutils.Translate("commoncode.LOG_CONTENT_TYPE.VA_BATCH_EXECUTE", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "commoncode.LOG_CONTENT_TYPE.EXECESS_AND_SHORTAGE_TARGET_EXECUTE", Name: i18nutils.Translate("commoncode.LOG_CONTENT_TYPE.EXECESS_AND_SHORTAGE_TARGET_EXECUTE", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: "commoncode.LOG_CONTENT_TYPE.PAYMENT_STATEMENT_BATCH_EXECUTE", Name: i18nutils.Translate("commoncode.LOG_CONTENT_TYPE.PAYMENT_STATEMENT_BATCH_EXECUTE", loc), Order: len(codes) + 1})

	return codes
}

var OperatorHistoryCSVHeaders = []string{
	"model.operator_history.created", // 処理日時
	"model.operator_history.name",    // 管理者名（メールアドレス）
	"model.operator_history.menu",    // ログ種別
	"model.operator_history.log",     // ログ
	"model.operator_history.ip",      // IP
}

func (OperatorHistory) GetAll(ctx context.Context, name, menu, state *string, startDate, endDate string, sortby, order string, offset, limit int) (int64, []OperatorHistory, error) {
	q := factory.DB(ctx).Table("operator_history").Select("operator_history.id, operator_history.user_id, operator_history.menu, operator_history.ip, operator_history.title, operator_history.state, operator_history.created, admin_user.name as name, admin_user.email as email").
		Join("INNER", "admin_user", "operator_history.user_id = admin_user.id")

	if err := modelhelper.SetSortOrder(q, sortby, order); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SearchLikeQuery(q, "admin_user", "name", name); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SearchEqQuery(q, "operator_history", "menu", menu); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SearchEqQuery(q, "operator_history", "state", state); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SetBetweenDate(q, "operator_history.created", startDate, endDate); err != nil {
		factory.Logger(ctx).Error(err)
	}

	var items []OperatorHistory
	var totalCount int64
	totalCount, err := q.Limit(limit, offset).FindAndCount(&items)
	if err != nil {
		return 0, nil, err
	}

	return totalCount, items, nil
}

func (o *OperatorHistory) Create(ctx context.Context) (int64, error) {
	return factory.DB(ctx).Insert(o)
}

func (OperatorHistory) GetById(ctx context.Context, id int64) (*OperatorHistory, error) {
	var item OperatorHistory
	if has, err := factory.DB(ctx).ID(id).Get(&item); err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}

	return &item, nil
}

func (OperatorHistory) Delete(ctx context.Context, id int64) error {
	_, err := factory.DB(ctx).ID(id).Delete(&OperatorHistory{})
	return err
}

func (OperatorHistory) Record(ctx context.Context, userId int64, email, menu, title, ip string, state OPERATOR_STATE) error {
	history := OperatorHistory{
		UserId: userId,
		Email:  email,
		Menu:   menu,
		Title:  title,
		Ip:     ip,
		State:  state,
	}

	_, err := history.Create(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (OperatorHistory) PrepareAndReturnCsv(items []OperatorHistory) [][]string {
	type history struct {
		Created jsonutils.JSONTime `json:"created" i18n:"model.operator_history.created"`
		Email   string             `json:"email" i18n:"model.operator_history.name"`
		Menu    string             `json:"menu" i18n:"model.operator_history.menu"`
		State   string             `json:"state" i18n:"model.operator_history.state"`
		Ip      string             `json:"ip" i18n:"model.operator_history.ip"`
	}

	var histories []history
	for _, item := range items {
		h := history{
			Email:   item.Email,
			Menu:    i18nutils.DefaultTranslate(item.Menu),
			Ip:      item.Ip,
			State:   i18nutils.DefaultTranslate("commoncode.OPERATOR_STATE." + string(item.State)),
			Created: item.Created,
		}

		if item.State == LOOKUP {
			h.State = h.Menu + " " + h.State
		} else if len(item.Title) > 0 {
			h.State = item.Title + " " + h.State
		}

		histories = append(histories, h)
	}

	cols := factory.ReflectJsonFields(reflect.ValueOf(&history{}).Elem())
	if len(cols) <= 0 {
		return nil
	}

	records := [][]string{}
	records = append(records, cols)

	for _, v := range histories {
		records = append(records, factory.DataWriteToCsv(reflect.ValueOf(&v).Elem()))
	}

	return records
}

func (OperatorHistory) DownloadCsv(ctx context.Context, w http.ResponseWriter, name, menu, state *string, startDate, endDate string, sortby, order string, offset, limit int) error {
	q := factory.DB(ctx).Table("operator_history").Select("operator_history.id, operator_history.user_id, operator_history.menu, operator_history.ip, operator_history.title, operator_history.state, operator_history.created, admin_user.name as name, admin_user.email as email").
		Join("INNER", "admin_user", "operator_history.user_id = admin_user.id")

	if err := modelhelper.SetSortOrder(q, sortby, order); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SearchLikeQuery(q, "admin_user", "name", name); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SearchEqQuery(q, "operator_history", "menu", menu); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SearchEqQuery(q, "operator_history", "state", state); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SetBetweenDate(q, "operator_history.created", startDate, endDate); err != nil {
		factory.Logger(ctx).Error(err)
	}

	var items []OperatorHistory
	_, err := q.FindAndCount(&items)
	if err != nil {
		return err
	}

	records := OperatorHistory{}.PrepareAndReturnCsv(items)
	csvutils.WriteHeader(w, "energyshares_operator_history.csv")

	return csv.NewWriter(w).WriteAll(records)
}
