package sms_msg

import (
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

var (
	mockContent = "[EnergyShares]\nエナジーシェアーズをご利用頂き、誠にありがとうございます。\nお客様のパスワードが変更されましたので、ご案内致します。"

	mockSendSmsRequest = cdomain.SendSmsRequest{
		Template:   cdomain.SMS_PASSWORD_RESET_COMPLETE_TEMPLATE,
		Title:      cdomain.SMS_PASSWORD_RESET_COMPLETE_TITLE,
		Sender:     "test@email.com",
		SenderName: "sender name",
		Receiver:   "010-0000-0000",
		SmsType:    domain.PASSWORD_RESET_COMPLETE_SMS,
		Details:    cdomain.SmsPasswordResetDetail{ServiceCall: "0000-0000-0000"},
	}

	mockSmsMsg = domain.SmsMsg{
		Id:         0,
		MsgId:      "",
		Title:      "Password Changed",
		Name:       "sender name",
		Sender:     "test@email.com",
		Receiver:   []string(nil),
		Status:     "REQUEST_SMS",
		SmsType:    "PASSWORD_RESET_COMPLETE_SMS",
		SmsContent: "[EnergyShares]\nエナジーシェアーズをご利用頂き、誠にありがとうございます。\nお客様のパスワードが変更されましたので、ご案内致します。",
		Count:      1,
	}

	mockMapSmsReceiver = domain.MapSmsReceiver{
		Id:       int64(0),
		SmsId:    int64(0),
		Receiver: "010-0000-0000",
	}
)
