package mysql

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

func Test_userTransactionMysqlRepository_Create(t *testing.T) {
	type args struct {
		ctx             context.Context
		userTransaction domain.UserTransaction
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "UserTransaction 모델 생성 => 정상적으로 생성되기를 기대합니다.",
			args: args{
				ctx:             context.Background(),
				userTransaction: domain.UserTransaction{},
			},
			wantErr: false,
		},
		{
			name: "기존에 이미 있는 UserTransaction 모델 생성  : 1000 => 에러가 발생하기를 기대합니다.",
			args: args{
				ctx:             context.Background(),
				userTransaction: mockUserTransaction1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := userTransactionMysqlRepository{
				dbFactory: dbFactoryMock{},
			}

			_, err := repo.Create(tt.args.ctx, tt.args.userTransaction)
			assert.Equalf(t, tt.wantErr, err != nil, "Create() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}
