package apirender

import (
	"net/http"
	"reflect"

	"xorm.io/xorm"

	"github.com/labstack/echo/v4"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

const (
	MaxPinCount = 4
	Bearer      = "Bearer "
)

type ApiResult struct {
	Result  interface{}   `json:"result"`
	Success bool          `json:"success"`
	Error   factory.Error `json:"error"`
}

type ArrayResult struct {
	Items      interface{} `json:"items"`
	TotalCount int64       `json:"total_count"`
}

func RenderSucc(c echo.Context, status int, result interface{}) error {
	req := c.Request()
	if req.Method == "POST" || req.Method == "PUT" || req.Method == "DELETE" || req.Method == "PATCH" {
		if session, ok := factory.DB(req.Context()).(*xorm.Session); ok {

			// TODO テスト用に空のオブジェクトの場合はCommitを実行せずに返す
			// 該当条件はテストでのみ進入可能だがテスト用コードではないため修正が必要
			// このように別途既存のecho.Contextをラップするのではなく
			// ミドルウェアでエラーを返すか成功時に処理する形式でDBを処理する形式に変更が必要
			if reflect.DeepEqual(session, &xorm.Session{}) {
				return writeSucc(c, status, result)
			}

			err := session.Commit()
			if err != nil {
				return factory.ErrorDB.New(err)
			}
		}
	}

	return writeSucc(c, status, result)
}

// 204などの本文を持てない状態で本文を書くと、net/httpはErrBodyNotAllowedを返す。
// 応答は本文なしで届くが、ハンドラがエラーを返すため、DBのミドルウェアは失敗として
// 扱い、コミット後の処理(RunAfterCommit)を実行しない(MIR-1392)。
func writeSucc(c echo.Context, status int, result interface{}) error {
	if !statusAllowsBody(status) {
		return c.NoContent(status)
	}
	return c.JSON(status, ApiResult{
		Success: true,
		Result:  result,
	})
}

// net/httpのbodyAllowedForStatusと同じ条件。
func statusAllowsBody(status int) bool {
	switch {
	case status >= 100 && status <= 199:
		return false
	case status == http.StatusNoContent:
		return false
	case status == http.StatusNotModified:
		return false
	}
	return true
}
