package http

import (
	"net/http"
	"testing"

	adomain "github.com/illuminarean/energy-share-core-api/domain"

	pkgtest "github.com/illuminarean/energy-share-core-api/delivery/testing"
	"github.com/illuminarean/energy-share-core-api/domain/mocks"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/validator"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_fileHttpDelivery_CreatePreSignedUrls(t *testing.T) {
	tests := []struct {
		name        string
		echoCtx     func(e *echo.Echo) echo.Context
		fileUsecase func() *mocks.FileUsecase
		wantErr     bool
	}{
		{
			name: "POST /api/v1/files/pre-signed-urls 정상적으로 호출",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/users/pre-signed-urls",
				Method: http.MethodPost,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Body: `{
					"files" : [
						{
							"file_name" : "attachment.jpg",
							"file_type" : "attachment"
						},
						{
							"file_name" : "image.jpg",
							"file_type" : "content_image"
						}
					]
				}`,
				Claims: map[string]interface{}{
					"user":  "1",
					"email": "admin@energyshares.jp",
					"name":  "Admin",
					"role":  []string{"ROLE_SUPER_ADMIN", "ROLE_ADMIN", "ROLE_OPERATOR"},
				},
			}),
			fileUsecase: func() *mocks.FileUsecase {
				usecase := &mocks.FileUsecase{}
				usecase.On("CreatePreSignedUrls", mock.Anything, adomain.FileCreatePreSignedUrlsRequest{
					Files: []adomain.PreSignedFileRequest{
						{
							FileName: "attachment.jpg",
							FileType: "attachment",
						},
						{
							FileName: "image.jpg",
							FileType: "content_image",
						},
					},
				}).Return(adomain.FileCreatePreSignedUrlsResponse{}, nil)
				return usecase
			},
			wantErr: false,
		},
		{
			name: "POST /api/v1/files/pre-signed-urls 호출시 사전에 정의되지 않는 파일타입으로 요청 (file_type:faq)",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/api/v1/users/pre-signed-urls",
				Method: http.MethodPost,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Body: `{
					"files" : [
						{
							"file_name" : "attachment.jpg",
							"file_type" : "faq"
						},
					]
				}`,
				Claims: map[string]interface{}{
					"user":  "1",
					"email": "admin@energyshares.jp",
					"name":  "Admin",
					"role":  []string{"ROLE_SUPER_ADMIN", "ROLE_ADMIN", "ROLE_OPERATOR"},
				},
			}),
			fileUsecase: func() *mocks.FileUsecase {
				usecase := &mocks.FileUsecase{}
				return usecase
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := fileHttpDelivery{
				fileUsecase: tt.fileUsecase(),
				extractor:   jwthelper.NewUserExtractor(),
			}

			e := echo.New()
			e.Validator = validator.NewCustomValidator()

			err := delivery.CreatePreSignedUrls(tt.echoCtx(e))
			assert.Equalf(t, tt.wantErr, err != nil, "CreatePreSignedUrls() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}
