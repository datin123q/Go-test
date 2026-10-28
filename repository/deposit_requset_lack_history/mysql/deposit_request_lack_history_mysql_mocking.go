package mysql

import (
	"context"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"xorm.io/xorm"
)

// db factory mock
type dbFactoryMock struct{}

var (
	mockDepositRequestLackHistory1 = domain.DepositRequestLackHistory{
		Id:               1000,
		UserId:           2000,
		DepositRequestId: 3000,
		Amount:           10000,
		AccountNumber:    "123-123-123",
		ProjectType:      domain.EQUITY,
		ProjectId:        4000,
		ShareId:          "SHARE_ID_1",
	}
)

func (d dbFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	xormEngine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	xormEngine.Sync(new(domain.DepositRequestLackHistory))
	xormEngine.ShowSQL(false)

	// テストデータを挿入
	xormEngine.Insert(mockDepositRequestLackHistory1)

	return xormEngine, nil
}

func (d dbFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	panic("implement me")
}
