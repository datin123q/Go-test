package jobs

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"

	"github.com/sirupsen/logrus"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"
)

type ReportDateInfo struct {
	Year    string `json:"year"`
	Quarter string `json:"quarter"`
}

func (r ReportDateInfo) getDateInfo() (ReportDateInfo, error) {
	// 비어있지 않으면 그대로 반환
	if !r.IsEmpty() {
		return r, nil
	}

	// 비어있다면 현재 날짜 기준으로 연도와 분기를 구함
	year := strconv.Itoa(time.Now().Year())
	month := time.Now().Month()
	var quarter domain.QUARTER_TYPE

	switch month {
	case 1, 2, 3:
		year = strconv.Itoa(time.Now().Year() - 1)
		quarter = domain.QUARTER_FOURTH
	case 4, 5, 6:
		quarter = domain.QUARTER_FIRST
	case 7, 8, 9:
		quarter = domain.QUARTER_SECOND
	case 10, 11, 12:
		quarter = domain.QUARTER_THIRD
	default:
		quarter = domain.QUARTER_FOURTH
	}

	return ReportDateInfo{Year: year, Quarter: string(quarter)}, nil
}

func (r ReportDateInfo) IsEmpty() bool {
	if len(r.Year) == 0 || len(r.Quarter) == 0 {
		return true
	}

	return false
}

// FIXME context에 저장된 값을 가져옴, 인터페이스 변경이 가장 적은 방법으로 context에 저장된 값(request)을 가져오도록 처리
// 향후에 좀 더 깔끔한 방법으로 변경할 수 있도록 리팩토링 필요
func (ReportDateInfo) FromContext(ctx context.Context) ReportDateInfo {
	// ctx가 비어있으면 빈 값 반환
	if ctx == nil {
		return ReportDateInfo{}
	}

	request := ctx.Value("request")

	// request가 ReportDateInfo 타입이 아니면 빈 값 반환
	reportDateInfo, ok := request.(ReportDateInfo)
	if !ok {
		return ReportDateInfo{}
	}

	return reportDateInfo
}

type ReportJobs struct{}

func transactionBalanceReportErrorFields(err error) logrus.Fields {
	fields := logrus.Fields{
		"err":        err,
		"error_type": "UNKNOWN",
	}

	var reportServiceErr *domain.ReportServiceError
	if errors.As(err, &reportServiceErr) {
		fields["error_code"] = reportServiceErr.ErrorCode
		fields["error_message"] = reportServiceErr.Message
		switch {
		case reportServiceErr.IsPDFError():
			fields["error_type"] = "PDF_ERROR"
		case reportServiceErr.IsS3Error():
			fields["error_type"] = "S3_ERROR"
		default:
			fields["error_type"] = "REPORT_SERVICE_ERROR"
		}
		return fields
	}

	var apiErr *factory.Error
	if errors.As(err, &apiErr) {
		fields["factory_error_code"] = apiErr.Code
		fields["factory_error_message"] = apiErr.Message
		if apiErr.Code == factory.ErrorNotFoundReport.Code {
			fields["error_type"] = "NOT_FOUND"
			return fields
		}
		fields["error_type"] = "FACTORY_ERROR"
		return fields
	}

	errText := err.Error()
	switch {
	case strings.Contains(errText, "fail http transaction balance report status"):
		fields["error_type"] = "HTTP_ERROR"
	case strings.Contains(errText, "Unmarshal"):
		fields["error_type"] = "JSON_PARSE_ERROR"
	case strings.Contains(errText, "fail request transaction balance report"):
		fields["error_type"] = "REQUEST_ERROR"
	case strings.Contains(errText, "insert") || strings.Contains(errText, "Insert"):
		fields["error_type"] = "DB_INSERT_ERROR"
	}
	return fields
}

func isNotFoundReportError(err error) bool {
	var apiErr *factory.Error
	return errors.As(err, &apiErr) && apiErr.Code == factory.ErrorNotFoundReport.Code
}

// 투자자용 거래 잔액 보고서 생성
func (ReportJobs) CreateInvestorTransactionBalanceReportJob(ctx context.Context) error {
	request := ReportDateInfo{}.FromContext(ctx)
	return runJobWithExecutionHistory(ctx, JobExecutionHistoryInput{
		JobType:        domain.JOB_EXECUTION_TYPE_CREATE_INVESTOR_TRANSACTION_BALANCE_REPORT,
		JobName:        jobNameCreateInvestorTransactionBalanceReport,
		RequestPayload: jobExecutionRequestPayload(request),
	}, createInvestorTransactionBalanceReportJob)
}

func createInvestorTransactionBalanceReportJob(ctx context.Context) error {
	rdi, err := ReportDateInfo{}.FromContext(ctx).getDateInfo()
	if err != nil {
		return err
	}

	logrus.WithFields(logrus.Fields{
		"category": "create report",
		"time":     time.Now(),
	}).Info("CreateInvestorTransactionBalanceReportJob Jobs Start")

	quarterReport := &domain.ReportOutput{
		UserId:         0,
		SpvId:          0,
		IsOpen:         false,
		SubmissionType: domain.SUBMISSION_INVESTORS,
		ReportType:     domain.REPORT_TRANSACTION_BALANCE_ADMIN_REPORT,
		Year:           rdi.Year,
		Quarter:        domain.QUARTER_TYPE(rdi.Quarter),
	}
	_, err = quarterReport.CreateTransactionBalanceReportAll(ctx)
	if err != nil {
		fields := transactionBalanceReportErrorFields(err)
		fields["category"] = "create transaction balance report"
		fields["job_name"] = "CreateInvestorTransactionBalanceReportJob"
		fields["report_type"] = quarterReport.ReportType
		fields["submission_type"] = quarterReport.SubmissionType
		fields["year"] = rdi.Year
		fields["quarter"] = domain.QUARTER_TYPE(quarterReport.Quarter)
		logrus.WithFields(fields).Error("not created transaction balance report all")
		return errors.Wrap(err, "fail create transaction balance report all")
	}

	items, err := domain.User{}.FindByAuditStatus(ctx, domain.COMPLETED_AUDIT)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"category":   "create transaction balance report",
			"job_name":   "CreateInvestorTransactionBalanceReportJob",
			"error_type": "DB_ERROR",
			"year":       rdi.Year,
			"err":        err,
		}).Error("failed find audit completed users")
		return errors.Wrap(err, "fail find audit completed users")
	}
	for _, user := range items {

		ro := &domain.ReportOutput{
			UserId:         user.Id,
			SpvId:          0,
			IsOpen:         false,
			SubmissionType: domain.SUBMISSION_INVESTORS,
			ReportType:     domain.REPORT_TRANSACTION_BALANCE_REPORT,
			Year:           rdi.Year,
			Quarter:        domain.QUARTER_TYPE(rdi.Quarter),
		}
		_, err = ro.CreateTransactionBalanceReport(ctx)
		if err != nil {
			fields := transactionBalanceReportErrorFields(err)
			fields["category"] = "create transaction balance report"
			fields["job_name"] = "CreateInvestorTransactionBalanceReportJob"
			fields["report_type"] = ro.ReportType
			fields["submission_type"] = ro.SubmissionType
			fields["user_id"] = user.Id
			fields["year"] = rdi.Year
			fields["quarter"] = domain.QUARTER_TYPE(ro.Quarter)
			if isNotFoundReportError(err) {
				logrus.WithFields(fields).Info("not created transaction balance report")
				continue
			}
			logrus.WithFields(fields).Error("not created transaction balance report")
			return errors.Wrap(err, "fail create transaction balance report")
		}
	}

	return nil
}

// 투자자용 지불조서 생성
func (ReportJobs) CreateInvestorPaymentStatementJob(ctx context.Context) error {
	request := ReportDateInfo{}.FromContext(ctx)
	return runJobWithExecutionHistory(ctx, JobExecutionHistoryInput{
		JobType:        domain.JOB_EXECUTION_TYPE_CREATE_INVESTOR_PAYMENT_STATEMENT,
		JobName:        jobNameCreateInvestorPaymentStatement,
		RequestPayload: jobExecutionRequestPayload(request),
	}, createInvestorPaymentStatementJob)
}

func createInvestorPaymentStatementJob(ctx context.Context) error {
	rdi, err := ReportDateInfo{}.FromContext(ctx).getDateInfo()
	if err != nil {
		return err
	}

	logrus.WithFields(logrus.Fields{
		"category": "create report",
		"time":     time.Now(),
	}).Info("CreateInvestorPaymentStatementJob Jobs Start")

	type Item struct {
		UserId int64 `json:"user_id"`
		SpvId  int64 `json:"spv_id"`
	}

	type PaymentRequest struct {
		AccountId int64 `json:"account_id"`
	}

	var items []Item
	query := factory.DB(ctx).Table("dividend").Select("dividend.user_id as user_id,dividend.spv_id as spv_id").GroupBy("dividend.user_id").
		Where("dividend.transaction_type = 'DEPOSIT_DIVIDEND'").And("dividend.dividend_date >= ?", rdi.Year+"-01-01").And("dividend.dividend_date <= ?", rdi.Year+"-12-31")
	if err := query.Find(&items); err != nil {
		return err
	}

	for _, item := range items {

		dividends, err := domain.Dividend{}.GetAllByUserId(ctx, item.UserId, rdi.Year+"-01-01", rdi.Year+"-12-31")
		if err != nil {
			return errors.Wrap(err, "fail get spv for payment statement")
		}
		spvs := map[int64]*domain.Spv{}
		for _, dividend := range dividends {
			if _, ok := spvs[dividend.SpvId]; !ok {
				if dividend.SpvId == 0 {
					continue
				}
				spv, err := domain.Spv{}.GetById(ctx, dividend.SpvId)
				if err != nil {
					return errors.Wrap(err, "fail get spv for payment statement")
				}
				if spv == nil {
					return errors.Wrap(err, "fail get spv for payment statement")
				}
				spvs[dividend.SpvId] = spv

			}
		}

		for _, spv := range spvs {
			ro := &domain.ReportOutput{
				UserId:         item.UserId,
				SpvId:          0,
				IsOpen:         false,
				SubmissionType: domain.SUBMISSION_INVESTORS,
				ReportType:     domain.REPORT_PAYMENT_STATEMENT,
				Year:           rdi.Year,
				Quarter:        "",
			}
			account, err := domain.User{}.GetById(ctx, item.UserId)
			if err != nil {
				return errors.Wrap(err, "fail get user for investor payment statement")
			}
			_, err = ro.CreateInvestorPaymentStatement(ctx, account, spv)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"category": "create investor payment statement",
					"user_id":  item.UserId,
					"year":     rdi.Year,
					"err":      err,
				}).Warning("not created investor payment statement")
			}
		}

	}

	adminId, err := domain.AdminUser{}.GetSuperAdminId(ctx)
	if err != nil {
		return err
	}
	user, err := domain.User{}.GetById(ctx, adminId)
	if err != nil {
		return err
	}
	history := domain.OperatorHistory{
		UserId: user.Id,
		Email:  user.Email,
		Menu:   i18nutils.DefaultTranslate("commoncode.LOG_CONTENT_TYPE.PAYMENT_STATEMENT_BATCH_EXECUTE"),
		Ip:     "::1",
		State:  domain.OPERATOR_STATE("CHANGE"),
	}
	_, err = history.Create(ctx)
	if err != nil {
		return err
	}

	return nil
}

// 세무서용 지불조서 생성
func (ReportJobs) CreateTaxOfficePaymentStatementJob(ctx context.Context) error {
	request := ReportDateInfo{}.FromContext(ctx)
	return runJobWithExecutionHistory(ctx, JobExecutionHistoryInput{
		JobType:        domain.JOB_EXECUTION_TYPE_CREATE_TAX_OFFICE_PAYMENT_STATEMENT,
		JobName:        jobNameCreateTaxOfficePaymentStatement,
		RequestPayload: jobExecutionRequestPayload(request),
	}, createTaxOfficePaymentStatementJob)
}

func createTaxOfficePaymentStatementJob(ctx context.Context) error {
	rdi, err := ReportDateInfo{}.FromContext(ctx).getDateInfo()
	if err != nil {
		return err
	}

	logrus.WithFields(logrus.Fields{
		"category": "create report",
		"time":     time.Now(),
	}).Info("CreateTaxOfficePaymentStatementJob Jobs Start")

	type Item struct {
		UserId int64 `json:"user_id"`
		SpvId  int64 `json:"spv_id"`
	}

	var items []Item
	query := factory.DB(ctx).Table("dividend").Select("dividend.user_id as user_id, dividend.spv_id as spv_id").GroupBy("dividend.spv_id").
		Where("dividend.transaction_type = 'DEPOSIT_DIVIDEND'").And("dividend.dividend_date >= ?", rdi.Year+"-01-01").And("dividend.dividend_date <= ?", rdi.Year+"-12-31")
	if err := query.Find(&items); err != nil {
		return errors.Wrap(err, "fail get spv for payment statement")
	}

	for _, item := range items {
		dividends, err := domain.Dividend{}.GetAllBySpvId(ctx, item.SpvId, rdi.Year+"-01-01", rdi.Year+"-12-31")
		if err != nil {
			return errors.Wrap(err, "fail get spv for tax office payment statement")
		}

		accounts := map[int64]*domain.User{}
		for _, dividend := range dividends {
			if _, ok := accounts[dividend.UserId]; !ok {
				accounts[dividend.UserId], err = domain.User{}.GetById(ctx, dividend.UserId)
				if err != nil {
					return errors.Wrap(err, "fail get spv for tax office payment statement")
				}
			}
		}
		spv, err := domain.Spv{}.GetById(ctx, item.SpvId)
		if err != nil {
			return errors.Wrap(err, "fail get spv for tax office payment statement")
		}
		if spv == nil {
			return errors.Wrap(err, "fail get spv for tax office payment statement")
		}

		ro := &domain.ReportOutput{
			UserId:         0,
			SpvId:          item.SpvId,
			IsOpen:         false,
			SubmissionType: domain.SUBMISSION_TAX_OFFICE,
			ReportType:     domain.REPORT_PAYMENT_STATEMENT,
			Year:           rdi.Year,
			Quarter:        "",
		}

		_, err = ro.CreateTaxOfficePaymentStatement(ctx, spv, accounts)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"category": "create tax office payment statement",
				"user_id":  item.UserId,
				"year":     rdi.Year,
				"err":      err,
			}).Warning("not created tax office investor payment statement")
		}
	}

	userId, err := domain.AdminUser{}.GetSuperAdminId(ctx)
	if err != nil {
		return err
	}
	user, err := domain.User{}.GetById(ctx, userId)
	if err != nil {
		return err
	}
	history := domain.OperatorHistory{
		UserId: user.Id,
		Email:  user.Email,
		Menu:   "commoncode.LOG_CONTENT_TYPE.PAYMENT_STATEMENT_BATCH_EXECUTE",
		Ip:     "::1",
		State:  domain.OPERATOR_STATE("CHANGE"),
	}
	_, err = history.Create(ctx)
	if err != nil {
		return err
	}

	return nil
}

// 세무서용 지불조서 합계표 생성
func (ReportJobs) CreateTaxOfficePaymentStatementSumJob(ctx context.Context) error {
	request := ReportDateInfo{}.FromContext(ctx)
	return runJobWithExecutionHistory(ctx, JobExecutionHistoryInput{
		JobType:        domain.JOB_EXECUTION_TYPE_CREATE_TAX_OFFICE_PAYMENT_STATEMENT_SUM,
		JobName:        jobNameCreateTaxOfficePaymentStatementSum,
		RequestPayload: jobExecutionRequestPayload(request),
	}, createTaxOfficePaymentStatementSumJob)
}

func createTaxOfficePaymentStatementSumJob(ctx context.Context) error {
	rdi, err := ReportDateInfo{}.FromContext(ctx).getDateInfo()
	if err != nil {
		return err
	}

	logrus.WithFields(logrus.Fields{
		"category": "create report",
		"time":     time.Now(),
	}).Info("CreateTaxOfficePaymentStatementSumJob Jobs Start")

	type Item struct {
		UserId int64 `json:"user_id"`
		SpvId  int64 `json:"spv_id"`
	}

	var items []Item
	query := factory.DB(ctx).Table("dividend").Select("dividend.spv_id as spv_id").GroupBy("dividend.spv_id").
		Where("dividend.transaction_type = 'DEPOSIT_DIVIDEND'").And("dividend.dividend_date >= ?", rdi.Year+"-01-01").And("dividend.dividend_date <= ?", rdi.Year+"-12-31")
	if err := query.Find(&items); err != nil {
		return errors.Wrap(err, "fail get spv for tax office payment statement sum ")
	}

	for _, item := range items {
		if item.SpvId == 0 {
			continue
		}
		spv, err := domain.Spv{}.GetById(ctx, item.SpvId)
		if err != nil {
			return errors.Wrap(err, "fail get spv for tax office payment statement sum ")
		}
		if spv == nil {
			return errors.Wrap(err, "fail get spv for tax office payment statement sum ")
		}

		ro := &domain.ReportOutput{
			UserId:         0,
			SpvId:          item.SpvId,
			IsOpen:         false,
			SubmissionType: domain.SUBMISSION_TAX_OFFICE,
			ReportType:     domain.REPORT_PAYMENT_STATEMENT_SUM,
			Year:           rdi.Year,
			Quarter:        "",
		}

		_, err = ro.CreateTaxOfficePaymentStatementSum(ctx, spv)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"category": "create tax office payment statement",
				"spv_id":   item.SpvId,
				"year":     rdi.Year,
				"spv_name": spv.Name,
				"err":      err,
			}).Warning("not created tax office investor payment statement")
			continue
		}
	}
	adminId, err := domain.AdminUser{}.GetSuperAdminId(ctx)
	if err != nil {
		return err
	}
	user, err := domain.User{}.GetById(ctx, adminId)
	if err != nil {
		return err
	}
	history := domain.OperatorHistory{
		UserId: user.Id,
		Email:  user.Email,
		Menu:   "commoncode.LOG_CONTENT_TYPE.PAYMENT_STATEMENT_SUM_BATCH_EXECUTE",
		Ip:     "::1",
		State:  domain.OPERATOR_STATE("CHANGE"),
	}
	_, err = history.Create(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (ReportJobs) findInvestorReportOutput(ctx context.Context, reportType, quarter, year string, userId, spvId int64) (bool, error) {

	q := factory.DB(ctx).Table("report_output").Select("report_output.*")

	if len(quarter) > 0 {
		q.Where("submission_type=? AND report_type=? AND year=? AND quarter=? AND user_id=?", "SUBMISSION_INVESTORS", reportType, year, quarter, userId)
	} else {
		q.Where("submission_type=? AND report_type=? AND year=? AND user_id=?", "SUBMISSION_INVESTORS", reportType, year, userId)
	}

	if spvId > 0 {
		q.Where("spv_id=?", spvId)
	}

	items := make([]domain.ReportOutput, 0, 0)
	totalCount, err := q.FindAndCount(&items)
	if err != nil {
		return false, err
	}

	if totalCount == 0 {
		return false, nil
	}

	return true, nil
}

func (ReportJobs) findTaxOfficeReportOutput(ctx context.Context, reportType, year string, userId, spvId int64) (bool, error) {

	q := factory.DB(ctx).Table("report_output").Select("report_output.*")

	if userId > 0 {
		q.Where("submission_type=? AND report_type=? AND year=? AND user_id=? AND spv_id=?", "SUBMISSION_TAX_OFFICE", reportType, year, userId, spvId)
	} else {
		q.Where("submission_type=? AND report_type=? AND year=? AND spv_id=?", "SUBMISSION_TAX_OFFICE", reportType, year, spvId)
	}

	items := make([]domain.ReportOutput, 0, 0)
	totalCount, err := q.FindAndCount(&items)
	if err != nil {
		return false, err
	}

	if totalCount == 0 {
		return false, nil
	}

	return true, nil
}

func (ReportJobs) findAuditResponseReportOutput(ctx context.Context, reportType, year, quarter string) (bool, error) {

	q := factory.DB(ctx).Table("report_output").Select("report_output.*")

	q.Where("submission_type=? AND report_type=? AND year=? AND quarter=?", "SUBMISSION_AUDIT_RESPONSE", reportType, year, quarter)

	items := make([]domain.ReportOutput, 0, 0)
	totalCount, err := q.FindAndCount(&items)
	if err != nil {
		return false, err
	}

	if totalCount == 0 {
		return false, nil
	}

	return true, nil
}
