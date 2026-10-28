package gmo_aozora

import (
	"context"
	"database/sql"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

const (
	tokenRefreshLockName = "gmo_aozora_token_refresh"
	// GET_LOCKの待機秒数。RefreshTokenの区間(DB読み直し→銀行呼び出し→
	// DB書き込みの再試行)が収まる程度の長さにし、負けたタスクが勝った
	// タスクの完了を実質的に待ってから自分のDB読み直しへ進めるようにする。
	tokenRefreshLockTimeoutSeconds = 30
)

// ErrTokenRefreshLockNotAvailable は、待機秒数内にロックを取得できなかった場合に返す。
var ErrTokenRefreshLockNotAvailable = errors.New("gmo aozora token refresh: lock not available")

// withTokenRefreshLock は、MySQLの名前付きロック(GET_LOCK)でRefreshTokenの
// 「DB読み直し→銀行呼び出し→DB書き込み」区間を排他する(MIR-1298)。
//
// job_execution_lock(MIR-1194)のFOR UPDATE NOWAIT方式を使わなかったのは、
// この区間が銀行へのHTTP呼び出しを挟み数秒かかりうるため、その間トランザクション
// と行ロックを開いたままにしたくないため(DB接続をロック保持のまま長時間
// 占有する・InnoDBのロック待ちを他のクエリに波及させる、といった副作用を
// 避けたい)。GET_LOCKはトランザクション外で使えるアドバイザリーロックで、
// 行ロックを伴わない。
//
// GET_LOCKとRELEASE_LOCKは同じDB接続でなければ意味を持たない(異なる接続からの
// RELEASE_LOCKは何もしない)。cronのセッションはBegin()しないため、ctx経由の
// セッションで素朴にExecを2回呼ぶと、コネクションプールが別々の物理接続を
// 割り当てうる。ここでは*sql.Connを明示的に1本取得してGET_LOCK/RELEASE_LOCKの
// 両方をそれに固定する。ロック区間内の業務ロジック(fn)は通常どおりctxのセッション
// (別の接続になりうる)を使ってよい——GET_LOCKはMySQLサーバ側の名前空間で
// 管理されるアドバイザリーロックで、対象の行やテーブルを実際にロックしている
// わけではないため、どの接続がデータを読み書きするかとは独立している。
// withTokenRefreshLockは変数にして、テストが実MySQL(GET_LOCK)を介さずに
// 差し替えられるようにする(②レビュー2周目の指摘: RequestAozoraNewTokenが
// このロックを通ることにテストが無かった)。既存のgetAccountInfoClient等と
// 同じ間接呼び出しの方針。
var withTokenRefreshLock = withTokenRefreshLockMySQL

func withTokenRefreshLockMySQL(ctx context.Context, fn func(ctx context.Context) error) error {
	engine, err := factory.Engine(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get db engine for gmo aozora token refresh lock")
	}

	conn, err := engine.DB().Conn(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get db connection for gmo aozora token refresh lock")
	}
	defer conn.Close()

	var got sql.NullInt64
	row := conn.QueryRowContext(ctx, "SELECT GET_LOCK(?, ?)", tokenRefreshLockName, tokenRefreshLockTimeoutSeconds)
	if err := row.Scan(&got); err != nil {
		return errors.Wrap(err, "failed to acquire gmo aozora token refresh lock")
	}
	if !got.Valid || got.Int64 != 1 {
		return ErrTokenRefreshLockNotAvailable
	}
	defer func() {
		// 解放はベストエフォート。このあとconn.Close()で接続がプールへ返却
		// されるため、仮にRELEASE_LOCKが失敗してもプロセスの生存期間を超えて
		// ロックが残ることはない(MySQLは接続が切れると名前付きロックを
		// 自動的に解放する)。親ctxが既にキャンセルされていても解放は試みる。
		releaseCtx := context.WithoutCancel(ctx)
		if _, err := conn.ExecContext(releaseCtx, "SELECT RELEASE_LOCK(?)", tokenRefreshLockName); err != nil {
			logrus.WithError(err).Warn("failed to release gmo aozora token refresh lock")
		}
	}()

	return fn(ctx)
}
