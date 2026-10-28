package gmo_aozora

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	corporateclient "github.com/gmoaozora/gmo-aozora-api-go/corporateclient"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// HTTPStatusCode は、*http.Response をそのまま logrus.Fields に渡すと
// Request.GetBody(関数値)を含むため標準JSONエンコーダで直列化に失敗し、
// ログ行そのものが出力されなくなる問題への対応(MIR-1122/MIR-1124)。
// ステータスコードだけを取り出せば直列化に失敗しない。
func TestHTTPStatusCode(t *testing.T) {
	tests := []struct {
		name string
		resp *http.Response
		want int
	}{
		{
			name: "nilのときは0を返す",
			resp: nil,
			want: 0,
		},
		{
			name: "200のときは200を返す",
			resp: &http.Response{StatusCode: http.StatusOK},
			want: http.StatusOK,
		},
		{
			name: "429のときは429を返す",
			resp: &http.Response{StatusCode: http.StatusTooManyRequests},
			want: http.StatusTooManyRequests,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, HTTPStatusCode(tt.resp))
		})
	}
}

// TestCheckAuthorizationはMIR-1126の検証。CheckAuthorizationは不一致でも
// trueを返す既存の不具合(2024-03-07の初回実装から)があるが、本番ログで
// 一致率を実測するまでは挙動を変えない方針のため、返り値は一致・不一致の
// どちらもtrueのままであることと、一致・不一致(matched)だけがログへ出て、
// 認証情報(authヘッダーの値・生成した期待値)が一切ログへ出ないことを確認する。
func TestCheckAuthorization(t *testing.T) {
	originalClientInfo := clientInfo
	originalOut := logrus.StandardLogger().Out
	originalFormatter := logrus.StandardLogger().Formatter
	originalLevel := logrus.GetLevel()
	t.Cleanup(func() {
		clientInfo = originalClientInfo
		logrus.SetOutput(originalOut)
		logrus.SetFormatter(originalFormatter)
		logrus.SetLevel(originalLevel)
	})
	// Infoログをbytes.Bufferで捕まえて解析するため、パッケージ内の別テストが
	// 将来レベルをInfo未満へ下げても空バッファでjson.Unmarshalが落ちないよう固定する。
	logrus.SetLevel(logrus.InfoLevel)

	clientInfo = ClientInfo{ClientId: "test-client-id", ClientSecret: "test-client-secret"}
	expected := generateAuthorization()

	tests := []struct {
		name string
		auth string
	}{
		{name: "一致する場合", auth: expected},
		{name: "一致しない場合", auth: "Basic invalid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logrus.SetOutput(&buf)
			logrus.SetFormatter(&logrus.JSONFormatter{})

			got := CheckAuthorization(context.Background(), tt.auth)
			assert.True(t, got, "現状は一致・不一致どちらでもtrueを返す仕様のまま(MIR-1126はまだ検証ロジックを直さない)")

			// 秘密情報が漏れていないかの確認は、ログ行のJSONパースが失敗しても
			// 必ず実行されるようにrequire.NoErrorより先に置く。ここが後だと、
			// ログが複数行になる等でパースに失敗した場合にfail-safeにならず、
			// 漏洩の検知自体がスキップされてしまう。
			output := buf.String()
			assert.NotContains(t, output, "test-client-id", "clientIdをログへ出さないこと")
			assert.NotContains(t, output, "test-client-secret", "clientSecretをログへ出さないこと")
			assert.NotContains(t, output, expected, "生成した期待値をログへ出さないこと")
			assert.NotContains(t, output, tt.auth, "受け取ったauthヘッダーの値をログへ出さないこと")

			var logLine map[string]any
			require.NoError(t, json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &logLine))
			assert.Equal(t, tt.auth == expected, logLine["matched"])
		})
	}
}

// TestCheckAuthorization_LogsRequestIdはMIR-1126の検証。不一致件数を実測
// しても、同じリクエストの他のログ(echoutils.CustomLogger()のアクセスログ)
// と突き合わせられなければ、GMOからの本物の通知か無関係な通信かを区別
// できない(敵対的検証で指摘)。当初はfactory.Logger経由で試したが、それが
// 読むechomiddleware.ContextLoggerNameはこのアプリのミドルウェアに配線され
// ておらず本番では常にフォールバックすることが分かったため、実際に
// request_idが乗る生の文字列キー"request_id"(pkg/echoutils/echo.goの
// CustomContextが設定するもの)を直接読む形に直した。このテストはその実装を
// 検証する。
func TestCheckAuthorization_LogsRequestId(t *testing.T) {
	originalClientInfo := clientInfo
	originalOut := logrus.StandardLogger().Out
	originalFormatter := logrus.StandardLogger().Formatter
	originalLevel := logrus.GetLevel()
	t.Cleanup(func() {
		clientInfo = originalClientInfo
		logrus.SetOutput(originalOut)
		logrus.SetFormatter(originalFormatter)
		logrus.SetLevel(originalLevel)
	})
	logrus.SetLevel(logrus.InfoLevel)

	clientInfo = ClientInfo{ClientId: "test-client-id", ClientSecret: "test-client-secret"}

	var buf bytes.Buffer
	logrus.SetOutput(&buf)
	logrus.SetFormatter(&logrus.JSONFormatter{})

	ctx := context.WithValue(context.Background(), "request_id", "test-request-id") //nolint:staticcheck // pkg/echoutils/echo.goのCustomContextと同じ生の文字列キーで実配線を再現する

	CheckAuthorization(ctx, "Basic invalid")

	var logLine map[string]any
	require.NoError(t, json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &logLine))
	assert.Equal(t, "test-request-id", logLine["request_id"], "同じリクエストの他のログと突き合わせるためrequest_idが載ること")
}

// TestGetAccountInfoはMIR-1271の検証。以前は口座数ぶん残高・入金明細取得を
// 繰り返していたが、口座一覧取得(AccountsUsingGET)の1回だけで完結すること、
// および設定済み口座がGMOの応答に含まれない場合にエラーになることを確認する。
func TestGetAccountInfo(t *testing.T) {
	originalGetClient := getAccountInfoClient
	originalGetAccounts := getAccounts
	t.Cleanup(func() {
		getAccountInfoClient = originalGetClient
		getAccounts = originalGetAccounts
	})

	getAccountInfoClient = func() (*corporateclient.APIClient, error) {
		return &corporateclient.APIClient{}, nil
	}

	t.Run("設定済み口座がない場合はAPIを呼び出さずエラー", func(t *testing.T) {
		callCount := 0
		getAccounts = func(context.Context, *corporateclient.APIClient) (corporateclient.AccountsResponse, *http.Response, error) {
			callCount++
			return corporateclient.AccountsResponse{}, nil, nil
		}

		err := GetAccountInfo(context.Background(), Accounts{
			Account{}.New("", "holder", "0310", "101", EQUITY_DEPOSIT),
		})
		require.Error(t, err)
		assert.ErrorContains(t, err, "empty configured bank accounts")
		assert.Zero(t, callCount)
	})

	t.Run("Config対象口座だけを初期化し、APIは1回だけ呼ぶ", func(t *testing.T) {
		callCount := 0
		getAccounts = func(context.Context, *corporateclient.APIClient) (corporateclient.AccountsResponse, *http.Response, error) {
			callCount++
			return corporateclient.AccountsResponse{Accounts: []corporateclient.Account{
				{AccountNumber: "1234567", AccountId: "account-id-1", AccountTypeCode: "01", TransferLimitAmount: "5000000"},
				{AccountNumber: "7654321", AccountId: "account-id-2", AccountTypeCode: "01", TransferLimitAmount: "3000000"},
				{AccountNumber: "9999999", AccountId: "not-configured", AccountTypeCode: "01", TransferLimitAmount: "1000000"},
			}}, nil, nil
		}

		accounts := Accounts{
			Account{}.New("1234567", "holder", "0310", "101", EQUITY_DEPOSIT),
			Account{}.New("1234567", "holder", "0310", "101", LOAN_DEPOSIT),
			Account{}.New("7654321", "holder", "0310", "102", EQUITY_WITHDRAWAL),
			Account{}.New("", "holder", "0310", "103", LOAN_WITHDRAWAL),
		}

		err := GetAccountInfo(context.Background(), accounts)
		require.NoError(t, err, "口座番号未設定の口座(LOANの空欄想定)は未検出エラーの対象にならないこと")
		assert.Equal(t, 1, callCount)
		for _, account := range accounts[:2] {
			assert.Equal(t, "account-id-1", account.AccountId)
		}
		assert.Equal(t, "account-id-2", accounts[2].AccountId)
		assert.Empty(t, accounts[3].AccountId)
	})

	t.Run("Config対象口座がGMOの取得結果にない場合はエラー", func(t *testing.T) {
		getAccounts = func(context.Context, *corporateclient.APIClient) (corporateclient.AccountsResponse, *http.Response, error) {
			return corporateclient.AccountsResponse{Accounts: []corporateclient.Account{
				{AccountNumber: "1234567", AccountId: "account-id-1", AccountTypeCode: "01"},
			}}, nil, nil
		}

		accounts := Accounts{
			Account{}.New("1234567", "holder", "0310", "101", EQUITY_DEPOSIT),
			Account{}.New("7654321", "holder", "0310", "102", EQUITY_WITHDRAWAL),
		}

		err := GetAccountInfo(context.Background(), accounts)
		require.Error(t, err)
		assert.ErrorContains(t, err, "EQUITY_WITHDRAWAL")
		assert.NotContains(t, err.Error(), "7654321", "口座番号はエラー文言に含めないこと")
	})

	t.Run("口座一覧APIのエラーを返す", func(t *testing.T) {
		callCount := 0
		getAccounts = func(context.Context, *corporateclient.APIClient) (corporateclient.AccountsResponse, *http.Response, error) {
			callCount++
			return corporateclient.AccountsResponse{}, nil, errors.New("account API error")
		}

		err := GetAccountInfo(context.Background(), Accounts{
			Account{}.New("1234567", "holder", "0310", "101", EQUITY_DEPOSIT),
		})
		require.Error(t, err)
		assert.ErrorContains(t, err, "fail account get")
		assert.Equal(t, 1, callCount)
	})

	t.Run("口座APIクライアントの初期化エラーを返す", func(t *testing.T) {
		getAccountInfoClient = func() (*corporateclient.APIClient, error) {
			return nil, errors.New("client error")
		}
		t.Cleanup(func() {
			getAccountInfoClient = func() (*corporateclient.APIClient, error) {
				return &corporateclient.APIClient{}, nil
			}
		})

		err := GetAccountInfo(context.Background(), Accounts{
			Account{}.New("1234567", "holder", "0310", "101", EQUITY_DEPOSIT),
		})
		require.Error(t, err)
		assert.ErrorContains(t, err, "fail load corporate client")
	})
}
