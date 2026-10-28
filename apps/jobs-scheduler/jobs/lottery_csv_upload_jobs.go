package jobs

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"xorm.io/xorm"

	adomain "github.com/illuminarean/energy-share-core-api/domain"
	lotteryUsecase "github.com/illuminarean/energy-share-core-api/usecase/lottery"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/emailutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/pathutils"
	lotteryApplicationRepository "github.com/illuminarean/energy-shares-backend-repo/repository/lottery_application/mysql"
	lotteryHistoryRepository "github.com/illuminarean/energy-shares-backend-repo/repository/lottery_history/mysql"
	mailRetryRepository "github.com/illuminarean/energy-shares-backend-repo/repository/mail_retry/mysql"
	operatorHistoryRepository "github.com/illuminarean/energy-shares-backend-repo/repository/operator_history/mysql"
	projectRepository "github.com/illuminarean/energy-shares-backend-repo/repository/project/mysql"
	userRepository "github.com/illuminarean/energy-shares-backend-repo/repository/user/mysql"
	lotteryService "github.com/illuminarean/energy-shares-backend-repo/service/lottery"
	mailService "github.com/illuminarean/energy-shares-backend-repo/service/mail"
)

type LotteryCSVUploadJobs struct{}

func buildLotteryCSVUploadUsecase(ctx context.Context) (adomain.LotteryUseCase, error) {
	var cfg config.Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, errors.Wrap(err, "failed to load config for lottery csv upload job")
	}

	var engine *xorm.Engine
	switch db := factory.DB(ctx).(type) {
	case *xorm.Engine:
		engine = db
	case *xorm.Session:
		engine = db.Engine()
	default:
		return nil, fmt.Errorf("unsupported db type for lottery csv upload job: %T", db)
	}

	dbFactory := factory.NewDBFactory()
	loggerFactory := factory.NewLoggerFactory()
	emailRequester := emailutils.New(&cfg, loggerFactory)
	mailSvc := mailService.New(
		&cfg,
		emailRequester,
		loggerFactory,
		mailRetryRepository.New(dbFactory),
	)

	return lotteryUsecase.New(
		&cfg,
		lotteryApplicationRepository.New(dbFactory),
		lotteryHistoryRepository.New(dbFactory),
		projectRepository.New(dbFactory),
		userRepository.New(dbFactory),
		operatorHistoryRepository.New(dbFactory),
		mailSvc,
		lotteryService.New(i18nutils.New(pathutils.GetPathOfAssets())),
		factory.NewTxManager(engine),
	), nil
}

func (LotteryCSVUploadJobs) ProcessQueuedLotteryCSVUploadJob(ctx context.Context) error {
	return runJobWithExecutionHistory(ctx, JobExecutionHistoryInput{
		JobType: domain.JOB_EXECUTION_TYPE_LOTTERY_CSV_UPLOAD,
		JobName: jobNameLotteryCSVUpload,
	}, processQueuedLotteryCSVUploadJob)
}

func processQueuedLotteryCSVUploadJob(ctx context.Context) error {
	usecase, err := buildLotteryCSVUploadUsecase(ctx)
	if err != nil {
		return err
	}

	jobs, err := (domain.LotteryCSVUploadJob{}).GetQueuedJobs(ctx, 20)
	if err != nil {
		return errors.Wrap(err, "failed to get queued lottery csv upload jobs")
	}

	for _, job := range jobs {
		if err := usecase.ProcessCSVUploadJob(ctx, job.JobID); err != nil {
			factory.Logger(ctx).WithError(err).Errorf("failed to process lottery csv upload job: job_id=%s", job.JobID)
		}
	}

	return nil
}

func (LotteryCSVUploadJobs) ProcessLotteryCSVUploadJob(ctx context.Context, jobID string) error {
	return runJobWithExecutionHistory(ctx, JobExecutionHistoryInput{
		JobType:        domain.JOB_EXECUTION_TYPE_LOTTERY_CSV_UPLOAD,
		JobName:        jobNameLotteryCSVUpload,
		RequestPayload: jobExecutionRequestPayload(map[string]string{"job_id": jobID}),
	}, func(ctx context.Context) error {
		return processLotteryCSVUploadJob(ctx, jobID)
	})
}

func processLotteryCSVUploadJob(ctx context.Context, jobID string) error {
	usecase, err := buildLotteryCSVUploadUsecase(ctx)
	if err != nil {
		return err
	}
	return usecase.ProcessCSVUploadJob(ctx, jobID)
}
