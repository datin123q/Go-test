package controllers

import (
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/s3utils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
)

type AuthApiController struct {
}

func (c AuthApiController) Init(g echoswagger.ApiGroup) {

	g.POST("/update-token", c.UpdateToken).
		SetSummary("JWT 토큰 업데이트")
}

func (AuthApiController) UpdateToken(c echo.Context) error {

	// 쿠키에서 토큰 추출
	cookie, err := c.Cookie("refresh_token")
	if err != nil {
		return factory.ErrorTokenInvalid.New(err) // 쿠키가 없거나 에러가 발생한 경우
	}
	refreshTokenInCookie := cookie.Value

	refreshToken, err := domain.RefreshToken{}.GetByClientInfo(c.Request().Context(), refreshTokenInCookie, c.RealIP(), c.Request().UserAgent())
	if err != nil {
		return factory.ErrorTokenInvalid.New(err)
	}
	if refreshToken == nil {
		return factory.ErrorTokenInvalid.New(errors.New("not found token"))
	}

	if refreshToken.ExpireIn < time.Now().Unix() {
		return factory.ErrorTokenInvalid.New(errors.New("refresh token expired"))
	}

	// 会員のrefresh_tokenだけを受け付ける。管理者のrefresh_tokenのuser_idは管理者のIDで、userテーブルのIDと重なる
	// ため、通すと同じ番号の会員のトークンが出てしまう(MIR-1407)。
	if refreshToken.UserType != domain.PRIVATE && refreshToken.UserType != domain.CORPORATE {
		return factory.ErrorTokenInvalid.New(errors.New("not member refresh token"))
	}

	// 사용자 조회
	user, err := domain.User{}.GetById(c.Request().Context(), refreshToken.UserId)
	if err != nil {
		return factory.ErrorTokenInvalid.New(err)
	}
	if user == nil {
		return factory.ErrorTokenInvalid.New(errors.New("not Found User"))
	}

	// refresh token 재발급
	if err := refreshToken.UpdateExpireIn(c.Request().Context(), *user); err != nil {
		return factory.ErrorDB.New(err)
	}

	// access token 재발급
	accessToken, err := domain.CreateToken(c.Request().Context(), userForRefreshedToken(*user))
	if err != nil {
		return factory.ErrorTokenInvalid.New(err)
	}

	// cookie 에 refresh token 저장
	cookie.Name = "refresh_token"
	cookie.Value = refreshToken.Token

	if serviceutils.EnvIsDev() {
		cookie.Expires = time.Now().Add(domain.TestRefreshTokenCookieExpireTime)
	} else {
		cookie.Expires = time.Now().Add(domain.RefreshTokenCookieExpireTime)
	}

	cookie.HttpOnly = true
	cookie.Secure = true
	cookie.Path = "/"
	c.SetCookie(cookie)

	cookies, err := s3utils.GetCookieSign()
	if err != nil {
		return factory.ErrorSystem.New(err)
	}
	for _, cookie := range cookies {
		c.SetCookie(cookie)
	}

	// 반환은 access token 만
	return renderSucc(c, http.StatusOK, domain.ResultToken{
		AccessToken: accessToken,
	})
}

// userForRefreshedToken は、トークン更新でaccess tokenに載せる会員を返す。ログイン直後(CreateUserAccessToken)と
// 同じくROLE_USERだけにする。GetByIdはroleテーブルを会員のIDで引くが、roleテーブルは管理者のIDで権限を
// 持つため、IDが重なる会員のトークンに管理者の権限(ROLE_PAYMENTなど)が入ってしまう(MIR-1407)。
func userForRefreshedToken(user domain.User) domain.User {
	user.Roles = []string{string(domain.ROLE_USER)}
	return user
}
