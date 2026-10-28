package auth

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/hillfolk/goutils/jwtutil"
	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/tool"
)

const (
	SERVICE_NAME = "energyshares"
	// ユーザー用のトークンの有効期限
	USER_REFRESH_TOKEN_EXPIRE_TIME = time.Hour * 1
	USER_AUTH_TOKEN_EXPIRE_TIME    = time.Minute * 5
	// 管理用のトークンの有効期限
	ADMIN_REFRESH_TOKEN_EXPIRE_TIME = time.Hour * 24
	ADMIN_AUTH_TOKEN_EXPIRE_TIME    = time.Minute * 5
	// テスト用のトークンの有効期限
	TEST_REFRESH_TOKEN_EXPIRE_TIME = time.Hour * 4
	TEST_AUTH_TOKEN_EXPIRE_TIME    = time.Minute * 20
)

type authService struct {
	logger                 factory.LoggerFactory
	cfg                    *config.Config
	refreshTokenRepository cdomain.RefreshTokenRepository
}

func New(
	logger factory.LoggerFactory,
	cfg *config.Config,
	refreshTokenRepository cdomain.RefreshTokenRepository,
) cdomain.AuthService {
	return &authService{
		logger:                 logger,
		cfg:                    cfg,
		refreshTokenRepository: refreshTokenRepository,
	}
}

func (a authService) CreateUserAccessToken(ctx context.Context, user domain.User) (string, error) {
	user.Roles = []string{string(domain.ROLE_USER)}

	tokenMap := map[string]interface{}{
		"iss":          SERVICE_NAME,
		"aud":          SERVICE_NAME,
		"user":         strconv.Itoa(int(user.Id)),
		"role":         user.Roles,
		"email":        user.Email,
		"name":         user.Name,
		"audit_status": user.AuditStatus,
		"user_type":    user.UserType,
	}
	if strings.ToUpper(a.cfg.Env) == "DEV" {
		jwtutil.SetExpDuration(TEST_AUTH_TOKEN_EXPIRE_TIME)
	} else {
		jwtutil.SetExpDuration(USER_AUTH_TOKEN_EXPIRE_TIME)
	}

	token, err := jwtutil.NewToken(tokenMap)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (a authService) CreateUserRefreshToken(ctx context.Context, input cdomain.UserRefreshTokenInput) (string, error) {
	// create refresh token
	jti, err := tool.RandomString(16)
	if err != nil {
		return "", err
	}

	tokenMap := map[string]interface{}{
		"iss":   SERVICE_NAME,
		"aud":   SERVICE_NAME,
		"user":  strconv.Itoa(int(input.User.Id)),
		"email": input.User.Email,
		"iat":   time.Now().Unix(),
		"jti":   jti,
	}
	if strings.ToUpper(a.cfg.Env) == "DEV" {
		jwtutil.SetExpDuration(TEST_REFRESH_TOKEN_EXPIRE_TIME)
	} else {
		jwtutil.SetExpDuration(USER_REFRESH_TOKEN_EXPIRE_TIME)
	}

	token, err := jwtutil.NewToken(tokenMap)
	if err != nil {
		return "", err
	}

	// store refresh token
	claimInfo, err := jwtutil.Extract(token)
	if err != nil {
		return "", err
	}
	exp, ok := claimInfo["exp"]
	if !ok {
		return "", errors.New("token does not contain an 'exp' claim")
	}

	if _, err := a.refreshTokenRepository.Create(ctx, domain.RefreshToken{
		UserId:   input.User.Id,
		UserType: input.User.UserType,
		Token:    token,
		IsActive: true,
		Email:    input.User.Email,
		ExpireIn: int64(exp.(float64)),
		IpAddr:   input.IP,
		Agent:    input.Agent,
	}); err != nil {
		return "", err
	}

	return token, nil
}

func (a authService) CreateAdminAccessToken(ctx context.Context, admin domain.AdminUser) (string, error) {
	tokenMap := map[string]interface{}{
		"iss":   SERVICE_NAME,
		"aud":   SERVICE_NAME,
		"user":  strconv.Itoa(int(admin.Id)),
		"role":  admin.Roles,
		"email": admin.Email,
		"name":  admin.Name,
	}

	jwtutil.SetExpDuration(ADMIN_AUTH_TOKEN_EXPIRE_TIME)
	token, err := jwtutil.NewToken(tokenMap)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (a authService) CreateAdminRefreshToken(ctx context.Context, input cdomain.AdminRefreshTokenInput) (string, error) {
	// create refresh token
	jti, err := tool.RandomString(16)
	if err != nil {
		return "", err
	}

	tokenMap := map[string]interface{}{
		"iss":   SERVICE_NAME,
		"aud":   SERVICE_NAME,
		"user":  strconv.Itoa(int(input.Admin.Id)),
		"email": input.Admin.Email,
		"iat":   time.Now().Unix(),
		"jti":   jti,
	}

	jwtutil.SetExpDuration(ADMIN_REFRESH_TOKEN_EXPIRE_TIME)
	token, err := jwtutil.NewToken(tokenMap)
	if err != nil {
		return "", err
	}

	// store refresh token
	claimInfo, err := jwtutil.Extract(token)
	if err != nil {
		return "", err
	}
	exp, ok := claimInfo["exp"]
	if !ok {
		return "", errors.New("token does not contain an 'exp' claim")
	}

	if _, err := a.refreshTokenRepository.Create(ctx, domain.RefreshToken{
		UserId:   input.Admin.Id,
		UserType: domain.ADMIN,
		Token:    token,
		IsActive: true,
		Email:    input.Admin.Email,
		ExpireIn: int64(exp.(float64)),
		IpAddr:   input.IP,
		Agent:    input.Agent,
	}); err != nil {
		return "", err
	}

	return token, nil
}
