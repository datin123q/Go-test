package domain

import (
	"context"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
)

type FundingRequestUsecase interface {
	GetAll(ctx context.Context, request FundingRequestGetAllRequest) (FundingRequestGetAllResponse, error)
	GetOne(ctx context.Context, request FundingRequestGetOneRequest) (FundingRequestGetOneResponse, error)
	Update(ctx context.Context, request FundingRequestUpdateRequest) error
}
type FundingRequestGetAllRequest struct {
	JWTUser  jwthelper.User
	ClientIp string
	cdomain.SearchInput
}

type FundingRequestGetAllResponse struct {
	TotalCount int64                   `json:"total_count" description:"총 개수" format:"int64" example:"1"`
	Items      []domain.FundingRequest `json:"items" description:"펀딩 요청 목록"`
}

type FundingRequestGetOneResponse struct {
	domain.FundingRequest
}

type FundingRequestGetOneRequest struct {
	JWTUser  jwthelper.User
	ClientIp string
	Id       int64 `param:"id" path:"id" description:"단건조회를 위한 ID" example:"1" format:"int64"`
}

type FundingRequestUpdateRequest struct {
}
