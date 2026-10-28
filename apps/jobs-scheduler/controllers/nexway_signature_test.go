package controllers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	logrustest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/nexwaysign"
)

const (
	signatureTestSecret = "9ab6bf804c3937fcf27ddaad49721c6b"
	jpkiResultBody      = `{"applicant_id":13,"reference_id":"ref-1","ekyc_type":"jpki","status":"completed","result":"match"}`
)

var signatureTestNow = time.Date(2026, 10, 15, 12, 0, 0, 0, time.FixedZone("JST", 9*3600))

// callSignedWebhook は署名検証ミドルウェアを通した受信を1回行い、HTTPステータスと
// 後続のハンドラに渡った本体(渡らなければnil)を返す。
func callSignedWebhook(t *testing.T, secret, body string, headers map[string]string) (int, *string) {
	t.Helper()
	var received *string
	handler := func(c echo.Context) error {
		b, err := io.ReadAll(c.Request().Body)
		require.NoError(t, err)
		s := string(b)
		received = &s
		return c.JSON(http.StatusOK, true)
	}
	mw := verifyNexwayBpoSignature(func() string { return secret }, func() time.Time { return signatureTestNow })

	req := httptest.NewRequest(http.MethodPost, "/nexway/bpo/webhook", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	require.NoError(t, mw(handler)(echo.New().NewContext(req, rec)))
	return rec.Code, received
}

func signedHeaders(t *testing.T, body string, at time.Time) map[string]string {
	t.Helper()
	signature, timestamp, err := nexwaysign.SignRequest(signatureTestSecret, []byte(body), at)
	require.NoError(t, err)
	return map[string]string{nexwaysign.HeaderSignature: signature, nexwaysign.HeaderTime: timestamp}
}

func TestVerifyNexwayBpoSignature_Jpki(t *testing.T) {
	tests := []struct {
		name       string
		secret     string
		body       string
		headers    func(t *testing.T) map[string]string
		wantStatus int
		wantPassed bool
	}{
		{
			name: "正しい署名: 通り、後続のハンドラが本体を読める", secret: signatureTestSecret, body: jpkiResultBody,
			headers: func(t *testing.T) map[string]string {
				return signedHeaders(t, jpkiResultBody, signatureTestNow.Add(-time.Minute))
			},
			wantStatus: http.StatusOK, wantPassed: true,
		},
		{
			name: "ekyc_typeが大文字でも署名を検証する", secret: signatureTestSecret, body: strings.Replace(jpkiResultBody, `"jpki"`, `"JPKI"`, 1),
			headers:    func(t *testing.T) map[string]string { return map[string]string{} },
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "署名ヘッダーが無い", secret: signatureTestSecret, body: jpkiResultBody,
			headers:    func(t *testing.T) map[string]string { return map[string]string{} },
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "タイムスタンプのヘッダーだけ無い", secret: signatureTestSecret, body: jpkiResultBody,
			headers: func(t *testing.T) map[string]string {
				h := signedHeaders(t, jpkiResultBody, signatureTestNow)
				delete(h, nexwaysign.HeaderTime)
				return h
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "署名が不一致(別の鍵で作った署名)", secret: signatureTestSecret, body: jpkiResultBody,
			headers: func(t *testing.T) map[string]string {
				signature, timestamp, err := nexwaysign.SignRequest("another-secret-another-secret-0123", []byte(jpkiResultBody), signatureTestNow)
				require.NoError(t, err)
				return map[string]string{nexwaysign.HeaderSignature: signature, nexwaysign.HeaderTime: timestamp}
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "本体が改ざんされている(署名は元の本体のもの)", secret: signatureTestSecret,
			body: strings.Replace(jpkiResultBody, `"match"`, `"unmatch"`, 1),
			headers: func(t *testing.T) map[string]string {
				return signedHeaders(t, jpkiResultBody, signatureTestNow)
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "10分以上前の署名(再生攻撃)", secret: signatureTestSecret, body: jpkiResultBody,
			headers: func(t *testing.T) map[string]string {
				return signedHeaders(t, jpkiResultBody, signatureTestNow.Add(-11*time.Minute))
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "署名用シークレットが未設定: 検証なしで通さず500(BPOの再送で設定後に受け直す)", secret: "", body: jpkiResultBody,
			headers:    func(t *testing.T) map[string]string { return signedHeaders(t, jpkiResultBody, signatureTestNow) },
			wantStatus: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, received := callSignedWebhook(t, tt.secret, tt.body, tt.headers(t))

			assert.Equal(t, tt.wantStatus, status)
			if tt.wantPassed {
				require.NotNil(t, received)
				assert.Equal(t, tt.body, *received, "後続のハンドラへ本体がそのまま渡る")
			} else {
				assert.Nil(t, received, "検証に失敗した通知は後続の処理へ渡さない")
			}
		})
	}
}

func TestVerifyNexwayBpoSignature_NonJpkiPassesWithoutSignature(t *testing.T) {
	// ホ方式(ProTech ID Checker)の通知には署名オプションが無い。検証しない。
	bodies := map[string]string{
		"ekyc_typeなし(ホ方式)":     `{"applicant_id":13,"reference_id":"ref-1","status":"completed","result":"match"}`,
		"ekyc_typeが空":          `{"applicant_id":13,"ekyc_type":""}`,
		"ekyc_typeがjpki以外":     `{"applicant_id":13,"ekyc_type":"ic_chip"}`,
		"JSONでない(後続のBindが400)": `not json`,
		"空の本体":                 ``,
	}
	for name, body := range bodies {
		t.Run(name, func(t *testing.T) {
			status, received := callSignedWebhook(t, signatureTestSecret, body, nil)

			assert.Equal(t, http.StatusOK, status)
			require.NotNil(t, received)
			assert.Equal(t, body, *received)
		})
	}
}

func TestVerifyNexwayBpoSignature_NonJpkiPassesEvenWithoutSecret(t *testing.T) {
	// 署名用シークレットが未設定でも、ホ方式の通知は止めない(JPKI用の設定不備でホ方式を巻き込まない)。
	status, received := callSignedWebhook(t, "", `{"applicant_id":13,"status":"completed"}`, nil)

	assert.Equal(t, http.StatusOK, status)
	assert.NotNil(t, received)
}

func TestVerifyNexwayBpoSignature_RejectionDoesNotLogBody(t *testing.T) {
	hook := logrustest.NewGlobal()
	t.Cleanup(func() { logrus.StandardLogger().ReplaceHooks(make(logrus.LevelHooks)) })

	callSignedWebhook(t, signatureTestSecret, jpkiResultBody, nil)
	callSignedWebhook(t, "", jpkiResultBody, nil)

	require.NotEmpty(t, hook.AllEntries())
	for _, entry := range hook.AllEntries() {
		line := entry.Message
		for _, v := range entry.Data {
			line += " " + toLogString(v)
		}
		assert.NotContains(t, line, "ref-1")
		assert.NotContains(t, line, "applicant_id")
	}
}

// 認証の無い公開口のため、本文の大きさに上限がある。上限ちょうどは通し、超えれば413で、後続へ渡さない。
func TestVerifyNexwayBpoSignature_BodyLimit(t *testing.T) {
	padded := func(size int) string {
		// ekyc_typeがjpkiでない(ホ方式として素通しされる)有効なJSONを、指定の大きさで作る。
		prefix := `{"applicant_id":13,"padding":"`
		suffix := `"}`
		return prefix + strings.Repeat("a", size-len(prefix)-len(suffix)) + suffix
	}

	t.Run("上限ちょうど: 通す", func(t *testing.T) {
		body := padded(nexwayWebhookMaxBodyBytes)
		status, received := callSignedWebhook(t, signatureTestSecret, body, nil)

		assert.Equal(t, http.StatusOK, status)
		require.NotNil(t, received)
		assert.Equal(t, len(body), len(*received))
	})

	t.Run("上限を1バイト超える: 413で止め、後続へ渡さない", func(t *testing.T) {
		status, received := callSignedWebhook(t, signatureTestSecret, padded(nexwayWebhookMaxBodyBytes+1), nil)

		assert.Equal(t, http.StatusRequestEntityTooLarge, status)
		assert.Nil(t, received)
	})

	t.Run("JPKIの通知でも、上限を超えれば署名の検証より前に止める", func(t *testing.T) {
		body := `{"ekyc_type":"jpki","padding":"` + strings.Repeat("a", nexwayWebhookMaxBodyBytes) + `"}`
		status, received := callSignedWebhook(t, signatureTestSecret, body, signedHeaders(t, body, signatureTestNow))

		assert.Equal(t, http.StatusRequestEntityTooLarge, status)
		assert.Nil(t, received)
	})
}
