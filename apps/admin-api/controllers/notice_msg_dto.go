package controllers

import (
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

type NoticeSearchInput struct {
	Title     *string `query:"title"`
	StartDate string  `query:"start_date" swagger:"desc(開始日)"`
	EndDate   string  `query:"end_date" swagger:"desc(終了日)"`
	SearchInput
}

type NoticeSendInput struct {
	Title       string `json:"title" form:"title" swagger:"desc(タイトル)" validate:"required"`
	ContentHtml string `json:"content_html" form:"content_html" swagger:"desc(コンテンツ)" validate:"required"`
}

type NoticeSendUserTypeInput struct {
	NoticeSendInput
	UserType string `json:"user_type" form:"user_type" swagger:"desc(user_type)" validate:"required"`
}

type NoticeSendProjectInput struct {
	NoticeSendInput
	ProjectId int64 `json:"project_id" form:"project_id" swagger:"desc(project_id)" validate:"required"`
}

func (noticeSendInput *NoticeSendInput) ToModel(noticeUserType domain.NOTICE_USER_TYPE) (*domain.NoticeMsg, error) {
	return &domain.NoticeMsg{
		Title:          noticeSendInput.Title,
		ContentHtml:    noticeSendInput.ContentHtml,
		NoticeDate:     jsonutils.JSONTime(time.Now()),
		NoticeUserType: noticeUserType,
	}, nil
}

func (noticeSendInput *NoticeSendProjectInput) ToModel(noticeUserType domain.NOTICE_USER_TYPE) (*domain.NoticeMsg, error) {
	return &domain.NoticeMsg{
		Title:          noticeSendInput.Title,
		ContentHtml:    noticeSendInput.ContentHtml,
		NoticeDate:     jsonutils.JSONTime(time.Now()),
		NoticeUserType: noticeUserType,
		ProjectId:      noticeSendInput.ProjectId,
	}, nil
}
