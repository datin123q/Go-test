package mysql

import (
	"context"

	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

// db factory mock
type dbFactoryMock struct{}

var (
	mockUserAntisocialCheck1 = domain.UserAntisocialCheck{
		CheckUserId:   1000,
		Checker:       "checker1",
		InvalidReason: "反社会 1",
	}
	mockUserAntisocialCheck2 = domain.UserAntisocialCheck{
		CheckUserId:   2000,
		Checker:       "checker1",
		InvalidReason: "反社会 1",
	}
)

func (d dbFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	xormEngine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	xormEngine.Sync(new(domain.UserAntisocialCheck))
	xormEngine.ShowSQL(false)

	// テストデータ挿入
	xormEngine.Insert(mockUserAntisocialCheck1)

	return xormEngine, nil
}

func (d dbFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	panic("implement me")
}
