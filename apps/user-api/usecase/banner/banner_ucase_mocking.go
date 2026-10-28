package banner

import (
	"time"

	"github.com/samber/lo"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

var (
	mockRollingSecond = domain.BannerSetting{
		Id:           1,
		SettingKey:   string(domain.ROLLING_SECOND),
		SettingValue: "5",
	}

	mockBanner1 = domain.Banner{
		Id:          1,
		Title:       "title1",
		IsPosted:    true,
		StartDate:   jsonutils.JSONTime(time.Now().AddDate(0, 0, -3)),
		EndDate:     jsonutils.JSONTime(time.Now().AddDate(0, 0, 5)),
		Url:         "url1",
		SortOrder:   lo.ToPtr(int64(1)),
		Description: "description1",
	}
	mockBanner2 = domain.Banner{
		Id:          2,
		Title:       "title2",
		IsPosted:    true,
		StartDate:   jsonutils.JSONTime(time.Now().AddDate(0, 0, -3)),
		EndDate:     jsonutils.JSONTime(time.Now().AddDate(0, 0, 5)),
		Url:         "url2",
		SortOrder:   lo.ToPtr(int64(2)),
		Description: "description2",
	}
	mockBanner3 = domain.Banner{
		Id:          3,
		Title:       "title3",
		IsPosted:    true,
		StartDate:   jsonutils.JSONTime(time.Now().AddDate(0, 0, 3)),
		EndDate:     jsonutils.JSONTime(time.Now().AddDate(0, 0, 7)),
		Url:         "url3",
		SortOrder:   lo.ToPtr(int64(3)),
		Description: "description3",
	}
)
