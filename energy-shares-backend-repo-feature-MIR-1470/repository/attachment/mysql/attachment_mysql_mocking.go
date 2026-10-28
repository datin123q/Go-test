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
	mockAttachment1 = domain.Attachment{
		Id:            1000,
		FileName:      "bpbmq9297s524a3jhgn1",
		OriginalName:  "bpbmhsbipt35d0tqvg3f.pdf",
		FileSize:      1398,
		FileExtension: "pdf",
		FilePath:      "/attachment_file/2702602/",
		FileUrl:       "https://dev-energyshares-file-storage.s3.ap-northeast-2.amazonaws.com/attachment_file/2702602/bpbmq9297s524a3jhgn1.pdf",
		ResourceId:    "e3ae2c14-b617-4870-b057-a8ae915cdee7",
		Created:       jsonutils.JSONTime(time.Date(2020, 02, 27, 16, 17, 57, 0, time.UTC)),
		Updated:       jsonutils.JSONTime(time.Date(2020, 02, 27, 16, 17, 57, 0, time.UTC)),
	}
	mockAttachment2 = domain.Attachment{
		Id:            1001,
		FileName:      "bpbmq9297s524a3jhgn0",
		OriginalName:  "bpbmhsbipt35d0tqvg3g.pdf",
		FileSize:      1398,
		FileExtension: "pdf",
		FilePath:      "/attachment_file/2702602/",
		FileUrl:       "https://dev-energyshares-file-storage.s3.ap-northeast-2.amazonaws.com/attachment_file/2702602/bpbmq9297s524a3jhgn0.pdf",
		ResourceId:    "e3ae2c14-b617-4870-b057-a8ae915cdee6",
		Created:       jsonutils.JSONTime(time.Date(2020, 02, 27, 16, 17, 57, 0, time.UTC)),
		Updated:       jsonutils.JSONTime(time.Date(2020, 02, 27, 16, 17, 57, 0, time.UTC)),
	}

	mockAttachment3 = domain.Attachment{
		Id:            1002,
		FileName:      "bpbmq9a97s524a3jhgng",
		OriginalName:  "bpbmknript35e7ukrpmg.png",
		FileSize:      102160,
		FileExtension: ".png",
		FilePath:      "/attachment_file/2702602/",
		FileUrl:       "https://dev-energyshares-file-storage.s3.ap-northeast-2.amazonaws.com/attachment_file/2702602/bpbmq9a97s524a3jhgng.png",
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

	xormEngine.Sync(new(domain.Attachment))
	xormEngine.ShowSQL(false)

	// テストデータを挿入
	xormEngine.Insert(mockAttachment1, mockAttachment2, mockAttachment3)

	return xormEngine, nil
}

func (d dbFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	panic("implement me")
}
