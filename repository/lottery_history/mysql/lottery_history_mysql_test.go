package mysql

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

func Test_lotteryHistoryRepository_Create(t *testing.T) {
	type args struct {
		ctx     context.Context
		history domain.LotteryHistory
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "LotteryHistoryモデル作成 => 正常に作成されることが期待されます。",
			args: args{
				ctx: context.Background(),
				history: domain.LotteryHistory{
					ApplicationId: "app-001",
					PrevStatus:    domain.LOTTERY_STATUS_APPLIED,
					NewStatus:     domain.LOTTERY_STATUS_CANCELED,
					ActorId:       999,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := lotteryHistoryRepository{
				dbFactory: dbFactoryMock{},
			}

			err := repo.Create(tt.args.ctx, tt.args.history)
			assert.Equalf(t, tt.wantErr, err != nil, "Create() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}
