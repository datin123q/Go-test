package controllers

import (
	"net/http"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
	"github.com/stretchr/testify/assert"
)

// MIR-1415: 認証コードを確かめずにメールアドレスを書き換える古いルート(PUT /users/:id/email-change)を外した。
// メールアドレスの変更は、認証コードを照合してから書き換える今の経路(POST /api/v1/users-verify/email/check-code)で行う。
func TestUserApiController_Init_NoEmailChangeRoute(t *testing.T) {
	e := echo.New()
	UserApiController{}.Init(echoswagger.NewNop(e).Group("users", "/users"))

	methodsByPath := map[string][]string{}
	for _, r := range e.Routes() {
		methodsByPath[r.Path] = append(methodsByPath[r.Path], r.Method)
	}

	assert.NotContains(t, methodsByPath, "/users/:id/email-change", "古いメール変更のルートは登録しない")
	assert.Contains(t, methodsByPath["/users/email-code-check"], http.MethodPost, "同じグループのほかのルートは残る")
}
