package controllers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/apirender"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/service"
)

type ColumnApiController struct {
	columnService service.IColumnService
}

func NewColumnApiController() *ColumnApiController {
	return &ColumnApiController{
		columnService: service.NewColumnService(),
	}
}

func (api ColumnApiController) Init(g echoswagger.ApiGroup) {

	g.GET("", api.GetAll).
		AddParamQueryNested(ColumnSearchInput{}).
		AddResponse(http.StatusOK, "컬럼 목록 조회", []ColumnOutput{}, nil).
		SetDescription("컬럼 목록을 조회 합니다. (검색 옵션: 제목)").
		SetSummary("컬럼 목록 조회")

	g.GET("/:id", api.GetOne).
		AddParamPath("", "id", "컬럼 ID").
		AddResponse(http.StatusOK, "컬럼 조회", ColumnDetailOutput{}, nil).
		SetDescription("컬럼을 조회 합니다.").
		SetSummary("컬럼 조회")

	g.GET("/:id/preview/:user_id", api.GetPreview).
		AddParamPath("", "id", "컬럼 ID").
		AddParamPath("", "userId", "유저 ID").
		AddResponse(http.StatusOK, "컬럼 미리 보기 조회", ColumnDetailOutput{}, nil).
		SetDescription("어드민이 컬럼을 미리 보기 형태로 조회 합니다.").
		SetSummary("컬럼 미리 보기 조회")
}

func (api ColumnApiController) GetAll(c echo.Context) error {

	var input ColumnSearchInput
	if err := c.Bind(&input); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if input.MaxResultCount == 0 {
		input.MaxResultCount = DefaultMaxResultCount
	}

	total, columnList, err := api.columnService.GetPostedColumnList(c.Request().Context(), input.Title, input.Sortby, input.Order, input.SkipCount, input.MaxResultCount)
	if err != nil {
		return err
	}

	type ArrayResult struct {
		Total int64       `json:"total_count"`
		Items interface{} `json:"items"`
	}

	return apirender.RenderSucc(c, http.StatusOK, ArrayResult{
		Total: total,
		Items: ColumnOutput{}.NewList(columnList),
	})
}

func (api ColumnApiController) GetOne(c echo.Context) error {

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	column, err := api.columnService.GetColumnById(c.Request().Context(), id)
	if err != nil {
		return err
	}

	if column.IsPosted == false {
		return factory.ErrorNotFound.New(nil)
	}

	return apirender.RenderSucc(c, http.StatusOK, ColumnDetailOutput{}.New(column))
}

func (api ColumnApiController) GetPreview(c echo.Context) error {

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	userId, err := strconv.ParseInt(c.Param("user_id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	admin, err := domain.AdminUser{}.GetById(c.Request().Context(), userId)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	if admin == nil {
		return factory.ErrorNotFound.New(err)
	}

	column, err := api.columnService.GetColumnById(c.Request().Context(), id)
	if err != nil {
		return err
	}

	return apirender.RenderSucc(c, http.StatusOK, ColumnDetailOutput{}.New(column))
}
