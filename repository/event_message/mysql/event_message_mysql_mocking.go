package mysql

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"xorm.io/xorm"
)

// db factory mock
type dbFactoryMock struct{}

var (
	mockEventMessage1 = domain.EventMessage{
		Category:      domain.EVENT_FUND_TRANSFER,
		Content:       `{"text" : "hello world"}`,
		ContentString: "",
		Uuid:          "29f24996-e3ae-4012-b25e-b31773c6cbfb",
		IsPublish:     true,
	}
)

func (d dbFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	xormEngine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	xormEngine.Sync(new(domain.EventMessage))
	xormEngine.ShowSQL(false)

	// テストデータを挿入
	xormEngine.Insert(mockEventMessage1)

	return xormEngine, nil
}

func (d dbFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	panic("implement me")
}
