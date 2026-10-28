package domain

import (
	"context"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

func setupGmoAozoraOAuthStateTestDB(t *testing.T) context.Context {
	t.Helper()

	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	require.NoError(t, err)
	engine.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, engine.Close()) })

	require.NoError(t, engine.Sync2(new(GmoAozoraOAuthState)))

	return context.WithValue(context.Background(), factory.ConextDBType, engine)
}

func TestGmoAozoraOAuthState_CreateAndGetByLookup(t *testing.T) {
	ctx := setupGmoAozoraOAuthStateTestDB(t)

	lookup := Sha256Hex("S-value")
	expiresAt := time.Now().UTC().Add(30 * time.Minute)
	require.NoError(t, (GmoAozoraOAuthState{}).Create(ctx, lookup, 42, expiresAt))

	row, err := (GmoAozoraOAuthState{}).GetByLookup(ctx, lookup)
	require.NoError(t, err)
	require.NotNil(t, row)
	assert.Equal(t, int64(42), row.AdminUserId)
	assert.Nil(t, row.StateHash)
	assert.Nil(t, row.UsedAt)
	assert.WithinDuration(t, expiresAt, row.ExpiresAt, time.Second)
}

func TestGmoAozoraOAuthState_GetByLookup_NotFound(t *testing.T) {
	ctx := setupGmoAozoraOAuthStateTestDB(t)

	row, err := (GmoAozoraOAuthState{}).GetByLookup(ctx, "does-not-exist")
	require.NoError(t, err)
	assert.Nil(t, row)
}

// TestGmoAozoraOAuthState_SetStateHashOnce_WritesOnlyOnce は、GetLoginが同じ
// state行で2回呼ばれても、2回目はstate_hash/nonceを書き戻さないことを確認する
// (MIR-1298)。2回目まで書き戻すと、1回目に発行したURLでのログインの照合が
// 壊れる。
func TestGmoAozoraOAuthState_SetStateHashOnce_WritesOnlyOnce(t *testing.T) {
	ctx := setupGmoAozoraOAuthStateTestDB(t)

	lookup := Sha256Hex("S-value")
	require.NoError(t, (GmoAozoraOAuthState{}).Create(ctx, lookup, 1, time.Now().UTC().Add(time.Hour)))
	row, err := (GmoAozoraOAuthState{}).GetByLookup(ctx, lookup)
	require.NoError(t, err)
	require.NotNil(t, row)

	updated, err := (GmoAozoraOAuthState{}).SetStateHashOnce(ctx, row.Id, "hash-1", "nonce-1")
	require.NoError(t, err)
	assert.True(t, updated, "1回目は書き戻せること")

	updated, err = (GmoAozoraOAuthState{}).SetStateHashOnce(ctx, row.Id, "hash-2", "nonce-2")
	require.NoError(t, err)
	assert.False(t, updated, "2回目は書き戻さないこと")

	reloaded, err := (GmoAozoraOAuthState{}).GetByLookup(ctx, lookup)
	require.NoError(t, err)
	require.NotNil(t, reloaded)
	require.NotNil(t, reloaded.StateHash)
	require.NotNil(t, reloaded.Nonce)
	assert.Equal(t, "hash-1", *reloaded.StateHash, "1回目の値のまま残ること")
	assert.Equal(t, "nonce-1", *reloaded.Nonce)
}

// TestGmoAozoraOAuthState_ClaimByStateHash はstateの未使用・期限切れ・使用済み・
// 正常の4パターンを確認する(MIR-1298の5節「テスト」)。
func TestGmoAozoraOAuthState_ClaimByStateHash(t *testing.T) {
	now := time.Now().UTC()

	t.Run("未発行のstate_hashはnilを返す", func(t *testing.T) {
		ctx := setupGmoAozoraOAuthStateTestDB(t)
		row, err := (GmoAozoraOAuthState{}).ClaimByStateHash(ctx, "unknown-hash", now)
		require.NoError(t, err)
		assert.Nil(t, row)
	})

	t.Run("期限切れは弾く", func(t *testing.T) {
		ctx := setupGmoAozoraOAuthStateTestDB(t)
		lookup := Sha256Hex("S-expired")
		require.NoError(t, (GmoAozoraOAuthState{}).Create(ctx, lookup, 1, now.Add(-time.Minute)))
		created, err := (GmoAozoraOAuthState{}).GetByLookup(ctx, lookup)
		require.NoError(t, err)
		_, err = (GmoAozoraOAuthState{}).SetStateHashOnce(ctx, created.Id, "hash-expired", "nonce")
		require.NoError(t, err)

		row, err := (GmoAozoraOAuthState{}).ClaimByStateHash(ctx, "hash-expired", now)
		require.NoError(t, err)
		assert.Nil(t, row)
	})

	t.Run("正常な行は1回だけ使用済みにできる", func(t *testing.T) {
		ctx := setupGmoAozoraOAuthStateTestDB(t)
		lookup := Sha256Hex("S-ok")
		require.NoError(t, (GmoAozoraOAuthState{}).Create(ctx, lookup, 7, now.Add(30*time.Minute)))
		created, err := (GmoAozoraOAuthState{}).GetByLookup(ctx, lookup)
		require.NoError(t, err)
		_, err = (GmoAozoraOAuthState{}).SetStateHashOnce(ctx, created.Id, "hash-ok", "nonce-ok")
		require.NoError(t, err)

		claimed, err := (GmoAozoraOAuthState{}).ClaimByStateHash(ctx, "hash-ok", now)
		require.NoError(t, err)
		require.NotNil(t, claimed)
		assert.Equal(t, int64(7), claimed.AdminUserId)
		require.NotNil(t, claimed.UsedAt)

		// 同じstateで2回目(F5・二重送信)は使用済みなので弾く。
		second, err := (GmoAozoraOAuthState{}).ClaimByStateHash(ctx, "hash-ok", now)
		require.NoError(t, err)
		assert.Nil(t, second, "使用済みのstateを2回claimできないこと")
	})
}

func TestGmoAozoraOAuthState_RecordResultAndReload(t *testing.T) {
	ctx := setupGmoAozoraOAuthStateTestDB(t)
	now := time.Now().UTC()

	lookup := Sha256Hex("S-result")
	require.NoError(t, (GmoAozoraOAuthState{}).Create(ctx, lookup, 3, now.Add(time.Hour)))
	created, err := (GmoAozoraOAuthState{}).GetByLookup(ctx, lookup)
	require.NoError(t, err)
	_, err = (GmoAozoraOAuthState{}).SetStateHashOnce(ctx, created.Id, "hash-result", "nonce")
	require.NoError(t, err)

	claimed, err := (GmoAozoraOAuthState{}).ClaimByStateHash(ctx, "hash-result", now)
	require.NoError(t, err)
	require.NotNil(t, claimed)

	require.NoError(t, (GmoAozoraOAuthState{}).RecordResult(ctx, claimed.Id, GmoAozoraOAuthResultSuccess))

	reloaded, err := (GmoAozoraOAuthState{}).GetByStateHash(ctx, "hash-result")
	require.NoError(t, err)
	require.NotNil(t, reloaded)
	require.NotNil(t, reloaded.Result)
	assert.Equal(t, GmoAozoraOAuthResultSuccess, *reloaded.Result)
	require.NotNil(t, reloaded.UsedAt)
}
