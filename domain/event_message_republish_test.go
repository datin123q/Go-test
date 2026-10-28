package domain

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

// RePublish(MIR-1114)のテスト。Kafka送信の成否はpublishToEventBusを差し替えて注入する。

func setupRePublishTestDB(t *testing.T) *xorm.Engine {
	t.Helper()
	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, engine.Close()) })
	require.NoError(t, engine.Sync2(new(EventMessage)))
	return engine
}

// insertUnpublishedEvent は、rePublishMinAgeより十分古い作成時刻で行を挿入する。
// RePublishが「直近rePublishMinAge以内の行は対象外にする」ようになった
// (MIR-1202のOutbox化に伴う二重送信対策)ため、既存のテストが前提としている
// 「挿入した行がRePublishの対象になる」を保つには、デフォルトで古い時刻に
// しておく必要がある。直近の行の挙動を確認したいテストは
// insertUnpublishedEventCreatedAt を使う。
func insertUnpublishedEvent(t *testing.T, engine *xorm.Engine, contentString string) EventMessage {
	t.Helper()
	return insertUnpublishedEventCreatedAt(t, engine, contentString, time.Now().Add(-rePublishMinAge-time.Hour))
}

// insertUnpublishedEventCreatedAt は Created(自動タイムスタンプ)を指定した
// 時刻で挿入する。xormはCreatedをInsert時に自動でtime.Now()に上書きするため、
// 生SQLで直接UPDATEすることで確実に指定した時刻にする。
func insertUnpublishedEventCreatedAt(t *testing.T, engine *xorm.Engine, contentString string, createdAt time.Time) EventMessage {
	t.Helper()
	id, err := ksuid.NewRandom()
	require.NoError(t, err)
	e := EventMessage{
		Category:      EVENT_INVESTMENT_REQUEST,
		ContentString: contentString,
		Uuid:          id.String(),
		EventTime:     time.Now(),
		IsPublish:     false,
	}
	_, err = engine.Insert(&e)
	require.NoError(t, err)

	// engine.Exec経由の生SQLはRePublish本体のWhere条件とは別の経路で、
	// .UTC()の有無で格納値が変わらないことを実測で確認済み(不要だが、
	// 呼んでも実害はない)。本体側のWhere条件のように.UTC()の有無で
	// 結果が変わるわけではないので、本体側と混同しないこと。
	_, err = engine.Exec("UPDATE event_message SET created = ? WHERE uuid = ?", createdAt, e.Uuid)
	require.NoError(t, err)
	e.Created = createdAt

	return e
}

func TestEventMessage_RePublish(t *testing.T) {
	t.Run("送信に成功したらis_publishがtrueになり、次回の再送対象から外れる", func(t *testing.T) {
		engine := setupRePublishTestDB(t)
		e := insertUnpublishedEvent(t, engine, `{"user_id":1}`)

		original := publishToEventBus
		t.Cleanup(func() { publishToEventBus = original })
		publishToEventBus = func(data []byte) error { return nil }

		ctx := context.WithValue(context.Background(), factory.ConextDBType, engine)
		require.NoError(t, EventMessage{}.RePublish(ctx))

		var reloaded EventMessage
		has, err := engine.ID(e.Uuid).Get(&reloaded)
		require.NoError(t, err)
		require.True(t, has)
		assert.True(t, reloaded.IsPublish)
	})

	t.Run("送信するペイロードに、DTOへ復元したcontentが実際に含まれる", func(t *testing.T) {
		// toSendを送信用のコピーに分けたことで、送信データ自体が壊れる余地が
		// あるため(このコピーからContentを消せば本テストは落ちる)、
		// callCountだけでなくペイロードの中身まで確認する。ログ側が
		// toSend.LogFields()を正しく参照しているかどうかまではここでは
		// 検証していない(logrusのフックを使ったテストが別途必要)。
		engine := setupRePublishTestDB(t)
		insertUnpublishedEvent(t, engine, `{"user_id":123}`)

		var sentData []byte
		original := publishToEventBus
		t.Cleanup(func() { publishToEventBus = original })
		publishToEventBus = func(data []byte) error {
			sentData = data
			return nil
		}

		ctx := context.WithValue(context.Background(), factory.ConextDBType, engine)
		require.NoError(t, EventMessage{}.RePublish(ctx))

		require.NotNil(t, sentData)
		var payload struct {
			Content InvestmentRequestDto `json:"content"`
		}
		require.NoError(t, json.Unmarshal(sentData, &payload))
		assert.EqualValues(t, 123, payload.Content.UserId)
	})

	t.Run("送信に成功しても、DBのcontent列は書き換わらない", func(t *testing.T) {
		engine := setupRePublishTestDB(t)
		// legacy_fieldはInvestmentRequestDtoに無いフィールド。Update経由だと
		// DTOへのUnmarshal→再マーシャルで消えてしまう(MIR-1114の敵対的検証で発見)。
		originalContent := `{"user_id":1,"legacy_field":"keep me"}`
		e := insertUnpublishedEvent(t, engine, originalContent)

		original := publishToEventBus
		t.Cleanup(func() { publishToEventBus = original })
		publishToEventBus = func(data []byte) error { return nil }

		ctx := context.WithValue(context.Background(), factory.ConextDBType, engine)
		require.NoError(t, EventMessage{}.RePublish(ctx))

		var reloaded EventMessage
		has, err := engine.ID(e.Uuid).Get(&reloaded)
		require.NoError(t, err)
		require.True(t, has)
		assert.True(t, reloaded.IsPublish)
		assert.Equal(t, originalContent, reloaded.ContentString, "DTOに無いフィールドを含む元のcontentが破壊されないこと")
	})

	t.Run("送信に失敗したらis_publishはfalseのまま残る", func(t *testing.T) {
		engine := setupRePublishTestDB(t)
		e := insertUnpublishedEvent(t, engine, `{"user_id":1}`)

		original := publishToEventBus
		t.Cleanup(func() { publishToEventBus = original })
		callCount := 0
		publishToEventBus = func(data []byte) error {
			callCount++
			return assert.AnError
		}

		ctx := context.WithValue(context.Background(), factory.ConextDBType, engine)
		require.NoError(t, EventMessage{}.RePublish(ctx))
		assert.Equal(t, 1, callCount, "行がis_publish=falseのまま残るだけでなく、実際にPublishが呼ばれたことも確認する")

		var reloaded EventMessage
		has, err := engine.ID(e.Uuid).Get(&reloaded)
		require.NoError(t, err)
		require.True(t, has)
		assert.False(t, reloaded.IsPublish, "送信失敗時はfalseのまま残り、次回も再送対象になること")
	})

	t.Run("1件が送信失敗しても、残りの未送信は処理され続ける", func(t *testing.T) {
		engine := setupRePublishTestDB(t)
		failing := insertUnpublishedEvent(t, engine, `{"user_id":1}`)
		succeeding := insertUnpublishedEvent(t, engine, `{"user_id":2}`)

		original := publishToEventBus
		t.Cleanup(func() { publishToEventBus = original })
		callCount := 0
		publishToEventBus = func(data []byte) error {
			callCount++
			if callCount == 1 {
				return assert.AnError
			}
			return nil
		}

		ctx := context.WithValue(context.Background(), factory.ConextDBType, engine)
		require.NoError(t, EventMessage{}.RePublish(ctx))
		assert.Equal(t, 2, callCount, "1件目が失敗しても2件目のPublishが呼ばれること")

		var reloadedFailing, reloadedSucceeding EventMessage
		has, err := engine.ID(failing.Uuid).Get(&reloadedFailing)
		require.NoError(t, err)
		require.True(t, has)
		has, err = engine.ID(succeeding.Uuid).Get(&reloadedSucceeding)
		require.NoError(t, err)
		require.True(t, has)
		assert.False(t, reloadedFailing.IsPublish)
		assert.True(t, reloadedSucceeding.IsPublish)
	})

	t.Run("contentが壊れている行はエラーのままスキップし、残りは処理され続ける", func(t *testing.T) {
		engine := setupRePublishTestDB(t)
		corrupted := insertUnpublishedEvent(t, engine, `not a valid json`)
		valid := insertUnpublishedEvent(t, engine, `{"user_id":2}`)

		original := publishToEventBus
		t.Cleanup(func() { publishToEventBus = original })
		callCount := 0
		publishToEventBus = func(data []byte) error {
			callCount++
			return nil
		}

		ctx := context.WithValue(context.Background(), factory.ConextDBType, engine)
		require.NoError(t, EventMessage{}.RePublish(ctx))
		assert.Equal(t, 1, callCount, "壊れた行はPublish自体を呼ばずスキップすること")

		var reloadedCorrupted, reloadedValid EventMessage
		has, err := engine.ID(corrupted.Uuid).Get(&reloadedCorrupted)
		require.NoError(t, err)
		require.True(t, has)
		has, err = engine.ID(valid.Uuid).Get(&reloadedValid)
		require.NoError(t, err)
		require.True(t, has)
		assert.False(t, reloadedCorrupted.IsPublish, "壊れた行はis_publishが更新されずfalseのまま残ること(エラーとして分かる)")
		assert.True(t, reloadedValid.IsPublish)
	})

	// このテストは本体側(RePublish)の.UTC()の欠落を非UTC環境でのみ検出する
	// (CIはAsia/Tokyoで実行されるため確実に捕まえる。.github/workflows/go.yml
	// 参照。開発機がUTCだと.UTC()を外しても偶然通ってしまう)。
	t.Run("直近rePublishMinAge以内に作られた行は再送対象から除外される(MIR-1202のOutbox化に伴う二重送信対策)", func(t *testing.T) {
		engine := setupRePublishTestDB(t)
		recent := insertUnpublishedEventCreatedAt(t, engine, `{"user_id":1}`, time.Now().Add(-1*time.Minute))
		old := insertUnpublishedEventCreatedAt(t, engine, `{"user_id":2}`, time.Now().Add(-rePublishMinAge-time.Minute))

		original := publishToEventBus
		t.Cleanup(func() { publishToEventBus = original })
		callCount := 0
		publishToEventBus = func(data []byte) error {
			callCount++
			return nil
		}

		ctx := context.WithValue(context.Background(), factory.ConextDBType, engine)
		require.NoError(t, EventMessage{}.RePublish(ctx))
		assert.Equal(t, 1, callCount, "十分古い行だけが再送されること")

		var reloadedRecent, reloadedOld EventMessage
		has, err := engine.ID(recent.Uuid).Get(&reloadedRecent)
		require.NoError(t, err)
		require.True(t, has)
		has, err = engine.ID(old.Uuid).Get(&reloadedOld)
		require.NoError(t, err)
		require.True(t, has)
		assert.False(t, reloadedRecent.IsPublish, "直近の行はまだ送信中かもしれないため再送対象から除外されること")
		assert.True(t, reloadedOld.IsPublish, "十分古い行は再送されis_publish=trueになること")
	})
}
