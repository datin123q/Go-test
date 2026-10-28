package jwthelper

import (
	"path/filepath"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/pathutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/skipurls"
)

var (
	once        sync.Once
	jwtSkipUrls map[string][]string
)

const (
	jwtUserKey  = "user"
	jwtEmailKey = "email"
	jwtNameKey  = "name"
)

type User struct {
	Id    int64  `query:"-" json:"-"`
	Email string `query:"-" json:"-"`
	Name  string `query:"-" json:"-"`
	Roles map[string]int
}

func (j User) HasRole(role domain.ROLE_TYPE) bool {
	_, ok := j.Roles[string(role)]

	return ok
}

func (j User) IsEmpty() bool {
	return reflect.DeepEqual(j, User{})
}

type userExtractor struct{}

type UserExtractor interface {
	Extract(c echo.Context) (User, error)
}

func NewUserExtractor() UserExtractor {
	return &userExtractor{}
}

func (j userExtractor) Extract(c echo.Context) (User, error) {
	user, ok := c.Get(jwtUserKey).(*jwt.Token)
	if !ok {
		return User{}, errors.New("invalid token")
	}

	claims, ok := user.Claims.(jwt.MapClaims)
	if !ok {
		return User{}, errors.New("invalid claims")
	}

	userIdStr, ok := claims[jwtUserKey].(string)
	if !ok {
		return User{}, errors.New("user ID missing or invalid in token")
	}
	id, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		return User{}, errors.New("user ID conversion error")
	}

	email, ok := claims[jwtEmailKey].(string)
	if !ok {
		return User{}, errors.New("no email in token")
	}

	name, ok := claims[jwtNameKey].(string)
	if !ok {
		return User{}, errors.New("no name in token")
	}

	roles, _ := TokenToRoleMap(c)

	return User{
		Id:    id,
		Email: email,
		Name:  name,
		Roles: roles,
	}, nil
}

func TokenToUser(c echo.Context) (int64, string, error) {
	user, ok := c.Get("user").(*jwt.Token)
	if ok {
		claims := user.Claims.(jwt.MapClaims)
		id, err := strconv.ParseInt(claims["user"].(string), 10, 64)
		if err != nil {
			return 0, "", factory.ErrorTokenInvalid.New(err)
		}

		email, ok := claims["email"].(string)
		if !ok {
			return 0, "", factory.ErrorTokenInvalid.New(errors.New("no email in token"))
		}

		return id, email, nil
	} else {
		return 0, "", factory.ErrorTokenInvalid.New(errors.New("invalid token"))
	}
}

func TokenToUserExtended(c echo.Context) (int64, string, string, error) {
	user, ok := c.Get("user").(*jwt.Token)
	if ok {
		claims := user.Claims.(jwt.MapClaims)
		id, err := strconv.ParseInt(claims["user"].(string), 10, 64)
		if err != nil {
			return 0, "", "", factory.ErrorTokenInvalid.New(err)
		}

		name, ok := claims["name"].(string)
		if !ok {
			return 0, "", "", factory.ErrorTokenInvalid.New(errors.New("no name in token"))
		}

		email, ok := claims["email"].(string)
		if !ok {
			return 0, "", "", factory.ErrorTokenInvalid.New(errors.New("no email in token"))
		}

		return id, name, email, nil
	} else {
		return 0, "", "", factory.ErrorTokenInvalid.New(errors.New("invalid token"))
	}
}

func GetNameFromToken(c echo.Context) (string, error) {
	user, ok := c.Get("user").(*jwt.Token)
	if ok {
		claims := user.Claims.(jwt.MapClaims)
		name, ok := claims["name"].(string)
		if !ok {
			return "", factory.ErrorTokenInvalid.New(errors.New("no name in token"))
		}
		return name, nil
	} else {
		return "", factory.ErrorTokenInvalid.New(errors.New("invalid token"))
	}
}

// トークンとRequestユーザーIDを検証
func ValidateUserToken(c echo.Context, userId int64) error {
	user, ok := c.Get("user").(*jwt.Token)
	if ok {
		claims := user.Claims.(jwt.MapClaims)
		id, err := strconv.ParseInt(claims["user"].(string), 10, 64)
		if err != nil {
			return factory.ErrorTokenInvalid.New(err)
		}
		if id != userId {
			return factory.ErrorNotHaveRole.New(errors.New("not have role"))
		}
		return nil
	} else {
		return factory.ErrorNotHaveRole.New(errors.New("invalid token"))
	}
}

func GetRoleFromToken(c echo.Context) ([]string, error) {
	// 一時的な使用（今後クリーンアーキテクチャで全体リファクタリング時に削除）

	user, ok := c.Get("user").(*jwt.Token)
	if !ok {
		return nil, errors.New("invalid token: user not found")
	}
	claims, ok := user.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token: claims not found")
	}

	object := reflect.ValueOf(claims["role"])
	var roles = make([]string, 0)
	for i := 0; i < object.Len(); i++ {
		roles = append(roles, object.Index(i).Interface().(string))
	}

	return roles, nil
}

func TokenToRoleMap(c echo.Context) (map[string]int, error) {
	user, ok := c.Get("user").(*jwt.Token)
	if !ok {
		return nil, errors.New("not find user info")
	}
	claims, ok := user.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("not find user info")
	}
	object := reflect.ValueOf(claims["role"])
	var roleMap = make(map[string]int)
	for i := 0; i < object.Len(); i++ {
		roleMap[object.Index(i).Interface().(string)] = 1
	}
	return roleMap, nil
}

func HasROLE(c echo.Context, r domain.ROLE_TYPE) bool {
	roleMap, err := TokenToRoleMap(c)
	if err != nil {
		return false
	}
	_, ok := roleMap[string(r)]
	return ok
}

func CreateAdminTokenForTest(id int) (*jwt.Token, string) {
	tokenMap := map[string]interface{}{
		"iss":   domain.SERVICE_NAME,
		"aud":   domain.SERVICE_NAME,
		"user":  strconv.Itoa(int(id)),
		"email": "admin@energyshares.jp",
		"name":  "admin",
		"role":  []string{"ROLE_ADMIN", "ROLE_PRODUCT", "ROLE_OPERATOR", "ROLE_USER", "ROLE_PAYMENT", "ROLE_INVESTOR", "ROLE_MESSENGER", "ROLE_REPORT", "ROLE_SUPER_ADMIN"},
	}

	jwtSigningMethod := jwt.SigningMethodHS256
	jwtSecret := "JWT_SECRET"
	expDuration := time.Hour * 24

	claims := jwt.MapClaims{
		"nbf": time.Now().Unix(),
		"exp": time.Now().Add(expDuration).Unix(),
	}
	for k, v := range tokenMap {
		claims[k] = v
	}
	token := jwt.NewWithClaims(jwtSigningMethod, claims)
	secret, _ := token.SignedString([]byte(jwtSecret))

	return token, secret
}

func CreateUserTokenForTest(id int, email string) (*jwt.Token, string) {
	tokenMap := map[string]interface{}{
		"iss":   domain.SERVICE_NAME,
		"aud":   domain.SERVICE_NAME,
		"user":  strconv.Itoa(int(id)),
		"email": email,
		"name":  "user",
		"role":  []string{"ROLE_USER"},
	}

	jwtSigningMethod := jwt.SigningMethodHS256
	jwtSecret := "JWT_SECRET"
	expDuration := time.Hour * 24

	claims := jwt.MapClaims{
		"nbf": time.Now().Unix(),
		"exp": time.Now().Add(expDuration).Unix(),
	}
	for k, v := range tokenMap {
		claims[k] = v
	}
	token := jwt.NewWithClaims(jwtSigningMethod, claims)
	secret, _ := token.SignedString([]byte(jwtSecret))

	return token, secret
}

func JwtErrorHandler() func(c echo.Context, err error) error {
	return func(c echo.Context, err error) error {
		return factory.ErrorTokenExpired.New(err)
	}
}

func JwtSkipper(path string) func(c echo.Context) bool {
	return func(c echo.Context) bool {
		// 最初の1回のみ実行され、その後はグローバル変数を参照するためJwtSkipUrlsFromYaml()関数を繰り返し呼び出さない。
		once.Do(func() {
			var err error
			assetsPath := pathutils.GetPathOfAssets()
			jwtSkipUrls, err = skipurls.JwtSkipUrlsFromYaml(filepath.Join(assetsPath, path), "jwt_skip.yml")
			if err != nil {
				// ログを記録してサーバーを終了します。
				factory.Logger(c.Request().Context()).Fatal(errors.Wrap(err, "fail to get user id for check permission"))
			}
		})

		methodList, ok := jwtSkipUrls[c.Path()]
		if !ok { // (do not skip) 一致するAPI URLなし
			return false
		}

		for _, m := range methodList {
			if len(m) == 0 {
				return true // (skip) 一致するAPI URLがあり全メソッド許可
			}
			if m == c.Request().Method {
				return true // (skip) 一致するAPI URLがあり特定メソッド許可
			}
		}

		// (do not skip) 一致するAPI URLがあるがメソッド不一致
		return false
	}
}
