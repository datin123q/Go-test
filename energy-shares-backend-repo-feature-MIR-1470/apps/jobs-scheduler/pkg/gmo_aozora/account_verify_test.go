package gmo_aozora

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"

	corporateclient "github.com/gmoaozora/gmo-aozora-api-go/corporateclient"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func withVerifyAccountsStubs(t *testing.T, clientFn func() (*corporateclient.APIClient, error), accountsFn func(context.Context, *corporateclient.APIClient, string) (corporateclient.AccountsResponse, *http.Response, error)) {
	t.Helper()
	originalClient := getAccountInfoClient
	originalAccounts := getAccountsWithToken
	t.Cleanup(func() {
		getAccountInfoClient = originalClient
		getAccountsWithToken = originalAccounts
	})
	getAccountInfoClient = clientFn
	getAccountsWithToken = accountsFn
}

// TestVerifyAccountsIncludeConfigured_UsesGivenTokenNotGlobal はMIR-1298の
// 要求そのもの: 保存確定前のトークンを検証するためのものなので、呼び出し元が
// 渡したaccessTokenだけを使い、グローバルなclientInfo(GetAccessToken)は見ない
// ことを確認する。
func TestVerifyAccountsIncludeConfigured_UsesGivenTokenNotGlobal(t *testing.T) {
	originalClientInfo := clientInfo
	t.Cleanup(func() { clientInfo = originalClientInfo })
	clientInfo = ClientInfo{AccessToken: "global-token-should-not-be-used"}

	var gotToken string
	withVerifyAccountsStubs(t,
		func() (*corporateclient.APIClient, error) { return &corporateclient.APIClient{}, nil },
		func(ctx context.Context, client *corporateclient.APIClient, accessToken string) (corporateclient.AccountsResponse, *http.Response, error) {
			gotToken = accessToken
			return corporateclient.AccountsResponse{Accounts: []corporateclient.Account{
				{AccountNumber: "1234567"},
			}}, nil, nil
		},
	)

	err := VerifyAccountsIncludeConfigured(context.Background(), "fresh-token", Accounts{
		Account{}.New("1234567", "holder", "0310", "101", EQUITY_DEPOSIT),
	})
	require.NoError(t, err)
	assert.Equal(t, "fresh-token", gotToken)
}

// TestVerifyAccountsIncludeConfigured_PartialMatchSucceeds はMIR-1298、
// 2026-09-25の河内さんの判断: 本番で設定口座番号とGMO側の実際の口座一覧が
// ズレているケースが見つかったため、全件一致ではなく1件以上の一致で
// 通すよう変更した。
func TestVerifyAccountsIncludeConfigured_PartialMatchSucceeds(t *testing.T) {
	withVerifyAccountsStubs(t,
		func() (*corporateclient.APIClient, error) { return &corporateclient.APIClient{}, nil },
		func(context.Context, *corporateclient.APIClient, string) (corporateclient.AccountsResponse, *http.Response, error) {
			return corporateclient.AccountsResponse{Accounts: []corporateclient.Account{
				{AccountNumber: "1234567"},
			}}, nil, nil
		},
	)

	err := VerifyAccountsIncludeConfigured(context.Background(), "token", Accounts{
		Account{}.New("1234567", "holder", "0310", "101", EQUITY_DEPOSIT),
		Account{}.New("7654321", "holder", "0310", "102", EQUITY_WITHDRAWAL),
	})
	require.NoError(t, err, "設定した口座のうち1つでもGMOの応答にあれば通ること")
}

func TestVerifyAccountsIncludeConfigured_NoMatchFails(t *testing.T) {
	withVerifyAccountsStubs(t,
		func() (*corporateclient.APIClient, error) { return &corporateclient.APIClient{}, nil },
		func(context.Context, *corporateclient.APIClient, string) (corporateclient.AccountsResponse, *http.Response, error) {
			return corporateclient.AccountsResponse{Accounts: []corporateclient.Account{
				{AccountNumber: "9999999"},
			}}, nil, nil
		},
	)

	err := VerifyAccountsIncludeConfigured(context.Background(), "token", Accounts{
		Account{}.New("1234567", "holder", "0310", "101", EQUITY_DEPOSIT),
		Account{}.New("7654321", "holder", "0310", "102", EQUITY_WITHDRAWAL),
	})
	require.Error(t, err)
	assert.ErrorContains(t, err, "none of the 2 configured", "口座の不一致による失敗だと分かること")
	assert.NotContains(t, err.Error(), "1234567", "口座番号はエラー文言に含めないこと")
	assert.NotContains(t, err.Error(), "7654321", "口座番号はエラー文言に含めないこと")
}

// TestVerifyAccountsIncludeConfigured_PartialMatchLogsWarn は②レビューの
// 提案(MIR-1298): 一部だけ一致した場合(通しはするが)、口座番号を含まない
// 形でWarnログを残す。今回の本番の設定ズレが、既存のGetAccountInfoの
// ログだけでは長期間気付かれなかったため。
func TestVerifyAccountsIncludeConfigured_PartialMatchLogsWarn(t *testing.T) {
	originalOut := logrus.StandardLogger().Out
	originalFormatter := logrus.StandardLogger().Formatter
	originalLevel := logrus.GetLevel()
	t.Cleanup(func() {
		logrus.SetOutput(originalOut)
		logrus.SetFormatter(originalFormatter)
		logrus.SetLevel(originalLevel)
	})
	logrus.SetLevel(logrus.WarnLevel)
	var buf bytes.Buffer
	logrus.SetOutput(&buf)
	logrus.SetFormatter(&logrus.JSONFormatter{})

	withVerifyAccountsStubs(t,
		func() (*corporateclient.APIClient, error) { return &corporateclient.APIClient{}, nil },
		func(context.Context, *corporateclient.APIClient, string) (corporateclient.AccountsResponse, *http.Response, error) {
			return corporateclient.AccountsResponse{Accounts: []corporateclient.Account{
				{AccountNumber: "1234567"},
			}}, nil, nil
		},
	)

	err := VerifyAccountsIncludeConfigured(context.Background(), "token", Accounts{
		Account{}.New("1234567", "holder", "0310", "101", EQUITY_DEPOSIT),
		Account{}.New("7654321", "holder", "0310", "102", EQUITY_WITHDRAWAL),
	})
	require.NoError(t, err)

	output := buf.String()
	assert.NotContains(t, output, "1234567", "口座番号をログに出さないこと")
	assert.NotContains(t, output, "7654321", "口座番号をログに出さないこと")

	var logLine map[string]any
	require.NoError(t, json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &logLine))
	assert.Equal(t, "some configured accounts were not found in GMO response", logLine["msg"])
	assert.Equal(t, float64(2), logLine["required"])
	assert.Equal(t, float64(1), logLine["matched"])
}

func TestVerifyAccountsIncludeConfigured_IgnoresBankSideEmptyAccountNumbers(t *testing.T) {
	withVerifyAccountsStubs(t,
		func() (*corporateclient.APIClient, error) { return &corporateclient.APIClient{}, nil },
		func(context.Context, *corporateclient.APIClient, string) (corporateclient.AccountsResponse, *http.Response, error) {
			return corporateclient.AccountsResponse{Accounts: []corporateclient.Account{
				{AccountNumber: ""}, // 定期・外貨等、口座番号が空の行
				{AccountNumber: "1234567"},
			}}, nil, nil
		},
	)

	err := VerifyAccountsIncludeConfigured(context.Background(), "token", Accounts{
		Account{}.New("1234567", "holder", "0310", "101", EQUITY_DEPOSIT),
	})
	require.NoError(t, err)
}

func TestVerifyAccountsIncludeConfigured_EmptyRequiredFails(t *testing.T) {
	callCount := 0
	withVerifyAccountsStubs(t,
		func() (*corporateclient.APIClient, error) { return &corporateclient.APIClient{}, nil },
		func(context.Context, *corporateclient.APIClient, string) (corporateclient.AccountsResponse, *http.Response, error) {
			callCount++
			return corporateclient.AccountsResponse{}, nil, nil
		},
	)

	err := VerifyAccountsIncludeConfigured(context.Background(), "token", Accounts{
		Account{}.New("", "holder", "0310", "101", EQUITY_DEPOSIT),
	})
	require.Error(t, err)
	assert.ErrorContains(t, err, "empty configured bank accounts")
	assert.Zero(t, callCount, "設定済み口座が無ければAPIを呼ばないこと")
}

// TestVerifyAccountsIncludeConfigured_RetriesCommunicationErrors はMIR-1298の
// 「通信エラーは最大3回まで再試行する」の確認。
func TestVerifyAccountsIncludeConfigured_RetriesCommunicationErrors(t *testing.T) {
	callCount := 0
	withVerifyAccountsStubs(t,
		func() (*corporateclient.APIClient, error) { return &corporateclient.APIClient{}, nil },
		func(context.Context, *corporateclient.APIClient, string) (corporateclient.AccountsResponse, *http.Response, error) {
			callCount++
			if callCount < 3 {
				return corporateclient.AccountsResponse{}, nil, errors.New("connection reset")
			}
			return corporateclient.AccountsResponse{Accounts: []corporateclient.Account{
				{AccountNumber: "1234567"},
			}}, nil, nil
		},
	)

	err := VerifyAccountsIncludeConfigured(context.Background(), "token", Accounts{
		Account{}.New("1234567", "holder", "0310", "101", EQUITY_DEPOSIT),
	})
	require.NoError(t, err)
	assert.Equal(t, 3, callCount, "2回失敗しても3回目で成功すれば通ること")
}

func TestVerifyAccountsIncludeConfigured_GivesUpAfterMaxRetries(t *testing.T) {
	callCount := 0
	withVerifyAccountsStubs(t,
		func() (*corporateclient.APIClient, error) { return &corporateclient.APIClient{}, nil },
		func(context.Context, *corporateclient.APIClient, string) (corporateclient.AccountsResponse, *http.Response, error) {
			callCount++
			return corporateclient.AccountsResponse{}, nil, errors.New("connection reset")
		},
	)

	err := VerifyAccountsIncludeConfigured(context.Background(), "token", Accounts{
		Account{}.New("1234567", "holder", "0310", "101", EQUITY_DEPOSIT),
	})
	require.Error(t, err)
	assert.ErrorContains(t, err, "fail account verify get")
	assert.Equal(t, 3, callCount)
}
