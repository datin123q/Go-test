package gmo_aozora

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"time"
	"unicode/utf8"

	corporateclient "github.com/gmoaozora/gmo-aozora-api-go/corporateclient"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	maxRetryFor429         = 3
	defaultRetryAfterDelay = 5 * time.Second
)

var (
	errMax429Retry = errors.New("gmo aozora api: exceeded retry attempts for 429 responses")
)

// callWith429Retry wraps GMO AOZORA API requests and retries when HTTP 429 is returned.
func callWith429Retry[T any](ctx context.Context, operation string, fn func(context.Context) (T, *http.Response, error)) (T, *http.Response, error) {
	var zero T
	var lastResp *http.Response
	var lastErr error
	var lastResult T

	for attempt := 1; attempt <= maxRetryFor429; attempt++ {
		res, resp, err := fn(ctx)
		if !shouldRetry429(resp) {
			return res, resp, err
		}

		lastResp = resp
		lastErr = err
		lastResult = res

		wait := parseRetryAfter(resp)
		waitSeconds := wait.Seconds()
		logrus.WithFields(logrus.Fields{
			"operation":    operation,
			"attempt":      attempt,
			"wait_seconds": fmt.Sprintf("%f(s)", waitSeconds),
		}).Warn("GMO AOZORA API returned 429, retrying")

		drainBody(resp)

		if attempt == maxRetryFor429 {
			if lastErr == nil {
				lastErr = errMax429Retry
			} else {
				lastErr = errors.Wrap(lastErr, errMax429Retry.Error())
			}
			return lastResult, lastResp, errors.Wrap(lastErr, operation)
		}

		if err := waitContext(ctx, wait); err != nil {
			return zero, resp, err
		}
	}

	return zero, lastResp, errors.Wrap(errMax429Retry, operation)
}

// CallWith429Retry exposes the retry helper for other packages.
func CallWith429Retry[T any](ctx context.Context, operation string, fn func(context.Context) (T, *http.Response, error)) (T, *http.Response, error) {
	return callWith429Retry(ctx, operation, fn)
}

type swaggerError interface {
	error
	Body() []byte
	Model() interface{}
}

// client.GenericSwaggerError が swaggerError を満たすことをコンパイル時に固定する。errors.As は
// この一致に依存しており、SDKの再生成でメソッドが変わると errDetail が常に nil になり、エラーの
// 原因がログから静かに消える(MIR-1127)。
var _ swaggerError = corporateclient.GenericSwaggerError{}

// GetSwaggerErrorDetail は err.(client.GenericSwaggerError) という無条件の型アサーションを避ける
// ためのヘルパー(MIR-1127)。CallWith429Retry は 429 リトライを使い切ると errors.Wrap で元の
// エラーをラップして返すため、無条件の型アサーションはこの経路で必ず panic する。errors.As は
// ラップされたエラーチェーンを辿って一致する型を探すため、ラップの有無に関わらず安全に取り出せる。
// 一致しない場合は nil を返し、呼び出し側はそのままログに出力すればよい。
// Model() は SDK が値型で返すため、Content-Type が application/json のレスポンスであれば、
// GMOあおぞらが定義するフィールド名と一致しない JSON(GMO以外の装置が返す別形式のエラー等)でも
// デコード自体は成功し、ゼロ値の構造体が非nilで返ってくることがある。この場合に構造体を優先すると
// 生のレスポンス本文が失われるため、ゼロ値のときは Body() へフォールバックする。
func GetSwaggerErrorDetail(err error) interface{} {
	var swaggerErr swaggerError
	if !errors.As(err, &swaggerErr) {
		return nil
	}

	if model := swaggerErr.Model(); model != nil && !reflect.ValueOf(model).IsZero() {
		return model
	}
	if body := swaggerErr.Body(); len(body) > 0 {
		return string(body)
	}
	return nil
}

// maxSwaggerErrorDetailLen は GetSwaggerErrorDetailString が返す文字列に掛ける上限(MIR-1128)。
// 中間装置(WAF/ALB等)が返す HTML のエラーページが Body() フォールバック経由で丸ごとログへ載る
// ケースだけでなく、GMO が明細単位のエラー(TransferError.TransferErrorDetails 等)を大量の
// 明細に対して返し、Model() 経由でも1行のログが肥大化しうるため、Model()/Body() どちらの経路
// で取り出した値かに関わらず、ログ・エラーメッセージへ渡す直前の最終文字列へ一律に掛ける。
const maxSwaggerErrorDetailLen = 2000

// truncateSwaggerErrorDetailString は上限を超えた分を切り捨て、切り詰めたことが分かるよう末尾に
// 印を付ける。GMOのエラーメッセージは日本語(マルチバイトのUTF-8)を含むため、単純にバイト数で
// 切ると文字の途中で分割されて不正なUTF-8になりうる。有効な文字境界まで後ろへ戻してから切り詰める。
func truncateSwaggerErrorDetailString(s string) string {
	if len(s) <= maxSwaggerErrorDetailLen {
		return s
	}
	truncated := s[:maxSwaggerErrorDetailLen]
	for len(truncated) > 0 && !utf8.ValidString(truncated) {
		truncated = truncated[:len(truncated)-1]
	}
	return truncated + "...(truncated)"
}

// GetSwaggerErrorDetailString はログ出力・エラーメッセージへの埋め込み専用の文字列版。
// GetSwaggerErrorDetail の戻り値をそのまま fmt.Sprintf("%+v", ...) すると、nil のときに
// リテラル文字列 "<nil>" が生成されログやエラーメッセージへ混入するため、nil のときは
// 空文字を返す。
func GetSwaggerErrorDetailString(err error) string {
	detail := GetSwaggerErrorDetail(err)
	if detail == nil {
		return ""
	}
	return truncateSwaggerErrorDetailString(fmt.Sprintf("%+v", detail))
}

func shouldRetry429(resp *http.Response) bool {
	return resp != nil && resp.StatusCode == http.StatusTooManyRequests
}

func drainBody(resp *http.Response) {
	if resp == nil || resp.Body == nil {
		return
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
}

func parseRetryAfter(resp *http.Response) time.Duration {
	if resp == nil {
		return defaultRetryAfterDelay
	}
	raw := resp.Header.Get("Retry-After")
	if raw == "" {
		return defaultRetryAfterDelay
	}

	if seconds, err := strconv.Atoi(raw); err == nil && seconds > 0 {
		return time.Duration(seconds) * time.Second
	}

	if retryTime, err := http.ParseTime(raw); err == nil {
		delay := time.Until(retryTime)
		if delay > 0 {
			return delay
		}
	}

	return defaultRetryAfterDelay
}

func waitContext(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
