package mysql

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"xorm.io/xorm"
)

// db factory mock
type dbFactoryMock struct{}

var (
	mockProject1 = domain.Project{
		Id:             1000,
		CommonId:       "0cb05069-19e1-4ea8-8cfd-a50388eec4c5",
		Name:           "水資源プロジェクト",
		CategoryType:   domain.SOLAR,
		SelectionType:  domain.PROJECT_SELECTION_LOTTERY,
		OwnerId:        2000,
		TargetShareQty: 1000,
		IsWaitInvest:   false,
	}

	mockShare1 = domain.Share{
		Id:        "SHARE_ID1",
		ProjectId: 1000,
	}
)

func (d dbFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	xormEngine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	xormEngine.Sync(new(domain.Project), new(domain.Share))

	xormEngine.ShowSQL(false)

	// テストデータ挿入
	xormEngine.Insert(mockProject1, mockShare1)

	return xormEngine, nil
}

func (d dbFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	panic("implement me")
}
