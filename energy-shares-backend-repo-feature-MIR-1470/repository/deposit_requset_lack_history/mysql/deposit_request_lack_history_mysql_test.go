package mysql

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

func Test_depositRequestLackHistoryMysqlRepository_Create(t *testing.T) {
	type args struct {
		ctx                       context.Context
		depositRequestLackHistory domain.DepositRequestLackHistory
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "DepositRequestLackHistory 모델 생성 => 정상적으로 생성되기를 기대합니다.",
			args: args{
				ctx: context.Background(),
				depositRequestLackHistory: domain.DepositRequestLackHistory{
					UserId:           2000,
					DepositRequestId: 3000,
					Amount:           10000,
					AccountNumber:    "123-123-123",
					ProjectType:      domain.EQUITY,
					ProjectId:        4000,
					ShareId:          "SHARE_ID_1",
				},
			},
			wantErr: false,
		},
		{
			name: "기존에 이미 있는 DepositRequestLackHistory 모델 생성  : 1000 => 에러가 발생하기를 기대합니다.",
			args: args{
				ctx:                       context.Background(),
				depositRequestLackHistory: mockDepositRequestLackHistory1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := depositRequestLackHistoryMysqlRepository{
				dbFactory: dbFactoryMock{},
			}

			err := repo.Create(tt.args.ctx, tt.args.depositRequestLackHistory)
			assert.Equalf(t, tt.wantErr, err != nil, "Create() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}
