package mysql

import (
	"context"

	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/orm"
)

type privateProfileMysqlRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.PrivateProfileRepository {
	return &privateProfileMysqlRepository{
		dbFactory: dbFactory,
	}
}

func (u privateProfileMysqlRepository) Create(ctx context.Context, privateProfile domain.PrivateProfile) error {
	db, err := u.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return err
	}

	if _, err = db.UseBool().Insert(privateProfile); err != nil {
		return errors.Wrap(err, "err insert privateProfile")
	}

	return nil
}

// GetById - 使用する機会があるだろうか？
func (u privateProfileMysqlRepository) GetById(ctx context.Context, id int64) (domain.PrivateProfile, error) {
	db, err := u.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return domain.PrivateProfile{}, err
	}

	var privateProfile domain.PrivateProfile
	has, err := orm.Table(db, domain.PrivateProfile{}).Id(id).Get(&privateProfile)
	if err != nil {
		return domain.PrivateProfile{}, errors.Wrap(err, "failed to get private profile")
	}
	if !has {
		return domain.PrivateProfile{}, cdomain.ErrItemNotFound
	}

	return privateProfile, nil
}

func (u privateProfileMysqlRepository) GetByUserId(ctx context.Context, userId int64) (domain.PrivateProfile, error) {
	db, err := u.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return domain.PrivateProfile{}, errors.Wrap(err, "failed to get db")
	}

	var privateProfile domain.PrivateProfile

	has, err := orm.Table(db, domain.PrivateProfile{}).
		Equal(domain.PrivateProfile{}.TableName()+".user_id", userId).
		Get(&privateProfile)
	if err != nil {
		return domain.PrivateProfile{}, errors.Wrap(err, "failed to get private profile")
	}

	if !has {
		return domain.PrivateProfile{}, cdomain.ErrItemNotFound
	}

	return privateProfile, nil
}

func (u privateProfileMysqlRepository) Update(ctx context.Context, privateProfile domain.PrivateProfile) error {
	db, err := u.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return err
	}

	_, err = db.ID(privateProfile.Id).UseBool().Update(&privateProfile)
	if err != nil {
		return errors.Wrap(err, "failed to update private profile")
	}

	return nil
}
