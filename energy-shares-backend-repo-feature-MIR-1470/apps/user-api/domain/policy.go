package domain

import (
	"context"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

type PolicyUsecase interface {
	GetAll(ctx context.Context, request PolicyGetAllRequest) (PolicyGetAllResponse, error)
	GetAllHistory(ctx context.Context, request PolicyGetAllHistoryRequest) (PolicyGetAllHistoryResponse, error)
}

type PolicyGetAllRequest struct {
	PolicyGetAllQueryParam
}

type PolicyGetAllQueryParam struct {
	cdomain.SearchInput
	PolicyType     *string `query:"policy_type" description:"約款タイプ" enum:"POLICY_CAUTION,POLICY_PROTECTION,POLICY_ANTISOCIAL_FORCES_BAGIC,POLICY_NON_ANTISOCIAL_FORCES_AGREEMENT,POLICY_SUBSTANTIVE_RULER,POLICY_SOLICITATION,POLICY_COMMISSION_RISK,POLICY_CUSTOMER_CENTERED_WORK,POLICY_ELECTRONIC_TEACHING,POLICY_NON_PEPS,POLICY_ADULT_GUARDIAN"`
	PolicyViewType *string `query:"policy_view_type" description:"約款表示タイプ" enum:"POLICY_VIEW_FOOTER,POLICY_VIEW_JOIN"`
}

type PolicyGetAllResponse struct {
	TotalCount int64              `json:"total_count" description:"合計数" format:"int64" example:"1"`
	Items      []PolicyGetAllItem `json:"items" description:"約款リスト"`
}

type PolicyGetAllItem struct {
	Id             int64                   `json:"id" description:"ID"`
	PolicyViewType domain.POLICY_VIEW_TYPE `json:"policy_view_type" description:"約款表示タイプ" enum:"POLICY_VIEW_FOOTER,POLICY_VIEW_JOIN"`
	PolicyType     domain.POLICY_TYPE      `json:"policy_type" description:"約款タイプ" enum:"POLICY_CAUTION,POLICY_PROTECTION,POLICY_ANTISOCIAL_FORCES_BAGIC,POLICY_NON_ANTISOCIAL_FORCES_AGREEMENT,POLICY_SUBSTANTIVE_RULER,POLICY_SOLICITATION,POLICY_COMMISSION_RISK,POLICY_CUSTOMER_CENTERED_WORK,POLICY_ELECTRONIC_TEACHING,POLICY_NON_PEPS,POLICY_ADULT_GUARDIAN"`
	ContentHtml    string                  `json:"content_html" description:"内容"`
	Attachment     *domain.Attachment      `json:"attachment" description:"添付ファイル"`
	Created        jsonutils.JSONTime      `json:"created" description:"作成日" format:"yyyy-MM-dd HH:mm:ss"`
	Updated        jsonutils.JSONTime      `json:"updated" description:"更新日" format:"yyyy-MM-dd HH:mm:ss"`
}

type PolicyGetAllHistoryRequest struct {
	PolicyGetAllHistoryQueryParam
}

type PolicyGetAllHistoryQueryParam struct {
	cdomain.SearchInput
	PolicyType     *string `query:"policy_type" description:"約款タイプ" required:"true" validate:"required" enum:"POLICY_CAUTION,POLICY_PROTECTION"`
	PolicyViewType *string `query:"policy_view_type" description:"約款表示タイプ" required:"true" validate:"required" enum:"POLICY_VIEW_FOOTER"`
}

type PolicyGetAllHistoryResponse struct {
	TotalCount int64                     `json:"total_count" description:"合計数" format:"int64" example:"1"`
	Items      []PolicyGetAllHistoryItem `json:"items" description:"約款履歴リスト"`
}

type PolicyGetAllHistoryItem struct {
	Id             int64                   `json:"id" description:"ID"`
	PolicyViewType domain.POLICY_VIEW_TYPE `json:"policy_view_type" description:"約款表示タイプ" enum:"POLICY_VIEW_FOOTER"`
	PolicyType     domain.POLICY_TYPE      `json:"policy_type" description:"約款タイプ" enum:"POLICY_CAUTION,POLICY_PROTECTION"`
	ContentHtml    string                  `json:"content_html" description:"内容"`
	IsPosted       bool                    `json:"is_posted" description:"掲載の有無" format:"bool" example:"true"`
	Label          *string                 `json:"label" description:"ラベル" format:"yyyy-MM-dd" example:"2024-07-26"`
	Created        jsonutils.JSONTime      `json:"created" description:"作成日" format:"yyyy-MM-dd HH:mm:ss"`
}
