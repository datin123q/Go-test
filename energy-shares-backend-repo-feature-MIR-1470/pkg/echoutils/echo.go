package echoutils

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/hillfolk/goutils/behaviorlog"
	"github.com/hillfolk/goutils/ctxbase"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/eventbus"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
)

var skipper = func(c echo.Context) bool {
	// GET /api/version` パスへのリクエストはログに出力しない
	return c.Request().RequestURI == "/api/version"
}

func CustomErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	if err == nil {
		err = factory.ErrorSystem.New(nil)
	}

	var (
		apiError  *factory.Error
		httpError *echo.HTTPError
	)

	// errors.Asはerrors.Wrapのラップを何段挟んでいても内側の型を見つける。httpErrorが
	// typed-nilで見つかった場合、直後のWithError(err)がerr.Error()を呼んだ時点でpanicする
	// (echo.HTTPError.Error()はnilレシーバのフィールドを直接参照するため)。*factory.Error側は
	// Error()がnilセーフなのでここで先に弾く必要はなく、switch内のnilガードで十分。
	if errors.As(err, &httpError) && httpError == nil {
		renderAPIError(c, factory.ErrorSystem.New(nil))
		return
	}

	behaviorlog.FromCtx(c.Request().Context()).WithError(err)

	switch {
	case errors.As(err, &apiError):
		renderAPIError(c, apiError)
	case httpError != nil:
		c.JSON(httpError.Code, echo.HTTPError{
			Code:    httpError.Code,
			Message: err.Error(),
		})
	default:
		// errors.Wrap で包んだだけの素のエラーは *factory.Error にも *echo.HTTPError にも
		// 一致しない。以前はここに default が無く、レスポンスに何も書かれないまま関数を
		// 抜けてしまい、net/http の既定動作でステータス200・ボディ空のレスポンスが返って
		// いた(呼び出し元からは成功と区別が付かない)。500として明示的に返すことで、
		// 今後同じ書き方の抜け漏れが増えても機械的に安全側へ倒す(MIR-1133)。
		renderAPIError(c, factory.ErrorSystem.New(err))
	}
}

// renderAPIError は *factory.Error をJSONレスポンスとして書き出す。500系のときはスタック
// トレースを標準出力へ残す(将来エラータイプを使って改善できるようにするための暫定対応)。
// apiErrorがtyped-nilの場合、errors.Asはerror interfaceとしては非nilと判定するため
// switch側では弾けない。*apiErrorの参照でpanicするため、ここでnilを弾く(MIR-1133で
// このswitchにdefaultを足すのに合わせて手当てした、既存の潜在バグ)。
func renderAPIError(c echo.Context, apiError *factory.Error) {
	if apiError == nil {
		apiError = factory.ErrorSystem.New(nil)
	}

	c.JSON(apiError.Status(), EchoApiResult{
		Success: false,
		Error:   *apiError,
	})

	if c.Response().Status >= http.StatusInternalServerError {
		if apiError.StackTrace() != nil {
			requestId := c.Response().Header().Get(echo.HeaderXRequestID)
			fmt.Printf("[INTERNAL ERROR : %v]%+v\n", requestId, apiError.StackTrace())
		}
	}
}

func CustomContext() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			ctx := req.Context()
			rid := req.Header.Get(echo.HeaderXRequestID)
			if rid == "" {
				rid = ctxbase.NewID()
			}
			c.Response().Header().Set(echo.HeaderXRequestID, rid)

			cb := ctxbase.ContextBase{
				RequestID: rid,
				ActionID:  ctxbase.NewID(),
			}
			c.SetRequest(req.WithContext(context.WithValue(ctx, ctxbase.ContextBaseName, &cb)))
			c.SetRequest(req.WithContext(context.WithValue(ctx, "request_id", rid)))
			return next(c)
		}
	}
}

func CustomDebugContext(isDebug bool) []echo.MiddlewareFunc {
	if !isDebug {
		return nil
	}

	logrus.SetLevel(logrus.DebugLevel)

	return []echo.MiddlewareFunc{
		middleware.BodyDumpWithConfig(middleware.BodyDumpConfig{
			Skipper: skipper,
			Handler: func(c echo.Context, reqBody, resBody []byte) {
				// /docサブパスへのリクエストはロギングから除外
				if strings.HasPrefix(c.Request().URL.Path, "/doc") {
					return
				}

				if strings.HasPrefix(c.Request().URL.Path, "/api/swagger") {
					return
				}

				// 本文サイズが10KBを超える場合は切り詰める
				maxSize := 10240 // 10KB

				entry := logrus.WithFields(logrus.Fields{
					"request_id": c.Response().Header().Get(echo.HeaderXRequestID),
					"request":    truncateBody(reqBody, maxSize),
					"response":   truncateBody(resBody, maxSize),
				})
				entry.Debug("request/response body")
			},
		}),

		func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				req := c.Request()

				requestId := c.Response().Header().Get(echo.HeaderXRequestID)
				c.SetRequest(req.WithContext(context.WithValue(req.Context(), "__xorm_session_id", requestId)))
				return next(c)
			}
		},
	}
}

// redactAozoraAuthQuery は、/gmo-aozora/auth/ 配下(state・codeを含む
// GmoあおぞらのOAuth関連エンドポイント)のアクセスログから、クエリ文字列
// (state・code・admin-apiが発行したワンタイム値等)を落とす(MIR-1298)。
// アクセスログはURIをクエリごと出す仕様のため、この2エンドポイントだけ例外的に
// 削る。
func redactAozoraAuthQuery(uri string) string {
	path := uri
	if idx := strings.Index(uri, "?"); idx != -1 {
		path = uri[:idx]
	} else {
		return uri
	}
	if strings.Contains(path, "/gmo-aozora/auth/") {
		return path
	}
	return uri
}

func CustomLogger() echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		Skipper: skipper,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			entry := logrus.WithFields(logrus.Fields{
				"log_type":   "request",
				"request_id": v.RequestID,
				"remote_ip":  c.RealIP(),
				"host":       v.Host,
				"uri":        redactAozoraAuthQuery(v.URI),
				"method":     v.Method,
				"referer":    v.Referer,
				"user_agent": v.UserAgent,
				"status":     v.Status,
				"latency":    v.Latency.String(),
				"bytes_in":   v.ContentLength,
				"bytes_out":  v.ResponseSize,
			})

			if v.Error == nil {
				entry.Info("request info")
			} else {
				var err *factory.Error

				if errors.As(v.Error, &err) {
					entry = entry.WithField("status", err.Status())
				}

				entry.WithError(v.Error).Error("error occurred")
			}
			return nil
		},
		LogRequestID:     true,
		LogLatency:       true,
		LogRemoteIP:      true,
		LogHost:          true,
		LogMethod:        true,
		LogURI:           true,
		LogReferer:       true,
		LogStatus:        true,
		LogUserAgent:     true,
		LogContentLength: true,
		LogResponseSize:  true,
		LogError:         true,
	})
}

// kafkaへアクセスログを送信するロガーミドルウェア (admin-api専用、一時的に保持)
func CustomKafkaLogger(eventBus eventbus.IEventBus, cfg config.Config) []echo.MiddlewareFunc {
	var resBytes []byte
	return []echo.MiddlewareFunc{
		func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				if c.Path() == "/api/v1/admin-users/login" {
					handler := func(c echo.Context, reqBody, resBody []byte) {
						resBytes = resBody
					}
					bodyDumpMiddleware := middleware.BodyDump(handler)
					return bodyDumpMiddleware(next)(c)
				}
				return next(c)
			}
		},

		middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
			Skipper: skipper,
			LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
				startTime := v.StartTime.UnixMilli()
				entry := logrus.WithFields(logrus.Fields{
					"log_type":   "request",
					"request_id": v.RequestID,
					"remote_ip":  c.RealIP(),
					"host":       v.Host,
					"path":       c.Path(),
					"uri":        v.URI,
					"method":     v.Method,
					"referer":    v.Referer,
					"user_agent": v.UserAgent,
					"status":     v.Status,
					"latency":    v.Latency.String(),
					"bytes_in":   v.ContentLength,
					"bytes_out":  v.ResponseSize,
					"date":       startTime,
				})

				var actor domain.OperatorHistoryActor
				if c.Request().Header.Get(echo.HeaderAuthorization) != "" {
					user, err := extractFromContext(c)
					// 認証期限切れの場合、トークンを取得できずエラーになるのでコメントアウト
					// if err != nil {
					// 	return echo.NewHTTPError(http.StatusBadRequest, err.Error())
					// }
					if err == nil {
						actor = domain.OperatorHistoryActor{
							Id:    user.Id,
							Name:  user.Name,
							Email: user.Email,
						}
					}
				}
				if c.Path() == "/api/v1/admin-users/login" {
					var res struct {
						Result struct {
							AccessToken string `json:"access_token" description:"アクセストークン"`
						} `json:"result"`
						Success bool `json:"success"`
						Error   interface{}
					}
					err := json.Unmarshal(resBytes, &res)
					if err != nil {
						return echo.NewHTTPError(http.StatusBadRequest, err.Error())
					}

					user, err := extractFromTokenString(res.Result.AccessToken, cfg.Jwt.Secret)
					if err != nil {
						return echo.NewHTTPError(http.StatusBadRequest, err.Error())
					}

					actor = domain.OperatorHistoryActor{
						Id:    user.Id,
						Name:  user.Name,
						Email: user.Email,
					}
				}

				entry = entry.WithFields(logrus.Fields{
					"user_id": actor.Id,
				})

				if v.Error == nil {
					entry.Info("request info")
				} else {
					var err *factory.Error
					if errors.As(v.Error, &err) {
						entry = entry.WithField("status", err.Status())
					}
					entry.WithError(v.Error).Error("error occurred")
				}

				return nil
			},
			LogRequestID:     true,
			LogLatency:       true,
			LogRemoteIP:      true,
			LogHost:          true,
			LogMethod:        true,
			LogURI:           true,
			LogReferer:       true,
			LogStatus:        true,
			LogUserAgent:     true,
			LogContentLength: true,
			LogResponseSize:  true,
			LogError:         true,
		}),
	}
}

func CustomRecover() echo.MiddlewareFunc {
	return middleware.RecoverWithConfig(middleware.RecoverConfig{
		LogErrorFunc: func(c echo.Context, err error, stack []byte) error {
			requestId := c.Response().Header().Get(echo.HeaderXRequestID)
			fmt.Printf("[INTERNAL ERROR: %v] %v\n%s\n", requestId, err, stack)

			return c.JSON(http.StatusInternalServerError, EchoApiResult{
				Success: false,
				Error:   *factory.ErrorSystem.New(err),
			})
		},
	})
}

type EchoApiResult struct {
	Result  interface{}   `json:"result"`
	Success bool          `json:"success"`
	Error   factory.Error `json:"error"`
}

func truncateBody(body []byte, maxSize int) string {
	if len(body) > maxSize {
		return string(body[:maxSize]) + " ... [truncated]"
	}
	return string(body)
}

// RefreshTokenCookieExtendはrefresh_tokenクッキーの有効期限を延長するミドルウェア
func RefreshTokenCookieExtend() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cookie, err := c.Cookie("refresh_token")
			if err == nil && cookie != nil {
				if serviceutils.EnvIsDev() {
					cookie.Expires = time.Now().Add(domain.TestRefreshTokenCookieExpireTime)
				} else {
					cookie.Expires = time.Now().Add(domain.RefreshTokenCookieExpireTime)
				}
				cookie.Path = "/"
				cookie.HttpOnly = true
				cookie.Secure = true
				c.SetCookie(cookie)
			}
			return next(c)
		}
	}
}

// AllowRolesは許可されたロールのみがアクセスできるようにするミドルウェアです。
func AllowRoles(allowedRoles ...domain.ROLE_TYPE) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userRoleMap, err := tokenToUserRoleMap(c)
			if err != nil {
				return echo.NewHTTPError(http.StatusForbidden, err.Error())
			}

			if !userHasAllowedRole(userRoleMap, allowedRoles) {
				return echo.NewHTTPError(http.StatusForbidden, "insufficient permissions to access this resource")
			}

			return next(c)
		}
	}
}

func AllowAuditStatuses(allowedStatuses ...domain.AUDIT_STATUS) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			auditStatus, err := tokenToUserAuditStatus(c)
			if err != nil {
				return factory.ErrorNotMatchAudit.New(err)
			}

			if !userHasAllowedAuditStatus(auditStatus, allowedStatuses) {
				return factory.ErrorNotMatchAudit.New(nil)
			}

			return next(c)
		}
	}
}

func tokenToUserAuditStatus(c echo.Context) (domain.AUDIT_STATUS, error) {
	user, ok := c.Get("user").(*jwt.Token)
	if !ok {
		return "", errors.New("invalid or missing user token")
	}

	claims, ok := user.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid token claims")
	}

	auditStatus, ok := claims["audit_status"].(string)
	if !ok {
		return "", errors.New("missing audit_status in token claims")
	}

	return domain.AUDIT_STATUS(auditStatus), nil
}

func userHasAllowedAuditStatus(auditStatus domain.AUDIT_STATUS, allowedStatuses []domain.AUDIT_STATUS) bool {
	return slices.Contains(allowedStatuses, auditStatus)
}

func tokenToUserRoleMap(c echo.Context) (map[string]struct{}, error) {
	user, ok := c.Get("user").(*jwt.Token)
	if !ok {
		return nil, errors.New("invalid or missing user token")
	}

	claims, ok := user.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	roleClaims, ok := claims["role"].([]interface{})
	if !ok {
		return nil, errors.New("missing roles in token claims")
	}

	roleMap := make(map[string]struct{}, len(roleClaims))
	for _, role := range roleClaims {
		if roleStr, ok := role.(string); ok {
			roleMap[roleStr] = struct{}{}
		} else {
			return nil, errors.New("invalid role format in token claims")
		}
	}

	return roleMap, nil
}

func userHasAllowedRole(userRoles map[string]struct{}, allowedRoles []domain.ROLE_TYPE) bool {
	for _, role := range allowedRoles {
		if _, exists := userRoles[string(role)]; exists {
			return true
		}
	}
	return false
}

func extractFromContext(c echo.Context) (jwthelper.User, error) {
	const (
		jwtUserKey  = "user"
		jwtEmailKey = "email"
		jwtNameKey  = "name"
	)

	user, ok := c.Get(jwtUserKey).(*jwt.Token)
	if !ok {
		return jwthelper.User{}, errors.New("invalid token")
	}

	claims, ok := user.Claims.(jwt.MapClaims)
	if !ok {
		return jwthelper.User{}, errors.New("invalid claims")
	}

	userIdStr, ok := claims[jwtUserKey].(string)
	if !ok {
		return jwthelper.User{}, errors.New("user ID missing or invalid in token")
	}
	id, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		return jwthelper.User{}, errors.New("user ID conversion error")
	}

	email, ok := claims[jwtEmailKey].(string)
	if !ok {
		return jwthelper.User{}, errors.New("no email in token")
	}

	name, ok := claims[jwtNameKey].(string)
	if !ok {
		return jwthelper.User{}, errors.New("no name in token")
	}

	roles, _ := jwthelper.TokenToRoleMap(c)

	return jwthelper.User{
		Id:    id,
		Email: email,
		Name:  name,
		Roles: roles,
	}, nil
}

func extractFromTokenString(tokenString, secret string) (jwthelper.User, error) {
	const (
		jwtUserKey  = "user"
		jwtEmailKey = "email"
		jwtNameKey  = "name"
	)

	user, err := jwt.ParseWithClaims(tokenString, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return jwthelper.User{}, errors.New("invalid token")
	}

	claims, ok := user.Claims.(jwt.MapClaims)
	if !ok {
		return jwthelper.User{}, errors.New("invalid claims")
	}

	userIdStr, ok := claims[jwtUserKey].(string)
	if !ok {
		return jwthelper.User{}, errors.New("user ID missing or invalid in token")
	}
	id, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		return jwthelper.User{}, errors.New("user ID conversion error")
	}

	email, ok := claims[jwtEmailKey].(string)
	if !ok {
		return jwthelper.User{}, errors.New("no email in token")
	}

	name, ok := claims[jwtNameKey].(string)
	if !ok {
		return jwthelper.User{}, errors.New("no name in token")
	}

	return jwthelper.User{
		Id:    id,
		Email: email,
		Name:  name,
	}, nil
}
