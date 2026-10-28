package domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/modelhelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/s3utils"
)

const COLUMN_FILE_S3_PATH = "column_file"

type COLUMN_FILE_TYPE string

const (
	COLUMN_THUMBNAIL COLUMN_FILE_TYPE = "THUMBNAIL"
)

type Column struct {
	Id           int64              `json:"id" xorm:"pk autoincr"`
	CommonId     string             `json:"common_id" xorm:"notnull UUID" swagger:"desc(공용 ID)"`
	Title        string             `json:"title" xorm:"not null" swagger:"desc(제목)"`
	Content      string             `json:"content" xorm:"LONGTEXT not null" swagger:"desc(내용)"`
	UserId       int64              `json:"user_id" xorm:"not null" swagger:"desc(작성자)"`
	ThumbnailUrl string             `json:"thumbnail_url" xorm:"not null" swagger:"desc(썸네일 URL)"`
	Thumbnail    *ColumnFile        `json:"thumbnail" xorm:"-"`
	IsPosted     bool               `json:"is_posted" xorm:"not null default false" swagger:"desc(게시 여부)"`
	Created      jsonutils.JSONTime `json:"created" xorm:"created"`
	Updated      jsonutils.JSONTime `json:"updated" xorm:"updated"`
	Deleted      jsonutils.JSONTime `json:"deleted" xorm:"deleted"`
	//Id           int64              `json:"id"`
	//Title        string             `json:"title"`
	//Content      string             `json:"content"`
	//UserId       int64              `json:"user_id"`
	//IsPosted     bool               `json:"is_posted"`
	//ThumbnailUrl string             `json:"thumbnail_url"`
	//Thumbnail    *ColumnFile        `json:"thumbnail"`
	//Created      jsonutils.JSONTime `json:"created"`
	//Updated      jsonutils.JSONTime `json:"updated"`
}

type ColumnFile struct {
	Id            int64              `json:"id" xorm:"pk autoincr"`
	ColumnId      int64              `json:"column_id" xorm:"not null"`
	UserId        int64              `json:"user_id" xorm:"not null" swagger:"desc(작성자)"`
	FileName      string             `json:"file_name" xorm:"not null" swagger:"desc(S3 파일명)"`
	OriginalName  string             `json:"original_name" xorm:"NOT NULL" swagger:"desc(오리지날 파일명)"`
	FileType      COLUMN_FILE_TYPE   `json:"file_type" xorm:"NOT NULL" swagger:"desc(첨부 파일 타입)"`
	FileSize      int64              `json:"file_size" xorm:"NOT NULL" swagger:"desc(파일 사이즈)"`
	FileExtension string             `json:"file_extension" xorm:"NOT NULL" swagger:"desc(파일 확장자)"`
	FilePath      string             `json:"file_path" xorm:"NOT NULL" swagger:"desc(S3 경로)"`
	FileUrl       string             `json:"file_url" xorm:"NOT NULL" swagger:"desc(파일 URL)"`
	Created       jsonutils.JSONTime `json:"created" xorm:"created"`
	Deleted       jsonutils.JSONTime `json:"deleted" xorm:"deleted"`
	//Id            int64              `json:"id"`
	//ColumnId      int64              `json:"column_id"`
	//UserId        int64              `json:"user_id"`
	//FileName      string             `json:"file_name"`
	//OriginalName  string             `json:"original_name"`
	//FileType      COLUMN_FILE_TYPE   `json:"file_type"`
	//FileSize      int64              `json:"file_size"`
	//FileExtension string             `json:"file_extension"`
	//FilePath      string             `json:"file_path"`
	//FileUrl       string             `json:"file_url"`
	//Created       jsonutils.JSONTime `json:"created"`
}

func (r *ColumnFile) GetKey() string {
	return r.FilePath + r.FileName + r.FileExtension
}

func (r *ColumnFile) GetCDNUrl() string {
	return s3utils.MakeCDNURL(r.GetKey())
}

func (r *ColumnFile) GetCDNSignUrl() string {
	return s3utils.MakeCDNSignURL(r.GetKey())
}

func (r *ColumnFile) GetFileType() string {
	return PROJECT_IMG_FILE_TYPE
}

func (r *ColumnFile) GetFileName() string {
	return r.OriginalName
}

func (r *ColumnFile) SetLocation(loc string) {
	r.FileUrl = loc
}

// Create Deprecated
func (c *Column) Create(ctx context.Context) (int64, error) {
	return factory.DB(ctx).Insert(c)
}

// FindById Deprecated
func (Column) FindById(ctx context.Context, id int64) (*Column, error) {
	var item Column

	_, err := factory.DB(ctx).Where("id = ?", id).Get(&item)
	if err != nil {
		return nil, err
	}

	return &item, nil
}

// FindAll Deprecated
func (Column) FindAll(ctx context.Context, title, id, start, end *string, sortBy, order string, offset, limit int) (int64, []*Column, error) {
	q := factory.DB(ctx).Table("column").Select("*")

	if err := modelhelper.SetSortOrder(q, sortBy, order); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SearchLikeQuery(q, "column", "title", title); err != nil {
		return 0, nil, err
	}
	if err := modelhelper.SearchEqQuery(q, "column", "id", id); err != nil {
		return 0, nil, err
	}
	if start != nil && end != nil {
		if err := modelhelper.SetBetweenDate(q, "created", *start, *end); err != nil {
			return 0, nil, err
		}
	}

	var items []*Column
	total, err := q.Limit(limit, offset).FindAndCount(&items)
	if err != nil {
		return 0, nil, err
	}

	return total, items, nil
}

// FindPostedAll Deprecated
func (Column) FindPostedAll(ctx context.Context, title *string, sortBy, order string, offset, limit int) (int64, []*Column, error) {
	q := factory.DB(ctx).Table("column").Select("*").Where("is_posted = ?", true)

	if err := modelhelper.SetSortOrder(q, sortBy, order); err != nil {
		factory.Logger(ctx).Error(err)
	}

	if err := modelhelper.SearchLikeQuery(q, "column", "title", title); err != nil {
		return 0, nil, err
	}

	var items []*Column
	total, err := q.Limit(limit, offset).FindAndCount(&items)
	if err != nil {
		return 0, nil, err
	}

	return total, items, nil
}

// Update Deprecated
func (c *Column) Update(ctx context.Context) (int64, error) {
	return factory.DB(ctx).UseBool("is_posted").ID(c.Id).Update(c)
}

// Delete Deprecated
func (c *Column) Delete(ctx context.Context) (int64, error) {
	return factory.DB(ctx).ID(c.Id).Delete(c)
}

// Create Deprecated
func (c *ColumnFile) Create(ctx context.Context) (int64, error) {
	return factory.DB(ctx).Insert(c)
}

// FindByColumnIdAndFileType Deprecated
func (c ColumnFile) FindByColumnIdAndFileType(ctx context.Context, columnId int64, fileType COLUMN_FILE_TYPE) (*ColumnFile, error) {
	var item ColumnFile

	_, err := factory.DB(ctx).Where("column_id = ?", columnId).And("file_type = ?", fileType).Get(&item)
	if err != nil {
		return nil, err
	}

	return &item, nil
}

// Delete Deprecated
func (c *ColumnFile) Delete(ctx context.Context) (int64, error) {
	return factory.DB(ctx).ID(c.Id).Delete(c)
}
