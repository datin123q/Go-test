package mysql

import (
	"context"

	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type dbFactoryMock struct{}

var (
	mockRole1 = domain.Role{
		Id:       1,
		UserId:   1000,
		RoleCode: string(domain.ROLE_ADMIN),
	}
	mockRole2 = domain.Role{
		Id:       2,
		UserId:   1000,
		RoleCode: string(domain.ROLE_INVESTOR),
	}
	mockRole3 = domain.Role{
		Id:       3,
		UserId:   1000,
		RoleCode: string(domain.ROLE_MYNUMBER),
	}
	mockRole4 = domain.Role{
		Id:       4,
		UserId:   2000,
		RoleCode: string(domain.ROLE_ADMIN),
	}
)

func (d dbFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	xormEngine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	xormEngine.Sync(domain.Role{})
	xormEngine.ShowSQL(false)

	// テストデータ挿入
	xormEngine.Insert(mockRole1, mockRole2, mockRole3, mockRole4)

	return xormEngine, nil
}

func (d dbFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	panic("implement me")
}
