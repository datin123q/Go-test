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
	mockUser1 = domain.User{
		Id:           1000,
		Uid:          "UID1",
		CommonId:     "410d3741-7180-4024-a969-648723a57bbe",
		Email:        "gildong@email.com",
		IsNewsletter: true,
		IsSmsCert:    true,
	}
	mockUser1Role1 = domain.Role{Id: 2000, UserId: 1000, RoleCode: string(domain.ROLE_SUPER_ADMIN)}
	mockUser1Role2 = domain.Role{Id: 2001, UserId: 1000, RoleCode: string(domain.ROLE_ADMIN)}

	mockUser2 = domain.User{
		Id:           2000,
		Uid:          "UID2",
		CommonId:     "410d3741-4321-1234-a969-648723a57bbe",
		Email:        "chulsu@email.com",
		IsNewsletter: false,
		IsSmsCert:    false,
	}

	mockUser3 = domain.User{
		Id:           3000,
		Uid:          "UID3",
		CommonId:     "410d3741-4321-1234-a969-648723a57bba",
		Email:        "minsu@email.com",
		IsNewsletter: false,
		IsSmsCert:    false,
		Deleted:      jsonutils.JSONTime(time.Now()),
	}

	mockUser2Role1 = domain.Role{Id: 2000, UserId: 1000, RoleCode: string(domain.ROLE_USER)}
)

func (d dbFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	xormEngine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	xormEngine.Sync(new(domain.User), new(domain.Role))
	xormEngine.ShowSQL(true)

	// テストデータ挿入
	xormEngine.Insert(mockUser1, mockUser1Role1, mockUser1Role2)
	xormEngine.Insert(mockUser2, mockUser2Role1)
	xormEngine.Insert(mockUser3)

	return xormEngine, nil
}

func (d dbFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	panic("implement me")
}
