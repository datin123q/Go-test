package controllers

import (
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/apirender"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/s3utils"
)

type AuthApiController struct {
}

func (c AuthApiController) Init(g echoswagger.ApiGroup) {

	g.POST("/update-token", c.UpdateToken).
		SetSummary("管理者トークン更新")
}

func (AuthApiController) UpdateToken(c echo.Context) error {

	// クッキーからトークンを抽出
	cookie, err := c.Cookie("refresh_token")
	if err != nil {
		return factory.ErrorTokenInvalid.New(err) // クッキーがない場合またはエラーが発生した場合
	}
	refreshTokenInCookie := cookie.Value

	// リフレッシュトークンを照会および有効期限確認
	refreshToken, err := domain.RefreshToken{}.GetByClientInfo(c.Request().Context(), refreshTokenInCookie, c.RealIP(), c.Request().UserAgent())
	if err != nil {
		return factory.ErrorTokenInvalid.New(err)
	}
	if refreshToken == nil {
		return factory.ErrorTokenInvalid.New(errors.New("not found refresh token"))
	}

	if refreshToken.ExpireIn < time.Now().Unix() {
		return factory.ErrorTokenInvalid.New(errors.New("refresh token expired"))
	}

	// 管理者のrefresh_tokenだけを受け付ける。会員のrefresh_tokenのuser_idは会員のIDで、admin_userのIDと重なる
	// ため、通すと同じ番号の管理者のトークンが出てしまう(MIR-1407)。
	if refreshToken.UserType != domain.ADMIN {
		return factory.ErrorTokenInvalid.New(errors.New("not admin refresh token"))
	}

	// ユーザーを照会
	user, err := domain.AdminUser{}.GetById(c.Request().Context(), refreshToken.UserId)
	if err != nil {
		return factory.ErrorTokenInvalid.New(err)
	}
	if user == nil {
		return factory.ErrorTokenInvalid.New(errors.New("not found user"))
	}

	accessToken, err := domain.CreateAdminToken(c.Request().Context(), *user)
	if err != nil {
		return factory.ErrorTokenInvalid.New(err)
	}

	cookies, err := s3utils.GetCookieSign()
	if err != nil {
		return factory.ErrorSystem.New(err)
	}
	for _, cookie := range cookies {
		c.SetCookie(cookie)
	}

	// リフレッシュトークンは更新しない。
	return apirender.RenderSucc(c, http.StatusOK, domain.ResultToken{
		AccessToken: accessToken,
	})
}
