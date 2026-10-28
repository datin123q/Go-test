package controllers

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"xorm.io/xorm"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/labstack/echo/v4"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

type ApiResult struct {
	Result  interface{}   `json:"result"`
	Success bool          `json:"success"`
	Error   factory.Error `json:"error"`
}

type ArrayResult struct {
	Items      interface{} `json:"items"`
	TotalCount int64       `json:"total_count"`
}

func tokenToUser(c echo.Context) (int64, string, error) {
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
		return 0, "", factory.ErrorTokenInvalid.New(errors.New("no email in token"))
	}
}

func mustTokenUserId(c echo.Context) int64 {
	uid, _, _ := tokenToUser(c) // must関数は特殊な目的を持ち、エラーとメール値を無視してユーザーIDのみを返します。
	return uid
}

func renderSucc(c echo.Context, status int, result interface{}) error {
	req := c.Request()
	if req.Method == "POST" || req.Method == "PUT" || req.Method == "DELETE" {
		if session, ok := factory.DB(req.Context()).(*xorm.Session); ok {
			err := session.Commit()
			if err != nil {
				return factory.ErrorDB.New(err)
			}
		}
	}

	return c.JSON(status, ApiResult{
		Success: true,
		Result:  result,
	})
}

func CreateUserTokenForTest(id int, email string) (*jwt.Token, string) {
	tokenMap := map[string]interface{}{
		"iss":   domain.SERVICE_NAME,
		"aud":   domain.SERVICE_NAME,
		"user":  strconv.Itoa(int(id)),
		"email": email,
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

func jwtFromHeaderMiddleware(jwtSecret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Skipper ロジックを使用して特定のリクエストをスキップできます
			// 例えば、JWT ミドルウェアに既に適用されているリクエストはここで処理しません

			authHeader := c.Request().Header.Get(echo.HeaderAuthorization)
			if authHeader == "" {
				return next(c)
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				return next(c)
			}

			// ヘッダーからJWTトークンを抽出およびパースします
			token, err := jwt.ParseWithClaims(tokenString, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
				return []byte(jwtSecret), nil
			})

			if err != nil || !token.Valid {
				return next(c)
			}

			c.Set("user", token)

			return next(c)
		}
	}
}
