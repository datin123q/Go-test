package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hillfolk/goutils/jwtutil"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
)

// MIR-1407: 会員のトークン更新では、roleテーブルに同じIDの管理者権限があっても、ROLE_USERだけを載せる。
func TestAuthApiController_UpdateToken_OnlyRoleUser(t *testing.T) {
	prevEnv := serviceutils.GetEnv()
	serviceutils.SetEnv("LOCAL") // CloudFrontの署名付きクッキーを作らない
	jwtutil.SetJwtSecret("test-secret")
	t.Cleanup(func() { serviceutils.SetEnv(prevEnv) })

	tests := []struct {
		name  string
		roles []string // roleテーブルに、会員と同じIDで入っている権限
	}{
		{name: "同じIDの管理者がROLE_PAYMENTを持つ", roles: []string{"ROLE_PAYMENT"}},
		{name: "同じIDの管理者がROLE_SUPER_ADMIN・ROLE_MYNUMBER・ROLE_ADMINを持つ", roles: []string{"ROLE_SUPER_ADMIN", "ROLE_MYNUMBER", "ROLE_ADMIN"}},
		{name: "roleテーブルに行が無い", roles: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine, err := xorm.NewEngine("sqlite3", ":memory:")
			require.NoError(t, err)
			engine.SetMaxOpenConns(1)
			t.Cleanup(func() { require.NoError(t, engine.Close()) })
			require.NoError(t, engine.Sync(new(domain.User), new(domain.Role), new(domain.RefreshToken)))

			_, err = engine.Insert(&domain.User{Id: 7, Email: "member@example.invalid", UserType: domain.PRIVATE})
			require.NoError(t, err)
			for _, code := range tt.roles {
				_, err = engine.Insert(&domain.Role{UserId: 7, RoleCode: code})
				require.NoError(t, err)
			}
			_, err = engine.Insert(&domain.RefreshToken{
				UserId: 7, UserType: domain.PRIVATE, Token: "refresh-token-7", IsActive: true,
				ExpireIn: time.Now().Add(time.Hour).Unix(), IpAddr: "192.0.2.1", Agent: "test-agent",
			})
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/auth/update-token", nil)
			req.Header.Set("User-Agent", "test-agent")
			req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "refresh-token-7"})
			req = req.WithContext(context.WithValue(req.Context(), factory.ConextDBType, engine))
			rec := httptest.NewRecorder()
			c := echo.New().NewContext(req, rec)

			require.NoError(t, AuthApiController{}.UpdateToken(c))
			require.Equal(t, http.StatusOK, rec.Code)

			var res struct {
				Result struct {
					AccessToken string `json:"access_token"`
				} `json:"result"`
			}
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
			claims, err := jwtutil.Extract(res.Result.AccessToken)
			require.NoError(t, err)

			assert.Equal(t, "7", claims["user"])
			assert.Equal(t, []interface{}{"ROLE_USER"}, claims["role"], "管理者の権限が入っていないこと")
		})
	}
}

func TestUserForRefreshedToken(t *testing.T) {
	in := domain.User{Id: 3, Email: "a@example.invalid", Name: "n", UserType: domain.CORPORATE, Roles: []string{"ROLE_PAYMENT", "ROLE_SUPER_ADMIN"}}

	got := userForRefreshedToken(in)

	assert.Equal(t, []string{"ROLE_USER"}, got.Roles)
	assert.Equal(t, in.Id, got.Id)
	assert.Equal(t, in.Email, got.Email)
	assert.Equal(t, in.UserType, got.UserType)
}

// 会員のトークン更新は、会員(PRIVATE/CORPORATE)のrefresh_tokenだけを受け付ける。管理者のrefresh_tokenは、
// 同じIDの会員がいても拒否する。
func TestAuthApiController_UpdateToken_MemberRefreshTokenOnly(t *testing.T) {
	prevEnv := serviceutils.GetEnv()
	serviceutils.SetEnv("LOCAL") // CloudFrontの署名付きクッキーを作らない
	jwtutil.SetJwtSecret("test-secret")
	t.Cleanup(func() { serviceutils.SetEnv(prevEnv) })

	call := func(t *testing.T, userType domain.USER_TYPE) (*httptest.ResponseRecorder, error) {
		engine, err := xorm.NewEngine("sqlite3", ":memory:")
		require.NoError(t, err)
		engine.SetMaxOpenConns(1)
		t.Cleanup(func() { require.NoError(t, engine.Close()) })
		require.NoError(t, engine.Sync(new(domain.User), new(domain.Role), new(domain.RefreshToken)))

		// 会員7がいて、user_idが同じ7のrefresh_tokenを、種別を変えて試す。
		_, err = engine.Insert(&domain.User{Id: 7, Email: "member@example.invalid", UserType: domain.PRIVATE})
		require.NoError(t, err)
		_, err = engine.Insert(&domain.RefreshToken{
			UserId: 7, UserType: userType, Token: "rt", IsActive: true,
			ExpireIn: time.Now().Add(time.Hour).Unix(), IpAddr: "192.0.2.1", Agent: "test-agent",
		})
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/auth/update-token", nil)
		req.Header.Set("User-Agent", "test-agent")
		req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "rt"})
		req = req.WithContext(context.WithValue(req.Context(), factory.ConextDBType, engine))
		rec := httptest.NewRecorder()
		return rec, AuthApiController{}.UpdateToken(echo.New().NewContext(req, rec))
	}

	for _, userType := range []domain.USER_TYPE{domain.ADMIN, domain.SYSTEM, ""} {
		t.Run("拒否: refresh_tokenのuser_typeが"+string(userType)+"(同じIDの会員がいても)", func(t *testing.T) {
			_, err := call(t, userType)

			require.Error(t, err)
			factoryErr, ok := err.(*factory.Error)
			require.True(t, ok, "error type = %T", err)
			assert.Equal(t, factory.ErrorTokenInvalid.Code, factoryErr.Code)
		})
	}

	for _, userType := range []domain.USER_TYPE{domain.PRIVATE, domain.CORPORATE} {
		t.Run("受け付ける: refresh_tokenのuser_typeが"+string(userType), func(t *testing.T) {
			rec, err := call(t, userType)

			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, rec.Code)
		})
	}
}
