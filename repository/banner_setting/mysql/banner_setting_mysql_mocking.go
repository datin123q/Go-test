package mysql

import (
	"context"

	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

var (
	mockBannerSetting1 = domain.BannerSetting{
		Id:           1,
		SettingKey:   string(domain.ROLLING_SECOND),
		SettingValue: "5",
	}
)

type dbFactoryMock struct {
}

func (d dbFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	xormEngine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	xormEngine.Sync(new(domain.BannerSetting))
	xormEngine.ShowSQL(true)

	xormEngine.Insert(mockBannerSetting1)

	return xormEngine, nil
}

func (d dbFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	panic("implement me")
}
