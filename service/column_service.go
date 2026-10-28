package service

import (
	"context"
	"mime/multipart"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/s3utils"
)

type IColumnService interface {
	CreateColumn(context.Context, *domain.Column, *multipart.FileHeader) error
	GetColumnById(context.Context, int64) (*domain.Column, error)
	GetColumnList(context.Context, *string, *string, *string, *string, string, string, int, int) (int64, []*domain.Column, error)
	GetPostedColumnList(context.Context, *string, string, string, int, int) (int64, []*domain.Column, error)
	UpdateColumn(context.Context, *domain.Column) error
	UpdateColumnAndThumbnail(context.Context, *domain.Column, *multipart.FileHeader) error
	DeleteColumnById(context.Context, int64) error
	GetThumbnailByColumnId(context.Context, int64) (*domain.ColumnFile, error)
	DeleteThumbnailByColumnId(context.Context, int64) error
}

var _ IColumnService = (*ColumnService)(nil)

type ColumnService struct {
	// repository
}

func NewColumnService() IColumnService {
	return &ColumnService{}
}

func (s *ColumnService) CreateColumn(ctx context.Context, column *domain.Column, thumbnail *multipart.FileHeader) error {
	var err error

	// upload thumbnail
	column.Thumbnail = &domain.ColumnFile{
		FileName:      s3utils.MakeS3FileName(),
		OriginalName:  thumbnail.Filename,
		ColumnId:      column.Id,
		UserId:        column.UserId,
		FileSize:      thumbnail.Size,
		FileType:      domain.COLUMN_THUMBNAIL,
		FileExtension: s3utils.GetExtension(thumbnail.Filename),
		FilePath:      s3utils.MakeS3Path(domain.COLUMN_FILE_S3_PATH),
		FileUrl:       "",
	}

	column.ThumbnailUrl, err = s3utils.UploadAttachmentFile(ctx, s3utils.BUCKET_PUBLIC_NAME, column.Thumbnail, thumbnail)
	if err != nil {
		return factory.ErrorAWSUpload.New(err)
	}

	// create column
	if _, err = column.Create(ctx); err != nil {
		return factory.ErrorDB.New(err)
	}

	// create thumb nail
	column.Thumbnail.ColumnId = column.Id
	if _, err = column.Thumbnail.Create(ctx); err != nil {
		return factory.ErrorDB.New(err)
	}

	return nil
}

func (s *ColumnService) GetColumnById(ctx context.Context, id int64) (*domain.Column, error) {
	column, err := domain.Column{}.FindById(ctx, id)
	if err != nil {
		return nil, factory.ErrorDB.New(err)
	} else if column == nil {
		return nil, factory.ErrorNotFound.New(nil)
	}

	column.Thumbnail, err = domain.ColumnFile{}.FindByColumnIdAndFileType(ctx, column.Id, domain.COLUMN_THUMBNAIL)
	if err != nil {
		return nil, factory.ErrorDB.New(err)
	}

	return column, nil
}

func (s *ColumnService) GetColumnList(ctx context.Context, title, id, start, end *string, sortBy, order string, offset, limit int) (int64, []*domain.Column, error) {
	total, columnList, err := domain.Column{}.FindAll(ctx, title, id, start, end, sortBy, order, offset, limit)
	if err != nil {
		return 0, nil, factory.ErrorDB.New(err)
	}

	return total, columnList, nil
}

func (s *ColumnService) GetPostedColumnList(ctx context.Context, title *string, sortBy, order string, offset, limit int) (int64, []*domain.Column, error) {
	total, columnList, err := domain.Column{}.FindPostedAll(ctx, title, sortBy, order, offset, limit)
	if err != nil {
		return 0, nil, factory.ErrorDB.New(err)
	}

	return total, columnList, nil
}

func (s *ColumnService) UpdateColumn(ctx context.Context, column *domain.Column) error {
	if _, err := column.Update(ctx); err != nil {
		return factory.ErrorDB.New(err)
	}

	return nil
}

func (s *ColumnService) UpdateColumnAndThumbnail(ctx context.Context, column *domain.Column, thumbnail *multipart.FileHeader) error {
	var err error

	// upload thumbnail
	column.Thumbnail = &domain.ColumnFile{
		FileName:      s3utils.MakeS3FileName(),
		OriginalName:  thumbnail.Filename,
		ColumnId:      column.Id,
		UserId:        column.UserId,
		FileSize:      thumbnail.Size,
		FileType:      domain.COLUMN_THUMBNAIL,
		FileExtension: s3utils.GetExtension(thumbnail.Filename),
		FilePath:      s3utils.MakeS3Path(domain.COLUMN_FILE_S3_PATH),
		FileUrl:       "",
	}

	column.ThumbnailUrl, err = s3utils.UploadAttachmentFile(ctx, s3utils.BUCKET_PUBLIC_NAME, column.Thumbnail, thumbnail)
	if err != nil {
		return factory.ErrorAWSUpload.New(err)
	}

	// update column
	if _, err = column.Update(ctx); err != nil {
		return factory.ErrorDB.New(err)
	}

	// create thumb nail
	if _, err = column.Thumbnail.Create(ctx); err != nil {
		return factory.ErrorDB.New(err)
	}

	return nil
}

func (s *ColumnService) DeleteColumnById(ctx context.Context, id int64) error {
	column := domain.Column{Id: id}
	if _, err := column.Delete(ctx); err != nil {
		return factory.ErrorDB.New(err)
	}

	if err := s.DeleteThumbnailByColumnId(ctx, id); err != nil {
		return err
	}

	return nil
}

func (s *ColumnService) GetThumbnailByColumnId(ctx context.Context, columnId int64) (*domain.ColumnFile, error) {
	thumbnail, err := domain.ColumnFile{}.FindByColumnIdAndFileType(ctx, columnId, domain.COLUMN_THUMBNAIL)
	if err != nil {
		return nil, factory.ErrorDB.New(err)
	}

	return thumbnail, nil
}

func (s *ColumnService) DeleteThumbnailByColumnId(ctx context.Context, columnId int64) error {
	thumbnail, err := domain.ColumnFile{}.FindByColumnIdAndFileType(ctx, columnId, domain.COLUMN_THUMBNAIL)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	if _, err := thumbnail.Delete(ctx); err != nil {
		return factory.ErrorDB.New(err)
	}

	err = s3utils.DeleteFile(ctx, s3utils.BUCKET_PUBLIC_NAME, thumbnail)
	if err != nil {
		return factory.ErrorAWSDelete.New(err)
	}

	return nil
}
