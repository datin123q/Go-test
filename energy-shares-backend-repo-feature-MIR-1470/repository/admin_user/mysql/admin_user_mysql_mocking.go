package mysql

import (
	"context"
	"time"

	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

var (
	mockAdminUser1 = domain.AdminUser{Id: 1000, CommonId: "d0b7dad8-9a33-4a23-b45d-70825a6a3f31", Email: "kildong@email.com", Name: "홍길동", MobilePhone: "01012345678", Roles: []string{string(domain.ROLE_ADMIN), string(domain.ROLE_SUPER_ADMIN)}, IsSuperAdmin: true}
	mockAdminUser2 = domain.AdminUser{Id: 1001, CommonId: "531e32a4-0775-4fe3-9bab-5d6d78d835d9", Email: "dool2@email.com", Name: "둘리", MobilePhone: "01087654321", Roles: []string{string(domain.ROLE_BASE)}, RolesString: "ROLE_BASE", IsSuperAdmin: false}

	mockCreateAdminUser3 = domain.AdminUser{Id: 1002, CommonId: "d0b7dad8-9a33-4a23-b45d-70825a6a3f31", Email: "kii@test-email.com", Name: "홍길동", Password: "password1234", Roles: []string{string(domain.ROLE_ADMIN), string(domain.ROLE_SUPER_ADMIN)}, MobilePhone: "01012345678", IsSuperAdmin: false,
		ValidityStartDate: jsonutils.JSONTime(func() time.Time {
			t, _ := time.ParseInLocation("2006-01-02 15:04:05", "2024-01-01 00:00:00", time.Local)
			return t
		}()),
		ValidityEndDate: jsonutils.JSONTime(func() time.Time {
			t, _ := time.ParseInLocation("2006-01-02 15:04:05", "2024-12-31 23:59:59", time.Local)
			return t
		}()),
	}

	mockUserRole1 = domain.Role{Id: 3000, UserId: 1000, RoleCode: string(domain.ROLE_SUPER_ADMIN)}
	mockUserRole2 = domain.Role{Id: 3001, UserId: 1000, RoleCode: string(domain.ROLE_ADMIN)}
	mockUserRole3 = domain.Role{Id: 3002, UserId: 1001, RoleCode: string(domain.ROLE_BASE)}
)

// db factory mock
type dbFactoryMock struct{}

func (d dbFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	xormEngine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}
	xormEngine.Sync(new(domain.AdminUser), new(domain.Role))
	xormEngine.ShowSQL(true)

	// テストデータを挿入
	xormEngine.Insert(mockAdminUser1, mockAdminUser2)
	xormEngine.Insert(mockUserRole1, mockUserRole2, mockUserRole3)

	return xormEngine, nil
}

func (d dbFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	panic("implement me")
}
