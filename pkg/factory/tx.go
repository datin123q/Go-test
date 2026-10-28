package factory

import (
	"context"
	"errors"

	"xorm.io/xorm"
)

type TxContext struct {
	context.Context
	parentContext    context.Context
	session          *xorm.Session
	committed        bool
	afterCommitFuncs []func(context.Context) error
}

type afterCommitContextKey struct{}

type afterCommitStore struct {
	funcs []func(context.Context) error
	ran   bool
}

// addは、確定後キューへfnを登録する。既にrun済み(s.ran)の状態で呼ばれた場合、
// 登録してもキューはもう回らないため誰にも実行されず黙って捨てられていた
// (MIR-1318 設計v2で発見。確定後の処理の実行中にさらにAfterCommitを呼ぶと
// このケースに当たる)。その場で実行することで、「確定後に実行される」という
// 約束を守りつつ、処理が消えることを防ぐ。想定外の経路のため、気付けるよう
// Warnログを残す。パニックは通常の実行経路(run)と同じく変換せずそのまま
// 伝播させる(②レビューで指摘: パニックをエラーへ変換する案は、この入れ子
// 登録のときだけでなく通常の実行経路にまで影響する一歩大きい変更になるため、
// 今回のスコープには含めない)。
func (s *afterCommitStore) add(ctx context.Context, fn func(context.Context) error) {
	if s.ran {
		Logger(ctx).Warn("factory: AfterCommit registered after the after-commit queue already ran; running it immediately")
		if err := fn(ctx); err != nil {
			Logger(ctx).WithError(err).Error("factory: after-commit callback (run immediately due to late registration) failed")
		}
		return
	}
	s.funcs = append(s.funcs, fn)
}

func (s *afterCommitStore) run(ctx context.Context) error {
	if s.ran {
		return nil
	}
	s.ran = true

	var afterCommitErrors []error
	for _, fn := range s.funcs {
		if err := fn(ctx); err != nil {
			afterCommitErrors = append(afterCommitErrors, err)
		}
	}
	return errors.Join(afterCommitErrors...)
}

// WithAfterCommit は任意のトランザクション境界で使う after-commit キューを context に設定する。
// 既に設定済みの場合は同じ context と false を返す。
func WithAfterCommit(ctx context.Context) (context.Context, bool) {
	if _, ok := ctx.Value(afterCommitContextKey{}).(*afterCommitStore); ok {
		return ctx, false
	}
	return context.WithValue(ctx, afterCommitContextKey{}, &afterCommitStore{}), true
}

// RunAfterCommit は WithAfterCommit で登録された処理を実行する。
func RunAfterCommit(ctx, callbackCtx context.Context) error {
	store, ok := ctx.Value(afterCommitContextKey{}).(*afterCommitStore)
	if !ok {
		return nil
	}
	return store.run(callbackCtx)
}

func (tc *TxContext) Commit() error {
	if tc.committed {
		return nil
	}
	if err := tc.session.Commit(); err != nil {
		return err
	}

	tc.committed = true
	return nil
}

func (tc *TxContext) Rollback() error {
	if !tc.committed {
		return tc.session.Rollback()
	}
	return nil
}

func (tc *TxContext) Close() {
	if !tc.committed {
		tc.session.Rollback()
	}
	tc.session.Close()
}

// addAfterCommitは、確定後キューへfnを登録する。既にcommitted済みの状態で
// 呼ばれた場合、登録してもキューはもう回らないため誰にも実行されず黙って
// 捨てられていた(afterCommitStore.addと同じ理由、MIR-1318 設計v2)。その場で
// 実行することで約束を守る。実行に使うctxはrunAfterCommitと同じ
// tc.parentContext(トランザクションスコープのctxではなく、確定後の処理に
// 使うべきctx)を使う。
func (tc *TxContext) addAfterCommit(fn func(context.Context) error) {
	if tc.committed {
		Logger(tc.parentContext).Warn("factory: AfterCommit registered after commit already ran; running it immediately")
		if err := fn(tc.parentContext); err != nil {
			Logger(tc.parentContext).WithError(err).Error("factory: after-commit callback (run immediately due to late registration) failed")
		}
		return
	}
	tc.afterCommitFuncs = append(tc.afterCommitFuncs, fn)
}

func (tc *TxContext) runAfterCommit() error {
	var afterCommitErrors []error
	for _, fn := range tc.afterCommitFuncs {
		if err := fn(tc.parentContext); err != nil {
			afterCommitErrors = append(afterCommitErrors, err)
		}
	}
	return errors.Join(afterCommitErrors...)
}

// AfterCommit はトランザクション内ではコミット成功後に処理を実行し、
// トランザクション外では即時実行する。
func AfterCommit(ctx context.Context, fn func(context.Context) error) error {
	if txCtx, ok := ctx.(*TxContext); ok {
		txCtx.addAfterCommit(fn)
		return nil
	}
	if store, ok := ctx.Value(afterCommitContextKey{}).(*afterCommitStore); ok {
		store.add(ctx, fn)
		return nil
	}
	return fn(ctx)
}

type TxManager interface {
	WithTx(ctx context.Context, fn func(context.Context) error) error
}

type txManager struct {
	engine *xorm.Engine
}

func NewTxManager(engine *xorm.Engine) TxManager {
	return &txManager{engine: engine}
}

func (tm *txManager) WithTx(ctx context.Context, fn func(context.Context) error) error {
	session := tm.engine.NewSession()
	err := session.Begin()
	if err != nil {
		session.Close()
		return err
	}

	txCtx := &TxContext{
		Context:       context.WithValue(ctx, ConextDBType, session),
		parentContext: ctx,
		session:       session,
		committed:     false,
	}
	defer txCtx.Close()

	err = fn(txCtx)
	if err != nil {
		txCtx.Rollback()
		return err
	}

	if err = txCtx.Commit(); err != nil {
		return err
	}

	return txCtx.runAfterCommit()
}
