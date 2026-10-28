package jobs

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type EventMessageJobs struct{}

// RePublishJob は is_publish=false の行を定期的に拾って再送する(MIR-1114)。
// 従来 apps/jobs-scheduler/controllers/events.go の手動エンドポイント
// (POST /events/republish) としてしか存在せず、取りこぼしに気付く手段が
// 定期実行では無かった。
func (EventMessageJobs) RePublishJob(ctx context.Context) error {
	return runJobWithExecutionHistory(ctx, JobExecutionHistoryInput{
		JobType: domain.JOB_EXECUTION_TYPE_RE_PUBLISH,
		JobName: jobNameRePublish,
	}, domain.EventMessage{}.RePublish)
}
