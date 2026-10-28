package controllers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
	"github.com/pkg/errors"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type EventApiControllers struct {
}

func (c EventApiControllers) Init(g echoswagger.ApiGroup) {
	g.POST("/events/republish", c.RePublish).
		AddResponse(http.StatusOK, "미발송 이벤트 메시지 송신 결과", nil, nil).
		SetDescription("미발송 이벤트 메시지 송신")
}

func (EventApiControllers) RePublish(c echo.Context) error {
	err := domain.EventMessage{}.RePublish(c.Request().Context())
	if err != nil {
		return errors.Wrap(err, "fail to republish event")
	}
	return nil
}
