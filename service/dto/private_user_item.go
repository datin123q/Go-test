package dto

import (
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

type PrivateUserItem struct {
	Id            int64                `json:"id"`
	UserId        int64                `json:"user_id"`
	AuditStatus   domain.AUDIT_STATUS  `json:"audit_status"`
	InvestorType  domain.INVESTOR_TYPE `json:"investor_type"`
	Email         string               `json:"email"`
	Name          string               `json:"name"`
	LastLoginDate jsonutils.JSONTime   `json:"last_login_date"`
	MobilePhone   string               `json:"mobile_phone"`
	RequestId     string               `json:"request_id"`
	SignUpDate    jsonutils.JSONTime   `json:"sign_up_date"`
}

func (PrivateUserItem) New(user *domain.PrivateUser) *PrivateUserItem {
	return &PrivateUserItem{
		Id:            user.User.Id,
		UserId:        user.User.Id,
		AuditStatus:   user.AuditStatus,
		Name:          user.Name,
		LastLoginDate: user.LastLoginDate,
		MobilePhone:   user.MobilePhone,
		Email:         user.Email,
		InvestorType:  user.InvestorType,
		RequestId:     user.RequestId,
		SignUpDate:    user.SignUpDate,
	}
}
