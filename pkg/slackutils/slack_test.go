package slackutils

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostMessage(t *testing.T) {
	t.Run("200を返せば成功しペイロードにtextが載る", func(t *testing.T) {
		// httptestのハンドラは別goroutineで動く。requireの失敗はt.FailNow()経由で
		// runtime.Goexitを呼ぶが、testingパッケージはテスト本体以外の
		// goroutineからのFailNowを未定義動作としているため、ハンドラ内はassertに
		// 統一し、検証はハンドラの外(サーバ終了後)で行う。
		var gotMethod, gotContentType string
		var gotBody map[string]string
		var bodyErr, unmarshalErr error
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotMethod = r.Method
			gotContentType = r.Header.Get("Content-Type")
			body, err := io.ReadAll(r.Body)
			bodyErr = err
			if err == nil {
				unmarshalErr = json.Unmarshal(body, &gotBody)
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		err := PostMessage(context.Background(), server.URL, "09月18日(金)の勤怠連絡をお願いします。")
		require.NoError(t, err)

		require.NoError(t, bodyErr)
		require.NoError(t, unmarshalErr)
		assert.Equal(t, http.MethodPost, gotMethod)
		assert.Equal(t, "application/json", gotContentType)
		assert.Equal(t, "09月18日(金)の勤怠連絡をお願いします。", gotBody["text"])
	})

	t.Run("200以外はエラーを返す", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("invalid_payload"))
		}))
		defer server.Close()

		err := PostMessage(context.Background(), server.URL, "text")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "400")
	})

	t.Run("webhook URLが空ならリクエストせずエラーを返す", func(t *testing.T) {
		err := PostMessage(context.Background(), "", "text")
		require.Error(t, err)
	})

	// webhook URLはそれ自体が認証情報(トークンを含む)であり、*url.Errorの
	// 文字列表現はリクエストURLを丸ごと含むため、DNS失敗等でそのまま返すと
	// ログ・job_execution_histories・Sentryへ漏れる(①敵対的検証で指摘)。
	t.Run("接続失敗時のエラーにwebhook URLが含まれない", func(t *testing.T) {
		secretURL := "http://127.0.0.1:1/services/T00000000/B00000000/SuperSecretToken123"
		err := PostMessage(context.Background(), secretURL, "text")
		require.Error(t, err)
		assert.NotContains(t, err.Error(), "SuperSecretToken123")
		assert.NotContains(t, err.Error(), secretURL)
	})

	// ②レビューで見つかった穴: http.NewRequestWithContextがURLのパースに
	// 失敗した場合(制御文字混入など)も*url.Errorを返すため、Doの失敗だけ
	// sanitizeしていると、こちらの経路からwebhook URLが漏れる。
	t.Run("リクエスト構築失敗時のエラーにwebhook URLが含まれない", func(t *testing.T) {
		secretURL := "http://hooks.slack.com/services/T00/B00/SuperSecretToken123\n"
		err := PostMessage(context.Background(), secretURL, "text")
		require.Error(t, err)
		assert.NotContains(t, err.Error(), "SuperSecretToken123")
	})
}
