package mysql

import (
	"context"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_marketingCodeMysqlRepository_Create(t *testing.T) {
	type args struct {
		ctx           context.Context
		marketingCode domain.MarketingCode
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "마케팅 코드 생성 -> 정상적으로 생성되기를 기대합니다.",
			args: args{
				ctx: context.Background(),
				marketingCode: domain.MarketingCode{
					Code:            "marketing_code_2",
					Campaign:        "marketing_campaign_2",
					SignUpRequestId: "sign_up_request_id_2",
					Email:           "dool2@email.com",
					UserId:          101,
				},
			},
			wantErr: false,
			want:    2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := marketingCodeMysqlRepository{
				dbFactory: dbFactoryMock{},
			}
			got, err := repo.Create(tt.args.ctx, tt.args.marketingCode)
			assert.Equalf(t, tt.wantErr, err != nil, "marketingCodeMysqlRepository.Create() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.want, got, "marketingCodeMysqlRepository.Create() = %v, want %v", got, tt.want)
		})
	}
}
