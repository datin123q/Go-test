package mysql

import (
	"context"

	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/orm"
)

type userFileRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.UserFileRepository {
	return &userFileRepository{
		dbFactory: dbFactory,
	}
}

func (repo userFileRepository) Create(ctx context.Context, userFile domain.UserFile) error {
	db, err := repo.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return err
	}

	if _, err = db.Insert(&userFile); err != nil {
		return errors.Wrap(err, "err insert privateProfile")
	}

	return nil
}

func (repo userFileRepository) CreateMany(ctx context.Context, userFiles []domain.UserFile) error {
	db, err := repo.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get db")
	}

	if _, err = db.Insert(&userFiles); err != nil {
		return errors.Wrap(err, "failed to insert userFiles")
	}

	return nil

}

// GetAll は条件に該当するUserFileを取得します。
// 注意: domain.UserFile の FileUrl フィールドは別途入力する必要があります。
func (repo userFileRepository) GetAll(ctx context.Context, options ...cdomain.UserFileQueryOption) (int64, domain.UserFiles, error) {
	db, err := repo.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, nil, errors.Wrap(err, "err get db")
	}

	opts := cdomain.UserFileQueryOptionFunc()
	for _, opt := range options {
		opt(&opts)
	}

	var userFiles domain.UserFiles

	totalCount, err := orm.Table(db, domain.UserFile{}).
		Equal(domain.UserFile{}.TableName()+".id", opts.Id()).
		Equal(domain.UserFile{}.TableName()+".user_id", opts.UserId()).
		Equal(domain.UserFile{}.TableName()+".file_type", opts.FileType()).
		Equal(domain.UserFile{}.TableName()+".etc", opts.Etc()).
		SortOrder(opts.SortBy(), opts.Order()).
		Limit(opts.Limit(), opts.Offset()).
		FindAndCount(&userFiles)
	if err != nil {
		return 0, nil, errors.Wrap(err, "failed to get userFiles")
	}

	return totalCount, userFiles, nil
}

func (repo userFileRepository) ExistsByFileName(ctx context.Context, fileName string) (bool, error) {
	db, err := repo.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return false, errors.Wrap(err, "err get db")
	}

	// unique 制約は deleted を見ないため、論理削除済みの行も数える
	count, err := orm.Table(db, domain.UserFile{}).
		Equal(domain.UserFile{}.TableName()+".file_name", &fileName).
		Unscoped().
		Count()
	if err != nil {
		return false, errors.Wrap(err, "failed to count userFiles by file name")
	}

	return count > 0, nil
}

func (repo userFileRepository) GetById(ctx context.Context, id int64) (*domain.UserFile, error) {
	db, err := repo.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get db")
	}

	var userFile domain.UserFile
	has, err := orm.Table(db, domain.UserFile{}).Id(id).Get(&userFile)
	if err != nil {
		return nil, errors.Wrap(err, "failed to find by id")
	}
	if !has {
		return nil, cdomain.ErrItemNotFound
	}

	return &userFile, nil
}

func (repo userFileRepository) Delete(ctx context.Context, id int64) error {
	db, err := repo.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get db")
	}

	if _, err := db.ID(id).Delete(&domain.UserFile{}); err != nil {
		return errors.Wrap(err, "failed to delete user file")
	}

	return nil
}
