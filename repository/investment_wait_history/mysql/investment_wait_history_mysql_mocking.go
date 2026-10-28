package mysql

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"xorm.io/xorm"
)

// db factory mock
type dbFactoryMock struct{}

var (
	mockInvestmentWaitHistory1 = domain.InvestmentWaitHistory{
		Id:               1000,
		InvestmentWaitId: 2000,
		Name:             "ホン・ギルドン",
	}

	mockInvestmentWaitHistory2 = domain.InvestmentWaitHistory{
		Id:               1001,
		InvestmentWaitId: 2000,
		Name:             "ホン・ギルドン",
	}

	mockInvestmentWaitHistory3 = domain.InvestmentWaitHistory{
		Id:               1002,
		InvestmentWaitId: 2001,
		Name:             "ドゥリ",
	}
)

func (d dbFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	xormEngine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	xormEngine.Sync(new(domain.InvestmentWaitHistory))
	xormEngine.ShowSQL(false)

	// テストデータ挿入
	xormEngine.Insert(mockInvestmentWaitHistory1, mockInvestmentWaitHistory2, mockInvestmentWaitHistory3)

	return xormEngine, nil
}

func (d dbFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	panic("implement me")
}
