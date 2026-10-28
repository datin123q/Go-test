package controllers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/apirender"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/echoutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/s3utils"
)

type StorageApiController struct {
}

func NewStorageApiController() *StorageApiController {
	return &StorageApiController{}
}

func (s *StorageApiController) Init(g echoswagger.ApiGroup) {

	g.POST("/upload", s.Upload,
		echoutils.AllowRoles(domain.ROLE_ADMIN)).
		AddParamFile("content_file", "一時画像ファイル", true).
		AddResponse(http.StatusCreated, "一時画像が登録されました。", nil, nil).
		SetDescription("コンテンツ画像をS3一時ストレージに保存します。").
		SetSummary("一時ストレージに画像を保存します。").
		SetSecurity("Authorization")
}

const TEMP_FILE = "temp_file"

type StorageObject struct {
	FileName      string
	FilePath      string
	FileType      string
	FileExtension string
	FileUrl       string
}

func (s StorageObject) GetKey() string {
	return s.FilePath + s.FileName + s.FileExtension
}

func (s StorageObject) GetFileType() string {
	return s.FileType
}

func (s StorageObject) GetFileName() string {
	return s.FileName
}

func (s *StorageObject) SetLocation(loc string) {
	s.FileUrl = loc
}

func (s StorageApiController) Upload(c echo.Context) error {

	file, err := c.FormFile("content_file")
	if err != nil {
		return factory.ErrorInvalidForm.New(err)
	}
	obj := &StorageObject{
		FileName:      s3utils.MakeS3FileName(),
		FilePath:      s3utils.MakeS3Path(TEMP_FILE),
		FileExtension: s3utils.GetExtension(file.Filename),
		FileUrl:       "",
		FileType:      TEMP_FILE,
	}

	url, err := s3utils.UploadShowFile(c.Request().Context(), s3utils.BUCKET_PUBLIC_NAME, obj, file)
	if err != nil {
		return factory.ErrorAWSUpload.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, url)
}

func (s StorageObject) GetCDNUrl() string {
	return s3utils.MakeCDNURL(s.GetKey())
}

func (s StorageObject) GetCDNSignUrl() string {
	return s3utils.MakeCDNSignURL(s.GetKey())
}
