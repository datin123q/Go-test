package sms_msg

import (
	"context"
	"testing"
	"text/template"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain/mocks"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/pathutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/snsutils"
)

func Test_smsMsgService_SendSms(t *testing.T) {
	type args struct {
		ctx     context.Context
		request cdomain.SendSmsRequest
	}
	tests := []struct {
		name                     string
		args                     args
		snsPkg                   func() snsutils.Sns
		smsMsgRepository         func() cdomain.SmsMsgRepository
		mapSmsReceiverRepository func() cdomain.MapSmsReceiverRepository
		wantErr                  error
	}{
		{
			name: "(성공) SMS 전송 => 성공",
			args: args{
				ctx:     context.Background(),
				request: mockSendSmsRequest,
			},
			snsPkg: func() snsutils.Sns {
				snsPkg := &mocks.Sns{}
				snsPkg.On("SendSMS", context.Background(), mockSendSmsRequest.Receiver, mock.Anything).Return("messageID", nil)
				template, _ := template.ParseFiles(pathutils.GetPathOfAssets() + mockSendSmsRequest.Template)
				snsPkg.On("LoadTextTemplate", context.Background(), mockSendSmsRequest.Template).Return(template, nil)
				return snsPkg
			},
			smsMsgRepository: func() cdomain.SmsMsgRepository {
				smsMsgRepository := &mocks.SmsMsgRepository{}
				smsMsgRepository.On("Create", context.Background(), mock.Anything).Return(int64(0), nil)
				smsMsgRepository.On("Update", context.Background(), mock.Anything).Return(nil)
				return smsMsgRepository
			},
			mapSmsReceiverRepository: func() cdomain.MapSmsReceiverRepository {
				mapSmsReceiverRepository := &mocks.MapSmsReceiverRepository{}
				mapSmsReceiverRepository.On("Create", context.Background(), mockMapSmsReceiver).Return(int64(0), nil)

				return mapSmsReceiverRepository
			},
			wantErr: nil,
		},
		{
			name: "(실패) SMS 전송 => SMS 메시지 전송 실패",
			args: args{
				ctx:     context.Background(),
				request: mockSendSmsRequest,
			},
			snsPkg: func() snsutils.Sns {
				snsPkg := &mocks.Sns{}
				snsPkg.On("SendSMS", context.Background(), mockSendSmsRequest.Receiver, mock.Anything).Return("", errors.New("failed to send sms"))
				template, _ := template.ParseFiles(pathutils.GetPathOfAssets() + mockSendSmsRequest.Template)
				snsPkg.On("LoadTextTemplate", context.Background(), mockSendSmsRequest.Template).Return(template, nil)
				return snsPkg
			},
			smsMsgRepository: func() cdomain.SmsMsgRepository {
				smsMsgRepository := &mocks.SmsMsgRepository{}
				smsMsgRepository.On("Create", context.Background(), mock.Anything).Return(int64(0), nil)
				smsMsgRepository.On("Update", context.Background(), mock.Anything).Return(nil)
				return smsMsgRepository
			},
			mapSmsReceiverRepository: func() cdomain.MapSmsReceiverRepository {
				mapSmsReceiverRepository := &mocks.MapSmsReceiverRepository{}
				mapSmsReceiverRepository.On("Create", context.Background(), mockMapSmsReceiver).Return(int64(0), nil)

				return mapSmsReceiverRepository
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := smsMsgService{
				snsPkg:                   tt.snsPkg(),
				loggerFactory:            factory.NewLoggerFactory(),
				smsMsgRepository:         tt.smsMsgRepository(),
				mapSmsReceiverRepository: tt.mapSmsReceiverRepository(),
			}

			err := service.SendSms(tt.args.ctx, tt.args.request)
			if err != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("smsMsgService.SendSms() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

		})
	}
}
