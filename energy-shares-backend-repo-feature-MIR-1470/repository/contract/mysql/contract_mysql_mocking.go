package mysql

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"xorm.io/xorm"
)

// db factory mock
type dbFactoryMock struct{}

func (d dbFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	xormEngine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	xormEngine.Sync(new(domain.Contract))
	xormEngine.ShowSQL(false)

	// テストデータを挿入
	xormEngine.Insert(
		domain.Contract{
			ShareId:     "SHARE_ID1",
			ShareQty:    4,
			ShareAmount: 1000,
			Status:      domain.CONTRACT_SIGNED,
			ProjectId:   1234,
			UserName:    "Jinyong Kim",
		},

		domain.Contract{
			ShareId:     "SHARE_ID2",
			ShareQty:    5,
			ShareAmount: 2000,
			Status:      domain.CONTRACT_SIGNED,
			ProjectId:   1235,
			UserName:    "Jinyong Kim",
		},
	)

	return xormEngine, nil
}

func (d dbFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	panic("implement me")
}
