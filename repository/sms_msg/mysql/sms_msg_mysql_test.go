package mysql

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

func Test_smsMsgMysqlRepository_Create(t *testing.T) {
	type args struct {
		ctx    context.Context
		smsMsg domain.SmsMsg
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "(성공) SMS 메시지 생성",
			args: args{
				ctx: context.Background(),
				smsMsg: domain.SmsMsg{
					MsgId:      "MSG100",
					Title:      "제목",
					Name:       "홍길동",
					Sender:     "test@email.com",
					Status:     "REQUEST_SMS",
					SmsType:    "PASSWORD_RESET_COMPLETE_SMS",
					SmsContent: "[EnergyShares] test message",
					Count:      1,
					SendDate:   jsonutils.JSONTime(time.Now()),
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := smsMsgMysqlRepository{
				dbFactory: dbFactoryMock{},
			}
			id, err := repo.Create(tt.args.ctx, tt.args.smsMsg)

			assert.Equalf(t, tt.wantErr, err != nil, "smsMsgMysqlRepository.Create() error = %v, wantErr %v", err, tt.wantErr)
			assert.Greaterf(t, id, int64(0), "smsMsgMysqlRepository.Create() id = %v, want > 0", id)
		})
	}
}

func Test_smsMsgMysqlRepository_Update(t *testing.T) {
	type args struct {
		ctx    context.Context
		smsMsg domain.SmsMsg
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "(성공) SMS 메시지 수정",
			args: args{
				ctx: context.Background(),
				smsMsg: domain.SmsMsg{
					Status: "SUCCESS_SMS",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := smsMsgMysqlRepository{
				dbFactory: dbFactoryMock{},
			}
			err := repo.Update(tt.args.ctx, tt.args.smsMsg)

			assert.Equalf(t, tt.wantErr, err != nil, "smsMsgMysqlRepository.Update() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}
