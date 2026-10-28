package domain

import (
	"github.com/illuminarean/energy-shares-backend-repo/pkg/commoncode"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

const WITH_HOLDING_TAX = 0.2042

type SETTLEMENT_EXECUTION_STATUS string

const (
	SETTLEMENT_EXECUTION_STATUS_CODE                             = "SETTLEMENT_EXECUTION_STATUS"
	SETTLEMENT_COMPLETE              SETTLEMENT_EXECUTION_STATUS = "SETTLEMENT_COMPLETE"
	SETTLEMENT_INCOMPLETE            SETTLEMENT_EXECUTION_STATUS = "SETTLEMENT_INCOMPLETE"
)

type SettlementSchedule struct {
	Id              int64                       `json:"id" xorm:"pk autoincr not null"`
	ProjectId       int64                       `json:"project_id"`
	StartDate       jsonutils.JSONTime          `json:"start_date"`
	EndDate         jsonutils.JSONTime          `json:"end_date"`
	DividendDate    jsonutils.JSONTime          `json:"dividend_date"`
	IsConfirm       bool                        `json:"is_confirm"`
	ExecutionStatus SETTLEMENT_EXECUTION_STATUS `json:"execution_status"` // 배치후 완료 처리
}

func (SettlementSchedule) GetSettlementExecutionStatusCodeGroup(loc string) commoncode.CodeGroup {

	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(SETTLEMENT_COMPLETE), Name: i18nutils.Translate("commoncode.SETTLEMENT_EXECUTION_STATUS.SETTLEMENT_COMPLETE", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(SETTLEMENT_INCOMPLETE), Name: i18nutils.Translate("commoncode.SETTLEMENT_EXECUTION_STATUS.SETTLEMENT_INCOMPLETE", loc), Order: len(codes) + 1})

	return commoncode.CodeGroup{
		GroupCode: SETTLEMENT_EXECUTION_STATUS_CODE,
		Codes:     codes,
	}
}

type SETTLEMENT_TYPE string

const (
	SETTLEMENT_TYPE_CODE                          = "SETTLEMENT_TYPE"
	INVESTMENT_SETTLEMENT         SETTLEMENT_TYPE = "INVESTMENT_SETTLEMENT"
	DIVIDEND_SETTLEMENT           SETTLEMENT_TYPE = "DIVIDEND_SETTLEMENT"
	REPAYMENT_SETTLEMENT          SETTLEMENT_TYPE = "REPAYMENT_SETTLEMENT"
	DIVIDEND_REPAYMENT_SETTLEMENT SETTLEMENT_TYPE = "DIVIDEND_REPAYMENT_SETTLEMENT"
)

func (Settlement) GetSettlementTypeCodeGroup(loc string) commoncode.CodeGroup {

	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(INVESTMENT_SETTLEMENT), Name: i18nutils.Translate("commoncode.SETTLEMENT_TYPE.INVESTMENT_SETTLEMENT", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(DIVIDEND_SETTLEMENT), Name: i18nutils.Translate("commoncode.SETTLEMENT_TYPE.DIVIDEND_SETTLEMENT", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(REPAYMENT_SETTLEMENT), Name: i18nutils.Translate("commoncode.SETTLEMENT_TYPE.REPAYMENT_SETTLEMENT", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(DIVIDEND_REPAYMENT_SETTLEMENT), Name: i18nutils.Translate("commoncode.SETTLEMENT_TYPE.DIVIDEND_REPAYMENT_SETTLEMENT", loc), Order: len(codes) + 1})

	return commoncode.CodeGroup{
		GroupCode: SETTLEMENT_TYPE_CODE,
		Codes:     codes,
	}
}

// 상태값 : 대출 완료 / 배당 예정 / 배당 확정 / 배당완료 / 상환 완료
type SETTLEMENT_STATUS string

const (
	SETTLEMENT_STATUS_CODE                   = "SETTLEMENT_STATUS"
	COMPLETED              SETTLEMENT_STATUS = "COMPLETED" // 완료
	SCHEDULED              SETTLEMENT_STATUS = "SCHEDULED" // 예정
	CONFIRM                SETTLEMENT_STATUS = "CONFIRM"   // 확정
)

func (Settlement) GetSettlementStatusCodeGroup(loc string) commoncode.CodeGroup {

	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(COMPLETED), Name: i18nutils.Translate("commoncode.SETTLEMENT_STATUS.COMPLETED", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(CONFIRM), Name: i18nutils.Translate("commoncode.SETTLEMENT_STATUS.CONFIRM", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(SCHEDULED), Name: i18nutils.Translate("commoncode.SETTLEMENT_STATUS.SCHEDULED", loc), Order: len(codes) + 1})

	return commoncode.CodeGroup{
		GroupCode: SETTLEMENT_STATUS_CODE,
		Codes:     codes,
	}
}

type Settlement struct {
	Id                     int64              `json:"id" xorm:"pk autoincr not null"` //  배당 ID
	ProjectId              int64              `json:"project_id" xorm:"project_id "`  // 상품 ID
	ProjectType            PROJECT_TYPE       `json:"project_type" xorm:"project_type"`
	ScheduleId             int64              `json:"schedule_id" xorm:"schedule_id not null unique"` // 결산 스케줄 ID
	Round                  int                `json:"round" xorm:"round"`                             // 결산 회차
	SettlementType         SETTLEMENT_TYPE    `json:"settlement_type" xorm:"settlement_type"`
	Status                 SETTLEMENT_STATUS  `json:"status" xorm:"status notnull"`                             // 결산 상태
	ShareDividendAmount    float64            `json:"share_dividend_amount" xorm:"share_dividend_amount"`       // 주당 분배금
	ShareRepaymentAmount   float64            `json:"share_repayment_amount" xorm:"share_repayment_amount"`     // 주당 상환금
	ProjectShareQty        int64              `json:"project_share_qty" xorm:"project_share_qty"`               // 상품 판매 총 구수
	ProjectShareAmount     int64              `json:"project_share_amount" xorm:"project_share_amount"`         // 상품 판매 총 금액
	DividendAmount         int64              `json:"dividend_amount" xorm:"dividend_amount"`                   // 배당금액
	InterestAmount         int64              `json:"interest_amount" xorm:"interest_amount"`                   // 이자 금액
	RepaymentAmount        int64              `json:"repayment_amount" xorm:"repayment_amount"`                 // 상환 원금
	BrokerFee              int64              `json:"broker_fee" xorm:"broker_fee"`                             // 영업자 보수
	WithHoldingTax         int64              `json:"with_holding_tax" xorm:"with_holding_tax"`                 // 원천징수 세금
	EtcProfit              int64              `json:"etc_profit" xorm:"etc_profit"`                             // 잡이익
	BeforeEtcProfit        int64              `json:"before_etc_profit" xorm:"before_etc_profit"`               // 직전회차 상환 적립금
	RepaymentReserve       int64              `json:"repayment_reserve" xorm:"repayment_reserve"`               // 상환 적립금
	BeforeRepaymentReserve int64              `json:"before_repayment_reserve" xorm:"before_repayment_reserve"` // 이전 회차 상환 적립금
	Created                jsonutils.JSONTime `json:"created" xorm:"created"`
	Updated                jsonutils.JSONTime `json:"updated" xorm:"updated"`
	Deleted                jsonutils.JSONTime `json:"-" xorm:"deleted"`
}

func (Settlement) TableName() string {
	return "settlement"
}

func (s Settlement) SelectAll() string {
	return s.TableName() + ".*"
}
