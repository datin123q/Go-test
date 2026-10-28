package withdrawal

import (
	"context"
	"fmt"
	"reflect"
	"strconv"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/gmoaozorautils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"
	"github.com/samber/lo"
)

type withdrawalCSVData struct {
	BankCode     string // 銀行番号
	BranchCode   string // 支店番号
	DepositType  string // 預金種目
	BankAccount  string // 口座番号 (6桁の場合は前に0)
	Depositor    string // 預金者名
	Amount       int64  // 振込金額
	EmptyColumn1 string // empty column 1
	EmptyColumn2 string // empty column 2
	InvestorId   string // 預金者ID
	RequestDate  string // 要求日時
	CompleteDate string // 完了日時
	Status       string // ステータス
}

var withdrawalCSVColumns = []string{
	"model.withdrawal.bank_code",     // 銀行番号
	"model.withdrawal.branch_code",   // 支店番号
	"model.withdrawal.deposit_type",  // 預金種目
	"model.withdrawal.bank_account",  // 口座番号 (6桁の場合は前に0)
	"model.withdrawal.orderer",       // 預金者名
	"model.withdrawal.amount",        // 振込金額
	"",                               // empty column 1
	"",                               // empty column 2
	"model.withdrawal.investor_id",   // 預金者ID
	"model.withdrawal.request_date",  // 要求日時
	"model.withdrawal.complete_date", // 完了日時
	"model.withdrawal.status",        // ステータス
}

type withdrawalService struct {
	logger     factory.LoggerFactory
	translator i18nutils.Translator
}

func New(
	logger factory.LoggerFactory,
	translator i18nutils.Translator,
) cdomain.WithdrawalService {
	return &withdrawalService{
		logger:     logger,
		translator: translator,
	}
}

func (s withdrawalService) ExportCSV(ctx context.Context, items []domain.WithdrawalDetail) cdomain.CSVData {

	csvHeaders := s.translator.Translates(withdrawalCSVColumns)

	var withdrawals []withdrawalCSVData
	for _, item := range items {
		record := withdrawalCSVData{
			BankCode:     lo.FromPtr(item.BankCode),
			BranchCode:   lo.FromPtr(item.BranchCode),
			DepositType:  lo.If(item.DepositType == domain.ORDINARY_ACCOUNT, "1").ElseIf(item.DepositType == domain.CHECKING_ACCOUNT, "2").Else("ERROR#INVALID_DEPOSIT_TYPE"),
			BankAccount:  formatBankAccount(item.BankAccount),
			Depositor:    gmoaozorautils.ZenkakuToHankaku(item.User.Depositor),
			Amount:       item.Amount,
			EmptyColumn1: "",
			EmptyColumn2: "",
			InvestorId:   item.InvestorId,
			RequestDate:  item.RequestDate.ToString(),
			CompleteDate: item.CompleteDate.ToString(),
			Status:       item.Status.DescriptionJapanese(),
		}

		withdrawals = append(withdrawals, record)
	}

	var records [][]string
	records = append(records, csvHeaders)

	for _, w := range withdrawals {
		records = append(records, factory.DataWriteToCsv(reflect.ValueOf(&w).Elem()))
	}

	return records
}

func formatBankAccount(bankAccount string) string {

	if bankAccount == "" {
		return "ERROR#DATA_EMPTY"
	}

	val, err := strconv.ParseInt(bankAccount, 10, 64)
	if err != nil {
		return "ERROR#DATA_INVALID"
	}

	// 6桁以下は前に0で埋める
	return fmt.Sprintf("%07d", val)
}
