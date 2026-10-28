package http

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/apirender"

	openapigo "github.com/swaggest/openapi-go"

	adomain "github.com/illuminarean/energy-share-core-api/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/echoutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
)

type fileHttpDelivery struct {
	fileUsecase adomain.FileUsecase
	extractor   jwthelper.UserExtractor
}

func New(
	api *echoutils.EchoSwagger,
	fileUsecase adomain.FileUsecase,
	extractor jwthelper.UserExtractor,
) *fileHttpDelivery {
	handler := &fileHttpDelivery{
		fileUsecase: fileUsecase,
		extractor:   extractor,
	}

	g := api.Group("/api/v1/files")

	g.POST("/pre-signed-urls", handler.CreatePreSignedUrls,
		echoutils.AllowRoles(domain.ROLE_OPERATOR)).
		SetTags("파일").
		SetSummary("서명된 URL을 생성합니다.").
		AddRequest(adomain.FileCreatePreSignedUrlsRequest{}).
		AddResponse(adomain.FileCreatePreSignedUrlsResponse{}, func(cu *openapigo.ContentUnit) {
			cu.HTTPStatus = http.StatusOK
			cu.Description = "서명된 URL 응답값"
		}).
		AddErrors(
			factory.ErrorFailAuthorized,
			factory.ErrorParameter,
			factory.ErrorDB,
		).
		AddSecurity("Authorization")

	return handler
}

func (f fileHttpDelivery) CreatePreSignedUrls(c echo.Context) error {
	if _, err := f.extractor.Extract(c); err != nil {
		return factory.ErrorFailAuthorized.New(err)
	}

	ctx := c.Request().Context()

	// Request DTO 생성
	var request = adomain.FileCreatePreSignedUrlsRequest{}

	// Request DTO 바인딩
	if err := c.Bind(&request); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// Request DTO 유효성 검증
	if err := request.Validate(); err != nil {
		return factory.ErrorParameter.New(err)
	}

	response, err := f.fileUsecase.CreatePreSignedUrls(ctx, request)

	switch {
	case err == nil:
		return apirender.RenderSucc(c, http.StatusOK, response)
	default:
		return factory.ErrorDB.New(err)
	}
}
