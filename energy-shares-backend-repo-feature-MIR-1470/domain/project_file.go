package domain

import (
	"context"
	"time"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/s3utils"
)

type PROJECT_FILE_TYPE string

const (
	CONTRACT_BEFORE_PAGE PROJECT_FILE_TYPE = "CONTRACT_BEFORE_PAGE"
	CONTRACT_AFTER_PAGE  PROJECT_FILE_TYPE = "CONTRACT_AFTER_PAGE"
	SILENT_PARTNERSHIPS  PROJECT_FILE_TYPE = "SILENT_PARTNERSHIPS"
	PROJECT_REPORT       PROJECT_FILE_TYPE = "PROJECT_REPORT"
)

// RequiredProjectFileTypes は、空にしたまま登録できない書面の種別と、
// 差し替え時に管理画面が新しいファイルを送るマルチパートフォームの
// フィールド名の対応表。削除リクエストと同じフォームにこのフィールド名で
// 新規ファイルが無い場合、対象種別の削除を拒否するために使う
// (差し替えのつもりが削除だけ通り、書面が空になる事象への対策)。
// PROJECT_REPORT(ファンドレポート)は必須ではないため含まない。
var RequiredProjectFileTypes = map[PROJECT_FILE_TYPE]string{
	CONTRACT_BEFORE_PAGE: "contract_before_page",
	CONTRACT_AFTER_PAGE:  "contract_after_page",
	SILENT_PARTNERSHIPS:  "silent_partnerships",
}

const PROJECT_FILE = "project_file"

type ProjectFile struct {
	Id            int64             `json:"id" swagger:"desc(프로젝트 이미지 고유 ID)"`
	ProjectId     int64             `json:"project_id" xorm:"NOT NULL" swagger:"desc(상품 ID)"`
	FileName      string            `json:"file_name" xorm:"UNIQUE NOT NULL" swagger:"desc(파일명)"`
	OriginalName  string            `json:"original_name" xorm:"NOT NULL" swagger:"desc(오리지날 파일명)"`
	FileType      PROJECT_FILE_TYPE `json:"file_type" xorm:"NOT NULL" swagger:"desc(첨부파일 타입,공통코드 [PROJECT_FILE_TYPE])"`
	FileSize      int64             `json:"file_size" xorm:"NOT NULL" swagger:"desc(파일 사이즈)"`
	FileExtension string            `json:"file_extension" xorm:"NOT NULL" swagger:"desc(파일 확장자)"`
	FilePath      string            `json:"file_path" xorm:"NOT NULL" swagger:"desc(파일 경로)"`
	FileUrl       string            `json:"file_url" xorm:"NOT NULL" swagger:"desc(URL)"`
	OwnerId       int64             `json:"owner_id" xorm:"NOT NULL" swagger:"desc(소유주 ID)"`
	UpdateId      int64             `json:"update_id"  swagger:"desc(최근 업데이트 사용자)"`
	Created       time.Time         `json:"created" xorm:"created"`
	Updated       time.Time         `json:"updated" xorm:"updated"`
	Deleted       time.Time         `json:"deleted" xorm:"deleted"`
}

func (ProjectFile) TableName() string {
	return "project_file"
}

func (p ProjectFile) SelectAll() string {
	return p.TableName() + ".*"
}

func (p *ProjectFile) Compare(t *ProjectFile) bool {
	if p.Id != t.Id {
		return false
	}
	if p.ProjectId != t.ProjectId {
		return false
	}
	if p.FileName != t.FileName {
		return false
	}
	if p.OriginalName != t.OriginalName {
		return false
	}
	if p.FileType != t.FileType {
		return false
	}
	if p.FileSize != t.FileSize {
		return false
	}
	if p.FileExtension != t.FileExtension {
		return false
	}
	if p.FilePath != t.FilePath {
		return false
	}
	if p.FileUrl != t.FileUrl {
		return false
	}
	if p.OwnerId != t.OwnerId {
		return false
	}
	if p.UpdateId != t.UpdateId {
		return false
	}
	return true
}
func (p *ProjectFile) SetLocation(loc string) {
	p.FileUrl = loc
}

func (p *ProjectFile) GetCDNUrl() string {
	return s3utils.MakeCDNURL(p.GetKey())
}

func (p *ProjectFile) GetCDNSignUrl() string {
	return s3utils.MakeCDNSignURL(p.GetKey())
}

func (p *ProjectFile) GetFileType() string {
	return PROJECT_FILE
}

func (p *ProjectFile) GetKey() string {
	return p.FilePath + p.FileName + p.FileExtension
}

func (p *ProjectFile) GetFileName() string {
	return p.OriginalName
}

func (p *ProjectFile) Create(ctx context.Context) (int64, error) {
	return factory.DB(ctx).Insert(p)
}

func (ProjectFile) GetById(ctx context.Context, id int64) (*ProjectFile, error) {
	var v ProjectFile
	if has, err := factory.DB(ctx).ID(id).Get(&v); err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}
	return &v, nil
}

func (ProjectFile) FindByProjectIdAndFileType(ctx context.Context, id int64, file_type PROJECT_FILE_TYPE) (*ProjectFile, error) {
	var v ProjectFile
	if has, err := factory.DB(ctx).Where("project_id = ? AND file_type=?", id, file_type).Get(&v); err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}

	return &v, nil
}

func (ProjectFile) Delete(ctx context.Context, id int64) (err error) {
	_, err = factory.DB(ctx).ID(id).Delete(&ProjectFile{})
	return err
}

type PROJECT_IMG_TYPE string

const (
	MAIN_IMG   PROJECT_IMG_TYPE = "MAIN_IMG"
	SUB_IMG    PROJECT_IMG_TYPE = "SUB_IMG"
	BANNER_IMG                  = "BANNER_IMG"
)

const PROJECT_IMG_FILE_TYPE = "project_img_file"

type ProjectImage struct {
	Id            int64            `json:"id" swagger:"desc(프로젝트 이미지 고유 ID)"`
	ProjectId     int64            `json:"project_id" xorm:"NOT NULL" swagger:"desc(프로젝트 ID)"`
	FileName      string           `json:"file_name" xorm:"UNIQUE NOT NULL" swagger:"desc(파일명)"`
	OriginalName  string           `json:"original_name" xorm:"NOT NULL" swagger:"desc(오리지날 파일명)"`
	FileType      string           `json:"file_type" xorm:"NOT NULL" swagger:"desc(파일 타입)"`
	FileSize      int64            `json:"file_size" xorm:"NOT NULL" swagger:"desc(파일 사이즈)"`
	FileExtension string           `json:"file_extension" xorm:"NOT NULL" swagger:"desc(파일 확장자)"`
	FilePath      string           `json:"file_path" xorm:"NOT NULL" swagger:"desc(파일 경로)"`
	FileUrl       string           `json:"file_url" xorm:"NOT NULL" swagger:"desc(URL)"`
	OwnerId       int64            `json:"owner_id" xorm:"NOT NULL" swagger:"desc(소유주 ID)"`
	UpdateId      int64            `json:"updator"  swagger:"desc(최근 업데이트 사용자)"`
	ImageType     PROJECT_IMG_TYPE `json:"image_type" xorm:"NOT NULL" swagger:"desc(이미지 타입),enum(MAIN_IMG|SUB_IMG) "` // 메인,썸네일 Big, Small
	Order         int64            `json:"order" form:"order" swagger:"desc(表示順)"`                                    // NOTE: メイン画像の場合は0固定、サブ画像の場合は0は非表示、1以上は表示で制御する
	Created       time.Time        `json:"created" xorm:"created"`
	Updated       time.Time        `json:"updated" xorm:"updated"`
	Deleted       time.Time        `json:"deleted" xorm:"deleted"`
}

func (ProjectImage) TableName() string {
	return "project_image"
}

func (p ProjectImage) SelectAll() string {
	return p.TableName() + ".*"
}

func (p *ProjectImage) Compare(t *ProjectImage) bool {
	if p.Id != t.Id {
		return false
	}
	if p.ProjectId != t.ProjectId {
		return false
	}
	if p.FileName != t.FileName {
		return false
	}
	if p.OriginalName != t.OriginalName {
		return false
	}
	if p.FileType != t.FileType {
		return false
	}
	if p.FileSize != t.FileSize {
		return false
	}
	if p.FileExtension != t.FileExtension {
		return false
	}
	if p.FilePath != t.FilePath {
		return false
	}
	if p.FileUrl != t.FileUrl {
		return false
	}
	if p.OwnerId != t.OwnerId {
		return false
	}
	if p.UpdateId != t.UpdateId {
		return false
	}
	if p.ImageType != t.ImageType {
		return false
	}
	if p.Order != t.Order {
		return false
	}

	return true
}

func (p *ProjectImage) GetKey() string {
	return p.FilePath + p.FileName + p.FileExtension
}

func (p *ProjectImage) GetCDNUrl() string {
	return s3utils.MakeCDNURL(p.GetKey())
}

func (p *ProjectImage) GetCDNSignUrl() string {
	return s3utils.MakeCDNSignURL(p.GetKey())
}

func (p *ProjectImage) GetFileType() string {
	return PROJECT_IMG_FILE_TYPE
}

func (p *ProjectImage) GetFileName() string {
	return p.OriginalName
}

func (p *ProjectImage) SetLocation(loc string) {
	p.FileUrl = loc
}

func (p *ProjectImage) Create(ctx context.Context) (int64, error) {
	return factory.DB(ctx).Insert(p)
}

func (ProjectImage) CreateList(ctx context.Context, images []*ProjectImage) (int64, error) {
	return factory.DB(ctx).Insert(images)
}

func (ProjectImage) Delete(ctx context.Context, id int64) error {
	_, err := factory.DB(ctx).ID(id).Delete(&ProjectImage{})
	return err
}

func (p *ProjectImage) Update(ctx context.Context) (int64, error) {
	return factory.DB(ctx).ID(p.Id).
		Cols(
			"file_name",
			"original_name",
			"file_type",
			"file_size",
			"file_extension",
			"file_path",
			"file_url",
			"owner_id",
			"update_id",
			"image_type",
			"order",
		).
		Update(p)
}

func (ProjectImage) GetById(ctx context.Context, id int64) (*ProjectImage, error) {

	var pi ProjectImage
	if has, err := factory.DB(ctx).ID(id).Get(&pi); err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}

	return &pi, nil
}

func (ProjectImage) FindByProjectId(ctx context.Context, projectId int64) ([]*ProjectImage, error) {
	items := make([]*ProjectImage, 0)
	if err := factory.DB(ctx).Where("project_id = ?", projectId).Find(&items); err != nil {
		return nil, err
	}

	return items, nil
}

func (ProjectImage) FindByProjectIdAndImageType(ctx context.Context, projectId int64, imageType PROJECT_IMG_TYPE) ([]*ProjectImage, error) {
	items := make([]*ProjectImage, 0)
	if err := factory.DB(ctx).Where("project_id = ? AND image_type = ?", projectId, imageType).Find(&items); err != nil {
		return nil, err
	}

	return items, nil
}
