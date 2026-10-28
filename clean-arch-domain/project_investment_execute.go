package clean_arch_domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type ProjectInvestmentExecuteRepository interface {
	Create(ctx context.Context, projectInvestmentExecute domain.ProjectInvestmentExecute) (domain.ProjectInvestmentExecute, error)
	GetByProjectId(ctx context.Context, projectId int64) (domain.ProjectInvestmentExecute, error)
	// GetAllByProjectId はproject_idの全行を返す(MIR-1295)。project_idには
	// ユニーク制約が無く複数行が入り得るため、単一行しか返さないGetByProjectId
	// では出資実行の進捗状態(記録なし/途中/完了)を安全に判定できない。
	// 1件も無ければ空スライス・nilエラーを返す(ErrItemNotFoundにしない)。
	GetAllByProjectId(ctx context.Context, projectId int64) ([]domain.ProjectInvestmentExecute, error)
	// Update はExecuteStatusだけを更新する(MIR-1204)。分割実行の途中は
	// COMPLETED以外(SCHEDULED)にしておき、全件処理し終えたときだけ
	// COMPLETEDへ更新するために使う。
	Update(ctx context.Context, projectInvestmentExecute domain.ProjectInvestmentExecute) error
	// IncrementPublishFailedCount は法定帳簿へのイベント送信に失敗した件数を
	// publish_failed_count へ加算する(MIR-1202)。イベント送信はコミット後
	// (AfterCommit)に行われるため、このメソッドもAfterCommit経由で呼ばれる
	// 想定。delta<=0のときは何もしない。
	IncrementPublishFailedCount(ctx context.Context, id int64, delta int64) error
}
