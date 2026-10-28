package jobs

import (
	"context"

	"github.com/pkg/errors"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/emailutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

type MailRetryJobs struct{}

// 통합 메일 서비스 이용
func (MailRetryJobs) MailRequestRetryJob(ctx context.Context) error {
	return runJobWithExecutionHistory(ctx, JobExecutionHistoryInput{
		JobType: domain.JOB_EXECUTION_TYPE_MAIL_REQUEST_RETRY,
		JobName: jobNameMailRequestRetry,
	}, mailRequestRetryJob)
}

func mailRequestRetryJob(ctx context.Context) error {
	// PENDING 인 메일을 모두 가져오기
	mailRetries, err := domain.MailRetry{}.GetAllByStatus(ctx, "PENDING")
	if err != nil {
		return errors.Wrap(err, "fail to get mail_retry by status PENDING")
	}

	// DB 업데이트: 읽어온 PENDING 메일을 IN_PROGRESS 로 변경
	err = domain.MailRetry{}.UpdateStatusMany(ctx, mailRetries, "IN_PROGRESS")
	if err != nil {
		return errors.Wrap(err, "fail to update mail_retry status to IN_PROGRESS")
	}

	// 재발신 시도
	for _, mailRetry := range mailRetries {

		emailRequest, err := mailRetry.ToEmailRequest()
		if err != nil {
			factory.Logger(ctx).Error(errors.Wrap(err, "fail to convert mail_retry to email_request"))
			continue
		}

		// 메일 발신 요청 재시도 2회
		mailIds, err := emailutils.SendEmailRequestRetry(ctx, emailRequest)
		if err != nil {
			// 최종 실패 메일 저장 - 같은 구조체, 다른 테이블
			mailRetry.RetryStatus = "FAILED"
			mailFailed := mailRetry.ToMailFailed()
			_, err := mailFailed.Create(ctx)
			if err != nil {
				factory.Logger(ctx).Error(errors.Wrap(err, "fail to create mail_failed"))
			}
			continue
		}
		factory.Logger(ctx).Infof("success to send mail of ids(in mail service): %v", mailIds)
	}

	// 모든 메일 재시도후 IN_PROGRESS 상태인 메일을 모두 삭제
	err = domain.MailRetry{}.DeleteMany(ctx, mailRetries)
	if err != nil {
		factory.Logger(ctx).Error(errors.Wrap(err, "fail to delete mail_retry of retried"))
		return errors.Wrap(err, "fail to delete mail_retry")
	}

	return nil
}
