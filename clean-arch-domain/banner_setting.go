package clean_arch_domain

import (
	"context"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
)

type BannerSettingRepository interface {
	Create(ctx context.Context, bannerSetting domain.BannerSetting) error
	GetByKey(ctx context.Context, settingKey domain.BANNER_SETTING_KEY) (domain.BannerSetting, error)
	Update(ctx context.Context, bannerSetting domain.BannerSetting) error
	DeleteByKey(ctx context.Context, settingKey domain.BANNER_SETTING_KEY) error
}
