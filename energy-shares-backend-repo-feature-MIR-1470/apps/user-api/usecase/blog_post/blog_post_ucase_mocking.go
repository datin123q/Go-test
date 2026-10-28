package blog_post

import (
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	udomain "github.com/illuminarean/energyshares-user-api/domain"
)

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
	PublishedAt: jsonutils.JSONTime(func() time.Time {
		t, _ := time.ParseInLocation(time.RFC3339, "2021-01-02T00:00:00+09:00", time.Local)
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

var mockBlogPostResponse1 = udomain.BlogPostResponse{
	Id:       mockBlogPost1.Id,
	Title:    mockBlogPost1.Title,
	Link:     mockBlogPost1.Link,
	Data:     mockBlogPost1.Data,
	Category: mockBlogPost1.Category,
	Img:      mockBlogPost1.SourceUrl,
	Date:     mockBlogPost1.PublishedAt.Time().Format(time.DateOnly),
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
	PublishedAt: jsonutils.JSONTime(func() time.Time {
		t, _ := time.ParseInLocation(time.RFC3339, "2021-01-02T00:00:00+09:00", time.Local)
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

var mockBlogPostResponse2 = udomain.BlogPostResponse{
	Id:       mockBlogPost2.Id,
	Title:    mockBlogPost2.Title,
	Link:     mockBlogPost2.Link,
	Data:     mockBlogPost2.Data,
	Category: mockBlogPost2.Category,
	Img:      mockBlogPost2.SourceUrl,
	Date:     mockBlogPost2.PublishedAt.Time().Format(time.DateOnly),
}
