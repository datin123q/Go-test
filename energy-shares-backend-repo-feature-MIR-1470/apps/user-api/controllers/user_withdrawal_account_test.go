package controllers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pangpanglabs/echoswagger/v2"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/echoutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/validator"
)

// MIR-1481: マイページで出金口座を変えても、取引停止・メルマガ同意・SMS 認証済み・
// ハガキ発送済みが外れない。
func TestUpdateWithdrawalAccount_KeepsFlags(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	require.NoError(t, err)
	engine.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, engine.Close()) })
	require.NoError(t, engine.Sync2(new(domain.User), new(domain.Role)))
	ctx := context.WithValue(context.Background(), factory.ConextDBType, engine)

	u := domain.User{Uid: "u10", Email: "u10@example.invalid", BlockInvest: true, IsNewsletter: true, IsSmsCert: true,
		PostalCodeStatus: true, AuditStatus: domain.COMPLETED_AUDIT, BankName: "旧銀行", BankAccount: "1111111"}
	_, err = engine.UseBool().Insert(&u)
	require.NoError(t, err)

	e := echo.New()
	e.Validator = validator.NewCustomValidator()
	e.HTTPErrorHandler = echoutils.CustomErrorHandler
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.SetRequest(c.Request().WithContext(ctx))
			c.Set("user", jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"user": strconv.FormatInt(u.Id, 10), "email": "u10@example.invalid",
				"role": []interface{}{"ROLE_USER"}, "audit_status": string(domain.COMPLETED_AUDIT)}))
			return next(c)
		}
	})
	UserApiController{}.Init(echoswagger.NewNop(e).Group("users", "/api/users"))

	body := `{"deposit_type":"ORDINARY_ACCOUNT","bank_name":"新銀行","branch_name":"新支店","bank_account":"2222222","bank_code":"0005","branch_code":"123"}`
	req := httptest.NewRequest(http.MethodPut, "/api/users/"+strconv.FormatInt(u.Id, 10)+"/withdrawal-account", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	var got domain.User
	_, err = engine.ID(u.Id).Get(&got)
	require.NoError(t, err)
	assert.True(t, got.BlockInvest, "取引停止が外れない")
	assert.True(t, got.IsNewsletter)
	assert.True(t, got.IsSmsCert)
	assert.True(t, got.PostalCodeStatus)
	assert.Equal(t, "新銀行", got.BankName)
	assert.Equal(t, "2222222", got.BankAccount)
	assert.Equal(t, lo.ToPtr("0005"), got.BankCode)
}
