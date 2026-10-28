package domain

import (
	"github.com/illuminarean/energy-shares-backend-repo/pkg/commoncode"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

type POLICY_TYPE string

const (
	POLICY_TYPE_CODE                                   = "POLICY_TYPE"
	POLICY_CAUTION                         POLICY_TYPE = "POLICY_CAUTION"                         // 1
	POLICY_PROTECTION                      POLICY_TYPE = "POLICY_PROTECTION"                      // 2
	POLICY_ANTISOCIAL_FORCES_BAGIC         POLICY_TYPE = "POLICY_ANTISOCIAL_FORCES_BAGIC"         // 3
	POLICY_NON_ANTISOCIAL_FORCES_AGREEMENT POLICY_TYPE = "POLICY_NON_ANTISOCIAL_FORCES_AGREEMENT" // 3-2
	POLICY_SUBSTANTIVE_RULER               POLICY_TYPE = "POLICY_SUBSTANTIVE_RULER"               // 3-3
	POLICY_SOLICITATION                    POLICY_TYPE = "POLICY_SOLICITATION"                    // 4
	POLICY_COMMISSION_RISK                 POLICY_TYPE = "POLICY_COMMISSION_RISK"                 // 5
	POLICY_CUSTOMER_CENTERED_WORK          POLICY_TYPE = "POLICY_CUSTOMER_CENTERED_WORK"          // 6
	POLICY_ELECTRONIC_TEACHING             POLICY_TYPE = "POLICY_ELECTRONIC_TEACHING"             // 7
	POLICY_NON_PEPS                        POLICY_TYPE = "POLICY_NON_PEPS"                        // 8
	POLICY_ADULT_GUARDIAN                  POLICY_TYPE = "POLICY_ADULT_GUARDIAN"                  // 9
	POLICY_SPECIFIC_INVESTORS              POLICY_TYPE = "POLICY_SPECIFIC_INVESTORS"              // 10
	POLICY_COOLING_OFF                     POLICY_TYPE = "POLICY_COOLING_OFF"                     // 11
	POLICY_CUSTOMER_HARASSMENT             POLICY_TYPE = "POLICY_CUSTOMER_HARASSMENT"             // 12
)

func (Policy) GetTypeCodeGroup(loc string) commoncode.CodeGroup {

	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(POLICY_CAUTION), Name: i18nutils.Translate("commoncode.POLICY_TYPE.POLICY_CAUTION", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(POLICY_PROTECTION), Name: i18nutils.Translate("commoncode.POLICY_TYPE.POLICY_PROTECTION", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(POLICY_ANTISOCIAL_FORCES_BAGIC), Name: i18nutils.Translate("commoncode.POLICY_TYPE.POLICY_ANTISOCIAL_FORCES_BAGIC", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(POLICY_NON_ANTISOCIAL_FORCES_AGREEMENT), Name: i18nutils.Translate("commoncode.POLICY_TYPE.POLICY_NON_ANTISOCIAL_FORCES_AGREEMENT", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(POLICY_SUBSTANTIVE_RULER), Name: i18nutils.Translate("commoncode.POLICY_TYPE.POLICY_SUBSTANTIVE_RULER", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(POLICY_SOLICITATION), Name: i18nutils.Translate("commoncode.POLICY_TYPE.POLICY_SOLICITATION", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(POLICY_COMMISSION_RISK), Name: i18nutils.Translate("commoncode.POLICY_TYPE.POLICY_COMMISSION_RISK", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(POLICY_CUSTOMER_CENTERED_WORK), Name: i18nutils.Translate("commoncode.POLICY_TYPE.POLICY_CUSTOMER_CENTERED_WORK", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(POLICY_ELECTRONIC_TEACHING), Name: i18nutils.Translate("commoncode.POLICY_TYPE.POLICY_ELECTRONIC_TEACHING", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(POLICY_NON_PEPS), Name: i18nutils.Translate("commoncode.POLICY_TYPE.POLICY_NON_PEPS", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(POLICY_ADULT_GUARDIAN), Name: i18nutils.Translate("commoncode.POLICY_TYPE.POLICY_ADULT_GUARDIAN", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(POLICY_SPECIFIC_INVESTORS), Name: i18nutils.Translate("commoncode.POLICY_TYPE.POLICY_SPECIFIC_INVESTORS", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(POLICY_COOLING_OFF), Name: i18nutils.Translate("commoncode.POLICY_TYPE.POLICY_COOLING_OFF", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(POLICY_CUSTOMER_HARASSMENT), Name: i18nutils.Translate("commoncode.POLICY_TYPE.POLICY_CUSTOMER_HARASSMENT", loc), Order: len(codes) + 1})

	return commoncode.CodeGroup{
		GroupCode: POLICY_TYPE_CODE,
		Codes:     codes,
	}
}

type POLICY_VIEW_TYPE string

const (
	POLICY_VIEW_TYPE_CODE                  = "POLICY_VIEW_TYPE"
	POLICY_VIEW_FOOTER    POLICY_VIEW_TYPE = "POLICY_VIEW_FOOTER" // フッター
	POLICY_VIEW_JOIN      POLICY_VIEW_TYPE = "POLICY_VIEW_JOIN"   // 会員登録
)

func (Policy) GetViewTypeCodeGroup(loc string) commoncode.CodeGroup {

	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(POLICY_VIEW_FOOTER), Name: i18nutils.Translate("commoncode.POLICY_VIEW_TYPE.POLICY_VIEW_FOOTER", loc), Order: 1})
	codes = append(codes, commoncode.Code{Code: string(POLICY_VIEW_JOIN), Name: i18nutils.Translate("commoncode.POLICY_VIEW_TYPE.POLICY_VIEW_JOIN", loc), Order: 2})

	return commoncode.CodeGroup{
		GroupCode: POLICY_VIEW_TYPE_CODE,
		Codes:     codes,
	}
}

type Policy struct {
	Id             int64              `json:"id"`
	CommonId       string             `json:"common_id" xorm:"notnull UUID"`
	OwnerId        int64              `json:"owner_id" xorm:"index"`
	ModifierId     int64              `json:"modifier_id" xorm:"index"`
	PolicyType     POLICY_TYPE        `json:"policy_type" xorm:"not null"`
	PolicyViewType POLICY_VIEW_TYPE   `json:"policy_view_type" xorm:"not null"`
	ContentHtml    string             `json:"content_html" xorm:"LONGTEXT not null"`
	ModifierEmail  string             `json:"modifier_email" xorm:"not null"`
	Attachment     *Attachment        `json:"attachment" xorm:"-"`
	AttachmentId   int64              `json:"attachment_id"`
	Created        jsonutils.JSONTime `json:"created" xorm:"created"`
	Updated        jsonutils.JSONTime `json:"updated" xorm:"updated"`
	Deleted        jsonutils.JSONTime `json:"deleted" xorm:"deleted"`
}

func (p Policy) TableName() string {
	return "policy"
}

func (p Policy) SelectAll() string {
	return p.TableName() + ".*"
}
