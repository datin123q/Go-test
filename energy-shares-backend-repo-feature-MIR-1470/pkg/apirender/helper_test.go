package apirender

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func NewContext() echo.Context {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec)
}

func TestRenderSucc(t *testing.T) {
	type args struct {
		c      echo.Context
		status int
		result interface{}
	}
	tests := []struct {
		name     string
		args     args
		wantErr  bool
		Success  bool
		wantBody string
	}{
		{
			name: "status is 200",
			args: args{
				c:      NewContext(),
				status: 200,
				result: nil,
			},
			wantErr:  false,
			Success:  true,
			wantBody: "{\"result\":null,\"success\":true,\"error\":{}}\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := RenderSucc(tt.args.c, tt.args.status, tt.args.result); (err != nil) != tt.wantErr {
				t.Errorf("RenderSucc() error = %v, wantErr %v", err, tt.wantErr)
			}
			status := tt.args.c.Response().Status
			message := tt.args.c.Response().Writer.(*httptest.ResponseRecorder).Body.String()

			if status != http.StatusOK || message != tt.wantBody {
				{
					t.Errorf("fail to RenderSucc() status = %v message = %v", status, message)
				}
			}
		})
	}
}

func TestRenderSucc_StatusWithoutBody(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		wantBody string
	}{
		{name: "204は本文なしでエラーを返さない", status: http.StatusNoContent, wantBody: ""},
		{name: "304は本文なしでエラーを返さない", status: http.StatusNotModified, wantBody: ""},
		{name: "201は従来どおり本文を書く", status: http.StatusCreated, wantBody: "{\"result\":null,\"success\":true,\"error\":{}}\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewContext()

			err := RenderSucc(c, tt.status, nil)

			assert.NoError(t, err)
			assert.Equal(t, tt.status, c.Response().Status)
			assert.Equal(t, tt.wantBody, c.Response().Writer.(*httptest.ResponseRecorder).Body.String())
		})
	}
}

func Test_statusAllowsBody(t *testing.T) {
	tests := []struct {
		name   string
		status int
		want   bool
	}{
		{name: "99は本文を書ける", status: 99, want: true},
		{name: "100は本文を持てない", status: 100, want: false},
		{name: "101は本文を持てない", status: 101, want: false},
		{name: "199は本文を持てない", status: 199, want: false},
		{name: "200は本文を書ける", status: 200, want: true},
		{name: "201は本文を書ける", status: 201, want: true},
		{name: "203は本文を書ける", status: 203, want: true},
		{name: "204は本文を持てない", status: 204, want: false},
		{name: "205は本文を書ける", status: 205, want: true},
		{name: "303は本文を書ける", status: 303, want: true},
		{name: "304は本文を持てない", status: 304, want: false},
		{name: "305は本文を書ける", status: 305, want: true},
		{name: "400は本文を書ける", status: 400, want: true},
		{name: "500は本文を書ける", status: 500, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, statusAllowsBody(tt.status), "status=%d", tt.status)
		})
	}
}
