package mysql

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"xorm.io/xorm"
)

var (
	mockWithdrawalCloseMapping1 = domain.WithdrawalCloseMapping{
		Id:           1,
		CloseId:      1,
		WithdrawalId: 1,
	}

	mockWithdrawalCloseMapping2 = domain.WithdrawalCloseMapping{
		Id:           2,
		CloseId:      1,
		WithdrawalId: 2,
	}

	mockWithdrawalCloseMapping3 = domain.WithdrawalCloseMapping{
		Id:           3,
		CloseId:      2,
		WithdrawalId: 3,
	}
)

type dbFactoryMock struct {
}

func (d dbFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	xormEngine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	xormEngine.ShowSQL(true)

	xormEngine.Sync(new(domain.WithdrawalCloseMapping))

	xormEngine.Insert(mockWithdrawalCloseMapping1, mockWithdrawalCloseMapping2, mockWithdrawalCloseMapping3)

	return xormEngine, nil
}

func (d dbFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	panic("implement me")
}
