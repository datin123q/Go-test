package filehelper

import (
	"net/url"
	"path"
	"strings"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/s3utils"
	"github.com/pkg/errors"
)

type FileHelper interface {
	ExtractFilePath(fileURL string) (string, error)
	ExtractFileName(filePath string) string
	ExtractFileNames(filePaths []string) []string
}

type fileHelper struct{}

func NewFileHelper() FileHelper {
	return &fileHelper{}
}

func (f fileHelper) ExtractFilePath(fileURL string) (string, error) {
	parsedURL, err := url.Parse(fileURL)
	if err != nil {
		return "", errors.Wrap(err, "err in url.Parse")
	}

	return trimBucketPrefix(parsedURL.Path), nil
}

// trimBucketPrefix はパスの先頭に付いたバケット名を取り除く。
//
// 本番は CDN や仮想ホスト形式の URL のためパスにバケット名は含まれないが、
// ローカルの S3 互換エミュレータはパス形式(S3ForcePathStyle)で動くため
// `/<bucket>/<key>` の形になり、そのままでは S3 のキーとして使えない。
//
// CDN のパスを誤って削らないよう、先頭セグメントが設定済みのバケット名と
// 完全一致する場合のみ取り除く。
func trimBucketPrefix(urlPath string) string {
	for _, bucket := range s3utils.BucketNames() {
		if bucket == "" {
			continue
		}

		prefix := "/" + bucket
		if urlPath == prefix {
			return "/"
		}
		if strings.HasPrefix(urlPath, prefix+"/") {
			return strings.TrimPrefix(urlPath, prefix)
		}
	}

	return urlPath
}

func (f fileHelper) ExtractFileName(filePath string) string {
	fileName := path.Base(filePath)
	extension := path.Ext(fileName)

	return strings.TrimSuffix(fileName, extension)
}

func (f fileHelper) ExtractFileNames(filePaths []string) []string {
	fileNames := make([]string, 0)

	for _, filePath := range filePaths {
		fileName := path.Base(filePath)
		extension := path.Ext(fileName)
		fileNames = append(fileNames, strings.TrimSuffix(fileName, extension))
	}

	return fileNames
}

func NormalizeS3FilePath(bucket, filePath string) string {
	if filePath == "" {
		return filePath
	}

	normalized := filePath

	if !strings.HasPrefix(normalized, "/") {
		normalized = "/" + normalized
	}

	if bucket != "" {
		bucket = strings.TrimPrefix(bucket, "/")
		bucketPrefix := "/" + bucket
		if strings.HasPrefix(normalized, bucketPrefix+"/") {
			normalized = strings.TrimPrefix(normalized, bucketPrefix)
		} else if normalized == bucketPrefix {
			normalized = "/"
		}
	}

	return normalized
}
