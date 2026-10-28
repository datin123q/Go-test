package controllers

import (
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/stringutils"
)

// ContactInput はお問い合わせメール送信のリクエストボディを表す
type ContactInput struct {
	Name     string `json:"name" swagger:"desc(お名前)" validate:"required"`
	Email    string `json:"email" swagger:"desc(メールアドレス)" validate:"required"`
	Title    string `json:"title" swagger:"desc(件名)" validate:"required"`
	Contents string `json:"contents" swagger:"desc(内容)" validate:"required"`
}

// Sanitize はXSS攻撃を防ぐためにHTMLタグをエスケープする
func (ci *ContactInput) Sanitize() {
	ci.Name = stringutils.SanitizeHTML(ci.Name)
	ci.Title = stringutils.SanitizeHTML(ci.Title)
	ci.Contents = stringutils.SanitizeHTMLKeepBR(ci.Contents)
}

func (ci ContactInput) ToModel() (*domain.Contact, error) {
	return &domain.Contact{
		Name:     ci.Name,
		Email:    ci.Email,
		Title:    ci.Title,
		Contents: ci.Contents,
	}, nil
}
