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
	createdTime     = time.Date(2024, 4, 15, 0, 0, 0, 0, time.Local)
	urlExpireTime   = createdTime.Add(time.Hour * 24) // 24時間後に失効
	emailExpireTime = createdTime.Add(time.Hour * 24) // 24時間後に失効

	mockSignUpRequest1 = domain.SignUpRequest{
		Id:        1,
		RequestId: "bcdb9d00-fb92-11ee-a951-0242ac120002",
		Email:     "gildong@email.com",
		Url:       "test-url",
		SendCount: 1,
		UrlExpire: jsonutils.JSONTime(urlExpireTime),
		Created:   jsonutils.JSONTime(createdTime),
	}

	mockSignUpRequest2 = domain.SignUpRequest{
		Id:          2,
		RequestId:   "bcdb9d00-fb92-11ee-a951-0242ac120003",
		Email:       "dool2@email.com",
		Url:         "test-url",
		EmailExpire: jsonutils.JSONTime(emailExpireTime),
		SendCount:   1,
	}
)

func (d dbFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	xormEngine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	xormEngine.Sync(new(domain.SignUpRequest))
	xormEngine.ShowSQL(false)

	// テストデータ挿入
	xormEngine.Insert(mockSignUpRequest1, mockSignUpRequest2)

	return xormEngine, nil
}

func (d dbFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	panic("implement me")
}
