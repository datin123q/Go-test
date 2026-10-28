package domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/commoncode"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

type EKYC_REQUEST_STATUS string

const (
	EKYC_REQUEST_STATUS_CODE                        = "EKYC_REQUEST_STATUS"
	EKYC_REQUEST_STATUS_NONE    EKYC_REQUEST_STATUS = "NONE"
	EKYC_REQUEST_STATUS_SUCCESS EKYC_REQUEST_STATUS = "SUCCESS"
	EKYC_REQUEST_STATUS_FAIL    EKYC_REQUEST_STATUS = "FAIL"
)

type EKYC_RESULT_STATUS string

const (
	EKYC_RESULT_STATUS_CODE                       = "EKYC_RESULT_STATUS"
	EKYC_RESULT_STATUS_MATCH   EKYC_RESULT_STATUS = "MATCH"
	EKYC_RESULT_STATUS_UNMATCH EKYC_RESULT_STATUS = "UNMATCH"
)

type EKYC_OVERALL_RESULT_STATUS string

const (
	EKYC_OVERALL_RESULT_STATUS_CODE                                = "EKYC_OVERALL_RESULT_STATUS"
	EKYC_OVERALL_RESULT_STATUS_CLEARED  EKYC_OVERALL_RESULT_STATUS = "CLEARED"  // 本人確認OK+反社チェックOK
	EKYC_OVERALL_RESULT_STATUS_DETECTED EKYC_OVERALL_RESULT_STATUS = "DETECTED" // 本人確認OK+反社チェックNG
	EKYC_OVERALL_RESULT_STATUS_UNMATCH  EKYC_OVERALL_RESULT_STATUS = "UNMATCH"  // 本人確認NG
)

func (EkycRequestLog) GetResultStatusCodeGroup(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0)
	codes = append(codes, commoncode.Code{Code: string(EKYC_RESULT_STATUS_MATCH), Name: i18nutils.Translate("commoncode.EKYC_RESULT_STATUS.MATCH", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(EKYC_RESULT_STATUS_UNMATCH), Name: i18nutils.Translate("commoncode.EKYC_RESULT_STATUS.UNMATCH", loc), Order: len(codes) + 1})

	return commoncode.CodeGroup{
		GroupCode: EKYC_RESULT_STATUS_CODE,
		Codes:     codes,
	}
}

type EkycRequestLog struct {
	Id             int64               `json:"id" xorm:"pk autoincr"`
	Email          string              `json:"email" xorm:"notnull" swagger:"desc(email)"`
	UserId         int64               `json:"user_id" xorm:"notnull" swagger:"desc(user ID)"`
	EkycId         string              `json:"ekyc_id" xorm:"notnull" swagger:"desc(eKYC 요청 시 생성한 ID)"`
	RequestId      *string             `json:"request_id,omitempty" swagger:"desc(JPKI方式の検証依頼ID)"` // MIR-1469: JPKI方式でNexwayが発行する検証依頼ID。ホ方式の行はnil
	ApplicantId    int64               `json:"applicant_id" xorm:"notnull" swagger:"desc(eKYC 요청 시 nexway 에서 발급 받은 ID)"`
	Status         EKYC_REQUEST_STATUS `json:"status" xorm:"notnull" swagger:"desc(eKYC 요청 성공 여부: NONE|SUCCESS|FAIL)"` // NONE: 요청 결과 없음
	Created        jsonutils.JSONTime  `json:"created" xorm:"created"`
	Result         EKYC_RESULT_STATUS  `json:"result" swagger:"desc(eKYC 요청 결과: MATCH|UNMATCH)"`     // webhook 으로 온 eKYC 최종 결과
	ErrorMessage   *string             `json:"error_message" swagger:"desc(eKYC 최종 결과 실패 시 에러 메시지)"` // Result 가 UNMATCH 인 경우에만 값이 존재
	ResultResponse *EkycResultOutput   `json:"result_response,omitempty" xorm:"json" swagger:"desc(eKYCレスポンスJSON)"`

	// eKYC照合結果
	EkycFrontMatch     *string             `json:"ekyc_front_match,omitempty" swagger:"desc(eKYC表面照合結果)"`
	EkycThicknessMatch *string             `json:"ekyc_thickness_match,omitempty" swagger:"desc(eKYC厚み照合結果)"`
	EkycLiveMatch      *string             `json:"ekyc_live_match,omitempty" swagger:"desc(eKYC生体検知結果)"`
	EkycEndAt          *jsonutils.JSONTime `json:"ekyc_end_at,omitempty" swagger:"desc(eKYC終了日時)"`

	// リスク情報検索結果
	RiskDetectDetectedAt *jsonutils.JSONTime `json:"risk_detect_detected_at,omitempty" swagger:"desc(リスク情報検索日時)"`
	RiskDetectResult     *string             `json:"risk_detect_result,omitempty" swagger:"desc(リスク情報検索結果: cleared, detected, not_checked)"`
}

func (EkycRequestLog) TableName() string {
	return "ekyc_request_log"
}

func (e EkycRequestLog) SelectAll() string {
	return e.TableName() + ".*"
}

func (c *EkycRequestLog) Create(ctx context.Context) (int64, error) {
	return factory.DB(ctx).Insert(c)
}

// Updateはxormのデフォルトどおりゼロ値のカラムをUPDATE文から除外する。
// MIR-1223: 検証未了のためApplicantId/Result等をクリア(ゼロ値/nil)したい
// 呼び出し元は、colsで対象カラムを明示すること。指定が無ければUpdate(c)は
// ゼロ値のカラムをすべて無視するため、「クリアしたつもり」で実際にはDBの
// 既存値が変わらないまま残る(xormの罠、実測で確認済み)。
func (c *EkycRequestLog) Update(ctx context.Context, cols ...string) (int64, error) {
	if len(cols) > 0 {
		return factory.DB(ctx).ID(c.Id).Cols(cols...).Update(c)
	}
	return factory.DB(ctx).ID(c.Id).Update(c)
}

// UpdateWithMustColsはUpdateと異なり、他のカラムをColsで絞り込まない
// (ゼロ値でなければ従来どおり自動検出でUPDATE対象に含む)。mustColsに
// 指定したカラムだけ、ゼロ値でも強制的にUPDATE文に含める。MIR-1223:
// 「ある1カラムだけゼロ値でも確実にクリアしたいが、他のカラムは通常どおり
// 自動検出させたい」場合に使う。Cols(全体を絞り込む)とMustCols(一部だけ
// 強制する)はxorm内部で別の仕組みであり、Colsで絞り込むと指定していない
// カラムはゼロ値でなくてもUPDATE対象から外れてしまうため、Updateの
// cols引数(=Cols)を流用してはいけない(実測で確認済み)。
func (c *EkycRequestLog) UpdateWithMustCols(ctx context.Context, mustCols ...string) (int64, error) {
	return factory.DB(ctx).ID(c.Id).MustCols(mustCols...).Update(c)
}

func (c *EkycRequestLog) GetLatest(ctx context.Context, userId int64) (*EkycRequestLog, error) {
	var item EkycRequestLog

	has, err := factory.DB(ctx).Where("user_id = ?", userId).Desc("created").Get(&item)
	if err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}

	return &item, nil
}

func (c *EkycRequestLog) GetLatestByEmail(ctx context.Context, email string) (*EkycRequestLog, error) {
	var item EkycRequestLog

	has, err := factory.DB(ctx).Where("email = ?", email).Desc("created").Get(&item)
	if err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}

	return &item, nil
}

func (c *EkycRequestLog) GetByEkycId(ctx context.Context, ekycId string) (*EkycRequestLog, error) {
	var item EkycRequestLog

	_, err := factory.DB(ctx).Where("ekyc_id = ?", ekycId).Get(&item)
	if err != nil {
		return nil, err
	}

	return &item, nil
}
