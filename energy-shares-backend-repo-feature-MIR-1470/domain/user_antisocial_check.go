package domain

import (
	"context"
	"mime/multipart"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/s3utils"
)

const ANTISOCIAL_CHECK_FILE_TYPE = "antisocial_check_file"

type ANTISOCIAL_CHECK_STATUS int

const (
	ANTISOCIAL_CHECK_NONE ANTISOCIAL_CHECK_STATUS = 0
	ANTISOCIAL_CHECK_YES  ANTISOCIAL_CHECK_STATUS = 1
	ANTISOCIAL_CHECK_NO   ANTISOCIAL_CHECK_STATUS = 2
)

// 소셜 데이터 파일
type AntisocialCheckFile struct {
	Id            int64              `json:"id" swagger:"desc(반사회 세력 체크 파일 고유 ID)"`
	CheckId       int64              `json:"check_id" xorm:"NOT NULL" swagger:"desc(체크 데이터 ID)"`
	FileName      string             `json:"file_name" xorm:"UNIQUE NOT NULL" swagger:"desc(파일명)"`
	OriginalName  string             `json:"original_name" xorm:"NOT NULL" swagger:"desc(오리지날 파일명)"`
	FileSize      int64              `json:"file_size" xorm:"NOT NULL" swagger:"desc(파일 사이즈)"`
	FileExtension string             `json:"file_extension" xorm:"NOT NULL" swagger:"desc(파일 확장자)"`
	FilePath      string             `json:"file_path" xorm:"NOT NULL" swagger:"desc(파일 경로)"`
	FileUrl       string             `json:"file_url" xorm:"NOT NULL" swagger:"desc(URL)"`
	OwnerId       int64              `json:"owner_id" xorm:"NOT NULL" swagger:"desc(소유주 ID)"`
	UpdateId      int64              `json:"update_id"  swagger:"desc(최근 업데이트 사용자)"`
	Created       jsonutils.JSONTime `json:"created" xorm:"created"`
	Updated       jsonutils.JSONTime `json:"updated" xorm:"updated"`
	Deleted       jsonutils.JSONTime `json:"deleted" xorm:"deleted"`
}

type AntisocialCheckFiles []AntisocialCheckFile

func (a AntisocialCheckFiles) ToPtrSlice() []*AntisocialCheckFile {
	var ptrSlice []*AntisocialCheckFile
	for i := range a {
		ptrSlice = append(ptrSlice, &a[i])
	}
	return ptrSlice
}

func (AntisocialCheckFile) TableName() string {
	return "antisocial_check_file"
}

func (a AntisocialCheckFile) SelectAll() string {
	return a.TableName() + ".*"
}

func (a *AntisocialCheckFile) GetKey() string {
	return a.FilePath + a.FileName + a.FileExtension
}

func (a *AntisocialCheckFile) GetCDNUrl() string {
	return s3utils.MakeCDNURL(a.GetKey())
}

func (a *AntisocialCheckFile) GetCDNSignUrl() string {
	return s3utils.MakeCDNSignURL(a.GetKey())
}

func (a *AntisocialCheckFile) GetFileType() string {
	return ANTISOCIAL_CHECK_FILE_TYPE
}

func (a *AntisocialCheckFile) GetFileName() string {
	return a.OriginalName
}

func (a *AntisocialCheckFile) SetLocation(loc string) {
	a.FileUrl = loc
}

type UserAntisocialCheck struct {
	Id            int64                   `json:"id"`
	CheckUserId   int64                   `json:"check_user_id" xorm:"NOT NULL"  swagger:"desc(반사회적 세력 여부를 체크한 사용자)"`
	Checker       string                  `json:"checker"  swagger:"desc(반사회 체크자)"`
	CheckDate     jsonutils.JSONTime      `json:"check_date" swagger:"desc(반사회 체크일)"`
	CheckStatus   ANTISOCIAL_CHECK_STATUS `json:"check_status"  swagger:"desc(NONE:0 승인:1 비승인:2)"`
	InvalidReason string                  `json:"invalid_reason"  swagger:"desc(비승인 사유)"`
	Updated       jsonutils.JSONTime      `json:"-" xorm:"updated"`
	Created       jsonutils.JSONTime      `json:"-" xorm:"created"`
	CheckFiles    []*AntisocialCheckFile  `json:"check_files" xorm:"-" swagger:"desc(확인 자료)"`
}

func (u UserAntisocialCheck) IsEmpty() bool {
	return u.Id == 0
}

func (UserAntisocialCheck) TableName() string {
	return "user_antisocial_check"
}

func (u UserAntisocialCheck) SelectAll() string {
	return u.TableName() + ".*"
}

func (a *AntisocialCheckFile) Create(ctx context.Context) (int64, error) {
	q := factory.DB(ctx)
	return q.Insert(a)
}

func (AntisocialCheckFile) GetById(ctx context.Context, id int64) (*AntisocialCheckFile, error) {
	var v AntisocialCheckFile

	if has, err := factory.DB(ctx).ID(id).Get(&v); err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}

	v.FileUrl = v.GetCDNSignUrl()

	return &v, nil
}

// Deprecated: Use AntisocialCheckFileRepository.GetByCheckId instead
func (AntisocialCheckFile) FindByCheckId(ctx context.Context, limit int, id int64) ([]*AntisocialCheckFile, error) {
	// 반사회 확인 파일은 최신파일을 최상단에 보여준다.
	q := factory.DB(ctx)
	items := make([]*AntisocialCheckFile, 0)
	_, err := q.Where("check_id = ?", id).Limit(limit).Desc("created").FindAndCount(&items)
	if err != nil {
		return nil, err
	}
	for i := range items {
		items[i].FileUrl = items[i].GetCDNSignUrl()
	}

	return items, nil
}

func (AntisocialCheckFile) AddCheckFiles(ctx context.Context, id int64, form *multipart.Form) ([]*AntisocialCheckFile, error) {
	var err error
	antisocialCheckFiles := make([]*AntisocialCheckFile, 0)
	for _, value := range form.File {

		for _, file := range value {
			antisocialCheckFile := &AntisocialCheckFile{
				FileName:      s3utils.MakeS3FileName(),
				OriginalName:  file.Filename,
				FileSize:      file.Size,
				FileExtension: s3utils.GetExtension(file.Filename),
				FilePath:      s3utils.MakeS3Path(ANTISOCIAL_CHECK_FILE_TYPE),
				FileUrl:       "",
				CheckId:       id,
			}
			antisocialCheckFile.FileUrl, err = s3utils.UploadShowFile(ctx, s3utils.BUCKET_PRIVATE_NAME, antisocialCheckFile, file)
			if err != nil {
				return nil, err
			}
			if _, err = antisocialCheckFile.Create(ctx); err != nil {
				return nil, err
			}
			antisocialCheckFiles = append(antisocialCheckFiles, antisocialCheckFile)
		}
	}
	return antisocialCheckFiles, nil
}

// FindByIdsUnscoped は、IDで反社チェックの書類をまとめて取る(論理削除済みも含む)。MIR-1413で、削除の指定が対象の
// 会員の書類かを確かめるために使う。削除済みも取るのは、画面からの送り直しで届く自分の削除済みの書類を、他の会員の
// 書類と区別して黙って飛ばすため。
func (AntisocialCheckFile) FindByIdsUnscoped(ctx context.Context, ids []int64) ([]AntisocialCheckFile, error) {
	items := make([]AntisocialCheckFile, 0)
	if len(ids) == 0 {
		return items, nil
	}
	if err := factory.DB(ctx).Unscoped().In("id", ids).Find(&items); err != nil {
		return nil, err
	}
	return items, nil
}

// DeleteByIdAndCheckId は、反社チェックの書類を論理削除する。紐づけ先(check_id。会員のID)も条件に入れ、判定を
// 通った書類だけを消す。
func (AntisocialCheckFile) DeleteByIdAndCheckId(ctx context.Context, id, checkId int64) error {
	_, err := factory.DB(ctx).Where("id = ? AND check_id = ?", id, checkId).Delete(&AntisocialCheckFile{})
	return err
}

func (AntisocialCheckFile) Delete(ctx context.Context, id int64) (err error) {
	_, err = factory.DB(ctx).ID(id).Delete(&AntisocialCheckFile{})
	return err
}

func (a *UserAntisocialCheck) Create(ctx context.Context) (int64, error) {
	q := factory.DB(ctx)
	return q.Insert(a)
}

func (a *UserAntisocialCheck) Update(ctx context.Context) error {
	_, err := factory.DB(ctx).Where("check_user_id = ?", a.CheckUserId).UseBool("check_status").Update(a)
	return err
}

func (UserAntisocialCheck) GetById(ctx context.Context, id int64) (*UserAntisocialCheck, error) {
	var v UserAntisocialCheck
	if _, err := factory.DB(ctx).ID(id).Get(&v); err != nil {
		return nil, err
	}
	return &v, nil
}

// Deprecated: Use UserAntisocialCheckService.GetByUserId instead
func (UserAntisocialCheck) GetByCheckUserId(ctx context.Context, id int64) (*UserAntisocialCheck, error) {
	var v UserAntisocialCheck

	if has, err := factory.DB(ctx).Where("check_user_id = ?", id).Get(&v); err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}
	var err error
	v.CheckFiles, err = AntisocialCheckFile{}.FindByCheckId(ctx, 10, id)
	if err != nil {
		return nil, err
	}
	return &v, nil
}
