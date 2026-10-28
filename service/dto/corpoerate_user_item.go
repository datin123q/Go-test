package dto

import (
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

type CorporateUserItem struct {
	Id            int64                `json:"id"`
	UserId        int64                `json:"user_id"`
	AuditStatus   domain.AUDIT_STATUS  `json:"audit_status"`
	InvestorType  domain.INVESTOR_TYPE `json:"investor_type"`
	Email         string               `json:"email"`
	Name          string               `json:"name"`
	LastLoginDate jsonutils.JSONTime   `json:"last_login_date"`
	MobilePhone   string               `json:"mobile_phone"`
	SignUpDate    jsonutils.JSONTime   `json:"sign_up_date"`
}

func (CorporateUserItem) New(user domain.CorporateUser) *CorporateUserItem {

	return &CorporateUserItem{
		Id:            user.User.Id,
		UserId:        user.User.Id,
		AuditStatus:   user.AuditStatus,
		Name:          user.Name,
		LastLoginDate: user.LastLoginDate,
		MobilePhone:   user.MobilePhone,
		Email:         user.Email,
		InvestorType:  user.InvestorType,
		SignUpDate:    user.SignUpDate,
	}
}

func (CorporateUserItem) NewList(users []domain.CorporateUser) []*CorporateUserItem {
	var items []*CorporateUserItem
	for i := range users {
		items = append(items, CorporateUserItem{}.New(users[i]))
	}
	return items

}
