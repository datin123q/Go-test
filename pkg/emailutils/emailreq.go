package emailutils

import (
	"context"
	"fmt"
	"net/http"
	url2 "net/url"
	"time"

	"github.com/pkg/errors"
	"github.com/samber/lo"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	"github.com/go-resty/resty/v2"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
)

// WIP - 作業中のファイルのため、レビュー不要
// クリーンアーキテクチャ前のメール送信を実装予定
var (
	client      *resty.Client
	retryClient *resty.Client
)

func init() {
	var httpClient = &http.Client{
		Transport: otelhttp.NewTransport(
			http.DefaultTransport,
			otelhttp.WithTracerProvider(otel.GetTracerProvider()),
			otelhttp.WithPropagators(otel.GetTextMapPropagator()),
			otelhttp.WithSpanNameFormatter(func(_ string, r *http.Request) string {
				return fmt.Sprintf("%s %s", r.Method, r.URL.Host)
			}),
		),
	}

	client = resty.NewWithClient(httpClient).
		SetTimeout(10*time.Second).
		SetHeader("Content-Type", "application/json")

	retryClient = client.Clone().
		SetRetryCount(2).
		SetRetryWaitTime(5 * time.Second)

}

// SendEmailRequest 함수를 통해 메일 발송 요청을 보낼 수 있다.
// EmailRequest 는 통합 메일 서비스로 메일 발송 요청을 보내기 위한 구조체이다.
func SendEmailRequest(ctx context.Context, emailRequest EmailRequest) ([]string, error) {

	var messageIds []int64

	url, err := url2.JoinPath(serviceutils.GetEmailServiceUrl(), SEND_REQUEST_PATH)
	if err != nil {
		return nil, errors.Wrap(err, "error joining url")
	}

	// MIR-1345: テンプレート名と宛先の件数だけをログに出力する。宛先そのもの
	// (メールアドレス・氏名)は出さない。成功時はこの後のログに出るメッセージ
	// IDでメールサービスのDBを引ける。応答が返らなかった場合(タイムアウト等)
	// はメッセージIDも残らないため、その場合はこのバックエンド側のmail_retry/
	// mail_failedテーブルかテンプレート名・時刻で引く。
	factory.Logger(ctx).
		WithTime(time.Now()).
		WithField("templateName", emailRequest.TemplateName).
		WithField("receiverCount", countReceivers(emailRequest.MailData)).
		Info("SendEmailRequest")

	traceOnlyCtx := extractTraceOnlyContext(ctx)

	resp, err := client.R().
		SetContext(traceOnlyCtx).
		SetBody(emailRequest).
		SetResult(&messageIds). // 結果をmessageIdsに格納する。
		Post(url)

	factory.Logger(ctx).
		WithTime(time.Now()).
		WithField("status code", resp.StatusCode()).
		WithField("responseBody", resp.String()).
		WithField("responseTime", resp.Time()).
		Info("Legacy SendEmailRequest")

	if err != nil {
		return nil, fmt.Errorf("error sending email request: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("received non-2xx response: %s", resp.Status())
	}

	// int64 slice to string slice
	int64SliceToString := func(slice []int64) []string {
		strSlice := lo.Map(slice, func(x int64, _ int) string {
			return fmt.Sprintf("%d", x)
		})
		return strSlice
	}

	return int64SliceToString(messageIds), nil
}

// SendEmailRequestRetry 는 SendEmailRequest 함수와 동일하지만, resty client의 retry 설정이 다르다.
func SendEmailRequestRetry(ctx context.Context, emailReq EmailRequest) ([]string, error) {

	var messageIds []int64

	url, err := url2.JoinPath(serviceutils.GetEmailServiceUrl(), SEND_REQUEST_PATH)
	if err != nil {
		return nil, errors.Wrap(err, "error joining url")
	}

	// MIR-1345: SendEmailRequestと同じ理由・同じ調べ方(宛先そのものは出さず、
	// 件数だけ出す)。
	factory.Logger(ctx).
		WithTime(time.Now()).
		WithField("templateName", emailReq.TemplateName).
		WithField("receiverCount", countReceivers(emailReq.MailData)).
		Info("SendEmailRequestRetry")

	traceOnlyCtx := extractTraceOnlyContext(ctx)

	resp, err := retryClient.R().
		SetContext(traceOnlyCtx).
		SetBody(emailReq).
		SetResult(&messageIds). // 結果をmessageIdsに格納する。
		Post(url)

	factory.Logger(ctx).
		WithTime(time.Now()).
		WithField("status code", resp.StatusCode()).
		WithField("responseBody", resp.String()).
		WithField("responseTime", resp.Time()).
		Info("SendEmailRequestRetry")

	if err != nil {
		return nil, fmt.Errorf("error sending email request: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("received non-2xx response: %s", resp.Status())
	}

	// int64 slice to string slice
	int64SliceToString := func(slice []int64) []string {
		strSlice := lo.Map(slice, func(x int64, _ int) string {
			return fmt.Sprintf("%d", x)
		})
		return strSlice
	}

	return int64SliceToString(messageIds), nil
}

// MIR-1345: 宛先の件数だけをログに出すため、メールアドレスそのものには
// 触れずに件数へ変換する。MailDataは1件のEmailRequestに複数含まれうるため、
// 全件のReceiversを合算する。
func countReceivers(mailData []MailData) int {
	count := 0
	for _, data := range mailData {
		count += len(data.Receivers)
	}
	return count
}

func extractTraceOnlyContext(ctx context.Context) context.Context {
	spanCtx := trace.SpanContextFromContext(ctx)
	if !spanCtx.IsValid() {
		return context.Background()
	}
	return trace.ContextWithSpanContext(context.Background(), spanCtx)
}
