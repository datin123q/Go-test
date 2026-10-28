package domain

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"sort"
	"strconv"
	"time"

	"github.com/gofrs/uuid"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/commoncode"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/csvutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/modelhelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/pathutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/pdfgen"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
	"github.com/jung-kurt/gofpdf"
	pdfcpuapi "github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"xorm.io/builder"
	"xorm.io/xorm"
)

func init() {
}

type PROJECT_CATEGORY string

const (
	PROJECT_CATEGORY_CODE                  = "PROJECT_CATEGORY"
	SOLAR                 PROJECT_CATEGORY = "SOLAR"          // 太陽光
	LOGISTICS             PROJECT_CATEGORY = "LOGISTICS"      // 物流
	FOOD_SERVICE          PROJECT_CATEGORY = "FOOD_SERVICE"   // 飲食
	PRECIOUS_METAL        PROJECT_CATEGORY = "PRECIOUS_METAL" // 貴金属
	MEDICAL               PROJECT_CATEGORY = "MEDICAL"        // 医療
	RETAIL                PROJECT_CATEGORY = "RETAIL"         // 店舗
	REAL_ESTATE           PROJECT_CATEGORY = "REAL_ESTATE"    // 不動産
	BESS                  PROJECT_CATEGORY = "BESS"           // 蓄電池
	EVENT                 PROJECT_CATEGORY = "EVENT"          // イベント
	FACTORING             PROJECT_CATEGORY = "FACTORING"      // ファクタリング
	ETC                   PROJECT_CATEGORY = "ETC"            // その他
)

func (Project) GetCategoryCodeGroup(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(SOLAR), Name: i18nutils.Translate("commoncode.PROJECT_CATEGORY.SOLAR", loc), Order: 1})
	codes = append(codes, commoncode.Code{Code: string(LOGISTICS), Name: i18nutils.Translate("commoncode.PROJECT_CATEGORY.LOGISTICS", loc), Order: 2})
	codes = append(codes, commoncode.Code{Code: string(FOOD_SERVICE), Name: i18nutils.Translate("commoncode.PROJECT_CATEGORY.FOOD_SERVICE", loc), Order: 3})
	codes = append(codes, commoncode.Code{Code: string(PRECIOUS_METAL), Name: i18nutils.Translate("commoncode.PROJECT_CATEGORY.PRECIOUS_METAL", loc), Order: 4})
	codes = append(codes, commoncode.Code{Code: string(MEDICAL), Name: i18nutils.Translate("commoncode.PROJECT_CATEGORY.MEDICAL", loc), Order: 5})
	codes = append(codes, commoncode.Code{Code: string(RETAIL), Name: i18nutils.Translate("commoncode.PROJECT_CATEGORY.RETAIL", loc), Order: 6})
	codes = append(codes, commoncode.Code{Code: string(REAL_ESTATE), Name: i18nutils.Translate("commoncode.PROJECT_CATEGORY.REAL_ESTATE", loc), Order: 7})
	codes = append(codes, commoncode.Code{Code: string(BESS), Name: i18nutils.Translate("commoncode.PROJECT_CATEGORY.BESS", loc), Order: 8})
	codes = append(codes, commoncode.Code{Code: string(EVENT), Name: i18nutils.Translate("commoncode.PROJECT_CATEGORY.EVENT", loc), Order: 9})
	codes = append(codes, commoncode.Code{Code: string(FACTORING), Name: i18nutils.Translate("commoncode.PROJECT_CATEGORY.FACTORING", loc), Order: 10})
	codes = append(codes, commoncode.Code{Code: string(ETC), Name: i18nutils.Translate("commoncode.PROJECT_CATEGORY.ETC", loc), Order: 11})

	return commoncode.CodeGroup{
		GroupCode: PROJECT_CATEGORY_CODE,
		Codes:     codes,
	}
}

type PROJECT_TYPE string

const (
	PROJECT_TYPE_CODE              = "PROJECT_TYPE"
	LOAN              PROJECT_TYPE = "LOAN"
	EQUITY            PROJECT_TYPE = "EQUITY"
)

type OFFERING_TYPE string

const (
	OFFERING_TYPE_CODE               = "OFFERING_TYPE"
	OFFERING_PRIVATE   OFFERING_TYPE = "OFFERING_PRIVATE"
	OFFERING_PUBLIC    OFFERING_TYPE = "OFFERING_PUBLIC"
	OFFERING_CANCEL    OFFERING_TYPE = "OFFERING_CANCEL"
)

func (o OFFERING_TYPE) ToRecruitment() Recruitment {
	switch o {
	case OFFERING_PRIVATE:
		return PRIVATE_OFFERING
	case OFFERING_PUBLIC:
		return PUBLIC_OFFERING
	case OFFERING_CANCEL:
		return CANCELLATION
	default:
		return PRIVATE_OFFERING
	}
	return ""
}

func (Project) GetTypeCodeGroup(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(LOAN), Name: i18nutils.Translate("commoncode.PROJECT_TYPE.LOAN", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(EQUITY), Name: i18nutils.Translate("commoncode.PROJECT_TYPE.EQUITY", loc), Order: len(codes) + 1})

	return commoncode.CodeGroup{
		GroupCode: PROJECT_TYPE_CODE,
		Codes:     codes,
	}
}

type PROJECT_STATUS string

const (
	PROJECT_STATUS_CODE                = "PROJECT_STATUS"
	STANDBY_FUNDING     PROJECT_STATUS = "STANDBY_FUNDING"   // 募集開始前
	OPEN_FUNDING        PROJECT_STATUS = "OPEN_FUNDING"      // 募集中
	END_FUNDING         PROJECT_STATUS = "END_FUNDING"       // 募集終了
	SUCCESS_FUNDING     PROJECT_STATUS = "SUCCESS_FUNDING"   // 成立
	FAIL_FUNDING        PROJECT_STATUS = "FAIL_FUNDING"      // 不成立
	OPERATION_FUNDING   PROJECT_STATUS = "OPERATION_FUNDING" // 運用中
	// OPERATION_END_FUNDING は運用を終えて償還を待つ状態(MIR-272)。出資は運用中のまま残し、
	// 最後の償還を払ったとき(ジョブ)か償還完了のときに償還済みへ変わる。
	OPERATION_END_FUNDING PROJECT_STATUS = "OPERATION_END_FUNDING" // 運用終了
	CLOSE_FUNDING         PROJECT_STATUS = "CLOSE_FUNDING"         // 償還完了(MIR-272 より前は「運用終了」)
	INVALID_FUNDING       PROJECT_STATUS = "INVALID_FUNDING"       // 無効
)

func (Project) GetStatusCodeGroup(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(STANDBY_FUNDING), Name: i18nutils.Translate("commoncode.PROJECT_STATUS.STANDBY_FUNDING", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(OPEN_FUNDING), Name: i18nutils.Translate("commoncode.PROJECT_STATUS.OPEN_FUNDING", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(END_FUNDING), Name: i18nutils.Translate("commoncode.PROJECT_STATUS.END_FUNDING", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(SUCCESS_FUNDING), Name: i18nutils.Translate("commoncode.PROJECT_STATUS.SUCCESS_FUNDING", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(FAIL_FUNDING), Name: i18nutils.Translate("commoncode.PROJECT_STATUS.FAIL_FUNDING", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(OPERATION_FUNDING), Name: i18nutils.Translate("commoncode.PROJECT_STATUS.OPERATION_FUNDING", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(OPERATION_END_FUNDING), Name: i18nutils.Translate("commoncode.PROJECT_STATUS.OPERATION_END_FUNDING", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(CLOSE_FUNDING), Name: i18nutils.Translate("commoncode.PROJECT_STATUS.CLOSE_FUNDING", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(INVALID_FUNDING), Name: i18nutils.Translate("commoncode.PROJECT_STATUS.INVALID_FUNDING", loc), Order: len(codes) + 1})

	return commoncode.CodeGroup{
		GroupCode: PROJECT_STATUS_CODE,
		Codes:     codes,
	}
}

type PROJECT_STATUS_LIST_TYPE string

const (
	PROJECT_STATUS_LIST_TYPE_CODE                            = "PROJECT_STATUS_LIST_TYPE"
	ALL_PROJECT_LIST_TYPE           PROJECT_STATUS_LIST_TYPE = "ALL_PROJECT"
	STANDBY_FUNDING_LIST_TYPE       PROJECT_STATUS_LIST_TYPE = "STANDBY_FUNDING"       // 募集開始前
	OPEN_FUNDING_LIST_TYPE          PROJECT_STATUS_LIST_TYPE = "OPEN_FUNDING"          // 募集中
	END_FUNDING_LIST_TYPE           PROJECT_STATUS_LIST_TYPE = "END_FUNDING"           // 募集終了
	OPERATION_FUNDING_LIST_TYPE     PROJECT_STATUS_LIST_TYPE = "OPERATION_FUNDING"     // 運営中
	OPERATION_END_FUNDING_LIST_TYPE PROJECT_STATUS_LIST_TYPE = "OPERATION_END_FUNDING" // 運用終了(償還を待つ状態。MIR-272)
	CLOSE_FUNDING_LIST_TYPE         PROJECT_STATUS_LIST_TYPE = "CLOSE_FUNDING"         // 償還完了(MIR-272 より前は「運営終了」)
)

func (Project) GetProjrctListStatusCodeGroup(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(STANDBY_FUNDING_LIST_TYPE), Name: i18nutils.Translate("commoncode.PROJECT_STATUS.STANDBY_FUNDING", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(OPEN_FUNDING_LIST_TYPE), Name: i18nutils.Translate("commoncode.PROJECT_STATUS.OPEN_FUNDING", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(END_FUNDING_LIST_TYPE), Name: i18nutils.Translate("commoncode.PROJECT_STATUS.END_FUNDING", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(OPERATION_FUNDING_LIST_TYPE), Name: i18nutils.Translate("commoncode.PROJECT_STATUS.OPERATION_FUNDING", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(OPERATION_END_FUNDING_LIST_TYPE), Name: i18nutils.Translate("commoncode.PROJECT_STATUS.OPERATION_END_FUNDING", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(CLOSE_FUNDING_LIST_TYPE), Name: i18nutils.Translate("commoncode.PROJECT_STATUS.CLOSE_FUNDING", loc), Order: len(codes) + 1})

	return commoncode.CodeGroup{
		GroupCode: PROJECT_STATUS_LIST_TYPE_CODE,
		Codes:     codes,
	}
}

type PROJECT_LIST_TYPE string

const (
	PROJECT_LIST_TYPE_CODE                   = "PROJECT_LIST_TYPE"
	ALL_PROJECT            PROJECT_LIST_TYPE = "ALL_PROJECT"
	LOAN_PROJECT           PROJECT_LIST_TYPE = "LOAN_PROJECT"
	EQUITY_PROJECT         PROJECT_LIST_TYPE = "EQUITY_PROJECT"
	FUND_PROJECT           PROJECT_LIST_TYPE = "FUND_PROJECT"
)

type PROJECT_SELECTION_TYPE string

const (
	PROJECT_SELECTION_TYPE_CODE                                      = "PROJECT_SELECTION_TYPE"
	PROJECT_SELECTION_LOTTERY                 PROJECT_SELECTION_TYPE = "LOTTERY"                 // 抽選
	PROJECT_SELECTION_FIRST_COME_FIRST_SERVED PROJECT_SELECTION_TYPE = "FIRST_COME_FIRST_SERVED" // 先着
)

func (Project) GetProjectSelectionTypeCodeGroup(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0)
	codes = append(codes, commoncode.Code{Code: string(PROJECT_SELECTION_LOTTERY), Name: i18nutils.Translate("commoncode.PROJECT_SELECTION_TYPE.LOTTERY", loc), Order: 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_SELECTION_FIRST_COME_FIRST_SERVED), Name: i18nutils.Translate("commoncode.PROJECT_SELECTION_TYPE.FIRST_COME_FIRST_SERVED", loc), Order: 2})

	return commoncode.CodeGroup{
		GroupCode: PROJECT_SELECTION_TYPE_CODE,
		Codes:     codes,
	}
}

type PROJECT_COUNTRY_TYPE string

const (
	PROJECT_COUNTRY_TYPE_CODE = "PROJECT_COUNTRY_TYPE"

	PROJECT_COUNTRY_JP PROJECT_COUNTRY_TYPE = "JP" // "国内"
	PROJECT_COUNTRY_US PROJECT_COUNTRY_TYPE = "US" // "米国"
)

type DIVIDEND_TYPE string

const (
	DIVIDEND_TYPE_CODE                    = "DIVIDEND_TYPE"
	INCOME_AND_CAPITAL_GAIN DIVIDEND_TYPE = "INCOME_AND_CAPITAL_GAIN"
	INCOME_GAIN             DIVIDEND_TYPE = "INCOME_GAIN"
	CAPITAL_GAIN            DIVIDEND_TYPE = "CAPITAL_GAIN"
)

func (Project) GetProjecDividendTypeGroup(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(INCOME_AND_CAPITAL_GAIN), Name: i18nutils.Translate("commoncode.DIVIDEND_TYPE.INCOME_AND_CAPITAL_GAIN", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(INCOME_GAIN), Name: i18nutils.Translate("commoncode.DIVIDEND_TYPE.INCOME_GAIN", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(CAPITAL_GAIN), Name: i18nutils.Translate("commoncode.DIVIDEND_TYPE.CAPITAL_GAIN", loc), Order: len(codes) + 1})

	return commoncode.CodeGroup{
		GroupCode: DIVIDEND_TYPE_CODE,
		Codes:     codes,
	}
}

func (Project) GetListCodeGroup(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(ALL_PROJECT), Name: i18nutils.Translate("commoncode.PROJECT_LIST_TYPE.ALL_PROJECT", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(LOAN_PROJECT), Name: i18nutils.Translate("commoncode.PROJECT_LIST_TYPE.LOAN_PROJECT", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(EQUITY_PROJECT), Name: i18nutils.Translate("commoncode.PROJECT_LIST_TYPE.EQUITY_PROJECT", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(FUND_PROJECT), Name: i18nutils.Translate("commoncode.PROJECT_LIST_TYPE.FUND_PROJECT", loc), Order: len(codes) + 1})

	return commoncode.CodeGroup{
		GroupCode: PROJECT_LIST_TYPE_CODE,
		Codes:     codes,
	}
}

func (Project) GetProjectCountryGroup(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(PROJECT_COUNTRY_JP), Name: i18nutils.Translate("commoncode.PROJECT_COUNTRY_TYPE.PROJECT_COUNTRY_JP", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(PROJECT_COUNTRY_US), Name: i18nutils.Translate("commoncode.PROJECT_COUNTRY_TYPE.PROJECT_COUNTRY_US", loc), Order: len(codes) + 1})

	return commoncode.CodeGroup{
		GroupCode: PROJECT_COUNTRY_TYPE_CODE,
		Codes:     codes,
	}
}

const (
	PROJECT_LIST_CLOUMN_LIST = "project.id,project.common_id,project.fund_number,project.spv_id," +
		"project.name,project.description,project.country,project.offering_type ,project.category_type,project.selection_type AS selection_type,project.start_at,project.end_at,project.planed_operation_start_date,project.planed_operation_end_date,project.block_invest," +
		"project.project_type,project.project_status,project.owner_id,project.update_id," +
		"project.capacity,project.operation_term,project.country,project.dividend_rate,project.is_investor_limit, project.capital_raising_entity_id," +
		"(SELECT IFNULL(SUM(amount),0) FROM share WHERE share.project_id = project.id  AND invest_status IN ('SUCCESS_INVEST','REQUEST_INVEST')) AS current_share_amount," +
		"(SELECT COUNT(qty) FROM share WHERE share.project_id = project.id  AND invest_status IN ('SUCCESS_INVEST','REQUEST_INVEST')) AS current_share_qty, project.share_price,project.target_share_qty,project.target_share_amount,project.bank_account," +
		"project.bank_account,project.main_image_url,project.posted,project.created,project.updated," +
		" project.is_wait_invest,(SELECT IFNULL(SUM(amount),0) FROM share WHERE share.project_id = project.id  AND invest_status IN ('SUCCESS_INVEST','OPERATION_INVEST','REFUND_INVEST')) AS raised_amount," +
		" (SELECT count(share.id) FROM share WHERE share.project_id = project.id) as transaction_count, project_view_count.view_count as view_count, project.dividend_type, project.min_share_qty, project.final_funding_amount"

	PROJECT_PAGE_STRINGS string = "project.id,project.common_id,project.spv_id,project.name," +
		"project.category_type,project.selection_type AS selection_type,project.start_at,project.end_at,project.planed_operation_start_date,project.planed_operation_end_date,project.project_type,project.project_status," +
		"project.owner_id,project.update_id,project.capacity,project.country,project.dividend_rate," +
		"project.operation_term," +
		"(SELECT IFNULL(SUM(amount),0) FROM share WHERE share.project_id = project.id  AND invest_status IN ('SUCCESS_INVEST','REQUEST_INVEST')) AS current_share_amount, " +
		"(SELECT COUNT(qty) FROM share WHERE share.project_id = project.id  AND invest_status IN ('SUCCESS_INVEST','REQUEST_INVEST','REFUND_INVEST')) AS current_share_qty,project.min_share_qty,project.share_price,project.target_min_share_amount,project.target_share_qty," +
		"project.target_share_amount,project.bank_account,project.posted,project.main_image_url,project.created,project.updated,project.deleted,project.project_progress_percent,project.dividend_type"
)

const (
	// https://illuminarean.atlassian.net/wiki/spaces/ESJ/pages/190906518/ES 상품 관련 정보
	PUBLIC_COLUMN = PUBLIC_PROGRESS_INVEST_PERCENT + " AS public_progress_invest_percent," +
		PUBLIC_VIEW_SHARE_AMOUNT + " AS public_view_share_amount"
	PUBLIC_VIEW_SHARE_AMOUNT = "CASE WHEN" + // ユーザー表示用投資金額
		" project.project_status = 'SUCCESS_FUNDING' THEN (SELECT IFNULL(SUM(amount),0) FROM share WHERE share.project_id = project.id AND share.invest_status = 'SUCCESS_INVEST')" + // 成立
		" WHEN project.project_status = 'FAIL_FUNDING' THEN (SELECT IFNULL(SUM(amount),0) FROM share WHERE share.project_id = project.id AND share.invest_status = 'FAIL_INVEST')" + // 不成立
		" WHEN project.project_status IN ('OPERATION_FUNDING','OPERATION_END_FUNDING') THEN (SELECT IFNULL(SUM(amount),0) FROM share WHERE share.project_id = project.id AND share.invest_status IN ('OPERATION_INVEST','REFUND_INVEST'))" + // 運用中・運用終了(最後の償還を払うと出資は償還済みになるが、償還完了まで金額は変えない。MIR-272)
		" WHEN project.project_status = 'CLOSE_FUNDING' THEN (SELECT IFNULL(SUM(amount),0) FROM share WHERE share.project_id = project.id AND share.invest_status = 'REFUND_INVEST')" + // 償還完了
		" ELSE (SELECT IFNULL(SUM(amount),0) FROM share WHERE share.project_id = project.id  AND invest_status IN ('SUCCESS_INVEST','REQUEST_INVEST'))" + // 無効
		" END"
	PUBLIC_PROGRESS_INVEST_PERCENT = "CASE WHEN" + // ユーザー表示用パーセンテージ
		" project.project_status = 'STANDBY_FUNDING' OR project.project_status = 'OPEN_FUNDING' OR project.project_status = 'END_FUNDING' THEN (((SELECT IFNULL(SUM(amount),0) FROM share WHERE share.project_id = project.id AND invest_status IN ('SUCCESS_INVEST','REQUEST_INVEST')) / project.target_share_amount) * 100)" + // 募集開始前、募集中、募集終了
		" WHEN project.project_status = 'SUCCESS_FUNDING' THEN (((SELECT IFNULL(SUM(amount),0) FROM share WHERE share.project_id = project.id AND share.invest_status = 'SUCCESS_INVEST') / project.target_share_amount) * 100)" + // 成立
		" WHEN project.project_status = 'FAIL_FUNDING' THEN (((SELECT IFNULL(SUM(amount),0) FROM share WHERE share.project_id = project.id AND share.invest_status = 'FAIL_INVEST') / project.target_share_amount) * 100)" + // 不成立
		" WHEN project.project_status IN ('OPERATION_FUNDING','OPERATION_END_FUNDING') THEN (((SELECT IFNULL(SUM(amount),0) FROM share WHERE share.project_id = project.id AND share.invest_status IN ('OPERATION_INVEST','REFUND_INVEST')) / project.target_share_amount) * 100)" + // 運用中・運用終了(MIR-272)
		" WHEN project.project_status = 'CLOSE_FUNDING' THEN (((SELECT IFNULL(SUM(amount),0) FROM share WHERE share.project_id = project.id AND share.invest_status = 'REFUND_INVEST') / project.target_share_amount) * 100)" + // 償還完了
		" ELSE (((SELECT IFNULL(SUM(amount),0) FROM share WHERE share.project_id = project.id  AND share.invest_status = 'SUCCESS_INVEST') / project.target_share_amount) * 100)" + // 無効
		" END"
	PUBLIC_COLUMN_AMOUNT_ORDER = PUBLIC_VIEW_SHARE_AMOUNT + " DESC"
	PUBLIC_COLUMN_END_AT_ORDER = " project.project_status = 'END_FUNDING' ASC, project.end_at > now() DESC, project.end_at ASC"
)

const (
	PROJECT_DETAIL_COLUMN_LIST = "project.id, project.common_id, project.fund_number, project.spv_id, project.name,project.offering_type, project.category_type, project.selection_type AS selection_type, project.admin_note, project.project_info_html, project.company_info_html, project.description, project.start_at, project.end_at, project.planed_operation_start_date, project.planed_operation_end_date, project.project_type, project.project_status, project.owner_id, project.is_closed_fund, project.closed_fund_slug, " +
		"project.update_id, project.capacity, project.country, project.dividend_rate, project.operation_term, project.min_share_qty, project.share_price, project.target_min_share_amount," +
		" project.target_share_qty, project.target_share_amount, project.posted, project.block_invest, project.main_image_id, project.main_image_url, project.created, project.updated, " +
		" project.caution_info_html, project.bank_name, project.branch_name, project.bank_code, project.branch_code, project.bank_account, project.depositor, project.contract_before_page_date, project.contract_after_page_date, project.silent_partnerships_date, project.project_report_date, project.capital_raising_entity_id," +
		" (SELECT IFNULL(SUM(amount),0) FROM share WHERE share.project_id = project.id  AND invest_status IN ('SUCCESS_INVEST','REQUEST_INVEST')) AS current_share_amount, " +
		" project.is_wait_invest,(SELECT IFNULL(SUM(amount),0) FROM share WHERE share.project_id = project.id  AND invest_status IN ('SUCCESS_INVEST','OPERATION_INVEST','REFUND_INVEST')) AS raised_amount, project.is_manual_invest, " +
		" (SELECT count(share.id) FROM share WHERE share.project_id = project.id) as transaction_count, (target_share_qty - (SELECT IFNULL(SUM(qty),0) FROM share WHERE share.project_id = project.id  AND invest_status IN ('SUCCESS_INVEST','REQUEST_INVEST'))) AS remaining_qty , (target_share_qty - (SELECT IFNULL(SUM(qty),0) FROM share WHERE share.project_id = project.id  AND invest_status IN ('SUCCESS_INVEST'))) AS invest_wait_remaining_qty," +
		" project.is_investor_limit, project.is_amount_limit,project.project_progress_percent, project.is_visible_investment_info, project.dividend_type, project.final_funding_amount, project.detail_version"
)

const TargetInvestorCount = 499

type Project struct {
	Id       int64  `json:"id"`
	CommonId string `json:"common_id" xorm:"notnull UUID" swagger:"desc(共通ID)"`
	// 未入力を許すため NULL 許容とする。DB 側でユニーク制約を張っており、
	// MySQL の仕様で NULL 同士は重複とみなされない。
	FundNumber               *int64                 `json:"fund_number" xorm:"BIGINT null" i18n:"commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_FUND_NUMBER" swagger:"desc(投資家へ提示するファンド番号)"`
	OfferingType             OFFERING_TYPE          `json:"offering_type" xorm:"varchar(50)" swagger:"desc(ファンディング情報)"`
	SpvId                    int64                  `json:"spv_id" xorm:"index" i18n:"commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_SPV" swagger:"desc(開発会社ID)"`
	Spv                      *Spv                   `json:"spv" xorm:"-" swagger:"desc(開発会社 SPV / デベロッパー)"` // TODO 削除予定（APIレスポンス側で処理）
	Name                     string                 `json:"name" xorm:"VARCHAR(100) notnull" i18n:"commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_NAME" swagger:"desc(商品名)"`
	CategoryType             PROJECT_CATEGORY       `json:"category_type" i18n:"commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_POWER_PLANT" xorm:"notnull" swagger:"desc(ファンディングカテゴリ)"`                   // ファンディングカテゴリ
	SelectionType            PROJECT_SELECTION_TYPE `json:"selection_type" xorm:"'selection_type'" i18n:"commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_SELECTION_METHOD" swagger:"desc(配分方法 1:抽選 2:先着)"` // 配分方法
	AdminNote                string                 `json:"admin_note" xorm:"TEXT" swagger:"desc(管理者メモ)"`
	ProjectInfoHtml          string                 `json:"project_info_html" xorm:"LONGTEXT" i18n:"commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_SUMMARY" swagger:"desc(プロジェクト情報)"`      // プロジェクトコンテンツ HTML
	CompanyInfoHtml          string                 `json:"company_info_html" xorm:"LONGTEXT" i18n:"commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_QUESTIONS" swagger:"desc(お問い合わせ／会社情報)"` // 会社コンテンツ HTML
	Description              string                 `json:"description" xorm:"not null" i18n:"commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_DESCRIPTION" swagger:"desc(商品説明)"`
	StartAt                  jsonutils.JSONTime     `json:"start_at" i18n:"commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_FUNDRAISING_START_DATE" swagger:"desc(募集開始日／ファンド開始日)"`
	EndAt                    jsonutils.JSONTime     `json:"end_at" i18n:"commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_FUNDRAISING_END_DATE" swagger:"desc(募集終了日)"`
	PlanedOperationStartDate jsonutils.JSONTime     `json:"planed_operation_start_date" swagger:"desc(予定運用開始日)"`
	PlanedOperationEndDate   jsonutils.JSONTime     `json:"planed_operation_end_date" swagger:"desc(予定運用終了日)"`
	ProjectType              PROJECT_TYPE           `json:"project_type" i18n:"commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_INVESTMENT_METHOD" swagger:"desc(プロジェクトタイプ 1:LOAN 2:EQUITY)"`
	ProjectStatus            PROJECT_STATUS         `json:"project_status" i18n:"commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_PROJECT_STATUS" swagger:"desc(プロジェクト状態（ファンディング状態） 1:募集待機 2:募集中 3:募集成功 4:募集失敗 5:投資制限)"`
	OwnerId                  int64                  `json:"owner_id" xorm:"notnull" swagger:"desc(登録者)"`
	UpdateId                 int64                  `json:"update_id" swagger:"desc(最終更新者)"`
	Capacity                 string                 `json:"capacity" i18n:"commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_SCALE" swagger:"desc(開発容量（kW、フォームP DC規模）)"`
	Country                  PROJECT_COUNTRY_TYPE   `json:"country" i18n:"commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_COUNTRY"`               // 国
	DividendRate             float64                `json:"dividend_rate" i18n:"commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_DIVIDEND_RATE"`   // 配当率
	OperationTerm            int                    `json:"operation_term" i18n:"commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_OPERATION_TERM"` // 運用期間（ローン期間）
	CurrentShareAmount       int64                  `json:"current_share_amount" swagger:"desc(募集済みShare金額)"`
	CurrentShareQty          int64                  `json:"current_share_qty" swagger:"desc(現在募集済みShare数量)"`
	MinShareQty              int64                  `json:"min_share_qty" i18n:"commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_MINIMUM_INVESTMENT" swagger:"desc(最小購入数量)"`
	SharePrice               int64                  `json:"share_price" i18n:"commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_PRICE_PER_SHARE" swagger:"desc(1口あたりの価格)"`
	TargetMinShareAmount     int64                  `json:"target_min_share_amount" i18n:"commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_MINIMUM_FUNDRAISING_AMOUNT" swagger:"desc(目標募集金額)"`
	TargetShareQty           int64                  `json:"target_share_qty" i18n:"commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_SHARES" swagger:"desc(目標株式数（総発行株式）)"`
	TargetShareAmount        int64                  `json:"target_share_amount" i18n:"commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_TARGET_FUNDRAISING_AMOUNT" swagger:"desc(上限募集金額)"`
	Posted                   bool                   `json:"posted" xorm:"posted" i18n:"commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_POSTED_STATE" swagger:"desc(ファンド掲載有無)"`
	ViewCount                int64                  `json:"view_count" xorm:"-" swagger:"desc(パブリック閲覧数)"` // TODO 削除予定

	BlockInvest             bool               `json:"block_invest" i18n:"commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_LIMITED_INVEST" swagger:"desc(投資制限 True:投資不可 False:投資可)"`
	MainImageId             int64              `json:"main_image_id" swagger:"desc(プロジェクトメイン画像ID)"`
	MainImageUrl            string             `json:"main_image_url" i18n:"commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_MAINIMAGE_URL" swagger:"desc(ファンディングメイン画像URL)"`
	MainImage               *ProjectImage      `json:"main_image" xorm:"-" swagger:"desc(プロジェクト画像)"` // TODO 削除予定
	SubImageInfoList        []*SubImageInfo    `json:"sub_image_info_list" xorm:"-" swagger:"desc(サブ画像情報リスト)"`
	Attachments             []*Attachment      `json:"attachments" xorm:"-" swagger:"desc(プロジェクト添付ファイル一覧)"`
	News                    []*ProjectNews     `json:"project_news" xorm:"-" swagger:"desc(プロジェクトニュース一覧)"`
	ClosedFundSlug          string             `json:"closed_fund_slug" swagger:"desc(限定公開ファンドURLスラッグ)"`
	IsWaitInvest            bool               `json:"is_wait_invest" swagger:"desc(投資待機状態 True:投資待機商品 False:投資可能)"`
	IsManualInvest          bool               `json:"is_manual_invest" swagger:"desc(手動投資有無)"`
	IsInvestorLimit         bool               `json:"is_investor_limit" i18n:"commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_IS_INVESTOR_LIMIT" swagger:"desc(投資者数制限有無)"`
	IsAmountLimit           bool               `json:"is_amount_limit" i18n:"commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_IS_AMOUNT_LIMIT" swagger:"desc(投資金額制限有無)"`
	IsVisibleInvestmentInfo bool               `json:"is_visible_investment_info" xorm:"notnull default 0" i18n:"commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_IS_VISIBLE_INVESTMENT_INFO" swagger:"desc(投資情報公開有無／初期値は非公開)"`
	IsClosedFund            bool               `json:"is_closed_fund" swagger:"desc(限定公開ファンドフラグ)"`
	Created                 jsonutils.JSONTime `json:"created" xorm:"created"`
	Updated                 jsonutils.JSONTime `json:"updated" xorm:"updated"`
	Deleted                 jsonutils.JSONTime `json:"-" xorm:"deleted"`
	// MIR-1492: ファンド詳細の保存回数。UpdateForAdminDetail が SQL 側で1増やす。
	// 構造体の値で書き戻すと、同時に保存した別の管理者の加算を打ち消すため、
	// 読み取り専用にしている。notnull・default は本番の列定義(マイグレーション)と
	// そろえ、sqlite のテストでも新しい行が 0 から数えられるようにするため。
	DetailVersion int64 `json:"detail_version" xorm:"<- notnull default 0"`

	TransactionCount            int64                       `json:"transaction_count" xorm:"<-" swagger:"desc(取引件数)"`
	CautionInfoHtml             string                      `json:"caution_info_html" xorm:"LONGTEXT" i18n:"commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_CAUTION" swagger:"desc(注意事項)"`
	BankName                    string                      `json:"bank_name" i18n:"commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_BANK_NAME" swagger:"desc(金融機関名)"`
	BranchName                  string                      `json:"branch_name" i18n:"commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_BRANCH_NAME" swagger:"desc(支店名)"`
	BankCode                    string                      `json:"bank_code" swagger:"desc(銀行コード)"`
	BranchCode                  string                      `json:"branch_code" swagger:"desc(銀行支店コード)"`
	BankAccount                 string                      `json:"bank_account" i18n:"commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_BANK_ACCOUNT" swagger:"desc(口座番号)"`
	Depositor                   string                      `json:"depositor" i18n:"commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_DEPOSITOR" swagger:"desc(口座名義人)"`
	ContractBeforePage          *ProjectFile                `json:"contract_before_page" xorm:"-" swagger:"desc(契約成立前書面)"`
	ContractAfterPage           *ProjectFile                `json:"contract_after_page" xorm:"-" swagger:"desc(契約成立時書面)"`
	SilentPartnerships          *ProjectFile                `json:"silent_partnerships" xorm:"-" swagger:"desc(匿名組合契約書)"`
	ProjectReport               *ProjectFile                `json:"project_report" xorm:"-" swagger:"desc(ファンド報告書)"`
	ContractBeforePageDate      jsonutils.JSONTime          `json:"contract_before_page_date" swagger:"desc(契約成立前書面 登録日)"`
	ContractAfterPageDate       jsonutils.JSONTime          `json:"contract_after_page_date" swagger:"desc(契約成立時書面 登録日)"`
	SilentPartnershipsDate      jsonutils.JSONTime          `json:"silent_partnerships_date" swagger:"desc(匿名組合契約書 登録日)"`
	ProjectReportDate           jsonutils.JSONTime          `json:"project_report_date" swagger:"desc(ファンド報告書 登録日)"`
	TransactionDate             jsonutils.JSONTime          `json:"transaction_date" xorm:"<-" swagger:"desc(取引日)"`
	TransactionAmount           int                         `json:"transaction_amount" xorm:"<-" swagger:"desc(取引募集金額)"`
	PublicProgressInvestPercent float64                     `json:"public_progress_invest_percent" xorm:"<-" swagger:"desc(ユーザー向け進捗バー用 募集額％（投資キャンセル含む）)"`
	RemainingQty                int64                       `json:"remaining_qty" xorm:"<-" swagger:"desc(購入可能口数)"`
	CurrentInvestorCount        int64                       `json:"current_investor_count" xorm:"<-" swagger:"desc(投資者数)"`
	RaisedAmount                int64                       `json:"raised_amount" xorm:"<-" swagger:"desc(調達金額)"`
	RepaymentDividendSchedules  []RepaymentDividendSchedule `json:"repayment_dividend_schedules" xorm:"-"`
	ProjectInvestmentExecute    *ProjectInvestmentExecute   `json:"project_investment_execute" xorm:"-"`
	InvestWaitRemainingQty      int64                       `json:"invest_wait_remaining_qty" xorm:"<-" swagger:"desc(投資待機可能口数)"`
	PublicViewShareAmount       int64                       `json:"public_view_share_amount" swagger:"desc(商品状態別 Share 金額)"`
	ProjectProgressPercent      int                         `json:"project_progress_percent" i18n:"commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_PROJECT_PROGRESS_PERCENT" swagger:"desc(プロジェクト進捗率)"`
	ProjectMilestones           []ProjectMilestone          `json:"project_milestones" xorm:"-" swagger:"desc(プロジェクトマイルストーン一覧)"`
	DividendType                *DIVIDEND_TYPE              `json:"dividend_type" xorm:"dividend_type" swagger:"desc(分配方式)"`
	FinalFundingAmount          int64                       `json:"final_funding_amount" xorm:"final_funding_amount"`           // 募集終了時点の申込金額
	CapitalRaisingEntityId      *int64                      `json:"capital_raising_entity_id" xorm:"capital_raising_entity_id"` // 資金調達者ID
	CapitalRaisingEntity        *CapitalRaisingEntity       `json:"capital_raising_entity" xorm:"-"`                            // 資金調達者
}

type SubImageInfo struct {
	SubImageId  int64         `json:"sub_image_id" swagger:"desc(プロジェクトサブ画像固有ID)"`
	SubImageUrl string        `json:"sub_image_url"  i18n:"commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_SUBIMAGE_URL"`
	SubImage    *ProjectImage `json:"sub_image"  xorm:"-" swagger:"desc(プロジェクトサブイメージ)"`
}

func (p Project) TableName() string {
	return "project"
}

func (p Project) SelectAll() string {
	return p.TableName() + ".*"
}

func (p Project) WithSpv(spv Spv) Project {
	if reflect.ValueOf(spv).IsZero() {
		return p
	}

	p.Spv = &spv
	return p
}

func (p Project) WithMainImage(mainImage ProjectImage) Project {
	if reflect.ValueOf(mainImage).IsZero() {
		return p
	}

	p.MainImage = &mainImage
	return p
}

func (p Project) WithProjectNews(projectNews []ProjectNews) Project {
	if len(projectNews) == 0 {
		return p
	}

	p.News = lo.Map(projectNews, func(p ProjectNews, _ int) *ProjectNews {
		return &p
	})

	return p
}

func (p Project) WithAttachments(attachments []Attachment) Project {
	if len(attachments) == 0 {
		return p
	}

	p.Attachments = lo.Map(attachments, func(a Attachment, _ int) *Attachment {
		return &a
	})

	return p
}

func (p Project) WithContractBeforePage(contractBeforePage ProjectFile) Project {
	if reflect.ValueOf(contractBeforePage).IsZero() {
		return p
	}

	p.ContractBeforePage = &contractBeforePage
	return p
}

func (p Project) WithContractAfterPage(contractAfterPage ProjectFile) Project {
	if reflect.ValueOf(contractAfterPage).IsZero() {
		return p
	}

	p.ContractAfterPage = &contractAfterPage
	return p
}

func (p Project) WithSilentPartnerships(silentPartnerships ProjectFile) Project {
	if reflect.ValueOf(silentPartnerships).IsZero() {
		return p
	}

	p.SilentPartnerships = &silentPartnerships
	return p
}

func (p Project) WithProjectReport(projectReport ProjectFile) Project {
	if reflect.ValueOf(projectReport).IsZero() {
		return p
	}

	p.ProjectReport = &projectReport
	return p
}

func (p Project) WithRepaymentDividendSchedules(repaymentDividendSchedules []RepaymentDividendSchedule) Project {
	if len(repaymentDividendSchedules) == 0 {
		return p
	}

	p.RepaymentDividendSchedules = lo.Map(repaymentDividendSchedules, func(r RepaymentDividendSchedule, _ int) RepaymentDividendSchedule {
		return r
	})

	return p
}

func (p Project) WithProjectInvestmentExecute(projectInvestmentExecute ProjectInvestmentExecute) Project {
	if reflect.ValueOf(projectInvestmentExecute).IsZero() {
		return p
	}

	p.ProjectInvestmentExecute = &projectInvestmentExecute
	return p
}

func (p Project) WithProjectMilestones(projectMilestones []ProjectMilestone) Project {
	if len(projectMilestones) == 0 {
		return p
	}

	p.ProjectMilestones = lo.Map(projectMilestones, func(p ProjectMilestone, _ int) ProjectMilestone {
		return p
	})

	return p
}

func (p Project) WithProjectViewCount(projectViewCount int64) Project {
	p.ViewCount = projectViewCount
	return p
}

func (p *Project) Create(ctx context.Context, tokenId int64) (int64, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return 0, err
	}
	p.CommonId = id.String()
	p.OwnerId = tokenId
	p.UpdateId = tokenId
	p.ProjectStatus = STANDBY_FUNDING
	p.OfferingType = offeringType(p.IsInvestorLimit)

	return factory.DB(ctx).Insert(p)
}

// プロジェクト照会
func (Project) GetById(ctx context.Context, id int64) (*Project, error) {
	var p Project
	if has, err := factory.DB(ctx).Table("project").Select(PROJECT_DETAIL_COLUMN_LIST).ID(id).Get(&p); err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}
	spv, err := Spv{}.GetById(ctx, p.SpvId)
	if err != nil {
		return nil, err
	} else {
		p.Spv = spv
	}

	if p.MainImageId > 0 {
		projectImages, err := ProjectImage{}.FindByProjectId(ctx, p.Id)
		if err != nil {
			return nil, err
		}
		for _, img := range projectImages {
			if img.Id == p.MainImageId {
				p.MainImage = img
			} else {
				subImageInfo := SubImageInfo{
					SubImageId:  img.Id,
					SubImageUrl: img.FileUrl,
					SubImage:    img,
				}
				p.SubImageInfoList = append(p.SubImageInfoList, &subImageInfo)
			}
		}
		sort.Slice(p.SubImageInfoList, func(i, j int) bool {
			return p.SubImageInfoList[i].SubImage.Order < p.SubImageInfoList[j].SubImage.Order
		})
	}

	_, news, err := ProjectNews{}.GetProjectById(ctx, p.Id, "created", "asc", 0, 30)
	if err != nil {
		return nil, err
	} else {
		p.News = news
	}
	attachments, err := Attachment{}.FindByResourceId(ctx, 10, p.CommonId)
	if err != nil {
		return nil, err
	} else {
		p.Attachments = attachments
	}

	p.ContractBeforePage, err = ProjectFile{}.FindByProjectIdAndFileType(ctx, p.Id, CONTRACT_BEFORE_PAGE)
	if err != nil {
		return nil, err
	}
	p.ContractAfterPage, err = ProjectFile{}.FindByProjectIdAndFileType(ctx, p.Id, CONTRACT_AFTER_PAGE)
	if err != nil {
		return nil, err
	}
	p.SilentPartnerships, err = ProjectFile{}.FindByProjectIdAndFileType(ctx, p.Id, SILENT_PARTNERSHIPS)
	if err != nil {
		return nil, err
	}
	p.ProjectReport, err = ProjectFile{}.FindByProjectIdAndFileType(ctx, p.Id, PROJECT_REPORT)
	if err != nil {
		return nil, err
	}

	_, p.RepaymentDividendSchedules, err = RepaymentDividendSchedule{}.GetByProjectId(ctx, p.Id, SearchInput{MaxResultCount: 100})
	if err != nil {
		return nil, err
	}

	p.ProjectInvestmentExecute, err = ProjectInvestmentExecute{}.GetByProjectId(ctx, p.Id)
	if err != nil {
		return nil, err
	}
	pv, err := ProjectViewCount{}.GetByProjectId(ctx, p.Id)
	if err != nil {
		return nil, err
	}
	p.ViewCount = pv.ViewCount

	p.ProjectMilestones, err = ProjectMilestone{}.GetAllByProjectId(ctx, p.Id)
	if err != nil {
		return nil, err
	}

	p.CurrentInvestorCount, err = Project{}.GetCurrentInvestorCount(ctx, p.Id)
	if err != nil {
		return nil, err
	}

	if p.CapitalRaisingEntityId != nil {
		capitalRaisingEntity, err := CapitalRaisingEntity{}.GetById(ctx, *p.CapitalRaisingEntityId)
		if err != nil {
			return nil, err
		}

		p.CapitalRaisingEntity = capitalRaisingEntity
	}

	return &p, nil
}

// GetByIdForPublic プロジェクト情報取得
func (Project) GetByIdForPublic(ctx context.Context, id int64, isPreview bool) (*Project, error) {
	var p Project
	q := factory.DB(ctx).Table("project").Select(PROJECT_DETAIL_COLUMN_LIST + "," + PUBLIC_COLUMN).ID(id)
	if !isPreview {
		q.Where("project.posted = ?", true)
	}

	if has, err := q.Get(&p); err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}

	spv, err := Spv{}.GetById(ctx, p.SpvId)
	if err != nil {
		return nil, err
	} else {
		p.Spv = spv
	}

	if p.MainImageId > 0 {
		projectImages, err := ProjectImage{}.FindByProjectId(ctx, p.Id)
		if err != nil {
			return nil, err
		}
		for _, img := range projectImages {
			if img.Id == p.MainImageId {
				p.MainImage = img
			} else if img.Order == 0 {
				// 表示順:0のものは非表示のためスキップ
				continue
			} else {
				subImageInfo := SubImageInfo{
					SubImageId:  img.Id,
					SubImageUrl: img.FileUrl,
					SubImage:    img,
				}
				p.SubImageInfoList = append(p.SubImageInfoList, &subImageInfo)
			}
		}
		sort.Slice(p.SubImageInfoList, func(i, j int) bool {
			return p.SubImageInfoList[i].SubImage.Order < p.SubImageInfoList[j].SubImage.Order
		})
	}

	_, news, err := ProjectNews{}.GetPostedList(ctx, p.Id, 0, 30)
	if err != nil {
		return nil, err
	} else {
		p.News = news
	}
	attachments, err := Attachment{}.FindByResourceId(ctx, 10, p.CommonId)
	if err != nil {
		return nil, err
	} else {
		p.Attachments = attachments
	}

	p.ContractBeforePage, err = ProjectFile{}.FindByProjectIdAndFileType(ctx, p.Id, CONTRACT_BEFORE_PAGE)
	if err != nil {
		return nil, err
	}
	p.ContractAfterPage, err = ProjectFile{}.FindByProjectIdAndFileType(ctx, p.Id, CONTRACT_AFTER_PAGE)
	if err != nil {
		return nil, err
	}
	p.SilentPartnerships, err = ProjectFile{}.FindByProjectIdAndFileType(ctx, p.Id, SILENT_PARTNERSHIPS)
	if err != nil {
		return nil, err
	}
	p.ProjectReport, err = ProjectFile{}.FindByProjectIdAndFileType(ctx, p.Id, PROJECT_REPORT)
	if err != nil {
		return nil, err
	}

	_, p.RepaymentDividendSchedules, err = RepaymentDividendSchedule{}.GetByProjectId(ctx, p.Id, SearchInput{MaxResultCount: 100})
	if err != nil {
		return nil, err
	}

	p.ProjectInvestmentExecute, err = ProjectInvestmentExecute{}.GetByProjectId(ctx, p.Id)
	if err != nil {
		return nil, err
	}
	pv, err := ProjectViewCount{}.GetByProjectId(ctx, p.Id)
	if err != nil {
		return nil, err
	}
	err = pv.UpdateCount(ctx)
	if err != nil {
		return nil, err
	}

	p.ViewCount = pv.ViewCount

	p.ProjectMilestones, err = ProjectMilestone{}.GetAllByProjectId(ctx, p.Id)
	if err != nil {
		return nil, err
	}

	p.CurrentInvestorCount, err = Project{}.GetCurrentInvestorCount(ctx, p.Id)
	if err != nil {
		return nil, err
	}

	if p.CapitalRaisingEntityId != nil {
		capitalRaisingEntity, err := CapitalRaisingEntity{}.GetById(ctx, *p.CapitalRaisingEntityId)
		if err != nil {
			return nil, err
		}

		p.CapitalRaisingEntity = capitalRaisingEntity
	}

	return &p, nil
}

// 掲示日を過ぎた未掲示商品検索
func OutOfDateQuery(q xorm.Interface) error {
	queryCond := builder.NewCond()
	queryCond = queryCond.And(builder.Eq{"project.posted": false}).And(builder.Lte{"date(project.start_at)": time.Now().Local()})

	if ok := queryCond.IsValid(); ok {
		q.(*xorm.Session).And(queryCond)
	} else {
		return errors.New("fail query valid")
	}

	return nil
}

func (Project) GetProjectsByCompanyId(ctx context.Context, sortby, order string, offset, limit int, companyId int64) (int64, []Project, error) {
	q := factory.DB(ctx)
	q.Table("project").Select(PROJECT_LIST_CLOUMN_LIST)
	q.Join("LEFT", "project_view_count", "project.id = project_view_count.project_id")
	q.Join("LEFT", "share", "project.id = share.project_id")
	q.GroupBy("project.id")

	if err := modelhelper.SetSortOrder(q, sortby, order); err != nil {
		factory.Logger(ctx).Error(err)
	}

	var totalCount int64
	var items []Project

	totalCount, err := q.Where("spv_id = ?", companyId).Limit(limit, offset).FindAndCount(&items)
	if err != nil {
		return 0, nil, err
	}

	return totalCount, items, nil
}

func (Project) GetAll(ctx context.Context, sortby, order string, offset, limit int, projectId, projectStatus, targetShareAmount, projectName, projectType, categoryType *string, blockInvest *bool, startDate, endDate string, outOfDate bool, posted *bool) (int64, []Project, error) {
	q := factory.DB(ctx)

	var totalCount int64
	var items []Project
	q.Table("project").Select(PROJECT_LIST_CLOUMN_LIST)
	q.Join("LEFT", "project_view_count", "project.id = project_view_count.project_id")
	q.Join("LEFT", "share", "project.id = share.project_id")

	q.GroupBy("project.id")
	if outOfDate {
		if err := OutOfDateQuery(q); err != nil {
			factory.Logger(ctx).Error(err)
		}
	}

	if err := modelhelper.SetSortOrder(q, sortby, order); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SearchEqQuery(q, "project", "id", projectId); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SearchEqQuery(q, "project", "project_status", projectStatus); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SearchEqQuery(q, "project", "target_share_amount", targetShareAmount); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SearchLikeQuery(q, "project", "name", projectName); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SearchBoolQuery(q, "project", "block_invest", blockInvest); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SearchBoolQuery(q, "project", "posted", posted); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SearchEqQuery(q, "project", "project_type", projectType); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SearchEqQuery(q, "project", "category_type", categoryType); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SetBetweenDate(q, "project.created", startDate, endDate); err != nil {
		factory.Logger(ctx).Error(err)
	}

	// GROUP BYを使用するとCOUNTが正確に表示されないため、Findに変更した後、個別にCOUNTを記述
	totalCount, err := q.Limit(limit, offset).FindAndCount(&items)
	if err != nil {
		return 0, nil, err
	}

	return totalCount, items, nil
}

func (Project) GetDashboardList(ctx context.Context, offset, limit int) (int64, []Project, error) {
	q := factory.DB(ctx).Table("project").Select(PROJECT_PAGE_STRINGS)

	var totalCount int64

	var items []Project
	totalCount, err := q.Where("posted = false").Where("DATE(CONVERT_TZ(start_at, 'UTC', 'Asia/Tokyo')) <= DATE(CONVERT_TZ(?, 'UTC', 'Asia/Tokyo'))", time.Now()).Desc("start_at").Limit(limit, offset).FindAndCount(&items)
	if err != nil {
		return 0, nil, err
	}
	if items == nil {
		items = make([]Project, 0, 0)
	}

	return totalCount, items, nil
}

// 499名制限がある場合はプライベートに指定される。
func offeringType(isInvestorLimit bool) OFFERING_TYPE {
	if isInvestorLimit {
		return OFFERING_PRIVATE
	}
	return OFFERING_PUBLIC
}

// Update は project を丸ごと保存する。project_status は書かない(MIR-1314)。
// 「丸ごと保存」は読み込んだ時点の状態を一緒に書き戻すだけで、状態を変える
// 目的の処理ではない。並行して動くジョブや状態変更ボタンが変えた状態を、
// 保存のたびに古い値へ戻してしまう不具合の対策として Omit した。状態を
// 変えるのは UpdateStatus(状態変更API・ジョブ経由)だけにする。
func (p *Project) Update(ctx context.Context) (err error) {
	p.OfferingType = offeringType(p.IsInvestorLimit)

	query := factory.DB(ctx).Nullable("company_info_html", "dividend_type", "planed_operation_start_date", "planed_operation_end_date", "fund_number").UseBool("posted", "block_invest", "is_wait_invest", "is_investor_limit", "is_manual_invest", "is_amount_limit", "is_visible_investment_info", "is_closed_fund", "project_progress_percent").Omit("project_status")
	_, err = query.ID(p.Id).Update(p)
	return err
}

// UpdateForAdminDetail は、管理画面のファンド詳細保存(admin-api
// controllers/project.go の Update)専用の保存メソッド(MIR-1314)。
// project_status に加えて is_wait_invest も書かない。画面にこの2項目を
// 編集する欄は無く、開いた時点の値をフォームの初期値としてそのまま送り
// 返しているだけなので、書き戻すと並行した変更を上書きしてしまう。
// 状態を変えるのは UpdateStatus・ReopenFunding だけにする。募集再開時に
// キャンセル待ち(is_wait_invest)を自動解除する処理は、この保存では行わない
// (MIR-1314設計で解除しない方針に決定済み)。
//
// expectedDetailVersion が nil でない場合は、detail_version がその値のときだけ
// 書き込み、別の保存が先に確定していれば何も書かずに false を返す(MIR-1492)。
// 呼び出し元の事前チェックはロックしない読み取りのため、ほぼ同時に始まった
// 2つの保存がどちらも事前チェックを通ることがある。後から書く側の UPDATE は
// 先の保存の確定を待ってから最新の値で条件を判定するので、ここで弾ける。
func (p *Project) UpdateForAdminDetail(ctx context.Context, expectedDetailVersion *int64) (updated bool, err error) {
	p.OfferingType = offeringType(p.IsInvestorLimit)

	query := factory.DB(ctx).Nullable("company_info_html", "dividend_type", "planed_operation_start_date", "planed_operation_end_date", "fund_number").UseBool("posted", "block_invest", "is_wait_invest", "is_investor_limit", "is_manual_invest", "is_amount_limit", "is_visible_investment_info", "is_closed_fund", "project_progress_percent").Omit("project_status", "is_wait_invest").
		// MIR-1492: 同時編集の検知用に保存回数を増やす(DetailVersion のコメント参照)。
		// 必ず値が変わるため、条件に合った行は影響行数に数えられる。
		Incr("detail_version").
		ID(p.Id)
	if expectedDetailVersion != nil {
		query = query.Where("detail_version = ?", *expectedDetailVersion)
	}
	affected, err := query.Update(p)
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}

// allowedFromStatuses は、行き先の状態ごとに、UpdateStatus が更新前の状態
// として許す遷移元を返す(MIR-1314)。表に無い行き先(STANDBY_FUNDINGや
// 不正な値)は空を返す。「募集終了→募集中」(ファンド詳細の保存で終了日時を
// 延長したときの再開)はここに含めない。状態変更API経由の遷移ではなく、
// ReopenFunding が別に扱う専用の遷移のため。
func allowedFromStatuses(to PROJECT_STATUS) []PROJECT_STATUS {
	switch to {
	case OPEN_FUNDING:
		return []PROJECT_STATUS{STANDBY_FUNDING}
	case END_FUNDING:
		return []PROJECT_STATUS{OPEN_FUNDING}
	case SUCCESS_FUNDING, FAIL_FUNDING:
		return []PROJECT_STATUS{END_FUNDING}
	case OPERATION_FUNDING:
		return []PROJECT_STATUS{SUCCESS_FUNDING}
	case OPERATION_END_FUNDING:
		return []PROJECT_STATUS{OPERATION_FUNDING}
	case CLOSE_FUNDING:
		return []PROJECT_STATUS{OPERATION_END_FUNDING}
	case INVALID_FUNDING:
		return []PROJECT_STATUS{STANDBY_FUNDING, OPEN_FUNDING, END_FUNDING}
	default:
		return nil
	}
}

// UpdateStatus は、p.ProjectStatus(呼び出し元が既に書き換えた行き先の状態)
// への遷移を、DBの現在の状態が allowedFromStatuses の表にある遷移元で
// あることを条件にして行う(MIR-1314)。画面を開いている間にジョブや別の
// 管理者が状態を変えていた場合、影響行数が0件になるので
// ErrorProjectStatusChanged を返す。行き先が表に無い値なら、更新前に
// 既存の "not supported status" を返す(呼び出し元 ChangeStatus の default
// ケースと同じ文言)。
func (p *Project) UpdateStatus(ctx context.Context) (err error) {
	return p.updateStatus(ctx, "", nil)
}

// updateStatusForJob は、UpdateStatus と同じ条件付き更新に、ジョブが一覧を
// 読んだときと同じ時刻条件を追加する(MIR-1314)。一覧取得後に管理者が
// start_at/end_at を変えていた場合や、同じジョブが二重に実行された場合を
// 弾くため。
func (p *Project) updateStatusForJob(ctx context.Context, jobWhere string, jobArg interface{}) error {
	return p.updateStatus(ctx, jobWhere, jobArg)
}

func (p *Project) updateStatus(ctx context.Context, jobWhere string, jobArg interface{}) error {
	from := allowedFromStatuses(p.ProjectStatus)
	if len(from) == 0 {
		// ErrorTemplate.NewはSprintf(t.Message, v...)でメッセージを組み立てる
		// ため、ひな形にプレースホルダーが無いのに第2引数以降へ文言を渡すと
		// "%!(EXTRA string=...)"が末尾に付く。詳細はerr引数(Details)側に含める。
		return factory.ErrorParameter.New(errors.New("not supported status"))
	}

	fromValues := make([]interface{}, len(from))
	for i, s := range from {
		fromValues[i] = string(s)
	}

	query := factory.DB(ctx).Table("project").
		ID(p.Id).In("project_status", fromValues...)
	if jobWhere != "" {
		query = query.Where(jobWhere, jobArg)
	}

	cols := []string{"project_status"}
	if p.ProjectStatus == END_FUNDING {
		cols = append(cols, "final_funding_amount")
	}

	affected, err := query.Cols(cols...).Update(p)
	if err != nil {
		return err
	}
	if affected == 0 {
		return factory.ErrorProjectStatusChanged.New(errors.New("project status already changed"))
	}
	return nil
}

// ReopenFunding は、募集終了(END_FUNDING)のファンドを募集中(OPEN_FUNDING)へ
// 戻す(MIR-1314)。ファンド詳細の保存で終了日時が延長されたとき(画面の状態に
// 関わらず)に呼ばれる。状態変更API(UpdateStatus)のallowedFromStatusesには
// 含めない専用の遷移で、事前の状態チェックはしない(条件付き更新のWHERE句
// だけで判定する)。戻り値は実際に切り替わったか(影響行数>0)を表す。呼び出し
// 元は、画面が募集終了を表示していたのにfalseだった場合だけ
// ErrorProjectStatusChangedを返す(それ以外の状態を表示していた場合は、
// 再開できなくても保存を成功として扱う)。
func (p *Project) ReopenFunding(ctx context.Context) (bool, error) {
	affected, err := factory.DB(ctx).Table("project").ID(p.Id).
		Where("project_status = ?", string(END_FUNDING)).
		Cols("project_status").Update(&Project{ProjectStatus: OPEN_FUNDING})
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}

func (Project) Delete(ctx context.Context, id int64) (err error) {
	_, err = factory.DB(ctx).ID(id).Delete(&Project{})
	return err
}

func (p *Project) CheckShareQty(qty int64, amt int64) bool {
	if p.TargetShareQty < p.CurrentShareQty+qty {
		return false
	}
	if p.TargetShareAmount < p.CurrentShareAmount+amt {
		return false
	}
	return true
}

func (p Project) HasReachedTarget() bool {
	if p.CurrentShareQty >= p.TargetShareQty {
		return true
	}
	if p.CurrentShareAmount >= p.TargetShareAmount {
		return true
	}
	return false
}

// isWaitInvestFinalPortion は、繰上げ(isWaitInvest=true)によってその待機者の
// 繰上げ可能残をちょうど使い切る投資(=この待機者への最後の繰上げ)かどうか(MIR-1068)。
// 申請10口・最小投資口数3口で8口→残り2口のように、正規の申込の残りが最小投資口数を
// 下回るケースを救うためのスキップで、投資提案(isProposalInvest)経路の既存の
// スキップに倣う。部分繰上げ(投資後も繰上げ可能残が残る)には適用しない。
func (p *Project) InvestProject(uid int64, qty int64, price int64, amt int64, isWaitInvest bool, isProposalInvest bool, isWaitInvestFinalPortion bool) (*Share, error) {
	if !p.Posted {
		return nil, factory.ErrorNotPostedProject.New(errors.New("Not Posted"))
	}

	if !p.IsInvestmentOpen(isWaitInvest) {
		return nil, factory.ErrorNotOpenFunding.New(errors.New("Not OPEN_FUNDING"))
	}

	if !((isProposalInvest && p.MinShareQty > p.RemainingQty) || (isWaitInvest && isWaitInvestFinalPortion)) {
		if p.MinShareQty > qty {
			// 投資口数が最小投資口数未満の場合はエラー
			// NOTE: 投資提案による投資かつ、残口数が最小投資口数未満の場合は、最小投資口数チェックをスキップ
			// NOTE: 繰上げによってその待機者の繰上げ可能残を使い切る場合も、同じ理由でスキップ(MIR-1068)
			return nil, factory.ErrorLessThanMinQuantity.New(nil)
		}
	}

	if p.SharePrice != price {
		return nil, factory.ErrorNotMatchInvest.New(nil)
	}
	if p.TargetShareQty < p.CurrentShareQty+qty {
		return nil, factory.ErrorInsufficientQty.New(nil)
	}

	if p.TargetShareAmount < p.CurrentShareAmount+amt {
		return nil, factory.ErrorInsufficientAmount.New(nil)
	}

	return &Share{
		ProjectId:                   p.Id,
		CategoryType:                p.CategoryType,
		ProjectType:                 p.ProjectType,
		OwnerId:                     uid,
		UpdateId:                    uid,
		Desc:                        p.Name,
		InvestStatus:                REQUEST_INVEST,
		InvestDate:                  jsonutils.JSONTime(time.Now()),
		Amount:                      amt,
		Qty:                         qty,
		Price:                       price,
		ContractBeforePageCheckDate: jsonutils.JSONTime(time.Now()),
		SilentPartnershipsCheckDate: jsonutils.JSONTime(time.Now()),
	}, nil
}

// GetByProjectType はユーザー向けのファンド一覧を返す。
func (Project) GetByProjectType(ctx context.Context, listType PROJECT_STATUS_LIST_TYPE, offset, limit int, sortby string, order string, projectName *string, startDate, endDate string) (int64, []Project, error) {
	q := factory.DB(ctx)

	var totalCount int64
	var items []Project
	q = q.Table("project").Select(PROJECT_LIST_CLOUMN_LIST + "," + PUBLIC_COLUMN)
	q = q.Join("LEFT", "project_view_count", "project.id = project_view_count.project_id")
	q = q.Where("is_closed_fund = false AND posted = true")
	switch listType {
	case ALL_PROJECT_LIST_TYPE:
		{
			q.Where("NOT project_status = ?", INVALID_FUNDING)
		}
	case STANDBY_FUNDING_LIST_TYPE:
		{
			q.Where("project_status = ?", STANDBY_FUNDING)
		}
	case OPEN_FUNDING_LIST_TYPE:
		{
			q.Where("project_status = ?", OPEN_FUNDING)
		}
	case END_FUNDING_LIST_TYPE:
		{
			q.Where("(project_status = ? OR  project_status = ? OR project_status = ? )", END_FUNDING, SUCCESS_FUNDING, FAIL_FUNDING)
		}
	case OPERATION_FUNDING_LIST_TYPE:
		{
			q.Where("project_status = ?", OPERATION_FUNDING)
		}
	case OPERATION_END_FUNDING_LIST_TYPE:
		{
			q.Where("project_status = ?", OPERATION_END_FUNDING)
		}
	case CLOSE_FUNDING_LIST_TYPE:
		{
			q.Where("project_status = ?", CLOSE_FUNDING)
		}
	default:
		{
			q.Where("NOT project_status = ?", INVALID_FUNDING)
		}

	}

	ApplyProjectListSortOrder(ctx, q, sortby, order)

	if err := modelhelper.SearchLikeQuery(q, "project", "name", projectName); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SetBetweenDate(q, "project.created", startDate, endDate); err != nil {
		factory.Logger(ctx).Error(err)
	}
	totalCount, err := q.Limit(limit, offset).FindAndCount(&items)
	if err != nil {
		return 0, nil, err
	}

	for i := range items {
		if items[i].CapitalRaisingEntityId == nil {
			continue
		}

		capitalRaisingEntity, err := CapitalRaisingEntity{}.GetById(ctx, *items[i].CapitalRaisingEntityId)
		if err != nil {
			return 0, nil, err
		}

		items[i].CapitalRaisingEntity = capitalRaisingEntity
	}

	return totalCount, items, nil
}

// プロジェクト別投資者リスト照会
// Deprecated: Use ShareMysqlRepository.GetInvestors
func (p *Project) GetInvestors(ctx context.Context) ([]User, error) {
	var userList []User

	err := factory.DB(ctx).Table("share").Select("share.project_id, user.*").Join("INNER", "user", "share.owner_id = user.id").
		Where("share.project_id = ?", p.Id).GroupBy("share.owner_id").Find(&userList)
	if err != nil {
		return nil, err
	}

	return userList, err
}

// プロジェクト別ニュースレターを許可した投資待機または投資完了した投資者リスト
// Deprecated: Use ShareMysqlRepository.GetInvestors
func (p *Project) GetAcceptedNewsletterAndCompletedInvestors(ctx context.Context) ([]User, error) {
	var userList []User

	err := factory.DB(ctx).Table("share").Select("share.project_id, user.*").
		Join("INNER", "user", "share.owner_id = user.id AND (share.invest_status = 'REQUEST_INVEST' OR share.invest_status = 'SUCCESS_INVEST')").
		Where("share.project_id = ? and user.is_newsletter = true", p.Id).GroupBy("email").Find(&userList)
	if err != nil {
		return nil, err
	}

	return userList, err
}

func (Project) PrepareAndReturnCsv(items []Project) [][]string {
	// if items == nil {
	//	return nil
	// }

	type project struct {
		Id                int64              `json:"id" i18n:"model.project.id"`                                   // ID
		Posted            string             `json:"posted" i18n:"model.project.post"`                             // 掲示有無
		Status            string             `json:"status" i18n:"model.project.project_status"`                   // 募集状態
		ProjectType       string             `json:"project_type" i18n:"model.project.project_type"`               // 投資タイプ
		CategoryType      string             `json:"category_type" i18n:"model.project.category_type"`             // 発電タイプ
		Name              string             `json:"name" i18n:"model.project.project_name"`                       // 商品名
		Period            string             `json:"start_date" i18n:"model.project.period"`                       // ファンディング期間
		TargetShareAmount int64              `json:"target_share_amount" i18n:"model.project.target_share_amount"` // 上限募金額
		Created           jsonutils.JSONTime `json:"created" i18n:"model.project.created"`                         // 登録日付
	}

	var requests []project
	for _, item := range items {
		c := project{
			Id:                item.Id,
			ProjectType:       string(item.ProjectType),
			CategoryType:      string(item.CategoryType),
			Name:              item.Name,
			TargetShareAmount: item.TargetShareAmount,
			Created:           item.Created,
		}

		if item.Posted {
			c.Posted = i18nutils.DefaultTranslate("commoncode.PROJECT_PUBLISH_TYPE.PROJECT_PUBLISH_POSTED")
		} else {
			c.Posted = i18nutils.DefaultTranslate("commoncode.PROJECT_PUBLISH_TYPE.PROJECT_PUBLISH_UNPOSTED")
		}

		c.Status = i18nutils.DefaultTranslate("commoncode.PROJECT_STATUS." + string(item.ProjectStatus))
		if item.BlockInvest {
			c.Status += " -" + i18nutils.DefaultTranslate("commoncode.PROJECT_CHANGED_CONTENT_TYPE.PROJECT_CHANGED_CONTENT_LIMITED_INVEST") + "-"
		}

		c.Period = item.StartAt.ToString() + " ~ " + item.EndAt.ToString()

		if value, ok := reflect.ValueOf(&item.ProjectType).Elem().Interface().(PROJECT_TYPE); ok {
			c.ProjectType = "commoncode.PROJECT_TYPE." + string(value)
		}
		if value, ok := reflect.ValueOf(&item.CategoryType).Elem().Interface().(PROJECT_CATEGORY); ok {
			c.CategoryType = "commoncode.PROJECT_CATEGORY." + string(value)
		}

		requests = append(requests, c)
	}

	cols := factory.ReflectJsonFields(reflect.ValueOf(&project{}).Elem())
	if len(cols) <= 0 {
		return nil
	}

	records := [][]string{}
	records = append(records, cols)

	for _, v := range requests {
		records = append(records, factory.DataWriteToCsv(reflect.ValueOf(&v).Elem()))
	}

	return records
}

func (Project) DownloadCsv(ctx context.Context, w http.ResponseWriter, sortby, order string, projectId, projectStatus, targetShareAmount, projectName, projectType, categoryType *string, blockInvest *bool, startDate, endDate string, outOfDate bool, posted *bool) error {
	q := factory.DB(ctx)

	var items []Project
	q.Table("project").Select(PROJECT_LIST_CLOUMN_LIST)
	q.Join("LEFT", "project_view_count", "project.id = project_view_count.project_id")
	q.Join("LEFT", "share", "project.id = share.project_id")

	q.GroupBy("project.id")
	if outOfDate {
		if err := OutOfDateQuery(q); err != nil {
			factory.Logger(ctx).Error(err)
		}
	}

	if err := modelhelper.SetSortOrder(q, sortby, order); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SearchEqQuery(q, "project", "id", projectId); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SearchEqQuery(q, "project", "project_status", projectStatus); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SearchEqQuery(q, "project", "target_share_amount", targetShareAmount); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SearchLikeQuery(q, "project", "name", projectName); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SearchBoolQuery(q, "project", "block_invest", blockInvest); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SearchBoolQuery(q, "project", "posted", posted); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SearchEqQuery(q, "project", "project_type", projectType); err != nil {
		factory.Logger(ctx).Error(err)
	}
	if err := modelhelper.SearchEqQuery(q, "project", "category_type", categoryType); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SetBetweenDate(q, "project.created", startDate, endDate); err != nil {
		factory.Logger(ctx).Error(err)
	}

	err := q.Find(&items)
	if err != nil {
		return err
	}

	records := Project{}.PrepareAndReturnCsv(items)
	csvutils.WriteHeader(w, "energyshares_product_list.csv")

	return csv.NewWriter(w).WriteAll(records)
}

// 商品の投資者配当金の合計
func (p *Project) SumInvestorsDividend(projectShareQty, sharePerDividendAmount int64) (int64, error) {
	sumProjectDividend := sharePerDividendAmount * projectShareQty

	return sumProjectDividend, nil
}

func (Project) GetShareRemainingQty(ctx context.Context, id int64) (*Project, error) {
	var v Project
	if has, err := factory.DB(ctx).Table("project").Select("project.is_wait_invest, project.target_share_qty - (SELECT IFNULL(SUM(qty),0) FROM share WHERE share.project_id = project.id  AND invest_status IN ('SUCCESS_INVEST','REQUEST_INVEST')) AS remaining_qty").ID(id).Get(&v); err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}

	return &v, nil
}

func (Project) GetCurrentShareAmountsForUserView(ctx context.Context, projectIds []int64) (map[int64]int64, error) {
	result := make(map[int64]int64, len(projectIds))
	if len(projectIds) == 0 {
		return result, nil
	}

	type projectAmount struct {
		ProjectId int64 `xorm:"project_id"`
		Amount    int64 `xorm:"amount"`
	}

	shareStatuses := []INVEST_STATUS{SUCCESS_INVEST, REQUEST_INVEST}
	var shareAmounts []projectAmount
	// shareの論理削除は考慮していない。PROJECT_LIST_CLOUMN_LIST・PROJECT_DETAIL_COLUMN_LIST
	// 等のcurrent_share_amount/current_share_qty/remaining_qtyも同じ条件で論理削除を
	// 考慮しておらず、ここだけ考慮すると画面間で金額が食い違う。MIR-1163(旧列フォールバック
	// 削除)のスコープ外の既存の設計であり、直すなら関連する集計ロジックすべてを合わせて
	// 見直す必要がある。
	if err := factory.DB(ctx).Table("share").
		Select("project_id, IFNULL(SUM(amount),0) AS amount").
		In("project_id", projectIds).
		In("invest_status", shareStatuses).
		GroupBy("project_id").
		Find(&shareAmounts); err != nil {
		return nil, err
	}
	for _, amount := range shareAmounts {
		result[amount.ProjectId] += amount.Amount
	}

	waiterStatuses := []INVESTMENT_WAITING_TYPE{WAITING, DEPOSIT_REQUEST, PROPOSAL}
	var waitingAmounts []projectAmount
	// 当選済み(DEPOSIT_REQUEST/PROPOSAL)の待機者はshare側の金額と二重に計上されるため除外
	// する。当選の有無はinvestment_waiters_winning側で判定する(MIR-1118)。旧列
	// investment_waiters.share_idを条件に含める案も検討したが、バックフィル(MIR-1163)
	// 自体が対策であり(INSERT...SELECTは部分成功しないため、成功すれば取りこぼしは無い)、
	// ここへ追加のSQL条件を足す必要は無い。取り返しのつかない操作(旧列のカラム自体を削除する
	// PR4)の前に、取りこぼしが無いことを人が確認する運用にしている(MIR-1163のPR4節参照)。
	// 集計先のprojectAmountにDeletedフィールドが無くxormのsoft-deleteフィルタが自動付与
	// されないため、論理削除済みの待機者を拾わないようNotDeletedConditionを明示している。
	// share側のクエリ(直前)はこの対応をしていない(MIR-1163のスコープ外の既存問題のため、
	// 別途判断する)。
	if err := factory.DB(ctx).Table("investment_waiters").
		Select("project_id, IFNULL(SUM(amount),0) AS amount").
		In("project_id", projectIds).
		In("status", waiterStatuses).
		Where(NotDeletedCondition("investment_waiters")+" AND NOT "+InvestmentWaitersWinningExistsByWaiterIdCond).
		GroupBy("project_id").
		Find(&waitingAmounts); err != nil {
		return nil, err
	}
	for _, amount := range waitingAmounts {
		result[amount.ProjectId] += amount.Amount
	}

	return result, nil
}

// プロジェクト別投資者照会
func (p *Project) GetProjectInvestorsByProjectStatus(ctx context.Context) ([]Investment, error) {
	// クリーンアーキテクチャ作業時に改善要望
	// 現在使用するフィールドはuserテーブルのname、email、およびshareテーブルのinvest_dateのみを使用する。
	// さらに現在user.nameがInvestment.Project.Nameとして解析されているようだ
	var investors []Investment

	var investStatuses []INVEST_STATUS
	switch p.ProjectStatus {
	case INVALID_FUNDING: // 無効
		investStatuses = []INVEST_STATUS{FAIL_INVEST, SUCCESS_INVEST, REQUEST_INVEST, OPERATION_INVEST}

	case END_FUNDING: // 募集終了
		investStatuses = []INVEST_STATUS{REQUEST_INVEST, SUCCESS_INVEST}

	case SUCCESS_FUNDING, FAIL_FUNDING, OPERATION_FUNDING: // 投資成立、投資不成立、投資運用中
		investStatuses = []INVEST_STATUS{SUCCESS_INVEST}

	case CLOSE_FUNDING: // 償還完了。最後の償還をジョブが払うと出資は償還済みになるため、償還済みの出資者も宛先に含める(MIR-272)。この集合は出資の付け替えの前後で変わらない。
		investStatuses = []INVEST_STATUS{OPERATION_INVEST, REFUND_INVEST}

	case OPERATION_END_FUNDING: // 運用終了はメールを送らない(MIR-272)。
		return []Investment{}, nil
	default: // STANDBY_FUNDING、OPEN_FUNDINGはメールを送信しない。
		return []Investment{}, nil
	}

	err := factory.DB(ctx).Table("share").Select("share.*, user.*").Join("INNER", "user", "share.owner_id = user.id").
		Where("share.project_id = ?", p.Id).In("share.invest_status", investStatuses).GroupBy("owner_id").Find(&investors)
	if err != nil {
		return nil, err
	}

	return investors, err
}

// JobStatusCondition は、ジョブがChangeStatusを呼ぶときに追加する、一覧取得時
// と同じ時刻条件(MIR-1314)。一覧取得後に管理者がstart_at/end_atを変えていた
// 場合や、同じジョブが二重に実行された場合を、この条件付き更新のWHERE句で
// 弾くために使う。ボタン経由の呼び出し(条件無し)と区別するため、
// ChangeStatusの可変長引数として渡す。
type JobStatusCondition struct {
	Where string      // 例: "start_at <= ?"
	Arg   interface{}
}

func (p *Project) ChangeStatus(ctx context.Context, status PROJECT_STATUS, updateId int64, jobCond ...JobStatusCondition) error {
	switch status {
	case END_FUNDING:
		// NOTE: 募集終了ステータスではキャンセル待ちフラグを維持する（手動延長時に必要なため）
		investorCount, err := p.GetCurrentInvestorCount(ctx, p.Id)
		if err != nil {
			return factory.ErrorDB.New(errors.Wrap(err, "fail to get current investor count"))
		}
		if investorCount >= TargetInvestorCount {
			p.OfferingType = OFFERING_PUBLIC
		}

		if p.SelectionType != PROJECT_SELECTION_LOTTERY {
			// 募集終了時点の申込金額を保存する。
			// 集計は GROUP BY のため、投資も申込も無いとキー自体が作られない。
			// 未登録なら申込金額0として扱う。存在しないことをエラーにすると、
			// 投資が1件も無いファンドを募集終了にできない。
			amountMapByProjectId, err := Project{}.GetCurrentShareAmountsForUserView(ctx, []int64{p.Id})
			if err != nil {
				return err
			}
			p.FinalFundingAmount = amountMapByProjectId[p.Id]
		}

	case OPERATION_END_FUNDING: // 出資実行が済んでいない出資が残っている場合、運用終了にできない(MIR-272)。
		if err := lockProjectRowForTransition(ctx, p.Id, OPERATION_END_FUNDING); err != nil {
			return err
		}
		blocking, err := countInvestmentsBlockingOperationEnd(ctx, p.Id)
		if err != nil {
			return factory.ErrorDB.New(errors.Wrap(err, "fail to count investments blocking operation end"))
		}
		if err := checkOperationEnd(blocking); err != nil {
			return err
		}

	case CLOSE_FUNDING: // 実施されていない分配/返済スケジュールがある場合、償還完了にできない。元本が残っているときは確認が要る。
		if err := lockProjectRowForTransition(ctx, p.Id, CLOSE_FUNDING); err != nil {
			return err
		}
		closeCheck, err := Project{}.GetCloseCheck(ctx, p.Id, true)
		if err != nil {
			return factory.ErrorDB.New(err)
		}
		if err := closeCheck.Check(isRemainingPrincipalConfirmed(ctx)); err != nil {
			return err
		}

	case INVALID_FUNDING: // 募集開始前、募集中、募集終了の状態でのみ無効処理が可能
		if p.ProjectStatus != STANDBY_FUNDING && p.ProjectStatus != OPEN_FUNDING && p.ProjectStatus != END_FUNDING {
			return factory.ErrorParameter.New(
				fmt.Errorf("cannot change status to INVALID_FUNDING from current status: %s, allowed statuses are STANDBY_FUNDING, OPEN_FUNDING, and END_FUNDING", p.ProjectStatus),
			)
		}
	}

	beforeProject := *p
	p.ProjectStatus = status

	var updateErr error
	if len(jobCond) > 0 {
		updateErr = p.updateStatusForJob(ctx, jobCond[0].Where, jobCond[0].Arg)
	} else {
		updateErr = p.UpdateStatus(ctx)
	}
	if updateErr != nil {
		return updateErr
	}
	if err := p.ChangeStatusSendEmail(ctx); err != nil {
		return factory.ErrorAWSSendMail.New(err, "fail Change Status Send Email ")
	}

	switch p.ProjectStatus {
	case SUCCESS_FUNDING:
		err := Share{}.ChangeOfInvestStatusByProject(ctx, p.ProjectStatus, p.Id, updateId)
		if err != nil {
			return factory.ErrorDB.New(err)
		}

	case CLOSE_FUNDING: // 残っている運用中の出資を償還済みにする(MIR-272)。
		if err := (Share{}).RefundOperationInvestsByProject(ctx, p.Id, updateId); err != nil {
			return factory.ErrorDB.New(err)
		}

	case FAIL_FUNDING:
		fundFailureCtx := WithFundFailureProcessing(ctx)

		iwList, err := InvestmentWaiters{}.GetByProjectIdAndStatusesWithoutGrouping(fundFailureCtx, p.Id, []INVESTMENT_WAITING_TYPE{WAITING, DEPOSIT_REQUEST, PROPOSAL})
		if err != nil {
			return errors.Wrap(err, "fail to get by project id and status")
		}
		sortFundFailureInvestmentWaiters(iwList)
		for _, value := range iwList {
			err = value.ChangeStatus(fundFailureCtx, WAITING_END, updateId)
			if err != nil {
				return factory.ErrorDB.New(errors.Wrap(err, "InvestmentWaiters change status failed."))
			}
		}
		_, shares, err := Share{}.GetByProjectId(fundFailureCtx, p.Id)
		if err != nil {
			return err
		}
		targetShares := make([]Share, 0, len(shares))
		for _, share := range shares {
			// 商品が募集中の場合のみ無効処理が可能
			if share.InvestStatus == SUCCESS_INVEST || share.InvestStatus == REQUEST_INVEST {
				targetShares = append(targetShares, share)
			}
		}
		sortFundFailureShares(targetShares)
		for _, share := range targetShares {
			if err := share.UpdateInvestStatus(fundFailureCtx, FAIL_INVEST, updateId, NOT_IMMEDIATE); err != nil {
				return factory.ErrorDB.New(errors.Wrap(err, "share status changed failed"))
			}
		}
		if err := CreateFundFailureEvent(fundFailureCtx, p); err != nil {
			return factory.ErrorDB.New(err)
		}

	case INVALID_FUNDING:
		iwList, err := InvestmentWaiters{}.GetByProjectIdAndStatusesWithoutGrouping(ctx, p.Id, []INVESTMENT_WAITING_TYPE{PROPOSAL})
		if err != nil {
			return err
		}
		for _, value := range iwList {
			err = value.ChangeStatus(ctx, WAITING_END, updateId)
			if err != nil {
				return factory.ErrorDB.New(errors.Wrap(err, "investment waiters status changed failed"))
			}
		}

		_, shares, err := Share{}.GetByProjectId(ctx, p.Id)
		if err != nil {
			return errors.Wrap(err, "share status changed failed")
		}
		for _, share := range shares {
			// 상품이 모집중인 경우만 무효 처리가 가능함
			if share.InvestStatus == SUCCESS_INVEST || share.InvestStatus == REQUEST_INVEST {
				if err := share.UpdateInvestStatus(ctx, CANCEL_INVEST, updateId, NOT_IMMEDIATE); err != nil {
					return factory.ErrorDB.New(errors.Wrap(err, "share status changed failed"))
				}
			}
		}

	case STANDBY_FUNDING, OPEN_FUNDING, OPERATION_FUNDING, OPERATION_END_FUNDING, END_FUNDING:

	// do nothing(運用終了は出資の状態を変えない。償還で償還済みになる。MIR-272)

	default:
		// ErrorTemplate.Newの第2引数以降はSprintfの差し込み値であり、ひな形に
		// プレースホルダーが無いと"%!(EXTRA string=...)"が末尾に付く。詳細は
		// err引数側に含める。
		return factory.ErrorParameter.New(errors.New("not supported status"))
	}

	_, err := ProjectInfoHistory{}.Create(ctx, &beforeProject, p, updateId)
	if err != nil {
		return factory.ErrorDB.New(errors.Wrap(err, "share status changed failed"))
	}

	return nil
}

// sendOperationCloseMail は、償還完了の「払戻し完了のご案内」を1人へ送る。単体テストで差し替える。
var sendOperationCloseMail = func(ctx context.Context, email, name, projectName string) error {
	return Mail{}.ProjectStatusOperationCloseMail(ctx, serviceutils.GetServiceInfoMail(), email, name, projectName)
}

// メモ：現在は各投資者に1つずつメールを送信している。統合メールサービスを改善して、一括で送信できるように改善する必要がある。
func (p *Project) ChangeStatusSendEmail(ctx context.Context) error {
	investors, err := p.GetProjectInvestorsByProjectStatus(ctx)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	// 償還完了のメールは、コミットの後に送る(MIR-272)。トランザクションの途中で送ると、後段の失敗で
	// 状態は巻き戻るのにメールだけ出たままになり、押し直すと全員に2通目が届く。宛先はここ(トランザクション
	// の中)で取り、送信だけを後ろへ回す。
	if p.ProjectStatus == CLOSE_FUNDING {
		projectName := p.Name
		return factory.AfterCommit(ctx, func(afterCommitCtx context.Context) error {
			failed := 0
			for _, investor := range investors {
				if err := sendOperationCloseMail(afterCommitCtx, investor.Email, investor.Name, projectName); err != nil {
					failed++
					// 氏名・アドレスは出さず、会員 ID だけで追えるようにする。
					factory.Logger(afterCommitCtx).WithError(err).WithField("owner_id", investor.Share.OwnerId).Error("fail to send operation close mail")
				}
			}
			if failed > 0 {
				return fmt.Errorf("fail to send operation close mail: %d of %d recipients", failed, len(investors))
			}
			return nil
		})
	}

	var mail Mail
	for _, investor := range investors {
		switch p.ProjectStatus {
		case END_FUNDING:
			if err := mail.ProjectStatusEndMail(ctx, serviceutils.GetServiceInfoMail(), investor.Email, investor.Name, p.Name, jsonutils.JSONTime.ToString(investor.InvestDate)); err != nil {
				return factory.ErrorAWSSendMail.New(err)
			}
		case SUCCESS_FUNDING:
			if err := mail.ProjectStatusSuccessMail(ctx, serviceutils.GetServiceInfoMail(), investor.Email, investor.Name, p.Name, serviceutils.GetUserServiceDomain()+"/product/view"+"/"+strconv.Itoa(int(p.Id))+"?tab=news"); err != nil {
				return factory.ErrorAWSSendMail.New(err)
			}
		case FAIL_FUNDING:
			if err := mail.ProjectStatusFailMail(ctx, serviceutils.GetServiceInfoMail(), investor.Email, investor.Name, p.Name); err != nil {
				return factory.ErrorAWSSendMail.New(err)
			}
		case OPERATION_FUNDING:
			if err := mail.ProjectStatusOperatingMail(ctx, serviceutils.GetServiceInfoMail(), investor.Email, investor.Name, p.Name, serviceutils.GetUserServiceDomain()+"/product/view"+"/"+strconv.Itoa(int(p.Id))+"?tab=news"); err != nil {
				return factory.ErrorAWSSendMail.New(err)
			}
		}
	}

	// 成立の場合、待機中の投資待機者にもメールを追加
	if p.ProjectStatus == SUCCESS_FUNDING {
		investmentWaiters, err := InvestmentWaiters{}.GetByProjectIdAndStatuses(ctx, p.Id, []INVESTMENT_WAITING_TYPE{WAITING})
		if err != nil {
			return factory.ErrorDB.New(err)
		}
		for _, values := range investmentWaiters {
			if err := mail.InvestRecruitmentSuccessMail(ctx, serviceutils.GetServiceInfoMail(), values.UserEmail, values.UserName, p.Name); err != nil {
				return factory.ErrorAWSSendMail.New(err)
			}
		}
	}
	return nil
}

func (p Project) GetCurrentInvestorCount(ctx context.Context, projectId int64) (int64, error) {
	var results []map[string]string
	var err error

	results, err = factory.DB(ctx).QueryString("SELECT COUNT(*) AS cnt FROM (SELECT share.owner_id FROM share WHERE share.project_id = ?  AND invest_status IN (?, ?) GROUP BY share.owner_id) AS a", projectId, SUCCESS_INVEST, REQUEST_INVEST)
	if err != nil {
		return 0, nil
	}

	res, err := strconv.ParseInt(results[0]["cnt"], 10, 64)
	if err != nil {
		return 0, err
	}
	return res, nil
}

func (p Project) GetInvestmentDividendProjects(ctx context.Context, userId int64, sortby string, order string, offset int, limit int) ([]Project, error) {
	var projects []Project
	q := factory.DB(ctx).Table("project").Select("project.*").
		Join("INNER", "dividend", "dividend.project_id = project.id").
		Join("INNER", "repayment_dividend_schedule", "repayment_dividend_schedule.project_id = project.id").
		Where("dividend.user_id = ? AND repayment_dividend_schedule.status =?", userId, RETURN_COMPLETED).GroupBy("project.id")
	if err := modelhelper.SetSortOrder(q, sortby, order); err != nil {
		factory.Logger(ctx).Error(err)
	}

	err := q.Limit(limit, offset).Find(&projects)
	if err != nil {
		return nil, err
	}

	return projects, err
}

func (Project) GetProjectTypeById(ctx context.Context, id int64) (*PROJECT_TYPE, error) {
	var v PROJECT_TYPE
	if has, err := factory.DB(ctx).ID(id).Get(&v); err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}
	return &v, nil
}

func (p *Project) CheckInvestmentWait(ctx context.Context, amount, qty int64) error {
	// 投資待機条件
	// 目標募金額達成（取引状態：投資申請中＋投資完了＋現在申請金額＞目標金額）＋投資中（OPEN_FUNDING）の場合
	// 目標投資者数達成（取引状態：投資申請中＋投資完了投資者数≧目標投資者数）＋投資中（OPEN_FUNDING）の場合
	// 最小投資単位＞残投資単位＋投資中（OPEN_FUNDING）の場合

	currentInvestorCount, err := p.GetCurrentInvestorCount(ctx, p.Id)
	if err != nil {
		return factory.ErrorDB.New(errors.Wrap(err, "fail Get Current Investor Count"))
	}
	if (p.CurrentShareAmount+amount >= p.TargetShareAmount || (p.IsInvestorLimit && (currentInvestorCount >= TargetInvestorCount)) || p.MinShareQty > p.RemainingQty-qty) && p.ProjectStatus == OPEN_FUNDING {
		p.IsWaitInvest = true
		if err := p.Update(ctx); err != nil {
			return factory.ErrorDB.New(err)
		}
	}

	return nil
}

var assetPath string

func (Project) CreateProjectCatalogPDF(ctx context.Context, p *Project) (*bytes.Buffer, error) {
	assetPath = pathutils.GetPathOfAssets()
	// var obj *s3manager.UploadOutput
	projectInfo := pdfgen.ProductCatalogInfo{
		ProjectName:          p.Name,
		ProjectType:          i18nutils.DefaultTranslate(string(p.ProjectType)),
		MainImageUrl:         p.MainImageUrl,
		Description:          p.Description,
		SpvName:              p.Spv.Name,
		SpvCorpNo:            p.Spv.CorpNo,
		SpvPhone:             p.Spv.Phone,
		SpvMail:              p.Spv.Email,
		SpvZipCode:           p.Spv.ZipCode,
		SpvProvince:          p.Spv.Province,
		SpvCity:              p.Spv.City,
		SpvAddr:              p.Spv.Address,
		SpvAddr2:             p.Spv.Address2,
		StatAt:               p.StartAt,
		EndAt:                p.EndAt,
		CategoryType:         i18nutils.DefaultTranslate(string(p.CategoryType)),
		Capacity:             p.Capacity,
		DividendRate:         strconv.FormatFloat(p.DividendRate, 'f', 1, 64), // 配当率
		OperationTerm:        strconv.Itoa(p.OperationTerm),                   //
		MinShareQty:          modelhelper.ConvertNumberToPrice(fmt.Sprint(p.MinShareQty)),
		SharePrice:           modelhelper.ConvertNumberToPrice(fmt.Sprint(p.SharePrice)),
		TargetMinShareAmount: modelhelper.ConvertNumberToPrice(fmt.Sprint(p.TargetMinShareAmount)),
		TargetShareAmount:    modelhelper.ConvertNumberToPrice(fmt.Sprint(p.TargetShareAmount)),
		TargetShareQty:       modelhelper.ConvertNumberToPrice(fmt.Sprint(p.TargetShareQty)),
		LoanAmount:           modelhelper.ConvertNumberToPrice(fmt.Sprint(p.TargetShareAmount)),
	}

	catalog, err := pdfgen.CreateCatelogPDF()
	if err != nil {
		return nil, err
	}
	header := func() {
		catalog.Image(assetPath+"/img/wave.png", 0, 0, 210, 0, false, "PNG", 0, "")
	}
	footer := func() {
		catalog.SetY(-8)
		catalog.SetFont("noto", "D", 9)
		catalog.SetTextColor(255, 255, 255)
		catalog.CellFormat(0, 4, fmt.Sprintf("%d", catalog.PageNo()-1),
			"", 0, "C", false, 0, "")
		catalog.AliasNbPages("")
	}

	cover, err := pdfgen.CreateCoverPage(catalog, projectInfo.ProjectName, projectInfo.SpvName, projectInfo.EndAt.Time().Format(time.DateOnly))
	if err != nil {
		return nil, err
	}

	infoPage, err := pdfgen.CreateInfoPage(catalog, projectInfo)
	spvTable := pdfgen.SpvInfoTable(catalog, projectInfo)

	catalog.AddPage()
	catalog.UseTemplate(*cover)
	catalog.SetHeaderFunc(header)
	catalog.AddPage()
	catalog.UseTemplate(*infoPage)
	catalog.UseTemplate(*spvTable)
	catalog.SetFooterFunc(footer)
	catalog.AddPage()

	investInfoTable := pdfgen.InvestInfoTableForEquity(catalog, projectInfo)
	catalog.UseTemplate(*investInfoTable)

	catalog.SetFooterFunc(footer)

	if catalog.Err() {
		return nil, catalog.Error()
	}
	buf := &bytes.Buffer{}
	catalog.Close()
	err = catalog.Output(buf)
	if err != nil {
		return nil, err
	}

	src := make([]io.ReadSeeker, 0, 0)
	src = append(src, bytes.NewReader(buf.Bytes()))
	var index int
	for _, i2 := range p.Attachments {
		if i2.FileExtension == ".pdf" {

			resp, err := http.Get(i2.FileUrl)
			if err != nil {
				return nil, err
			}
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}

			cfg := model.NewDefaultConfiguration()
			cfg.ValidationMode = model.ValidationRelaxed
			if err := pdfcpuapi.Validate(bytes.NewReader(body), cfg); err != nil {
				continue
			}

			index++
			cl := attachmentPdf(index)
			if cl.Err() {
				return nil, cl.Error()
			}
			buf2 := &bytes.Buffer{}
			cl.Close()
			err = cl.Output(buf2)
			if err != nil {
				return nil, err
			}
			src = append(src, bytes.NewReader(buf2.Bytes()))

			src = append(src, bytes.NewReader(body))
		}
	}
	buf2 := &bytes.Buffer{}

	cfg := model.NewDefaultConfiguration()
	cfg.ValidationMode = model.ValidationRelaxed

	pdfcpuapi.MergeRaw(src, buf2, true, cfg)

	if err := pdfcpuapi.Validate(bytes.NewReader(buf2.Bytes()), cfg); err != nil {
		return nil, err
	}

	return buf2, nil
}

func attachmentPdf(index int) *gofpdf.Fpdf {
	catelog, err := pdfgen.CreateCatelogPDF()
	if err != nil {
		return nil
	}

	header := func() {
		w, h := catelog.GetPageSize()
		l, r, t, _ := catelog.GetMargins()
		catelog.Image(assetPath+"/img/wave_top.png", 0, 0, 210, 0, false, "PNG", 0, "")
		catelog.Image(assetPath+"/img/wave_bottom.png", 0, h-24, 210, 0, false, "PNG", 0, "")

		catelog.SetY(h*0.5 - t)
		catelog.SetFont("noto", "D", 25)
		catelog.SetTextColor(0, 0, 0)
		catelog.CellFormat(w-(l+r), 25, fmt.Sprintf("%s %d", i18nutils.DefaultTranslate("pdf.attachment.reference"), index), "0", 0, "C", false, 0, "")

		catelog.Image(assetPath+"/img/logo_bottom.png", w-43, 6, 35, 0, false, "PNG", 0, "")
		catelog.Image(assetPath+"/img/logo.png", w-43, 6, 35, 0, false, "PNG", 0, "")
	}

	catelog.SetHeaderFunc(header)
	catelog.AddPage()

	return catelog
}

func (p Project) IsInvestmentOpen(isWaitInvest bool) bool {
	if isWaitInvest {
		// キャンセル待ち繰り上げ投資の場合は「募集中」と「募集終了」のときだけ投資可能
		if p.ProjectStatus != OPEN_FUNDING && p.ProjectStatus != END_FUNDING {
			return false
		}
	} else {
		// 通常投資は「募集中」のときだけ投資可能
		if p.ProjectStatus != OPEN_FUNDING {
			return false
		}
	}

	return true
}

func (Project) GetByClosedSlugForClosed(ctx context.Context, slug string, isPreview bool) (*Project, error) {
	var p Project
	q := factory.DB(ctx).Table("project").Select(PROJECT_DETAIL_COLUMN_LIST+","+PUBLIC_COLUMN).
		Where("project.closed_fund_slug = ? and project.is_closed_fund = ?", slug, true)

	if !isPreview {
		q.Where("project.posted = ?", true)
	}

	if has, err := q.Get(&p); err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}

	spv, err := Spv{}.GetById(ctx, p.SpvId)
	if err != nil {
		return nil, err
	} else {
		p.Spv = spv
	}

	if p.MainImageId > 0 {
		projectImages, err := ProjectImage{}.FindByProjectId(ctx, p.Id)
		if err != nil {
			return nil, err
		}
		for _, img := range projectImages {
			if img.Id == p.MainImageId {
				p.MainImage = img
			} else if img.Order == 0 {
				// 表示順:0のものは非表示のためスキップ
				continue
			} else {
				subImageInfo := SubImageInfo{
					SubImageId:  img.Id,
					SubImageUrl: img.FileUrl,
					SubImage:    img,
				}
				p.SubImageInfoList = append(p.SubImageInfoList, &subImageInfo)
			}
		}
		sort.Slice(p.SubImageInfoList, func(i, j int) bool {
			return p.SubImageInfoList[i].SubImage.Order < p.SubImageInfoList[j].SubImage.Order
		})
	}

	_, news, err := ProjectNews{}.GetPostedList(ctx, p.Id, 0, 30)
	if err != nil {
		return nil, err
	} else {
		p.News = news
	}
	attachments, err := Attachment{}.FindByResourceId(ctx, 10, p.CommonId)
	if err != nil {
		return nil, err
	} else {
		p.Attachments = attachments
	}

	p.ContractBeforePage, err = ProjectFile{}.FindByProjectIdAndFileType(ctx, p.Id, CONTRACT_BEFORE_PAGE)
	if err != nil {
		return nil, err
	}
	p.ContractAfterPage, err = ProjectFile{}.FindByProjectIdAndFileType(ctx, p.Id, CONTRACT_AFTER_PAGE)
	if err != nil {
		return nil, err
	}
	p.SilentPartnerships, err = ProjectFile{}.FindByProjectIdAndFileType(ctx, p.Id, SILENT_PARTNERSHIPS)
	if err != nil {
		return nil, err
	}
	p.ProjectReport, err = ProjectFile{}.FindByProjectIdAndFileType(ctx, p.Id, PROJECT_REPORT)
	if err != nil {
		return nil, err
	}

	_, p.RepaymentDividendSchedules, err = RepaymentDividendSchedule{}.GetByProjectId(ctx, p.Id, SearchInput{MaxResultCount: 100})
	if err != nil {
		return nil, err
	}

	p.ProjectInvestmentExecute, err = ProjectInvestmentExecute{}.GetByProjectId(ctx, p.Id)
	if err != nil {
		return nil, err
	}
	pv, err := ProjectViewCount{}.GetByProjectId(ctx, p.Id)
	if err != nil {
		return nil, err
	}
	err = pv.UpdateCount(ctx)
	if err != nil {
		return nil, err
	}

	p.ViewCount = pv.ViewCount

	p.ProjectMilestones, err = ProjectMilestone{}.GetAllByProjectId(ctx, p.Id)
	if err != nil {
		return nil, err
	}

	p.CurrentInvestorCount, err = Project{}.GetCurrentInvestorCount(ctx, p.Id)
	if err != nil {
		return nil, err
	}

	if p.CapitalRaisingEntityId != nil {
		capitalRaisingEntity, err := CapitalRaisingEntity{}.GetById(ctx, *p.CapitalRaisingEntityId)
		if err != nil {
			return nil, err
		}

		p.CapitalRaisingEntity = capitalRaisingEntity
	}

	return &p, nil
}

func (p Project) CheckExistClosedSlug(ctx context.Context) (bool, error) {
	if p.ClosedFundSlug == "" {
		return false, nil
	}
	q := factory.DB(ctx).Table("project").Where("closed_fund_slug = ? and id != ?", p.ClosedFundSlug, p.Id)
	count, err := q.Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// CheckExistFundNumber は同じファンド番号のファンドが他に存在するかを返す。
// 未入力(NULL)は重複とみなさないため、値が無ければ常に false を返す。
// 新規登録時は id が 0 のため、id != 0 の条件で全件が比較対象になる。
// 論理削除済みも対象に含める。DB のユニーク制約は deleted を見ないため、
// ここで除外すると画面では通るのに保存で落ちる、という食い違いが起きる。
func (p Project) CheckExistFundNumber(ctx context.Context) (bool, error) {
	if p.FundNumber == nil {
		return false, nil
	}
	q := factory.DB(ctx).Table("project").
		Where("fund_number = ? and id != ?", *p.FundNumber, p.Id)
	count, err := q.Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
