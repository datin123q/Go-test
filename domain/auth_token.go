package domain

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/modelhelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/tool"

	"github.com/hillfolk/goutils/jwtutil"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

const SERVICE_NAME = "energyshares"

const (
	AccessTokenExpireTime        = 5 * time.Minute
	RefreshTokenExpireTime       = 24 * time.Hour
	RefreshTokenCookieExpireTime = 1 * time.Hour

	TestAccessTokenExpireTime        = 20 * time.Minute
	TestRefreshTokenExpireTime       = 4 * time.Hour
	TestRefreshTokenCookieExpireTime = 4 * time.Hour
)

type ResultToken struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type AuthToken struct {
	Id       int64              `json:"id"`
	UserId   int64              `json:"user_id" xorm:"INDEX"`
	Token    string             `json:"token" xorm:"UNIQUE varchar(500)"`
	IsActive bool               `json:"is_active"`
	ExpireIn int64              `json:"expire_in"`
	Name     string             `json:"name" xorm:"-"`
	Email    string             `json:"email" xorm:"-"`
	Roles    []string           `json:"roles" xorm:"-"`
	Created  jsonutils.JSONTime `json:"created" xorm:"created"`
	Updated  jsonutils.JSONTime `json:"updated" xorm:"updated"`
	Deleted  jsonutils.JSONTime `json:"deleted" xorm:"deleted"`
}

type RefreshToken struct {
	Id       int64              `json:"id"`
	UserId   int64              `json:"user_id" xorm:"INDEX"`
	UserType USER_TYPE          `json:"user_type"`
	Token    string             `json:"token" xorm:"UNIQUE varchar(500)"`
	IsActive bool               `json:"is_active" xorm:"is_active"`
	ExpireIn int64              `json:"expire_in" xorm:"expire_in"`
	IpAddr   string             `json:"ip_addr" xorm:"ip_addr"`
	Agent    string             `json:"agent" xorm:"agent"`
	Name     string             `json:"name" xorm:"-"`
	Email    string             `json:"email" xorm:"-"`
	Created  jsonutils.JSONTime `json:"created" xorm:"created"`
	Updated  jsonutils.JSONTime `json:"updated" xorm:"updated"`
	Deleted  jsonutils.JSONTime `json:"deleted" xorm:"deleted"`
}

func (r RefreshToken) TableName() string {
	return "refresh_token"
}

func (r RefreshToken) SelectAll() string {
	return r.TableName() + ".*"
}

// 토큰 생성
func (AuthToken) Create(ctx context.Context, u User) (*AuthToken, error) {
	var v int64
	t, err := CreateToken(ctx, u)
	if err != nil {
		return nil, err
	}
	c, err := jwtutil.Extract(t)
	if err != nil {
		return nil, err
	}
	exp, ok := c["exp"]
	if !ok {
		return nil, errors.New("Fail Case exp")
	}
	v = int64(exp.(float64))
	token := &AuthToken{
		UserId:   u.Id,
		Token:    t,
		Name:     u.Name,
		IsActive: true,
		Email:    u.Email,
		Roles:    u.Roles,
		ExpireIn: v,
	}
	// AccessToken DB저장 X
	// _, err = factory.DB(ctx).Insert(token)
	// if err != nil {
	//	return nil, err
	// }
	return token, nil
}

func (AuthToken) CreateAdmin(ctx context.Context, u AdminUser) (*AuthToken, error) {
	var v int64
	t, err := CreateAdminToken(ctx, u)
	if err != nil {
		return nil, err
	}
	c, err := jwtutil.Extract(t)
	if err != nil {
		return nil, err
	}
	exp, ok := c["exp"]
	if !ok {
		return nil, errors.New("Fail Case exp")
	}
	v = int64(exp.(float64))
	token := &AuthToken{
		UserId:   u.Id,
		Token:    t,
		Name:     u.Name,
		IsActive: true,
		Email:    u.Email,
		Roles:    u.Roles,
		ExpireIn: v,
	}
	// AccessToken DB저장 X
	// _, err = factory.DB(ctx).Insert(token)
	// if err != nil {
	//	return nil, err
	// }
	return token, nil
}

func CreateToken(ctx context.Context, u User) (string, error) {
	tokenMap := map[string]interface{}{
		"iss":   SERVICE_NAME,
		"aud":   SERVICE_NAME,
		"user":  strconv.Itoa(int(u.Id)),
		"role":  u.Roles,
		"email": u.Email,
		// 밑 3가지 항목 추가
		"name":         u.Name,
		"audit_status": u.AuditStatus,
		"user_type":    u.UserType,
	}
	if serviceutils.GetEnv() == "DEV" {
		jwtutil.SetExpDuration(TestAccessTokenExpireTime)
	} else {
		jwtutil.SetExpDuration(AccessTokenExpireTime)
	}
	return jwtutil.NewToken(tokenMap)

}

// CreateRefreshToken 리스플레쉬 토큰 추가
func CreateRefreshToken(ctx context.Context, u User) (string, error) {
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
	if serviceutils.GetEnv() == "DEV" {
		jwtutil.SetExpDuration(TestRefreshTokenExpireTime)
	} else {
		jwtutil.SetExpDuration(RefreshTokenExpireTime)
	}
	return jwtutil.NewToken(tokenMap)

}

func (RefreshToken) Create(ctx context.Context, u User, ipAddr, agent string) (*RefreshToken, error) {
	var v int64
	t, err := CreateRefreshToken(ctx, u)
	if err != nil {
		return nil, err
	}
	c, err := jwtutil.Extract(t)
	if err != nil {
		return nil, err
	}
	exp, ok := c["exp"]
	if !ok {
		return nil, errors.New("Fail Case exp")
	}
	v = int64(exp.(float64))

	token := &RefreshToken{
		UserId:   u.Id,
		UserType: u.UserType,
		Token:    t,
		IsActive: true,
		Email:    u.Email,
		ExpireIn: v,
		IpAddr:   ipAddr,
		Agent:    agent,
	}
	_, err = factory.DB(ctx).Insert(token)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (RefreshToken) GetAll(ctx context.Context, sortby, order string, offset, limit int) (int64, []RefreshToken, error) {
	q := factory.DB(ctx)
	if err := modelhelper.SetSortOrder(q, sortby, order); err != nil {
		factory.Logger(ctx).Error(err)
	}

	var items []RefreshToken
	totalCount, err := q.Limit(limit, offset).FindAndCount(&items)
	if err != nil {
		return 0, nil, err
	}
	return totalCount, items, nil
}

func (t RefreshToken) Get(ctx context.Context, token string) (*RefreshToken, error) {
	var res RefreshToken
	if has, err := factory.DB(ctx).Table("refresh_token").Where("token=?", token).Get(&res); err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}
	return &res, nil
}

func (t RefreshToken) GetByClientInfo(ctx context.Context, token, ipAddr, agent string) (*RefreshToken, error) {
	var res RefreshToken
	if has, err := factory.DB(ctx).Table("refresh_token").Where("token=?", token).And("ip_addr = ?", ipAddr).And("agent = ?", agent).Get(&res); err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}
	return &res, nil
}

func (t *RefreshToken) UpdateExpireIn(ctx context.Context, u User) error {
	var v int64
	token, err := CreateRefreshToken(ctx, u)
	if err != nil {
		return err
	}
	c, err := jwtutil.Extract(token)
	if err != nil {
		return err
	}
	exp, ok := c["exp"]
	if !ok {
		return errors.New("Fail Case exp")
	}
	v = int64(exp.(float64))

	t.Token = token
	t.ExpireIn = v
	_, err = factory.DB(ctx).ID(t.Id).Update(t)

	return err
}

func (t *RefreshToken) Update(ctx context.Context) (int64, error) {
	return factory.DB(ctx).ID(t.Id).Update(t)
}
