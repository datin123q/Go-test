package mysql

import (
	"context"
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"xorm.io/xorm"
)

type dbFactoryMock struct {
	db xorm.Interface
}

var (
	// withdrawal
	mockCompleteWithdrawal = domain.Withdrawal{
		Id:           1,
		InvestorId:   "1",
		InvestorName: "investor1",
		Amount:       1000,
		Fee:          10,
		RequestDate:  jsonutils.JSONTime(time.Now().AddDate(0, 0, -5)),
		CompleteDate: jsonutils.JSONTime(time.Now().AddDate(0, 0, -5)),
		Status:       domain.WITHDRAWAL_STATUS_COMPLETE,
		IsManual:     false,
		Created:      jsonutils.JSONTime(time.Now().AddDate(0, 0, -5)),
	}

	mockInCompleteWithdrawal = domain.Withdrawal{
		Id:           2,
		InvestorId:   "2",
		InvestorName: "investor2",
		Amount:       2000,
		Fee:          25,
		RequestDate:  jsonutils.JSONTime(time.Now().AddDate(0, 0, -1)),
		CompleteDate: jsonutils.JSONTime(time.Now().AddDate(0, 0, -1)),
		Status:       domain.WITHDRAWAL_STATUS_INCOMPLETE,
		IsManual:     false,
		Created:      jsonutils.JSONTime(time.Now().AddDate(0, 0, -1)),
	}

	mockWaitWithdrawal = domain.Withdrawal{
		Id:           3,
		InvestorId:   "1",
		InvestorName: "investor1",
		Amount:       3000,
		Fee:          30,
		RequestDate:  jsonutils.JSONTime(time.Now().AddDate(0, 0, -1)),
		CompleteDate: jsonutils.JSONTime(time.Now().AddDate(0, 0, -1)),
		Status:       domain.WITHDRAWAL_STATUS_WAIT,
		IsManual:     true,
		Created:      jsonutils.JSONTime(time.Now().AddDate(0, 0, -1)),
	}

	// user
	mockUser1 = domain.User{
		Id:    1,
		Uid:   "uid1",
		Email: "user1@example.com",
	}

	mockUser2 = domain.User{
		Id:    2,
		Uid:   "uid2",
		Email: "user2@example.com",
	}

	// withdrawal_close_mapping
	mockWithdrawalCloseMapping1 = domain.WithdrawalCloseMapping{
		Id:           1,
		WithdrawalId: 1,
		CloseId:      1,
	}

	mockWithdrawalCloseMapping2 = domain.WithdrawalCloseMapping{
		Id:           2,
		WithdrawalId: 2,
		CloseId:      1,
	}

	mockWithdrawalCloseMapping3 = domain.WithdrawalCloseMapping{
		Id:           3,
		WithdrawalId: 3,
		CloseId:      2,
	}

	// withdrawal_transfer_mapping
	mockWithdrawalTransferMapping1 = domain.WithdrawalTransferMapping{
		Id:            1,
		TransactionId: "transaction_id_1",
		WithdrawalId:  1,
	}
)

func (d *dbFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	if d.db != nil {
		return d.db, nil
	}

	xormEngine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	xormEngine.ShowSQL(false)

	xormEngine.Sync(new(domain.Withdrawal))
	xormEngine.Sync(new(domain.User))
	xormEngine.Sync(new(domain.WithdrawalCloseMapping))
	xormEngine.Sync(new(domain.WithdrawalTransferMapping))

	xormEngine.Insert(mockCompleteWithdrawal, mockInCompleteWithdrawal, mockWaitWithdrawal)
	xormEngine.Insert(mockWithdrawalCloseMapping1, mockWithdrawalCloseMapping2, mockWithdrawalCloseMapping3)
	xormEngine.Insert(mockUser1, mockUser2)
	xormEngine.Insert(mockWithdrawalTransferMapping1)

	d.db = xormEngine
	return xormEngine, nil
}

func (d *dbFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	panic("implement me")
}
