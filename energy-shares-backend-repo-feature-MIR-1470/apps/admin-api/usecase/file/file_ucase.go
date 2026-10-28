package file

import (
	"context"
	"fmt"
	"mime"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/filehelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/pkg/errors"

	"github.com/aws/aws-sdk-go/aws"

	adomain "github.com/illuminarean/energy-share-core-api/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/s3utils"
)

type fileUsecase struct {
	cfg        *config.Config
	s3         s3utils.S3
	fileHelper filehelper.FileHelper
}

func New(
	cfg *config.Config,
	s3 s3utils.S3,
	fileHelper filehelper.FileHelper,
) adomain.FileUsecase {
	return &fileUsecase{
		cfg:        cfg,
		s3:         s3,
		fileHelper: fileHelper,
	}
}

func (f fileUsecase) CreatePreSignedUrls(ctx context.Context, request adomain.FileCreatePreSignedUrlsRequest) (adomain.FileCreatePreSignedUrlsResponse, error) {
	var preSignedFiles []adomain.PreSignedFileResponse

	now := time.Now()

	for _, file := range request.Files {
		filePath := file.FileType.Path() + "/" + now.Format(time.DateOnly)

		escapedFilename := url.QueryEscape(file.FileName)
		escapedFilename = strings.ReplaceAll(escapedFilename, "+", " ")
		contentType := mime.TypeByExtension(path.Ext(file.FileName))

		if contentType == "" {
			contentType = "application/octet-stream"
		}

		contentDisposition := fmt.Sprintf("attachment; filename=\"%s\"; filename*=UTF-8''%s", escapedFilename, escapedFilename)

		urlString, err := f.s3.CreatePreSignedUrl(ctx, s3utils.CreatePreSignedUrlRequest{
			BucketName:         file.FileType.Bucket(f.cfg),
			FilePath:           filePath,
			FileName:           file.FileName,
			ContentType:        aws.String(contentType),
			ContentDisposition: aws.String(contentDisposition),
		})
		if err != nil {
			return adomain.FileCreatePreSignedUrlsResponse{}, err
		}

		pathFromURL, err := f.fileHelper.ExtractFilePath(urlString)
		if err != nil {
			return adomain.FileCreatePreSignedUrlsResponse{}, errors.Wrap(err, "err in url.Parse")
		}
		pathFromURL = filehelper.NormalizeS3FilePath(file.FileType.Bucket(f.cfg), pathFromURL)

		preSignedFiles = append(preSignedFiles, adomain.PreSignedFileResponse{
			PreSignedUrl:       urlString,
			CdnUrl:             s3utils.MakeCDNURL(pathFromURL),
			ExpiredAt:          jsonutils.JSONTime(now.Add(time.Minute * 30)),
			ContentType:        contentType,
			ContentDisposition: contentDisposition,
		})
	}

	return adomain.FileCreatePreSignedUrlsResponse{
		Files: preSignedFiles,
	}, nil
}
