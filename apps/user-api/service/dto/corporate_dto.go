package dto

import (
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

type CorporateShareholderInput struct {
	UserId         int64  `json:"user_id" form:"user_id"`
	FamilyName     string `json:"family_name" form:"family_name" swagger:"desc(株主名姓)"`
	FirstName      string `json:"first_name" form:"first_name" swagger:"desc(株主名名前)"`
	FamilyNameKana string `json:"family_name_kana" form:"family_name_kana"  swagger:"desc(株主名姓振り仮名)"`
	FirstNameKana  string `json:"first_name_kana" form:"first_name_kana"  swagger:"desc(株主名名前振り仮名)"`
	Birthday       string `json:"birthday" form:"birthday" swagger:"desc(生年月日)"`
	ZipCode        string `json:"zip_code" form:"zip_code" swagger:"desc(郵便番号)"`
	Address        string `json:"address" form:"address" swagger:"desc(住所)"`
	Address2       string `json:"address2" form:"address2" swagger:"desc(詳細住所)"`
	Province       string `json:"province"  form: "province" swagger:"desc(都道府県名)"`
	City           string `json:"city"  form:"city" swagger:"desc(市区町村)"`
}

func (c *CorporateShareholderInput) ToModel() (*domain.CorporateShareholder, error) {
	bd, err := jsonutils.ParseJSONTimeJST(time.DateTime, c.Birthday)
	if err != nil {
		return nil, err
	}

	return &domain.CorporateShareholder{
		UserId:         c.UserId,
		FirstName:      c.FirstName,
		FamilyName:     c.FamilyName,
		FirstNameKana:  c.FirstNameKana,
		FamilyNameKana: c.FamilyNameKana,
		City:           c.City,
		Province:       c.Province,
		Address:        c.Address,
		Address2:       c.Address2,
		ZipCode:        c.ZipCode,
		Birthday:       bd,
	}, nil
}
