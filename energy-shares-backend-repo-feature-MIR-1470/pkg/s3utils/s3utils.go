package s3utils

import (
	"bytes"
	"context"
	"crypto/rsa"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront/sign"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/rs/xid"
	"github.com/sirupsen/logrus"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
)

var (
	BUCKET_PRIVATE_NAME     string
	BUCKET_PUBLIC_NAME      string
	BUCKET_EKYC_IMAGES_NAME string
)

const EXPIRE_TIME = 5

func init() {
	BUCKET_PRIVATE_NAME = "private-energyshares-file-storage"
	BUCKET_PUBLIC_NAME = "public-energyshares-file-storage"
	BUCKET_EKYC_IMAGES_NAME = "ekyc-images"
}

type S3Object interface {
	GetKey() string
	GetFileType() string
	GetFileName() string
	SetLocation(loc string)
	GetCDNUrl() string
	GetCDNSignUrl() string
}

func SetS3PrivateBucket(bucketName string) {
	BUCKET_PRIVATE_NAME = bucketName
}

func SetS3PublicBucket(bucketName string) {
	BUCKET_PUBLIC_NAME = bucketName
}

func SetS3EkycImagesBucket(bucketName string) {
	BUCKET_EKYC_IMAGES_NAME = bucketName
}

// BucketNames は設定済みのバケット名を返す。
// パス形式(S3ForcePathStyle)の URL からバケット名を取り除く判定に使う。
func BucketNames() []string {
	return []string{BUCKET_PRIVATE_NAME, BUCKET_PUBLIC_NAME, BUCKET_EKYC_IMAGES_NAME}
}

func InitializeS3Bucket(s3privateBucket, s3publicBucket string) {
	SetS3PrivateBucket(s3privateBucket)
	SetS3PublicBucket(s3publicBucket)
}

func InitializeS3BucketWithEkyc(s3privateBucket, s3publicBucket, s3ekycImagesBucket string) {
	SetS3PrivateBucket(s3privateBucket)
	SetS3PublicBucket(s3publicBucket)
	SetS3EkycImagesBucket(s3ekycImagesBucket)
}

// Upload S3
func DownloaderS3File(ctx context.Context, bucket string, s3obj S3Object) (file *os.File, err error) {
	f, err := os.Create(s3obj.GetFileName())
	if err != nil {
		return nil, fmt.Errorf("failed to create file %q, %v", s3obj.GetFileName(), err)
	}

	_, err = factory.S3Downloader(ctx).Download(f, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(s3obj.GetKey()),
	})
	if err != nil {
		return nil, err
	}
	return f, nil
}

// Upload S3
func UploadAttachmentFile(ctx context.Context, bucket string, s3obj S3Object, f *multipart.FileHeader) (path string, err error) {
	var ct string
	src, err := f.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	urlEncode := strings.ReplaceAll(url.QueryEscape(s3obj.GetFileName()), "+", " ")
	ct = getContentType(s3obj, ct)

	_, err = factory.S3Uploader(ctx).Upload(&s3manager.UploadInput{
		ACL:                aws.String(s3.BucketCannedACLPrivate),
		Body:               src,
		Bucket:             aws.String(bucket),
		Key:                aws.String(s3obj.GetKey()),
		ContentDisposition: aws.String("attachment; filename*=utf-8''" + urlEncode),
		ContentType:        aws.String(ct),
	})
	if err != nil {
		return "", err
	}

	s3obj.SetLocation(s3obj.GetCDNUrl())
	return s3obj.GetCDNUrl(), err
}

// S3メモリベースのアップロード
func UploadBuffFile(ctx context.Context, bucket string, s3obj S3Object, buf *bytes.Buffer) (path string, err error) {
	var ct string
	urlEncode := strings.ReplaceAll(url.QueryEscape(s3obj.GetFileName()), "+", " ")

	ct = getContentType(s3obj, ct)
	_, err = factory.S3Uploader(ctx).Upload(&s3manager.UploadInput{
		ACL:                aws.String(s3.BucketCannedACLPrivate),
		Body:               bytes.NewReader(buf.Bytes()),
		Bucket:             aws.String(bucket),
		Key:                aws.String(s3obj.GetKey()),
		ContentDisposition: aws.String("inline; filename*=utf-8''" + urlEncode),
		ContentType:        aws.String(ct),
	})
	if err != nil {
		return "", err
	}
	s3obj.SetLocation(s3obj.GetCDNUrl())
	return s3obj.GetCDNUrl(), err
}

func getContentType(s3obj S3Object, ct string) string {
	switch GetExtension(s3obj.GetFileName()) {
	case ".png":
		ct = "image/png"
	case ".pdf":
		ct = "application/pdf"
	case ".jpg", ".jpeg":
		ct = "image/jpeg"
	case ".gif":
		ct = "image/gif"
	default:
		ct = "application/octet-stream"
	}
	return ct
}

func UploadShowFile(ctx context.Context, bucket string, s3obj S3Object, f *multipart.FileHeader) (path string, err error) {
	var ct string
	src, err := f.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()
	// png, jpg, jpeg, pdf
	switch GetExtension(s3obj.GetFileName()) {
	case ".png":
		ct = "image/png"
	case ".pdf":
		ct = "application/pdf"
	case ".jpg", ".jpeg":
		ct = "image/jpeg"
	default:
		ct = "application/octet-stream"
	}
	urlEncode := strings.ReplaceAll(url.QueryEscape(s3obj.GetFileName()), "+", " ")
	_, err = factory.S3Uploader(ctx).Upload(&s3manager.UploadInput{
		ACL:                aws.String(s3.BucketCannedACLPrivate),
		Body:               src,
		Bucket:             aws.String(bucket),
		Key:                aws.String(s3obj.GetKey()),
		ContentDisposition: aws.String("inline; filename*=utf-8''" + urlEncode),
		ContentType:        aws.String(ct),
	})
	if err != nil {
		return "", err
	}
	s3obj.SetLocation(s3obj.GetCDNUrl())
	return s3obj.GetCDNUrl(), err
}

func DeleteFile(ctx context.Context, bucket string, s3obj S3Object) (err error) {
	_, err = factory.S3(ctx).DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(s3obj.GetKey()),
	})
	return err
}

func CopyFile(ctx context.Context, targetBucket, targetKey, sourceBucket, sourceItem string) (string, error) {
	source := sourceBucket + sourceItem

	out, err := factory.S3(ctx).CopyObject(&s3.CopyObjectInput{
		ACL:        aws.String(s3.BucketCannedACLPrivate),
		Bucket:     aws.String(targetBucket),
		CopySource: aws.String(url.PathEscape(source)),
		Key:        aws.String(targetKey),
	})

	// Wait to see if the item got copied
	// err = svc.WaitUntilObjectExists(&s3.HeadObjectInput{Bucket: aws.String(other), Key: aws.String(item)})
	// if err != nil {
	//	exitErrorf("Error occurred while waiting for item %q to be copied to bucket %q, %v", bucket, item, other, err)
	// }

	return out.String(), err
}

func MakeS3FileName() string {
	gen := xid.New()
	return gen.String()
}

func MakeS3Path(fileType string) (filePath string) {
	t := time.Now()
	return "/" + fileType + "/" + t.Format(time.DateOnly) + "/"
}

func GetExtension(fName string) (result string) {
	return path.Ext(fName)
}

func MakeCDNURL(key string) string {
	if serviceutils.EnvIsLocal() {
		cdn := serviceutils.GetCDN()
		if cdn == "" {
			return key
		}

		cleanKey := strings.TrimPrefix(key, "/")
		cleanBase := strings.TrimSuffix(cdn, "/")

		if strings.HasPrefix(cdn, "http://") || strings.HasPrefix(cdn, "https://") {
			return cleanBase + "/" + cleanKey
		}
		return "https://" + cleanBase + "/" + cleanKey
	}

	return "https://" + serviceutils.GetCDN() + key
}

func MakeCDNSignURL(key string) string {
	if serviceutils.EnvIsLocal() {
		return MakeCDNURL(key)
	}
	return GetUrlSign(MakeCDNURL(key))
}

func GetCookieSign() ([]*http.Cookie, error) {
	if serviceutils.EnvIsLocal() {
		// ローカル環境ではクッキーを使用しない
		return []*http.Cookie{}, nil
	}

	privKey, err := loadCloudFrontPrivateKey()
	if err != nil {
		return nil, err
	}
	s := sign.NewCookieSigner(serviceutils.GetPublicKeyId(), privKey)
	cookies, err := s.Sign("https://"+serviceutils.GetCDN()+"/*", time.Now().Add(EXPIRE_TIME*time.Minute), func(o *sign.CookieOptions) {
		o.Path = "/"
		o.Domain = "." + serviceutils.GetDomain()
	})
	// policy := &sign.Policy{
	//	Statements: []sign.Statement{
	//		{
	//			// Read the provided documentation on how to set this
	//			// correctly, you'll probably want to use wildcards.
	//			Resource: serviceutils.GetCDN() + "/*",
	//			Condition: sign.Condition{
	//				// Optional ClientIp source address range
	//				IPAddress: &sign.IPAddress{SourceIP: "0.0.0,0"},
	//				// Optional date URL is not valid until
	//				DateGreaterThan: &sign.AWSEpochTime{time.Now().Add(30 * time.Minute)},
	//				// Required date the URL will expire after
	//				DateLessThan: &sign.AWSEpochTime{time.Now().Add(1 * time.Hour)},
	//			},
	//		},
	//	},
	// }
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err,
		}).
			Error("failed to create signed cookies")
		return nil, err
	}

	return cookies, nil
}

func GetUrlSign(url string) string {
	privKey, err := loadCloudFrontPrivateKey()
	if err != nil {
		return ""
	}
	// policy := &sign.Policy{
	//	Statements: []sign.Statement{
	//		{
	//			// Read the provided documentation on how to set this
	//			// correctly, you'll probably want to use wildcards.
	//			Resource: serviceutils.GetCDN() + "/*",
	//			Condition: sign.Condition{
	//				// Optional ClientIp source address range
	//				//IPAddress: &sign.IPAddress{SourceIP: "0.0.0.0"},
	//				// Optional date URL is not valid until
	//				DateGreaterThan: &sign.AWSEpochTime{time.Now().Add(30 * time.Minute).UTC()},
	//				// Required date the URL will expire after
	//				DateLessThan: &sign.AWSEpochTime{time.Now().Add(30 * time.Hour).UTC()},
	//			},
	//		},
	//	},
	// }

	s := sign.NewURLSigner(serviceutils.GetPublicKeyId(), privKey)
	signUrl, err := s.Sign(url, time.Now().Add(EXPIRE_TIME*time.Minute).UTC())
	//	signUrl, err := s.SignWithPolicy(url,policy)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err,
		}).
			Error("failed to create signed url")
	}

	return signUrl
}

func loadCloudFrontPrivateKey() (*rsa.PrivateKey, error) {
	key, err := serviceutils.LoadPrivateKey()
	if err != nil {
		return nil, err
	}

	return sign.LoadPEMPrivKey(strings.NewReader(key))
}
