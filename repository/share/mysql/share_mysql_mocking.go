package mysql

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"xorm.io/xorm"
)

// db factory mock
type dbFactoryMock struct{}

var (
	mockShare1 = domain.Share{
		Id:           "SHARE_ID1",
		OwnerId:      1000,
		ProjectId:    2000,
		InvestStatus: domain.SUCCESS_INVEST,
	}

	mockShare2 = domain.Share{
		Id:           "SHARE_ID2",
		OwnerId:      1001,
		ProjectId:    2000,
		InvestStatus: domain.REQUEST_INVEST,
	}

	mockShare3 = domain.Share{
		Id:           "SHARE_ID3",
		OwnerId:      1002,
		ProjectId:    2000,
		InvestStatus: domain.FAIL_INVEST,
	}
)

func (d dbFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	xormEngine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	xormEngine.Sync(new(domain.Share))
	xormEngine.ShowSQL(false)

	// テストデータ挿入
	xormEngine.Insert(mockShare1, mockShare2, mockShare3)

	return xormEngine, nil
}

func (d dbFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	panic("implement me")
}
