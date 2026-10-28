package controllers

import (
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

type ColumnOutput struct {
	Id           int64              `json:"id" swagger:"desc(아이디)"`                 // 아이디
	Title        string             `json:"title" swagger:"desc(제목)"`               // 제목
	Content      string             `json:"content" swagger:"desc(내용)"`             // 내용
	ThumbnailUrl string             `json:"thumbnail_url"  swagger:"desc(썸네일 URL)"` // 썸네일
	UserId       int64              `json:"user_id"  swagger:"desc(작성자)"`           // 작성자
	Created      jsonutils.JSONTime `json:"created"  swagger:"desc(작성일)"`           // 작성일
	Updated      jsonutils.JSONTime `json:"updated"  swagger:"desc(수정일)"`           // 수정일
}

type ColumnDetailOutput struct {
	Column    ColumnOutput `json:"column" swagger:"desc(コラム)"`
	Thumbnail FileOutput   `json:"thumbnail" swagger:"desc(サムネイル)"`
}

type FileOutput struct {
	Id            int64                   `json:"id" swagger:"desc(ID)"`                // ID
	OriginalName  string                  `json:"file_name" swagger:"desc(ファイル名)"`         // ファイル名
	FileUrl       string                  `json:"file_url" swagger:"desc(ファイルURL)"`       // ファイルURL
	FileSize      int64                   `json:"file_size" swagger:"desc(ファイルサイズ)"`       // ファイルサイズ
	FileExtension string                  `json:"file_extension" swagger:"desc(ファイル拡張子)"` // ファイル拡張子
	UserId        int64                   `json:"user_id" swagger:"desc(著者)"`           // 著者
	FileType      domain.COLUMN_FILE_TYPE `json:"file_type" swagger:"desc(添付ファイルタイプ)"`    // 添付ファイルタイプ
	Created       jsonutils.JSONTime      `json:"created"`                               // 作成日
}

type ColumnSearchInput struct {
	SearchInput
	Title *string `query:"title" swagger:"desc(タイトル)"` // タイトル
}

func (cr ColumnOutput) New(column *domain.Column) *ColumnOutput {
	return &ColumnOutput{
		Id:           column.Id,
		Title:        column.Title,
		Content:      column.Content,
		ThumbnailUrl: column.ThumbnailUrl,
		UserId:       column.UserId,
		Created:      column.Created,
		Updated:      column.Updated,
	}
}

func (cr ColumnOutput) NewList(columns []*domain.Column) []*ColumnOutput {
	var list []*ColumnOutput
	for _, column := range columns {
		list = append(list, cr.New(column))
	}

	return list
}

func (cr ColumnDetailOutput) New(column *domain.Column) *ColumnDetailOutput {
	return &ColumnDetailOutput{
		Column: ColumnOutput{
			Id:           column.Id,
			Title:        column.Title,
			Content:      column.Content,
			ThumbnailUrl: column.ThumbnailUrl,
			UserId:       column.UserId,
			Created:      column.Created,
			Updated:      column.Updated,
		},
		Thumbnail: FileOutput{
			Id:            column.Thumbnail.Id,
			OriginalName:  column.Thumbnail.OriginalName,
			FileUrl:       column.Thumbnail.FileUrl,
			FileSize:      column.Thumbnail.FileSize,
			FileExtension: column.Thumbnail.FileExtension,
			UserId:        column.Thumbnail.UserId,
			FileType:      column.Thumbnail.FileType,
			Created:       column.Thumbnail.Created,
		},
	}
}
