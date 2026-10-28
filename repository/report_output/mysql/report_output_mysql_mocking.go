package mysql

import (
	"context"
	"time"

	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

// REPORT_PAYMENT_STATEMENT
var (
	mockReportPaymentStatementOutput1 = domain.ReportOutput{
		Id:             1001,
		UserId:         1,
		SpvId:          0,
		IsOpen:         true,
		Title:          "energyshares 2024年 支払調書",
		SubmissionType: domain.SUBMISSION_INVESTORS,
		ReportType:     domain.REPORT_PAYMENT_STATEMENT,
		Year:           "2023",
		Quarter:        "",
		FileUrl:        "https://staturoy-cdn-sb-jp-dev.illuminarean.com/public/statutory_books/ledger_file/payment_statement/user/6/2024-02-15/0ea90e56-57ca-42c2-9d9e-6a82533f0fcb.pdf",
		AtOpen:         jsonutils.JSONTime(time.Date(2024, 8, 29, 9, 34, 19, 0, time.UTC)),
		Created:        jsonutils.JSONTime(time.Date(2024, 2, 15, 1, 21, 41, 0, time.UTC)),
		Updated:        jsonutils.JSONTime(time.Date(2024, 8, 29, 9, 34, 19, 0, time.UTC)),
		Deleted:        jsonutils.JSONTime(time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)),
	}

	mockReportPaymentStatementOutput2 = domain.ReportOutput{
		Id:             1002,
		UserId:         2,
		SpvId:          0,
		IsOpen:         true,
		Title:          "すべての法人 2024年 支払調書",
		SubmissionType: domain.SUBMISSION_INVESTORS,
		ReportType:     domain.REPORT_PAYMENT_STATEMENT,
		Year:           "2023",
		Quarter:        "",
		FileUrl:        "https://staturoy-cdn-sb-jp-dev.illuminarean.com/public/statutory_books/ledger_file/payment_statement/user/6/2024-02-15/228e5f3e-a09c-4cf4-acb1-eb50fed978bb.pdf",
		AtOpen:         jsonutils.JSONTime(time.Date(2024, 8, 29, 9, 34, 19, 0, time.UTC)),
		Created:        jsonutils.JSONTime(time.Date(2024, 2, 15, 1, 21, 41, 0, time.UTC)),
		Updated:        jsonutils.JSONTime(time.Date(2024, 8, 29, 9, 34, 19, 0, time.UTC)),
		Deleted:        jsonutils.JSONTime(time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)),
	}

	mockReportPaymentStatementOutput3 = domain.ReportOutput{
		Id:             1003,
		UserId:         3,
		SpvId:          0,
		IsOpen:         true,
		Title:          "法定帳簿テスト用開発会社 2024年 支払調書",
		SubmissionType: domain.SUBMISSION_INVESTORS,
		ReportType:     domain.REPORT_PAYMENT_STATEMENT,
		Year:           "2024",
		Quarter:        "",
		FileUrl:        "https://staturoy-cdn-sb-jp-dev.illuminarean.com/public/statutory_books/ledger_file/payment_statement/user/6/2024-02-15/56292ec9-645d-42fb-86ce-d9761e590e48.pdf",
		AtOpen:         jsonutils.JSONTime(time.Date(2024, 8, 29, 9, 34, 19, 0, time.UTC)),
		Created:        jsonutils.JSONTime(time.Date(2024, 2, 15, 1, 21, 42, 0, time.UTC)),
		Updated:        jsonutils.JSONTime(time.Date(2024, 8, 29, 9, 34, 19, 0, time.UTC)),
		Deleted:        jsonutils.JSONTime(time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)),
	}

	mockReportPaymentStatementOutputNotOpened = domain.ReportOutput{
		Id:             1004,
		UserId:         4,
		SpvId:          0,
		IsOpen:         false,
		Title:          "すべての法人 2024年 支払調書",
		SubmissionType: domain.SUBMISSION_INVESTORS,
		ReportType:     domain.REPORT_PAYMENT_STATEMENT,
		Year:           "2024",
		Quarter:        "",
		FileUrl:        "https://staturoy-cdn-sb-jp-dev.illuminarean.com/public/statutory_books/ledger_file/payment_statement/user/107/2024-02-15/228e5f3e-a09c-4cf4-acb1-eb50fed978bb.pdf",
		AtOpen:         jsonutils.JSONTime(time.Date(2024, 8, 29, 9, 34, 19, 0, time.UTC)),
		Created:        jsonutils.JSONTime(time.Date(2024, 2, 15, 1, 21, 42, 0, time.UTC)),
		Updated:        jsonutils.JSONTime(time.Date(2024, 8, 29, 9, 34, 19, 0, time.UTC)),
		Deleted:        jsonutils.JSONTime(time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)),
	}
)

// REPORT_TRANSACTION_BALANCE_ADMIN_REPORT
var (
	mockReportTransactionBalanceReportOutput1 = domain.ReportOutput{
		Id:             1005,
		UserId:         5,
		SpvId:          0,
		IsOpen:         true,
		Title:          "energyshares 2024年 支払調書",
		SubmissionType: domain.SUBMISSION_INVESTORS,
		ReportType:     domain.REPORT_TRANSACTION_BALANCE_ADMIN_REPORT,
		Year:           "2023",
		Quarter:        domain.QUARTER_FIRST,
		FileUrl:        "https://staturoy-cdn-sb-jp-dev.illuminarean.com/public/statutory_books/ledger_file/payment_statement/user/6/2024-02-15/0ea90e56-57ca-42c2-9d9e-6a82533f0fcb.pdf",
		AtOpen:         jsonutils.JSONTime(time.Date(2024, 8, 29, 9, 34, 19, 0, time.UTC)),
		Created:        jsonutils.JSONTime(time.Date(2024, 2, 15, 1, 21, 41, 0, time.UTC)),
		Updated:        jsonutils.JSONTime(time.Date(2024, 8, 29, 9, 34, 19, 0, time.UTC)),
		Deleted:        jsonutils.JSONTime(time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)),
	}

	mockReportTransactionBalanceReportOutput2 = domain.ReportOutput{
		Id:             1006,
		UserId:         6,
		SpvId:          0,
		IsOpen:         true,
		Title:          "すべての法人 2024年 支払調書",
		SubmissionType: domain.SUBMISSION_INVESTORS,
		ReportType:     domain.REPORT_TRANSACTION_BALANCE_ADMIN_REPORT,
		Year:           "2023",
		Quarter:        domain.QUARTER_FIRST,
		FileUrl:        "https://staturoy-cdn-sb-jp-dev.illuminarean.com/public/statutory_books/ledger_file/payment_statement/user/6/2024-02-15/228e5f3e-a09c-4cf4-acb1-eb50fed978bb.pdf",
		AtOpen:         jsonutils.JSONTime(time.Date(2024, 8, 29, 9, 34, 19, 0, time.UTC)),
		Created:        jsonutils.JSONTime(time.Date(2024, 2, 15, 1, 21, 41, 0, time.UTC)),
		Updated:        jsonutils.JSONTime(time.Date(2024, 8, 29, 9, 34, 19, 0, time.UTC)),
		Deleted:        jsonutils.JSONTime(time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)),
	}

	mockReportTransactionBalanceReportOutput3 = domain.ReportOutput{
		Id:             1007,
		UserId:         7,
		SpvId:          0,
		IsOpen:         true,
		Title:          "法定帳簿テスト用開発会社 2024年 支払調書",
		SubmissionType: domain.SUBMISSION_INVESTORS,
		ReportType:     domain.REPORT_TRANSACTION_BALANCE_ADMIN_REPORT,
		Year:           "2024",
		Quarter:        domain.QUARTER_THIRD,
		FileUrl:        "https://staturoy-cdn-sb-jp-dev.illuminarean.com/public/statutory_books/ledger_file/payment_statement/user/6/2024-02-15/56292ec9-645d-42fb-86ce-d9761e590e48.pdf",
		AtOpen:         jsonutils.JSONTime(time.Date(2024, 8, 29, 9, 34, 19, 0, time.UTC)),
		Created:        jsonutils.JSONTime(time.Date(2024, 2, 15, 1, 21, 42, 0, time.UTC)),
		Updated:        jsonutils.JSONTime(time.Date(2024, 8, 29, 9, 34, 19, 0, time.UTC)),
		Deleted:        jsonutils.JSONTime(time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)),
	}

	mockReportTransactionBalanceReportOutput4 = domain.ReportOutput{
		Id:             1008,
		UserId:         7,
		SpvId:          8,
		IsOpen:         true,
		Title:          "すべての法人 2024年 支払調書",
		SubmissionType: domain.SUBMISSION_INVESTORS,
		ReportType:     domain.REPORT_TRANSACTION_BALANCE_ADMIN_REPORT,
		Year:           "2024",
		Quarter:        domain.QUARTER_THIRD,
		FileUrl:        "https://staturoy-cdn-sb-jp-dev.illuminarean.com/public/statutory_books/ledger_file/payment_statement/user/107/2024-02-15/228e5f3e-a09c-4cf4-acb1-eb50fed978bb.pdf",
		AtOpen:         jsonutils.JSONTime(time.Date(2024, 8, 29, 9, 34, 19, 0, time.UTC)),
		Created:        jsonutils.JSONTime(time.Date(2024, 2, 15, 1, 21, 42, 0, time.UTC)),
		Updated:        jsonutils.JSONTime(time.Date(2024, 8, 29, 9, 34, 19, 0, time.UTC)),
		Deleted:        jsonutils.JSONTime(time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)),
	}
)

// db factory mock
type dbFactoryMock struct{}

func (d dbFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	xormEngine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	xormEngine.Sync(new(domain.ReportOutput))
	xormEngine.ShowSQL(false)

	// テストデータ挿入
	xormEngine.Insert(
		mockReportPaymentStatementOutput1,
		mockReportPaymentStatementOutput2,
		mockReportPaymentStatementOutput3,
		mockReportPaymentStatementOutputNotOpened,
		mockReportTransactionBalanceReportOutput1,
		mockReportTransactionBalanceReportOutput2,
		mockReportTransactionBalanceReportOutput3,
		mockReportTransactionBalanceReportOutput4,
	)
	return xormEngine, nil
}

// インターフェース満足のために必要
func (d dbFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	panic("implement me")
}
