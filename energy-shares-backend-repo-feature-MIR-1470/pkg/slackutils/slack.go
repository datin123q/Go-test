package slackutils

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"
)

// httpClientはタイムアウトを明示する。http.DefaultClientはタイムアウトが
// 無いため、相手のサーバが応答しないまま接続を保持し続けると、cronジョブの
// goroutineが無期限に張り付いてしまう(MIR-1194と同じ理由)。
var httpClient = &http.Client{Timeout: 20 * time.Second}

// PostMessage はSlack Incoming Webhookへtextをそのまま投稿する。
func PostMessage(ctx context.Context, webhookURL, text string) error {
	if webhookURL == "" {
		return errors.New("slack webhook url is empty")
	}

	body, err := json.Marshal(map[string]string{"text": text})
	if err != nil {
		return errors.Wrap(err, "failed to marshal slack payload")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(body))
	if err != nil {
		return errors.Wrap(sanitizeURLError(err), "failed to build slack request")
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return errors.Wrap(sanitizeURLError(err), "failed to post to slack")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4000))
		return errors.Errorf("slack webhook returned status %d: %s", resp.StatusCode, string(respBody))
	}
	return nil
}

// sanitizeURLErrorは、webhook URL(それ自体が認証情報)がエラー文言に残らない
// ようにする。*url.Errorの文字列表現はリクエストURLを丸ごと含むため、DNS失敗や
// タイムアウトのたびにWebhook URLがログ・job_execution_histories・Sentryへ
// 漏れてしまう(apps/admin-api/controllers/job_execute_time.goのcallJobSchedulerと
// 同じ理由で、内部情報を含みうるerrはそのまま返さない)。
func sanitizeURLError(err error) error {
	var urlErr *url.Error
	if errors.As(err, &urlErr) && urlErr.Err != nil {
		return urlErr.Err
	}
	return err
}
