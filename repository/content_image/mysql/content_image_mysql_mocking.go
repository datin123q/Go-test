package mysql

import (
	"context"
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"xorm.io/xorm"
)

// db factory mock
type dbFactoryMock struct{}

var (
	mockContentImage1 = domain.ContentImage{
		Id:            1000,
		FileName:      "bpbmq9297s524a3jhgn1",
		OriginalName:  "bpbmhsbipt35d0tqvg3f.pdf",
		FileSize:      1398,
		FileExtension: "pdf",
		FilePath:      "/content_image/2702602/",
		FileUrl:       "https://dev-energyshares-file-storage.s3.ap-northeast-2.amazonaws.com/content_image/2702602/bpbmq9297s524a3jhgn1.pdf",
		ResourceId:    "e3ae2c14-b617-4870-b057-a8ae915cdee7",
		Created:       jsonutils.JSONTime(time.Date(2020, 02, 27, 16, 17, 57, 0, time.UTC)),
		Updated:       jsonutils.JSONTime(time.Date(2020, 02, 27, 16, 17, 57, 0, time.UTC)),
	}
	mockContentImage2 = domain.ContentImage{
		Id:            1001,
		FileName:      "bpbmq9297s524a3jhgn0",
		OriginalName:  "bpbmhsbipt35d0tqvg3g.pdf",
		FileSize:      1398,
		FileExtension: "pdf",
		FilePath:      "/content_image/2702602/",
		FileUrl:       "https://dev-energyshares-file-storage.s3.ap-northeast-2.amazonaws.com/content_image/2702602/bpbmq9297s524a3jhgn0.pdf",
		ResourceId:    "e3ae2c14-b617-4870-b057-a8ae915cdee6",
		Created:       jsonutils.JSONTime(time.Date(2020, 02, 27, 16, 17, 57, 0, time.UTC)),
		Updated:       jsonutils.JSONTime(time.Date(2020, 02, 27, 16, 17, 57, 0, time.UTC)),
	}

	mockContentImage3 = domain.ContentImage{
		Id:            1002,
		FileName:      "bpbmq9a97s524a3jhgng",
		OriginalName:  "bpbmknript35e7ukrpmg.png",
		FileSize:      102160,
		FileExtension: ".png",
		FilePath:      "/content_image/2702602/",
		FileUrl:       "https://dev-energyshares-file-storage.s3.ap-northeast-2.amazonaws.com/content_image/2702602/bpbmq9a97s524a3jhgng.png",
		ResourceId:    "e3ae2c14-b617-4870-b057-a8ae915cdee6",
		Created:       jsonutils.JSONTime(time.Date(2020, 02, 27, 16, 17, 57, 0, time.UTC)),
		Updated:       jsonutils.JSONTime(time.Date(2020, 02, 27, 16, 17, 57, 0, time.UTC)),
	}
)

func (d dbFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	xormEngine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	xormEngine.Sync(new(domain.ContentImage))
	xormEngine.ShowSQL(false)

	// テストデータを挿入
	xormEngine.Insert(mockContentImage1, mockContentImage2, mockContentImage3)

	return xormEngine, nil
}

func (d dbFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	panic("implement me")
}
