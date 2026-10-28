package controllers

import "github.com/illuminarean/energy-shares-backend-repo/domain"

type ConsultNoteInput struct {
	Title   string `json:"title" swagger:"desc(タイトル)" validate:"required"`
	Content string `json:"content" swagger:"desc(コンテンツ)" validate:"required"`
}

func (cni *ConsultNoteInput) ToModel() *domain.ConsultNote {
	return &domain.ConsultNote{
		Title:   cni.Title,
		Content: cni.Content,
	}
}

type ConsultNoteUpdateInput struct {
	Title   *string `json:"title" swagger:"desc(タイトル)"`
	Content *string `json:"content" swagger:"desc(コンテンツ)"`
}

func (cni *ConsultNoteUpdateInput) ToModel() *domain.ConsultNote {
	consultNote := &domain.ConsultNote{}
	if cni.Title != nil {
		consultNote.Title = *cni.Title
	}
	if cni.Content != nil {
		consultNote.Content = *cni.Content
	}

	return consultNote
}
