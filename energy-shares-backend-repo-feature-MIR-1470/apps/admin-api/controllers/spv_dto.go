package controllers

import (
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/samber/lo"
)

type SpvSearchInput struct {
	Id    *string `query:"id"`
	Name  *string `query:"name"`
	Email *string `query:"email"`
	Phone *string `query:"phone"`
	SearchInput
}

type SpvInput struct {
	Name     string  `json:"name" form:"name" swagger:"desc(会社名)" validate:"required"`
	Email    *string `json:"email" form:"email" swagger:"desc(メール)"`
	Address  string  `json:"address" form:"address" swagger:"desc(住所)" validate:"required"`
	Address2 *string `json:"address2" form:"address2" swagger:"desc(住所2)"`
	City     string  `json:"city" form:"city" swagger:"desc(都市)" validate:"required"`
	ZipCode  string  `json:"zip_code" form:"zip_code" swagger:"desc(住所)" validate:"required"`
	Province string  `json:"province" form:"province" swagger:"desc(都道府県名)" validate:"required"` // 都道府県名
	Phone    string  `json:"phone" form:"phone" swagger:"desc(電話番号)" validate:"required"`
	CorpNo   string  `json:"corp_no" form:"corp_no" swagger:"desc(事業者登録番号)" validate:"required"`
}

type SpvCropNoInput struct {
	CorpNo string `json:"corp_no" form:"corp_no" query:"CorpNo"  swagger:"dsec(事業者登録番号)" validate:"required"`
}

func (d *SpvInput) ToModel() (*domain.Spv, error) {

	return &domain.Spv{
		Name:     d.Name,
		Email:    lo.FromPtr(d.Email),
		Address:  d.Address,
		Address2: lo.FromPtr(d.Address2),
		City:     d.City,
		ZipCode:  d.ZipCode,
		Province: d.Province,
		Phone:    d.Phone,
		CorpNo:   d.CorpNo,
	}, nil
}
