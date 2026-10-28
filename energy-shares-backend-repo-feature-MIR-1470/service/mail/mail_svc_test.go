package mail

import (
	"context"
	"testing"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/emailutils"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"

	"github.com/stretchr/testify/assert"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain/mocks"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

func Test_mailService_SendEmail(t *testing.T) {
	type args struct {
		ctx              context.Context
		SendEmailRequest cdomain.SendMailRequest
	}
	tests := []struct {
		name     string
		cfg      func() *config.Config
		emailpkg func() *mocks.EmailRequester
		args     args
		wantErr  bool
	}{
		{
			name: "メールを正常に送信します。",
			args: args{
				ctx: context.Background(),
				SendEmailRequest: cdomain.SendMailRequest{
					TemplateType: domain.INVEST_SUCCESS_TEMPLATE,
					Sender:       "gildong@email.com",
					Receivers:    []string{"dool2@email.com"},
					Details: map[string]interface{}{
						"test": "test",
					},
					MailType: domain.MAIL_SEND_PRIORITY_FAST,
				},
			},

			cfg: func() *config.Config {
				cfg := &config.Config{}
				cfg.EmailService.Url = "https://localhost:5000"
				cfg.Env = "DEV"
				return cfg
			},

			emailpkg: func() *mocks.EmailRequester {
				emailpkg := &mocks.EmailRequester{}
				emailpkg.On("SendEmailRequest", context.Background(), emailutils.EmailRequest{
					TemplateName: string(domain.INVEST_SUCCESS_TEMPLATE),
					ServiceName:  "ESJP",
					ServiceEnv:   "DEV", // 実際には、これはこのように表示されるべきではありません。テストのために空の値を使用しています。
					MailData: []emailutils.MailData{
						{
							TemplateFields: map[string]interface{}{
								"test": "test",
							},
							Receivers: []string{
								"dool2@email.com",
							},
							SendType: "FAST",
						},
					},
				}).Return([]string{"msgId"}, nil)

				return emailpkg
			},

			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mailService{
				cfg:           tt.cfg(),
				emailpkg:      tt.emailpkg(),
				loggerFactory: factory.NewLoggerFactory(),
			}
			err := svc.SendMail(tt.args.ctx, tt.args.SendEmailRequest)

			assert.Equalf(t, tt.wantErr, err != nil, "SendMail() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_mailService_SendDirectEmail(t *testing.T) {
	type args struct {
		ctx             context.Context
		SendMailRequest cdomain.SendMailRequest
	}
	tests := []struct {
		name     string
		cfg      func() *config.Config
		emailpkg func() *mocks.EmailRequester
		args     args
		wantErr  bool
	}{
		{
			name: "이메일을 정상적으로 발송합니다.",
			args: args{
				ctx: context.Background(),
				SendMailRequest: cdomain.SendMailRequest{
					TemplateType: domain.INVEST_SUCCESS_TEMPLATE,
					Sender:       "gildong@email.com",
					Receivers:    []string{"dool2@email.com"},
					Details: map[string]interface{}{
						"test": "test",
					},
					MailType: domain.MAIL_SEND_PRIORITY_FAST,
				},
			},

			cfg: func() *config.Config {
				cfg := &config.Config{}
				cfg.EmailService.Url = "https://localhost:5000"
				cfg.Env = "DEV"
				return cfg
			},

			emailpkg: func() *mocks.EmailRequester {
				emailpkg := &mocks.EmailRequester{}
				emailpkg.On("SendEmailRequest", context.Background(), emailutils.EmailRequest{
					TemplateName: string(domain.INVEST_SUCCESS_TEMPLATE),
					ServiceName:  "ESJP",
					ServiceEnv:   "DEV", // 실제로는 이렇게 나오면 안된다. 테스트를 위해 빈값으로 둔다.
					MailData: []emailutils.MailData{
						{
							TemplateFields: map[string]interface{}{
								"test": "test",
							},
							Receivers: []string{
								"dool2@email.com",
							},
							SendType: "FAST",
						},
					},
				}).Return([]string{"msgId"}, nil)

				return emailpkg
			},

			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mailService{
				cfg:           tt.cfg(),
				emailpkg:      tt.emailpkg(),
				loggerFactory: factory.NewLoggerFactory(),
			}
			err := svc.SendMail(tt.args.ctx, tt.args.SendMailRequest)

			assert.Equalf(t, tt.wantErr, err != nil, "SendMail() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}
