package controllers

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	neturl "net/url"
	"testing"
	"time"

	"github.com/hillfolk/goutils/echomiddleware"
	"github.com/labstack/echo/v4"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pangpanglabs/echoswagger/v2"
	"github.com/sirupsen/logrus"
	logrustest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
)

func TestIsJobSchedulerCallFailed(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		want       bool
	}{
		{"200は失敗ではない", http.StatusOK, false},
		{"201は失敗ではない", http.StatusCreated, false},
		{"299は失敗ではない(2xx境界)", 299, false},
		{"300は失敗", http.StatusMultipleChoices, true},
		{"400は失敗", http.StatusBadRequest, true},
		{"500は失敗", http.StatusInternalServerError, true},
		{"0(未設定)は失敗", 0, true},
		{"199は失敗(2xx未満)", 199, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isJobSchedulerCallFailed(tt.statusCode))
		})
	}
}

// withFakeJobScheduler は jobs-scheduler を模したサーバーへ差し替えて fn を実行し、
// jobs-scheduler が受け取ったリクエストパスを返す。テスト間で
// serviceutils のパッケージ変数を共有するため、このパッケージのテストは
// 並行実行しない前提(t.Parallel() を呼ばない)。
func withFakeJobScheduler(t *testing.T, statusCode int, body string, fn func(c echo.Context) error) (gotPath string, handlerErr error) {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(statusCode)
		_, _ = w.Write([]byte(body))
	}))
	defer server.Close()

	original := serviceutils.GetJobSchedulerUrl()
	serviceutils.SetJobSchedulerUrl(server.URL)
	defer serviceutils.SetJobSchedulerUrl(original)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handlerErr = fn(c)

	return gotPath, handlerErr
}

func TestDepositCheckJobExecute(t *testing.T) {
	tests := []struct {
		name            string
		schedulerStatus int
		schedulerBody   string
		wantErr         bool
	}{
		{"jobs-schedulerが200を返せば成功", http.StatusOK, `{}`, false},
		{"jobs-schedulerが500を返せば失敗", http.StatusInternalServerError, `{"error":{"message":"boom"}}`, true},
		{"jobs-schedulerが404を返せば失敗", http.StatusNotFound, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPath, err := withFakeJobScheduler(t, tt.schedulerStatus, tt.schedulerBody, func(c echo.Context) error {
				return JobApiController{}.DepositCheckJobExecute(c)
			})

			assert.Equal(t, "/va-accounts/job/execute", gotPath)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// DepositAutoProcessJobExecuteは、jobs-scheduler側に存在しないURL
// (/depositprocess/job/execute。e60922f5で削除済み)を今も呼んでいる。
// execute グループの後継ルートはQA専用ガード付きで、本番から呼ぶと
// 空ボディ200を返すため失敗を検知できず、付け替えるとかえって悪化する
// (MIR-1186で確認)。本番で使える手動実行ルートの新設は別チケットへ
// 切り出した。それまでは404のまま、「常に失敗」を正直に返す。
// 実URLへ差し替えるときは、このテストの期待パスも合わせて直すこと。
func TestDepositAutoProcessJobExecute_StillCallsRemovedUrl(t *testing.T) {
	gotPath, err := withFakeJobScheduler(t, http.StatusOK, `{}`, func(c echo.Context) error {
		return JobApiController{}.DepositAutoProcessJobExecute(c)
	})

	require.NoError(t, err)
	assert.Equal(t, "/depositprocess/job/execute", gotPath)
}

func TestDepositAutoProcessJobExecute_FailsOnNon2xx(t *testing.T) {
	_, err := withFakeJobScheduler(t, http.StatusNotFound, "", func(c echo.Context) error {
		return JobApiController{}.DepositAutoProcessJobExecute(c)
	})

	require.Error(t, err)
}

// withFakeJobSchedulerAndAuth はwithFakeJobSchedulerと同じくjobs-schedulerを
// 模したサーバーへ差し替えるが、GmoConnectUrl(MIR-1298)がjwthelper.TokenToUserで
// 管理者IDを取り、domain.GmoAozoraOAuthStateの保存にDBを使うため、JWTと
// sqliteのDBコンテキストも用意する。gotQueryは、admin-apiが発行したstateが
// jobs-schedulerへのリクエストに載っているかを確認するためのクエリ文字列。
func withFakeJobSchedulerAndAuth(t *testing.T, statusCode int, body string, fn func(c echo.Context) error) (gotPath, gotQuery string, handlerErr error) {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		w.WriteHeader(statusCode)
		_, _ = w.Write([]byte(body))
	}))
	defer server.Close()

	original := serviceutils.GetJobSchedulerUrl()
	serviceutils.SetJobSchedulerUrl(server.URL)
	defer serviceutils.SetJobSchedulerUrl(original)

	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	require.NoError(t, err)
	engine.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, engine.Close()) })
	require.NoError(t, engine.Sync2(new(domain.GmoAozoraOAuthState)))

	e := echo.New()
	dbCtx := context.WithValue(context.Background(), factory.ConextDBType, engine)
	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(dbCtx)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	token, _ := jwthelper.CreateAdminTokenForTest(99)
	c.Set("user", token)

	handlerErr = fn(c)

	return gotPath, gotQuery, handlerErr
}

func TestGmoConnectUrl(t *testing.T) {
	tests := []struct {
		name            string
		schedulerStatus int
		schedulerBody   string
		wantErr         bool
		// wantErrCode は0以外を指定すると、返るfactory.Errorのコードまで
		// 検証する。500ケースはErrorServiceUnavailable(10002)であることを
		// 確認し、json.Unmarshal後の!result.SuccessによるErrorSystem(10001)
		// へ誤って進んでいないこと(=ステータス検査が先に効いていること)を
		// 固定する。
		wantErrCode int
	}{
		{
			"jobs-schedulerが200かつsuccess:trueを返せば成功",
			http.StatusOK,
			`{"result":"https://example.com/auth","success":true}`,
			false,
			0,
		},
		{
			"jobs-schedulerが500を返せば失敗(json.Unmarshalへ進まない)",
			http.StatusInternalServerError,
			`{"error":{"message":"boom"}}`,
			true,
			10002,
		},
		{"jobs-schedulerが404を返せば失敗", http.StatusNotFound, "", true, 10002},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPath, gotQuery, err := withFakeJobSchedulerAndAuth(t, tt.schedulerStatus, tt.schedulerBody, func(c echo.Context) error {
				return JobApiController{}.GmoConnectUrl(c)
			})

			assert.Equal(t, "/gmo-aozora/auth/login", gotPath)
			assert.Contains(t, gotQuery, "state=", "admin-apiが発行したstateをjobs-schedulerへ渡すこと(MIR-1298)")
			if tt.wantErr {
				require.Error(t, err)
				if tt.wantErrCode != 0 {
					var factoryErr *factory.Error
					require.True(t, errors.As(err, &factoryErr))
					assert.Equal(t, tt.wantErrCode, factoryErr.Code)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestGmoConnectUrl_PersistsOneTimeState はMIR-1298の要求そのもの:
// GmoConnectUrlは呼ぶたびに新しいワンタイムstateを発行し、
// gmo_aozora_oauth_stateへ保存してからjobs-schedulerを呼ぶ。
func TestGmoConnectUrl_PersistsOneTimeState(t *testing.T) {
	var gotState string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotState = r.URL.Query().Get("state")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"result":"https://example.com/auth","success":true}`))
	}))
	defer server.Close()

	original := serviceutils.GetJobSchedulerUrl()
	serviceutils.SetJobSchedulerUrl(server.URL)
	defer serviceutils.SetJobSchedulerUrl(original)

	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	require.NoError(t, err)
	engine.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, engine.Close()) })
	require.NoError(t, engine.Sync2(new(domain.GmoAozoraOAuthState)))

	e := echo.New()
	dbCtx := context.WithValue(context.Background(), factory.ConextDBType, engine)
	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(dbCtx)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	token, _ := jwthelper.CreateAdminTokenForTest(123)
	c.Set("user", token)

	require.NoError(t, JobApiController{}.GmoConnectUrl(c))

	count, err := engine.Count(new(domain.GmoAozoraOAuthState))
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	var row domain.GmoAozoraOAuthState
	has, err := engine.Get(&row)
	require.NoError(t, err)
	require.True(t, has)
	assert.Equal(t, int64(123), row.AdminUserId)
	assert.True(t, row.ExpiresAt.After(time.Now().UTC()), "有効期限が未来に設定されていること")

	// ②レビューの指摘(MIR-1298): 件数と行の存在だけでなく、jobs-schedulerへ
	// 実際に送ったstateのsha256が、保存したlookupと一致することまで確認する
	// (別の値を送ってしまうバグでも従来のアサーションは通ってしまうため)。
	require.NotEmpty(t, gotState)
	assert.Equal(t, domain.Sha256Hex(gotState), row.Lookup)
}

// TestGmoConnectUrl_NoJWTReturnsTokenInvalidNotPermissionDenied は②レビューの
// 指摘(MIR-1298): jwthelper.TokenToUserは既にfactory.ErrorTokenInvalid(401)で
// ラップして返す。ここでErrorPermissionDenied(403)へ包み直さないことを確認する。
func TestGmoConnectUrl_NoJWTReturnsTokenInvalidNotPermissionDenied(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	// JWTを設定しない(c.Set("user", ...)を呼ばない)。

	err := JobApiController{}.GmoConnectUrl(c)
	require.Error(t, err)
	var factoryErr *factory.Error
	require.True(t, errors.As(err, &factoryErr))
	assert.Equal(t, factory.ErrorTokenInvalid.Code, factoryErr.Code)
}

// TestRedactURLFromErr はMIR-1298の確認: net/httpの*url.Errorはメッセージに
// リクエストURLをそのまま含むため、GmoConnectUrlが呼ぶ?state=<S>付きのURLへの
// 接続失敗ログにstateが漏れる。rawURLのクエリ文字列がエラーメッセージに
// 含まれていれば取り除くことを確認する。
func TestRedactURLFromErr(t *testing.T) {
	t.Run("クエリを含むURLはメッセージから取り除く", func(t *testing.T) {
		rawURL := "http://jobs-scheduler/gmo-aozora/auth/login?state=super-secret-state"
		err := fmt.Errorf(`Get "%s": dial tcp: connection refused`, rawURL)

		redacted := redactURLFromErr(err, rawURL)
		assert.NotContains(t, redacted.Error(), "super-secret-state")
		assert.Contains(t, redacted.Error(), "http://jobs-scheduler/gmo-aozora/auth/login")
	})

	t.Run("クエリが無いURLはそのまま", func(t *testing.T) {
		rawURL := "http://jobs-scheduler/va-accounts/job/execute"
		err := fmt.Errorf(`Get "%s": dial tcp: connection refused`, rawURL)

		redacted := redactURLFromErr(err, rawURL)
		assert.Equal(t, err.Error(), redacted.Error())
	})

	t.Run("nilはnilのまま", func(t *testing.T) {
		assert.Nil(t, redactURLFromErr(nil, "http://x?y=z"))
	})

	// ①敵対的検証の指摘(MIR-1298): net/httpが実際に返す*url.Errorは、rawURLの
	// 生の文字列と一致するとは限らない(StripPassword等で書き換えられることが
	// ある)。文字列一致に頼らず、*url.Error自体のURLフィールドから
	// クエリを取り除けることを、実際のurl.Errorで確認する。
	t.Run("実際のurl.Errorは構造体のURLフィールドから直接取り除く", func(t *testing.T) {
		rawURL := "http://jobs-scheduler/gmo-aozora/auth/login?state=super-secret-state"
		urlErr := &neturl.Error{Op: "Get", URL: rawURL, Err: errors.New("connection refused")}

		redacted := redactURLFromErr(urlErr, rawURL)
		assert.NotContains(t, redacted.Error(), "super-secret-state")

		var gotUrlErr *neturl.Error
		require.True(t, errors.As(redacted, &gotUrlErr), "返り値も*url.Errorのままであること(型情報を保つ)")
		assert.Equal(t, "http://jobs-scheduler/gmo-aozora/auth/login", gotUrlErr.URL)
	})
}

func TestNewGmoConnectUrlState(t *testing.T) {
	a, err := newGmoConnectUrlState()
	require.NoError(t, err)
	b, err := newGmoConnectUrlState()
	require.NoError(t, err)

	assert.Len(t, a, 64, "32バイトのhexは64文字")
	assert.NotEqual(t, a, b, "毎回異なる値になること")
}

func TestRespondJobSchedulerResult(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantErr    bool
	}{
		{"200は成功", http.StatusOK, false},
		{"500は失敗", http.StatusInternalServerError, true},
		{"404は失敗", http.StatusNotFound, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := respondJobSchedulerResult(context.Background(), tt.statusCode, []byte(`{}`), "test")

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// withCapturedLogger は、factory.Loggerが返すロガーを差し替え、出力されたログを
// 後から検査できるようにする。
func withCapturedLogger() (context.Context, *logrustest.Hook) {
	logger, hook := logrustest.NewNullLogger()
	return context.WithValue(context.Background(), echomiddleware.ContextLoggerName, logrus.NewEntry(logger)), hook
}

// assertNoSecretInLogs は、どのログの本文・フィールドにもsecretが含まれないことを確かめる。
func assertNoSecretInLogs(t *testing.T, hook *logrustest.Hook, secret string) {
	t.Helper()
	for _, entry := range hook.AllEntries() {
		line, err := entry.String()
		require.NoError(t, err)
		assert.NotContains(t, line, secret)
	}
}

// GmoConnectUrlの成功時の本文はGMOあおぞら銀行の認可URL(state・nonce・client_id
// を含む)そのものなので、成功時は本文をログに出さない。失敗時の本文は診断に
// 要るので残す。staging(MIR-1194の取り消し)を取り込んだときに、成功時も本文を
// 出す形へ戻ってしまったことがあるため固定する。
func TestRespondJobSchedulerResult_LogsBodyOnlyOnFailure(t *testing.T) {
	const secret = "super-secret-state"
	body := []byte(`{"url":"https://bank.example/auth?state=` + secret + `&nonce=n"}`)

	tests := []struct {
		name        string
		statusCode  int
		wantBodyLog bool
	}{
		{"200は本文をログに出さない", http.StatusOK, false},
		{"500は本文をログに残す", http.StatusInternalServerError, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, hook := withCapturedLogger()

			_ = respondJobSchedulerResult(ctx, tt.statusCode, body, "test")

			require.NotEmpty(t, hook.AllEntries())
			if tt.wantBodyLog {
				assert.Equal(t, string(body), hook.LastEntry().Data["body"])
			} else {
				assertNoSecretInLogs(t, hook, secret)
			}
		})
	}
}

// リクエストを組み立てられない場合も、state入りのURLを画面へ返すエラーやログに
// 出さず、ErrorServiceUnavailable(500)だけを返す(MIR-1298)。
func TestCallJobScheduler_BuildRequestErrorHidesURL(t *testing.T) {
	const secret = "super-secret-state"
	ctx, hook := withCapturedLogger()

	_, _, err := callJobScheduler(ctx, http.MethodGet, "http://jobs-scheduler/gmo-aozora/auth/login?state="+secret+"\x00", "test")

	require.Error(t, err)
	var factoryErr *factory.Error
	require.True(t, errors.As(err, &factoryErr))
	assert.Equal(t, http.StatusInternalServerError, factoryErr.Status())
	assert.NotContains(t, err.Error(), secret)
	require.NotEmpty(t, hook.AllEntries())
	assertNoSecretInLogs(t, hook, secret)
}

// jobs-scheduler自体に接続できない(DNS失敗・タイムアウト等)場合も、
// 管理者の入力ではなく上流の障害なので factory.ErrorParameter(400)
// ではなく factory.ErrorServiceUnavailable(500)として扱われることを固定する。
// 固定ポート(127.0.0.1:1等)への接続だと、実行環境のファイアウォール設定に
// よってはSYNが黙って捨てられ、タイムアウトの無いhttp.DefaultClientが
// テストごとハングしうる。ローカルで確実に「即座に接続拒否」させるため、
// 一度listenしてすぐcloseしたアドレスを使う。
func TestDepositCheckJobExecute_TransportErrorIsServiceUnavailable(t *testing.T) {
	closedListener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	closedAddr := closedListener.Addr().String()
	require.NoError(t, closedListener.Close())

	original := serviceutils.GetJobSchedulerUrl()
	serviceutils.SetJobSchedulerUrl("http://" + closedAddr)
	defer serviceutils.SetJobSchedulerUrl(original)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handlerErr := JobApiController{}.DepositCheckJobExecute(c)

	require.Error(t, handlerErr)
	var factoryErr *factory.Error
	require.True(t, errors.As(handlerErr, &factoryErr))
	assert.Equal(t, http.StatusInternalServerError, factoryErr.Status())
}

// MIR-1243: ジョブを実行する2本のルートはPOSTで登録し、GETでは受け付けない
// (GETだとブラウザの再読み込みや戻る操作で再実行されうる)。照会や
// GMO連携のURL取得は副作用が無い、または理由があってGETのままにしてある。
func TestJobApiController_Init_RouteMethods(t *testing.T) {
	e := echo.New()
	JobApiController{}.Init(echoswagger.NewNop(e).Group("jobs", "/jobs"))

	methodsByPath := map[string][]string{}
	for _, r := range e.Routes() {
		methodsByPath[r.Path] = append(methodsByPath[r.Path], r.Method)
	}

	tests := []struct {
		name string
		path string
		want string
	}{
		{name: "入金確認APIの実行はPOST", path: "/jobs/execute-deposit-check-api", want: http.MethodPost},
		{name: "保有預託金自動入金の実行はPOST", path: "/jobs/execute-deposit-process-api", want: http.MethodPost},
		{name: "実行時間の照会はGETのまま", path: "/jobs/execute-time", want: http.MethodGet},
		{name: "GMO連携URLの取得はGETのまま", path: "/jobs/gmo-connect-url", want: http.MethodGet},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, []string{tt.want}, methodsByPath[tt.path])
		})
	}
}
