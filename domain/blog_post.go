package domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

// Blog Post: 블로그 포스트 모델 (WordPress & Ghost 지원ㄴ)
type BlogPost struct {
	Id          string             `json:"id" xorm:"id pk uuid"`
	Title       string             `json:"title"`                            // 제목
	Link        string             `json:"link"`                             // 블로그 링크
	Data        string             `json:"data" xorm:"text"`                 // 내용
	Category    string             `json:"category"`                         // 카테고리
	SourceUrl   string             `json:"source_url" xorm:"source_url"`     // 이미지 소스 URL
	WriteAt     jsonutils.JSONTime `json:"write_at" xorm:"write_at"`         // 작성일
	PublishedAt jsonutils.JSONTime `json:"published_at" xorm:"published_at"` // 발행일
	Created     jsonutils.JSONTime `json:"created" xorm:"created"`
	Updated     jsonutils.JSONTime `json:"updated" xorm:"updated"`
}

func (p *BlogPost) Create(ctx context.Context) (int64, error) {
	return factory.DB(ctx).Insert(p)
}

func (p *BlogPost) Update(ctx context.Context) (err error) {
	_, err = factory.DB(ctx).ID(p.Id).Update(p)
	return
}

func (BlogPost) TableName() string {
	return "blog_post"
}

func (p *BlogPost) GetById(ctx context.Context) (bool, error) {
	return factory.DB(ctx).ID(p.Id).Get(p)
}

func (BlogPost) IsExist(ctx context.Context, id string) (bool, error) {
	return factory.DB(ctx).ID(id).Exist(&BlogPost{})
}

func (p *BlogPost) Delete(ctx context.Context) (err error) {
	_, err = factory.DB(ctx).ID(p.Id).Delete(p)
	return
}

func (BlogPost) GetAll(ctx context.Context, offset, limit int) (int64, []BlogPost, error) {
	var items []BlogPost
	total, err := factory.DB(ctx).Table(BlogPost{}).OrderBy("published_at DESC").Limit(limit, offset).FindAndCount(&items)
	return total, items, err
}
