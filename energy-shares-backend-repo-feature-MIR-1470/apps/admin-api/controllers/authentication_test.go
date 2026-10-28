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

// MIR-1407: 管理者のトークン更新は、管理者(user_type=ADMIN)のrefresh_tokenだけを受け付ける。
func TestAuthApiController_UpdateToken_AdminRefreshTokenOnly(t *testing.T) {
	prevEnv := serviceutils.GetEnv()
	serviceutils.SetEnv("LOCAL") // CloudFrontの署名付きクッキーを作らない
	jwtutil.SetJwtSecret("test-secret")
	t.Cleanup(func() { serviceutils.SetEnv(prevEnv) })

	// 管理者7(強い権限を持つ)がいて、user_idが同じ7のrefresh_tokenを、会員の種別などで作って試す。
	setup := func(t *testing.T) context.Context {
		engine, err := xorm.NewEngine("sqlite3", ":memory:")
		require.NoError(t, err)
		engine.SetMaxOpenConns(1)
		t.Cleanup(func() { require.NoError(t, engine.Close()) })
		require.NoError(t, engine.Sync(new(domain.AdminUser), new(domain.Role), new(domain.RefreshToken)))

		_, err = engine.Table("admin_user").Insert(&domain.AdminUser{Id: 7, Email: "admin@example.invalid", Name: "admin"})
		require.NoError(t, err)
		for _, code := range []string{"ROLE_PAYMENT", "ROLE_SUPER_ADMIN"} {
			_, err = engine.Insert(&domain.Role{UserId: 7, RoleCode: code})
			require.NoError(t, err)
		}
		return context.WithValue(context.Background(), factory.ConextDBType, engine)
	}
	insertRefreshToken := func(t *testing.T, ctx context.Context, userType domain.USER_TYPE, token string) {
		_, err := ctx.Value(factory.ConextDBType).(*xorm.Engine).Insert(&domain.RefreshToken{
			UserId: 7, UserType: userType, Token: token, IsActive: true,
			ExpireIn: time.Now().Add(time.Hour).Unix(), IpAddr: "192.0.2.1", Agent: "test-agent",
		})
		require.NoError(t, err)
	}
	call := func(ctx context.Context, token string) (*httptest.ResponseRecorder, error) {
		req := httptest.NewRequest(http.MethodPost, "/api/auth/update-token", nil)
		req.Header.Set("User-Agent", "test-agent")
		req.AddCookie(&http.Cookie{Name: "refresh_token", Value: token})
		rec := httptest.NewRecorder()
		c := echo.New().NewContext(req.WithContext(ctx), rec)
		return rec, AuthApiController{}.UpdateToken(c)
	}

	for _, userType := range []domain.USER_TYPE{domain.PRIVATE, domain.CORPORATE, domain.SYSTEM, ""} {
		t.Run("拒否: refresh_tokenのuser_typeが"+string(userType)+"(同じIDの管理者がいても)", func(t *testing.T) {
			ctx := setup(t)
			insertRefreshToken(t, ctx, userType, "rt")

			_, err := call(ctx, "rt")

			require.Error(t, err)
			factoryErr, ok := err.(*factory.Error)
			require.True(t, ok, "error type = %T", err)
			assert.Equal(t, factory.ErrorTokenInvalid.Code, factoryErr.Code)
		})
	}

	t.Run("管理者のrefresh_tokenは通り、その管理者の権限のトークンが出る", func(t *testing.T) {
		ctx := setup(t)
		insertRefreshToken(t, ctx, domain.ADMIN, "admin-rt")

		rec, err := call(ctx, "admin-rt")

		require.NoError(t, err)
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
		assert.ElementsMatch(t, []interface{}{"ROLE_PAYMENT", "ROLE_SUPER_ADMIN"}, claims["role"])
	})
}
