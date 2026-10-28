package event_message

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type MockEventBus struct {
	PublishFunc func(v []byte) error
	PublishWithTopicFunc func(topic string, v []byte) error
	PublishDataFunc func(ctx context.Context, key []byte, value []byte, headers ...kafka.Header) error
	PublishMessageFunc func(ctx context.Context, message kafka.Message) error
}

func (m *MockEventBus) Publish(v []byte) error {
	if m.PublishFunc != nil {
		return m.PublishFunc(v)
	}
	return nil
}

func (m *MockEventBus) PublishWithTopic(topic string, v []byte) error {
	if m.PublishWithTopicFunc != nil {
		return m.PublishWithTopicFunc(topic, v)
	}
	return nil
}

func (m *MockEventBus) PublishData(ctx context.Context, key []byte, value []byte, headers ...kafka.Header) error {
	if m.PublishDataFunc != nil {
		return m.PublishDataFunc(ctx, key, value, headers...)
	}
	return nil
}

func (m *MockEventBus) PublishMessage(ctx context.Context, message kafka.Message) error {
	if m.PublishMessageFunc != nil {
		return m.PublishMessageFunc(ctx, message)
	}
	return nil
}
