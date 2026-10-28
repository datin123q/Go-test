package withdrawal

import (
	"context"
	"testing"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/pathutils"
	"github.com/stretchr/testify/assert"
)

func Test_withdrawalService_ExportCSV(t *testing.T) {
	type args struct {
		ctx   context.Context
		items []domain.WithdrawalDetail
	}

	tests := []struct {
		name       string
		args       args
		translator i18nutils.Translator
		want       cdomain.CSVData
	}{
		{
			name: "(성공) CSV 데이터 생성",
			args: args{
				items: []domain.WithdrawalDetail{
					mockWithdrawal1,
					mockWithdrawal2,
				},
			},
			translator: i18nutils.New(pathutils.GetPathOfAssets()),
			want: [][]string{
				{"被仕向銀行番号", "被仕向支店番号", "預金種目", "口座番号", "受取人名", "振込金額", "", "", "顧客ID", "依頼日時", "完了日時", "ステータス"},
				{"001", "031", "1", "0001234", "ﾃｽﾄ", "1000", "", "", "1", "2024-02-01 09:00:00", "2024-03-01 09:00:00", "完了"},
				{"322", "018", "1", "0123456", "ﾃｽﾄ ﾃｽﾄ", "5000", "", "", "2", "2024-02-01 09:00:00", "", "送金待ち"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := withdrawalService{
				translator: tt.translator,
			}

			got := service.ExportCSV(tt.args.ctx, tt.args.items)

			assert.Equalf(t, tt.want, got, "ExportCSV() got = %v, want %v", got, tt.want)
		})
	}
}
