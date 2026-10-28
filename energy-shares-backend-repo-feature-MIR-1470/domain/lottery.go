package domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/commoncode"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

type LOTTERY_STATUS string

const (
	LOTTERY_STATUS_CODE                           = "LOTTERY_STATUS"
	LOTTERY_STATUS_APPLIED         LOTTERY_STATUS = "APPLIED"
	LOTTERY_STATUS_WINNING         LOTTERY_STATUS = "WINNING"
	LOTTERY_STATUS_PARTIAL_WINNING LOTTERY_STATUS = "PARTIAL_WINNING"
	LOTTERY_STATUS_LOST            LOTTERY_STATUS = "LOST"
	LOTTERY_STATUS_CANCELED        LOTTERY_STATUS = "CANCELED"
)

type LotteryApplication struct {
	// 抽選申込ヘッダ。実際の当選 share は Winnings 側で管理する。
	Id                 string             `json:"id" xorm:"pk"`
	ProjectId          int64              `json:"project_id"`
	UserId             int64              `json:"user_id"`
	Qty                int64              `json:"qty"`
	UnitPrice          int64              `json:"unit_price"`
	Amount             int64              `json:"amount"`
	AppliedAt          jsonutils.JSONTime `json:"applied_at"`
	LostAt             jsonutils.JSONTime `json:"lost_at"`
	WinningAt          jsonutils.JSONTime `json:"winning_at"`
	CanceledAt         jsonutils.JSONTime `json:"canceled_at"`
	Created            jsonutils.JSONTime `json:"-" xorm:"created"`
	Updated            jsonutils.JSONTime `json:"-" xorm:"updated"`
	Deleted            jsonutils.JSONTime `json:"-" xorm:"deleted"`
	Status             LOTTERY_STATUS     `json:"status" xorm:"-"`
	Winnings           []Share            `json:"shares,omitempty" xorm:"-"`
	ProjectName        string             `json:"project_name,omitempty" xorm:"-"`
	ProjectStatus      PROJECT_STATUS     `json:"project_status,omitempty" xorm:"-"`
	TargetShareQty     int64              `json:"target_share_qty,omitempty" xorm:"-"`
	MinShareQty        int64              `json:"min_share_qty,omitempty" xorm:"-"`
	UserName           string             `json:"user_name,omitempty" xorm:"-"`
	UserEmail          string             `json:"user_email,omitempty" xorm:"-"`
	UserAccountBalance *int64             `json:"user_account_balance,omitempty" xorm:"-"`
	UserType           USER_TYPE          `json:"user_type,omitempty" xorm:"-"`
	winningSetsLoaded  bool               // 当選レコードがロード済みかどうか（未ロード時の誤判定防止）
}

func (l LotteryApplication) TableName() string { return "lottery_application" }
func (l LotteryApplication) SelectAll() string { return l.TableName() + ".*" }
func (l LotteryApplication) IsEmpty() bool     { return l.Id == "" }

func (l LotteryApplication) WinningAmount() int64 {
	var total int64
	for _, winning := range l.Winnings {
		total += winning.Amount
	}
	return total
}

func (l LotteryApplication) CanceledWinningAmount() int64 {
	var total int64
	for _, winning := range l.Winnings {
		if winning.InvestStatus == CANCEL_INVEST {
			total += winning.Amount
		}
	}
	return total
}

// ComputeStatus はタイムスタンプと当選情報からステータスを導出する。
// winningSetsLoaded が true の場合、当選口数×単価と申込金額を比較して一部当選を判定する。
// 落選後の繰り上げ当選（lost_at と winning_at の両方がセット）は WINNING/PARTIAL_WINNING として扱う。
func (l *LotteryApplication) ComputeStatus() {
	var winningQty int64
	for _, s := range l.Winnings {
		winningQty += s.Qty
	}

	switch {
	case !l.CanceledAt.IsZero():
		l.Status = LOTTERY_STATUS_CANCELED
	case !l.WinningAt.IsZero():
		// 当選レコードがロード済かつ当選口数が申込口数未満の場合は一部当選
		if l.winningSetsLoaded && winningQty > 0 && winningQty < l.Qty {
			l.Status = LOTTERY_STATUS_PARTIAL_WINNING
		} else {
			l.Status = LOTTERY_STATUS_WINNING
		}
	case !l.LostAt.IsZero():
		l.Status = LOTTERY_STATUS_LOST
	default:
		l.Status = LOTTERY_STATUS_APPLIED
	}
}

// SetWinnings は当選レコード（InvestStatus付き）をセットし、ステータスを再計算する。詳細取得時に使用する。
func (l *LotteryApplication) SetWinnings(winnings []Share) {
	// 一覧/詳細取得時に当選 share を後付けでセットして、集約ステータスを再計算する。
	l.Winnings = winnings

	l.winningSetsLoaded = true
	l.ComputeStatus()
}

// CanMarkAsWinning は、当選状態にマーク可能かを返します。
func (l *LotteryApplication) CanMarkAsWinning() bool {
	// 現状は「取消済みでないこと」のみを条件にしている。
	return l.CanceledAt.IsZero()
}

// CanMarkAsLost は、落選状態にマーク可能かを返します。
func (l *LotteryApplication) CanMarkAsLost() bool {
	// 部分当選済みでも残口数があれば落選確定できるため、取消済みだけを除外する。
	return l.CanceledAt.IsZero()
}

// CanMarkAsCanceled は、キャンセル状態にマーク可能かを返します。
func (l *LotteryApplication) CanMarkAsCanceled() bool {
	// ユーザ取消は未当選・未落選・未取消の申込だけを対象とする。
	return l.WinningAt.IsZero() && l.LostAt.IsZero() && l.CanceledAt.IsZero()
}

// LotteryWinning は lottery_application と share の中間テーブル。
// repository 層でのJOINにのみ使用し、ドメイン層では意識しない。
type LotteryWinning struct {
	Id                   int64              `xorm:"pk autoincr"`
	LotteryApplicationId string             `xorm:"not null"`
	ShareId              string             `xorm:"not null"`
	Created              jsonutils.JSONTime `xorm:"created"`
	Updated              jsonutils.JSONTime `xorm:"updated"`
	Deleted              jsonutils.JSONTime `xorm:"deleted"`
}

func (l LotteryWinning) TableName() string { return "lottery_winning" }

func (l *LotteryWinning) Create(ctx context.Context) (int64, error) {
	// 手動当選や CSV 当選で生成した share と抽選申込を紐付ける。
	return factory.DB(ctx).Insert(l)
}

func (LotteryApplication) GetStatusCodeGroup(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0, 5)
	codes = append(codes, commoncode.Code{Code: string(LOTTERY_STATUS_APPLIED), Name: i18nutils.Translate("commoncode.LOTTERY_STATUS.APPLIED", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(LOTTERY_STATUS_WINNING), Name: i18nutils.Translate("commoncode.LOTTERY_STATUS.WINNING", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(LOTTERY_STATUS_PARTIAL_WINNING), Name: i18nutils.Translate("commoncode.LOTTERY_STATUS.PARTIAL_WINNING", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(LOTTERY_STATUS_LOST), Name: i18nutils.Translate("commoncode.LOTTERY_STATUS.LOST", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(LOTTERY_STATUS_CANCELED), Name: i18nutils.Translate("commoncode.LOTTERY_STATUS.CANCELED", loc), Order: len(codes) + 1})

	return commoncode.CodeGroup{
		GroupCode: LOTTERY_STATUS_CODE,
		Codes:     codes,
	}
}
