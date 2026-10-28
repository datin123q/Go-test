package controllers

import (
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

type ColumnInput struct {
	Title    string `json:"title" form:"title" swagger:"desc(タイトル)" validate:"required"`        // タイトル
	Content  string `json:"content" form:"content" swagger:"desc(コンテンツ)" validate:"required"`    // コンテンツ
	IsPosted bool   `json:"is_posted" form:"is_posted" swagger:"desc(投稿有無、デフォルト: false)"` // 投稿有無
}

type ColumnOutput struct {
	Id           int64              `json:"id" swagger:"desc(ID)"`                 // ID
	Title        string             `json:"title" swagger:"desc(タイトル)"`               // タイトル
	Content      string             `json:"content" swagger:"desc(コンテンツ)"`             // コンテンツ
	ThumbnailUrl string             `json:"thumbnail_url"  swagger:"desc(サムネイルURL)"` // サムネイル
	UserId       int64              `json:"user_id"  swagger:"desc(작성자)"`           // 작성자
	IsPosted     bool               `json:"is_posted"  swagger:"desc(投稿有無)"`       // 投稿有無
	Created      jsonutils.JSONTime `json:"created"  swagger:"desc(작성일)"`           // 작성일
	Updated      jsonutils.JSONTime `json:"updated"  swagger:"desc(수정일)"`           // 수정일
}

type ColumnDetailOutput struct {
	Column    ColumnOutput `json:"column" swagger:"desc(カラム)"`
	Thumbnail FileOutput   `json:"thumbnail" swagger:"desc(サムネイル)"`
}

type FileOutput struct {
	Id            int64                   `json:"id" swagger:"desc(아이디)"`                // 아이디
	OriginalName  string                  `json:"file_name" swagger:"desc(ファイル名)"`         // ファイル名
	FileUrl       string                  `json:"file_url" swagger:"desc(ファイルURL)"`       // ファイルURL
	FileSize      int64                   `json:"file_size" swagger:"desc(파일 크기)"`       // 파일 크기
	FileExtension string                  `json:"file_extension" swagger:"desc(파일 확장자)"` // 파일 확장자
	UserId        int64                   `json:"user_id" swagger:"desc(작성자)"`           // 작성자
	FileType      domain.COLUMN_FILE_TYPE `json:"file_type" swagger:"desc(첨부 파일 타입)"`    // 첨부 파일 타입
	Created       jsonutils.JSONTime      `json:"created"`                               // 작성일
}

type ColumnSearchInput struct {
	SearchInput
	Title     *string `query:"title" swagger:"desc(タイトル)"` // タイトル
	Id        *string `query:"id" swagger:"desc(ID)"`   // ID
	StartDate *string `query:"start_date" swagger:"desc(作成日-from)"`
	EndDate   *string `query:"end_date" swagger:"desc(作成日-to)"`
}

func (cr *ColumnInput) ToModel() *domain.Column {
	return &domain.Column{
		Title:    cr.Title,
		Content:  cr.Content,
		IsPosted: cr.IsPosted,
	}
}

func (cr ColumnOutput) New(column *domain.Column) *ColumnOutput {
	return &ColumnOutput{
		Id:           column.Id,
		Title:        column.Title,
		Content:      column.Content,
		ThumbnailUrl: column.ThumbnailUrl,
		IsPosted:     column.IsPosted,
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
			IsPosted:     column.IsPosted,
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
