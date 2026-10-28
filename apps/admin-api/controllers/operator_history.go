package controllers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
	"github.com/sirupsen/logrus"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/apirender"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/commoncode"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/echoutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
)

type OperatorHistoryApiController struct{}

func (c OperatorHistoryApiController) Init(g echoswagger.ApiGroup) {
	g.GET("", c.GetAll,
		echoutils.AllowRoles(domain.ROLE_SUPER_ADMIN)).
		AddParamQueryNested(OperatorHistorySearchInput{}).
		AddResponse(http.StatusOK, "管理者ログの照会", domain.OperatorHistory{}, nil).
		SetSummary("管理者ログの照会").
		SetDescription("管理者ログを照会します.").
		SetSecurity("Authorization")

	g.GET("/download", c.Download,
		echoutils.AllowRoles(domain.ROLE_SUPER_ADMIN)).
		SetSummary("管理者ログのExcelファイルダウンロード").
		SetDescription("管理者ログのExcelファイルをダウンロードします.").
		SetSecurity("Authorization")

	g.GET("/menu-codes", c.GetMenuCodeList,
		echoutils.AllowRoles(domain.ROLE_ADMIN)).
		AddResponse(http.StatusOK, "", commoncode.CodeGroup{}, nil).
		AddParamHeader("", "locale", "", false).
		SetSummary("メニュー公開コードの照会").
		SetDescription("メニュー公開コードを照会します.").
		SetSecurity("Authorization")

	g.GET("/log-codes", c.GetLogCodeList,
		echoutils.AllowRoles(domain.ROLE_ADMIN)).
		AddResponse(http.StatusOK, "", commoncode.CodeGroup{}, nil).
		AddParamHeader("", "locale", "", false).
		SetSummary("処理ログ公開コードの照会").
		SetDescription("処理ログ公開コードを照会します.").
		SetSecurity("Authorization")
}

func (OperatorHistoryApiController) GetAll(c echo.Context) error {
	var v OperatorHistorySearchInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if v.MaxResultCount == 0 {
		v.MaxResultCount = DefaultMaxResultCount
	}

	userId, email, err := jwthelper.TokenToUser(c)
	if err != nil {
		return err
	}

	factory.Logger(c.Request().Context()).WithFields(logrus.Fields{
		"sortby":           v.Sortby,
		"order":            v.Order,
		"max_result_count": v.MaxResultCount,
		"skip_count":       v.SkipCount,
	}).Info("SearchStart")

	totalCount, items, err := domain.OperatorHistory{}.GetAll(c.Request().Context(), v.Name, v.Menu, v.State, v.StartDate, v.EndDate, v.Sortby, v.Order, v.SkipCount, v.MaxResultCount)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	err = domain.OperatorHistory{}.Record(c.Request().Context(), userId, email,
		"menu.manage_operating_system.look_up_administrators_access_history", "", c.RealIP(), domain.LOOKUP)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, apirender.ArrayResult{
		Items:      items,
		TotalCount: totalCount,
	})
}

func (OperatorHistoryApiController) Download(c echo.Context) error {

	var v OperatorHistorySearchInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if v.MaxResultCount == 0 {
		v.MaxResultCount = DefaultMaxResultCount
	}

	userId, email, err := jwthelper.TokenToUser(c)
	if err != nil {
		return err
	}

	factory.Logger(c.Request().Context()).WithFields(logrus.Fields{
		"sortby":           v.Sortby,
		"order":            v.Order,
		"max_result_count": v.MaxResultCount,
		"skip_count":       v.SkipCount,
	}).Info("SearchStart")

	err = domain.OperatorHistory{}.DownloadCsv(c.Request().Context(), c.Response().Writer, v.Name, v.Menu, v.State, v.StartDate, v.EndDate, v.Sortby, v.Order, v.SkipCount, v.MaxResultCount)
	if err != nil {
		return factory.ErrorFailDownloadCSV.New(err)
	}

	err = domain.OperatorHistory{}.Record(c.Request().Context(), userId, email,
		"menu.manage_operating_system.look_up_administrators_access_history", "", c.RealIP(), domain.DOWNLOAD)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return nil
}

func (OperatorHistoryApiController) GetMenuCodeList(c echo.Context) error {
	lang := c.Request().Header.Get("locale")
	if lang == "" {
		lang = "ja"
	}
	loc := lang

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

	return apirender.RenderSucc(c, http.StatusOK, domain.MenuGroup{Codes: codes})
}

func (OperatorHistoryApiController) GetLogCodeList(c echo.Context) error {
	lang := c.Request().Header.Get("locale")
	if lang == "" {
		lang = "ja"
	}
	loc := lang

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

	return apirender.RenderSucc(c, http.StatusOK, domain.MenuGroup{Codes: codes})
}
