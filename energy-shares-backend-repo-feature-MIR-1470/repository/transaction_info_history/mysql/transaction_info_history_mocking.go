package mysql

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"xorm.io/xorm"
)

// db factory mock
type dbFactoryMock struct{}

var (
	mockTransactionInfoHistory1 = domain.TransactionInfoHistory{
		Id:            1000,
		ShareId:       "SHAREID_1",
		SubstanceType: string(domain.INVESTMENT_WAIT_HISTORY_STATUS),
		ProcessMethod: domain.USER_HISTORY_PROCESS_MEMBER,
		Name:          "고길동",
		Email:         "gildong@email.com",
	}
)

func (d dbFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	xormEngine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	xormEngine.Sync(new(domain.TransactionInfoHistory))
	xormEngine.ShowSQL(false)

	// テストデータ挿入
	xormEngine.Insert(mockTransactionInfoHistory1)

	return xormEngine, nil
}

func (d dbFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	panic("implement me")
}
