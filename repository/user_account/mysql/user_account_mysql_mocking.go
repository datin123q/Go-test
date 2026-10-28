package mysql

import (
	"context"
	"github.com/samber/lo"

	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

// db factory mock
type dbFactoryMock struct{}

var (
	mockUser1 = domain.User{
		Id:       1000,
		Uid:      "UID1",
		CommonId: "410d3741-7180-4024-a969-648723a57bbe",
		Email:    "gildong@email.com",
	}

	mockUserAccount1 = domain.UserAccount{
		Id:       2000,
		UserId:   1000,
		UserName: "고길동",
		UserType: domain.PRIVATE,
		Balance:  lo.ToPtr(int64(100000)),
		Version:  1,
	}

	mockUserAccountToInsert = domain.UserAccount{
		UserId:   2000,
		UserName: "김둘리",
		UserType: domain.PRIVATE,
		Balance:  lo.ToPtr(int64(200000)),
		Version:  2,
	}
)

func (d dbFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	xormEngine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	xormEngine.Sync(new(domain.User), new(domain.UserAccount))
	xormEngine.ShowSQL(false)

	// テストデータ挿入
	xormEngine.Insert(mockUser1, mockUserAccount1)

	return xormEngine, nil
}

func (d dbFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	panic("implement me")
}
