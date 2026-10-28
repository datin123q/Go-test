package mail

import (
	"context"

	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/emailutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

type mailService struct {
	cfg                 *config.Config
	emailpkg            emailutils.EmailRequester
	loggerFactory       factory.LoggerFactory
	mailRetryRepository cdomain.MailRetryRepository
}

func New(
	cfg *config.Config,
	emailpkg emailutils.EmailRequester,
	loggerFactory factory.LoggerFactory,
	mailRetryRepository cdomain.MailRetryRepository,
) cdomain.MailService {
	return &mailService{
		cfg:                 cfg,
		emailpkg:            emailpkg,
		loggerFactory:       loggerFactory,
		mailRetryRepository: mailRetryRepository,
	}
}

func (m mailService) SendMail(ctx context.Context, request cdomain.SendMailRequest) error {

	// 統合メールサービスにメール送信要求
	env := m.cfg.Env
	if env == "GITHUB" {
		env = "DEV"
	}

	emailRequest := request.ToEmailRequest(env)
	_, err := m.emailpkg.SendEmailRequest(ctx, emailRequest) // 統合メールサービスはメール要求のDBテーブル保存時にIDを返します
	if err != nil {

		m.loggerFactory.Entry(ctx).WithError(err).Errorf("err send email request")

		// メール送信失敗時にメール再試行テーブルに保存
		var mailRetry domain.MailRetry
		err := mailRetry.EmailRequestToMailRetry(emailRequest)
		if err != nil {
			m.loggerFactory.Entry(ctx).WithError(err).Error("fail convert email request to mail retry")
			return errors.Wrap(err, "fail convert email request to mail retry")
		}

		_, err = m.mailRetryRepository.Create(ctx, mailRetry)
		if err != nil {
			m.loggerFactory.Entry(ctx).WithError(err).Error("fail create mail retry")
			return errors.Wrap(err, "fail create mail retry")
		}
	}

	return nil
}
