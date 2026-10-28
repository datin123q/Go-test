package mysql

import (
	"context"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"time"
	"xorm.io/xorm"
)

// db factory mock
type dbFactoryMock struct{}

var (
	expireDepositDate = jsonutils.JSONTime(time.Now().AddDate(0, 0, 1))

	mockShare1 = domain.Share{
		Id: "SHARE_ID1",
	}

	mockDepositRequest1 = domain.DepositRequest{
		Id:                1000,
		ShareId:           "SHARE_ID1",
		UserId:            2000,
		ExpireDepositDate: expireDepositDate,
	}
)

func (d dbFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	xormEngine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	xormEngine.Sync(new(domain.Share), new(domain.DepositRequest))
	xormEngine.ShowSQL(false)

	// テストデータを挿入
	xormEngine.Insert(mockShare1, mockDepositRequest1)

	return xormEngine, nil
}

func (d dbFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	panic("implement me")
}
