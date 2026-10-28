package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/s3utils"
)

// SaveApplicantImages Nexway API から取得した申請画像（元画像）をS3に保存しDBに記録
func SaveApplicantImages(ctx context.Context, ekycRequestLogId int64, applicantId int64, images []domain.NexwayApplicantImageDto) error {
	for _, img := range images {
		if img.ImageData == "" {
			continue
		}

		imageBytes, ext, err := decodeAndDetectImage(img.ImageData, img.ImageType)
		if err != nil {
			return err
		}

		fileName := makeImageFileName(applicantId, img.ImageType, ext)
		filePath := s3utils.MakeS3Path(domain.NEXWAY_APPLICANT_IMAGE_FILE_TYPE)
		s3Image := &domain.NexwayApplicantImage{
			EkycRequestLogId: ekycRequestLogId,
			ImageType:        img.ImageType,
			S3Key:            filePath + fileName,
		}

		if err := uploadToS3AndSaveDB(ctx, s3Image, imageBytes, "applicant"); err != nil {
			return err
		}

		logrus.WithFields(logrus.Fields{
			"ekyc_request_log_id": ekycRequestLogId,
			"image_type":          img.ImageType,
			"s3_key":              s3Image.S3Key,
		}).Info("Saved applicant image")
	}

	return nil
}

// decodeAndDetectImage base64デコードと画像形式判定を行う
func decodeAndDetectImage(base64Data string, imageType domain.NexwayImageType) ([]byte, string, error) {
	imageBytes, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":      err,
			"image_type": imageType,
		}).Error("Failed to decode base64 image")
		return nil, "", fmt.Errorf("failed to decode base64 image (type=%s): %w", imageType, err)
	}

	ext, err := detectImageExtension(imageBytes)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":      err,
			"image_type": imageType,
		}).Error("Failed to detect image format")
		return nil, "", fmt.Errorf("failed to detect image format (type=%s): %w", imageType, err)
	}

	return imageBytes, ext, nil
}

// uploadToS3AndSaveDB S3へのアップロードとDB保存を行う
func uploadToS3AndSaveDB(ctx context.Context, s3Obj s3utils.S3Object, imageBytes []byte, category string) error {
	// S3にアップロード
	buf := bytes.NewBuffer(imageBytes)
	_, err := s3utils.UploadBuffFile(ctx, s3utils.BUCKET_EKYC_IMAGES_NAME, s3Obj, buf)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":    err,
			"s3_key":   s3Obj.GetKey(),
			"category": category,
		}).Error("Failed to upload image to S3")
		return fmt.Errorf("failed to upload %s image to S3: %w", category, err)
	}

	// DBに保存
	var dbErr error
	switch v := s3Obj.(type) {
	case *domain.NexwayApplicantImage:
		_, dbErr = v.Create(ctx)
	default:
		return fmt.Errorf("unsupported S3Object type: %T", s3Obj)
	}

	if dbErr != nil {
		logrus.WithFields(logrus.Fields{
			"error":    dbErr,
			"s3_key":   s3Obj.GetKey(),
			"category": category,
		}).Error("Failed to save image record to DB")
		return fmt.Errorf("failed to save %s image to DB: %w", category, dbErr)
	}

	return nil
}

// detectImageExtension 画像のバイナリからファイル拡張子を判定
// Nexway APIはJPEGまたはPNG形式のみ返却する
func detectImageExtension(data []byte) (string, error) {
	if len(data) < 12 {
		return "", fmt.Errorf("image data too short (%d bytes)", len(data))
	}

	contentType := http.DetectContentType(data)
	switch contentType {
	case "image/jpeg":
		return ".jpg", nil
	case "image/png":
		return ".png", nil
	default:
		return "", fmt.Errorf("unsupported image format: %s (expected jpeg or png)", contentType)
	}
}

// makeImageFileName 画像ファイル名を生成 (形式: {applicantId}_{imageType}_{uniqueId}.ext)
func makeImageFileName(applicantId int64, imageType domain.NexwayImageType, ext string) string {
	uniqueId := s3utils.MakeS3FileName()
	return fmt.Sprintf("%d_%s_%s%s", applicantId, imageType, uniqueId, ext)
}
