package mysql

import (
	"time"

	"context"
	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

// db factory mock
type dbFactoryMock struct{}

var mockBlogPost1 = domain.BlogPost{
	Id:        "c71c759d-5ff3-47d7-bfee-a034b2028040",
	Title:     "title1",
	Data:      "data1",
	Category:  "category1",
	SourceUrl: "sourceUrl1",
	WriteAt: jsonutils.JSONTime(func() time.Time {
		t, _ := time.ParseInLocation(time.RFC3339, "2021-01-01T00:00:00+09:00", time.Local)
		return t
	}()),
	Created: jsonutils.JSONTime(func() time.Time {
		t, _ := time.ParseInLocation(time.RFC3339, "2021-01-01T00:00:00+09:00", time.Local)
		return t
	}()),
	Updated: jsonutils.JSONTime(func() time.Time {
		t, _ := time.ParseInLocation(time.RFC3339, "2021-01-01T00:00:00+09:00", time.Local)
		return t
	}()),
}
var mockBlogPost2 = domain.BlogPost{
	Id:        "n71c759d-5ff3-47d7-bfee-a034b2028040",
	Title:     "title2",
	Data:      "data2",
	Category:  "category2",
	SourceUrl: "sourceUrl2",
	WriteAt: jsonutils.JSONTime(func() time.Time {
		t, _ := time.ParseInLocation(time.RFC3339, "2021-01-01T00:00:00+09:00", time.Local)
		return t
	}()),
	Created: jsonutils.JSONTime(func() time.Time {
		t, _ := time.ParseInLocation(time.RFC3339, "2021-01-01T00:00:00+09:00", time.Local)
		return t
	}()),
	Updated: jsonutils.JSONTime(func() time.Time {
		t, _ := time.ParseInLocation(time.RFC3339, "2021-01-01T00:00:00+09:00", time.Local)
		return t
	}()),
}
var mockCreateBlogPost = domain.BlogPost{
	Id:        "f71c759d-5ff3-47d7-bfee-a034b2028040",
	Title:     "title1",
	Data:      "data1",
	Category:  "category1",
	SourceUrl: "sourceUrl1",
	WriteAt: jsonutils.JSONTime(func() time.Time {
		t, _ := time.ParseInLocation(time.RFC3339, "2021-01-01T00:00:00+09:00", time.Local)
		return t
	}()),
	Created: jsonutils.JSONTime(func() time.Time {
		t, _ := time.ParseInLocation(time.RFC3339, "2021-01-01T00:00:00+09:00", time.Local)
		return t
	}()),
	Updated: jsonutils.JSONTime(func() time.Time {
		t, _ := time.ParseInLocation(time.RFC3339, "2021-01-01T00:00:00+09:00", time.Local)
		return t
	}()),
}

func (d dbFactoryMock) EnergySharesDB(ctx context.Context) (xorm.Interface, error) {
	xormEngine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	xormEngine.Sync(new(domain.BlogPost))
	xormEngine.ShowSQL(false)

	// テストデータを挿入
	xormEngine.Insert(mockBlogPost1, mockBlogPost2)

	return xormEngine, nil
}

func (d dbFactoryMock) MyNumberDB(ctx context.Context) (xorm.Interface, error) {
	panic("implement me")
}
