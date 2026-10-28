package domain

import (
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

type BANNER_SETTING_KEY string

const (
	ROLLING_SECOND BANNER_SETTING_KEY = "ROLLING_SECOND"
)

type BannerSetting struct {
	Id           int64              `json:"id" xorm:"pk autoincr"`
	SettingKey   string             `json:"setting_key" xorm:"not null" swagger:"desc(설정 키)"`
	SettingValue string             `json:"setting_value" xorm:"not null" swagger:"desc(설정 값)"`
	Created      jsonutils.JSONTime `json:"created" xorm:"created"`
	Updated      jsonutils.JSONTime `json:"updated" xorm:"updated"`
	Deleted      jsonutils.JSONTime `json:"deleted" xorm:"deleted"`
}

func (BannerSetting) TableName() string {
	return "banner_setting"
}

func (b BannerSetting) IsEmpty() bool {
	return b.Id == 0
}
