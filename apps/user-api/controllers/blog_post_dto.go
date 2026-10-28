package controllers

import (
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

// Deprecated 사용 만료 예정
type BlogPostResponse struct {
	Id       string `json:"id"`
	Title    string `json:"title"`
	Link     string `json:"link"`
	Data     string `json:"data"`
	Category string `json:"category"`
	Img      string `json:"img"`
	Date     string `json:"date"`
}

func (BlogPostResponse) New(blogPost *domain.BlogPost) BlogPostResponse {
	date := blogPost.WriteAt.Time().Format(time.DateOnly)
	return BlogPostResponse{
		Id:       blogPost.Id,
		Title:    blogPost.Title,
		Link:     blogPost.Link,
		Data:     blogPost.Data,
		Category: blogPost.Category,
		Img:      blogPost.SourceUrl,
		Date:     date,
	}
}

func (b BlogPostResponse) NewList(blogList []domain.BlogPost) []BlogPostResponse {
	var list []BlogPostResponse
	for _, v := range blogList {
		list = append(list, b.New(&v))
	}

	return list
}
