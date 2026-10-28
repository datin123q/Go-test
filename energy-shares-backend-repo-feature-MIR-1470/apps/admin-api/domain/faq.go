package domain

import (
	"context"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
)

type FaqUsecase interface {
	Create(ctx context.Context, request FaqCreateRequest) error
	Update(ctx context.Context, request FaqUpdateRequest) error
	GetAll(ctx context.Context, request FaqGetAllRequest) (FaqGetAllResponse, error)
	GetOne(ctx context.Context, request FaqGetOneRequest) (FaqGetOneResponse, error)
	Delete(ctx context.Context, request FaqDeleteRequest) error
}

type FaqCreateRequest struct {
	Pin              *bool               `json:"pin" description:"고정핀" default:"false" example:"false"`
	ContentHtml      string              `json:"content_html" description:"게시판 내용" example:"<body><h1>테스트입니다.</h1></body>"`
	Title            string              `json:"title" description:"제목" required:"true" validate:"min=1,max=200" minLength:"1" maxLength:"200" example:"테스트 입니다."`
	Posted           bool                `json:"posted" description:"게시 여부" default:"false" example:"true"`
	ClosingDate      *jsonutils.JSONTime `json:"closing_date" description:"마감일" format:"yyyy-mm-dd HH:mm:ss" example:"2024-10-10 00:00:00"`
	CategoryType     string              `json:"category_type" description:"카테고리" enum:"FAQ_CATEGORY_ACCOUNT_REGISTRATION,FAQ_CATEGORY_PUT_TAKE_MONEY,FAQ_CATEGORY_INVESTING,FAQ_CATEGORY_FUNDS,FAQ_CATEGORY_REPAY_DISTRIBUTION" validate:"oneof=FAQ_CATEGORY_ACCOUNT_REGISTRATION FAQ_CATEGORY_PUT_TAKE_MONEY FAQ_CATEGORY_INVESTING FAQ_CATEGORY_FUNDS FAQ_CATEGORY_REPAY_DISTRIBUTION"`
	ContentImageUrls []string            `json:"content_image_urls" description:"게시판 이미지 URL 리스트" example:"[https://es-jp-devv2-public-file-storage.s3.ap-northeast-2.amazonaws.com/content_img_file/2024-07-26/cqhkh82nulri1.jpg]"`
	AttachmentUrls   []string            `json:"attachment_urls" description:"첨부파일 URL 리스트" example:"[https://es-jp-devv2-public-file-storage.s3.ap-northeast-2.amazonaws.com/attachment_file/2024-07-26/cqhkh82nulri2.jpg]"`
	JWTUser          jwthelper.User
	ClientIp         string
}

type FaqUpdateRequest struct {
	Id               int64               `param:"id" path:"id" description:"단건 업데이트를 위한 FAQ ID" example:"1" format:"int64"`
	Pin              *bool               `json:"pin" description:"고정핀" default:"false" example:"false"`
	ContentHtml      string              `json:"content_html" description:"게시판 내용" example:"<body><h1>테스트입니다.</h1></body>"`
	Title            string              `json:"title" description:"제목" required:"true" validate:"min=1,max=200" minLength:"1" maxLength:"200" example:"테스트 입니다."`
	Posted           bool                `json:"posted" description:"게시 여부" default:"false" example:"true"`
	ClosingDate      *jsonutils.JSONTime `json:"closing_date" description:"마감일" format:"yyyy-mm-dd HH:mm:ss" example:"2024-10-10 00:00:00"`
	CategoryType     string              `json:"category_type" description:"카테고리" enum:"FAQ_CATEGORY_ACCOUNT_REGISTRATION,FAQ_CATEGORY_PUT_TAKE_MONEY,FAQ_CATEGORY_INVESTING,FAQ_CATEGORY_FUNDS,FAQ_CATEGORY_REPAY_DISTRIBUTION" validate:"oneof=FAQ_CATEGORY_ACCOUNT_REGISTRATION FAQ_CATEGORY_PUT_TAKE_MONEY FAQ_CATEGORY_INVESTING FAQ_CATEGORY_FUNDS FAQ_CATEGORY_REPAY_DISTRIBUTION"`
	ContentImageUrls []string            `json:"content_image_urls" description:"게시판 이미지 URL 리스트" example:"[https://es-jp-devv2-public-file-storage.s3.ap-northeast-2.amazonaws.com/content_img_file/2024-07-26/cqhkh82nulri1.jpg]"`
	AttachmentUrls   []string            `json:"attachment_urls" description:"첨부파일 URL 리스트" example:"[https://es-jp-devv2-public-file-storage.s3.ap-northeast-2.amazonaws.com/attachment_file/2024-07-26/cqhkh82nulri2.jpg]"`
	JWTUser          jwthelper.User
	ClientIp         string
}

type FaqGetAllRequest struct {
	JWTUser  jwthelper.User
	ClientIp string
	FaqGetAllQueryParam
}

type FaqGetAllQueryParam struct {
	Title        *string `query:"title" description:"제목"`
	StartDate    *string `query:"start_date" description:"등록일자 시작일" format:"yyyy-mm-dd HH:mm:ss"`
	EndDate      *string `query:"end_date" description:"등록일자 종료일" format:"yyyy-mm-dd HH:mm:ss"`
	Posted       *bool   `query:"posted" description:"게시여부"`
	CategoryType *string `query:"category_type" description:"카테고리\n* FAQ_CATEGORY_ACCOUNT_REGISTRATION(회원등록)\n* FAQ_CATEGORY_PUT_TAKE_MONEY(입금·출금)\n* FAQ_CATEGORY_INVESTING(투자)\n* FAQ_CATEGORY_FUNDS(펀드)\n* FAQ_CATEGORY_REPAY_DISTRIBUTION(상환·분배)" enum:"FAQ_CATEGORY_ACCOUNT_REGISTRATION,FAQ_CATEGORY_PUT_TAKE_MONEY,FAQ_CATEGORY_INVESTING,FAQ_CATEGORY_FUNDS,FAQ_CATEGORY_REPAY_DISTRIBUTION"`
	cdomain.SearchInput
}

type FaqGetAllResponse struct {
	TotalCount int64        `json:"total_count" description:"총 개수" format:"int64" example:"1"`
	Items      []domain.Faq `json:"items" description:"FAQ 리스트"`
}

type FaqGetOneRequest struct {
	JWTUser  jwthelper.User
	ClientIp string
	Id       int64 `param:"id" path:"id" description:"단건조회를 위한 FAQ ID" example:"1" format:"int64"`
}

type FaqGetOneResponse struct {
	domain.Faq
}

type FaqDeleteRequest struct {
	JWTUser  jwthelper.User
	ClientIp string
	Id       int64 `param:"id" path:"id" description:"삭제를 위한 FAQ ID" example:"1" format:"int64"`
}
