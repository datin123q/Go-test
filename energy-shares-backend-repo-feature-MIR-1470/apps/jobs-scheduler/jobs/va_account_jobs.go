package jobs

import (
	"context"
	"net/http"
	"strconv"
	"time"
	"unicode/utf8"

	"golang.org/x/text/width"

	"energyshares-job-scheduler/pkg/gmo_aozora"
	"energyshares-job-scheduler/pkg/timeutils"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"

	"github.com/antihax/optional"
	client "github.com/gmoaozora/gmo-aozora-api-go/corporateclient"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

func (a *AozoraJobs) GetEquityVaTransactionJob(ctx context.Context) error {
	return runJobWithExecutionHistory(ctx, JobExecutionHistoryInput{
		JobType: domain.JOB_EXECUTION_TYPE_GET_EQUITY_VA_TRANSACTION,
		JobName: jobNameGetEquityVaTransaction,
	}, a.getEquityVaTransactionJob)
}

func (a *AozoraJobs) getEquityVaTransactionJob(ctx context.Context) error {
	jobExecuteTime, err := domain.JobExecuteTime{}.GetJobLastExecuteTime(ctx)
	if err != nil {
		return err
	}

	depositAccount, err := gmo_aozora.BankAccounts.GetAccountByAccountUseType(gmo_aozora.EQUITY_DEPOSIT)
	if err != nil {
		return errors.Wrap(err, "fail GetEquityVaTransactionJob not found deposit account")
	}

	lastExecuteTime := timeutils.LocalZoneTime()
	if jobExecuteTime.DepositCheckTime.IsEmpty() {
		// 初回実行のみ
		jobExecuteTime.DepositCheckTime = jsonutils.JSONTime(lastExecuteTime.AddDate(0, 0, -7))
	}
	dataFrom := jobExecuteTime.DepositCheckTime.Time().Format(time.DateOnly)
	dataTo := lastExecuteTime.Format(time.DateOnly)

	// 입금 내역 조회
	transactions, err := gmo_aozora.GetVaTransactions(ctx, depositAccount.AccountId, dataFrom, dataTo)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"data_to":   dataTo,
			"data_from": dataFrom,
			"AccountId": depositAccount.AccountId,
			"error":     err,
		}).Error("GetEquityVaTransactionJob: Failed Va Transactions Request")
		return errors.Wrap(err, "fail GetEquityVaTransactionJob")
	}

	logrus.WithFields(logrus.Fields{
		"data_to":   dataTo,
		"data_from": dataFrom,
		"AccountId": depositAccount.AccountId,
		"count":     len(transactions),
	}).Info("GetEquityVaTransactionJob: Success Va Transactions Request")

	if err := ProcessEquityVaTransactions(ctx, transactions); err != nil {
		return err
	}

	// JobExecuteTime 업데이트
	jobExecuteTime.DepositCheckTime = jsonutils.JSONTime(lastExecuteTime)
	if err = jobExecuteTime.Update(ctx); err != nil {
		return errors.Wrap(err, "fail deposit check time update")
	}
	logrus.WithFields(logrus.Fields{
		"job execute time": jobExecuteTime.DepositCheckTime.ToString(),
	}).Info("GetEquityVaTransactionJob: Updated job execute time")

	// audit log: GMO API 실행
	UserId, err := domain.AdminUser{}.GetSuperAdminId(ctx)
	if err != nil {
		return errors.Wrap(err, "fail super admin id")
	}
	user, err := domain.User{}.GetById(ctx, UserId)
	if err != nil {
		return errors.Wrap(err, "fail get by id")
	}
	history := domain.OperatorHistory{
		UserId: user.Id,
		Email:  user.Email,
		Menu:   "commoncode.LOG_CONTENT_TYPE.GMO_API_REQUEST",
		Ip:     "::1",
		State:  domain.OPERATOR_STATE("REGISTER"),
	}
	_, err = history.Create(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to create audit log")
	}

	return nil

}

// ProcessEquityVaTransactions は定期照会で取得したVA入金を取引順に処理する。
func ProcessEquityVaTransactions(ctx context.Context, transactions []client.VaTransaction) error {
	var userId int64
	for i, t := range transactions {
		if err := domain.ValidateVaDepositItemKey(t.ItemKey); err != nil {
			factory.Logger(ctx).WithFields(logrus.Fields{
				"idx":   i,
				"error": err,
			}).Error("GetEquityVaTransactionJob: Invalid va deposit item_key")
			return errors.Wrap(err, "invalid va deposit item_key")
		}

		depositAmount, err := strconv.ParseInt(t.DepositAmount, 10, 64)
		if err != nil {
			return err
		}

		u, err := domain.UserVirtualAccount{}.GetUserVirtualAccountByBranchAndAccountNumber(ctx, t.VaBranchCode, t.VaAccountNumber)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"item_key":       t.ItemKey,
				"va_id":          t.VaId,
				"va_branch_code": t.VaBranchCode,
				"error":          err,
			}).Error("GetEquityVaTransactionJob: Fail Deposit To Virtual Account Connect Account Number")
			return err
		}
		if u == nil {
			logrus.WithFields(logrus.Fields{
				"item_key":          t.ItemKey,
				"va_id":             t.VaId,
				"va_branch_code":    t.VaBranchCode,
				"va_account_number": t.VaAccountNumber,
			}).Error("GetEquityVaTransactionJob: Not Found Virtual Account By Account Number")
			userId = 0
		} else {
			userId = u.UserId
		}

		depositTime, err := jsonutils.ParseJSONTimeJST(time.DateOnly, t.TransactionDate)
		if err != nil {
			return err
		}
		vaDeposit := &domain.VaDeposit{
			UserId:                userId,
			VaId:                  t.VaId,
			TransactionDate:       t.TransactionDate,
			ValueDate:             t.ValueDate,
			VaBranchCode:          t.VaBranchCode,
			VaBranchNameKana:      t.VaBranchNameKana,
			VaAccountNameKana:     t.VaAccountNameKana,
			RemitterNameKana:      t.RemitterNameKana,
			PaymentBankName:       t.PaymentBankName,
			PaymentBranchName:     t.PaymentBranchName,
			PartnerName:           t.PartnerName,
			Remarks:               t.Remarks,
			ItemKey:               t.ItemKey,
			VaAccountNumber:       t.VaAccountNumber,
			DepositAmount:         depositAmount,
			DepositDate:           depositTime,
			DepositStatus:         domain.NOT_USE_DEPOSIT,
			TransactionDetailType: domain.DEPOSIT_VIRTUAL_ACCOUNT,
		}

		created, err := vaDeposit.CreateOrConfirmExisting(ctx)
		if err != nil {
			if errors.Is(err, domain.ErrVaDepositTransactionConflict) {
				factory.Logger(ctx).WithFields(logrus.Fields{
					"item_key": vaDeposit.ItemKey,
					"error":    err,
				}).Error("GetEquityVaTransactionJob: Va deposit transaction conflict")
			}

			return errors.Wrap(err, "fail register va deposit")
		}
		if !created {
			factory.Logger(ctx).WithFields(logrus.Fields{
				"item_key": vaDeposit.ItemKey,
			}).Info("GetEquityVaTransactionJob: Va deposit already registered")
			continue
		}

		if vaDeposit.UserId == 0 {
			continue
		}

		if err := vaDeposit.DepositEquityProcess(ctx); err != nil {
			logrus.WithFields(logrus.Fields{
				"item_key": vaDeposit.ItemKey,
				"va_id":    vaDeposit.VaId,
				"error":    err,
			}).Error("GetEquityVaTransactionJob: Fail Deposit Equity Process VaDeposit")
			return errors.Wrap(err, "fail Deposit Equity Process VaDeposit")
		}

		logrus.WithFields(logrus.Fields{
			"idx":      i,
			"item_key": t.ItemKey,
			"va_id":    t.VaId,
		}).Info("GetEquityVaTransactionJob: Success Deposit Check Process")
	}

	return nil
}

func (a *AozoraJobs) IssueVirtualAccountJob(ctx context.Context) error {
	// 가상 계좌 요청
	logrus.WithFields(logrus.Fields{
		"desc": "check virtual accounts",
		"time": time.Now(),
	}).Info("Issue Virtual Accounts")

	isNotSufficient, err := domain.UserVirtualAccount{}.CheckIfNotSufficient(ctx)
	if err != nil {
		return err
	}
	if !isNotSufficient {
		logrus.WithFields(logrus.Fields{
			"desc": "Sufficient Virtual Accounts",
			"time": time.Now(),
		}).Info("Issue Virtual Accounts")
		return nil
	}

	jobLastExecuteTime, err := domain.JobExecuteTime{}.GetJobLastExecuteTime(ctx)
	if err != nil {
		return err
	}

	depositAccount, err := gmo_aozora.BankAccounts.GetAccountByAccountUseType(gmo_aozora.EQUITY_DEPOSIT)
	if err != nil {
		return errors.Wrap(err, "fail get by deposit account")
	}

	corporateClient, err := gmo_aozora.GetCorporateClient()
	if err != nil {
		return errors.Wrap(err, "fail get corporate client")
	}

	body := client.VaIssueRequest{VaTypeCode: VIRTUAL_ACCOUNT_TYPE, IssueRequestCount: VIRTUAL_ACCOUNT_ISSUE_COUNT, RaId: depositAccount.AccountId}

	opts := client.IssueUsingPOSTOpts{Body: optional.NewInterface(body)}

	res, httpResp, err := gmo_aozora.CallWith429Retry(ctx, "VirtualAccountApi.IssueUsingPOST", func(ctx context.Context) (client.VaIssueResponse, *http.Response, error) {
		return corporateClient.VirtualAccountApi.IssueUsingPOST(ctx, gmo_aozora.GetAccessToken(), &opts)
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err":            err,
			"httpStatusCode": gmo_aozora.HTTPStatusCode(httpResp),
			"time":           time.Now(),
		}).Error("Fail Issue Virtual Account")

		return err
	}

	accounts := VaProcess(res)

	_, err = domain.UserVirtualAccount{}.InsertMany(ctx, accounts)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err":   err,
			"count": len(accounts),
			"time":  time.Now(),
		}).Error("Fail Insert Virtual Account")

		return err
	}

	logrus.WithFields(logrus.Fields{
		"count": len(accounts),
		"desc":  "success update",
	}).Info("IssueVirtualAccountJob")
	jobLastExecuteTime.VaAccountIssueTime = jsonutils.JSONTime(time.Now())
	err = jobLastExecuteTime.Update(ctx)
	if err != nil {
		return err
	}

	// 가상계좌 발행 API실행 로그 적용
	user_id, err := domain.AdminUser{}.GetSuperAdminId(ctx)
	if err != nil {
		return err
	}
	user, err := domain.User{}.GetById(ctx, user_id)
	if err != nil {
		return err
	}
	history := domain.OperatorHistory{
		UserId: user.Id,
		Email:  user.Email,
		Menu:   "commoncode.LOG_CONTENT_TYPE.VA_BATCH_EXECUTE",
		Ip:     "::1",
		State:  domain.OPERATOR_STATE("CHANGE"),
	}
	_, err = history.Create(ctx)
	if err != nil {
		return err
	}

	return nil
}

// GetStatusVirtualAccount 가상 계좌
func (a *AozoraJobs) GetStatusVirtualAccount(ctx context.Context) error {
	// 가상 계좌 요청
	logrus.WithFields(logrus.Fields{
		"desc": "get status virtual accounts",
		"time": time.Now(),
	}).Info("Get Status Virtual Accounts")

	jobLastExecuteTime, err := domain.JobExecuteTime{}.GetJobLastExecuteTime(ctx)
	if err != nil {
		return err
	}

	corporateClient, err := gmo_aozora.GetCorporateClient()
	if err != nil {
		return errors.Wrap(err, "fail get corporate client")
	}

	body := client.VaListRequest{}
	opts := client.VaListUsingPOSTOpts{Body: optional.NewInterface(body)}

	res, httpResp, err := gmo_aozora.CallWith429Retry(ctx, "VirtualAccountApi.VaListUsingPOST", func(ctx context.Context) (client.VaListResponse, *http.Response, error) {
		return corporateClient.VirtualAccountApi.VaListUsingPOST(ctx, gmo_aozora.GetAccessToken(), &opts)
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err":            err,
			"httpStatusCode": gmo_aozora.HTTPStatusCode(httpResp),
			"time":           time.Now(),
		}).Error("Fail List Virtual Account")

		return err
	}
	for _, account := range res.VAccounts {
		logrus.WithFields(logrus.Fields{
			"va_id":                      account.VaId,
			"va_account_status_code":     account.VaStatusCode,
			"va_issue_date_time":         account.VaIssueDateTime,
			"va_change_status_date_time": account.VaChangeStatusDateTime,
		}).Info("Status VirtualAccount")
	}

	jobLastExecuteTime.VaAccountIssueTime = jsonutils.JSONTime(time.Now())
	err = jobLastExecuteTime.Update(ctx)
	if err := jobLastExecuteTime.UpdateVaAccountIssueTime(ctx, jsonutils.JSONTime(time.Now())); err != nil {
		return errors.New("fail update va account issue time")
	}
	if err != nil {
		return err
	}

	return nil
}

func VaProcess(response client.VaIssueResponse) []domain.UserVirtualAccount {
	var accounts []domain.UserVirtualAccount
	accounts = make([]domain.UserVirtualAccount, 0)
	for _, va := range response.VaList {

		newAccount := domain.UserVirtualAccount{
			VaBankName:       domain.GMO_AOZORA_BANK_NAME,
			VaBranchCode:     va.VaBranchCode,
			VaBranchName:     width.Widen.String(convertDakutenAndHandakuten(va.VaBranchNameKana)),
			VaAccountNumber:  va.VaAccountNumber,
			VaHolderNameKana: width.Widen.String(convertDakutenAndHandakuten(response.VaHolderNameKana)),
			ProjectType:      domain.EQUITY,
			LinkedStatus:     domain.OFF_VA_ACCOUNT,
			UpdateDate:       jsonutils.JSONTime(time.Now()),
		}
		accounts = append(accounts, newAccount)
	}
	return accounts
}

// ｱｼﾞｻｲ 데이터에서 다쿠텐 "ﾞ"이 포함 되어있는 상태로 변환 하고 있어서 해당 데이터를 다쿠텐을 결합한 문자(ジ)로 변환할 필요가 있습니다.
// (ｼﾞ != ジ) 해당 문자 모두 반각입니다.
func convertDakutenAndHandakuten(s string) string {
	// 반각 카타카나 문자와 다쿠텐/반탁음을 결합하여 전각 카타카나로 변환하는 매핑 테이블
	halfWidthToFullWidthWithDakuten := map[string]rune{
		"ｶﾞ": 'ガ', "ｷﾞ": 'ギ', "ｸﾞ": 'グ', "ｹﾞ": 'ゲ', "ｺﾞ": 'ゴ',
		"ｻﾞ": 'ザ', "ｼﾞ": 'ジ', "ｽﾞ": 'ズ', "ｾﾞ": 'ゼ', "ｿﾞ": 'ゾ',
		"ﾀﾞ": 'ダ', "ﾁﾞ": 'ヂ', "ﾂﾞ": 'ヅ', "ﾃﾞ": 'デ', "ﾄﾞ": 'ド',
		"ﾊﾞ": 'バ', "ﾋﾞ": 'ビ', "ﾌﾞ": 'ブ', "ﾍﾞ": 'ベ', "ﾎﾞ": 'ボ',
		"ﾊﾟ": 'パ', "ﾋﾟ": 'ピ', "ﾌﾟ": 'プ', "ﾍﾟ": 'ペ', "ﾎﾟ": 'ポ',
	}

	var result []rune
	i := 0
	for i < len(s) {
		r, size := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError {
			result = append(result, r)
			i += size
			continue
		}

		// Check if the next character is Dakuten or Handakuten
		if i+size < len(s) {
			nextRune, nextSize := utf8.DecodeRuneInString(s[i+size:])
			if nextRune == 'ﾞ' || nextRune == 'ﾟ' {
				composite := string([]rune{r, nextRune})
				if fw, ok := halfWidthToFullWidthWithDakuten[composite]; ok {
					result = append(result, fw)
					i += size + nextSize
					continue
				}
			}
		}

		// Append the current rune to the result
		result = append(result, r)
		i += size
	}

	return string(result)
}
