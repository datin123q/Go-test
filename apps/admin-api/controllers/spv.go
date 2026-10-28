package controllers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
	"github.com/sirupsen/logrus"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/apirender"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/echoutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
)

type SpvApiController struct {
}

func (c SpvApiController) Init(g echoswagger.ApiGroup) {
	g.GET("", c.GetAll,
		echoutils.AllowRoles(domain.ROLE_PRODUCT)).
		AddParamQueryNested(SpvSearchInput{}).
		AddResponse(http.StatusOK, "開発会社リスト", domain.Spv{}, nil).
		SetDescription("SPV開発社リストをリクエストしています。").
		SetSummary("SPV開発社リストをリクエスト").
		SetSecurity("Authorization")

	g.POST("", c.Create,
		echoutils.AllowRoles(domain.ROLE_PRODUCT)).
		AddParamBody(SpvInput{}, "body", "", true).
		AddResponse(http.StatusCreated, "開発会社を登録します。", domain.Spv{}, nil).
		SetDescription("SPV開発社リストを追加します。").
		SetSummary("SPV開発社リストを追加します。").
		SetSecurity("Authorization")

	g.GET("/existscorpno", c.ExistsCropNo,
		echoutils.AllowRoles(domain.ROLE_PRODUCT)).
		AddParamQueryNested(SpvCropNoInput{}).
		AddResponse(http.StatusOK, "重複事業者番号チェック", "true or false", nil).
		SetDescription("同じ事業者番号の存在有無チェック").
		SetSummary("同じ事業者番号の存在有無チェック")

	g.GET("/:id", c.GetOne,
		echoutils.AllowRoles(domain.ROLE_PRODUCT)).
		AddParamPath("", "id", "").
		AddResponse(http.StatusOK, "開発会社詳細情報", domain.Spv{}, nil).
		SetDescription("単一SPV開発社エンティティを照会します。").
		SetSummary("単一SPV開発社詳細照会").
		SetSecurity("Authorization")

	g.PUT("/:id", c.Update,
		echoutils.AllowRoles(domain.ROLE_PRODUCT)).
		AddParamPath("", "id", "").
		AddParamBody(SpvInput{}, "body", "", true).
		AddResponse(http.StatusOK, "開発会社情報修正", domain.Spv{}, nil).
		SetDescription("単一SPV開発社エンティティを修正します。").
		SetSummary("単一SPV開発社エンティティを修正").
		SetSecurity("Authorization")

	g.GET("/:id/projects", c.GetProjects,
		echoutils.AllowRoles(domain.ROLE_PRODUCT)).
		AddParamPath("", "id", "").
		AddParamQueryNested(ProjectSearchInput{}).
		AddResponse(http.StatusOK, "開発会社と連結されたプロジェクトリスト", domain.Project{}, nil).
		SetDescription("開発社と連結されたプロジェクトリスト情報をリクエストしています。").
		SetSummary("SPV開発社プロジェクト照会")

	g.GET("/download", c.Download,
		echoutils.AllowRoles(domain.ROLE_PRODUCT)).
		AddParamQueryNested(SpvSearchInput{}).
		AddResponse(http.StatusOK, "開発会社リストダウンロード", domain.Spv{}, nil).
		SetDescription("開発会社リストをダウンロードします。").
		SetSummary("開発会社リストダウンロード").
		SetSecurity("Authorization")
}

func (SpvApiController) GetAll(c echo.Context) error {
	var v SpvSearchInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if v.MaxResultCount == 0 {
		v.MaxResultCount = DefaultMaxResultCount
	}

	// console log
	factory.Logger(c.Request().Context()).WithFields(logrus.Fields{
		"sortby":           v.Sortby,
		"order":            v.Order,
		"max_result_count": v.MaxResultCount,
		"skip_count":       v.SkipCount,
	}).Info("SearchStart")

	totalCount, items, err := domain.Spv{}.GetAll(c.Request().Context(), v.Id, v.Name, v.Email, v.Phone, v.Sortby, v.Order, v.SkipCount, v.MaxResultCount)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, apirender.ArrayResult{
		TotalCount: totalCount,
		Items:      items,
	})
}

func (SpvApiController) Create(c echo.Context) error {
	var v SpvInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}

	uid, email, err := jwthelper.TokenToUser(c)
	if err != nil {
		return err
	}

	company, err := v.ToModel()
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	company.OwnerId = uid
	company.UpdateId = uid

	if _, err := company.Create(c.Request().Context()); err != nil {
		return factory.ErrorDB.New(err)
	}

	err = domain.OperatorHistory{}.Record(c.Request().Context(), uid, email,
		"menu.manage_product.developer_register", company.Name, c.RealIP(), domain.REGISTER)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	return apirender.RenderSucc(c, http.StatusCreated, company)
}

// 重複事業者番号チェック
func (SpvApiController) ExistsCropNo(c echo.Context) error {
	var v SpvCropNoInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	company, err := domain.Spv{}.GetByCropNo(c.Request().Context(), v.CorpNo)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}
	if company != nil {
		return apirender.RenderSucc(c, http.StatusOK, true)
	}
	return apirender.RenderSucc(c, http.StatusOK, false)
}

func (SpvApiController) GetOne(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}
	v, err := domain.Spv{}.GetById(c.Request().Context(), id)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	if v == nil {
		return factory.ErrorNotFound.New(err)
	}
	uid, email, err := jwthelper.TokenToUser(c)
	if err != nil {
		return factory.ErrorTokenInvalid.New(err)
	}

	err = domain.OperatorHistory{}.Record(c.Request().Context(), uid, email,
		"menu.manage_product.developer_detail", v.Name, c.RealIP(), domain.LOOKUP)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	return apirender.RenderSucc(c, http.StatusOK, v)
}

func (SpvApiController) Update(c echo.Context) error {
	var v SpvInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}

	uid, email, err := jwthelper.TokenToUser(c)
	if err != nil {
		return err
	}

	spv, err := v.ToModel()
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}
	spv.Id = id

	before, err := domain.Spv{}.GetById(c.Request().Context(), id)
	if err != nil {
		return factory.ErrorParameter.New(err)
	} else if before == nil {
		return factory.ErrorNotFound.New(err)
	}

	spv.UpdateId = uid

	if err := spv.Update(c.Request().Context()); err != nil {
		return factory.ErrorDB.New(err)
	}

	after, err := domain.Spv{}.GetById(c.Request().Context(), spv.Id)
	if err != nil {
		return factory.ErrorParameter.New(err)
	} else if after == nil {
		return factory.ErrorDB.New(err)
	}
	_, err = domain.SpvInfoHistory{}.Create(c.Request().Context(), before, after, uid)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	err = domain.OperatorHistory{}.Record(c.Request().Context(), uid, email,
		"menu.manage_product.developer_detail", v.Name, c.RealIP(), domain.CHANGE)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, spv)
}

func (SpvApiController) GetProjects(c echo.Context) error {
	var v ProjectSearchInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if v.MaxResultCount == 0 {
		v.MaxResultCount = DefaultMaxResultCount
	}

	// console log
	factory.Logger(c.Request().Context()).WithFields(logrus.Fields{
		"sortby":           v.Sortby,
		"order":            v.Order,
		"max_result_count": v.MaxResultCount,
		"skip_count":       v.SkipCount,
	}).Info("SearchStart")

	spvId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	totalCount, items, err := domain.Project{}.GetProjectsByCompanyId(c.Request().Context(), v.Sortby, v.Order, v.SkipCount, v.MaxResultCount, spvId)

	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, apirender.ArrayResult{
		TotalCount: totalCount,
		Items:      items,
	})
}

func (SpvApiController) Download(c echo.Context) error {
	var v SpvSearchInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if v.MaxResultCount == 0 {
		v.MaxResultCount = DefaultMaxResultCount
	}
	userId, email, err := jwthelper.TokenToUser(c)
	if err != nil {
		return factory.ErrorTokenInvalid.New(err)
	}

	// console log
	factory.Logger(c.Request().Context()).WithFields(logrus.Fields{
		"sortby":           v.Sortby,
		"order":            v.Order,
		"max_result_count": v.MaxResultCount,
		"skip_count":       v.SkipCount,
	}).Info("SearchStart")

	err = domain.Spv{}.DownloadCsv(c.Request().Context(), c.Response().Writer, v.Id, v.Name, v.Email, v.Phone, v.Sortby, v.Order, v.SkipCount, v.MaxResultCount)

	if err != nil {
		factory.ErrorFailDownloadCSV.New(err)
	}

	err = domain.OperatorHistory{}.Record(c.Request().Context(), userId, email,
		"menu.manage_product.developer_list", "", c.RealIP(), domain.DOWNLOAD)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return nil
}
