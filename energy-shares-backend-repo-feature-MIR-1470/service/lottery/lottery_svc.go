package lottery

import (
	"fmt"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"
)

var lotteryCSVColumns = []string{
	"model.lottery_application.id",                      // 抽選ID
	"model.lottery_application.status",                  // 抽選ステータス
	"model.lottery_application.user_name",               // 投資家名
	"model.lottery_application.user_email",              // メールアドレス
	"model.lottery_application.project_name",            // ファンド名
	"model.lottery_application.qty",                     // 申込口数
	"model.lottery_application.amount",                  // 申込金額
	"model.lottery_application.winning_amount",          // 当選総額
	"model.lottery_application.canceled_winning_amount", // 当選後キャンセル額
	"model.lottery_application.user_account_balance",    // 投資可能金額
	"model.lottery_application.applied_at",              // 申込日時
}

var bulkLotteryCSVColumns = append(append([]string{}, lotteryCSVColumns...),
	"model.bulk_lottery.result",                   // 当落
	"model.bulk_lottery.requested_winning_amount", // 当選させたい金額
)

type lotteryService struct {
	translator i18nutils.Translator
}

func New(translator i18nutils.Translator) cdomain.LotteryService {
	return &lotteryService{
		translator: translator,
	}
}

func (s lotteryService) ExportCSV(items []domain.LotteryApplication) cdomain.CSVData {
	csvHeaders := s.translator.Translates(lotteryCSVColumns)

	records := [][]string{csvHeaders}

	for _, item := range items {
		userAccountBalance := int64(0)
		if item.UserAccountBalance != nil {
			userAccountBalance = *item.UserAccountBalance
		}

		record := []string{
			item.Id,
			lotteryStatusLabel(s.translator, item.Status),
			item.UserName,
			item.UserEmail,
			item.ProjectName,
			fmt.Sprint(item.Qty),
			fmt.Sprint(item.Amount),
			fmt.Sprint(item.WinningAmount()),
			fmt.Sprint(item.CanceledWinningAmount()),
			fmt.Sprint(userAccountBalance),
			item.AppliedAt.ToString(),
		}

		records = append(records, record)
	}

	return records
}

func (s lotteryService) ExportCSVForBulkLottery(items []domain.LotteryApplication) cdomain.CSVData {
	records := s.ExportCSV(items)
	if len(records) == 0 {
		return records
	}

	records[0] = s.translator.Translates(bulkLotteryCSVColumns)
	for i := 1; i < len(records); i++ {
		records[i] = append(records[i], "", "")
	}

	return records
}

func lotteryStatusLabel(translator i18nutils.Translator, status domain.LOTTERY_STATUS) string {
	if status == "" {
		return ""
	}
	return translator.Translate("commoncode.LOTTERY_STATUS." + string(status))
}
