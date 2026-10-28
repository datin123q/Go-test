package domain

import (
	"context"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
)

type ChangelogUsecase interface {
	GetAll(ctx context.Context, request ChangelogGetAllRequest) (ChangelogGetAllResponse, error)
}

type ChangelogGetAllRequest struct {
	JWTUser  jwthelper.User
	ClientIp string
	ChangelogGetAllQueryParam
}

type ChangelogGetAllQueryParam struct {
	Id *string `query:"id" description:"조회할 ID" format:"string" example:"123"`
	cdomain.KsqlSearchInput
}

type ChangelogGetAllResponse struct {
	Items []domain.ChangelogRow `json:"items" description:"변경로그 목록"`
}
