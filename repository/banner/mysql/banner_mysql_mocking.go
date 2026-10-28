package mysql

import (
	"context"
	"time"

	"github.com/samber/lo"
	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

var (
	mockBanner1 = domain.Banner{
		Id:          1,
		Title:       "test banner 1",
		StartDate:   jsonutils.JSONTime(time.Now().AddDate(0, 0, -5)),
		EndDate:     jsonutils.JSONTime(time.Now().AddDate(0, 0, 5)),
		IsPosted:    true,
		SortOrder:   lo.ToPtr(int64(1)),
		Description: "test banner description",
		NewTab:      false,
		Created:     jsonutils.JSONTime(time.Now().AddDate(0, 0, -20)),
	}

	mockBanner2 = domain.Banner{
		Id:        2,
		Title:     "test banner 2",
		StartDate: jsonutils.JSONTime(time.Now().AddDate(0, 0, -1)),
		EndDate:   jsonutils.JSONTime(time.Now().AddDate(0, 0, 1)),
		IsPosted:  true,
		SortOrder: lo.ToPtr(int64(2)),
		NewTab:    true,
		Created:   jsonutils.JSONTime(time.Now().AddDate(0, 0, -10)),
	}

	mockBanner3 = domain.Banner{
		Id:          3,
		Title:       "for update",
		StartDate:   jsonutils.JSONTime(time.Now().AddDate(0, 0, -1)),
		EndDate:     jsonutils.JSONTime(time.Now().AddDate(0, 0, 1)),
		IsPosted:    false,
		SortOrder:   nil,
		Description: "for update test",
		NewTab:      true,
		Created:     jsonutils.JSONTime(time.Now().AddDate(0, 0, -5)),
	}

	mockBanner4 = domain.Banner{
		Id:          4,
		Title:       "for delete",
		StartDate:   jsonutils.JSONTime(time.Now().AddDate(0, 0, -1)),
		EndDate:     jsonutils.JSONTime(time.Now().AddDate(0, 0, 1)),
		IsPosted:    false,
		SortOrder:   nil,
		Description: "for delete test",
		NewTab:      true,
		Created:     jsonutils.JSONTime(time.Now().AddDate(0, 0, -5)),
	}
)

type dbFactoryMock struct {
}

func (d dbFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	xormEngine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	xormEngine.Sync(new(domain.Banner))
	xormEngine.ShowSQL(false)

	xormEngine.Insert(mockBanner1, mockBanner2, mockBanner3, mockBanner4)

	return xormEngine, nil
}

func (d dbFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	panic("implement me")
}
