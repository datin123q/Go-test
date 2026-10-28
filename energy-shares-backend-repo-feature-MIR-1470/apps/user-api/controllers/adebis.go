package controllers

import (
	"net/http"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
)

type AdEbisAPIController struct {
}

func (a AdEbisAPIController) Init(g echoswagger.ApiGroup) {
	// Implementation will be added in the future
	g.GET("/itp", a.HandleITP).
		AddResponse(http.StatusOK, "ITP対応用のCokkieをセット", nil, nil).
		SetDescription("Handle ITP requests").
		SetSummary("ITP Handler")
	g.POST("/itp", a.HandleITP).
		AddResponse(http.StatusOK, "ITP対応用のCokkieをセット", nil, nil).
		SetDescription("Handle ITP requests").
		SetSummary("ITP Handler")
}

func (a AdEbisAPIController) HandleITP(c echo.Context) error {
	domain.AdEbis{}.SetITPMeasuresCookie(c.Request(), c.Response())
	return renderSucc(c, http.StatusOK, nil)
}
