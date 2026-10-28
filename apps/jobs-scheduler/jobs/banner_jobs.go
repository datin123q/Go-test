package jobs

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

type BannerJobs struct{}

func (BannerJobs) CloseBannersJob(ctx context.Context) error {
	return runJobWithExecutionHistory(ctx, JobExecutionHistoryInput{
		JobType: domain.JOB_EXECUTION_TYPE_CLOSE_BANNERS,
		JobName: jobNameCloseBanners,
	}, closeBannersJob)
}

func closeBannersJob(ctx context.Context) error {
	logrus.Info("CloseBannersJob started")
	now := time.Now().Format(time.DateTime)

	q := factory.DB(ctx)
	q.Table("banner").Select("banner.*").Where("end_date < ?", now).And("is_posted = true")

	var items []domain.Banner
	if err := q.Find(&items); err != nil {
		return errors.Wrap(err, "CloseBannersJob: failed to fetch banners")
	}

	for _, item := range items {
		item.IsPosted = false
		item.SortOrder = nil

		if _, err := q.ID(item.Id).Cols("is_posted", "sort_order").UseBool("is_posted").Update(&item); err != nil {
			return errors.Wrap(err, "CloseBannersJob: failed to update banner")
		}
	}

	logrus.Info("CloseBannersJob: ", len(items), " banners closed")

	return nil
}
