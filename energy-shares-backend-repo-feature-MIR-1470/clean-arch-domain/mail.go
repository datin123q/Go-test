package clean_arch_domain

import (
	"context"
	"maps"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/emailutils"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type MailService interface {
	SendMail(ctx context.Context, request SendMailRequest) error
}

type SendMailRequest struct {
	TemplateType domain.TEMPLATE_TYPE
	MailType     domain.MAIL_TYPE
	Sender       string
	Receivers    []string
	Details      map[string]interface{}
}

// ToEmailRequest converts SendDirectMailRequest to emailutils.EmailRequest
func (r SendMailRequest) ToEmailRequest(serviceEnv string) emailutils.EmailRequest {
	// 参照で直接代入すると問題が生じる可能性があるため、コピーを作成して代入します。
	templateFields := make(map[string]interface{})
	if r.Details != nil {
		templateFields = make(map[string]interface{}, len(r.Details))
		maps.Copy(templateFields, r.Details)
	}

	return emailutils.EmailRequest{
		ServiceName:  "ESJP", // 固定値
		ServiceEnv:   serviceEnv,
		TemplateName: r.TemplateType.TemplateName(),
		MailData: []emailutils.MailData{
			{
				TemplateFields: templateFields,
				Receivers:      r.Receivers,
				SendType:       string(r.MailType),
			},
		},
	}
}
