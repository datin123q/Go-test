package domain

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strconv"
	"time"

	"github.com/gofrs/uuid"
	"github.com/hillfolk/goutils/jwtutil"
	"golang.org/x/crypto/pbkdf2"
	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/tool"
)

type AdminUser struct {
	Id                int64              `json:"id" xorm:"pk autoincr"`
	CommonId          string             `json:"common_id" xorm:"common_id"`
	Email             string             `json:"email" xorm:"email unique"`
	Name              string             `json:"name" xorm:"name varchar(100)"`
	Password          string             `json:"-" xorm:"password ->"`
	Salt              string             `json:"-" xorm:"salt ->"`
	Roles             []string           `json:"roles" xorm:"-"`
	MobilePhone       string             `json:"mobile_phone" xorm:"mobile_phone varchar(20)"`
	LastLoginDate     jsonutils.JSONTime `json:"last_login_date" xorm:"last_login_date"`
	LoginFailCnt      int                `json:"login_fail_cnt" xorm:"login_fail_cnt"`
	LastLoginIp       string             `json:"last_login_ip" xorm:"last_login_ip varchar(50)"`
	LastLoginAgent    string             `json:"last_login_agent" xorm:"last_login_agent varchar(255)"`
	RolesString       string             `json:"-" xorm:"<-"`
	IsSuperAdmin      bool               `json:"is_super_admin" xorm:"-"`
	RoleUpdated       jsonutils.JSONTime `json:"role_updated" xorm:"role_updated"`
	ValidityStartDate jsonutils.JSONTime `json:"validity_start_date" xorm:"validity_start_date"`
	ValidityEndDate   jsonutils.JSONTime `json:"validity_end_date" xorm:"validity_end_date"`
	Updated           jsonutils.JSONTime `json:"-" xorm:"updated"`
	Created           jsonutils.JSONTime `json:"-" xorm:"created"`
	Deleted           jsonutils.JSONTime `json:"-" xorm:"deleted"`
}

func (u AdminUser) GetById(ctx context.Context, id int64) (*AdminUser, error) {
	var user AdminUser
	if has, err := factory.DB(ctx).Table(u.TableName()).Select(AdminUser{}.SelectAll()).ID(id).Get(&user); err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}
	user.Roles, _ = UserRoles{}.GetRoles(ctx, user.Id)
	user.IsSuperAdmin = tool.StrSliceContains(user.Roles, string(ROLE_SUPER_ADMIN))
	return &user, nil
}

func (AdminUser) TableName() string {
	return "admin_user"
}

func (a AdminUser) SelectAll() string {
	return a.TableName() + ".*"
}
func (u AdminUser) GetSuperAdmin(ctx context.Context) (*AdminUser, error) {

	query := factory.DB(ctx).Table(AdminUser{}.TableName()).Select(AdminUser{}.SelectAll())

	var adminUsers []AdminUser
	// No 에 로우 카운터 적용 필요
	err := query.Find(&adminUsers)
	if err != nil {
		return nil, err
	}

	var superAdmin AdminUser
	// 사용자 권한
	for _, user := range adminUsers {
		user.Roles, _ = UserRoles{}.GetRoles(ctx, user.Id)
		user.IsSuperAdmin = tool.StrSliceContains(user.Roles, string(ROLE_SUPER_ADMIN))

		if user.IsSuperAdmin {
			superAdmin = user
			break
		}
	}

	return &superAdmin, nil
}

func (u AdminUser) CreateAdmin(engine *xorm.Engine) error {

	var uid uuid.UUID
	var err error
	if uid, err = uuid.NewV1(); err != nil {
		return err
	}

	u.CommonId = uuid.NewV5(uid, u.Email).String()

	if u.Salt, err = GetUserSalt(); err != nil {
		return err
	}
	u.EncodePasswd()

	if _, err = engine.Table(AdminUser{}.TableName()).Insert(&u); err != nil {
		return err
	}

	// 사용자 권한 부여
	codes := []string{"ROLE_ADMIN", "ROLE_OPERATOR", "ROLE_PRODUCT", "ROLE_PAYMENT", "ROLE_MESSENGER", "ROLE_INVESTOR", "ROLE_BASE", "ROLE_REPORT", "ROLE_SUPER_ADMIN"}
	for _, v := range codes {
		role := &Role{
			UserId:   u.Id,
			RoleCode: v,
		}

		if _, err = engine.Table("role").Insert(role); err != nil {
			return err
		}
	}
	return nil
}

// EncodePasswd encodes password to safe format.
func (u *AdminUser) EncodePasswd() {
	newPasswd := pbkdf2.Key([]byte(u.Password), []byte(u.Salt), 10000, 50, sha256.New)
	u.Password = fmt.Sprintf("%x", newPasswd)
}

func (AdminUser) GetSuperAdminId(ctx context.Context) (int64, error) {

	query := factory.DB(ctx).Table(AdminUser{}.TableName()).Select(AdminUser{}.SelectAll())
	var userList []AdminUser
	// No 에 로우 카운터 적용 필요
	err := query.Find(&userList)
	if err != nil {
		return 0, err
	}

	var superAdminId int64
	// 사용자 권한
	for _, user := range userList {
		user.Roles, _ = UserRoles{}.GetRoles(ctx, user.Id)
		user.IsSuperAdmin = tool.StrSliceContains(user.Roles, string(ROLE_SUPER_ADMIN))

		if user.IsSuperAdmin {
			superAdminId = user.Id
			break
		}
	}

	return superAdminId, nil
}

func CreateAdminToken(ctx context.Context, u AdminUser) (string, error) {
	tokenMap := map[string]interface{}{
		"iss":   SERVICE_NAME,
		"aud":   SERVICE_NAME,
		"user":  strconv.Itoa(int(u.Id)),
		"role":  u.Roles,
		"email": u.Email,
		"name":  u.Name,
	}
	jwtutil.SetExpDuration(AccessTokenExpireTime)
	return jwtutil.NewToken(tokenMap)
}

// CreateRefreshToken 리스플레쉬 토큰 추가
func CreateAdminRefreshToken(ctx context.Context, u AdminUser) (string, error) {
	jti, err := tool.RandomString(16)
	if err != nil {
		return "", err
	}

	tokenMap := map[string]interface{}{
		"iss":   SERVICE_NAME,
		"aud":   SERVICE_NAME,
		"user":  strconv.Itoa(int(u.Id)),
		"email": u.Email,
		"iat":   time.Now().Unix(),
		"jti":   jti,
	}
	jwtutil.SetExpDuration(RefreshTokenExpireTime)
	return jwtutil.NewToken(tokenMap)
}
