package mysql

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

func Test_eventMessageMysqlRepository_Create(t *testing.T) {
	type args struct {
		ctx          context.Context
		eventMessage domain.EventMessage
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "EventMessage 모델 생성 => 정상적으로 생성되기를 기대합니다.",
			args: args{
				ctx: context.Background(),
				eventMessage: domain.EventMessage{
					Category:      domain.EVENT_FUND_TRANSFER,
					Content:       `{"text" : "hello world"}`,
					ContentString: "",
					Uuid:          "d962f932-7324-43e6-8d29-bc31f9bde697",
					IsPublish:     true,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := eventMessageMysqlRepository{
				dbFactory: dbFactoryMock{},
			}

			err := repo.Create(tt.args.ctx, tt.args.eventMessage)
			assert.Equalf(t, tt.wantErr, err != nil, "Create() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_eventMessageMysqlRepository_Update(t *testing.T) {
	type args struct {
		ctx     context.Context
		message domain.EventMessage
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "EventMessage 모델 수정 => 정상적으로 수정되기를 기대합니다.",
			args: args{
				ctx: context.Background(),
				message: domain.EventMessage{
					Uuid:      "29f24996-e3ae-4012-b25e-b31773c6cbfb",
					IsPublish: true,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := eventMessageMysqlRepository{
				dbFactory: dbFactoryMock{},
			}

			err := repo.Update(tt.args.ctx, tt.args.message)
			assert.Equalf(t, tt.wantErr, err != nil, "Update() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}
