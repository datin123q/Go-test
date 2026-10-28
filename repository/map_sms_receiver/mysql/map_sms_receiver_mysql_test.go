package mysql

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

func Test_mapSmsReceiverMysqlRepository_Create(t *testing.T) {
	type args struct {
		ctx    context.Context
		smsMsg domain.MapSmsReceiver
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "(성공) MapSmsReceiver 생성",
			args: args{
				ctx: context.Background(),
				smsMsg: domain.MapSmsReceiver{
					SmsId:    1,
					Receiver: "01012345678",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mapSmsReceiverMysqlRepository{
				dbFactory: dbFactoryMock{},
			}
			id, err := repo.Create(tt.args.ctx, tt.args.smsMsg)

			assert.Equalf(t, tt.wantErr, err != nil, "mapSmsReceiverMysqlRepository.Create() error = %v, wantErr %v", err, tt.wantErr)
			assert.Greaterf(t, id, int64(0), "mapSmsReceiverMysqlRepository.Create() id = %v, want > 0", id)
		})
	}
}
