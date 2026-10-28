package domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/s3utils"
)

const SHARE_FILE_TYPE = "share_file"

type ShareFile struct {
	Id            int64              `json:"id"`
	ShareId       string             `json:"share_id"`
	FileName      string             `json:"file_name" xorm:"UNIQUE NOT NULL"`
	OriginalName  string             `json:"original_name" xorm:"NOT NULL"`
	FileSize      int64              `json:"file_size" xorm:"NOT NULL"`
	FileExtension string             `json:"file_extension" xorm:"NOT NULL"`
	FilePath      string             `json:"file_path" xorm:"NOT NULL"`
	FileUrl       string             `json:"file_url" xorm:"NOT NULL"`
	Created       jsonutils.JSONTime `json:"created" xorm:"created"`
	Updated       jsonutils.JSONTime `json:"updated" xorm:"updated"`
	Deleted       jsonutils.JSONTime `json:"deleted" xorm:"deleted"`
}

func (s *ShareFile) SetLocation(loc string) {
	s.FileUrl = loc
}

func (s *ShareFile) GetFileType() string {
	return SHARE_FILE_TYPE
}

func (s *ShareFile) GetKey() string {
	return s.FilePath + s.FileName + s.FileExtension
}

func (s *ShareFile) GetFileName() string {
	return s.OriginalName
}

func (s *ShareFile) Create(ctx context.Context) (int64, error) {
	return factory.DB(ctx).Insert(s)
}

func (ShareFile) GetById(ctx context.Context, id int64) (*ShareFile, error) {
	var s ShareFile
	if has, err := factory.DB(ctx).ID(id).Get(&s); err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}

	s.FileUrl = s.GetCDNSignUrl()

	return &s, nil
}

func (ShareFile) FindBySharesId(ctx context.Context, id string) (*ShareFile, error) {
	var s ShareFile
	if has, err := factory.DB(ctx).Where("share_id = ?", id).Get(&s); err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}

	s.FileUrl = s.GetCDNSignUrl()

	return &s, nil
}

func (ShareFile) Delete(ctx context.Context, id int64) (err error) {
	_, err = factory.DB(ctx).ID(id).Delete(&ShareFile{})
	return err
}

func (s ShareFile) GetCDNUrl() string {
	return s3utils.MakeCDNURL(s.GetKey())
}

func (s ShareFile) GetCDNSignUrl() string {
	return s3utils.MakeCDNSignURL(s.GetKey())
}
