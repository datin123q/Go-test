package domain

import (
	"context"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
)

type NewsUsecase interface {
	GetAll(ctx context.Context, request NewsGetAllRequest) (NewsGetAllResponse, error)
}

type NewsGetAllRequest struct {
	JWTUser  jwthelper.User
	ClientIp string
	NewsGetAllQueryParam
}

type NewsGetAllQueryParam struct {
	Title          string `query:"title" description:"タイトル"`
	Posted         *bool  `query:"posted"  description:"掲載有無"`
	OnlyNotExpired *bool  `query:"only_not_expired" description:"掲載日期限前のお知らせのみ表示"`
	StartDate      string `query:"start_date" description:"登録日の開始日"`
	EndDate        string `query:"end_date" description:"登録日の終了日"`
	cdomain.SearchInput
}

type NewsGetAllResponse struct {
	Items      []domain.News `json:"items" description:"お知らせリスト"`
	TotalCount int64         `json:"total_count" description:"総件数" example:"1"`
}
