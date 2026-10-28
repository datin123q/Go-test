package jobs

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// tryLockJobExecution は、指定したjobTypeについてDBレベルの排他ロックを取得する
// (MIR-1194)。project.TryLockById(MIR-1204)と同じ考え方で、FOR UPDATE NOWAITを
// 使う。「実行中フラグを立てて、一定時間経過したものは無視する」という設計にしな
// かったのは、プロセスが異常終了した場合にフラグが残り続けて以後ずっと弾かれる
// 問題を、MIR-1204で河内さんが避けた判断に合わせるため。ロックは戻り値の release
// を呼ぶまで保持され、プロセスがクラッシュしてDB接続が切れれば自動的に解放される。
// ctxが保持している(cronやHTTPミドルウェアが用意した)セッション・トランザクション
// とは別の専用セッションでロックするため、ジョブ本体のデータ更新のトランザクション
// 設計には影響しない。ロック対象のjob_type行は実行時にINSERT IGNOREで用意せず、
// マイグレーション側で事前に1行だけ用意している。実行時にINSERT IGNOREすると、
// 既に他方がFOR UPDATEで行ロックを保持している場合、重複キー検出のためにMySQLが
// その行へ共有ロックを要求してブロックしてしまい、NOWAITの意味が無くなるため
// (code-review指摘)。
func tryLockJobExecution(ctx context.Context, jobType domain.JOB_EXECUTION_TYPE) (release func(), err error) {
	engine, err := factory.Engine(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get db engine for job execution lock")
	}
	session := engine.NewSession()

	// 名前付き戻り値errが非nilのままここに到達した経路は、すべてロック取得に
	// 失敗した経路(=ここでCloseすべき経路)と一致する。成功時はerrにnilを
	// 明示して return するため、専用のフラグ変数を持たずに済む(code-review指摘)。
	// xorm@v1.3.9のSession.Closeは、Commit/Rollback未実行のトランザクションを
	// 自動でRollbackするため、個別にRollbackを呼ぶ必要も無い。
	defer func() {
		if err != nil {
			session.Close()
		}
	}()

	if err := session.Begin(); err != nil {
		return nil, errors.Wrap(err, "failed to begin job execution lock session")
	}

	var lockedRow struct {
		JobType string `xorm:"job_type"`
	}
	has, err := session.SQL("SELECT job_type FROM job_execution_lock WHERE job_type = ? FOR UPDATE NOWAIT", jobType).Get(&lockedRow)
	if err != nil {
		if factory.IsLockNotAvailableError(err) {
			return nil, factory.ErrJobExecutionLocked.New(err)
		}
		return nil, errors.Wrap(err, "failed to try-lock job execution")
	}
	if !has {
		return nil, errors.Errorf("job execution lock row not found for job type %s (add it to the job_execution_lock migration seed)", jobType)
	}

	return func() {
		// Commit失敗時もこのあとのCloseでロックは解放される(接続がプールへ
		// 返却されるため)。原因調査の手がかりとしてログにだけ残す(code-review指摘)。
		if commitErr := session.Commit(); commitErr != nil {
			logrus.WithError(commitErr).WithFields(logrus.Fields{
				"jobType": jobType,
			}).Warn("failed to commit job execution lock release")
		}
		session.Close()
	}, nil
}

// isJobExecutionLockedError は、tryLockJobExecutionが「他の実行中とぶつかった
// ので弾いた」ことを表すエラーを返したかどうかを判定する。
func isJobExecutionLockedError(err error) bool {
	var apiErr *factory.Error
	return errors.As(err, &apiErr) && apiErr.Code == factory.ErrJobExecutionLocked.Code
}
