package mysql

import (
	"context"

	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

// db factory mock
type dbFactoryMock struct{}

var (
	mockAntisocialCheckFile1 = domain.AntisocialCheckFile{
		Id:            1000,
		CheckId:       1000,
		FileName:      "test",
		OriginalName:  "test",
		FileSize:      1000,
		FileExtension: "pdf",
		FilePath:      "test",
		FileUrl:       "http://test.com/test",
		OwnerId:       1000,
		UpdateId:      1000,
	}
	mockAntisocialCheckFile2 = domain.AntisocialCheckFile{
		Id:            2000,
		CheckId:       1001,
		FileName:      "test",
		OriginalName:  "test",
		FileSize:      1000,
		FileExtension: "pdf",
		FilePath:      "test",
		FileUrl:       "http://test.com/test",
		OwnerId:       1000,
		UpdateId:      1000,
	}
)

func (d dbFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	xormEngine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	xormEngine.Sync(new(domain.AntisocialCheckFile))
	xormEngine.ShowSQL(false)

	// テストデータを挿入
	xormEngine.Insert(mockAntisocialCheckFile1, mockAntisocialCheckFile2)

	return xormEngine, nil
}

func (d dbFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	panic("implement me")
}
