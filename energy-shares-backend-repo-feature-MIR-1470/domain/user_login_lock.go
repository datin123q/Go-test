package domain

import "github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"

const (
	LOGIN_FAIL_WARN_CNT = 4
	LOGIN_FAIL_LOCK_CNT = 7
	LOGIN_LOCK_MINUTES  = 30
)

type UserLoginLock struct {
	Id           int64              `json:"id" xorm:"pk autoincr"`
	UserId       int64              `json:"user_id" xorm:"user_id"`
	LoginFailCnt int64              `json:"login_fail_cnt" xorm:"login_fail_cnt"`
	LockAt       jsonutils.JSONTime `json:"lock_at" xorm:"lock_at"`
	Created      jsonutils.JSONTime `json:"-" xorm:"created"`
	Updated      jsonutils.JSONTime `json:"-" xorm:"updated"`
	Deleted      jsonutils.JSONTime `json:"-" xorm:"deleted"`
}

func (u UserLoginLock) TableName() string {
	return "user_login_lock"
}

func (u UserLoginLock) SelectAll() string {
	return u.TableName() + ".*"
}

func (u UserLoginLock) IsEmpty() bool {
	return u.Id == 0
}
