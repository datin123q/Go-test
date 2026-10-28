package mysql

import (
	"context"

	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type dbFactoryMock struct{}

var (
	mockMapSmsReceiver = domain.MapSmsReceiver{
		Id:       1000,
		SmsId:    1,
		Receiver: "01012345678",
	}
)

func (d dbFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	xormEngine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	xormEngine.Sync(new(domain.MapSmsReceiver))
	xormEngine.ShowSQL(false)

	// テストデータ挿入
	xormEngine.Insert(mockMapSmsReceiver)

	return xormEngine, nil
}

func (d dbFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	panic("implement me")
}
