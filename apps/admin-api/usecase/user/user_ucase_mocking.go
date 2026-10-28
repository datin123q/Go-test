package user

import (
	adomain "github.com/illuminarean/energy-share-core-api/domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/emailutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

var (
	mockUser1 = domain.User{
		Id:    146,
		Email: "mockuser1@example.com",
	}

	mockMailResponse1 = adomain.MailResponse{
		Id:           101,
		Subject:      "title1",
		Sender:       "info@example.com",
		TemplateType: "TEST",
		MailType:     domain.MAIL_SEND_PRIORITY_FAST,
		SendDate:     jsonutils.JSONTime{},
		Count:        3,
		Status:       domain.MAIL_SENT,
	}

	mockMailResponse2 = adomain.MailResponse{
		Id:           102,
		Subject:      "title2",
		Sender:       "info@example.com",
		TemplateType: "TEST",
		MailType:     domain.MAIL_SEND_PRIORITY_FAST,
		SendDate:     jsonutils.JSONTime{},
		Count:        3,
		Status:       domain.MAIL_SENT,
	}

	mockEmailResponse1 = emailutils.Email{
		ID:           1,
		Subject:      "title1",
		Status:       adomain.MAIL_SENT,
		SendType:     "FAST",
		Sender:       "info@example.com",
		TemplateName: "TEST",
		SendDateTime: jsonutils.JSONTime{},
	}

	mockEmailResponse2 = emailutils.Email{
		ID:           2,
		Subject:      "title2",
		Status:       adomain.MAIL_SENT,
		SendType:     "FAST",
		Sender:       "info@example.com",
		TemplateName: "TEST",
		SendDateTime: jsonutils.JSONTime{},
	}
)
