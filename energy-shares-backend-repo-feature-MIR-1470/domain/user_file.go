package domain

import (
	"context"
	"regexp"
	"strings"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/s3utils"
)

const USER_FILE = "user_file"

const MY_NUMBER_USER_FILE = "my_number"

type UserFile struct {
	Id            int64              `json:"id" swagger:"desc(프로젝트 이미지 고유 ID)"`
	UserId        int64              `json:"user_id" xorm:"NOT NULL" swagger:"desc(유저 ID)"`
	FileName      string             `json:"file_name" xorm:"UNIQUE NOT NULL" swagger:"desc(파일명)"`
	OriginalName  string             `json:"original_name" xorm:"NOT NULL" swagger:"desc(오리지날 파일명)"`
	FileType      USER_FILE_TYPE     `json:"file_type" xorm:"NOT NULL" swagger:"desc(첨부파일 타입,공통코드 [USER_FILE_TYPE])"`
	FileSize      int64              `json:"file_size" xorm:"NOT NULL" swagger:"desc(파일 사이즈)"`
	FileExtension string             `json:"file_extension" xorm:"NOT NULL" swagger:"desc(파일 확장자)"`
	FilePath      string             `json:"file_path" xorm:"NOT NULL" swagger:"desc(파일 경로)"`
	FileUrl       string             `json:"file_url" xorm:"NOT NULL" swagger:"desc(URL)"`
	OwnerId       int64              `json:"owner_id" xorm:"NOT NULL" swagger:"desc(소유주 ID)"`
	UpdateId      int64              `json:"update_id"  swagger:"desc(최근 업데이트 사용자)"`
	UserType      USER_TYPE          `json:"user_type" xorm:"NOT NULL" swagger:"desc(유저 타입 공통코드 [USER_TYPE])"`
	Etc           string             `json:"etc" swagger:"기타 항목"`
	Created       jsonutils.JSONTime `json:"created" xorm:"created"`
	Updated       jsonutils.JSONTime `json:"updated" xorm:"updated"`
	Deleted       jsonutils.JSONTime `json:"deleted" xorm:"deleted"`
}

type UserFiles []UserFile

func (userFiles UserFiles) ToPtrSlice() []*UserFile {
	var ptrSlice []*UserFile
	for i := range userFiles {
		ptrSlice = append(ptrSlice, &userFiles[i])
	}
	return ptrSlice
}

func (userFiles UserFiles) GetCDNSignUrl() {
	for i := range userFiles {
		key := (&userFiles[i]).GetKey()
		userFiles[i].FileUrl = s3utils.MakeCDNSignURL(key)
	}
}

func (u UserFile) IsEmpty() bool {
	return u.Id == 0
}

func (u UserFile) TableName() string {
	return "user_file"
}

func (u UserFile) SelectAll() string {
	return u.TableName() + ".*"
}

func (uf *UserFile) GetKey() string {
	return uf.FilePath + uf.FileName + uf.FileExtension
}

func (uf *UserFile) GetCDNUrl() string {
	return s3utils.MakeCDNURL(uf.GetKey())
}

func (uf *UserFile) GetCDNSignUrl() string {
	return s3utils.MakeCDNSignURL(uf.GetKey())
}

func (uf *UserFile) GetFileType() string {
	return USER_FILE
}

func (uf *UserFile) GetFileName() string {
	return uf.OriginalName
}

func (uf *UserFile) SetLocation(loc string) {
	uf.FileUrl = loc
}

func (uf *UserFile) Create(ctx context.Context) (int64, error) {
	return factory.DB(ctx).Insert(uf)
}

func (uf *UserFile) Update(ctx context.Context) (err error) {
	_, err = factory.DB(ctx).ID(uf.Id).Update(uf)
	return err
}

func (UserFile) GetById(ctx context.Context, id int64) (*UserFile, error) {
	var v UserFile
	if has, err := factory.DB(ctx).ID(id).Get(&v); err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}
	// DONE URL 생성 로직 추가
	v.FileUrl = v.GetCDNSignUrl()
	return &v, nil
}

func (UserFile) FindOneUserIdByAndFileType(ctx context.Context, id int64, fileType USER_FILE_TYPE) (*UserFile, error) {
	var v UserFile
	if has, err := factory.DB(ctx).Where("user_id = ? AND file_type=?", id, fileType).Desc("created").Get(&v); err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}
	v.FileUrl = v.GetCDNSignUrl()
	return &v, nil
}

func (UserFile) FindByUserIdAndFileType(ctx context.Context, limit int, id int64, fileType USER_FILE_TYPE) ([]*UserFile, error) {
	q := factory.DB(ctx)

	items := make([]*UserFile, 0)
	_, err := q.Where("user_id = ? AND file_type=?", id, fileType).Desc("created").Limit(limit).FindAndCount(&items)
	if err != nil {
		return nil, err
	}
	// DONE URL 생성 로직 추가
	for i := range items {
		items[i].FileUrl = items[i].GetCDNSignUrl()
	}
	return items, nil
}

func (UserFile) GetOneMyNumberFile(ctx context.Context, userId int64, file_type USER_FILE_TYPE, myNumberId string) (*UserFile, error) {
	var userFile UserFile
	if has, err := factory.DB(ctx).Where("user_id = ? AND file_type=? AND etc=?", userId, file_type, myNumberId).Get(&userFile); err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}
	// DONE URL 생성 로직 추가
	userFile.FileUrl = userFile.GetCDNSignUrl()
	return &userFile, nil
}

// FindByIds は、IDで書類をまとめて取る(論理削除済みは含まない)。MIR-1403で、削除の指定が対象の会員の書類かを
// 確かめるために使う。
func (UserFile) FindByIds(ctx context.Context, ids []int64) ([]UserFile, error) {
	items := make([]UserFile, 0)
	if len(ids) == 0 {
		return items, nil
	}
	if err := factory.DB(ctx).In("id", ids).Find(&items); err != nil {
		return nil, err
	}
	return items, nil
}

func (UserFile) Delete(ctx context.Context, id int64) (err error) {
	_, err = factory.DB(ctx).ID(id).Delete(&UserFile{})
	return err
}

func (p *UserFile) CreateMyNumberDB(ctx context.Context) (int64, error) {
	return factory.MyNumberDB(ctx).Insert(p)
}

func UserFileTypePicker(name string) USER_FILE_TYPE {
	const (
		IDENTIFICATION1_FRONT_FILE = "identification1_front_file" // 담당자 본인 확인 서류 1
		IDENTIFICATION1_BACK_FILE  = "identification1_back_file"
		IDENTIFICATION2_FRONT_FILE = "identification2_front_file" // 담당자 본인 확인 서류 2
		IDENTIFICATION2_BACK_FILE  = "identification2_back_file"
		EKYC_IDENTIFICATION_FILES  = "ekyc_identification_files"
		CERTIFIED_COPY_FILES       = "certified_copy_files" // 이력사항 전부 증명서
		POA_FILES                  = "poa_files"            // 위임장 (power of attorney)

	)

	// name: identification_files[0] -> identification_files
	re := regexp.MustCompile(`\[.*\]$`)
	name = re.ReplaceAllString(name, "")

	if strings.EqualFold(name, IDENTIFICATION1_FRONT_FILE) {
		return IDENTIFICATION1_FRONT
	}
	if strings.EqualFold(name, IDENTIFICATION1_BACK_FILE) {
		return IDENTIFICATION1_BACK
	}
	if strings.EqualFold(name, IDENTIFICATION2_FRONT_FILE) {
		return IDENTIFICATION2_FRONT
	}
	if strings.EqualFold(name, IDENTIFICATION2_BACK_FILE) {
		return IDENTIFICATION2_BACK
	}
	if strings.EqualFold(name, EKYC_IDENTIFICATION_FILES) {
		return EKYC_IDENTIFICATION
	}
	if strings.EqualFold(name, CERTIFIED_COPY_FILES) {
		return CERTIFIED_COPY
	}
	if strings.EqualFold(name, POA_FILES) {
		return POA
	}

	return ""

}
