package domain

import "github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"

// コード種別
type CodeType string

const (
	CodeTypeReferral CodeType = "REFERRAL" // 紹介コード
	// NOTE: キャンペーンコード等別の意味を持つコードを追加する場合はここに定義を増やす
)

type CodeOwnership struct {
	Id        int64              `json:"id" xorm:"pk autoincr"`
	OwnerType USER_TYPE          `json:"owner_type" xorm:"owner_type"`
	OwnerId   int64              `json:"owner_id" xorm:"owner_id"`
	CodeType  CodeType           `json:"code_type" xorm:"code_type"`
	Code      string             `json:"code" xorm:"code"`
	Notes     *string            `json:"notes,omitempty" xorm:"notes"`
	Created   jsonutils.JSONTime `json:"created" xorm:"created"`
	Updated   jsonutils.JSONTime `json:"updated" xorm:"updated"`
	Deleted   jsonutils.JSONTime `json:"deleted" xorm:"deleted"`
}

func (CodeOwnership) TableName() string {
	return "code_ownership"
}

type CodeUsage struct {
	Id              int64               `json:"id" xorm:"pk autoincr"`
	UsedOwnerType   USER_TYPE           `json:"used_owner_type" xorm:"used_owner_type"`
	UsedOwnerId     int64               `json:"used_owner_id" xorm:"used_owner_id"`
	CodeOwnershipId int64               `json:"code_ownership_id" xorm:"code_ownership_id"`
	CodeInputAt     *jsonutils.JSONTime `json:"code_input_at" xorm:"code_input_at"`
	Created         jsonutils.JSONTime  `json:"created" xorm:"created"`
	Updated         jsonutils.JSONTime  `json:"updated" xorm:"updated"`
	Deleted         jsonutils.JSONTime  `json:"deleted" xorm:"deleted"`
}

func (CodeUsage) TableName() string {
	return "code_usage"
}
