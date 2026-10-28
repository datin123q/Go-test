package controllers

import (
	"context"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
	"github.com/pkg/errors"

	"energyshares-job-scheduler/jobs"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
	"github.com/illuminarean/energy-shares-backend-repo/service"
)

type ExecuteApiControllers struct {
}

type GetEkycByApplicantIdRequest struct {
	ApplicantId string `json:"applicant_id" form:"applicant_id" swagger:"desc(Nexway applicant id)"`
}

type LotteryCSVUploadJobRequest struct {
	JobID string `json:"job_id" form:"job_id" swagger:"desc(Target lottery csv upload job id. If empty, process queued jobs)"`
}

// Initはバッチ作業ジョブを手動で実行するAPIを登録します。
// - QAのテスト用途であり、PROD環境では実行されてはいけません。
// - SetDescriptionには実際のバッチジョブがいつ、何をするかを簡単に記入します。
func (c ExecuteApiControllers) Init(g echoswagger.ApiGroup) {

	g.POST("/refresh-token-job", c.RefreshToken).
		SetSummary("GMOトークン更新").
		SetDescription("GMOトークン更新実施。スケジュール：毎月20日10時に実行").
		AddResponse(http.StatusOK, "GMOトークン更新実行", nil, nil)

	g.POST("/get-account-info-job", c.GetAccountInfoJob).
		SetSummary("GMO口座情報照会").
		SetDescription("GMO口座情報照会。スケジュール：30分ごと").
		AddResponse(http.StatusOK, "GMO口座情報照会実行", nil, nil)

	g.POST("/bulk-transfer-job", c.BulkTransferJob).
		SetSummary("一括送金").
		SetDescription("一括送金作業。スケジュール：10分ごと").
		AddResponse(http.StatusOK, "一括送金実行", nil, nil)

	g.POST("/bulk-transfer-result-check-job", c.BulkTransferResultCheckJob).
		SetSummary("一括送金結果確認").
		SetDescription("一括送金結果確認作業。スケジュール：10分ごと").
		AddResponse(http.StatusOK, "一括送金結果確認実行", nil, nil)

	g.POST("/bulk-transfer-status-check-job", c.BulkTransferStatusCheckJob).
		SetSummary("一括送金ステータス確認").
		SetDescription("一括送金ステータス確認作業。スケジュール：毎日12時、16時に実行").
		AddResponse(http.StatusOK, "一括送金ステータス確認実行", nil, nil)

	g.POST("/get-equity-va-transaction-job", c.GetEquityVaTransactionJob).
		SetSummary("GMO仮想口座入金確認").
		SetDescription("GMO仮想口座入金確認作業。スケジュール：毎日9時").
		AddResponse(http.StatusOK, "GMO仮想口座入金確認実行", nil, nil)

	g.POST("/auto-investment-job", c.AutoInvestmentJob).
		SetSummary("自動投資").
		SetDescription("投資待機自動投資作業。スケジュール：毎日9時").
		AddResponse(http.StatusOK, "自動投資実行", nil, nil)

	g.POST("/check-investment-propose-job", c.CheckInvestmentProposeJob).
		SetSummary("投資リクエスト有効期限確認").
		SetDescription("投資待機者投資リクエスト有効期限確認作業。スケジュール：毎日午前0時").
		AddResponse(http.StatusOK, "投資リクエスト有効期限確認実行", nil, nil)

	g.POST("/funding-open-job", c.FundingOpenJob).
		SetSummary("ファンディングステータス更新（募集中）").
		SetDescription("ファンディングステータスを募集中に更新。スケジュール：毎時正時").
		AddResponse(http.StatusOK, "ファンディングステータス更新（募集中）実行", nil, nil)

	g.POST("/funding-close-job", c.FundingCloseJob).
		SetSummary("ファンディングステータス更新（募集終了）").
		SetDescription("ファンディングステータスを募集終了に更新。スケジュール：毎時正時").
		AddResponse(http.StatusOK, "ファンディングステータス更新（募集終了）実行", nil, nil)

	g.POST("/repayment-dividend-job", c.RepaymentDividendJob).
		SetSummary("ファンド配当/返済").
		SetDescription("ファンド配当/返済実施作業。スケジュール：毎日9時").
		AddResponse(http.StatusOK, "ファンド配当/返済実行", nil, nil)

	g.POST("/va-deposit-not-use-item-check-job", c.VaDepositNotUseItemCheckJob).
		SetSummary("未使用入金額確認").
		SetDescription("未使用入金額確認作業。スケジュール：10分ごと").
		AddResponse(http.StatusOK, "未使用入金額確認実行", nil, nil)

	g.POST("/deposit-request-expire-job", c.DepositRequestExpireJob).
		SetSummary("入金リクエスト有効期限確認").
		SetDescription("入金リクエスト有効期限メール送信作業。スケジュール：毎日午前0時").
		AddResponse(http.StatusOK, "入金リクエスト有効期限メール送信実行", nil, nil)

	g.POST("/user-address-check-job", c.UserAddressCheckJob).
		SetSummary("郵便コード送信").
		SetDescription("郵便コード送信作業。スケジュール：毎日9時").
		AddResponse(http.StatusOK, "郵便コード送信実行", nil, nil)

	g.POST("/my-number-delete-file-job", c.MyNumberDeleteFileJob).
		SetSummary("マイナンバー削除").
		SetDescription("マイナンバー削除作業。スケジュール：毎日9時").
		AddResponse(http.StatusOK, "マイナンバー削除実行", nil, nil)

	g.POST("/mail-request-retry-job", c.MailRequestRetryJob).
		SetSummary("メール送信再リクエスト").
		SetDescription("失敗したメール送信再リクエスト作業。再リクエスト後は成功/失敗に関わらずテーブルから削除。スケジュール：1分ごと").
		AddResponse(http.StatusOK, "メール送信再リクエスト実行", nil, nil)

	g.POST("/contract-create-job", c.ContractCreateJob).
		SetSummary("契約書作成").
		SetDescription("契約時書面作成作業。スケジュール：1分ごと").
		AddResponse(http.StatusOK, "契約書作成実行", nil, nil)

	g.POST("/deposit-auto-process-job", c.DepositAutoProcessJob).
		SetSummary("未使用預金自動投資").
		SetDescription("未使用預金自動投資作業。スケジュール：毎日9時").
		AddResponse(http.StatusOK, "未使用預金自動投資実行", nil, nil)

	g.POST("/create-investor-transaction-balance-report-job", c.CreateInvestorTransactionBalanceReportJob).
		AddParamBody(jobs.ReportDateInfo{}, "body", "法定帳簿四半期情報", false).
		SetSummary("投資家取引残高報告").
		SetDescription("投資家取引残高報告作業。スケジュール：四半期ごと9時").
		AddResponse(http.StatusOK, "投資家取引残高報告実行", nil, nil)

	g.POST("/create-investor-payment-statement-job", c.CreateInvestorPaymentStatementJob).
		AddParamBody(jobs.ReportDateInfo{}, "body", "法定帳簿四半期情報", false).
		SetSummary("投資家支払調書作成").
		SetDescription("投資家支払調書作成作業。スケジュール：毎年1月1日9時").
		AddResponse(http.StatusOK, "投資家支払調書作成実行", nil, nil)

	g.POST("/create-tax-office-payment-statement-job", c.CreateTaxOfficePaymentStatementJob).
		AddParamBody(jobs.ReportDateInfo{}, "body", "法定帳簿四半期情報", false).
		SetSummary("税務署用支払調書作成").
		SetDescription("税務署用支払調書作成作業。スケジュール：毎年1月1日9時").
		AddResponse(http.StatusOK, "税務署用支払調書作成実行", nil, nil)

	g.POST("/create-tax-office-payment-statement-sum-job", c.CreateTaxOfficePaymentStatementSumJob).
		AddParamBody(jobs.ReportDateInfo{}, "body", "法定帳簿四半期情報", false).
		SetSummary("税務署用支払調書合計表作成").
		SetDescription("税務署用支払調書合計表作成作業。スケジュール：毎年1月1日9時").
		AddResponse(http.StatusOK, "税務署用支払調書合計表作成実行", nil, nil)

	g.POST("/blog-post-request", c.BlogPostRequest).
		SetSummary("ブログポスト申請").
		SetDescription("ブログポスト申請作業。スケジュール：毎時正時").
		AddResponse(http.StatusOK, "ブログポスト申請実行", nil, nil)

	g.POST("/create-holiday-calendar-job", c.CreateHolidayCalendarJob).
		SetSummary("祝日カレンダー作成").
		SetDescription("祝日カレンダー作成作業。スケジュール：毎日午前0時").
		AddResponse(http.StatusOK, "祝日カレンダー作成実行", nil, nil)

	g.POST("/close-banners-job", c.CloseBannersJob).
		SetSummary("期間終了バナー非表示処理").
		SetDescription("期間終了バナー非表示処理作業。スケジュール：毎日午前0時").
		AddResponse(http.StatusOK, "期間終了バナー非表示処理実行", nil, nil)

	g.POST("/delete-testusers-job", c.DeleteTestUsersJob).
		SetSummary("QA、STGテストユーザー削除").
		SetDescription("QA、STGテストユーザーアカウント削除 - 毎月1日午前1時").
		AddResponse(http.StatusOK, "QA、STGテストユーザーアカウント削除完了", nil, nil)

	g.POST("/attendance-reminder-job", c.AttendanceReminderJob).
		SetSummary("勤怠連絡リマインド").
		SetDescription("翌営業日の勤怠連絡を促すSlack投稿。営業日の15:53(JST)に実行").
		AddResponse(http.StatusOK, "勤怠連絡リマインド実行", nil, nil)

	g.POST("/attendance-morning-reminder-job", c.AttendanceMorningReminderJob).
		SetSummary("勤怠連絡リマインド（当日朝）").
		SetDescription("本日の勤怠連絡を促すSlack投稿。営業日の6:53(JST)に実行").
		AddResponse(http.StatusOK, "勤怠連絡リマインド（当日朝）実行", nil, nil)

	g.POST("/lottery-csv-upload-job", c.LotteryCSVUploadJob).
		AddParamBody(LotteryCSVUploadJobRequest{}, "body", "lottery csv upload job id", false).
		SetSummary("抽選CSV一括当落").
		SetDescription("抽選CSV一括当落処理。スケジュール実行またはjob_id指定で単発実行").
		AddResponse(http.StatusOK, "抽選CSV一括当落実行", nil, nil)

	g.POST("/ekyc/applicant", c.GetEkycByApplicantId).
		AddParamBody(GetEkycByApplicantIdRequest{}, "body", "applicant id", true).
		AddResponse(http.StatusOK, "applicant id で eKYC を取得", service.EkycResultOutput{}, nil).
		SetDescription("applicant id を指定して Nexway の eKYC 結果を取得します（検証環境での手動実行用）").
		SetSummary("applicant id で eKYC 結果を取得")
}

// executeJobはバッチ作業を実行する関数です。
func executeJob(ctx echo.Context, jobFunc func(context.Context) error) error {
	if serviceutils.EnvIsProd() {
		// 同ファイルのGetEkycByApplicantIdと同じ本番禁止ガードなので、返すステータスも揃える
		// (MIR-1133でdefaultが500を返すようになるまでは、素のerrors.Newでも空の200が返って
		// 気付けなかった)。
		return factory.ErrorPermissionDenied.New(errors.New("this API is not for production"))
	}

	jobCtx := jobs.WithJobExecutionTriggerType(ctx.Request().Context(), domain.JOB_EXECUTION_TRIGGER_TYPE_MANUAL)
	if err := jobFunc(jobCtx); err != nil {
		// jobFuncが既に*factory.Errorを返している場合(例: report_jobsのErrorNotFoundReport)は
		// そのステータス・コードをそのまま活かす。ここで一律ErrorSystemに包み直すと、errors.Asが
		// 外側から先に一致するため元のステータスが失われ、404が500に化けるような退行になる。
		var apiError *factory.Error
		if errors.As(err, &apiError) {
			return err
		}
		// 素のエラーは一律500が妥当なため、明示的にErrorSystemとして返す(このexecuteJobが
		// MIR-1133の本題である「空の200」を最も広く引き起こしていた経路)。
		return factory.ErrorSystem.New(errors.Wrap(err, "fail to execute job"))
	}

	return renderSucc(ctx, http.StatusOK, nil)
}

func (c ExecuteApiControllers) RefreshToken(ctx echo.Context) error {
	return executeJob(ctx, (&jobs.AozoraJobs{}).RefreshToken)
}

func (c ExecuteApiControllers) GetAccountInfoJob(ctx echo.Context) error {
	return executeJob(ctx, (&jobs.AozoraJobs{}).GetAccountInfo)
}

func (c ExecuteApiControllers) BulkTransferJob(ctx echo.Context) error {
	return executeJob(ctx, (&jobs.AozoraJobs{}).BulkTransferJob)
}

func (c ExecuteApiControllers) BulkTransferResultCheckJob(ctx echo.Context) error {
	return executeJob(ctx, (&jobs.AozoraJobs{}).BulkTransferResultCheckJob)
}

func (c ExecuteApiControllers) BulkTransferStatusCheckJob(ctx echo.Context) error {
	return executeJob(ctx, (&jobs.AozoraJobs{}).BulkTransferStatusCheckJob)
}

func (c ExecuteApiControllers) GetEquityVaTransactionJob(ctx echo.Context) error {
	return executeJob(ctx, (&jobs.AozoraJobs{}).GetEquityVaTransactionJob)
}

func (c ExecuteApiControllers) AutoInvestmentJob(ctx echo.Context) error {
	return executeJob(ctx, (&jobs.WaitingInvestmentJobs{}).AutoInvestmentJob)
}

func (c ExecuteApiControllers) CheckInvestmentProposeJob(ctx echo.Context) error {
	return executeJob(ctx, (&jobs.WaitingInvestmentJobs{}).CheckInvestmentProposeJob)
}

func (c ExecuteApiControllers) FundingOpenJob(ctx echo.Context) error {
	return executeJob(ctx, jobs.ProductJobs{}.FundingOpenJob)
}

func (c ExecuteApiControllers) FundingCloseJob(ctx echo.Context) error {
	return executeJob(ctx, jobs.ProductJobs{}.FundingCloseJob)
}

func (c ExecuteApiControllers) RepaymentDividendJob(ctx echo.Context) error {
	return executeJob(ctx, jobs.RepaymentDividendJobs{}.RepaymentDividendProcess)
}

func (c ExecuteApiControllers) VaDepositNotUseItemCheckJob(ctx echo.Context) error {
	return executeJob(ctx, jobs.DepositCheckJobs{}.VaDepositNotUseItemCheckJob)
}

func (c ExecuteApiControllers) DepositRequestExpireJob(ctx echo.Context) error {
	return executeJob(ctx, jobs.DepositCheckJobs{}.DepositRequestExpireJob)
}

func (c ExecuteApiControllers) DepositAutoProcessJob(ctx echo.Context) error {
	return executeJob(ctx, jobs.DepositCheckJobs{}.DepositAutoProcessJob)
}

func (c ExecuteApiControllers) UserAddressCheckJob(ctx echo.Context) error {
	return executeJob(ctx, jobs.UserAddressCheckJobs{}.UserAddressCheckJob)
}

func (c ExecuteApiControllers) MyNumberDeleteFileJob(ctx echo.Context) error {
	return executeJob(ctx, jobs.MyNumberJobs{}.MyNumberDeleteFileJob)
}

func (c ExecuteApiControllers) MailRequestRetryJob(ctx echo.Context) error {
	return executeJob(ctx, jobs.MailRetryJobs{}.MailRequestRetryJob)
}

func (c ExecuteApiControllers) ContractCreateJob(ctx echo.Context) error {
	return executeJob(ctx, jobs.ContractJobs{}.ContractCreateJob)
}

func (c ExecuteApiControllers) CreateInvestorTransactionBalanceReportJob(ctx echo.Context) error {
	var request jobs.ReportDateInfo

	if err := ctx.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	ctxWithValue := context.WithValue(ctx.Request().Context(), "request", request)
	ctx.SetRequest(ctx.Request().WithContext(ctxWithValue))

	return executeJob(ctx, jobs.ReportJobs{}.CreateInvestorTransactionBalanceReportJob)
}

func (c ExecuteApiControllers) CreateInvestorPaymentStatementJob(ctx echo.Context) error {
	var request jobs.ReportDateInfo

	if err := ctx.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	ctxWithValue := context.WithValue(ctx.Request().Context(), "request", request)
	ctx.SetRequest(ctx.Request().WithContext(ctxWithValue))

	return executeJob(ctx, jobs.ReportJobs{}.CreateInvestorPaymentStatementJob)
}

func (c ExecuteApiControllers) CreateTaxOfficePaymentStatementJob(ctx echo.Context) error {
	var request jobs.ReportDateInfo

	if err := ctx.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	ctxWithValue := context.WithValue(ctx.Request().Context(), "request", request)
	ctx.SetRequest(ctx.Request().WithContext(ctxWithValue))

	return executeJob(ctx, jobs.ReportJobs{}.CreateTaxOfficePaymentStatementJob)
}

func (c ExecuteApiControllers) CreateTaxOfficePaymentStatementSumJob(ctx echo.Context) error {
	var request jobs.ReportDateInfo

	if err := ctx.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	ctxWithValue := context.WithValue(ctx.Request().Context(), "request", request)
	ctx.SetRequest(ctx.Request().WithContext(ctxWithValue))

	return executeJob(ctx, jobs.ReportJobs{}.CreateTaxOfficePaymentStatementSumJob)
}

func (c ExecuteApiControllers) BlogPostRequest(ctx echo.Context) error {
	return executeJob(ctx, jobs.BlogPostJobs{}.BlogPostRequest)
}

func (c ExecuteApiControllers) CreateHolidayCalendarJob(ctx echo.Context) error {
	return executeJob(ctx, jobs.CalendarJobs{}.CreateHolidayCalendarJob)
}

func (c ExecuteApiControllers) CloseBannersJob(ctx echo.Context) error {
	return executeJob(ctx, jobs.BannerJobs{}.CloseBannersJob)
}

func (c ExecuteApiControllers) DeleteTestUsersJob(ctx echo.Context) error {
	return executeJob(ctx, jobs.TestUsersJobs{}.DeleteTestUsersJob)
}

func (c ExecuteApiControllers) AttendanceReminderJob(ctx echo.Context) error {
	return executeJob(ctx, jobs.AttendanceReminderJobs{}.AttendanceReminderJob)
}

func (c ExecuteApiControllers) AttendanceMorningReminderJob(ctx echo.Context) error {
	return executeJob(ctx, jobs.AttendanceReminderJobs{}.AttendanceMorningReminderJob)
}

func (c ExecuteApiControllers) LotteryCSVUploadJob(ctx echo.Context) error {
	var req LotteryCSVUploadJobRequest

	if err := ctx.Bind(&req); err != nil {
		return factory.ErrorParameter.New(err)
	}

	req.JobID = strings.TrimSpace(req.JobID)
	if req.JobID == "" {
		return executeJob(ctx, jobs.LotteryCSVUploadJobs{}.ProcessQueuedLotteryCSVUploadJob)
	}

	return executeJob(ctx, func(jobCtx context.Context) error {
		return jobs.LotteryCSVUploadJobs{}.ProcessLotteryCSVUploadJob(jobCtx, req.JobID)
	})
}

func (c ExecuteApiControllers) GetEkycByApplicantId(ctx echo.Context) error {
	if serviceutils.EnvIsProd() {
		return factory.ErrorPermissionDenied.New(errors.New("not allowed in production"))
	}

	var req GetEkycByApplicantIdRequest
	if err := ctx.Bind(&req); err != nil {
		return factory.ErrorParameter.New(err)
	}
	req.ApplicantId = strings.TrimSpace(req.ApplicantId)
	if req.ApplicantId == "" {
		return factory.ErrorParameter.New(errors.New("applicant_id is required"))
	}

	result, err := service.NewNexwayService().GetEkycResultByApplicantId(ctx.Request().Context(), req.ApplicantId)
	if err != nil {
		return factory.ErrorSystem.New(errors.Wrap(err, "Failed call nexway service"))
	}
	if result != nil && result.ErrorCode != "" {
		return factory.ErrorSystem.New(errors.New("Failed call nexway service"))
	}

	return renderSucc(ctx, http.StatusOK, service.SanitizeEkycResultOutputForLog(result))
}
