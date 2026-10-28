package jobs

import (
	"context"
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/s3utils"
	"github.com/sirupsen/logrus"
)

type MyNumberJobs struct{}

func (MyNumberJobs) MyNumberDeleteFileJob(ctx context.Context) error {
	return runJobWithExecutionHistory(ctx, JobExecutionHistoryInput{
		JobType: domain.JOB_EXECUTION_TYPE_MY_NUMBER_DELETE_FILE,
		JobName: jobNameMyNumberDeleteFile,
	}, myNumberDeleteFileJob)
}

func myNumberDeleteFileJob(ctx context.Context) error {
	logrus.WithFields(logrus.Fields{
		"title": "=======Start=======",
		"desc":  "=======Start Jobs=======",
		"time":  time.Now(),
	}).Info("MyNumber Jobs")

	q := factory.DB(ctx)

	var items []domain.UserFile

	q.Where("etc IN (SELECT id FROM my_number WHERE status = ? AND now() > approval_date + 90)", domain.DISAPPROVED).Find(&items)

	for _, userFile := range items {
		err := s3utils.DeleteFile(ctx, s3utils.BUCKET_PRIVATE_NAME, &userFile)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"title": "=======End MyNumber Jobs=======",
				"desc":  "Fail S3 Delete My Number File",
				"err":   err,
			}).Error("MyNumber Jobs")
			return err
		}

		err = domain.UserFile{}.Delete(ctx, userFile.Id)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"title": "=======End MyNumber Jobs=======",
				"desc":  "Fail Delete My Number File",
				"err":   err,
			}).Error("MyNumber Jobs")
			return err
		}
	}

	logrus.WithFields(logrus.Fields{
		"title": "=======End=======",
		"desc":  "=======Success Jobs=======",
		"time":  time.Now(),
	}).Info("MyNumber Jobs")

	return nil
}
