package mysql

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/orm"

	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

type faqRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.FaqRepository {
	return &faqRepository{
		dbFactory: dbFactory,
	}
}

func (f faqRepository) Create(ctx context.Context, faq domain.Faq) (int64, error) {
	db, err := f.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "failed to get db")
	}

	_, err = db.InsertOne(&faq)
	if err != nil {
		return 0, errors.Wrap(err, "failed to insert")
	}

	return faq.Id, nil

}

func (f faqRepository) Update(ctx context.Context, faq domain.Faq) error {
	db, err := f.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get db")
	}

	_, err = db.ID(faq.Id).UseBool().Update(&faq)
	if err != nil {
		return errors.Wrap(err, "failed to update")
	}

	return nil
}

func (f faqRepository) GetAll(ctx context.Context, options ...cdomain.FaqQueryOption) (int64, []domain.Faq, error) {
	db, err := f.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, nil, errors.Wrap(err, "failed to get db")
	}

	queryOpts := cdomain.FaqQueryOptionFunc()
	for _, opt := range options {
		opt(&queryOpts)
	}

	var faqs []domain.Faq

	count, err := orm.Table(db, domain.Faq{}).
		Like(domain.Faq{}.TableName()+".title", queryOpts.Title()).
		Equal(domain.Faq{}.TableName()+".category_type", queryOpts.CategoryType()).
		Equal(domain.Faq{}.TableName()+".posted", queryOpts.Posted()).
		Equal(domain.Faq{}.TableName()+".pin", queryOpts.Pin()).
		Before("created", queryOpts.EndDate()).
		After("created", queryOpts.StartDate()).
		SortOrder(queryOpts.SortBy(), queryOpts.Order()).
		Limit(queryOpts.Limit(), queryOpts.Offset()).
		FindAndCount(&faqs)
	if err != nil {
		return 0, nil, errors.Wrap(err, "err find and count")
	}

	return count, faqs, nil
}

func (f faqRepository) GetById(ctx context.Context, id int64) (domain.Faq, error) {
	db, err := f.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return domain.Faq{}, errors.Wrap(err, "failed to get db")
	}

	var faq domain.Faq

	has, err := orm.Table(db, domain.Faq{}).Id(id).Get(&faq)
	if err != nil {
		return domain.Faq{}, errors.Wrap(err, "failed to find by id")
	}
	if !has {
		return domain.Faq{}, cdomain.ErrItemNotFound
	}

	return faq, nil
}

func (f faqRepository) Delete(ctx context.Context, id int64) error {
	db, err := f.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get db")
	}

	if _, err = db.ID(id).Delete(&domain.Faq{}); err != nil {
		return errors.Wrap(err, "failed to delete")
	}

	return nil
}
