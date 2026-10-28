package jobs

import (
	"context"
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type ProductJobs struct {
}

// 펀딩 오픈 잡
func (ProductJobs) FundingOpenJob(ctx context.Context) error {
	logrus.WithFields(logrus.Fields{
		"category": "funding open",
	}).Info("start FundingOpenJob")

	var items []domain.Project
	// Start everyday 1 o'clock
	// MIR-1314: 一覧取得(start_at <= nowStr)と、各件のChangeStatus条件付き
	// 更新に同じ時刻文字列を使う。一覧取得後に管理者がstart_atを変えていた
	// 場合や、同じジョブが二重に実行された場合を、この条件付き更新の
	// WHERE句で弾くため。
	nowStr := time.Now().In(time.Local).Format(time.DateTime)
	q := factory.DB(ctx)
	totalCnt, err := q.Table("project").Select(domain.Project{}.SelectAll()).Where("project_status = ? AND "+
		"start_at <= ?", domain.STANDBY_FUNDING, nowStr).
		FindAndCount(&items)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"category": "project close funding",
			"err":      err,
			"time":     time.Now(),
		}).Error("Failed to get the fund table that needs to be closed FundingCloseJob")
	}

	for _, item := range items {
		logrus.WithFields(logrus.Fields{
			"category":     "project open funding",
			"project":      item.Name,
			"id":           item.Id,
			"project_type": item.ProjectType,
			"status":       item.ProjectStatus,
			"end_at":       item.EndAt.ToString(),
			"time":         time.Now(),
		}).Info("fund open job start FundingOpenJob")

		err := item.ChangeStatus(ctx, domain.OPEN_FUNDING, item.UpdateId,
			domain.JobStatusCondition{Where: "start_at <= ?", Arg: nowStr})
		if err != nil {
			logJobStatusChangeError(ctx, item.Id, err, "FundingOpenJob")
		}
		logrus.WithFields(logrus.Fields{
			"category":     "project open funding",
			"project":      item.Name,
			"id":           item.Id,
			"project_type": item.ProjectType,
			"status":       item.ProjectStatus,
			"end_at":       item.EndAt.ToString(),
			"time":         time.Now(),
		}).Info("success change item status FundingOpenJob")
	}
	// console log
	logrus.WithFields(logrus.Fields{
		"category": "project end EQUITY",
		"count":    totalCnt,
		"time":     time.Now(),
	}).Info("success FundingCloseJob")
	return nil
}

// 펀드 마감 잡
func (ProductJobs) FundingCloseJob(ctx context.Context) error {
	logrus.WithFields(logrus.Fields{
		"category": "project close",
	}).Info("start FundingCloseJob")

	var items []domain.Project

	q := factory.DB(ctx)
	now := jsonutils.JSONTime(time.Now())
	// MIR-1314: 一覧取得(end_at < todayMidnight)と、各件のChangeStatus
	// 条件付き更新に同じ時刻文字列を使う。理由はFundingOpenJobと同じ。
	todayMidnight := now.Time().Format("2006-01-02 00:00:00")

	logrus.WithFields(logrus.Fields{
		"time": now.Time().In(time.Local).Format("2006-01-02 00:00:00"),
	}).Info("start time FundingCloseJob")

	totalCnt, err := q.Select(domain.Project{}.SelectAll()).
		Where("project_status = ?", domain.OPEN_FUNDING).
		And("end_at < ?", todayMidnight).
		FindAndCount(&items)

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"category": "project close funding",
			"err":      err,
			"time":     time.Now(),
		}).Error("Failed to get the fund table that needs to be closed FundingCloseJob")
	}

	for _, item := range items {
		logrus.WithFields(logrus.Fields{
			"category":     "project close EQUITY",
			"project":      item.Name,
			"id":           item.Id,
			"project_type": item.ProjectType,
			"status":       item.ProjectStatus,
			"end_at":       item.EndAt.ToString(),
			"time":         time.Now(),
		}).Info("fund end job start FundingCloseJob")

		err := item.ChangeStatus(ctx, domain.END_FUNDING, item.UpdateId,
			domain.JobStatusCondition{Where: "end_at < ?", Arg: todayMidnight})
		if err != nil {
			logJobStatusChangeError(ctx, item.Id, err, "FundingCloseJob")
		}
		logrus.WithFields(logrus.Fields{
			"category":     "project ending EQUITY",
			"project":      item.Name,
			"id":           item.Id,
			"project_type": item.ProjectType,
			"status":       item.ProjectStatus,
			"end_at":       item.EndAt.ToString(),
			"time":         time.Now(),
		}).Info("success change item status FundingCloseJob")
	}
	// console log
	logrus.WithFields(logrus.Fields{
		"category": "project end EQUITY",
		"count":    totalCnt,
		"time":     time.Now(),
	}).Info("success FundingCloseJob")
	return nil
}

// isProjectStatusChangedは、errがErrorProjectStatusChanged(一覧取得後に状態や
// start_at/end_atが変わっていた、または同じジョブの二重実行)かを判定する
// (MIR-1314)。errors.Asはpkg/errors.Wrapで包まれていても中の*factory.Errorを
// たどれる。
func isProjectStatusChanged(err error) bool {
	var apiErr *factory.Error
	return errors.As(err, &apiErr) && apiErr.Code == factory.ErrorProjectStatusChanged.Code
}

// logJobStatusChangeErrorは、ChangeStatusのエラーをログに残す(MIR-1314)。
// isProjectStatusChangedがtrueの場合は、その1件を飛ばして次へ進む前提の
// 想定内の結果なのでWarnで記録する。それ以外は従来どおりErrorで記録する。
func logJobStatusChangeError(ctx context.Context, projectId int64, err error, jobName string) {
	fields := logrus.Fields{
		"projectId": projectId,
		"err":       err,
		"time":      time.Now(),
	}
	if isProjectStatusChanged(err) {
		factory.Logger(ctx).WithFields(fields).Warn("skip: project status changed since listing " + jobName)
		return
	}
	factory.Logger(ctx).WithFields(fields).Error("fail project status update " + jobName)
}
