package mysql

import (
	"context"

	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

// db factory mock
type dbFactoryMock struct{}

var (
	mockUserVirtualAccount1 = domain.UserVirtualAccount{
		Id:           1000,
		UserId:       2000,
		ProjectType:  domain.EQUITY,
		LinkedStatus: domain.ON_VA_ACCOUNT,
	}
	mockUserVirtualAccount2 = domain.UserVirtualAccount{
		Id:           1001,
		UserId:       2001,
		ProjectType:  domain.EQUITY,
		LinkedStatus: domain.OFF_VA_ACCOUNT,
	}
)

func (d dbFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	xormEngine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	xormEngine.Sync(new(domain.UserVirtualAccount))
	xormEngine.ShowSQL(false)

	// テストデータ挿入
	xormEngine.Insert(mockUserVirtualAccount1)
	xormEngine.Insert(mockUserVirtualAccount2)

	return xormEngine, nil
}

func (d dbFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	panic("implement me")
}
