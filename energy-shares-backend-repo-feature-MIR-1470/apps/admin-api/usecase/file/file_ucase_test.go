package file

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/filehelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	adomain "github.com/illuminarean/energy-share-core-api/domain"
	"github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain/mocks"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/s3utils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
)

func Test_fileUsecase_CreatePreSignedUrls(t *testing.T) {
	timeNowDateOnly := time.Now().Format(time.DateOnly)

	type args struct {
		ctx     context.Context
		request adomain.FileCreatePreSignedUrlsRequest
	}

	tests := []struct {
		name    string
		args    args
		s3pkg   func() *mocks.S3
		want    adomain.FileCreatePreSignedUrlsResponse
		wantErr bool
	}{
		{
			name: "正常に呼び出し",
			args: args{
				ctx: context.Background(),
				request: adomain.FileCreatePreSignedUrlsRequest{
					Files: []adomain.PreSignedFileRequest{
						{
							FileName: "attachment.jpg",
							FileType: adomain.FileTypeAttachment,
						},
						{
							FileName: "image.jpg",
							FileType: adomain.FileTypeContentImage,
						},
					},
				},
			},

			s3pkg: func() *mocks.S3 {
				s3pkg := new(mocks.S3)
				s3pkg.On("CreatePreSignedUrl", context.Background(), mock.MatchedBy(func(req s3utils.CreatePreSignedUrlRequest) bool {
					return req.BucketName == "test-bucket" &&
						req.FilePath == fmt.Sprintf("%s/%s", "attachment_file", timeNowDateOnly) &&
						req.FileName == "attachment.jpg" &&
						req.ContentType != nil && *req.ContentType == "image/jpeg" &&
						req.ContentDisposition != nil && *req.ContentDisposition == "attachment; filename=\"attachment.jpg\"; filename*=UTF-8''attachment.jpg"
				})).Return(fmt.Sprintf("https://test-bucket/attachment_file/%s/attachment.jpg?abcd=1", timeNowDateOnly), nil)

				s3pkg.On("CreatePreSignedUrl", context.Background(), mock.MatchedBy(func(req s3utils.CreatePreSignedUrlRequest) bool {
					return req.BucketName == "test-bucket" &&
						req.FilePath == fmt.Sprintf("%s/%s", "content_img_file", timeNowDateOnly) &&
						req.FileName == "image.jpg" &&
						req.ContentType != nil && *req.ContentType == "image/jpeg" &&
						req.ContentDisposition != nil && *req.ContentDisposition == "attachment; filename=\"image.jpg\"; filename*=UTF-8''image.jpg"
				})).Return(fmt.Sprintf("https://test-bucket/content_img_file/%s/image.jpg?abcd=1", timeNowDateOnly), nil)

				return s3pkg
			},
			want: adomain.FileCreatePreSignedUrlsResponse{
				Files: []adomain.PreSignedFileResponse{
					{
						PreSignedUrl:       fmt.Sprintf("https://test-bucket/attachment_file/%s/attachment.jpg?abcd=1", timeNowDateOnly),
						CdnUrl:             fmt.Sprintf("https://eshares.cdn.jp/attachment_file/%s/attachment.jpg", timeNowDateOnly),
						ContentType:        "image/jpeg",
						ContentDisposition: "attachment; filename=\"attachment.jpg\"; filename*=UTF-8''attachment.jpg",
					},
					{
						PreSignedUrl:       fmt.Sprintf("https://test-bucket/content_img_file/%s/image.jpg?abcd=1", timeNowDateOnly),
						CdnUrl:             fmt.Sprintf("https://eshares.cdn.jp/content_img_file/%s/image.jpg", timeNowDateOnly),
						ContentType:        "image/jpeg",
						ContentDisposition: "attachment; filename=\"image.jpg\"; filename*=UTF-8''image.jpg",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// CDN URLはserviceutilsを経由して参照されるため、テストで明示的に設定する
			serviceutils.SetCDN("eshares.cdn.jp")
			t.Cleanup(func() {
				serviceutils.SetCDN("")
			})

			usecase := fileUsecase{
				cfg: &config.Config{
					CommonConfig: config.CommonConfig{
						AWS: struct {
							Region                  string
							SES_Region              string
							SNS_Region              string
							CDN                     string
							S3_PRIVATE_BUCKET       string
							S3_PUBLIC_BUCKET        string
							S3_EKYC_IMAGES_BUCKET   string
							SES_ARN                 string
							SNS_ARN                 string
							PUBLIC_KEY_ID           string
							PRIVATE_KEY_PATH        string
							BASE_ENDPOINT           string
							INVEST_WORKER_QUEUE_URL string
						}{
							S3_PUBLIC_BUCKET: "test-bucket",
							CDN:              "eshares.cdn.jp",
						},
					},
				},
				fileHelper: filehelper.NewFileHelper(),
				s3:         tt.s3pkg(),
			}

			got, err := usecase.CreatePreSignedUrls(tt.args.ctx, tt.args.request)
			ignoreTime := cmpopts.IgnoreTypes(jsonutils.JSONTime{})

			assert.Equalf(t, tt.wantErr, err != nil, "CreatePreSignedUrls() error = %v, wantErr %v", err, tt.wantErr)
			assert.Truef(t, cmp.Equal(tt.want, got, ignoreTime), "CreatePreSignedUrls() = %v, want %v", got, tt.want)
		})
	}
}
