package mysql

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"xorm.io/xorm"
)

var (
	mockTransfer1 = domain.Transfer{
		Id:                    int64(1),
		ItemId:                1,
		RequestId:             int64(1),
		TransactionId:         "cr3csfclss4s73e2qa8g",
		UserId:                int64(1),
		TransferAmount:        int64(3000),
		AccountNumber:         "123456",
		BankCode:              "0310",
		BankName:              "ゆうちょ銀行",
		BranchCode:            "019",
		BranchName:            "〇一九",
		Depositor:             "ｺﾊﾞﾔｼ ｴﾘﾄ",
		AccountTypeCode:       "1",
		TransferFee:           int64(0),
		TransactionDetailType: domain.WITHDRAWAL_INVESTOR,
		TransferStatus:        domain.TRANSFER_STATUS_SUCCESS,
	}

	mockTransfer2 = domain.Transfer{
		Id:                    int64(2),
		ItemId:                1,
		RequestId:             int64(2),
		TransactionId:         "cmimpr901u34v3gekemg",
		UserId:                int64(2),
		TransferAmount:        int64(3500),
		AccountNumber:         "3293",
		BankCode:              "0310",
		BankName:              "ゆうちょ銀行",
		BranchCode:            "019",
		BranchName:            "〇一九",
		Depositor:             "ｺﾊﾞﾔｼ ｴﾘﾄ",
		AccountTypeCode:       "1",
		TransferFee:           int64(0),
		TransactionDetailType: domain.WITHDRAWAL_INVESTOR,
		TransferStatus:        domain.TRANSFER_STATUS_SUCCESS,
	}

	mockTransfer3 = domain.Transfer{
		Id:                    int64(3),
		ItemId:                1,
		RequestId:             int64(1),
		TransactionId:         "ch152rnmq22vjpim47l0",
		UserId:                int64(1),
		TransferAmount:        int64(5000),
		AccountNumber:         "123456",
		BankCode:              "0310",
		BankName:              "ゆうちょ銀行",
		BranchCode:            "019",
		BranchName:            "〇一九",
		Depositor:             "ｺﾊﾞﾔｼ ｴﾘﾄ",
		AccountTypeCode:       "1",
		TransferFee:           int64(0),
		TransactionDetailType: domain.DEPOSIT_REPAYMENT,
		TransferStatus:        domain.TRANSFER_STATUS_FAIL,
	}

	mockWithdrawalTransferMapping1 = domain.WithdrawalTransferMapping{
		Id:            int64(1),
		TransactionId: "cr3csfclss4s73e2qa8g",
		WithdrawalId:  int64(1),
	}
	mockWithdrawalTransferMapping2 = domain.WithdrawalTransferMapping{
		Id:            int64(2),
		TransactionId: "cmimpr901u34v3gekemg",
		WithdrawalId:  int64(2),
	}
)

type dbFactoryMock struct{}

func (d dbFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	xormEngine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	xormEngine.ShowSQL(false)
	xormEngine.Sync(new(domain.Transfer), new(domain.WithdrawalTransferMapping))
	xormEngine.Insert(mockTransfer1, mockTransfer2, mockTransfer3)
	xormEngine.Insert(mockWithdrawalTransferMapping1, mockWithdrawalTransferMapping2)

	return xormEngine, nil
}

func (d dbFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	panic("implement me")
}
