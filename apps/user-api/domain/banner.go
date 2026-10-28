package domain

import (
	"context"
)

type BannerUsecase interface {
	GetAll(context.Context) (BannerGetAllResponse, error)
}

type BannerGetAllResponse struct {
	TotalCount    int64              `json:"total_count" swagger:"desc(掲載中の件数)"`
	RollingSecond int64              `json:"rolling_second" swagger:"desc(ローリング時間（単位：秒）)"`
	Items         []BannerGetAllItem `json:"items" swagger:"desc(バナーリスト)"`
}

type BannerGetAllItem struct {
	Id        int64  `json:"id" swagger:"desc(バナーID)"`
	Title     string `json:"title" swagger:"desc(タイトル)"`
	IsPosted  bool   `json:"posted" swagger:"desc(掲載有無)"`
	SortOrder int64  `json:"sort_order" swagger:"desc(順序)"`
	Url       string `json:"url" swagger:"desc(url)"`
	NewTab    bool   `json:"new_tab" swagger:"desc(新しいタブで開く)"`
}
