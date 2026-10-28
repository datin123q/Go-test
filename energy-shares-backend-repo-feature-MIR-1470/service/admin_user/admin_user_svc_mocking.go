package admin_user

import (
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

var (
	mockSuperAdmin = domain.AdminUser{
		Id:          1,
		CommonId:    "e00be1bb-1c99-4fe2-8b84-d01783a25aee",
		Email:       "admin@energyshares.jp",
		Name:        "Admin",
		Password:    "password1234",
		Salt:        "salt",
		MobilePhone: "012345678",
		LastLoginDate: func() jsonutils.JSONTime {
			lastLoginDate, _ := time.ParseInLocation(time.DateTime, "2024-02-01 00:00:00", time.UTC)
			return jsonutils.JSONTime(lastLoginDate)
		}(),
		LoginFailCnt: 0,
		RolesString:  "",
		IsSuperAdmin: true,
		RoleUpdated:  jsonutils.JSONTime(time.Now()),
		ValidityStartDate: func() jsonutils.JSONTime {
			validityStartDate, _ := time.ParseInLocation(time.DateTime, "2024-01-01 00:00:00", time.UTC)
			return jsonutils.JSONTime(validityStartDate)
		}(),
		ValidityEndDate: func() jsonutils.JSONTime {
			validityStartDate, _ := time.ParseInLocation(time.DateTime, "2100-12-01 23:59:59", time.UTC)
			return jsonutils.JSONTime(validityStartDate)
		}(),
	}
)
