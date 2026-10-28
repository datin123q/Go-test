package user_verify

import (
	"time"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

var (
	mockUser = domain.User{
		Id:          int64(1),
		Name:        "ホン ギルドン",
		Email:       "gildong@email.com",
		MobilePhone: "010-1234-5678",
		Password:    "password",
		Salt:        "salt",
		AuditStatus: domain.COMPLETED_AUDIT,
	}

	mockSendMailResetComplete = cdomain.SendMailRequest{
		TemplateType: domain.PASSWORD_RESET_COMPLETE_TEMPLATE,
		Sender:       "info@illuminarean.com",
		Receivers:    []string{"gildong@email.com"},
		MailType:     domain.MAIL_SEND_PRIORITY_NORMAL,
		Details: map[string]interface{}{
			"Receiver":   "ホン ギルドン",
			"Time":       jsonutils.JSONTime(time.Now()).ToString(),
			"ServiceUrl": "https://dev.energyshares.jp",
		},
	}

	mockSendSmsResetComplete = cdomain.SendSmsRequest{
		Template: cdomain.SMS_PASSWORD_RESET_COMPLETE_TEMPLATE,
		Title:    cdomain.SMS_PASSWORD_RESET_COMPLETE_TITLE,
		Sender:   "ホン ギルドン",
		Receiver: "010-1234-5678",
		SmsType:  domain.PASSWORD_RESET_COMPLETE_SMS,
		Details:  cdomain.SmsPasswordResetDetail{ServiceCall: "0000-0000-0000"},
	}

	mockSendMailPasswordReset = cdomain.SendMailRequest{
		TemplateType: domain.PASSWORD_RESET_TEMPLATE,
		Sender:       "",
		Receivers:    []string{"gildong@email.com"},
		MailType:     domain.MAIL_SEND_PRIORITY_NORMAL,
		Details: map[string]interface{}{
			"ServiceUrl":       "https://dev.energyshares.jp",
			"Receiver":         "ホン ギルドン",
			"PasswordResetUrl": "url",
		},
	}
)
