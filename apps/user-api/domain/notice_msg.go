package domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
)

type NoticeMsgUsecase interface {
	GetAll(ctx context.Context, request NoticeMsgGetAllRequest) (NoticeMsgGetAllResponse, error)
	GetOne(ctx context.Context, request NoticeMsgGetOneRequest) (NoticeMsgGetOneResponse, error)
}

type NoticeMsgGetAllRequest struct {
	JWTUser  jwthelper.User
	ClientIp string
	NoticeMsgGetAllQueryParam
}

type NoticeMsgGetAllQueryParam struct {
	Title     *string `query:"title" description:"タイトル" example:"title"`
	StartDate *string `query:"start_date" description:"開始日" format:"yyyy-mm-dd HH:mm:ss" validate:"omitempty,datetime=2006-01-02 15:04:05" example:"2021-01-01 00:00:00"`
	EndDate   *string `query:"end_date" description:"終了日" format:"yyyy-mm-dd HH:mm:ss" validate:"omitempty,datetime=2006-01-02 15:04:05" example:"2021-01-01 23:59:59"`
	domain.SearchInput
}

type NoticeMsgGetAllResponse struct {
	TotalCount int64              `json:"total_count" description:"合計数" example:"1"`
	Items      []domain.NoticeMsg `json:"items"`
}

type NoticeMsgGetOneRequest struct {
	Id       int64 `param:"id" path:"id" description:"照会するユーザー通知メッセージID" format:"int64" example:"1"`
	JWTUser  jwthelper.User
	ClientIp string
}

type NoticeMsgGetOneResponse struct {
	domain.NoticeMsg
}
