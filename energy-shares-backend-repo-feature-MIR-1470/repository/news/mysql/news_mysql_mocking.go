package mysql

import (
	"context"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"

	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

var (
	mockNews1 = domain.News{
		Id:            1000,
		CommonId:      "2000",
		Pin:           true,
		Title:         "Title1",
		OwnerId:       1000,
		UpdateId:      1000,
		ContentHtml:   "",
		Posted:        true,
		ClosingDate:   jsonutils.JSONTime{},
		Created:       jsonutils.JSONTime{},
		Updated:       jsonutils.JSONTime{},
		Deleted:       jsonutils.JSONTime{},
		Attachments:   nil,
		AttachmentCnt: 0,
	}
	mockNews2 = domain.News{
		Id:            1001,
		CommonId:      "2001",
		Pin:           false,
		Title:         "Title2",
		OwnerId:       1000,
		UpdateId:      1000,
		ContentHtml:   "",
		Posted:        false,
		ClosingDate:   jsonutils.JSONTime{},
		Created:       jsonutils.JSONTime{},
		Updated:       jsonutils.JSONTime{},
		Deleted:       jsonutils.JSONTime{},
		Attachments:   nil,
		AttachmentCnt: 0,
	}
	mockNewsToCreate = domain.News{
		Id:            1002,
		CommonId:      "2003",
		Pin:           false,
		Title:         "Title Created",
		OwnerId:       1000,
		UpdateId:      1000,
		ContentHtml:   "",
		Posted:        true,
		ClosingDate:   jsonutils.JSONTime{},
		Created:       jsonutils.JSONTime{},
		Updated:       jsonutils.JSONTime{},
		Deleted:       jsonutils.JSONTime{},
		Attachments:   nil,
		AttachmentCnt: 0,
	}
)

// db factory mock
type dbFactoryMock struct{}

func (d dbFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	xormEngine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	xormEngine.Sync(new(domain.News))
	xormEngine.ShowSQL(false)

	// テストデータ挿入
	xormEngine.Insert(mockNews1, mockNews2)

	return xormEngine, nil
}

// インターフェース満足のために必要
func (d dbFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	panic("implement me")
}
