package domain

import (
	"context"
	"encoding/json"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/commoncode"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

type ROLE_TYPE string

const (
	ROLE_TYPE_CODE             = "ROLE_TYPE"
	ROLE_USER        ROLE_TYPE = "ROLE_USER"
	ROLE_ADMIN       ROLE_TYPE = "ROLE_ADMIN"
	ROLE_OPERATOR    ROLE_TYPE = "ROLE_OPERATOR"
	ROLE_PRODUCT     ROLE_TYPE = "ROLE_PRODUCT"
	ROLE_PAYMENT     ROLE_TYPE = "ROLE_PAYMENT"
	ROLE_INVESTOR    ROLE_TYPE = "ROLE_INVESTOR"
	ROLE_MESSENGER   ROLE_TYPE = "ROLE_MESSENGER"
	ROLE_BASE        ROLE_TYPE = "ROLE_BASE"
	ROLE_REPORT      ROLE_TYPE = "ROLE_REPORT"
	ROLE_MYNUMBER    ROLE_TYPE = "ROLE_MYNUMBER"
	ROLE_SUPER_ADMIN ROLE_TYPE = "ROLE_SUPER_ADMIN"
)

func (UserRoles) GetRoleTypeCodeGroup(loc string) commoncode.CodeGroup {

	codes := make([]commoncode.Code, 0, 0)

	codes = append(codes, commoncode.Code{Code: string(ROLE_OPERATOR), Name: i18nutils.Translate("commoncode.ROLE_TYPE.ROLE_OPERATOR", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(ROLE_PRODUCT), Name: i18nutils.Translate("commoncode.ROLE_TYPE.ROLE_PRODUCT", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(ROLE_PAYMENT), Name: i18nutils.Translate("commoncode.ROLE_TYPE.ROLE_PAYMENT", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(ROLE_INVESTOR), Name: i18nutils.Translate("commoncode.ROLE_TYPE.ROLE_INVESTOR", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(ROLE_MESSENGER), Name: i18nutils.Translate("commoncode.ROLE_TYPE.ROLE_MESSENGER", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(ROLE_BASE), Name: i18nutils.Translate("commoncode.ROLE_TYPE.ROLE_BASE", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(ROLE_REPORT), Name: i18nutils.Translate("commoncode.ROLE_TYPE.ROLE_REPORT", loc), Order: len(codes) + 1})
	codes = append(codes, commoncode.Code{Code: string(ROLE_MYNUMBER), Name: i18nutils.Translate("commoncode.ROLE_TYPE.ROLE_MYNUMBER", loc), Order: len(codes) + 1})
	return commoncode.CodeGroup{
		GroupCode: ROLE_TYPE_CODE,
		Codes:     codes,
	}
}

type UserRoles struct {
	Id       int64              `json:"id"`
	RoleCode ROLE_TYPE          `json:"role_code" xorm:"varchar(100)"`
	RoleName ROLE_TYPE          `json:"role_name" xorm:"varchar(100)"`
	Created  jsonutils.JSONTime `json:"created" xorm:"created"`
	Updated  jsonutils.JSONTime `json:"updated" xorm:"updated"`
}

func (ur *UserRoles) MarshalJSON() ([]byte, error) {

	return json.Marshal(map[string]interface{}{
		"role_code": string(ur.RoleCode), // code
		"role_name": string(ur.RoleName), // name
	})
}

type Role struct {
	Id       int64              `json:"id" xorm:"pk autoincr"`
	UserId   int64              `json:"user_id"`
	RoleCode string             `json:"role_code" xorm:"varchar(100)"`
	Created  jsonutils.JSONTime `json:"created" xorm:"created"`
	Updated  jsonutils.JSONTime `json:"updated" xorm:"updated"`
}

func (r Role) TableName() string {
	return "role"
}

func (r Role) SelectAll() string {
	return r.TableName() + ".*"
}

func (UserRoles) GetAll(ctx context.Context) (int64, []UserRoles, error) {
	q := factory.DB(ctx)

	var items []UserRoles
	var totalCount int64
	totalCount, err := q.Where("role_code != 'ROLE_USER' AND role_code != 'ROLE_ADMIN' AND role_code != 'ROLE_SUPER_ADMIN' ").FindAndCount(&items)
	if err != nil {
		return 0, nil, err
	}

	return totalCount, items, nil
}

func (UserRoles) GetAllUserRoles(ctx context.Context) ([]UserRoles, error) {
	q := factory.DB(ctx)

	var items []UserRoles
	_, err := q.FindAndCount(&items)
	if err != nil {
		return nil, err
	}

	return items, nil
}

func (ur *UserRoles) Create(ctx context.Context) (int64, error) {
	return factory.DB(ctx).Insert(ur)
}

func (Role) Delete(ctx context.Context, id int64) error {
	_, err := factory.DB(ctx).ID(id).Delete(&Role{})
	return err
}

func (UserRoles) GetRoles(ctx context.Context, userId int64) ([]string, error) {
	q := factory.DB(ctx)

	var roleArr []Role
	_, err := q.Table("role").Where("user_id = ?", userId).FindAndCount(&roleArr)
	if err != nil {
		return make([]string, 0), err
	}

	var roles []string
	for _, role := range roleArr {
		roles = append(roles, role.RoleCode)
	}

	return roles, nil
}
