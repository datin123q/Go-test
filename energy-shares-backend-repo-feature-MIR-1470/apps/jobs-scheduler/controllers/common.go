package controllers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
)

type CommonApiControllers struct {
}

func (d CommonApiControllers) Init(g echoswagger.ApiGroup) {
	g.GET("/api/health-check", d.HealthCheck).
		SetSummary("헬스 체크").
		SetDescription("서버 헬스 체크 요청에 대한 결과를 반환합니다.").
		AddResponse(http.StatusOK, "서버 헬스 체크 요청에 대한 결과", nil, nil)
}

func (d CommonApiControllers) HealthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status": "UP",
	})
}
