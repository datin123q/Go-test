package mysql

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"xorm.io/xorm"
)

var (
	mockPolicy1 = domain.Policy{
		Id:             1,
		CommonId:       "eeae3a9d-1ca8-43c6-a804-a235983bd102",
		OwnerId:        1,
		ModifierId:     1,
		PolicyViewType: domain.POLICY_VIEW_JOIN,
		PolicyType:     domain.POLICY_PROTECTION,
		ContentHtml:    "content",
		ModifierEmail:  "admin@energyshares.jp",
	}

	mockPolicy2 = domain.Policy{
		Id:             2,
		CommonId:       "c2c546ab-0715-4527-a40f-16c126fb1c6b",
		OwnerId:        1,
		ModifierId:     1,
		PolicyViewType: domain.POLICY_VIEW_FOOTER,
		PolicyType:     domain.POLICY_CAUTION,
		ContentHtml:    "content",
		ModifierEmail:  "admin@energyshares.jp",
	}
)

type dbFactoryMock struct{}

func (d dbFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	xormEngine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	xormEngine.ShowSQL(false)

	xormEngine.Sync(new(domain.Policy))

	xormEngine.Insert(mockPolicy1, mockPolicy2)

	return xormEngine, nil
}

func (d dbFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	panic("implement me")
}
