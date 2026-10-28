package clean_arch_domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

const (
	SMS_SIGN_UP_COMPLETE_TITLE           = "Certification"
	SMS_SIGN_UP_COMPLETE_TEMPLATE        = "/template/sms/ja/sms-templates-smscert.txt"
	SMS_PASSWORD_RESET_COMPLETE_TITLE    = "Password Changed"
	SMS_PASSWORD_RESET_COMPLETE_TEMPLATE = "/template/sms/ja/sms-templates-password-changed.txt"
)

type SmsMsgService interface {
	SendSms(ctx context.Context, request SendSmsRequest) error
}

type SmsMsgRepository interface {
	Create(ctx context.Context, smsMsg domain.SmsMsg) (int64, error)
	Update(ctx context.Context, smsMsg domain.SmsMsg) error
}

type SendSmsRequest struct {
	Template   string
	Title      string
	Sender     string
	SenderName string
	Receiver   string
	SmsType    domain.SMS_TYPE
	Details    interface{}
}

type SmsSignUpDetail struct {
	Code string
}

type SmsPasswordResetDetail struct {
	ServiceCall string
}
