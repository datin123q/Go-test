package domain

import (
	"context"
	"net/http"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
)

type AdminUserUsecase interface {
	GetAll(ctx context.Context, request AdminUserGetAllRequest) (AdminUserGetAllResponse, error)
	GetOne(ctx context.Context, request AdminUserGetOneRequest) (AdminUserGetOneResponse, error)
	Create(ctx context.Context, request AdminUserCreateRequest) (AdminUserCreateResponse, error)
	Delete(ctx context.Context, request AdminUserDeleteRequest) (AdminUserDeleteResponse, error)
	Update(ctx context.Context, request AdminUserUpdateRequest) (AdminUserUpdateResponse, error)
	DownloadHistories(ctx context.Context, request AdminUserDownloadHistoriesRequest) (AdminUserDownloadHistoriesResponse, error)
	Login(ctx context.Context, request AdminUserLoginRequest) (AdminUserLoginResponse, error)
	Logout(ctx context.Context, request AdminUserLogoutRequest) (AdminUserLogoutResponse, error)
}

type AdminUserItem struct {
	Id                int64              `json:"id" description:"ID"`
	CommonId          string             `json:"common_id" description:"common ID"`
	Email             string             `json:"email" description:"이메일"`
	Name              string             `json:"name" description:"이름"`
	Roles             []string           `json:"roles" description:"권한 목록"`
	LastLoginDate     jsonutils.JSONTime `json:"last_login_date" description:"마지막 로그인 일시"`
	MobilePhone       string             `json:"mobile_phone" description:"휴대전화"`
	IsSuperAdmin      bool               `json:"is_super_admin" description:"슈퍼 관리자 여부"`
	RoleUpdated       jsonutils.JSONTime `json:"role_updated" description:"권한 업데이트 일시"`
	ValidityStartDate jsonutils.JSONTime `json:"validity_start_date" description:"비밀번호 유효기간 시작"`
	ValidityEndDate   jsonutils.JSONTime `json:"validity_end_date" description:"비밀번호 유효기간 종료"`
}

type AdminUserGetAllQueryParam struct {
	cdomain.SearchInput
	Name        *string `query:"name" description:"이름"`
	MobilePhone *string `query:"mobile_phone" description:"휴대전화" example:"01012345678"`
}

type AdminUserGetAllRequest struct {
	JWTUser  jwthelper.User
	ClientIp string
	AdminUserGetAllQueryParam
}

type AdminUserGetAllResponse struct {
	TotalCount int64           `json:"total_count" description:"총 개수"`
	Items      []AdminUserItem `json:"items" description:"운영자 목록"`
}

type AdminUserGetOneRequest struct {
	JWTUser  jwthelper.User
	ClientIp string
	Id       int64 `param:"id" path:"id" description:"조회할 운영자 ID" example:"1" format:"int64"`
}

type AdminUserGetOneResponse struct {
	AdminUserItem
}

type AdminUserCreateRequest struct {
	JWTUser  jwthelper.User
	ClientIp string

	Email             string   `json:"email" description:"메일" required:"true" validate:"required" example:"test@mail.com"`
	Password          string   `json:"password" description:"비밀번호" required:"true" validate:"required"`
	Name              string   `json:"name" description:"이름" required:"true" validate:"required"`
	MobilePhone       string   `json:"mobile_phone" description:"휴대전화" required:"true" validate:"required" example:"01012345678"`
	ValidityStartDate string   `json:"validity_start_date" description:"비밀번호 유효기간 시작" required:"true" validate:"required" example:"2021-01-01 00:00:00"`
	ValidityEndDate   string   `json:"validity_end_date" description:"비밀번호 유효기간 종료" required:"true" validate:"required" example:"2022-01-01 23:59:59"`
	Roles             []string `json:"roles" description:"권한 목록"`
}

type AdminUserCreateResponse struct {
	AdminUserItem
}

type AdminUserDeleteRequest struct {
	JWTUser  jwthelper.User
	ClientIp string
	Id       int64 `param:"id" path:"id" description:"삭제할 운영자 ID" example:"1" format:"int64"`
}

type AdminUserDeleteResponse struct {
}

type AdminUserUpdateRequest struct {
	JWTUser  jwthelper.User
	ClientIp string
	Id       int64 `param:"id" path:"id" description:"수정할 운영자 ID" example:"1" format:"int64"`

	Email             string   `json:"email" required:"true" validate:"required" description:"메일"`
	Password          *string  `json:"password" description:"비밀번호"`
	Name              string   `json:"name" required:"true" validate:"required" description:"이름"`
	MobilePhone       string   `json:"mobile_phone" required:"true" validate:"required" description:"휴대전화"`
	ValidityStartDate string   `json:"validity_start_date" required:"true" validate:"required" description:"비밀번호 유효기간 시작" example:"2021-01-01 00:00:00"`
	ValidityEndDate   string   `json:"validity_end_date" required:"true" validate:"required" description:"비밀번호 유효기간 종료" example:"2022-01-01 23:59:59"`
	Roles             []string `json:"roles" description:"권한 목록"`
}

type AdminUserUpdateResponse struct {
	AdminUserItem
}

type AdminUserDownloadHistoriesResponse struct {
	Data []byte
}

type AdminUserDownloadQueryParam struct {
	Name        *string `query:"name" description:"이름"`
	MobilePhone *string `query:"mobile_phone" description:"휴대전화"`
}

type AdminUserDownloadHistoriesRequest struct {
	JWTUser  jwthelper.User
	ClientIp string
	cdomain.SearchInput
	AdminUserDownloadQueryParam
}

type AdminUserLoginRequest struct {
	Email     string `json:"email" description:"이메일" required:"true" validate:"required"`
	Password  string `json:"password" description:"비밀번호" required:"true" validate:"required"`
	ClientIp  string
	UserAgent string
}

type AdminUserLoginResponse struct {
	AccessToken  string         `json:"access_token" description:"액세스 토큰"`
	RefreshToken string         `json:"-"`
	Cookies      []*http.Cookie `json:"-"`
}

type AdminUserLogoutRequest struct {
	JWTUser      jwthelper.User
	RefreshToken string
	ClientIp     string
	UserAgent    string
}

type AdminUserLogoutResponse struct {
}
