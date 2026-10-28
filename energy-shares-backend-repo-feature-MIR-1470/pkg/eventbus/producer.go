package eventbus

import (
	"context"
	"crypto/tls"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsCfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/aws_msk_iam_v2"
	"github.com/sirupsen/logrus"
)

type ContextEventBusType string

const (
	ContextStatutoryLedgerEventBus ContextEventBusType = "STATUTORY_LEDGER_EVENT_BUS"
)

var eventbus IEventBus

type IEventBus interface {
	PublishWithTopic(topic string, v []byte) error
	Publish(v []byte) error

	PublishData(ctx context.Context, key []byte, value []byte, headers ...kafka.Header) error
	PublishMessage(ctx context.Context, message kafka.Message) error
}

type eventBus struct {
	writer       *kafka.Writer
	topic        string
	brokers      []string
	flushTimeout int
}

func newKafkaWriter(region string, brokers []string, auth AuthenticationType) (*kafka.Writer, error) {
	if len(brokers) == 0 {
		return nil, nil
	}

	writer := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Balancer: &kafka.LeastBytes{},
	}

	switch auth {
	case AuthenticationNone, AuthenticationPlain:
		return writer, nil
	case AuthenticationAWSMskIAM:
		if region == "" {
			return nil, errors.New("eventBus.newKafkaWriter: region is required for AWS_MSK_IAM authentication")
		}
		ctx := context.Background()
		cfg, err := awsCfg.LoadDefaultConfig(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "eventBus.newKafkaWriter: load aws config")
		}

		mechanism := aws_msk_iam_v2.NewMechanism(aws.Config{
			Region:      region,
			Credentials: cfg.Credentials,
		})

		writer.Transport = &kafka.Transport{
			SASL:        mechanism,
			TLS:         &tls.Config{MinVersion: tls.VersionTLS12},
			MetadataTTL: 120 * time.Second,
		}
		return writer, nil
	default:
		return nil, errors.Errorf("eventBus.newKafkaWriter: unsupported authentication type %q", auth)
	}
}

func (p *eventBus) Publish(v []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(p.flushTimeout)*time.Second)
	defer cancel()
	if p.writer == nil {
		return nil
	}
	message := kafka.Message{
		Key:   []byte(p.topic),
		Value: v,
	}
	err := p.writer.WriteMessages(ctx, message)
	if err != nil {
		return errors.Wrap(err, "eventBus.Publish")
	}
	return nil
}

func (p *eventBus) PublishData(ctx context.Context, key []byte, value []byte, headers ...kafka.Header) error {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(p.flushTimeout)*time.Second)
	defer cancel()
	if p.writer == nil {
		return nil
	}
	message := kafka.Message{
		Key:     key,
		Value:   value,
		Headers: headers,
	}
	err := p.writer.WriteMessages(ctx, message)
	if err != nil {
		return errors.Wrap(err, "eventBus.PublishData")
	}
	return nil
}

func (p *eventBus) PublishMessage(ctx context.Context, message kafka.Message) error {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(p.flushTimeout)*time.Second)
	defer cancel()
	if p.writer == nil {
		return nil
	}
	// Always rely on the default topic configured on the writer.
	message.Topic = ""
	err := p.writer.WriteMessages(ctx, message)
	if err != nil {
		return errors.Wrap(err, "eventBus.PublishMessage")
	}
	return nil
}

func newProducerLocal(brokers []string, topic string, timeout int) (IEventBus, error) {
	w, err := newKafkaWriter("", brokers, AuthenticationNone)
	if err != nil {
		return nil, err
	}
	if w != nil {
		w.Topic = topic
	}

	return &eventBus{
		writer:       w,
		topic:        topic,
		brokers:      brokers,
		flushTimeout: timeout,
	}, nil
}

func newProducerMock() (IEventBus, error) {

	return &eventBus{
		writer: nil,
	}, nil
}

func newProducerMsk(region string, brokers []string, topic string, timeout int) (IEventBus, error) {
	if len(brokers) == 0 {
		return nil, errors.New("eventBus.NewProducer: brokers is empty")
	}

	w, err := newKafkaWriter(region, brokers, AuthenticationAWSMskIAM)
	if err != nil {
		return nil, err
	}

	batchSize := 100 // 100 messages per batch
	batchTimeout := 5 * time.Millisecond

	if w != nil {
		w.Topic = topic
		w.BatchSize = batchSize
		w.BatchTimeout = batchTimeout
	}

	logrus.WithFields(
		logrus.Fields{
			"brokers":   brokers,
			"topic":     topic,
			"timeout":   timeout,
			"region":    region,
			"transport": w.Transport,
		}).Debug("newProducerMsk")

	if w != nil && topic != "" {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
		defer cancel()
		err = w.WriteMessages(ctx, kafka.Message{
			Key: []byte("producer connection"),
		})
		if err != nil {
			return nil, errors.Wrap(err, "eventBus.WriteMessages")
		}
	}

	return &eventBus{
		writer:       w,
		topic:        topic,
		brokers:      brokers,
		flushTimeout: timeout,
	}, nil
}

func NewEventBus(b []string, t string, timeout int) (IEventBus, error) {
	if eventbus == nil {
		var err error
		eventbus, err = newProducerLocal(b, t, timeout)
		if err != nil {
			return nil, err
		}
	}
	return eventbus, nil
}

func NewMskEventBus(r string, b []string, t string, timeout int) (IEventBus, error) {
	if eventbus == nil {
		var err error
		eventbus, err = newProducerMsk(r, b, t, timeout)
		if err != nil {
			return nil, err
		}
	}
	return eventbus, nil
}

// メッセージ検証テスト用のMock
func NewMockEventBus() (IEventBus, error) {
	if eventbus == nil {
		var err error
		eventbus, err = newProducerMock()
		if err != nil {
			return nil, err
		}
	}
	return eventbus, nil
}

func GetEventBus() IEventBus {
	return eventbus
}

func (p *eventBus) PublishWithTopic(topic string, v []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(p.flushTimeout)*time.Second)
	defer cancel()
	if p.writer == nil {
		return nil
	}
	message := kafka.Message{
		Key:   []byte(topic),
		Value: v,
	}
	err := p.writer.WriteMessages(ctx, message)
	if err != nil {
		return errors.Wrap(err, "eventBus.Publish")
	}
	defer p.writer.Close()
	return nil
}

func sliceToString(s []string) string {
	var str string
	for i, v := range s {
		if i == len(s)-1 {
			str += v
			break
		}
		str += v + ","
	}
	return str
}
