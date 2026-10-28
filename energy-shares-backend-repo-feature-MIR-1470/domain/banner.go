package domain

import (
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

type Banner struct {
	Id          int64              `json:"id" xorm:"pk autoincr"`
	Title       string             `json:"title" xorm:"not null" swagger:"desc(제목)"`
	IsPosted    bool               `json:"is_posted" xorm:"not null default true" swagger:"desc(게시 여부)"`
	StartDate   jsonutils.JSONTime `json:"start_date" swagger:"desc(게시 시작일, yyyy-MM-dd HH:mm:ss)"`
	EndDate     jsonutils.JSONTime `json:"end_date" swagger:"desc(게시 종료일, yyyy-MM-dd 23:59:59)"`
	Url         string             `json:"url" xorm:"not null" swagger:"desc(url)"`
	SortOrder   *int64             `json:"sort_order" swagger:"desc(순서)"`
	Description string             `json:"description" swagger:"desc(설명)"`
	NewTab      bool               `json:"new_tab" xorm:"not null default false" swagger:"desc(새 탭 여부)"`
	Created     jsonutils.JSONTime `json:"created" xorm:"created"`
	Updated     jsonutils.JSONTime `json:"updated" xorm:"updated"`
	Deleted     jsonutils.JSONTime `json:"deleted" xorm:"deleted"`
}

func (Banner) TableName() string {
	return "banner"
}

func (b Banner) SelectAll() string {
	return b.TableName() + ".*"
}

func (b Banner) IsEmpty() bool {
	return b.Id == 0
}
