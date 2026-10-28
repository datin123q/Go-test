package mysql

import (
	"context"

	"github.com/pkg/errors"
	"github.com/samber/lo"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/orm"
)

type bannerRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.BannerRepository {
	return &bannerRepository{
		dbFactory: dbFactory,
	}
}

func (r bannerRepository) Create(ctx context.Context, banner domain.Banner) (int64, error) {
	db, err := r.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "err get db")
	}

	if _, err = db.Insert(&banner); err != nil {
		return 0, errors.Wrap(err, "err insert banner")
	}

	return banner.Id, nil
}

func (r bannerRepository) GetAll(ctx context.Context, options ...cdomain.BannerQueryOption) (int64, []domain.Banner, error) {
	db, err := r.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return 0, nil, errors.Wrap(err, "err get db")
	}

	queryOpts := cdomain.BannerQueryOptionFunc()
	for _, opt := range options {
		opt(&queryOpts)
	}

	banners := make([]domain.Banner, 0)

	totalCount, err := orm.Table(db, domain.Banner{}).
		Like("title", lo.ToPtr(queryOpts.Title())).
		Equal("is_posted", queryOpts.IsPosted()).
		GTE("start_date", queryOpts.GteStartDate()).
		LTE("start_date", queryOpts.LteStartDate()).
		GTE("end_date", queryOpts.GteEndDate()).
		LTE("end_date", queryOpts.LteEndDate()).
		Between("created", queryOpts.CreatedStartDate(), queryOpts.CreatedEndDate()).
		SortOrder(queryOpts.SortBy(), queryOpts.Order()).
		Limit(queryOpts.Limit(), queryOpts.Offset()).
		FindAndCount(&banners)
	if err != nil {
		return 0, nil, errors.Wrap(err, "err find")
	}

	return totalCount, banners, nil
}

func (r bannerRepository) GetById(ctx context.Context, id int64) (domain.Banner, error) {
	db, err := r.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return domain.Banner{}, errors.Wrap(err, "err get db")
	}

	var banner domain.Banner
	has, err := orm.Table(db, domain.Banner{}).
		Equal("id", id).
		Get(&banner)
	if err != nil {
		return domain.Banner{}, errors.Wrap(err, "err get")
	}
	if !has {
		return domain.Banner{}, cdomain.ErrItemNotFound
	}

	return banner, nil
}

func (r bannerRepository) Update(ctx context.Context, banner domain.Banner) error {
	db, err := r.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "err get db")
	}

	if _, err = db.ID(banner.Id).UseBool().AllCols().Update(&banner); err != nil {
		return errors.Wrap(err, "err update")
	}

	return nil
}

func (r bannerRepository) Delete(ctx context.Context, id int64) error {
	db, err := r.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "err get db")
	}

	if _, err = db.ID(id).Delete(&domain.Banner{}); err != nil {
		return errors.Wrap(err, "err delete")
	}

	return nil
}
