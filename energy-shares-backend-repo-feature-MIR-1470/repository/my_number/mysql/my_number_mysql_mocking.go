package mysql

import (
	"context"

	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

// db factory mock
type dbFactoryMock struct{}

var (
	mockMyNumber1 = domain.MyNumber{
		Id:       1000,
		CommonId: "1000",
		UserId:   3000,
		Status:   domain.APPROVE,
	}
	mockMyNumber2 = domain.MyNumber{
		Id:       2000,
		CommonId: "2000",
		UserId:   2000,
		Status:   domain.APPROVE,
	}
	mockMyNumberToCreate = domain.MyNumber{
		Id:       1001,
		CommonId: "2001",
		UserId:   3001,
		Status:   domain.APPROVE,
	}
)

func (d dbFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	xormEngine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	xormEngine.Sync(new(domain.MyNumber))
	xormEngine.ShowSQL(false)

	// テストデータ挿入
	xormEngine.Insert(mockMyNumber1, mockMyNumber2)

	return xormEngine, nil
}

func (d dbFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	panic("implement me")
}
