package domain

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"github.com/samber/lo"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
)

type BannerUsecase interface {
	Create(context.Context, BannerCreateRequest) (BannerCreateResponse, error)
	GetAll(context.Context, BannerGetAllRequest) (BannerGetAllResponse, error)
	GetOne(context.Context, BannerGetOneRequest) (BannerGetOneResponse, error)
	Update(context.Context, BannerUpdateRequest) (BannerUpdateResponse, error)
	Delete(context.Context, BannerDeleteRequest) error
	UpdateRollingTime(context.Context, BannerUpdateRollingTimeRequest) error
}

type BannerCreateRequest struct {
	JWTUser     jwthelper.User `json:"-" swaggerignore:"true"`
	Title       string         `json:"title" swagger:"desc(제목, 최대 80자)" valid:"required,max=80"`
	IsPosted    *bool          `json:"posted" swagger:"desc(게시 여부 (default: true))" valid:"required"`
	StartDate   string         `json:"start_date" swagger:"desc(게시 시작일, yyyy-MM-dd HH:mm:ss)" valid:"required"`
	EndDate     string         `json:"end_date" swagger:"desc(게시 종료일, yyyy-MM-dd 23:59:59)" valid:"required"`
	Url         string         `json:"url" swagger:"desc(url)" valid:"required"`
	SortOrder   *int64         `json:"sort_order" swagger:"desc(순서, 1~10)"`
	Description *string        `json:"description" swagger:"desc(설명)"`
	NewTab      *bool          `json:"new_tab" swagger:"desc(새 탭 여부 (default: false))" valid:"required"`
}

func (b *BannerCreateRequest) Validate() error {
	validate := validator.New()
	if err := validate.Struct(b); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// default: true
	if b.IsPosted == nil {
		b.IsPosted = lo.ToPtr(true)
	}

	// IsPosted 가 true 인 경우 Order 는 필수
	if lo.FromPtr(b.IsPosted) {
		if b.SortOrder == nil {
			return factory.ErrorParameter.New(errors.New("sort_order is required"))
		}

		// SortOrder 는 1~10 사이의 값이어야 함
		if *b.SortOrder < 1 || *b.SortOrder > 10 {
			return factory.ErrorParameter.New(errors.New("sort_order must be between 1 and 10"))
		}
	}

	return nil
}

type BannerCreateResponse struct {
	Id          int64              `json:"id" swagger:"desc(배너 ID)"`
	Title       string             `json:"title" swagger:"desc(제목)"`
	IsPosted    bool               `json:"posted" swagger:"desc(게시 여부)"`
	StartDate   jsonutils.JSONTime `json:"start_date" swagger:"desc(게시 시작일, yyyy-MM-dd HH:mm:ss)"`
	EndDate     jsonutils.JSONTime `json:"end_date" swagger:"desc(게시 종료일, yyyy-MM-dd 23:59:59)"`
	Url         string             `json:"url" swagger:"desc(url)"`
	SortOrder   *int64             `json:"sort_order" swagger:"desc(순서)"`
	Description string             `json:"description" swagger:"desc(설명)"`
	NewTab      bool               `json:"new_tab" swagger:"desc(새 탭 여부)"`
	Created     jsonutils.JSONTime `json:"created" swagger:"desc(생성일, yyyy-MM-dd HH:mm:ss)"`
	Updated     jsonutils.JSONTime `json:"updated" swagger:"desc(수정일, yyyy-MM-dd HH:mm:ss)"`
}

type BannerGetAllQueryParam struct {
	cdomain.SearchInput
	Title            *string `query:"title" swagger:"desc(제목)"`
	StartDate        *string `query:"start_date" swagger:"desc(게시 시작일, yyyy-MM-dd HH:mm:ss)"`
	EndDate          *string `query:"end_date" swagger:"desc(게시 종료일, yyyy-MM-dd 23:59:59)"`
	IsPosted         *bool   `query:"posted" swagger:"desc(게시 여부)"`
	CreatedStartDate *string `query:"created_start_date" swagger:"desc(생성일, yyyy-MM-dd HH:mm:ss)"`
	CreatedEndDate   *string `query:"created_end_date" swagger:"desc(생성일, yyyy-MM-dd HH:mm:ss)"`
}

type BannerGetAllRequest struct {
	JWTUser jwthelper.User `json:"-" swaggerignore:"true"`
	BannerGetAllQueryParam
}

type BannerGetAllResponse struct {
	TotalCount        int64              `json:"total_count" swagger:"desc(총 개수)"`
	TotalBannerPosted int64              `json:"total_banner_posted" swagger:"desc(게시중 개수)"`
	RollingSecond     int64              `json:"rolling_second" swagger:"desc(롤링 시간 (단위: 초))"`
	Items             []BannerGetAllItem `json:"items" swagger:"desc(배너 목록)"`
}

type BannerGetAllItem struct {
	Id        int64              `json:"id" swagger:"desc(배너 ID)"`
	Title     string             `json:"title" swagger:"desc(제목)"`
	IsPosted  bool               `json:"posted" swagger:"desc(게시 여부)"`
	SortOrder *int64             `json:"sort_order" swagger:"desc(순서)"`
	StartDate jsonutils.JSONTime `json:"start_date" swagger:"desc(게시 시작일, yyyy-MM-dd HH:mm:ss)"`
	EndDate   jsonutils.JSONTime `json:"end_date" swagger:"desc(게시 종료일, yyyy-MM-dd 23:59:59)"`
	Created   jsonutils.JSONTime `json:"created" swagger:"desc(생성일)"`
}

type BannerGetOneRequest struct {
	JWTUser jwthelper.User `json:"-" swaggerignore:"true"`
	Id      int64          `param:"id" swagger:"desc(배너 ID)" valid:"required"`
}

type BannerGetOneResponse struct {
	Id          int64              `json:"id" swagger:"desc(배너 ID)"`
	Title       string             `json:"title" swagger:"desc(제목)"`
	IsPosted    bool               `json:"posted" swagger:"desc(게시 여부)"`
	StartDate   jsonutils.JSONTime `json:"start_date" swagger:"desc(게시 시작일, yyyy-MM-dd HH:mm:ss)"`
	EndDate     jsonutils.JSONTime `json:"end_date" swagger:"desc(게시 종료일, yyyy-MM-dd 23:59:59)"`
	Url         string             `json:"url" swagger:"desc(url)"`
	SortOrder   *int64             `json:"sort_order" swagger:"desc(순서)"`
	Description string             `json:"description" swagger:"desc(설명)"`
	NewTab      bool               `json:"new_tab" swagger:"desc(새 탭 여부)"`
	Created     jsonutils.JSONTime `json:"created" swagger:"desc(생성일, yyyy-MM-dd HH:mm:ss)"`
	Updated     jsonutils.JSONTime `json:"updated" swagger:"desc(수정일, yyyy-MM-dd HH:mm:ss)"`
}

type BannerUpdateRequest struct {
	JWTUser     jwthelper.User `json:"-" swaggerignore:"true"`
	Id          int64          `param:"id" swagger:"desc(배너 ID)" valid:"required"`
	Title       string         `json:"title" swagger:"desc(제목, 최대 80자)" valid:"required,max=80"`
	IsPosted    *bool          `json:"posted" swagger:"desc(게시 여부 (default: true))" valid:"required"`
	StartDate   string         `json:"start_date" swagger:"desc(게시 시작일, YYYY-MM-DD HH:mm:ss)" valid:"required"`
	EndDate     string         `json:"end_date" swagger:"desc(게시 종료일, YYYY-MM-DD 23:59:59)" valid:"required"`
	Url         string         `json:"url" swagger:"desc(url)" valid:"required"`
	SortOrder   *int64         `json:"sort_order" swagger:"desc(순서, 1~10)"`
	Description *string        `json:"description" swagger:"desc(설명)"`
	NewTab      *bool          `json:"new_tab" swagger:"desc(새 탭 여부 (default: false))" valid:"required"`
}

func (b *BannerUpdateRequest) Validate() error {
	validate := validator.New()
	if err := validate.Struct(b); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// default: true
	if b.IsPosted == nil {
		b.IsPosted = lo.ToPtr(true)
	}

	// IsPosted 가 true 인 경우 Order 는 필수
	if lo.FromPtr(b.IsPosted) {
		if b.SortOrder == nil {
			return factory.ErrorParameter.New(errors.New("order is required"))
		}

		// Order 는 1~10 사이의 값이어야 함
		if *b.SortOrder < 1 || *b.SortOrder > 10 {
			return factory.ErrorParameter.New(errors.New("order must be between 1 and 10"))
		}
	} else {
		// IsPosted 가 false 인 경우 Order 는 무시
		b.SortOrder = nil
	}

	return nil
}

type BannerUpdateResponse struct {
	Id          int64              `json:"id" swagger:"desc(배너 ID)"`
	Title       string             `json:"title" swagger:"desc(제목)"`
	IsPosted    bool               `json:"posted" swagger:"desc(게시 여부)"`
	StartDate   jsonutils.JSONTime `json:"start_date" swagger:"desc(게시 시작일)"`
	EndDate     jsonutils.JSONTime `json:"end_date" swagger:"desc(게시 종료일)"`
	Url         string             `json:"url" swagger:"desc(url)"`
	SortOrder   *int64             `json:"sort_order" swagger:"desc(순서)"`
	Description string             `json:"description" swagger:"desc(설명)"`
	NewTab      bool               `json:"new_tab" swagger:"desc(새 탭 여부)"`
	Created     jsonutils.JSONTime `json:"created" swagger:"desc(생성일)"`
	Updated     jsonutils.JSONTime `json:"updated" swagger:"desc(수정일)"`
}

type BannerUpdateRollingTimeRequest struct {
	JWTUser       jwthelper.User `json:"-" swaggerignore:"true"`
	RollingSecond int64          `json:"rolling_second" swagger:"desc(롤링 시간 (단위: 초))" valid:"required"`
}

type BannerDeleteRequest struct {
	JWTUser jwthelper.User `json:"-" swaggerignore:"true"`
	Id      int64          `param:"id" swagger:"desc(배너 ID)" valid:"required"`
}
