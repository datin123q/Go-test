package domain

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

const (
	filePathAttachmentFile = "attachment_file"
	filePathContentImage   = "content_img_file"
)

type FileUsecase interface {
	CreatePreSignedUrls(ctx context.Context, request FileCreatePreSignedUrlsRequest) (FileCreatePreSignedUrlsResponse, error)
}

type FileType string

func (f FileType) Path() string {
	switch f {
	case FileTypeAttachment:
		return filePathAttachmentFile
	case FileTypeContentImage:
		return filePathContentImage
	default:
		return ""
	}
}

func (f FileType) Bucket(cfg *config.Config) string {
	switch f {
	case FileTypeAttachment:
		return cfg.AWS.S3_PUBLIC_BUCKET
	case FileTypeContentImage:
		return cfg.AWS.S3_PUBLIC_BUCKET
	default:
		return ""
	}
}

func (f FileType) Valid() error {
	switch f {
	case FileTypeAttachment, FileTypeContentImage:
		return nil
	default:
		return fmt.Errorf("invalid file type: %s", f)
	}
}

const (
	FileTypeAttachment   FileType = "attachment"
	FileTypeContentImage FileType = "content_image"
)

type FileCreatePreSignedUrlsRequest struct {
	Files []PreSignedFileRequest `json:"files" description:"파일 목록"`
}

func (f FileCreatePreSignedUrlsRequest) Validate() error {
	validate := validator.New()
	if err := validate.Struct(f); err != nil {
		return factory.ErrorParameter.New(err)
	}

	for _, file := range f.Files {
		if err := file.FileType.Valid(); err != nil {
			return factory.ErrorParameter.New(err)
		}
	}

	return nil
}

type PreSignedFileRequest struct {
	FileName string   `json:"file_name" validate:"required" required:"true" description:"파일명" example:"test.jpg"`
	FileType FileType `json:"file_type" validate:"required" required:"true" description:"### 해당하는 값 외에 요청되면 에러가 발생합니다.\n파일타입\n* attachment(첨부파일)\n* content_image(컨텐츠 이미지)" enum:"attachment,content_image" example:"attachment"`
}

type FileCreatePreSignedUrlsResponse struct {
	Files []PreSignedFileResponse `json:"files"`
}

type PreSignedFileResponse struct {
	PreSignedUrl       string             `json:"pre_signed_url" description:"서명된 URL" example:"https://es-jp-devv2-public-file-storage.s3.ap-northeast-2.amazonaws.com/attachment_file/2024-09-09/crfbl6cldh5hujq0ulq0.jpg?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=ASIAX3VRTD32MJW2LNN4%2F20240909%2Fap-northeast-2%2Fs3%2Faws4_request&X-Amz-Date=20240909T090633Z&X-Amz-Expires=1800&X-Amz-Security-Token=IQoJb3JpZ2luX2VjEDEaDmFwLW5vcnRoZWFzdC0yIkYwRAIgJ6IYFoha%2F5rncjLWPtfaMIRuX06TrV%2BWFx3zpXSYqOECIGLiFJ9yG9e%2BKTkpnYg3bxQbCENRKDILaxC%2F5TPVTpv1Ku4BCFoQBRoMNTQwNDY0NTg2NDg0Igy15W%2F6hkjthh13v%2F4qywHWmD2sop8LWdE9Ln3myQxJVhOyQNJUvDAGkw%2BUKfcIN%2FBXDcqspAXx6VoUmR7GwHQh60WteRBtirqhRh8ZxSOgCoHnrw48hhCFg%2BiaOkipop6EHRe6d6%2B6UGmie4QdZvjf4R4WWkiI31Qqjo84FQyVrNRPHK5Ig%2FVrnQqRqvSbRblVZCGYB83S6F4qNcHxBDIjOOTweAXxEZpPoO5PLS8dAJq4E5nuMlfA3QdRr2JadvARmel0YUp%2B0LXTo33pkDwZNAy3OABoBkOaEDC79Pq2BjqZATp5J2nnJUkj574d1r7Bt6QpgVOlDD39CNm%2BLRwTbfD95T1lYXD6%2F%2BW6%2F%2F05QaRRWuTqHhrF3st3qy1zssEA%2FdBry%2FIbJqAO1KB9jke9i274zEt1vh%2FFuxn0MevBmYdATloB31cFfK0JcewBjceZPs6ntVH06YPHkItb77e3R3AqcIV1AxSWElTMWCiUUaPlWXYfzERGsXklZA%3D%3D&X-Amz-SignedHeaders=content-disposition%3Bcontent-type%3Bhost&X-Amz-Signature=92c51d623d8079b4c6748693ae3e1131e7e811e3d0f29de4eb3c7855d1ceab99"`
	CdnUrl             string             `json:"cdn_url" description:"CDN URL" example:"https://cdn.dev.energyshares.jp/attachment_file/2024-09-09/crfbl6cldh5hujq0ulq0.jpg"`
	ExpiredAt          jsonutils.JSONTime `json:"expired_at" description:"만료일시" example:"2024-09-09 18:36:33"`
	ContentType        string             `json:"content_type" description:"컨텐츠 타입" example:"image/jpeg"`
	ContentDisposition string             `json:"content_disposition" description:"컨텐츠 표시" example:"attachment; filename=\"test.jpg\""`
}
