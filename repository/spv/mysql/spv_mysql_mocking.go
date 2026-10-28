package mysql

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"xorm.io/xorm"
)

// db factory mock
type dbFactoryMock struct{}

var (
	mockSpv1 = domain.Spv{
		Id:      1000,
		Name:    "水資源公社",
		Email:   "gildong@email.com",
		Phone:   "010-1234-5678",
		OwnerId: 2000,
	}
)

func (d dbFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	xormEngine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	xormEngine.Sync(new(domain.Spv))
	xormEngine.ShowSQL(false)

	// テストデータ挿入
	xormEngine.Insert(mockSpv1)

	return xormEngine, nil
}

func (d dbFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	panic("implement me")
}
