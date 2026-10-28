package mysql

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"xorm.io/xorm"
)

// db factory mock
type dbFactoryMock struct{}

var (
	mockUserTransaction1 = domain.UserTransaction{
		Id:                    1000,
		UserId:                2000,
		UserType:              domain.PRIVATE,
		UserName:              "고길동",
		ProjectId:             3000,
		AccountNumber:         "1234-5678-9101112",
		ProjectType:           domain.EQUITY,
		Amount:                10_000_000,
		Balance:               50_000,
		TransactionType:       domain.DEPOSIT_TRANSACTION,
		TransactionDetailType: domain.DEPOSIT_WITHDRAWAL_CANCEL,
	}
)

func (d dbFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	xormEngine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	xormEngine.Sync(new(domain.UserTransaction))
	xormEngine.ShowSQL(false)

	// テストデータ挿入
	xormEngine.Insert(mockUserTransaction1)

	return xormEngine, nil
}

func (d dbFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	panic("implement me")
}
