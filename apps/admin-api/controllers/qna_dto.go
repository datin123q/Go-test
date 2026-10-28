package controllers

import (
	"github.com/illuminarean/energy-shares-backend-repo/domain"
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
	UserId   int64  `json:"user_id" form:"user_id"`
	Title    string `json:"title" form:"title" validate:"required" swagger:"desc(問い合わせ回答題目)"`
	Content  string `json:"content" form:"content" validate:"required" swagger:"desc(問い合わせ内容)"`
	CommonId string `json:"common_id" form:"common_id"`
}

func (q *QnaInput) ToModel() (*domain.Qna, error) {

	return &domain.Qna{
		UserId:   q.UserId,
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
