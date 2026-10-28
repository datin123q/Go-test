package domain

import (
	"context"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

type PolicyUsecase interface {
	Update(ctx context.Context, request PolicyUpdateRequest) (PolicyUpdateResponse, error)
	GetAll(ctx context.Context, request PolicyGetAllRequest) (PolicyGetAllResponse, error)
	GetOne(ctx context.Context, request PolicyGetOneRequest) (PolicyGetOneResponse, error)
	GetAllHistory(ctx context.Context, request PolicyGetAllHistoryRequest) (PolicyGetAllHistoryResponse, error)
	UpdateHistory(ctx context.Context, request PolicyUpdateHistoryRequest) (PolicyUpdateHistoryResponse, error)
	Delete(ctx context.Context, request PolicyDeleteRequest) error
	CheckLabelDuplicate(ctx context.Context, request PolicyCheckLabelDuplicateRequest) (PolicyCheckLabelDuplicateResponse, error)
}

type PolicyUpdateRequest struct {
	JWTUser        jwthelper.User
	Id             int64                   `param:"id" path:"id" description:"단건 업데이트를 위한 policy ID" format:"int64"`
	PolicyViewType domain.POLICY_VIEW_TYPE `json:"policy_view_type" description:"약관 보기 타입" required:"true" format:"string" enum:"POLICY_VIEW_FOOTER,POLICY_VIEW_JOIN" validate:"oneof=POLICY_VIEW_FOOTER POLICY_VIEW_JOIN"`
	ContentHtml    string                  `json:"content_html" description:"내용" format:"html" example:"<body><h1>테스트입니다.</h1></body>"`
	AttachmentUrls []string                `json:"attachment_urls" description:"첨부파일 URL 리스트" format:"[]string" example:"[https://es-jp-devv2-public-file-storage.s3.ap-northeast-2.amazonaws.com/attachment_file/2024-07-26/cqhkh82nulri2.jpg]"`
}

type PolicyUpdateResponse struct {
	Id             int64                   `json:"id" description:"ID"`
	CommonId       string                  `json:"common_id" description:"공통 ID"`
	OwnerId        int64                   `json:"owner_id" description:"소유자 ID"`
	ModifierId     int64                   `json:"modifier_id" description:"수정자 ID"`
	PolicyViewType domain.POLICY_VIEW_TYPE `json:"policy_view_type" description:"약관 보기 타입" enum:"POLICY_VIEW_FOOTER,POLICY_VIEW_JOIN"`
	PolicyType     domain.POLICY_TYPE      `json:"policy_type" description:"약관 타입" enum:"POLICY_CAUTION,POLICY_PROTECTION,POLICY_ANTISOCIAL_FORCES_BAGIC,POLICY_NON_ANTISOCIAL_FORCES_AGREEMENT,POLICY_SUBSTANTIVE_RULER,POLICY_SOLICITATION,POLICY_COMMISSION_RISK,POLICY_CUSTOMER_CENTERED_WORK,POLICY_ELECTRONIC_TEACHING,POLICY_NON_PEPS,POLICY_ADULT_GUARDIAN"`
	ContentHtml    string                  `json:"content_html" description:"내용"`
	ModifierEmail  string                  `json:"modifier_email" description:"수정자 이메일"`
	Attachment     *domain.Attachment      `json:"attachment" description:"첨부파일"`
	Created        jsonutils.JSONTime      `json:"created" description:"생성일" format:"yyyy-MM-dd HH:mm:ss"`
	Updated        jsonutils.JSONTime      `json:"updated" description:"업데이트일" format:"yyyy-MM-dd HH:mm:ss"`
}

type PolicyGetAllRequest struct {
	JWTUser jwthelper.User
	PolicyGetAllQueryParam
}

type PolicyGetAllQueryParam struct {
	cdomain.SearchInput
	PolicyViewType *string `query:"policy_view_type" description:"약관 보기 타입" enum:"POLICY_VIEW_FOOTER,POLICY_VIEW_JOIN"`
	PolicyType     *string `query:"policy_type" description:"약관 타입" enum:"POLICY_CAUTION,POLICY_PROTECTION,POLICY_ANTISOCIAL_FORCES_BAGIC,POLICY_NON_ANTISOCIAL_FORCES_AGREEMENT,POLICY_SUBSTANTIVE_RULER,POLICY_SOLICITATION,POLICY_COMMISSION_RISK,POLICY_CUSTOMER_CENTERED_WORK,POLICY_ELECTRONIC_TEACHING,POLICY_NON_PEPS,POLICY_ADULT_GUARDIAN"`
}

type PolicyGetAllResponse struct {
	TotalCount int64              `json:"total_count" description:"총 개수" format:"int64" example:"1"`
	Items      []PolicyGetAllItem `json:"items" description:"policy 리스트"`
}

type PolicyGetAllItem struct {
	Id             int64                   `json:"id" description:"ID"`
	CommonId       string                  `json:"common_id" description:"공통 ID"`
	OwnerId        int64                   `json:"owner_id" description:"소유자 ID"`
	ModifierId     int64                   `json:"modifier_id" description:"수정자 ID"`
	PolicyViewType domain.POLICY_VIEW_TYPE `json:"policy_view_type" description:"약관 보기 타입" enum:"POLICY_VIEW_FOOTER,POLICY_VIEW_JOIN"`
	PolicyType     domain.POLICY_TYPE      `json:"policy_type" description:"약관 타입" enum:"POLICY_CAUTION,POLICY_PROTECTION,POLICY_ANTISOCIAL_FORCES_BAGIC,POLICY_NON_ANTISOCIAL_FORCES_AGREEMENT,POLICY_SUBSTANTIVE_RULER,POLICY_SOLICITATION,POLICY_COMMISSION_RISK,POLICY_CUSTOMER_CENTERED_WORK,POLICY_ELECTRONIC_TEACHING,POLICY_NON_PEPS,POLICY_ADULT_GUARDIAN"`
	Created        jsonutils.JSONTime      `json:"created" description:"생성일" format:"yyyy-MM-dd HH:mm:ss"`
	Updated        jsonutils.JSONTime      `json:"updated" description:"업데이트일" format:"yyyy-MM-dd HH:mm:ss"`
}

type PolicyGetOneRequest struct {
	JWTUser jwthelper.User
	Id      int64 `param:"id" path:"id" description:"단건 조회를 위한 policy ID" example:"1" format:"int64"`
}

type PolicyGetOneResponse struct {
	Id             int64                   `json:"id" description:"ID"`
	CommonId       string                  `json:"common_id" description:"공통 ID"`
	OwnerId        int64                   `json:"owner_id" description:"소유자 ID"`
	ModifierId     int64                   `json:"modifier_id" description:"수정자 ID"`
	PolicyViewType domain.POLICY_VIEW_TYPE `json:"policy_view_type" description:"약관 보기 타입" enum:"POLICY_VIEW_FOOTER,POLICY_VIEW_JOIN"`
	PolicyType     domain.POLICY_TYPE      `json:"policy_type" description:"약관 타입" enum:"POLICY_CAUTION,POLICY_PROTECTION,POLICY_ANTISOCIAL_FORCES_BAGIC,POLICY_NON_ANTISOCIAL_FORCES_AGREEMENT,POLICY_SUBSTANTIVE_RULER,POLICY_SOLICITATION,POLICY_COMMISSION_RISK,POLICY_CUSTOMER_CENTERED_WORK,POLICY_ELECTRONIC_TEACHING,POLICY_NON_PEPS,POLICY_ADULT_GUARDIAN"`
	ContentHtml    string                  `json:"content_html" description:"내용"`
	ModifierEmail  string                  `json:"modifier_email" description:"수정자 이메일"`
	Attachment     *domain.Attachment      `json:"attachment" description:"첨부파일"`
	Created        jsonutils.JSONTime      `json:"created" description:"생성일" format:"yyyy-MM-dd HH:mm:ss"`
	Updated        jsonutils.JSONTime      `json:"updated" description:"업데이트일" format:"yyyy-MM-dd HH:mm:ss"`
}

type PolicyGetAllHistoryRequest struct {
	JWTUser jwthelper.User
	Id      int64 `param:"id" path:"id" description:"히스토리 조회를 위한 policy ID" example:"1" format:"int64"`
	PolicyGetAllHistoryQueryParam
}

type PolicyGetAllHistoryQueryParam struct {
	cdomain.SearchInput
	IsPosted *bool `query:"is_posted" description:"게시 여부" format:"bool" example:"true"`
}

type PolicyGetAllHistoryResponse struct {
	TotalCount int64               `json:"total_count" description:"총 개수" format:"int64" example:"1"`
	Items      []PolicyHistoryItem `json:"items" description:"policy 히스토리 리스트"`
}

type PolicyHistoryItem struct {
	Id              int64              `json:"id" description:"ID"`
	PolicyId        int64              `json:"policy_id" description:"약관 ID"`
	SubstanceType   string             `json:"substance_type" description:"내용 표기방식"`
	ContentType     string             `json:"content_type" description:"변경된 내용 타입"`
	Name            string             `json:"name" description:"처리자"`
	Email           string             `json:"email" description:"처리자 이메일"`
	IsPosted        bool               `json:"is_posted" description:"게시 여부" format:"bool" example:"true"`
	Label           *string            `json:"label" description:"라벨" format:"yyyy-MM-dd" example:"2024-07-26"`
	PrevContent     string             `json:"prev_content" description:"변경 전 내용"`
	CurrContent     string             `json:"curr_content" description:"변경 후 내용"`
	PrevContentHtml string             `json:"prev_content_html" description:"변경 전 HTML"`
	CurrContentHtml string             `json:"curr_content_html" description:"변경 후 HTML"`
	PrevUrl         string             `json:"prev_url" description:"변경 전 URL"`
	CurrUrl         string             `json:"curr_url" description:"변경 후 URL"`
	Created         jsonutils.JSONTime `json:"created" description:"처리일시" format:"yyyy-MM-dd HH:mm:ss"`
}

type PolicyUpdateHistoryRequest struct {
	JWTUser   jwthelper.User
	Id        int64   `param:"id" path:"id" description:"policy ID" example:"1" format:"int64"`
	HistoryId int64   `param:"history_id" path:"history_id" description:"policy history ID" example:"1" format:"int64"`
	IsPosted  bool    `json:"is_posted" description:"게시 여부" format:"bool" example:"true"`
	Label     *string `json:"label" description:"라벨" format:"yyyy-MM-dd" example:"2024-07-26"`
}

func (r PolicyUpdateHistoryRequest) Valid(c echo.Context) error {
	if err := c.Validate(&r); err != nil {
		return errors.Wrap(err, "PolicyUpdateRequest validation error")
	}

	if r.IsPosted && r.Label == nil {
		return errors.New("Label is required when is_posted is true")
	}

	return nil
}

type PolicyUpdateHistoryResponse struct {
	PolicyHistoryItem
}

type PolicyDeleteRequest struct {
	JWTUser jwthelper.User
	Id      int64 `param:"id" path:"id" description:"단건 삭제를 위한 policy ID" example:"1" format:"int64"`
}

type PolicyCheckLabelDuplicateRequest struct {
	JWTUser jwthelper.User
	Id      int64   `param:"id" path:"id" description:"policy ID" example:"1" format:"int64"`
	Label   *string `query:"label" description:"약관 이력 라벨" format:"yyyy-MM-dd" required:"true" validate:"required" example:"2024-07-26"`
}

type PolicyCheckLabelDuplicateResponse struct {
	IsDuplicate   bool               `json:"is_duplicate" description:"중복 여부" format:"bool" example:"true"`
	PolicyHistory *PolicyHistoryItem `json:"policy_history" description:"중복된 policy 히스토리"`
}
