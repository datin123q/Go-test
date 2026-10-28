package domain

import (
	"context"
	"fmt"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
)

func newSmsCertTestRequest(t *testing.T) (context.Context, *MobilePhoneVerificationRequest) {
	t.Helper()
	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	require.NoError(t, err)
	engine.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, engine.Close()) })
	require.NoError(t, engine.Sync(new(MobilePhoneVerificationRequest)))
	p := &MobilePhoneVerificationRequest{RequestId: "req-1", UserId: 10, Code: "123456", MobilePhone: "09012345678"}
	_, err = engine.Insert(p)
	require.NoError(t, err)
	return context.WithValue(context.Background(), factory.ConextDBType, engine), p
}

func withEnv(t *testing.T, env string) {
	t.Helper()
	prev := serviceutils.GetEnv()
	serviceutils.SetEnv(env)
	t.Cleanup(func() { serviceutils.SetEnv(prev) })
}

// MIR-1420: 本番では、申請した会員・申請ID・保存したコードがそろったときだけ通す。どの6桁でも、保存したコード以外は
// 通らない(固定のコードも含む)ことを全件で確かめる。環境の名前が想定と違う場合も本番と同じに扱う。
func TestMobilePhoneVerificationRequest_VerifyRequest_OnlyStoredCodeInProduction(t *testing.T) {
	for _, env := range []string{"prod", "PROD", "production", ""} {
		t.Run(fmt.Sprintf("env=%q", env), func(t *testing.T) {
			withEnv(t, env)
			ctx, p := newSmsCertTestRequest(t)
			for i := 0; i < 1000000; i++ {
				code := fmt.Sprintf("%06d", i)
				if code == p.Code {
					continue
				}
				if ok, err := p.VerifyRequest(ctx, 10, "req-1", code, time.Now()); ok || err == nil {
					t.Fatal("保存したコード以外の6桁が通った")
				}
			}

			ok, err := p.VerifyRequest(ctx, 11, "req-1", "123456", time.Now())
			assert.False(t, ok, "他の会員の申請では通らない")
			assert.Error(t, err)
			ok, err = p.VerifyRequest(ctx, 10, "req-2", "123456", time.Now())
			assert.False(t, ok, "申請IDが違えば通らない")
			assert.Error(t, err)
			ok, err = p.VerifyRequest(ctx, 10, "req-1", "123456", time.Now())
			assert.True(t, ok)
			assert.NoError(t, err)
		})
	}
}

// staging・開発環境では、SMSを受け取らずに試せるよう固定のコードで通す。
func TestMobilePhoneVerificationRequest_VerifyRequest_FixedCodeOnlyInDevelopment(t *testing.T) {
	for _, env := range []string{"stg", "dev", "qa", "local"} {
		t.Run(env, func(t *testing.T) {
			withEnv(t, env)
			ctx, p := newSmsCertTestRequest(t)
			ok, err := p.VerifyRequest(ctx, 10, "req-1", FixedSmsCode, time.Now())
			assert.True(t, ok)
			assert.NoError(t, err)
			ok, err = p.VerifyRequest(ctx, 10, "req-1", "654321", time.Now())
			assert.False(t, ok, "固定のコード以外の誤ったコードは通らない")
			assert.Error(t, err)
		})
	}
}
