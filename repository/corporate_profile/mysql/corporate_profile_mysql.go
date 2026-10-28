package mysql

import (
	"context"

	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/orm"
)

type corporateProfileMysqlRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.CorporateProfileRepository {
	return &corporateProfileMysqlRepository{
		dbFactory: dbFactory,
	}
}

func (c corporateProfileMysqlRepository) Create(ctx context.Context, corporateProfile domain.CorporateProfile) (int64, error) {
	db, err := c.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "err get db")
	}

	if _, err = db.InsertOne(&corporateProfile); err != nil {
		return 0, errors.Wrap(err, "err insert corporate profile")
	}

	return corporateProfile.Id, nil
}

func (c corporateProfileMysqlRepository) GetById(ctx context.Context, id int64) (domain.CorporateProfile, error) {
	db, err := c.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return domain.CorporateProfile{}, errors.Wrap(err, "err get db")
	}

	var corporateProfile domain.CorporateProfile

	has, err := orm.Table(db, domain.CorporateProfile{}).
		Equal(domain.CorporateProfile{}.TableName()+".id", id).
		Get(&corporateProfile)
	if err != nil {
		return domain.CorporateProfile{}, errors.Wrap(err, "err get corporate profile")
	}

	if !has {
		return domain.CorporateProfile{}, errors.Wrap(cdomain.ErrItemNotFound, "corporate profile not found")
	}

	return corporateProfile, nil
}

func (c corporateProfileMysqlRepository) GetByCorpNo(ctx context.Context, corpNo string) (domain.CorporateProfile, error) {
	db, err := c.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return domain.CorporateProfile{}, errors.Wrap(err, "err get db")
	}

	var corporateProfile domain.CorporateProfile

	has, err := orm.Table(db, domain.CorporateProfile{}).
		Equal(domain.CorporateProfile{}.TableName()+".corp_no", corpNo).
		Get(&corporateProfile)
	if err != nil {
		return domain.CorporateProfile{}, errors.Wrap(err, "err get corporate profile")
	}

	if !has {
		return domain.CorporateProfile{}, errors.Wrap(cdomain.ErrItemNotFound, "corporate profile not found")
	}

	return corporateProfile, nil
}

func (c corporateProfileMysqlRepository) GetByUserId(ctx context.Context, userId int64) (domain.CorporateProfile, error) {
	db, err := c.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return domain.CorporateProfile{}, errors.Wrap(err, "err get db")
	}

	var corporateProfile domain.CorporateProfile

	has, err := orm.Table(db, domain.CorporateProfile{}).
		Equal(domain.CorporateProfile{}.TableName()+".user_id", userId).
		Get(&corporateProfile)
	if err != nil {
		return domain.CorporateProfile{}, errors.Wrap(err, "err get corporate profile")
	}

	if !has {
		return domain.CorporateProfile{}, errors.Wrap(cdomain.ErrItemNotFound, "corporate profile not found")
	}

	return corporateProfile, nil
}

func (c corporateProfileMysqlRepository) Update(ctx context.Context, corporateProfile domain.CorporateProfile) error {
	db, err := c.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "err get db")
	}

	_, err = db.ID(corporateProfile.Id).UseBool().Update(&corporateProfile)
	if err != nil {
		return errors.Wrap(err, "err update corporate profile")
	}

	return nil
}
