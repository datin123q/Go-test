package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hillfolk/goutils/echomiddleware"
	"github.com/hillfolk/goutils/kafka"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/eventbus"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

// admin-apiのcontextDBWithNameへAfterCommit機構を組み込んだこと(MIR-1202)を検証する。
// user-api側(apps/user-api/server/server.go)には既に同じ形が入っており、
// admin-apiだけがこの仕組みを持たなかった。

func newContextDBAfterCommitTestEngine(t *testing.T) *xorm.Engine {
	t.Helper()
	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, engine.Close()) })
	return engine
}

func newContextDBAfterCommitTestEcho(engines ...*xorm.Engine) (*echo.Echo, []echo.MiddlewareFunc) {
	names := []echomiddleware.ContextDBType{factory.ConextDBType, factory.ConextMyNumberDBType}
	middlewares := make([]echo.MiddlewareFunc, 0, len(engines))
	for i, engine := range engines {
		middlewares = append(middlewares, contextDBWithName("test-service", names[i], engine, kafka.Config{}, map[eventbus.ContextEventBusType]eventbus.IEventBus{}))
	}
	return echo.New(), middlewares
}

func TestContextDBWithName_AfterCommit(t *testing.T) {
	t.Run("ハンドラが成功したら、コミット後にAfterCommit登録処理が実行される", func(t *testing.T) {
		engine := newContextDBAfterCommitTestEngine(t)
		e, middlewares := newContextDBAfterCommitTestEcho(engine)

		afterCommitCalled := false
		e.POST("/test", func(c echo.Context) error {
			return factory.AfterCommit(c.Request().Context(), func(context.Context) error {
				afterCommitCalled = true
				return nil
			})
		}, middlewares...)

		req := httptest.NewRequest(http.MethodPost, "/test", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.True(t, afterCommitCalled)
	})

	t.Run("ハンドラがエラーを返したら、ロールバックされAfterCommit登録処理は実行されない", func(t *testing.T) {
		engine := newContextDBAfterCommitTestEngine(t)
		e, middlewares := newContextDBAfterCommitTestEcho(engine)

		afterCommitCalled := false
		e.POST("/test", func(c echo.Context) error {
			if err := factory.AfterCommit(c.Request().Context(), func(context.Context) error {
				afterCommitCalled = true
				return nil
			}); err != nil {
				return err
			}
			return echo.NewHTTPError(http.StatusBadRequest, "fail")
		}, middlewares...)

		req := httptest.NewRequest(http.MethodPost, "/test", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.False(t, afterCommitCalled, "ハンドラが失敗したらAfterCommit登録処理は実行されないこと")
	})

	t.Run("ハンドラが500を返したら、ロールバックされAfterCommit登録処理は実行されない", func(t *testing.T) {
		engine := newContextDBAfterCommitTestEngine(t)
		e, middlewares := newContextDBAfterCommitTestEcho(engine)

		afterCommitCalled := false
		e.POST("/test", func(c echo.Context) error {
			if err := factory.AfterCommit(c.Request().Context(), func(context.Context) error {
				afterCommitCalled = true
				return nil
			}); err != nil {
				return err
			}
			return c.NoContent(http.StatusInternalServerError)
		}, middlewares...)

		req := httptest.NewRequest(http.MethodPost, "/test", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.False(t, afterCommitCalled, "500応答時はロールバックされAfterCommit登録処理は実行されないこと")
	})

	t.Run("2本のミドルウェア(通常DB・マイナンバーDB)を重ねても、AfterCommit登録処理は1回だけ実行される", func(t *testing.T) {
		mainEngine := newContextDBAfterCommitTestEngine(t)
		myNumberEngine := newContextDBAfterCommitTestEngine(t)
		e, middlewares := newContextDBAfterCommitTestEcho(mainEngine, myNumberEngine)

		calledCount := 0
		e.POST("/test", func(c echo.Context) error {
			return factory.AfterCommit(c.Request().Context(), func(context.Context) error {
				calledCount++
				return nil
			})
		}, middlewares...)

		req := httptest.NewRequest(http.MethodPost, "/test", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, 1, calledCount, "外側のミドルウェアだけがRunAfterCommitを実行し、二重実行されないこと")
	})

	t.Run("GETなど対象外メソッドでは、コミット処理自体を通らない", func(t *testing.T) {
		engine := newContextDBAfterCommitTestEngine(t)
		e, middlewares := newContextDBAfterCommitTestEcho(engine)

		handlerCalled := false
		e.GET("/test", func(c echo.Context) error {
			handlerCalled = true
			return c.NoContent(http.StatusOK)
		}, middlewares...)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.True(t, handlerCalled)
	})

	t.Run("session.Begin()が失敗したら500を返し、ハンドラは呼ばれない", func(t *testing.T) {
		engine := newContextDBAfterCommitTestEngine(t)
		require.NoError(t, engine.Close())
		e, middlewares := newContextDBAfterCommitTestEcho(engine)

		handlerCalled := false
		e.POST("/test", func(c echo.Context) error {
			handlerCalled = true
			return c.NoContent(http.StatusOK)
		}, middlewares...)

		req := httptest.NewRequest(http.MethodPost, "/test", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.False(t, handlerCalled, "Beginに失敗したら、そのままハンドラを続行しないこと")
	})

	t.Run("AfterCommit登録処理には素のxorm.Engineが渡り、Close済みセッションを掴まない", func(t *testing.T) {
		engine := newContextDBAfterCommitTestEngine(t)
		e, middlewares := newContextDBAfterCommitTestEcho(engine)

		var receivedDB any
		e.POST("/test", func(c echo.Context) error {
			return factory.AfterCommit(c.Request().Context(), func(afterCommitCtx context.Context) error {
				receivedDB = afterCommitCtx.Value(factory.ConextDBType)
				return nil
			})
		}, middlewares...)

		req := httptest.NewRequest(http.MethodPost, "/test", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		if _, ok := receivedDB.(*xorm.Engine); !ok {
			t.Fatalf("AfterCommit登録処理へ渡るctxのDBは*xorm.Engineであるべきだが、実際は%T", receivedDB)
		}
	})
}
