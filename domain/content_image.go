package domain

import (
	"context"
	"net/url"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/modelhelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/s3utils"
)

const CONTENT_IMG_FILE = "content_img_file"

type ContentImage struct {
	Id            int64              `json:"id"`
	FileName      string             `json:"file_name" xorm:"UNIQUE NOT NULL"`
	OriginalName  string             `json:"original_name" xorm:"NOT NULL"`
	OriginalUrl   string             `json:"original_url" xorm:"NOT NULL"`
	FileSize      int64              `json:"file_size" xorm:"NOT NULL"`
	FileExtension string             `json:"file_extension" xorm:"NOT NULL"`
	FilePath      string             `json:"file_path" xorm:"NOT NULL"`
	FileUrl       string             `json:"file_url" xorm:"NOT NULL"`
	ResourceId    string             `json:"resource_id" xorm:"UUID NOT NULL INDEX"`
	Created       jsonutils.JSONTime `json:"created" xorm:"created"`
	Updated       jsonutils.JSONTime `json:"updated" xorm:"updated"`
	Deleted       jsonutils.JSONTime `json:"deleted" xorm:"deleted"`
}

func (c ContentImage) TableName() string {
	return "content_image"
}

func (c ContentImage) SelectAll() string {
	return c.TableName() + ".*"
}

func (c *ContentImage) GetFileType() string {
	return CONTENT_IMG_FILE
}

func (c *ContentImage) GetKey() string {
	return c.FilePath + c.FileName + c.FileExtension
}

func (c *ContentImage) SetLocation(loc string) {
	c.FileUrl = loc
}

func (c *ContentImage) GetFileName() string {
	return c.OriginalName
}
func (ContentImage) AddContentImages(ctx context.Context, resourceId, bucket string, urls []string) ([]*ContentImage, error) {
	contentImages := make([]*ContentImage, 0)
	for _, u := range urls {
		d, err := url.Parse(u)
		if err != nil {
			return nil, err
		}

		extension := filepath.Ext(d.Path)
		filename := filepath.Base(d.Path)[0 : len(filepath.Base(d.Path))-len(extension)]

		img := &ContentImage{
			FileName:      filename,
			OriginalName:  filepath.Base(d.Path),
			FileSize:      0,
			FileExtension: extension,
			FilePath:      s3utils.MakeS3Path(CONTENT_IMG_FILE),
			FileUrl:       "",
			ResourceId:    resourceId,
			OriginalUrl:   u,
		}
		img.FileUrl, err = s3utils.CopyFile(ctx, s3utils.BUCKET_PUBLIC_NAME, img.GetKey(), bucket, d.Path)
		if err != nil {
			return nil, errors.Wrap(err, "failed s3 Copy File")
		}
		if _, err = img.Create(ctx); err != nil {
			return nil, errors.Wrap(err, "failed create content image")
		}
		contentImages = append(contentImages, img)
	}

	return contentImages, nil
}
func (c *ContentImage) Create(ctx context.Context) (int64, error) {
	return factory.DB(ctx).Insert(c)
}

func (ContentImage) GetById(ctx context.Context, id int64) (*ContentImage, error) {
	var v ContentImage
	if has, err := factory.DB(ctx).ID(id).Get(&v); err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}
	return &v, nil
}

func (ContentImage) GetAll(ctx context.Context, sortby, order string, offset, limit int) (int64, []ContentImage, error) {
	q := factory.DB(ctx)
	if err := modelhelper.SetSortOrder(q, sortby, order); err != nil {
		factory.Logger(ctx).Error(err)
	}

	var items []ContentImage
	totalCount, err := q.Limit(limit, offset).FindAndCount(&items)
	if err != nil {
		return 0, nil, err
	}
	return totalCount, items, nil
}

func (ContentImage) FindByResourceId(ctx context.Context, limit int, id string) ([]*ContentImage, error) {
	q := factory.DB(ctx)

	items := make([]*ContentImage, 0)
	_, err := q.Where("resource_id = ?", id).Limit(limit).FindAndCount(&items)
	if err != nil {
		return nil, err
	}

	for i := range items {
		items[i].FileUrl = items[i].GetCDNUrl()
	}

	return items, nil
}

func (ContentImage) FindByResourceIdToMap(ctx context.Context, limit int, id string) (map[string]*ContentImage, error) {
	m := map[string]*ContentImage{}
	items, err := ContentImage{}.FindByResourceId(ctx, limit, id)
	if err != nil {
		return nil, err
	}
	for i := range items {
		m[items[i].GetCDNUrl()] = items[i]
	}
	return m, nil
}

func (ContentImage) Delete(ctx context.Context, id int64) (err error) {
	_, err = factory.DB(ctx).ID(id).Delete(&ContentImage{})
	return err
}

func (c *ContentImage) GetCDNUrl() string {
	return s3utils.MakeCDNURL(c.GetKey())
}

func (s *ContentImage) GetCDNSignUrl() string {
	return s3utils.MakeCDNSignURL(s.GetKey())
}
