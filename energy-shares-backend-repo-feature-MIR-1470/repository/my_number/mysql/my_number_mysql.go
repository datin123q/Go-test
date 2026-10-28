package mysql

import (
	"context"

	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/orm"
)

type myNumberRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.MyNumberRepository {
	return &myNumberRepository{
		dbFactory: dbFactory,
	}
}

func (repo myNumberRepository) Create(ctx context.Context, myNumber domain.MyNumber) (int64, error) {
	db, err := repo.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "err get db")
	}

	if _, err = db.Insert(&myNumber); err != nil {
		return 0, errors.Wrap(err, "err insert myNumber")
	}

	return myNumber.Id, nil
}

func (repo myNumberRepository) GetByUserId(ctx context.Context, userId int64) (domain.MyNumber, error) {
	db, err := repo.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return domain.MyNumber{}, errors.Wrap(err, "err get db")
	}

	var myNumber domain.MyNumber
	has, err := orm.Table(db, domain.MyNumber{}).
		Equal(domain.MyNumber{}.TableName()+".user_id", userId).
		Get(&myNumber)

	if err != nil {
		return domain.MyNumber{}, errors.Wrap(err, "failed to find by user_id")
	}
	if !has {
		return domain.MyNumber{}, cdomain.ErrItemNotFound
	}

	return myNumber, nil
}

func (repo myNumberRepository) GetById(ctx context.Context, id int64) (domain.MyNumber, error) {
	db, err := repo.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return domain.MyNumber{}, errors.Wrap(err, "err get db")
	}

	var myNumber domain.MyNumber
	if has, err := db.ID(id).Get(&myNumber); err != nil {
		return domain.MyNumber{}, errors.Wrap(err, "err get")
	} else if !has {
		return domain.MyNumber{}, cdomain.ErrItemNotFound
	}

	return myNumber, nil
}

func (repo myNumberRepository) Delete(ctx context.Context, id int64) error {
	db, err := repo.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "err get db")
	}

	if _, err = db.ID(id).Delete(&domain.MyNumber{}); err != nil {
		return errors.Wrap(err, "failed to delete")
	}

	return nil
}
