package eventbus

import (
	"testing"
)

func Test_KafkaProducer_SendEvent(t *testing.T) {
	t.Skip("Skip TestKafkaProducer_SendEvent")

	type fields struct {
		topic        string
		brokers      []string
		flushTimeout int
	}
	type args struct {
		v []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "TestKafkaProducer_SendEvent",
			fields: fields{
				topic:        "test",
				brokers:      []string{"localhost:9092", "localhost:9092", "localhost:9092"},
				flushTimeout: 1500,
			},

			args: args{
				v: []byte("1W54f9aVr5v2n7hzf"),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewEventBus(tt.fields.brokers, tt.fields.topic, tt.fields.flushTimeout)
			if err != nil {
				t.Errorf("NewProducer() error = %v", err)
				return
			}
			if err := p.Publish(tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("Publish() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
