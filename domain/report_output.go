package domain

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/commoncode"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/modelhelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/s3utils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
)

type QUARTER_TYPE string

const (
	QUARTER_TYPE_CODE              = "QUARTER_TYPE"
	QUARTER_FIRST     QUARTER_TYPE = "QUARTER_FIRST"
	QUARTER_SECOND    QUARTER_TYPE = "QUARTER_SECOND"
	QUARTER_THIRD     QUARTER_TYPE = "QUARTER_THIRD"
	QUARTER_FOURTH    QUARTER_TYPE = "QUARTER_FOURTH"
)

func (q QUARTER_TYPE) Quarter() int {
	switch q {
	case QUARTER_FIRST:
		return 1
	case QUARTER_SECOND:
		return 2
	case QUARTER_THIRD:
		return 3
	case QUARTER_FOURTH:
		return 4

	}
	return 0
}

func searchErr(errCode string, messages ...string) error {
	message := "report system error"
	if len(messages) > 0 && strings.TrimSpace(messages[0]) != "" {
		message = messages[0]
	}
	switch errCode {
	case "s001":
		return factory.ErrorNotFoundReport.New(errors.New(errCode + ":" + message))
	case "s002":
		return &ReportServiceError{ErrorCode: errCode, Message: message}
	case "s003":
		return &ReportServiceError{ErrorCode: errCode, Message: message}
	default:
		return &ReportServiceError{ErrorCode: errCode, Message: message}
	}
}

type Response struct {
	Result    string `json:"result,omitempty"`
	Success   bool   `json:"success,omitempty"`
	Error     string `json:"error,omitempty"`
	ErrorCode string `json:"error_code,omitempty"`
}

var reportServiceHTTPClient = http.Client{
	Timeout: 20 * time.Second,
}

var reportServiceBulkHTTPClient = http.Client{
	Timeout: 300 * time.Second,
}

type ReportServiceError struct {
	ErrorCode string
	Message   string
}

func (e *ReportServiceError) Error() string {
	if e == nil {
		return ""
	}
	if e.Message == "" {
		return e.ErrorCode + ":report service error"
	}
	return e.ErrorCode + ":" + e.Message
}

func (e *ReportServiceError) IsPDFError() bool {
	return e != nil && e.ErrorCode == "s003"
}

func (e *ReportServiceError) IsS3Error() bool {
	return e != nil && e.ErrorCode == "s002"
}

func addressToString[ADDR2 any](city, province, address1 string, address2 ADDR2) string {
	switch val := any(address2).(type) {
	case string:
		if len(val) == 0 {
			return fmt.Sprintf("%s %s %s", city, province, address1)
		}
		return fmt.Sprintf("%s %s %s %s", city, province, address1, val)
	case *string:
		if val == nil {
			return fmt.Sprintf("%s %s %s", city, province, address1)
		}
		return fmt.Sprintf("%s %s %s %s", city, province, address1, *val)
	default:
		return fmt.Sprintf("%s %s %s", city, province, address1)
	}
}

func (ReportOutput) GetQuarterTypeCodeGroup(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(QUARTER_FIRST), Name: i18nutils.Translate("commoncode.QUATER_TYPE.QUARTER_FIRST", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(QUARTER_SECOND), Name: i18nutils.Translate("commoncode.QUATER_TYPE.QUARTER_SECOND", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(QUARTER_THIRD), Name: i18nutils.Translate("commoncode.QUATER_TYPE.QUARTER_THIRD", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(QUARTER_FOURTH), Name: i18nutils.Translate("commoncode.QUATER_TYPE.QUARTER_FOURTH", loc), Order: len(codes) + 1})

	return commoncode.CodeGroup{
		GroupCode: QUARTER_TYPE_CODE,
		Codes:     codes,
	}
}

type REPORT_TYPE string

const (
	REPORT_TYPE_CODE                                    = "REPORT_TYPE"
	REPORT_TRANSACTION_BALANCE_REPORT       REPORT_TYPE = "REPORT_TRANSACTION_BALANCE_REPORT"       // 거래잔액 보고서
	REPORT_TRANSACTION_BALANCE_ADMIN_REPORT REPORT_TYPE = "REPORT_TRANSACTION_BALANCE_ADMIN_REPORT" // 거래잔액 관리자 보고서
	REPORT_PAYMENT_STATEMENT                REPORT_TYPE = "REPORT_PAYMENT_STATEMENT"                // 지불조서
	REPORT_PAYMENT_STATEMENT_SUM            REPORT_TYPE = "REPORT_PAYMENT_STATEMENT_SUM"            // 지불조서 합계표
	REPORT_TRADING_DIARY                    REPORT_TYPE = "REPORT_TRADING_DIARY"                    // 거래일기장
	REPORT_TRANSACTION_RECORD               REPORT_TYPE = "REPORT_TRANSACTION_RECORD"               // 모집 취급에 관한 거래 기록 (사모공모 거래 기록)
	REPORT_CUSTOMER_ACCOUNT                 REPORT_TYPE = "REPORT_CUSTOMER_ACCOUNT"                 // 고객계정 원장
)

func (ReportOutput) GetReportTypeCodeGroup(loc string) commoncode.CodeGroup {
	codes := make([]commoncode.Code, 0, 0)
	codes = append(codes, commoncode.Code{Code: string(REPORT_TRANSACTION_BALANCE_REPORT), Name: i18nutils.Translate("commoncode.REPORT_TYPE.REPORT_TRANSACTION_BALANCE_REPORT", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(REPORT_TRANSACTION_BALANCE_ADMIN_REPORT), Name: i18nutils.Translate("commoncode.REPORT_TYPE.REPORT_TRANSACTION_BALANCE_ADMIN_REPORT", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(REPORT_PAYMENT_STATEMENT), Name: i18nutils.Translate("commoncode.REPORT_TYPE.REPORT_PAYMENT_STATEMENT", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(REPORT_PAYMENT_STATEMENT_SUM), Name: i18nutils.Translate("commoncode.REPORT_TYPE.REPORT_PAYMENT_STATEMENT_SUM", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(REPORT_TRADING_DIARY), Name: i18nutils.Translate("commoncode.REPORT_TYPE.REPORT_TRADING_DIARY", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(REPORT_TRANSACTION_RECORD), Name: i18nutils.Translate("commoncode.REPORT_TYPE.REPORT_TRANSACTION_RECORD", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(REPORT_CUSTOMER_ACCOUNT), Name: i18nutils.Translate("commoncode.REPORT_TYPE.REPORT_CUSTOMER_ACCOUNT", loc), Order: len(codes) + 1})

	return commoncode.CodeGroup{
		GroupCode: REPORT_TYPE_CODE,
		Codes:     codes,
	}
}

type SUBMISSION_TYPE string

const (
	SUBMISSION_INVESTORS      SUBMISSION_TYPE = "SUBMISSION_INVESTORS"      // 투자자용
	SUBMISSION_TAX_OFFICE     SUBMISSION_TYPE = "SUBMISSION_TAX_OFFICE"     // 세무서용
	SUBMISSION_AUDIT_RESPONSE SUBMISSION_TYPE = "SUBMISSION_AUDIT_RESPONSE" // 감사대응
)

type ReportOutput struct {
	Id             int64              `json:"id"`
	UserId         int64              `json:"user_id"`
	SpvId          int64              `json:"spv_id"`
	IsOpen         bool               `json:"is_open"`
	Title          string             `json:"title" xorm:"not null"`
	SubmissionType SUBMISSION_TYPE    `json:"submission_type" xorm:"not null"`
	ReportType     REPORT_TYPE        `json:"report_type" xorm:"not null"`
	Year           string             `json:"year" xorm:"not null"`
	Quarter        QUARTER_TYPE       `json:"quarter" xorm:"not null"`
	FileUrl        string             `json:"file_url"`
	AtOpen         jsonutils.JSONTime `json:"at_open" xorm:"at_open"`
	Created        jsonutils.JSONTime `json:"created" xorm:"created"`
	Updated        jsonutils.JSONTime `json:"updated" xorm:"updated"`
	Deleted        jsonutils.JSONTime `json:"-" xorm:"deleted"`
}

func (ReportOutput) TableName() string {
	return "report_output"
}

func (r *ReportOutput) Open() {
	r.IsOpen = true
	r.AtOpen = jsonutils.JSONTime(time.Now())
}

// isNewerReportOutputはcandidateがcurrentより新しいかを判定する
// Createdが異なれば新しい方を優先し、同一の場合はIdが大きい方を新しいとみなす
func isNewerReportOutput(current, candidate ReportOutput) bool {
	currentCreated := time.Time(current.Created)
	candidateCreated := time.Time(candidate.Created)
	if !candidateCreated.Equal(currentCreated) {
		return candidateCreated.After(currentCreated)
	}

	return candidate.Id > current.Id
}

// latestReportOutputIDsByUserはReportOutputのスライスからユーザーごとに最新（Created→Id順）のReportOutput.Idを返す
func latestReportOutputIDsByUser(items []ReportOutput) map[int64]int64 {
	latestByUser := make(map[int64]ReportOutput, len(items))
	for _, item := range items {
		current, exists := latestByUser[item.UserId]
		if !exists || isNewerReportOutput(current, item) {
			latestByUser[item.UserId] = item
		}
	}

	latestIDs := make(map[int64]int64, len(latestByUser))
	for userID, item := range latestByUser {
		latestIDs[userID] = item.Id
	}

	return latestIDs
}

func (ReportOutput) GetQuarterTransactionBalanceAdminReportList(ctx context.Context, sortby, order string) (int64, []ReportOutput, error) {
	q := factory.DB(ctx).Table("report_output").Select("report_output.*")

	if err := modelhelper.SetSortOrder(q, sortby, order); err != nil {
		factory.Logger(ctx).Error(err)
	}

	items := make([]ReportOutput, 0, 0)
	var totalCount int64
	totalCount, err := q.Where("submission_type = 'SUBMISSION_INVESTORS' AND report_type = ?", REPORT_TRANSACTION_BALANCE_ADMIN_REPORT).GroupBy("year,quarter").
		FindAndCount(&items)
	if err != nil {
		return 0, nil, errors.Wrap(err, "failed to get report output list")
	}

	for index := range items {
		items[index].FileUrl = s3utils.GetUrlSign(items[index].FileUrl)
	}

	return totalCount, items, nil
}

func (ReportOutput) GetAllInvestorPaymentStatementList(ctx context.Context, sortby, order string, offset, limit int) (int64, []ReportOutput, error) {
	q := factory.DB(ctx).Table("report_output").Select("report_output.*")

	if err := modelhelper.SetSortOrder(q, sortby, order); err != nil {
		factory.Logger(ctx).Error(err)
	}

	items := make([]ReportOutput, 0, 0)
	var totalCount int64
	totalCount, err := q.Where("submission_type = 'SUBMISSION_INVESTORS' AND report_type = ?", REPORT_PAYMENT_STATEMENT).GroupBy("year").
		FindAndCount(&items)
	if err != nil {
		return 0, nil, err
	}

	for index := range items {
		items[index].FileUrl = s3utils.GetUrlSign(items[index].FileUrl)
	}

	return totalCount, items, nil
}

// 개발회사 별로 지불조서가 존재하는 연도를 제공
func (ReportOutput) GetAllTaxOfficeReportGroupByYear(ctx context.Context, spvId int64) (int64, []string, error) {
	q := factory.DB(ctx).Table("report_output").Select("report_output.year")

	if err := modelhelper.SetSortOrder(q, "year", "desc"); err != nil {
		factory.Logger(ctx).Error(err)
	}

	items := make([]string, 0)
	var totalCount int64
	totalCount, err := q.Where("submission_type = 'SUBMISSION_TAX_OFFICE'").And("spv_id = ?", spvId).GroupBy("year").
		FindAndCount(&items)
	if err != nil {
		return 0, nil, err
	}

	return totalCount, items, nil
}

// OpenTransactionBalanceReportは指定されたyear/quarterの取引残高報告書を公開する
// 全件を一度is_open=falseにした上で、ユーザーごとに最新の1件だけをis_open=trueにする
func (ReportOutput) OpenTransactionBalanceReport(ctx context.Context, year, quarter string) (bool, error) {
	q := factory.DB(ctx).Table("report_output").Select("report_output.*")

	q.Where("submission_type=? AND (report_type=? or report_type=?) AND year=? AND quarter=?",
		SUBMISSION_INVESTORS, REPORT_TRANSACTION_BALANCE_REPORT, REPORT_TRANSACTION_BALANCE_ADMIN_REPORT, year, quarter)
	items := make([]ReportOutput, 0)
	_, err := q.FindAndCount(&items)
	if err != nil {
		return false, err
	}

	if len(items) == 0 {
		factory.Logger(ctx).WithFields(logrus.Fields{
			"year":    year,
			"quarter": quarter,
		}).Warn("no transaction balance reports found for the specified period")
		return false, nil
	}

	latestIDsByUser := latestReportOutputIDsByUser(items)

	for i := range items {
		isLatest := latestIDsByUser[items[i].UserId] == items[i].Id
		if isLatest {
			items[i].Open()
		} else {
			items[i].IsOpen = false
		}

		if err := items[i].Update(ctx); err != nil {
			factory.Logger(ctx).WithFields(logrus.Fields{
				"report_output_id": items[i].Id,
				"error":            err,
			}).Error("fail open report output")
			return false, errors.Wrap(err, "fail open report")
		}
	}

	return false, nil
}

// OpenPaymentStatementReportは指定されたyearの支払計算書を公開する
// 未公開の全件をis_open=trueにする
func (ReportOutput) OpenPaymentStatementReport(ctx context.Context, year string) (bool, error) {
	q := factory.DB(ctx).Table("report_output").Select("report_output.*")

	q.Where("submission_type=? AND report_type=? AND year=?", SUBMISSION_INVESTORS, REPORT_PAYMENT_STATEMENT, year)
	items := make([]ReportOutput, 0)
	_, err := q.FindAndCount(&items)
	if err != nil {
		return false, err
	}

	if len(items) == 0 {
		factory.Logger(ctx).WithFields(logrus.Fields{
			"year": year,
		}).Warn("no payment statement reports found for the specified year")
		return false, nil
	}

	alreadyAllOpen := true
	for i := range items {
		if !items[i].IsOpen {
			alreadyAllOpen = false
			break
		}
	}
	if alreadyAllOpen {
		return true, errors.New("already all report opened")
	}

	for i := range items {
		if !items[i].IsOpen {
			items[i].Open()
			if err := items[i].Update(ctx); err != nil {
				factory.Logger(ctx).WithFields(logrus.Fields{
					"report_output_id": items[i].Id,
					"error":            err,
				}).Error("fail open report output")
				return false, errors.Wrap(err, "fail open report")
			}
		}
	}

	return false, nil
}

func (ReportOutput) FindAdminInvestorReportByReportType(ctx context.Context, reportType REPORT_TYPE, year, quarter string) (*ReportOutput, error) {
	q := factory.DB(ctx).Table("report_output").Select("report_output.*")

	if len(quarter) > 0 {
		q.Where("submission_type=? AND report_type=? AND year=? AND quarter=?", SUBMISSION_INVESTORS, reportType, year, quarter)
	} else {
		q.Where("submission_type=? AND report_type=? AND year=?", SUBMISSION_INVESTORS, reportType, year)
	}

	items := make([]ReportOutput, 0, 0)
	_, err := q.FindAndCount(&items)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, nil
	}
	for index := range items {
		items[index].FileUrl = s3utils.GetUrlSign(items[index].FileUrl)
	}

	item := items[len(items)-1]

	return &item, nil
}

func (ReportOutput) FindAdminTaxOfficeReportByReportType(ctx context.Context, reportType REPORT_TYPE, spvId int64, year string) (*ReportOutput, error) {
	q := factory.DB(ctx).Table("report_output").Select("report_output.*")

	q.Where("submission_type=? AND spv_id = ? AND report_type=? AND year=?", SUBMISSION_TAX_OFFICE, spvId, reportType, year)

	items := make([]ReportOutput, 0, 0)
	_, err := q.FindAndCount(&items)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, nil
	}
	for index := range items {
		items[index].FileUrl = s3utils.GetUrlSign(items[index].FileUrl)
	}

	item := items[len(items)-1]

	return &item, nil
}

func (ReportOutput) GetById(ctx context.Context, id int64) (*ReportOutput, error) {
	var reportOutput ReportOutput
	has, err := factory.DB(ctx).ID(id).Get(&reportOutput)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}

	reportOutput.FileUrl = s3utils.GetUrlSign(reportOutput.FileUrl)

	return &reportOutput, nil
}

func (ro *ReportOutput) Update(ctx context.Context) error {
	_, err := factory.DB(ctx).UseBool("is_open").ID(ro.Id).Update(ro)
	return err
}

func (ReportOutput) Delete(ctx context.Context, id int64) (err error) {
	_, err = factory.DB(ctx).ID(id).Delete(&ReportOutput{})
	return
}

func (ro *ReportOutput) CreateTitleTransactionBalanceReport() {
	ro.Title = ro.Year + i18nutils.DefaultTranslate("pdf.report.year")
	ro.Title += " " + i18nutils.DefaultTranslate("commoncode.QUATER_TYPE."+string(ro.Quarter))
	ro.Title += " " + i18nutils.DefaultTranslate("commoncode.REPORT_TYPE."+string(ro.ReportType))
}

func (ro *ReportOutput) CreateTitleTransactionBalanceAdminReport() {
	ro.Title = ro.Year + i18nutils.DefaultTranslate("pdf.report.year")
	ro.Title += " " + i18nutils.DefaultTranslate("commoncode.REPORT_TYPE."+string(ro.ReportType))
}

func (ro *ReportOutput) CreateTitleInvestorPaymentStatment(spvName string) {
	ro.Title = spvName
	ro.Title += " " + ro.Year + i18nutils.DefaultTranslate("pdf.report.year")
	ro.Title += " " + i18nutils.DefaultTranslate("commoncode.REPORT_TYPE."+string(ro.ReportType))
}

func (ro *ReportOutput) CreateTitleTaxOfficePaymentStatment() {
	ro.Title = ro.Year + i18nutils.DefaultTranslate("pdf.report.year")
	ro.Title += " " + i18nutils.DefaultTranslate("commoncode.REPORT_TYPE."+string(ro.ReportType))
}

func (ReportOutput) GetByUserReport(ctx context.Context, userId int64, reportsType REPORT_TYPE, sortby string, order string, offset int, limit int) (int64, []ReportOutput, error) {
	q := factory.DB(ctx).Table("report_output").Select("report_output.*").
		Where("submission_type=? AND report_type=? AND user_id=? AND is_open=?", SUBMISSION_INVESTORS, reportsType, userId, true)

	if err := modelhelper.SetSortOrder(q, sortby, order); err != nil {
		factory.Logger(ctx).Error(err)
	}

	items := make([]ReportOutput, 0, 0)
	var totalCount int64
	totalCount, err := q.Limit(limit, offset).FindAndCount(&items)
	if err != nil {
		return 0, nil, err
	}

	for index := range items {
		items[index].FileUrl = s3utils.GetUrlSign(items[index].FileUrl)
	}

	return totalCount, items, nil
}

func (ro *ReportOutput) CreateTransactionBalanceReport(ctx context.Context) (int64, error) {
	url, err := ro.RequestTransactionBalanceReport(ctx, ro.UserId, ro.Year, ro.Quarter)
	if err != nil {
		var notFoundErr *factory.Error
		if errors.As(err, &notFoundErr) && notFoundErr.Code == factory.ErrorNotFoundReport.Code {
			return 0, err
		}
	}
	ro.CreateTitleTransactionBalanceReport()
	ro.FileUrl = url
	if err != nil {
		return 0, errors.Wrap(err, "fail request transaction balance report")
	}
	return factory.DB(ctx).Insert(ro)
}

func (ro *ReportOutput) CreateTransactionBalanceReportAll(ctx context.Context) (int64, error) {
	url, err := ro.RequestTransactionBalanceReportAll(ctx, ro.UserId, ro.Year, ro.Quarter)
	if err != nil {
		var notFoundErr *factory.Error
		if errors.As(err, &notFoundErr) && notFoundErr.Code == factory.ErrorNotFoundReport.Code {
			return 0, err
		}
	}
	ro.CreateTitleTransactionBalanceReport()
	ro.FileUrl = url
	if err != nil {
		return 0, errors.Wrap(err, "fail request transaction balance report")
	}
	return factory.DB(ctx).Insert(ro)
}

func (ReportOutput) RequestTransactionBalanceReport(ctx context.Context, userId int64, year string, quarter QUARTER_TYPE) (string, error) {
	const URL = "/api/download-quarterly-balance-report"
	u, err := User{}.GetById(ctx, userId)
	if err != nil {
		return "", errors.Wrap(err, "fail get user for transaction balance report")
	}

	params := url.Values{}
	params.Add("user_id", fmt.Sprint(userId))
	params.Add("user_name", u.Name)
	params.Add("year", fmt.Sprint(year))
	params.Add("quarter", fmt.Sprint(quarter.Quarter()))
	params.Add("recreate", "true")

	resp, err := RequestReportService(URL, params)
	if err != nil {
		return "", errors.Wrap(err, "fail request transaction balance report")
	}

	if !resp.Success {
		return "", searchErr(resp.ErrorCode, resp.Error)
	}
	return resp.Result, nil
}

func (ReportOutput) RequestTransactionBalanceReportAll(ctx context.Context, userId int64, year string, quarter QUARTER_TYPE) (string, error) {
	const URL = "/api/download-quarterly-balance-report-all"

	params := url.Values{}
	params.Add("year", fmt.Sprint(year))
	params.Add("quarter", fmt.Sprint(quarter.Quarter()))
	params.Add("recreate", "true")

	resp, err := RequestReportServiceBulk(URL, params)
	if err != nil {
		return "", errors.Wrap(err, "fail request transaction balance report")
	}

	if !resp.Success {
		return "", searchErr(resp.ErrorCode, resp.Error)
	}
	return resp.Result, nil
}

func RequestReportService(URL string, params url.Values) (*Response, error) {
	return doRequestReportService(reportServiceHTTPClient, URL, params)
}

func RequestReportServiceBulk(URL string, params url.Values) (*Response, error) {
	return doRequestReportService(reportServiceBulkHTTPClient, URL, params)
}

func doRequestReportService(client http.Client, URL string, params url.Values) (*Response, error) {
	var respBody Response
	resp, err := client.Get(serviceutils.GetStatutoryBooksServiceUrl() + URL + "?" + params.Encode())
	if err != nil {
		return nil, errors.Wrap(err, "fail http request")
	}
	defer resp.Body.Close()

	buff, err2 := io.ReadAll(resp.Body)
	if err2 != nil {
		return nil, errors.Wrap(err2, "fail read http transaction balance report response body")
	}

	if resp.StatusCode != http.StatusOK {
		if err := json.Unmarshal(buff, &respBody); err != nil {
			return nil, errors.Errorf("fail http transaction balance report status=%d body=%s", resp.StatusCode, string(buff))
		}
		return nil, errors.Errorf("fail http transaction balance report status=%d error_code=%s error=%s", resp.StatusCode, respBody.ErrorCode, respBody.Error)
	}

	if err := json.Unmarshal(buff, &respBody); err != nil {
		return nil, errors.Wrap(err, "fail http transaction balance report Unmarshal")
	}

	logrus.WithFields(logrus.Fields{
		"url":      URL,
		"response": respBody,
		"params":   params,
		"time":     time.Now(),
	}).Info("report service response")

	return &respBody, nil
}

func (ro *ReportOutput) CreateInvestorPaymentStatement(ctx context.Context, u *User, spv *Spv) (int64, error) {
	ro.CreateTitleInvestorPaymentStatment(spv.Name)
	var err error
	ro.FileUrl, err = RequestInvestorPaymentStatement(ctx, ro, u, spv)
	if err != nil {
		return 0, errors.Wrap(err, "fail request payment statement")
	}
	return factory.DB(ctx).Insert(ro)
}

// 투자자의 경우 마이넘버, 법인의 경우 법인번호를 가져온다.
func GetIdentificationNumber(ctx context.Context, u *User) (string, error) {
	var identificationNumber string
	if u.UserType == PRIVATE {
		m, err := MyNumber{}.GetByUserId(ctx, u.Id)
		if err != nil {
			return "", errors.Wrap(err, "fail get my number for payment statement")
		}
		if m == nil {
			identificationNumber = ""
			return "", errors.Wrap(err, "fail get my number for payment statement")
		}
		identificationNumber = m.MyNumber
	} else if u.UserType == CORPORATE {
		corpUser, err := CorporateUser{}.GetById(ctx, u.Id)
		if err != nil {
			return "", errors.Wrap(err, "fail get corporate user for payment statement")
		}
		identificationNumber = corpUser.CorpNo
	} else {
		identificationNumber = ""
	}
	return identificationNumber, nil
}

type ReportUserInput struct {
	UserId               int64  `json:"user_id"`
	UserAddress          string `json:"user_address"`
	UserName             string `json:"user_name"`
	IdentificationNumber string `json:"identification_number" example:"투자자인 경우 마이넘버, 법인인 경우 법인번호, 빈문자열도 받음"`
}

type SpvInput struct {
	SpvId      int64  `json:"spv_id"`
	SpvNo      string `json:"spv_no" example:"SPV 법인번호"`
	SpvAddress string `json:"spv_address"`
	SpvName    string `json:"spv_name"`
	SpvPhone   string `json:"spv_phone"`
}
type UserPaymentStatementInput struct {
	Year int64 `json:"year"`
	ReportUserInput
	IdentificationNumber string     `json:"identification_number" example:"투자자인 경우 마이넘버, 법인인 경우 법인번호, 빈문자열도 받음"`
	SpvList              []SpvInput `json:"spv_list"`
	Recreate             bool       `json:"recreate"`
}

// recreate

// RequestInvestorPaymentStatement  지불조서 요청
func RequestInvestorPaymentStatement(ctx context.Context, ro *ReportOutput, u *User, spv *Spv) (string, error) {
	var dto UserPaymentStatementInput
	var respBody []byte
	var err error

	if u == nil {
		return "", errors.New("user model is nil")
	}

	if ro == nil {
		return "", errors.New("report model is nil")
	}
	const URL = "/api/download-user-payment-statement"

	year, err := strconv.ParseInt(ro.Year, 10, 64)
	if err != nil {
		return "", errors.Wrap(err, "fail get year for payment statement")
	}

	address2 := ""
	if u.Address2 != nil {
		address2 = *u.Address2
	}

	dto.Year = year
	dto.UserId = u.Id
	dto.UserAddress = addressToString(u.City, u.Province, u.Address, address2)
	dto.UserName = u.Name
	dto.SpvList = make([]SpvInput, 0)
	dto.SpvList = append(dto.SpvList, SpvInput{
		SpvId:      spv.Id,
		SpvNo:      spv.CorpNo,
		SpvAddress: addressToString(spv.City, spv.Province, spv.Address, spv.Address2),
		SpvName:    spv.Name,
		SpvPhone:   PhoneNumberSplit(spv.Phone),
	})
	dto.Recreate = true

	pbytes, _ := json.Marshal(dto)
	body := bytes.NewBuffer(pbytes)
	resp, err := http.Post(serviceutils.GetStatutoryBooksServiceUrl()+URL, "application/json", body)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"category": "create investor payment statement",
			"user_id":  u.Id,
			"year":     ro.Year,
			"quarter":  ro.Quarter,
			"err":      err,
		}).Warning("not created investor payment statement")
		return "", errors.Wrap(err, "fail http payment statement request")
	}
	if resp.StatusCode != http.StatusOK {
		buff, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", errors.Wrap(err, "fail http transaction balance report request")
		}
		if resp.StatusCode == http.StatusNotFound {
			return "", factory.ErrorNotFound.New(errors.New("not found payment statement data"))
		}

		err = jsonutils.Unmarshal(buff, &respBody)
		if err != nil {
			return "", errors.Wrap(err, "fail http transaction balance report Unmarshal")
		}
	}

	respBody, err = io.ReadAll(resp.Body)
	if err != nil {
		str := string(respBody)
		factory.Logger(ctx).WithFields(logrus.Fields{
			"user_id":   u.Id,
			"user_name": u.Name,
			"resp":      str,
		}).Error("fail http payment statement request")
		return "", errors.Wrap(err, "fail http fee request")
	}
	var dat map[string]interface{}
	if err := json.Unmarshal(respBody, &dat); err != nil {
		factory.Logger(ctx).WithFields(logrus.Fields{
			"user_id":   u.Id,
			"user_name": u.Name,
			"response":  dat,
			"err":       err,
		}).Error("fail http payment statement Unmarshal")
		return "", errors.Wrap(err, "fail http fee Unmarshal")
	}
	if success, ok := dat["success"].(bool); !ok || !success {
		factory.Logger(ctx).WithFields(logrus.Fields{
			"user_id":   u.Id,
			"user_name": u.Name,
			"response":  dat,
		}).Error("fail http payment statement request")
		return "", errors.New("fail http payment statement request")
	}
	result, ok := dat["result"].(string)
	if !ok {
		return "", errors.New("fail interface assertion")
	}
	return result, nil
}

type TaxOfficePaymentStatementInput struct {
	Year     int64             `json:"year"`
	UserList []ReportUserInput `json:"user_list"`
	SpvInput
	Recreate bool `json:"recreate"`
}

// RequestTaxOfficePaymentStatement 세무서용 지불조서 요청
func RequestTaxOfficePaymentStatement(ctx context.Context, ro *ReportOutput, spv *Spv, userList map[int64]*User) (string, error) {
	var dto TaxOfficePaymentStatementInput
	var respBody []byte

	if spv == nil {
		return "", errors.New("spv is nil")
	}

	if ro == nil {
		return "", errors.New("report is nil")
	}
	const URL = "/api/download-tax-office-payment-statement"

	year, err := strconv.ParseInt(ro.Year, 10, 64)
	if err != nil {
		return "", errors.Wrap(err, "fail get year for payment statement")
	}

	dto.Year = year
	dto.SpvId = spv.Id
	dto.SpvNo = spv.CorpNo
	dto.SpvAddress = addressToString(spv.City, spv.Province, spv.Address, spv.Address2)
	dto.SpvName = spv.Name
	dto.SpvPhone = PhoneNumberSplit(spv.Phone)
	dto.UserList = make([]ReportUserInput, 0)
	for _, user := range userList {
		identificationNumber, err := GetIdentificationNumber(ctx, user)
		if err != nil {
			return "", errors.Wrap(err, "fail get identification number for tax office payment statement request")
		}
		address2 := ""
		if user.Address2 != nil {
			address2 = *user.Address2
		}

		dto.UserList = append(dto.UserList, ReportUserInput{
			UserId:               user.Id,
			UserAddress:          addressToString(user.City, user.Province, user.Address, address2),
			UserName:             user.Name,
			IdentificationNumber: identificationNumber,
		})
	}
	dto.Recreate = true

	pbytes, _ := json.Marshal(dto)
	body := bytes.NewBuffer(pbytes)
	resp, err := http.Post(serviceutils.GetStatutoryBooksServiceUrl()+URL, "application/json", body)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"category": "create tax office payment statement request",
			"spv_id":   spv.Id,
			"year":     ro.Year,
			"quarter":  ro.Quarter,
			"err":      err,
		}).Warning("not created tax office payment statement request")
		return "", errors.Wrap(err, "fail http tax office payment statement request")
	}
	if resp.StatusCode != http.StatusOK {
		buff, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", errors.Wrap(err, "fail http tax office payment statement request")
		}
		if resp.StatusCode == http.StatusNotFound {
			return "", factory.ErrorNotFound.New(errors.New("not found payment statement data"))
		}

		err = jsonutils.Unmarshal(buff, &respBody)
	}

	respBody, err = io.ReadAll(resp.Body)
	if err != nil {
		str := string(respBody)
		factory.Logger(ctx).WithFields(logrus.Fields{
			"spv_id":   spv.Id,
			"spv_name": spv.Name,
			"resp":     str,
		}).Error("fail http tax office payment statement request")
		return "", errors.Wrap(err, "fail http tax office payment statement request")
	}
	var dat map[string]interface{}
	if err := json.Unmarshal(respBody, &dat); err != nil {
		factory.Logger(ctx).WithFields(logrus.Fields{
			"spv_id":   spv.Id,
			"spv_name": spv.Name,
			"response": dat,
			"err":      err,
		}).Error("fail http payment statement Unmarshal")
		return "", errors.Wrap(err, "fail http fee Unmarshal")
	}
	if success, ok := dat["success"].(bool); !ok || !success {
		factory.Logger(ctx).WithFields(logrus.Fields{
			"spv_id":   spv.Id,
			"spv_name": spv.Name,
			"response": dat,
		}).Error("fail http payment statement request")
		return "", errors.New("fail http payment statement request")
	}
	result, ok := dat["result"].(string)
	if !ok {
		return "", errors.New("fail interface assertion")
	}
	return result, nil
}

func (ro *ReportOutput) CreateTaxOfficePaymentStatement(ctx context.Context, spv *Spv, accountList map[int64]*User) (int64, error) {
	ro.CreateTitleTaxOfficePaymentStatment()
	var err error
	ro.FileUrl, err = RequestTaxOfficePaymentStatement(ctx, ro, spv, accountList)
	if err != nil {
		return 0, errors.Wrap(err, "fail request payment statement")
	}
	return factory.DB(ctx).Insert(ro)
}

type PaymentStatementSummaryInput struct {
	Year                          int64  `json:"year"`
	SpvId                         int64  `json:"spv_id"`
	CompanyAddress                string `json:"company_address"`
	CompanyName                   string `json:"company_name"`
	FuriganaCompanyName           string `json:"furigana_company_name"`
	CompanyPhone                  string `json:"company_phone"`
	CompanyNo                     string `json:"company_no"`
	CompanyRepresentative         string `json:"company_representative"`
	FuriganaCompanyRepresentative string `json:"furigana_company_representative"`
	Recreate                      bool   `json:"recreate"`
}

// RequestInvestorPaymentStatement  지불조서 요청
func RequestTaxOfficePaymentStatementSum(ctx context.Context, ro *ReportOutput, spv *Spv) (string, error) {
	var dto PaymentStatementSummaryInput
	var respBody []byte
	var err error

	if spv == nil {
		return "", errors.New("spv is nil")
	}

	if ro == nil {
		return "", errors.New("report is nil")
	}

	const URL = "/api/download-payment-statement-summary"

	year, err := strconv.ParseInt(ro.Year, 10, 64)
	if err != nil {
		return "", errors.Wrap(err, "fail get year for payment statement")
	}

	dto.Year = year
	dto.SpvId = spv.Id
	dto.CompanyAddress = addressToString(spv.City, spv.Province, spv.Address, spv.Address2)
	dto.CompanyName = spv.Name
	dto.FuriganaCompanyName = "" // TODO: 후리가나 법인명 현재 수집하지 않음
	dto.CompanyPhone = PhoneNumberSplit(spv.Phone)
	dto.CompanyNo = spv.CorpNo
	dto.CompanyRepresentative = ""         // TODO: 대표자명 현재 수집하지 않음
	dto.FuriganaCompanyRepresentative = "" // TODO: 후리가나 대표자명 현재 수집하지 않음
	dto.Recreate = true

	pbytes, _ := json.Marshal(dto)
	body := bytes.NewBuffer(pbytes)
	resp, err := http.Post(serviceutils.GetStatutoryBooksServiceUrl()+URL, "application/json", body)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"category": "create  payment statement sum",
			"spv_id":   spv.Id,
			"year":     ro.Year,
			"quarter":  ro.Quarter,
			"err":      err,
		}).Warning("not created payment statement sum")
		return "", errors.Wrap(err, "fail http payment statement sum request")
	}
	if resp.StatusCode != http.StatusOK {
		buff, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", errors.Wrap(err, "fail http payment statement sum request")
		}
		if resp.StatusCode == http.StatusNotFound {
			return "", factory.ErrorNotFound.New(errors.New("not found payment statement data"))
		}

		err = jsonutils.Unmarshal(buff, &respBody)
	}

	respBody, err = io.ReadAll(resp.Body)
	if err != nil {
		str := string(respBody)
		factory.Logger(ctx).WithFields(logrus.Fields{
			"user_id":   spv.Id,
			"user_name": spv.Name,
			"resp":      str,
		}).Error("fail http payment statement sum request")
		return "", errors.Wrap(err, "fail http payment statement sum request")
	}
	var dat map[string]interface{}
	if err := json.Unmarshal(respBody, &dat); err != nil {
		factory.Logger(ctx).WithFields(logrus.Fields{
			"user_id":   spv.Id,
			"user_name": spv.Name,
			"response":  dat,
			"err":       err,
		}).Error("fail http payment statement Unmarshal")
		return "", errors.Wrap(err, "fail http response Unmarshal")
	}
	if success, ok := dat["success"].(bool); !ok || !success {
		factory.Logger(ctx).WithFields(logrus.Fields{
			"user_id":   spv.Id,
			"user_name": spv.Name,
			"response":  dat,
		}).Error("fail http payment statement sum request")
		return "", errors.New("fail http payment statement sum request")
	}
	result, ok := dat["result"].(string)
	if !ok {
		return "", errors.New("fail interface assertion")
	}
	return result, nil
}

func PhoneNumberSplit(s string) string {
	s = strings.Replace(s, "+", "", -1)
	res := strings.Replace(s, "-", "", -1)
	res1 := res[0 : len(res)-8]
	res2 := res[len(res)-8 : len(res)-4]
	res3 := res[len(res)-4:]
	return res1 + "-" + res2 + "-" + res3
}

func (ro *ReportOutput) CreateTaxOfficePaymentStatementSum(ctx context.Context, spv *Spv) (int64, error) {
	ro.CreateTitleTaxOfficePaymentStatment()
	var err error
	ro.FileUrl, err = RequestTaxOfficePaymentStatementSum(ctx, ro, spv)
	if err != nil {
		return 0, errors.Wrap(err, "fail payment statement sum request")
	}
	return factory.DB(ctx).Insert(ro)
}

func (ReportOutput) RequestCustomerAccount(startDate, endDate *time.Time) (string, error) {
	const URL = "/api/download-customer-ledger"

	params := url.Values{}
	if startDate != nil {
		params.Add("start_date", startDate.Format(time.DateOnly))
	}
	if endDate != nil {
		params.Add("end_date", endDate.Format(time.DateOnly))
	}
	params.Add("recreate", "true")

	resp, err := RequestReportService(URL, params)
	if err != nil {
		return "", errors.Wrap(err, "fail request transaction balance report")
	}

	if !resp.Success {
		return "", searchErr(resp.ErrorCode)
	}
	return resp.Result, nil
}

func (ReportOutput) RequestTransactionRecord(startDate, endDate *time.Time) (string, error) {
	const URL = "/api/download-offering-record"
	params := url.Values{}
	if startDate != nil {
		params.Add("start_date", startDate.Format(time.DateOnly))
	}
	if endDate != nil {
		params.Add("end_date", endDate.Format(time.DateOnly))
	}
	params.Add("recreate", "true")
	resp, err := RequestReportService(URL, params)
	if err != nil {
		return "", errors.Wrap(err, "fail request transaction balance report")
	}

	if !resp.Success {
		return "", searchErr(resp.ErrorCode)
	}
	return resp.Result, nil
}

func (ReportOutput) RequestTradingDiaryReport(startDate, endDate *time.Time) (string, error) {
	const URL = "/api/download-transaction-daily-record"
	params := url.Values{}
	if startDate != nil {
		params.Add("start_date", startDate.Format(time.DateOnly))
	}
	if endDate != nil {
		params.Add("end_date", endDate.Format(time.DateOnly))
	}
	params.Add("recreate", "true")
	resp, err := RequestReportService(URL, params)
	if err != nil {
		return "", errors.Wrap(err, "fail request transaction balance report")
	}
	if !resp.Success {
		return "", searchErr(resp.ErrorCode)
	}
	return resp.Result, nil
}
