package lottery

import (
	"testing"
	"time"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/pathutils"
	"github.com/stretchr/testify/assert"
)

func Test_lotteryService_ExportCSV(t *testing.T) {
	appliedAt := jsonutils.JSONTime(time.Date(2026, 4, 10, 1, 2, 3, 0, time.UTC))
	balance := int64(12345)

	type args struct {
		items []domain.LotteryApplication
	}
	tests := []struct {
		name       string
		args       args
		translator i18nutils.Translator
		want       cdomain.CSVData
	}{
		{
			name: "（成功）CSVデータ生成",
			args: args{
				items: []domain.LotteryApplication{
					{
						Id:                 "app-001",
						Status:             domain.LOTTERY_STATUS_PARTIAL_WINNING,
						UserName:           "投資家 太郎",
						UserEmail:          "investor@example.com",
						ProjectName:        "ファンドA",
						Qty:                10,
						Amount:             10000,
						UserAccountBalance: &balance,
						AppliedAt:          appliedAt,
						Winnings: []domain.Share{
							{Amount: 3000, Qty: 3, Price: 1000},
							{Amount: 2000, Qty: 2, Price: 1000, InvestStatus: domain.CANCEL_INVEST},
						},
					},
				},
			},
			translator: i18nutils.New(pathutils.GetPathOfAssets()),
			want: cdomain.CSVData{
				[]string{"抽選ID", "抽選ステータス", "投資家名", "メールアドレス", "ファンド名", "申込口数", "申込金額", "当選総額", "当選後キャンセル額", "投資可能金額", "申込日時"},
				[]string{"app-001", "一部当選", "投資家 太郎", "investor@example.com", "ファンドA", "10", "10000", "5000", "2000", "12345", "2026-04-10 10:02:03"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := lotteryService{
				translator: tt.translator,
			}
			got := svc.ExportCSV(tt.args.items)

			assert.Equalf(t, tt.want, got, "ExportCSV() = %v, want %v", got, tt.want)
		})
	}
}

func Test_lotteryService_ExportCSVForBulkLottery(t *testing.T) {
	appliedAt := jsonutils.JSONTime(time.Date(2026, 4, 10, 1, 2, 3, 0, time.UTC))
	balance := int64(12345)

	type args struct {
		items []domain.LotteryApplication
	}
	tests := []struct {
		name       string
		args       args
		translator i18nutils.Translator
		want       cdomain.CSVData
	}{
		{
			name: "（成功）一括抽選用CSVデータ生成",
			args: args{
				items: []domain.LotteryApplication{
					{
						Id:                 "app-001",
						Status:             domain.LOTTERY_STATUS_PARTIAL_WINNING,
						UserName:           "投資家 太郎",
						UserEmail:          "investor@example.com",
						ProjectName:        "ファンドA",
						Qty:                10,
						Amount:             10000,
						UserAccountBalance: &balance,
						AppliedAt:          appliedAt,
						Winnings: []domain.Share{
							{Amount: 3000},
							{Amount: 2000, InvestStatus: domain.CANCEL_INVEST},
						},
					},
				},
			},
			translator: i18nutils.New(pathutils.GetPathOfAssets()),
			want: cdomain.CSVData{
				[]string{"抽選ID", "抽選ステータス", "投資家名", "メールアドレス", "ファンド名", "申込口数", "申込金額", "当選総額", "当選後キャンセル額", "投資可能金額", "申込日時", "当落", "当選させたい金額"},
				[]string{"app-001", "一部当選", "投資家 太郎", "investor@example.com", "ファンドA", "10", "10000", "5000", "2000", "12345", "2026-04-10 10:02:03", "", ""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := lotteryService{
				translator: tt.translator,
			}
			got := svc.ExportCSVForBulkLottery(tt.args.items)

			assert.Equalf(t, tt.want, got, "ExportCSVForBulkLottery() = %v, want %v", got, tt.want)
		})
	}
}
