package domain

import (
	"context"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/commoncode"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
)

type OperatorHistoryUsecase interface {
	GetAll(ctx context.Context, request OperatorHistoryGetAllRequest) (OperatorHistoryGetAllResponse, error)
	Download(ctx context.Context, request OperatorHistoryDownloadRequest) (OperatorHistoryDownloadResponse, error)
	GetMenuCodes(ctx context.Context, request GetCodeListRequest) (GetCodeListResponse, error)
	GetLogCodes(ctx context.Context, request GetCodeListRequest) (GetCodeListResponse, error)
}

type OperatorHistoryGetAllRequest struct {
	JWTUser  jwthelper.User
	ClientIp string
	OperatorHistoryGetAllQueryParam
}

type OperatorHistoryGetAllQueryParam struct {
	cdomain.KsqlSearchInput
}

type OperatorHistoryGetAllResponse struct {
	TotalCount int64                               `json:"total_count" description:"총 개수" format:"int64" example:"1"`
	Items      []domain.OperatorHistoryRowResponse `json:"items" description:"운영자 로그 리스트"`
}

type OperatorHistoryDownloadRequest struct {
	JWTUser  jwthelper.User
	ClientIp string
	OperatorHistoryGetAllQueryParam
}

type OperatorHistoryDownloadResponse struct {
	Data []byte `json:"data" description:"운영자 로그 엑셀파일 데이터"`
}

type GetCodeListRequest struct {
	JWTUser  jwthelper.User
	ClientIp string
	Locale   string `query:"locale" description:"로케일"`
}

type GetCodeListResponse struct {
	Codes []commoncode.Code `json:"codes"`
}
