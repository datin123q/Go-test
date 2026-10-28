package clean_arch_domain

import (
	"context"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type EventMessageRepository interface {
	Create(ctx context.Context, message domain.EventMessage) error
	Update(ctx context.Context, message domain.EventMessage) error
}

type EventMessageService interface {
	PublishEvent(ctx context.Context, category domain.EventCategory, content any) error
	// RecordEvent と SendRecordedEvent は Outbox パターン(MIR-1202)で使う。
	// トランザクション内で RecordEvent を呼び、コミット成功後に factory.AfterCommit
	// 経由で SendRecordedEvent を実行することで、コミット確定時点で必ず
	// event_message 行が存在する状態を作れる。
	RecordEvent(ctx context.Context, category domain.EventCategory, content any) (domain.EventMessage, error)
	// SendRecordedEvent の戻り値のboolは、実際にKafka送信が成功したかどうかを示す
	// (呼び出し元が送信結果を知りたい場合に使う)。errorはMarshal・EventBusFactory
	// 取得・DB更新自体の失敗のみを示し、Kafka送信自体の失敗はerrorに含めない。
	SendRecordedEvent(ctx context.Context, eventMessage domain.EventMessage) (bool, error)
}
