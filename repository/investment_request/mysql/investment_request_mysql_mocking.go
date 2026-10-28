package mysql

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"xorm.io/xorm"
)

// db factory mock
type dbFactoryMock struct{}

var (
	mockInvestmentRequest1 = domain.InvestmentRequest{
		Id:                  1000,
		InvestmentWaitersId: 2000,
		ProjectId:           3000,
		IsExpire:            false,
	}
)

func (d dbFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	xormEngine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	xormEngine.Sync(new(domain.InvestmentRequest))
	xormEngine.ShowSQL(false)

	// テストデータを挿入
	xormEngine.Insert(mockInvestmentRequest1)

	return xormEngine, nil
}

func (d dbFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	panic("implement me")
}
