package domain

import "github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"

type LotteryHistory struct {
	Id            int64              `json:"id" xorm:"pk autoincr"`
	ApplicationId string             `json:"application_id"`
	PrevStatus    LOTTERY_STATUS     `json:"prev_status"`
	NewStatus     LOTTERY_STATUS     `json:"new_status"`
	ActorId       int64              `json:"actor_id"`
	Created       jsonutils.JSONTime `json:"created" xorm:"created"`
}

func (LotteryHistory) TableName() string { return "lottery_history" }
