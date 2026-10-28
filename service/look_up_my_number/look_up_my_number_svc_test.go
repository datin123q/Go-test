package look_up_my_number

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain/mocks"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/cipher"
)

func Test_LookUpMyNumberService_GetDecryptedMyNumberByMyNumberId(t *testing.T) {
	type args struct {
		ctx        context.Context
		myNumberId int64
	}
	tests := []struct {
		name                     string
		args                     args
		lookUpMyNumberRepository func() *mocks.LookUpMyNumberRepository
		want                     string
		wantErr                  error
	}{
		{
			name: "(성공) GetDecryptedMyNumberByMyNumberId 모델 조회 (myNumberId:1000) -> 정상적으로 복호화한 문자열을 기대합니다.",
			args: args{
				ctx:        context.Background(),
				myNumberId: mockLookUpMyNumber1.MyNumberId,
			},
			lookUpMyNumberRepository: func() *mocks.LookUpMyNumberRepository {
				lookUpMyNumberRepository := &mocks.LookUpMyNumberRepository{}
				lookUpMyNumberRepository.On("GetByMyNumberIdIncludeDeleted",
					context.Background(),
					mockLookUpMyNumber1.MyNumberId).
					Return(mockLookUpMyNumber1, nil)
				return lookUpMyNumberRepository
			},

			want:    mockLookUpMyNumber2.MyNumber,
			wantErr: nil,
		},
	}

	crypto, _ := cipher.New()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := lookUpMyNumberService{
				crypto:                   crypto,
				lookUpMyNumberRepository: tt.lookUpMyNumberRepository(),
			}

			got, err := service.GetDecryptedMyNumberByMyNumberId(tt.args.ctx, tt.args.myNumberId)
			if err != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("lookUpMyNumberService.GetDecryptedMyNumberByMyNumberId() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equalf(t, tt.want, got, "lookUpMyNumberService.GetDecryptedMyNumberByMyNumberId() = %v, want %v", got, tt.want)
		})
	}
}
