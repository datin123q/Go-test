package service

import (
	"context"
	"strings"

	"github.com/pkg/errors"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type ContentImageService struct {
}

func NewContentImageService() *ContentImageService {
	return &ContentImageService{}
}

const CONTENTS_IMAGE_LIMIT = 30

func (s *ContentImageService) ReplaceContentImage(ctx context.Context, bucket, contents string, resourceId string, imageUrls []string) (string, error) {

	imageMap, err := domain.ContentImage{}.FindByResourceIdToMap(ctx, CONTENTS_IMAGE_LIMIT, resourceId)
	if err != nil {
		return "", errors.Wrap(err, "failed found contents images")
	}

	newImageUrls := make([]string, 0)

	for _, url := range imageUrls {
		_, has := imageMap[url]
		if has {
			delete(imageMap, url)
		} else {
			if strings.Contains(url, domain.CONTENT_IMG_FILE) {
				continue
			}
			if strings.Contains(contents, url) {
				newImageUrls = append(newImageUrls, url)
			}
		}
	}

	for _, image := range imageMap {
		err := image.Delete(ctx, image.Id)
		if err != nil {
			return "", errors.Wrap(err, "failed delete content image models")
		}
	}

	newImages, err := domain.ContentImage{}.AddContentImages(
		ctx,
		resourceId,
		bucket,
		newImageUrls,
	)
	if err != nil {
		return "", errors.Wrap(err, "failed add content images")
	}

	for i := range newImages {
		contents = strings.Replace(contents, newImages[i].OriginalUrl, newImages[i].GetCDNUrl(), -1)
	}
	return contents, nil
}
