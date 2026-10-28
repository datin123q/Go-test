package mysql

import (
	"context"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"xorm.io/xorm"
)

// db factory mock
type dbFactoryMock struct{}

var (
	mockUserId1 = int64(1000)

	mockCorporateShareholder1 = domain.CorporateShareholder{
		Id:     1,
		UserId: mockUserId1,
		Uid:    "UID1",
	}
)

func (d dbFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	xormEngine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	xormEngine.Sync(new(domain.CorporateShareholder))
	xormEngine.ShowSQL(false)

	// テストデータを挿入
	xormEngine.Insert(mockCorporateShareholder1)

	return xormEngine, nil
}

func (d dbFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	panic("implement me")
}
