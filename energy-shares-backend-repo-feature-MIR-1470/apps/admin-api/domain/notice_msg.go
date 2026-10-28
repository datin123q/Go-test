package domain

import (
	"context"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
)

type NoticeMsgUsecase interface {
	GetAll(ctx context.Context, request NoticeMsgGetAllRequest) (NoticeMsgGetAllResponse, error)
	GetOne(ctx context.Context, request NoticeMsgGetOneRequest) (NoticeMsgGetOneResponse, error)
	Create(ctx context.Context, request NoticeMsgCreateRequest) error
	CreateByProject(ctx context.Context, request NoticeMsgCreateByProjectRequest) error
	CreateByNewsletter(ctx context.Context, request NoticeMsgCreateByNewsletterRequest) error
	CreateByUserType(ctx context.Context, request NoticeMsgCreateByUserTypeRequest) error
}

type NoticeMsgGetAllRequest struct {
	JWTUser  jwthelper.User
	ClientIp string
	NoticeMsgGetAllQueryParam
}

type NoticeMsgGetAllQueryParam struct {
	Title     *string `query:"title" description:"제목" format:"string" example:"title"`
	StartDate *string `query:"start_date" description:"시작일" format:"yyyy-mm-dd HH:mm:ss" validate:"omitempty,datetime=2006-01-02 15:04:05" example:"2021-01-01 00:00:00"`
	EndDate   *string `query:"end_date" description:"종료일" format:"yyyy-mm-dd HH:mm:ss" validate:"omitempty,datetime=2006-01-02 15:04:05" example:"2021-01-01 23:59:59"`
	cdomain.SearchInput
}

type NoticeMsgGetAllResponse struct {
	TotalCount int64              `json:"total_count" description:"총 개수" format:"int64" example:"1"`
	Items      []domain.NoticeMsg `json:"items" description:"사용자 소식 리스트"`
}

type NoticeMsgGetOneRequest struct {
	Id       int64 `param:"id" path:"id" description:"조회할 사용자 소식 ID" format:"int64" example:"1"`
	JWTUser  jwthelper.User
	ClientIp string
}

type NoticeMsgGetOneResponse struct {
	domain.NoticeMsg
}

type NoticeMsgCreateRequestBase struct {
	Title       string `json:"title" description:"제목" required:"true" validate:"required" minLength:"1" maxLength:"200" example:"타이틀"`
	ContentHtml string `json:"content_html" description:"소식 내용" required:"true" validate:"required" example:"<body><h1>테스트입니다.</h1></body>"`
}

type NoticeMsgCreateRequest struct {
	UserIds  []int64 `json:"user_ids" description:"사용자 ID 리스트" example:"[1,2,3]"`
	JWTUser  jwthelper.User
	ClientIp string
	NoticeMsgCreateRequestBase
}

type NoticeMsgCreateByProjectRequest struct {
	ProjectId int64 `json:"project_id" description:"상품 ID" required:"true" validate:"required" format:"int64" example:"1"`
	JWTUser   jwthelper.User
	ClientIp  string
	NoticeMsgCreateRequestBase
}

type NoticeMsgCreateByNewsletterRequest struct {
	JWTUser  jwthelper.User
	ClientIp string
	NoticeMsgCreateRequestBase
}

type NoticeMsgCreateByUserTypeRequest struct {
	UserTypes []string `json:"user_types" description:"투자자 유형 리스트 (PRIVATE_INVESTOR, PRIVATE_PRO_INVESTOR, CORPORATE_INVESTOR, CORPORATE_PRO_INVESTOR)" required:"true" validate:"required" example:"[\"PRIVATE_INVESTOR\",\"PRIVATE_PRO_INVESTOR\"]"`
	JWTUser   jwthelper.User
	ClientIp  string
	NoticeMsgCreateRequestBase
}
