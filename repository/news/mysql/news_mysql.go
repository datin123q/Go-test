package mysql

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/orm"

	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
)

type newsRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.NewsRepository {
	return &newsRepository{
		dbFactory: dbFactory,
	}
}

func (repo newsRepository) Create(ctx context.Context, news domain.News) (int64, error) {
	db, err := repo.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "failed to get db")
	}

	if _, err = db.Insert(&news); err != nil {
		return 0, errors.Wrap(err, "failed to insert news")
	}

	return news.Id, nil
}

func (repo newsRepository) Update(ctx context.Context, news domain.News) error {
	db, err := repo.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get db")
	}

	if _, err := db.ID(news.Id).UseBool().Update(&news); err != nil {
		return errors.Wrap(err, "failed to update news")
	}

	return nil
}

func (repo newsRepository) GetAll(ctx context.Context, options ...cdomain.NewsQueryOption) (int64, []domain.News, error) {
	db, err := repo.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, nil, errors.Wrap(err, "failed to get db")
	}

	queryOpts := cdomain.NewsQueryOptionFunc()
	for _, opt := range options {
		opt(&queryOpts)
	}

	var newsList []domain.News

	total, err := orm.Table(db, domain.News{}).
		Equal(domain.News{}.TableName()+".pin", queryOpts.Pin()).
		Like(domain.News{}.TableName()+".title", queryOpts.Title()).
		Equal(domain.News{}.TableName()+".posted", queryOpts.Posted()).
		GTE(domain.News{}.TableName()+".closing_date", queryOpts.ClosingDate()).
		Between("created", queryOpts.StartDate(), queryOpts.EndDate()).
		SortOrder(queryOpts.SortBy(), queryOpts.Order()).
		Limit(queryOpts.Limit(), queryOpts.Offset()).
		FindAndCount(&newsList)
	if err != nil {
		return 0, nil, errors.Wrap(err, "err find and count")
	}

	return total, newsList, nil
}

func (repo newsRepository) GetById(ctx context.Context, id int64) (domain.News, error) {
	db, err := repo.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return domain.News{}, errors.Wrap(err, "failed to get db")
	}

	var news domain.News
	has, err := orm.Table(db, domain.News{}).Id(id).Get(&news)
	if err != nil {
		return domain.News{}, errors.Wrap(err, "failed to find by id")
	}
	if !has {
		return domain.News{}, cdomain.ErrItemNotFound
	}

	return news, nil
}

func (repo newsRepository) Delete(ctx context.Context, id int64) error {
	db, err := repo.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get db")
	}

	if _, err := db.ID(id).Delete(&domain.News{}); err != nil {
		return errors.Wrap(err, "failed to delete news")
	}

	return nil
}
