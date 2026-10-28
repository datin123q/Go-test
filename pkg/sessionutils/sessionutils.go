package sessionutils

import (
	"context"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/hillfolk/goutils/ctxaws"
	"github.com/labstack/echo/v4"
)

// Deprecated: Use AwsSession instead
func ContextAwsSes(session *session.Session) echo.MiddlewareFunc {

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

// Deprecated: Use AwsSession instead
func ContextAwsSns(session *session.Session) echo.MiddlewareFunc {

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
