package domain

import (
	"context"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
)

type NewsUsecase interface {
	GetAll(ctx context.Context, request NewsGetAllRequest) (NewsGetAllResponse, error)
	GetOne(ctx context.Context, request NewsGetOneRequest) (NewsGetOneResponse, error)
	Create(ctx context.Context, request NewsCreateRequest) error
	Update(ctx context.Context, request NewsUpdateRequest) error
	Delete(ctx context.Context, request NewsDeleteRequest) error
}

type NewsGetAllRequest struct {
	JWTUser  jwthelper.User
	ClientIp string
	NewsGetAllQueryParam
}

type NewsGetAllQueryParam struct {
	Title          string `query:"title" description:"제목"`
	Posted         *bool  `query:"posted"  description:"게시여부"`
	OnlyNotExpired *bool  `query:"only_not_expired" description:"게시일 만료전 공지사항만 노출"`
	StartDate      string `query:"start_date" description:"등록일자 시작일" format:"yyyy-mm-dd HH:mm:ss"`
	EndDate        string `query:"end_date" description:"등록일자 종료일" format:"yyyy-mm-dd HH:mm:ss"`
	cdomain.SearchInput
}

type NewsGetAllResponse struct {
	Items      []domain.News `json:"items" description:"공지사항 리스트"`
	TotalCount int64         `json:"total_count" description:"총 개수" example:"1"`
}

type NewsCreateRequest struct {
	Pin              *bool               `json:"pin" description:"고정핀" default:"false" example:"false"`
	ContentHtml      string              `json:"content_html" description:"공지 내용" example:"<body><h1>테스트입니다.</h1></body>"`
	Title            string              `json:"title" description:"제목" required:"true" validate:"min=1,max=200" minLength:"1" maxLength:"200" example:"테스트 입니다."`
	Posted           bool                `json:"posted" description:"게시 여부" default:"true" example:"true"`
	ClosingDate      *jsonutils.JSONTime `json:"closing_date" description:"마감일" format:"yyyy-mm-dd HH:mm:ss" example:"2024-10-10 00:00:00"`
	SendMailStatus   bool                `json:"send_mail_status" description:"메일 전송 여부" default:"false" example:"true"`
	ContentImageUrls []string            `json:"content_image_urls" description:"게시판 이미지 URL 리스트" example:"[https://es-jp-devv2-public-file-storage.s3.ap-northeast-2.amazonaws.com/content_img_file/2024-07-26/cqhkh82nulri1.jpg]"`
	AttachmentUrls   []string            `json:"attachment_urls" description:"첨부파일 URL 리스트" example:"[https://es-jp-devv2-public-file-storage.s3.ap-northeast-2.amazonaws.com/attachment_file/2024-07-26/cqhkh82nulri2.jpg]"`
	JWTUser          jwthelper.User
	ClientIp         string
}

type NewsGetOneRequest struct {
	JWTUser  jwthelper.User
	ClientIp string
	NewsId   int64 `param:"id" path:"id" swagger:"desc(뉴스 아이디)"`
}

type NewsGetOneResponse struct {
	domain.News
}

type NewsUpdateRequest struct {
	Id               int64               `param:"id" path:"id" description:"단건 업데이트를 위한 공지사항 ID" example:"1" format:"int64"`
	Pin              *bool               `json:"pin" description:"고정핀" default:"false" example:"false"`
	ContentHtml      string              `json:"content_html" description:"공지 내용" example:"<body><h1>테스트입니다.</h1></body>"`
	Title            string              `json:"title" description:"제목" required:"true" validate:"min=1,max=200" minLength:"1" maxLength:"200" example:"테스트 입니다."`
	Posted           bool                `json:"posted" description:"게시 여부" default:"true" example:"true"`
	ClosingDate      *jsonutils.JSONTime `json:"closing_date" description:"마감일" format:"yyyy-mm-dd HH:mm:ss" example:"2024-10-10 00:00:00"`
	ContentImageUrls []string            `json:"content_image_urls" description:"게시판 이미지 URL 리스트" example:"[https://es-jp-devv2-public-file-storage.s3.ap-northeast-2.amazonaws.com/content_img_file/2024-07-26/cqhkh82nulri1.jpg]"`
	AttachmentUrls   []string            `json:"attachment_urls" description:"첨부파일 URL 리스트" example:"[https://es-jp-devv2-public-file-storage.s3.ap-northeast-2.amazonaws.com/attachment_file/2024-07-26/cqhkh82nulri2.jpg]"`
	JWTUser          jwthelper.User
	ClientIp         string
}

type NewsDeleteRequest struct {
	JWTUser  jwthelper.User
	ClientIp string
	NewsId   int64 `param:"id" path:"id" swagger:"desc(뉴스 아이디)"`
}
