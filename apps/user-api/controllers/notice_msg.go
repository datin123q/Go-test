package controllers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

type NoticeMsgApiController struct {
}

func (c NoticeMsgApiController) Init(g echoswagger.ApiGroup) {
	g.GET("/:id", c.GetOne).
		AddParamPath("", "id", "").
		SetDescription("ユーザー通知情報をリクエストします。").
		SetSummary("ユーザー通知情報リクエスト").
		SetSecurity("Authorization")

}

func (NoticeMsgApiController) GetOne(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	tokenId, _, err := tokenToUser(c)
	if err != nil {
		return err
	}

	noticeMsg, err := domain.NoticeMsg{}.GetByIdWithUserId(c.Request().Context(), id, tokenId)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	if noticeMsg == nil {
		return factory.ErrorNotFound.New(err)
	}

	return renderSucc(c, http.StatusOK, noticeMsg)
}
