package mysql

import (
	"context"

	"github.com/pkg/errors"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/orm"
)

type bannerSettingRepository struct {
	dbFactory factory.DBFactory
}

func New(dbFactory factory.DBFactory) cdomain.BannerSettingRepository {
	return &bannerSettingRepository{
		dbFactory: dbFactory,
	}
}

func (r bannerSettingRepository) Create(ctx context.Context, bannerSetting domain.BannerSetting) error {
	db, err := r.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "err get db")
	}

	if _, err = db.Insert(&bannerSetting); err != nil {
		return errors.Wrap(err, "err insert banner setting")
	}

	return nil
}

func (r bannerSettingRepository) GetByKey(ctx context.Context, settingKey domain.BANNER_SETTING_KEY) (domain.BannerSetting, error) {
	db, err := r.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return domain.BannerSetting{}, errors.Wrap(err, "err get db")
	}

	var bannerSetting domain.BannerSetting
	has, err := orm.Table(db, domain.BannerSetting{}).
		Equal("setting_key", settingKey).
		Get(&bannerSetting)
	if err != nil {
		return domain.BannerSetting{}, errors.Wrap(err, "err get")
	}
	if !has {
		return domain.BannerSetting{}, cdomain.ErrItemNotFound
	}
	return bannerSetting, nil
}

func (r bannerSettingRepository) Update(ctx context.Context, bannerSetting domain.BannerSetting) error {
	db, err := r.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "err get db")
	}

	if _, err = db.ID(bannerSetting.Id).Update(&bannerSetting); err != nil {
		return errors.Wrap(err, "err update banner setting")
	}

	return nil
}

func (r bannerSettingRepository) DeleteByKey(ctx context.Context, settingKey domain.BANNER_SETTING_KEY) error {
	db, err := r.dbFactory.EnergySharesDB(ctx)
	if err != nil {
		return errors.Wrap(err, "err get db")
	}

	if _, err := db.Delete(&domain.BannerSetting{SettingKey: string(settingKey)}); err != nil {
		return errors.Wrap(err, "err delete")
	}

	return nil
}
