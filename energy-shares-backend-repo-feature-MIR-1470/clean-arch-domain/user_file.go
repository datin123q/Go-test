package clean_arch_domain

import (
	"context"
	"fmt"

	"github.com/samber/lo"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type UserFileRepository interface {
	Create(ctx context.Context, userFile domain.UserFile) error
	CreateMany(ctx context.Context, userFiles []domain.UserFile) error
	GetAll(ctx context.Context, options ...UserFileQueryOption) (int64, domain.UserFiles, error)
	// ExistsByFileName は同じ file_name の行があるかを返す。
	// user_file.file_name はテーブル全体で一意で、制約は deleted を見ないため、
	// 論理削除済みの行も含めて調べる。
	ExistsByFileName(ctx context.Context, fileName string) (bool, error)
	// GetById は id で1件取得する。見つからなければ ErrItemNotFound を返す。
	GetById(ctx context.Context, id int64) (*domain.UserFile, error)
	// Delete は id で論理削除する。
	Delete(ctx context.Context, id int64) error
}

type UserFileService interface {
	CreateUserFile(ctx context.Context, userFile domain.UserFile) (domain.UserFile, error)
	GetPrivateUserIdentificationFiles(ctx context.Context, userId int64) (PrivateUserIdentificationFiles, error)
	GetCorporateUserIdentificationFiles(ctx context.Context, userId int64) (CorporateUserIdentificationFiles, error)
}

type UserFileQueryOption = func(*UserFileQueryOptionFields)

type UserFileQueryOptionFields struct {
	QueryOptions
	id       *int64
	userId   *int64
	fileType *string
	etc      *string
}

func UserFileQueryOptionFunc() UserFileQueryOptionFields {
	return UserFileQueryOptionFields{}
}

func (qo UserFileQueryOptionFields) WithId(id int64) func(*UserFileQueryOptionFields) {
	return func(qo *UserFileQueryOptionFields) {
		qo.id = &id
	}
}

func (qo UserFileQueryOptionFields) WithUserId(userId int64) func(*UserFileQueryOptionFields) {
	return func(qo *UserFileQueryOptionFields) {
		qo.userId = &userId
	}
}

func (qo UserFileQueryOptionFields) WithFileType(fileType domain.USER_FILE_TYPE) func(*UserFileQueryOptionFields) {
	return func(qo *UserFileQueryOptionFields) {
		qo.fileType = lo.ToPtr(string(fileType))
	}
}

func (qo UserFileQueryOptionFields) WithEtc(myNumberId string) func(*UserFileQueryOptionFields) {
	return func(qo *UserFileQueryOptionFields) {
		qo.etc = lo.ToPtr(string(myNumberId))
	}
}

func (qo UserFileQueryOptionFields) WithSortBy(sortBy string) func(*UserFileQueryOptionFields) {
	return func(qo *UserFileQueryOptionFields) {
		qo.sortBy = &sortBy
	}
}

func (qo UserFileQueryOptionFields) WithOrder(order string) func(*UserFileQueryOptionFields) {
	return func(qo *UserFileQueryOptionFields) {
		qo.order = &order
	}
}

func (qo UserFileQueryOptionFields) WithOffset(offset int) func(*UserFileQueryOptionFields) {
	return func(qo *UserFileQueryOptionFields) {
		qo.offset = &offset
	}
}

func (qo UserFileQueryOptionFields) WithLimit(limit int) func(*UserFileQueryOptionFields) {
	return func(qo *UserFileQueryOptionFields) {
		qo.limit = &limit
	}
}

func (qo UserFileQueryOptionFields) Id() *string {
	if qo.id == nil {
		return nil
	}
	return lo.ToPtr(fmt.Sprintf("%v", *qo.id))
}

func (qo UserFileQueryOptionFields) UserId() *string {
	if qo.userId == nil {
		return nil
	}
	return lo.ToPtr(fmt.Sprintf("%v", *qo.userId))
}

func (qo UserFileQueryOptionFields) FileType() *string {
	if qo.fileType == nil {
		return nil
	}
	return qo.fileType
}

func (qo UserFileQueryOptionFields) Etc() *string {
	if qo.etc == nil {
		return nil
	}
	return qo.etc
}

type PrivateUserIdentificationFiles struct {
	Front1 domain.UserFile
	Back1  domain.UserFile
	Front2 domain.UserFile
	Back2  domain.UserFile
}

type CorporateUserIdentificationFiles struct {
	Front1        *domain.UserFile
	Back1         *domain.UserFile
	Front2        *domain.UserFile
	Back2         *domain.UserFile
	CertifiedCopy *domain.UserFile
	POA           *domain.UserFile
	AOI           *domain.UserFile
	Shareholders  *domain.UserFile
	Others        *domain.UserFile
}
