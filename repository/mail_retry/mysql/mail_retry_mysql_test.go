package mysql

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

func Test_mailRetryMysqlRepository_Create(t *testing.T) {
	type args struct {
		ctx       context.Context
		mailRetry domain.MailRetry
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    int64
	}{
		{
			name: "MailRetry 모델 생성 => 정상적으로 생성되기를 기대합니다.",
			args: args{
				ctx:       context.Background(),
				mailRetry: mockMailRetry1,
			},
			wantErr: false,
			want:    1001,
		},
		{
			name: "기존에 이미 있는 MailRetry 모델 생성: 1002 => 에러가 발생하기를 기대합니다.",
			args: args{
				ctx:       context.Background(),
				mailRetry: mockMailRetry2,
			},
			wantErr: true,
			want:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mailRetryMysqlRepository{
				dbFactory: dbFactoryMock{},
			}

			got, err := repo.Create(tt.args.ctx, tt.args.mailRetry)
			assert.Equalf(t, tt.wantErr, err != nil, "Create() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want, got, "Create() = %v, want %v", got, tt.want)
		})
	}
}
