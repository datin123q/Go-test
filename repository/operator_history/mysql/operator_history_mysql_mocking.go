package mysql

import (
	"context"

	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

// db factory mock
type dbFactoryMock struct{}

var (
	mockOperatorHistory1 = domain.OperatorHistory{
		Id:     1000,
		UserId: 2000,
		Name:   "コ・ギルドン",
		Email:  "kildong@email.com",
		Menu:   "menu.manage_product.funding_request_detail",
		Ip:     "127.0.0.1",
		Title:  "テストページ",
		State:  domain.LOOKUP,
	}
)

var (
	adminUser1 = domain.AdminUser{
		Id:    2000,
		Email: "kildong@email.com",
		Name:  "コ・ギルドン",
		Roles: []string{"ROLE_SUPER_ADMIN", "ROLE_ADMIN", "ROLE_OPERATOR"},
	}
)

func (d dbFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	xormEngine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	xormEngine.Sync(new(domain.OperatorHistory), new(domain.AdminUser))
	xormEngine.ShowSQL(false)

	// テストデータ挿入
	xormEngine.Insert(mockOperatorHistory1, adminUser1)

	return xormEngine, nil
}

func (d dbFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	panic("implement me")
}
