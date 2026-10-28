package gmo_aozora

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
)

type client struct {
	authorization string
	basePath      string
}

func NewWebHookClient() client {
	return client{
		authorization: generateAuthorization(),
		basePath:      webHookClientCfg.BasePath,
	}
}

func (c client) Subscribe(ctx context.Context, status string, eventTypes map[string]string) error {
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

	client := resty.NewWithClient(httpClient)
	req := client.R().SetContext(ctx)
	req.SetHeader("Authorization", c.authorization)
	req.SetHeader("Content-Type", "application/json")
	req.SetBody(map[string]interface{}{
		"subscribeStatus": status,
		"eventTypes":      eventTypes,
	})
	resp, err := req.Post(c.basePath + "/subscribe")
	if err != nil {
		return err
	}
	if resp.StatusCode() != 200 {
		return err
	}
	return nil
}

func (c client) SubscribeStatus(ctx context.Context) (string, error) {
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

	client := resty.NewWithClient(httpClient)
	req := client.R().SetContext(ctx)
	req.SetHeader("Authorization", c.authorization)
	req.SetHeader("Content-Type", "application/json;charset=UTF-8")
	req.SetHeader("Accept", "application/json;charset=UTF-8")
	resp, err := req.Get(c.basePath + "/subscribe-status")
	if err != nil {
		return "", err
	}
	if resp.StatusCode() != 200 {
		return resp.String(), nil
	}
	return resp.String(), nil
}

func (c client) GetList(ctx context.Context) (string, error) {
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

	client := resty.NewWithClient(httpClient)
	req := client.R().SetContext(ctx)
	req.SetHeader("Authorization", c.authorization)
	req.SetHeader("Content-Type", "application/json;charset=UTF-8")
	req.SetHeader("Accept", "application/json;charset=UTF-8")
	resp, err := req.Get(c.basePath + "/unsentlist/va-deposit-transaction")
	if err != nil {
		return "", err
	}
	if resp.StatusCode() != 200 {
		return resp.String(), nil
	}
	return resp.String(), nil
}
