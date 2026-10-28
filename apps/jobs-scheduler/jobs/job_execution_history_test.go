package jobs

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
)

type fakeJobExecutionHistoryRepository struct {
	created []domain.JobExecutionHistory
	updated []domain.JobExecutionHistory
	nextID  int64
}

func (r *fakeJobExecutionHistoryRepository) Create(_ context.Context, history domain.JobExecutionHistory) (int64, error) {
	if r.nextID == 0 {
		r.nextID = 1
	}
	history.Id = r.nextID
	r.created = append(r.created, history)
	r.nextID++
	return history.Id, nil
}

func (r *fakeJobExecutionHistoryRepository) GetById(_ context.Context, id int64) (domain.JobExecutionHistory, error) {
	for _, history := range append(r.created, r.updated...) {
		if history.Id == id {
			return history, nil
		}
	}
	return domain.JobExecutionHistory{}, errors.New("not found")
}

func (r *fakeJobExecutionHistoryRepository) Update(_ context.Context, history domain.JobExecutionHistory) error {
	r.updated = append(r.updated, history)
	return nil
}

func TestJobExecutionHistoryRunner_RunSuccess(t *testing.T) {
	repo := &fakeJobExecutionHistoryRepository{}
	now := time.Date(2026, 7, 13, 9, 0, 0, 0, time.UTC)
	requestPayload := `{"year":"2026","quarter":"QUARTER_FIRST"}`

	runner := JobExecutionHistoryRunner{
		Repository: repo,
		Now: func() time.Time {
			return now
		},
	}

	err := runner.Run(context.Background(), JobExecutionHistoryInput{
		JobType:        domain.JOB_EXECUTION_TYPE_CREATE_INVESTOR_TRANSACTION_BALANCE_REPORT,
		JobName:        jobNameCreateInvestorTransactionBalanceReport,
		TriggerType:    domain.JOB_EXECUTION_TRIGGER_TYPE_SCHEDULED,
		RequestPayload: &requestPayload,
	}, func(ctx context.Context) error {
		return nil
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if len(repo.created) != 1 {
		t.Fatalf("created count = %d, want 1", len(repo.created))
	}
	if repo.created[0].Status != domain.JOB_EXECUTION_HISTORY_STATUS_RUNNING {
		t.Fatalf("created status = %s, want RUNNING", repo.created[0].Status)
	}
	if repo.created[0].TriggerType != domain.JOB_EXECUTION_TRIGGER_TYPE_SCHEDULED {
		t.Fatalf("created trigger type = %s, want SCHEDULED", repo.created[0].TriggerType)
	}
	if repo.created[0].RequestPayload == nil || *repo.created[0].RequestPayload != requestPayload {
		t.Fatalf("created request payload = %v, want %s", repo.created[0].RequestPayload, requestPayload)
	}
	if len(repo.updated) != 1 {
		t.Fatalf("updated count = %d, want 1", len(repo.updated))
	}
	if repo.updated[0].Status != domain.JOB_EXECUTION_HISTORY_STATUS_SUCCESS {
		t.Fatalf("updated status = %s, want SUCCESS", repo.updated[0].Status)
	}
	if repo.updated[0].FinishedAt == nil {
		t.Fatal("updated finished_at is nil")
	}
}

func TestJobExecutionHistoryRunner_RunSuccessWithFinishedAt(t *testing.T) {
	repo := &fakeJobExecutionHistoryRepository{}
	requestPayload := `{"target_year":2026,"target_quarter":"QUARTER_FIRST"}`
	nowValues := []time.Time{
		time.Date(2026, 7, 9, 10, 0, 0, 0, time.UTC),
		time.Date(2026, 7, 9, 10, 1, 0, 0, time.UTC),
	}
	nowIndex := 0
	runner := JobExecutionHistoryRunner{
		Repository: repo,
		Now: func() time.Time {
			got := nowValues[nowIndex]
			nowIndex++
			return got
		},
	}

	err := runner.Run(context.Background(), JobExecutionHistoryInput{
		JobType:        domain.JOB_EXECUTION_TYPE_CREATE_INVESTOR_TRANSACTION_BALANCE_REPORT,
		JobName:        jobNameCreateInvestorTransactionBalanceReport,
		TriggerType:    domain.JOB_EXECUTION_TRIGGER_TYPE_SCHEDULED,
		RequestPayload: &requestPayload,
	}, func(context.Context) error {
		return nil
	})

	assert.NoError(t, err)
	assert.Len(t, repo.created, 1)
	assert.Len(t, repo.updated, 1)
	assert.Equal(t, domain.JOB_EXECUTION_HISTORY_STATUS_RUNNING, repo.created[0].Status)
	assert.Equal(t, requestPayload, *repo.created[0].RequestPayload)
	assert.Equal(t, domain.JOB_EXECUTION_HISTORY_STATUS_SUCCESS, repo.updated[0].Status)
	assert.Nil(t, repo.updated[0].ErrorType)
	assert.Nil(t, repo.updated[0].ErrorLog)
	assert.Equal(t, jsonutils.JSONTime(nowValues[1]), *repo.updated[0].FinishedAt)
}

func TestJobExecutionHistoryRunner_RunError(t *testing.T) {
	repo := &fakeJobExecutionHistoryRepository{}
	runner := JobExecutionHistoryRunner{
		Repository: repo,
		Now: func() time.Time {
			return time.Date(2026, 7, 13, 9, 0, 0, 0, time.UTC)
		},
	}

	jobErr := errors.New("failed to get db")
	err := runner.Run(context.Background(), JobExecutionHistoryInput{
		JobType:     domain.JOB_EXECUTION_TYPE_DEPOSIT_AUTO_PROCESS,
		JobName:     jobNameDepositAutoProcess,
		TriggerType: domain.JOB_EXECUTION_TRIGGER_TYPE_MANUAL,
	}, func(ctx context.Context) error {
		return jobErr
	})
	if err == nil {
		t.Fatal("Run() error is nil")
	}
	if len(repo.updated) != 1 {
		t.Fatalf("updated count = %d, want 1", len(repo.updated))
	}
	updated := repo.updated[0]
	if updated.Status != domain.JOB_EXECUTION_HISTORY_STATUS_FAILED {
		t.Fatalf("updated status = %s, want FAILED", updated.Status)
	}
	if updated.ErrorType == nil || *updated.ErrorType != domain.JOB_EXECUTION_ERROR_TYPE_DB_ERROR {
		t.Fatalf("updated error type = %v, want DB_ERROR", updated.ErrorType)
	}
	if updated.ErrorLog == nil || *updated.ErrorLog != jobErr.Error() {
		t.Fatalf("updated error log = %v, want %s", updated.ErrorLog, jobErr.Error())
	}
}

func TestJobExecutionHistoryRunner_RunErrorWithFinishedAt(t *testing.T) {
	repo := &fakeJobExecutionHistoryRepository{}
	runner := JobExecutionHistoryRunner{
		Repository: repo,
		Now: func() time.Time {
			return time.Date(2026, 7, 9, 10, 0, 0, 0, time.UTC)
		},
	}
	wantErr := errors.New("failed to get db")

	err := runner.Run(context.Background(), JobExecutionHistoryInput{
		JobType:     domain.JOB_EXECUTION_TYPE_BANK_TRANSFER,
		JobName:     "BulkTransferJob",
		TriggerType: domain.JOB_EXECUTION_TRIGGER_TYPE_MANUAL,
	}, func(context.Context) error {
		return wantErr
	})

	assert.ErrorIs(t, err, wantErr)
	assert.Len(t, repo.created, 1)
	assert.Len(t, repo.updated, 1)
	assert.Equal(t, domain.JOB_EXECUTION_HISTORY_STATUS_FAILED, repo.updated[0].Status)
	assert.NotNil(t, repo.updated[0].FinishedAt)
	assert.NotNil(t, repo.updated[0].ErrorType)
	assert.Equal(t, domain.JOB_EXECUTION_ERROR_TYPE_DB_ERROR, *repo.updated[0].ErrorType)
	assert.NotNil(t, repo.updated[0].ErrorLog)
	assert.Equal(t, wantErr.Error(), *repo.updated[0].ErrorLog)
}

func TestJobExecutionTriggerTypeFromContext(t *testing.T) {
	if got := JobExecutionTriggerTypeFromContext(context.Background()); got != domain.JOB_EXECUTION_TRIGGER_TYPE_SCHEDULED {
		t.Fatalf("default trigger type = %s, want SCHEDULED", got)
	}

	ctx := WithJobExecutionTriggerType(context.Background(), domain.JOB_EXECUTION_TRIGGER_TYPE_MANUAL)
	if got := JobExecutionTriggerTypeFromContext(ctx); got != domain.JOB_EXECUTION_TRIGGER_TYPE_MANUAL {
		t.Fatalf("context trigger type = %s, want MANUAL", got)
	}
}

func TestJobExecutionRequestPayload(t *testing.T) {
	payload := jobExecutionRequestPayload(ReportDateInfo{Year: "2026", Quarter: "QUARTER_FIRST"})
	if payload == nil {
		t.Fatal("payload is nil")
	}
	if *payload != `{"year":"2026","quarter":"QUARTER_FIRST"}` {
		t.Fatalf("payload = %s", *payload)
	}
}

func TestRunJobWithExecutionHistory_PersistsFailedHistoryAfterManualTransactionRollback(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite3", "file:job_execution_history_manual_rollback?mode=memory&cache=shared")
	require.NoError(t, err)
	defer engine.Close()

	_, err = engine.Exec(`
CREATE TABLE job_execution_histories (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	job_type TEXT NOT NULL,
	job_name TEXT NOT NULL,
	status TEXT NOT NULL,
	started_at DATETIME NOT NULL,
	finished_at DATETIME NULL,
	error_type TEXT NULL,
	error_log TEXT NULL,
	request_payload TEXT NULL,
	result_payload TEXT NULL,
	trigger_type TEXT NOT NULL,
	created DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated DATETIME DEFAULT CURRENT_TIMESTAMP
)`)
	require.NoError(t, err)

	session := engine.NewSession()
	defer session.Close()
	require.NoError(t, session.Begin())

	ctx := context.WithValue(context.Background(), factory.ConextDBType, session)
	ctx = WithJobExecutionTriggerType(ctx, domain.JOB_EXECUTION_TRIGGER_TYPE_MANUAL)

	jobErr := errors.New("failed to get db")
	err = runJobWithExecutionHistory(ctx, JobExecutionHistoryInput{
		JobType: domain.JOB_EXECUTION_TYPE_DEPOSIT_AUTO_PROCESS,
		JobName: jobNameDepositAutoProcess,
	}, func(ctx context.Context) error {
		return jobErr
	})
	require.Error(t, err)
	require.NoError(t, session.Rollback())

	var histories []domain.JobExecutionHistory
	require.NoError(t, engine.Find(&histories))
	require.Len(t, histories, 1)
	require.Equal(t, domain.JOB_EXECUTION_HISTORY_STATUS_FAILED, histories[0].Status)
	require.Equal(t, domain.JOB_EXECUTION_TRIGGER_TYPE_MANUAL, histories[0].TriggerType)
	require.NotNil(t, histories[0].ErrorType)
	require.Equal(t, domain.JOB_EXECUTION_ERROR_TYPE_DB_ERROR, *histories[0].ErrorType)
	require.NotNil(t, histories[0].ErrorLog)
	require.Equal(t, jobErr.Error(), *histories[0].ErrorLog)
}

func TestJobExecutionHistoryRunner_RunPanic(t *testing.T) {
	repo := &fakeJobExecutionHistoryRepository{}
	runner := JobExecutionHistoryRunner{
		Repository: repo,
		Now: func() time.Time {
			return time.Date(2026, 7, 9, 10, 0, 0, 0, time.UTC)
		},
	}

	err := runner.Run(context.Background(), JobExecutionHistoryInput{
		JobType:     domain.JOB_EXECUTION_TYPE_POINT_GRANT,
		JobName:     "PointGrantJob",
		TriggerType: domain.JOB_EXECUTION_TRIGGER_TYPE_SCHEDULED,
	}, func(context.Context) error {
		panic("unexpected panic")
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "panic recovered in job execution")
	assert.Len(t, repo.updated, 1)
	assert.Equal(t, domain.JOB_EXECUTION_HISTORY_STATUS_FAILED, repo.updated[0].Status)
	assert.NotNil(t, repo.updated[0].ErrorType)
	assert.Equal(t, domain.JOB_EXECUTION_ERROR_TYPE_UNKNOWN, *repo.updated[0].ErrorType)
	assert.NotNil(t, repo.updated[0].ErrorLog)
	assert.Contains(t, *repo.updated[0].ErrorLog, "unexpected panic")
}

func TestFormatJobExecutionError(t *testing.T) {
	tests := []struct {
		name          string
		err           error
		wantErrorType domain.JOB_EXECUTION_ERROR_TYPE
	}{
		{
			name:          "PDF_ERROR",
			err:           &domain.ReportServiceError{ErrorCode: "s003", Message: "pdf error"},
			wantErrorType: domain.JOB_EXECUTION_ERROR_TYPE_PDF_ERROR,
		},
		{
			name:          "S3_ERROR",
			err:           &domain.ReportServiceError{ErrorCode: "s002", Message: "s3 error"},
			wantErrorType: domain.JOB_EXECUTION_ERROR_TYPE_S3_ERROR,
		},
		{
			name:          "HTTP_ERROR",
			err:           errors.New("fail http transaction balance report status=500"),
			wantErrorType: domain.JOB_EXECUTION_ERROR_TYPE_HTTP_ERROR,
		},
		{
			name:          "JSON_PARSE_ERROR",
			err:           errors.New("fail http transaction balance report Unmarshal"),
			wantErrorType: domain.JOB_EXECUTION_ERROR_TYPE_JSON_PARSE_ERROR,
		},
		{
			name:          "REQUEST_ERROR",
			err:           errors.New("fail request transaction balance report"),
			wantErrorType: domain.JOB_EXECUTION_ERROR_TYPE_REQUEST_ERROR,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatJobExecutionError(tt.err)
			assert.Equal(t, tt.wantErrorType, got.ErrorType)
			assert.Equal(t, tt.err.Error(), got.ErrorLog)
		})
	}
}

func TestFormatJobExecutionErrorTruncatesErrorLog(t *testing.T) {
	err := errors.New(strings.Repeat("a", maxJobExecutionErrorLogLength+1))

	got := FormatJobExecutionError(err)

	assert.Len(t, got.ErrorLog, maxJobExecutionErrorLogLength)
}
