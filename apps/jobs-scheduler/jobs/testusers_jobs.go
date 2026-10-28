package jobs

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
)

type TestUsersJobs struct{}

func (TestUsersJobs) DeleteTestUsersJob(ctx context.Context) error {
	return runJobWithExecutionHistory(ctx, JobExecutionHistoryInput{
		JobType: domain.JOB_EXECUTION_TYPE_DELETE_TEST_USERS,
		JobName: jobNameDeleteTestUsers,
	}, deleteTestUsersJob)
}

func deleteTestUsersJob(ctx context.Context) error {
	if !serviceutils.EnvIsQa() && !serviceutils.EnvIsStg() {
		return nil
	}

	logrus.Info("DeleteTestUsersJob started")

	// 지라: https://illuminarean.atlassian.net/browse/ESJP-1155
	rawsAffected, err := factory.DB(ctx).Table("user").
		Where("email like ? OR email like ?",
			"qa.illuminarean+7%", "qa.illuminarean+8%",
		).Delete(&domain.User{})

	if err != nil {
		logrus.Error("DeleteTestUsersJob: failed to delete test users")
		return err
	}

	logrus.Infof("DeleteTestUsersJob: %d user(s) deleted", rawsAffected)

	return nil
}
