package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	jobExecutionHistoryMysql "github.com/illuminarean/energy-shares-backend-repo/repository/job_execution_history/mysql"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"xorm.io/xorm"
)

const maxJobExecutionErrorLogLength = 4000

type jobExecutionTriggerTypeContextKey struct{}

const (
	jobNameCreateInvestorTransactionBalanceReport = "投資家向け取引残高報告書作成"
	jobNameCreateInvestorPaymentStatement         = "投資家向け支払調書作成"
	jobNameCreateTaxOfficePaymentStatement        = "税務署向け支払調書作成"
	jobNameCreateTaxOfficePaymentStatementSum     = "税務署向け支払調書合計表作成"
	jobNameContractCreate                         = "契約締結時書面作成"
	jobNameBulkTransfer                           = "一括振込依頼"
	jobNameBulkTransferResultCheck                = "一括振込結果確認"
	jobNameBulkTransferStatusCheck                = "一括振込ステータス確認"
	jobNameGetEquityVaTransaction                 = "出資用VA入金明細取得"
	jobNameVaDepositNotUseItemCheck               = "VA入金未使用明細確認"
	jobNameDepositRequestExpire                   = "入金申請期限切れ処理"
	jobNameDepositAutoProcess                     = "入金自動消込処理"
	jobNameRepaymentDividend                      = "償還・分配処理"
	jobNameLotteryCSVUpload                       = "抽選CSVアップロード処理"
	jobNameMailRequestRetry                       = "メール送信リトライ"
	jobNameUserAddressCheck                       = "ユーザー住所確認"
	jobNameMyNumberDeleteFile                     = "マイナンバーファイル削除"
	jobNameCreateHolidayCalendar                  = "休日カレンダー作成"
	jobNameCloseBanners                           = "バナー公開終了処理"
	jobNameBlogPostRequest                        = "ブログ投稿予約処理"
	jobNameDeleteTestUsers                        = "テストユーザー削除"
	jobNameRePublish                              = "未送信イベント再送"
	jobNameAttendanceReminder                     = "勤怠連絡リマインド"
	jobNameAttendanceMorningReminder              = "勤怠連絡リマインド（当日朝）"
)

type JobExecutionHistoryInput struct {
	JobType        domain.JOB_EXECUTION_TYPE
	JobName        string
	TriggerType    domain.JOB_EXECUTION_TRIGGER_TYPE
	RequestPayload *string
}

func WithJobExecutionTriggerType(ctx context.Context, triggerType domain.JOB_EXECUTION_TRIGGER_TYPE) context.Context {
	return context.WithValue(ctx, jobExecutionTriggerTypeContextKey{}, triggerType)
}

func JobExecutionTriggerTypeFromContext(ctx context.Context) domain.JOB_EXECUTION_TRIGGER_TYPE {
	if ctx == nil {
		return domain.JOB_EXECUTION_TRIGGER_TYPE_SCHEDULED
	}
	if triggerType, ok := ctx.Value(jobExecutionTriggerTypeContextKey{}).(domain.JOB_EXECUTION_TRIGGER_TYPE); ok && triggerType != "" {
		return triggerType
	}
	return domain.JOB_EXECUTION_TRIGGER_TYPE_SCHEDULED
}

func runJobWithExecutionHistory(
	ctx context.Context,
	input JobExecutionHistoryInput,
	jobFunc func(context.Context) error,
) error {
	input.TriggerType = JobExecutionTriggerTypeFromContext(ctx)
	return JobExecutionHistoryRunner{
		Repository:     jobExecutionHistoryMysql.New(factory.NewDBFactory()),
		HistoryContext: jobExecutionHistoryContext(ctx),
	}.Run(ctx, input, jobFunc)
}

func jobExecutionHistoryContext(ctx context.Context) context.Context {
	if ctx == nil {
		return ctx
	}
	if session, ok := ctx.Value(factory.ConextDBType).(*xorm.Session); ok {
		return context.WithValue(ctx, factory.ConextDBType, session.Engine())
	}
	return ctx
}

func jobExecutionRequestPayload(v interface{}) *string {
	bytes, err := json.Marshal(v)
	if err != nil {
		logrus.WithError(err).Warn("failed to marshal job execution request payload")
		return nil
	}
	payload := string(bytes)
	return &payload
}

type JobExecutionHistoryRunner struct {
	Repository     cdomain.JobExecutionHistoryRepository
	HistoryContext context.Context
	Now            func() time.Time
}

func (r JobExecutionHistoryRunner) Run(ctx context.Context, input JobExecutionHistoryInput, jobFunc func(context.Context) error) (err error) {
	if r.Repository == nil {
		return errors.New("job execution history repository is nil")
	}
	if jobFunc == nil {
		return errors.New("job function is nil")
	}

	now := r.now()
	history := domain.JobExecutionHistory{
		JobType:        input.JobType,
		JobName:        input.JobName,
		Status:         domain.JOB_EXECUTION_HISTORY_STATUS_RUNNING,
		StartedAt:      jsonutils.JSONTime(now),
		RequestPayload: input.RequestPayload,
		TriggerType:    input.TriggerType,
	}
	historyCtx := r.historyContext(ctx)

	id, createErr := r.Repository.Create(historyCtx, history)
	if createErr != nil {
		return errors.Wrap(createErr, "failed to create job execution history")
	}
	history.Id = id

	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("panic recovered in job execution: %v", recovered)
		}

		finishedAt := jsonutils.JSONTime(r.now())
		history.FinishedAt = &finishedAt
		if err != nil {
			history.Status = domain.JOB_EXECUTION_HISTORY_STATUS_FAILED
			errorInfo := FormatJobExecutionError(err)
			history.ErrorType = &errorInfo.ErrorType
			history.ErrorLog = &errorInfo.ErrorLog
		} else {
			history.Status = domain.JOB_EXECUTION_HISTORY_STATUS_SUCCESS
			history.ErrorType = nil
			history.ErrorLog = nil
		}

		if updateErr := r.Repository.Update(historyCtx, history); updateErr != nil {
			wrappedUpdateErr := errors.Wrap(updateErr, "failed to update job execution history")
			if err == nil {
				err = wrappedUpdateErr
				return
			}
			logrus.WithError(wrappedUpdateErr).Error("failed to update job execution history")
		}
	}()

	return jobFunc(ctx)
}

func (r JobExecutionHistoryRunner) historyContext(ctx context.Context) context.Context {
	if r.HistoryContext != nil {
		return r.HistoryContext
	}
	return ctx
}

func (r JobExecutionHistoryRunner) now() time.Time {
	if r.Now != nil {
		return r.Now()
	}
	return time.Now()
}

func RunWithJobExecutionHistory(
	ctx context.Context,
	repository cdomain.JobExecutionHistoryRepository,
	input JobExecutionHistoryInput,
	jobFunc func(context.Context) error,
) error {
	return JobExecutionHistoryRunner{Repository: repository}.Run(ctx, input, jobFunc)
}

type JobExecutionErrorInfo struct {
	ErrorType domain.JOB_EXECUTION_ERROR_TYPE
	ErrorLog  string
}

func FormatJobExecutionError(err error) JobExecutionErrorInfo {
	if err == nil {
		return JobExecutionErrorInfo{}
	}

	return JobExecutionErrorInfo{
		ErrorType: classifyJobExecutionError(err),
		ErrorLog:  truncateJobExecutionErrorLog(err.Error()),
	}
}

func classifyJobExecutionError(err error) domain.JOB_EXECUTION_ERROR_TYPE {
	var reportServiceErr *domain.ReportServiceError
	if errors.As(err, &reportServiceErr) {
		switch {
		case reportServiceErr.IsPDFError():
			return domain.JOB_EXECUTION_ERROR_TYPE_PDF_ERROR
		case reportServiceErr.IsS3Error():
			return domain.JOB_EXECUTION_ERROR_TYPE_S3_ERROR
		default:
			return domain.JOB_EXECUTION_ERROR_TYPE_UNKNOWN
		}
	}

	var apiErr *factory.Error
	if errors.As(err, &apiErr) {
		switch apiErr.Code {
		case factory.ErrorParameter.Code, factory.ErrorMissParameter.Code, factory.ErrorIllegalRequest.Code:
			return domain.JOB_EXECUTION_ERROR_TYPE_VALIDATION_ERROR
		case factory.ErrorDB.Code:
			return domain.JOB_EXECUTION_ERROR_TYPE_DB_ERROR
		case factory.ErrorAWSUpload.Code, factory.ErrorAWSDelete.Code, factory.ErrorAWSGet.Code:
			return domain.JOB_EXECUTION_ERROR_TYPE_S3_ERROR
		default:
			return domain.JOB_EXECUTION_ERROR_TYPE_UNKNOWN
		}
	}

	errText := err.Error()
	lowerErrText := strings.ToLower(errText)
	switch {
	case strings.Contains(errText, "fail http transaction balance report status"):
		return domain.JOB_EXECUTION_ERROR_TYPE_HTTP_ERROR
	case strings.Contains(errText, "Unmarshal") || strings.Contains(lowerErrText, "json"):
		return domain.JOB_EXECUTION_ERROR_TYPE_JSON_PARSE_ERROR
	case strings.Contains(errText, "fail request transaction balance report"):
		return domain.JOB_EXECUTION_ERROR_TYPE_REQUEST_ERROR
	case strings.Contains(lowerErrText, "insert") ||
		strings.Contains(lowerErrText, "update") ||
		strings.Contains(lowerErrText, "failed to get db") ||
		strings.Contains(lowerErrText, "db error"):
		return domain.JOB_EXECUTION_ERROR_TYPE_DB_ERROR
	}

	return domain.JOB_EXECUTION_ERROR_TYPE_UNKNOWN
}

func truncateJobExecutionErrorLog(errorLog string) string {
	if len(errorLog) <= maxJobExecutionErrorLogLength {
		return errorLog
	}
	return errorLog[:maxJobExecutionErrorLogLength]
}
