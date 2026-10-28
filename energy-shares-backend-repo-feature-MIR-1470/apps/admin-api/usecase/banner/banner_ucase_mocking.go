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

	mockBannerCreate = domain.Banner{
		Id:          1,
		Title:       "title",
		IsPosted:    true,
		StartDate:   jsonutils.JSONTime(time.Date(2024, 1, 1, 0, 0, 0, 0, lo.Must(time.LoadLocation("Asia/Tokyo")))),
		EndDate:     jsonutils.JSONTime(time.Date(2024, 1, 1, 23, 59, 59, 0, lo.Must(time.LoadLocation("Asia/Tokyo")))),
		Url:         "url",
		SortOrder:   lo.ToPtr(int64(1)),
		Description: "",
		NewTab:      true,
	}

	mockBanner1 = domain.Banner{
		Id:          1,
		Title:       "title1",
		IsPosted:    true,
		StartDate:   jsonutils.JSONTime(time.Date(2024, 1, 1, 0, 0, 0, 0, lo.Must(time.LoadLocation("Asia/Tokyo")))),
		EndDate:     jsonutils.JSONTime(time.Date(2024, 1, 2, 23, 59, 59, 0, lo.Must(time.LoadLocation("Asia/Tokyo")))),
		Url:         "url1",
		SortOrder:   lo.ToPtr(int64(1)),
		Description: "description1",
		NewTab:      true,
	}
	mockBanner2 = domain.Banner{
		Id:          2,
		Title:       "title2",
		IsPosted:    true,
		StartDate:   jsonutils.JSONTime(time.Date(2024, 1, 1, 0, 0, 0, 0, lo.Must(time.LoadLocation("Asia/Tokyo")))),
		EndDate:     jsonutils.JSONTime(time.Date(2024, 1, 5, 23, 59, 59, 0, lo.Must(time.LoadLocation("Asia/Tokyo")))),
		Url:         "url2",
		SortOrder:   lo.ToPtr(int64(2)),
		Description: "description2",
		NewTab:      false,
	}
	mockBanner3 = domain.Banner{
		Id:          3,
		Title:       "title3",
		IsPosted:    true,
		StartDate:   jsonutils.JSONTime(time.Date(2024, 3, 1, 0, 0, 0, 0, lo.Must(time.LoadLocation("Asia/Tokyo")))),
		EndDate:     jsonutils.JSONTime(time.Date(2024, 3, 2, 23, 59, 59, 0, lo.Must(time.LoadLocation("Asia/Tokyo")))),
		Url:         "url3",
		SortOrder:   lo.ToPtr(int64(3)),
		Description: "description3",
		NewTab:      true,
	}

	mockBannerUpdateBefore = domain.Banner{
		Id:          1,
		Title:       "title before",
		IsPosted:    true,
		StartDate:   jsonutils.JSONTime(time.Date(2024, 1, 1, 0, 0, 0, 0, lo.Must(time.LoadLocation("Asia/Tokyo")))),
		EndDate:     jsonutils.JSONTime(time.Date(2024, 1, 1, 23, 59, 59, 0, lo.Must(time.LoadLocation("Asia/Tokyo")))),
		Url:         "url after",
		SortOrder:   lo.ToPtr(int64(1)),
		Description: "description before",
		NewTab:      true,
	}
	mockBannerUpdateAfter = domain.Banner{
		Id:          1,
		Title:       "title after",
		IsPosted:    true,
		StartDate:   jsonutils.JSONTime(time.Date(2025, 12, 1, 0, 0, 0, 0, lo.Must(time.LoadLocation("Asia/Tokyo")))),
		EndDate:     jsonutils.JSONTime(time.Date(2025, 12, 31, 23, 59, 59, 0, lo.Must(time.LoadLocation("Asia/Tokyo")))),
		Url:         "url before",
		SortOrder:   lo.ToPtr(int64(5)),
		Description: "description after",
		NewTab:      false,
	}
)
