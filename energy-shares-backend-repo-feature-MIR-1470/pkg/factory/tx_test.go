package factory

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
)

type txTestRecord struct {
	Id    int64  `xorm:"pk autoincr"`
	Value string `xorm:"varchar(255)"`
}

func (txTestRecord) TableName() string {
	return "tx_test_record"
}

func TestTxManager_WithTx(t *testing.T) {
	tests := []struct {
		name      string
		callback  func(context.Context) error
		wantCount int64
		wantErr   error
	}{
		{
			name: "処理成功時はコミットする",
			callback: func(ctx context.Context) error {
				_, err := DB(ctx).Insert(&txTestRecord{Value: "committed"})
				return err
			},
			wantCount: 1,
		},
		{
			name: "処理失敗時はロールバックする",
			callback: func(ctx context.Context) error {
				if _, err := DB(ctx).Insert(&txTestRecord{Value: "rolled back"}); err != nil {
					return err
				}
				return assert.AnError
			},
			wantCount: 0,
			wantErr:   assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine, err := xorm.NewEngine("sqlite3", filepath.Join(t.TempDir(), "transaction.db"))
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, engine.Close()) })
			require.NoError(t, engine.Sync(new(txTestRecord)))

			err = NewTxManager(engine).WithTx(context.Background(), tt.callback)

			if tt.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.True(t, errors.Is(err, tt.wantErr))
			}
			count, countErr := engine.Count(new(txTestRecord))
			require.NoError(t, countErr)
			assert.Equal(t, tt.wantCount, count)
		})
	}
}

func TestTxManager_AfterCommit(t *testing.T) {
	newEngine := func(t *testing.T) *xorm.Engine {
		t.Helper()

		engine, err := xorm.NewEngine("sqlite3", filepath.Join(t.TempDir(), "transaction.db"))
		require.NoError(t, err)
		t.Cleanup(func() { require.NoError(t, engine.Close()) })
		require.NoError(t, engine.Sync(new(txTestRecord)))
		return engine
	}

	t.Run("コミット成功後に処理を実行する", func(t *testing.T) {
		engine := newEngine(t)
		ctx := context.WithValue(context.Background(), ConextDBType, engine)
		afterCommitCalled := false

		err := NewTxManager(engine).WithTx(ctx, func(txCtx context.Context) error {
			if _, err := DB(txCtx).Insert(&txTestRecord{Value: "committed"}); err != nil {
				return err
			}
			return AfterCommit(txCtx, func(afterCommitCtx context.Context) error {
				afterCommitCalled = true
				count, err := DB(afterCommitCtx).Count(new(txTestRecord))
				require.NoError(t, err)
				assert.Equal(t, int64(1), count)
				return nil
			})
		})

		assert.NoError(t, err)
		assert.True(t, afterCommitCalled)
	})

	t.Run("ロールバック時は処理を実行しない", func(t *testing.T) {
		engine := newEngine(t)
		ctx := context.WithValue(context.Background(), ConextDBType, engine)
		afterCommitCalled := false

		err := NewTxManager(engine).WithTx(ctx, func(txCtx context.Context) error {
			if _, err := DB(txCtx).Insert(&txTestRecord{Value: "rolled back"}); err != nil {
				return err
			}
			if err := AfterCommit(txCtx, func(context.Context) error {
				afterCommitCalled = true
				return nil
			}); err != nil {
				return err
			}
			return assert.AnError
		})

		assert.ErrorIs(t, err, assert.AnError)
		assert.False(t, afterCommitCalled)
		count, countErr := engine.Count(new(txTestRecord))
		require.NoError(t, countErr)
		assert.Equal(t, int64(0), count)
	})

	t.Run("コミット後処理の失敗時もDB更新はコミット済み", func(t *testing.T) {
		engine := newEngine(t)
		ctx := context.WithValue(context.Background(), ConextDBType, engine)
		secondAfterCommitCalled := false

		err := NewTxManager(engine).WithTx(ctx, func(txCtx context.Context) error {
			if _, err := DB(txCtx).Insert(&txTestRecord{Value: "committed"}); err != nil {
				return err
			}
			if err := AfterCommit(txCtx, func(context.Context) error {
				return assert.AnError
			}); err != nil {
				return err
			}
			return AfterCommit(txCtx, func(context.Context) error {
				secondAfterCommitCalled = true
				return nil
			})
		})

		assert.ErrorIs(t, err, assert.AnError)
		assert.True(t, secondAfterCommitCalled)
		count, countErr := engine.Count(new(txTestRecord))
		require.NoError(t, countErr)
		assert.Equal(t, int64(1), count)
	})
}

func TestAfterCommit_OutsideTransaction(t *testing.T) {
	called := false

	err := AfterCommit(context.Background(), func(context.Context) error {
		called = true
		return nil
	})

	assert.NoError(t, err)
	assert.True(t, called)
}

func TestAfterCommit_WithAfterCommitContext(t *testing.T) {
	ctx, ownsAfterCommit := WithAfterCommit(context.Background())
	require.True(t, ownsAfterCommit)

	sameCtx, ownsAfterCommit := WithAfterCommit(ctx)
	require.False(t, ownsAfterCommit)
	assert.Equal(t, ctx, sameCtx)

	called := false
	err := AfterCommit(ctx, func(context.Context) error {
		called = true
		return nil
	})
	require.NoError(t, err)
	assert.False(t, called)

	require.NoError(t, RunAfterCommit(ctx, context.Background()))
	assert.True(t, called)
}

func TestRunAfterCommit_RunsOnlyOnce(t *testing.T) {
	ctx, _ := WithAfterCommit(context.Background())

	var called int
	require.NoError(t, AfterCommit(ctx, func(context.Context) error {
		called++
		return nil
	}))

	require.NoError(t, RunAfterCommit(ctx, context.Background()))
	require.NoError(t, RunAfterCommit(ctx, context.Background()))
	assert.Equal(t, 1, called)
}

// MIR-1318 設計v2で発見: 確定後キューの実行中、またはその後にさらにAfterCommitを
// 呼ぶと、キューはもう二度と回らないため誰にも実行されず黙って捨てられていた
// (afterCommitStore.run は ran=true を先に立ててから s.funcs を回すため、実行中に
// store.add されたものは今回のループに入らず、次にRunAfterCommitを呼んでも
// ran==true で即return nilになる)。この入れ子登録を「その場で実行する」形に
// 直したことを確認する。

func TestAfterCommitStore_NestedRegistrationDuringRun_RunsImmediately(t *testing.T) {
	ctx, _ := WithAfterCommit(context.Background())

	outerCalled := false
	innerCalled := false

	require.NoError(t, AfterCommit(ctx, func(runCtx context.Context) error {
		outerCalled = true
		// 実行中(ran=trueが立った直後)にさらに登録する。
		return AfterCommit(runCtx, func(context.Context) error {
			innerCalled = true
			return nil
		})
	}))

	require.NoError(t, RunAfterCommit(ctx, context.Background()))

	assert.True(t, outerCalled)
	assert.True(t, innerCalled, "実行中に登録された処理も、その場で実行されて消えないこと")

	// 二重実行にもなっていないことを確認する。
	innerCalled = false
	require.NoError(t, RunAfterCommit(ctx, context.Background()))
	assert.False(t, innerCalled, "2回目のRunAfterCommitでは何も再実行されないこと")
}

func TestAfterCommitStore_RegistrationAfterRun_RunsImmediately(t *testing.T) {
	ctx, _ := WithAfterCommit(context.Background())
	require.NoError(t, RunAfterCommit(ctx, context.Background()))

	called := false
	// run済みのstoreへ、実行が終わったあとに登録する。
	require.NoError(t, AfterCommit(ctx, func(context.Context) error {
		called = true
		return nil
	}))

	assert.True(t, called, "run済みのstoreへ登録された処理も、その場で実行されて消えないこと")
}

// その場で実行される入れ子登録のパニックは、通常の実行経路(run)と同じく
// 変換せずそのまま伝播する(②レビューで指摘: パニックをエラーへ変換すると
// 通常の実行経路にまで影響する一歩大きい変更になるため、今回のスコープには
// 含めず、入れ子登録のときだけ挙動を変えることはしない)。
func TestAfterCommitStore_NestedRegistration_PanicPropagates(t *testing.T) {
	ctx, _ := WithAfterCommit(context.Background())
	require.NoError(t, RunAfterCommit(ctx, context.Background()))

	assert.Panics(t, func() {
		_ = AfterCommit(ctx, func(context.Context) error {
			panic("boom")
		})
	}, "その場で実行された処理のpanicは、通常の実行経路と同じく呼び出し元へ伝播すること")
}

// tc.runAfterCommit はコールバックに tc.parentContext (*TxContext ではない生のctx) を
// 渡すため、確定後コールバックが「受け取ったctx」経由でAfterCommitを呼び直しても
// afterCommitStore 側の分岐にしか乗らない。tc.addAfterCommit の committed 分岐へ
// 実際に乗るのは、呼び出し側が元の *TxContext への参照を保持していて、確定後に
// それへ向けて直接 AfterCommit(txCtx, ...) を呼び直す経路であり、下記はその経路を
// 直接再現する。

func TestTxContext_NestedRegistrationAfterCommit_RunsImmediately(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite3", filepath.Join(t.TempDir(), "transaction.db"))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, engine.Close()) })
	require.NoError(t, engine.Sync(new(txTestRecord)))
	ctx := context.WithValue(context.Background(), ConextDBType, engine)

	var capturedTxCtx context.Context
	err = NewTxManager(engine).WithTx(ctx, func(txCtx context.Context) error {
		capturedTxCtx = txCtx
		_, insertErr := DB(txCtx).Insert(&txTestRecord{Value: "committed"})
		return insertErr
	})
	require.NoError(t, err)

	called := false
	// 確定(Commit)が既に済んだ *TxContext へ、保持していた参照から登録し直す。
	require.NoError(t, AfterCommit(capturedTxCtx, func(context.Context) error {
		called = true
		return nil
	}))

	assert.True(t, called, "確定済みのTxContextへ登録された処理も、その場で実行されて消えないこと")
}

func TestTxContext_NestedRegistrationAfterCommit_PanicPropagates(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite3", filepath.Join(t.TempDir(), "transaction.db"))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, engine.Close()) })
	require.NoError(t, engine.Sync(new(txTestRecord)))
	ctx := context.WithValue(context.Background(), ConextDBType, engine)

	var capturedTxCtx context.Context
	err = NewTxManager(engine).WithTx(ctx, func(txCtx context.Context) error {
		capturedTxCtx = txCtx
		return nil
	})
	require.NoError(t, err)

	assert.Panics(t, func() {
		_ = AfterCommit(capturedTxCtx, func(context.Context) error {
			panic("boom")
		})
	}, "その場で実行された処理のpanicは、通常の実行経路と同じく呼び出し元へ伝播すること")
}

func TestEngine(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite3", filepath.Join(t.TempDir(), "engine.db"))
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, engine.Close())
	})

	t.Run("Engineを取得できる", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), ConextDBType, engine)
		got, err := Engine(ctx)
		require.NoError(t, err)
		assert.Equal(t, engine, got)
	})

	t.Run("Sessionから親Engineを取得できる", func(t *testing.T) {
		session := engine.NewSession()
		t.Cleanup(func() { session.Close() })

		ctx := context.WithValue(context.Background(), ConextDBType, session)
		got, err := Engine(ctx)
		require.NoError(t, err)
		assert.Equal(t, engine, got)
	})

	t.Run("DBが無い場合はエラーを返す", func(t *testing.T) {
		_, err := Engine(context.Background())
		require.Error(t, err)
	})
}
