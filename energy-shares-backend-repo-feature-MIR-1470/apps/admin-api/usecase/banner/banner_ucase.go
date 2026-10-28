package banner

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/samber/lo"

	adomain "github.com/illuminarean/energy-share-core-api/domain"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
)

type bannerUsecase struct {
	bannerRepository        cdomain.BannerRepository
	bannerSettingRepository cdomain.BannerSettingRepository
}

func New(
	bannerRepository cdomain.BannerRepository,
	bannerSettingRepository cdomain.BannerSettingRepository,
) adomain.BannerUsecase {
	return &bannerUsecase{
		bannerRepository:        bannerRepository,
		bannerSettingRepository: bannerSettingRepository,
	}
}

func (b bannerUsecase) Create(ctx context.Context, request adomain.BannerCreateRequest) (adomain.BannerCreateResponse, error) {
	startDate, err := jsonutils.ParseJSONTimeJST(time.DateTime, request.StartDate)
	if err != nil {
		return adomain.BannerCreateResponse{}, errors.Wrap(err, "failed to parse startDate")
	}
	endDate, err := jsonutils.ParseJSONTimeJST(time.DateTime, request.EndDate)
	if err != nil {
		return adomain.BannerCreateResponse{}, errors.Wrap(err, "failed to parse endDate")
	}

	// create banner
	banner := domain.Banner{
		Title:       request.Title,
		IsPosted:    lo.FromPtr(request.IsPosted),
		StartDate:   startDate,
		EndDate:     endDate,
		Url:         request.Url,
		SortOrder:   request.SortOrder,
		Description: lo.FromPtr(request.Description),
		NewTab:      lo.FromPtr(request.NewTab),
	}

	banner.Id, err = b.bannerRepository.Create(ctx, banner)
	if err != nil {
		return adomain.BannerCreateResponse{}, errors.Wrap(err, "failed to create banner")
	}

	// return response
	return adomain.BannerCreateResponse{
		Id:          banner.Id,
		Title:       banner.Title,
		IsPosted:    banner.IsPosted,
		StartDate:   banner.StartDate,
		EndDate:     banner.EndDate,
		Url:         banner.Url,
		SortOrder:   banner.SortOrder,
		Description: banner.Description,
		NewTab:      banner.NewTab,
		Created:     banner.Created,
		Updated:     banner.Updated,
	}, nil
}

func (b bannerUsecase) GetAll(ctx context.Context, request adomain.BannerGetAllRequest) (adomain.BannerGetAllResponse, error) {
	var err error

	// check time type
	if request.StartDate != nil {
		_, err = jsonutils.ParseJSONTimeJST(time.DateTime, lo.FromPtr(request.StartDate))
		if err != nil {
			return adomain.BannerGetAllResponse{}, errors.Wrap(err, "failed to parse startDate")
		}
	}

	if request.EndDate != nil {
		_, err = jsonutils.ParseJSONTimeJST(time.DateTime, lo.FromPtr(request.EndDate))
		if err != nil {
			return adomain.BannerGetAllResponse{}, errors.Wrap(err, "failed to parse endDate")
		}
	}

	if request.CreatedStartDate != nil {
		_, err = jsonutils.ParseJSONTimeJST(time.DateTime, lo.FromPtr(request.CreatedStartDate))
		if err != nil {
			return adomain.BannerGetAllResponse{}, errors.Wrap(err, "failed to parse createdStartDate")
		}
	}

	if request.CreatedEndDate != nil {
		_, err = jsonutils.ParseJSONTimeJST(time.DateTime, lo.FromPtr(request.CreatedEndDate))
		if err != nil {
			return adomain.BannerGetAllResponse{}, errors.Wrap(err, "failed to parse createdEndDate")
		}
	}

	// 정렬 조건에서 posted -> is_posted 로 변경
	request.Sortby = strings.Replace(request.Sortby, "posted", "is_posted", -1)

	opts := cdomain.BannerQueryOptionFunc()
	totalCount, banners, err := b.bannerRepository.GetAll(ctx,
		opts.WithSortBy(request.Sortby),
		opts.WithOrder(request.Order),
		opts.WithOffset(request.SkipCount),
		opts.WithLimit(request.MaxResultCount),
		opts.WithTitle(lo.FromPtr(request.Title)),
		opts.WithIsPosted(request.IsPosted),
		opts.WithGTEEndDate(lo.FromPtr(request.StartDate)),
		opts.WithLTEStartDate(lo.FromPtr(request.EndDate)),
		opts.WithCreatedStartDate(lo.FromPtr(request.CreatedStartDate)),
		opts.WithCreatedEndDate(lo.FromPtr(request.CreatedEndDate)),
	)
	if err != nil {
		return adomain.BannerGetAllResponse{}, errors.Wrap(err, "failed to get all banners")
	}

	items := lo.Map(banners, func(banner domain.Banner, _ int) adomain.BannerGetAllItem {
		return adomain.BannerGetAllItem{
			Id:        banner.Id,
			Title:     banner.Title,
			IsPosted:  banner.IsPosted,
			SortOrder: banner.SortOrder,
			StartDate: banner.StartDate,
			EndDate:   banner.EndDate,
			Created:   banner.Created,
		}
	})

	// count posted banners
	totalBannerPosted := lo.CountBy(banners, func(banner domain.Banner) bool {
		return banner.IsPosted
	})

	// get rolling time
	bannerSetting, err := b.bannerSettingRepository.GetByKey(ctx, domain.ROLLING_SECOND)
	if err != nil {
		return adomain.BannerGetAllResponse{}, errors.Wrap(err, "failed to get banner setting")
	}
	rollingSec, err := strconv.ParseInt(bannerSetting.SettingValue, 10, 64)
	if err != nil {
		return adomain.BannerGetAllResponse{}, errors.Wrap(err, "failed to parse banner setting")
	}

	// return response
	return adomain.BannerGetAllResponse{
		TotalCount:        totalCount,
		TotalBannerPosted: int64(totalBannerPosted),
		RollingSecond:     rollingSec,
		Items:             items,
	}, nil
}

func (b bannerUsecase) GetOne(ctx context.Context, request adomain.BannerGetOneRequest) (adomain.BannerGetOneResponse, error) {

	banner, err := b.bannerRepository.GetById(ctx, request.Id)
	if err != nil {
		return adomain.BannerGetOneResponse{}, errors.Wrap(err, "failed to get banner")
	}

	return adomain.BannerGetOneResponse{
		Id:          banner.Id,
		Title:       banner.Title,
		IsPosted:    banner.IsPosted,
		StartDate:   banner.StartDate,
		EndDate:     banner.EndDate,
		Url:         banner.Url,
		SortOrder:   banner.SortOrder,
		Description: banner.Description,
		NewTab:      banner.NewTab,
		Created:     banner.Created,
		Updated:     banner.Updated,
	}, nil
}

func (b bannerUsecase) Update(ctx context.Context, request adomain.BannerUpdateRequest) (adomain.BannerUpdateResponse, error) {

	// string to time
	startDate, err := jsonutils.ParseJSONTimeJST(time.DateTime, request.StartDate)
	if err != nil {
		return adomain.BannerUpdateResponse{}, errors.Wrap(err, "failed to parse startDate")
	}
	endDate, err := jsonutils.ParseJSONTimeJST(time.DateTime, request.EndDate)
	if err != nil {
		return adomain.BannerUpdateResponse{}, errors.Wrap(err, "failed to parse endDate")
	}

	// get banner
	banner, err := b.bannerRepository.GetById(ctx, request.Id)
	if err != nil {
		return adomain.BannerUpdateResponse{}, errors.Wrap(err, "failed to get banner")
	}

	// update banner
	banner.Title = request.Title
	banner.IsPosted = lo.FromPtr(request.IsPosted)
	banner.Url = request.Url
	banner.SortOrder = request.SortOrder
	banner.Description = lo.FromPtr(request.Description)
	banner.NewTab = lo.FromPtr(request.NewTab)
	banner.StartDate = startDate
	banner.EndDate = endDate

	if err := b.bannerRepository.Update(ctx, banner); err != nil {
		return adomain.BannerUpdateResponse{}, errors.Wrap(err, "failed to update banner")
	}

	return adomain.BannerUpdateResponse{
		Id:          banner.Id,
		Title:       banner.Title,
		IsPosted:    banner.IsPosted,
		StartDate:   banner.StartDate,
		EndDate:     banner.EndDate,
		Url:         banner.Url,
		SortOrder:   banner.SortOrder,
		Description: banner.Description,
		NewTab:      banner.NewTab,
		Created:     banner.Created,
		Updated:     banner.Updated,
	}, nil
}

func (b bannerUsecase) Delete(ctx context.Context, request adomain.BannerDeleteRequest) error {

	if err := b.bannerRepository.Delete(ctx, request.Id); err != nil {
		return errors.Wrap(err, "failed to delete banner")
	}

	return nil
}

func (b bannerUsecase) UpdateRollingTime(ctx context.Context, request adomain.BannerUpdateRollingTimeRequest) error {

	// get banner setting - ROLLING_SECOND
	rollingTime, err := b.bannerSettingRepository.GetByKey(ctx, domain.ROLLING_SECOND)
	if err != nil {
		return errors.Wrap(err, "failed to get banner setting")
	}

	// set rolling time
	rollingTime.SettingValue = strconv.FormatInt(request.RollingSecond, 10)

	// update banner setting - ROLLING_SECOND
	if err := b.bannerSettingRepository.Update(ctx, rollingTime); err != nil {
		return errors.Wrap(err, "failed to update banner setting")
	}

	return nil
}
