package controllers

import (
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

type ProjectNewsInput struct {
	ProjectId   int64  `json:"project_id" form:"project_id" validate:"required"`
	Title       string `json:"title" form:"title" validate:"required"`
	ContentHtml string `json:"content_html" form:"content_html"`
	Posted      bool   `json:"posted" form:"posted"`
	PostDate    string `json:"post_date" form:"post_date" swagger:"desc(掲載日),default(2019-12-25 00:00:00)"`
}

func (d *ProjectNewsInput) ToModel() (*domain.ProjectNews, error) {
	if d.PostDate == "" {
		d.PostDate = "2020-01-01 00:00:00"
	}
	postDate, err := jsonutils.ParseJSONTimeJST(time.DateTime, d.PostDate)
	if err != nil {
		return nil, err
	}

	return &domain.ProjectNews{
		ProjectId:   d.ProjectId,
		Title:       d.Title,
		ContentHtml: d.ContentHtml,
		Posted:      d.Posted,
		PostDate:    postDate,
	}, nil
}
