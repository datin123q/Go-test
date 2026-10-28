package sessionutils

import (
	"context"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/hillfolk/goutils/ctxaws"
	"github.com/labstack/echo/v4"
)

var (
	ContextAwsSesName = "AWS_SES"
	ContextAwsSnsName = "AWS_SNS"
)

type AwsSession interface {
}

type awsSession struct{}

func New() AwsSession {
	return &awsSession{}
}

func (a awsSession) ContextAwsSes(session *session.Session) echo.MiddlewareFunc {

	ctxAws := ctxaws.New(session)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			ctx := req.Context()

			session := ctxAws.NewSession(ctx)

			c.SetRequest(req.WithContext(context.WithValue(ctx, ContextAwsSesName, session)))

			return next(c)
		}
	}
}

func (a awsSession) ContextAwsSns(session *session.Session) echo.MiddlewareFunc {

	ctxAws := ctxaws.New(session)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			ctx := req.Context()

			session := ctxAws.NewSession(ctx)

			c.SetRequest(req.WithContext(context.WithValue(ctx, ContextAwsSnsName, session)))

			return next(c)
		}
	}
}
