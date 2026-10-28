package gmo_aozora

import (
	"context"
	"net/http"

	corporateclient "github.com/gmoaozora/gmo-aozora-api-go/corporateclient"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type ACCOUNT_USE_TYPE string

const (
	LOAN_DEPOSIT      ACCOUNT_USE_TYPE = "LOAN_DEPOSIT"      //대출입금
	LOAN_WITHDRAWAL   ACCOUNT_USE_TYPE = "LOAN_WITHDRAWAL"   //대출출금
	EQUITY_DEPOSIT    ACCOUNT_USE_TYPE = "EQUITY_DEPOSIT"    //자본입금
	EQUITY_WITHDRAWAL ACCOUNT_USE_TYPE = "EQUITY_WITHDRAWAL" //자본출금
)

type Accounts []*Account

var BankAccounts Accounts

type Account struct {
	AccountUseType      ACCOUNT_USE_TYPE `json:"account_type"`
	AccountId           string           `json:"account_id"`
	AccountTypeCode     string           `json:"account_type_code"`
	AccountNumber       string           `json:"account_number"`
	RemitterName        string           `json:"remitter_name"`
	BankCode            string           `json:"bank_code"`
	BranchCode          string           `json:"branch_code"`
	AuthKey             string           `json:"auth_key"`
	TransferLimitAmount string           `json:"transfer_limit_amount"`
	LastTransactionDate string           `json:"last_transaction_date"`
}

func (a Account) New(accountNumber, accountHolder, bankCode, branchCode string, accountType ACCOUNT_USE_TYPE) *Account {
	return &Account{
		AccountNumber:  accountNumber,
		RemitterName:   accountHolder,
		AccountUseType: accountType,
		BankCode:       bankCode,
		BranchCode:     branchCode,
	}
}

func (a *Account) GetAccountType() ACCOUNT_USE_TYPE {
	return a.AccountUseType
}

func (a *Account) SetTransferLimitAmount(transferLimitAmount string) {
	a.TransferLimitAmount = transferLimitAmount
}

func (a *Account) GetTransferLimitAmount() string {
	return a.TransferLimitAmount
}

func (a *Account) SetLastTransactionDate(lastTransactionDate string) {
	a.LastTransactionDate = lastTransactionDate
}

func (a *Account) GetLastTransactionDate() string {
	return a.LastTransactionDate
}

func (a *Account) GetAccountId() string {
	return a.AccountId
}

func (a *Account) GetAccountNumber() string {
	return a.AccountNumber
}

func (a *Account) GetRemitterName() string {
	return a.RemitterName
}

func (a *Account) GetBankCode() string {
	return a.BankCode
}

func (a *Account) GetBranchCode() string {
	return a.BranchCode
}

func (a *Account) GetAuthKey() string {
	return a.AuthKey
}

func (a *Account) GetAccountTypeCode() string {
	return a.AccountTypeCode
}

func (a *Account) SetAccountId(accountId string) {
	a.AccountId = accountId
}

func (a *Account) SetAccountNumber(accountNumber string) {
	a.AccountNumber = accountNumber
}

func (a *Account) SetRemitterName(remitterName string) {
	a.RemitterName = remitterName
}

func (a *Account) SetBankCode(bankCode string) {
	a.BankCode = bankCode
}

func (a *Account) SetBranchCode(branchCode string) {
	a.BranchCode = branchCode
}

func (a *Account) SetAuthKey(authKey string) {
	a.AuthKey = authKey
}

func (a *Account) SetAccountTypeCode(accountTypeCode string) {
	a.AccountTypeCode = accountTypeCode
}

func (items Accounts) GetAccountByAccountUseType(accountUseType ACCOUNT_USE_TYPE) (*Account, error) {
	for _, account := range items {
		if account.AccountUseType == accountUseType {
			return account, nil
		}
	}

	return nil, errors.New("not found account by account use type " + string(accountUseType))
}

func (items Accounts) GetByAccountNumber(accountNumber string) *Account {
	for i := range items {
		if items[i].GetAccountNumber() == accountNumber {
			return items[i]
		}
	}
	return nil
}

func (items Accounts) InitAccountInfo(accountNumber, accountId, accountTypeCode, amountLimit string) {
	for i := range items {
		if items[i].GetAccountNumber() == accountNumber {
			items[i].SetAccountId(accountId)
			items[i].SetAccountTypeCode(accountTypeCode)
			items[i].SetTransferLimitAmount(amountLimit)
		}
	}
}

func (items Accounts) GetByAccountId(accountId string) *Account {
	for i := range items {
		if items[i].GetAccountId() == accountId {
			return items[i]
		}
	}
	return nil
}

// VerifyAccountsIncludeConfigured は、accessTokenで取得したGMOの口座一覧に、この
// サービスが実際にAPIで使う設定済み口座(accountsのうち口座番号が空でないもの)が
// 「1つ以上」含まれているかを確認する(MIR-1298、河内さんの判断で2026-09-25に
// 条件を変更)。トークン交換・更新の直後、DBへ保存する前のゲートとして使う。
// GetAccountInfoと違い、内部でグローバルなclientInfoのトークンを使わず、
// 呼び出し元が渡すaccessTokenだけを使う——保存が確定する前のトークンを検証する
// ためのものなので、まだグローバルへ反映していない値で確認する必要がある。
//
// 全件一致ではなく1件以上の一致にした理由: 本番で設定口座番号とGMO側の実際の
// 口座一覧がズレているケースが確認されており(設定側の不備。別チケットで対応)、
// 全件一致を要求すると正規のログイン・更新まで機械的に弾いてしまう。第三者の
// トークンには自社の口座番号が1つも含まれないはずなので、なりすまし対策としては
// 1件以上の一致で足りる。設定の口座番号が全て空の場合(比較対象が無い)は、
// これまでどおり失敗として扱う。
//
// 口座番号の完全一致で比べる(支店コードは比べない。設定の支店コードの形式が
// 確かめられていないため)。銀行側の口座番号が空の行(定期・外貨)は無視する。
// 通信エラーと口座の不一致は区別し、通信エラーは最大3回まで再試行する
// (リクエスト間隔は全リクエスト共有のレートリミッタ(3秒に1回)に従う)。
func VerifyAccountsIncludeConfigured(ctx context.Context, accessToken string, accounts Accounts) error {
	required := make([]string, 0, len(accounts))
	for _, account := range accounts {
		if account != nil && account.AccountNumber != "" {
			required = append(required, account.AccountNumber)
		}
	}
	if len(required) == 0 {
		return errors.New("empty configured bank accounts")
	}

	client, err := getAccountInfoClient()
	if err != nil {
		return errors.Wrap(err, "fail load corporate client")
	}

	const maxAttempts = 3
	var res corporateclient.AccountsResponse
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		var httpResp *http.Response
		res, httpResp, lastErr = CallWith429Retry(ctx, "AccountApi.AccountsUsingGET(verify)", func(ctx context.Context) (corporateclient.AccountsResponse, *http.Response, error) {
			return getAccountsWithToken(ctx, client, accessToken)
		})
		if lastErr == nil {
			break
		}
		logrus.WithFields(logrus.Fields{
			"attempt":        attempt,
			"httpStatusCode": HTTPStatusCode(httpResp),
			"error":          lastErr,
		}).Warn("fail account verify get, retrying")
	}
	if lastErr != nil {
		return errors.Wrap(lastErr, "fail account verify get")
	}

	present := make(map[string]struct{}, len(res.Accounts))
	for _, apiAccount := range res.Accounts {
		if apiAccount.AccountNumber == "" {
			continue
		}
		present[apiAccount.AccountNumber] = struct{}{}
	}

	matchedCount := 0
	for _, accountNumber := range required {
		if _, ok := present[accountNumber]; ok {
			matchedCount++
		}
	}
	if matchedCount > 0 && matchedCount < len(required) {
		// ②レビューの指摘(MIR-1298): 一部だけ一致した場合(通しはする)、
		// 警告を残す。全滅の場合はこのあとエラーで返すため、ここでは
		// 重ねてログを出さない。今回の本番の設定ズレ(EQUITY_*・LOAN_*が
		// 長期間見つからなかった)は、既存のGetAccountInfoがログに出す
		// だけで気付かれずにいたため長引いた。口座番号自体はログに出さない
		// (既存のGetAccountInfoと同じ方針。CLAUDE.mdのReview Checklist
		// 「口座番号もログに出さない」)。
		logrus.WithFields(logrus.Fields{
			"required": len(required),
			"matched":  matchedCount,
		}).Warn("some configured accounts were not found in GMO response")
	}
	if matchedCount == 0 {
		return errors.Errorf("none of the %d configured accounts were found in GMO response", len(required))
	}
	return nil
}
