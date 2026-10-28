package domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/modelhelper"
)

const CAPITAL_RAISING_ENTITY = "capital_raising_entity"

// CapitalRaisingEntity 資金調達者
type CapitalRaisingEntity struct {
	Id                           int64              `json:"id" swagger:"desc(ID)"`
	CorporateName                string             `json:"corporate_name" xorm:"NOT NULL" swagger:"desc(法人名)"`
	ZipCode                      string             `json:"zip_code" xorm:"NOT NULL" swagger:"desc(郵便番号)"`
	Province                     string             `json:"province" xorm:"NOT NULL" swagger:"desc(都道府県)"`
	City                         string             `json:"city" xorm:"NOT NULL" swagger:"desc(市区町村)"`
	Address                      string             `json:"address" xorm:"NOT NULL" swagger:"desc(住所)"`
	Address2                     string             `json:"address2" xorm:"NOT NULL" swagger:"desc(住所補足)"`
	RepresentativeFamilyName     string             `json:"representative_family_name" xorm:"NOT NULL" swagger:"desc(代表者 姓)"`
	RepresentativeFirstName      string             `json:"representative_first_name" xorm:"NOT NULL" swagger:"desc(代表者 名)"`
	RepresentativeFamilyNameKana string             `json:"representative_family_name_kana" xorm:"NOT NULL" swagger:"desc(代表者 姓（カナ）)"`
	RepresentativeFirstNameKana  string             `json:"representative_first_name_kana" xorm:"NOT NULL" swagger:"desc(代表者 名（カナ）)"`
	ImageUrl                     *string            `json:"image_url" swagger:"desc(画像URL)"`
	Created                      jsonutils.JSONTime `json:"created" xorm:"created"`
	Updated                      jsonutils.JSONTime `json:"updated" xorm:"updated"`
	Deleted                      jsonutils.JSONTime `json:"deleted" xorm:"deleted"`
}

func (CapitalRaisingEntity) TableName() string {
	return CAPITAL_RAISING_ENTITY
}

func (c CapitalRaisingEntity) SelectAll() string {
	return c.TableName() + ".*"
}

func (c *CapitalRaisingEntity) Compare(t *CapitalRaisingEntity) bool {
	if c.Id != t.Id {
		return false
	}
	if c.CorporateName != t.CorporateName {
		return false
	}
	if c.ZipCode != t.ZipCode {
		return false
	}
	if c.Province != t.Province {
		return false
	}
	if c.City != t.City {
		return false
	}
	if c.Address != t.Address {
		return false
	}
	if c.Address2 != t.Address2 {
		return false
	}
	if c.RepresentativeFamilyName != t.RepresentativeFamilyName {
		return false
	}
	if c.RepresentativeFirstName != t.RepresentativeFirstName {
		return false
	}
	if c.RepresentativeFamilyNameKana != t.RepresentativeFamilyNameKana {
		return false
	}
	if c.RepresentativeFirstNameKana != t.RepresentativeFirstNameKana {
		return false
	}
	if c.ImageUrl != t.ImageUrl {
		return false
	}
	return true
}

func (c *CapitalRaisingEntity) Create(ctx context.Context) (int64, error) {
	return factory.DB(ctx).Insert(c)
}

func (e *CapitalRaisingEntity) Update(ctx context.Context) error {
	query := factory.DB(ctx).Nullable("image_url")
	_, err := query.ID(e.Id).Update(e)
	return err
}

func (CapitalRaisingEntity) GetById(ctx context.Context, id int64) (*CapitalRaisingEntity, error) {
	var v CapitalRaisingEntity
	if has, err := factory.DB(ctx).ID(id).Get(&v); err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}
	return &v, nil
}

func (CapitalRaisingEntity) GetAll(
	ctx context.Context,
	sortby, order string,
	offset, limit int,
) (int64, []CapitalRaisingEntity, error) {

	q := factory.DB(ctx)

	var (
		totalCount int64
		items      []CapitalRaisingEntity
	)

	// 対象テーブル
	q.Table("capital_raising_entity")

	// 並び順
	if err := modelhelper.SetSortOrder(q, sortby, order); err != nil {
		factory.Logger(ctx).Error(err)
	}

	// 論理削除除外（deleted IS NULL）
	q.Where("capital_raising_entity.deleted IS NULL")

	// LIMIT / OFFSET + COUNT
	totalCount, err := q.
		Limit(limit, offset).
		FindAndCount(&items)
	if err != nil {
		return 0, nil, err
	}

	return totalCount, items, nil
}

func (CapitalRaisingEntity) Delete(ctx context.Context, id int64) error {
	_, err := factory.DB(ctx).ID(id).Delete(&CapitalRaisingEntity{})
	return err
}
