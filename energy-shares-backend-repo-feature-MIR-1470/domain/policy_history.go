package domain

import (
	"github.com/illuminarean/energy-shares-backend-repo/pkg/commoncode"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

type POLICY_HISTORY_TYPE string

const (
	POLICY_HISTORY_TYPE_CODE                     = "POLICY_HISTORY_TYPE"
	POLICY_HISTORY_CONTENTS  POLICY_HISTORY_TYPE = "POLICY_HISTORY_CONTENTS" // 内容変更
	POLICY_CHANGED_CONTENTS  POLICY_HISTORY_TYPE = "POLICY_CHANGED_CONTENTS" // 変更内容
)

func (PolicyInfoHistory) GetPolicyHistoryTypeCodeGroup(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(POLICY_HISTORY_CONTENTS), Name: i18nutils.Translate("commoncode.POLICY_HISTORY_TYPE.POLICY_HISTORY_CONTENTS", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(POLICY_CHANGED_CONTENTS), Name: i18nutils.Translate("commoncode.POLICY_HISTORY_TYPE.POLICY_CHANGED_CONTENTS", loc), Order: len(codes) + 1})
	return commoncode.CodeGroup{
		GroupCode: POLICY_HISTORY_TYPE_CODE,
		Codes:     codes,
	}
}

type PolicyInfoHistory struct {
	Id              int64              `json:"id"`                                // ユニークID
	PolicyId        int64              `json:"policy_id"`                         // 状態変更された注意事項ID
	SubstanceType   string             `json:"substance_type"`                    // 内容表記方式
	ContentType     string             `json:"content_type"`                      // 変更内容タイプ
	Name            string             `json:"name"`                              // 処理者
	Email           string             `json:"email"`                             // 処理者メール
	IsPosted        bool               `json:"is_posted" xorm:"not null"`         // 掲示有無
	Label           jsonutils.JSONTime `json:"label" xorm:"date"`                 // ラベル
	PrevContent     string             `json:"prev_content" xorm:"LONGTEXT"`      // 変更前内容
	CurrContent     string             `json:"curr_content" xorm:"LONGTEXT"`      // 変更後内容
	PrevContentHtml string             `json:"prev_content_html" xorm:"LONGTEXT"` // 変更前HTML
	CurrContentHtml string             `json:"curr_content_html" xorm:"LONGTEXT"` // 変更後HTML
	PrevUrl         string             `json:"prev_url"`                          // 変更前URL
	CurrUrl         string             `json:"curr_url"`                          // 変更後URL
	Created         jsonutils.JSONTime `json:"created" xorm:"created"`            // 処理日時
}

func (p PolicyInfoHistory) TableName() string {
	return "policy_info_history"
}

func (p PolicyInfoHistory) SelectAll() string {
	return p.TableName() + ".*"
}

/*
현재 policyInfoHistory 에서
substance_type 은 항상 POLICY_HISTORY_CONTENTS,
content_type 은 항상 POLICY_CHANGED_CONTENTS 을 갖는다.
-
Q. 한 가지 타입만 사용되는 이유는?
A. (준용님 답변)
History 데이터 구조상 통일되도록 작업 하신것 같은데 아마 Policy 의 상황과 맞지 않아서 한가지 타입만 들어간거 같습니다.
저 분류 자체가 꼭 필요하다기 보다는 현재 디스플레이에 필요해서 추가된 정보 같습니다.
-
향후 CDC 를 적용 하여 해당 내용을 변경/개선 할 수 있도록 한다.
*/
