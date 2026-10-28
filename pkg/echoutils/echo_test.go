package echoutils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

// CustomErrorHandler の switch には default 節が無く、errors.Wrap で包んだだけの素のエラーは
// *factory.Error にも *echo.HTTPError にも一致しないため、レスポンスに何も書かれないまま関数を
// 抜けていた。呼び出し元には net/http の既定動作でステータス200・ボディ空のレスポンスが返り、
// 失敗が成功と区別できなくなっていた(MIR-1133)。default節を追加して500を明示的に返すことを
// 固定する。あわせて既存の *factory.Error / *echo.HTTPError の経路が従来どおりであることも
// 確認する。
func TestCustomErrorHandler(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   int
	}{
		{
			name:       "factory.Errorはそのステータス・コードで返る",
			err:        factory.ErrorParameter.New(errors.New("bad request")),
			wantStatus: http.StatusBadRequest,
			wantCode:   factory.ErrorParameter.Code,
		},
		{
			name:       "echo.HTTPErrorはそのコードで返る",
			err:        echo.NewHTTPError(http.StatusNotFound, "not found"),
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "factory.ErrorでもEcho.HTTPErrorでもない素のエラーはdefaultで500として返る",
			err:        errors.Wrap(errors.New("gmo api error"), "fail to call gmo api"),
			wantStatus: http.StatusInternalServerError,
			wantCode:   factory.ErrorSystem.Code,
		},
		{
			name:       "nilエラーはErrorSystemに差し替わり500で返る",
			err:        nil,
			wantStatus: http.StatusInternalServerError,
			wantCode:   factory.ErrorSystem.Code,
		},
		{
			name:       "typed-nilの*factory.Errorはerrorインターフェースとしてはnilでないためcase1に入るが、panicせず500で返る",
			err:        typedNilFactoryError(),
			wantStatus: http.StatusInternalServerError,
			wantCode:   factory.ErrorSystem.Code,
		},
		{
			name:       "typed-nilの*echo.HTTPErrorは事前チェックで弾かれ、panicせず500で返る",
			err:        typedNilHTTPError(),
			wantStatus: http.StatusInternalServerError,
			wantCode:   factory.ErrorSystem.Code,
		},
		{
			name:       "errors.Wrapで包んだtyped-nilの*echo.HTTPErrorも、WithError呼び出し前の事前チェックで検出され、panicせず500で返る",
			err:        errors.Wrap(typedNilHTTPError(), "context"),
			wantStatus: http.StatusInternalServerError,
			wantCode:   factory.ErrorSystem.Code,
		},
		{
			name:       "errors.Wrapで包んだtyped-nilの*factory.Errorは、Error()がnilセーフなためWithErrorはpanicせず、case1のrenderAPIError内のnilガードで500になる",
			err:        errors.Wrap(typedNilFactoryError(), "context"),
			wantStatus: http.StatusInternalServerError,
			wantCode:   factory.ErrorSystem.Code,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			assert.NotPanics(t, func() {
				CustomErrorHandler(tt.err, c)
			})

			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.NotEmpty(t, rec.Body.Bytes(), "レスポンスボディが空(=呼び出し元からは成功と区別できない)")

			if tt.wantCode != 0 {
				var body EchoApiResult
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
				assert.False(t, body.Success)
				assert.Equal(t, tt.wantCode, body.Error.Code)
			}
		})
	}
}

// typedNilFactoryError は、値がnilだが型情報を持つ*factory.Errorをerrorインターフェースに
// 入れて返す。errors.Asはこれを「一致した」と判定するため、CustomErrorHandlerのswitchの
// 1つ目のcaseに入るが、apiError自体はnilのままである。
func typedNilFactoryError() error {
	var apiError *factory.Error
	return apiError
}

// typedNilHTTPError は、値がnilだが型情報を持つ*echo.HTTPErrorをerrorインターフェースに
// 入れて返す。errors.Wrapで包んでも、errors.Asはラップの奥までこの値を見つけられる。
// echo.HTTPError.Error()はnilレシーバのフィールドを直接参照するため、CustomErrorHandlerが
// switchより前にerrors.Asで先に検出して除外しないと、behaviorlogのWithError呼び出しが
// err.Error()経由でpanicする。
func typedNilHTTPError() error {
	var httpError *echo.HTTPError
	return httpError
}

// TestRedactAozoraAuthQuery はMIR-1298の確認: アクセスログはURIをクエリごと
// 出す仕様だが、/gmo-aozora/auth/*(state・codeを含む)だけはクエリを落とす。
// 他のパスには影響しないこと、クエリが無いURIはそのままなことも確認する。
func TestRedactAozoraAuthQuery(t *testing.T) {
	tests := []struct {
		name string
		uri  string
		want string
	}{
		{
			name: "auth/loginのstateを落とす",
			uri:  "/gmo-aozora/auth/login?state=super-secret-state",
			want: "/gmo-aozora/auth/login",
		},
		{
			name: "auth/access-tokenのcode・stateを落とす",
			uri:  "/gmo-aozora/auth/access-token?code=abc&state=super-secret-state",
			want: "/gmo-aozora/auth/access-token",
		},
		{
			name: "クエリが無ければそのまま",
			uri:  "/gmo-aozora/auth/login",
			want: "/gmo-aozora/auth/login",
		},
		{
			name: "無関係なパスのクエリは落とさない",
			uri:  "/gmo-aozora/webhook?foo=bar",
			want: "/gmo-aozora/webhook?foo=bar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, redactAozoraAuthQuery(tt.uri))
		})
	}
}
