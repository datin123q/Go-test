package mysql

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

var (
	mockFaq1 = domain.Faq{
		Id:            1000,
		CommonId:      "2000",
		Pin:           false,
		Title:         "중요 공지사항입니다.",
		OwnerId:       1000,
		UpdateId:      1000,
		ContentHtml:   "",
		Posted:        true,
		CategoryType:  domain.FAQ_CATEGORY_INVESTING,
		ClosingDate:   jsonutils.JSONTime{},
		Created:       jsonutils.JSONTime{},
		Updated:       jsonutils.JSONTime{},
		Deleted:       jsonutils.JSONTime{},
		Attachments:   nil,
		AttachmentCnt: 0,
	}
	mockFaq2 = domain.Faq{
		Id:            1001,
		CommonId:      "2001",
		Pin:           false,
		Title:         "서브 공지사항입니다.",
		OwnerId:       1000,
		UpdateId:      1000,
		ContentHtml:   "",
		Posted:        false,
		CategoryType:  domain.FAQ_CATEGORY_ACCOUNT_REGISTRATION,
		ClosingDate:   jsonutils.JSONTime{},
		Created:       jsonutils.JSONTime{},
		Updated:       jsonutils.JSONTime{},
		Deleted:       jsonutils.JSONTime{},
		Attachments:   nil,
		AttachmentCnt: 0,
	}

	mockFaq3 = domain.Faq{
		Id:            1002,
		CommonId:      "2001",
		Pin:           true,
		Title:         "핀이 박힌 공지사항입니다.",
		OwnerId:       1000,
		UpdateId:      1000,
		ContentHtml:   "",
		CategoryType:  domain.FAQ_CATEGORY_FUNDS,
		Posted:        false,
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

	xormEngine.Sync(new(domain.Faq))
	xormEngine.ShowSQL(false)

	// テストデータを挿入
	xormEngine.Insert(mockFaq1, mockFaq2, mockFaq3)
	return xormEngine, nil
}

// インターフェース仕様を満たすために必要
func (d dbFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	panic("implement me")
}
