package s3utils

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudfront/sign"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"
	"github.com/rs/xid"
	"github.com/sirupsen/logrus"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
)

type S3 interface {
	CreateSignedUrl(ctx context.Context, url string) (string, error)
	CreatePreSignedUrl(ctx context.Context, request CreatePreSignedUrlRequest) (string, error)
	CreateCookieSign() ([]*http.Cookie, error)
	GetHeadObject(bucket string, key string) (S3ObjectInfo, error)
	DeleteObject(bucket string, key string) error
}

type s3pkg struct {
	cfg     *config.Config
	session *s3.S3
}

type CreatePreSignedUrlRequest struct {
	BucketName         string
	FilePath           string
	FileName           string
	ContentType        *string
	ContentDisposition *string
}

func New(cfg *config.Config) S3 {
	awsCfg := &aws.Config{
		Region: aws.String(cfg.AWS.Region),
	}
	if cfg.IsLocal() && strings.TrimSpace(cfg.AWS.BASE_ENDPOINT) != "" {
		awsCfg.Endpoint = aws.String(cfg.AWS.BASE_ENDPOINT)
		awsCfg.S3ForcePathStyle = aws.Bool(true)
	}

	sess, err := session.NewSession(awsCfg)
	if err != nil {
		panic(err)
	}

	return &s3pkg{
		cfg:     cfg,
		session: s3.New(sess),
	}
}

func (s s3pkg) CreateSignedUrl(ctx context.Context, url string) (string, error) {
	privateKey, err := loadCloudFrontPrivateKey()
	if err != nil {
		return "", errors.Wrap(err, "err load private key")
	}

	urlSigner := sign.NewURLSigner(s.cfg.AWS.PUBLIC_KEY_ID, privateKey)

	signedUrl, err := urlSigner.Sign(url, time.Now().Add(EXPIRE_TIME*time.Minute).UTC())
	if err != nil {
		return "", errors.Wrap(err, "err sign url")
	}

	return signedUrl, nil
}

func (s s3pkg) CreatePreSignedUrl(ctx context.Context, request CreatePreSignedUrlRequest) (string, error) {
	extension := path.Ext(request.FileName)

	fileName := xid.New().String() + extension

	req, _ := s.session.PutObjectRequest(&s3.PutObjectInput{
		Bucket:             aws.String(request.BucketName),
		Key:                aws.String(fmt.Sprintf("%s/%s", request.FilePath, fileName)),
		ContentType:        request.ContentType,
		ContentDisposition: request.ContentDisposition,
	})

	urlString, err := req.Presign(30 * time.Minute)
	if err != nil {
		return "", errors.Wrap(err, "err presign url")
	}

	return urlString, nil
}

func (s s3pkg) CreateCookieSign() ([]*http.Cookie, error) {
	if s.cfg.IsLocal() {
		// ローカル環境では生成をスキップする
		return []*http.Cookie{}, nil
	}

	privKey, err := loadCloudFrontPrivateKey()
	if err != nil {
		return nil, err
	}
	signer := sign.NewCookieSigner(s.cfg.AWS.PUBLIC_KEY_ID, privKey)
	cookies, err := signer.Sign("https://"+serviceutils.GetCDN()+"/*", time.Now().Add(EXPIRE_TIME*time.Minute), func(o *sign.CookieOptions) {
		o.Path = "/"
		o.Domain = "." + s.cfg.Domain
	})

	if err != nil {
		return nil, errors.Wrap(err, "err create signed cookies")
	}

	return cookies, nil
}

type S3ObjectInfo struct {
	FileSize         int64
	OriginalFileName string
}

// GetHeadObject はS3からオブジェクトのメタデータを取得します。
// 使用する際、必要に応じて追加的なフィールドを追加してください。
func (s s3pkg) GetHeadObject(bucket, key string) (S3ObjectInfo, error) {
	input := &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	output, err := s.session.HeadObject(input)
	if err != nil {
		return S3ObjectInfo{}, errors.Wrap(err, "err head object")
	}

	contentDisposition := *output.ContentDisposition
	if contentDisposition == "" {
		return S3ObjectInfo{}, errors.New("content-disposition is empty")
	}
	filename, err := extractFilenameFromContentDisposition(contentDisposition)
	// 以前にアップロードされたファイルを処理するための下位互換性関連のコード
	// content-dispositionからファイル名を抽出できない場合
	if err != nil {
		logrus.Errorf("failed to extract filename from content-disposition: %s, error: %v", contentDisposition, err)
	}

	result := S3ObjectInfo{
		FileSize:         *output.ContentLength,
		OriginalFileName: filename,
	}
	return result, nil
}

func (s s3pkg) DeleteObject(bucket string, key string) error {
	input := s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	if _, err := s.session.DeleteObject(&input); err != nil {
		return errors.Wrap(err, "err delete object")
	}

	return nil
}

func extractFilenameFromContentDisposition(contentDisposition string) (string, error) {

	if contentDisposition == "" {
		return "", errors.New("content-disposition is empty")
	}

	// filenameパラメータの抽出
	for _, part := range strings.Split(contentDisposition, ";") {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "filename=") || strings.HasPrefix(part, "filename*=") {
			fileName := strings.TrimPrefix(part, "filename=")
			fileName = strings.TrimPrefix(fileName, "filename*=")
			fileName = strings.Trim(fileName, "\"")
			decodedFileName, err := url.QueryUnescape(fileName)
			if err != nil {
				return "", errors.Wrapf(err, "failed to decode filename from contentDisposition '%s'", contentDisposition)
			}
			return decodedFileName, nil
		}
	}

	return "", fmt.Errorf("filename parameter not found in content-disposition: %s", contentDisposition)
}
