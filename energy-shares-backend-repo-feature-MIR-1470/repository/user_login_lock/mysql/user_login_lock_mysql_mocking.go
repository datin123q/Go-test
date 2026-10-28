package mysql

import (
	"context"
	"time"

	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

var (
	// モックレコードはUTCで保存されるため、ここでローカルタイムを適用する必要はない
	baseTime   = time.Now()
	startOfDay = time.Date(baseTime.Year(), baseTime.Month(), baseTime.Day(), 0, 0, 0, 0, baseTime.Location())
	endOfDay   = time.Date(baseTime.Year(), baseTime.Month(), baseTime.Day(), 23, 59, 59, 0, baseTime.Location())
)

var (
	mockUserLoginLock1 = domain.UserLoginLock{
		Id:           1,
		UserId:       1001,
		LoginFailCnt: 1,
	}
	mockUserLoginLock2 = domain.UserLoginLock{
		Id:           2,
		UserId:       1002,
		LoginFailCnt: 4,
	}
	mockUserLoginLock3 = domain.UserLoginLock{
		Id:           3,
		UserId:       1003,
		LoginFailCnt: 7,
		LockAt:       jsonutils.JSONTime(baseTime.Add(30 * time.Minute)),
	}
)

type dbFactoryMock struct{}

func (d dbFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	xormEngine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	xormEngine.Sync(new(domain.UserLoginLock))
	xormEngine.ShowSQL(true)

	// テストデータ挿入
	xormEngine.Insert(mockUserLoginLock1, mockUserLoginLock2, mockUserLoginLock3)

	return xormEngine, nil
}

func (d dbFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	panic("implement me")
}
