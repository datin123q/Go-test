package admin_user

import (
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

var (
	mockSuperAdmin = domain.AdminUser{
		Id:            int64(1),
		CommonId:      "commonId-super-admin",
		Email:         "admin@energyshares.jp",
		Name:          "super-admin",
		Password:      "password1234",
		Salt:          "salt",
		Roles:         []string{"ROLE_SUPER_ADMIN", "ROLE_ADMIN"},
		MobilePhone:   "01012341234",
		LastLoginDate: jsonutils.JSONTime(time.Now()),
		LoginFailCnt:  0,
		IsSuperAdmin:  true,
		RoleUpdated:   jsonutils.JSONTime(time.Now()),
		ValidityStartDate: func() jsonutils.JSONTime {
			validityStartDate, _ := time.ParseInLocation(time.DateTime, "2024-02-09 00:00:00", time.Local)
			return jsonutils.JSONTime(validityStartDate)
		}(),
		ValidityEndDate: func() jsonutils.JSONTime {
			validityStartDate, _ := time.ParseInLocation(time.DateTime, "2039-12-19 23:59:59", time.Local)
			return jsonutils.JSONTime(validityStartDate)
		}(),
	}
	mockAdminUser = domain.AdminUser{
		Id:            int64(2),
		CommonId:      "commonId-admin-user",
		Email:         "admin-user@energyshares.jp",
		Name:          "admin-user",
		Password:      "password1234",
		Salt:          "salt",
		MobilePhone:   "01011112222",
		LastLoginDate: jsonutils.JSONTime(time.Now()),
		LoginFailCnt:  0,
		Roles:         []string{"ROLE_INVESTOR", "ROLE_ADMIN"},
		IsSuperAdmin:  false,
		ValidityStartDate: func() jsonutils.JSONTime {
			validityStartDate, _ := time.ParseInLocation(time.DateTime, "2024-02-09 00:00:00", time.Local)
			return jsonutils.JSONTime(validityStartDate)
		}(),
		ValidityEndDate: func() jsonutils.JSONTime {
			validityStartDate, _ := time.ParseInLocation(time.DateTime, "2025-12-19 23:59:59", time.Local)
			return jsonutils.JSONTime(validityStartDate)
		}(),
	}
)
