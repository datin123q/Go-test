package controllers

import (
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/stringutils"
)

type QnaSearchInput struct {
	UserId    *string `query:"user_id"`
	Name      *string `query:"name"`
	Title     *string `query:"title"`
	Status    *string `query:"status"`
	StartDate string  `query:"start_date"`
	EndDate   string  `query:"end_date"`
	SearchInput
}

type QnaInput struct {
	Title    string `json:"title" form:"title" validate:"min=1,max=50" swagger:"desc(Q&A件名)"`
	Content  string `json:"content" form:"content" validate:"min=1,max=1000" swagger:"desc(Q&A内容)"`
	CommonId string `json:"common_id" form:"common_id"`
}

// Sanitize はXSS攻撃を防ぐためにHTMLタグをエスケープする
func (q *QnaInput) Sanitize() {
	q.Title = stringutils.SanitizeHTML(q.Title)
	q.Content = stringutils.SanitizeHTMLKeepBR(q.Content)
}

func (q *QnaInput) ToModel() (*domain.Qna, error) {
	return &domain.Qna{
		Title:    q.Title,
		Content:  q.Content,
		CommonId: q.CommonId,
	}, nil
}

type AnswerInput struct {
	Answer   string `json:"answer" form:"answer" validate:"required" swagger:"desc(回答)"`
	CommonId string `json:"common_id" form:"common_id"`
}

func (a *AnswerInput) ToModel() (*domain.Answer, error) {
	return &domain.Answer{
		Answer:   a.Answer,
		CommonId: a.CommonId,
	}, nil
}
