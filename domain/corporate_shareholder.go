package domain

import (
	"context"
	"fmt"
	"strconv"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

const MAX_SHAREHOLDER = 1000

type CorporateShareholder struct {
	Id             int64              `json:"-"`
	UserId         int64              `json:"user_id" xorm:"not null"`
	Uid            string             `json:"uid" swagger:"desc(주주 유니크 ID)"`
	FamilyName     string             `json:"family_name" xorm:"not null" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_NAME" swagger:"desc(주주 명 성)"`
	FirstName      string             `json:"first_name" xorm:"not null" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_FAMILY_NAME" swagger:"desc(주주 명 이름)"`
	FamilyNameKana string             `json:"family_name_kana" xorm:"not null" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_NAME2" swagger:"desc(주주 명 성 후리가나)"`
	FirstNameKana  string             `json:"first_name_kana" xorm:"not null" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_FAMILY_NAME2" swagger:"desc(주주 명 이름 후리가나)"`
	Birthday       jsonutils.JSONTime `json:"birthday" xorm:"not null" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_BIRTHDAY" swagger:"desc(생년월일)"`
	ZipCode        string             `json:"zip_code" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_ZIPCODE" swagger:"desc(우편번호)"`
	Address        string             `json:"address" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_ADDRESS" swagger:"desc(주소)"`
	Address2       string             `json:"address2" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_ADDRESS2" swagger:"desc(상세 주소)"`
	Province       string             `json:"province" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_PROVINCE" swagger:"desc(현 이름)"`
	City           string             `json:"city" i18n:"commoncode.USER_CHANGED_CONTENT_TYPE.USER_CHANGED_CONTENT_CITY" xorm:"not null" swagger:"desc(시 또는 구, 군)"`
	Created        jsonutils.JSONTime `json:"created" xorm:"created"`
	Updated        jsonutils.JSONTime `json:"updated" xorm:"updated"`
	Deleted        jsonutils.JSONTime `json:"deleted" xorm:"deleted"`
}

func (c CorporateShareholder) TableName() string {
	return "corporate_shareholder"
}

func (c *CorporateShareholder) SelectAll() string {
	return c.TableName() + ".*"
}

func (c *CorporateShareholder) Create(ctx context.Context) (int64, error) {

	var item CorporateShareholder
	has, err := factory.DB(ctx).Table("corporate_shareholder").
		Where("corporate_shareholder.user_id = ?", c.UserId).Desc("uid").Limit(1).Get(&item)
	if err != nil {
		return 0, err
	}

	var uid int
	if has {
		uid, err = strconv.Atoi(item.Uid)
		if err != nil {
			return 0, err
		}
	}

	uid++
	if uid < MAX_SHAREHOLDER {
		c.Uid = fmt.Sprintf("%03d", uid)
	} else {
		c.Uid = fmt.Sprintf("%d", uid)
	}

	return factory.DB(ctx).Insert(c)
}

func (CorporateShareholder) FindByUserId(ctx context.Context, limit int, id int64) ([]*CorporateShareholder, error) {
	q := factory.DB(ctx)

	items := make([]*CorporateShareholder, 0)
	_, err := q.Where("user_id = ? ", id).Limit(limit).FindAndCount(&items)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (c *CorporateShareholder) Update(ctx context.Context) (err error) {
	if _, err := factory.DB(ctx).UseBool("address", "address2").ID(c.Id).Update(c); err != nil {
		return err
	}

	return err
}

func (CorporateShareholder) DeleteByUid(ctx context.Context, uid string) (err error) {
	_, err = factory.DB(ctx).Where("corporate_shareholder.uid = ?", uid).Delete(&CorporateShareholder{})
	return err
}
