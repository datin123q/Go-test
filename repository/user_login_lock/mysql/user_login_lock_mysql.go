package mysql

import (
	"context"
	"time"

	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/orm"
)

type userLoginLockMysqlRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.UserLoginLockRepository {
	return &userLoginLockMysqlRepository{
		dbFactory: dbFactory,
	}
}

func (u userLoginLockMysqlRepository) GetAll(ctx context.Context, options ...cdomain.UserLoginLockQueryOption) (int64, []domain.UserLoginLock, error) {
	db, err := u.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, nil, errors.Wrap(err, "failed to get db")
	}

	opts := cdomain.UserLoginLockQueryOptionFunc()
	for _, opt := range options {
		opt(&opts)
	}

	loginLocks := make([]domain.UserLoginLock, 0)

	now := time.Now()
	_ = now
	totalCount, err := orm.Table(db, domain.UserLoginLock{}).
		Equal(domain.UserLoginLock{}.TableName()+".user_id", opts.UserId()).
		Between(domain.UserLoginLock{}.TableName()+".created", opts.StartDate(), opts.EndDate()).
		SortOrder(opts.SortBy(), opts.Order()).
		Limit(opts.Limit(), opts.Offset()).
		FindAndCount(&loginLocks)
	if err != nil {
		return 0, nil, errors.Wrap(err, "failed to get user login locks")
	}

	return totalCount, loginLocks, nil
}

func (u userLoginLockMysqlRepository) GetByUserId(ctx context.Context, userId int64) (domain.UserLoginLock, error) {
	db, err := u.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return domain.UserLoginLock{}, errors.Wrap(err, "failed to get db")
	}

	loginLock := domain.UserLoginLock{}

	has, err := orm.Table(db, domain.UserLoginLock{}).
		Equal(domain.UserLoginLock{}.TableName()+".user_id", userId).
		Get(&loginLock)
	if err != nil {
		return domain.UserLoginLock{}, errors.Wrap(err, "failed to get user login lock")
	}

	if !has {
		return domain.UserLoginLock{}, cdomain.ErrItemNotFound
	}

	return loginLock, nil
}

func (u userLoginLockMysqlRepository) Create(ctx context.Context, loginLock domain.UserLoginLock) (int64, error) {
	db, err := u.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "failed to get db")
	}

	if _, err = db.Insert(&loginLock); err != nil {
		return 0, errors.Wrap(err, "failed to insert user login lock")
	}

	return loginLock.Id, nil
}

func (u userLoginLockMysqlRepository) Update(ctx context.Context, loginLock domain.UserLoginLock) error {
	db, err := u.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get db")
	}

	affected, err := db.ID(loginLock.Id).Update(&loginLock)
	if err != nil {
		return errors.Wrap(err, "failed to update user login lock")
	}
	if affected == 0 {
		return cdomain.ErrItemNotAffected
	}

	return nil
}

func (u userLoginLockMysqlRepository) Delete(ctx context.Context, id int64) error {
	db, err := u.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get db")
	}

	affected, err := db.ID(id).Delete(domain.UserLoginLock{})
	if err != nil {
		return errors.Wrap(err, "failed to delete user login lock")
	}
	if affected == 0 {
		return cdomain.ErrItemNotAffected
	}

	return nil
}
