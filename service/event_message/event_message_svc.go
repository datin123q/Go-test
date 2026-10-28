package event_message

import (
	"context"
	"encoding/json"
	"time"

	"github.com/pkg/errors"

	"github.com/samber/lo"
	"github.com/segmentio/ksuid"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type eventMessageService struct {
	loggerFactory          factory.LoggerFactory
	eventBusFactory        factory.EventBusFactory
	eventMessageRepository cdomain.EventMessageRepository
}

func New(
	loggerFactory factory.LoggerFactory,
	eventBusFactory factory.EventBusFactory,
	eventMessageRepository cdomain.EventMessageRepository,
) cdomain.EventMessageService {
	return &eventMessageService{
		loggerFactory:          loggerFactory,
		eventBusFactory:        eventBusFactory,
		eventMessageRepository: eventMessageRepository,
	}
}

// RecordEvent は event_message へ is_publish=false で記録するだけで、送信はしない
// (MIR-1202)。呼び出し元がトランザクション内でこれを呼び、コミット成功後に
// SendRecordedEvent を実行する(Outbox パターン)ことで、コミットが確定した時点で
// 必ず行が存在する状態を作れる。送信の途中でプロセスが落ちても、この行が
// is_publish=false のまま残るためRePublish(MIR-1114)で拾い直せる。
func (e eventMessageService) RecordEvent(ctx context.Context, category domain.EventCategory, content any) (domain.EventMessage, error) {
	eventMessage := domain.EventMessage{
		Category:      category,
		Content:       content,
		ContentString: string(lo.Must1(json.Marshal(content))),
		Uuid:          lo.Must1(ksuid.NewRandom()).String(),
		EventTime:     time.Now(),
		IsPublish:     false,
	}

	if err := e.eventMessageRepository.Create(ctx, eventMessage); err != nil {
		return domain.EventMessage{}, errors.Wrap(err, "eventMessageService.RecordEvent")
	}

	return eventMessage, nil
}

// SendRecordedEvent は RecordEvent で記録済みの EventMessage を送信し、結果に
// 応じて is_publish を更新する(MIR-1202)。戻り値のboolは実際にKafka送信が
// 成功したかどうかを示す。errorはMarshal・EventBusFactory取得・DB更新自体の
// 失敗を示し、Kafka送信自体の失敗はerrorに含めない(その場合はfalse, nilを返す。
// 呼び出し元がエラーとして扱いロールバックしてしまわないようにするため)。
// 呼び出し元が送信結果を知りたい場合(実行レコードへの失敗件数記録など)は
// 戻り値のboolを見る。
func (e eventMessageService) SendRecordedEvent(ctx context.Context, eventMessage domain.EventMessage) (bool, error) {
	data, err := json.Marshal(eventMessage)
	if err != nil {
		return false, errors.Wrap(err, "eventMessageService.SendRecordedEvent")
	}

	bus, err := e.eventBusFactory.StatutoryLedgerEventBus(ctx)
	if err != nil {
		return false, errors.Wrap(err, "eventMessageService.SendRecordedEvent")
	}

	publishErr := bus.Publish(data)
	published := publishErr == nil
	if publishErr != nil {
		// NOTE: ここで return してはいけない。トランザクション内の呼び出し元
		// （user-api の抽選申込）がロールバックして event_message 行ごと消え、
		// 出資成立の一括ループも途中で中断して再実行時に二重計上になる。
		e.loggerFactory.Entry(ctx).
			WithTime(time.Now()).
			WithError(publishErr).
			WithFields(eventMessage.LogFields()).
			Error("fail published EventMessage.SendRecordedEvent")
	} else {
		e.loggerFactory.Entry(ctx).
			WithTime(time.Now()).
			WithFields(eventMessage.LogFields()).
			Info("success published EventMessage.SendRecordedEvent")
	}

	// 送信結果に合わせてis_publishを更新する(MIR-1202)。以前は送信の成否に
	// かかわらずtrueを立てており、失敗した行がRePublish(is_publish=falseを
	// 拾う再送ジョブ、MIR-1114)の対象から漏れていた。
	eventMessage.IsPublish = published

	if err := e.eventMessageRepository.Update(ctx, eventMessage); err != nil {
		e.loggerFactory.Entry(ctx).
			WithTime(time.Now()).
			WithFields(eventMessage.LogFields()).
			WithError(err).
			Error("fail update EventMessage.SendRecordedEvent")

		return published, errors.Wrap(err, "fail update EventMessage.SendRecordedEvent")
	}

	return published, nil
}

// PublishEvent は既存呼び出し元向けの互換ラッパーで、RecordEvent と
// SendRecordedEvent を同期的に呼ぶ(記録とほぼ同時に送信したい場合に使う)。
// 呼び出し元がトランザクション境界を持ち、コミット後にまとめて送信したい場合は
// RecordEvent 単体を呼び、factory.AfterCommit 経由で SendRecordedEvent を遅延実行
// すること(MIR-1202、apps/admin-api/usecase/repayment_dividend 参照)。
func (e eventMessageService) PublishEvent(ctx context.Context, category domain.EventCategory, content any) error {
	eventMessage, err := e.RecordEvent(ctx, category, content)
	if err != nil {
		return err
	}
	_, err = e.SendRecordedEvent(ctx, eventMessage)
	return err
}
