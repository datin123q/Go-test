package factory

import (
	"context"

	"github.com/getsentry/sentry-go"
	"github.com/hillfolk/goutils/cronjob"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type SentryHook struct{}

func (hook *SentryHook) Fire(entry *logrus.Entry) error {
	if entry.Level <= logrus.ErrorLevel {
		err := errors.New(entry.Message)
		if errField, exists := entry.Data["err"]; exists {
			if actualErr, ok := errField.(error); ok {
				err = errors.Wrap(err, actualErr.Error())
			}
		}
		if errField, exists := entry.Data["error"]; exists {
			if actualErr, ok := errField.(error); ok {
				err = errors.Wrap(err, actualErr.Error())
			}
		}

		sentry.CaptureException(err)
	}

	return nil
}

func (hook *SentryHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.WarnLevel,
		logrus.ErrorLevel,
		logrus.FatalLevel,
		logrus.PanicLevel,
	}
}

func CronSentryMiddleware() cronjob.Middleware {
	return func(job, action string, next cronjob.HandlerFunc) cronjob.HandlerFunc {
		return func(ctx context.Context) error {
			tracer := otel.Tracer("jobs-scheduler")
			jobCtx, span := tracer.Start(ctx, action, trace.WithSpanKind(trace.SpanKindInternal))
			defer span.End()

			hub := sentry.CurrentHub().Clone()
			jobCtx = sentry.SetHubOnContext(jobCtx, hub)
			hub.ConfigureScope(func(scope *sentry.Scope) {
				scope.SetTag("job", action)
				scope.SetTag("scheduler", job)
			})

			defer func() {
				if recoverErr := recover(); recoverErr != nil {
					span.RecordError(errors.Errorf("panic: %v", recoverErr))
					span.SetStatus(codes.Error, "panic")
					if hub := sentry.GetHubFromContext(jobCtx); hub != nil {
						hub.Recover(recoverErr)
					} else {
						sentry.CurrentHub().Recover(recoverErr)
					}
					panic(recoverErr)
				}
			}()

			err := next(jobCtx)
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
				if hub := sentry.GetHubFromContext(jobCtx); hub != nil {
					hub.CaptureException(err)
				} else {
					sentry.CaptureException(err)
				}
				return err
			}

			span.SetStatus(codes.Ok, "ok")
			return nil
		}
	}
}
