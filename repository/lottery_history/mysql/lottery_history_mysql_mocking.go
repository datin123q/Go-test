package mysql

import (
	"context"
	"time"

	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

// db factory mock
type dbFactoryMock struct{}

var (
	now = time.Date(2026, 4, 14, 10, 0, 0, 0, time.UTC)

	mockLotteryApplication1 = domain.LotteryApplication{
		Id:         "app-001",
		ProjectId:  1,
		UserId:     100,
		Qty:        10,
		UnitPrice:  1000,
		Amount:     10000,
		AppliedAt:  jsonutils.JSONTime(now),
		CanceledAt: jsonutils.JSONTime{},
	}
)

func (d dbFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	xormEngine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	xormEngine.Sync(new(domain.LotteryApplication), new(domain.LotteryHistory))
	xormEngine.ShowSQL(false)

	// テストデータ挿入
	xormEngine.Insert(mockLotteryApplication1)

	return xormEngine, nil
}

func (d dbFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	panic("implement me")
}
