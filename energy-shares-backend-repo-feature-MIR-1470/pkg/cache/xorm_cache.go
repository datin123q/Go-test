package cache

import (
	"context"

	"github.com/labstack/echo/v4"
	"xorm.io/xorm"
)

var (
	XormEngine string = "XORM_ENGINE"
)

func ContextXormEngine(xormEngine *xorm.Engine) echo.MiddlewareFunc {

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			ctx := req.Context()

			c.SetRequest(req.WithContext(context.WithValue(ctx, XormEngine, xormEngine)))

			return next(c)
		}
	}
}

func GetObject(ctx context.Context, tableName, id string) interface{} {

	if engine, ok := ctx.Value(XormEngine).(*xorm.Engine); ok {

		cacher := engine.GetDefaultCacher()
		if cacher == nil {
			return nil
		}

		return cacher.GetBean(tableName, id)
	}

	return nil
}

func PutObject(ctx context.Context, tableName, id string, obj interface{}) {

	if engine, ok := ctx.Value(XormEngine).(*xorm.Engine); ok {
		cacher := engine.GetDefaultCacher()
		cacher.PutBean(tableName, id, obj)
	}
}

func DelObject(ctx context.Context, tableName, id string) {

	if engine, ok := ctx.Value(XormEngine).(*xorm.Engine); ok {
		cacher := engine.GetDefaultCacher()
		cacher.DelBean(tableName, id)
	}
}

func ClearObjects(ctx context.Context, tableName string) {

	if engine, ok := ctx.Value(XormEngine).(*xorm.Engine); ok {
		cacher := engine.GetDefaultCacher()
		cacher.ClearBeans(tableName)
	}
}

func GetSqlObject(ctx context.Context, tableName, sql string) interface{} {

	if engine, ok := ctx.Value(XormEngine).(*xorm.Engine); ok {
		cacher := engine.GetDefaultCacher()
		return cacher.GetIds(tableName, sql)
	}

	return nil
}

func PutSqlObject(ctx context.Context, tableName, sql string, obj interface{}) {

	if engine, ok := ctx.Value(XormEngine).(*xorm.Engine); ok {
		cacher := engine.GetDefaultCacher()
		cacher.PutIds(tableName, sql, obj)
	}
}

func DelSqlObject(ctx context.Context, tableName, sql string) {

	if engine, ok := ctx.Value(XormEngine).(*xorm.Engine); ok {
		cacher := engine.GetDefaultCacher()
		cacher.DelIds(tableName, sql)
	}
}

func ClearSqlObjects(ctx context.Context, tableName string) {

	if engine, ok := ctx.Value(XormEngine).(*xorm.Engine); ok {
		cacher := engine.GetDefaultCacher()
		cacher.ClearIds(tableName)
	}
}
