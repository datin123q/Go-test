package controllers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
	"github.com/pkg/errors"

	"energyshares-job-scheduler/pkg/gmo_aozora"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

type WebHookApiControllers struct {
}

func (c WebHookApiControllers) Init(g echoswagger.ApiGroup) {
	g.POST("/start-subscription", c.StartSubscription).
		AddResponse(http.StatusOK, "GMO Webhook 시작", nil, nil).
		SetDescription("수동으로 웹 시작을 요청하는 경우")

	g.POST("/stop-subscription", c.StopSubscription).
		AddResponse(http.StatusOK, "GMO Webhook 종료", nil, nil).
		SetDescription("수동으로 웹 종료를 요청하는 경우")

	g.GET("/status-subscription", c.StatusSubscription).
		AddResponse(http.StatusOK, "GMO Webhook 구독 상태 조회", nil, nil).
		SetDescription("웹훅 상태를 수동으로 조회 하는 경우")

	g.GET("/unsubscribed-history", c.GetUnsubscribedHistory).
		AddResponse(http.StatusOK, "GMO Webhook 미수신 이력 조회", nil, nil).
		SetDescription("수동으로 웹훅 미수신 이력을 조회 하는 경우. 장기간 웹훅이 작동하지 않는 경우 수행이 필요함")
}

func (WebHookApiControllers) StartSubscription(c echo.Context) error {
	err := gmo_aozora.WebhookSubscription(gmo_aozora.START_SUBSCRIPTION)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}
	return renderSucc(c, http.StatusOK, nil)
}

func (WebHookApiControllers) StopSubscription(c echo.Context) error {
	err := gmo_aozora.WebhookSubscription(gmo_aozora.STOP_SUBSCRIPTION)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}
	return renderSucc(c, http.StatusOK, nil)
}

func (WebHookApiControllers) StatusSubscription(c echo.Context) error {
	client := gmo_aozora.NewWebHookClient()
	resp, err := client.SubscribeStatus(c.Request().Context())
	if err != nil {
		return factory.ErrorParameter.New(errors.Wrap(err, "Webhook Subscription Error"))
	}
	return renderSucc(c, http.StatusOK, resp)
}

func (WebHookApiControllers) GetUnsubscribedHistory(c echo.Context) error {
	client := gmo_aozora.NewWebHookClient()
	resp, err := client.GetList(c.Request().Context())
	if err != nil {
		return factory.ErrorParameter.New(errors.Wrap(err, "Webhook Subscription Error"))
	}
	return renderSucc(c, http.StatusOK, resp)
}
