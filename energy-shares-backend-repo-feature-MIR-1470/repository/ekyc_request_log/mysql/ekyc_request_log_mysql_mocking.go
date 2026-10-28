package mysql

import (
	"context"
	"time"

	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

// db factory mock
type dbFactoryMock struct{}

var (
	mockEkycRequestLogToInsert = domain.EkycRequestLog{
		Email:        "gildong@gmail.com",
		UserId:       1,
		EkycId:       "EKYC-1000",
		ApplicantId:  1000,
		ErrorMessage: nil,
	}

	mockEkycRequestLog1 = domain.EkycRequestLog{
		Id:           1000,
		Email:        "gildong@gmail.com",
		UserId:       1,
		EkycId:       "EKYC-1110",
		ApplicantId:  1000,
		Status:       "NONE",
		Result:       "UNMATCH",
		ErrorMessage: nil,
	}

	mockEkycRequestLog2 = domain.EkycRequestLog{
		Id:           1001,
		Email:        "gildong@gmail.com",
		UserId:       1,
		EkycId:       "EKYC-1111",
		ApplicantId:  1000,
		Status:       "NONE",
		Result:       "UNMATCH",
		ErrorMessage: nil,
	}
)

func (d dbFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	xormEngine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	xormEngine.Sync(new(domain.EkycRequestLog))
	xormEngine.ShowSQL(false)

	// テストデータを挿入
	xormEngine.Insert(mockEkycRequestLog1)
	time.Sleep(1000 * time.Millisecond)
	xormEngine.Insert(mockEkycRequestLog2)

	return xormEngine, nil
}

func (d dbFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	panic("implement me")
}
