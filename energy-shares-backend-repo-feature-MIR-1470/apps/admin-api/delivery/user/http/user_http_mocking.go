package http

import (
	adomain "github.com/illuminarean/energy-share-core-api/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

var (
	mockMail1 = adomain.MailResponse{
		Id:           1,
		Subject:      "Mail1",
		Sender:       "sender@email.com",
		TemplateType: "001",
		MailType:     "FAST",
		SendDate:     jsonutils.JSONTime{},
		Count:        1,
		Status:       "SENT",
	}
	mockMail2 = adomain.MailResponse{
		Id:           2,
		Subject:      "Mail2",
		Sender:       "sender@email.com",
		TemplateType: "002",
		MailType:     "FAST",
		SendDate:     jsonutils.JSONTime{},
		Count:        1,
		Status:       "SENT",
	}
)
