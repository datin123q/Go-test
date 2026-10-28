package controllers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

var DISPLAY_LIMIT = 6

type blogPostsApiController struct {
}

func NewBlogPostApiController() *blogPostsApiController {
	return &blogPostsApiController{}
}

func (c blogPostsApiController) Init(g echoswagger.ApiGroup) {
	g.GET("", c.GetAll).
		SetSummary("Get all blog posts").
		SetDeprecated()
}

func (a blogPostsApiController) GetAll(c echo.Context) error {
	var items []domain.BlogPost

	totalCnt, items, err := domain.BlogPost{}.GetAll(c.Request().Context(), 0, DISPLAY_LIMIT)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return renderSucc(c, http.StatusOK, ArrayResult{
		TotalCount: totalCnt,
		Items:      BlogPostResponse{}.NewList(items),
	})

}
