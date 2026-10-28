package controllers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"

	adminservice "github.com/illuminarean/energy-share-core-api/service"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/apirender"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/echoutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/formutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/s3utils"
	"github.com/illuminarean/energy-shares-backend-repo/service"
)

type ColumnApiController struct {
	columnService  service.IColumnService
	contentService *adminservice.ContentImageService
}

func NewColumnApiController() *ColumnApiController {
	return &ColumnApiController{
		columnService:  service.NewColumnService(),
		contentService: adminservice.NewContentImageService(),
	}
}

func (api ColumnApiController) Init(g echoswagger.ApiGroup) {

	g.GET("", api.GetAll,
		echoutils.AllowRoles(domain.ROLE_OPERATOR)).
		AddParamQueryNested(ColumnSearchInput{}).
		AddResponse(http.StatusOK, "カラムリスト照会", []ColumnOutput{}, nil).
		SetDescription("カラムリストを照会します。(検索オプション: タイトル、ID、作成日)").
		SetSummary("カラムリスト照会")

	g.GET("/:id", api.GetOne,
		echoutils.AllowRoles(domain.ROLE_OPERATOR)).
		AddParamPath("", "id", "カラムID").
		AddResponse(http.StatusOK, "カラム照会", ColumnDetailOutput{}, nil).
		SetDescription("カラムを照会します.").
		SetSummary("カラム照会")

	g.POST("", api.Create,
		echoutils.AllowRoles(domain.ROLE_OPERATOR)).
		AddParamForm(ColumnInput{}, "body", "", true).
		AddParamFile("thumbnail", "サムネイル", true).
		AddResponse(http.StatusCreated, "カラム登録", ColumnOutput{}, nil).
		SetDescription("カラムを登録します.").
		SetSummary("カラム登録").
		SetRequestContentType("multipart/form-data").
		SetSecurity("Authorization")

	g.PUT("/:id", api.Update,
		echoutils.AllowRoles(domain.ROLE_OPERATOR)).
		AddParamPath("", "id", "カラムID").
		AddParamForm(ColumnInput{}, "body", "", false).
		AddParamFile("thumbnail", "サムネイル", false).
		AddResponse(http.StatusOK, "カラム修正", ColumnOutput{}, nil).
		SetDescription("カラムを修正します.").
		SetSummary("カラム修正").
		SetRequestContentType("multipart/form-data").
		SetSecurity("Authorization")

	g.DELETE("/:id", api.Delete,
		echoutils.AllowRoles(domain.ROLE_OPERATOR)).
		AddParamPath("", "id", "カラムID").
		AddResponse(http.StatusOK, "カラム削除", nil, nil).
		SetDescription("カラムを削除します.").
		SetSummary("カラム削除").
		SetSecurity("Authorization")
}

func (api ColumnApiController) Create(c echo.Context) error {

	var input ColumnInput
	if err := c.Bind(&input); err != nil {
		factory.ErrorParameter.New(err)
	}
	if err := c.Validate(&input); err != nil {
		return factory.ErrorParameter.New(err)
	}

	userId, email, err := jwthelper.TokenToUser(c)
	if err != nil {
		factory.ErrorFailAuthorized.New(err)
	}

	_, thumbnail, err := c.Request().FormFile(strings.ToLower(string(domain.COLUMN_THUMBNAIL)))
	if err != nil {
		factory.ErrorInvalidForm.New(err)
	}

	form, err := c.MultipartForm()
	if err != nil {
		return factory.ErrorInvalidForm.New(err)
	}

	column := input.ToModel()
	column.UserId = userId

	id, err := uuid.NewV4()
	if err != nil {
		return factory.ErrorSystem.New(err)
	}
	column.CommonId = id.String()

	column.Content, err = api.contentService.ReplaceContentImage(c.Request().Context(), s3utils.BUCKET_PUBLIC_NAME, column.Content, column.CommonId, formutils.FromToUrls(*form, "content_images"))

	if err := api.columnService.CreateColumn(c.Request().Context(), column, thumbnail); err != nil {
		return err
	}

	err = domain.OperatorHistory{}.Record(c.Request().Context(), userId, email,
		"menu.manage_website.manage_column", column.Title, c.RealIP(), domain.REGISTER)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusCreated, ColumnOutput{}.New(column))
}

func (api ColumnApiController) GetAll(c echo.Context) error {

	var input ColumnSearchInput
	if err := c.Bind(&input); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if input.MaxResultCount == 0 {
		input.MaxResultCount = DefaultMaxResultCount
	}

	total, columnList, err := api.columnService.GetColumnList(c.Request().Context(), input.Title, input.Id, input.StartDate, input.EndDate, input.Sortby, input.Order, input.SkipCount, input.MaxResultCount)
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

	return apirender.RenderSucc(c, http.StatusOK, ColumnDetailOutput{}.New(column))
}

func (api ColumnApiController) Update(c echo.Context) error {

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	var input ColumnInput
	if err := c.Bind(&input); err != nil {
		factory.ErrorParameter.New(err)
	}
	if err := c.Validate(&input); err != nil {
		factory.ErrorParameter.New(err)
	}

	userId, email, err := jwthelper.TokenToUser(c)
	if err != nil {
		factory.ErrorFailAuthorized.New(err)
	}

	_, thumbnail, err := c.Request().FormFile(strings.ToLower(string(domain.COLUMN_THUMBNAIL)))
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			thumbnail = nil
		} else {
			factory.ErrorInvalidForm.New(err)
		}
	}

	form, err := c.MultipartForm()
	if err != nil {
		return err
	}

	column := input.ToModel()
	column.Id = id
	column.UserId = userId
	column.Content, err = api.contentService.ReplaceContentImage(c.Request().Context(), s3utils.BUCKET_PUBLIC_NAME, column.Content, column.CommonId, formutils.FromToUrls(*form, "content_images"))

	if thumbnail != nil {
		if err := api.columnService.DeleteThumbnailByColumnId(c.Request().Context(), id); err != nil {
			return err
		}
		if err := api.columnService.UpdateColumnAndThumbnail(c.Request().Context(), column, thumbnail); err != nil {
			return err
		}
	} else {
		if err := api.columnService.UpdateColumn(c.Request().Context(), column); err != nil {
			return err
		}
	}

	updated, err := api.columnService.GetColumnById(c.Request().Context(), id)
	if err != nil {
		return err
	}

	err = domain.OperatorHistory{}.Record(c.Request().Context(), userId, email,
		"menu.manage_website.manage_column", column.Title, c.RealIP(), domain.CHANGE)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, ColumnOutput{}.New(updated))
}

func (api ColumnApiController) Delete(c echo.Context) error {

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	userId, email, err := jwthelper.TokenToUser(c)
	if err != nil {
		factory.ErrorFailAuthorized.New(err)
	}

	if err := api.columnService.DeleteColumnById(c.Request().Context(), id); err != nil {
		return err
	}

	err = domain.OperatorHistory{}.Record(c.Request().Context(), userId, email,
		"menu.manage_website.manage_column", c.Param("id"), c.RealIP(), domain.DELETE)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, nil)
}
