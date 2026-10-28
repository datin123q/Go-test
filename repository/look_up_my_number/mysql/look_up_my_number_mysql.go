package mysql

import (
	"context"

	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/orm"
)

type lookUpMyNumberRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.LookUpMyNumberRepository {
	return &lookUpMyNumberRepository{
		dbFactory: dbFactory,
	}
}

func (repo lookUpMyNumberRepository) Create(ctx context.Context, lookUpMyNumber domain.LookUpMyNumber) (int64, error) {
	db, err := repo.dbFactory.MyNumberDB(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "err get db")
	}

	if _, err = db.Insert(&lookUpMyNumber); err != nil {
		return 0, errors.Wrap(err, "err insert lookUpMyNumber")
	}

	return lookUpMyNumber.Id, nil
}

// GetByMyNumberId は GetByMyNumberIdIncludeDeleted と非常に似ていますが、メソッド規約を維持するために別々に構成されています。
func (repo lookUpMyNumberRepository) GetByMyNumberId(ctx context.Context, myNumberId int64) (domain.LookUpMyNumber, error) {
	db, err := repo.dbFactory.MyNumberDB(ctx)
	if err != nil {
		return domain.LookUpMyNumber{}, errors.Wrap(err, "failed to get db")
	}

	var lookUpMyNumber domain.LookUpMyNumber
	has, err := orm.Table(db, domain.LookUpMyNumber{}).
		Equal(domain.LookUpMyNumber{}.TableName()+".my_number_id", myNumberId).
		Get(&lookUpMyNumber)

	if err != nil {
		return domain.LookUpMyNumber{}, errors.Wrap(err, "failed to get lookUpMyNumber")
	}
	if !has {
		return domain.LookUpMyNumber{}, cdomain.ErrItemNotFound
	}

	return lookUpMyNumber, nil
}

func (repo lookUpMyNumberRepository) GetByMyNumberIdIncludeDeleted(ctx context.Context, myNumberId int64) (domain.LookUpMyNumber, error) {
	db, err := repo.dbFactory.MyNumberDB(ctx)
	if err != nil {
		return domain.LookUpMyNumber{}, errors.Wrap(err, "failed to get db")
	}

	var lookUpMyNumber domain.LookUpMyNumber
	has, err := orm.Table(db, domain.LookUpMyNumber{}).
		Equal(domain.LookUpMyNumber{}.TableName()+".my_number_id", myNumberId).
		Unscoped().
		Get(&lookUpMyNumber)

	if err != nil {
		return domain.LookUpMyNumber{}, errors.Wrap(err, "failed to get lookUpMyNumber")
	}
	if !has {
		return domain.LookUpMyNumber{}, cdomain.ErrItemNotFound
	}

	return lookUpMyNumber, nil
}

func (repo lookUpMyNumberRepository) GetByMyNumber(ctx context.Context, myNumber string) (domain.LookUpMyNumber, error) {
	db, err := repo.dbFactory.MyNumberDB(ctx)
	if err != nil {
		return domain.LookUpMyNumber{}, errors.Wrap(err, "failed to get db")
	}

	var lookUpMyNumber domain.LookUpMyNumber
	has, err := orm.Table(db, domain.LookUpMyNumber{}).
		Equal(domain.LookUpMyNumber{}.TableName()+".my_number", myNumber).
		Get(&lookUpMyNumber)

	if err != nil {
		return domain.LookUpMyNumber{}, errors.Wrap(err, "failed to get lookUpMyNumber")
	}
	if !has {
		return domain.LookUpMyNumber{}, cdomain.ErrItemNotFound
	}

	return lookUpMyNumber, nil
}

func (repo lookUpMyNumberRepository) GetByMyNumberIncludeDeleted(ctx context.Context, myNumber string) (domain.LookUpMyNumber, error) {
	db, err := repo.dbFactory.MyNumberDB(ctx)
	if err != nil {
		return domain.LookUpMyNumber{}, errors.Wrap(err, "failed to get db")
	}

	var lookUpMyNumber domain.LookUpMyNumber
	has, err := orm.Table(db, domain.LookUpMyNumber{}).
		Equal(domain.LookUpMyNumber{}.TableName()+".my_number", myNumber).
		Unscoped().
		Get(&lookUpMyNumber)

	if err != nil {
		return domain.LookUpMyNumber{}, errors.Wrap(err, "failed to get lookUpMyNumber")
	}
	if !has {
		return domain.LookUpMyNumber{}, cdomain.ErrItemNotFound
	}

	return lookUpMyNumber, nil
}

func (repo lookUpMyNumberRepository) DeleteByMyNumberId(ctx context.Context, myNumberId int64) error {
	db, err := repo.dbFactory.MyNumberDB(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get db")
	}

	_, err = db.Where(domain.LookUpMyNumber{}.TableName()+".my_number_id = ?", myNumberId).Delete(&domain.LookUpMyNumber{})
	if err != nil {
		return errors.Wrap(err, "failed to delete lookUpMyNumber")
	}

	return nil
}
