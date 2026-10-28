package domain

import (
	"context"
)

type BlogPostUsecase interface {
	GetAll(ctx context.Context, request BlogPostGetAllRequest) (int64, BlogPostGetAllResponse, error)
}

type BlogPostGetAllRequest struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

type BlogPostResponse struct {
	Id       string `json:"id"`
	Title    string `json:"title"`
	Link     string `json:"link"`
	Data     string `json:"data"`
	Category string `json:"category"`
	Img      string `json:"img"`
	Date     string `json:"date"`
}

type BlogPostGetAllResponse struct {
	TotalCount int64              `json:"total_count"`
	Items      []BlogPostResponse `json:"items"`
}
