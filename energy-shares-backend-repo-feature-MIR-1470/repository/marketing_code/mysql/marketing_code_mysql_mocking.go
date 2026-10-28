package mysql

import (
	"context"
	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

// db factory mock
type dbFactoryMock struct{}

var (
	mockMarketingCode1 = domain.MarketingCode{
		Id:              1,
		Code:            "marketing_code_1",
		Campaign:        "marketing_campaign_1",
		SignUpRequestId: "sign_up_request_id_1",
		Email:           "gildong@email.com",
		UserId:          100,
	}
)

func (d dbFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	xormEngine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	xormEngine.Sync(new(domain.MarketingCode))
	xormEngine.ShowSQL(false)

	// テストデータ挿入
	xormEngine.Insert(mockMarketingCode1)

	return xormEngine, nil
}

func (d dbFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	panic("implement me")
}
