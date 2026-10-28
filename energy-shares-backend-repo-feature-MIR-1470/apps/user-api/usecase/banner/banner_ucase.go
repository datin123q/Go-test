package banner

import (
	"context"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/samber/lo"

	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	udomain "github.com/illuminarean/energyshares-user-api/domain"
)

type bannerUsecase struct {
	bannerRepository        cdomain.BannerRepository
	bannerSettingRepository cdomain.BannerSettingRepository
}

func New(
	bannerRepository cdomain.BannerRepository,
	bannerSettingRepository cdomain.BannerSettingRepository,
) udomain.BannerUsecase {
	return &bannerUsecase{
		bannerRepository:        bannerRepository,
		bannerSettingRepository: bannerSettingRepository,
	}
}

func (b bannerUsecase) GetAll(ctx context.Context) (udomain.BannerGetAllResponse, error) {

	today := time.Now()

	// 현재 게시 중인 배너 목록을 sort_order 순으로 조회
	opts := cdomain.BannerQueryOptionFunc()
	totalCount, banners, err := b.bannerRepository.GetAll(ctx,
		opts.WithSortBy("sort_order"),
		opts.WithOrder("asc"),
		opts.WithOffset(0),
		opts.WithLimit(10),
		opts.WithIsPosted(lo.ToPtr(true)),
		opts.WithLTEStartDate(today.Format(time.DateTime)),
		opts.WithGTEEndDate(today.Format(time.DateTime)),
	)
	if err != nil {
		return udomain.BannerGetAllResponse{}, errors.Wrap(err, "failed to get all banners")
	}

	items := lo.Map(banners, func(banner domain.Banner, _ int) udomain.BannerGetAllItem {
		return udomain.BannerGetAllItem{
			Id:        banner.Id,
			Title:     banner.Title,
			IsPosted:  banner.IsPosted,
			SortOrder: lo.FromPtr(banner.SortOrder),
			Url:       banner.Url,
			NewTab:    banner.NewTab,
		}
	})

	// get rolling time
	bannerSetting, err := b.bannerSettingRepository.GetByKey(ctx, domain.ROLLING_SECOND)
	if err != nil {
		return udomain.BannerGetAllResponse{}, errors.Wrap(err, "failed to get banner setting")
	}
	rollingSec, err := strconv.ParseInt(bannerSetting.SettingValue, 10, 64)
	if err != nil {
		return udomain.BannerGetAllResponse{}, errors.Wrap(err, "failed to parse banner setting")
	}

	// return response
	return udomain.BannerGetAllResponse{
		TotalCount:    totalCount,
		RollingSecond: rollingSec,
		Items:         items,
	}, nil
}
