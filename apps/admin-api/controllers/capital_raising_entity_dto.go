package controllers

import (
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

type CapitalRaisingEntitySearchInput struct {
	SearchInput
}

type CapitalRaisingEntityInput struct {
	CorporateName                string `json:"corporate_name" form:"corporate_name" swagger:"desc(法人名)"`
	ZipCode                      string `json:"zip_code" form:"zip_code" swagger:"desc(郵便番号)"`
	Province                     string `json:"province" form:"province" swagger:"desc(都道府県)"`
	City                         string `json:"city" form:"city" swagger:"desc(市区町村)"`
	Address                      string `json:"address" form:"address" swagger:"desc(住所)"`
	Address2                     string `json:"address2" form:"address2" swagger:"desc(建物名・部屋番号など)"`
	RepresentativeFamilyName     string `json:"representative_family_name" form:"representative_family_name" swagger:"desc(代表者 姓)"`
	RepresentativeFirstName      string `json:"representative_first_name" form:"representative_first_name" swagger:"desc(代表者 名)"`
	RepresentativeFamilyNameKana string `json:"representative_family_name_kana" form:"representative_family_name_kana" swagger:"desc(代表者 姓（カナ）)"`
	RepresentativeFirstNameKana  string `json:"representative_first_name_kana" form:"representative_first_name_kana" swagger:"desc(代表者 名（カナ）)"`
}

type CapitalRaisingEntityAdminResponseDto struct {
	Id                           int64              `json:"id"`                              // ID
	CorporateName                string             `json:"corporate_name"`                  // 法人名
	ZipCode                      string             `json:"zip_code"`                        // 郵便番号
	Province                     string             `json:"province"`                        // 都道府県
	City                         string             `json:"city"`                            // 市区町村
	Address                      string             `json:"address"`                         // 住所
	Address2                     string             `json:"address2"`                        // 建物名・部屋番号など
	RepresentativeFamilyName     string             `json:"representative_family_name"`      // 代表者 姓
	RepresentativeFirstName      string             `json:"representative_first_name"`       // 代表者 名
	RepresentativeFamilyNameKana string             `json:"representative_family_name_kana"` // 代表者 姓（カナ）
	RepresentativeFirstNameKana  string             `json:"representative_first_name_kana"`  // 代表者 名（カナ）
	ImageUrl                     *string            `json:"image_url"`                       // 画像URL
	Created                      jsonutils.JSONTime `json:"created"`                         // 作成日時
	Updated                      jsonutils.JSONTime `json:"updated"`                         // 更新日時
}

func (receiver CapitalRaisingEntityAdminResponseDto) NewCapitalRaisingEntityAdminResponseDto(c *domain.CapitalRaisingEntity) *CapitalRaisingEntityAdminResponseDto {
	if c == nil {
		return nil
	}

	return &CapitalRaisingEntityAdminResponseDto{
		Id:                           c.Id,
		CorporateName:                c.CorporateName,
		ZipCode:                      c.ZipCode,
		Province:                     c.Province,
		City:                         c.City,
		Address:                      c.Address,
		Address2:                     c.Address2,
		RepresentativeFamilyName:     c.RepresentativeFamilyName,
		RepresentativeFirstName:      c.RepresentativeFirstName,
		RepresentativeFamilyNameKana: c.RepresentativeFamilyNameKana,
		RepresentativeFirstNameKana:  c.RepresentativeFirstNameKana,
		ImageUrl:                     c.ImageUrl,
		Created:                      c.Created,
		Updated:                      c.Updated,
	}
}

func (receiver *CapitalRaisingEntityInput) ToModel() (*domain.CapitalRaisingEntity, error) {
	return &domain.CapitalRaisingEntity{
		CorporateName:                receiver.CorporateName,
		ZipCode:                      receiver.ZipCode,
		Province:                     receiver.Province,
		City:                         receiver.City,
		Address:                      receiver.Address,
		Address2:                     receiver.Address2,
		RepresentativeFamilyName:     receiver.RepresentativeFamilyName,
		RepresentativeFirstName:      receiver.RepresentativeFirstName,
		RepresentativeFamilyNameKana: receiver.RepresentativeFamilyNameKana,
		RepresentativeFirstNameKana:  receiver.RepresentativeFirstNameKana,
	}, nil
}
