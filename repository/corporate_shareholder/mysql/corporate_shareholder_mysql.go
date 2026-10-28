package mysql

import (
	"context"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/orm"
	"github.com/pkg/errors"
)

type corporateShareholderMysqlRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.CorporateShareholderRepository {
	return &corporateShareholderMysqlRepository{
		dbFactory: dbFactory,
	}
}

func (c corporateShareholderMysqlRepository) Create(ctx context.Context, corporateShareholder domain.CorporateShareholder) (int64, error) {
	db, err := c.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "err get db")
	}

	if _, err = db.InsertOne(&corporateShareholder); err != nil {
		return 0, errors.Wrap(err, "err insert corporate shareholder")
	}

	return corporateShareholder.Id, nil
}

func (c corporateShareholderMysqlRepository) GetAll(ctx context.Context, options ...cdomain.CorporateShareholderOption) (int64, []domain.CorporateShareholder, error) {
	db, err := c.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, []domain.CorporateShareholder{}, errors.Wrap(err, "err get db")
	}

	opts := cdomain.CorporateShareholderQueryOptionFunc()
	for _, opt := range options {
		opt(&opts)
	}

	corporateShareholders := make([]domain.CorporateShareholder, 0)

	count, err := orm.Table(db, domain.CorporateShareholder{}).
		Equal(domain.CorporateShareholder{}.TableName()+".user_id", opts.UserId()).
		SortOrder(opts.SortBy(), opts.Order()).
		FindAndCount(&corporateShareholders)
	if err != nil {
		return 0, []domain.CorporateShareholder{}, errors.Wrap(err, "err get corporate shareholders")
	}

	return count, corporateShareholders, nil
}

func (c corporateShareholderMysqlRepository) DeleteByUserId(ctx context.Context, userId int64) error {
	db, err := c.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "err get db")
	}

	// 実際に削除
	attached, err := db.Unscoped().Delete(&domain.CorporateShareholder{
		UserId: userId,
	})
	if err != nil {
		return errors.Wrap(err, "failed to delete corporate shareholder by user id")
	}

	if attached == 0 {
		return cdomain.ErrItemNotAffected
	}

	return nil
}
