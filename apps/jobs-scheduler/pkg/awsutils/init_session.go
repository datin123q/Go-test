package awsutils

import (
	"context"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/hillfolk/goutils/cronjob"
	"github.com/hillfolk/goutils/ctxaws"
	"github.com/hillfolk/goutils/echomiddleware"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/sessionutils"
)

func ContextAws(session *session.Session) cronjob.Middleware {

	ctxaws := ctxaws.New(session)

	return func(job, action string, next cronjob.HandlerFunc) cronjob.HandlerFunc {
		return func(ctx context.Context) error {

			session := ctxaws.NewSession(ctx)
			ctx = context.WithValue(ctx, echomiddleware.ContextAWSName, session)

			return next(ctx)
		}
	}
}

func ContextAwsSes(session *session.Session) cronjob.Middleware {

	ctxaws := ctxaws.New(session)

	return func(job, action string, next cronjob.HandlerFunc) cronjob.HandlerFunc {
		return func(ctx context.Context) error {

			session := ctxaws.NewSession(ctx)
			ctx = context.WithValue(ctx, sessionutils.ContextAwsSesName, session)

			return next(ctx)
		}
	}
}

func ContextAwsSns(session *session.Session) cronjob.Middleware {

	ctxaws := ctxaws.New(session)

	return func(job, action string, next cronjob.HandlerFunc) cronjob.HandlerFunc {
		return func(ctx context.Context) error {

			session := ctxaws.NewSession(ctx)
			ctx = context.WithValue(ctx, sessionutils.ContextAwsSnsName, session)

			return next(ctx)
		}
	}
}
