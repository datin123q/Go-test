package mysql

import (
	"context"
	"time"

	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

// db factory mock
type dbFactoryMock struct{}

var (
	expireIn = time.Now().Add(time.Hour * 24)
	mockCtx  = context.Background()

	mockAdminUser1 = domain.AdminUser{
		Id:    1,
		Name:  "My Admin1",
		Email: "admin@energyshares.jp",
	}
	mockRefreshToken1 = domain.RefreshToken{
		Id:       1,
		UserId:   mockAdminUser1.Id,
		Token:    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJlbmVyZ3lzaGFyZXMiLCJhdWRpdF9zdGF0dXMiOiJJTlZBTElESVRZX0FVRElUIiwiZW1haWwiOiJhZG1pbkBlbmVyZ3lzaGFyZXMuanAiLCJleHAiOjE3MDM2NDAxNTQsImlzcyI6ImVuZXJneXNoYXJlcyIsIm5hbWUiOiJBRE1JTiIsIm5iZiI6MTcwMzYzOTg1NCwicm9sZSI6WyJST0xFX0lOVkVTVE9SIiwiUk9MRV9NRVNTRU5HRVIiLCJST0xFX1BBWU1FTlQiLCJST0xFX1BST0RVQ1QiLCJST0xFX09QRVJBVE9SIiwiUk9MRV9CQVNFIiwiUk9MRV9SRVBPUlQiLCJST0xFX0FETUlOIiwiUk9MRV9TVVBFUl9BRE1JTiJdLCJ1c2VyIjoiMSIsInVzZXJfdHlwZSI6IkFETUlOIn0.w8b3v8-bBtr5YmDBCGVY5FOV9TA_dAtdkTBDx_7knMU",
		IsActive: true,
		ExpireIn: expireIn.Unix(),
		IpAddr:   "192.168.0.2",
		Agent:    "Mozilla/5.0",
		Name:     mockAdminUser1.Name,
		Email:    mockAdminUser1.Email,
		Created:  jsonutils.JSONTime(time.Now()),
		Updated:  jsonutils.JSONTime(time.Now()),
		Deleted:  jsonutils.JSONTime(time.Now()),
	}

	mockAdminUser2 = domain.AdminUser{
		Id:    2,
		Name:  "My Admin2",
		Email: "admin2@example.com",
	}
	mockRefreshToken2 = domain.RefreshToken{
		Id:       2,
		UserId:   mockAdminUser2.Id,
		Token:    "token2",
		IsActive: true,
		ExpireIn: expireIn.Unix(),
		Name:     mockAdminUser2.Name,
		Email:    mockAdminUser2.Email,
		Created:  jsonutils.JSONTime(time.Now()),
		Updated:  jsonutils.JSONTime(time.Now()),
		Deleted:  jsonutils.JSONTime(time.Now()),
	}

	mockUser1 = domain.User{
		Id:    1000,
		Name:  "My User1",
		Email: "user1@example.com",
	}
	mockRefreshToken3 = domain.RefreshToken{
		Id:       3,
		UserId:   mockUser1.Id,
		Token:    "token3",
		IsActive: true,
		ExpireIn: expireIn.Unix(),
		Name:     mockUser1.Name,
		Email:    mockUser1.Email,
		Created:  jsonutils.JSONTime(time.Now()),
		Updated:  jsonutils.JSONTime(time.Now()),
		Deleted:  jsonutils.JSONTime(time.Now()),
	}

	mockUser2 = domain.User{
		Id:    1001,
		Name:  "My User2",
		Email: "user2@example.com",
	}
	mockRefreshToken4 = domain.RefreshToken{
		Id:       4,
		UserId:   mockUser2.Id,
		Token:    "token4",
		IsActive: true,
		ExpireIn: expireIn.Unix(),
		Name:     mockUser2.Name,
		Email:    mockUser2.Email,
		Created:  jsonutils.JSONTime(time.Now()),
		Updated:  jsonutils.JSONTime(time.Now()),
		Deleted:  jsonutils.JSONTime(time.Now()),
	}
)

func (d dbFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	xormEngine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	xormEngine.Sync(new(domain.RefreshToken))
	xormEngine.ShowSQL(false)

	// テストデータ挿入
	xormEngine.Insert(mockRefreshToken1, mockRefreshToken2, mockRefreshToken3, mockRefreshToken4)

	return xormEngine, nil
}

func (d dbFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	panic("implement me")
}
