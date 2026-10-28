package service

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/nexway"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/nexwaysign"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
)

const jpkiTestSecret = "9ab6bf804c3937fcf27ddaad49721c6b"

// setJpkiTestConfig はグローバルなJPKI設定を書き換える。他のテストへ影響しないよう、
// t.Cleanupで元の値に戻す。
func setJpkiTestConfig(t *testing.T, cfg nexway.Config) {
	t.Helper()
	original := nexway.Config{
		JpkiAuthBaseUrl:      serviceutils.GetServiceNexwayJpkiAuthBaseUrl(),
		JpkiAuthToken:        serviceutils.GetServiceNexwayJpkiAuthToken(),
		JpkiBpoBaseUrl:       serviceutils.GetServiceNexwayJpkiBpoBaseUrl(),
		JpkiBpoToken:         serviceutils.GetServiceNexwayJpkiBpoToken(),
		JpkiBpoSigningSecret: serviceutils.GetServiceNexwayJpkiBpoSigningSecret(),
	}
	t.Cleanup(func() {
		serviceutils.SetNexwayConfig(original)
		serviceutils.SetNexwayJpkiConfig(original)
	})
	serviceutils.SetNexwayConfig(cfg)
	serviceutils.SetNexwayJpkiConfig(cfg)
}

func newJpkiTestService(now time.Time) *JpkiNexwayService {
	return &JpkiNexwayService{
		httpClient: &http.Client{Timeout: NexwayApiTimeout},
		now:        func() time.Time { return now },
	}
}

func TestJpkiNexwayService_GetSignatureResult(t *testing.T) {
	const resultWithPersonalInfo = `{
		"reference_id":"ref-1","request_id":"req-1","status":"completed","result":"succeeded",
		"completed_at":"2026-10-15T10:00:00+09:00",
		"jpki_result":{"name":"山田太郎","birthday":"520000101","address":"東京都千代田区","gender":"male"}
	}`

	t.Run("正常: 結果を返し、個人情報は構造体に持たない", func(t *testing.T) {
		var gotMethod, gotPath, gotAuth string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotMethod, gotPath, gotAuth = r.Method, r.URL.Path, r.Header.Get("Authorization")
			_, _ = w.Write([]byte(resultWithPersonalInfo))
		}))
		defer server.Close()
		setJpkiTestConfig(t, nexway.Config{JpkiAuthBaseUrl: server.URL + "/", JpkiAuthToken: "auth-token"})

		got, err := newJpkiTestService(time.Now()).GetSignatureResult(context.Background(), "req-1")
		require.NoError(t, err)

		assert.Equal(t, http.MethodGet, gotMethod)
		assert.Equal(t, "/api/verification/req-1/result", gotPath, "ベースURL末尾のスラッシュがあっても二重にならない")
		assert.Equal(t, "Token auth-token", gotAuth)
		assert.Equal(t, &JpkiSignatureResult{
			ReferenceId: "ref-1", RequestId: "req-1", Status: "completed", Result: "succeeded",
			CompletedAt: "2026-10-15T10:00:00+09:00",
		}, got)
	})

	t.Run("検証依頼IDはパスとして安全にエスケープする", func(t *testing.T) {
		var gotRawPath string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotRawPath = r.URL.EscapedPath()
			_, _ = w.Write([]byte(`{}`))
		}))
		defer server.Close()
		setJpkiTestConfig(t, nexway.Config{JpkiAuthBaseUrl: server.URL, JpkiAuthToken: "t"})

		_, err := newJpkiTestService(time.Now()).GetSignatureResult(context.Background(), "a/b?c")
		require.NoError(t, err)
		assert.Equal(t, "/api/verification/a%2Fb%3Fc/result", gotRawPath)
	})

	t.Run("404・4040002はErrNexwayJpkiRequestNotFound", func(t *testing.T) {
		// 試験環境で、存在しないrequest_idを照会したときの実際の応答(2026-10-26)。
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error_code":"4040002","error_detail":"not found"}`))
		}))
		defer server.Close()
		setJpkiTestConfig(t, nexway.Config{JpkiAuthBaseUrl: server.URL, JpkiAuthToken: "t"})

		_, err := newJpkiTestService(time.Now()).GetSignatureResult(context.Background(), "req-1")
		assert.ErrorIs(t, err, ErrNexwayJpkiRequestNotFound)
		assert.NotErrorIs(t, err, ErrNexwayInvalidResponse)
	})

	// 4040002以外の404は、request_idが存在しない証拠にならない(ベースURLの設定誤りで、別のサーバーが
	// 返した404など)。存在しないとして扱うと、定期ジョブが会員の記録を失敗にしてメールを送ってしまう。
	t.Run("4040002以外の404はErrNexwayInvalidResponse(あとで確認し直す)", func(t *testing.T) {
		for name, body := range map[string]string{
			"別のエラーコード":  `{"error_code":"4040001","error_detail":"not found"}`,
			"JSONでない本文": `<html><body>404 Not Found</body></html>`,
			"本文なし":      ``,
		} {
			body := body
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(body))
			}))
			setJpkiTestConfig(t, nexway.Config{JpkiAuthBaseUrl: server.URL, JpkiAuthToken: "t"})

			_, err := newJpkiTestService(time.Now()).GetSignatureResult(context.Background(), "req-1")
			assert.ErrorIs(t, err, ErrNexwayInvalidResponse, name)
			assert.NotErrorIs(t, err, ErrNexwayJpkiRequestNotFound, name)
			server.Close()
		}
	})

	t.Run("その他の異常ステータスはErrNexwayInvalidResponse", func(t *testing.T) {
		// 403はIP制限(4030001)だけ。503のコードは現在 503001 で、今後 5030001 に直る(ネクスウェイ、
		// 2026-10-23)。コードの値ではなくHTTPステータスで判断するため、どちらでも同じ扱いになる。
		for _, tc := range []struct {
			status int
			code   string
		}{
			{http.StatusBadRequest, "4000001"},
			{http.StatusUnauthorized, "4010001"},
			{http.StatusForbidden, "4030001"},
			{http.StatusInternalServerError, "5000001"},
			{http.StatusServiceUnavailable, "503001"},
			{http.StatusServiceUnavailable, "5030001"},
		} {
			status, code := tc.status, tc.code
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(status)
				_, _ = w.Write([]byte(`{"error_code":"` + code + `"}`))
			}))
			setJpkiTestConfig(t, nexway.Config{JpkiAuthBaseUrl: server.URL, JpkiAuthToken: "t"})

			_, err := newJpkiTestService(time.Now()).GetSignatureResult(context.Background(), "req-1")
			assert.ErrorIs(t, err, ErrNexwayInvalidResponse, "status=%d", status)
			server.Close()
		}
	})

	t.Run("本文が壊れていればErrNexwayInvalidResponse。本文はエラーに含めない", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(`{"name":"山田太郎" broken`))
		}))
		defer server.Close()
		setJpkiTestConfig(t, nexway.Config{JpkiAuthBaseUrl: server.URL, JpkiAuthToken: "t"})

		_, err := newJpkiTestService(time.Now()).GetSignatureResult(context.Background(), "req-1")
		require.ErrorIs(t, err, ErrNexwayInvalidResponse)
		assert.NotContains(t, err.Error(), "山田")
	})

	t.Run("通信できなければErrNexwayUnavailable", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		url := server.URL
		server.Close() // 接続先を閉じる
		setJpkiTestConfig(t, nexway.Config{JpkiAuthBaseUrl: url, JpkiAuthToken: "t"})

		_, err := newJpkiTestService(time.Now()).GetSignatureResult(context.Background(), "req-1")
		assert.ErrorIs(t, err, ErrNexwayUnavailable)
	})

	t.Run("URLまたはトークンが未設定なら通信せずErrNexwayJpkiNotConfigured。ホ方式の設定は流用しない", func(t *testing.T) {
		var called atomic.Int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called.Add(1)
		}))
		defer server.Close()

		for name, cfg := range map[string]nexway.Config{
			"URLなし":  {JpkiAuthToken: "t", ApplicantsApi: server.URL, Token: "protech-token"},
			"トークンなし": {JpkiAuthBaseUrl: server.URL, ApplicantsApi: server.URL, Token: "protech-token"},
		} {
			setJpkiTestConfig(t, cfg)
			_, err := newJpkiTestService(time.Now()).GetSignatureResult(context.Background(), "req-1")
			assert.ErrorIs(t, err, ErrNexwayJpkiNotConfigured, name)
		}
		assert.Zero(t, called.Load())
	})

	t.Run("request_idが空なら通信しない", func(t *testing.T) {
		var called atomic.Int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called.Add(1) }))
		defer server.Close()
		setJpkiTestConfig(t, nexway.Config{JpkiAuthBaseUrl: server.URL, JpkiAuthToken: "t"})

		_, err := newJpkiTestService(time.Now()).GetSignatureResult(context.Background(), "")
		assert.ErrorIs(t, err, ErrNexwayInvalidResponse)
		assert.Zero(t, called.Load())
	})
}

func TestJpkiNexwayService_RegisterBpoApplicant(t *testing.T) {
	input := JpkiBpoApplicantInput{
		ReferenceId: "ref-1", RequestId: "req-1",
		Name: "山田太郎", Birthday: "1990-01-02", Address: "東京都千代田区丸の内1-1-1 ビル101",
	}
	now := time.Date(2026, 10, 15, 12, 0, 0, 0, time.FixedZone("JST", 9*3600))

	t.Run("正常: 署名付きで申請し、申請IDを返す", func(t *testing.T) {
		var (
			gotMethod, gotPath, gotAuth, gotContentType string
			gotSignature, gotTime                       string
			gotBody                                     []byte
		)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotMethod, gotPath = r.Method, r.URL.Path
			gotAuth, gotContentType = r.Header.Get("Authorization"), r.Header.Get("Content-Type")
			gotSignature, gotTime = r.Header.Get(nexwaysign.HeaderSignature), r.Header.Get(nexwaysign.HeaderTime)
			gotBody, _ = io.ReadAll(r.Body)
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"applicant_id":9001,"reference_id":"ref-1","accepted_at":"2026-10-15T12:00:00+09:00"}`))
		}))
		defer server.Close()
		setJpkiTestConfig(t, nexway.Config{JpkiBpoBaseUrl: server.URL + "/", JpkiBpoToken: "bpo-token", JpkiBpoSigningSecret: jpkiTestSecret})

		got, err := newJpkiTestService(now).RegisterBpoApplicant(context.Background(), input)
		require.NoError(t, err)
		assert.Equal(t, &JpkiBpoApplicantOutput{ApplicantId: 9001, ReferenceId: "ref-1", AcceptedAt: "2026-10-15T12:00:00+09:00"}, got)

		assert.Equal(t, http.MethodPost, gotMethod)
		assert.Equal(t, "/api/v2/applicants/profiles", gotPath)
		assert.Equal(t, "Token bpo-token", gotAuth, "署名用キーではなくAPIトークンをAuthorizationへ")
		assert.Equal(t, "application/json", gotContentType)
		assert.JSONEq(t, `{"reference_id":"ref-1","request_id":"req-1","name":"山田太郎","birthday":"1990-01-02","address":"東京都千代田区丸の内1-1-1 ビル101"}`, string(gotBody))

		// 署名は「実際に送った本体のバイト列 + x-kyc-time」に対するもの。サーバー側で
		// 受信した本体から再計算して、一致することを確認する(受信側の検証と同じ手順)。
		assert.Equal(t, "2026-10-15T12:00:00+09:00", gotTime)
		assert.NoError(t, nexwaysign.Verify(jpkiTestSecret, gotBody, gotSignature, gotTime, now))
	})

	t.Run("申請済み(4000010・4000011)はErrNexwayJpkiAlreadyRegistered", func(t *testing.T) {
		for _, code := range []string{"4000010", "4000011"} {
			code := code
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"error_code":"` + code + `","message":"bad request","details":["x"]}`))
			}))
			setJpkiTestConfig(t, nexway.Config{JpkiBpoBaseUrl: server.URL, JpkiBpoToken: "t", JpkiBpoSigningSecret: jpkiTestSecret})

			_, err := newJpkiTestService(now).RegisterBpoApplicant(context.Background(), input)
			assert.ErrorIs(t, err, ErrNexwayJpkiAlreadyRegistered, code)
			assert.NotErrorIs(t, err, ErrNexwayJpkiApplicantRejected, code)
			server.Close()
		}
	})

	t.Run("申請内容の不備(4000008・4000009)はErrNexwayJpkiApplicantRejected。detailsはエラーに含めない", func(t *testing.T) {
		for _, code := range []string{"4000008", "4000009"} {
			code := code
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"error_code":"` + code + `","message":"bad request","details":["nameを入力してください。","山田太郎"]}`))
			}))
			setJpkiTestConfig(t, nexway.Config{JpkiBpoBaseUrl: server.URL, JpkiBpoToken: "t", JpkiBpoSigningSecret: jpkiTestSecret})

			_, err := newJpkiTestService(now).RegisterBpoApplicant(context.Background(), input)
			require.ErrorIs(t, err, ErrNexwayJpkiApplicantRejected, code)
			assert.NotContains(t, err.Error(), "山田")
			server.Close()
		}
	})

	t.Run("認証・署名・IP制限・メンテナンス・サーバーエラーは再送してよい(ErrNexwayInvalidResponse)", func(t *testing.T) {
		cases := []struct {
			status int
			code   string
		}{
			{http.StatusUnauthorized, "4010001"}, {http.StatusUnauthorized, "4010003"},
			{http.StatusUnauthorized, "4010004"}, {http.StatusUnauthorized, "4010005"},
			{http.StatusForbidden, "4030001"}, {http.StatusForbidden, "4030009"},
			{http.StatusInternalServerError, "5000001"}, {http.StatusServiceUnavailable, "5030001"},
			{http.StatusBadGateway, ""},
		}
		for _, tc := range cases {
			tc := tc
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(`{"error_code":"` + tc.code + `"}`))
			}))
			setJpkiTestConfig(t, nexway.Config{JpkiBpoBaseUrl: server.URL, JpkiBpoToken: "t", JpkiBpoSigningSecret: jpkiTestSecret})

			_, err := newJpkiTestService(now).RegisterBpoApplicant(context.Background(), input)
			require.ErrorIs(t, err, ErrNexwayInvalidResponse, "%d %s", tc.status, tc.code)
			assert.NotErrorIs(t, err, ErrNexwayJpkiApplicantRejected, "%d %s", tc.status, tc.code)
			assert.NotErrorIs(t, err, ErrNexwayJpkiAlreadyRegistered, "%d %s", tc.status, tc.code)
			server.Close()
		}
	})

	t.Run("成功(201)なのに申請IDが無い・本文が壊れている場合は成功と扱わない", func(t *testing.T) {
		for name, body := range map[string]string{"申請IDなし": `{"reference_id":"ref-1"}`, "壊れた本文": `not json`} {
			body := body
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(body))
			}))
			setJpkiTestConfig(t, nexway.Config{JpkiBpoBaseUrl: server.URL, JpkiBpoToken: "t", JpkiBpoSigningSecret: jpkiTestSecret})

			_, err := newJpkiTestService(now).RegisterBpoApplicant(context.Background(), input)
			assert.ErrorIs(t, err, ErrNexwayInvalidResponse, name)
			server.Close()
		}
	})

	t.Run("通信できなければErrNexwayUnavailable", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		url := server.URL
		server.Close()
		setJpkiTestConfig(t, nexway.Config{JpkiBpoBaseUrl: url, JpkiBpoToken: "t", JpkiBpoSigningSecret: jpkiTestSecret})

		_, err := newJpkiTestService(now).RegisterBpoApplicant(context.Background(), input)
		assert.ErrorIs(t, err, ErrNexwayUnavailable)
	})

	t.Run("URL・トークン・署名用キーのどれかが未設定なら、通信せずErrNexwayJpkiNotConfigured", func(t *testing.T) {
		var called atomic.Int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called.Add(1) }))
		defer server.Close()

		for name, cfg := range map[string]nexway.Config{
			"URLなし":  {JpkiBpoToken: "t", JpkiBpoSigningSecret: jpkiTestSecret, ApplicantsApi: server.URL, Token: "protech"},
			"トークンなし": {JpkiBpoBaseUrl: server.URL, JpkiBpoSigningSecret: jpkiTestSecret},
			"署名用キーなし(署名なしでは送らない)": {JpkiBpoBaseUrl: server.URL, JpkiBpoToken: "t"},
		} {
			setJpkiTestConfig(t, cfg)
			_, err := newJpkiTestService(now).RegisterBpoApplicant(context.Background(), input)
			assert.ErrorIs(t, err, ErrNexwayJpkiNotConfigured, name)
		}
		assert.Zero(t, called.Load())
	})

	t.Run("署名はリクエストごとの時刻で作る", func(t *testing.T) {
		var times []string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			times = append(times, r.Header.Get(nexwaysign.HeaderTime))
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"applicant_id":1}`))
		}))
		defer server.Close()
		setJpkiTestConfig(t, nexway.Config{JpkiBpoBaseUrl: server.URL, JpkiBpoToken: "t", JpkiBpoSigningSecret: jpkiTestSecret})

		_, err := newJpkiTestService(now).RegisterBpoApplicant(context.Background(), input)
		require.NoError(t, err)
		_, err = newJpkiTestService(now.Add(90*time.Second)).RegisterBpoApplicant(context.Background(), input)
		require.NoError(t, err)
		assert.Equal(t, []string{"2026-10-15T12:00:00+09:00", "2026-10-15T12:01:30+09:00"}, times)
	})
}
