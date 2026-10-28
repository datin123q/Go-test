package mysql

import (
	"context"
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

// db factory mock
type dbFactoryMock struct{}

var (
	mockEmailCodeRequest1 = domain.EmailCodeRequest{
		Id:       1,
		Email:    "gildong@gmail.com",
		Code:     "178237",
		UserId:   100,
		FirstReq: false,
		Expire:   jsonutils.JSONTime(time.Now().Add(time.Minute * 3)),
	}
)

func (d dbFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	xormEngine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	xormEngine.Sync(new(domain.EmailCodeRequest))
	xormEngine.ShowSQL(false)

	// テストデータを挿入
	xormEngine.Insert(mockEmailCodeRequest1)

	return xormEngine, nil
}

func (d dbFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	panic("implement me")
}
