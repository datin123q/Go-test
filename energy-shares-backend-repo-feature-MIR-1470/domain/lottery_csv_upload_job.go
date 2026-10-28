package domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

type LOTTERY_CSV_UPLOAD_JOB_STATUS string

const (
	LOTTERY_CSV_UPLOAD_JOB_STATUS_QUEUED         LOTTERY_CSV_UPLOAD_JOB_STATUS = "QUEUED"
	LOTTERY_CSV_UPLOAD_JOB_STATUS_RUNNING        LOTTERY_CSV_UPLOAD_JOB_STATUS = "RUNNING"
	LOTTERY_CSV_UPLOAD_JOB_STATUS_COMPLETED      LOTTERY_CSV_UPLOAD_JOB_STATUS = "COMPLETED"
	LOTTERY_CSV_UPLOAD_JOB_STATUS_PARTIAL_FAILED LOTTERY_CSV_UPLOAD_JOB_STATUS = "PARTIAL_FAILED"
	LOTTERY_CSV_UPLOAD_JOB_STATUS_FAILED         LOTTERY_CSV_UPLOAD_JOB_STATUS = "FAILED"
	LOTTERY_CSV_UPLOAD_JOB_STATUS_CANCELED       LOTTERY_CSV_UPLOAD_JOB_STATUS = "CANCELED"
)

type LOTTERY_CSV_UPLOAD_PROCESS_TYPE string

const (
	LOTTERY_CSV_UPLOAD_PROCESS_TYPE_WIN  LOTTERY_CSV_UPLOAD_PROCESS_TYPE = "WIN"
	LOTTERY_CSV_UPLOAD_PROCESS_TYPE_LOSE LOTTERY_CSV_UPLOAD_PROCESS_TYPE = "LOSE"
	LOTTERY_CSV_UPLOAD_PROCESS_TYPE_SKIP LOTTERY_CSV_UPLOAD_PROCESS_TYPE = "SKIP"
)

type LOTTERY_CSV_UPLOAD_DETAIL_STATUS string

const (
	LOTTERY_CSV_UPLOAD_DETAIL_STATUS_PENDING LOTTERY_CSV_UPLOAD_DETAIL_STATUS = "PENDING"
	LOTTERY_CSV_UPLOAD_DETAIL_STATUS_SUCCESS LOTTERY_CSV_UPLOAD_DETAIL_STATUS = "SUCCESS"
	LOTTERY_CSV_UPLOAD_DETAIL_STATUS_FAILED  LOTTERY_CSV_UPLOAD_DETAIL_STATUS = "FAILED"
	LOTTERY_CSV_UPLOAD_DETAIL_STATUS_SKIPPED LOTTERY_CSV_UPLOAD_DETAIL_STATUS = "SKIPPED"
)

type LotteryCSVUploadJob struct {
	Id              int64                         `json:"id" xorm:"pk autoincr"`
	JobID           string                        `json:"job_id" xorm:"job_id"`
	ProjectID       int64                         `json:"project_id" xorm:"project_id"`
	RequestedBy     string                        `json:"requested_by" xorm:"requested_by"`
	FileName        string                        `json:"file_name" xorm:"file_name"`
	Status          LOTTERY_CSV_UPLOAD_JOB_STATUS `json:"status" xorm:"status"`
	TotalRowCount   int                           `json:"total_row_count" xorm:"total_row_count"`
	SuccessRowCount int                           `json:"success_row_count" xorm:"success_row_count"`
	FailedRowCount  int                           `json:"failed_row_count" xorm:"failed_row_count"`
	StartedAt       jsonutils.JSONTime            `json:"started_at" xorm:"started_at"`
	FinishedAt      jsonutils.JSONTime            `json:"finished_at" xorm:"finished_at"`
	ErrorMessage    *string                       `json:"error_message" xorm:"error_message"`
	Created         jsonutils.JSONTime            `json:"created" xorm:"created"`
	Updated         jsonutils.JSONTime            `json:"updated" xorm:"updated"`
	Deleted         jsonutils.JSONTime            `json:"deleted" xorm:"deleted"`
}

type LotteryCSVUploadJobDetail struct {
	Id                      int64                            `json:"id" xorm:"pk autoincr"`
	JobID                   int64                            `json:"job_id" xorm:"job_id"`
	RowNumber               int                              `json:"row_number" xorm:"row_number"`
	LotteryApplicationID    string                           `json:"lottery_application_id" xorm:"lottery_application_id"`
	ExpectedLotteryStatus   *LOTTERY_STATUS                  `json:"expected_lottery_status" xorm:"expected_lottery_status"`
	ExpectedAppliedAt       *string                          `json:"expected_applied_at" xorm:"expected_applied_at"`
	InputAppliedQty         *int64                           `json:"input_applied_qty" xorm:"input_applied_qty"`
	InputWinningQty         int64                            `json:"input_winning_qty" xorm:"input_winning_qty"`
	ProcessType             LOTTERY_CSV_UPLOAD_PROCESS_TYPE  `json:"process_type" xorm:"process_type"`
	Status                  LOTTERY_CSV_UPLOAD_DETAIL_STATUS `json:"status" xorm:"status"`
	ErrorCode               *string                          `json:"error_code" xorm:"error_code"`
	ErrorMessage            *string                          `json:"error_message" xorm:"error_message"`
	ProcessedAt             jsonutils.JSONTime               `json:"processed_at" xorm:"processed_at"`
	CreatedShareID          *string                          `json:"created_share_id" xorm:"created_share_id"`
	CreatedDepositRequestID *int64                           `json:"created_deposit_request_id" xorm:"created_deposit_request_id"`
	Created                 jsonutils.JSONTime               `json:"created" xorm:"created"`
	Updated                 jsonutils.JSONTime               `json:"updated" xorm:"updated"`
	Deleted                 jsonutils.JSONTime               `json:"deleted" xorm:"deleted"`
}

func (LotteryCSVUploadJob) TableName() string {
	return "lottery_csv_upload_job"
}

func (LotteryCSVUploadJobDetail) TableName() string {
	return "lottery_csv_upload_job_detail"
}

func (j *LotteryCSVUploadJob) Create(ctx context.Context) (int64, error) {
	if _, err := factory.DB(ctx).Insert(j); err != nil {
		return 0, err
	}
	return j.Id, nil
}

func (d *LotteryCSVUploadJobDetail) Create(ctx context.Context) (int64, error) {
	if _, err := factory.DB(ctx).Insert(d); err != nil {
		return 0, err
	}
	return d.Id, nil
}

func (LotteryCSVUploadJob) GetByJobID(ctx context.Context, jobID string) (*LotteryCSVUploadJob, error) {
	var job LotteryCSVUploadJob
	has, err := factory.DB(ctx).Where("job_id = ?", jobID).Get(&job)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return &job, nil
}

func (LotteryCSVUploadJobDetail) GetByJobID(ctx context.Context, jobID int64) ([]LotteryCSVUploadJobDetail, error) {
	var details []LotteryCSVUploadJobDetail
	err := factory.DB(ctx).
		Where("job_id = ?", jobID).
		Asc("row_number").
		Find(&details)
	return details, err
}

func (LotteryCSVUploadJob) GetQueuedJobs(ctx context.Context, limit int) ([]LotteryCSVUploadJob, error) {
	var jobs []LotteryCSVUploadJob
	q := factory.DB(ctx).
		Where("status = ?", LOTTERY_CSV_UPLOAD_JOB_STATUS_QUEUED).
		Asc("created")
	if limit > 0 {
		q = q.Limit(limit)
	}
	err := q.Find(&jobs)
	return jobs, err
}

func (LotteryCSVUploadJob) HasActiveJobByProjectID(ctx context.Context, projectID int64) (bool, error) {
	count, err := factory.DB(ctx).
		Where("project_id = ?", projectID).
		In("status", LOTTERY_CSV_UPLOAD_JOB_STATUS_QUEUED, LOTTERY_CSV_UPLOAD_JOB_STATUS_RUNNING).
		Count(LotteryCSVUploadJob{})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (j *LotteryCSVUploadJob) TryMarkRunningFromQueued(ctx context.Context, startedAt jsonutils.JSONTime) (bool, error) {
	update := &LotteryCSVUploadJob{
		Status:       LOTTERY_CSV_UPLOAD_JOB_STATUS_RUNNING,
		StartedAt:    startedAt,
		ErrorMessage: nil,
	}

	affected, err := factory.DB(ctx).
		Table(j.TableName()).
		Where("id = ?", j.Id).
		And("status = ?", LOTTERY_CSV_UPLOAD_JOB_STATUS_QUEUED).
		Cols("status", "started_at", "error_message").
		Update(update)
	if err != nil {
		return false, err
	}
	if affected == 0 {
		return false, nil
	}

	j.Status = LOTTERY_CSV_UPLOAD_JOB_STATUS_RUNNING
	j.StartedAt = startedAt
	j.ErrorMessage = nil
	return true, nil
}

func (j *LotteryCSVUploadJob) Update(ctx context.Context) error {
	_, err := factory.DB(ctx).ID(j.Id).AllCols().Update(j)
	return err
}

func (d *LotteryCSVUploadJobDetail) Update(ctx context.Context) error {
	_, err := factory.DB(ctx).ID(d.Id).AllCols().Update(d)
	return err
}
