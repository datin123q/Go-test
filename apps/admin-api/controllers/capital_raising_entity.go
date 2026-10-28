package controllers

import (
	"net/http"
	"strconv"

	"github.com/illuminarean/energy-share-core-api/service"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/apirender"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/s3utils"
	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
	"github.com/sirupsen/logrus"
)

type CapitalRaisingEntityApiController struct {
	contentImageService *service.ContentImageService
}

func (cre CapitalRaisingEntityApiController) Init(g echoswagger.ApiGroup) {
	cre.contentImageService = service.NewContentImageService()

	g.POST("", cre.Create).
		AddParamForm(CapitalRaisingEntityInput{}, "body", "", true).
		AddParamFile("image_file", "画像", false).
		AddResponse(http.StatusOK, "", CapitalRaisingEntityAdminResponseDto{}, nil).
		SetDescription("資金調達者を登録します。").
		SetSummary("資金調達者登録").
		SetSecurity("Authorization")

	g.GET("/:id", cre.GetOne).
		AddParamPath("", "id", "").
		AddResponse(http.StatusOK, "", CapitalRaisingEntityAdminResponseDto{}, nil).
		SetDescription("資金調達者の詳細を取得します。").
		SetSummary("資金調達者詳細取得").
		SetSecurity("Authorization")

	g.GET("", cre.GetAll).
		AddParamQueryNested(CapitalRaisingEntitySearchInput{}).
		AddResponse(http.StatusOK, "", CapitalRaisingEntityAdminResponseDto{}, nil).
		SetDescription("資金調達者一覧を取得します。").
		SetSummary("資金調達者一覧取得").
		SetSecurity("Authorization")

	g.PUT("/:id", cre.Update).
		AddParamForm(ProjectInput{}, "body", "", true).
		AddParamPath("", "id", "").
		AddParamFile("image_file", "画像", false).
		AddResponse(http.StatusOK, "", CapitalRaisingEntityAdminResponseDto{}, nil).
		SetDescription("資金調達者を更新します。").
		SetSummary("資金調達者更新").
		SetSecurity("Authorization")
}

func (cre CapitalRaisingEntityApiController) Create(c echo.Context) error {
	var input CapitalRaisingEntityInput
	if err := c.Bind(&input); err != nil {
		return factory.ErrorParameter.New(err)
	}

	if err := c.Validate(&input); err != nil {
		return factory.ErrorParameter.New(err)
	}

	capitalRaisingEntity, err := input.ToModel()
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	// 先にDB作成（ID確定させる）
	if _, err := capitalRaisingEntity.Create(c.Request().Context()); err != nil {
		return factory.ErrorDB.New(err)
	}

	// ---- 画像アップロード処理
	file, fileHeader, err := c.Request().FormFile("image_file")
	if err == nil && fileHeader != nil {
		defer file.Close()

		imageUrl, err := s3utils.UploadAttachmentFile(
			c.Request().Context(),
			s3utils.BUCKET_PUBLIC_NAME,
			&domain.Attachment{
				FileName:      s3utils.MakeS3FileName(),
				OriginalName:  fileHeader.Filename,
				ResourceId:    strconv.FormatInt(capitalRaisingEntity.Id, 10),
				FileSize:      fileHeader.Size,
				FileExtension: s3utils.GetExtension(fileHeader.Filename),
				FilePath:      s3utils.MakeS3Path(domain.CAPITAL_RAISING_ENTITY),
				FileUrl:       "",
			},
			fileHeader,
		)
		if err != nil {
			return factory.ErrorAWSUpload.New(err)
		}

		capitalRaisingEntity.ImageUrl = &imageUrl

		// image_url 更新
		if err := capitalRaisingEntity.Update(c.Request().Context()); err != nil {
			return factory.ErrorDB.New(err)
		}
	}

	return apirender.RenderSucc(c, http.StatusOK, CapitalRaisingEntityAdminResponseDto{}.NewCapitalRaisingEntityAdminResponseDto(capitalRaisingEntity))
}

func (cre CapitalRaisingEntityApiController) GetOne(c echo.Context) error {
	var err error
	_, _, err = jwthelper.TokenToUser(c)
	if err != nil {
		return err
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}
	capitalRaisingEntity, err := domain.CapitalRaisingEntity{}.GetById(c.Request().Context(), id)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	if capitalRaisingEntity == nil {
		return factory.ErrorNotFound.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, CapitalRaisingEntityAdminResponseDto{}.NewCapitalRaisingEntityAdminResponseDto(capitalRaisingEntity))
}

func (cre CapitalRaisingEntityApiController) GetAll(c echo.Context) error {
	var v CapitalRaisingEntitySearchInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}

	uid, uemail, err := jwthelper.TokenToUser(c)
	if err != nil {
		return err
	}

	if v.MaxResultCount == 0 {
		v.MaxResultCount = DefaultMaxResultCount
	}

	// console log
	factory.Logger(c.Request().Context()).WithFields(logrus.Fields{
		"user":             uid,
		"email":            uemail,
		"sortby":           v.Sortby,
		"order":            v.Order,
		"max_result_count": v.MaxResultCount,
		"skip_count":       v.SkipCount,
	}).Info("SearchStart")

	totalCount, items, err := domain.CapitalRaisingEntity{}.GetAll(
		c.Request().Context(),
		v.Sortby,
		v.Order,
		v.SkipCount,
		v.MaxResultCount,
	)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, apirender.ArrayResult{
		TotalCount: totalCount,
		Items:      items,
	})
}

func (cre CapitalRaisingEntityApiController) Update(c echo.Context) error {
	// 認証チェック
	_, _, err := jwthelper.TokenToUser(c)
	if err != nil {
		return err
	}

	// パスパラメータ id
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	// 既存データ取得
	capitalRaisingEntity, err := domain.CapitalRaisingEntity{}.GetById(c.Request().Context(), id)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	if capitalRaisingEntity == nil {
		return factory.ErrorNotFound.New(nil)
	}

	// 入力 bind
	var input CapitalRaisingEntityInput
	if err := c.Bind(&input); err != nil {
		return factory.ErrorParameter.New(err)
	}

	if err := c.Validate(&input); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// 更新内容を反映
	capitalRaisingEntity.CorporateName = input.CorporateName
	capitalRaisingEntity.ZipCode = input.ZipCode
	capitalRaisingEntity.Province = input.Province
	capitalRaisingEntity.City = input.City
	capitalRaisingEntity.Address = input.Address
	capitalRaisingEntity.Address2 = input.Address2
	capitalRaisingEntity.RepresentativeFamilyName = input.RepresentativeFamilyName
	capitalRaisingEntity.RepresentativeFirstName = input.RepresentativeFirstName
	capitalRaisingEntity.RepresentativeFamilyNameKana = input.RepresentativeFamilyNameKana
	capitalRaisingEntity.RepresentativeFirstNameKana = input.RepresentativeFirstNameKana

	// ---- 画像アップロード処理
	file, fileHeader, err := c.Request().FormFile("image_file")
	if err == nil && fileHeader != nil {
		defer file.Close()

		imageUrl, err := s3utils.UploadAttachmentFile(
			c.Request().Context(),
			s3utils.BUCKET_PUBLIC_NAME,
			&domain.Attachment{
				FileName:      s3utils.MakeS3FileName(),
				OriginalName:  fileHeader.Filename,
				ResourceId:    strconv.FormatInt(capitalRaisingEntity.Id, 10),
				FileSize:      fileHeader.Size,
				FileExtension: s3utils.GetExtension(fileHeader.Filename),
				FilePath:      s3utils.MakeS3Path(domain.CAPITAL_RAISING_ENTITY),
				FileUrl:       "",
			},
			fileHeader,
		)
		if err != nil {
			return factory.ErrorAWSUpload.New(err)
		}

		capitalRaisingEntity.ImageUrl = &imageUrl
	}

	// 更新実行
	if err := capitalRaisingEntity.Update(c.Request().Context()); err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(
		c,
		http.StatusOK,
		CapitalRaisingEntityAdminResponseDto{}.NewCapitalRaisingEntityAdminResponseDto(capitalRaisingEntity),
	)
}
