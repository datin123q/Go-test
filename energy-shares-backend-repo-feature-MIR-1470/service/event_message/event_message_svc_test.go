package event_message

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"

	"github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain/mocks"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

func TestEventMessageService_PublishEvent(t *testing.T) {
	type args struct {
		ctx      context.Context
		category domain.EventCategory
		content  any
	}
	tests := []struct {
		name                   string
		args                   args
		eventMessageRepository func() *mocks.EventMessageRepository
		eventBusFactory        func() *mocks.EventBusFactory
		wantErr                bool
	}{
		{
			name: "イベント発行が正常に実行されます。",
			args: args{
				ctx:      context.Background(),
				category: domain.EVENT_INVESTMENT_REQUEST,
				content:  `{"investment_id": 1000}`,
			},
			eventMessageRepository: func() *mocks.EventMessageRepository {
				eventMessageRepository := new(mocks.EventMessageRepository)
				eventMessageRepository.On("Create", mock.Anything, mock.Anything).Return(nil)
				eventMessageRepository.On("Update", mock.Anything, mock.MatchedBy(func(message domain.EventMessage) bool {
					// イベント発行フラグがtrueに変更されたかどうかを確認します。
					return message.IsPublish == true
				})).Return(nil)
				return eventMessageRepository
			},
			eventBusFactory: func() *mocks.EventBusFactory {
				eventBusFactory := new(mocks.EventBusFactory)
				eventBusFactory.On("StatutoryLedgerEventBus", mock.Anything).Return(&MockEventBus{
					PublishFunc: func(v []byte) error {
						data := struct {
							Category  domain.EventCategory `json:"category"`
							Content   string               `json:"content"`
							UUID      string               `json:"uuid"`
							EventTime string               `json:"event_time"`
						}{}
						err := json.Unmarshal(v, &data)
						assert.NoErrorf(t, err, "json.Unmarshal() error = %v", err)
						assert.NotNilf(t, data, "the data is nil")
						assert.Equal(t, domain.EVENT_INVESTMENT_REQUEST, data.Category)
						assert.Equal(t, `{"investment_id": 1000}`, data.Content)
						return nil
					},
				}, nil)
				return eventBusFactory
			},
			wantErr: false,
		},
		{
			// NOTE(eventMessageService.SendRecordedEvent 内のコメント参照): bus.Publish が
			// 失敗しても return してはいけない(トランザクション内の呼び出し元が
			// ロールバックしてしまうため)。ここではその不変条件、つまり
			// エラーを返さないことを確認する。is_publishは送信結果に合わせてfalseで
			// 更新される(MIR-1202。以前はtrueで更新されており、送信失敗行がRePublish
			// の再送対象から漏れていた)。
			name: "イベント送信が失敗してもエラーを返さず、is_publish=falseで更新される",
			args: args{
				ctx:      context.Background(),
				category: domain.EVENT_INVESTMENT_REQUEST,
				content:  `{"investment_id": 1000}`,
			},
			eventMessageRepository: func() *mocks.EventMessageRepository {
				eventMessageRepository := new(mocks.EventMessageRepository)
				eventMessageRepository.On("Create", mock.Anything, mock.Anything).Return(nil)
				eventMessageRepository.On("Update", mock.Anything, mock.MatchedBy(func(message domain.EventMessage) bool {
					return message.IsPublish == false
				})).Return(nil)
				return eventMessageRepository
			},
			eventBusFactory: func() *mocks.EventBusFactory {
				eventBusFactory := new(mocks.EventBusFactory)
				eventBusFactory.On("StatutoryLedgerEventBus", mock.Anything).Return(&MockEventBus{
					PublishFunc: func(v []byte) error {
						return errors.New("bus down")
					},
				}, nil)
				return eventBusFactory
			},
			wantErr: false,
		},
		{
			name: "更新に失敗したらエラーを返す",
			args: args{
				ctx:      context.Background(),
				category: domain.EVENT_INVESTMENT_REQUEST,
				content:  `{"investment_id": 1000}`,
			},
			eventMessageRepository: func() *mocks.EventMessageRepository {
				eventMessageRepository := new(mocks.EventMessageRepository)
				eventMessageRepository.On("Create", mock.Anything, mock.Anything).Return(nil)
				eventMessageRepository.On("Update", mock.Anything, mock.Anything).Return(errors.New("update failed"))
				return eventMessageRepository
			},
			eventBusFactory: func() *mocks.EventBusFactory {
				eventBusFactory := new(mocks.EventBusFactory)
				eventBusFactory.On("StatutoryLedgerEventBus", mock.Anything).Return(&MockEventBus{
					PublishFunc: func(v []byte) error {
						return nil
					},
				}, nil)
				return eventBusFactory
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := eventMessageService{
				loggerFactory:          factory.NewLoggerFactory(),
				eventBusFactory:        tt.eventBusFactory(),
				eventMessageRepository: tt.eventMessageRepository(),
			}

			err := e.PublishEvent(tt.args.ctx, tt.args.category, tt.args.content)
			assert.Equalf(t, tt.wantErr, err != nil, "PublishEvent() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

// TestEventMessageService_RecordEvent は、RecordEvent(MIR-1202)が記録のみ行い
// 送信は一切しないことを確認する。Outboxパターンでトランザクション内から
// 呼ばれる想定のため、ここで送信してしまうとコミット前に外部通知が飛んでしまう。
func TestEventMessageService_RecordEvent(t *testing.T) {
	t.Run("記録に成功したら、送信は一切行わずEventMessageを返す", func(t *testing.T) {
		eventMessageRepository := new(mocks.EventMessageRepository)
		eventMessageRepository.On("Create", mock.Anything, mock.MatchedBy(func(m domain.EventMessage) bool {
			return m.Category == domain.EVENT_INVESTMENT_REQUEST && m.IsPublish == false
		})).Return(nil)

		eventBusFactory := new(mocks.EventBusFactory)
		// StatutoryLedgerEventBusが呼ばれたら即失敗させ、送信していないことを担保する。
		eventBusFactory.On("StatutoryLedgerEventBus", mock.Anything).Return(nil, errors.New("must not be called"))

		e := eventMessageService{
			loggerFactory:          factory.NewLoggerFactory(),
			eventBusFactory:        eventBusFactory,
			eventMessageRepository: eventMessageRepository,
		}

		eventMessage, err := e.RecordEvent(context.Background(), domain.EVENT_INVESTMENT_REQUEST, `{"investment_id": 1000}`)
		require.NoError(t, err)
		assert.NotEmpty(t, eventMessage.Uuid)
		assert.False(t, eventMessage.IsPublish)
	})

	t.Run("記録に失敗したらエラーを返す", func(t *testing.T) {
		eventMessageRepository := new(mocks.EventMessageRepository)
		eventMessageRepository.On("Create", mock.Anything, mock.Anything).Return(errors.New("create failed"))

		e := eventMessageService{
			loggerFactory:          factory.NewLoggerFactory(),
			eventMessageRepository: eventMessageRepository,
		}

		_, err := e.RecordEvent(context.Background(), domain.EVENT_INVESTMENT_REQUEST, `{"investment_id": 1000}`)
		require.Error(t, err)
	})
}

// TestEventMessageService_SendRecordedEvent は、RecordEventで記録済みの
// EventMessageを送信し、送信結果に応じてis_publishを更新することを確認する
// (MIR-1202の③、送信失敗時にtrueを立てていた既存の不具合の修正)。
func TestEventMessageService_SendRecordedEvent(t *testing.T) {
	baseEventMessage := domain.EventMessage{
		Category:      domain.EVENT_INVESTMENT_REQUEST,
		ContentString: `{"investment_id": 1000}`,
		Uuid:          "test-uuid",
		IsPublish:     false,
	}

	t.Run("送信に成功したらis_publish=trueで更新される", func(t *testing.T) {
		eventMessageRepository := new(mocks.EventMessageRepository)
		eventMessageRepository.On("Update", mock.Anything, mock.MatchedBy(func(m domain.EventMessage) bool {
			return m.IsPublish == true
		})).Return(nil)

		eventBusFactory := new(mocks.EventBusFactory)
		eventBusFactory.On("StatutoryLedgerEventBus", mock.Anything).Return(&MockEventBus{
			PublishFunc: func(v []byte) error { return nil },
		}, nil)

		e := eventMessageService{
			loggerFactory:          factory.NewLoggerFactory(),
			eventBusFactory:        eventBusFactory,
			eventMessageRepository: eventMessageRepository,
		}

		published, err := e.SendRecordedEvent(context.Background(), baseEventMessage)
		require.NoError(t, err)
		assert.True(t, published)
	})

	t.Run("送信に失敗してもエラーを返さず、is_publish=falseで更新される", func(t *testing.T) {
		eventMessageRepository := new(mocks.EventMessageRepository)
		eventMessageRepository.On("Update", mock.Anything, mock.MatchedBy(func(m domain.EventMessage) bool {
			return m.IsPublish == false
		})).Return(nil)

		eventBusFactory := new(mocks.EventBusFactory)
		eventBusFactory.On("StatutoryLedgerEventBus", mock.Anything).Return(&MockEventBus{
			PublishFunc: func(v []byte) error { return errors.New("bus down") },
		}, nil)

		e := eventMessageService{
			loggerFactory:          factory.NewLoggerFactory(),
			eventBusFactory:        eventBusFactory,
			eventMessageRepository: eventMessageRepository,
		}

		published, err := e.SendRecordedEvent(context.Background(), baseEventMessage)
		require.NoError(t, err)
		assert.False(t, published, "送信結果(published)にも失敗が正しく反映されること")
	})

	t.Run("Updateに失敗したらエラーを返す", func(t *testing.T) {
		eventMessageRepository := new(mocks.EventMessageRepository)
		eventMessageRepository.On("Update", mock.Anything, mock.Anything).Return(errors.New("update failed"))

		eventBusFactory := new(mocks.EventBusFactory)
		eventBusFactory.On("StatutoryLedgerEventBus", mock.Anything).Return(&MockEventBus{
			PublishFunc: func(v []byte) error { return nil },
		}, nil)

		e := eventMessageService{
			loggerFactory:          factory.NewLoggerFactory(),
			eventBusFactory:        eventBusFactory,
			eventMessageRepository: eventMessageRepository,
		}

		_, err := e.SendRecordedEvent(context.Background(), baseEventMessage)
		require.Error(t, err)
	})
}

func Test_EventMessage_LogFields(t *testing.T) {
	// content には氏名・取引金額・残高が入る。ログへ出すのは識別子だけにする。
	eventMessage := domain.EventMessage{
		Uuid:     "01HZX0000000000000000000",
		Category: domain.EVENT_TRANSACTION,
		Content: map[string]any{
			"user_name": "ホン・ギルドン",
			"amount":    1000000,
			"balance":   5000000,
		},
		ContentString: `{"user_name":"ホン・ギルドン","amount":1000000}`,
	}

	got := eventMessage.LogFields()

	assert.Equal(t, "01HZX0000000000000000000", got["event_uuid"])
	assert.Equal(t, domain.EVENT_TRANSACTION, got["event_category"])
	assert.Contains(t, got, "event_time")

	// 中身を持ち出していないこと。キーが増えたら気付けるよう件数も固定する
	assert.Len(t, got, 3)
	assert.NotContains(t, got, "event")
	assert.NotContains(t, got, "content")

	for key, value := range got {
		assert.NotContains(t, fmt.Sprint(value), "ホン・ギルドン", "%s に氏名が含まれている", key)
		assert.NotContains(t, fmt.Sprint(value), "1000000", "%s に金額が含まれている", key)
	}
}

// Test_eventLogFields_WithBusinessIdentifier は、admin-api の一括分配実行
// (repayment_dividend usecase) が直接呼び出す本経路で、実際に業務上の
// 識別子がログに残ることを確かめる。2026-08-19 の障害では1,166件のログが
// この経路(EVENT_TRANSACTION / EVENT_FUND_TRANSFER)から出ており、
// event_uuid・event_category・event_time だけでは対象の投資/ユーザーを
// 個別に特定できなかった。
func Test_EventMessage_LogFields_WithBusinessIdentifier(t *testing.T) {
	eventMessage := domain.EventMessage{
		Uuid:     "01HZX0000000000000000000",
		Category: domain.EVENT_TRANSACTION,
		Content: domain.TransactionDto{
			UserId:       1,
			InvestmentId: "inv-1",
			ProductId:    2,
			UserName:     "ホン・ギルドン",
			Amount:       1000000,
			Balance:      5000000,
		},
	}

	got := eventMessage.LogFields()

	assert.Equal(t, int64(1), got["user_id"])
	assert.Equal(t, "inv-1", got["investment_id"])
	assert.Equal(t, int64(2), got["product_id"])
	assert.Equal(t, "01HZX0000000000000000000", got["event_uuid"])
	assert.Equal(t, domain.EVENT_TRANSACTION, got["event_category"])
	assert.Contains(t, got, "event_time")

	for key, value := range got {
		assert.NotContains(t, fmt.Sprint(value), "ホン・ギルドン", "%s に氏名が含まれている", key)
		assert.NotContains(t, fmt.Sprint(value), "1000000", "%s に金額が含まれている", key)
	}
}
