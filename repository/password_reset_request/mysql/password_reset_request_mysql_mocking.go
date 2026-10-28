package mysql

import (
	"context"
	"time"

	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

type dbFactoryMock struct{}

var (
	mockPasswordResetRequest1 = domain.PasswordResetRequest{
		Id:        1,
		RequestId: "00000000-0000-0000-0000-000000000000",
		Email:     "gildong@email.com",
		Url:       "https://es-jp-dev.illuminarean.com/reset/password/00000000-0000-0000-0000-000000000000",
		SendCount: 0,
		Expire:    jsonutils.JSONTime(time.Now().Add(time.Minute * 5).In(time.Local)),
	}

	mockPasswordResetRequest2 = domain.PasswordResetRequest{
		Id:        2,
		RequestId: "20000000-0000-0000-0000-000000000000",
		Email:     "gildong@email.com",
		Url:       "https://es-jp-dev.illuminarean.com/reset/password/20000000-0000-0000-0000-000000000000",
		SendCount: 1,
		Expire:    jsonutils.JSONTime(time.Now().Add(time.Minute * 20).In(time.Local)),
	}
)

func (d dbFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	xormEngine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	xormEngine.Sync(new(domain.PasswordResetRequest))
	xormEngine.ShowSQL(true)

	// テストデータ挿入
	xormEngine.Insert(mockPasswordResetRequest1)
	xormEngine.Insert(mockPasswordResetRequest2)

	return xormEngine, nil
}

func (d dbFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	panic("implement me")
}
